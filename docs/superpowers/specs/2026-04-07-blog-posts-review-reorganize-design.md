# 技术博客文档 Review 与重组设计

**日期**: 2026-04-07
**状态**: Approved
**范围**: 14 篇面试/工作经验文档的全面重组与质量统一

---

## 1. 背景

### 1.1 现状

博客 `system-design/` 目录下积累了 14 篇面试和工作踩坑文档，覆盖计算机基础、编程语言、中间件、架构设计、系统可靠性等主题。这些文档跨越较长时间积累，存在以下问题：

- **目录混乱**：编程语言基础（OS、网络、bash、Python、C++、Go）和系统设计文档混放在同一目录
- **质量参差不齐**：新文档（08-cache-redis、14-system-reliability、17-high-frequency）质量高，旧文档（01-06、09）质量低
- **格式不统一**：大多数文档缺少 `tags` 字段，违反项目 `.cursorrules` 规范
- **文件名问题**：拼写错误（`7-storage-desgin.md`）、大小写不一致（`9-kafka-MQ.md`）
- **内容问题**：`04-python` 有多个 Front Matter；`12-tech-design` 过于臃肿（1400 行）；`09-kafka` 过于单薄（173 行）
- **适度重复可接受**：专题文章讲原理深度，`17-high-frequency` 作为面试速查保留精简版

### 1.2 涉及文档清单

| 编号 | 原文件名 | 行数 | 质量评级 |
|------|----------|------|----------|
| 00 | `00-system-design-fundamentals.md` | 333 | 中（导航页，需更新链接） |
| 01 | `1-OS-interview.md` | 79 | 低 |
| 02 | `2-Internet-interview.md` | 392 | 低 |
| 03 | `3-bash-shell.md` | 121 | 低 |
| 04 | `4-python-experience.md` | 451 | 低（多 Front Matter） |
| 05 | `5-cpp-interview.md` | 750 | 中 |
| 06 | `6-golang-interview.md` | 1408 | 中 |
| 07 | `7-storage-desgin.md` | 1371 | 中（文件名拼写错误） |
| 08 | `8-cache-redis.md` | 1806 | 高（质量标杆） |
| 09 | `9-kafka-MQ.md` | 173 | 低（内容过少） |
| 10 | `10-elasticsearch.md` | 708 | 中 |
| 11 | `11-k8s-docker.md` | 245 | 中 |
| 12 | `12-tech-design.md` | 1399 | 中（过于臃肿） |
| 14 | `14-system-reliability.md` | 1541 | 高 |
| 17 | `17-high-frequency-system-design.md` | 1805 | 高 |

### 1.3 目标

- 自用优先，兼顾公开发布
- 全面改造：格式统一 + 内容重组 + 质量提升
- 编程语言/基础类文档迁移到独立目录
- 允许专题与速查之间的适度内容重复

### 1.4 不在本次范围

- `30-clean-architecture-ddd-cqrs.md`、`31-clean-code.md`、`32-architecture-checklist.md`（已有独立设计 spec）
- `source/_posts/AI/` 和 `source/_posts/other/` 目录
- 主题配置和 Hexo 设置

---

## 2. 设计方案

### 2.1 目录结构重组

```
source/_posts/
├── fundamentals/                          # 新目录：计算机基础 & 编程语言
│   ├── 1-os-fundamentals.md              ← 原 1-OS-interview.md
│   ├── 2-network-fundamentals.md         ← 原 2-Internet-interview.md
│   ├── 3-bash-shell.md                   ← 原 3-bash-shell.md
│   ├── 4-python-practice.md              ← 原 4-python-experience.md
│   ├── 5-cpp-practice.md                 ← 原 5-cpp-interview.md
│   └── 6-golang-practice.md              ← 原 6-golang-interview.md
│
└── system-design/                         # 聚焦系统设计
    ├── 00-system-design-fundamentals.md   # 导航页
    ├── 07-storage-mysql.md               ← 原 7-storage-desgin.md
    ├── 08-cache-redis.md                 # 保留
    ├── 09-kafka-mq.md                    ← 原 9-kafka-MQ.md
    ├── 10-elasticsearch.md               # 保留
    ├── 11-k8s-docker.md                  # 保留
    ├── 12-tech-design.md                 # 拆分精简
    ├── 14-system-reliability.md          # 保留
    └── 17-high-frequency-system-design.md # 保留
```

### 2.2 文件名变更映射

| 原文件名 | 新文件名 | 新目录 | 变更原因 |
|----------|----------|--------|----------|
| `1-OS-interview.md` | `1-os-fundamentals.md` | fundamentals/ | 去 interview，小写统一 |
| `2-Internet-interview.md` | `2-network-fundamentals.md` | fundamentals/ | 更准确名称 |
| `3-bash-shell.md` | `3-bash-shell.md` | fundamentals/ | 仅迁移 |
| `4-python-experience.md` | `4-python-practice.md` | fundamentals/ | 统一 practice |
| `5-cpp-interview.md` | `5-cpp-practice.md` | fundamentals/ | 统一 practice |
| `6-golang-interview.md` | `6-golang-practice.md` | fundamentals/ | 统一 practice |
| `7-storage-desgin.md` | `07-storage-mysql.md` | system-design/ | 修正拼写 + 补零 |
| `9-kafka-MQ.md` | `09-kafka-mq.md` | system-design/ | 小写统一 + 补零 |

### 2.3 每篇文档改造计划

#### 2.3.1 fundamentals/ — 格式修复 + 轻度内容优化

| 文件 | 改造内容 |
|------|----------|
| `1-os-fundamentals.md` | 补 tags；修复重复编号（两个 7.）；修正"进bai程"编码乱码；整理 Q&A 结构 |
| `2-network-fundamentals.md` | 补 tags；修复重复编号（两个 14.）；修正"小路"笔误；合并 HTTP/HTTPS 散落内容 |
| `3-bash-shell.md` | 补 tags；修正 typo（comamand1, taf→tar）；移除 `<font>` 标签 |
| `4-python-practice.md` | 补 tags；删除多余第二段 Front Matter（~line 39-43）；修正 JIL→JIT；统一 categories |
| `5-cpp-practice.md` | 补 tags；检查代码块语言标注 |
| `6-golang-practice.md` | 补 tags；移除 `<font>` 标签；修复开头空行 |

所有 6 篇统一操作：
- Front Matter 补充 `tags` 和 `toc: true`
- `categories` 改为对应分类（"计算机基础" 或 "编程语言"）
- 确保代码块都标注语言

#### 2.3.2 system-design/ — 分三档处理

**第一档：微调（已高质量）**

| 文件 | 改造内容 |
|------|----------|
| `08-cache-redis.md` | 基本不动，检查 emoji 渲染 |
| `14-system-reliability.md` | 修复"字节跳动 SRE 实践"占位链接（当前是 `xxx`） |
| `17-high-frequency-system-design.md` | 基本不动 |

**第二档：中度改造**

| 文件 | 改造内容 |
|------|----------|
| `00-system-design-fundamentals.md` | 更新迁移后的内部导航链接；精简虚构"学员评价"；更新学习路径文章链接 |
| `07-storage-mysql.md` | 修正文件名；补 tags；纯链接罗列改为简短描述 + 链接；补充建表规范代码示例 |
| `10-elasticsearch.md` | 补 tags；验证 `<details>` 渲染；补充 DSL 查询示例 |
| `11-k8s-docker.md` | 补 tags；修正 iptabels→iptables |

**第三档：重度改造**

| 文件 | 改造内容 |
|------|----------|
| `09-kafka-mq.md` | 大幅扩充至 600-800 行：补充 Kafka 架构图说明、消费者组原理、Exactly-once 语义、性能调优、Go 代码示例；修复 GitHub blob 图片链接；删除空标题 |
| `12-tech-design.md` | 拆分精简：保留"技术方案写作方法论"和"架构设计模式"；删除与 14 重复的"系统稳定性建设"章节；将"中间件和存储"精简为链接索引（各专题已有详细内容） |

---

## 3. 统一质量标准

### 3.1 Front Matter 模板

```yaml
---
title: [中文标题]
date: YYYY-MM-DD
categories:
- [主分类]
tags:
- [tag1]
- [tag2]
- [tag3]
toc: true
---
```

分类映射：
- `fundamentals/` 的 OS、网络 → `计算机基础`
- `fundamentals/` 的编程语言 → `编程语言`
- `system-design/` 全部 → `系统设计`

### 3.2 文档结构标准

```markdown
<!-- toc -->

导读段落：1-3 句话说明本文覆盖什么、适合谁、怎么用。

## 一、[主题章节]
### 1. [子主题]
（内容 + 代码示例 + 表格对比）

## 二、[主题章节]
...

## 参考资料
- [标题](链接) - 一句话说明
```

### 3.3 代码规范

- 所有代码块必须标注语言（`go`、`sql`、`redis`、`bash`、`text`）
- Go 代码遵循 idiomatic Go 风格
- 移除所有 HTML 标签（`<font>`），改用标准 Markdown
- 图片保留 `<p align="center">` 格式（Hexo NexT 主题需要）

### 3.4 链接规范

- 内部链接使用 Hexo 路径格式
- 外部链接确保可访问
- 每篇文末"参考资料"集中管理外部链接

---

## 4. 执行计划

分 3 个阶段，每阶段独立可交付：

### P0：基础设施（优先级最高）

| 任务 | 涉及文件 | 说明 |
|------|----------|------|
| 创建 `fundamentals/` 目录 | — | 新建目录 |
| 迁移并重命名 6 篇基础文档 | 01-06 | 迁移到 fundamentals/ + 改名 |
| 修复所有 14 篇 Front Matter | 全部 | 补 tags、toc、修正 categories |
| 修正文件名拼写 | 07, 09 | desgin→design, MQ→mq |

### P1：中度改造

| 任务 | 涉及文件 | 说明 |
|------|----------|------|
| 更新导航页链接 | 00 | 适配迁移后路径 |
| 修复 typo 和格式问题 | 07, 10, 11 | tags + typo |
| 拆分精简 12-tech-design | 12 | 删除重复内容，聚焦方案写作 |
| 修复占位链接 | 14 | 字节跳动 SRE 链接 |

### P2：重度改造

| 任务 | 涉及文件 | 说明 |
|------|----------|------|
| 扩充 09-kafka-mq | 09 | 173 行 → 600-800 行 |
| fundamentals 内容优化 | 01-06 | 修复编号、乱码、结构整理 |

### 构建验证

每个阶段完成后执行：

```bash
npm run clean && npm run build
```

确保无构建错误后才提交。

---

## 5. 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 文件迁移导致内部链接失效 | 已发布页面 404 | P0 阶段同步更新所有内部引用链接 |
| 文件重命名导致 Git 历史断裂 | 无法追溯修改历史 | 使用 `git mv` 保留历史 |
| 12 拆分后内容遗漏 | 有用信息丢失 | 拆分前先备份，拆分后检查完整性 |
| 09 扩充质量不达标 | 与高质量文档风格不一致 | 以 08-cache-redis 为模板扩写 |
| Hexo 构建兼容性 | 迁移后渲染异常 | 每阶段构建验证 |

---

## 6. 成功标准

- [ ] 所有 14 篇文档 Front Matter 完整（title, date, categories, tags, toc）
- [ ] 所有代码块标注语言
- [ ] 无文件名拼写错误
- [ ] `fundamentals/` 目录包含 6 篇基础文档
- [ ] `system-design/` 目录聚焦架构和设计
- [ ] `09-kafka-mq` 行数 ≥ 600
- [ ] `12-tech-design` 行数 ≤ 800（拆分后）
- [ ] `npm run clean && npm run build` 无错误
- [ ] 所有内部链接可访问
