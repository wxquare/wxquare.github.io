# 第9章 Agent 执行编排与平台架构：工作流、状态机、多 Agent 与框架生态

> Agent 执行编排的目标，不是让模型“多想几步”，而是把不确定的模型行为放进可恢复、可审计、可治理的运行时结构里。

## 引言：从可靠执行到平台化运行时

前面几章分别讨论了 Agent 的架构边界、工具系统、知识系统和记忆系统。到这里，一个 Agent 已经具备了“看上下文、查知识、调用工具、保留连续性”的能力。但这些能力本身还不能保证任务可靠完成。

生产环境中的 Agent 往往要处理更长的任务：

- 一次任务包含多个步骤；
- 某些步骤依赖工具结果；
- 某些动作有副作用，需要审批；
- 中途可能失败、超时或被用户打断；
- 任务可能隔天继续；
- 多个 Agent 或多个工作区可能并行执行；
- 结果必须能被追踪、回放和评估。

这就是执行编排的位置。

执行编排不是多 Agent 才需要。单 Agent 只要任务变长、动作变多、风险变高，也需要 workflow、state machine、checkpoint、human-in-the-loop 和 trace。多 Agent 只是执行编排的一种复杂形态。

本章把两类内容合并讨论：

```text
执行编排：如何把任务组织成可靠过程
平台架构：如何把编排、状态、工具、审批、恢复和观测沉淀为可复用运行时
```

这也是为什么 LangGraph、AutoGen、Microsoft Agent Framework 这类框架不能只当作“Agent 框架”看。它们背后真正解决的是：如何让 Agent 从 demo 变成可运行、可恢复、可审计、可扩展的系统。

---

## 9.1 从 Agent Loop 到 Workflow Runtime

### 9.1.1 最小 Agent Loop 的边界

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

### 9.1.2 为什么单 Agent 也需要显式编排

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

### 9.1.3 控制流、状态、副作用、人工介入与可观测性

生产级 Agent Runtime 至少要显式处理五类问题。

| 问题 | Demo 做法 | 生产做法 |
|:---|:---|:---|
| 控制流 | 让模型自由决定下一步 | workflow、graph、router、state machine |
| 状态 | 拼进 prompt 或内存变量 | task state、checkpoint、session store |
| 副作用 | 直接调用工具 | tool runtime、幂等键、审批、补偿 |
| 人工介入 | 聊天里问“可以吗” | approval node、interrupt、resume token |
| 可观测性 | 打印日志 | trace、span、evidence、cost、eval replay |

这五类问题决定了 Agent 能不能上线。模型可以生成计划，但系统必须管理执行计划的生命周期。

### 9.1.4 Agent Demo 到 Agent Platform 的演进路径

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

## 9.2 单 Agent 执行编排：把任务组织成可靠过程

### 9.2.1 Plan-and-Execute：计划与执行分离

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

### 9.2.2 ReAct Loop：观察、思考、行动、反馈

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

### 9.2.3 Workflow：固定流程中的 LLM 节点

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

### 9.2.4 State Machine：显式状态转移

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

### 9.2.5 Checkpoint、Resume、Replay 与幂等

长任务必须考虑失败。模型调用可能超时，工具可能失败，进程可能重启，人类审批可能隔天才发生。

Checkpoint 的基本思想是：

```text
Step completed -> save state
External action requested -> save intent and idempotency key
Human approval pending -> save approval state
Resume -> continue from last safe checkpoint
```

关键点有三个：

- checkpoint 保存的是结构化状态，不只是聊天历史；
- 外部副作用必须有幂等键；
- replay 时不能重复执行已经成功的副作用。

幂等设计尤其重要。一个“创建退款单”的工具如果在恢复时重复执行，就会造成真实业务事故。Runtime 必须记录 action id、request payload、response、状态和重试策略。

### 9.2.6 Human-in-the-loop：审批、暂停与恢复

Human-in-the-loop 不是在聊天里问一句“要继续吗”，而是 workflow state 的一部分。

```text
Propose Action
  -> Risk Classifier
  -> Approval Node
  -> Execute or Reject
```

审批节点至少要记录：

- 谁发起；
- 审批什么动作；
- 证据是什么；
- 风险等级是什么；
- 谁批准或拒绝；
- 什么时候处理；
- 恢复后从哪里继续。

这样人工介入才具备可审计性和可恢复性。

---

## 9.3 Workflow Pattern：单 Agent 和多 Agent 都适用

### 9.3.1 Sequential：顺序执行

Sequential 是最基础的 workflow pattern。

```text
Input -> Step 1 -> Step 2 -> Step 3 -> Output
```

适合步骤有明确依赖的任务，例如：

- 先检索证据，再生成回答；
- 先解析需求，再生成代码；
- 先跑测试，再总结结果。

Sequential 可以由一个 Agent 执行，也可以由多个节点分别执行。重点是步骤关系，而不是 Agent 数量。

### 9.3.2 Parallel：并行执行

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

### 9.3.3 Router：按意图分流

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

### 9.3.4 Evaluator-Optimizer：评估与迭代优化

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

### 9.3.5 Orchestrator-Workers：编排者与工作者

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

### 9.3.6 Fan-out：批量任务与并行处理

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

### 9.3.7 Workflow 的验收标准、超时与预算

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

## 9.4 Multi-Agent 协作：执行编排的复杂形态

### 9.4.1 多 Agent 什么时候才必要

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

### 9.4.2 Planner / Executor

Planner / Executor 把计划和执行分离。

```text
Planner: 生成计划、约束、风险和验收标准
Executor: 按计划执行工具、修改文件或生成结果
```

适合长任务和高风险任务。Planner 不直接执行副作用动作，Executor 也不能随意改变目标。计划变更需要回到 Planner 或用户确认。

### 9.4.3 Writer / Reviewer

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

### 9.4.4 Researcher / Synthesizer

Researcher / Synthesizer 用于复杂研究任务。

```text
Researcher: 搜索、阅读、抽取证据
Synthesizer: 组织结论、处理冲突、生成答案
```

Researcher 不应该直接生成最终结论，Synthesizer 也不应该编造证据。两者之间应该传递 Evidence Packet，而不是自由文本摘要。

### 9.4.5 Coordinator / Specialist

Coordinator / Specialist 适合多领域任务。

```text
Coordinator
  -> Security Specialist
  -> Performance Specialist
  -> API Specialist
  -> Docs Specialist
```

Coordinator 负责拆分、分配、合并和停止条件。Specialist 负责明确领域内的判断。

### 9.4.6 多 Agent 的风险：成本、上下文丢失、无限循环、权限绕过

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

## 9.5 并行执行基础设施

### 9.5.1 Git Worktrees：隔离工作区

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

### 9.5.2 Subagents：专家上下文

Subagent 的核心价值是独立上下文。主 Agent 可以把一个边界清晰的任务交给专家 Agent，让它在自己的上下文中完成。

适合：

- 安全审查；
- 性能分析；
- 测试补齐；
- 文档改写；
- 代码库局部探索。

不适合把关键路径上的阻塞任务随意交给 subagent。下一步必须依赖的结果，主流程通常应该自己处理，或明确等待。

### 9.5.3 Agent Teams：自动协调

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

### 9.5.4 Batch / Fan-out：批量迁移、批量修复与批量评审

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

### 9.5.5 并行任务的合并、冲突和审查边界

并行执行的难点不在启动 worker，而在合并。

合并前要检查：

- 是否有文件冲突；
- 是否有接口不一致；
- 是否有重复实现；
- 是否破坏共享测试；
- 是否有未声明副作用。

并行执行的原则是：并行可以提升吞吐，但不能降低审查标准。

---

## 9.6 LangGraph：把 Agent 表示成有状态图

### 9.6.1 State、Node、Edge 与 Graph

LangGraph 的核心抽象是有状态图。

| 抽象 | 含义 | 工程价值 |
|:---|:---|:---|
| State | 节点之间共享的结构化状态 | 避免所有信息都塞进 prompt |
| Node | 一步计算，可以是 LLM、工具、函数、人工节点 | 让步骤可测试、可替换 |
| Edge | 状态转移，可以是固定边或条件边 | 把控制流显式化 |
| Graph | 节点和边组成的运行时结构 | 支持可视化、恢复和审计 |

这种设计适合把 Agent loop 拆成可观察的步骤。

### 9.6.2 Durable Execution

Durable Execution 让长任务可以在失败后恢复。关键是每一步状态都能持久化。

```text
Node A completed -> checkpoint
Node B waiting approval -> checkpoint
Resume after approval -> Node C
```

这比保存聊天历史更可靠，因为 Runtime 知道当前状态、已完成动作和下一步允许动作。

### 9.6.3 Human-in-the-loop

LangGraph 这类图式编排天然适合插入人工节点。

```text
Diagnose -> Propose Action -> Human Approval -> Execute
```

审批不是模型输出的一句话，而是图中的 interrupt / approval state。这样才能恢复和审计。

### 9.6.4 Persistence 与 Replay

Persistence 保存状态，Replay 用于调试和评估。

Replay 时要区分：

- 可以重放的纯计算节点；
- 不能重复执行的副作用节点；
- 需要 mock 或固定响应的外部工具；
- 需要人工重新确认的审批节点。

这也是为什么工具副作用要有幂等键和 action log。

### 9.6.5 适用场景与局限

适合 LangGraph 思路的场景：

- 长任务；
- 多步骤流程；
- 需要恢复和人审；
- 需要可视化状态；
- 需要严格 trace。

局限是设计成本更高。对于一次性问答或低风险探索，完整图式编排可能过重。

---

## 9.7 AutoGen：从多 Agent 对话到协作拓扑

### 9.7.1 AgentChat、Core 与 Extensions

AutoGen 的价值在于多 Agent 编程模型。它把不同角色的 Agent、消息流、工具执行和团队模式组织起来。

从工程角度看，它解决的是：

- 多 Agent 如何通信；
- 谁先说，谁后说；
- 什么时候停止；
- 工具由谁执行；
- 多角色结果如何合并。

### 9.7.2 Team Pattern：多 Agent 不是群聊

多 Agent 系统不应该是开放群聊，而应该是协作拓扑。

```text
RoundRobin Team
Selector Team
Swarm
Planner / Executor Team
Writer / Reviewer Team
```

不同拓扑对应不同控制策略。生产系统需要明确谁有最终决策权。

### 9.7.3 工具执行和代码执行边界

多 Agent 中工具权限更容易出问题。不能因为某个 Agent 是“子角色”，就绕过工具权限。

每个 Agent 都应该有：

- 可用工具列表；
- 风险等级；
- sandbox；
- 审批规则；
- trace identity。

### 9.7.4 多 Agent 系统的终止条件

多 Agent 最容易无限循环。终止条件必须外置。

```yaml
team_policy:
  max_rounds: 6
  stop_when:
    - reviewer_approved
    - coordinator_decided
    - budget_exceeded
    - human_required
```

终止条件不应该完全交给参与对话的 Agent 自己判断。

### 9.7.5 适用场景与局限

适合 AutoGen 思路的场景：

- 多角色协作；
- 互审；
- 研究和综合；
- 复杂任务分解；
- 需要模拟团队工作方式。

局限是成本和不确定性更高。角色越多，协调协议越重要。

---

## 9.8 Microsoft Agent Framework：企业化 Agent Runtime

### 9.8.1 Agents vs Workflows

企业级 Agent 系统通常同时需要 Agent 和 Workflow。

```text
Agent: 处理开放任务和动态判断
Workflow: 管理确定流程、状态、审批和恢复
```

两者不是替代关系。生产系统常常是 workflow 调用 Agent，Agent 在节点内完成推理或工具选择。

### 9.8.2 企业编排模式

企业编排更关注：

- 长任务；
- 组织权限；
- 审批流程；
- 多系统连接；
- 审计留痕；
- 版本治理；
- 部署和监控。

这些能力往往比“模型能不能回答”更决定系统能否上线。

### 9.8.3 Middleware、Telemetry 与企业接入

企业平台需要在运行时统一处理横切能力：

- authentication；
- authorization；
- policy；
- logging；
- tracing；
- cost accounting；
- rate limiting；
- data boundary。

Middleware 和 telemetry 的价值是让这些能力不散落在业务代码里。

### 9.8.4 审批、恢复、部署与治理

企业场景中，Agent Runtime 要能回答：

- 哪个版本执行了这个任务；
- 哪些工具被调用；
- 谁批准了高风险动作；
- 失败后是否能恢复；
- 新版本是否通过 eval；
- 线上质量是否下降。

这些能力和第 10 章生产治理直接衔接。

### 9.8.5 适用场景与局限

适合企业级 Agent Framework 的场景：

- 多团队共用 Agent 能力；
- 需要统一权限和审计；
- 需要工作流审批；
- 需要部署、监控和治理；
- Agent 是长期运行服务，而不是一次性脚本。

局限是平台成本较高。单一场景验证阶段不应过早平台化。

---

## 9.9 平台能力矩阵

### 9.9.1 编排能力

编排能力包括：

- graph / workflow；
- router；
- conditional edge；
- loop；
- interrupt；
- human approval；
- retry；
- compensation。

编排层决定任务怎么走。

### 9.9.2 状态与持久化能力

状态能力包括：

- task state；
- session state；
- checkpoint；
- thread；
- replay；
- state migration；
- action log。

状态层决定任务能不能恢复。

### 9.9.3 多 Agent 能力

多 Agent 能力包括：

- role definition；
- team topology；
- handoff；
- shared evidence；
- independent context；
- termination policy；
- merge protocol。

多 Agent 层决定多个角色能不能协作而不是互相干扰。

### 9.9.4 工具和资源接入能力

工具和资源接入可以通过函数调用、内部 API Gateway、Connector、MCP Server 等方式实现。

MCP 的价值在于统一暴露 Tool、Resource 和 Prompt，但它不是编排框架。它不负责 workflow、checkpoint、human approval、multi-agent routing 或 release gate。

因此平台设计时应把 MCP 放在 Integration Layer，而不是 Orchestration Layer。

```text
Orchestration Layer: workflow / state / approval / routing
Integration Layer: tools / resources / connectors / MCP servers
Governance Layer: policy / trace / eval / release gate
```

### 9.9.5 治理、观测和部署能力

治理能力包括：

- policy engine；
- tool permission；
- trace；
- eval harness；
- release gate；
- audit log；
- cost control；
- deployment strategy。

这些能力决定 Agent 能不能长期运行，而不仅是 demo 能不能跑通。

### 9.9.6 框架选型对比表

| 维度 | LangGraph | AutoGen | Microsoft Agent Framework | 自研 Runtime |
|:---|:---|:---|:---|:---|
| 核心强项 | 有状态图、durable execution | 多 Agent 编程模型 | 企业工作流和平台化 | 完全贴合内部系统 |
| 适合任务 | 长任务、可恢复流程 | 多角色协作 | 企业级 Agent 应用 | 特殊约束或轻量场景 |
| 状态管理 | 强 | 中 | 强 | 取决于实现 |
| 多 Agent | 中 | 强 | 中到强 | 取决于实现 |
| 工具接入 | 需集成 | 需集成 | 生态集成 | 自行建设 |
| 治理能力 | 需补齐 | 需补齐 | 较强 | 自行建设 |
| 主要风险 | 图设计成本 | 协调成本 | 平台复杂度 | 重复造轮子 |

---

## 9.10 设计自己的 Agent 平台

### 9.10.1 最小生产平台清单

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

### 9.10.2 控制面、执行面、集成面与治理面

可以把平台拆成四个面：

```text
Control Plane: registry / policy / config / release
Execution Plane: workflow runtime / task runner / sandbox
Integration Plane: tools / resources / connectors / MCP
Governance Plane: trace / eval / audit / cost / monitoring
```

这四个面不一定要拆成四个服务，但职责要清楚。

### 9.10.3 生产级运行时事件模型

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

### 9.10.4 多租户与权限边界

平台化后必须处理多租户：

- 用户身份；
- 项目边界；
- 工具权限；
- 数据权限；
- Memory scope；
- 知识库 ACL；
- 审计归属。

多 Agent 和 subagent 不能继承无限权限。每次 handoff 都应该重新计算权限和上下文边界。

### 9.10.5 从单应用 Agent 到共享平台的演进路线

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

## 9.11 工程取舍与选型清单

### 9.11.1 图式编排 vs 自由 Agent Loop

自由 loop 灵活，图式编排可控。

建议：

- 低风险探索用自由 loop；
- 生产流程用显式 workflow；
- 高风险动作必须进入审批节点；
- 长任务必须有 checkpoint。

### 9.11.2 单 Agent vs 多 Agent

优先从单 Agent + Workflow 开始。只有当存在清晰角色边界、并行收益或互审需求时，再引入多 Agent。

多 Agent 的收益必须超过协调成本。

### 9.11.3 框架 vs 自研

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

### 9.11.4 平台化 vs 单应用内嵌

单应用内嵌适合早期验证。平台化适合多团队复用。

平台化触发信号：

- 多个团队重复接模型；
- 多个团队重复封装工具；
- 权限和审计开始分散；
- 需要统一成本控制；
- 需要统一 eval 和 release gate；
- 长任务恢复成为共性问题。

### 9.11.5 选型决策树

```text
任务是否有明确流程？
  ├─ 是：优先 workflow / graph
  └─ 否：继续问

任务是否需要长时间运行或恢复？
  ├─ 是：必须 checkpoint / durable execution
  └─ 否：继续问

是否需要多个角色协作？
  ├─ 是：考虑 multi-agent topology
  └─ 否：单 Agent + workflow

是否多团队复用？
  ├─ 是：建设平台控制面
  └─ 否：先应用内嵌

是否主要问题是外部能力接入？
  ├─ 是：建设 tool runtime / connector / MCP integration
  └─ 否：不要把 MCP 当编排层
```

---

## 9.12 常见失败模式与修复路径

### 9.12.1 工作流卡死

表现：

- 等待一个永远不会发生的事件；
- 状态无法转移；
- Agent 反复尝试同一步。

修复：

- 每个等待状态设置 timeout；
- 每个状态定义允许动作；
- 增加 fallback 和 human escalation；
- trace 中记录卡住原因。

### 9.12.2 状态丢失或重复执行

表现：

- 进程重启后任务从头开始；
- 工具副作用重复发生；
- 审批后找不到上下文。

修复：

- checkpoint 关键状态；
- 外部动作使用幂等键；
- action log 记录请求和响应；
- resume 时从状态恢复，不从聊天历史猜。

### 9.12.3 多 Agent 循环争论

表现：

- reviewer 不断要求修改；
- writer 不断生成新版本；
- coordinator 无法决策。

修复：

- 设置最大轮数；
- 定义验收标准；
- 引入 final decision owner；
- 低置信度转人工。

### 9.12.4 工具副作用失控

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

### 9.12.5 平台抽象过重

表现：

- 简单任务也要配置复杂 graph；
- 业务团队接入成本高；
- 框架概念多于业务价值。

修复：

- 从 MVP runtime 开始；
- 常见模式模板化；
- 保留轻量 escape hatch；
- 平台能力按复用痛点演进。

### 9.12.6 Trace 不足导致无法复盘

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

## 本章小结

执行编排是 Agent 从 demo 走向生产的关键层。它不只属于多 Agent，单 Agent 只要任务变长、动作变多、风险变高，也需要 workflow、state machine、checkpoint 和 human-in-the-loop。

本章的核心结论是：

1. Agent Loop 适合原型，生产系统需要 Workflow Runtime；
2. 单 Agent 也需要显式编排；
3. Workflow Pattern 可以服务单 Agent，也可以服务多 Agent；
4. Multi-Agent 是执行编排的复杂形态，不是默认起点；
5. LangGraph、AutoGen、Microsoft Agent Framework 分别从状态图、多 Agent 编程模型和企业运行时角度提供抽象；
6. MCP 是 Integration Layer，不是 Orchestration Layer；
7. 平台化的关键不是框架名称，而是状态、工具、权限、审批、恢复、trace、eval 和发布治理是否形成闭环。

一句话总结：

```text
生产级 Agent 不是一个更聪明的 while loop，而是一个能组织任务、保存状态、控制副作用、插入人工、恢复失败、记录证据的运行时系统。
```

---

## 参考资料

1. [LangGraph Overview - LangChain Docs](https://docs.langchain.com/oss/python/langgraph/overview)
2. [LangGraph Durable Execution - LangChain Docs](https://docs.langchain.com/oss/python/langgraph/durable-execution)
3. [LangGraph Persistence - LangChain Docs](https://docs.langchain.com/oss/python/langgraph/persistence)
4. [AutoGen Stable Documentation - Microsoft](https://microsoft.github.io/autogen/stable/index.html)
5. [AutoGen AgentChat User Guide - Microsoft](https://microsoft.github.io/autogen/stable/user-guide/agentchat-user-guide/index.html)
6. [AutoGen: Enabling Next-Gen LLM Applications via Multi-Agent Conversation - Microsoft Research](https://www.microsoft.com/en-us/research/publication/autogen-enabling-next-gen-llm-applications-via-multi-agent-conversation-framework/)
7. [Microsoft Agent Framework Overview - Microsoft Learn](https://learn.microsoft.com/en-us/agent-framework/overview/)
8. [Agent Framework Workflow Orchestrations - Microsoft Learn](https://learn.microsoft.com/en-us/agent-framework/workflows/orchestrations/)
9. [Model Context Protocol Specification](https://modelcontextprotocol.io/specification/2025-03-26/architecture)
