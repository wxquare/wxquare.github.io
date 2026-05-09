# 八层商品交易模型 Demo

这个 Demo 展示如何把“异构数字商品”落成一套可执行的工程结构。它不是生产级完整系统，而是把书中讨论的八层模型映射到代码：

```text
Product Definition
  → Resource
  → Offer / Rate Plan
  → Availability
  → Input Schema
  → Booking / Lock
  → Fulfillment Contract
  → Refund Rule
```

## 为什么需要八层模型

数字商品的“商品”并不总是一个固定 SKU：

| 品类 | 难点 |
|---|---|
| Topup | 商品像一个固定面额，但履约依赖充值账户和供应商结果 |
| Gift Card | 商品是面额，但库存来自本地券码池 |
| Flight | 商品不是固定航班 SKU，而是一次实时查询后的报价和座位资源 |
| Hotel | 商品静态信息可沉淀，但价格和房态取决于日期、间夜、入住人和 Rate Plan |

八层模型的目标不是把所有品类强行变成一张宽表，而是把“交易前必须理解的上下文”拆成稳定的几个层次。

## 代码结构

```text
internal/domain/
├── category_capability.go      # 品类能力矩阵
├── runtime_context.go          # ProductRuntimeContext 八层上下文
├── category_strategy.go        # CategoryStrategy 策略接口
└── strategy/
    ├── topup_strategy.go
    ├── giftcard_strategy.go
    ├── flight_strategy.go
    └── hotel_strategy.go

internal/application/
├── dto/runtime_context_dto.go
├── dto/category_action_dto.go
├── service/runtime_context_service.go
└── service/category_action_service.go

internal/infrastructure/persistence/
└── capability_repository.go    # 内存示例数据

internal/interfaces/http/
├── runtime_context_handler.go  # 八层上下文查询接口
└── category_action_handler.go  # 品类动作/垂直搜索接口
```

## 示例 API

启动服务：

```bash
go run cmd/main.go
```

查询八层上下文：

```bash
curl "http://localhost:8080/api/v1/products/runtime-context?sku_id=10001&category_id=10102&scene=detail"
curl "http://localhost:8080/api/v1/products/runtime-context?sku_id=10002&category_id=30105&scene=detail"
curl "http://localhost:8080/api/v1/products/runtime-context?sku_id=40001&category_id=40102&scene=checkout"
curl "http://localhost:8080/api/v1/products/runtime-context?sku_id=40002&category_id=40104&scene=checkout"
```

品类动作接口和垂直实时供给接口：

```bash
curl -X POST "http://localhost:8080/api/v1/topup/validate-account" \
  -H "Content-Type: application/json" \
  -d '{"sku_id":10001,"mobile_number":"13800138000"}'

curl "http://localhost:8080/api/v1/travel/flights/search?from=SHA&to=BJS&date=2026-05-01&adult=1"
```

这两个接口体现了一个关键原则：`ProductRuntimeContext` 是统一的交易前上下文模型，但 C 端接口不需要强行统一成一个万能接口。

| 场景 | 接口 | 使用八层模型的方式 |
|---|---|---|
| Topup 详情 | `GET /api/v1/products/runtime-context` | 返回固定面额、无限库存、手机号输入、充值履约规则 |
| Topup 账号校验 | `POST /api/v1/topup/validate-account` | 复用商品运行时数据和输入规则，对用户账号做独立校验 |
| Flight 搜索 | `GET /api/v1/travel/flights/search` | 使用航线/航司等静态资源，返回实时报价、实时库存和占座要求 |
| Flight 创单前校验 | `GET /api/v1/products/runtime-context` 或报价详情接口 | 根据 offer token 重建关键上下文，重新校验价格、库存和占座规则 |

样例数据：

| SKU | Category | 品类 | 策略 |
|---:|---:|---|---|
| 10001 | 10102 | Mobile Topup | 固定面额 + 无限库存 + 支付后充值 |
| 10002 | 30105 | Gift Card | 固定面额 + 券码池 + 支付后发码 |
| 40001 | 40102 | Flight | 静态航线 + 实时报价库存 + 创单前占座 |
| 40002 | 40104 | Hotel | 酒店房型 + Rate Plan + 动态房态房价 |

## 生产表映射

当前 Demo 使用内存 Repository。生产环境通常会把这些数据拆到 MySQL 表中：

| Demo 模型 | 生产表建议 | 说明 |
|---|---|---|
| `CategoryCapability` | `category_capability_tab` | 定义品类能力矩阵 |
| `ProductRuntimeSource` | `product_spu_tab`、`product_sku_tab` | 商品主数据 |
| `ResourceContext` | `hotel_tab`、`route_tab`、`brand_tab`、`merchant_tab` | 稳定资源实体 |
| `OfferContext` | `product_offer_tab`、`rate_plan_tab` | 固定价、报价规则、Rate Plan |
| `AvailabilityContext` | `product_stock_tab`、`voucher_code_pool_tab`、供应商查询结果缓存 | 可售状态与库存来源 |
| `InputSchema` | `input_schema_tab`、`input_field_tab` | 动态表单和校验规则 |
| `FulfillmentContract` | `fulfillment_rule_tab`、`supplier_product_mapping_tab` | 履约参数和供应商映射 |
| `RefundRule` | `refund_rule_tab` | 售后规则 |

## 设计要点

1. `CategoryCapability` 决定一个品类需要哪些能力，而不是把差异写死在主流程。
2. `CategoryStrategy` 隔离品类差异，主流程只负责组装上下文。
3. `ProductRuntimeContext` 是交易前读模型，适合详情页、结算页、创单前校验等场景。
4. Flight 这种实时资源不沉淀完整 SKU，只沉淀航线、城市、航司等静态资源，报价和库存实时获取。
5. Hotel 静态资源和动态房态房价分离，Rate Plan 表达早餐、取消规则、结算方式等商业条件。
