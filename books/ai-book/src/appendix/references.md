# 附录B 学习路径与延伸阅读

本附录按学习目的导航。**每章末尾的参考资料是该章论点、数字和产品行为的权威来源清单**；本页不维护第二套全集，也不把论文名或工具名当作覆盖证明。先读机制，再沿章末原始论文、标准或官方文档阅读，最后用练习检验理解。

## 按问题选择阅读路径

| 学习目的 | 先读 | 再读与原始来源入口 | 阅读时核对什么 |
| --- | --- | --- | --- |
| 从张量理解模型 | [模型结构](../part1/01-llm-foundations-transformer.md) | [预训练](../part1/02-pretraining-data-scaling.md)的目标函数及两章参考资料 | Attention 的维度、mask、位置、GQA/MLA 与任务目标 |
| 理解行为如何塑形 | [后训练](../part1/03-post-training-alignment.md) | [模型适配与能力扩展](../part1/05-finetuning-compression-multimodal.md)的适配数据与量化来源 | 监督位置、PPO/GAE、DPO/GRPO、数据独立性 |
| 评估推理与多模态 | [推理期扩展](../part1/04-reasoning-generation-algorithms.md) | [模型适配与能力扩展](../part1/05-finetuning-compression-multimodal.md)、[具身智能与 Physical AI](../part1/06-world-models-embodied-ai.md) | 额外预算、验证器、未观察范围与物理安全 |
| 建立资源账本 | [AI Infra 全景](../part2/01-ai-infra-overview.md) | [训练平台与数据管线](../part2/02-training.md)、[推理服务与资源调度](../part2/03-inference.md) | 并行切分轴、通信、瞬时峰值、工作负载与版本 |
| 设计可恢复运行时 | [Agent Runtime](../part3/03-agent-runtime.md#delivery-idempotency) | [LLM API 协议](../part3/01-llm-api-protocol.md)、[Agent Runtime恢复机制](../part3/03-agent-runtime.md#recovery) | 提交、投递、幂等、unknown、owner、断线与迟到结果 |
| 组织任务与上下文 | [Agent 的上下文工程](../part3/02-context-engineering.md) | [任务协议](../part3/02-context-engineering.md#task-protocol)、[装配与预算](../part3/02-context-engineering.md#context-builder)、[长期记忆](../part3/02-context-engineering.md#memory) | 指令权威、摘要保真、来源失效与事实状态 |
| 选择工具与协议 | [Agent 工具系统工程](../part3/04-skills-tools-mcp-connectors.md) | [Agent Runtime Runtime 组成](../part3/03-agent-runtime.md#runtime-components)及[附录C](tools.md) | Schema、执行权、MCP 传输迁移、A2A 版本与未知能力 |
| 构建知识助手 | [Agent 的上下文工程](../part3/02-context-engineering.md#knowledge) | — | 任务协议、知识检索、证据组织和上下文边界 |
| 理解 Coding 系统 | [Coding Agent 系统](../part4/01-coding-agent-systems.md) | [Pi Agent 系统 Pi](../part4/02-pi-agent-architecture.md) | 共通执行闭环、产品实现差异、开源范围与独立验证 |
| 观察长期助手 | [Hermes Agent 系统 Hermes](../part4/04-hermes-agent-architecture.md) | [OpenClaw 系统 OpenClaw](../part4/03-openclaw-architecture.md) | 身份、会话、记忆、投递与业务副作用分别取证 |
| 形成质量与发布闭环 | [评估与反馈闭环](../part2/04-data-eval.md) | [生产运营与治理控制](../part2/05-production-operations-governance.md)、[Agent 治理与调优](../part3/05-agent-governance-optimization.md) | 固定评估集、Judge 校准、回归、成本及回滚证据 |
| 阅读研究报告 | [AI 智能体研究现状、工程瓶颈与未来理想能力架构报告](../part5/01-agent-frontier-research.md) | 沿其参考资料回到对应论文版本、任务和评测协议 | 历史结果、研究提案、公开实现与实测结论的区别 |

## 章末练习入口

以下仅链接原练习，不复制答案或创建另一套题库。卡片定义输入、输出与验收条件；**设计完成不表示代码实现、GPU 实测或性能验证完成**。

- [模型结构 Attention / GQA 练习卡](../part1/01-llm-foundations-transformer.md)：张量、mask 与增量一致性。
- [后训练 SFT 对齐练习卡](../part1/03-post-training-alignment.md)：固定模板、监督位置与单次 shift。
- [模型适配与能力扩展校准练习卡](../part1/05-finetuning-compression-multimodal.md)：零块、异常值、尾组及误差报告。
- [训练平台与数据管线 GPU 与规约练习卡组](../part2/02-training.md)：层级、同步、布局、正确性与环境化性能。
- [推理服务与资源调度显存账本练习卡](../part2/03-inference.md)：固定模型配置、长度、并发、预留与敏感性分析。

## 从面试题定位正文

[题库的书稿覆盖与定位说明](https://github.com/wxquare/leetcode-primer/blob/master/system-design-interview/question-bank/ai-agent-engineering.md#书稿覆盖与定位)定义覆盖口径；每个题目下方给出对应章节和具体小节，以及无法由本书回答的部分。逐题映射只在题库维护，本附录提供学习路径，不复制题目与答案。

做题时先读主定位，跨层问题再读补充章节。例如工具成功但响应丢失，应先读[Agent Runtime 的 Operation 与 Attempt](../part3/03-agent-runtime.md#delivery-idempotency)，再读工具运行时的重试分类；不能只看 MCP 协议。手写题则必须自行完成实现和验证，章末练习卡仅定义验收要求。

## 如何继续找资料

稳定概念优先读经典教材、原始论文和正式标准；实现行为读对应版本的官方文档与源码。课程、博客、演讲、论文聚合页和厂商主页可用于发现问题，不能代替支撑具体结论的原始材料。收藏链接时记录作者/机构、标题、版本或年份、原始 URL、访问日期及可支撑范围。

产品排名、价格、模型窗口、硬件支持与基准分数变化很快。无法取得同版本证据就保留未知；不要把别的模型、硬件或评测集上的结果迁移为自己的结论。工具候选从[附录C](tools.md)进入，论点依据回到相应章节的参考资料。
