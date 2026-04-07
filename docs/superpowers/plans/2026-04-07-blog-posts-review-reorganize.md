# 技术博客文档 Review 与重组实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 14 篇面试/工作经验文档全面重组：迁移目录、统一格式、修复质量问题、拆分臃肿文档、扩充单薄文档。

**Architecture:** 分 3 个阶段（P0 基础设施 → P1 中度改造 → P2 重度改造），每阶段独立可交付。每阶段结束后 `npm run clean && npm run build` 验证。

**Tech Stack:** Hexo 7.2.0, hexo-theme-next, Markdown, Git

**Spec:** `docs/superpowers/specs/2026-04-07-blog-posts-review-reorganize-design.md`

---

## Phase P0：基础设施

### Task 1: 创建 fundamentals 目录并迁移 6 篇基础文档

**Files:**
- Create: `source/_posts/fundamentals/` (directory)
- Move+Rename: 6 files from `source/_posts/system-design/` to `source/_posts/fundamentals/`

- [ ] **Step 1: 创建 fundamentals 目录**

```bash
mkdir -p source/_posts/fundamentals
```

- [ ] **Step 2: 迁移并重命名 6 篇文件**

```bash
git mv source/_posts/system-design/1-OS-interview.md source/_posts/fundamentals/1-os-fundamentals.md
git mv source/_posts/system-design/2-Internet-interview.md source/_posts/fundamentals/2-network-fundamentals.md
git mv source/_posts/system-design/3-bash-shell.md source/_posts/fundamentals/3-bash-shell.md
git mv source/_posts/system-design/4-python-experience.md source/_posts/fundamentals/4-python-practice.md
git mv source/_posts/system-design/5-cpp-interview.md source/_posts/fundamentals/5-cpp-practice.md
git mv source/_posts/system-design/6-golang-interview.md source/_posts/fundamentals/6-golang-practice.md
```

- [ ] **Step 3: 重命名 system-design 中的 2 个拼写错误文件**

```bash
git mv source/_posts/system-design/7-storage-desgin.md source/_posts/system-design/07-storage-mysql.md
git mv source/_posts/system-design/9-kafka-MQ.md source/_posts/system-design/09-kafka-mq.md
```

- [ ] **Step 4: 验证文件结构**

```bash
ls source/_posts/fundamentals/
ls source/_posts/system-design/
```

Expected: fundamentals/ 有 6 个文件，system-design/ 中不再有 1-6 开头的文件，且有 `07-storage-mysql.md` 和 `09-kafka-mq.md`。

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "refactor: migrate fundamentals docs to dedicated directory and fix filenames"
```

---

### Task 2: 修复 fundamentals/ 6 篇文档的 Front Matter

**Files:**
- Modify: `source/_posts/fundamentals/1-os-fundamentals.md`
- Modify: `source/_posts/fundamentals/2-network-fundamentals.md`
- Modify: `source/_posts/fundamentals/3-bash-shell.md`
- Modify: `source/_posts/fundamentals/4-python-practice.md`
- Modify: `source/_posts/fundamentals/5-cpp-practice.md`
- Modify: `source/_posts/fundamentals/6-golang-practice.md`

- [ ] **Step 1: 修复 `1-os-fundamentals.md` Front Matter**

将原 Front Matter 替换为：

```yaml
---
title: 计算机基础：操作系统
date: 2024-03-01
categories:
- 计算机基础
tags:
- 操作系统
- 进程
- 内存管理
- 面试
toc: true
---
```

- [ ] **Step 2: 修复 `2-network-fundamentals.md` Front Matter**

```yaml
---
title: 计算机基础：计算机网络实践
date: 2024-03-02
categories:
- 计算机基础
tags:
- 计算机网络
- TCP
- HTTP
- 面试
toc: true
---
```

- [ ] **Step 3: 修复 `3-bash-shell.md` Front Matter**

```yaml
---
title: 编程语言：bash shell 实践
date: 2024-03-03
categories:
- 编程语言
tags:
- bash
- linux
- shell
- 运维
toc: true
---
```

- [ ] **Step 4: 修复 `4-python-practice.md` Front Matter**

将原 Front Matter 替换为以下内容，并**删除文件中约 line 39-43 处的第二段 `---...---` Front Matter 块**（搜索第二个以 `---` 开头的 YAML 块并删除）：

```yaml
---
title: 编程语言：Python 实践记录
date: 2024-03-04
categories:
- 编程语言
tags:
- python
- swig
- 性能优化
- C++调用
toc: true
---
```

- [ ] **Step 5: 修复 `5-cpp-practice.md` Front Matter**

```yaml
---
title: 编程语言：C/C++ 实践
date: 2024-03-05
categories:
- 编程语言
tags:
- C++
- STL
- 内存管理
- 面试
toc: true
---
```

- [ ] **Step 6: 修复 `6-golang-practice.md` Front Matter**

检查文件开头是否有空行（在 `---` 之前），如有则删除。替换 Front Matter：

```yaml
---
title: 编程语言：Go 实践
date: 2024-03-06
categories:
- 编程语言
tags:
- golang
- 并发
- GMP
- 内存管理
- 面试
toc: true
---
```

- [ ] **Step 7: Commit**

```bash
git add source/_posts/fundamentals/
git commit -m "fix: standardize front matter for all fundamentals docs"
```

---

### Task 3: 修复 system-design/ 8 篇文档的 Front Matter

**Files:**
- Modify: `source/_posts/system-design/00-system-design-fundamentals.md`
- Modify: `source/_posts/system-design/07-storage-mysql.md`
- Modify: `source/_posts/system-design/08-cache-redis.md` (检查，可能不需改)
- Modify: `source/_posts/system-design/09-kafka-mq.md`
- Modify: `source/_posts/system-design/10-elasticsearch.md`
- Modify: `source/_posts/system-design/11-k8s-docker.md`
- Modify: `source/_posts/system-design/12-tech-design.md`
- Modify: `source/_posts/system-design/14-system-reliability.md` (检查，可能不需改)
- Check: `source/_posts/system-design/17-high-frequency-system-design.md` (检查，可能不需改)

- [ ] **Step 1: 修复 `07-storage-mysql.md` Front Matter**

```yaml
---
title: 中间件 - 存储与 MySQL 数据库
date: 2024-03-04
categories:
- 系统设计
tags:
- MySQL
- 数据库
- 索引
- 分库分表
- 存储设计
toc: true
---
```

- [ ] **Step 2: 修复 `09-kafka-mq.md` Front Matter**

```yaml
---
title: 中间件 - 异步和消息队列
date: 2024-03-10
categories:
- 系统设计
tags:
- Kafka
- 消息队列
- 异步
- 分布式
toc: true
---
```

- [ ] **Step 3: 修复 `10-elasticsearch.md` Front Matter**

```yaml
---
title: 中间件 - 搜索和 Elasticsearch
date: 2024-03-07
categories:
- 系统设计
tags:
- Elasticsearch
- 搜索引擎
- 倒排索引
- DSL
toc: true
---
```

- [ ] **Step 4: 修复 `11-k8s-docker.md` Front Matter**

```yaml
---
title: 互联网基础设施：Kubernetes 与 Docker 实践
date: 2024-12-20
categories:
- 系统设计
tags:
- kubernetes
- docker
- 容器化
- 云原生
toc: true
---
```

- [ ] **Step 5: 修复 `12-tech-design.md` Front Matter**

```yaml
---
title: 互联网系统设计 - 概述和技术方案写作
date: 2025-04-01
categories:
- 系统设计
tags:
- 技术方案
- 架构设计
- API设计
- 微服务
toc: true
---
```

- [ ] **Step 6: 检查 `00-system-design-fundamentals.md`**

确认已有 `tags` 和 `toc`。如果缺少 `toc: true`，添加之。

- [ ] **Step 7: 检查 `08-cache-redis.md`、`14-system-reliability.md`、`17-high-frequency-system-design.md`**

这 3 篇已有较完整 Front Matter，确认 `toc: true` 存在即可。

- [ ] **Step 8: Commit**

```bash
git add source/_posts/system-design/
git commit -m "fix: standardize front matter for all system-design docs"
```

---

### Task 4: P0 构建验证

- [ ] **Step 1: 执行构建验证**

```bash
npm run clean && npm run build
```

Expected: 构建成功，无错误。如有错误，排查并修复（常见问题：Front Matter 格式错误、文件路径问题）。

- [ ] **Step 2: 如有构建错误，修复后再次提交**

```bash
git add -A
git commit -m "fix: resolve build errors after P0 restructure"
```

---

## Phase P1：中度改造

### Task 5: 更新 00-system-design-fundamentals 导航链接

**Files:**
- Modify: `source/_posts/system-design/00-system-design-fundamentals.md`

- [ ] **Step 1: 更新学习路径中的文章链接**

搜索文件中所有指向旧路径的内部链接，更新为新路径：

需要更新的链接映射：
- `/system-design/7-storage-desgin/` → `/system-design/07-storage-mysql/`
- `/system-design/9-kafka-MQ/` → `/system-design/09-kafka-mq/`

同时检查其他内部链接是否指向正确的路径。

- [ ] **Step 2: 精简"学员评价"章节**

删除 `## 🌟 成功学员经验分享` 整个章节（约 line 299-305），这些是虚构内容。

- [ ] **Step 3: Commit**

```bash
git add source/_posts/system-design/00-system-design-fundamentals.md
git commit -m "fix: update navigation links and remove placeholder content in fundamentals guide"
```

---

### Task 6: 修复 07, 10, 11 的 typo 和格式问题

**Files:**
- Modify: `source/_posts/system-design/07-storage-mysql.md`
- Modify: `source/_posts/system-design/10-elasticsearch.md`
- Modify: `source/_posts/system-design/11-k8s-docker.md`

- [ ] **Step 1: 修复 `07-storage-mysql.md`**

- 将文章标题 `title` 中如果引用旧文件名的地方统一
- 检查文章开头几节的纯链接罗列，添加一句话描述

- [ ] **Step 2: 修复 `11-k8s-docker.md` typo**

搜索并替换 `iptabels` → `iptables`（全文替换）。

- [ ] **Step 3: 验证 `10-elasticsearch.md` 的 `<details>` 标签**

确认 `<details>` + `<summary>` HTML 标签在 Hexo 构建后能正确渲染。如果不能，改为标准 Markdown 格式。

- [ ] **Step 4: Commit**

```bash
git add source/_posts/system-design/07-storage-mysql.md source/_posts/system-design/10-elasticsearch.md source/_posts/system-design/11-k8s-docker.md
git commit -m "fix: typos and formatting in storage, elasticsearch, and k8s docs"
```

---

### Task 7: 拆分精简 12-tech-design

**Files:**
- Modify: `source/_posts/system-design/12-tech-design.md`

- [ ] **Step 1: 删除与 14-system-reliability 重复的"系统稳定性建设"章节**

删除 `## 系统稳定性建设` 整个章节（约从 line 1252 到 line 1327），包括其下的所有子章节。在原位置添加一行链接引用：

```markdown
## 系统稳定性建设

> 详见 [互联网系统稳定性建设：方法论与实践](/2025/05/15/system-design/14-system-reliability/)
```

- [ ] **Step 2: 精简"中间件和存储"章节**

将 `## 中间件和存储` 下的详细内容（约 line 906 到 line 1180）精简为链接索引：

```markdown
## 中间件和存储

各中间件详细原理与实践参见专题文章：

- [存储与 MySQL 数据库](/2024/03/04/system-design/07-storage-mysql/)
- [Redis 原理与实践](/2024/03/06/system-design/08-cache-redis/)
- [异步和消息队列](/2024/03/10/system-design/09-kafka-mq/)
- [搜索和 Elasticsearch](/2024/03/07/system-design/10-elasticsearch/)
- [Kubernetes 与 Docker 实践](/2024/12/20/system-design/11-k8s-docker/)
```

保留 `### 如何选择存储组件` 图片和简要说明（约 5-10 行），其余删除。

- [ ] **Step 3: 精简"大数据存储和计算"章节**

`## 大数据存储和计算` 仅有 6 行外链，保留不变。

- [ ] **Step 4: 精简"云原生和服务部署CI/CD"章节**

保留不变（仅 4 行链接）。

- [ ] **Step 5: 验证精简后的文件行数**

```bash
wc -l source/_posts/system-design/12-tech-design.md
```

Expected: ≤ 800 行。

- [ ] **Step 6: Commit**

```bash
git add source/_posts/system-design/12-tech-design.md
git commit -m "refactor: slim down tech-design doc, deduplicate with reliability and middleware articles"
```

---

### Task 8: 修复 14-system-reliability 占位链接

**Files:**
- Modify: `source/_posts/system-design/14-system-reliability.md`

- [ ] **Step 1: 修复占位链接**

搜索 `xxx` 占位符（在"学习资料"章节的"字节跳动 SRE 实践"链接中），替换为有效链接或删除该条目：

将：
```markdown
- [字节跳动 SRE 实践](https://mp.weixin.qq.com/s/xxx)
```

替换为：
```markdown
- 字节跳动 SRE 实践（链接已失效，可搜索"字节跳动 SRE"获取相关文章）
```

- [ ] **Step 2: Commit**

```bash
git add source/_posts/system-design/14-system-reliability.md
git commit -m "fix: replace broken placeholder link in system-reliability doc"
```

---

### Task 9: P1 构建验证

- [ ] **Step 1: 执行构建验证**

```bash
npm run clean && npm run build
```

Expected: 构建成功，无错误。

- [ ] **Step 2: 如有构建错误，修复后提交**

```bash
git add -A
git commit -m "fix: resolve build errors after P1 improvements"
```

---

## Phase P2：重度改造

### Task 10: 扩充 09-kafka-mq

**Files:**
- Modify: `source/_posts/system-design/09-kafka-mq.md`

- [ ] **Step 1: 修复现有内容问题**

- 删除空标题（搜索 `##` 后面没有内容的行）
- 修正 typo：`mannul` → `manual`
- 修复 GitHub blob 图片链接：如果图片是 GitHub blob URL，改为本地 `/images/` 路径或 raw URL

- [ ] **Step 2: 扩充内容 — Kafka 核心架构**

在现有 Kafka 特点章节后，补充以下内容（以 08-cache-redis 风格为模板）：

```markdown
## Kafka 核心架构

### 整体架构

Kafka 采用分布式发布-订阅消息系统架构，核心组件包括 Producer、Broker、Consumer、ZooKeeper/KRaft。

### Partition 与副本机制

每个 Topic 可分为多个 Partition，Partition 是 Kafka 并行度的基本单位。

| 概念 | 说明 |
|------|------|
| **Leader** | 负责读写的主副本 |
| **Follower** | 同步数据的从副本 |
| **ISR** | In-Sync Replicas，与 Leader 保持同步的副本集合 |
| **OSR** | Out-of-Sync Replicas，落后的副本 |

### Consumer Group 与 Rebalance

消费者组内的消费者共同消费一个 Topic，每个 Partition 只能被组内一个消费者消费。

Rebalance 触发条件：
- 消费者加入/离开消费者组
- 订阅的 Topic Partition 数量变化
- 消费者心跳超时

### 消息投递语义

| 语义 | 生产端配置 | 消费端配置 | 适用场景 |
|------|-----------|-----------|----------|
| **At-most-once** | acks=0 | 自动提交 offset | 日志采集（允许丢失） |
| **At-least-once** | acks=all + 重试 | 手动提交 offset | 订单处理（不允许丢失） |
| **Exactly-once** | 幂等生产者 + 事务 | 事务性消费 | 金融场景 |
```

- [ ] **Step 3: 扩充内容 — 消息不丢失保障**

```markdown
## 消息不丢失全链路保障

### 生产端

| 配置项 | 推荐值 | 说明 |
|--------|--------|------|
| `acks` | `all` | 等待所有 ISR 副本确认 |
| `retries` | `3` | 发送失败重试次数 |
| `max.in.flight.requests.per.connection` | `1` | 保证消息顺序（配合重试） |

### Broker 端

| 配置项 | 推荐值 | 说明 |
|--------|--------|------|
| `min.insync.replicas` | `2` | 至少 2 个副本同步才允许写入 |
| `unclean.leader.election.enable` | `false` | 禁止非 ISR 副本成为 Leader |
| `log.flush.interval.messages` | 根据场景 | 刷盘策略 |

### 消费端

- 手动提交 offset（`enable.auto.commit=false`）
- 消费成功后再提交，失败则重试
- 消费逻辑做幂等处理
```

- [ ] **Step 4: 扩充内容 — 性能调优与 Go 示例**

```markdown
## 性能调优

### Kafka 高性能原因

1. **顺序写磁盘**：追加写 log 文件，避免随机 IO
2. **Page Cache**：利用 OS 页缓存，减少用户态拷贝
3. **零拷贝（sendfile）**：数据直接从磁盘到网卡，绕过用户空间
4. **批量发送 + 压缩**：减少网络往返

### 消费积压排查

1. 检查消费者 lag：`kafka-consumer-groups --describe --group xxx`
2. 常见原因：消费逻辑慢（DB/外部调用）、Partition 数不够、消费者数不匹配
3. 应急方案：临时扩容消费者（不超过 Partition 数）、跳过非关键消息

### Go 生产者示例

```go
import "github.com/segmentio/kafka-go"

func ProduceMessage(topic, key, value string) error {
    w := &kafka.Writer{
        Addr:         kafka.TCP("localhost:9092"),
        Topic:        topic,
        Balancer:     &kafka.LeastBytes{},
        RequiredAcks: kafka.RequireAll,
        MaxAttempts:  3,
    }
    defer w.Close()

    return w.WriteMessages(context.Background(),
        kafka.Message{
            Key:   []byte(key),
            Value: []byte(value),
        },
    )
}
```

### Go 消费者示例

```go
func ConsumeMessages(topic, groupID string) {
    r := kafka.NewReader(kafka.ReaderConfig{
        Brokers:  []string{"localhost:9092"},
        GroupID:  groupID,
        Topic:    topic,
        MinBytes: 10e3,
        MaxBytes: 10e6,
    })
    defer r.Close()

    for {
        m, err := r.ReadMessage(context.Background())
        if err != nil {
            log.Printf("read error: %v", err)
            break
        }
        log.Printf("topic=%s partition=%d offset=%d key=%s value=%s",
            m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
    }
}
```
```

- [ ] **Step 5: 整理参考资料到文末**

将散落在文中的外部链接统一移到文末 `## 参考资料` 章节。

- [ ] **Step 6: 验证行数**

```bash
wc -l source/_posts/system-design/09-kafka-mq.md
```

Expected: ≥ 600 行。

- [ ] **Step 7: Commit**

```bash
git add source/_posts/system-design/09-kafka-mq.md
git commit -m "feat: significantly expand kafka-mq doc with architecture, semantics, tuning, and Go examples"
```

---

### Task 11: fundamentals 内容质量优化

**Files:**
- Modify: `source/_posts/fundamentals/1-os-fundamentals.md`
- Modify: `source/_posts/fundamentals/2-network-fundamentals.md`
- Modify: `source/_posts/fundamentals/3-bash-shell.md`
- Modify: `source/_posts/fundamentals/4-python-practice.md`
- Modify: `source/_posts/fundamentals/6-golang-practice.md`

- [ ] **Step 1: 修复 `1-os-fundamentals.md`**

- 修复重复编号：搜索两个 `7.` 开头的条目，将第二个改为正确编号
- 搜索并修正"进bai程"编码乱码为"进程"
- 确保所有代码块标注语言

- [ ] **Step 2: 修复 `2-network-fundamentals.md`**

- 修复重复编号：搜索两个 `14.` 开头的条目，将第二个改为正确编号
- 搜索并修正"小路"（如果是笔误应为"效率"之类的词）
- 确保所有代码块标注语言

- [ ] **Step 3: 修复 `3-bash-shell.md`**

- 搜索并替换 `comamand1` → `command1`
- 搜索并替换 `taf` → `tar`（如果是 tar 的笔误）
- 移除所有 `<font color=red>...</font>` 标签，改为 Markdown **加粗** 标记

- [ ] **Step 4: 修复 `4-python-practice.md`**

- 确认第二段 Front Matter 已在 Task 2 中删除
- 搜索并替换 `JIL` → `JIT`
- 检查文件中约 line 359 处是否还有第三个 `---` 块，如有删除

- [ ] **Step 5: 修复 `6-golang-practice.md`**

- 移除所有 `<font color=red>...</font>` 标签，改为 Markdown **加粗** 标记

- [ ] **Step 6: Commit**

```bash
git add source/_posts/fundamentals/
git commit -m "fix: content quality improvements for fundamentals docs (typos, encoding, formatting)"
```

---

### Task 12: P2 构建验证 + 最终检查

- [ ] **Step 1: 执行最终构建验证**

```bash
npm run clean && npm run build
```

Expected: 构建成功，无错误。

- [ ] **Step 2: 验证成功标准检查清单**

逐项检查：

```bash
# 1. fundamentals 目录包含 6 篇文档
ls source/_posts/fundamentals/ | wc -l
# Expected: 6

# 2. 无旧文件名残留
ls source/_posts/system-design/ | grep -E "^[1-6]-|desgin|kafka-MQ"
# Expected: 无输出

# 3. 09-kafka-mq 行数
wc -l source/_posts/system-design/09-kafka-mq.md
# Expected: >= 600

# 4. 12-tech-design 行数
wc -l source/_posts/system-design/12-tech-design.md
# Expected: <= 800
```

- [ ] **Step 3: 如有构建错误或检查不通过，修复后提交**

```bash
git add -A
git commit -m "fix: final adjustments after P2 review"
```

---

## 任务依赖关系

```
Phase P0 (Tasks 1-4)
  Task 1: 目录迁移
  Task 2: fundamentals Front Matter (依赖 Task 1)
  Task 3: system-design Front Matter (依赖 Task 1)
  Task 4: P0 构建验证 (依赖 Task 2, 3)

Phase P1 (Tasks 5-9)
  Task 5: 导航链接更新 (依赖 P0)
  Task 6: typo 修复 (独立)
  Task 7: 12 拆分精简 (独立)
  Task 8: 14 链接修复 (独立)
  Task 9: P1 构建验证 (依赖 Task 5-8)

Phase P2 (Tasks 10-12)
  Task 10: 09 扩充 (独立)
  Task 11: fundamentals 内容优化 (独立)
  Task 12: P2 构建验证 (依赖 Task 10-11)
```

Tasks 5/6/7/8 可并行执行。Tasks 10/11 可并行执行。
