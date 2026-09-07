# 书籍介绍

欢迎阅读《AI Agent 工程实践：从大模型基础到生产级智能体系统》。

这不是一本单纯介绍工具按钮怎么点的书。它关注的是一个更长期的问题：当 AI 已经可以读代码、改代码、调用工具、规划任务、检索知识、长期运行甚至并行协作时，工程师应该怎样重新设计自己的工作方式、上下文系统、工具接口、验证回路和生产治理。

全书围绕核心公式 **Agent = LLM + 上下文 + 工具** 展开，沿着“大模型基础 -> Agent 工程 -> 应用实战 -> 前沿与研究”的路径层层递进。

## 全书工程主线图

本书按“**大模型 → Agent 工程 → Agent 应用与实战 → 前沿与研究**”四部分展开。主线不是“学一堆工具名”，而是先理解模型能力与边界，再把模型放进可约束、可验证、可治理的 Agent 运行环境，用这套方法完成真实工作，最后以研究型与前沿专题判断下一步能力边界。

```mermaid
flowchart LR
    A["第一部分：大模型<br/>原理 / 训练 / 推理 / 边界"] --> B["第二部分：Agent 工程<br/>Prompt / Context / Harness"]
    B --> C["Agent Runtime<br/>工具 / 知识 / 记忆 / 编排"]
    C --> D["生产治理<br/>Evals / Guardrails / Observability"]
    D --> E["第三部分：Agent 应用与实战<br/>Coding / 企业知识 / 告警 / PKM / 生活 Agent"]
    E --> F["第四部分：前沿与研究<br/>Research Agent / 多模态 / 具身 / 长期学习"]
```

## 本书解决什么问题

AI 工程实践里最容易踩的坑，不是模型不会生成内容，而是我们把不清楚的意图、不完整的上下文、没有边界的工具和没有验证回路的流程交给了模型。结果往往是：第一版看起来很聪明，第二版开始补洞，第三版引入新问题，最后人和 AI 一起在上下文里迷路。

本书的主线是把这种不稳定的协作方式，逐步收敛成可复用、可审查、可验证、可治理的工程系统。

读完之后，你应该能够回答五类问题：

- **模型边界**：LLM 擅长什么、不擅长什么，这些边界如何影响系统设计？
- **任务协议**：如何通过 Prompt、结构化输出和 Context 降低任务歧义与输出漂移？
- **系统落地**：如何通过 Harness、工具、知识、记忆和工作流构建 Agent Runtime？
- **生产治理**：如何评估、监控、调试、回滚并持续优化一个 Agent 系统？
- **现实应用**：如何把大模型与 Agent 用到编程、告警处理、企业知识、个人知识管理和研究任务中？

## 适合谁读

本书主要面向已经参与过真实软件项目的读者，包括后端工程师、AI 应用工程师、技术负责人，以及正在把 AI 编程或 Agent 系统引入团队流程的人。

你不需要是大模型研究员，但最好具备以下基础：

- 能读懂一种主流编程语言的代码示例；
- 理解 API、数据库、测试、日志、部署、权限等基本工程概念；
- 对 LLM、RAG、Agent、MCP、Evals 等术语有初步印象，遇到细节时愿意回查。

如果你刚开始接触 AI 编程，建议按章节顺序建立基础；如果你已经在团队里落地 Agent，则可以直接从第二部分的 Agent 架构、工具、知识系统和生产治理部分切入。

## 内容结构

### 第一部分：大模型

这一部分建立全书的模型基础：理解 LLM 的能力边界、训练与推理机制，以及模型能力如何影响系统设计。它回答“模型能做什么、不能做什么、工程上为什么不能把模型当成万能组件”。

| 章 | 章节 | 一句话概述 |
| --- | --- | --- |
| 第1章 | [大模型范式演进与工程学习地图](part1/01-llm-paradigm-roadmap.md) | 先建立大模型工程的全局视角，说明学习路线、关键概念和能力边界。 |
| 第2章 | [大模型基础知识：Token、Embedding、Transformer 与能力边界](part1/02-llm-basics.md) | 讲清大模型最核心的底层概念，以及这些概念如何决定系统设计。 |
| 第3章 | [大模型训练过程详解：Pretraining、Post-training、SFT、RL 与偏好优化](part1/03-llm-training.md) | 梳理模型从预训练到后训练的全过程，理解能力是如何被塑造出来的。 |
| 第4章 | [大模型推理机制详解：Prefill、Decode、Sampling、KV Cache 与 Reasoning Budget](part1/04-llm-inference.md) | 解释模型在推理时如何工作，以及延迟、成本和效果之间怎么权衡。 |
| 第5章 | [微调、量化与部署：LoRA、QLoRA、Serving Engine](part1/05-llm-adaptation-serving.md) | 讨论模型适配、压缩和部署的工程手段，帮助模型真正跑进生产环境。 |
| 第6章 | [世界模型与具身智能：从预测世界到行动系统](part1/06-world-models-embodied-ai.md) | 介绍世界模型与具身智能的关系，理解“会说”如何走向“会做”。 |
| 第7章 | [LLM 能力边界与架构约束](part1/07-llm-boundaries.md) | 从模型局限出发，反推工程系统该如何加约束、补短板。 |

### 第二部分：Agent 工程

这一部分讨论如何把大模型组装为真正可运行的 Agent。它先解释 Agent 如何从对话和 Prompt 应用演化为受控 Runtime，再依次展开任务协议、上下文、Harness、模型协议、工具、知识、记忆、编排和生产治理。

| 章 | 章节 | 一句话概述 |
| --- | --- | --- |
| 第8章 | [Agent 的演化与架构总纲：从对话应用到可治理 Runtime](part2/01-agent-architecture.md) | 说明 Agent 为什么不是简单聊天框，而是需要被治理的运行时系统。 |
| 第9章 | [Prompt Engineering 与结构化输出：从提示词到任务协议](part2/02-prompt-engineering.md) | 把提示词升级成可执行的任务协议，让输入和输出更稳定。 |
| 第10章 | [Context Engineering：从上下文注入到信息架构](part2/03-context-engineering.md) | 讲清上下文如何组织、压缩和注入，决定 Agent 的认知上限。 |
| 第11章 | [Harness Engineering：从模型调用到 Agent 运行环境](part2/04-harness-engineering.md) | 构建模型之外的执行外壳，负责控制、状态、权限和失败处理。 |
| 第12章 | [LLM API 协议：模型能力如何被系统消费](part2/05-llm-api-protocol.md) | 把模型接口当成系统契约来设计，而不是只看一次调用是否成功。 |
| 第13章 | [Agent 工具系统工程：Tool Calling、Skills 与 MCP](part2/06-tool-calling-mcp.md) | 说明工具如何成为 Agent 的手脚，以及如何安全、可扩展地接入工具。 |
| 第14章 | [Agent 知识系统：从知识源、RAG 到 Agentic RAG](part2/07-agent-knowledge-systems.md) | 介绍知识检索与增强生成的系统化做法，让 Agent 知道去哪找答案。 |
| 第15章 | [Agent 记忆系统：Memory、会话与长期上下文](part2/08-agent-memory.md) | 讨论记忆如何跨会话保存、更新和检索，支撑长期协作。 |
| 第16章 | [Agent 执行编排与平台架构：工作流、状态机、多 Agent 与框架生态](part2/09-workflow-orchestration.md) | 说明任务如何被拆解、编排和调度成稳定的执行流程。 |
| 第17章 | [Agent 生产治理：Evals、Guardrails 与可观测性](part2/10-agent-evals-guardrails-observability.md) | 把评估、约束和可观测性连成闭环，保证系统可用、可控、可迭代。 |

### 第三部分：Agent 应用与实战

这一部分将前两部分的能力放到现实工作中：先拆解成熟产品，再落到企业和个人任务。它关心的是“如何用大模型与 Agent 完成一项可交付、可验证、可治理的真实工作”。

| 章 | 章节 | 一句话概述 |
| --- | --- | --- |
| 第18章 | [AI Coding Agent 系统解析：从工作协议到 Harness 工程](part3/01-coding-agent-systems.md) | 从系统视角拆解 Coding Agent，理解它为什么能帮人写代码。 |
| 第19章 | [企业知识助手：RAG、搜索、权限与知识治理落地实践](part3/02-enterprise-knowledge-assistant.md) | 展示企业知识助手如何把检索、权限和知识治理真正结合起来。 |
| 第20章 | [Pi Agent 架构解析：终端原生 Coding Agent Runtime、扩展系统与上下文工程](part3/03-pi-agent-architecture.md) | 分析一个终端原生 Coding Agent 如何组织运行时、扩展和上下文。 |
| 第21章 | [OpenClaw 架构解析：个人 AI 助手的 Gateway、Runtime 与工具生态](part3/04-openclaw-architecture.md) | 讲述个人 AI 助手如何通过 Gateway、Runtime 和工具生态协同工作。 |
| 第22章 | [Hermes Agent 架构解析：自我进化、记忆与多入口 Agent Gateway](part3/05-hermes-agent-architecture.md) | 说明一个可自我进化的 Agent 如何通过记忆和入口管理持续扩展。 |
| 第23章 | [DoD Agent：企业级告警处理与知识答疑系统](part3/06-dod-agent-case-study.md) | 用告警与答疑场景展示企业级 Agent 的生产落地方式。 |
| 第24章 | [从零实现一个可观测 Coding Agent](part3/08-mini-agent-observability.md) | 通过一个可复现项目展示 Coding Agent 的观测、调试与扩展方法。 |
| 第25章 | [个人知识管理 Agent 实践](part3/07-pkm-agent-case-study.md) | 说明个人知识管理如何借助 RAG、记忆和工作流形成闭环。 |
| 第26章 | [持续进化的生活 Agent：从日常反馈到可信能力闭环](part3/10-daily-life-evolving-agent.md) | 讨论生活 Agent 如何在反馈、审批和回滚中安全地持续进化。 |

### 第四部分：前沿与研究

这一部分讨论工程实践之外仍在快速演化的前沿问题。它不局限于 Research Agent，后续可扩展到多模态、具身智能、长期学习和世界模型等专题。

| 章 | 章节 | 一句话概述 |
| --- | --- | --- |
| 第27章 | [AI 智能体研究现状、工程瓶颈与未来理想能力架构报告](part4/01-agent-frontier-research.md) | 从研究和工程两条线梳理智能体的现状、瓶颈与未来方向。 |

### 附录

附录提供术语、参考资料、工具清单、系统设计思考题与项目实践模板。它们不再作为正文主线，而是作为随用随查的材料。

## 在线阅读

- **GitHub Pages**：https://wxquare.github.io/ai-book/
- **源码仓库**：https://github.com/wxquare/wxquare.github.io

## 反馈与贡献

欢迎通过 Issue、Pull Request 或博客留言提供反馈。尤其欢迎三类反馈：

- 哪些章节读起来跳跃，缺少承上启下；
- 哪些代码或架构图难以复现；
- 哪些工具、模型或最佳实践已经过时。

## 版本信息

- **当前版本**：v1.0
- **发布日期**：2026 年 4 月
- **更新计划**：持续更新，优先深化大模型基础、Agent 运行时、成熟系统解析、案例、评估体系和生产治理实践。
