# AI 工程实践书籍结构优化设计文档

**项目**：《AI 工程实践：从编程到 Agent 的完整指南》  
**日期**：2026-04-27  
**作者**：wxquare + AI Assistant  
**状态**：已根据用户反馈重构，待用户审阅

---

## 1. 背景与目标

当前书稿已经形成完整 mdBook 结构，但目录仍有明显的“内容归类”痕迹：AI 编程、Agent 设计、案例、理论补充、面试实战被分成几个板块，读者能看到很多好内容，却不容易感受到一条足够强的 AI 架构方法论主线。

用户明确提出：希望深入讲解 **Prompt Engineering**、**Context Engineering**、**Harness Engineering**，并且三者各自成为独立章节。这个反馈改变了本书的结构重心。

新版结构应从“章节拼盘”升级为“AI 架构能力栈”：

```text
Prompt Engineering
  ↓
Context Engineering
  ↓
Harness Engineering
  ↓
Agent Runtime
  ↓
Knowledge / Memory / Retrieval
  ↓
Evals / Guardrails / Observability
  ↓
Case Studies
  ↓
Interview / Portfolio
```

优化目标：

- 把 Prompt、Context、Harness 三章确立为全书方法论主干。
- 用架构师视角组织后续内容，而不是按“专题”或“理论补充”堆放。
- 将面试内容收束为表达、题库和作品集，不再承载核心架构章节。
- 将 Evals、Guardrails、Tool Calling、Debug 等内容放回生产级 Agent 架构体系。
- 让读者读完后形成一套从单次模型调用到生产级 Agent 系统的完整设计框架。

---

## 2. 核心定位

本书的新定位：

> 以 Prompt Engineering、Context Engineering、Harness Engineering 为主线，讲解工程师如何把 LLM 从一次性文本生成能力，逐步组织成可控、可评估、可治理的生产级 Agent 系统。

三层方法论关系：

| 层次 | 核心问题 | 典型产出 | 对应能力 |
|:---|:---|:---|:---|
| Prompt Engineering | 如何把意图变成模型可执行的任务协议 | Prompt 模板、输出契约、任务说明 | 单次调用可控 |
| Context Engineering | 如何把正确的信息放进模型工作区 | 上下文包、文档索引、记忆策略、上下文优先级 | 信息供给可控 |
| Harness Engineering | 如何构建让 Agent 稳定运行的外部系统 | 工具、工作流、验证、权限、观测、反馈回路 | 系统行为可控 |

这三者不是并列技巧，而是递进关系：

- Prompt 解决“怎么说清楚”。
- Context 解决“让模型知道什么”。
- Harness 解决“让模型在什么系统中行动”。

---

## 3. 推荐新版目录

### 第一部分：AI 工程方法论基础

这一部分是全书的理论和实践根基。它回答：AI 工程为什么不能停留在提示词技巧，应该如何从 Prompt 走向 Context，再走向 Harness。

1. 第 1 章 从 Vibe Coding 到 Spec Coding：AI 编程范式演进
2. 第 2 章 Prompt Engineering：从提示词到任务协议
3. 第 3 章 Context Engineering：从上下文注入到信息架构
4. 第 4 章 Harness Engineering：从模型调用到 Agent 运行环境
5. 第 5 章 Claude Code：终端原生的 AI Agent 实践

章节顺序说明：

- 第 1 章建立“即兴使用 AI”和“规范驱动 AI”的差异。
- 第 2 至第 4 章构成本书核心三部曲。
- 第 5 章再进入 Claude Code，让具体工具承接前面的方法论，而不是让工具先定义读者视角。

### 第二部分：Agent 架构与运行时设计

这一部分回答：当 AI 不再只是回答问题，而要在系统中执行任务时，运行时架构如何设计。

6. 第 6 章 LLM 能力边界与架构约束
7. 第 7 章 Agent 架构设计与决策框架
8. 第 8 章 Tool Calling 与 MCP 工程架构
9. 第 9 章 Agent 工作流、状态机与多 Agent 编排

章节边界：

- LLM 能力边界是架构设计的前置约束，不应放在后面的“理论补充”。
- Tool Calling 与 MCP 是 Agent 行动能力的基础，不应只作为面试深化章节。
- 多 Agent 协作应和状态机、工作流一起讲，形成运行时编排章节。

### 第三部分：知识、上下文与记忆系统

这一部分回答：Agent 如何获得事实、保持连续性，并避免上下文污染。

10. 第 10 章 RAG 与上下文工程基础
11. 第 11 章 检索系统工程：从文档管道到评估闭环
12. 第 12 章 Agent Memory 与状态管理

章节边界：

- RAG 章节讲基本链路和上下文构建。
- 检索系统章节讲生产级检索，包括文档管道、metadata、hybrid search、rerank、权限过滤和检索 eval。
- Memory 章节讲任务状态、会话连续性、长期记忆、历史案例和遗忘机制。

### 第四部分：生产级 Agent 治理

这一部分回答：如何证明 Agent 有效，如何限制风险，如何定位失败，如何持续改进。

13. 第 13 章 Agent Evals：从离线评估到线上质量闭环
14. 第 14 章 Guardrails 与 Agent 安全架构
15. 第 15 章 可观测性、成本优化与 Trace 诊断
16. 第 16 章 Agent 系统全生命周期工程
17. 第 17 章 Agent 失败模式、Debug 与持续改进

章节边界：

- Evals 是生产级 Agent 的质量底座，不应只放在面试部分。
- Guardrails 是安全架构，不应只作为面试知识点。
- Debug 与失败模式是系统治理能力，作品集表达只保留在第六部分。
- 生命周期章节作为第四部分收束，连接需求、架构、开发、测试、上线和持续优化。

### 第五部分：完整架构案例

这一部分回答：前四部分的方法如何组合成可落地系统。

18. 第 18 章 DoD Agent：电商告警自动处理系统
19. 第 19 章 个人知识管理 Agent 实践
20. 第 20 章 从零实现一个可观测 Mini Agent

章节边界：

- DoD Agent 展示企业生产环境中的 Agent 架构。
- 个人知识管理 Agent 展示长期知识、Memory、RAG 和个人工作流。
- Mini Agent 作为可复现案例，服务读者实操和作品集。

### 第六部分：Agent 应用工程师面试与作品集

这一部分回答：如何把前面的方法论和项目经验讲清楚。

21. 第 21 章 Agent 应用工程师能力地图与 30 天冲刺计划
22. 第 22 章 Agent 系统设计面试题库
23. 第 23 章 项目复盘、作品集模板与面试表达

章节边界：

- 面试部分不再承载 Evals、Guardrails、Tool Calling 等核心内容。
- 它负责把前面章节重新组织成岗位能力、系统设计回答和项目表达。
- 面试话术从主线章节迁移到这里，主线章节保留工程表达和设计模板。

---

## 4. 三个核心章节的深度设计

### 4.1 第 2 章 Prompt Engineering：从提示词到任务协议

本章定位：

Prompt Engineering 不是“写神奇提示词”，而是设计模型可执行的任务协议。它是 AI 工程的最小控制面。

建议结构：

```text
2.1 Prompt Engineering 的边界：能解决什么，不能解决什么
2.2 Prompt 的四层结构：Role / Task / Context / Output Contract
2.3 任务协议设计：目标、输入、约束、步骤、验收标准
2.4 结构化输出：JSON Schema、函数调用、格式校验、失败重试
2.5 Few-shot 与反例：降低歧义，而不是堆示例
2.6 Prompt 与工具 Schema 的协同设计
2.7 Prompt 版本管理、评估与回滚
2.8 常见失败模式：指令冲突、输出漂移、过度拒答、编造依据
```

必须讲深的点：

- Prompt 是运行时协议，不是自然语言愿望。
- Prompt 的目标不是让模型“更聪明”，而是减少歧义和输出自由度。
- 结构化输出不是格式美化，而是让模型输出进入后端 workflow。
- Prompt 不能承担权限、安全、事实校验，这些必须由 Harness 承担。
- Prompt 要版本化，并和 eval dataset 绑定。

素材来源：

- `part4/prompt-engineering.md`
- `part4/chapter10.md` 中 Prompt Engineering 相关内容
- `part1/chapter3.md` 中三次范式跃迁相关内容

### 4.2 第 3 章 Context Engineering：从上下文注入到信息架构

本章定位：

Context Engineering 是本书的关键升级点。它不是简单“多塞资料”，而是设计模型在当前任务中应该看到的信息集合、优先级、来源、时效、权限和压缩策略。

建议结构：

```text
3.1 Context Engineering 的本质：模型当前应该知道什么
3.2 上下文类型：用户意图、项目规范、业务事实、工具结果、历史记忆
3.3 上下文优先级：当前输入、工具结果、权威文档、记忆、历史案例
3.4 上下文预算：token、成本、延迟与信息密度
3.5 上下文压缩：摘要、滑动窗口、事件化、引用裁剪
3.6 上下文检索：RAG、metadata、query rewrite、rerank
3.7 上下文污染：过期信息、错误记忆、权限越界、Context Rot
3.8 上下文防火墙：子代理、新会话、任务隔离
3.9 项目级上下文：CLAUDE.md、AGENTS.md、docs/、specs/ 的组织方式
```

必须讲深的点：

- 上下文是信息架构，不是 prompt 的附属字段。
- 上下文质量比上下文长度更重要。
- Context Rot 是长会话 Agent 的核心失败模式。
- RAG、Memory、工具结果、项目规则都是上下文来源，但可信度不同。
- Context Engineering 要同时考虑相关性、时效性、权限、成本和可追溯性。

素材来源：

- `part1/chapter3.md` 中 Context Engineering、上下文防火墙内容
- `part4/memory.md`
- `part4/chapter11.md`
- `part4/vector-search.md`
- `part2/chapter7.md` 中 Prompt 压缩和成本优化内容

### 4.3 第 4 章 Harness Engineering：从模型调用到 Agent 运行环境

本章定位：

Harness Engineering 是 Prompt 和 Context 的系统级外壳。它负责把模型能力放进一个可验证、可约束、可观测、可迭代的工程环境。

建议结构：

```text
4.1 Harness 的定义：Agent = Model + Harness
4.2 Harness 的六层架构：Context / Tools / Workflow / Guardrails / Evals / Observability
4.3 工具系统：Tool Registry、MCP、权限、超时、重试、审计
4.4 工作流控制：状态机、DAG、Router、Plan-and-Execute
4.5 验证回路：测试、lint、eval、review、human-in-the-loop
4.6 安全护栏：输入、上下文、工具、输出四层 Guardrails
4.7 可观测性：trace、step、tool call、cost、failure category
4.8 Harness 迭代：每次 Agent 犯错，如何沉淀成规则、测试或工具改造
```

必须讲深的点：

- Harness 不是单个框架，而是模型周围的运行环境。
- 好的 Harness 会主动缩小模型自由度。
- 工具、工作流、guardrails、evals、observability 都是 Harness 的组成部分。
- Harness 的优化对象不是某个回答，而是 Agent 行为分布。
- 每次失败都应该沉淀为规则、测试、数据集、工具 schema 或工作流调整。

素材来源：

- `part1/chapter3.md`
- `part2/chapter5.md`
- `part2/chapter6.md`
- `part2/chapter7.md`
- `part5/chapter13.md`
- `part5/chapter14.md`
- `part5/chapter15.md`

---

## 5. 现有文件重组策略

### 5.1 第一部分重组

当前 `part1/chapter3.md` 同时讲 Prompt、Context、Harness，内容应拆分：

- Prompt Engineering 的概念和实践迁入新第 2 章。
- Context Engineering 的概念、项目上下文、上下文防火墙迁入新第 3 章。
- Harness Engineering 的六大组件、行业案例、检查清单保留并扩展为新第 4 章。
- Claude Code 从当前第 2 章调整为新第 5 章，用作工具实践承接。

### 5.2 第二部分重组

当前 Agent 系统设计章节整体保留，但顺序应跟运行时架构一致：

- LLM 能力边界提前到第 6 章。
- Agent 架构决策作为第 7 章。
- 工具系统与 Tool Calling / MCP 工程深化合并为第 8 章。
- 多 Agent 协作与 workflow / state machine 合并为第 9 章。

### 5.3 第三部分重组

当前 RAG、Memory、Vector Search 不应作为零散理论专题，而应组成知识系统部分：

- RAG 基础讲端到端链路。
- 检索系统工程讲生产级检索。
- Memory 讲状态、会话、长期记忆和历史经验。

### 5.4 第四部分重组

当前面试部分中的 `Agent Evals`、`Guardrails`、`Tool Calling`、`失败模式/Debug` 应拆分处理：

- Evals 移到生产治理。
- Guardrails 移到生产治理。
- Tool Calling 移到 Agent 运行时设计。
- 失败模式与 Debug 移到生产治理。
- 作品集模板和面试表达保留在第六部分。

### 5.5 第五、第六部分重组

案例部分应放在治理体系之后，因为案例需要综合使用 Prompt、Context、Harness、RAG、工具、Evals、Guardrails 和 Observability。

面试部分应放最后，只承担表达和转译职责。

---

## 6. 面试内容迁移原则

主线章节中不应频繁出现“面试官问”“面试表达”这类措辞。它会让读者误以为主书定位从工程实践切换成面试资料。

处理规则：

1. 第一至第五部分保留工程设计、检查清单、决策模板。
2. 直接面试话术迁移到第六部分。
3. 主线章节中的“面试表达”改为“设计表达模板”“架构说明模板”或删除。
4. 第 21 章负责把前五部分内容重新组织为岗位能力地图。
5. 第 22 章负责系统设计题。
6. 第 23 章负责项目复盘、作品集和面试表达。

---

## 7. 实施范围

本次结构优化建议的实施范围：

- 更新 `ai-book/src/SUMMARY.md` 的目录结构、部分名称和章节编号。
- 更新 `ai-book/src/README.md` 的内容结构说明和阅读路径。
- 新增或拆分 Prompt Engineering、Context Engineering、Harness Engineering 三个核心章节。
- 移动或合并现有章节内容，使其符合新版能力栈。
- 更新章节间的“下一章预告”和编号引用。
- 迁移或改写主线章节中的面试表达。
- 清理重复内容，尤其是 Prompt、Context、RAG、Memory、检索系统之间的边界。

暂不处理：

- mdBook 主题和构建配置。
- 书籍视觉样式。
- 大规模重写所有案例正文。
- 删除已有高质量内容。

---

## 8. 验收标准

结构优化完成后应满足：

1. 第一部分明确包含 Prompt Engineering、Context Engineering、Harness Engineering 三个独立深度章节。
2. `SUMMARY.md` 中不再出现承担主线职责的“专题”章节。
3. Evals、Guardrails、Tool Calling、Debug 等核心架构内容不再主要放在面试部分。
4. 第六部分只承担面试表达、题库、作品集和复盘整理。
5. Prompt、Context、Harness 三章之间递进关系清晰，不互相重复。
6. RAG、检索系统、Memory 三章边界清晰。
7. 章节编号连续，正文标题与目录一致。
8. `README.md` 的内容结构和阅读路径与新目录一致。
9. `mdbook build` 可以通过。

---

## 9. 风险与注意事项

主要风险：

- 章节编号会大幅顺延，需要系统修复交叉引用。
- 第 3 章 Context Engineering 需要补充内容最多，不能只从 Harness 章节里摘几段。
- 第 4 章 Harness Engineering 需要避免过度重复后面的 Tool、Workflow、Guardrails、Evals、Observability 章节。
- 现有 `ai-book/book/` 构建产物可能已有未提交变更，实施时应避免误覆盖用户改动。
- 如果只改目录不改导语，会导致“目录升级，正文仍像旧专题”的割裂感。

实施建议：

- 先改 `SUMMARY.md`，确认新版能力栈。
- 再处理第一部分三章，因为它们决定全书叙事。
- 接着移动第二至第四部分中被放错位置的核心章节。
- 最后调整案例和面试部分。
- 全文搜索并修复“专题：”“理论专题：”“面试表达”“下一章”“第 X 章”等关键词。
- 构建前注意当前工作区已有用户改动，避免误提交无关文件。
