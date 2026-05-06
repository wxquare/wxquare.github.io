# 附录 C：工具与构建说明

## mdBook

与仓库内 `books/ai-book`、`books/ecommerce-book` 相同：安装 [mdBook](https://github.com/rust-lang/mdBook)，可选安装 `mdbook-mermaid` 以渲染 Mermaid。

```bash
cd books/system-design-book
mdbook build
mdbook serve
```

## Mermaid

`book.toml` 中 `additional-js` 指向与 `book.toml` 同目录下的 `mermaid.min.js` 与 `mermaid-init.js`；Mermaid 代码块由 `books/scripts/mermaid-preprocessor.py` 统一预处理。若升级版本，请同步检查这两个静态文件。
