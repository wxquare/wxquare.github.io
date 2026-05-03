# 电商全局 ID 体系设计规格

## 1. 目标

在 `ecommerce-book/src/appendix` 中新增一篇工程级附录，暂定为：

```text
附录H 全局 ID 体系设计
文件：ecommerce-book/src/appendix/id-system.md
```

这篇附录要解决的问题不是“如何实现一个 Snowflake”，而是帮助读者理解电商系统中不同 ID 的业务语义、技术约束、选型取舍、治理方式和工程落地。它需要覆盖商品、供给、库存、购物车、结算、订单、支付、售后、营销、搜索、履约、财务、事件和链路追踪等主要场景，并能解释示例代码中 `NextID(ctx, "draft")`、`sku_id`、`spu_id`、`order_id`、`checkout_id` 等写法应该如何演进。

## 2. 背景与问题

当前书稿和示例代码中已经大量出现 ID：

- 商品中心：`item_id`、`spu_id`、`sku_id`、`offer_id`、`rate_plan_id`、`snapshot_id`
- 供给链路：`draft_id`、`staging_id`、`qc_review_id`、`operation_id`、`task_id`、`batch_id`
- 交易链路：`cart_id`、`checkout_id`、`order_id`、`payment_id`、`refund_id`
- 库存履约：`inventory_key`、`reservation_id`、`deduct_id`、`fulfillment_id`
- 集成治理：`event_id`、`outbox_event_id`、`trace_id`、`idempotency_key`

这些 ID 的要求并不相同。商品主数据 ID 要长期稳定、适合索引和跨系统引用；订单号要对外展示、不可轻易枚举、便于客服和对账；草稿和审核单 ID 更强调流程追踪；事件 ID 和幂等键则用于去重、重放和一致性治理。如果所有场景都用同一种生成方式，要么过度设计，要么在生产环境里留下冲突、枚举、时钟回拨、跨机房扩展和数据泄露等风险。

## 3. 范围

本次附录写作包含：

1. 新增 `appendix/id-system.md`。
2. 在 `SUMMARY.md` 附录区加入入口，作为附录 H。
3. 建立电商全场景 ID 分类和选型矩阵。
4. 对比 DB 自增、DB Sequence、Redis INCR、Snowflake、Segment 号段、UUIDv7/ULID 等常见方案。
5. 推荐一套生产可行的混合架构：Segment 号段 + Snowflake + ULID/UUIDv7 + 幂等键治理。
6. 给出 ID 服务的逻辑架构、核心接口、表结构、容灾策略、监控指标和迁移建议。
7. 衔接示例代码，说明 `repo.NextID(ctx, "draft")`、`NextItemID`、`NextOrderID` 应迁移到统一 `idgen` 能力。

本次附录写作不直接改造 `example-codes` 的 Go 代码。代码改造可以作为后续独立任务，避免附录写作和工程重构互相牵连。

## 4. 读者收益

读完附录后，读者应该能够回答：

1. 为什么 `sku_id`、`order_id`、`draft_id`、`event_id` 和 `idempotency_key` 不能混为一谈？
2. 什么场景适合 DB Sequence、Redis INCR、Snowflake、Segment 或 ULID？
3. 电商系统里哪些 ID 应该是 `BIGINT`，哪些应该是字符串业务单号？
4. 订单号为什么不应直接使用连续自增值？
5. `checkout_id` 与 `idempotency_key` 的关系是什么？
6. 多服务、多实例、多机房下如何避免重复发号？
7. 统一 ID 服务应该有哪些表、接口、监控和降级策略？
8. 示例代码里的简化发号逻辑应该如何演进到生产级设计？

## 5. 附录结构

建议正文结构如下：

```text
1. 为什么电商系统需要全局 ID 体系
2. 电商 ID 分类：实体 ID、业务单号、流程单据、事件 ID、幂等键、链路 ID
3. 全场景 ID 清单：商品、供给、库存、购物车、结算、订单、支付、售后、营销、搜索、履约、财务
4. 常见方案对比：DB 自增、DB Sequence、Redis INCR、Snowflake、Segment 号段、UUIDv7/ULID
5. 推荐混合架构：Segment + Snowflake + ULID/UUIDv7 + 幂等键治理
6. ID 服务架构：Registry、Generator、SDK、号段缓存、批量取号、多机房
7. 关键业务设计：sku_id/spu_id、order_id、checkout_id、payment_id、draft_id、event_id
8. 容灾与风险：时钟回拨、号段浪费、重复发号、ID 暴露、枚举攻击、跨地域冲突
9. 数据库与接口设计：namespace 表、segment 表、worker 表、SDK 接口
10. example-codes 改造建议：替换 repo.NextID(ctx, "draft")，统一 idgen
11. 面试和架构评审要点
```

## 6. 核心观点

附录要明确给出以下结论：

1. **ID 不是一种东西**：实体 ID、业务单号、流程单据 ID、事件 ID、幂等键和链路 ID 的设计目标不同。
2. **Repository 不应该定义发号规则**：仓储负责持久化，ID 生成属于基础设施能力，应通过 `idgen` 或 ID 服务提供。
3. **业务服务不应该传裸字符串 prefix**：`NextID(ctx, "draft")` 应替换为受控 namespace，例如 `NamespaceSupplyDraft`。
4. **核心实体 ID 适合 `int64`**：`item_id`、`spu_id`、`sku_id`、库存账本 ID 等高频索引字段优先使用 `BIGINT`。
5. **对外单号应与内部主键解耦**：订单、支付、退款可以有内部 `BIGINT` 主键和外部业务单号，避免暴露业务量和被枚举。
6. **幂等键不是普通 ID**：幂等键表达“同一次业务请求”的唯一性，必须配合唯一索引、状态机和重试语义。
7. **生产系统通常采用混合方案**：不存在一个算法覆盖所有电商场景，推荐用规则治理不同类型的 ID。

## 7. 推荐 ID 分配策略

| 场景 | 示例 | 推荐类型 | 推荐方案 |
|------|------|----------|----------|
| 商品主数据 | `item_id`、`spu_id`、`sku_id` | `BIGINT` | Segment 号段或 Snowflake |
| 商品组合对象 | `offer_id`、`rate_plan_id` | `BIGINT` 或字符串 | 中小规模可用 Segment，大量外部映射可用字符串 |
| 供给流程单据 | `draft_id`、`staging_id`、`qc_review_id` | 字符串 | ULID/UUIDv7 + 受控 prefix |
| 供给任务 | `task_id`、`batch_id` | 字符串 | ULID/UUIDv7 或业务时间分区编码 |
| 库存事实 | `inventory_record_id`、`stock_ledger_id` | `BIGINT` | Segment 或 Snowflake |
| 库存业务键 | `inventory_key` | 字符串 | 业务组合键，受控格式 |
| 购物车 | `cart_id` | 字符串或 `BIGINT` | 登录用户可绑定 `user_id`，游客车可用 ULID |
| 结算会话 | `checkout_id` | 字符串 | ULID/UUIDv7，配合幂等键 |
| 订单 | `order_id`、`order_no` | 内部 `BIGINT` + 外部字符串 | Snowflake 派生业务单号 |
| 支付 | `payment_id`、`payment_no` | 内部 `BIGINT` + 外部字符串 | Snowflake 或渠道请求号 |
| 退款售后 | `refund_id`、`after_sale_id` | 字符串或 `BIGINT` | Snowflake 派生单号 |
| 营销 | `campaign_id`、`coupon_id` | `BIGINT` | Segment 或 Snowflake |
| 事件 | `event_id`、`outbox_event_id` | 字符串 | ULID/UUIDv7 或确定性事件 ID |
| 链路追踪 | `trace_id`、`operation_id` | 字符串 | Trace 标准或 ULID |
| 幂等 | `idempotency_key` | 字符串 | 客户端请求 ID 或业务语义组合键 |

## 8. 推荐架构

附录中的推荐架构包含以下组件：

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

职责划分：

- **ID Registry**：登记 namespace、ID 类型、生成策略、前缀、长度、是否对外暴露、容量规划和负责人。
- **ID SDK**：业务服务依赖的本地接口，支持 `NextInt64`、`NextString`、`NextBatch`。
- **Segment Generator**：面向主数据、高频实体和不希望依赖时钟的 ID。
- **Snowflake Generator**：面向高并发交易、趋势递增和低延迟发号。
- **ULID/UUIDv7 Generator**：面向流程单据、事件、日志和跨系统追踪。
- **Business Number Formatter**：把底层数字 ID 转成订单号、支付单号、退款单号等对外格式。
- **Observability**：记录发号 QPS、失败率、号段余量、时钟回拨、重复冲突、耗尽预警和 namespace 使用情况。

## 9. 数据模型要求

附录应给出最小可落地表结构，包括：

```text
id_namespace
  记录 namespace、业务域、ID 类型、生成策略、前缀、长度、容量、负责人和状态。

id_segment
  记录每个 namespace 的当前 max_id、step、version，用于号段分配。

id_worker
  记录 Snowflake worker_id、region、datacenter、实例租约和心跳。

id_issue_log
  可选审计表，记录关键 namespace 的发号请求、调用方和错误。
```

这些表以说明治理模型为主，不要求所有业务都同步记录每一次发号。高频发号只记录指标和异常日志，避免把 ID 服务拖成审计瓶颈。

## 10. 示例代码衔接

附录应明确指出当前示例代码中的简化写法：

```go
s.repo.NextID(ctx, "draft")
s.repo.NextID(ctx, "staging")
s.repo.NextID(ctx, "qc")
s.repo.NextID(ctx, "log")
s.repo.NextItemID(ctx)
s.repo.NextID(ctx) // order-service
```

生产演进方向：

```go
type Generator interface {
    NextInt64(ctx context.Context, ns Namespace) (int64, error)
    NextString(ctx context.Context, ns Namespace) (string, error)
    NextBatchInt64(ctx context.Context, ns Namespace, size int) ([]int64, error)
}
```

调用方从裸 prefix 改成受控 namespace：

```go
draftID, err := idgen.NextString(ctx, id.NamespaceSupplyDraft)
itemID, err := idgen.NextInt64(ctx, id.NamespaceProductItem)
orderNo, err := idgen.NextString(ctx, id.NamespaceTradeOrder)
```

该附录只描述改造方向，不直接承担代码重构。

## 11. 风险与边界

附录需要覆盖以下风险：

1. **时钟回拨**：Snowflake 必须处理回拨检测、短暂等待、切换 worker 或熔断。
2. **号段浪费**：Segment 预取会在重启时浪费 ID，这是可接受成本，但要容量规划。
3. **重复发号**：worker_id 租约、DB 乐观锁、唯一索引和冲突监控必须同时存在。
4. **ID 枚举**：对外单号不能直接暴露连续自增，必要时加入日期、随机扰动、校验位或编码。
5. **跨地域冲突**：多机房需要 region bits、独立号段或中心化规划。
6. **字段类型失控**：同一业务域内避免 `string`、`int64`、`varchar number` 混用。
7. **把幂等键当 ID**：幂等键需要业务唯一约束，不能只靠随机字符串。

## 12. 实施顺序

正式写作建议按以下顺序执行：

1. 新增 `appendix/id-system.md`，完成完整附录正文。
2. 更新 `SUMMARY.md`，将其加入附录 H。
3. 在相关章节需要处增加轻量交叉引用，优先考虑商品中心、购物车与结算、订单、支付章节。
4. 更新参考资料，如引用 Snowflake、Leaf、UUIDv7/ULID 等资料。
5. 运行 `npm run clean && npm run build` 验证 mdBook/Hexo 构建。

## 13. 验收标准

1. 附录覆盖电商核心业务域和常见 ID 类型。
2. 读者可以根据表格判断每类 ID 的推荐生成方式。
3. 附录提供至少 5 种方案对比，并明确推荐混合架构。
4. 附录包含 ID 服务接口、表结构、容灾、监控和迁移建议。
5. 附录能解释并改进示例代码里的 `NextID(ctx, "draft")` 类问题。
6. `SUMMARY.md` 中出现附录 H 入口。
7. 构建命令 `npm run clean && npm run build` 通过。
