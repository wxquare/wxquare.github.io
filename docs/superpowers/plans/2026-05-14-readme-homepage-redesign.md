# README Homepage Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite `README.md` into a visitor-first homepage that highlights the blog's core themes, featured entry points, and selected content while retaining a compact maintenance appendix.

**Architecture:** Replace the current flat repository manual with a layered document. The new top half will present positioning, core entry points, and featured content; the middle will provide a compact content map; the bottom will keep only the minimum local development and writing-maintenance notes needed for repo collaborators.

**Tech Stack:** Markdown, Hexo repository structure, existing mdBook/book links

---

### Task 1: Collect stable entry points and representative content

**Files:**
- Modify: `README.md`
- Reference: `books/ai-book/src/README.md`
- Reference: `books/ai-book/src/SUMMARY.md`
- Reference: `books/system-design-architecture-book/src/README.md`
- Reference: `books/system-design-architecture-book/src/SUMMARY.md`

- [ ] **Step 1: Inspect existing book entry documents**

Run:

```bash
sed -n '1,220p' books/ai-book/src/README.md
sed -n '1,220p' books/system-design-architecture-book/src/README.md
sed -n '1,220p' books/ai-book/src/SUMMARY.md
sed -n '1,220p' books/system-design-architecture-book/src/SUMMARY.md
```

Expected: clear book positioning plus stable chapter paths for featured links.

- [ ] **Step 2: Select featured entries**

Use these repository-relative links as the initial featured set:

```text
./books/ai-book/
./books/system-design-architecture-book/
./source/_posts/AI/01-claude-code-practices.md
./source/_posts/AI/06-harness-engineering.md
./source/_posts/system-design/00-system-design-overview.md
./source/_posts/system-design/07-system-reliability-engineering.md
./source/about/material/homepage-performance-interview-material.md
./source/about/material/pricing-engine-interview-material.md
```

Expected: a balanced list centered on AI Agent plus system design.

### Task 2: Rewrite README information architecture

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Replace the current intro and navigation sections**

Write a new opening structure with:

```md
# wxquare.github.io

> 一个长期维护的技术博客与知识库，重点关注系统设计与后端架构，以及 AI 与 Agent 工程实践。
```

Then add compact core entry points for the online blog, AI book, and system design book.

- [ ] **Step 2: Add featured content and content map**

Add:

```md
## 精选内容
```

with 6-8 selected links, then:

```md
## 内容地图
```

with four short theme blocks:

- 系统设计与后端架构
- AI 与 Agent 工程实践
- 电商架构与性能优化
- 计算机基础

Expected: visitor can identify what to read first without scanning long nested lists.

- [ ] **Step 3: Compress maintenance-oriented sections**

Keep only:

- `## 本地运行`
- `## 仓库结构`
- `## 写作与维护约定`

Remove or fold:

- long topic enumerations
- 推荐工具
- 开发规范
- 参考资源
- 外部链接资源

Expected: README reads like a homepage first and a repository manual second.

### Task 3: Verify readability and repository accuracy

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Review the rewritten README in plain text**

Run:

```bash
sed -n '1,260p' README.md
```

Expected: top sections appear in this order:

```text
标题
一句定位
核心入口
精选内容
内容地图
本地运行
仓库结构
写作与维护约定
```

- [ ] **Step 2: Check for stale or misleading paths**

Run:

```bash
rg -n "system-design-architecture-book|ai-book|homepage-performance-interview-material|pricing-engine-interview-material|01-claude-code-practices|06-harness-engineering|00-system-design-overview|07-system-reliability-engineering" README.md
```

Expected: every featured path exists and matches current repository structure.

- [ ] **Step 3: Commit checkpoint**

```bash
git add README.md docs/superpowers/specs/2026-05-14-readme-homepage-redesign-design.md docs/superpowers/plans/2026-05-14-readme-homepage-redesign.md
git commit -m "docs: redesign README homepage structure"
```

Expected: only run this commit step if the user asks for a commit.
