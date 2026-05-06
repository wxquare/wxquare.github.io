# 第14章 Agent 平台与编排框架解析：LangGraph、AutoGen、MCP 生态

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

---

## 14.1 Agent 编排为什么需要框架

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

---

## 14.2 LangGraph：把 Agent 表示成有状态图

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

### 适用场景

LangGraph 式编排适合：

- 有明确状态流转的 Agent；
- 需要中断和恢复的长任务；
- 需要人工审批的流程；
- 需要可视化和调试执行路径的系统；
- 需要把 Agent 放进后端服务的团队。

---

## 14.3 AutoGen：从多 Agent 对话到协作模式

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

---

## 14.4 Microsoft Agent Framework：企业化 Agent Runtime

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

---

## 14.5 MCP 生态：Agent 工具和资源的连接协议

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

---

## 14.6 平台能力矩阵

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

---

## 14.7 设计自己的 Agent 平台：分层架构

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

---

## 14.8 工程取舍

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

---

## 14.9 可借鉴点

从成熟框架里可以抽象出这些设计原则：

- Agent 流程要显式建模，不要藏在无限循环里；
- 状态要持久化，不能只存在上下文窗口；
- 人工审批要成为 workflow 节点；
- 多 Agent 要有角色边界和终止条件；
- 工具连接协议和业务工具实现要解耦；
- Trace、成本和失败原因必须贯穿每一步；
- 平台层应该提供统一模型网关、工具注册、权限和观测。

---

## 本章小结

Agent 平台和编排框架的共同趋势是：把不确定的模型行为放进确定的工程结构里。

- LangGraph 代表有状态图和 durable execution；
- AutoGen 代表多 Agent 对话和角色协作；
- Microsoft Agent Framework 代表企业级 Agent Runtime；
- MCP 代表工具和资源连接协议。

一句话总结：

> 生产级 Agent 不是一个更聪明的 while loop，而是一个能保存状态、连接工具、插入人工、恢复失败、记录证据的运行时系统。

---

## 参考资料

1. [LangGraph Durable Execution - LangChain Docs](https://docs.langchain.com/oss/python/langgraph/durable-execution)
2. [Microsoft Agent Framework Overview - Microsoft Learn](https://learn.microsoft.com/en-us/agent-framework/overview/)
3. [AutoGen: Enabling Next-Gen LLM Applications via Multi-Agent Conversation - Microsoft Research](https://www.microsoft.com/en-us/research/publication/autogen-enabling-next-gen-llm-applications-via-multi-agent-conversation-framework/)
4. [Model Context Protocol Specification](https://modelcontextprotocol.io/specification/2025-03-26/architecture)
