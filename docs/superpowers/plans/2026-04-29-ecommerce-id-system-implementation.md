# Ecommerce ID System Appendix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a complete appendix for production-grade ID design in ecommerce systems, covering scenarios, option tradeoffs, recommended architecture, governance, and example-code migration guidance.

**Architecture:** This is a documentation-only implementation. The appendix will be added as `ecommerce-book/src/appendix/id-system.md`, linked from `ecommerce-book/src/SUMMARY.md`, and lightly connected to references. Existing example code is used as explanatory context but is not modified.

**Tech Stack:** Markdown, mdBook-style `SUMMARY.md`, existing `ecommerce-book` source tree, npm build commands from repo instructions.

---

### Task 1: Reconfirm Book Context And Source Anchors

**Files:**
- Read: `docs/superpowers/specs/2026-04-29-ecommerce-id-system-design.md`
- Read: `ecommerce-book/src/SUMMARY.md`
- Read: `ecommerce-book/src/appendix/product-supply-ops.md`
- Read: `ecommerce-book/src/appendix/supplier-sync.md`
- Read: `ecommerce-book/src/part2/supply/chapter7.md`
- Read: `ecommerce-book/src/part2/overview/chapter5.md`
- Read: `ecommerce-book/example-codes/product-service/internal/application/service/supply_ops_service.go`
- Read: `ecommerce-book/example-codes/product-service/internal/infrastructure/persistence/supply_ops_repository.go`
- Read: `ecommerce-book/example-codes/order-service/internal/infrastructure/mysql/mysql_db.go`

- [ ] **Step 1: Inspect current book navigation**

Run:

```bash
sed -n '1,120p' ecommerce-book/src/SUMMARY.md
```

Expected: the appendix section currently ends at `附录G 商品供给与运营治理平台`.

- [ ] **Step 2: Inspect appendix writing style**

Run:

```bash
sed -n '1,120p' ecommerce-book/src/appendix/product-supply-ops.md
sed -n '1,120p' ecommerce-book/src/appendix/supplier-sync.md
```

Expected: appendices use numbered sections, Chinese prose, tables, and fenced code blocks.

- [ ] **Step 3: Capture current ID pain points from example code**

Run:

```bash
rg -n 'NextID|NextItemID|NextOrderID|sku_id|spu_id|checkout_id|order_id|idempotency_key' ecommerce-book/example-codes ecommerce-book/src
```

Expected: output includes `s.repo.NextID(ctx, "draft")`, `NextItemID`, `NextOrderID`, and existing book references to `idempotency_key`.

### Task 2: Create Appendix H With Problem Statement, Classification, And Scenario Matrix

**Files:**
- Create: `ecommerce-book/src/appendix/id-system.md`

- [ ] **Step 1: Create the appendix skeleton**

Add `ecommerce-book/src/appendix/id-system.md` with these top-level sections:

```markdown
# 附录H 全局 ID 体系设计

## 1. 为什么电商系统需要全局 ID 体系

## 2. 电商 ID 分类

## 3. 全场景 ID 清单

## 4. 常见发号方案对比

## 5. 推荐混合架构

## 6. ID 服务架构

## 7. 关键业务 ID 设计

## 8. 容灾、风险与治理

## 9. 数据库与接口设计

## 10. 示例代码改造建议

## 11. 面试和架构评审要点

## 12. 小结
```

- [ ] **Step 2: Fill section 1 with the motivating example**

Write section 1 around this concrete contrast:

```go
s.repo.NextID(ctx, "draft")
```

and explain that repository-local prefix generation is acceptable for demos but unsafe for production because namespace governance, multi-instance uniqueness, error handling, observability, and cross-service consistency are missing.

- [ ] **Step 3: Fill section 2 with the ID taxonomy table**

Add a table with these rows:

```markdown
| 类型 | 典型字段 | 设计重点 | 不适合的做法 |
|------|----------|----------|--------------|
| 实体 ID | `item_id`、`spu_id`、`sku_id` | 长期稳定、索引友好、跨系统引用 | 每个服务各自自增 |
| 业务单号 | `order_no`、`payment_no`、`refund_no` | 对外展示、客服查询、对账、不可枚举 | 直接暴露连续自增 |
| 流程单据 ID | `draft_id`、`staging_id`、`qc_review_id` | 流程追踪、审计、低耦合 | 与正式商品 ID 混用 |
| 事件 ID | `event_id`、`outbox_event_id` | 幂等消费、重放、排障 | 用时间戳字符串拼接 |
| 幂等键 | `idempotency_key`、`request_id` | 表达同一次业务请求 | 当作普通随机 ID |
| 链路 ID | `trace_id`、`operation_id` | 跨服务追踪和审计 | 每层重新生成 |
```

- [ ] **Step 4: Fill section 3 with the ecommerce scenario matrix**

Add a table covering at least these domains:

```text
商品、供给、库存、购物车、结算、订单、支付、售后、营销、搜索、履约、财务、事件、链路追踪
```

For each domain, include key IDs, recommended type, and recommended generator. Ensure rows include:

```text
item_id/spu_id/sku_id -> BIGINT -> Segment or Snowflake
draft_id/staging_id/qc_review_id -> string -> ULID/UUIDv7 + prefix
checkout_id -> string -> ULID/UUIDv7 + idempotency
order_id/order_no -> BIGINT + string -> Snowflake-derived business number
event_id/outbox_event_id -> string -> ULID/UUIDv7 or deterministic event ID
idempotency_key -> string -> business unique key
```

### Task 3: Add Scheme Comparison And Recommended Hybrid Architecture

**Files:**
- Modify: `ecommerce-book/src/appendix/id-system.md`

- [ ] **Step 1: Fill section 4 with scheme comparison**

Add subsections for:

```markdown
### 4.1 DB 自增
### 4.2 DB Sequence 表
### 4.3 Redis INCR
### 4.4 Snowflake
### 4.5 Segment 号段
### 4.6 UUIDv7、ULID 与 KSUID
```

Each subsection must include: core idea, advantages, drawbacks, and suitable ecommerce scenarios.

- [ ] **Step 2: Add a comparison table**

Add this table shape after the subsections:

```markdown
| 方案 | 是否中心化 | 是否趋势递增 | 主要优点 | 主要风险 | 推荐场景 |
|------|------------|--------------|----------|----------|----------|
```

Rows must cover DB 自增、DB Sequence、Redis INCR、Snowflake、Segment、UUIDv7/ULID.

- [ ] **Step 3: Fill section 5 with the recommended hybrid architecture**

Use this architecture block:

```text
业务服务
  -> ID SDK
  -> ID Registry
  -> Generator Router
      -> Segment Generator
      -> Snowflake Generator
      -> ULID / UUIDv7 Generator
      -> Business Number Formatter
  -> Observability / Audit / Admin
```

State the recommendation explicitly:

```text
商品和库存主数据：Segment 或 Snowflake 的 BIGINT
交易单号：Snowflake 派生业务单号
供给流程、事件和链路：ULID/UUIDv7 + 受控 prefix
幂等：业务语义唯一约束，不等同于普通 ID
```

### Task 4: Add ID Service Design, Critical Business Cases, And Risk Governance

**Files:**
- Modify: `ecommerce-book/src/appendix/id-system.md`

- [ ] **Step 1: Fill section 6 with ID service components**

Describe these components:

```text
ID Registry
ID SDK
Generator Router
Segment Generator
Snowflake Generator
ULID/UUIDv7 Generator
Business Number Formatter
Observability / Audit / Admin
```

Include the boundary rule: business services depend on the SDK; repositories do not define ID rules.

- [ ] **Step 2: Fill section 7 with critical business case designs**

Add subsections:

```markdown
### 7.1 `sku_id`、`spu_id` 与 `item_id`
### 7.2 `order_id` 与 `order_no`
### 7.3 `checkout_id` 与 `idempotency_key`
### 7.4 `payment_id`、渠道单号与对账
### 7.5 `draft_id`、`staging_id` 与供给审核单
### 7.6 `event_id` 与 Outbox 去重
```

Each subsection must state the recommended generator and why it matches the business semantics.

- [ ] **Step 3: Fill section 8 with risk governance**

Cover these exact risks:

```text
时钟回拨
号段浪费
重复发号
ID 枚举
跨地域冲突
字段类型失控
把幂等键当 ID
```

For each risk, include at least one mitigation.

### Task 5: Add Database/API Design And Example-Code Migration Guidance

**Files:**
- Modify: `ecommerce-book/src/appendix/id-system.md`

- [ ] **Step 1: Fill section 9 with SQL tables**

Add SQL examples for:

```sql
CREATE TABLE id_namespace (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    namespace VARCHAR(64) NOT NULL COMMENT '业务命名空间，例如 product.sku、trade.order',
    biz_domain VARCHAR(64) NOT NULL COMMENT '业务域，例如 product、trade、supply',
    id_type VARCHAR(32) NOT NULL COMMENT 'INT64/STRING/BUSINESS_NO/IDEMPOTENCY_KEY',
    generator_type VARCHAR(32) NOT NULL COMMENT 'SEGMENT/SNOWFLAKE/ULID/UUIDV7/BUSINESS',
    prefix VARCHAR(32) DEFAULT NULL COMMENT '字符串 ID 或业务单号前缀',
    expose_scope VARCHAR(32) NOT NULL COMMENT 'INTERNAL/EXTERNAL/MIXED',
    step INT NOT NULL DEFAULT 1000 COMMENT 'Segment 号段步长',
    max_capacity BIGINT DEFAULT NULL COMMENT '容量规划上限',
    owner_team VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL COMMENT 'ENABLED/DISABLED/DEPRECATED',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_namespace (namespace),
    KEY idx_domain_status (biz_domain, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ID 命名空间注册表';

CREATE TABLE id_segment (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    namespace VARCHAR(64) NOT NULL,
    max_id BIGINT NOT NULL COMMENT '当前已经分配到的最大 ID',
    step INT NOT NULL COMMENT '每次申请的号段大小',
    version BIGINT NOT NULL COMMENT '乐观锁版本',
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_namespace (namespace)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Segment 号段表';

CREATE TABLE id_worker (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    worker_id INT NOT NULL,
    region_code VARCHAR(32) NOT NULL,
    datacenter_code VARCHAR(32) NOT NULL,
    instance_id VARCHAR(128) NOT NULL,
    lease_token VARCHAR(64) NOT NULL,
    lease_until DATETIME NOT NULL,
    heartbeat_at DATETIME NOT NULL,
    status VARCHAR(32) NOT NULL COMMENT 'ACTIVE/EXPIRED/DISABLED',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_worker_region_dc (worker_id, region_code, datacenter_code),
    UNIQUE KEY uk_instance (instance_id),
    KEY idx_status_lease (status, lease_until)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Snowflake worker 租约表';

CREATE TABLE id_issue_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    request_id VARCHAR(64) NOT NULL,
    namespace VARCHAR(64) NOT NULL,
    caller VARCHAR(128) NOT NULL,
    issue_type VARCHAR(32) NOT NULL COMMENT 'SUCCESS/FAILED/ROLLBACK/SEGMENT_ALLOCATED',
    issued_value VARCHAR(128) DEFAULT NULL,
    error_message VARCHAR(512) DEFAULT NULL,
    created_at DATETIME NOT NULL,
    UNIQUE KEY uk_request_id (request_id),
    KEY idx_namespace_time (namespace, created_at),
    KEY idx_issue_type_time (issue_type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='关键 ID 发号审计与异常记录';
```

The appendix can shorten comments if prose around the SQL already explains the intent, but it must preserve primary keys, unique keys, status fields, and timestamps.

- [ ] **Step 2: Add Go SDK interface**

Add this interface as the minimum shape:

```go
type Namespace string

const (
    NamespaceProductItem  Namespace = "product.item"
    NamespaceProductSPU   Namespace = "product.spu"
    NamespaceProductSKU   Namespace = "product.sku"
    NamespaceSupplyDraft  Namespace = "supply.draft"
    NamespaceSupplyStage  Namespace = "supply.staging"
    NamespaceTradeOrder   Namespace = "trade.order"
    NamespaceTradePayment Namespace = "trade.payment"
)

type Generator interface {
    NextInt64(ctx context.Context, ns Namespace) (int64, error)
    NextString(ctx context.Context, ns Namespace) (string, error)
    NextBatchInt64(ctx context.Context, ns Namespace, size int) ([]int64, error)
}
```

- [ ] **Step 3: Fill section 10 with before/after examples**

Show the current simplified code:

```go
DraftID: s.repo.NextID(ctx, "draft"),
```

and the target shape:

```go
draftID, err := s.idgen.NextString(ctx, id.NamespaceSupplyDraft)
if err != nil {
    return nil, err
}
```

State clearly that this appendix only documents the migration direction and does not require changing `example-codes` in this task.

- [ ] **Step 4: Fill sections 11 and 12**

Section 11 must include interview/review questions covering uniqueness, ordering, exposure, clock rollback, disaster recovery, idempotency, and cross-region behavior.

Section 12 must summarize:

```text
统一 ID 体系的重点不是某个算法，而是按业务语义治理 namespace、生成策略、暴露形式和失败处理。
```

### Task 6: Update Navigation And References

**Files:**
- Modify: `ecommerce-book/src/SUMMARY.md`
- Modify: `ecommerce-book/src/appendix/references.md`

- [ ] **Step 1: Add appendix H to SUMMARY**

Add one line after appendix G:

```markdown
- [附录H 全局 ID 体系设计](appendix/id-system.md)
```

- [ ] **Step 2: Add references**

Append references to `ecommerce-book/src/appendix/references.md` under the existing title:

```markdown
## 分布式 ID 与唯一性

- Twitter Snowflake：分布式趋势递增 ID 的经典方案。
- 美团 Leaf：Segment 号段与 Snowflake 模式的工程化参考。
- UUID Version 7：适合按时间排序的 UUID 标准。
- ULID：按时间排序、可读性较好的 128 位标识。
```

### Task 7: Validate Formatting And Build

**Files:**
- Verify: `ecommerce-book/src/appendix/id-system.md`
- Verify: `ecommerce-book/src/SUMMARY.md`
- Verify: `ecommerce-book/src/appendix/references.md`

- [ ] **Step 1: Search for unresolved markers**

Run:

```bash
rg -n 'T[B]D|T[O]DO|待[定]|待补[充]|F[I]XME' ecommerce-book/src/appendix/id-system.md ecommerce-book/src/SUMMARY.md ecommerce-book/src/appendix/references.md
```

Expected: no matches.

- [ ] **Step 2: Check links to the new appendix**

Run:

```bash
rg -n '附录H|id-system.md' ecommerce-book/src
```

Expected: at least `SUMMARY.md` and `appendix/id-system.md` appear.

- [ ] **Step 3: Build the book/site**

Run:

```bash
npm run clean && npm run build
```

Expected: command exits with status 0.

- [ ] **Step 4: Review git diff**

Run:

```bash
git diff -- ecommerce-book/src/appendix/id-system.md ecommerce-book/src/SUMMARY.md ecommerce-book/src/appendix/references.md
```

Expected: diff contains only appendix, navigation, and reference updates for the ID system topic.

- [ ] **Step 5: Commit implementation changes**

Stage only the implementation files that changed:

```bash
git add ecommerce-book/src/appendix/id-system.md ecommerce-book/src/SUMMARY.md ecommerce-book/src/appendix/references.md
git commit -m "Add ecommerce global ID system appendix"
```
