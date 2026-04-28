# 第23章 Agent 系统设计面试题库

> 系统设计题考的不是“架构图画得复杂”，而是你能否在不确定性中划清边界。

## 引言

Agent 应用工程师面试中，系统设计题是最能拉开差距的环节。很多候选人会直接说“用 RAG + Agent + MCP”，但面试官真正想听的是：为什么需要 Agent，工具怎么设计，怎么评估效果，失败了怎么兜底。

本章提供一组高频题和回答框架。

---

## 17.1 通用回答框架

所有 Agent 系统设计题都可以按以下顺序回答：

```text
1. 需求澄清
2. 成功指标
3. Agent 选型判断
4. 总体架构
5. RAG / Memory 设计
6. Tool Calling / MCP 设计
7. Workflow / State Machine
8. Guardrails
9. Evals
10. Observability
11. 成本、延迟与扩展性
12. 失败模式与兜底
```

回答时要反复强调：不确定性放在 Agent，确定性放在后端系统。

---

## 17.2 题目一：企业知识库问答Agent

### 需求

设计一个企业内部知识库问答 Agent，支持员工查询制度、项目文档和技术 runbook。

### 关键澄清

- 文档来源有哪些？
- 是否有权限隔离？
- 是否要求引用来源？
- 是否支持多轮追问？
- 是否允许回答“我不知道”？

### 架构

```text
User
  ↓
Auth / Permission
  ↓
Query Rewriter
  ↓
Hybrid Retriever
  ↓
Reranker
  ↓
Answer Generator
  ↓
Citation Checker
  ↓
Response
```

### 重点设计

- 文档 ingest 时保留部门、项目、密级等 metadata；
- 检索前按用户权限过滤；
- 使用 hybrid search 提升召回；
- 使用 reranker 提升排序；
- 输出必须带 citation；
- 上下文不足时明确拒答。

### Evals

- retrieval recall@5；
- citation accuracy；
- answer faithfulness；
- permission violation rate；
- no-answer correctness。

### 常见追问

问题：如果用户问的问题没有文档怎么办？

回答：

```text
我会让 Agent 明确说明知识库没有足够依据，而不是编造。
同时记录 no-answer trace，用于判断是知识库缺口还是检索失败。
如果同类问题频繁出现，就进入文档补齐流程。
```

---

## 17.3 题目二：客服工单处理Agent

### 需求

设计一个客服 Agent，自动回答用户问题，必要时创建或升级工单。

### 架构

```text
User Message
  ↓
Intent Router
  ├─ FAQ RAG
  ├─ Order Status Tool
  ├─ Refund Policy RAG + Tool
  └─ Human Handoff
```

### 工具

- `get_order_status`：只读；
- `get_refund_status`：只读；
- `create_ticket`：中风险；
- `escalate_ticket`：中风险；
- `issue_refund`：高风险，默认不允许 Agent 直接执行。

### Guardrails

- 用户只能查询自己的订单；
- 退款政策必须引用来源；
- 高风险操作转人工；
- 输出中隐藏敏感字段。

### Evals

- first contact resolution；
- escalation correctness；
- policy citation accuracy；
- tool selection accuracy；
- user satisfaction。

---

## 17.4 题目三：代码审查Agent

### 需求

设计一个 Agent，自动审查 Pull Request，发现 bug、风险和测试缺口。

### 架构

```text
PR Webhook
  ↓
Diff Parser
  ↓
Context Collector
  ↓
Review Agent
  ↓
Static Tools
  ↓
Finding Ranker
  ↓
PR Comment
```

### 工具

- 读取 diff；
- 搜索相关代码；
- 运行单元测试；
- 运行 lint；
- 查询历史 bug；
- 创建 review comment。

### 设计重点

- 不直接修改代码，默认只评论；
- findings 必须包含文件、行号、风险和建议；
- 区分 blocker、warning 和 nit；
- 低置信度建议不发评论，进入 summary；
- 对安全相关文件提高审查强度。

### Evals

- bug detection precision；
- false positive rate；
- actionable comment rate；
- developer acceptance rate；
- missed critical issue rate。

---

## 17.5 题目四：生产告警诊断Agent

### 需求

设计一个 Agent，帮助值班工程师诊断生产告警。

### 架构

```text
Alertmanager
  ↓
Alert Normalizer
  ↓
State Machine
  ↓
Diagnosis Agent
  ├─ Metrics Tool
  ├─ Logs Tool
  ├─ Deployment Tool
  ├─ Runbook RAG
  └─ Incident Tool
  ↓
Decision Engine
  ↓
Notify / Escalate
```

### 设计重点

- 状态机控制生命周期；
- ReACT 只负责诊断步骤；
- 高风险操作只给建议；
- 每次诊断生成证据链；
- 相似历史案例进入上下文。

### Evals

- diagnosis accuracy；
- MTTR reduction；
- correct escalation rate；
- tool path completeness；
- unsafe action rate。

---

## 17.6 题目五：销售线索分析Agent

### 需求

设计一个 Agent，帮助销售团队分析潜在客户，生成跟进建议。

### 架构

```text
Lead Input
  ↓
Company Enrichment Tools
  ↓
CRM History Retriever
  ↓
Scoring Agent
  ↓
Recommendation Generator
  ↓
CRM Update / Task Creation
```

### 工具

- 查询 CRM；
- 查询公司信息；
- 查询历史沟通；
- 创建 follow-up task；
- 更新 lead score。

### Guardrails

- 不编造客户事实；
- 对外发送内容必须人工确认；
- CRM 写入需要审计；
- 遵守隐私和合规要求。

### Evals

- lead score correlation；
- recommendation acceptance rate；
- CRM data accuracy；
- hallucinated fact rate。

---

## 17.7 题目六：数据分析Agent

### 需求

设计一个 Agent，让业务人员用自然语言查询数据并生成分析报告。

### 架构

```text
Natural Language Query
  ↓
Intent + Permission Check
  ↓
Semantic Layer
  ↓
SQL Generator
  ↓
SQL Validator
  ↓
Query Executor
  ↓
Chart / Report Generator
```

### 设计重点

- 不让模型直接访问任意表；
- 建 semantic layer；
- SQL 执行前做只读校验；
- 限制扫描量；
- 查询结果做隐私脱敏；
- 复杂分析输出方法说明。

### Evals

- SQL correctness；
- permission violation rate；
- chart correctness；
- insight usefulness；
- query cost。

---

## 17.8 题目七：多Agent协作任务调度系统

### 需求

设计一个多 Agent 协作系统，用于处理复杂业务任务。例如用户提交“分析一次线上故障并生成复盘报告”，系统需要自动检索指标、分析日志、查找变更、引用 runbook，最后生成诊断结论和改进建议。

### 关键澄清

- 任务是否需要实时完成，还是可以异步执行？
- 子任务之间是否存在依赖？
- Agent 是否允许执行有副作用的修复动作？
- 结果是否需要人工审核？
- 失败后是否要支持断点恢复？

### 架构

```text
Task API
  ↓
Task Store / State Machine
  ↓
Coordinator Agent
  ├─ Metrics Analyst Agent
  ├─ Log Analyst Agent
  ├─ Change Investigator Agent
  ├─ Runbook RAG Agent
  └─ Report Writer Agent
  ↓
Result Aggregator
  ↓
Human Review / Publish
```

### 任务调度设计

- Task Store 保存任务状态、输入、子任务、依赖关系和执行结果；
- Coordinator 只负责任务拆解、依赖编排和结果汇总；
- Worker Agent 只处理单一职责，避免一个 Agent 背负过长上下文；
- 子任务可以并行执行，依赖任务按 DAG 顺序执行；
- 每个子任务设置 timeout、retry、max step 和 fallback；
- 最终报告必须带证据链，不允许只给结论。

### 状态机

```text
CREATED
  → PLANNING
  → RUNNING
  → WAITING_REVIEW
  → COMPLETED

RUNNING
  → PARTIAL_FAILED
  → RETRYING
  → ESCALATED
```

### 工具与权限

- 指标、日志、变更查询属于只读工具，可以自动执行；
- 创建工单、通知负责人属于中风险工具，需要审计；
- 重启服务、回滚发布属于高风险动作，只能生成建议，不能自动执行；
- 工具权限必须由后端系统校验，不能只依赖 prompt。

### Evals

- plan correctness：任务拆解是否合理；
- dependency correctness：依赖顺序是否正确；
- tool path completeness：是否查询了必要证据；
- report faithfulness：报告是否忠实于证据；
- escalation correctness：高风险场景是否正确升级；
- task success rate：整体任务完成率。

### 常见追问

问题：为什么不让一个 ReACT Agent 自己循环完成所有事情？

回答：

```text
自由循环的 ReACT Agent 灵活，但生产系统里很难控制执行边界。
我会把任务生命周期交给状态机，把任务拆解交给 Coordinator，
把具体分析交给职责单一的 Worker Agent。
这样既保留推理能力，又能控制超时、重试、权限、审计和断点恢复。
```

问题：多个 Agent 的结果冲突怎么办？

回答：

```text
我不会让 Aggregator 直接“投票”得出结论。
每个 Agent 必须输出证据、置信度和不确定性。
如果指标和日志结论冲突，就进入二次验证或人工审核，而不是强行生成确定结论。
```

---

## 17.9 面试中的加分表达

### 1. 先判断是否需要 Agent

```text
如果流程固定、规则明确、准确率必须 100%，我不会优先用 Agent。
我会把 Agent 用在意图理解、多步骤诊断、复杂信息整合这些开放性环节。
```

### 2. 把工具权限说清楚

```text
只读工具可以自动执行，有副作用工具要风险分级，高风险工具需要人工确认。
```

### 3. 把评估说清楚

```text
我会分层评估：检索、工具调用、任务完成和多轮会话。
上线后把失败 trace 加入回归集。
```

### 4. 把失败模式说清楚

```text
我会重点防止三类问题：无依据回答、错误工具调用和越权操作。
对应措施是 citation、tool guardrails 和权限校验。
```

---

## 本章小结

Agent 系统设计题的核心不是堆技术名词，而是展示工程判断。

每道题都要回答：

- 为什么需要 Agent；
- 哪些部分不用 Agent；
- 工具边界在哪里；
- 安全边界在哪里；
- 如何证明系统有效；
- 失败后怎么复盘。

下一章我们整理 Agent 的常见失败模式、Debug 方法和作品集模板，帮助你把项目讲得更像真实工程经验。
