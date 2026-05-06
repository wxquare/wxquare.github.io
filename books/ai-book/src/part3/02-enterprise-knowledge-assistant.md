# 第13章 企业知识助手系统解析：RAG、搜索、权限与知识治理

> 企业知识助手不是“把文档丢进向量库”，而是一个带权限、证据、治理和反馈闭环的知识操作系统。

## 引言

RAG 是 Agent 系统里最常见、也最容易被低估的能力。很多团队以为企业知识助手的核心是 Embedding 和向量检索，真正上线后才发现主要难点在别处：

- 数据源很多，格式和权限各不相同；
- 搜索结果必须严格遵守源系统权限；
- 文档过期、重复、冲突非常普遍；
- 用户需要答案，也需要证据和引用；
- 组织知识会持续变化，索引必须持续更新；
- 回答错误会影响业务决策，必须可追责。

Glean、Perplexity Enterprise、Microsoft 365 Copilot 类系统的成熟之处，不是单点模型能力，而是把搜索、RAG、权限、引用、反馈和治理组合成完整系统。

本章从系统设计角度拆解企业知识助手。

---

## 13.1 系统定位：从搜索框到知识工作台

企业知识助手解决的不是“用户找不到文档”这么简单，而是三个层次的问题。

| 层次 | 用户问题 | 系统能力 |
|:---|:---|:---|
| Search | 这份文档在哪里？ | 跨系统检索、排序、权限过滤 |
| Answer | 这个问题的答案是什么？ | RAG、引用、摘要、冲突处理 |
| Action | 我下一步该怎么做？ | 工具调用、流程建议、任务创建 |

传统企业搜索通常停留在 Search 层；成熟知识助手会进入 Answer 和 Action 层。

### 典型用户工作流

```text
用户：为什么日本站订单取消率上周升高？

系统：
1. 识别问题类型：业务分析 + 时间范围 + 地区
2. 检索相关数据源：BI 报表、实验记录、客服工单、发布记录
3. 权限过滤：只返回用户有权查看的内容
4. 证据聚合：订单取消率、支付失败、物流异常、最近发布
5. 生成答案：给出结论、证据和不确定性
6. 建议动作：创建分析任务或订阅后续指标
```

这类系统的核心不是一次问答，而是把组织知识转化为可操作判断。

---

## 13.2 端到端架构

```text
┌─────────────────────────────────────────────────────────────┐
│              Enterprise Knowledge Assistant                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  User / Team / App Context                                  │
│      │                                                       │
│      ▼                                                       │
│  Query Understanding                                         │
│  ├─ Intent Classification                                    │
│  ├─ Entity / Time / Scope Extraction                         │
│  └─ Access Context                                           │
│      │                                                       │
│      ▼                                                       │
│  Retrieval Orchestrator                                      │
│  ├─ Keyword Search                                           │
│  ├─ Vector Search                                            │
│  ├─ Graph / Metadata Filter                                  │
│  └─ Freshness / Authority Ranker                             │
│      │                                                       │
│      ▼                                                       │
│  Evidence Builder                                            │
│  ├─ Permission Filtering                                     │
│  ├─ Dedup / Conflict Detection                               │
│  ├─ Snippet Selection                                        │
│  └─ Citation Package                                         │
│      │                                                       │
│      ▼                                                       │
│  Answer Generator                                            │
│  ├─ Grounded Answer                                          │
│  ├─ Uncertainty Statement                                    │
│  ├─ Sources / Citations                                      │
│  └─ Follow-up Actions                                        │
│      │                                                       │
│      ▼                                                       │
│  Feedback / Eval / Governance                                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

架构上最重要的一点是：**权限过滤和证据构建必须发生在答案生成之前**。模型不能看到用户无权访问的内容，也不能在无证据时编造答案。

---

## 13.3 数据摄取：连接器是第一道护城河

成熟企业知识助手通常通过连接器接入各种数据源：

- 文档系统：Google Drive、SharePoint、Confluence、Notion；
- 协作系统：Slack、Teams、邮件；
- 项目系统：Jira、Linear、GitHub；
- 客服和 CRM：Zendesk、Salesforce；
- BI 和数据平台：Tableau、Looker、内部报表；
- 代码和技术文档：GitHub、GitLab、Runbook。

连接器不只是“拉数据”。它至少要做四件事：

1. 拉取内容；
2. 解析结构；
3. 同步权限；
4. 维护增量更新。

### 权限同步比内容同步更重要

企业搜索最怕的不是“搜不到”，而是“搜到了不该看的东西”。

因此连接器需要同步源系统 ACL：

```text
Document
  ├─ content
  ├─ metadata
  ├─ owner
  ├─ source
  ├─ updated_at
  └─ allowed_principals
      ├─ users
      ├─ groups
      └─ domains
```

查询时再结合用户身份做过滤：

```text
candidate_docs
  │
  ▼
permission_filter(user_id, groups)
  │
  ▼
visible_docs
```

这意味着企业知识助手的检索系统必须同时处理“相关性”和“可见性”。

---

## 13.4 检索层：Hybrid Search 是默认选择

单纯向量检索不适合企业知识场景，因为企业问题经常包含精确实体：

- 工单号；
- 项目代号；
- 客户名；
- 服务名；
- 文档标题；
- 错误码；
- 表名和字段名。

成熟系统通常使用 Hybrid Search：

```text
Query
  ├─ BM25 / Keyword Search
  ├─ Dense Vector Search
  ├─ Metadata Filter
  ├─ Graph Expansion
  └─ Reranking
```

### 排序信号

企业知识排序不能只看语义相似度，还要考虑：

| 信号 | 作用 |
|:---|:---|
| 语义相似度 | 找到主题相关内容 |
| 关键词匹配 | 保证精确实体不丢 |
| 更新时间 | 避免旧文档误导 |
| 权威度 | 官方文档高于聊天记录 |
| 用户关系 | 同团队内容更可能相关 |
| 点击和反馈 | 组织内部使用行为 |
| 引用链 | 被其他文档引用的内容更重要 |

### Freshness 与 Authority 的冲突

新文档不一定权威，权威文档也可能过期。系统需要显式处理：

```text
高权威 + 新鲜：优先引用
高权威 + 过期：引用但标记过期风险
低权威 + 新鲜：作为辅助证据
低权威 + 过期：默认降权
```

这比“取 top-k 文档”更接近真实企业搜索。

---

## 13.5 Evidence Package：让答案有证据边界

RAG 的关键中间产物不应该是“拼接后的上下文”，而应该是 Evidence Package。

```json
{
  "question": "为什么日本站订单取消率上周升高？",
  "evidence": [
    {
      "source": "bi://orders/cancellation_dashboard",
      "title": "JP Order Cancellation Dashboard",
      "snippet": "Cancellation rate increased from 2.1% to 4.8% between Apr 20 and Apr 27.",
      "updated_at": "2026-04-28",
      "authority": "official_metric",
      "permission_checked": true
    },
    {
      "source": "jira://PAY-8842",
      "title": "Payment timeout increase in JP",
      "snippet": "Timeout errors increased after deploy payment-router-v3.",
      "updated_at": "2026-04-25",
      "authority": "incident_ticket",
      "permission_checked": true
    }
  ],
  "gaps": [
    "No logistics dashboard access for current user."
  ]
}
```

Evidence Package 有三个价值：

- 让模型只能基于证据回答；
- 让用户能追溯答案来源；
- 让评估系统能检查引用是否支持结论。

---

## 13.6 答案生成：引用、冲突与不确定性

企业知识助手的回答应该避免“全知口吻”。它需要表达三种东西：

1. 结论；
2. 证据；
3. 不确定性。

### 好的答案结构

```text
结论：
日本站订单取消率升高，最可能与 payment-router-v3 发布后的支付超时增加有关。

证据：
1. BI 报表显示取消率从 2.1% 升至 4.8%。
2. 支付工单 PAY-8842 显示 JP 支付超时在同一时间窗口上升。
3. 发布记录显示 payment-router-v3 在异常开始前 30 分钟上线。

不确定性：
我没有访问物流异常报表的权限，因此不能排除物流延迟因素。

建议：
1. 检查 payment-router-v3 的 JP 路由超时配置。
2. 让有权限的同事补充物流异常数据。
```

### 冲突处理

当证据冲突时，不要让模型“平均一下”。应该显式输出：

```text
证据冲突：
- BI 报表显示取消率升高；
- 客服周报说取消率无明显变化；
- 两者统计口径不同：BI 按订单创建时间，客服按投诉时间。
```

这类回答比强行给结论更可信。

---

## 13.7 企业治理：权限、审计和数据边界

成熟系统必须把安全设计放在架构里，而不是放在提示词里。

### 权限模型

```text
User Identity
  ├─ user_id
  ├─ groups
  ├─ department
  ├─ region
  └─ role

Document ACL
  ├─ users
  ├─ groups
  ├─ domains
  └─ source-specific rules
```

系统必须保证：

- 检索阶段不返回无权文档；
- 生成阶段不注入无权内容；
- 引用链接打开时仍由源系统校验权限；
- 审计日志能追踪谁问了什么、引用了什么。

### 数据边界

企业知识助手通常需要明确回答：

- 用户数据是否用于训练？
- 文件上传保留多久？
- 哪些连接器有写权限？
- 是否支持数据驻留要求？
- 是否支持 SSO、SCIM、审计导出？

这些不是合规团队才关心的问题，它们直接决定系统能否在企业内部上线。

---

## 13.8 反馈闭环与质量评估

企业知识助手的质量不能只看“用户点赞”。需要分层评估。

| 层次 | 指标 | 示例 |
|:---|:---|:---|
| 检索 | Recall@k、MRR、权限误召回 | 正确文档是否进入候选集 |
| 证据 | 引用支持率、冲突识别率 | 结论是否被证据支持 |
| 生成 | 正确性、完整性、拒答率 | 是否在证据不足时停止 |
| 安全 | 越权率、敏感信息泄露率 | 是否泄露无权内容 |
| 业务 | 节省时间、任务完成率 | 是否减少人工查找成本 |

### 线上反馈如何进入系统

```text
User Feedback
  ├─ helpful / not helpful
  ├─ wrong citation
  ├─ outdated document
  ├─ missing source
  └─ permission issue
        │
        ▼
Eval Dataset / Index Fix / Connector Fix / Prompt Fix
```

注意：不是所有问题都应该修 Prompt。

- 找不到文档：可能是连接器或索引问题；
- 引用错文档：可能是 reranker 问题；
- 答案越权：是权限系统问题；
- 编造结论：是生成约束和证据包问题；
- 文档过期：是知识治理问题。

---

## 13.9 工程取舍

### 1. 答案速度 vs 证据完整性

搜索助手需要快，但企业决策需要证据。可以用分层响应：

```text
快速答案：先给初步结论和 3 条证据
深度分析：后台继续检索更多源，更新答案
```

### 2. 个性化 vs 信息茧房

个性化排序能提高命中率，但也可能让用户只看到本部门视角。关键问题需要跨源证据，而不是只取“最像用户平时点击的内容”。

### 3. 内部知识 vs 外部知识

内部知识有权限和上下文优势，外部知识有时效和广度优势。成熟系统需要标记来源类型：

```text
internal_verified
internal_unofficial
external_web
premium_data
user_uploaded
```

不同来源进入答案时应有不同置信度和引用方式。

---

## 13.10 可借鉴点

设计企业知识助手时，可以直接借鉴这些原则：

- 连接器必须同步权限，不只是同步内容；
- 检索默认使用 Hybrid Search，而不是只用向量库；
- 中间产物设计成 Evidence Package；
- 生成答案必须包含引用、证据和不确定性；
- 过期文档需要降权或显式标记；
- 反馈要能路由到索引、权限、Prompt、文档治理等不同修复路径；
- 高风险领域宁可拒答，也不要无证据自信回答。

---

## 本章小结

企业知识助手的成熟度，不取决于向量库有多先进，而取决于它是否把知识系统当成工程系统来治理：

- 数据源需要连接器；
- 内容需要索引；
- 权限需要同步；
- 检索需要混合排序；
- 答案需要证据；
- 反馈需要闭环；
- 质量需要评估。

一句话总结：

> 企业知识助手的核心不是“让模型知道更多”，而是“让模型只基于用户有权访问、足够新鲜、可被引用的证据回答”。

---

## 参考资料

1. [Glean Connectors - Glean Docs](https://docs.glean.com/connectors/about)
2. [Perplexity Search API - Perplexity Docs](https://docs.perplexity.ai/guides/search-guide)
3. [Perplexity Enterprise Pro](https://enterprise.perplexity.ai/)
