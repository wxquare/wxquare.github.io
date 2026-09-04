# mdBook 教程：当前仓库的两本书

> 本文只说明当前仓库维护的 mdBook 项目。仓库不再创建 `ecommerce-book`，也不存在 `source/book`。

## 1. 项目目录

当前只有两个 mdBook 项目：

```text
books/
├── ai-book/
│   ├── book.toml
│   ├── src/                 # Markdown 源文件与 SUMMARY.md
│   └── book/                # mdBook 构建输出，不直接编辑
└── system-design-architecture-book/
    ├── book.toml
    ├── src/                 # Markdown 源文件与 SUMMARY.md
    ├── images/              # 本书使用的图片源文件
    └── book/                # mdBook 构建输出，不直接编辑
```

两个项目的构建约定相同：`book.toml` 位于项目根目录，正文位于 `src/`，HTML 输出到项目内的 `book/`。两本书都使用 `books/scripts/mermaid-preprocessor.py` 处理 Mermaid 代码块。

| 项目 | 内容 | 在线路径 |
| --- | --- | --- |
| `books/ai-book/` | AI Agent 工程实践 | `/ai-book/` |
| `books/system-design-architecture-book/` | 系统设计、架构与电商案例 | `/system-design-architecture-book/` |

系统设计书的 `book.toml` 还包含旧章节 URL 的重定向规则；修改章节路径时要同时检查这些规则。

## 2. 安装环境

需要 Rust、mdBook 和 Python 3：

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source "$HOME/.cargo/env"
cargo install mdbook

rustc --version
mdbook --version
python3 --version
```

本仓库没有要求使用 `mdbook-pdf` 或 `mdbook-mermaid`。Mermaid 由仓库内的 Python 预处理器和 mdBook 的 HTML 配置共同处理；安装其他插件不会替代这套配置。

## 3. 单独构建一本书

在仓库根目录执行：

```bash
# 构建 AI Agent 工程实践
mdbook build books/ai-book

# 构建系统设计与架构实战
mdbook build books/system-design-architecture-book
```

也可以使用仓库已有的 npm 命令：

```bash
npm run build:ai-book
npm run build:system-design-book

# 按顺序构建两本书
npm run build:books
```

构建后分别检查：

```text
books/ai-book/book/
books/system-design-architecture-book/book/
```

## 4. 本地预览

预览单本书时，在对应项目目录执行：

```bash
cd books/ai-book
mdbook serve --open
```

或：

```bash
cd books/system-design-architecture-book
mdbook serve --open
```

结束服务按 `Ctrl+C`。修改 `src/` 下的 Markdown 文件后，开发服务器会重新构建页面。

如果需要同时预览 Hexo 博客和两本书，从仓库根目录执行：

```bash
npm run server:site
```

该命令会清理并构建 Hexo、构建两本书、把书籍输出暂存到 `public/`，然后启动本地站点。系统设计书的兼容入口 `npm run server:system-design-architecture-book` 等价于该命令。

## 5. 修改书籍内容

正文修改位置：

```text
books/ai-book/src/
books/system-design-architecture-book/src/
```

每个新章节都必须同时完成两步：

1. 在对应的 `src/` 下创建 Markdown 文件。
2. 在对应的 `src/SUMMARY.md` 中注册文件，否则 mdBook 不会把它加入书籍。

不要直接编辑以下生成目录：

```text
books/ai-book/book/
books/system-design-architecture-book/book/
public/
```

图片应放在对应书籍的源码目录或 `books/system-design-architecture-book/images/` 中，并使用相对于当前 Markdown 文件的路径引用。

## 6. 配套示例代码

系统设计书的 Go 示例工程不属于 mdBook 项目，已独立放在：

```text
~/Projects/system-design-architecture-examples/
├── common-services/
├── order-service/
└── product-service/
```

运行示例：

```bash
cd ~/Projects/system-design-architecture-examples/product-service
go test ./...
go run cmd/main.go
```

书稿中的示例代码引用应指向 `~/Projects/system-design-architecture-examples/`，不要重新写成 `books/system-design-architecture-book/example-codes/`。示例工程的代码变更和 Go 测试也应在该独立目录中完成，不会被 `npm run build:books` 自动执行。

## 7. 完整站点构建与部署

根目录的 `package.json` 和 `.github/workflows/deploy-site.yml` 已经负责统一构建：

```bash
npm run clean
npm run build
npm run build:books
npm run stage:books
```

`stage:books` 将构建结果复制到：

```text
public/ai-book/
public/system-design-architecture-book/
```

推送到 `hexo` 分支后，`deploy-site.yml` 会构建 Hexo 和两本书，并将 `public/` 部署到 GitHub Pages。无需为任一本书创建独立 Git 仓库或独立的 `ecommerce-book` 工作流。

## 8. 常见问题

### 找不到 `source/book`

这是旧迁移教程中的路径。当前书籍正文分别在 `books/ai-book/src/` 和 `books/system-design-architecture-book/src/`；`source/` 是 Hexo 博客内容目录，不是 mdBook 源目录。

### 构建结果没有更新

确认修改的是 `src/` 而不是 `book/`，然后从项目根目录重新运行：

```bash
mdbook clean books/ai-book
mdbook build books/ai-book
```

另一项目同理。

### Mermaid 图表没有渲染

确认使用 Python 3，并检查对应项目的 `book.toml` 是否仍保留：

```toml
[preprocessor.mermaid]
command = "python3 ../scripts/mermaid-preprocessor.py"
optional = true
```

## 9. 快速命令参考

```bash
# 构建
npm run build:ai-book
npm run build:system-design-book
npm run build:books

# 单本书预览
cd books/ai-book && mdbook serve --open
cd books/system-design-architecture-book && mdbook serve --open

# 完整站点本地预览
npm run server:site
```

官方文档：[mdBook](https://rust-lang.github.io/mdBook/)、[Rust 安装指南](https://www.rust-lang.org/tools/install)。
