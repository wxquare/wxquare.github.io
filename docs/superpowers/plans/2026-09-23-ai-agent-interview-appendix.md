# AI Agent Interview Appendix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 AI 与 Agent 开发 50 题题库迁移为 `ai-book` 的附录 F，并归档原博客文章。

**Architecture:** 以 `books/ai-book/src/appendix/ai-agent-development-interview-50.md` 作为活动书稿源，以 `source/archive/AI/07-ai-agent-development-interview-50.md` 保留不可发布的历史博客副本。题库正文不改写，只移除 Hexo front matter、补充书稿标题，并在 `SUMMARY.md` 注册导航。

**Tech Stack:** Markdown、mdBook SUMMARY、Hexo front matter、Ruby editorial-integrity 检查脚本、Git。

## Global Constraints

- 保持题目、答案、顺序和参考链接原样，不改写题库内容。
- 原博客文章移动到 `source/archive/AI/07-ai-agent-development-interview-50.md`，并设置 `published: false`。
- 只修改 `books/ai-book/src/` 与 `source/archive/AI/` 的本次目标文件，不修改 `books/ai-book/book/` 生成目录。
- 保留工作区中与本任务无关的所有未提交改动，不执行清理、重置或覆盖操作。

---

### Task 1: 建立附录 F 书稿源

**Files:**
- Create: `books/ai-book/src/appendix/ai-agent-development-interview-50.md`
- Read: `source/_posts/AI/07-ai-agent-development-interview-50.md`

**Interfaces:**
- Consumes: 原博客文章的 front matter 和 50 题正文。
- Produces: 无 Hexo front matter、可被 mdBook 读取的附录 F Markdown 文件；从“50 题总览”到参考链接的正文保持逐字一致。

- [ ] **Step 1: 确认源文章边界**

  使用 `sed -n '1,12p' source/_posts/AI/07-ai-agent-development-interview-50.md` 确认 front matter 结束于 `---`，使用 `rg -n '^### [0-9]+\\.' source/_posts/AI/07-ai-agent-development-interview-50.md` 确认题号范围覆盖 1–50。

- [ ] **Step 2: 创建附录文件**

  新文件第一行使用：

  ```markdown
  # 附录 F AI 与 AI Agent 开发高频面试 50 题（工程与架构优先）
  ```

  随后接入原文章从 `## 50 题总览` 开始的全部正文；不得把 YAML front matter 复制到 mdBook 文件中，不得改动正文措辞、题号、链接或代码围栏。

- [ ] **Step 3: 检查附录正文完整性**

  运行：

  ```bash
  test -f books/ai-book/src/appendix/ai-agent-development-interview-50.md
  rg -n '^### (1|50)\\.' books/ai-book/src/appendix/ai-agent-development-interview-50.md
  test "$(rg -c '^### [0-9]+\\.' books/ai-book/src/appendix/ai-agent-development-interview-50.md)" -eq 50
  ```

  预期：文件存在，首尾题目存在，题目标题总数为 50。

### Task 2: 归档原博客文章

**Files:**
- Move: `source/_posts/AI/07-ai-agent-development-interview-50.md` → `source/archive/AI/07-ai-agent-development-interview-50.md`

**Interfaces:**
- Consumes: Task 1 已读取的原文章。
- Produces: 保留原 front matter 和全文、但不会作为 Hexo 活动文章发布的归档副本。

- [ ] **Step 1: 确认归档目标不存在**

  运行 `test ! -e source/archive/AI/07-ai-agent-development-interview-50.md`，确保不会覆盖现有文件。

- [ ] **Step 2: 移动并标记为不可发布**

  使用版本可追踪的移动操作将源文件放入 `source/archive/AI/`，并在 front matter 的 `tags` 后、结束分隔线前加入：

  ```yaml
  published: false
  ```

- [ ] **Step 3: 验证归档状态**

  运行：

  ```bash
  test ! -e source/_posts/AI/07-ai-agent-development-interview-50.md
  test -f source/archive/AI/07-ai-agent-development-interview-50.md
  rg -n '^published: false$' source/archive/AI/07-ai-agent-development-interview-50.md
  ```

  预期：原发布路径不存在，归档文件存在并包含唯一的 `published: false`。

### Task 3: 注册附录导航

**Files:**
- Modify: `books/ai-book/src/SUMMARY.md`

**Interfaces:**
- Consumes: Task 1 的附录文件路径。
- Produces: 附录列表中的附录 F 链接，链接目标为 `appendix/ai-agent-development-interview-50.md`。

- [ ] **Step 1: 在附录 E 后增加导航项**

  保持现有附录 A–E 不变，在附录 E 下一行加入：

  ```markdown
  - [附录F AI 与 AI Agent 开发高频面试 50 题](appendix/ai-agent-development-interview-50.md)
  ```

- [ ] **Step 2: 检查导航链接**

  运行 `rg -n '附录F|ai-agent-development-interview-50' books/ai-book/src/SUMMARY.md`，并确认链接目标 `test -f books/ai-book/src/appendix/ai-agent-development-interview-50.md` 成功。

### Task 4: 全量验证并复查 diff

**Files:**
- Test: `books/ai-book/scripts/test_verify_editorial_integrity.rb`
- Inspect: `books/ai-book/src/SUMMARY.md`
- Inspect: `git diff` and `git status`

**Interfaces:**
- Consumes: Tasks 1–3 的附录源、归档文章和导航改动。
- Produces: 可审查的最终迁移结果，不修改 mdBook 生成产物和无关工作区文件。

- [ ] **Step 1: 运行 editorial integrity 检查**

  运行 `ruby books/ai-book/scripts/test_verify_editorial_integrity.rb`。

  预期：检查通过；若失败，记录具体失败位置并只修复本次迁移引入的问题。

- [ ] **Step 2: 对比活动附录与归档正文**

  用脚本分别去除归档文件的 front matter 和附录文件首行标题，再比较两者正文，确认无题库内容差异；同时检查 `rg -n '^### [0-9]+\\.'` 的结果仍为 1–50。

- [ ] **Step 3: 检查 diff 范围**

  运行 `git status --short` 与 `git diff --stat`，确认本任务只新增附录、更新 SUMMARY、迁移并标记目标博客，以及新增的设计/计划文档；不得暂存或覆盖其他已有改动。

- [ ] **Step 4: 提交本任务变更**

  只暂存本任务文件：

  ```bash
  git add books/ai-book/src/appendix/ai-agent-development-interview-50.md \
    books/ai-book/src/SUMMARY.md \
    source/archive/AI/07-ai-agent-development-interview-50.md \
    docs/superpowers/specs/2026-09-23-ai-agent-interview-appendix-design.md \
    docs/superpowers/plans/2026-09-23-ai-agent-interview-appendix.md \
    source/_posts/AI/07-ai-agent-development-interview-50.md
  git commit -m "docs: add AI agent interview appendix"
  ```

  不要暂存工作区中其他用户改动。
