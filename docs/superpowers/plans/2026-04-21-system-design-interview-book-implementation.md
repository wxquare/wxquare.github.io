# 程序员系统设计与面试指南 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将现有 Hexo 系统设计、计算机基础与 LeetCode 长文迁移为可构建、可搜索、可持续维护的 `system-design-book/` mdBook 电子书。

**Architecture:** 以现有 `system-design-book/` 骨架为载体，按 spec 的「大而全手册型、先工程后面试、双目标平衡」原则分批迁移源稿。正文优先保留工程主线，每章末尾统一收束为「本章面试题与追问」，并在迁移过程中同步修正 Front Matter、站内链接、图片路径与 Mermaid 渲染。

**Tech Stack:** mdBook, mdbook-mermaid, Markdown, GitHub Actions, GitHub Pages

---

## 目标目录对齐规则

执行本计划时，所有章节迁移都必须对齐 `docs/superpowers/specs/2026-04-21-system-design-interview-book-design.md` 中的 **3.7 目标目录（最终建议版）**。

- 第一优先级：保持章节主线与小标题顺序一致
- 第二优先级：保留源稿中的高价值图表、表格、案例与代码
- 第三优先级：在不破坏结构的前提下做必要裁剪或重命名

若源稿结构与目标目录不一致，处理顺序如下：

1. 先按目标目录重组一级、二级结构
2. 再把源稿内容填回对应小节
3. 最后将面试题、高频问答、速查表统一沉到「本章面试题与追问」

禁止事项：

- 不得直接把整篇源稿原样贴入章节文件而不重组结构
- 不得让面试题区块跑到工程主线之前
- 不得把 LeetCode 章节重新并回附录或单章大杂烩

---

## 文件结构

```
system-design-book/
├── book.toml
├── mermaid.min.js
├── mermaid-init.js
├── src/
│   ├── README.md
│   ├── SUMMARY.md
│   ├── part01/chapter01.md
│   ├── part01/chapter02.md
│   ├── part02/chapter03.md
│   ├── part02/chapter04.md
│   ├── part02/chapter05.md
│   ├── part02/chapter06.md
│   ├── part02/chapter07.md
│   ├── part03/chapter08.md
│   ├── part04/chapter09.md
│   ├── part05/chapter10.md
│   ├── part05/chapter11.md
│   ├── part05/chapter12.md
│   ├── part05/chapter13.md
│   ├── part05/chapter14.md
│   ├── part05/chapter15.md
│   ├── part06/chapter16.md
│   ├── part06/chapter17.md
│   ├── part06/chapter18.md
│   ├── part06/chapter19.md
│   ├── part06/chapter20.md
│   ├── part06/chapter21.md
│   └── appendix/
│       ├── glossary.md
│       ├── references.md
│       └── tooling.md
└── .github/workflows/
    └── deploy-system-design-book.yml
```

---

## Task 1: 校准骨架与迁移规则

**Files:**
- Modify: `system-design-book/src/README.md`
- Modify: `system-design-book/src/SUMMARY.md`
- Modify: `system-design-book/src/appendix/references.md`
- Review: `system-design-book/book.toml`
- Review: `docs/superpowers/specs/2026-04-21-system-design-interview-book-design.md`

- [ ] **Step 1: 对照 spec 复核现有骨架**

Run:
```bash
sed -n '1,240p' docs/superpowers/specs/2026-04-21-system-design-interview-book-design.md
sed -n '1,220p' system-design-book/src/README.md
sed -n '1,220p' system-design-book/src/SUMMARY.md
sed -n '1,220p' system-design-book/book.toml
```
Expected: 目录、站点路径 `/system-design-book/`、六部分结构与 spec 一致。

- [ ] **Step 2: 在 README 中写清三层阅读分层**

将 `system-design-book/src/README.md` 调整为：
- 主线必读：第 1–9 章
- 支撑扩展：第 10–15 章
- 编码补充：第 16–21 章
- 明确本书是「先工程、后面试」的手册型电子书

- [ ] **Step 3: 在 SUMMARY 中保持章节命名统一**

检查并修正：
- 章节命名与 spec 一致
- 第 16–21 章保持 LeetCode 拆章后的语义
- 附录名称保持「术语表 / 参考文献与外链 / 工具与构建说明」

- [ ] **Step 4: 给 references 附录建立汇总规则**

将 `system-design-book/src/appendix/references.md` 改成可持续维护的结构：
- 说明本附录汇总各章参考资料
- 约定以章节为单位收录
- 约定去重原则与外链书写方式

- [ ] **Step 5: 构建一次骨架**

Run:
```bash
cd system-design-book
mdbook build
```
Expected: Build succeeds and `book/index.html` regenerates without missing chapter errors.

- [ ] **Step 6: 提交骨架校准**

```bash
git add system-design-book/src/README.md system-design-book/src/SUMMARY.md system-design-book/src/appendix/references.md
git commit -m "docs(system-design-book): align skeleton with design spec"
```

---

## Task 2: 迁移第一部分导论与方法论

**Files:**
- Modify: `system-design-book/src/part01/chapter01.md`
- Modify: `system-design-book/src/part01/chapter02.md`
- Source: `source/_posts/system-design/00-system-design-overview.md`
- Source: `source/_posts/system-design/06-tech-design-methodology.md`

- [ ] **Step 1: 读取两篇源稿并确认一级结构**

Run:
```bash
sed -n '1,260p' source/_posts/system-design/00-system-design-overview.md
sed -n '1,260p' source/_posts/system-design/06-tech-design-methodology.md
```
Expected: 明确每篇文章的主段落、面试题区块、参考资料区块。

- [ ] **Step 2: 迁移第 1 章正文**

编辑 `system-design-book/src/part01/chapter01.md`：
- 删除占位说明
- 去掉 Front Matter
- 对齐以下小标题：
  - 系统设计的核心问题
  - 性能、容量与成本的基本权衡
  - 一致性、可用性与扩展性的取舍
  - 从单机到分布式的思维切换
  - 常见系统设计误区
  - 本章小结
- 将原文中的总表类面试内容改为指向各章或本章末尾的收束内容

- [ ] **Step 3: 迁移第 2 章正文**

编辑 `system-design-book/src/part01/chapter02.md`：
- 删除占位说明
- 对齐以下小标题：
  - 需求澄清与问题定义
  - 容量估算与约束识别
  - 架构拆解与核心链路
  - 数据流、状态流与边界划分
  - 高可用、高性能与扩展性设计
  - 方案评审与风险管理
  - 本章小结
- 章末单独保留「本章面试题与追问」

- [ ] **Step 4: 补齐两章的章末结构**

确保第 1、2 章都含有：
- 本章小结
- 本章面试题与追问
- 推荐阅读 / 参考资料（如源稿已有）

- [ ] **Step 5: 构建并检查导论部分**

Run:
```bash
cd system-design-book
mdbook build
```
Expected: `part01/chapter01.html` 与 `part01/chapter02.html` 渲染正常，无占位文本残留。

- [ ] **Step 6: 提交第一部分**

```bash
git add system-design-book/src/part01/chapter01.md system-design-book/src/part01/chapter02.md
git commit -m "feat(system-design-book): migrate introduction and methodology chapters"
```

---

## Task 3: 迁移核心中间件上半部分

**Files:**
- Modify: `system-design-book/src/part02/chapter03.md`
- Modify: `system-design-book/src/part02/chapter04.md`
- Modify: `system-design-book/src/part02/chapter05.md`
- Source: `source/_posts/system-design/01-middleware-mysql.md`
- Source: `source/_posts/system-design/02-middleware-redis.md`
- Source: `source/_posts/system-design/03-middleware-kafka.md`

- [ ] **Step 1: 读取 MySQL、Redis、Kafka 源稿的章节边界**

Run:
```bash
rg -n '^#|^## ' source/_posts/system-design/01-middleware-mysql.md
rg -n '^#|^## ' source/_posts/system-design/02-middleware-redis.md
rg -n '^#|^## ' source/_posts/system-design/03-middleware-kafka.md
```
Expected: 能识别每篇的原理、使用、生产实践、面试题区域。

- [ ] **Step 2: 迁移第 3 章 MySQL**

编辑 `system-design-book/src/part02/chapter03.md`：
- 对齐以下小标题：
  - 存储引擎与整体架构
  - 索引与查询执行
  - 事务、隔离级别与锁
  - 日志、复制与高可用
  - SQL 优化与分库分表
  - 常见线上问题与排查
  - 本章小结
- 再落到「本章面试题与追问」
- 保留高价值表格、示意图、案例

- [ ] **Step 3: 迁移第 4 章 Redis**

编辑 `system-design-book/src/part02/chapter04.md`：
- 对齐以下小标题：
  - Redis 的核心模型与高性能来源
  - 数据结构与典型使用模式
  - 持久化、复制与集群
  - 缓存设计与一致性问题
  - 分布式锁与常见误区
  - 热点问题与线上治理
  - 本章小结
- 再整理缓存击穿/雪崩/一致性等面试追问

- [ ] **Step 4: 迁移第 5 章 Kafka**

编辑 `system-design-book/src/part02/chapter05.md`：
- 对齐以下小标题：
  - Kafka 基础模型
  - 写入、消费与消费组
  - 顺序性、可靠性与幂等
  - 高吞吐设计原理
  - 积压、延迟与回压处理
  - 常见线上问题与排查
  - 本章小结
- 再收束为 MQ 设计与排障类面试问题

- [ ] **Step 5: 本轮构建并检查大文件渲染**

Run:
```bash
cd system-design-book
mdbook build
```
Expected: 三个章节均可渲染，目录跳转正常，代码块语言标注完整。

- [ ] **Step 6: 提交中间件上半部分**

```bash
git add system-design-book/src/part02/chapter03.md system-design-book/src/part02/chapter04.md system-design-book/src/part02/chapter05.md
git commit -m "feat(system-design-book): migrate mysql redis kafka chapters"
```

---

## Task 4: 迁移核心中间件下半部分与可靠性

**Files:**
- Modify: `system-design-book/src/part02/chapter06.md`
- Modify: `system-design-book/src/part02/chapter07.md`
- Modify: `system-design-book/src/part03/chapter08.md`
- Source: `source/_posts/system-design/04-middleware-elasticsearch.md`
- Source: `source/_posts/system-design/05-infrastructure-k8s-docker.md`
- Source: `source/_posts/system-design/07-system-reliability-engineering.md`

- [ ] **Step 1: 迁移第 6 章 Elasticsearch**

编辑 `system-design-book/src/part02/chapter06.md`：
- 对齐以下小标题：
  - 倒排索引与搜索基础
  - 索引、分片与副本
  - 写入链路与查询链路
  - Mapping、分词与相关性
  - 深分页与性能优化
  - 集群治理与常见问题
  - 本章小结
- 再追加搜索系统常见面试题与排障追问

- [ ] **Step 2: 迁移第 7 章 Kubernetes 与 Docker**

编辑 `system-design-book/src/part02/chapter07.md`：
- 对齐以下小标题：
  - 容器化的价值与边界
  - Docker 基础模型
  - Kubernetes 核心对象
  - 调度、发布与服务治理
  - 网络、配置与资源管理
  - 集群运维与常见故障
  - 本章小结
- 再归纳容器编排相关问答

- [ ] **Step 3: 迁移第 8 章系统可靠性工程**

编辑 `system-design-book/src/part03/chapter08.md`：
- 对齐以下小标题：
  - 可靠性的核心指标
  - 高可用设计原则
  - 超时、重试、幂等与隔离
  - 限流、熔断与降级
  - 监控、日志与追踪
  - 容量规划、压测与演练
  - 本章小结
- 再统一沉淀为可靠性专题面试题

- [ ] **Step 4: 构建并检查第二、三部分完整度**

Run:
```bash
cd system-design-book
mdbook build
```
Expected: 第 6–8 章目录正常，Mermaid 与表格渲染正常，无源稿占位句。

- [ ] **Step 5: 提交中间件下半部分与可靠性**

```bash
git add system-design-book/src/part02/chapter06.md system-design-book/src/part02/chapter07.md system-design-book/src/part03/chapter08.md
git commit -m "feat(system-design-book): migrate elasticsearch k8s and reliability chapters"
```

---

## Task 5: 迁移系统设计面试专题

**Files:**
- Modify: `system-design-book/src/part04/chapter09.md`
- Source: `source/_posts/system-design/08-system-design-interview.md`

- [ ] **Step 1: 读取系统设计面试专题源稿结构**

Run:
```bash
rg -n '^#|^## ' source/_posts/system-design/08-system-design-interview.md
```
Expected: 明确大节边界与原文中的题型组织方式。

- [ ] **Step 2: 迁移第 9 章正文**

编辑 `system-design-book/src/part04/chapter09.md`：
- 按 spec 保持独立成篇
- 对齐以下小标题：
  - 系统设计面试在考什么
  - 回答框架与表达顺序
  - 需求澄清与容量估算
  - 存储、缓存、消息与搜索的选型表达
  - 高可用与一致性的答题方式
  - 高频题型拆解
  - 面试复盘与提升方法
  - 本章小结
- 再保留具体题型、追问与回答策略

- [ ] **Step 3: 补充与前文章节的交叉引用**

在第 9 章中为相关题型补充章节引用，例如：
- 存储相关题型指向 MySQL / Redis
- 搜索相关题型指向 Elasticsearch
- 高可用相关题型指向系统可靠性工程

- [ ] **Step 4: 构建并检查交叉链接**

Run:
```bash
cd system-design-book
mdbook build
```
Expected: 第 9 章可正常打开，内部锚点与相对链接有效。

- [ ] **Step 5: 提交系统设计面试专题**

```bash
git add system-design-book/src/part04/chapter09.md
git commit -m "feat(system-design-book): migrate system design interview chapter"
```

---

## Task 6: 迁移计算机基础部分

**Files:**
- Modify: `system-design-book/src/part05/chapter10.md`
- Modify: `system-design-book/src/part05/chapter11.md`
- Modify: `system-design-book/src/part05/chapter12.md`
- Modify: `system-design-book/src/part05/chapter13.md`
- Modify: `system-design-book/src/part05/chapter14.md`
- Modify: `system-design-book/src/part05/chapter15.md`
- Source: `source/_posts/fundamentals/1-os-fundamentals.md`
- Source: `source/_posts/fundamentals/2-network-fundamentals.md`
- Source: `source/_posts/fundamentals/3-bash-shell.md`
- Source: `source/_posts/fundamentals/4-python-practice.md`
- Source: `source/_posts/fundamentals/5-cpp-practice.md`
- Source: `source/_posts/fundamentals/6-golang-practice.md`

- [ ] **Step 1: 批量读取 fundamentals 目录结构**

Run:
```bash
for f in source/_posts/fundamentals/*.md; do
  printf '\n=== %s ===\n' "$f"
  rg -n '^#|^## ' "$f"
done
```
Expected: 明确 6 篇基础稿件的标题层级和迁移重点。

- [ ] **Step 2: 迁移第 10、11 章**

编辑 `chapter10.md` 与 `chapter11.md`：
- `chapter10.md` 对齐：
  - 进程、线程与调度
  - 内存管理与虚拟内存
  - 文件系统与 I/O 模型
  - 并发同步与死锁
  - 工程场景中的操作系统知识
  - 本章小结
- `chapter11.md` 对齐：
  - 网络分层与通信基础
  - TCP、UDP 与连接管理
  - HTTP、HTTPS 与应用层协议
  - DNS、CDN 与代理
  - 常见网络问题排查
  - 本章小结
- 章末补「本章面试题与追问」

- [ ] **Step 3: 迁移第 12–15 章**

编辑 `chapter12.md`、`chapter13.md`、`chapter14.md`、`chapter15.md`：
- `chapter12.md` 对齐：
  - Shell 基础与命令组织
  - 文本处理与管道
  - grep、sed、awk 实战
  - 日志排查与自动化脚本
  - 本章小结
- `chapter13.md` 对齐：
  - Python 的工程用途
  - 常用语法与标准库
  - 自动化与数据处理
  - 调试、性能与常见坑
  - 本章小结
- `chapter14.md` 对齐：
  - 内存模型与对象生命周期
  - STL 容器与算法
  - 字符串与字符处理
  - 性能优化与常见陷阱
  - 本章小结
- `chapter15.md` 对齐：
  - Go 的核心设计
  - Goroutine 与 Channel
  - 接口、组合与错误处理
  - 网络服务与并发实践
  - 本章小结
- Shell / Python / C++ / Go 保留对系统设计读者有价值的实用内容
- 避免写成语言教程，强调实践、面试、常见坑

- [ ] **Step 4: 构建并检查第五部分是否仍是“支撑扩展”**

Run:
```bash
cd system-design-book
mdbook build
```
Expected: 第五部分可完整浏览，但不抢主线章节叙事中心。

- [ ] **Step 5: 提交计算机基础部分**

```bash
git add system-design-book/src/part05/chapter10.md system-design-book/src/part05/chapter11.md system-design-book/src/part05/chapter12.md system-design-book/src/part05/chapter13.md system-design-book/src/part05/chapter14.md system-design-book/src/part05/chapter15.md
git commit -m "feat(system-design-book): migrate fundamentals chapters"
```

---

## Task 7: 拆分并迁移 LeetCode 大文

**Files:**
- Modify: `system-design-book/src/part06/chapter16.md`
- Modify: `system-design-book/src/part06/chapter17.md`
- Modify: `system-design-book/src/part06/chapter18.md`
- Modify: `system-design-book/src/part06/chapter19.md`
- Modify: `system-design-book/src/part06/chapter20.md`
- Modify: `system-design-book/src/part06/chapter21.md`
- Source: `source/_posts/system-design/09-leetcode.md`

- [ ] **Step 1: 标出 LeetCode 源稿一级标题与编号问题**

Run:
```bash
rg -n '^## ' source/_posts/system-design/09-leetcode.md
```
Expected: 能准确定位「数据结构 / 基本算法 / 数学 / 搜索 / DP / C++ 字符处理函数速查」六大区块。

- [ ] **Step 2: 迁移第 16–20 章主体内容**

将 `09-leetcode.md` 按 spec 拆分到 `chapter16.md` 至 `chapter20.md`：
- `chapter16.md` 对齐：
  - 数组、链表、栈与队列
  - 哈希表、堆、树与图
  - 常见题型总结
- `chapter17.md` 对齐：
  - 双指针、滑动窗口与前缀和
  - 二分、排序与贪心
  - 递归与回溯基础
  - 常见题型总结
- `chapter18.md` 对齐：
  - 数论与位运算
  - 概率、组合与常见技巧
  - 易错点总结
- `chapter19.md` 对齐：
  - DFS 与 BFS
  - 回溯与剪枝
  - 状态搜索与图搜索
  - 常见题型总结
- `chapter20.md` 对齐：
  - DP 的核心思想
  - 线性 DP 与背包问题
  - 区间 DP、树形 DP、状态压缩 DP
  - 常见套路总结
- 每章只保留本章主题对应内容
- 将原稿重复编号的小节重编
- 保持算法模式速查风格

- [ ] **Step 3: 迁移第 21 章与参考资料处理**

编辑 `chapter21.md`：
- 对齐以下小标题：
  - 常见字符与字符串处理函数
  - 使用场景与示例
  - 易错点与延伸阅读
- 保留 C++ 字符处理函数速查主体
- 原文 `## 参考` 要么压缩为本章延伸阅读，要么并入附录 B

- [ ] **Step 4: 为算法部分补最小章首说明**

在第 16–21 章开头补一句：
- 本部分属于编码补充
- 与主线系统设计章节互补，不要求线性通读

- [ ] **Step 5: 构建并检查第六部分导航**

Run:
```bash
cd system-design-book
mdbook build
```
Expected: 第 16–21 章导航连续、无重复编号导致的目录混乱。

- [ ] **Step 6: 提交 LeetCode 拆章**

```bash
git add system-design-book/src/part06/chapter16.md system-design-book/src/part06/chapter17.md system-design-book/src/part06/chapter18.md system-design-book/src/part06/chapter19.md system-design-book/src/part06/chapter20.md system-design-book/src/part06/chapter21.md
git commit -m "feat(system-design-book): split and migrate leetcode section"
```

---

## Task 8: 统一清理链接、图片、附录与术语

**Files:**
- Modify: `system-design-book/src/appendix/glossary.md`
- Modify: `system-design-book/src/appendix/references.md`
- Modify: `system-design-book/src/appendix/tooling.md`
- Modify: `system-design-book/src/**/*.md`

- [ ] **Step 1: 扫描站内绝对路径与 Hexo 遗留链接**

Run:
```bash
rg -n '/20[0-9]{2}/|source/_posts|\.html\)|\]\(/' system-design-book/src
```
Expected: 找出所有需改成 mdBook 相对链接或外链说明的链接。

- [ ] **Step 2: 扫描图片与 Mermaid 风险点**

Run:
```bash
rg -n '!\[|```mermaid' system-design-book/src
```
Expected: 明确哪些图片要保留相对路径，哪些需要搬运到 `system-design-book/src/assets/`。

- [ ] **Step 3: 汇总 glossary / references / tooling**

编辑附录文件：
- `glossary.md`：补充书中高频术语
- `references.md`：按章节归档外链与推荐阅读
- `tooling.md`：写本地构建、Mermaid、部署说明

- [ ] **Step 4: 逐章修复链接与资源路径**

按扫描结果修改对应章节：
- 相对链接优先指向书内 `.md`
- 无法书内化的链接保留为外链
- 图片路径在 `mdbook build` 后必须可访问

- [ ] **Step 5: 提交清理与附录**

```bash
git add system-design-book/src
git commit -m "docs(system-design-book): normalize links assets and appendices"
```

---

## Task 9: 增加部署工作流

**Files:**
- Create: `.github/workflows/deploy-system-design-book.yml`
- Review: `.github/workflows/deploy-ai-book.yml`
- Review: `.github/workflows/deploy-ecommerce-book.yml`

- [ ] **Step 1: 读取现有两本书的工作流**

Run:
```bash
sed -n '1,240p' .github/workflows/deploy-ai-book.yml
sed -n '1,240p' .github/workflows/deploy-ecommerce-book.yml
```
Expected: 明确安装 mdBook、构建、发布到 GitHub Pages 的现有做法。

- [ ] **Step 2: 创建 system-design-book 专用 workflow**

新建 `.github/workflows/deploy-system-design-book.yml`：
- 监听 `system-design-book/**`
- 工作目录切到 `system-design-book`
- 执行 `mdbook build`
- 将 `system-design-book/book` 发布到 Pages 流程或现有发布分支方案

- [ ] **Step 3: 检查 workflow 不影响现有两本书**

Run:
```bash
git diff -- .github/workflows
```
Expected: 只新增 `deploy-system-design-book.yml`，不修改现有 workflow。

- [ ] **Step 4: 提交部署工作流**

```bash
git add .github/workflows/deploy-system-design-book.yml
git commit -m "ci(system-design-book): add deployment workflow"
```

---

## Task 10: 全书最终验收

**Files:**
- Review: `system-design-book/book.toml`
- Review: `system-design-book/src/**/*.md`
- Review: `.github/workflows/deploy-system-design-book.yml`

- [ ] **Step 1: 全量构建**

Run:
```bash
cd system-design-book
mdbook build
```
Expected: Build succeeds with all 21 chapters and 3 appendices present.

- [ ] **Step 2: 扫描占位文案是否残留**

Run:
```bash
rg -n '占位|迁移中|本章为目录占位|_占位|源稿' system-design-book/src
```
Expected: 仅允许保留必要的来源说明，不再出现“请将正文迁移至此文件”之类占位语。

- [ ] **Step 3: 抽查关键页面**

Run:
```bash
for f in part01/chapter01 part02/chapter03 part03/chapter08 part04/chapter09 part06/chapter20 appendix/references; do
  test -f "system-design-book/book/${f}.html" && echo "OK ${f}"
done
```
Expected: 输出 6 行 `OK ...`，关键章节与附录均已生成。

- [ ] **Step 4: 记录最终交付说明**

在本次交付说明中确认：
- 主线必读 / 支撑扩展 / 编码补充 已在前言或目录处体现
- 各章遵循先工程、后面试
- LeetCode 已独立成第六部分

- [ ] **Step 5: 提交最终验收**

```bash
git add system-design-book .github/workflows/deploy-system-design-book.yml
git commit -m "feat(system-design-book): complete initial mdbook migration"
```

---

## Self-Review

- Spec coverage: 已覆盖骨架校准、第一至第六部分迁移、LeetCode 拆章、附录整合、链接图片清理、部署工作流与最终构建验收。
- Placeholder scan: 计划中未使用 `TBD`、`TODO` 或“后续补充”式空描述；每个 task 都给出了文件范围、执行命令与验收预期。
- Consistency: 全文统一使用 `system-design-book/`、`本章面试题与追问`、`主线必读 / 支撑扩展 / 编码补充` 三组核心术语，与 spec 保持一致。
