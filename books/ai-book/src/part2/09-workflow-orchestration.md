# 第21章 执行编排与多 Agent 协作：任务图、交接与平台边界

本章消费第11章持久Runtime，不重复定义可靠投递算法。版本与范围：资料核验于2026-10-18，A2A选读固定v0.3.0；以下消息、时序、批处理和控制流程均为教学设计，未运行多Agent或消息系统实验。

## 21.1 先选择确定性工作流还是 Agent

LangGraph文档区分固定Workflow与动态Agent路径；是否增加角色要比较任务收益与协调成本。[[1]](#ref-1)


### 21.1.1 最小 Agent Loop 的边界

最小 Agent Loop 通常长这样：

```python
while not done:
    action = llm(context)
    result = execute_tool(action)
    context.append(result)
```

这个 loop 适合原型，也适合低风险探索任务。它的优点是简单、灵活、实现快。

但它的边界也很明显：

- 控制流隐藏在模型输出里；
- 状态保存在上下文或内存变量里；
- 工具副作用缺少幂等和审计；
- 失败后不知道从哪一步恢复；
- 人工审批只能靠聊天确认；
- trace 不完整，无法稳定复盘；
- 任务越长，上下文越容易膨胀和污染。

最小 loop 的问题不是“模型不够聪明”，而是缺少运行时结构。模型可以决定下一步，但系统必须决定哪些步骤可执行、哪些动作要暂停、哪些状态要持久化、哪些证据必须保留。

### 21.1.2 为什么单 Agent 也需要显式编排

很多人把 workflow 和 state machine 误认为多 Agent 专属能力。其实只要单 Agent 需要执行多步任务，就已经需要显式编排。

例如一个单 Agent 告警诊断任务：

```text
接收告警
  -> 拉取指标
  -> 查询日志
  -> 生成假设
  -> 验证假设
  -> 生成修复建议
  -> 等待人工审批
  -> 执行低风险动作或输出操作手册
```

这里可以只有一个 Agent，但仍然需要：

- workflow：定义步骤顺序和分支；
- state：保存当前诊断进度；
- checkpoint：工具失败后恢复；
- approval：高风险动作前暂停；
- trace：记录证据和决策过程。

单 Agent 加上 workflow，不会让系统变复杂，反而会让复杂任务变得可控。

### 21.1.3 控制流、状态、副作用、人工介入与可观测性

生产级 Agent Runtime 至少要显式处理五类问题。

| 问题 | Demo 做法 | 生产做法 |
|:---|:---|:---|
| 控制流 | 让模型自由决定下一步 | workflow、graph、router、state machine |
| 状态 | 拼进 prompt 或内存变量 | task state、checkpoint、session store |
| 副作用 | 直接调用工具 | tool runtime、幂等键、审批、补偿 |
| 人工介入 | 聊天里问“可以吗” | approval node、interrupt、resume token |
| 可观测性 | 打印日志 | trace、span、evidence、cost、eval replay |

这五类问题决定了 Agent 能不能上线。模型可以生成计划，但系统必须管理执行计划的生命周期。

### 21.1.4 Agent Demo 到 Agent Platform 的演进路径

一个常见演进路径是：

```text
Single LLM Call
  -> Tool-using Agent
  -> Stateful Workflow
  -> Long-running Agent Runtime
  -> Multi-Agent Runtime
  -> Managed Agent Platform
```

每一步都不是为了“更炫”，而是因为出现了新的工程压力：

| 阶段 | 触发条件 | 新增能力 |
|:---|:---|:---|
| Tool-using Agent | 需要查询或执行外部系统 | tool schema、权限、错误处理 |
| Stateful Workflow | 任务超过一步 | task state、控制流、分支 |
| Long-running Runtime | 任务可能失败或暂停 | checkpoint、resume、replay |
| Multi-Agent Runtime | 需要多角色并行或互审 | role、handoff、team topology |
| Managed Platform | 多团队复用和上线治理 | registry、policy、trace、eval、release gate |

Agent Platform 不是一开始就要建设的东西。但如果每个团队都在重复实现工具注册、状态恢复、审批、trace 和 eval，那么平台化就开始有价值。

---

### 21.1.5 图式编排 vs 自由 Agent Loop

自由 loop 灵活，图式编排可控。

建议：

- 低风险探索用自由 loop；
- 生产流程用显式 workflow；
- 高风险动作必须进入审批节点；
- 长任务必须有 checkpoint。


### 21.1.6 单 Agent vs 多 Agent

优先从单 Agent + Workflow 开始。只有当存在清晰角色边界、并行收益或互审需求时，再引入多 Agent。

多 Agent 的收益必须超过协调成本。


### 21.1.7 框架 vs 自研

选择框架还是自研，取决于已有系统约束。

适合框架：

- 需要 durable execution；
- 需要多 Agent；
- 需要快速验证；
- 团队愿意接受框架抽象。

适合自研：

- 内部权限、工具、审计约束很强；
- 只需要轻量 workflow；
- 不希望引入重依赖；
- 平台能力需要深度定制。


### 21.1.8 平台化 vs 单应用内嵌

单应用内嵌适合早期验证。平台化适合多团队复用。

平台化触发信号：

- 多个团队重复接模型；
- 多个团队重复封装工具；
- 权限和审计开始分散；
- 需要统一成本控制；
- 需要统一 eval 和 release gate；
- 长任务恢复成为共性问题。

## 21.2 计划、任务图与有界执行

### 21.2.1 Plan-and-Execute：计划与执行分离

Plan-and-Execute 把任务拆成两个阶段：

```text
Plan: 生成可执行步骤、约束、证据要求和风险边界
Execute: 按步骤执行、验证、修正和结束
```

适合场景：

- 任务目标明确；
- 步骤之间有依赖；
- 需要用户或系统审查计划；
- 执行过程需要恢复或追踪。

关键不是“先让模型写一个计划”，而是计划必须变成结构化任务对象。

```json
{
  "goal": "diagnose_order_latency",
  "steps": [
    {"id": "metrics", "action": "query_metrics", "risk": "read"},
    {"id": "logs", "action": "search_logs", "risk": "read"},
    {"id": "hypothesis", "action": "generate_hypothesis", "risk": "none"},
    {"id": "verify", "action": "verify_hypothesis", "risk": "read"}
  ],
  "stop_conditions": ["root_cause_found", "evidence_insufficient", "budget_exceeded"],
  "requires_approval": []
}
```

计划一旦结构化，Runtime 就能检查预算、权限、风险动作和完成条件。


### 21.2.2 ReAct Loop：观察、思考、行动、反馈

ReAct Loop 强调在推理和行动之间循环：

```text
Observe -> Think -> Act -> Observe -> ...
```

它适合工具结果不确定、需要边查边判断的任务。问题是，如果 ReAct 完全自由运行，就容易出现：

- 工具调用过多；
- 反复查询同一信息；
- 中途忘记目标；
- 停止条件不清楚；
- trace 难以归因。

生产系统中更常见的做法是把 ReAct 放进受控边界：

```text
最多调用 N 次工具
每次工具调用必须有 purpose
每轮必须更新 evidence state
达到 stop condition 后停止
高风险工具必须暂停审批
```

也就是说，ReAct 是 Agent 的认知循环，Workflow Runtime 是它的运行边界。


### 21.2.3 Workflow：固定流程中的 LLM 节点

有些业务流程本身比较稳定，只是其中某些节点需要 LLM 判断或生成。

例如客服工单处理：

```text
Classify Ticket
  -> Retrieve Policy
  -> Draft Reply
  -> Risk Check
  -> Human Review
  -> Send
```

这里 LLM 不是整个流程的主人，而是某些节点的执行器。Workflow 负责顺序、分支、权限、审批和失败恢复。

这种模式适合生产系统，因为它把不确定性限制在节点内部，而不是让整个流程都由模型自由决定。


### 21.2.4 State Machine：显式状态转移

当任务有明确生命周期时，应该用 state machine 表示。

```text
created
  -> planning
  -> running
  -> waiting_approval
  -> executing_action
  -> completed
  -> failed
```

状态机的价值是：

- 当前任务处于哪个阶段一目了然；
- 每个状态允许哪些动作可以被系统校验；
- 失败和恢复路径可以明确建模；
- 人工接管时能看到稳定状态；
- trace 和 eval 可以按状态归因。

不要把状态机藏在 prompt 里。状态应该由 Runtime 保存，模型只能基于状态提出建议或选择允许的动作。


### 21.2.5 计划节点必须有外部验收

ReAct交替行动和观察，为下一步补充环境证据。[[2]](#ref-2)本书TaskGraph节点另含node_id、depends_on、input/artifact_refs、owner、acceptance、risk、budget/deadline与failure_route。Todo是图的视图，不替代执行状态。反思只在新失败证据或进展停滞时触发，有限轮数；没有新证据不能反复自我赞同。

重规划固定plan_version，复用仍有效的产物和既有Operation，只重算受影响节点。目标或授权改变需要重新确认，旧版本结果不能提交到新计划。最大步数、调用数、并发、token、成本和墙钟deadline来自父预算，子任务不得各自重新拥有全额预算。完成需满足验收与安全约束，预算耗尽或unknown不是成功。

Checkpoint、取消、等待审批、未知副作用与恢复按[第11章](infra-11-runtime.md)执行，本章只组合控制流；Replay观察不重复执行副作用。

## 21.3 可组合工作流模式

### 21.3.1 Sequential：顺序执行

Sequential 是最基础的 workflow pattern。

```text
Input -> Step 1 -> Step 2 -> Step 3 -> Output
```

适合步骤有明确依赖的任务，例如：

- 先检索证据，再生成回答；
- 先解析需求，再生成代码；
- 先跑测试，再总结结果。

Sequential 可以由一个 Agent 执行，也可以由多个节点分别执行。重点是步骤关系，而不是 Agent 数量。

### 21.3.2 Parallel：并行执行

Parallel 用于独立子任务并行处理。

```text
          -> Branch A ->
Input -> -> Branch B -> Aggregator -> Output
          -> Branch C ->
```

适合场景：

- 多个数据源并行查询；
- 多个文件并行分析；
- 多个 reviewer 从不同角度审查；
- 多个候选方案并行生成。

Parallel 的难点在聚合。Aggregator 不能只是拼接结果，它需要处理冲突、去重、排序、引用和证据充分性。

### 21.3.3 Router：按意图分流

Router 根据任务类型选择路径。

```text
User Input
  -> Intent Router
     -> QA Flow
     -> Coding Flow
     -> Data Analysis Flow
     -> Human Escalation
```

Router 可以由规则、分类模型或 LLM 实现。但生产系统中应该输出结构化结果：

```json
{
  "intent": "incident_diagnosis",
  "confidence": 0.86,
  "route": "diagnosis_workflow",
  "reason": "user asks to inspect alert and logs"
}
```

低置信度或高风险任务应该进入澄清或人工路径。

### 21.3.4 Evaluator-Optimizer：评估与迭代优化

Evaluator-Optimizer 用于需要多轮改进的任务。

```text
Generator -> Evaluator -> Feedback -> Generator
```

适合：

- 代码生成和审查；
- 文档草稿和编辑；
- 查询计划优化；
- 多候选答案评估。

必须设置退出条件：

- 最大迭代次数；
- 明确验收标准；
- 质量不再提升时停止；
- 成本超过预算时停止；
- 低置信度时交给人工。

没有退出条件的 evaluator loop 很容易变成无限循环。

### 21.3.5 Orchestrator-Workers：编排者与工作者

Orchestrator-Workers 把任务拆给多个 worker，再汇总结果。

```text
Orchestrator
  -> Worker A
  -> Worker B
  -> Worker C
  -> Merge
```

这个模式可以是单 Agent 内部的 task decomposition，也可以是真正的多 Agent 协作。

关键是 worker 的输入必须清晰：

- 目标；
- 边界；
- 可写范围；
- 预期输出格式；
- 验收标准；
- 禁止事项。

如果 worker 接到的是模糊自然语言，合并成本会急剧上升。

### 21.3.6 Fan-out：批量任务与并行处理

Fan-out 是 Orchestrator-Workers 的批量化形态。

```text
Files[1..N] -> N workers -> Results -> Review / Merge
```

适合：

- 批量迁移；
- 批量修复；
- 批量代码审查；
- 批量文档改写；
- 批量数据抽取。

Fan-out 的核心风险是冲突和质量不一致。应该尽量保证每个 worker 的写入范围不重叠，并在最后设置合并审查。

### 21.3.7 Workflow 的验收标准、超时与预算

Workflow 不能只定义步骤，还要定义运行边界。

```yaml
workflow_policy:
  max_steps: 12
  max_tool_calls: 20
  timeout_seconds: 900
  max_cost_usd: 3.0
  approval_required_for:
    - write_database
    - deploy
    - send_external_message
  stop_conditions:
    - completed
    - evidence_insufficient
    - user_cancelled
    - budget_exceeded
```

这些边界最好由 Runtime 执行，而不是靠 prompt 约束。

---


### 21.3.8 十万词分类：用批处理代替十万次自主规划

教学输入为十万余词及固定分类体系版本。先定义单/多标签、层级、歧义和拒判，人工标注分层基线；规范化保留original_word与稳定item_id，去重项保留原位置映射。规则/词典处理确定项，轻量分类或embedding给候选，LLM只处理难例；不按每个词创建自主Agent。

代码按token预算与输出长度确定batch，教学候选每批128项仅作起点，不能当通用最优值。每项输出item_id、taxonomy_version、labels、evidence/reason及abstain；检查输入输出一一对应、无缺/重项、枚举/层级/多标签约束。按batch_id/input_hash/model/prompt/taxonomy版本保存结果，失败只重跑未完成或失效项，重启按账本恢复。

标签更新保留旧版与迁移映射，低置信阈值在独立标注集校准，LLM自报分数不是概率。评价macro/micro、长尾混淆、层级一致、拒判、尾延迟和单位成功成本。仅独立领域判断确有收益时用专家Agent；机械切批、限速、join和结果去重由确定性调度器完成。此案例未运行。

### 漫剧创作工作流：从需求到可审查产物

下面是将本书编排方法迁移到创作场景的教学设计，不是已经实现的生成平台。先确认受众、时长、风格、角色与素材使用范围，再建立产物依赖：需求版本 → 剧本 → 角色资产与分镜 → 图像/配音 → 镜头片段 → 剪辑 → 成片。图像与配音可在依赖满足后并行，剪辑必须引用已经验收的片段版本。模型调用负责产生候选，推进节点由外部验收决定。

| 节点组 | 输入与产物契约 | 反馈后的处理 |
| --- | --- | --- |
| 剧本、角色与分镜 | 约束版本、角色 ID、镜头 ID、台词与时序；交付可预览的结构化草稿 | 剧情或服装改变，沿资产依赖标记受影响节点失效 |
| 图像、配音与片段 | 固定角色引用、模型配置和输入产物 hash；持久保存候选本体与失败原因 | 限额重试失败镜头，独立检查身份、音画及连续性；不能只改 Prompt 宣告修复 |
| 剪辑与成片 | 已验收镜头清单、顺序、字幕和音轨版本；交付预览与检查记录 | 用户确认新版本后再替换，旧版本可回滚；实际发布须单独授权 |

每次调用保存 `operation_id`、依赖版本、状态与费用；超时先查结果，不能把重复生成当免费重试。用户改服装后，旧版本的迟到成功仍进入产物账本，但不覆盖当前分镜。对象存储保存媒体，任务状态保存引用；seed 与参数帮助追溯，不能代替已验收媒体。按任务预算限制候选数量，失败时交付缺口与可继续的检查点。持久执行、补偿和人工接管复用[第11章 Agent Runtime](infra-11-runtime.md)，任务图与节点验收沿用本章。生成后端训练、专用身份控制算法与实际成片质量仍需独立项目验证。

## 21.4 多 Agent 的职责与交接协议

### 21.4.1 多 Agent 什么时候才必要

多 Agent 的价值不是“看起来更智能”，而是解决单 Agent 难以同时满足的工程诉求：

- 职责分离；
- 并行执行；
- 独立上下文；
- 互相审查；
- 专家能力隔离；
- 大任务分块处理。

不适合多 Agent 的场景：

- 简单问答；
- 单次工具查询；
- 没有明确角色边界的小任务；
- 成本比收益更高的低风险任务。

判断标准不是任务复杂度本身，而是是否存在清晰的角色边界和可合并的输出。

### 21.4.2 Planner / Executor

Planner / Executor 把计划和执行分离。

```text
Planner: 生成计划、约束、风险和验收标准
Executor: 按计划执行工具、修改文件或生成结果
```

适合长任务和高风险任务。Planner 不直接执行副作用动作，Executor 也不能随意改变目标。计划变更需要回到 Planner 或用户确认。

### 21.4.3 Writer / Reviewer

Writer / Reviewer 用于质量控制。

```text
Writer -> Draft
Reviewer -> Findings
Writer -> Revision
```

这个模式在代码、文档、方案设计中都常见。Reviewer 必须有明确标准，否则会变成泛泛评价。

好的 Reviewer 输出应该包含：

- 具体问题；
- 影响；
- 证据位置；
- 建议修复；
- 是否阻塞发布。

### 21.4.4 Researcher / Synthesizer

Researcher / Synthesizer 用于复杂研究任务。

```text
Researcher: 搜索、阅读、抽取证据
Synthesizer: 组织结论、处理冲突、生成答案
```

Researcher 不应该直接生成最终结论，Synthesizer 也不应该编造证据。两者之间应该传递 Evidence Packet，而不是自由文本摘要。

### 21.4.5 Coordinator / Specialist

Coordinator / Specialist 适合多领域任务。

```text
Coordinator
  -> Security Specialist
  -> Performance Specialist
  -> API Specialist
  -> Docs Specialist
```

Coordinator 负责拆分、分配、合并和停止条件。Specialist 负责明确领域内的判断。

### 21.4.6 多 Agent 的风险：成本、上下文丢失、无限循环、权限绕过

多 Agent 常见失败模式：

| 风险 | 表现 | 修复 |
|:---|:---|:---|
| 成本失控 | 多个 Agent 重复探索 | 设置预算、去重、共享证据 |
| 上下文丢失 | 交接后忘记约束 | 使用结构化 handoff |
| 无限循环 | reviewer 和 writer 反复争论 | 最大迭代次数和验收标准 |
| 权限绕过 | 子 Agent 调用不该用的工具 | 每个 Agent 独立权限校验 |
| 合并失败 | 输出格式不一致 | 明确输出 schema |

多 Agent 不是替代 workflow。多 Agent 更需要 workflow。

---


### 21.4.7 Modular Agent与交接契约

Modular Agent在本书指职责和接口可替换的设计方式，不是一个统一标准算法，也不要求每个模块都成为Agent。Multi-Agent则有多个各自上下文和决策循环的角色；模块化本身不会自动产生规划能力。

Handoff至少包含任务/父子身份、输入版本和证据、验收/返回格式、工具/数据授权范围、预算/deadline、owner与拒收/升级路径。接收方先验权限/版本/能力/预算，再原子认领目标Step并回accept或带原因reject；发送消息不等责任已交接。拒收仍由原owner负责重新分流，限定次数后升级人工，避免互相推诿。

## 21.5 mailbox、共享状态与事件总线

### 21.5.1 三种组件各司其职

共享状态持久保存TaskGraph、Step/owner/version、审批和不可变Artifact引用；mailbox承载定向命令及接收/处理状态；事件总线传播“状态已变”的通知，不能取代权威状态。结构为“编排器+状态→定向信箱→Worker→产物/状态→Outbox事件→编排器核验→后继节点”。

Correlation Identifier模式区分请求和响应的关联。[[3]](#ref-3)RabbitMQ将发布者确认与消费者确认分开；收到broker确认并不证明下游业务动作完成。[[4]](#ref-4)本书envelope是教学契约：

```json
{"schema_version":"1","message_id":"m17","kind":"review.request","task_id":"t1","run_id":"r1","step_id":"review","parent_task_id":"parent1","correlation_id":"h1","causation_id":"m16","sender":"A","recipient":"B","plan_version":3,"state_version":7,"owner_epoch":4,"deadline":"2026-10-18T10:10:00Z","budget_ref":"budget-r1","input_ref":"patch-v2","input_hash":"sha256-demo","evidence_refs":["test-v2"],"authorization_ref":"scope-review-v1","reply_to":"mailbox-A"}
```

身份与授权由可信通道和服务端验证，不能信任载荷自称sender或authorization。大内容传不可变引用，接收时重验读取权限；敏感原文不广播给所有订阅者。message_id用于传输去重，correlation_id贯穿交接，causation_id表达因果；业务Operation稳定键属于第11章，不能用message_id替代。

### 21.5.2 重复、乱序、取消和失联

消费者将inbox去重记录与创建/推进Step放同一受控事务，成功接收后才确认。投递重复查询既有处理结果，同message_id异payload拒绝；同业务动作跨消息仍按Operation去重。状态更新条件检查plan_version、state_version与epoch，早到依赖结果等待或对账，旧终态不得覆盖新终态。

取消持久写入权威状态，并以控制消息传播到子任务；停止新动作，在安全点确认。先到取消后到成功时保留外部事实，不能为保持“已取消”展示而抹掉副作用。owner失联按第11章租约/执行代次处理，租约过期不证明旧进程已停止；网关/资源支持时执行fencing，不支持则限制副作用并查询对账。

mailbox限制队列长度、并发、消息大小、委派深度与重投次数，超过总deadline停止新工作；永久失败进入带owner/修复入口的dead letter，通知不能代替认领。中心协调易于全局预算和审计但可能瓶颈；分布式执行提升隔离与吞吐，必须付出一致性、去重和跨分区恢复成本。逻辑中心协调可与分布式Worker组合，并非二选一。

## 21.6 共享资源与执行隔离

### 21.6.1 Git Worktrees：隔离工作区

Coding Agent 的并行执行通常需要隔离文件系统状态。Git worktree 是一个实用基础设施。

```bash
git worktree add ../feature-auth feature/auth
git worktree add ../feature-payment feature/payment
```

每个 Agent 在独立 worktree 中工作，可以减少文件冲突。适合：

- 大规模重构；
- 多模块并行开发；
- 多方案实验；
- 批量迁移。

关键要求是写入范围要提前划分，避免多个 Agent 修改同一文件。

### 21.6.2 Subagents：专家上下文

Subagent 的核心价值是独立上下文。主 Agent 可以把一个边界清晰的任务交给专家 Agent，让它在自己的上下文中完成。

适合：

- 安全审查；
- 性能分析；
- 测试补齐；
- 文档改写；
- 代码库局部探索。

不适合把关键路径上的阻塞任务随意交给 subagent。下一步必须依赖的结果，主流程通常应该自己处理，或明确等待。

### 21.6.3 Agent Teams：自动协调

Agent Team 把多个角色组织成固定拓扑。例如：

```text
Writer -> Reviewer -> Writer
Planner -> Executor -> Verifier
Coordinator -> Specialists -> Coordinator
```

Team 的重点不是“让多个 Agent 聊天”，而是：

- 每个角色有明确职责；
- 消息传递有结构；
- 有终止条件；
- 有合并规则；
- 有审计和成本记录。

### 21.6.4 Batch / Fan-out：批量迁移、批量修复与批量评审

批量任务适合 fan-out。

```text
Find targets
  -> shard targets
  -> dispatch workers
  -> collect results
  -> run verification
  -> merge
```

批量迁移尤其要注意：

- 每个 shard 的写入范围；
- 失败 shard 的重试策略；
- 全局一致性检查；
- 最后统一格式化和测试；
- 合并后的人工 review。

### 21.6.5 并行任务的合并、冲突和审查边界

并行执行的难点不在启动 worker，而在合并。

合并前要检查：

- 是否有文件冲突；
- 是否有接口不一致；
- 是否有重复实现；
- 是否破坏共享测试；
- 是否有未声明副作用。

并行执行的原则是：并行可以提升吞吐，但不能降低审查标准。

---


### 21.6.6 隔离的对象不止文件

Git worktree隔离checkout，不隔离同一数据库、远程服务、端口、凭据或共享缓存。每个角色使用独立工作目录/进程/必要sandbox，固定输入版本与hash；共享文件通过分工或受控合并更新，不让多个Agent同时覆写。

输入按“原始任务→权威状态→每角色Context Package视图”装配。公共目标/硬约束一致，专用证据按角色权限加载；各自消息列表、证据引用和摘要不可变或copy-on-write，不共享一个可任意改写的全量上下文。共享结果需明确来源/版本/可信度，敏感材料在读取和输出时再检权限。全局资源/预算/取消由唯一责任层控制。

## 21.7 框架与 SDK 的分层选型

### 21.7.1 先分模型SDK、Agent SDK、编排Runtime与产品

| 层 | 提供什么 | 例子与应用仍需承担的责任 |
| --- | --- | --- |
| 模型SDK | 请求/响应/流式事件客户端 | 原生OpenAI/Anthropic客户端；应用自建循环、工具执行和状态 |
| Agent SDK/Harness | 循环、工具、上下文与角色能力 | OpenAI Agents SDK、Claude Agent SDK；核对现成工具/交接和进程/工作区/权限边界 |
| 编排Runtime/框架 | 显式状态/节点/边/消息和恢复组合 | LangGraph、AutoGen Core、Microsoft Agent Framework；应用仍验证存储、业务幂等和总预算 |
| 面向用户产品 | IDE/CLI/工作台的完整使用体验 | Coding产品、Hermes等案例按第23/27章版本核验；产品名不证明底层标准或能力 |

LangChain提供高层agent/harness，官方说明其Agent建立于LangGraph之上；直接使用LangGraph可细化状态、节点、条件边和并发合并规则。[[5]](#ref-5)持久性还取决于checkpointer选择，内存Saver不能用于证明重启恢复。[[6]](#ref-6)

OpenAI Agents SDK提供代码式循环/工具/交接等入口，应用拥有工具实现、部署、状态与审批决策。[[7]](#ref-7)Claude Agent SDK面向把Claude Code循环/工具/上下文能力嵌入应用，运行者需管理进程、工作区和凭据；两者场景有重叠，不是业务/编码的排他分工。[[8]](#ref-8)

AutoGen Core强调Actor式异步消息与事件驱动。[[9]](#ref-9)Microsoft Agent Framework提供Agent与Workflow开发抽象。[[10]](#ref-10)这些是执行模型实例，不能仅按“强/中/弱”声称恢复、权限和治理完整。

### 21.7.2 组合选型要消除重复责任

固定路径分类先用普通代码+模型SDK；常见工具循环可评估高层Agent SDK；复杂分支/审批/暂停用显式图；跨角色异步协作先验证消息和owner。可以在图节点中运行Agent SDK，但必须明确内外循环、重试、状态、预算、取消与完成验收各由谁负责，防止双重重试和重复副作用。用最小真实任务和故障验证质量、延迟、成功成本与维护代价，再决定是否自研。

### 21.7.3 A2A仅作版本化协议选读

A2Av0.3.0定义Agent Card、Task/Message/Artifact以及认证授权等接口。[[11]](#ref-11)MCP连接能力服务，A2A面向具有独立执行过程的Agent协作；两者可以组合。本文自建envelope、子Agent调用或事件总线都不自动符合A2A。声明支持需要检查协议端点、版本/能力、消息/任务模型和兼容测试；本轮未实现或测试A2A。

## 21.8 平台组合、故障与选型

### 21.8.1 最小生产平台清单

一个最小生产 Agent 平台不应该先追求可视化拖拽，而应该优先建设运行时底座。

| 模块 | 最小能力 | 不做会怎样 |
|:---|:---|:---|
| Model Gateway | 统一模型调用、超时、重试、成本记录 | 各团队重复封装，成本不可控 |
| Tool Registry | 工具 schema、owner、风险等级 | 工具滥用，没人知道谁能做什么 |
| Workflow Runtime | 状态、checkpoint、interrupt、resume | 长任务失败无法恢复 |
| Policy Engine | read / write / execute 权限 | 只能靠 prompt 控制风险 |
| Trace Store | run、node、tool、token、latency、error | 调试和复盘困难 |
| Eval Harness | 回归集、失败样本、版本对比 | 模型或 prompt 一改就退化 |
| Artifact Store | 报告、diff、图表、日志片段 | 输出散落在聊天里，无法审查 |

### 21.8.2 控制面、执行面、集成面与治理面

可以把平台拆成四个面：

```text
Control Plane: registry / policy / config / release
Execution Plane: workflow runtime / task runner / sandbox
Integration Plane: tools / resources / connectors / MCP
Governance Plane: trace / eval / audit / cost / monitoring
```

这四个面不一定要拆成四个服务，但职责要清楚。

### 21.8.3 生产级运行时事件模型

Agent Runtime 应该把关键事件记录下来。

```json
{
  "run_id": "run_123",
  "event_type": "tool_call_completed",
  "node": "query_logs",
  "tool": "log_search",
  "input_hash": "sha256:...",
  "status": "success",
  "latency_ms": 842,
  "cost": 0.02,
  "timestamp": "2026-05-21T10:00:00Z"
}
```

事件模型是 trace、eval、debug、audit 和 billing 的共同基础。

### 21.8.4 多租户与权限边界

平台化后必须处理多租户：

- 用户身份；
- 项目边界；
- 工具权限；
- 数据权限；
- Memory scope；
- 知识库 ACL；
- 审计归属。

多 Agent 和 subagent 不能继承无限权限。每次 handoff 都应该重新计算权限和上下文边界。

### 21.8.5 从单应用 Agent 到共享平台的演进路线

推荐演进路线：

```text
single app agent
  -> shared tool runtime
  -> shared workflow runtime
  -> shared trace / eval
  -> shared policy / approval
  -> managed agent platform
```

不要一开始就做大而全平台。先从重复痛点最多的能力开始抽取。

---


### 21.8.6 工作流卡死

表现：

- 等待一个永远不会发生的事件；
- 状态无法转移；
- Agent 反复尝试同一步。

修复：

- 每个等待状态设置 timeout；
- 每个状态定义允许动作；
- 增加 fallback 和 human escalation；
- trace 中记录卡住原因。

### 21.8.7 状态丢失或重复执行

表现：

- 进程重启后任务从头开始；
- 工具副作用重复发生；
- 审批后找不到上下文。

修复：

- checkpoint 关键状态；
- 外部动作使用幂等键；
- action log 记录请求和响应；
- resume 时从状态恢复，不从聊天历史猜。

### 21.8.8 多 Agent 循环争论

表现：

- reviewer 不断要求修改；
- writer 不断生成新版本；
- coordinator 无法决策。

修复：

- 设置最大轮数；
- 定义验收标准；
- 引入 final decision owner；
- 低置信度转人工。

### 21.8.9 工具副作用失控

表现：

- Agent 执行了未授权写操作；
- 批量任务影响范围过大；
- 工具调用无法追责。

修复：

- 工具分级；
- 写操作审批；
- sandbox；
- dry-run；
- audit log；
- 最小权限。

### 21.8.10 平台抽象过重

表现：

- 简单任务也要配置复杂 graph；
- 业务团队接入成本高；
- 框架概念多于业务价值。

修复：

- 从 MVP runtime 开始；
- 常见模式模板化；
- 保留轻量 escape hatch；
- 平台能力按复用痛点演进。

### 21.8.11 Trace 不足导致无法复盘

表现：

- 不知道模型看到了什么；
- 不知道为什么调用工具；
- 不知道证据来自哪里；
- 不知道失败发生在哪一步。

修复：

- 记录 run / node / tool / model span；
- 保存 evidence id；
- 记录 policy decision；
- 记录 cost、latency、error；
- 将失败 trace 转成 eval case。

---


### 21.8.12 一次双Agent交接与故障复核

教学Coding任务由A负责父任务，B负责补丁审查子Step；B不自动获得发布权限。

| 顺序 | 事实/消息 | 权威变化与验证 |
| --- | --- | --- |
| 1 | A保存patch-v2/test-v2，提交review.request m17与Outbox | review Step版本7，目标B、预算与输入hash固定 |
| 2 | mailbox投递，B校验输入/权限并条件认领 | review版本8/owner=B/epoch=4，回accept；A仍拥有父任务 |
| 3 | m17重复投递 | inbox返回已有认领，不创建第二个review或新业务动作 |
| 4 | B输出不可变review-v1并提交result m18 | 带correlation=h1、causation=m17、plan=3与产物hash |
| 5 | A收到后查询review Step并独立验收 | 旧plan结果拒收，验收通过才join后继；拒收回有界修订 |
| 6 | B中途失联或父任务取消 | 由状态记录驱动owner恢复/取消传播，迟到产物重验，不复活终态 |

生产故障断言至少覆盖：提交后宕机、投递/回执重复、乱序、过期owner写入、同ID异参、拒收循环、权限撤销、取消后晚到成功、队列积压及缺失产物。每项复用[第11章](infra-11-runtime.md)原语和[第12章](infra-12-reliability-governance.md)修复owner/deadline；Trace和事件总线只是观测，不是可靠事实的替代。

## 本章小结

编排把计划、步骤、权限和验收连成有界过程。mailbox传定向工作，共享状态保存责任与版本，总线通知变化；消息关联、传输去重与业务幂等分别设计。角色越多越需要独立上下文、资源隔离和总预算。框架按层选型，自建通信不自动成为A2A，教学轨迹不是实现与实测。


## 参考资料

1. <a id="ref-1"></a> LangChain，[Workflows and agents](https://docs.langchain.com/oss/python/langgraph/workflows-agents)，rolling; 2026-10-18。
2. <a id="ref-2"></a> Yao et al.，[ReAct: Synergizing Reasoning and Acting in Language Models](https://arxiv.org/abs/2210.03629)，2022。
3. <a id="ref-3"></a> Hohpe and Woolf，[Correlation Identifier](https://www.enterpriseintegrationpatterns.com/patterns/messaging/CorrelationIdentifier.html)，Enterprise Integration Patterns。
4. <a id="ref-4"></a> RabbitMQ，[Consumer Acknowledgements and Publisher Confirms](https://www.rabbitmq.com/docs/confirms)，rolling; 2026-10-18。
5. <a id="ref-5"></a> LangChain，[LangChain overview](https://docs.langchain.com/oss/python/langchain/overview)，rolling; 2026-10-18。
6. <a id="ref-6"></a> LangChain，[Persistence](https://docs.langchain.com/oss/python/langgraph/persistence)，rolling; 2026-10-17。
7. <a id="ref-7"></a> OpenAI，[Agents SDK](https://developers.openai.com/api/docs/guides/agents/sdk)，rolling; 2026-10-17。
8. <a id="ref-8"></a> Anthropic，[Agent SDK overview](https://code.claude.com/docs/en/agent-sdk/overview)，rolling; 2026-10-18。
9. <a id="ref-9"></a> Microsoft，[AutoGen Core](https://microsoft.github.io/autogen/stable/user-guide/core-user-guide/index.html)，stable docs; 2026-10-18。
10. <a id="ref-10"></a> Microsoft，[Microsoft Agent Framework Overview](https://learn.microsoft.com/en-us/agent-framework/overview/)，rolling; 2026-10-18。
11. <a id="ref-11"></a> A2A project，[A2A Protocol Specification](https://a2a-protocol.org/v0.3.0/specification/)，v0.3.0。
