# 四域架构方法论文章 — Implementation Plan

> **For agentic workers:** 使用 `superpowers:subagent-driven-development` 或 `superpowers:executing-plans` 按任务逐项执行；步骤使用 `- [ ]` 勾选跟踪。

**Goal:** 在仓库中新增 Hexo 长文 `45-acc-four-domains-methodology.md`，落实规格 `docs/superpowers/specs/2026-04-11-architecture-four-domains-methodology-design.md`（流程驱动 + 四域工件 + 40–70 条清单 + 电商示例 + ACC 对照表）。

**Architecture:** 单文件、上半流程六步（输入 / 活动 / 输出 / 电商示例 / 反例可选）、下半四域 × 四维清单、文末复盘与扩展阅读；站内链接采用根路径 `/:year/:month/:day/:slug/`，`:slug` 与对应 `.md` 文件名（不含扩展名）一致（与 `_config.yml` 中 `permalink: :year/:month/:day/:title/` 及 Hexo 默认 `title` 变量行为一致；若本地生成路径不同，以 `public/` 下实际目录为准）。

**Tech Stack:** Hexo 7、Markdown、Front Matter（`toc: true`）。

**站内链接（重要）：** `系统设计基础` 分类下的文章在 `public/` 中实际路径为 `/:year/:month/:day/system-design/:slug/`（中间含 **`system-design`** 段）。Markdown 中根相对链接须写为 `/2026/04/01/system-design/41-acc-clean-arch-ddd-cqrs/` 等形式，而非省略分类段。

**规格覆盖自检:** 引言四域定义、总览表、六步流程、电商绑定、反例、ACC 对照表、四域清单（40–70 条）、结尾复盘、扩展阅读、非功能约束（不写 themes/ 等）均有对应任务。

---

## 站内链接速查（扩展阅读与 ACC，写入正文时使用）

根路径均为 `https://你的站点根` 在本地预览时为 `http://localhost:4000`；**仓库内建议使用相对根路径** `/YYYY/MM/DD/slug/`：

| 文章 | `date`（Front Matter） | 建议路径 |
|------|------------------------|----------|
| 41-acc-clean-arch-ddd-cqrs | 2026-04-01 | `/2026/04/01/41-acc-clean-arch-ddd-cqrs/` |
| 42-acc-clean-code | 2026-04-02 | `/2026/04/02/42-acc-clean-code/` |
| 43-acc-ddd-notes | 2026-04-03 | `/2026/04/03/43-acc-ddd-notes/` |
| 44-acc-code-review | 2026-04-04 | `/2026/04/04/44-acc-code-review/` |
| 20-ecommerce-overview | 2025-05-01 | `/2025/05/01/20-ecommerce-overview/` |
| 21-ecommerce-listing | 2025-08-21 | `/2025/08/21/21-ecommerce-listing/` |
| 22-ecommerce-inventory | 2025-05-29 | `/2025/05/29/22-ecommerce-inventory/` |
| 23-ecommerce-pricing-engine | 2025-06-26 | `/2025/06/26/23-ecommerce-pricing-engine/` |
| 24-ecommerce-pricing-ddd | 2025-07-10 | `/2025/07/10/24-ecommerce-pricing-ddd/` |
| 26-ecommerce-order-system | 2025-07-24 | `/2025/07/24/26-ecommerce-order-system/` |
| 27-ecommerce-product-center | 2025-05-15 | `/2025/05/15/27-ecommerce-product-center/` |
| 28-ecommerce-marketing-system | 2025-06-12 | `/2025/06/12/28-ecommerce-marketing-system/` |
| 29-ecommerce-payment-system | 2025-08-07 | `/2025/08/07/29-ecommerce-payment-system/` |
| 30-ecommerce-product-lifecycle-management | 2026-04-10 | `/2026/04/10/30-ecommerce-product-lifecycle-management/` |

**验证命令（链接路径以构建结果为准）:**

```bash
cd /Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io
npm run clean && npm run build
ls public/2026/04/01/ 2>/dev/null || ls public/2026/04/04/
```

**期望:** `public/` 下出现与上表一致的目录层级；若 `title` 与文件名不一致导致 slug 变化，以 `public/` 为准批量替换正文链接。

---

### Task 1: 创建文章文件与 Front Matter

**Files:**

- Create: `source/_posts/system-design/45-acc-four-domains-methodology.md`

- [ ] **Step 1: 写入完整 Front Matter 与 `<!-- toc -->`**

将下列内容作为文件开头（`date` 若晚于实际发布日可改，须为 `YYYY-MM-DD` 字符串）：

```yaml
---
title: 架构与整洁代码（五）：四域架构方法论 — 流程、工件与清单（电商实践向）
date: 2026-04-11
categories:
  - 系统设计基础
tags:
  - architecture-and-clean-code
  - architecture
  - e-commerce
  - system-design
  - four-domains
  - engineering-practice
  - 架构设计
toc: true
---

<!-- toc -->
```

- [ ] **Step 2: Commit**

```bash
git add source/_posts/system-design/45-acc-four-domains-methodology.md
git commit -m "post: scaffold 45 four-domains methodology article front matter"
```

---

### Task 2: 引言与四域定义

**Files:**

- Modify: `source/_posts/system-design/45-acc-four-domains-methodology.md`（在 `<!-- toc -->` 之后追加）

- [ ] **Step 1: 撰写「引言」一节（约 400–800 字）**

必含要点（可用你自己的句式，但信息须齐全）：

1. 单产品 / 技术负责人语境；与「企业 EA 治理」边界一句带过。  
2. 四域工程向定义：业务（能力与场景）、应用（边界与契约）、数据（所有权与一致性）、技术（约束与横切）。  
3. 与 ACC `41`–`44` 分工：本文管「方案成形与对齐」，彼系列管「建模、分层、CQRS、评审」；各用 Markdown 链接指向上表路径。  
4. 电商锚点：约 5 年电商实践作示例来源；默认 B2C / 平台型；跨境 / 多租户一句「超出本文默认假设」。  
5. 本文交付物预告：**六步流程** + **下半篇清单**。

- [ ] **Step 2: Commit**

```bash
git add source/_posts/system-design/45-acc-four-domains-methodology.md
git commit -m "post: add intro and four-domain definitions for 45"
```

---

### Task 3: 流程总览表

**Files:**

- Modify: `source/_posts/system-design/45-acc-four-domains-methodology.md`

- [ ] **Step 1: 新增章节「流程总览」并插入 Markdown 表**

列固定为：**步骤 | 主要问题 | 关键产出（工件）| 主要落在哪一域**。六行对应规格中步骤 1–6 名称与「四域落点」列（与规格第三节表一致）。

- [ ] **Step 2: Commit**

```bash
git commit -am "post: add workflow overview table for 45"
```

---

### Task 4: 流程步骤 1–2（问题与成功标准、业务架构轻量）

**Files:**

- Modify: `source/_posts/system-design/45-acc-four-domains-methodology.md`

- [ ] **Step 1: 为步骤 1、2 各写一节，结构统一为「输入 / 活动 / 输出 / 电商示例 / 常见反例（可选）」**

**步骤 1 电商要点（须出现）:** 峰值下单、支付成功率、库存准确性（零容忍或可接受策略）、价格展示与结算一致性、合规边界（支付牌照、个人信息）——写成叙述 + 可验收标准 bullet，不单列术语。

**步骤 2 电商要点:** 能力「下单、算价、占库、收款、发货、售后」；场景「大促 + 券 + 会员价 + 平台补贴」；明确本步**不写**技术选型长文。

- [ ] **Step 2: Commit**

```bash
git commit -am "post: add workflow steps 1-2 e-commerce examples for 45"
```

---

### Task 5: 流程步骤 3–4（应用架构、数据架构）

**Files:**

- Modify: `source/_posts/system-design/45-acc-four-domains-methodology.md`

- [ ] **Step 1: 步骤 3、4 各一节，结构同 Task 4**

**步骤 3 电商要点:** 上下文示例（订单、库存、商品与价格、营销、支付、物流）；同步 API 与领域事件各至少一例（如「订单已支付 → 出库指令」）；契约形态（REST / gRPC / 事件 schema 版本一句话）。

**步骤 4 电商要点:** 订单状态机与支付回调乱序；库存流水与可售库存；价格快照与订单行；Outbox、幂等、对账——各用**短段落**说明「输出工件」长什么样（非教程级展开）。

- [ ] **Step 2: Commit**

```bash
git commit -am "post: add workflow steps 3-4 for 45"
```

---

### Task 6: 流程步骤 5–6（技术架构、验证与固化）

**Files:**

- Modify: `source/_posts/system-design/45-acc-four-domains-methodology.md`

- [ ] **Step 1: 步骤 5、6 各一节，结构同 Task 4**

**步骤 5 电商要点:** 多 AZ、缓存与击穿、消息堆积、对账批处理、密钥与「卡数据不进业务库」级原则、成本敏感路径。

**步骤 6 电商要点:** 风险与假设列表模板；ADR 示例主题（「库存先扣后付 vs 先付后扣」「拆单与部分退款」）；评审关口与清单预告（指向下半篇）。

- [ ] **Step 2: Commit**

```bash
git commit -am "post: add workflow steps 5-6 for 45"
```

---

### Task 7: ACC 系列对照表

**Files:**

- Modify: `source/_posts/system-design/45-acc-four-domains-methodology.md`

- [ ] **Step 1: 插入二维 Markdown 表一节**

行：至少四行，对应 **业务 / 应用 / 数据 / 技术**（可与「流程横切」合并一行若你愿压缩，但规格要求能对照 ACC）。  
列：`41` | `42` | `43` | `44`。  
每格：**一句话**说明该文如何加深本域实践 + **同一格内 Markdown 链接**（路径用本计划「站内链接速查」表）。

- [ ] **Step 2: Commit**

```bash
git commit -am "post: add ACC cross-reference table for 45"
```

---

### Task 8: 分域审查清单（40–70 条）

**Files:**

- Modify: `source/_posts/system-design/45-acc-four-domains-methodology.md`

- [ ] **Step 1: 新章节「分域审查清单」，四大块：业务 / 应用 / 数据 / 技术**

每块下四个子标题：**完整性 | 一致性 | 可演进 | 可运维（含安全）**，使用 `- [ ]` 清单语法，便于读者打印勾选（Hexo 渲染为列表即可）。

- [ ] **Step 2: 条数与主题约束**

1. 总条数 **40–70**（含上下界）。  
2. 必须覆盖规格第四节「电商向清单主题」：大促与异常流程、第三方超时与补偿、幂等键、快照与审计、SLO / 告警 / 灰度等。  
3. 至少 **3 处** 用一句话指向 `44` 的详细 Code Review 维度（带链接），避免本篇无限膨胀。

- [ ] **Step 3: Commit**

```bash
git commit -am "post: add four-domain review checklist for 45"
```

---

### Task 9: 结尾、扩展阅读、总结

**Files:**

- Modify: `source/_posts/system-design/45-acc-four-domains-methodology.md`

- [ ] **Step 1: 撰写「复盘框架」小节**

固定三个反思点：**本轮最省的决策**、**本轮最险的决策**、**下一轮要提前问业务的一个问题**；各 2–4 句，可虚构「Composite 角色」式匿名示例，避免泄露敏感数据。

- [ ] **Step 2: 「扩展阅读」分组列表**

一组：**电商系列**（链接本计划速查表中 20–30 全部条目，标题用各文 `title` Front Matter 或中文文件名意译，与站内打开标题一致即可）。  
另一组：**ACC 系列**（`41`–`44`）。

- [ ] **Step 3: 「总结」短段（5–10 句）** 回扣成功标准（读者一次迭代可带走流程 + 清单 + 一类权衡）。

- [ ] **Step 4: Commit**

```bash
git commit -am "post: add closing, further reading, summary for 45"
```

---

### Task 10: 语言规范扫描与构建验证

**Files:**

- Modify: `source/_posts/system-design/45-acc-four-domains-methodology.md`（按需微调标点与中英文空格）

- [ ] **Step 1: 全文检查**

中文正文使用中文标点；英文术语首次出现可加中文括注；中英文之间空格（`.cursorrules`）。

- [ ] **Step 2: 构建**

```bash
cd /Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io
npm run clean && npm run build
```

**期望:** 命令退出码 `0`；若某篇 Front Matter YAML 报错导致 `0 files generated`，先修复**该篇**或跳过渲染配置，直至本新文章出现在 `public/` 对应日期目录。

- [ ] **Step 3: Commit（仅当有格式修复）**

```bash
git commit -am "post: polish language and fix build issues for 45"
```

---

### Task 11: 更新规格状态（可选但推荐）

**Files:**

- Modify: `docs/superpowers/specs/2026-04-11-architecture-four-domains-methodology-design.md`

- [ ] **Step 1: 将「设计状态」改为「已实现」或「正文已完成」并加一行指向正文路径**

- [ ] **Step 2: Commit**

```bash
git add docs/superpowers/specs/2026-04-11-architecture-four-domains-methodology-design.md
git commit -m "docs: mark four-domains methodology spec as implemented"
```

---

## Plan self-review

| 检查项 | 结果 |
|--------|------|
| 规格每节有任务 | 已映射引言、表、1–6 步、ACC 表、清单、结尾、扩展阅读、构建 |
| 占位符 | 无 TBD |
| 与仓库约束 | 仅 `source/_posts` 与可选 spec；未触碰 `themes/` |

---

**Plan 已保存至** `docs/superpowers/plans/2026-04-11-architecture-four-domains-methodology.md`。

**执行方式二选一：**

1. **Subagent-Driven（推荐）** — 每个 Task 新开子代理执行，任务间人工快速过目 diff。  
2. **Inline Execution** — 在当前会话按 Task 1→11 连续改稿与提交。

你回复 **「1」** 或 **「2」** 即可；若希望我现在就在本会话里从 Task 1 开始直接写正文，回复 **「2，开始」**。
