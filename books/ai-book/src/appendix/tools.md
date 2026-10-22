# 附录C 工具、框架与版本导航

本附录提供候选入口和选型记录，具体机制以责任章和其章末来源为准。表中的“核验”指公开文档/源码所支持的限定范围，**没有运行兼容性、生产负载或价格测试**。“滚动文档”不是已安装包版本；落地时须另锁包版本、模型、端点、配置、权限和工作负载。未知表示没有足够证据，不等于不支持。

## 已有正文依据的分层目录

| 工具/协议 | 层次与已核验能力范围 | 版本/日期 | 原始来源 | 仍需验证的边界 |
| --- | --- | --- | --- | --- |
| OpenAI API / 模型 SDK | 模型请求与工具配置；应用仍需处理自定义工具执行与状态 | 滚动官方文档；2026-10-18 | [Using tools](https://developers.openai.com/api/docs/guides/tools)；[第12章](../part3/01-llm-api-protocol.md) | 具体模型/工具/区域/账户可用性、SDK 版本和费用未知 |
| Anthropic 模型 SDK | 消息、流和工具的供应商适配入口 | [第12章](../part3/01-llm-api-protocol.md)核验；2026-10-18 | [API 文档](https://docs.anthropic.com/en/api/overview) | 不由模型 SDK 推出持久任务、审批或业务幂等 |
| OpenAI Agents SDK | 代码式 Agent 循环、工具和交接 | [第14章](../part3/03-agent-runtime.md#runtime-components)说明执行职责；滚动文档，2026-10-17 | [官方 SDK 指南](https://developers.openai.com/api/docs/guides/agents/sdk) | 应用自己的状态、工具实现及部署保证需另验 |
| Claude Agent SDK | 嵌入 Coding 循环、工具与上下文能力 | 第14章核验；滚动文档，2026-10-18 | [官方概览](https://code.claude.com/docs/en/agent-sdk/overview) | 进程、工作区、凭据与权限边界 |
| LangChain / LangGraph | 高层 Agent 抽象 / 显式状态图和 checkpoint 组合 | 第14章核验；滚动文档，2026-10-17–18 | [LangChain](https://docs.langchain.com/oss/python/langchain/overview)、[Persistence](https://docs.langchain.com/oss/python/langgraph/persistence) | 内存 saver 不证明重启恢复；业务幂等由应用承担 |
| AutoGen Core / Microsoft Agent Framework | 异步消息执行 / Agent 与 Workflow 抽象 | 第14章核验；stable/滚动文档，2026-10-18 | [Core](https://microsoft.github.io/autogen/stable/user-guide/core-user-guide/index.html)、[Framework](https://learn.microsoft.com/en-us/agent-framework/overview/) | 分布式存储、owner、取消和重复副作用 |
| MCP | Host/Client/Server 能力连接协议 | [第15章](../part3/04-skills-tools-mcp-connectors.md)；2025-11-25，核验2026-10-18 | [版本化规范](https://modelcontextprotocol.io/specification/2025-11-25) | 版本/能力协商、认证与传输；发现不等于授权 |
| A2A | 独立 Agent 的任务、消息与产物协议 | 第14章；v0.3.0，非“最新版”承诺 | [版本化规范](https://a2a-protocol.org/v0.3.0/specification/) | 本书未实现或测试兼容；自建消息不是兼容证明 |
| vLLM | 推理执行、调度和 KV 管理参考 | [第9章](../part2/03-inference.md)核验；V1 缓存设计/滚动文档，2026-10-17 | [APC](https://docs.vllm.ai/en/latest/features/automatic_prefix_caching/)、[设计](https://docs.vllm.ai/en/latest/design/prefix_caching/) | 具体模型、量化、GPU 与 shape 的支持及收益 |
| Triton 语言 | GPU 计算程序与算子层 | [第8章](../part2/02-training.md)核验；main 文档，2026-10-17 | [语言介绍](https://triton-lang.org/main/programming-guide/chapter-1/introduction.html) | 生成 kernel 的正确性与性能未实测 |
| Triton Inference Server | 模型仓库、版本与服务生命周期 | 第9章核验；滚动文档，2026-10-17 | [Model Management](https://docs.nvidia.com/deeplearning/triton-inference-server/user-guide/docs/user_guide/model_management.html) | 不是 Triton 语言；模型后端能力分别核验 |
| Claude Code / Cursor / Codex | 终端、IDE、本地/云任务产品形态 | [第17章](../part4/01-coding-agent-systems.md)核验；滚动文档，2026-10-18 | 原始产品文档及限定论点见第17章参考资料 | 无永久优劣榜；功能、套餐与本机开关分别核验 |
| Codex CLI / Claude Code / OpenCode | Coding Agent 执行循环、上下文、工具、权限与证据边界 | [第17章](../part4/01-coding-agent-systems.md)；源码/官方滚动文档，2026-10-22 | 机制与来源映射见第17章参考资料；OpenCode 按 V2 核验 | Codex CLI 与 OpenCode 有公开源码；Claude Code 核心闭源；版本和入口需分别核验 |
| Pi | 终端 Runtime 与 SDK、扩展及会话资源 | [第18章](../part4/02-pi-agent-architecture.md)核验；2026-10-18 | [Pi SDK](https://pi.dev/docs/latest/sdk)；具体包与来源见该章 | 以该章 @earendil-works 命名空间与 ModelRuntime 为准；不外推旧 SDK 示例 |
| OpenClaw | Gateway、自有 Runtime、工具和会话历史 | 提交 `62cb03ac7a95f4b9cb435f8efc57f4034ba2649b`；2026-10-18 | [固定源码](https://github.com/openclaw/openclaw/tree/62cb03ac7a95f4b9cb435f8efc57f4034ba2649b)；[第19章](../part4/03-openclaw-architecture.md) | 活动历史与旧 JSONL 区分；profile 不自动提供隔离 |
| Hermes | 入口、Memory、Skill、delegate 与限定投递记录 | 提交 `18c525b669b6ea4ba62e6967d228af21d89aa9d3`；2026-10-18 | [固定源码](https://github.com/NousResearch/hermes-agent/tree/18c525b669b6ea4ba62e6967d228af21d89aa9d3)；[第20章](../part4/04-hermes-agent-architecture.md) | 特定 Gateway 投递 ledger 不证明任意业务 Outbox/Checkpoint/补偿；教学扩展单列 |

## 其他候选入口

以下承接原工具清单，**仅保留检索入口，没有在本批核验版本和功能支持**。各行版本、核验日期和兼容状态统一为未知；不得把分类当作配置建议、许可证结论或性能推荐。

| 调研目的 | 原始入口 | 需补的版本、日期与能力证据 |
| --- | --- | --- |
| 模型 API 与 Embedding | [Gemini](https://ai.google.dev/)、[Cohere](https://cohere.com/)、[Voyage AI](https://www.voyageai.com/)、[Sentence Transformers](https://www.sbert.net/) | 精确模型与 revision、维度、tokenizer、非对称编码、限额与定价：未知 |
| 本地模型工作台 | [Ollama](https://ollama.com/)、[LM Studio](https://lmstudio.ai/) | 版本、模型格式、工具协议与硬件组合：未知 |
| 数据/RAG 框架 | [LlamaIndex](https://docs.llamaindex.ai/)、[Semantic Kernel](https://learn.microsoft.com/en-us/semantic-kernel/)、[AutoGPT](https://github.com/Significant-Gravitas/AutoGPT) | 具体 API、持久性、权限和恢复边界：未知 |
| 向量检索存储 | [Chroma](https://www.trychroma.com/)、[Pinecone](https://www.pinecone.io/)、[Weaviate](https://weaviate.io/)、[Milvus](https://milvus.io/)、[Zilliz](https://zilliz.com/)、[Qdrant](https://qdrant.tech/)、[pgvector](https://github.com/pgvector/pgvector) | 索引、过滤、删除一致性、量级、HA 与费用：未知；按第13/20章方法实验 |
| Coding 辅助产品 | [Copilot](https://github.com/features/copilot)、[Cline](https://github.com/cline/cline)、[Trae](https://www.trae.ai/) | 版本、权限、执行模式、工作区和价格：未知；不能套用其他产品结论 |
| 个人知识工作台 | [Obsidian](https://obsidian.md/)、[Notion](https://www.notion.so/)、[Logseq](https://logseq.com/) | 插件/同步/导出、ACL 与数据生命周期：未知；插件单独核验 |
| 观测与评估 | [LangSmith](https://docs.langchain.com/langsmith/)、[Helicone](https://www.helicone.ai/)、[Weights & Biases](https://wandb.ai/)、[Braintrust](https://www.braintrustdata.com/)、[Promptfoo](https://www.promptfoo.dev/) | 采样、脱敏、留存、数据驻留、评估与计费能力：未知 |
| 托管与执行平台 | [Replicate](https://replicate.com/)、[Modal](https://modal.com/)、[Together AI](https://www.together.ai/) | 镜像、GPU、扩缩容、数据路径、配额与费用：未知 |

## 一次选型的最小记录

先给出任务、质量底线、数据权限和故障模型，再选择所在层；不要按团队大小直接绑定某个产品。模型 SDK、Agent SDK、图运行时、IDE/CLI 产品与托管平台可以组合，但重试、状态、取消、预算和完成判定必须各有一个明确责任方。

记录产品/包/模型版本、官方来源与日期、已验证能力、未知项、权限范围、输入输出协议、依赖版本及固定回归集。运行时另测断线、重启、重复投递、撤权、部分输出与回滚；计费按当日模型和合同逐项核对输入、输出、缓存、工具、存储与失败重试，不沿用固定单价。只有文档支持而未运行的能力仍标为“文档核验”，实验数据另附环境与证据。
