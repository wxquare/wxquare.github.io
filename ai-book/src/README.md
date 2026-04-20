# 书籍介绍

本书面向**希望系统理解大模型、AI 编程与 Agent 工程**的开发者与架构师，从基础概念出发，串联工具链与落地方法，并以实战案例收束。

## 你将读到什么

- **大模型基础**：能力边界、训练 / 微调 / 推理工程、提示与 RAG 入门。
- **AI 编程**：人机协作范式、代码生成与审查、团队流程与安全边界。
- **AI Agent**：运行循环、工具与 MCP、多 Agent 与工作流。
- **实战**：从原型到上线、评测与持续迭代。

## 如何使用本书

各章目前为**占位骨架**，便于你在 mdbook 中先跑通目录与构建，再逐章填充正文。修改书名请在项目根目录的 `book.toml` 中调整 `[book] title` 与本页标题。

## 本地构建

需已安装 [mdBook](https://github.com/rust-lang/mdBook) 与 [mdbook-mermaid](https://github.com/badboy/mdbook-mermaid)。

```bash
cd ai-book
mdbook serve
```

浏览器打开提示的本地地址即可预览。
