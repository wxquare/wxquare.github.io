# AI工程实践：从编程到Agent的完整指南

[书籍介绍](README.md)

---

# 第一部分：AI编程实战

- [第1章 从Vibe Coding到Spec Coding：AI编程范式演进](part1/chapter1.md)
  - [1.1 AI编程工具的三次演进](part1/chapter1.md#11-ai编程工具的三次演进)
  - [1.2 Vibe Coding的本质与陷阱](part1/chapter1.md#12-vibe-coding的本质与陷阱)
  - [1.3 Spec Coding：规范驱动的工程化方法](part1/chapter1.md#13-spec-coding规范驱动的工程化方法)
  - [1.4 编写高质量Spec的完整指南](part1/chapter1.md#14-编写高质量spec的完整指南)
  - [1.5 Cursor IDE完整实践](part1/chapter1.md#15-cursor-ide完整实践)
  - [1.6 Claude Code工作流与最佳实践](part1/chapter1.md#16-claude-code工作流与最佳实践)

- [第2章 Claude Code：终端原生的AI Agent](part1/chapter2.md)
  - [2.1 Claude Code的革命性变化](part1/chapter2.md#21-claude-code的革命性变化)
  - [2.2 进阶对话技巧：让AI真正理解你](part1/chapter2.md#22-进阶对话技巧让ai真正理解你)
  - [2.3 Plan模式：先想清楚再动手](part1/chapter2.md#23-plan模式先想清楚再动手)
  - [2.4 Auto模式与安全防护](part1/chapter2.md#24-auto模式与安全防护)
  - [2.5 CLAUDE.md：Agent的宪法](part1/chapter2.md#25-claudemdagent的宪法)
  - [2.6 会话管理与上下文优化](part1/chapter2.md#26-会话管理与上下文优化)
  - [2.7 完整工作流案例](part1/chapter2.md#27-完整工作流案例)

- [第3章 Harness Engineering：驾驭AI的基础设施](part1/chapter3.md)
  - [3.1 从Prompt到Harness的范式革命](part1/chapter3.md#31-从prompt到harness的范式革命)
  - [3.2 Harness的七大核心组件](part1/chapter3.md#32-harness的七大核心组件)
  - [3.3 上下文工程：CLAUDE.md设计模式](part1/chapter3.md#33-上下文工程claudemd设计模式)
  - [3.4 验证回路与自检机制](part1/chapter3.md#34-验证回路与自检机制)
  - [3.5 架构约束与护栏设计](part1/chapter3.md#35-架构约束与护栏设计)
  - [3.6 可观测性与调试](part1/chapter3.md#36-可观测性与调试)
  - [3.7 Harness设计清单](part1/chapter3.md#37-harness设计清单)

---

# 第二部分：AI Agent系统设计

- [第4章 Agent架构设计与决策框架](part2/chapter4.md)
  - [4.1 Agent vs 传统后端系统](part2/chapter4.md#41-agent-vs-传统后端系统)
  - [4.2 主流Agent框架对比](part2/chapter4.md#42-主流agent框架对比)
  - [4.3 需求分析：何时需要Agent](part2/chapter4.md#43-需求分析何时需要agent)
  - [4.4 架构设计方法论](part2/chapter4.md#44-架构设计方法论)
  - [4.5 ReACT模式详解](part2/chapter4.md#45-react模式详解)
  - [4.6 Plan-and-Execute模式](part2/chapter4.md#46-plan-and-execute模式)
  - [4.7 状态机与混合架构](part2/chapter4.md#47-状态机与混合架构)
  - [4.8 数据流与状态管理](part2/chapter4.md#48-数据流与状态管理)

- [第5章 工具系统与MCP协议](part2/chapter5.md)
  - [5.1 Agent工具系统设计原则](part2/chapter5.md#51-agent工具系统设计原则)
  - [5.2 工具抽象与接口设计](part2/chapter5.md#52-工具抽象与接口设计)
  - [5.3 工具发现与动态注册](part2/chapter5.md#53-工具发现与动态注册)
  - [5.4 MCP协议详解](part2/chapter5.md#54-mcp协议详解)
  - [5.5 工具编排与组合](part2/chapter5.md#55-工具编排与组合)
  - [5.6 错误处理与容错](part2/chapter5.md#56-错误处理与容错)
  - [5.7 工具安全与权限控制](part2/chapter5.md#57-工具安全与权限控制)

- [第6章 多Agent协作与工作流编排](part2/chapter6.md)
  - [6.1 单Agent vs 多Agent场景](part2/chapter6.md#61-单agent-vs-多agent场景)
  - [6.2 多Agent协作模式](part2/chapter6.md#62-多agent协作模式)
  - [6.3 Agent间通信机制](part2/chapter6.md#63-agent间通信机制)
  - [6.4 工作流编排策略](part2/chapter6.md#64-工作流编排策略)
  - [6.5 冲突解决与一致性](part2/chapter6.md#65-冲突解决与一致性)
  - [6.6 协调Agent设计](part2/chapter6.md#66-协调agent设计)

- [第7章 可观测性与成本优化](part2/chapter7.md)
  - [7.1 Agent系统的可观测性挑战](part2/chapter7.md#71-agent系统的可观测性挑战)
  - [7.2 日志、指标与追踪](part2/chapter7.md#72-日志指标与追踪)
  - [7.3 LLM调用监控](part2/chapter7.md#73-llm调用监控)
  - [7.4 成本分析与优化策略](part2/chapter7.md#74-成本分析与优化策略)
  - [7.5 性能优化技巧](part2/chapter7.md#75-性能优化技巧)
  - [7.6 调试与故障排查](part2/chapter7.md#76-调试与故障排查)

---

# 第三部分：完整实战案例

- [第8章 DoD Agent：电商告警自动处理系统](part3/chapter8.md)
  - [8.1 项目背景与设计目标](part3/chapter8.md#81-项目背景与设计目标)
  - [8.2 需求分析与可行性评估](part3/chapter8.md#82-需求分析与可行性评估)
  - [8.3 架构设计：状态机+ReACT混合模式](part3/chapter8.md#83-架构设计状态机react混合模式)
  - [8.4 核心组件实现](part3/chapter8.md#84-核心组件实现)
  - [8.5 知识库与RAG集成](part3/chapter8.md#85-知识库与rag集成)
  - [8.6 工具系统设计](part3/chapter8.md#86-工具系统设计)
  - [8.7 分级决策与渐进式学习](part3/chapter8.md#87-分级决策与渐进式学习)
  - [8.8 部署与运维实践](part3/chapter8.md#88-部署与运维实践)
  - [8.9 效果评估与持续优化](part3/chapter8.md#89-效果评估与持续优化)

- [第9章 个人知识管理Agent实践](part3/chapter9.md)
  - [9.1 Karpathy的自进化知识库](part3/chapter9.md#91-karpathy的自进化知识库)
  - [9.2 知识管理系统架构](part3/chapter9.md#92-知识管理系统架构)
  - [9.3 OpenClaw框架深度解析](part3/chapter9.md#93-openclaw框架深度解析)
  - [9.4 技能系统设计](part3/chapter9.md#94-技能系统设计)
  - [9.5 记忆与上下文管理](part3/chapter9.md#95-记忆与上下文管理)
  - [9.6 多平台集成实践](part3/chapter9.md#96-多平台集成实践)
  - [9.7 从零搭建个人助手](part3/chapter9.md#97-从零搭建个人助手)

---

# 第四部分：基础理论补充

- [第10章 LLM能力边界与工程化要点](part4/chapter10.md)
  - [10.1 大语言模型概览](part4/chapter10.md#101-大语言模型概览)
  - [10.2 LLM的能力边界](part4/chapter10.md#102-llm的能力边界)
  - [10.3 推理与生成机制](part4/chapter10.md#103-推理与生成机制)
  - [10.4 模型选择与评估](part4/chapter10.md#104-模型选择与评估)
  - [10.5 Prompt工程基础](part4/chapter10.md#105-prompt工程基础)
  - [10.6 LLM工程化最佳实践](part4/chapter10.md#106-llm工程化最佳实践)

- [第11章 RAG与上下文工程基础](part4/chapter11.md)
  - [11.1 RAG系统架构](part4/chapter11.md#111-rag系统架构)
  - [11.2 文档处理与分块策略](part4/chapter11.md#112-文档处理与分块策略)
  - [11.3 向量检索与召回优化](part4/chapter11.md#113-向量检索与召回优化)
  - [11.4 上下文窗口管理](part4/chapter11.md#114-上下文窗口管理)
  - [11.5 混合检索策略](part4/chapter11.md#115-混合检索策略)
  - [11.6 RAG系统评估与优化](part4/chapter11.md#116-rag系统评估与优化)

---

# 附录

- [附录A 术语表](appendix/glossary.md)
- [附录B 参考资料与延伸阅读](appendix/references.md)
- [附录C 常用工具与框架](appendix/tools.md)
