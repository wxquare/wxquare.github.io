# 第27章 Embedding、Rerank 与 RAG 基础

LLM 的参数记忆不是生产系统的知识库。企业知识、实时数据、权限信息、交易状态、项目文档和用户私有数据，都需要从外部系统接入。Embedding、Rerank 和 RAG 就是这条链路的基础。

## 27.1 宏观理解：RAG 在解决什么问题

RAG（Retrieval-Augmented Generation）把检索和生成结合起来：

```mermaid
flowchart LR
    A["用户问题"] --> B["Query Rewrite"]
    B --> C["Retrieval"]
    C --> D["Rerank"]
    D --> E["Context Builder"]
    E --> F["LLM Generation"]
    F --> G["带证据回答"]
```

它解决的是：

- 模型不知道最新事实；
- 企业数据不能进入训练集；
- 回答需要引用证据；
- 权限和租户隔离必须可控；
- 知识频繁变化；
- 错误需要可追溯。

RAG 不是“把文档塞进 prompt”，而是一条上下文供应链。

## 27.2 Embedding 是什么

Embedding model 把文本映射成向量。语义相近的文本，向量距离通常更近。

例如：

```text
"订单退款规则" -> [0.12, -0.03, ...]
"如何退订酒店订单" -> [0.10, -0.01, ...]
```

向量检索会找和 query embedding 最相近的文档 chunk。

常见相似度包括：

- cosine similarity；
- dot product；
- L2 distance。

## 27.3 Sparse、Dense 与 Hybrid Retrieval

### 27.3.1 Sparse Retrieval

BM25 等 sparse 方法依赖词面匹配。优点是可解释、对关键词强，缺点是同义词和语义改写能力弱。

### 27.3.2 Dense Retrieval

Dense retrieval 使用 embedding 向量匹配语义。优点是能处理同义改写，缺点是可能忽略精确关键词、版本号、错误码、代码符号。

### 27.3.3 Hybrid Retrieval

Hybrid retrieval 结合 sparse 和 dense，常用于生产系统。它既保留关键词精确性，又利用语义召回。

## 27.4 Rerank 是什么

Embedding 检索通常负责从大量文档中召回候选，例如 top 50 或 top 100。

Reranker 负责对候选进行更精细排序，例如选出 top 5。

常见 reranker 是 cross-encoder：它同时读取 query 和 document，判断相关性。相比 embedding，它更准但更慢。

所以典型结构是：

```text
召回要快，排序要准。
```

## 27.5 Chunking：切块决定召回上限

文档进入向量库前要切成 chunk。

chunk 太小：

- 语义不完整；
- 答案跨 chunk；
- 上下文缺失。

chunk 太大：

- 噪声多；
- embedding 表示被稀释；
- 召回后占用太多上下文。

生产系统常用策略：

- 按标题层级切；
- 按段落切；
- 代码按函数/类切；
- 表格按行组或业务实体切；
- 保留 metadata；
- chunk overlap；
- parent-child retrieval；
- late chunking。

## 27.6 Metadata 是 RAG 的骨架

只存文本和向量是不够的。每个 chunk 应该有 metadata：

- 文档 ID；
- 标题路径；
- 来源系统；
- 作者或负责人；
- 更新时间；
- 版本；
- 权限标签；
- 业务域；
- 语言；
- 文档类型；
- chunk 顺序；
- 原文链接。

没有 metadata，RAG 很难做过滤、权限、溯源、更新和评估。

## 27.7 Context Builder：不是简单拼接

检索结果不能直接粗暴塞给模型。Context Builder 要做：

- 去重；
- 排序；
- 合并相邻 chunk；
- 控制 token 预算；
- 标注来源；
- 区分事实、规则、示例、日志；
- 处理冲突；
- 保留引用 ID。

好的 RAG 系统输出给模型的是结构化上下文，而不是一堆无序文本。

## 27.8 工业实践：RAG 质量链路

RAG 的错误可以拆成几类：

- 没召回：正确文档不在候选里；
- 排错序：正确文档有，但排名太低；
- 上下文污染：无关内容进入 prompt；
- 证据冲突：多个来源说法不一致；
- 生成错误：证据正确但模型总结错；
- 权限错误：召回了不该看的内容；
- 更新错误：索引不是最新版本。

所以 RAG eval 不能只看最终答案，还要分阶段评估：

- retrieval recall；
- rerank precision；
- citation accuracy；
- answer faithfulness；
- latency；
- token cost；
- permission correctness。

## 27.9 工业实践：RAG 与长上下文的关系

长上下文模型让我们可以放更多文档，但它没有消灭 RAG。

长上下文适合：

- 用户明确提供的一组文档；
- 小规模代码库或报告；
- 需要跨全文综合的任务。

RAG 适合：

- 大规模知识库；
- 频繁更新；
- 权限复杂；
- 需要溯源；
- 需要低成本和低延迟；
- 需要检索日志和评估。

很多生产系统会结合两者：先检索缩小范围，再把高质量证据放进较长上下文。

## 27.10 科研现状：截至 2026-05

### 1. 更强 Embedding

Embedding 模型从单纯语义相似，发展到多语言、多任务、多粒度和指令化 embedding。BGE-M3 等模型尝试同时支持 dense、sparse 和 multi-vector 检索。

### 2. Late Interaction

ColBERT 代表的 late interaction 方法保留 token-level 表示，在效率和精度之间折中。它比单向量 embedding 更细粒度，但索引和存储成本更高。

### 3. Reranker 强化

Reranker 从 cross-encoder 发展到 LLM reranker、多阶段 rerank、领域 rerank。工业上重点是质量和延迟平衡。

### 4. GraphRAG 与结构化检索

知识图谱、实体关系、社区摘要和结构化索引可以帮助模型处理复杂知识任务，尤其适合跨文档、多实体、多跳问题。

### 5. Agentic RAG

Agentic RAG 让模型可以多轮检索、改写 query、选择工具、验证证据和补充缺失信息。它比 naive RAG 更强，但更难评估，也更容易出现成本和循环问题。

### 6. RAG 评估

研究和工业界都在从“最终答案打分”转向链路级评估：检索是否命中、证据是否支持、引用是否准确、答案是否忠实、冲突是否处理。

## 27.11 工程清单

设计 RAG 系统时，检查：

- 文档解析是否保留结构？
- chunk 策略是否按文档类型区分？
- metadata 是否足够支持权限和溯源？
- 是否使用 hybrid retrieval？
- 是否有 reranker？
- 是否有 query rewrite？
- 是否处理冲突和过期文档？
- 是否输出 citation？
- 是否记录 retrieval trace？
- 是否有 retrieval recall eval？
- 是否有 answer faithfulness eval？
- 是否有权限测试？

## 27.12 面试表达

一句话版：

> Embedding 负责语义召回，Rerank 负责精排，RAG 负责把外部知识以证据形式注入上下文。生产级 RAG 不是向量库加 Prompt，而是文档解析、chunking、metadata、检索、重排、上下文构建、权限和评估的完整链路。

展开版：

> 我会把 RAG 拆成 retrieval pipeline 和 generation pipeline。检索侧关注文档解析、chunk、metadata、hybrid retrieval、rerank 和权限过滤；生成侧关注 context builder、引用、冲突处理和忠实性。评估也要分层：先看是否召回正确证据，再看排序和引用，最后看模型是否基于证据回答。长上下文可以增强 RAG，但不能替代 RAG 的权限、更新、溯源和评估能力。

## 27.13 深入理解：Embedding 检索为什么会错

Embedding 检索不是语义魔法。它把一段文本压缩成一个向量，这个压缩会丢失信息。

常见错误包括：

- **数字不敏感**：`v1.2` 和 `v1.3` 很接近，但业务含义可能完全不同。
- **否定不敏感**：“可以退款”和“不可以退款”可能向量距离很近。
- **实体混淆**：相似产品、相似服务、相似接口容易召回错。
- **代码符号弱**：函数名、路径、错误码需要精确匹配。
- **长 chunk 稀释**：一个 chunk 里包含多个主题，embedding 表示变得模糊。
- **查询过短**：用户问“这个怎么处理”，缺少可检索信号。
- **权限不可见**：向量相似度不知道用户是否有权访问。

所以生产 RAG 不应只依赖 dense vector search。更稳的做法是 dense + sparse + metadata + rerank + 权限过滤。

## 27.14 深入理解：Rerank 是质量杠杆，也是延迟成本

Rerank 的价值在于它能同时看 query 和候选文档，判断细粒度相关性。它比 embedding 更适合处理否定、数字、实体、上下文条件和问题意图。

但 rerank 也有成本：

- cross-encoder 需要对每个候选单独或批量推理；
- top-k 越大，延迟越高；
- reranker 本身也可能有领域偏差；
- 长文档 rerank 会被截断；
- 多语言和代码场景需要专门评估。

一个常见工业配置是：

```text
BM25 top 100 + dense top 100 -> merge/dedup -> rerank top 50 -> context top 5~10
```

这个配置不是固定答案，但体现了原则：召回阶段宁可多一点，rerank 阶段负责筛掉噪声，context builder 阶段负责控制 token 和证据结构。

## 27.15 工业实践：文档解析决定 RAG 上限

很多 RAG 项目失败，不是 embedding 模型不够强，而是文档解析太差。

PDF、网页、Markdown、Word、表格、PPT、代码库都有不同结构。如果解析阶段丢掉标题、表格关系、图片说明、代码块边界和页码，后面的 embedding 再强也只能检索残缺文本。

高质量解析至少要保留：

- 标题层级；
- 段落顺序；
- 表格结构；
- 图片和图注；
- 代码块语言；
- 列表层级；
- 页码或原文定位；
- 链接和引用；
- 文档版本和更新时间。

RAG 的第一性原理是：检索系统只能召回已经被正确表达和索引的东西。

## 27.16 工业实践：权限过滤必须前置

权限过滤有两种方式：

- 检索前过滤：只在用户有权访问的文档集合里检索。
- 检索后过滤：召回后再剔除无权限文档。

生产系统通常应优先检索前过滤，因为检索后过滤可能造成两个问题：

- 向量库或日志中已经暴露了不该访问的候选；
- top-k 被无权限文档占据，过滤后剩余结果不足。

权限还要进入 cache 设计。检索结果、rerank 结果、context package、LLM trace 都可能包含敏感信息，不能跨租户复用。

对企业 RAG 来说，权限正确性不是附加需求，而是系统正确性的一部分。

## 27.17 深入理解：GraphRAG 解决的是“全局理解”问题

传统 RAG 擅长回答局部问题，例如“某个退款规则是什么”。但遇到全局问题会变弱，例如：

- 这批客户投诉的主要主题是什么？
- 一个代码库的核心架构是什么？
- 多份报告中共同的风险趋势是什么？
- 一个组织知识库中的关键实体关系是什么？

这类问题不是找一个 chunk 就够，而是需要跨文档聚合、实体关系和层级摘要。

GraphRAG 的思路是先抽取实体和关系，构建图结构或社区摘要，再在查询时检索相关子图和摘要。它提升的是全局 sensemaking 能力，但代价是构建复杂、更新成本高、抽取错误会传播。

工程上，GraphRAG 适合相对稳定、实体关系重要、需要全局总结的知识库；不适合高频更新、小规模文档或简单 FAQ。

## 27.18 研究补充：Agentic RAG 的机会和风险

Agentic RAG 让模型可以多轮检索、改写查询、选择索引、读取目录、验证证据。2026 年 A-RAG 这类工作进一步强调 hierarchical retrieval interface，让模型通过层级接口逐步探索信息空间。

它的优势是：

- 能处理复杂问题；
- 能根据中间结果调整检索；
- 能减少一次性塞入上下文的 token；
- 更接近人类研究资料的过程。

风险是：

- 多轮检索增加延迟和成本；
- 模型可能陷入无效搜索；
- trace 更复杂；
- eval 更难；
- 工具权限和停止条件更重要。

因此 Agentic RAG 需要状态机、预算、最大轮数、检索 trace、证据验证和失败降级，不应该只是让模型自由调用 search。

## 27.19 研究补充：RAG Eval 正在从单分数走向诊断视图

早期 RAG eval 常看最终答案是否正确。但最终答案错了，原因可能很多：

- 没召回；
- 召回了但排序低；
- 证据被 context builder 丢掉；
- 模型没读懂证据；
- 模型读懂了但生成时编造；
- 引用错；
- 权限错。

RAGAS、ARES、RAGChecker、RAGVUE 等方向都说明：RAG 评估需要拆成多个诊断维度。

更好的评估报告应该能回答：

```text
retrieval 错了还是 generation 错了？
证据支持答案吗？
答案里的每个 claim 是否可追溯？
遗漏了哪些关键证据？
judge 自己是否可靠？
```

这和本书的 Agent Evals 思路一致：复杂系统不能只看最终分数，必须看链路。

## 27.20 工程案例：企业知识助手的 RAG 迭代路径

一个企业知识助手可以这样迭代：

1. **Baseline**：Markdown/网页解析，dense retrieval，top-k 拼接。
2. **结构化解析**：保留标题、表格、代码块、来源和更新时间。
3. **Hybrid Retrieval**：加入 BM25，解决关键词、错误码、版本号问题。
4. **Rerank**：提升 top-k 相关性。
5. **Context Builder**：合并相邻 chunk，去重，加入 citation。
6. **权限过滤**：接入用户、团队、租户和文档 ACL。
7. **Eval**：建立问题集、证据集、答案评估和权限测试。
8. **Agentic Retrieval**：对复杂问题允许多轮检索和工具查询。
9. **Feedback Loop**：把失败样本回流到 query rewrite、chunking 和 rerank。

这个路径比一开始追求“最强向量模型”更稳，因为它每一步都能定位收益。

## 27.21 常见误区：RAG

### 误区 1：有向量库就是 RAG

向量库只是检索基础设施。生产级 RAG 还需要解析、chunk、metadata、权限、rerank、context builder、引用、eval 和监控。

### 误区 2：召回 top-k 越多越好

top-k 太大可能引入噪声，增加 token 成本，让模型更难定位证据。更好的方式是提升召回质量、rerank 和上下文组织，而不是盲目增大 k。

### 误区 3：RAG 可以完全消除幻觉

RAG 降低幻觉，但不能消除。模型仍可能误读证据、忽略证据、错误归纳或编造没有被证据支持的 claim。

### 误区 4：长上下文会淘汰 RAG

长上下文减少一部分检索压力，但不能替代权限、更新、溯源、索引、trace 和评估。大规模企业知识仍然需要检索系统。

## 27.22 专家问答

**问：为什么 hybrid retrieval 通常比纯向量更稳？**

因为业务问题既有语义相似，也有精确匹配。错误码、版本号、函数名、产品名和日期常常需要 sparse retrieval；同义表达和自然语言问题需要 dense retrieval。

**问：RAG 的第一优先级是换 embedding 模型吗？**

不一定。很多时候更重要的是文档解析、chunk 策略、metadata、query rewrite 和 rerank。换 embedding 只能解决一部分召回问题。

**问：如何判断生成是否忠实？**

把答案拆成 claim，逐条检查是否被证据支持。只看最终答案相似度不够，因为答案可能语言流畅但证据不支持。

**问：Agentic RAG 什么时候值得做？**

当问题需要多跳检索、动态探索、跨文档综合或工具查询时值得做。简单 FAQ 不需要 Agentic RAG，否则只会增加延迟和不稳定性。

## 27.23 RAG 错误排查表

| 现象 | 可能原因 | 排查方向 |
| --- | --- | --- |
| 答案完全不相关 | query rewrite 错、召回错 | 查看原始 query、rewrite query、top-k 文档 |
| 答案缺关键条件 | chunk 太小或 evidence 被裁剪 | 检查 chunk 边界和 context builder |
| 引用了过期规则 | metadata 缺版本或更新时间 | 增加版本过滤和更新时间排序 |
| 引用了无权限文档 | 权限过滤后置或 cache 污染 | 检查 ACL filter 和 trace 隔离 |
| 文档召回了但没用 | 上下文噪声太多或排序低 | 加 rerank、减少 top-k、提升信息密度 |
| 答案有证据但总结错 | generation 忠实性问题 | 加 citation、claim check、低温采样 |
| 简单问题很慢 | 检索链路过重 | 增加意图分类和快速路径 |
| 多跳问题失败 | naive RAG 不够 | 引入 agentic retrieval 或分步检索 |

这个表的价值是把“RAG 不准”拆成可行动问题。不要一上来就换 embedding 模型。

## 27.24 案例：代码库 RAG 与文档 RAG 的差异

文档 RAG 主要处理自然语言段落，代码库 RAG 还要处理：

- 文件路径；
- import 关系；
- 函数调用图；
- 类型定义；
- 测试文件；
- 配置文件；
- 生成代码和手写代码边界；
- 最近修改历史。

代码库 RAG 的 chunk 不应该只按固定 token 数切。更好的方式是按函数、类、模块和调用关系切，并保留 symbol metadata。

代码问题经常需要精确匹配，因此 BM25、路径过滤、符号索引和 AST 分析往往比纯向量更重要。

这也解释了为什么 Coding Agent 的上下文工程会比普通问答更复杂：它需要把代码结构、运行状态、错误日志和修改意图一起组织给模型。

## 27.25 参考资料

- [Retrieval-Augmented Generation for Knowledge-Intensive NLP Tasks](https://arxiv.org/abs/2005.11401)
- [ColBERT: Efficient and Effective Passage Search via Contextualized Late Interaction](https://arxiv.org/abs/2004.12832)
- [BGE M3-Embedding](https://arxiv.org/abs/2402.03216)
- [Improving Text Embeddings with Large Language Models](https://arxiv.org/abs/2401.00368)
- [GraphRAG: From Local to Global](https://arxiv.org/abs/2404.16130)
- [RAGAS: Automated Evaluation of Retrieval Augmented Generation](https://arxiv.org/abs/2309.15217)
- [A-RAG: Scaling Agentic Retrieval-Augmented Generation via Hierarchical Retrieval Interfaces](https://arxiv.org/abs/2602.03442)
- [RAGVUE: A Diagnostic View for Explainable and Automated Evaluation of Retrieval-Augmented Generation](https://arxiv.org/abs/2601.04196)
