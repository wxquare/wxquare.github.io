# 附录A 术语与符号速查

本附录帮助定位正文中的定义。具体机制、来源与版本边界以表内责任章为准；同一个词在模型计算、任务运行和业务系统中可能指不同对象。

## 易混概念与责任边界

| 术语 | 本书约定 | 责任章 |
| --- | --- | --- |
| Context / Context Package | 当前一步装配给模型的指令、材料、约束与结果视图；受窗口、权限和版本限制 | [第15章](../part2/03-context-engineering.md) |
| Memory / Memory Item | 跨步骤或会话保存的候选经验、偏好和摘要；有来源、作用域、失效与写入门控 | [第20章](../part2/08-agent-memory.md) |
| State / Run、Step、Event | 权威执行事实、版本及事件；模型摘要和 Todo 只能引用，不能代替账本 | [第11章](../part2/infra-11-runtime.md) |
| Evidence | 能支撑特定论点、可回读原位置且当前有权访问的版本化证据；不等于任意 chunk | [第19章](../part2/07-agent-knowledge-systems.md) |
| Operation | 一次业务副作用的稳定身份与状态；超时后可为 unknown，不能直接按失败重做 | [第11章](../part2/infra-11-runtime.md) |
| Checkpoint | 已提交状态与恢复位置；不回滚已经发生的外部副作用。训练 checkpoint 另含优化器、随机数和数据位置 | [运行时第11章](../part2/infra-11-runtime.md)、[训练第8章](../part2/infra-08-training.md) |
| KV Cache | 前向计算中各层的键和值；用于复用历史计算，仍占用设备或分层存储资源 | [第9章 §9.3](../part2/infra-09-inference.md) |
| Prefix / Prompt Cache | 复用兼容的前缀计算；自托管实现与商业 API 的命中、计费、保留语义分别核验 | [第9章](../part2/infra-09-inference.md)、[Hermes 第27章](../part3/05-hermes-agent-architecture.md) |
| 精确响应缓存 / 语义响应缓存 | 前者匹配请求及其依赖版本，后者先召回相似候选再验证答案等价；两者都需授权、新鲜度和失效控制 | [第9章 §9.3](../part2/infra-09-inference.md) |
| Tool Calling / Skill / MCP | 分别是结构化行动请求、可复用方法上下文、能力连接协议；实际执行、权限和审计由 Runtime 承担 | [第18章](../part2/06-tool-calling-mcp.md) |
| MCP / A2A | 本书 MCP 主体按 2025-11-25，连接工具、资源和提示；A2A 按 v0.3.0，描述独立 Agent 的任务/消息/产物协作。自建 mailbox 不自动兼容 A2A | [第18章](../part2/06-tool-calling-mcp.md)、[第21章 §21.7](../part2/09-workflow-orchestration.md) |
| Triton 语言 / Triton Inference Server | 前者用于编写 GPU 计算程序；后者负责模型服务与模型管理，二者不是同一层的软件 | [第8章 §8.4](../part2/infra-08-training.md)、[第9章](../part2/infra-09-inference.md) |
| Model Manifest / Eval Case | 前者绑定模型、模板、格式与执行兼容条件；后者绑定输入、预期、不变量和版本化评估配置 | [第7章 §7.8](../part1/07-llm-boundaries.md)、[第10章](../part2/infra-10-data-eval.md) |

## 模型计算符号

下表沿用[第1章](../part1/01-llm-foundations-transformer.md)及[第9章](../part2/infra-09-inference.md)的记账口径。其他章节可用局部符号，但须重新定义。

| 符号 | 含义与单位 |
| --- | --- |
| B、T、S | batch 大小、当前查询长度、可见键/历史长度；不是字符数 |
| D、L | 残差宽度、层数 |
| H_q、H_kv、d_h | 查询头数、KV 头数、单头维度；GQA 不减少查询头数 |
| s | 每元素字节数；量化还需计入 scale、布局、工作区等开销 |
| logits / labels / loss mask | 预测分数、目标 token、有效监督位置；标签与预测错位只对齐一次 |
| TTFT / TPOT | 首 token 时间 / 每输出 token 时间；是否包含排队、网络和重试须说明 |
| prefill / decode | 处理输入序列 / 增量生成；两阶段的资源瓶颈和缓存收益不同 |
| token / Context Window | tokenizer 产生的模型单位 / 模型及接口允许的输入输出边界；汉字与 token 无固定比例，预算用目标 tokenizer 与实际 usage |

## 常用术语索引

| 术语 | 速查定义 | 详解 |
| --- | --- | --- |
| Agent / Harness | 在反馈中推进任务的系统 / 包围模型的控制、工具、状态、验证与观测环境 | [第13章](../part2/01-agent-architecture.md)、[第16章](../part2/04-harness-engineering.md) |
| Attention / Transformer / Positional Encoding | 内容匹配和加权 / 以注意力及逐位置网络组成的结构 / 注入顺序信息的方法 | [第1章](../part1/01-llm-foundations-transformer.md) |
| Embedding / Cosine Similarity | 向量表示 / 方向相似度；相似不等于事实正确或业务等价 | [第19章](../part2/07-agent-knowledge-systems.md) |
| Chunking / BM25 / Hybrid Search / RRF / Reranking | 文档切分 / 词法排序 / 多路召回 / 排名融合 / 候选重排 | [第19章](../part2/07-agent-knowledge-systems.md) |
| RAG / GraphRAG / Agentic RAG | 检索辅助生成 / 关系型组织的检索 / 受控循环中的检索决策；都需证据与权限验收 | [第19章](../part2/07-agent-knowledge-systems.md) |
| CoT / ReAct / Self-Consistency | 分步推理提示 / 交替行动与外部观察 / 多次采样后聚合；不保证答案正确，也不要求暴露隐藏内部推理 | [第4章](../part1/04-reasoning-generation-algorithms.md)、[第21章](../part2/09-workflow-orchestration.md) |
| Zero-shot / Few-shot / Prompt Engineering | 无示例 / 有少量示例 / 设计输入和输出契约；这里不表示权重已训练更新 | [第14章](../part2/02-prompt-engineering.md) |
| Plan-and-Execute / State Machine / Decision Tree | 计划与执行分工 / 显式状态迁移 / 条件决策树；可与模型循环组合 | [第21章](../part2/09-workflow-orchestration.md) |
| Sliding Window / Loop Detection / Early Stopping | 保留近期材料 / 按状态与观察检测空转 / 按预算、质量或停止条件结束 | [第15章](../part2/03-context-engineering.md)、[第11章](../part2/infra-11-runtime.md) |
| Fingerprint | 为内容、状态或请求构造的识别值；去重范围、碰撞和版本规则需定义，不天然代表同一业务操作 | [第11章](../part2/infra-11-runtime.md) |
| Hallucination / Guardrail / Observability | 缺少事实支持的生成 / 模型外可执行约束 / 通过事件与信号复盘系统行为 | [第22章](../part2/10-agent-evals-guardrails-observability.md) |
| MTTR / SLO / SOP | 恢复时间指标（须定义起止）/ 服务目标 / 标准操作流程 | [第12章](../part2/infra-12-reliability-governance.md) |
| Backlink / 知识编译 / Vector Database | 反向引用 / 将材料变成有来源的可用知识 / 向量索引与检索存储；均不代替原始事实源 | [第30章](../part3/07-pkm-agent-case-study.md)、[第19章](../part2/07-agent-knowledge-systems.md) |
| Spec Coding / Vibe Coding | 以规格验收驱动的协作 / 快速对话试探的工作方式；两者都需要版本、测试与授权 | [第23章](../part3/01-coding-agent-systems.md) |
| 多 Agent / 分级决策 / 混合架构 | 职责可分的多个执行主体 / 按风险选择控制方式 / 确定性流程与开放推理组合 | [第21章](../part2/09-workflow-orchestration.md)、[第22章](../part2/10-agent-evals-guardrails-observability.md) |
