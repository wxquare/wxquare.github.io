# 电商系统设计系列文章优化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 7 篇电商系统设计文章优化为 6 篇，统一编号 20-25，消除内容重复，建立系列导航。

**Architecture:** 温和重组方案 — 文件重命名（含 1 篇非电商文章移位）→ Front Matter 和导航块统一 → 文章 20 吸收原文章 22 独特内容 → 文章 25 去重 → 构建验证。

**Tech Stack:** Hexo 7.2.0, Git, npm

**Spec:** `docs/superpowers/specs/2026-04-07-ecommerce-series-optimization-design.md`

---

### Task 1: 备份原文章 22 内容并删除文件

**Files:**
- Delete: `source/_posts/system-design/22-unified-product-inventory-pricing-system.md`
- Create: `docs/superpowers/backup/22-unified-product-inventory-pricing-system.md` (备份，供后续 Task 提取内容)

- [ ] **Step 1: 备份文章 22 到 docs 目录**

```bash
mkdir -p docs/superpowers/backup
cp source/_posts/system-design/22-unified-product-inventory-pricing-system.md docs/superpowers/backup/
```

- [ ] **Step 2: 从 source 中删除文章 22**

```bash
git rm source/_posts/system-design/22-unified-product-inventory-pricing-system.md
```

- [ ] **Step 3: 验证文件已删除**

```bash
ls source/_posts/system-design/22-*
```

Expected: `No such file or directory`

- [ ] **Step 4: Commit**

```bash
git add docs/superpowers/backup/22-unified-product-inventory-pricing-system.md
git commit -m "chore: backup and remove article 22 (content to be merged into article 20)"
```

---

### Task 2: 文件重命名（无冲突顺序）

**Files:**
- Rename: 7 个文件（6 电商 + 1 广告定价）

重命名必须按以下顺序执行，以避免编号冲突：

- [ ] **Step 1: 19 → 21（21 号空闲）**

```bash
cd source/_posts/system-design
git mv 19-listing-upload-system-design.md 21-ecommerce-listing.md
```

- [ ] **Step 2: 20 → 19（19 号已空出）**

```bash
git mv 20-ad-realtime-pricing-optimization.md 19-ad-realtime-pricing.md
```

- [ ] **Step 3: 13 → 20（20 号已空出）**

```bash
git mv 13-e-commerce.md 20-ecommerce-overview.md
```

- [ ] **Step 4: 18 → 22（22 号已在 Task 1 删除后空出）**

```bash
git mv 18-inventory-system-design.md 22-ecommerce-inventory.md
```

- [ ] **Step 5: 25 → 临时文件（打破 23→25→24 循环依赖）**

```bash
git mv 25-ddd-pricing-engine-practice.md _temp-ecommerce-pricing-ddd.md
```

- [ ] **Step 6: 23 → 25（25 号已空出）**

```bash
git mv 23-b-side-operations-system.md 25-ecommerce-b-side-ops.md
```

- [ ] **Step 7: 24 → 23（23 号已空出）**

```bash
git mv 24-pricing-engine-design.md 23-ecommerce-pricing-engine.md
```

- [ ] **Step 8: 临时文件 → 24（24 号已空出）**

```bash
git mv _temp-ecommerce-pricing-ddd.md 24-ecommerce-pricing-ddd.md
```

- [ ] **Step 9: 回到项目根目录，验证文件结构**

```bash
cd /Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io
ls source/_posts/system-design/{19,20,21,22,23,24,25}*.md
```

Expected output（7 个文件）:
```
source/_posts/system-design/19-ad-realtime-pricing.md
source/_posts/system-design/20-ecommerce-overview.md
source/_posts/system-design/21-ecommerce-listing.md
source/_posts/system-design/22-ecommerce-inventory.md
source/_posts/system-design/23-ecommerce-pricing-engine.md
source/_posts/system-design/24-ecommerce-pricing-ddd.md
source/_posts/system-design/25-ecommerce-b-side-ops.md
```

- [ ] **Step 10: Commit**

```bash
git add -A source/_posts/system-design/
git commit -m "chore: rename ecommerce series to 20-25, move ad-pricing to 19"
```

---

### Task 3: 更新非电商文章 19（原 20）的 Front Matter

**Files:**
- Modify: `source/_posts/system-design/19-ad-realtime-pricing.md`

- [ ] **Step 1: 更新 Front Matter**

文件 `source/_posts/system-design/19-ad-realtime-pricing.md` 的 Front Matter 保持不变（标题和内容不改，仅文件名变了）。确认无需修改：

```bash
head -12 source/_posts/system-design/19-ad-realtime-pricing.md
```

Expected: Front Matter 中 title 仍为原标题，无需改动。

- [ ] **Step 2: Commit（如有改动）**

如 Step 1 确认无需改动，跳过此步。

---

### Task 4: 更新文章 21（原 19）— 商品上架系统

**Files:**
- Modify: `source/_posts/system-design/21-ecommerce-listing.md`

- [ ] **Step 1: 替换 Front Matter**

将原 Front Matter:

```yaml
---
title: 多品类统一商品上架系统设计：电商·虚拟商品·本地生活
date: 2025-06-28
categories:
- 系统设计
tags:
- 商品上架
- 电商
- 系统设计
- 状态机
toc: true
---
```

替换为:

```yaml
---
title: 电商系统设计（二）：商品上架系统
date: 2025-06-28
categories:
- 系统设计
tags:
- e-commerce
- system-design
- listing
- state-machine
- saga
toc: true
---
```

- [ ] **Step 2: 在 `<!-- toc -->` 之后插入系列导航块和开头引用**

在 `<!-- toc -->` 标签之后、正文第一个 `##` 标题之前，插入：

```markdown

> **电商系统设计系列**
> - [（一）全景概览与领域划分](/system-design/20-ecommerce-overview/)
> - **（二）商品上架系统**（本文）
> - [（三）库存系统](/system-design/22-ecommerce-inventory/)
> - [（四）计价引擎](/system-design/23-ecommerce-pricing-engine/)
> - [（五）计价系统 DDD 实践](/system-design/24-ecommerce-pricing-ddd/)
> - [（六）B 端运营系统](/system-design/25-ecommerce-b-side-ops/)

本文是电商系统设计系列的第二篇，建议先阅读[（一）全景概览与领域划分](/system-design/20-ecommerce-overview/)了解整体架构。

```

- [ ] **Step 3: 在文章末尾添加系列引用**

在文章最后（参考资料之后或文末）追加：

```markdown

---

> **系列导航**
> 上架完成后，商品的库存管理详见[（三）库存系统](/system-design/22-ecommerce-inventory/)，价格配置详见[（四）计价引擎](/system-design/23-ecommerce-pricing-engine/)。
```

- [ ] **Step 4: 构建验证**

```bash
npx hexo clean && npx hexo generate
```

Expected: 无报错

- [ ] **Step 5: Commit**

```bash
git add source/_posts/system-design/21-ecommerce-listing.md
git commit -m "docs: update article 21 - listing system title, navigation and cross-references"
```

---

### Task 5: 更新文章 22（原 18）— 库存系统

**Files:**
- Modify: `source/_posts/system-design/22-ecommerce-inventory.md`

- [ ] **Step 1: 替换 Front Matter**

将原 Front Matter:

```yaml
---
title: 多品类统一库存系统设计：电商·虚拟商品·本地生活
date: 2025-06-28
categories:
- 系统设计
tags:
- 库存系统
- 电商
- 系统设计
- 高并发
toc: true
---
```

替换为:

```yaml
---
title: 电商系统设计（三）：库存系统
date: 2025-06-28
categories:
- 系统设计
tags:
- e-commerce
- system-design
- inventory
- redis
- strategy-pattern
toc: true
---
```

- [ ] **Step 2: 在 `<!-- toc -->` 之后插入系列导航块和开头引用**

在 `<!-- toc -->` 标签之后、正文第一个 `##` 标题之前，插入：

```markdown

> **电商系统设计系列**
> - [（一）全景概览与领域划分](/system-design/20-ecommerce-overview/)
> - [（二）商品上架系统](/system-design/21-ecommerce-listing/)
> - **（三）库存系统**（本文）
> - [（四）计价引擎](/system-design/23-ecommerce-pricing-engine/)
> - [（五）计价系统 DDD 实践](/system-design/24-ecommerce-pricing-ddd/)
> - [（六）B 端运营系统](/system-design/25-ecommerce-b-side-ops/)

本文是电商系统设计系列的第三篇，聚焦库存系统的设计与实现。

```

- [ ] **Step 3: 在文章末尾添加系列引用**

```markdown

---

> **系列导航**
> 库存与价格在下单时的协作流程，详见[（一）全景概览与领域划分](/system-design/20-ecommerce-overview/)中的 C 端用户旅程章节。
```

- [ ] **Step 4: Commit**

```bash
git add source/_posts/system-design/22-ecommerce-inventory.md
git commit -m "docs: update article 22 - inventory system title, navigation and cross-references"
```

---

### Task 6: 更新文章 23（原 24）— 计价引擎

**Files:**
- Modify: `source/_posts/system-design/23-ecommerce-pricing-engine.md`

- [ ] **Step 1: 替换 Front Matter**

将原 Front Matter:

```yaml
---
title: 电商系统价格计算引擎设计与实现
date: 2026-02-27
categories:
- 系统设计
tags:
- 价格引擎
- 电商系统
- 计价中心
- 营销优惠
- 系统设计
toc: true
---
```

替换为:

```yaml
---
title: 电商系统设计（四）：计价引擎
date: 2026-02-27
categories:
- 系统设计
tags:
- e-commerce
- system-design
- pricing
- multi-level-cache
- degradation
toc: true
---
```

- [ ] **Step 2: 在 `<!-- toc -->` 之后插入系列导航块和开头引用**

在 `<!-- toc -->` 标签之后、正文第一个 `##` 标题之前，插入：

```markdown

> **电商系统设计系列**
> - [（一）全景概览与领域划分](/system-design/20-ecommerce-overview/)
> - [（二）商品上架系统](/system-design/21-ecommerce-listing/)
> - [（三）库存系统](/system-design/22-ecommerce-inventory/)
> - **（四）计价引擎**（本文）
> - [（五）计价系统 DDD 实践](/system-design/24-ecommerce-pricing-ddd/)
> - [（六）B 端运营系统](/system-design/25-ecommerce-b-side-ops/)

本文是电商系统设计系列的第四篇，详述计价引擎的设计与实现。

```

- [ ] **Step 3: 在文章末尾添加系列引用**

```markdown

---

> **系列导航**
> 计价系统的领域建模和 DDD 实践，详见[（五）计价系统 DDD 实践](/system-design/24-ecommerce-pricing-ddd/)。
```

- [ ] **Step 4: Commit**

```bash
git add source/_posts/system-design/23-ecommerce-pricing-engine.md
git commit -m "docs: update article 23 - pricing engine title, navigation and cross-references"
```

---

### Task 7: 更新文章 24（原 25）— 计价系统 DDD 实践

**Files:**
- Modify: `source/_posts/system-design/24-ecommerce-pricing-ddd.md`

- [ ] **Step 1: 替换 Front Matter**

将原 Front Matter:

```yaml
---
title: 领域驱动设计在电商计价系统中的实践
date: 2026-03-14
categories:
- 系统设计
tags:
- DDD
- 领域驱动设计
- 计价系统
- 架构设计
toc: true
---
```

替换为:

```yaml
---
title: 电商系统设计（五）：计价系统 DDD 实践
date: 2026-03-14
categories:
- 系统设计
tags:
- e-commerce
- system-design
- ddd
- hexagonal-architecture
- aggregate-root
toc: true
---
```

- [ ] **Step 2: 在 `<!-- toc -->` 之后插入系列导航块和开头引用**

在 `<!-- toc -->` 标签之后、正文第一个 `##` 标题之前，插入：

```markdown

> **电商系统设计系列**
> - [（一）全景概览与领域划分](/system-design/20-ecommerce-overview/)
> - [（二）商品上架系统](/system-design/21-ecommerce-listing/)
> - [（三）库存系统](/system-design/22-ecommerce-inventory/)
> - [（四）计价引擎](/system-design/23-ecommerce-pricing-engine/)
> - **（五）计价系统 DDD 实践**（本文）
> - [（六）B 端运营系统](/system-design/25-ecommerce-b-side-ops/)

本文是电商系统设计系列的第五篇，是[（四）计价引擎](/system-design/23-ecommerce-pricing-engine/)的姊妹篇，从 DDD 视角重新审视计价系统的建模。

```

- [ ] **Step 3: 在文章末尾添加系列引用**

```markdown

---

> **系列导航**
> 计价引擎的工程实现细节（多级缓存、降级策略等），详见[（四）计价引擎](/system-design/23-ecommerce-pricing-engine/)。
```

- [ ] **Step 4: Commit**

```bash
git add source/_posts/system-design/24-ecommerce-pricing-ddd.md
git commit -m "docs: update article 24 - pricing DDD title, navigation and cross-references"
```

---

### Task 8: 更新文章 25（原 23）— B 端运营系统

**Files:**
- Modify: `source/_posts/system-design/25-ecommerce-b-side-ops.md`

- [ ] **Step 1: 替换 Front Matter**

将原 Front Matter:

```yaml
---
title: 多品类统一商品运营管理系统设计
date: 2026-02-26
categories:
- 系统设计
tags:
- B端运营系统
- 多品类商品管理
- 统一上架系统
- 策略模式
- 电商
- 系统设计
toc: true
---
```

替换为:

```yaml
---
title: 电商系统设计（六）：B 端运营系统
date: 2026-02-26
categories:
- 系统设计
tags:
- e-commerce
- system-design
- b-side
- operations
- observability
toc: true
---
```

- [ ] **Step 2: 在 `<!-- toc -->` 之后插入系列导航块和开头引用**

在 `<!-- toc -->` 标签之后、正文第一个 `##` 标题之前，插入：

```markdown

> **电商系统设计系列**
> - [（一）全景概览与领域划分](/system-design/20-ecommerce-overview/)
> - [（二）商品上架系统](/system-design/21-ecommerce-listing/)
> - [（三）库存系统](/system-design/22-ecommerce-inventory/)
> - [（四）计价引擎](/system-design/23-ecommerce-pricing-engine/)
> - [（五）计价系统 DDD 实践](/system-design/24-ecommerce-pricing-ddd/)
> - **（六）B 端运营系统**（本文）

本文是电商系统设计系列的第六篇，聚焦 B 端运营系统的设计。

```

- [ ] **Step 3: 检查与文章 20 重复的稳定性/监控内容**

搜索文章 25 中与文章 20（全景概览）重复的大段内容，特别关注：
- 系统稳定性建设章节
- 可观测性设计章节
- 异常应急策略章节

如发现大段重复（>50 行相同内容），精简为摘要并添加引用：

```markdown
> 系统稳定性建设的完整设计（监控指标、告警规则、降级/限流/熔断策略），详见[（一）全景概览与领域划分](/system-design/20-ecommerce-overview/)中的系统稳定性章节。本文仅补充 B 端运营视角的特有监控需求。
```

- [ ] **Step 4: 在文章末尾添加系列引用**

```markdown

---

> **系列导航**
> 本系列全部文章索引，详见[（一）全景概览与领域划分](/system-design/20-ecommerce-overview/)。
```

- [ ] **Step 5: Commit**

```bash
git add source/_posts/system-design/25-ecommerce-b-side-ops.md
git commit -m "docs: update article 25 - B-side ops title, navigation, de-dup and cross-references"
```

---

### Task 9: 更新文章 20（原 13）— 全景概览与领域划分

这是改动最大的一篇。分为 3 个子步骤：Front Matter 更新、导航块添加、内容整合。

**Files:**
- Modify: `source/_posts/system-design/20-ecommerce-overview.md`
- Read: `docs/superpowers/backup/22-unified-product-inventory-pricing-system.md`（提取独特内容）

- [ ] **Step 1: 替换 Front Matter**

将原 Front Matter:

```yaml
---
title: 互联网业务系统 - 电商系统设计
date: 2025-05-01
categories:
- 系统设计
---
```

替换为:

```yaml
---
title: 电商系统设计（一）：全景概览与领域划分
date: 2025-05-01
categories:
- 系统设计
tags:
- e-commerce
- system-design
- ddd
- order
- architecture
toc: true
---
```

- [ ] **Step 2: 在正文开头（`## 电商系统整体架构设计` 之前）插入系列导航块**

```markdown
<!-- toc -->

> **电商系统设计系列**
> - **（一）全景概览与领域划分**（本文）
> - [（二）商品上架系统](/system-design/21-ecommerce-listing/)
> - [（三）库存系统](/system-design/22-ecommerce-inventory/)
> - [（四）计价引擎](/system-design/23-ecommerce-pricing-engine/)
> - [（五）计价系统 DDD 实践](/system-design/24-ecommerce-pricing-ddd/)
> - [（六）B 端运营系统](/system-design/25-ecommerce-b-side-ops/)

```

- [ ] **Step 3: 从备份的文章 22 中提取「背景与挑战」章节**

从 `docs/superpowers/backup/22-unified-product-inventory-pricing-system.md` 中读取以下章节：
- `## 一、背景与挑战`（包含 1.1 业务背景、1.2 多品类差异与挑战、1.3 核心痛点、1.4 设计目标）

将其插入到文章 20 的系列导航块之后、`## 电商系统整体架构设计` 之前，作为新的第一章节。

- [ ] **Step 4: 从备份的文章 22 中提取「整体架构 - 三大系统总览」**

从备份文章 22 中读取以下章节：
- `### 2.1 三大系统总览`
- `### 2.2 分层服务架构`
- `### 2.4 核心业务流`（含 2.4.1 流程职责划分）

将其插入到文章 20 的 `## 电商系统整体架构设计` 章节末尾（`## 商品管理 Product Center` 之前），作为新的子章节。

- [ ] **Step 5: 从备份的文章 22 中提取「C 端用户旅程」**

从备份文章 22 中读取以下章节：
- `## 五、用户交易 C 端流程（User Journey）`（包含 5.1 完整用户旅程、5.2 关键节点详细流程、5.3 用户体验优化）

将其作为文章 20 的新章节，插入到 `## 商品管理 Product Center` 之前，命名为 `## C 端用户旅程`。

- [ ] **Step 6: 从备份的文章 22 中提取「新品类接入指南」**

从备份文章 22 中读取以下章节：
- `## 十一、新品类接入指南`（包含 11.1 接入检查清单、11.2 四步接入示例）

将其作为文章 20 的新章节，插入到 `## 其它常见问题` 之前。

- [ ] **Step 7: 在「商品管理 Product Center」章节添加引用**

在 `## 商品管理 Product Center` 章节开头添加引用提示：

```markdown
> 商品上架系统的完整设计（状态机、审核策略、Saga 事务），详见[（二）商品上架系统](/system-design/21-ecommerce-listing/)。
> 库存系统的完整设计（二维分类模型、策略模式、Redis/MySQL 双写），详见[（三）库存系统](/system-design/22-ecommerce-inventory/)。
```

- [ ] **Step 8: 在「系统稳定性建设」章节添加引用**

在 `## 系统稳定性建设` 章节开头添加引用提示：

```markdown
> B 端运营系统的稳定性设计和监控体系，详见[（六）B 端运营系统](/system-design/25-ecommerce-b-side-ops/)。
```

- [ ] **Step 9: 在文章末尾添加系列引用**

在 `## 参考:` 之前追加：

```markdown

---

> **系列导航**
> - [（二）商品上架系统](/system-design/21-ecommerce-listing/) — 上架流程、状态机、审核策略、Saga 事务
> - [（三）库存系统](/system-design/22-ecommerce-inventory/) — 多品类库存模型、策略模式、Redis/MySQL 双写
> - [（四）计价引擎](/system-design/23-ecommerce-pricing-engine/) — 四层计价架构、多级缓存、降级策略
> - [（五）计价系统 DDD 实践](/system-design/24-ecommerce-pricing-ddd/) — 战略/战术设计、六边形架构、价格快照
> - [（六）B 端运营系统](/system-design/25-ecommerce-b-side-ops/) — 运营管理、监控告警、系统稳定性
```

- [ ] **Step 10: 构建验证**

```bash
npx hexo clean && npx hexo generate
```

Expected: 无报错

- [ ] **Step 11: Commit**

```bash
git add source/_posts/system-design/20-ecommerce-overview.md
git commit -m "docs: update article 20 - overview with series nav, absorb unique content from deleted article 22"
```

---

### Task 10: 全量构建验证

- [ ] **Step 1: 清理并完整构建**

```bash
npx hexo clean && npx hexo generate
```

Expected: 无报错，所有文章正常生成

- [ ] **Step 2: 本地预览验证**

```bash
npx hexo server
```

在浏览器中验证：
1. 访问 `/system-design/20-ecommerce-overview/` — 能正常打开，导航链接可点击
2. 访问 `/system-design/21-ecommerce-listing/` — 能正常打开
3. 访问 `/system-design/22-ecommerce-inventory/` — 能正常打开
4. 访问 `/system-design/23-ecommerce-pricing-engine/` — 能正常打开
5. 访问 `/system-design/24-ecommerce-pricing-ddd/` — 能正常打开
6. 访问 `/system-design/25-ecommerce-b-side-ops/` — 能正常打开
7. 确认旧 URL `/system-design/22-unified-product-inventory-pricing-system/` 已不存在

- [ ] **Step 3: 停止本地服务**

```bash
# Ctrl+C 停止 hexo server
```

- [ ] **Step 4: 最终 Commit（如有遗漏修复）**

```bash
git status
# 如有未提交的修复，执行：
# git add -A && git commit -m "fix: final adjustments after build verification"
```
