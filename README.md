# wxquare.github.io

> 一个长期维护的技术博客与知识库，重点关注系统设计与后端架构，以及 AI 与 Agent 工程实践。

## 核心入口

| 入口 | 说明 |
|------|------|
| [在线博客](https://wxquare.github.io) | 浏览完整博客内容与专题导航 |
| [AI Agent 工程实践](./books/ai-book/) | 从大模型基础到生产级智能体系统的系统化专题 |
| [系统设计与架构实战](./books/system-design-architecture-book/) | 面向中高级工程师的系统设计、电商架构与可靠性实战 |

## 精选内容

- [Claude Code 实践：从能写到写对](./source/_posts/AI/01-claude-code-practices.md)
- [Harness Engineering：把模型放进可验证的工程系统](./source/_posts/AI/06-harness-engineering.md)
- [系统设计完全指南：从问题定义到架构落地](./source/_posts/system-design/00-system-design-overview.md)
- [系统可靠性工程：从故障恢复到治理闭环](./source/_posts/system-design/07-system-reliability-engineering.md)
- [计价系统设计与实现](./source/_posts/system-design/24-ecommerce-pricing-engine.md)
- [搜索与导购系统设计](./source/_posts/system-design/31-ecommerce-search-discovery.md)
- [首页与导购链路性能优化面试材料](./source/about/material/homepage-performance-interview-material.md)
- [计价引擎面试材料](./source/about/material/pricing-engine-interview-material.md)

## 内容地图

### 系统设计与后端架构

围绕业务边界、系统拆分、数据一致性、可靠性工程和架构治理展开，覆盖 MySQL、Redis、Kafka、Elasticsearch、Kubernetes，以及商品、库存、计价、订单、支付等核心系统。

### AI 与 Agent 工程实践

关注大模型能力边界、Prompt Engineering、Context Engineering、Harness Engineering，以及 Tool Calling、MCP、RAG、Memory、Evals、Guardrails 和 Agent 平台化落地。

### 电商架构与性能优化

以电商链路为高密度样本，讨论首页、搜索、详情、购物车、计价、订单、支付、供应商同步与 B2B2C 平台演进，也包含面试表达与案例材料沉淀。

### 计算机基础

补充操作系统、网络、Shell、Python、C++、Go 等基础能力，帮助把系统设计、工程实现和编码实践串起来。

## 本地运行

### 环境要求

- Node.js >= 14
- npm >= 6.0

### 常用命令

```bash
# 安装依赖
npm install

# 启动本地预览服务器
npm run server

# 生成静态文件
npm run build

# 清理缓存
npm run clean
```

访问：

```text
http://localhost:4000
```

## 仓库结构

```text
.
├── source/
│   ├── _posts/                  # 博客文章
│   ├── about/                   # 关于页、简历、面试材料
│   └── diagrams/                # 图表源文件
├── books/
│   ├── ai-book/                 # AI Agent 工程实践专题
│   ├── system-design-architecture-book/  # 系统设计与架构专题
│   └── scripts/                 # 共用构建脚本
├── docs/                        # 规划、设计与过程文档
├── _config.yml                  # Hexo 配置
└── package.json                 # 项目依赖与脚本
```

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
3. 提交前运行 `npm run build`，先确认构建通过。
4. Front Matter 中 `date` 必须是字符串，不能写成对象。

### 相关说明

- Hexo 配置与博客搭建记录见 [基于 Github 双分支和 Hexo 搭建博客](./source/_posts/other/基于Github双分支和Hexo搭建博客.md)。

## 许可证

本项目采用 MIT 许可证。详见 `LICENSE`。



sk-af206d731b2045e2849d686c962c23d0