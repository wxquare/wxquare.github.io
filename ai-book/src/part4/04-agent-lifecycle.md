# 第17章 Agent 系统全生命周期工程

> 面试官说“全生命周期开发”，真正想听的不是你会调用模型，而是你能不能把 Agent 当成一个可靠的后端系统，从 0 到 1 设计、开发、测试、上线和持续优化。

## 引言

很多 Agent 项目停在 demo 阶段：能回答几个问题，能调用一两个工具，看起来很聪明。但生产系统的要求完全不同。

生产级 Agent 需要回答一组更工程化的问题：

- 这个问题为什么需要 Agent？
- 哪些步骤由模型判断，哪些步骤由后端确定性执行？
- 工作流如何编排，失败后如何恢复？
- RAG、工具调用、状态机和权限如何组合？
- 如何测试一个输出不固定的系统？
- 如何灰度上线，如何监控，如何持续优化？

这就是 AI Agent 系统的全生命周期开发。

---

## 1. 生命周期总览

一个完整的 Agent 系统可以拆成八个阶段：

```text
需求澄清
  ↓
架构设计
  ↓
工作流搭建
  ↓
RAG 与工具实现
  ↓
测试与评估
  ↓
上线部署
  ↓
监控与诊断
  ↓
持续优化
```

每个阶段都应该产出明确的工程资产：

| 阶段 | 关键问题 | 产出物 |
|:---|:---|:---|
| 需求澄清 | 是否真的需要 Agent | 业务目标、边界、成功指标 |
| 架构设计 | Agent 放在哪一层 | 架构图、模块边界、风险分级 |
| 工作流搭建 | 任务如何流转 | 状态机、DAG、Router、人工确认点 |
| RAG 与工具实现 | Agent 如何获取事实和行动能力 | Retriever、Tool Registry、MCP Server |
| 测试与评估 | 如何证明有效 | eval dataset、测试用例、质量指标 |
| 上线部署 | 如何安全发布 | 灰度策略、限流、熔断、回滚方案 |
| 监控与诊断 | 出错后怎么定位 | trace、metrics、日志、失败归因 |
| 持续优化 | 如何越跑越好 | 回归集、prompt 版本、检索优化、工具 schema 优化 |

面试中可以把这张表当成回答主线。

---

## 2. 需求澄清：先判断是否需要 Agent

不是所有自动化需求都适合 Agent。Agent 适合处理开放性、多步骤、信息不完整的任务；不适合替代强一致、强规则、强审计的核心交易逻辑。

判断问题时可以问五个问题：

```text
1. 任务是否需要理解自然语言或模糊意图？
2. 是否需要跨多个系统检索和整合信息？
3. 是否需要根据中间结果动态选择下一步？
4. 错误是否可以通过 guardrails、人工确认或回滚控制？
5. 成功标准是否可以被评估和持续优化？
```

如果答案大多是“是”，Agent 可能有价值。否则，规则引擎、普通后端服务或固定 workflow 可能更合适。

### 面试表达

```text
我不会一上来就说用 Agent。
我会先判断任务里哪些部分是不确定的，哪些部分是确定的。
不确定的部分交给模型做理解、规划和信息整合；
确定性的查询、权限、计算、状态变更和审计交给后端系统。
```

---

## 3. 架构设计：把 Agent 放进后端系统

生产级 Agent 通常不是一个孤立的模型调用，而是一组后端模块的组合。

```text
Client
  ↓
API Gateway / Auth
  ↓
Task Service
  ↓
Agent Runner
  ├─ State Machine / Workflow Engine
  ├─ RAG Retriever
  ├─ Tool Registry / MCP Client
  ├─ Memory / Session Store
  └─ Guardrails
  ↓
Trace / Metrics / Evals
```

### 核心模块

**Task Service**

负责任务创建、任务状态、幂等键、用户权限和结果查询。不要把关键任务状态只放在 Agent 内存里。

**Agent Runner**

负责调用模型、组织上下文、解释模型输出、执行工具调用和控制最大步数。

**Workflow / State Machine**

负责生命周期控制。例如告警处理可以是：

```text
NEW → ANALYZING → WAITING_CONFIRM → AUTO_RESOLVING → RESOLVED
                         ↓
                    ESCALATED
```

**RAG Retriever**

负责从知识库、runbook、历史工单和文档中检索事实，并保留来源。

**Tool Registry / MCP Client**

负责把后端能力包装成可调用工具，包括 schema、权限、超时、重试、风险等级和审计。

**Guardrails**

负责输入过滤、权限拦截、输出校验、高风险动作审批和敏感信息脱敏。

**Trace / Metrics / Evals**

负责记录 Agent 每一步行为，并把失败样本沉淀为回归测试。

---

## 4. 工作流搭建：用确定性框架约束不确定性

很多 Agent 系统失败，是因为把所有事情都交给一个自由循环的 Agent。更稳的做法是：用 workflow 控制主路径，把模型放在需要判断的位置。

### 常见工作流模式

| 模式 | 适用场景 | 示例 |
|:---|:---|:---|
| Router | 意图分流清晰 | FAQ、订单查询、工单创建 |
| Chain | 步骤固定 | 文档解析、摘要、格式化输出 |
| State Machine | 生命周期明确 | 告警处理、审批流、故障诊断 |
| DAG | 子任务有依赖 | 多来源分析、报告生成 |
| Coordinator + Worker | 多 Agent 协作 | 指标分析、日志分析、变更分析并行 |

### 一个告警诊断工作流

```text
Alert Received
  ↓
Normalize Alert
  ↓
Retrieve Runbook
  ↓
Query Metrics + Logs + Recent Deployments
  ↓
Diagnosis Agent
  ↓
Risk Decision
  ├─ Low Risk: Auto Suggest
  ├─ Medium Risk: Human Confirm
  └─ High Risk: Escalate
  ↓
Report + Trace + Eval Candidate
```

这里的关键点是：模型不负责随意决定系统状态，状态流转由后端控制。

---

## 5. RAG 与工具实现

Agent 的能力来自两类外部系统：RAG 提供事实，工具提供行动。

### RAG 设计要点

- 文档 ingest 时保留 metadata，例如服务名、环境、版本、权限、更新时间；
- 检索前做权限过滤，避免越权文档进入上下文；
- 使用 hybrid search 和 rerank 提升召回质量；
- 输出必须带 citation，无法找到依据时允许拒答；
- 把检索失败样本加入 eval dataset。

### 工具设计要点

工具不是简单暴露 API。生产工具应该包含：

```text
name: query_metrics
description: 查询指定服务在指定时间窗口内的指标
schema: 参数类型、必填字段、枚举值
risk_level: read_only
timeout: 3s
retry: 2
permission: sre:read_metrics
audit: enabled
```

工具按风险分级：

| 风险等级 | 示例 | 执行策略 |
|:---|:---|:---|
| 只读 | 查指标、查日志、查订单状态 | 可自动执行 |
| 中风险 | 创建工单、发送通知、更新备注 | 自动执行但必须审计 |
| 高风险 | 退款、回滚、重启服务、修改生产配置 | 只生成建议，需要人工确认 |
| 禁止 | 删除生产数据、绕过权限 | 不暴露给 Agent |

---

## 6. 测试与评估

Agent 系统既需要传统测试，也需要 Agent Evals。

### 传统测试

- Tool 单元测试：参数校验、权限校验、错误返回；
- Workflow 测试：状态流转、失败重试、人工确认；
- API 测试：鉴权、幂等、超时、结果查询；
- 集成测试：Agent Runner 与工具、RAG、状态机联调。

### Agent Evals

生产级 Agent 至少要分层评估：

| 层级 | 指标 |
|:---|:---|
| RAG | recall@k、MRR、citation accuracy |
| 工具调用 | tool selection accuracy、argument accuracy |
| 任务完成 | task success rate、escalation correctness |
| 安全 | unsafe action rate、permission violation rate |
| 体验 | 用户采纳率、人工升级率、平均处理时长 |

### 发布前检查

```text
1. 单元测试通过
2. 工具权限测试通过
3. 离线 eval 不低于基线版本
4. 高风险 case 全部正确拦截
5. trace 字段完整
6. 回滚开关可用
```

---

## 7. 上线部署

Agent 上线不要一次性全量发布。更稳的路径是：

```text
Shadow Mode → Internal Beta → Small Traffic → Gradual Rollout → Full Release
```

### Shadow Mode

Agent 只观察真实请求，不影响用户结果。用于收集 trace、验证工具路径和发现未知失败模式。

### Internal Beta

只开放给内部用户或值班工程师，允许人工反馈和快速修复。

### Small Traffic

按用户、团队、场景或风险等级小流量放开。例如只处理低风险告警，只允许只读工具自动执行。

### Gradual Rollout

逐步扩大范围，同时观察质量、安全、成本和延迟指标。

### 回滚策略

至少准备三类开关：

- model rollback：切回旧模型；
- prompt rollback：切回旧 prompt 版本；
- feature kill switch：关闭自动执行，仅保留建议模式。

---

## 8. 监控与诊断

上线后不能只看接口错误率。Agent 的问题经常表现为“没有报错，但做错了”。

### 必须监控的指标

**系统指标**

- 请求量；
- P95 / P99 延迟；
- 错误率；
- 超时率；
- 队列积压。

**Agent 指标**

- 平均 LLM 调用次数；
- 平均工具调用次数；
- 平均执行步数；
- early stopping 次数；
- fallback 次数。

**质量指标**

- task success rate；
- correct escalation rate；
- citation accuracy；
- tool selection accuracy；
- unsafe action rate。

**成本指标**

- token per task；
- cost per task；
- cache hit rate；
- 高成本请求占比。

### Trace 字段

```json
{
  "trace_id": "trace_001",
  "task_id": "task_001",
  "user_id": "u123",
  "workflow_state": "ANALYZING",
  "model": "gpt-4.1",
  "prompt_version": "dod_agent_v3",
  "retrieved_docs": [],
  "tool_calls": [],
  "risk_level": "medium",
  "final_decision": "WAITING_CONFIRM",
  "failure_category": "none"
}
```

有了 trace，失败复盘才不会停留在“模型不稳定”。

---

## 9. 持续优化

Agent 系统的优化应该围绕 trace 和 eval，而不是凭感觉调 prompt。

### 优化闭环

```text
线上请求
  ↓
Trace 采样
  ↓
失败归因
  ↓
加入 eval dataset
  ↓
修复 prompt / retriever / tool schema / workflow
  ↓
回归测试
  ↓
灰度发布
```

### 常见优化方向

| 问题 | 优化方式 |
|:---|:---|
| 检索不到正确文档 | 调整 chunk、metadata、rerank、query rewrite |
| 工具选错 | 改工具命名、description、schema 和 few-shot |
| 参数错误 | 收紧 schema，增加默认值和错误提示 |
| 执行太慢 | 并行工具调用、缓存、异步任务、模型分层 |
| 成本太高 | prompt 压缩、上下文裁剪、缓存、early stopping |
| 高风险输出 | 增加 guardrail、审批流和拒答策略 |

---

## 10. 面试表达模板

如果面试官问：“你如何负责 AI Agent 系统的全生命周期开发？”

可以这样回答：

```text
我会把 Agent 系统当成一个生产级后端系统来设计。

第一步是需求澄清，判断哪些环节真的需要 Agent，哪些环节应该用确定性后端逻辑。

第二步是架构设计，拆成 API、Task Service、Agent Runner、RAG、Tool Registry、Workflow、Guardrails 和 Trace。

第三步是工作流搭建，用 Router、状态机或 DAG 控制任务生命周期，把模型限制在意图理解、规划和诊断这些开放性环节。

第四步是实现 RAG 和工具调用。RAG 负责补充事实，工具负责连接外部系统。所有工具都要有 schema、权限、超时、重试、风险等级和审计。

第五步是测试和评估。除了传统单元测试、集成测试，还要做 RAG eval、tool eval、task eval 和 safety eval。

第六步是上线部署。先 shadow mode，再小流量灰度，配合限流、熔断、回滚和人工兜底。

最后是持续优化。上线后基于 trace 做失败归因，把失败样本加入 eval dataset，持续优化 prompt、retriever、工具 schema 和 workflow。
```

这个回答的重点是让面试官听到：你不是只会调模型，而是能负责一个 Agent 系统从 0 到 1 到稳定运行的完整闭环。

---

## 本章小结

AI Agent 系统的全生命周期开发，本质是把不确定的模型能力放进确定的工程框架里。

你需要同时具备四类能力：

1. 架构能力：能拆模块、定边界、做取舍；
2. 后端能力：能处理状态、队列、幂等、权限、部署和回滚；
3. Agent 能力：能设计 RAG、工具调用、工作流和多 Agent 协作；
4. 质量能力：能用 eval、trace 和监控持续证明系统有效。

下一章的 DoD Agent 案例，就是这套生命周期方法的一次完整落地。
