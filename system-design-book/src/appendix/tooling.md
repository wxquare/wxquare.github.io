# 附录 C：工具与构建说明

## mdBook

与仓库内 `ai-book`、`ecommerce-book` 相同：安装 [mdBook](https://github.com/rust-lang/mdBook)，可选安装 `mdbook-mermaid` 以渲染 Mermaid。

```bash
cd system-design-book
mdbook build
mdbook serve
```

## Mermaid

`book.toml` 中 `additional-js` 指向与 `book.toml` 同目录下的 `mermaid.min.js` 与 `mermaid-init.js`（已从 `ai-book` 复制）。若升级版本，请同步替换这两个文件。
