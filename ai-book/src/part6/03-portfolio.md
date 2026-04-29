# 第24章 项目复盘、作品集模板与面试表达

> 面试表达不是把术语背出来，而是把你的工程判断、系统边界、失败复盘和可验证结果讲清楚。

## 引言

前面的章节已经覆盖了 Prompt、Context、Harness、Agent Runtime、RAG、Memory、Evals、Guardrails、Observability 和完整案例。

本章不再引入新的架构主线，而是把这些能力转译成三类面试资产：

- 一个能讲清楚的项目；
- 一套系统设计回答框架；
- 一份能体现工程深度的作品集。

真正有说服力的 Agent 项目，不是“我用了 LangChain / MCP / RAG”，而是能讲清：

- 为什么这个问题适合 Agent；
- 哪些环节由模型判断，哪些环节由确定性系统控制；
- Prompt、Context、Harness 分别如何设计；
- 如何评估质量；
- 如何限制风险；
- 出错后如何定位和改进；
- 最后产生了什么可验证结果。

---

## 23.1 一页项目介绍

一页项目介绍适合放在简历、作品集首页或面试开场。

```markdown
# 项目名称：Production Alert Diagnosis Agent

## 背景

值班工程师每天需要处理大量生产告警。传统告警只告诉“哪个指标异常”，但根因分析需要同时查看 metrics、logs、runbook、部署记录和历史工单，处理成本高且依赖个人经验。

## 目标

构建一个告警诊断 Agent，帮助值班工程师快速汇总证据、判断可能原因、生成处理建议，并在高风险动作前强制人工确认。

## 架构

- Prompt Engineering：把诊断任务拆成风险分级、证据提取和建议动作输出协议。
- Context Engineering：按服务、环境、时间窗口检索 runbook、日志摘要、指标和历史案例。
- Harness Engineering：通过状态机、Tool Registry、Guardrails、Evals 和 Trace 控制执行过程。

## 我的贡献

- 设计 Agent 工作流和状态机；
- 设计工具 schema、权限、超时和审计；
- 构建 RAG 检索与 citation 输出；
- 建立离线 eval dataset 和线上 trace 分析；
- 梳理失败模式并沉淀为回归测试。

## 结果

- 平均告警初步诊断时间从 15 分钟降到 5 分钟；
- 低风险告警自动生成建议，高风险动作全部进入人工确认；
- 失败样本进入 eval dataset，支持 prompt、retriever 和 tool schema 的持续迭代。
```

注意：结果指标必须真实。如果没有线上数据，可以写“离线评估结果”或“预期验证方式”，不要编造上线效果。

---

## 23.2 技术决策表

面试官关心的不是你用了什么，而是为什么这么选。

| 决策点 | 选择 | 为什么 | 替代方案 | 风险与补救 |
|:---|:---|:---|:---|:---|
| 是否使用 Agent | 使用混合 Agent | 任务需要跨系统检索、动态判断和证据整合 | 固定 workflow / 规则引擎 | 高风险动作不自动执行 |
| 工作流 | 状态机 + ReACT 子步骤 | 主流程可控，诊断步骤保留灵活性 | 纯 ReACT | 状态机控制停止条件 |
| 知识系统 | RAG + metadata filter | runbook 和历史工单需要按服务、权限、时间过滤 | 纯向量检索 | 增加 hybrid search 和 rerank |
| 工具系统 | Tool Registry + MCP | 工具 schema、权限、审计统一管理 | 直接调 API | 每个工具定义 risk_level |
| 质量评估 | 分层 eval | 分别定位检索、工具、任务、安全问题 | 只看用户反馈 | 失败样本进入回归集 |
| 安全 | 四层 Guardrails | 输入、上下文、工具、输出风险不同 | 只写 Prompt 约束 | 高风险动作人工确认 |

这张表可以显著提升项目表达的可信度，因为它展示的是架构判断，而不是工具堆砌。

---

## 23.3 面试表达框架

回答 Agent 项目问题时，可以使用六段式结构。

### 1. 先讲问题，而不是先讲模型

```text
这个项目要解决的是生产告警诊断慢的问题。
单条告警本身信息不足，工程师需要跨 metrics、logs、runbook、部署记录和历史工单整合证据。
所以核心不是做一个聊天机器人，而是把诊断流程中的信息整合和初步判断自动化。
```

### 2. 说明为什么需要 Agent

```text
我没有直接选择固定规则引擎，因为告警原因组合很多，且需要根据中间证据动态决定下一步查什么。
但我也没有让模型自由循环，而是用状态机控制主流程，把模型限制在诊断、总结和风险判断这些开放环节。
```

### 3. 拆 Prompt、Context、Harness

```text
Prompt 层面，我把诊断输出定义成结构化协议，包括 summary、risk_level、evidence、recommended_actions 和 need_human_confirm。

Context 层面，我根据 service、env、time window 和 permission 检索 runbook、日志摘要、指标和历史案例，并保留 citation。

Harness 层面，我用 Tool Registry 管理工具 schema、权限、超时和风险等级，用 Guardrails 拦截高风险动作，用 Evals 和 Trace 做质量闭环。
```

### 4. 讲评估

```text
我把评估分成四层：
第一层是检索质量，看 recall@k、MRR 和 citation accuracy；
第二层是工具调用，看 tool selection 和 argument accuracy；
第三层是任务完成，看诊断是否有证据支持、风险分级是否正确；
第四层是安全，看 unsafe action rate 和 permission violation rate。
```

### 5. 讲失败和改进

```text
早期失败主要有三类：
一是检索到语义相近但服务不匹配的 runbook；
二是工具参数缺少时间窗口；
三是模型给出重启服务建议但没有标记人工确认。

对应改进是：增加 service metadata filter，收紧工具 schema，并把 high risk 动作的人审要求下沉到 guardrail，而不是只写在 prompt 里。
```

### 6. 讲结果和边界

```text
这个系统适合做辅助诊断和建议生成，不适合完全替代值班工程师。
低风险只读分析可以自动化，高风险生产变更必须人审。
```

这个结尾很重要。能讲清边界，比一味强调“自动化”更像成熟工程师。

---

## 23.4 作品集结构

一个完整作品集可以包含：

```text
portfolio/
├── README.md                # 一页项目介绍
├── architecture.md          # 架构图和模块说明
├── prompt-contract.md       # Prompt / output contract
├── context-design.md        # RAG、metadata、memory、context budget
├── tool-registry.md         # 工具 schema、权限、风险等级
├── evals.md                 # eval dataset、指标、样例
├── guardrails.md            # 风险分级和拦截策略
├── observability.md         # trace 字段、指标、失败分类
└── postmortem.md            # 失败复盘和迭代记录
```

作品集不一定要代码很多，但一定要展示工程闭环。

如果只能准备三份材料，优先准备：

1. `architecture.md`：证明你能做系统设计；
2. `evals.md`：证明你知道如何判断 Agent 是否变好；
3. `postmortem.md`：证明你真的理解失败模式，而不是只会讲成功案例。

---

## 23.5 失败复盘模板

失败复盘是 Agent 面试里最能拉开差距的部分。

```markdown
# 失败复盘：RAG 找错 runbook

## 现象

用户询问 order-service CPU 高，Agent 引用了 payment-service 的 CPU runbook。

## 影响

诊断建议方向错误，但未触发高风险动作。

## 根因

检索只使用向量相似度，没有按 service metadata 做过滤。
两个 runbook 都包含 “CPU high after deploy” 关键词，语义相似度很高。

## 修复

1. 文档 ingest 时增加 service、team、env、doc_type metadata；
2. 检索前根据 alert.service 做 metadata filter；
3. rerank 前保留更多候选；
4. eval dataset 增加跨服务相似 runbook 的负例。

## 防复发

- 新增 eval case：rag_wrong_service_001；
- 检索 trace 增加 selected_metadata；
- citation checker 校验引用文档 service 是否和告警服务一致。
```

这样的复盘比“我优化了 RAG”更有信息量。

---

## 本章小结

面试和作品集的本质，是把工程能力转译成别人能快速理解和验证的材料。

一个强 Agent 项目表达应该覆盖：

1. 问题背景；
2. 为什么需要 Agent；
3. Prompt、Context、Harness 如何设计；
4. 工具、RAG、Memory、Workflow 如何组合；
5. Evals、Guardrails、Observability 如何闭环；
6. 失败模式如何复盘和修复；
7. 系统边界在哪里。

如果能把这些讲清楚，你展示的就不是“会用 AI 工具”，而是具备 AI 架构设计和生产落地能力。
