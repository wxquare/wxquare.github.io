# wxquare.github.io - 个人技术博客

> 一个基于 Hexo 搭建的技术知识库，覆盖系统设计、AI 与 Agent、电商架构、计算机基础等核心领域。

## 📋 快速导航

| 资源 | 描述 |
|------|------|
| [🌐 在线博客](https://wxquare.github.io) | 访问完整的博客内容 |
| [📚 AI Agent 工程实践](./books/ai-book/) | mdBook 形式的系统教程 |
| [🛍️ 电商系统架构设计](./books/ecommerce-book/) | 电商架构完整实现 |
| [🎯 系统设计与面试指南](./books/system-design-book/) | 面试和系统设计经验总结 |

---

## 🚀 快速开始

### 环境要求
- Node.js >= 14
- npm >= 6.0

### 本地运行

```bash
# 安装依赖
npm install

# 启动本地预览服务器 (http://localhost:4000)
npm run server

# 生成静态文件
npm run build

# 清理缓存
npm run clean

# 部署到 GitHub Pages
npm run deploy
```

---

## 📁 项目结构

```
.
├── source/
│   ├── _posts/              # 博客文章
│   │   ├── AI/              # AI 与 Agent 相关
│   │   ├── system-design/   # 系统设计
│   │   └── other/           # 其他主题
│   ├── about/               # 关于页面、简历、面试资料
│   ├── diagrams/            # Excalidraw 图表
│   └── ...
├── books/
│   ├── ai-book/             # AI Agent 工程实践
│   ├── ecommerce-book/      # 电商系统架构设计
│   ├── system-design-book/  # 系统设计与面试指南
│   └── scripts/             # 共用构建脚本
├── public/                  # 生成的静态网站
├── _config.yml              # Hexo 配置
└── package.json             # 项目依赖
```

---

## 📚 核心内容领域

### 计算机基础

#### 操作系统与网络
- Linux 操作系统和常用命令（CPU、内存、网络、存储）
- 网络基础与协议原理
- 进程、线程、并发编程

#### 后台中间件
- **数据存储**：MySQL、Redis、Elasticsearch、HBase、S3、CDN
- **消息队列**：Kafka
- **缓存系统**
  - 本地缓存与 Remote Cache（双 buffer、LRU、并发安全）
  - 缓存更新机制（TTL、击穿、雪崩、singleflight）
  - 工具：groupcache

### 系统设计与架构

#### 网关与负载均衡
- DNS + LVS + Nginx 实现原理
- API 网关设计（Gin、Grpc-Gateway）
- 参考案例：Shopee、美团 Shepherd

#### 流量控制与可靠性
- **限流**：单机限流、分布式限流（Sentinel）
- **熔断**：熔断机制实现（Hystrix-Go）
- **重试**：指数退避重试（backoff）
- **监控**：ELK Stack、Prometheus、Grafana、Jaeger

#### 分布式服务
- gRPC 和 RPC 服务架构
- 服务发现与集群管理（Zookeeper、etcd）
- 工作流引擎与任务编排
- 定时任务调度（单机、分布式）
- 延时任务队列（Redis、LMSTFY）

#### 高级特性
- 规则引擎与风控（Gengine）
- 脚本执行引擎与低代码平台（Tengo、Anko）
- A/B Test 平台
- 大数据处理（Spark、Hive、Flink）

#### 部署与运维
- Docker 容器化
- Kubernetes 容器编排
- CI/CD（Jenkins、Git）

---

## 📝 写作规范

### Front Matter 要求
每篇文章必须包含：
```yaml
---
title: 文章标题
date: YYYY-MM-DD
categories:
  - 分类1
  - 分类2  # 最多 2 层
tags:
  - tag1
  - tag2  # 使用小写，多词用连字符连接
---
```

### 格式规范
- ✅ 中文文章使用中文标点，英文文章使用英文标点
- ✅ 代码块必须指定语言（```python、```go 等）
- ✅ 中英文之间加空格
- ✅ 图片使用相对路径（不要使用绝对路径）
- ✅ 标签用小写，多词用连字符（如 deep-learning）

### 文件命名规范
- **系列文章**：使用数字前缀（`22-ai-system-design.md`）
- **技术笔记**：描述性名称（`tensorflow-model-quantization.md`）
- **时效性强**：日期开头（`2026-03-07-OpenClaw深度调研.md`）

---

## 🛠️ 开发规范

### Git 规范
- 代码管理使用 Git
- 提交前运行 `npm run build` 确保无错误

### 编码规范
- Go：遵循 [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- API：RESTful API 设计规范、Swagger + YApi 文档

### 设计规范
- UML 标准
- DDD（Domain-Driven Design）设计模式
- 文档书写规范

---

## 🔨 推荐工具

| 工具 | 用途 |
|-----|------|
| **VS Code** | 轻量级代码编辑器，支持多语言 |
| **IntelliJ IDEA** | Java/Spark 开发 IDE |
| **DBeaver** | 数据库桌面管理工具 |
| **Postman** | API 测试与调试 |
| **Swagger** | API 文档与测试 |
| **PlantUML** | 时序图、架构图绘制 |
| **Diagrams.net** | 在线流程图、架构图工具 |
| **Charles Proxy** | HTTPS 代理抓包工具 |

---

## ⚠️ 常见陷阱

1. **修改 _config.yml**：必须重启 server
2. **新增文章**：需要运行 `hexo clean` 清理缓存
3. **部署前**：运行 `npm run build` 确保无错误
4. **Front Matter**：date 字段必须是字符串格式，不能是对象
5. **图片路径**：使用相对路径，不要使用绝对路径

---

## 📖 参考资源

### Hexo 配置完全指南
详见：[基于 Github 双分支和 Hexo 搭建博客](./source/_posts/other/基于Github双分支和Hexo搭建博客.md)

### AI 相关主题
- 计算机视觉（Computer Vision）
- 深度学习框架（TensorFlow）
- 编译器与优化（TVM）

### 系统设计主题
- 架构设计最佳实践
- 性能优化策略
- 可靠性工程

---

## 🔗 链接资源

### 技术文章
- [Shopee Games API 网关设计与实现](https://www.modb.pro/db/474513)
- [百亿规模 API 网关服务 Shepherd 的设计与实现](https://tech.meituan.com/2021/05/20/shepherd-api-gateway.html)

### 开源项目
- [gin - Go Web Framework](https://github.com/gin-gonic/gin)
- [Sentinel - 限流降级](https://github.com/alibaba/Sentinel)
- [Hystrix-Go - 熔断器](https://github.com/afex/hystrix-go)
- [Gengine - 规则引擎](https://github.com/bilibili/gengine)

### 相关技术
- [Zookeeper](https://zookeeper.apache.org)
- [etcd](https://coreos.com/etcd/docs/latest)
- [PlantUML](https://plantuml.com/)
- [Diagrams.net](https://app.diagrams.net/)

---

## 🤝 贡献指南

这是个人技术博客，但欢迎反馈和建议！

---

## 📄 许可证

本项目采用 MIT 许可证。详见 LICENSE 文件。

