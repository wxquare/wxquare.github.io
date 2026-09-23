---
title: 首页
layout: page
permalink: /
toc:
  enable: false
---

# AI、系统设计与工程实践

这里是 wxquare 的技术知识库，长期整理 AI Agent、系统设计、复杂业务和工程实践。

目前主要维护两本开放阅读的技术书稿：

- 《System Design Primer：系统设计与电商架构》
- 《AI Agent 工程实践：从大模型基础到生产级智能体系统》

博客文章主要用于记录书稿之外的实践、研究、工具变化和阶段性思考。

---

## 两本书

### [System Design Primer：系统设计与电商架构](/system-design-primer/)

从系统设计方法论出发，逐步进入电商系统的真实业务设计。

主要内容：

- 系统设计与架构方法论
- 编码、重构与 Code Review
- 生产系统治理与稳定性建设
- Saga、补偿、状态机与长事务
- 商品、库存、营销、计价与交易系统
- 系统设计题库与后端面试基础

适合希望系统学习后端架构、复杂业务和系统设计面试的读者。

[开始阅读 System Design Primer →](/system-design-primer/)

### [AI Agent 工程实践](/ai-book/)

从大模型基础、AI Infra 到生产级 Agent 系统，讨论 AI 能力如何被工程化、治理和落地。

主要内容：

- Transformer、预训练、后训练与推理
- 训练平台、推理服务与评估系统
- Context、Memory、Tool Calling 与 MCP
- Agent Runtime、工作流、多 Agent 与治理
- Coding Agent、企业知识助手和个人 Agent
- Agent 案例、前沿研究与工程实践

适合希望理解 AI Agent 原理，并构建可靠 AI 应用和工程系统的读者。

[开始阅读 AI Agent 工程实践 →](/ai-book/)

---

## 推荐阅读路线

### 系统设计学习路线

适合后端工程师、架构师和系统设计面试准备。

1. [系统设计与架构方法论](/system-design-primer/part01/01-system-design-guide-methodology.html)
2. [编码、重构与 Code Review](/system-design-primer/part01/02-coding-principles-design-patterns.html)
3. [生产系统治理与稳定性](/system-design-primer/part01/03-production-resilience-safeguards.html)
4. [大事务、补偿与最终一致性](/system-design-primer/part01/04-large-transaction-orchestration.html)
5. [电商系统全景图](/system-design-primer/part02/10-ecommerce-overview.html)
6. [系统设计题库](/system-design-primer/appendix/system-design-questionbank.html)

### AI Agent 学习路线

适合希望从大模型基础进入 Agent 工程实践的读者。

1. [从语言模型到 Agent](/ai-book/part1/01-llm-foundations-transformer.html)
2. [Agent 架构总纲](/ai-book/part2/01-agent-architecture.html)
3. [Context Engineering](/ai-book/part2/03-context-engineering.html)
4. [工具系统、Tool Calling 与 MCP](/ai-book/part2/06-tool-calling-mcp.html)
5. [Agent 执行编排与平台架构](/ai-book/part2/09-workflow-orchestration.html)
6. [Evals、Guardrails 与可观测性](/ai-book/part2/10-agent-evals-guardrails-observability.html)
7. [AI Agent 应用案例](/ai-book/part3/01-coding-agent-systems.html)

---

## 博客与研究动态

博客文章主要记录两本书之外的实践、研究、工具变化和阶段性思考。

### AI 与 Agent

- [AI Agent 工作流实践：从 Claude Code、Codex 到 OpenClaw、Hermes 与 DeepSeek](/2026/09/23/AI/01-ai-agent-workflow-practice/)
- [电商 Agent 项目全景调研报告](/2026/09/11/other/08-agent-electronic-commerce-research-report/)
- [从 Vibe Coding 到 Spec Coding](/2026/04/03/AI/00-vibe-coding-vs-spec-coding/)
- [Karpathy 的自我进化知识库](/2026/04/05/AI/02-karpathy-evolving-knowledge-base/)
- [从内容到短视频：用开源工具生成 AI 教程视频](/2026/05/08/other/ai-content-to-video-open-source-workflow/)
- [AI Agent 系统设计内容已整合至书稿](/ai-book/part2/01-agent-architecture.html)

### 系统设计与工程实践

- [系统设计完全指南](/2026/03/07/system-design/00-system-design-overview/)
- [架构与整洁代码：DDD、Clean Architecture 与 CQRS](/2026/04/01/system-design/41-acc-clean-arch-ddd-cqrs/)
- [领域驱动设计读书笔记](/2026/04/03/system-design/43-acc-ddd-notes/)
- [商品生命周期管理](/2026/04/10/system-design/30-ecommerce-product-lifecycle-management/)
- [价格日历系统设计](/2026/04/16/system-design/33-ecommerce-price-calendar/)
- [核心业务长事务怎么处理](/2026/06/09/system-design/34-ecommerce-long-transactions/)

### AI 算法与性能优化

- [TVM 算子优化实战](/2026/09/22/AI/05-tvm-operator-optimization-practice/)
- [TensorFlow 模型压缩与推理优化实战](/2026/09/22/AI/04-tensorflow-model-optimization/)
- [OpenCL、OpenCV UMat 与 KCF 性能测试](/2020/08/13/AI/03-opencl-mobile-performance-testing/)
- [视频目标追踪流程](/2020/08/13/AI/06-video-object-tracking/)

[浏览全部博客文章](/archives/) · [按分类浏览](/categories/) · [按标签浏览](/tags/)

---

## 其他项目与资料

### 找工作与面试准备

- [系统设计题库](/system-design-primer/appendix/system-design-questionbank.html)
- [后端面试基础知识题单](/system-design-primer/appendix/interview-basic-question-bank.html)
- [系统设计面试高频 50 题](/2026/04/07/system-design/08-system-design-interview/)
- [AI 与 Agent 开发高频 50 题](/ai-book/appendix/ai-agent-development-interview-50.html)
- [LeetCode Primer](https://github.com/wxquare/leetcode-primer)
- [AI 与 AI Agent 开发高频 50 题（旧版归档）](/archive/AI/07-ai-agent-development-interview-50.html)
- [LeetCode 500 精选题单](https://github.com/wxquare/leetcode-primer/blob/master/README.md)

### 其他项目

- [本站源码](https://github.com/wxquare/wxquare.github.io)
- [书籍与学习资料](/booklist/)
- [友情链接](/friends/)

### 其他资料

- [全部博客归档](/archives/)
- [文章分类](/categories/)
- [文章标签](/tags/)
- [阅读书单](/booklist/)

---

## 关于本站

本站的内容分为三层：

```text
两本书稿
  └── 系统化、稳定、完整的知识体系

博客文章
  └── 新实践、研究报告、工具变化和阶段性思考

项目与资料
  └── 面试题库、开源项目、书单和延伸阅读
```

内容会持续更新和重构。对于变化较快的 AI 工具、模型和框架，请结合文章更新时间和官方文档阅读。

---

## 联系与反馈

- [GitHub](https://github.com/wxquare)
- [本站源码](https://github.com/wxquare/wxquare.github.io)

---

## 实验与私有项目

- **SkillForge**：面向可复用 Agent Skills 的工具与工作流探索
- **Investment Assistant**：投资研究、信息整理与辅助决策实验
- **wxquare-private**：个人研究、实验与协作资产的私有工作区
