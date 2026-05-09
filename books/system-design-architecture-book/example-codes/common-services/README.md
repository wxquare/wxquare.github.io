# Common Services - ID Service

这个示例实现《附录H 全局 ID 体系设计》中的公共 ID 服务。它展示的是生产级边界：namespace 治理、Segment 号段、Snowflake worker 租约、ULID 流程单据、业务单号格式化、HTTP API、健康检查和指标。

## 能力

- Registry 统一管理 `product.sku`、`trade.order`、`supply.draft` 等 namespace。
- Segment 生成 `product.item`、`product.spu`、`product.sku`。
- Snowflake 生成 `trade.order`、`trade.payment`、`trade.refund`、`inventory.ledger`。
- ULID 生成 `supply.draft`、`checkout.session`、`event.outbox`、`trace.operation`。
- Business Number Formatter 生成 `ORD20260429...` 这类外部业务单号。
- HTTP API 暴露单个发号、批量发号、namespace 查询、健康检查和指标。

## 运行

默认使用内存模式，便于直接运行：

```bash
go test ./...
go run ./cmd/id-server
```

配置 `ID_MYSQL_DSN` 后会切换到 MySQL 持久化模式，并自动初始化 `id_namespace`、`id_segment`、`id_worker`、`id_issue_log`：

```bash
export ID_MYSQL_DSN='root:root@tcp(127.0.0.1:3306)/common_services?parseTime=true&charset=utf8mb4&loc=Local'
export ID_REGION_ID=1
export ID_DATACENTER_CODE=local-a
go run ./cmd/id-server
```

## API 示例

单个发号：

```bash
curl -X POST http://localhost:8090/api/v1/ids/next \
  -H 'Content-Type: application/json' \
  -d '{"namespace":"trade.order","caller":"order-service","request_id":"req-1"}'
```

批量发号：

```bash
curl -X POST http://localhost:8090/api/v1/ids/batch \
  -H 'Content-Type: application/json' \
  -d '{"namespace":"product.sku","caller":"product-service","request_id":"req-2","count":5}'
```

查看 namespace：

```bash
curl http://localhost:8090/api/v1/namespaces
```

健康检查与指标：

```bash
curl http://localhost:8090/healthz
curl http://localhost:8090/readyz
curl http://localhost:8090/metrics
```

## 生产语义

ID 服务允许跳号，不允许复用。业务表仍然需要唯一索引兜底。对外单号不直接暴露内部递增 ID。

Snowflake 的 `worker_id` 在 MySQL 模式下通过 `id_worker` 租约获取，租约丢失后 `/readyz` 返回不可用，发号链路不应静默继续生成交易 ID。

Segment 的号段申请通过 `id_segment.version` 乐观锁推进，适合商品、营销、财务这类容量可规划的主数据 ID。
