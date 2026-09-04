# Cursor 适配入口：Hexo 博客

本文件只负责在 Cursor 处理 Hexo 博客时选择上下文。规范唯一源是 [`AGENTS.md`](../../../AGENTS.md)。

处理文章、资源或构建任务前读取 `AGENTS.md` 第 3、4、7、10、11 节；需要解析博客主分类目录时使用 `.agents/config/post-categories.json`。

按任务路由到：

- 新建文章：`.agents/skills/new-post/SKILL.md`
- 整理文章：`.agents/skills/organize-posts/SKILL.md`
- 审查文章：`.agents/skills/review-post/SKILL.md`
- 检查链接：`.agents/skills/link-check/SKILL.md`
- 提交前验证：`tools/pre-commit-check.sh`

本文件不复制写作、分类、标签、图片、命名、教程或质量规则；长期规范只写入 `AGENTS.md`。
