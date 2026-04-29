# 附录H 全局 ID 体系设计

## 1. 为什么电商系统需要全局 ID 体系

在示例代码中，供给链路为了演示流程，使用了类似下面的写法：

```go
s.repo.NextID(ctx, "draft")
```

仓储内部再用时间戳、前缀和内存序列拼出一个 ID。这种写法适合教学 Demo，因为它能让读者把注意力放在 Draft、Staging、QC、Publish 的业务流程上；但在生产系统中，这类发号逻辑很快会失控。

问题不在于这行代码短，而在于它隐藏了很多关键决策：

1. `draft` 这个前缀是谁定义的？其他服务能不能复用？
2. 多实例部署后，内存序列是否仍然唯一？
3. 机器时钟回拨时，时间戳拼接是否会重复？
4. 发号失败时，业务层是否能感知并回滚？
5. 这个 ID 是否对外暴露？是否容易被枚举？
6. 未来从单机迁到多机房，是否还能保持唯一？
7. 日志、指标、容量预警和审计在哪里做？

电商系统里的 ID 远不止一个 `draft_id`。商品有 `item_id`、`spu_id`、`sku_id`，交易有 `checkout_id`、`order_id`、`payment_id`，供给有 `draft_id`、`staging_id`、`qc_review_id`，库存有 `inventory_key`、`reservation_id`，事件有 `event_id`、`outbox_event_id`，链路上还有 `trace_id` 和 `operation_id`。这些 ID 的业务语义、性能要求、暴露范围和容灾策略都不同。

所以，全局 ID 体系的目标不是发明一个“万能 ID 算法”，而是建立一套规则：

```text
什么业务对象使用什么 ID 类型；
什么 ID 由哪个 namespace 管理；
什么 ID 可以对外暴露；
什么 ID 需要趋势递增；
什么 ID 需要可读业务单号；
什么 ID 只是幂等键或链路追踪键；
发号失败、重复、耗尽、时钟回拨时如何处理。
```

一句话概括：**ID 体系是电商系统的基础设施治理问题，不只是一个工具函数问题。**

## 2. 电商 ID 分类

设计 ID 之前，先不要问“用不用 Snowflake”，而要问“这个 ID 表达什么业务语义”。下表给出电商系统中最常见的 ID 分类。

| 类型 | 典型字段 | 设计重点 | 不适合的做法 |
|------|----------|----------|--------------|
| 实体 ID | `item_id`、`spu_id`、`sku_id` | 长期稳定、索引友好、跨系统引用 | 每个服务各自自增 |
| 业务单号 | `order_no`、`payment_no`、`refund_no` | 对外展示、客服查询、对账、不可枚举 | 直接暴露连续自增 |
| 流程单据 ID | `draft_id`、`staging_id`、`qc_review_id` | 流程追踪、审计、低耦合 | 与正式商品 ID 混用 |
| 事件 ID | `event_id`、`outbox_event_id` | 幂等消费、重放、排障 | 用时间戳字符串拼接 |
| 幂等键 | `idempotency_key`、`request_id` | 表达同一次业务请求 | 当作普通随机 ID |
| 链路 ID | `trace_id`、`operation_id` | 跨服务追踪和审计 | 每层重新生成 |

这张表背后的关键判断是：**ID 的生成方式要服从它的业务用途。**

例如，`sku_id` 通常是商品主数据的稳定实体 ID，适合使用 `BIGINT`，方便数据库索引、缓存 Key、消息体和下游系统引用；`order_no` 是对外业务单号，除了唯一之外，还要考虑客服查询、对账、不可枚举和格式兼容；`idempotency_key` 则不是普通 ID，它表达“同一次业务请求”，必须配合唯一索引和状态机来防止重复下单、重复扣款或重复退票。

## 3. 全场景 ID 清单

下面的矩阵不是要求所有公司都照抄，而是给出一个可评审的默认选择。实际落地时，可以根据规模、团队能力、数据库类型、是否多机房和是否对外开放 API 做取舍。

| 业务域 | 关键 ID | 推荐类型 | 推荐生成方式 | 设计说明 |
|--------|---------|----------|--------------|----------|
| 商品中心 | `item_id`、`spu_id`、`sku_id` | `BIGINT` | Segment 号段或 Snowflake | 高频查询和跨系统引用，优先索引友好 |
| 商品组合 | `offer_id`、`rate_plan_id` | `BIGINT` 或字符串 | Segment，外部映射可用字符串 | 本地 Offer 用平台 ID，供应商编码单独保存 |
| 供给流程 | `draft_id`、`staging_id`、`qc_review_id` | 字符串 | ULID/UUIDv7 + 受控 prefix | 流程单据不应与正式商品 ID 混用 |
| 供给任务 | `task_id`、`batch_id`、`sync_batch_id` | 字符串 | ULID/UUIDv7 或业务时间分区编码 | 长任务、批处理和补偿需要可追踪 |
| 库存事实 | `stock_ledger_id`、`reservation_id` | `BIGINT` 或字符串 | Segment、Snowflake 或 ULID | 账本可用 `BIGINT`，预占凭证可用字符串 |
| 库存业务键 | `inventory_key` | 字符串 | 业务组合键 | 表达 SKU、范围、日期、渠道、供应商等维度 |
| 购物车 | `cart_id` | 字符串或 `BIGINT` | 登录态绑定 `user_id`，游客车用 ULID | 登录购物车可弱化独立 ID，游客车需要会话标识 |
| 结算 | `checkout_id` | 字符串 | ULID/UUIDv7 + 幂等键 | 一次结算会话要能重试、恢复和防重复 |
| 订单 | `order_id`、`order_no` | 内部 `BIGINT` + 外部字符串 | Snowflake 派生业务单号 | 内部主键和对外单号解耦 |
| 支付 | `payment_id`、`payment_no`、`channel_trade_no` | 内部 `BIGINT` + 外部字符串 | Snowflake 或渠道请求号 | 平台支付单和渠道单号都要保存 |
| 售后 | `refund_id`、`after_sale_id` | 字符串或 `BIGINT` | Snowflake 派生单号 | 便于客服、对账和售后流转 |
| 营销 | `campaign_id`、`coupon_id`、`promotion_id` | `BIGINT` | Segment 或 Snowflake | 营销对象数量大，需稳定引用 |
| 搜索 | `index_task_id`、`doc_id` | 字符串 | 业务 ID 或 ULID | 搜索文档通常以业务实体 ID 为主键 |
| 履约 | `fulfillment_id`、`delivery_order_no` | 字符串 | Snowflake 派生单号或外部单号 | 履约单经常要与供应商、物流系统对接 |
| 财务 | `ledger_id`、`settlement_id`、`reconciliation_id` | `BIGINT` 或字符串 | Segment、Snowflake、批次号 | 账务更重视可追溯、不可重复和对账批次 |
| 事件 | `event_id`、`outbox_event_id` | 字符串 | ULID/UUIDv7 或确定性事件 ID | 用于幂等消费、重放和排障 |
| 链路追踪 | `trace_id`、`operation_id` | 字符串 | Trace 标准或 ULID | 跨服务传递，不在每一层重新生成 |
| 幂等 | `idempotency_key` | 字符串 | 客户端请求 ID 或业务语义组合键 | 依赖唯一约束和状态机，不等同于随机 ID |

这里有几个容易混淆的点：

1. `order_id` 和 `order_no` 可以不是同一个字段。前者可以是内部主键，后者是对外业务单号。
2. `inventory_key` 通常不是随机 ID，而是业务维度组合，例如 `inv:sku:30001:global` 或 `inv:sku:40001:date:2026-05-01:channel:app`。
3. `checkout_id` 不是订单号。结算会话可能失败、过期或被重试，只有创单成功后才产生订单。
4. `idempotency_key` 的核心不是“看起来唯一”，而是业务上能判断“这是不是同一次请求”。

## 4. 常见发号方案对比

### 4.1 DB 自增

DB 自增是最简单的方案：表主键使用 `AUTO_INCREMENT` 或数据库原生 identity。它适合单库单表、小规模后台配置、内部字典表和教学示例。

优点是简单、强一致、无需额外服务。缺点也明显：强依赖单库，跨库分表困难；连续递增容易暴露业务量；高并发交易链路可能把数据库打成瓶颈。

在电商系统中，DB 自增可以用于后台低频配置表，但不建议直接作为对外订单号、支付单号或全局 SKU ID。

### 4.2 DB Sequence 表

Sequence 表通过插入一张专门的序列表获取 `LastInsertId`，示例中的订单服务就有类似思路：

```sql
CREATE TABLE order_id_seq (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(6) NOT NULL
);
```

每生成一个订单号，就插入一行序列表，再把自增值格式化成 `ORD-123`。这个方案比直接使用业务表自增稍微解耦，但本质仍是数据库中心化发号。

它适合早期系统、低中并发内部单据和容易理解的教学实现。不适合高并发交易核心，也不适合直接对外暴露连续序列。

### 4.3 Redis INCR

Redis `INCR` 可以按 key 递增，例如：

```text
INCR id:order:20260429
```

再格式化为：

```text
ORD2026042900012345
```

它的优点是性能高、实现简单、天然适合按天流水号。缺点是强依赖 Redis 高可用和持久化策略；主从切换、数据回滚、双活部署时要谨慎；同时，按天连续递增仍可能暴露业务量。

Redis INCR 适合活动流水、短期批次、低风险业务编号。核心订单和支付单如果使用 Redis INCR，必须设计持久化、主从切换和重复保护。

### 4.4 Snowflake

Snowflake 是经典的分布式趋势递增 ID 方案。常见实现把一个 64 位整数拆成：

```text
时间戳 + 机器 / 机房标识 + 毫秒内序列
```

例如常见切分是 41 位毫秒时间戳、10 位机器标识、12 位序列。它的优点是本地生成、低延迟、高吞吐、趋势递增、适合 `BIGINT` 主键。缺点是依赖时钟，必须治理 `worker_id`，还要处理时钟回拨。

Snowflake 适合订单内部 ID、支付内部 ID、库存账本 ID、营销 ID，以及需要高并发写入的实体 ID。对外单号可以基于 Snowflake 再格式化，而不是直接暴露原始数字。

### 4.5 Segment 号段

Segment 号段，也叫 Hi-Lo 模式。核心思想是数据库只负责分配一段 ID，服务实例拿到号段后在本地内存中发号：

```text
product.sku 申请到 1000000 - 1009999
product.sku 申请到 1010000 - 1019999
```

数据库中通常维护：

```text
namespace、max_id、step、version
```

服务用乐观锁推进 `max_id`，一次拿一段。这样既保留数据库的强一致分配，又避免每个 ID 都访问数据库。

优点是不强依赖时钟，容量可控，namespace 独立，适合主数据 ID。缺点是服务重启会浪费一段号；号段耗尽前要预取；如果数据库不可用，新的号段无法分配。

Segment 非常适合 `item_id`、`spu_id`、`sku_id`、`campaign_id`、`coupon_id` 等电商主数据 ID。

### 4.6 UUIDv7、ULID 与 KSUID

UUID、ULID、KSUID 都属于更偏字符串或 128 位标识的方案。相较传统 UUIDv4，UUIDv7、ULID 和 KSUID 更强调时间有序或近似时间有序，适合日志、事件、流程单据和跨服务追踪。

UUIDv7 已在 RFC 9562 中定义，它把 Unix 毫秒时间放在高位，并用随机位提供唯一性。ULID 也采用时间 + 随机的思路，字符串更短、更适合人类阅读和按字典序排序。

这类 ID 的优点是无需中心服务，跨服务生成方便，天然适合字符串前缀。缺点是比 `BIGINT` 长，索引和存储成本更高，不适合所有高频实体都使用。

推荐用于 `draft_id`、`staging_id`、`qc_review_id`、`operation_id`、`event_id`、`outbox_event_id`、`checkout_id` 等场景。

| 方案 | 是否中心化 | 是否趋势递增 | 主要优点 | 主要风险 | 推荐场景 |
|------|------------|--------------|----------|----------|----------|
| DB 自增 | 是 | 是 | 简单、强一致 | 单点瓶颈、暴露业务量 | 小规模后台表 |
| DB Sequence | 是 | 是 | 与业务表解耦、易理解 | 高并发瓶颈、跨库困难 | 早期单据号、教学示例 |
| Redis INCR | 是 | 是 | 高性能、适合按天流水 | 持久化和主从切换风险 | 活动流水、短期批次 |
| Snowflake | 否 | 大体是 | 低延迟、高吞吐、`BIGINT` 友好 | 时钟回拨、worker 分配 | 订单、支付、库存账本 |
| Segment 号段 | 半中心化 | 是 | 不依赖时钟、容量可控 | 号段浪费、依赖号段预取 | 商品、营销、主数据 |
| UUIDv7 / ULID | 否 | 是 | 无需协调、跨服务方便 | 字段较长、索引成本高 | 流程单据、事件、链路 |

## 5. 推荐混合架构

生产电商系统更常见的不是“全站只用一个算法”，而是混合架构：

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

推荐默认规则如下：

```text
商品和库存主数据：Segment 或 Snowflake 的 BIGINT
交易单号：Snowflake 派生业务单号
供给流程、事件和链路：ULID/UUIDv7 + 受控 prefix
幂等：业务语义唯一约束，不等同于普通 ID
```

也就是说：

1. `item_id`、`spu_id`、`sku_id` 这类主数据 ID 优先使用 `BIGINT`，便于索引和跨系统传递。
2. `order_no`、`payment_no`、`refund_no` 这类对外单号可以在底层 Snowflake ID 上增加日期、渠道、校验位或编码。
3. `draft_id`、`staging_id`、`qc_review_id` 这类流程单据 ID 使用字符串，更适合审计、日志和跨系统排障。
4. `idempotency_key` 不由 ID 服务随便生成，而要和用户、购物车快照、请求来源或业务动作绑定。

这个混合架构可以同时满足性能、可读性、治理和扩展性。

## 6. ID 服务架构

### 6.1 ID Registry

ID Registry 是 ID 体系的控制面，负责登记所有 namespace，例如：

```text
product.item
product.spu
product.sku
supply.draft
supply.staging
trade.order
trade.payment
event.outbox
```

每个 namespace 至少要记录：

| 字段 | 含义 |
|------|------|
| `namespace` | 全局唯一的业务命名空间 |
| `biz_domain` | 所属业务域 |
| `id_type` | `INT64`、`STRING`、`BUSINESS_NO` 或 `IDEMPOTENCY_KEY` |
| `generator_type` | `SEGMENT`、`SNOWFLAKE`、`ULID`、`UUIDV7` 或 `BUSINESS` |
| `prefix` | 字符串 ID 或业务单号前缀 |
| `expose_scope` | `INTERNAL`、`EXTERNAL` 或 `MIXED` |
| `owner_team` | 负责人团队 |
| `status` | `ENABLED`、`DISABLED` 或 `DEPRECATED` |

不要让业务代码直接传 `"draft"`、`"order"` 这种裸字符串。裸字符串无法治理，也无法做容量规划和审计。

### 6.2 ID SDK

业务服务应该依赖 SDK，而不是直接访问 ID 表或自己拼接字符串。SDK 至少提供：

```go
type Generator interface {
    NextInt64(ctx context.Context, ns Namespace) (int64, error)
    NextString(ctx context.Context, ns Namespace) (string, error)
    NextBatchInt64(ctx context.Context, ns Namespace, size int) ([]int64, error)
}
```

SDK 可以封装本地缓存、号段预取、熔断降级、指标上报和错误转换。业务服务只关心“我要哪个 namespace 的 ID”。

### 6.3 Generator Router

Generator Router 根据 namespace 配置路由到不同发号器：

```text
product.sku       -> Segment Generator
trade.order       -> Snowflake Generator + Business Number Formatter
supply.draft      -> ULID Generator
event.outbox      -> UUIDv7 Generator
checkout.session  -> ULID Generator
```

这样可以把“业务 ID 规则”从业务代码中拿出来，避免仓储层、应用层、HTTP 层各自发明一套规则。

### 6.4 Segment Generator

Segment Generator 从数据库申请号段，然后在本地内存中发号。为了避免号段耗尽造成请求抖动，应该支持双 Buffer：

```text
当前号段使用到 70% 时，后台预取下一段；
当前号段耗尽时，如果下一段已就绪，立即切换；
预取失败时，继续使用当前号段并告警；
当前号段完全耗尽且无法预取时，返回明确错误。
```

### 6.5 Snowflake Generator

Snowflake Generator 的关键不是位运算，而是 worker 治理：

1. `worker_id` 不能靠配置文件随手写，应该由租约表、注册中心或部署平台分配。
2. 实例启动时申请 worker，定期心跳，退出或过期后释放。
3. 发现时钟回拨时，要短暂等待、切换 worker 或熔断，而不是继续发号。
4. 多机房部署时，要预留 region 或 datacenter 位。

### 6.6 ULID / UUIDv7 Generator

这类生成器适合本地生成，但仍然要受 namespace 约束。推荐格式：

```text
draft_01JABCD...
staging_01JABCE...
qc_01JABCF...
evt_01JABCG...
op_01JABCH...
```

prefix 不是随意字符串，而是 Registry 中登记过的前缀。这样日志、排障和数据治理可以快速识别 ID 类型。

### 6.7 Business Number Formatter

业务单号通常不直接等于底层 ID。订单号可以设计为：

```text
ORD + yyyyMMdd + base36(snowflake_id) + check_digit
```

例如：

```text
ORD20260429CN7K3F9Q2X
```

这种格式便于客服和对账按日期定位，同时不直接暴露连续自增值。校验位可以降低人工录入错误。

### 6.8 Observability / Audit / Admin

ID 服务必须可观测：

| 指标 | 说明 |
|------|------|
| `idgen_qps` | 各 namespace 发号 QPS |
| `idgen_error_rate` | 发号失败率 |
| `segment_remaining` | 当前号段剩余比例 |
| `segment_alloc_latency` | 申请号段耗时 |
| `clock_rollback_count` | 时钟回拨次数 |
| `worker_lease_expired_count` | worker 租约过期次数 |
| `duplicate_key_error_count` | 下游唯一键冲突次数 |

高频 ID 不应把每次发号都同步写审计表，否则 ID 服务会被审计拖垮。更合理的方式是：常规路径打指标，异常路径写审计。

## 7. 关键业务 ID 设计

### 7.1 `sku_id`、`spu_id` 与 `item_id`

`item_id` 是前台商品入口，`spu_id` 是商品定义层的标准品，`sku_id` 是具体销售规格。它们都属于长期稳定的主数据 ID，推荐使用 `BIGINT`。

默认选择：

```text
product.item -> Segment
product.spu  -> Segment
product.sku  -> Segment
```

如果系统写入并发特别高，也可以改成 Snowflake，但要统一 worker 管理。无论使用哪种方案，ID 一旦发出就不应复用。草稿废弃、商品下架、SKU 删除都不应该回收 ID。

供给链路中是否提前生成 `sku_id`，取决于业务：

1. 如果外部供应商、图片、库存、审核都需要提前引用 SKU，可以在 Draft 阶段占号，状态为 `RESERVED`。
2. 如果希望未审核数据完全不污染正式商品空间，可以在 Publish 成功时生成正式 `sku_id`。

两种方案都可行，但必须在附录和代码中讲清楚边界。

### 7.2 `order_id` 与 `order_no`

订单建议内部主键和对外单号解耦：

```sql
CREATE TABLE orders (
    id BIGINT PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL,
    user_id BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL,
    created_at DATETIME NOT NULL,
    UNIQUE KEY uk_order_no (order_no)
);
```

其中：

```text
id       -> 内部主键，Snowflake 或 Segment
order_no -> 对外业务单号，Snowflake 派生格式
```

不要直接暴露 `ORD-1`、`ORD-2` 这类连续单号。它会暴露业务量，也容易被枚举。

### 7.3 `checkout_id` 与 `idempotency_key`

`checkout_id` 表达一次结算会话，`idempotency_key` 表达一次业务请求。它们可以相关，但不能混为一谈。

典型设计：

```text
checkout_id = ULID
idempotency_key = user_id + cart_snapshot_hash + client_request_id
```

创单时，订单系统需要唯一约束：

```sql
UNIQUE KEY uk_order_idempotency (user_id, idempotency_key)
```

这样用户重复点击“提交订单”时，系统返回同一笔订单，而不是生成多笔订单。

### 7.4 `payment_id`、渠道单号与对账

支付系统至少要区分三类编号：

| 字段 | 说明 |
|------|------|
| `payment_id` | 平台内部支付主键 |
| `payment_no` | 平台对外支付单号 |
| `channel_trade_no` | 支付渠道返回的交易号 |

平台调用渠道时，还需要一个稳定的渠道请求号，例如 `out_trade_no`。这个请求号通常应该由平台生成，并作为调用渠道的幂等键。不要用渠道返回单号作为平台支付单的唯一依据，因为渠道单号只有调用成功后才出现。

### 7.5 `draft_id`、`staging_id` 与供给审核单

供给流程 ID 推荐使用字符串：

```text
draft_01J...
staging_01J...
qc_01J...
```

原因是它们不是正式商品资产，不需要像 `sku_id` 一样参与高频交易查询。字符串 prefix 能快速表达流程类型，便于运营后台、日志检索和问题排查。

关键边界是：Draft、Staging、QC 阶段的 ID 不应替代正式 `item_id`、`spu_id`、`sku_id`。只有发布事务成功后，商品中心才持有正式商品主数据 ID。

### 7.6 `event_id` 与 Outbox 去重

事件 ID 需要支持幂等消费和重放。常见方案有两种：

1. 随机或时间有序 ID，例如 `evt_01J...`。
2. 确定性事件 ID，例如 `evt_product_published_{item_id}_{version}`。

对于 Outbox，确定性事件 ID 很有价值，因为同一个聚合版本只应该发布一次事件。消费者侧仍然要有处理表或唯一索引：

```sql
CREATE TABLE event_consume_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    consumer_group VARCHAR(64) NOT NULL,
    event_id VARCHAR(128) NOT NULL,
    consumed_at DATETIME NOT NULL,
    UNIQUE KEY uk_consumer_event (consumer_group, event_id)
);
```

这样即使消息系统 at-least-once 投递，也能实现业务上的精确一次效果。

## 8. 容灾、风险与治理

| 风险 | 表现 | 缓解策略 |
|------|------|----------|
| 时钟回拨 | Snowflake 生成重复或乱序 ID | 使用 NTP 单调配置；检测回拨；短暂等待；超过阈值熔断或切换 worker |
| 号段浪费 | Segment 服务重启后未用完的 ID 丢失 | 接受不连续；容量规划；合理设置 step；禁止回收已发号段 |
| 重复发号 | 多实例使用同一 worker 或并发申请同一号段 | worker 租约；DB 乐观锁；唯一索引；重复冲突告警 |
| ID 枚举 | 外部用户通过连续 ID 猜测订单量或访问资源 | 内外 ID 解耦；业务单号编码；权限校验；必要时加校验位 |
| 跨地域冲突 | 多机房各自发号后 ID 冲突 | 预留 region bits；按 region 分段；中心化 namespace 规划 |
| 字段类型失控 | 同一个 ID 在不同系统里一会儿是字符串，一会儿是数字 | 统一契约；IDL / OpenAPI 固化类型；迁移期双字段兼容 |
| 把幂等键当 ID | 重试请求仍然生成多笔订单或多次扣款 | 唯一约束；请求状态表；幂等返回；业务状态机保护 |

电商系统还要特别注意“唯一性不是只靠 ID 服务保证”。最终写入业务表时仍然要有唯一索引。ID 服务负责降低冲突概率和统一规则，业务数据库负责最后一道硬约束。

## 9. 数据库与接口设计

### 9.1 Namespace 注册表

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
```

### 9.2 Segment 号段表

```sql
CREATE TABLE id_segment (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    namespace VARCHAR(64) NOT NULL,
    max_id BIGINT NOT NULL COMMENT '当前已经分配到的最大 ID',
    step INT NOT NULL COMMENT '每次申请的号段大小',
    version BIGINT NOT NULL COMMENT '乐观锁版本',
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_namespace (namespace)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Segment 号段表';
```

申请号段时使用乐观锁：

```sql
UPDATE id_segment
SET max_id = max_id + step,
    version = version + 1,
    updated_at = NOW()
WHERE namespace = ?
  AND version = ?;
```

### 9.3 Snowflake Worker 租约表

```sql
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
```

### 9.4 发号审计与异常记录

```sql
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

审计表不应该记录所有高频发号请求。建议只记录关键 namespace、号段申请、异常、回拨和人工操作。

### 9.5 Go SDK 接口

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

应用服务依赖这个接口，仓储只负责保存实体：

```go
type SupplyOpsService struct {
    repo  SupplyRepository
    idgen id.Generator
}
```

## 10. 示例代码改造建议

当前示例中的写法是：

```go
DraftID: s.repo.NextID(ctx, "draft"),
```

生产级演进方向是：

```go
draftID, err := s.idgen.NextString(ctx, id.NamespaceSupplyDraft)
if err != nil {
    return nil, err
}
```

再构造领域对象：

```go
draft := &domain.ProductSupplyDraft{
    DraftID:     draftID,
    OperationID: operationID,
    Status:      domain.DraftStatusDraft,
    CreatedAt:   now,
    UpdatedAt:   now,
}
```

`ProductCenterRepository.NextItemID(ctx)` 也可以演进为：

```go
itemID, err := s.idgen.NextInt64(ctx, id.NamespaceProductItem)
if err != nil {
    return nil, err
}
```

订单服务中的 `NextOrderID(ctx)` 可以拆成两层：

```go
internalID, err := s.idgen.NextInt64(ctx, id.NamespaceTradeOrder)
if err != nil {
    return nil, err
}

orderNo := s.orderNoFormatter.Format(internalID, time.Now())
```

这样，仓储不再定义 ID 规则，业务服务也不再传裸 prefix。所有 namespace、生成策略和对外格式都由 ID 体系统一治理。

本附录只给出改造方向，不要求立刻重构示例代码。教学代码可以保留简化实现，但正文要让读者知道生产系统应该如何演进。

## 11. 面试和架构评审要点

设计全局 ID 体系时，可以用下面的问题自查：

1. 这个 ID 是内部实体 ID、对外业务单号、流程单据 ID、事件 ID，还是幂等键？
2. 这个 ID 是否会出现在 URL、订单详情、客服系统或开放 API 中？
3. 如果对外暴露，是否会泄露业务量或被枚举？
4. 这个 ID 是否需要趋势递增？是否真的需要严格递增？
5. 数据库主键是 `BIGINT` 还是字符串？索引成本是否可接受？
6. 多实例部署时，worker 或号段如何分配？
7. 机器时钟回拨时，系统等待、降级还是熔断？
8. 多机房部署时，是否预留 region 或 datacenter 位？
9. ID 服务不可用时，业务是失败、降级还是使用本地缓存号段？
10. 最终业务表是否有唯一索引兜底？
11. 幂等键是否有业务语义，还是只是随机字符串？
12. 老系统 ID 如何迁移？是否需要双写、双查或映射表？

如果这些问题没有答案，就说明 ID 体系还停留在工具函数层面，没有进入基础设施治理层面。

## 12. 小结

统一 ID 体系的重点不是某个算法，而是按业务语义治理 namespace、生成策略、暴露形式和失败处理。

电商系统中，`sku_id`、`order_no`、`draft_id`、`event_id` 和 `idempotency_key` 看起来都叫 ID，但它们解决的问题完全不同。生产级设计应该把它们分开建模：

```text
实体 ID：稳定引用，索引友好
业务单号：对外展示，防枚举，便于对账
流程单据：追踪流程，支持审计
事件 ID：幂等消费，支持重放
幂等键：表达同一次业务请求
链路 ID：贯穿调用链和操作链
```

推荐的默认架构是：**Segment 号段 + Snowflake + ULID/UUIDv7 + 幂等键治理**。这套组合不是最炫的方案，但足够贴近真实电商系统：它尊重不同业务场景的差异，也给后续多实例、多机房、开放平台和长期演进留下空间。
