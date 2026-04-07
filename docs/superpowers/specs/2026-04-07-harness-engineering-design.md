# Harness Engineering 博客文章设计规格

## 概述

- **文件名**: `06-harness-engineering.md`
- **路径**: `source/_posts/AI/06-harness-engineering.md`
- **系列位置**: AI 系列第 06 篇，承接 02-Agent 系统设计、03-DoD Agent 设计
- **定位**: 综合型（概念 + 实践），以演进叙事为主线
- **篇幅**: 5000-8000 字
- **案例来源**: 行业公开案例 + 个人实践经验

## Front Matter

```yaml
---
title: Harness Engineering：从驾驭 Prompt 到驾驭 Agent 的范式革命
date: 2026-04-07
categories:
  - AI
tags:
  - harness-engineering
  - ai-agent
  - context-engineering
  - prompt-engineering
  - cursor
  - claude-code
  - best-practices
toc: true
---
```

## 开篇引用

> "Agents aren't hard; the Harness is hard." —— Ryan Lopopolo, OpenAI

## 文章结构

### 前言（~500 字）

- **场景引入**: 用 Cursor/Claude Code 让 Agent 完成一个功能，第一次跑通，换个场景就崩。优化 Prompt 效果提升不到 3%。
- **核心公式**: Agent = Model + Harness
- **文章脉络**: 从 Prompt Engineering 到 Harness Engineering 的三次范式跃迁
- **系列衔接**: 呼应 00-Vibe/Spec Coding、01-Claude Code、02-Agent 系统设计，Harness Engineering 是这条认知升级线的最新节点

### 一、三次范式跃迁（~1500 字）

包含一张 Mermaid timeline 图展示三阶段演进。

#### 1.1 Prompt Engineering（2023-2024）

- 核心关注: 如何措辞指令
- 局限: 模型越强，Prompt 技巧的边际收益递减（<3%）
- 类比: 教人做事只靠口头说

#### 1.2 Context Engineering（2025）

- 核心关注: 给模型什么信息，CLAUDE.md / AGENTS.md 的出现
- 进步: 从"说什么"升级到"知道什么"
- 局限: 上下文是必要条件，但不充分——给对了信息，Agent 还是可能跑偏
- 类比: 教人做事不仅说了怎么做，还给了参考资料

#### 1.3 Harness Engineering（2026）

- 核心关注: 构建约束、反馈、验证的完整基础设施
- 核心洞察: 模型质量差异只有 10-15%，Harness 决定系统是否能用
- LangChain 实证: 仅改 Harness，准确率从 52.8% → 66.5%
- 类比: 不只是教人做事，而是设计整个工作环境——工具、流程、检查点、安全栏

章节末包含一张对比表格，从"关注点、核心产出、效果上限、工程师角色"四个维度对比三者。

### 二、Harness 的核心组件（~2500 字）

六大核心组件，每个组件结构: 概念 → 行业做法 → 个人实践。

#### 2.1 上下文工程（Context Engineering）

- Harness 的信息层: 项目地图、规范、约束
- OpenAI 实践: AGENTS.md 控制在 100 行以内，作为目录而非百科全书
- **个人实践**: 展示 `.cursorrules` 和 `CLAUDE.md` 的设计思路，踩过的坑，精简策略

#### 2.2 架构约束（Architectural Constraints）

- 核心洞察: 限制 Agent 自由度反而提高产出质量
- OpenAI 实践: 严格依赖分层 Types → Config → Repo → Service → Runtime → UI
- Stripe 实践: Minions Agent 用 linter + pre-commit hook 机械化执行规范
- **个人实践**: 项目中通过 linter、CI 检查约束 Agent 行为

#### 2.3 工具编排（Tools & MCP Servers）

- Agent 通过工具与外界交互，工具决定能力边界
- MCP（Model Context Protocol）统一工具接口
- **个人实践**: Cursor 中配置 MCP Server 的经验

#### 2.4 验证回路（Verification Loops）

- Agent 不能自我评估——它会给自己打高分
- 三层验证: 确定性工具（测试/linter）→ AI 驱动验证 → 人工审核
- **个人实践**: DoD Agent 中的验证机制设计（关联 03 号文章）

#### 2.5 子代理与上下文防火墙（Sub-Agents & Context Firewalls）

- 长会话导致上下文退化（context rot）
- 解法: 为每个子任务 spawn 新的子代理，保持上下文清洁
- **个人实践**: Claude Code 中的 Plan → 新会话执行模式（关联 01 号文章）

#### 2.6 可观测性与熵管理（Observability & Entropy Management）

- Agent 产出的代码会随时间积累技术债（代码熵）
- 需要"垃圾回收代理"定期清理
- 行业实践: OpenAI 的持续清理机制

### 三、行业实证（~1000 字）

#### 3.1 OpenAI：百万行零手写代码

- 3-7 人团队，5 个月，100% Agent 生成
- 关键不是模型，而是 Harness 的设计

#### 3.2 LangChain：Deep Agents 的 Harness 优化

- Terminal Bench 2.0 上的实验
- 不换模型，只改 Harness: 52.8% → 66.5%，排名从 30 开外进前 5
- 具体改了什么: 自验证回路、循环检测中间件、主动上下文工程

#### 3.3 Anthropic：三代理架构

- Planner → Generator → Evaluator
- 成本提升 22 倍，但功能完整度质变

### 四、工程师角色的转变（~500 字）

- 从"写代码的人"到"设计环境的人"
- 新的核心能力: 架构设计、文档工程、Agent 行为分析、反馈系统设计
- 与系列文章的呼应: 00 号讲 Spec Coding（规范驱动），Harness Engineering 是这个思想的系统级延伸

### 五、总结与展望（~300 字）

- Harness Engineering 的一句话总结
- 最小可行 Harness 检查清单（5-7 条，读者可以立即行动）:
  1. 精简的 AGENTS.md/CLAUDE.md 入口文件（<100 行）
  2. 可复现的开发环境（一键启动）
  3. 机械化的架构约束（linter + CI）
  4. 自验证回路（测试必须通过才能提交）
  5. Agent 可观测性（结构化日志、trace）
  6. 最小权限凭证 + 回滚能力
  7. 定期的代码熵清理
- 展望: Harness 将成为工程团队的核心竞争力

### 参考资料

列出所有引用链接:
- OpenAI Engineering Blog: Harness engineering: leveraging Codex in an agent-first world
- LangChain Blog: Improving Deep Agents with harness engineering
- LangChain Blog: The Anatomy of an Agent Harness
- Harrison Chase: Why Better LLMs Aren't Enough
- dev.to: Prompt Engineering vs Context Engineering vs Harness Engineering
- dev.to: Prompt Engineering Is Dead. Harness Engineering Is What Actually Works.
- dev.to: Beyond AGENTS.md: Harness Engineering, Loop-Based Delivery
- harness-engineering.ai: What Is Harness Engineering?
- gtcode.com: Harness Engineering: The Discipline of Building Systems That...
- InfoQ: OpenAI Introduces Harness Engineering
- 腾讯云: Agent 系列（三）：Harness Engineering
- 晨涧云: Harness Engineering：从驾驭百万行AI代码到软件工程的范式革命

## 图表需求

1. **Mermaid Timeline**: 三次范式跃迁（Prompt → Context → Harness）
2. **Mermaid Flowchart**: Harness 核心组件架构图（六大组件的关系）
3. **Markdown Table**: 三种范式的对比表格
4. **Markdown Table**: 最小可行 Harness 检查清单

## 写作风格要求

- 与系列文章保持一致: 正式技术中文，产品/工具名用英文
- 中英文之间有空格
- 代码块指定语言
- 英文术语首次出现时加中文注释
- 每章有实际案例，避免纯理论
- 类比生动（马具隐喻贯穿全文但不过度）
