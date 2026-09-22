# AI Agent 面试题库附录迁移设计

## 背景

`source/_posts/AI/07-ai-agent-development-interview-50.md` 是一篇包含 50 道 AI 与 Agent 工程面试题及参考答案的长文。`books/ai-book/src/appendix/` 已有题库素材和其他附录，但该题库目前仍作为博客文章维护，尚未进入书稿导航。

## 目标

1. 将 50 题题库纳入 `books/ai-book`，作为新的附录 F。
2. 保持题目、答案、顺序和参考链接原样，不进行内容改写或删减。
3. 将原博客文章移入 `source/archive/AI/`，并通过 `published: false` 停止博客发布，避免继续维护两份活动内容。
4. 在 `books/ai-book/src/SUMMARY.md` 中增加附录 F 的导航入口。

## 非目标

- 不调整题库的分类、题号、答案、面试回答套路或参考链接。
- 不修改现有附录 E、书稿章节和其他博客文章。
- 不直接修改 `books/ai-book/book/` 等 mdBook 生成目录。
- 不处理工作区中与本次题库迁移无关的已有未提交改动。

## 内容落点

### 活动书稿源

新建 `books/ai-book/src/appendix/ai-agent-development-interview-50.md`。文件去除 Hexo front matter，并增加书稿所需的一级标题“附录 F AI 与 AI Agent 开发高频面试 50 题（工程与架构优先）”；从原文章 `## 50 题总览` 开始的题库正文逐字保留。

### 历史博客归档

将原文章移动为 `source/archive/AI/07-ai-agent-development-interview-50.md`。归档文件保留原 front matter 和正文，只在 front matter 增加：

```yaml
published: false
```

### 书稿导航

在现有附录 E 后增加：

```markdown
- [附录F AI 与 AI Agent 开发高频面试 50 题](appendix/ai-agent-development-interview-50.md)
```

## 验证策略

1. 检查原 `_posts` 路径不存在、归档路径存在且包含 `published: false`。
2. 检查书稿附录文件存在，题号 `1` 到 `50` 均保留，一级标题和导航链接存在。
3. 检查新增书稿正文与原文章正文（去除 front matter、增加附录标题后的预期内容）一致。
4. 运行 `ruby books/ai-book/scripts/test_verify_editorial_integrity.rb`。
5. 复查 `git diff` 和 `git status`，确认不包含用户已有改动的误修改。

## 风险与应对

- **题库内容被意外改写**：迁移时只移除博客元数据，并用正文对比和题号扫描校验。
- **博客仍被发布**：归档路径移出 `source/_posts/`，同时设置 `published: false`。
- **生成目录被误改**：只操作 `books/ai-book/src/` 和 `SUMMARY.md`，不编辑 `books/ai-book/book/`。
