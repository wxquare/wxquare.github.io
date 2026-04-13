# 购物车与结算域文章 Implementation Plan

> **For agentic workers:** 按下方 checkbox 顺序执行；每完成一章可在 Git 中单次或分批提交。本仓库为 Hexo 静态站，验收以 `npm run clean && npm run build` 通过为准。

**Goal:** 在 `source/_posts/system-design/32-ecommerce-cart-checkout.md` 发布符合 `docs/superpowers/specs/2026-04-13-ecommerce-cart-checkout-design.md` 的系列长文（面试向 + 工程向），并与 `20`～`31` 篇建立交叉引用。

**Architecture:** 单文件 Markdown；主线为「购物车域（暂存与合并）+ 结算域（Saga 编排与预占）」；与 `26`/`23`/`22`/`28`/`27`/`29` 按接口契约分工，不重复实现细节。

**Tech Stack:** Hexo 7、Markdown、Mermaid、中文标点与中英文空格规范。

---

## 文件映射

| 文件 | 职责 |
|------|------|
| `source/_posts/system-design/32-ecommerce-cart-checkout.md` | 新建正文（Front Matter + 12 章结构 + 系列导航） |
| `docs/superpowers/specs/2026-04-13-ecommerce-cart-checkout-design.md` | 已存在，仅当正文偏离 scope 时回溯修订 |

---

### Task 1：Front Matter 与系列导航

**Files:**

- Create: `source/_posts/system-design/32-ecommerce-cart-checkout.md`（首屏含 Front Matter + 与 `31` 衔接的系列说明 + 指向 `20` 总索引）

- [ ] **Step 1:** `title` 使用「（十三）」与 `31` 的「（十二）」连续；`date: '2026-04-13'`（字符串）；`categories: [电商系统设计]`；`tags` 与 spec 一致（小写连字符）。

- [ ] **Step 2:** 文首引用块列出 `20`、`26`、`23`、`22`、`28`、`27`、`29`、`31` 的站内路径（与既有文章 `/system-design/slug/` 格式一致）。

**验收:** Front Matter 无 YAML 语法错误；系列篇次不自相矛盾。

---

### Task 2：第 1～2 章（引言、范围、分工表）

**Files:**

- Modify: `source/_posts/system-design/32-ecommerce-cart-checkout.md`

- [ ] **Step 1:** 写引言：购物车与结算在转化漏斗的位置；读多写少 vs 强校验的矛盾。

- [ ] **Step 2:** 显式「范围 / 非目标」列表（与 spec 一致：无支付流程主文、无订单状态机全展开）。

- [ ] **Step 3:** Markdown 表：与 `26`/`23`/`22`/`28`/`27`/`29` 分工（去重规则一段话 + 表）。

**验收:** 读者能一句话说清「购物车不做什么、结算页不做什么」。

---

### Task 3：第 3 章（核心场景与挑战）

**Files:**

- Modify: `source/_posts/system-design/32-ecommerce-cart-checkout.md`

- [ ] **Step 1:** 购物车场景表：未登录加购、登录合并、商品失效、批量操作、跨端同步。

- [ ] **Step 2:** 结算页场景表：价格试算、库存预占、营销校验、拆单编排、地址 & 运费、幂等与重试。

- [ ] **Step 3:** 挑战对比表：技术挑战 vs 方案（含未登录加购、预占库存、Saga 补偿、幂等）。

**验收:** 与第 4～8 章对应；面试可按此表逐项展开。

---

### Task 4：第 4 章（购物车设计）

**Files:**

- Modify: `source/_posts/system-design/32-ecommerce-cart-checkout.md`

- [ ] **Step 1:** `shopping_cart` 表 SQL（`user_id` / `cart_token` / `sku_id` / `quantity` / `selected`；唯一索引与索引）。

- [ ] **Step 2:** Redis HASH 结构说明（`key=cart:{user_id}`, `field=sku_id`, `value=quantity`）+ DB 双写策略。

- [ ] **Step 3:** 匿名与登录态合并策略（相同 SKU 数量相加、不同 SKU 追加）。

- [ ] **Step 4:** 批量操作幂等（乐观锁版本号）+ Go 伪代码示例（≤ 30 行）。

**验收:** 购物车表与 Redis 结构可直接用于实现。

---

### Task 5：第 5 章（结算页设计 Checkout Orchestrator）

**Files:**

- Modify: `source/_posts/system-design/32-ecommerce-cart-checkout.md`

- [ ] **Step 1:** Mermaid `sequenceDiagram`：结算页 Saga 编排（用户 → 结算服务 → 计价/库存/营销 → 订单系统；含补偿路径）。

- [ ] **Step 2:** 幂等与去重：`idempotency_key` + Redis `SET NX` + 订单表唯一索引。

- [ ] **Step 3:** 补偿路径说明：预占超时自动释放、订单创建失败显式释放（引用 `22` 与 `26`）。

- [ ] **Step 4:** 结算会话表（可选）：有状态 vs 无状态权衡（≤ 1 段）。

**验收:** Saga 序列图可读；幂等段落可单独背诵。

---

### Task 6：第 6 章（拆单与地址运费）

**Files:**

- Modify: `source/_posts/system-design/32-ecommerce-cart-checkout.md`

- [ ] **Step 1:** 拆单策略：跨店铺、跨仓、自营 + POP；结算页预览拆单结果，真正拆单在订单系统（引用 `26`）。

- [ ] **Step 2:** 地址服务集成：多地址切换、默认地址、收货人信息校验。

- [ ] **Step 3:** 运费规则引擎：运费实时计算（可缓存短时，秒级 TTL）。

**验收:** 明确结算页与订单系统在拆单上的边界。

---

### Task 7：第 7 章（与其他系统的集成与契约 - 重点章节）

**Files:**

- Modify: `source/_posts/system-design/32-ecommerce-cart-checkout.md`

- [ ] **Step 1:** 购物车边界表（与商品/计价/库存的调用场景与"不做什么"）。

- [ ] **Step 2:** 结算页边界表（与计价/库存/营销/地址/订单的接口契约、返回字段、失败处理、"不做什么"）。

- [ ] **Step 3:** 订单系统接收契约：JSON 示例（`price_snapshot_id`、`reserve_ids`、`idempotency_key`）+ 订单系统职责列表。

- [ ] **Step 4:** Mermaid `sequenceDiagram`：购物车查商品流程。

- [ ] **Step 5:** 边界陷阱反例表：5 个常见反模式与正确做法对比。

- [ ] **Step 6:** 事件消费语义（可选）：订单创建成功后清理购物车（异步、非强依赖）。

**验收:** 边界表可作为技术评审的契约文档。

---

### Task 8：第 8～10 章（一致性、可观测、工程清单）

**Files:**

- Modify: `source/_posts/system-design/32-ecommerce-cart-checkout.md`

- [ ] **Step 1:** 购物车弱一致、结算页强一致；预占超时释放机制（15 分钟过期）。

- [ ] **Step 2:** 指标表：加购率、进入结算率、提交订单成功率；漏斗分析。

- [ ] **Step 3:** 工程清单：购物车同步策略、结算页超时配置、预占释放监控。

**验收:** 每章有「可落地检查项」列表。

---

### Task 9：第 11～12 章（面试锦囊、总结与索引）

**Files:**

- Modify: `source/_posts/system-design/32-ecommerce-cart-checkout.md`

- [ ] **Step 1:** 15～20 条问答式要点，覆盖未登录加购、合并策略、预占库存、幂等、拆单时机、Saga 补偿、漏斗分析。

- [ ] **Step 2:** 总结 + 系列链接 + 下一篇可选方向（履约/物流、商家结算与对账）。

**验收:** 条数 ≥ 15。

---

### Task 10：构建与提交

**Files:**

- Shell: 仓库根目录

- [ ] **Step 1:** 运行 `npm run clean && npm run build`

**期望:** 进程退出码 0；无 Hexo 报错。

- [ ] **Step 2:** `git add` 新文章与 plan；`git commit -m "docs: add ecommerce cart & checkout post (13) and plan"`

**验收:** 工作区干净或仅剩无关本地改动。

---

## Plan 自检（对照 spec）

| Spec 章节 | Plan 覆盖 |
|-----------|-----------|
| 购物车域 + 结算域分域叙事 | Task 4 + 5 |
| 与 `26`/`23`/`22`/`28` 去重 | Task 2、7 |
| 边界与契约（重点章节） | Task 7 |
| 面试锦囊 | Task 9 |
| 写作约束 / build | Task 1、10 |

无 TBD 占位；Kafka Topic 名（如有）刻意泛化或对齐已有命名；后续系列若统一命名，可在修订提交中替换。
