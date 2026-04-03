# 文章规范：从 Vibe Coding 到 Spec Coding - AI 编程范式的演进与实践

## 文章元信息

**标题：** 从 Vibe Coding 到 Spec Coding：AI 编程范式的演进与实践  
**日期：** 2026-04-03  
**分类：** AI  
**标签：** ai-programming, cursor, claude-code, spec-coding, vibe-coding, best-practices  
**文件名：** 24-vibe-coding-vs-spec-coding.md  
**目标位置：** `/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/source/_posts/AI/`

## 目标读者

已在使用 Cursor/Claude Code 的后端开发者，特别是有电商或系统开发背景的工程师。

## 文章定位

- **类型：** 实战指南型
- **深度：** 理论+实践平衡（4000-6000 字）
- **核心观点：** Spec Coding 是生产环境必需品，Vibe Coding 是探索工具
- **阅读时间：** 20-25 分钟

## 文章结构（7 个部分）

### 第 1 部分：AI 编程工具的演进（500 字）

**内容要点：**
- 从 GitHub Copilot（2021）到 Cursor/Claude Code（2024-2025）的演进
- 工具能力三阶段：代码补全 → 多文件编辑 → Agent 自主执行
- 引出核心问题：工具能力提升了，但开发方法论跟上了吗？

**必须包含：**
- Mermaid 时间线图表展示 AI 编程工具演进
- 为后续 Vibe vs Spec 讨论做铺垫

---

### 第 2 部分：Vibe Coding - 第一代 AI 编程范式（600 字）

**内容要点：**
- 定义：即兴式 prompt，逐步迭代，"感觉对了就继续"
- 为什么它是自然选择：符合人类思维，上手门槛低
- 快速原型案例：用 Vibe Coding 5 分钟实现简单 REST API

**必须包含：**
- Go 语言实战示例（用户列表 API）
- 展示逐步迭代的过程（3-4 轮对话）
- Vibe Coding 的价值：探索、验证、学习

---

### 第 3 部分：Vibe Coding 的天花板（800 字）

**内容要点：**
- 通过 3 个真实场景展示局限性

**场景 1：订单状态机实现**
- Vibe 方式的问题：逐步添加状态，最后逻辑混乱
- 缺少整体规划，状态机不完整

**场景 2：支付接口集成**
- Vibe 方式的问题：先实现功能，后发现幂等性、重试、对账都没考虑
- 安全和可靠性要求容易被遗忘

**场景 3：数据库设计**
- Vibe 方式的问题：边写边改表结构，索引、外键、事务都有问题
- 缺少架构层面思考

**必须包含：**
- 量化数据：首次通过率、Bug 密度、重构频率
- 代码示例对比（简短）

---

### 第 4 部分：Spec Coding - 规范驱动的 AI 编程（900 字）

**内容要点：**

**理论基础：**
- Sean Grove "The New Code" 核心思想
  - "Code is a lossy projection of intent"
  - 规范才是 source of truth
- Robert C. Martin：精确描述需求就是编程

**Spec Coding 三层规范：**
1. 功能规范（What）：用户故事、验收标准
2. 架构规范（How - 语言无关）：数据模型、API、安全
3. 实现规范（How - 语言特定）：技术栈、代码规范、测试

**核心工作流：**
```
Specify → Plan → Tasks → Implement → Verify
```

**必须包含：**
- Mermaid 流程图
- 为什么适合生产环境（5 个要点）

---

### 第 5 部分：完整案例对比（1200 字）

**案例需求：电商库存扣减服务**

功能需求：
- 下单时扣减库存
- 超卖保护
- 订单取消时恢复库存
- 支持分布式并发

**5.1 Vibe Coding 实现过程**
- 展示 5-6 轮迭代过程
- 每轮发现的问题
- 总耗时、返工次数、代码质量

**5.2 Spec Coding 实现过程**
- 步骤 1：编写规范文档（完整示例）
- 步骤 2：AI 生成技术方案
- 步骤 3：逐步实现
- 步骤 4：验证
- 总耗时、返工次数、代码质量

**5.3 对比总结**
- 表格对比 7 个维度
- 关键洞察：前期投入 vs 后期收益

**必须包含：**
- 完整的规范文档示例（Markdown 格式）
- Go 代码示例（关键部分）
- 量化对比表格

---

### 第 6 部分：Spec Coding 工具链实践（1500 字）

#### 6.1 Cursor IDE 中的 Spec Coding 实践

**6.1.1 Cursor 配置文件体系**

**工具 1：`.cursorrules`**
- 位置：项目根目录
- 加载：Cursor 启动时自动加载
- 完整示例（电商后端项目）

**工具 2：`.cursor/rules/`**
- 位置：`.cursor/rules/` 目录
- 加载：通过 globs 条件加载
- 目录结构示例
- 完整示例：`10-api-design.md`、`11-database.md`

**6.1.2 Cursor 工作流实战**
- 4 个步骤：创建规范 → 引用规范 → Composer 编辑 → 验证
- 每个步骤的具体命令

**6.1.3 Cursor 常见陷阱**
- 3 个陷阱及解决方案（对比格式）

#### 6.2 Claude Code 中的 Spec Coding 实践

**6.2.1 Claude Code 配置文件体系**

**工具 1：`CLAUDE.md`**
- 位置：项目根目录
- 加载：Claude Code 启动时自动加载
- 完整示例（电商后端项目）

**工具 2：规范文档目录**
- 目录结构：`CLAUDE.md` + `docs/specs/` + `specs/`
- 如何在 CLAUDE.md 中引用其他规范

**6.2.2 Claude Code 工作流实战**
- 5 个步骤：创建规范 → 启动 Claude Code → /plan 命令 → 逐步实现 → 验证
- 终端命令示例

**6.2.3 Claude Code 特色功能**
- 规范验证
- 规范演进
- 命令示例

#### 6.3 Cursor vs Claude Code 对比

- 对比表格（7 个维度）
- 推荐使用场景
- 两者结合策略

**必须包含：**
- 完整的配置文件示例
- 实际操作命令
- 目录结构图

---

### 第 7 部分：混合策略与行动建议（500 字）

#### 7.1 渐进式工作流
- 3 个阶段：Vibe 探索 → 提炼规范 → Spec 重建
- 每个阶段的时间和目标

#### 7.2 决策框架
- 什么时候用 Vibe Coding（4 个场景）
- 什么时候用 Spec Coding（6 个场景）
- Mermaid 决策树图表

#### 7.3 行动建议
- 个人开发者（4 条建议）
- 团队负责人（5 条建议）
- 关键原则（4 条）

#### 7.4 常见问题
- 4 个 FAQ 及回答

**必须包含：**
- Mermaid 决策树
- 具体可执行的建议

---

### 第 8 部分：总结（300 字）

**内容要点：**
- 回顾核心观点
- 5 个关键要点
- 最终建议（4 条）
- 行动号召

---

## 写作规范

### Front Matter
```yaml
---
title: 从 Vibe Coding 到 Spec Coding：AI 编程范式的演进与实践
date: 2026-04-03
categories:
  - AI
tags:
  - ai-programming
  - cursor
  - claude-code
  - spec-coding
  - vibe-coding
  - best-practices
toc: true
---
```

### 格式要求
- 中英文之间有空格
- 代码块必须指定语言
- 使用中文标点
- 标题层级：# 只用一次，## 和 ### 组织内容
- 专业术语首次出现时加注释

### 代码示例要求
- 使用 Go 语言（符合读者背景）
- 代码必须完整可运行
- 关键部分有中文注释
- 代码块长度控制在 30 行以内

### 图表要求
- 使用 Mermaid 格式
- 3 个必需图表：
  1. AI 工具演进时间线
  2. Spec Coding 工作流程
  3. 决策树

### 引用规范
- 引用 Sean Grove 演讲
- 引用 Easy-Vibe 教程
- 引用 Robert C. Martin
- 提供完整的参考资料列表

## 验收标准

- [ ] 文章长度：5000-6000 字
- [ ] 包含所有 7 个部分
- [ ] 至少 3 个 Mermaid 图表
- [ ] 至少 5 个代码示例（Go 语言）
- [ ] 完整的配置文件示例（.cursorrules 和 CLAUDE.md）
- [ ] 完整的规范文档示例（库存扣减服务）
- [ ] 对比表格（Vibe vs Spec，Cursor vs Claude Code）
- [ ] 决策框架和行动建议
- [ ] 参考资料列表
- [ ] Front Matter 完整
- [ ] 所有代码块指定语言
- [ ] 中英文之间有空格
- [ ] 专业术语有注释

## 成功标准

读完文章后，读者应该能够：

1. 理解 Vibe Coding 和 Spec Coding 的本质区别
2. 知道什么场景用什么方法
3. 在 Cursor 或 Claude Code 中配置 Spec Coding 工具链
4. 编写一份合格的功能规范文档
5. 用 Spec Coding 方式完成一个生产级功能
6. 建立团队的规范管理体系

## 参考资料

- Sean Grove "The New Code" 演讲
- Easy-Vibe 教程：https://datawhalechina.github.io/easy-vibe/zh-cn/stage-3/core-skills/spec-coding/
- Robert C. Martin《Clean Code》
- Cursor 官方文档
- Claude Code 官方文档
