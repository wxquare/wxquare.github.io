# AGENTS.md

本文件是仓库内唯一的 AI 协作入口。与 AI 开发指导、协作约束、工具工作流相关的规则，统一维护在这里，不再拆散到 `CLAUDE.md`、`README-AI-SETUP.md`、`README-CURSOR-USAGE.md`。

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
- `source/_posts/other/`：其他文章
- `source/about/`：关于页、简历、面试材料
- `source/diagrams/`：图表源文件
- `books/ai-book/src/`：AI Agent 书稿源码
- `books/ai-book/labs/`：可运行 Agent 原型和实验
- `docs/`：调研、迁移、整理文档
- `.agents/`：统一的 AI 协作资产与工具配置目录

### 3.2 不要直接编辑的目录

- `public/`
- `.deploy_git/`
- `books/ai-book/book/`

## 4. 内容落点规则

每类内容只保留一个主事实源：

- 系统化 Agent 知识：`books/ai-book/src/`
- 面向读者的博客文章：`source/_posts/AI/`
- 可运行原型与 demo：`books/ai-book/labs/`
- AI 协作规则、共享技能与工具配置：`AGENTS.md`、`.agents/`
- 内部整理文档与迁移说明：`docs/`

不要把同一份长文同时维护在博客、书稿和 `docs/` 三处。

## 5. 统一 AI 协作资产

### 5.1 当前目录结构

```text
.
├── AGENTS.md
├── .cursorrules
├── .agents/
│   ├── agents/
│   ├── claude/
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
  - Cursor 的补充规则入口
  - 内容应与 `AGENTS.md` 保持一致，不再单独承载主规则
- `.agents/claude/settings.json`
  - Claude 类工具的 Hook 与权限配置
- `.agents/skills/`
  - 所有工具共享的技能主目录
- `.agents/commands/`
  - 快捷命令
- `.agents/agents/`
  - 专项助手
- `.agents/templates/`
  - 文章模板
- `.agents/cursor/rules/`
  - Cursor 专属加载入口
- `.agents/codex/config.toml`
  - Codex 专属配置
- `bin/pre-commit-check.sh`
  - 提交前检查脚本

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
- 分类：[AI/system-design/other]
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
- 标签使用统一命名风格
- 代码块必须声明语言
- 中英文之间保留空格
- 图片使用相对路径

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

### 11.2 Hook 行为

当前 `.agents/claude/settings.json` 中的 Hook 负责：

- 编辑文章后给出提示
- Git 提交前自动运行 `bin/pre-commit-check.sh`

### 11.3 构建即验证

对内容改动，最低验证标准是：

```bash
npm run clean
npm run build
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
- Cursor 专属补充才放进 `.cursorrules`
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
- `.cursorrules`：Cursor 适配规则
- `.agents/templates/review-checklist.md`：审查清单
- `docs/agent-development-guide.md`：Agent 内容地图与内容落点整理
