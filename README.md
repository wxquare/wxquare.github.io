# wxquare.github.io

> 一个长期维护的技术博客与知识库，重点关注系统设计与后端架构，以及 AI 与 Agent 工程实践。

## 核心入口

| 入口 | 说明 |
|------|------|
| [在线博客](https://wxquare.github.io) | 浏览完整博客内容与专题导航 |
| [公开参考资料库](./source/library/index.md) | 浏览公开且适合公开保存的第三方技术资料及原始来源 |
| [AI Agent 工程实践](./books/ai-book/) | 从大模型基础到生产级智能体系统的系统化专题 |
| [AGENTS.md](./AGENTS.md) | 仓库内 Agent 内容开发、维护边界与人机协作规范 |
| [CONTRIBUTING.md](./CONTRIBUTING.md) | 开源贡献流程与改动边界 |
| [系统设计与架构实战](./books/system-design-architecture-book/) | 面向中高级工程师的系统设计、电商架构与可靠性实战 |

## 精选内容

- [Claude Code 实践：从能写到写对](./source/_posts/AI/01-claude-code-practices.md)
- [Harness Engineering：把模型放进可验证的工程系统](./source/_posts/AI/06-harness-engineering.md)
- [系统设计完全指南：从问题定义到架构落地](./source/_posts/system-design/00-system-design-overview.md)
- [系统可靠性工程：从故障恢复到治理闭环](./source/_posts/system-design/07-system-reliability-engineering.md)
- [计价系统设计与实现](./source/_posts/system-design/24-ecommerce-pricing-engine.md)
- [搜索与导购系统设计](./source/_posts/system-design/31-ecommerce-search-discovery.md)

## 内容地图

### 系统设计与后端架构

围绕业务边界、系统拆分、数据一致性、可靠性工程和架构治理展开，覆盖 MySQL、Redis、Kafka、Elasticsearch、Kubernetes，以及商品、库存、计价、订单、支付等核心系统。

### AI 与 Agent 工程实践

关注大模型能力边界、Prompt Engineering、Context Engineering、Harness Engineering，以及 Tool Calling、MCP、RAG、Memory、Evals、Guardrails 和 Agent 平台化落地。

### 电商架构与性能优化

以电商链路为高密度样本，讨论首页、搜索、详情、购物车、计价、订单、支付、供应商同步与 B2B2C 平台演进，沉淀可公开复用的技术案例与架构实践。

### 计算机基础

补充操作系统、网络、Shell、Python、C++、Go 等基础能力，帮助把系统设计、工程实现和编码实践串起来。

### 公开资料库

`source/library/` 是统一资料根目录，按 books、papers、slides、presentations、other 组织内容：`slides/` 保存第三方演示资料，`presentations/` 保存本人创作或确认可公开的第一方演示资料。每条记录都包含来源、再分发说明和本地路径；无法确认重新托管边界的资料只保留原始链接。Other 只接收经批准且类型边界明确的补充资料。`source/booklist/` 是已阻塞的只读遗留目录，`source/pdf/` 不含活动源文件；旧 `/presentations/` 和 `/pdf/k8s-network.pdf` URL 由构建期 alias 兼容。博客与资料库只介绍或引用实验，不存放实验源码。

## 本地运行

### 环境要求

- Node.js >= 14
- npm >= 6.0
- mdBook 0.5.2（本地版本应与 CI 保持一致）

安装 mdBook：

```bash
# macOS
brew install mdbook

# 或已安装 Rust 时
cargo install mdbook --version 0.5.2 --locked

mdbook --version
```

### 常用命令

```bash
# 安装依赖
npm install

# 启动本地预览服务器
npm run server

# 生成静态文件
npm run build

# 构建 AI Agent 书籍
npm run build:ai-book

# 构建系统设计与架构书籍
npm run build:system-design-book

# 构建两本书
npm run build:books

# 将两本书汇总到 Hexo 发布目录
npm run stage:books

# 构建完整发布树并启动本地预览（端口 3000）
npm run server:site

# 清理缓存
npm run clean
```

Hexo 输出到 `public/`；两本 mdBook 分别输出到 `books/ai-book/book/` 和
`books/system-design-architecture-book/book/`。需要一次构建全部内容时，先运行
`npm run clean && npm run build && npm run build:books && npm run stage:books`，即可将两本书
汇总到 `public/ai-book/` 和 `public/system-design-architecture-book/` 下并预览完整发布树。

`npm run server` 访问博客：

```text
http://localhost:4000
```

`npm run server:site` 访问包含两本书的完整发布树：

```text
http://localhost:3000
```

### GitHub Actions 部署

`.github/workflows/deploy-site.yml` 是唯一的部署 Workflow。它在 `hexo` 分支相关内容
变更或手动触发时，依次调用 `npm run build`、`npm run build:books` 和 `npm run stage:books`，
将 Hexo 站点与两本书汇总到同一个 `public/` 目录后一次发布：

```text
public/
├── index.html
├── ai-book/
└── system-design-architecture-book/
```

部署后的访问路径分别是 `/`、`/ai-book/` 和 `/system-design-architecture-book/`。

## 仓库结构

```text
.
├── source/
│   ├── _posts/                  # 博客文章
│   ├── about/                   # 联系页（仅邮箱）
│   ├── diagrams/                # 图表源文件
│   ├── library/                 # 公开参考资料与演示资料目录
│   │   ├── books/
│   │   ├── papers/
│   │   ├── slides/
│   │   ├── presentations/
│   │   ├── other/
│   │   └── index.md
├── books/
│   ├── ai-book/                 # AI Agent 工程实践专题
│   └── system-design-architecture-book/  # 系统设计与架构专题
├── scripts/                     # 共用构建、校验与预处理脚本
├── docs/                        # 规划、设计与过程文档
├── _config.yml                  # Hexo 配置
└── package.json                 # 项目依赖与脚本
```

### 共用脚本

根目录 `scripts/` 下的脚本用于仓库级构建和校验，不承担博客业务逻辑：

- `tools/pre-commit-check.sh`：检查暂存 Markdown，并验证 Hexo 与 mdBook 构建，避免错误内容进入提交。
- `tools/stage-books.js`：将两本 mdBook 的生成结果复制到 `public/`，供 Hexo 统一发布。
- `tools/mermaid-preprocessor.py`：在 mdBook 构建期间把 Mermaid Markdown 代码块转换成 Mermaid.js 可渲染的 HTML。

注意：Hexo 会递归加载 `scripts/` 下的文件，因此非 Hexo 插件脚本统一放在根目录 `tools/`，并通过 npm 或 mdBook 显式调用。

## 写作与维护约定

### Front Matter

每篇文章必须包含：

```yaml
---
title: 文章标题
date: YYYY-MM-DD
categories:
  - 分类1
  - 分类2
tags:
  - tag1
  - tag2
---
```

### 基本规范

- 分类层级最多 2 层。
- 标签使用小写，多个词用连字符连接，如 `deep-learning`。
- 中文文章使用中文标点，英文文章使用英文标点。
- 代码块必须指定语言。
- 中英文之间保留空格。
- 图片使用相对路径，不要使用绝对路径。

### 常见陷阱

1. 修改 `_config.yml` 后必须重启本地服务。
2. 新增文章后建议先运行 `npm run clean`。
3. 提交前运行 `npm test`、`npm run build` 和 `npm run build:books`，确认博客与两本书都能构建。
4. Front Matter 中 `date` 必须是字符串，不能写成对象。

### 相关说明

- Hexo 配置与博客搭建记录见 [基于 Github 双分支和 Hexo 搭建博客](./source/_posts/other/基于Github双分支和Hexo搭建博客.md)。

## 许可证

本项目采用 MIT 许可证。详见 `LICENSE`。
