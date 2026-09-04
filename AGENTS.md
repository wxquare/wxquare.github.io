# AGENTS.md

本文件是仓库内唯一的 AI 协作入口和规范源。与 AI 开发指导、协作约束、工具工作流相关的规则，统一维护在这里，不再拆散到 `CLAUDE.md`、`README-AI-SETUP.md`、`README-CURSOR-USAGE.md`。其他入口只能引用本文件或执行其中定义的流程，不得复制规范。

## 1. 项目定位

- 仓库：`wxquare.github.io`
- 类型：Hexo 技术博客与知识库
- 主题：系统设计、后端工程、AI 与 Agent 工程实践
- 主题配置：`hexo-theme-next`
- 部署方式：GitHub Pages

## 2. 快速开始

在仓库根目录执行：

```bash
npm install
npm run server
npm run clean
npm run build
npm run build:books
```

本地预览地址：

```text
http://localhost:4000
```

## 3. 源码边界

只修改源码，不直接修改生成产物。

### 3.1 可编辑目录

- `source/_posts/AI/`：AI 与 Agent 相关文章
- `source/_posts/system-design/`：系统设计文章
- `source/_posts/fundamentals/`：计算机基础文章
- `source/_posts/other/`：其他文章
- `source/about/`：最小联系页，仅允许邮箱；不得存放简历、电话或工作资料
- `source/diagrams/`：图表源文件
- `source/library/`：公开且适合公开保存的第三方参考资料与来源目录
- `source/presentations/`：本人创作或经用户明确确认为第一方且可公开的演示资料；每份资料只有一个 Git 源载荷
- `books/ai-book/src/`：AI Agent 书稿源码
- `books/ai-book/labs/llm-from-scratch/`：唯一保留的遗留实验例外；不得在 `books/ai-book/labs/` 新增其他实验
- `docs/`：调研、迁移、整理文档
- `.agents/`：统一的 AI 协作资产与工具配置目录

博客主分类规范在本文件第 4.4 节定义：`AI`、`system-design`、`fundamentals`、`other`。`.agents/config/post-categories.json` 是供工具读取的非规范目录映射；新增或统计文章时复用它，目录 slug 与 Front Matter 展示名可能不同，以本文件和映射中的 `label`、`frontMatterLabels` 为准。

### 3.2 不要直接编辑的目录

- `public/`
- `.deploy_git/`
- `books/ai-book/book/`
- `source/booklist/`：已阻塞的只读遗留目录；逐条书目来源与元数据完成前保持原位，不得新增文件
- `source/pdf/`：不含活动源文件的遗留目录；`k8s-network.pdf` 已位于 `source/presentations/`，旧 URL 仅由构建期别名输出兼容，不得新增文件

## 4. 内容落点规则

每类内容只保留一个主事实源：

- 系统化 Agent 知识：`books/ai-book/src/`
- 面向读者的博客文章：`source/_posts/AI/`、`source/_posts/system-design/`、`source/_posts/fundamentals/`、`source/_posts/other/`
- 公开第三方参考资料：`source/library/`
- 第一方公开演示资料：`source/presentations/`
- AI 协作规则、共享技能与工具配置：`AGENTS.md`、`.agents/`
- 内部整理文档与迁移说明：`docs/`

不要把同一份长文同时维护在博客、书稿和 `docs/` 三处。

### 4.1 公开资料库治理

- books、papers、slides、tutorials、other 分别放入 `source/library/books/`、`papers/`、`slides/`、`tutorials/`、`other/`；博客草稿、个人或内部材料、秘密凭证和实验源码不得进入资料库。
- tutorials 用于公开教程、实验指南或课程型技术资料；它不是博客草稿或实验源码的存放目录。
- Other 仅用于经用户明确批准的公开项目补充文件，或类型边界明确但无法归入 book、paper、slide、tutorial 的参考资料；不得作为授权不明、来源不明或未分类内容的兜底，也不得绕过来源、再分发和用户批准要求。
- 每个条目必须在 `source/library/index.md` 记录标题、作者或机构、类型（book、paper、slide、tutorial 或 other）、主题、原始 URL、本地路径或“仅外链”、再分发说明和加入日期。
- 公开可访问不等于允许重新分发；授权或再分发边界无法确认时只保留原始链接，不上传本地副本。
- 单个二进制文件超过 10 MiB 时默认只保留外链；本地托管需要用户单独批准存储方案。
- 公开活动资料只在 `source/library/` 维护；其他文章需要引用时链接到资料库条目或原始来源，不复制活动副本。
- `source/booklist/` 保持阻塞和只读；其内容只能在权威 URL 与必填元数据齐全后逐条分解到资料库索引，逻辑索引目标不得作为 `git mv` 目标。
- `source/pdf/` 不含活动源文件；已批准的第一方演示资料位于 `source/presentations/`，旧 URL 由已验证的构建期别名兼容。

完整规则见 `docs/library-policy.md`。

### 4.2 第一方公开演示资料边界

- `source/presentations/` 只接收本人创作，或经用户明确确认为第一方且可公开的演示资料；第三方 slide 仍按来源与再分发规则进入 `source/library/slides/` 或仅保留外链。
- 资料不得包含密码、Token、API Key、私钥、个人隐私、公司内部、客户或未公开工作材料。
- 每份演示资料只保留一个 Git 跟踪的规范源文件。兼容旧 URL 时，Hexo 只能在构建阶段从规范源生成同字节输出；构建产物不是并行活动源。
- 禁止为兼容 URL 提交第二份源载荷、符号链接，或用 HTML 内容伪装 PDF 等原始格式。

### 4.3 实验仓库边界

- 新增可运行实验必须放在 `/Users/xianguiwang/Projects/<project>/`，一个实验对应一个独立 Git 仓库，GitHub 仓库默认设为 Private。
- `source/library/` 和博客文章不得包含实验源码；文章只保存实验介绍和仓库链接。
- `books/ai-book/labs/llm-from-scratch/` 是唯一保留的遗留例外；不得把 `books/ai-book/labs/` 作为新实验落点，也不得在其中恢复或新增其他实验。

### 4.4 博客文章规范（唯一来源）

以下规则只在本文件维护；Cursor 规则、共享技能和贡献指南只负责引用或执行：

- 博客主分类仅限 `AI`、`system-design`、`fundamentals`、`other`；目录映射由 `.agents/config/post-categories.json` 提供。`fundamentals` 对应计算机基础文章，默认 Front Matter 展示名为 `计算机基础`。
- `source/library/tutorials/` 仅用于公开第三方教程、实验指南或课程资料；博客教程仍是博客文章，按主题放入博客主分类并使用 `.agents/templates/tech-tutorial.md`。
- Front Matter 至少包含 `title`、`date`、`categories` 和 `tags`；`categories` 使用注册表允许的主分类映射，层级最多 2 层。
- 标签至少保留 2 个；英文多词标签使用小写 kebab-case（如 `deep-learning`），产品名、缩写和专有名词保留官方拼写，禁止仅大小写不同的重复标签。
- 图片路径使用站内 URL（如 `/images/diagram.png` 或 `/diagrams/flow.mmd`）；禁止本机绝对路径、仓库绝对路径和用外部镜像替代已托管的本地资源。
- 所有代码块必须标注语言，中英文混排保留空格；文章命名和目录专属要求见本文件第 10 节。

## 5. 统一 AI 协作资产

### 5.1 当前目录结构

```text
.
├── AGENTS.md
├── .cursorrules
├── .agents/
│   ├── agents/
│   ├── claude/
│   ├── config/
│   ├── codex/
│   ├── commands/
│   ├── cursor/
│   ├── skills/
│   └── templates/
└── bin/
    └── pre-commit-check.sh
```

### 5.2 各资产职责

- `AGENTS.md`
  - 仓库级统一入口
  - 存放项目规则、工作流、约束和 AI 协作方式
- `.cursorrules`
  - Cursor 的适配入口
  - 只负责加载本文件和路由执行流程，不承载独立规范
- `.agents/claude/settings.json`
  - Claude 类工具的 Hook 与权限配置
- `.agents/skills/`
  - 所有工具共享的技能主目录
- `.agents/commands/`
  - 快捷命令
- `.agents/config/post-categories.json`
  - 博客主分类的机器可读目录映射；规范解释以 `AGENTS.md` 为准
- `.agents/agents/`
  - 专项助手
- `.agents/templates/`
  - 文章模板
- `.agents/cursor/rules/`
  - Cursor 专属适配入口，不承载独立规范
- `.agents/codex/config.toml`
  - Codex 专属配置
- `bin/pre-commit-check.sh`
  - 提交前检查脚本

规则职责边界：

- `AGENTS.md` 是唯一规范源，维护分类、教程、Front Matter、标签、图片、命名和质量要求。
- `.cursorrules` 与 `.agents/cursor/rules/` 只负责让 Cursor 加载 `AGENTS.md` 的相关章节。
- `.agents/skills/` 与 `.agents/commands/` 只负责执行步骤、输入输出和命令；需要判断规范时回读 `AGENTS.md`。

## 6. 支持的 AI 工作面

### 6.1 Claude Code / 类 Claude 工作流

本仓库统一以 `AGENTS.md` 作为规则入口，不再维护单独的 `CLAUDE.md`。

可复用能力统一收敛到：

- `.agents/skills/new-post/`
- `.agents/skills/review-post/`
- `.agents/skills/organize-posts/`
- `.agents/skills/generate-summary/`
- `.agents/skills/link-check/`
- `.agents/skills/td-review/`

其他 Claude 专属能力来自：

- `.agents/commands/publish.md`
- `.agents/commands/stats.md`

### 6.2 Cursor 工作流

Cursor 侧通过 `.cursorrules` 和自然语言提示完成同类任务。统一要求是：

- 默认遵循本文件规则
- 用 `.cursorrules` 与 `.agents/cursor/rules/` 适配 Cursor 的加载方式
- 不再单独维护独立的 Cursor 指南文件

### 6.3 功能映射

| 目标 | Claude 风格 | Cursor 风格 |
|:---|:---|:---|
| 创建新文章 | `/new-post` | `Cmd+L` 或 `Cmd+K` 描述标题、分类、标签 |
| 审查文章 | `/review-post path/to/file.md` | 选中文章后要求按项目规范审查 |
| 生成摘要 | `/generate-summary path/to/file.md` | 选中文章后要求生成三种长度摘要 |
| 整理文章 | `/organize-posts` | 要求扫描文章结构并给整理建议 |
| 检查链接 | `/link-check` | 要求检查内部/外部链接和图片引用 |
| 发布前检查 | `bash bin/pre-commit-check.sh` | 终端执行同一脚本 |
| 统计博客 | `/stats` | 要求统计分类、标签、文章数量 |

## 7. 核心工作流

### 7.1 新建文章

创建文章时，至少确认：

1. 标题
2. 分类
3. 标签
4. 是否需要子分类
5. 是否属于系列文章

默认流程：

1. 确认文章主题和目录归属
2. 创建带完整 Front Matter 的文件
3. 参考模板补全结构
4. 编写内容
5. 运行文章审查
6. 运行提交前检查

### 7.2 文章审查

审查至少覆盖以下维度：

- Front Matter 完整性
- 内容结构
- 代码块语言标注
- 格式规范
- 技术准确性
- SEO 与标签
- 可读性
- 链接有效性

详细检查清单见：

- `.agents/templates/review-checklist.md`

### 7.3 摘要与 SEO

生成摘要时，默认输出：

- 一句话摘要（20-30 字）
- 段落摘要（100-150 字）
- 详细摘要（300-500 字）
- 标签建议
- SEO 关键词建议
- 内部链接建议

### 7.4 定期维护

推荐的低频维护节奏：

- 每周：查看统计信息
- 每月：整理文章结构
- 每季度：检查链接有效性

## 8. 专项助手与模板

### 8.1 专项助手

- `tech-reviewer`
  - 偏技术准确性审查
  - 适合复杂技术文章和架构类内容
- `seo-optimizer`
  - 偏标题、关键词、内部链接、可读性优化

### 8.2 文章模板

- `.agents/templates/tech-tutorial.md`
  - 技术教程类文章
- `.agents/templates/system-design.md`
  - 系统设计类文章
- `.agents/templates/interview-prep.md`
  - 面试准备类文章

## 9. Cursor 提示建议

在 Cursor 中，优先使用简短、结构清晰的任务提示。

### 9.1 常用提示词

创建文章：

```text
创建一篇新的博客文章：
- 标题：[标题]
- 分类：[AI/system-design/fundamentals/other]
- 标签：[tag1, tag2, tag3]
- 参考对应模板生成基础结构
```

审查文章：

```text
按照项目规范审查这篇文章，重点检查：
1. Front Matter
2. 结构
3. 代码块语言
4. 技术准确性
5. SEO 和链接
```

格式修复：

```text
检查并修复：
1. 中英文空格
2. 标点一致性
3. 代码块语言标注
4. 段落结构
```

### 9.2 Cursor 使用原则

- 先小步修改，再逐步放大任务范围
- 多文件改动优先用 `Composer`
- 讨论和规划优先用聊天模式
- 文章局部修订优先用内联编辑

## 10. 写作规范

### 10.1 Front Matter

每篇文章必须包含：

```yaml
---
title: 文章标题
date: YYYY-MM-DD
categories:
  - 分类
tags:
  - tag1
  - tag2
---
```

### 10.2 基本规则

- 日期必须为 `YYYY-MM-DD`
- 分类层级最多 2 层
- 主分类必须来自 `.agents/config/post-categories.json` 中的 `AI`、`system-design`、`fundamentals` 或 `other`
- 标签使用统一命名风格
- 代码块必须声明语言
- 中英文之间保留空格
- 图片路径按第 4.4 节使用站内 URL

### 10.6 教程与计算机基础文章

- 技术教程文章使用 `.agents/templates/tech-tutorial.md`，但目录归属仍由主分类注册表和主题决定。
- `source/_posts/fundamentals/` 专门收纳操作系统、网络、Shell 和编程语言基础文章；其默认 Front Matter 分类为 `计算机基础`。
- `tutorials` 是资料库类型，不是博客主分类；不得创建 `source/_posts/tutorials/` 作为新的博客分类。

### 10.3 文件命名

- 系列文章：数字前缀，如 `22-ai-system-design.md`
- 时效性强的研究文章：日期前缀
- 技术笔记：语义化文件名

### 10.4 AI 类文章补充规则

对于 `source/_posts/AI/`：

- 深度学习相关内容归入适当子目录
- 引用论文时尽量采用标准引用格式
- 引用代码时给出仓库或来源
- 英文术语首次出现时给出中文语义

### 10.5 系统设计文章补充规则

对于 `source/_posts/system-design/`，优先包含：

- 需求分析
- 容量估算
- 系统架构
- 数据库设计
- API 设计
- 高可用设计
- 性能优化
- 总结

## 11. 构建、校验与 Hook

### 11.1 提交前检查

统一执行：

```bash
bash bin/pre-commit-check.sh
```

脚本负责：

- 暂存 Markdown 文件的规范检查
- Hexo 清理与构建验证
- 两本 mdBook 书籍的构建验证（需要本地安装 mdBook 0.5.2）

### 11.2 Hook 行为

当前 `.agents/claude/settings.json` 中的 Hook 负责：

- 编辑文章后给出提示
- Git 提交前自动运行 `bin/pre-commit-check.sh`

### 11.3 构建即验证

对内容改动，最低验证标准是：

```bash
npm test
npm run clean
npm run build
npm run build:books
```

如果修改了 `_config.yml`，还需要重启本地预览服务。

## 12. 安全边界

除非任务明确要求，不要修改：

- `themes/`
- `node_modules/`
- `db.json`
- `.deploy_git/`
- `public/`
- `books/ai-book/book/`

同样不要：

- 在文章里使用绝对路径图片
- 把生成产物当源码直接改
- 在多个位置并行维护同一份长文

## 13. 常见陷阱

- Front Matter 的 `date` 字段不能写成对象
- 修改 `_config.yml` 后必须重启本地服务
- 新增文章后建议先执行 `npm run clean`
- 代码块没有语言标注会降低文章质量并增加构建风险
- 把 shell 脚本放进 Hexo `scripts/` 目录会被 Hexo 当作插件脚本加载

## 14. 最佳实践

### 14.1 规则维护

- 新规则优先补充进 `AGENTS.md`
    - Cursor 专属加载方式和交互适配才放进 `.cursorrules`；规范仍维护在 `AGENTS.md`
- 不再把流程性说明分散写进多个 README

### 14.2 质量控制

- 每篇文章发布前至少做一次审查
- 技术文章优先经过 `tech-reviewer`
- 重要文章优先经过 `seo-optimizer`

### 14.3 上下文控制

- 保持规则入口集中
- 复杂任务先规划，再执行
- 大任务拆分成多个小步骤

## 15. 相关文件

- `CONTRIBUTING.md`：开源贡献流程
- `.cursorrules`：Cursor 适配入口
- `.agents/templates/review-checklist.md`：审查清单
- `AGENTS.md`：Agent 内容地图、维护边界与内容落点整理
