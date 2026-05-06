# 第8章 Agent 平台与编排框架解析：LangGraph、AutoGen、MCP 生态

> Agent 平台的价值，不是替你写一个更长的 prompt，而是把推理、工具、状态、人工审批和恢复机制变成可组合的工程抽象。

## 引言

单个 Agent Demo 很容易写：一个模型、一组工具、一个循环。但当任务变长、工具变多、多人协作、需要审批和恢复时，Demo 架构很快会失控。

成熟 Agent 平台和编排框架通常要解决这些问题：

- 多步骤任务如何表示？
- 中间状态保存在哪里？
- 工具失败后如何恢复？
- 人类审批如何插入流程？
- 多 Agent 如何通信和分工？
- 如何观察每一步的输入、输出和成本？
- 如何把系统从 Demo 变成可运维服务？

LangGraph、AutoGen、Microsoft Agent Framework、MCP 生态分别从不同角度回答这些问题。本章不做框架教程，而是分析它们背后的系统设计模式。

本章放在第二部分，是因为它并不是某个成熟产品案例，而是前面几章工程抽象的汇合点：

- 第 5 章讲 Agent 架构决策；
- 第 6 章讲 Tool Calling、Skills 和 MCP；
- 第 7 章讲状态机、多 Agent 和工作流；
- 本章进一步讨论这些能力如何沉淀为框架、运行时和平台控制面。

读这一章时，不要把 LangGraph、AutoGen、Agent Framework 和 MCP 看成同一层的竞品。更准确的关系是：

```text
Agent Application
  │
  ├─ Orchestration Runtime: LangGraph / Agent Framework Workflows
  ├─ Multi-Agent Programming Model: AutoGen / Agent Framework Agents
  ├─ Tool Integration Protocol: MCP
  └─ Platform Governance: policy / trace / eval / deployment / cost
```

也就是说，一个生产级 Agent 系统通常不是“选一个框架就结束”，而是把编排、工具连接、状态持久化、权限治理和观测评估组合起来。

---

## 8.1 Agent 编排为什么需要框架

最小 Agent Loop 通常长这样：

```python
while not done:
    action = llm(context)
    result = execute_tool(action)
    context.append(result)
```

这段代码适合原型，但不适合生产。原因是生产系统需要显式处理：

- 状态保存；
- 步骤重放；
- 失败恢复；
- 超时和预算；
- 人工审批；
- 并发分支；
- 版本管理；
- trace 和评估。

一旦这些需求出现，Agent Loop 就会演化为 workflow runtime。

```text
Agent Demo
  │
  ▼
Tool-using Agent
  │
  ▼
Stateful Workflow
  │
  ▼
Multi-Agent Runtime
  │
  ▼
Agent Platform
```

从系统设计角度看，编排框架至少要把五类隐含问题显式化。

| 隐含问题 | Demo 里的做法 | 生产级平台里的做法 |
|:---|:---|:---|
| 控制流 | while loop 里临时判断 | graph、workflow、router、handoff |
| 状态 | 拼进 prompt 或内存变量 | checkpoint、thread、session、state store |
| 副作用 | 直接执行工具 | tool runtime、幂等键、审批、补偿 |
| 人工介入 | 聊天里问一句“可以吗” | workflow interrupt、approval node、resume token |
| 可观测性 | 打印日志 | trace、span、metrics、cost、eval replay |

这也是为什么“Agent 框架”通常不是模型 SDK 的简单包装。它更像一个小型 workflow engine，只是 workflow 的某些节点由 LLM 决策，某些边由工具结果、人工审批或策略引擎决定。

### 从 Agent Loop 到 Runtime

最小 loop 只有三件事：让模型想、调用工具、把结果塞回上下文。Runtime 要多做很多确定性工作。

```text
User Task
  │
  ▼
Task Envelope
  ├─ goal
  ├─ constraints
  ├─ permissions
  ├─ budget
  └─ expected evidence
  │
  ▼
Workflow Runtime
  ├─ plan / route
  ├─ execute node
  ├─ checkpoint state
  ├─ dispatch tool
  ├─ interrupt for human
  ├─ verify output
  └─ emit trace
```

所以框架的价值不是“让 Agent 更自动”，而是让 Agent 的自动化具备边界：哪些步骤由模型决定，哪些步骤由图和策略决定，哪些动作必须暂停等待人，哪些结果必须经过 verifier。

---

## 8.2 LangGraph：把 Agent 表示成有状态图

LangGraph 的核心抽象是图：节点表示计算步骤，边表示状态转移，状态在节点之间流动。

```text
┌──────────┐
│ Planner  │
└────┬─────┘
     ▼
┌──────────┐      tool needed      ┌──────────┐
│  Agent   │ ────────────────────► │  Tools   │
└────┬─────┘                       └────┬─────┘
     │ final answer                     │ result
     ▼                                  ▼
┌──────────┐ ◄──────────────────────────┘
│  End     │
└──────────┘
```

这种设计的价值是：Agent 不再是一段隐藏在 while loop 里的逻辑，而是一个可观察、可恢复、可测试的状态机。

从工程抽象看，LangGraph 的关键不是“画图”，而是把 Agent loop 拆成几类可组合对象：

| 抽象 | 含义 | 工程价值 |
|:---|:---|:---|
| State | 节点之间共享的结构化状态 | 避免把所有历史都塞进 prompt |
| Node | 一步计算，可以是 LLM、工具、函数、人工节点 | 让步骤可测试、可替换 |
| Edge | 下一步路由，可以是固定边或条件边 | 把控制流显式化 |
| Checkpointer | 保存每一步状态快照 | 支持恢复、回放、人审 |
| Thread | 一次工作流实例的状态序列 | 支持多会话和多任务隔离 |
| Interrupt | 在某一步暂停并等待外部输入 | 支持 human-in-the-loop |

一个典型 LangGraph 思维方式是：先定义业务状态，再定义哪些节点会改变状态。

```python
class IncidentState(TypedDict):
    alert: dict
    evidence: list[str]
    hypothesis: str | None
    remediation: str | None
    approval: Literal["pending", "approved", "rejected"]
    final_report: str | None
```

这和“把所有东西都放进 messages”不同。messages 是模型上下文，state 是运行时事实。生产系统应该尽量把关键事实保存在 state 里，再由 Context Builder 决定每一轮给模型看什么。

### Durable Execution

长任务最怕中途失败。如果状态只存在内存里，进程挂掉就要从头来过。

Durable execution 的思想是：在关键步骤保存状态，使流程可以从检查点恢复。

```text
Step 1 completed → checkpoint
Step 2 completed → checkpoint
Human approval pending → checkpoint
Resume after approval → continue
```

这对 Agent 尤其重要，因为 LLM 调用可能超时、工具可能失败、人类审批可能隔天才发生。

更深一层，durable execution 的难点在 **重放一致性**。如果一个节点里同时做了“调用模型、写数据库、发消息、改代码”，恢复时就很难判断哪些副作用已经发生、哪些可以重试。

成熟实现通常要求：

- 节点内部副作用要拆成可记录的 task；
- 每个外部动作要有幂等键；
- checkpoint 保存的是状态变化，不是只保存聊天文本；
- replay 时不能重复执行已经成功的副作用；
- 恢复点要能解释“上次停在哪、为什么停、下一步是什么”。

因此，LangGraph 更适合那些需要长期运行、可恢复、可审计的 Agent，而不是一次性问答。

### Human-in-the-loop

图式编排很适合插入人工节点：

```text
Diagnose Incident
  │
  ▼
Propose Remediation
  │
  ▼
Human Approval
  │
  ├─ approved → Execute
  └─ rejected → Explain / Alternative
```

关键是：人工审批不是聊天里的“确认一下”，而是 workflow state 的一部分。

把人工节点做成状态的一部分，能带来三个好处：

1. **可恢复**：审批人明天点批准，workflow 仍然能从原状态继续；
2. **可审计**：谁在什么时间批准了什么 action 可以进入 audit log；
3. **可治理**：不同风险等级的工具可以走不同审批路径。

例如告警自愈 Agent 可以这样分流：

```text
read logs             -> auto
query metrics         -> auto
restart one pod       -> ask on-call
scale deployment      -> ask service owner
change database       -> deny by default
```

这就是把权限模型和 workflow graph 结合起来。

### Time Travel 与 Debugging

有状态图的另一个价值是调试。Agent 出错时，团队需要回答：

- 哪个节点产生了错误假设？
- 当时 state 里有哪些证据？
- 哪个工具结果被误解了？
- 如果从某个 checkpoint 改一条状态继续，会发生什么？

这类问题靠普通日志很难回答。checkpoint + state history 可以让开发者回到某个中间状态，修改输入或策略后重新执行后续节点。这对 Agent eval 也很重要：失败样本不是一段聊天记录，而是一条可重放的状态轨迹。

### 适用场景

LangGraph 式编排适合：

- 有明确状态流转的 Agent；
- 需要中断和恢复的长任务；
- 需要人工审批的流程；
- 需要可视化和调试执行路径的系统；
- 需要把 Agent 放进后端服务的团队。

不适合的场景也要明确：

- 单次轻量问答；
- 需求还不稳定、流程每天变化的早期探索；
- 每一步都高度开放，没有稳定状态边界的研究任务；
- 团队还没有状态、测试、观测和审批的工程习惯。

---

## 8.3 AutoGen：从多 Agent 对话到协作模式

AutoGen 的核心思想是用多个可配置 Agent 通过对话完成任务。

```text
User Proxy
   │
   ▼
Planner Agent ──► Coder Agent ──► Reviewer Agent
      ▲                                │
      └────────── feedback ────────────┘
```

这种模型强调角色分工：

- Planner 负责拆任务；
- Coder 负责实现；
- Reviewer 负责审查；
- User Proxy 负责人类输入或工具执行；
- Group Chat Manager 负责路由对话。

### 多 Agent 的价值

多 Agent 不是为了“看起来智能”，而是为了三个工程目标：

1. **上下文隔离**：不同角色只保留相关上下文；
2. **职责分离**：分析、执行、审查分开；
3. **制衡机制**：Reviewer 能发现 Writer 的盲点。

### 多 Agent 的风险

多 Agent 也会带来额外复杂度：

- 对话轮数增加，成本上升；
- 角色边界不清会互相抢活；
- 缺少全局状态时容易循环；
- 聚合结果需要额外设计；
- 每个 Agent 的权限需要单独控制。

因此，多 Agent 应该用于真实需要分工、审查或并行的任务，而不是默认架构。

### AgentChat、Core 与 Extensions

新版 AutoGen 的一个重要变化是分层更清晰：

| 层级 | 适合谁 | 解决什么问题 |
|:---|:---|:---|
| Studio | 非代码化原型 | 快速组装和观察 agent team |
| AgentChat | 应用开发者 | 用高层 API 构建 conversational single / multi-agent app |
| Core | 平台和框架开发者 | 事件驱动、可扩展、可分布式的多 Agent runtime |
| Extensions | 集成层 | 模型客户端、MCP workbench、Docker executor、gRPC runtime 等 |

这说明多 Agent 框架正在从“几个 agent 对话”演进为更完整的 programming model。高层 API 让团队快速搭建 planner / coder / reviewer；底层 Core 则提供 message、topic、runtime、serialization、distributed execution 等平台能力。

### Team Pattern：多 Agent 不是群聊，而是拓扑

一个常见误区是把多 Agent 理解成“多个角色在同一个聊天室里自由发言”。生产系统更应该把它设计成明确拓扑。

| 模式 | 结构 | 适合场景 | 风险 |
|:---|:---|:---|:---|
| Round-robin | 按固定顺序发言 | 教学、固定审查链 | 容易产生无意义轮次 |
| Selector / Manager | 中央调度者选择下一个 Agent | 任务分工、研究综合 | manager 成为瓶颈 |
| Handoff | Agent 根据条件移交控制权 | 客服、诊断、专家系统 | 移交条件难调 |
| GraphFlow | 多 Agent 组成有向图 | 流程明确的协作任务 | 设计成本更高 |
| Concurrent | 多 Agent 并行探索 | 多方案比较、并行 review | 聚合和冲突处理复杂 |

多 Agent 的关键不是 agent 数量，而是 **谁拥有全局状态、谁决定下一步、谁负责终止**。

```text
Task
  │
  ▼
Manager
  ├─ Researcher A
  ├─ Researcher B
  ├─ Coder
  └─ Reviewer
       │
       ▼
   Aggregator
       │
       ▼
   Final Decision
```

如果没有 Aggregator，多个 Agent 的输出只是多份文本；如果没有 termination condition，多 Agent 很容易陷入互相补充、互相质疑但没有交付的循环。

### 工具执行和代码执行边界

AutoGen 生态常见的一个强能力是让 Agent 编写并执行代码，例如数据分析、仿真、批处理和代码生成。这个能力必须和 sandbox 绑定。

生产系统至少要回答：

- 生成的代码在哪里执行？
- 是否使用容器隔离？
- 能否访问网络？
- 能否读取本地文件和环境变量？
- stdout、stderr、文件产物如何回传？
- 失败时是否允许 Agent 自动修复并重跑？

这和 Coding Agent 的权限模型是同一个问题：模型可以提议代码，Runtime 决定代码能否执行、在哪里执行、执行结果如何进入上下文。

---

## 8.4 Microsoft Agent Framework：企业化 Agent Runtime

Microsoft Agent Framework 的方向是把 AutoGen 的多 Agent 抽象与 Semantic Kernel 的企业能力合并，形成更完整的 Agent 应用框架。

从系统设计角度看，它强调几个企业级能力：

- session-based state management；
- type safety；
- middleware / filters；
- telemetry；
- workflow orchestration；
- 多模型和 embedding 支持；
- 长任务和 human-in-the-loop。

这说明 Agent 平台正在从“研究框架”走向“企业应用运行时”。

### 企业平台的关键抽象

```text
Agent
  ├─ Instructions
  ├─ Tools
  ├─ State
  ├─ Middleware
  ├─ Telemetry
  └─ Workflow
```

对于企业来说，平台价值不只在模型调用，而在统一：

- 安全策略；
- 日志和审计；
- 状态管理；
- 工具注册；
- 模型路由；
- 版本发布；
- 成本追踪。

### Agents vs Workflows

Microsoft Agent Framework 的一个值得借鉴的判断是：不是所有任务都应该做成 Agent，也不是所有 Agent 任务都应该做成自由对话。

| 任务特征 | 更适合 Agent | 更适合 Workflow |
|:---|:---|:---|
| 目标是否开放 | 开放、需要探索 | 明确、有固定步骤 |
| 工具使用 | 由模型动态决定 | 由流程显式安排 |
| 控制需求 | 允许一定不确定性 | 需要强控制和审计 |
| 状态管理 | 会话状态为主 | checkpoint / durable state |
| 人工介入 | 临时追问 | 审批节点和恢复点 |

更进一步，如果一个任务可以用普通函数稳定完成，就不应该强行用 Agent。Agent Framework 的启示不是“万物 Agent 化”，而是把 Agent 放在普通程序、workflow 和企业治理体系中。

### 企业编排模式

企业平台通常会内置几类编排模式：

| 模式 | 含义 | 典型场景 |
|:---|:---|:---|
| Sequential | Agent 或函数按固定顺序执行 | 内容生成、审批链、报告生成 |
| Concurrent | 多个 Agent 并行执行后汇总 | 多方案评估、多维度审查 |
| Handoff | 一个 Agent 根据上下文移交给另一个 Agent | 客服、IT 支持、专家路由 |
| Group Chat | 多 Agent 在共享上下文中协作 | 复杂问题讨论、方案评审 |
| Magentic / Manager | manager 动态协调多个 specialist | 开放任务、研究、复杂执行 |

这些模式和第 7 章的 workflow / state machine / multi-agent pattern 是同一组思想。区别在于，平台框架把它们做成可复用运行时能力，而不是每个业务团队重复实现。

### Middleware、Telemetry 与企业接入

企业级 Agent Runtime 还需要几类“非智能但必要”的能力：

- **Middleware / Filters**：在模型调用、工具调用、消息处理前后插入策略；
- **Telemetry**：把 token、延迟、工具错误、节点耗时、用户反馈写入统一观测系统；
- **Session State**：保存多轮对话、审批状态、工具结果和业务上下文；
- **Model Provider Abstraction**：支持不同模型供应商和部署形态；
- **MCP Client**：以统一方式接入外部工具；
- **Type Safety**：让 workflow 节点输入输出可校验、可重构。

这些能力看起来不酷，却决定 Agent 能不能进入企业生产环境。

---

## 8.5 MCP 生态：Agent 工具和资源的连接协议

MCP（Model Context Protocol）解决的是另一个维度的问题：**AI 应用如何以标准方式连接外部工具和数据源。**

在 Agent 平台中，MCP 通常处于工具连接层：

```text
Agent Runtime
  │
  ▼
MCP Client
  │
  ├─ MCP Server: GitHub
  ├─ MCP Server: Postgres
  ├─ MCP Server: Slack
  └─ MCP Server: Internal Tools
```

MCP 的价值在于统一暴露：

- Resources：可读取上下文；
- Tools：可执行操作；
- Prompts：可复用任务模板。

### MCP 与编排框架的关系

MCP 不是 workflow engine，也不是 agent framework。更准确地说：

| 层次 | 负责什么 | 示例 |
|:---|:---|:---|
| Agent Framework | 决策、状态、流程 | LangGraph、Agent Framework |
| Tool Protocol | 工具发现和调用协议 | MCP |
| Business Service | 真实业务能力 | GitHub、数据库、工单系统 |

编排框架决定“什么时候调用什么”，MCP 决定“如何发现和调用外部能力”。

### MCP 在平台里的真实位置

在平台架构里，MCP 应该被放在 Integration Layer，而不是 Orchestration Layer。

```text
Workflow / Agent Runtime
  │  decides when to call
  ▼
Tool Runtime
  │  validates, authorizes, times out, logs
  ▼
MCP Client
  │  JSON-RPC / transport / capability negotiation
  ▼
MCP Server
  │  adapts external system
  ▼
Business API / DB / SaaS
```

这条链路里每一层责任不同：

- Workflow 决定当前任务是否需要某个能力；
- Tool Runtime 决定参数是否合法、权限是否允许、结果如何裁剪；
- MCP Client / Server 负责协议和能力发现；
- Business API 负责真实业务状态变更。

如果把 MCP Server 直接当成“可信工具”，平台很容易失控。尤其是 MCP Server 可能连接数据库、文件系统、工单、邮箱、日志平台和内部 API，它的安全边界应该和普通后端服务一样严肃。

### Tools、Resources、Prompts 的平台治理

MCP 暴露的三类能力在治理上应该区别对待。

| 能力 | 平台语义 | 风险 | 治理方式 |
|:---|:---|:---|:---|
| Resources | 可读上下文和外部资料 | 敏感数据泄露、上下文污染 | scope、脱敏、引用来源 |
| Tools | 可执行动作 | 越权、副作用、破坏性操作 | schema、权限、审批、审计 |
| Prompts | 可复用任务模板 | 注入错误流程、绕过团队规范 | 版本管理、review、测试 |

这和第 6 章的结论一致：MCP 是连接协议，不替代 Tool Runtime。真正的生产能力来自 MCP 之外的权限、审计、限流、重试、脱敏和 eval。

### MCP 与 Agent 平台的组合模式

常见组合有三种。

| 模式 | 做法 | 适合场景 |
|:---|:---|:---|
| Local MCP | Agent host 启动本地 MCP server | 本地开发、文件系统、CLI 工具 |
| Remote MCP | MCP server 独立部署，通过 HTTP 接入 | SaaS、企业内部平台、共享工具 |
| Managed MCP Marketplace | 平台统一发布和治理 MCP server | 大企业、多团队、合规环境 |

企业内部更推荐第二和第三种：把 MCP Server 当成受治理的 integration service，而不是让每个开发者在本机随意启动一堆连接内部系统的脚本。

---

## 8.6 平台能力矩阵

| 能力 | LangGraph | AutoGen | Microsoft Agent Framework | MCP |
|:---|:---|:---|:---|:---|
| 状态图 | 强 | 弱 | 强 | 不负责 |
| 多 Agent 会话 | 可实现 | 强 | 强 | 不负责 |
| Durable execution | 强 | 依赖实现 | 强 | 不负责 |
| Human-in-the-loop | 强 | 可实现 | 强 | 不负责 |
| 企业遥测 | 依赖平台 | 依赖实现 | 强 | 不负责 |
| 工具协议 | 依赖集成 | 依赖集成 | 依赖集成 | 强 |
| 资源发现 | 依赖实现 | 依赖实现 | 依赖实现 | 强 |
| 主要定位 | 状态化编排 | 多 Agent 协作 | 企业 Agent Runtime | 工具和资源连接 |

这张表的重点不是比较谁更好，而是提醒：这些框架解决的问题不同。

如果按生产平台的组成来看，可以进一步拆成“控制面”和“执行面”。

| 平台层 | 主要问题 | 更相关的框架或协议 |
|:---|:---|:---|
| Control Plane | 谁能创建 Agent、发布 workflow、配置权限、查看审计 | Agent Framework、企业自研平台 |
| Orchestration Plane | 流程如何执行、暂停、恢复、分支和合并 | LangGraph、Agent Framework Workflows |
| Collaboration Plane | 多 Agent 如何通信、分工、终止和聚合 | AutoGen、Agent Framework Agents |
| Integration Plane | 工具和资源如何被发现、调用和治理 | MCP、API Gateway、Connector |
| Execution Plane | 工具在哪里执行、如何隔离、如何处理副作用 | sandbox、container、serverless、worker |
| Governance Plane | 如何评估、观测、限流、审计和回滚 | Evals、Guardrails、Telemetry、Policy Engine |

成熟平台通常不会只采用其中一个开源框架，而是围绕这些平面进行组合。

---

## 8.7 设计自己的 Agent 平台：分层架构

一个企业内部 Agent 平台可以按五层设计。

```text
┌─────────────────────────────────────────────────────────────┐
│                     Agent Platform                           │
├─────────────────────────────────────────────────────────────┤
│  App Layer                                                   │
│  ├─ Support Agent                                            │
│  ├─ Coding Agent                                             │
│  └─ Data Analyst Agent                                       │
│                                                             │
│  Orchestration Layer                                         │
│  ├─ Workflow Graph                                           │
│  ├─ Multi-Agent Router                                       │
│  ├─ Human Approval                                           │
│  └─ State Store                                              │
│                                                             │
│  Runtime Layer                                               │
│  ├─ Model Gateway                                            │
│  ├─ Tool Runtime                                             │
│  ├─ Memory / RAG                                             │
│  └─ Guardrails                                               │
│                                                             │
│  Integration Layer                                           │
│  ├─ MCP Servers                                              │
│  ├─ API Gateway                                              │
│  ├─ Connectors                                               │
│  └─ Secret Manager                                           │
│                                                             │
│  Governance Layer                                            │
│  ├─ Trace / Metrics / Cost                                   │
│  ├─ Eval Datasets                                            │
│  ├─ Policy Engine                                            │
│  └─ Audit Logs                                               │
└─────────────────────────────────────────────────────────────┘
```

这个分层的好处是边界清楚：

- App 只关心具体业务；
- Orchestration 关心流程和状态；
- Runtime 关心模型、工具和安全；
- Integration 关心外部系统；
- Governance 关心质量和责任链。

### 控制面与执行面

更贴近生产的拆法，是把 Agent 平台分成 Control Plane 和 Data Plane。

```text
Control Plane
  ├─ Agent / Workflow Registry
  ├─ Tool / MCP Registry
  ├─ Policy & Permission
  ├─ Eval Dataset
  ├─ Deployment Config
  └─ Audit Console

Data Plane
  ├─ Runtime Worker
  ├─ Model Gateway
  ├─ Tool Executor
  ├─ State Store
  ├─ Trace Collector
  └─ Artifact Store
```

Control Plane 负责“定义和治理”，Data Plane 负责“执行和记录”。这条边界很关键：业务团队可以发布一个 Agent，但不应该直接绕过平台访问所有工具；Runtime Worker 可以执行任务，但不应该自己决定组织级安全策略。

### 生产级运行时事件模型

无论底层用 LangGraph、AutoGen 还是自研 runtime，事件日志都应该足够结构化。

```json
{
  "run_id": "run_123",
  "thread_id": "thread_456",
  "workflow": "incident-diagnosis",
  "node": "query_logs",
  "event": "tool_call_completed",
  "tool": "mcp.log.search",
  "input_ref": "redacted://inputs/789",
  "output_ref": "artifact://outputs/abc",
  "latency_ms": 842,
  "token_usage": {
    "input": 1200,
    "output": 320
  },
  "policy_decision": "allow",
  "created_at": "2026-05-06T10:00:00Z"
}
```

这类事件是后续 eval、debug、成本分析、审计和事故复盘的基础。没有事件模型，Agent 平台很快会退化成“跑过但说不清为什么”的黑盒。

### 多租户与权限边界

企业 Agent 平台还必须处理多租户问题：

- 不同团队能看到哪些 Agent？
- 一个 Agent 能调用哪些 MCP server？
- 一个用户的权限是否能透传到工具层？
- workflow 产生的 artifact 保存在哪里？
- trace 中的敏感字段如何脱敏？
- 离职用户创建的 automation 如何接管？

这些问题和模型能力无关，但决定平台能不能在组织里长期运行。

---

## 8.8 工程取舍

### 1. 图式编排 vs 自由 Agent Loop

自由循环灵活，但难以审计；图式编排可控，但前期设计成本更高。

建议：

- 低风险探索任务用自由 loop；
- 生产流程用显式 workflow；
- 高风险动作必须进入审批节点。

### 2. 单 Agent vs 多 Agent

单 Agent 更简单，多 Agent 更适合职责分离。判断标准不是任务复杂度，而是是否存在清晰的角色边界。

适合多 Agent：

- Writer / Reviewer；
- Planner / Executor；
- Researcher / Synthesizer；
- Security Reviewer / Performance Reviewer。

不适合多 Agent：

- 简单问答；
- 单次工具查询；
- 没有可并行边界的小任务。

### 3. 框架 vs 自研

自研适合学习和小型系统，但生产系统很快会遇到状态、恢复、观测和审批问题。

如果你的系统已经需要这些能力，就应该优先评估成熟框架，而不是继续堆 prompt 和 while loop。

### 4. 平台化 vs 单应用内嵌

不是所有团队都需要一开始就建设 Agent 平台。

适合先做单应用内嵌：

- 只有一个明确业务场景；
- 工具数量少；
- 用户群小；
- 还在验证 Agent 是否有价值；
- 没有跨团队复用需求。

适合平台化：

- 多个团队都在接模型和工具；
- 重复建设权限、日志、eval、MCP connector；
- 需要统一审批、审计和成本控制；
- Agent 任务开始长时间运行；
- 需要将能力发布给非工程用户。

一个常见演进路线是：

```text
single app agent
  -> shared tool runtime
  -> shared workflow runtime
  -> shared eval / trace platform
  -> managed agent platform
```

### 框架选择决策树

可以用下面的问题做初筛。

```text
任务是否有明确流程？
  ├─ 是：优先考虑 LangGraph / Agent Framework Workflow
  └─ 否：继续问

是否需要多个角色互相协作？
  ├─ 是：考虑 AutoGen / Agent Framework multi-agent
  └─ 否：继续问

核心问题是否是接入外部工具和数据源？
  ├─ 是：优先设计 MCP / Tool Runtime / Connector
  └─ 否：继续问

是否需要企业级权限、遥测、部署和多团队治理？
  ├─ 是：考虑 Agent Framework 或自研平台控制面
  └─ 否：先用轻量 agent loop + 明确 trace
```

这个决策树不是为了排他选择，而是避免把所有问题都扔给一个框架。

---

## 8.9 可借鉴点

从成熟框架里可以抽象出这些设计原则：

- Agent 流程要显式建模，不要藏在无限循环里；
- 状态要持久化，不能只存在上下文窗口；
- 人工审批要成为 workflow 节点；
- 多 Agent 要有角色边界和终止条件；
- 工具连接协议和业务工具实现要解耦；
- Trace、成本和失败原因必须贯穿每一步；
- 平台层应该提供统一模型网关、工具注册、权限和观测。

### 最小生产平台清单

如果要从零落地一个企业内部 Agent 平台，最小版本不应该先追求“可视化拖拽编排”。更稳妥的 MVP 是：

| 模块 | 最小能力 | 不做会怎样 |
|:---|:---|:---|
| Model Gateway | 统一模型调用、超时、重试、成本记录 | 各团队重复封装，成本不可控 |
| Tool Registry | 工具 schema、风险等级、owner、审计字段 | 工具滥用，没人知道谁能做什么 |
| Workflow Runtime | 状态、checkpoint、interrupt、resume | 长任务失败无法恢复 |
| Policy Engine | read / write / execute / network 权限 | 只能靠 prompt 控制风险 |
| Trace Store | run、node、tool、token、latency、error | 调试和复盘困难 |
| Eval Harness | 固定回归集、失败样本、版本对比 | 模型或 prompt 一改就退化 |
| Artifact Store | 报告、diff、图表、日志片段 | 输出散落在聊天里，无法审查 |

这套 MVP 看起来“平台味”很重，但它们正是生产级 Agent 和 Demo 的分界线。

---

## 本章小结

Agent 平台和编排框架的共同趋势是：把不确定的模型行为放进确定的工程结构里。

- LangGraph 代表有状态图和 durable execution；
- AutoGen 代表多 Agent programming model、team pattern 和协作拓扑；
- Microsoft Agent Framework 代表企业级 Agent Runtime、workflow、middleware 和 telemetry；
- MCP 代表工具和资源连接协议，但不替代 Tool Runtime 和权限治理；
- 生产级 Agent 平台需要同时建设控制面、执行面、集成面和治理面。

一句话总结：

> 生产级 Agent 不是一个更聪明的 while loop，而是一个能保存状态、连接工具、插入人工、恢复失败、记录证据的运行时系统。

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
