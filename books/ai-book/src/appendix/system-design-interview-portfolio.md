# 附录D 系统设计面试题与作品集模板

> 面试材料不是正文主线，但它能帮助读者把 Agent 工程能力表达出来。

## D.1 Agent 系统设计回答框架

回答 Agent 系统设计题时，不要一上来讲模型。先判断问题是否真的需要 Agent。

```text
1. 问题是否需要自然语言理解？
2. 是否需要多步骤推理？
3. 是否需要整合多个系统或数据源？
4. 是否允许概率性输出？
5. 是否有人工兜底和评估机制？
```

推荐回答结构：

```text
需求澄清
  → 是否需要 Agent
  → 核心架构
  → Prompt / Context / Tools / Memory
  → Guardrails
  → Evals
  → Observability
  → 失败模式和权衡
```

---

## D.2 题目一：企业知识库问答 Agent

### 需求

为公司内部文档、工单、聊天记录构建一个问答系统，支持员工用自然语言提问，并返回带引用的答案。

### 关键澄清

- 数据源有哪些？
- 权限是否要和源系统一致？
- 是否需要外部网页搜索？
- 答案是否必须引用来源？
- 对延迟和准确率有什么要求？

### 架构

```text
User
  │
  ▼
Query Understanding
  │
  ▼
Hybrid Retrieval
  ├─ Keyword Search
  ├─ Vector Search
  ├─ Metadata Filter
  └─ Permission Filter
  │
  ▼
Evidence Package
  │
  ▼
Answer Generator
  │
  ▼
Citation + Feedback + Trace
```

### 重点设计

- 权限过滤必须在生成前完成；
- 检索使用 hybrid search；
- Evidence Package 需要包含来源、时间、权限校验；
- 证据不足时拒答；
- 线上监控引用支持率和越权率。

---

## D.3 题目二：客服工单处理 Agent

### 需求

用户提交问题后，Agent 尝试基于知识库回答；无法解决时创建工单并路由到正确团队。

### 架构

```text
User Message
  │
  ▼
Intent Classifier
  ├─ FAQ Answer
  ├─ Need Clarification
  ├─ Create Ticket
  └─ Human Escalation
```

### 工具

- `search_kb`
- `get_ticket_status`
- `create_ticket`
- `route_ticket`
- `notify_support_team`

### Guardrails

- 涉及退款、支付、账号安全时转人工；
- 创建工单需要结构化字段；
- 禁止模型承诺 SLA 之外的处理时间；
- 输出必须避免泄露其他用户信息。

---

## D.4 题目三：代码审查 Agent

### 需求

为 PR 自动生成代码审查意见，覆盖 bug、安全、性能、可维护性和测试缺口。

### 架构

```text
PR Diff
  │
  ▼
Context Builder
  ├─ Changed Files
  ├─ Related Tests
  ├─ Ownership Rules
  └─ Project Guidelines
  │
  ▼
Review Agents
  ├─ Bug Reviewer
  ├─ Security Reviewer
  ├─ Performance Reviewer
  └─ Test Reviewer
  │
  ▼
Aggregator
```

### 设计重点

- 只评论可定位的问题；
- 输出必须包含文件和行号；
- 不要把风格偏好当 bug；
- 高置信度问题优先；
- 低置信度建议单独标记。

---

## D.5 题目四：生产告警诊断 Agent

### 需求

收到生产告警后，Agent 自动查询指标、日志、部署记录和历史案例，给出诊断建议。

### 架构

```text
Alert
  │
  ▼
State Machine
  ├─ Gather Alert Context
  ├─ Query Metrics
  ├─ Search Logs
  ├─ Check Deployments
  ├─ Retrieve Runbooks
  └─ Generate Diagnosis
```

### 重点设计

- 只读诊断自动执行；
- 重启、回滚、扩容必须人工审批；
- 结论必须标注证据；
- 记录完整 trace；
- 和人工处理结果对齐做 eval。

---

## D.6 一页项目介绍模板

```markdown
# 项目名称：Production Alert Diagnosis Agent

## 背景
值班工程师每天需要处理大量生产告警，诊断依赖指标、日志、部署记录和历史案例，响应慢且新人上手困难。

## 目标
- 自动收集诊断证据；
- 输出带引用的根因假设；
- 高风险动作转人工审批；
- 降低 MTTR。

## 架构
- Agent Runtime：状态机 + ReACT；
- Tools：Prometheus、Loki、Kubernetes、Runbook Search；
- Guardrails：工具风险分级和审批；
- Observability：Trace、成本、成功率；
- Evals：历史告警回放。

## 我的贡献
- 设计工具注册表和风险分级；
- 实现告警诊断工作流；
- 建立离线评估集；
- 接入 trace 和指标监控。

## 结果
- 自动诊断覆盖率：xx%；
- 诊断准确率：xx%；
- MTTR 降低：xx%；
- 高风险动作零自动执行。
```

---

## D.7 面试表达框架

讲 Agent 项目时，建议按这个顺序：

1. **先讲问题**：为什么传统规则或搜索不够。
2. **再讲 Agent 必要性**：哪里需要理解、推理、工具、反馈。
3. **拆架构**：Prompt、Context、Tools、Memory、Workflow。
4. **讲治理**：Evals、Guardrails、Observability。
5. **讲失败**：遇到过什么失败，怎么定位和修复。
6. **讲边界**：什么不让 Agent 自动做。
7. **讲结果**：量化指标和业务影响。

---

## D.8 失败复盘模板

```markdown
# 失败复盘：RAG 找错 Runbook

## 现象
Agent 在诊断 CPU 告警时引用了数据库连接池 Runbook，导致建议方向错误。

## 影响
值班工程师多花 15 分钟排查错误方向。

## 根因
- 检索只依赖向量相似度；
- Runbook 缺少服务和告警类型 metadata；
- reranker 没有考虑指标类型。

## 修复
- 为 Runbook 增加 service、metric、severity metadata；
- 检索时加入 metadata filter；
- 增加历史告警回放 eval。

## 防复发
- 新增 eval case：CPU、DB、网络三类告警；
- 监控引用支持率；
- 低置信度时要求输出多个假设。
```

---

## D.9 面试前检查清单

### 概念

- [ ] 能解释 Prompt、Context、Harness 的区别；
- [ ] 能解释 RAG、Memory、Tool Calling 的边界；
- [ ] 能说清楚 Agent 为什么需要 eval；
- [ ] 能说清楚 high-risk tool 为什么要审批。

### 项目

- [ ] 有一页项目介绍；
- [ ] 有架构图；
- [ ] 有 trace 示例；
- [ ] 有 eval dataset 示例；
- [ ] 有失败复盘；
- [ ] 有量化指标。

### 表达

- [ ] 不把所有问题都归因于“模型不够强”；
- [ ] 不把 demo 说成生产系统；
- [ ] 能讲 trade-off；
- [ ] 能讲边界和人工兜底；
- [ ] 能说明自己具体贡献。

---

## 小结

面试和作品集的目标不是展示“用了 AI”，而是展示你具备生产级 Agent 工程意识：

- 能判断是否需要 Agent；
- 能设计上下文、工具和工作流；
- 能治理风险；
- 能评估质量；
- 能复盘失败；
- 能把复杂系统讲清楚。
