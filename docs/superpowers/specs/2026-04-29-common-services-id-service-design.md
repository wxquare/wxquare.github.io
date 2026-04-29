# Common Services ID Service 生产级设计规格

## 1. 目标

在 `ecommerce-book/example-codes` 下新增一个独立的公共基础服务示例：

```text
ecommerce-book/example-codes/common-services
```

第一期实现其中的 `id-service`，用于承载电商全局 ID 体系。它不是一个简单的内存工具函数，而是对齐《附录H 全局 ID 体系设计》的生产级示例：统一 namespace、路由不同发号策略、治理 Snowflake worker、支持 Segment 号段、生成可对外展示的业务单号，并暴露 HTTP API、健康检查、指标和审计能力。

这个服务要让读者看到：生产系统里的 ID 生成是基础设施能力，而不是某个 repository 里的 `NextID(ctx, "draft")`。

## 2. 背景

当前示例代码中有三类简化发号逻辑：

```go
s.repo.NextID(ctx, "draft")
s.repo.NextItemID(ctx)
s.repo.NextID(txCtx) // order-service
```

这些写法适合教学早期阶段，但存在明显生产风险：

1. `draft`、`staging`、`qc` 这类裸 prefix 无法治理，也无法审计。
2. 内存序列在多实例部署后不能保证唯一。
3. 时间戳拼接在时钟回拨时可能重复或乱序。
4. 订单服务的 `order_id_seq` 依赖单库自增，跨库、多机房和高并发能力有限。
5. 对外暴露连续递增单号容易泄露业务量，也容易被枚举。

本次新增 `common-services`，先以 ID 服务作为第一个公共能力，后续可继续加入幂等服务、审计服务、配置服务、租户字典等基础设施能力。

## 3. 范围

本次设计覆盖：

1. 新增独立 Go module：`ecommerce-book/example-codes/common-services`。
2. 新增可运行服务入口：`cmd/id-server`。
3. 新增 ID 核心包：Registry、Router、Segment、Snowflake、ULID、Formatter。
4. 新增持久化接口与 MySQL 实现，用于 namespace、segment、worker lease 和 issue log。
5. 新增 HTTP API：单个发号、批量发号、namespace 查询、健康检查、指标。
6. 新增 README，解释运行方式、生产语义和与附录 H 的对应关系。
7. 新增单元测试，覆盖核心发号、路由、异常和并发唯一性。

本次不强制改造 `order-service` 和 `product-service`。示例代码可以先保留原来的简化发号方式，后续再通过独立任务接入 `common-services/id-service`。

## 4. 非目标

第一期不实现完整后台管理台，不实现 Prometheus 原生 client，不实现注册中心接入，不实现跨机房自动故障切换，不实现业务服务 SDK 的所有语言版本。

这些能力会在设计中预留接口，但示例代码保持可读、可运行、可测试。

## 5. 目录结构

目标目录如下：

```text
ecommerce-book/example-codes/common-services/
├── go.mod
├── README.md
├── cmd/
│   └── id-server/
│       └── main.go
└── internal/
    ├── bootstrap/
    │   └── app.go
    ├── idgen/
    │   ├── types.go
    │   ├── registry/
    │   ├── router/
    │   ├── segment/
    │   ├── snowflake/
    │   ├── ulid/
    │   └── formatter/
    ├── infrastructure/
    │   ├── mysql/
    │   ├── lease/
    │   ├── metrics/
    │   └── audit/
    └── interfaces/
        └── http/
```

`idgen` 承载核心模型和算法，`infrastructure` 承载 MySQL、租约、指标和审计，`interfaces/http` 承载对外 API。

## 6. Namespace 设计

业务服务不能传裸字符串 prefix，而要传受控 namespace：

```text
product.item
product.spu
product.sku
supply.draft
supply.staging
supply.qc_review
checkout.session
trade.order
trade.payment
trade.refund
inventory.ledger
inventory.reservation
event.outbox
trace.operation
```

每个 namespace 至少包含：

```go
type NamespaceConfig struct {
    Namespace     string
    BizDomain     string
    IDType        IDType
    GeneratorType GeneratorType
    Prefix        string
    ExposeScope   ExposeScope
    Step          int64
    MaxCapacity   int64
    OwnerTeam     string
    Status        NamespaceStatus
}
```

服务启动时从数据库加载 namespace。为了让示例开箱可跑，MySQL schema 初始化后会写入一组默认 namespace；测试可以使用内存 repository 注入相同配置。

## 7. 发号策略

### 7.1 Segment Generator

用于主数据和容量可规划的实体 ID：

```text
product.item
product.spu
product.sku
marketing.campaign
marketing.coupon
```

Segment 从 `id_segment` 表申请号段。申请时使用乐观锁推进 `max_id` 和 `version`：

```sql
UPDATE id_segment
SET max_id = max_id + step,
    version = version + 1,
    updated_at = NOW(6)
WHERE namespace = ?
  AND version = ?;
```

内存中维护当前号段和下一号段。当前号段使用到阈值后后台预取下一段。当前段耗尽且预取失败时返回明确错误，调用方可重试或失败返回。

### 7.2 Snowflake Generator

用于高并发、趋势递增、`BIGINT` 友好的 ID：

```text
trade.order
trade.payment
trade.refund
inventory.ledger
```

位分配采用 64 位正整数：

```text
1 bit  符号位，固定为 0
41 bit 毫秒时间戳，相对自定义 epoch
5 bit  region_id
5 bit  worker_id
12 bit 毫秒内序列
```

单 worker 每毫秒最多生成 4096 个 ID。`region_id` 支持最多 32 个地域，`worker_id` 支持每地域 32 个 worker。第一期示例使用 `ID_REGION_ID` 指定 region，但 `worker_id` 必须通过租约表申请。

时钟回拨处理：

1. 回拨小于等于可等待阈值时，短暂等待。
2. 回拨超过阈值时，返回 `CLOCK_ROLLBACK` 错误，并记录审计日志和指标。
3. 租约不可用或心跳失败时，Snowflake Generator 进入 not ready 状态，不继续发号。

### 7.3 ULID Generator

用于流程单据、事件、结算会话和链路操作：

```text
supply.draft
supply.staging
supply.qc_review
checkout.session
event.outbox
trace.operation
inventory.reservation
```

ID 格式：

```text
prefix_ulid
```

例如：

```text
draft_01JABCD...
staging_01JABCE...
chk_01JABCF...
evt_01JABCG...
op_01JABCH...
```

ULID 实现使用标准库生成 48 位毫秒时间和 80 位随机数，并用 Crockford Base32 编码，避免引入额外依赖。测试会验证长度、前缀、基本排序和并发唯一性。

### 7.4 Business Number Formatter

对外业务单号不直接暴露内部数字 ID。订单、支付、退款等 namespace 会返回内部 `raw_id` 和外部 `formatted_id`：

```text
ORD + yyyyMMdd + base36(raw_id) + check_digit
PAY + yyyyMMdd + base36(raw_id) + check_digit
RF  + yyyyMMdd + base36(raw_id) + check_digit
```

校验位使用轻量 mod 校验，降低人工录入错误。业务单号不保证连续，但保证可反查内部 ID。

## 8. 数据模型

### 8.1 `id_namespace`

```sql
CREATE TABLE IF NOT EXISTS id_namespace (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    namespace VARCHAR(64) NOT NULL,
    biz_domain VARCHAR(64) NOT NULL,
    id_type VARCHAR(32) NOT NULL,
    generator_type VARCHAR(32) NOT NULL,
    prefix VARCHAR(32) DEFAULT NULL,
    expose_scope VARCHAR(32) NOT NULL,
    step BIGINT NOT NULL DEFAULT 1000,
    max_capacity BIGINT DEFAULT 0,
    owner_team VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    UNIQUE KEY uk_namespace (namespace),
    KEY idx_domain_status (biz_domain, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 8.2 `id_segment`

```sql
CREATE TABLE IF NOT EXISTS id_segment (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    namespace VARCHAR(64) NOT NULL,
    max_id BIGINT NOT NULL,
    step BIGINT NOT NULL,
    version BIGINT NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    UNIQUE KEY uk_namespace (namespace)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 8.3 `id_worker`

```sql
CREATE TABLE IF NOT EXISTS id_worker (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    worker_id INT NOT NULL,
    region_id INT NOT NULL,
    datacenter_code VARCHAR(32) NOT NULL,
    instance_id VARCHAR(128) NOT NULL,
    lease_token VARCHAR(64) NOT NULL,
    lease_until DATETIME(6) NOT NULL,
    heartbeat_at DATETIME(6) NOT NULL,
    status VARCHAR(32) NOT NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    UNIQUE KEY uk_worker_region (worker_id, region_id),
    UNIQUE KEY uk_instance (instance_id),
    KEY idx_status_lease (status, lease_until)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 8.4 `id_issue_log`

```sql
CREATE TABLE IF NOT EXISTS id_issue_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    request_id VARCHAR(64) NOT NULL,
    namespace VARCHAR(64) NOT NULL,
    caller VARCHAR(128) NOT NULL,
    issue_type VARCHAR(32) NOT NULL,
    issued_value VARCHAR(128) DEFAULT NULL,
    error_code VARCHAR(64) DEFAULT NULL,
    error_message VARCHAR(512) DEFAULT NULL,
    created_at DATETIME(6) NOT NULL,
    UNIQUE KEY uk_request_id (request_id),
    KEY idx_namespace_time (namespace, created_at),
    KEY idx_issue_type_time (issue_type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

审计默认只记录异常、worker 租约、号段申请和关键 namespace 的请求，不同步记录所有高频成功发号。

## 9. HTTP API

### 9.1 单个发号

```http
POST /api/v1/ids/next
Content-Type: application/json
```

请求：

```json
{
  "namespace": "trade.order",
  "caller": "order-service",
  "request_id": "req-20260429-0001"
}
```

响应：

```json
{
  "namespace": "trade.order",
  "id_type": "BUSINESS_NO",
  "generator": "SNOWFLAKE",
  "raw_id": 1928475629384753152,
  "id": "ORD20260429CN7K3F9Q2X",
  "issued_at": "2026-04-29T12:00:00.123456+08:00"
}
```

### 9.2 批量发号

```http
POST /api/v1/ids/batch
```

请求：

```json
{
  "namespace": "product.sku",
  "caller": "product-service",
  "request_id": "req-20260429-0002",
  "count": 100
}
```

响应：

```json
{
  "namespace": "product.sku",
  "count": 100,
  "ids": ["600000001001", "600000001002"]
}
```

第一期限制单次批量大小，默认最大 1000，避免超大请求拖垮服务。

### 9.3 管理与健康接口

```text
GET /api/v1/namespaces
GET /healthz
GET /readyz
GET /metrics
```

`healthz` 表示进程存活，`readyz` 表示 namespace 已加载、worker 租约有效、核心 generator 可用。`metrics` 使用文本格式输出计数器和 gauge，便于阅读和后续接入 Prometheus。

## 10. 错误语义

服务返回结构化错误：

```json
{
  "error": {
    "code": "CLOCK_ROLLBACK",
    "message": "clock moved backwards by 800ms",
    "namespace": "trade.order",
    "retryable": false
  }
}
```

核心错误码：

| 错误码 | 含义 | 是否建议重试 |
|--------|------|--------------|
| `NAMESPACE_NOT_FOUND` | namespace 未注册 | 否 |
| `NAMESPACE_DISABLED` | namespace 被禁用 | 否 |
| `GENERATOR_NOT_READY` | generator 尚未就绪 | 是 |
| `SEGMENT_EXHAUSTED` | 当前号段耗尽且预取失败 | 是 |
| `WORKER_LEASE_LOST` | worker 租约丢失 | 是，需等待实例恢复 |
| `CLOCK_ROLLBACK` | 时钟回拨超过阈值 | 否，需人工或平台处理 |
| `BATCH_TOO_LARGE` | 批量数量超过限制 | 否 |
| `INVALID_REQUEST` | 请求参数错误 | 否 |

业务服务必须能感知发号失败。ID 服务允许跳号，不允许复用已经发出的 ID。

## 11. Worker 租约

Snowflake 的 `worker_id` 由租约表分配，不能靠配置文件随手写。

启动流程：

1. 实例生成或读取 `instance_id`。
2. 根据 `region_id` 扫描可用 `worker_id`。
3. 对过期租约执行抢占，写入新的 `lease_token`。
4. 启动后台心跳，定期延长 `lease_until`。
5. 心跳失败或租约被抢占时，服务进入 not ready，Snowflake 停止发号。

退出流程：

1. 正常退出时释放租约。
2. 异常退出时依赖 `lease_until` 自动过期。

## 12. 可观测性

第一期实现轻量指标，不引入第三方依赖：

```text
idgen_requests_total{namespace,generator,result}
idgen_batch_requests_total{namespace,result}
idgen_errors_total{namespace,code}
idgen_segment_allocations_total{namespace,result}
idgen_segment_remaining{namespace}
idgen_clock_rollback_total{namespace}
idgen_worker_lease_status{region_id,worker_id}
```

日志至少包含：

```text
request_id
caller
namespace
generator
raw_id
formatted_id
error_code
latency_ms
```

异常路径写 `id_issue_log`，常规高频成功路径只打指标和结构化日志。

## 13. 配置

通过环境变量配置：

```text
ID_SERVICE_ADDR=:8090
ID_MYSQL_DSN=root:root@tcp(127.0.0.1:3306)/common_services?parseTime=true&charset=utf8mb4&loc=Local
ID_REGION_ID=1
ID_DATACENTER_CODE=local-a
ID_INSTANCE_ID=hostname-pid
ID_WORKER_LEASE_TTL_SECONDS=30
ID_WORKER_HEARTBEAT_SECONDS=10
ID_SNOWFLAKE_EPOCH=2026-01-01T00:00:00Z
ID_MAX_BATCH_SIZE=1000
```

`ID_INSTANCE_ID` 未配置时使用 hostname、pid 和启动时间生成。生产部署中应该由平台注入稳定实例标识。

## 14. 测试策略

必须覆盖：

1. Registry 加载默认 namespace。
2. 非法 namespace 返回 `NAMESPACE_NOT_FOUND`。
3. Segment 单实例连续发号不重复。
4. Segment 并发发号不重复。
5. Segment 号段耗尽后申请下一段。
6. Snowflake 单实例连续发号不重复且趋势递增。
7. Snowflake 时钟回拨小于阈值时等待。
8. Snowflake 时钟回拨超过阈值时返回 `CLOCK_ROLLBACK`。
9. ULID 前缀、长度和并发唯一性。
10. Business Number Formatter 不直接暴露原始递增值。
11. HTTP 单个发号、批量发号、错误响应和 readyz。

核心算法测试使用内存 repository 和可注入 clock，避免依赖真实 MySQL。MySQL schema 和运行方式在 README 中提供。

## 15. 验收标准

实现完成后应满足：

1. `common-services` 是独立 Go module，可单独运行 `go test ./...`。
2. `go run ./cmd/id-server` 可以启动 HTTP 服务。
3. `POST /api/v1/ids/next` 能为 `trade.order` 返回业务单号。
4. `POST /api/v1/ids/batch` 能为 `product.sku` 返回批量 `BIGINT`。
5. 未注册 namespace 返回结构化错误。
6. Snowflake worker 通过租约获取，不通过硬编码 worker_id。
7. 时钟回拨和租约丢失不会静默发号。
8. README 能解释该实现与《附录H》的对应关系。

## 16. 后续演进

后续可以在这个 module 中继续扩展：

1. 订单服务接入 `id-service`，拆分内部 `order_id` 和外部 `order_no`。
2. 商品服务接入 `id-service`，替换 `NextItemID` 和供给流程中的裸 prefix。
3. 增加 SDK 包，封装 HTTP 调用、重试、超时、熔断和指标。
4. 增加 Prometheus client、OpenTelemetry trace 和管理后台。
5. 增加多机房策略：region bits 固化、跨 region 号段隔离、灾备演练。
