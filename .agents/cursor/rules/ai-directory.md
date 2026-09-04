# Cursor 适配入口：AI 文章

本文件只负责在 Cursor 处理 `source/_posts/AI/` 下的文章时选择上下文。规范唯一源是 [`AGENTS.md`](../../../AGENTS.md)。

处理 AI 文章前读取 `AGENTS.md` 第 3、4、7、10.1-10.4、11 节；目录和 Front Matter 映射按 `.agents/config/post-categories.json` 使用。

本文件不定义分类、子目录、命名、标签、图片、引用或代码规范。若发现缺少长期规则，应修改 `AGENTS.md`，不要在此追加副本。
