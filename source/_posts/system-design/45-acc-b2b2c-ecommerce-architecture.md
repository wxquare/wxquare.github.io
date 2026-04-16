---
title: B2B2C电商平台系统架构设计（200人团队、日订单200万级）
date: 2026-04-14
categories:
  - system-design
tags:
  - architecture
  - e-commerce
  - microservices
  - high-availability
  - distributed-systems
---

# B2B2C电商平台系统架构设计

> **项目背景**：设计一个中大型B2B2C电商平台，主要连接外部供应商（机票、酒店、充值、电影票等虚拟数字商品），同时支持自营商品（优惠券、礼品卡）。平台规模：200+人团队，日订单200万，大促峰值1000万订单/天。

---

## 一、业务背景与架构目标

### 1.1 业务模式

**核心业务**：
- **B2B2C聚合模式**：连接50+外部供应商，聚合机票、酒店、账单充值、电影票等虚拟商品
- **自营模式**：自营优惠券（e-voucher）、线下券、礼品卡等
- **无物流场景**：全部为虚拟数字商品，无需物流配送

**关键特征**：
- 供应商接口高度碎片化（实时查询 + 定时同步 + 推送混合）
- 核心品类（机票/酒店）零超卖容忍
- 长尾品类（充值/礼品卡）可事后补偿

### 1.2 品类业务模型差异

不同品类的业务模型存在显著差异，直接影响架构设计决策：

#### （1）机票（Flight）

```
业务特点：
• 库存模型：实时库存（供应商侧），强依赖供应商实时查询
• 价格模型：动态定价，实时波动（可能秒级变化）
• SKU复杂度：极高（航司+航班号+舱位+日期+...组合）
• 库存单位：座位数量（不可超卖）
• 扣减时机：下单即扣（预占）→ 支付确认 → 出票
• 履约流程：下单 → 支付 → 出票（调用GDS/供应商API）→ 发送电子票

架构影响：
✓ 必须支持实时库存查询（高频调用供应商API）
✓ 价格快照必须精确到秒级，防止价格变动纠纷
✓ 超卖零容忍 → 下单前二次确认库存
✓ 供应商故障需快速切换到备用供应商
✓ 订单状态复杂（待出票、出票中、出票失败、已出票）

技术要点：
• Redis缓存TTL：5分钟（库存）、10分钟（价格）
• 供应商调用超时：800ms（实时查询）
• 熔断阈值：错误率>50%，熔断10秒
```

#### （2）酒店（Hotel）

```
业务特点：
• 库存模型：房间数量（按日期维度管理）
• 价格模型：日历房价（每个日期不同价格）
• SKU复杂度：高（酒店ID+房型+日期范围+早餐+...）
• 库存单位：房间数/间夜数
• 扣减时机：下单预占 → 支付确认 → 供应商确认
• 履约流程：下单 → 支付 → 提交供应商 → 确认单 → 入住凭证

架构影响：
✓ 支持日期范围查询（check-in到check-out）
✓ 日历价格存储（每个日期一条记录）
✓ 库存按日期维度管理（某天无房不影响其他日期）
✓ 支持"担保"模式（先占房，入住时结算）
✓ 需处理"确认单延迟"（供应商异步确认）

技术要点：
• 价格存储：时间序列数据库或宽表（date维度）
• 库存粒度：SKU_ID + Date（复合键）
• 缓存策略：热门酒店30分钟，长尾酒店1小时
• 供应商确认：异步轮询（每30秒查询一次状态）
```

#### （3）充值（Top-up / Recharge）

```
业务特点：
• 库存模型：无限库存（供应商侧无限制）
• 价格模型：固定面额（10元、50元、100元）
• SKU复杂度：低（运营商+面额）
• 库存单位：无限
• 扣减时机：支付后
• 履约流程：下单 → 支付 → 调用供应商API → 充值成功/失败

架构影响：
✓ 无需库存管理（库存类型=无限）
✓ 价格简单（基础价+平台服务费）
✓ 超卖可接受（事后补偿）
✓ 供应商调用简单（同步API，3秒内返回）
✓ 失败重试友好（幂等性强）

技术要点：
• 库存管理：不需要预占，直接下单
• 价格缓存：1小时（价格稳定）
• 供应商调用：同步调用，3秒超时
• 重试策略：3次重试，指数退避（1s, 2s, 4s）
```

#### （4）账单缴费（Bill Payment）

```
业务特点：
• 库存模型：无库存概念（代收代付）
• 价格模型：查询实时账单金额
• SKU复杂度：低（账单类型+账号）
• 库存单位：无
• 扣减时机：支付后
• 履约流程：查询账单 → 下单 → 支付 → 缴费成功 → 回执

架构影响：
✓ 需要"查账单"接口（调用供应商）
✓ 金额动态（每次查询不同）
✓ 幂等性要求极高（避免重复缴费）
✓ 对账要求严格（需与供应商流水对账）

技术要点：
• 查账单缓存：5分钟（避免频繁查询）
• 幂等Token：前端生成，5分钟有效
• 对账频率：每小时一次
• 供应商调用：同步，5秒超时
```

#### （5）电影票（Movie Ticket）

```
业务特点：
• 库存模型：实时库存（座位级别）
• 价格模型：动态定价（场次+座位+时段）
• SKU复杂度：极高（影院+影片+场次+座位号）
• 库存单位：座位（精确到排号座号）
• 扣减时机：选座即锁定（15分钟）→ 支付确认
• 履约流程：选座 → 锁座（15min）→ 支付 → 出票码

架构影响：
✓ 座位锁定机制（15分钟倒计时）
✓ 实时库存（座位图需秒级更新）
✓ 超卖零容忍（用户体验极差）
✓ 高并发场景（热门场次抢票）
✓ 座位状态复杂（可售、锁定、已售、维修）

技术要点：
• 库存粒度：SKU_ID + SeatNo（精确到座位）
• 锁座机制：Redis SETNX + 15分钟TTL
• 座位图缓存：实时推送（WebSocket）
• 热门场次限流：令牌桶算法，QPS=500
```

#### （6）Deal/线下优惠券（Voucher）

```
业务特点：
• 库存模型：固定库存（券码池）
• 价格模型：固定折扣价
• SKU复杂度：中（商户+门店+商品+...）
• 库存单位：券码（一券一码）
• 扣减时机：支付后
• 履约流程：下单 → 支付 → 发券码 → 到店核销

架构影响：
✓ 券码池管理（预生成10万个券码）
✓ 券码发放（支付后随机分配）
✓ 核销系统（商户扫码核销）
✓ 过期管理（券有效期7天-180天）
✓ 退款逻辑（未核销可退，已核销不可退）

技术要点：
• 券码存储：Redis Set（未使用券码池）
• 发券逻辑：SPOP（原子弹出一个券码）
• 有效期管理：ZSet按过期时间排序
• 核销接口：幂等性（同一券码只能核销一次）
```

#### （7）礼品卡（Gift Card）

```
业务特点：
• 库存模型：无限库存（虚拟卡）
• 价格模型：固定面额 or 自定义金额
• SKU复杂度：低（面额）
• 库存单位：卡号
• 扣减时机：支付后
• 履约流程：下单 → 支付 → 生成卡号+卡密 → 发送

架构影响：
✓ 卡号生成算法（保证唯一性）
✓ 余额管理（卡内余额扣减）
✓ 多次使用（支持部分消费）
✓ 转赠功能（卡可以转给他人）
✓ 对账复杂（发卡、消费、退款流水）

技术要点：
• 卡号生成：雪花算法 + 校验位
• 余额存储：Redis Hash（cardNo -> balance）
• 消费记录：MySQL + ES（双写）
• 并发控制：乐观锁（版本号）
```

### 1.3 品类差异对架构的影响

| 维度 | 机票/酒店（核心品类） | 充值/账单（长尾品类） | Deal/礼品卡（自营） |
|-----|---------------------|---------------------|-------------------|
| **库存管理** | 实时同步，强一致 | 无限库存，无需管理 | 券码池，异步补充 |
| **价格策略** | 动态定价，实时变化 | 固定面额，稳定 | 固定折扣，活动价 |
| **超卖容忍度** | 零容忍（P0故障） | 可补偿（P2故障） | 低容忍（P1故障） |
| **供应商依赖** | 强依赖，需实时调用 | 弱依赖，异步批量 | 无依赖（自营） |
| **缓存TTL** | 5-10分钟 | 30-60分钟 | 1-24小时 |
| **熔断阈值** | 50%错误率 | 70%错误率 | 不需要熔断 |
| **对账频率** | 每5分钟 | 每小时 | 每天 |
| **履约复杂度** | 高（多状态机） | 低（同步返回） | 中（券码+核销） |

**架构设计启示**：
1. **不能一刀切**：不同品类需要不同的库存策略、缓存策略、对账策略
2. **策略模式**：通过策略模式实现品类差异化逻辑（见后续"库存服务设计"）
3. **优先级分级**：核心品类（机票/酒店）优先保障，长尾品类可降级
4. **供应商分级**：P0供应商（机票）熔断阈值更严格，P2供应商（充值）更宽松
5. **监控分级**：核心品类错误率>0.1%告警，长尾品类错误率>1%告警

### 1.4 规模指标

| 指标 | 日常 | 大促峰值 | 备注 |
|-----|------|---------|------|
| 日订单量 | 200万 | 1000万 | 5倍峰值 |
| 日活用户 | 500万 | 2000万 | 4倍峰值 |
| QPS峰值 | 5万 | 25万 | API Gateway总QPS |
| 商品SKU | 500万 | - | 包含供应商商品 |
| 团队规模 | 200人 | - | 研发+测试+运维 |

### 1.5 架构目标（优先级排序）

1. **高可用性**（P0）：核心服务SLA ≥ 99.95%
2. **数据一致性**（P0）：订单/库存/资金数据强一致
3. **供应商容错**（P1）：单个供应商故障不影响平台
4. **高性能**（P1）：P99延迟 < 1秒
5. **弹性扩展**（P2）：支持5-10倍弹性扩容

---

## 二、整体架构设计

### 2.1 七层架构模型

```
┌────────────────────────────────────────────────────────────────┐
│  L1: 用户层（User Layer）                                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  • Web端：React/Vue                                       │  │
│  │  • 移动端：iOS/Android原生 + RN/Flutter                   │  │
│  │  • 小程序：微信/支付宝小程序                              │  │
│  └──────────────────────────────────────────────────────────┘  │
├────────────────────────────────────────────────────────────────┤
│  L2: 网关层（Gateway Layer）                                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  • API Gateway (APISIX)：统一入口、限流、鉴权            │  │
│  │  • BFF (Backend For Frontend)：端特定逻辑聚合            │  │
│  └──────────────────────────────────────────────────────────┘  │
├────────────────────────────────────────────────────────────────┤
│  L3: 业务服务层（Business Service Layer）                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  核心域：                                                 │  │
│  │  • Order Service (订单)                                   │  │
│  │  • Payment Service (支付)                                 │  │
│  │  • Checkout Service (结算)                                │  │
│  │  • Inventory Service (库存)                               │  │
│  │                                                           │  │
│  │  支撑域：                                                 │  │
│  │  • Product Center (商品中心)                              │  │
│  │  • Pricing Service (计价引擎)                             │  │
│  │  • Marketing Service (营销)                               │  │
│  │  • Search Service (搜索)                                  │  │
│  │  • Aggregation Service (聚合服务，编排层)                │  │
│  │  • Cart Service (购物车)                                  │  │
│  │  • User Service (用户)                                    │  │
│  │  • Listing Service (商品上架)                             │  │
│  │                                                           │  │
│  │  供应商域：                                               │  │
│  │  • Supplier Gateway (供应商网关)                          │  │
│  │  • Supplier Sync (供应商同步)                             │  │
│  └──────────────────────────────────────────────────────────┘  │
├────────────────────────────────────────────────────────────────┤
│  L4: 供应商网关层（Supplier Gateway Layer）                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  • 供应商适配器（Plugin架构，每个供应商独立插件）         │  │
│  │  • 协议转换（HTTP/SOAP/gRPC统一适配）                     │  │
│  │  • 熔断降级（Hystrix）                                    │  │
│  │  • 限流重试（智能退避）                                   │  │
│  └──────────────────────────────────────────────────────────┘  │
├────────────────────────────────────────────────────────────────┤
│  L5: 中间件层（Middleware Layer）                              │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  • MySQL (分库分表、主从复制)                             │  │
│  │  • Redis Cluster (三级缓存：本地+Redis+DB)                │  │
│  │  • Kafka (事件总线、3副本、24h保留)                       │  │
│  │  • Elasticsearch (商品搜索、5分片×2副本)                  │  │
│  └──────────────────────────────────────────────────────────┘  │
├────────────────────────────────────────────────────────────────┤
│  L6: 基础设施层（Infrastructure Layer）                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  • Service Mesh (Istio/Linkerd)                           │  │
│  │  • 配置中心 (Nacos/Apollo)                                │  │
│  │  • 注册中心 (Consul/Etcd)                                 │  │
│  │  • 链路追踪 (Jaeger/Skywalking)                           │  │
│  │  • 监控告警 (Prometheus + Grafana)                        │  │
│  │  • 日志收集 (ELK Stack)                                   │  │
│  └──────────────────────────────────────────────────────────┘  │
├────────────────────────────────────────────────────────────────┤
│  L7: 部署层（Deployment Layer）                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  • Kubernetes (容器编排、HPA弹性伸缩)                     │  │
│  │  • 同城双活部署 (IDC-A + IDC-B)                           │  │
│  │  • DNS智能解析 (GeoDNS)                                   │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────┘
```

### 2.2 微服务拆分（19个核心服务）

| 服务名称 | 职责 | QPS（常态） | QPS（大促） | 副本数（常/大促） |
|---------|------|------------|------------|------------------|
| **api-gateway** | 统一入口、鉴权、限流 | 50000 | 250000 | 4 / 12 |
| **order-service** | 订单管理、状态机 | 3000 | 15000 | 10 / 30 |
| **payment-service** | 支付集成、回调处理 | 3000 | 15000 | 10 / 30 |
| **checkout-service** | 结算、试算、确认下单 | 2000 | 10000 | 8 / 24 |
| **inventory-service** | 库存管理、预占、扣减 | 5000 | 25000 | 8 / 24 |
| **product-center** | 商品主数据、SPU/SKU | 8000 | 40000 | 6 / 18 |
| **pricing-service** | 四层计价（基础价+促销+费用+券） | 4000 | 20000 | 8 / 24 |
| **marketing-service** | 优惠券、活动、规则引擎 | 2000 | 10000 | 6 / 18 |
| **search-service** | ES查询、索引管理 | 6000 | 30000 | 6 / 18 |
| **aggregation-service** | 数据聚合、编排多服务调用（搜索、详情等） | 6000 | 30000 | 6 / 18 |
| **cart-service** | 购物车增删改查 | 4000 | 20000 | 6 / 18 |
| **supplier-gateway** | 供应商调用适配、熔断、重试 | 10000 | 50000 | 10 / 30 |
| **supplier-sync** | 供应商数据同步（定时任务） | - | - | 4 / 4 |
| **listing-service** | 商品上架、审核、状态机 | 500 | 2000 | 4 / 12 |
| **user-service** | 用户信息、会员等级 | 2000 | 10000 | 4 / 12 |
| **notification-service** | 消息通知（短信/邮件/推送） | 2000 | 10000 | 4 / 12 |
| **analytics-service** | 数据上报、埋点 | 5000 | 25000 | 4 / 12 |
| **admin-service** | 运营后台管理 | 200 | 200 | 2 / 2 |
| **task-scheduler** | 定时任务调度 | - | - | 2 / 2 |

---

## 三、核心服务设计

### 3.1 聚合服务（Aggregation Service）

#### 服务定位

**职责**：通用聚合服务，编排多服务调用，按依赖关系顺序聚合数据。

**核心场景**：

##### 1. 商品列表场景（Item/SPU维度）

| 查询方式 | 查询维度 | ES查询字段 | 返回粒度 | 典型场景 |
|---------|---------|-----------|---------|---------|
| **关键字搜索** | keyword | title, description, tags | Item列表（SPU） | 用户输入"无线耳机" |
| **分类浏览** | category_id | category_id | Item列表（SPU） | 点击"数码配件"分类 |
| **筛选查询** | brand, price_range, attrs | brand, price, attributes | Item列表（SPU） | 筛选"苹果品牌+500-1000元" |
| **推荐列表** | user_id, item_id | 推荐算法 | Item列表（SPU） | "猜你喜欢"、"相关推荐" |

**编排流程**：ES查询（获取item_ids） → Product Center（商品基础信息+base_price） → Inventory（库存状态） → Marketing（营销活动） → Pricing（计算最终价格）

**数据特点**：
- ✅ 批量查询：一次返回20-50个商品
- ✅ 性能优先：支持降级（营销/库存可降级）
- ✅ 缓存友好：列表结果可缓存10分钟

##### 2. 商品详情场景（SKU维度）

| 查询方式 | 查询维度 | 返回粒度 | 典型场景 |
|---------|---------|---------|---------|
| **商品详情页** | item_id | Item信息 + 所有SKU详情 | 用户点击商品进入详情页 |
| **SKU详情** | sku_id | 单个SKU详细信息 | 用户选择规格后查询库存/价格 |

**编排流程**：Product Center（商品+SKU详情） → Inventory（SKU库存） → Marketing（SKU级营销） → Pricing（SKU价格） → Review（评价） → Recommendation（相关推荐）

**数据特点**：
- ✅ 细粒度查询：返回SKU级别的库存、价格、属性
- ✅ 完整性优先：不支持降级（库存/价格必须准确）
- ✅ 实时性强：缓存TTL较短（1-5分钟）

##### 3. 其他查询场景（扩展）

- **订单详情聚合**：订单 → 商品 → 物流 → 售后
- **用户中心聚合**：用户 → 订单 → 优惠券 → 积分
- **购物车聚合**：购物车 → 商品 → 库存 → 价格 → 营销

**与其他服务的区别**：
- **vs Search Service**：Search Service只负责ES查询，不包含业务逻辑
- **vs BFF**：不区分端（Web/App共用），专注于数据聚合编排
- **vs Checkout Service**：Checkout是交易编排，Aggregation是查询编排
- **vs Product Center**：Product Center是数据源，Aggregation是编排层

#### 目录结构

```
aggregation-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── application/
│   │   ├── dto/
│   │   │   ├── search_request.go         # 搜索/列表请求
│   │   │   ├── search_response.go        # 搜索/列表响应
│   │   │   ├── detail_request.go         # 详情请求
│   │   │   └── detail_response.go        # 详情响应
│   │   └── service/
│   │       ├── search_orchestrator.go    # 列表场景编排器（Item/SPU维度）
│   │       ├── detail_orchestrator.go    # 详情场景编排器（SKU维度）
│   │       └── cart_orchestrator.go      # 购物车场景编排器（扩展）
│   ├── infrastructure/
│   │   ├── rpc/
│   │   │   ├── search_client.go          # Search Service客户端
│   │   │   ├── product_client.go         # Product Center客户端
│   │   │   ├── inventory_client.go       # Inventory客户端
│   │   │   ├── marketing_client.go       # Marketing客户端
│   │   │   ├── pricing_client.go         # Pricing客户端
│   │   │   ├── review_client.go          # Review Service客户端
│   │   │   └── recommendation_client.go  # Recommendation客户端
│   │   ├── cache/
│   │   │   └── redis_cache.go            # Redis缓存
│   │   ├── circuitbreaker/
│   │   │   └── breaker.go                # 熔断器
│   │   └── event/
│   │       └── kafka_publisher.go        # 用户行为事件
│   └── interfaces/
│       ├── http/
│       │   ├── search_handler.go         # 搜索/列表接口
│       │   └── detail_handler.go         # 详情接口
│       └── grpc/
│           ├── search_handler.go
│           └── detail_handler.go
├── config/
│   └── config.yaml
└── go.mod
```

#### 核心编排逻辑

```go
// SearchOrchestrator 搜索编排器
type SearchOrchestrator struct {
    searchClient    rpc.SearchClient
    productClient   rpc.ProductClient
    inventoryClient rpc.InventoryClient
    marketingClient rpc.MarketingClient
    pricingClient   rpc.PricingClient
    cache           cache.Cache
}

// Search 搜索商品（编排多服务调用）
func (o *SearchOrchestrator) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
    // Step 1: 缓存检查
    cacheKey := o.buildCacheKey(req)
    if cached, err := o.cache.Get(ctx, cacheKey); err == nil {
        return cached, nil
    }
    
    // Step 2: 调用Search Service查询ES，获取sku_ids
    searchResult, err := o.searchClient.SearchES(ctx, &SearchESRequest{
        Keyword: req.Keyword,
        Page:    req.Page,
        Size:    req.Size,
    })
    if err != nil {
        return nil, fmt.Errorf("search ES failed: %w", err)
    }
    
    skuIDs := searchResult.SkuIDs  // [1001, 1002, ..., 1020]
    
    // Step 3: 并发调用Product + Inventory（无依赖关系）
    var (
        products  []*Product
        stocks    map[int64]*StockInfo
        wg        sync.WaitGroup
        errChan   = make(chan error, 2)
    )
    
    wg.Add(2)
    
    // 并发调用1：Product Center
    go func() {
        defer wg.Done()
        var err error
        products, err = o.productClient.BatchGetProducts(ctx, skuIDs)
        if err != nil {
            errChan <- fmt.Errorf("get products failed: %w", err)
        }
    }()
    
    // 并发调用2：Inventory Service
    go func() {
        defer wg.Done()
        var err error
        stocks, err = o.inventoryClient.BatchCheckStock(ctx, skuIDs)
        if err != nil {
            errChan <- fmt.Errorf("check stock failed: %w", err)
        }
    }()
    
    wg.Wait()
    close(errChan)
    
    // 检查错误（Product是核心依赖，必须成功）
    for err := range errChan {
        if err != nil && strings.Contains(err.Error(), "get products") {
            return nil, err  // Product失败直接返回
        }
        // Inventory失败降级处理（隐藏库存）
    }
    
    // 构建商品基础价格map（来自Product Center）
    basePriceMap := make(map[int64]float64)
    for _, p := range products {
        basePriceMap[p.SkuID] = p.BasePrice
    }
    
    // Step 4: 调用Marketing Service获取营销信息
    promos, err := o.marketingClient.BatchGetPromotions(ctx, &PromotionRequest{
        SkuIDs: skuIDs,
        UserID: req.UserID,  // 个性化营销
    })
    if err != nil {
        // Marketing失败降级：无促销信息
        promos = make(map[int64]*PromotionInfo)
    }
    
    // Step 5: 调用Pricing Service计算最终价格
    // 输入：base_price + 营销信息
    priceItems := make([]*PriceCalculateItem, 0, len(skuIDs))
    for _, skuID := range skuIDs {
        priceItems = append(priceItems, &PriceCalculateItem{
            SkuID:      skuID,
            BasePrice:  basePriceMap[skuID],  // 来自Product Center
            PromoInfo:  promos[skuID],        // 来自Marketing Service
            Quantity:   1,
        })
    }
    
    prices, err := o.pricingClient.BatchCalculatePrice(ctx, priceItems)
    if err != nil {
        // Pricing失败降级：只展示base_price
        prices = o.buildFallbackPrices(basePriceMap)
    }
    
    // Step 6: 数据聚合
    items := o.aggregateResults(products, stocks, promos, prices)
    
    // Step 7: 写入缓存（异步）
    go func() {
        o.cache.Set(context.Background(), cacheKey, items, 10*time.Minute)
    }()
    
    // Step 8: 发布搜索事件（异步）
    go func() {
        o.publishSearchEvent(context.Background(), req, searchResult)
    }()
    
    return &SearchResponse{
        Total:   searchResult.Total,
        Items:   items,
        Filters: searchResult.Filters,
    }, nil
}
```

```go
// DetailOrchestrator 详情编排器（SKU维度）
type DetailOrchestrator struct {
    productClient        rpc.ProductClient
    inventoryClient      rpc.InventoryClient
    marketingClient      rpc.MarketingClient
    pricingClient        rpc.PricingClient
    reviewClient         rpc.ReviewClient
    recommendationClient rpc.RecommendationClient
    cache                cache.Cache
}

// GetItemDetail 获取商品详情（Item + 所有SKU信息）
func (o *DetailOrchestrator) GetItemDetail(ctx context.Context, req *DetailRequest) (*DetailResponse, error) {
    // Step 1: 缓存检查（详情页缓存TTL较短：1-5分钟）
    cacheKey := fmt.Sprintf("item_detail:%d:user:%d", req.ItemID, req.UserID)
    if cached, err := o.cache.Get(ctx, cacheKey); err == nil {
        return cached, nil
    }
    
    // Step 2: 获取商品基础信息（Item + 所有SKU）
    itemDetail, err := o.productClient.GetItemDetail(ctx, req.ItemID)
    if err != nil {
        return nil, fmt.Errorf("get item detail failed: %w", err)
    }
    
    skuIDs := itemDetail.SkuIDs  // [1001, 1002, 1003] - 所有规格的SKU
    
    // Step 3: 并发调用多个服务（提升性能）
    var (
        stocks    map[int64]*StockInfo
        promos    map[int64]*PromoInfo
        reviews   *ReviewSummary
        recommend []*RecommendItem
        wg        sync.WaitGroup
        mu        sync.Mutex
        errs      []error
    )
    
    wg.Add(4)
    
    // 并发调用1：Inventory Service（SKU库存）
    go func() {
        defer wg.Done()
        var err error
        stocks, err = o.inventoryClient.BatchCheckStock(ctx, skuIDs)
        if err != nil {
            mu.Lock()
            errs = append(errs, fmt.Errorf("inventory failed: %w", err))
            mu.Unlock()
        }
    }()
    
    // 并发调用2：Marketing Service（SKU营销）
    go func() {
        defer wg.Done()
        var err error
        promos, err = o.marketingClient.BatchGetPromotions(ctx, skuIDs, req.UserID)
        if err != nil {
            mu.Lock()
            errs = append(errs, fmt.Errorf("marketing failed: %w", err))
            mu.Unlock()
        }
    }()
    
    // 并发调用3：Review Service（评价汇总）
    go func() {
        defer wg.Done()
        var err error
        reviews, err = o.reviewClient.GetReviewSummary(ctx, req.ItemID)
        if err != nil {
            mu.Lock()
            errs = append(errs, fmt.Errorf("review failed: %w", err))
            mu.Unlock()
        }
    }()
    
    // 并发调用4：Recommendation Service（相关推荐）
    go func() {
        defer wg.Done()
        var err error
        recommend, err = o.recommendationClient.GetRelatedItems(ctx, req.ItemID, 10)
        if err != nil {
            mu.Lock()
            errs = append(errs, fmt.Errorf("recommendation failed: %w", err))
            mu.Unlock()
        }
    }()
    
    wg.Wait()
    
    // 详情页关键数据失败不降级，直接返回错误
    if len(errs) > 0 {
        for _, err := range errs {
            if strings.Contains(err.Error(), "inventory failed") {
                return nil, err  // 库存是详情页核心数据，不可降级
            }
        }
    }
    
    // Step 4: 调用Pricing Service计算每个SKU的价格
    prices, err := o.pricingClient.BatchCalculatePrice(ctx, &PriceRequest{
        Items: buildPriceItems(itemDetail, promos),
    })
    if err != nil {
        return nil, fmt.Errorf("calculate price failed: %w", err)
    }
    
    // Step 5: 数据聚合（包含营销信息处理）
    skuDetails := make([]*SkuDetail, 0, len(skuIDs))
    for _, skuID := range skuIDs {
        skuDetails = append(skuDetails, &SkuDetail{
            SkuID:         skuID,
            Attributes:    itemDetail.SkuAttributes[skuID],  // 颜色、尺码等
            Stock:         stocks[skuID],                     // 库存状态
            Price:         prices[skuID],                     // 最终价格
            Promotion:     promos[skuID],                     // 营销活动
        })
    }
    
    // 构建营销信息（吸引用户多买）
    promotions := buildPromotionDetails(promos, itemDetail)
    // promotions包含：
    // - active_promotions: 多买优惠、组合优惠、品类促销
    // - coupons: 可用优惠券列表
    // - saving_tips: "再买1件，可省XXX元"
    
    resp := &DetailResponse{
        Item:           itemDetail.Item,
        SkuDetails:     skuDetails,
        Promotions:     promotions,           // 营销信息（促进多买）
        ReviewSummary:  reviews,
        Recommendation: recommend,
    }
    
    // Step 6: 写入缓存（TTL: 1-5分钟）
    _ = o.cache.Set(ctx, cacheKey, resp, 5*time.Minute)
    
    // Step 7: 发布用户行为事件（异步）
    go o.publishViewEvent(ctx, req.ItemID, req.UserID)
    
    return resp, nil
}
```

#### 列表 vs 详情场景对比

| 维度 | 搜索/列表场景 | 商品详情场景 |
|-----|------------|------------|
| **查询粒度** | Item/SPU（商品级别） | SKU（规格级别） |
| **返回数量** | 20-50个商品 | 1个商品 + N个SKU |
| **ES查询** | 需要（关键字/分类/筛选） | 不需要（直接通过item_id查询） |
| **库存查询** | 可降级（隐藏库存） | 不可降级（核心数据） |
| **营销信息** | 简单（单品折扣） | 详细（多买优惠、组合优惠、省钱提示） |
| **营销目的** | 吸引点击 | 吸引多买、提升客单价 |
| **价格计算** | 可降级（base_price） | 不可降级（必须准确） |
| **缓存TTL** | 10分钟（列表变化慢） | 1-5分钟（价格/库存变化快） |
| **降级策略** | 支持多级降级 | 关键数据不降级 |
| **用户行为** | 浏览、点击 | 详细查看、加购 |

#### 调用链路可视化

**场景1：搜索/列表场景（SearchOrchestrator）**

```
Aggregation Service编排流程：

Stage 1: ES查询（独立）
    ↓ item_ids / sku_ids
Stage 2: 并发调用（无依赖）
    ├─ Product Center (base_price, info)
    └─ Inventory Service (stock_info)
    ↓
Stage 3: Marketing Service（依赖sku_ids）
    ↓ promo_info
Stage 4: Pricing Service（依赖base_price + promo_info）
    ↓ final_price
Stage 5: 聚合返回

总耗时：50 + 50 + 70 + 100 + 20 = 290ms
缓存命中：5ms（80%场景）
降级场景：50 + 50 = 100ms（只返回商品基础信息）
```

**场景2：商品详情场景（DetailOrchestrator）**

```
Aggregation Service编排流程：

Stage 1: Product Center查询（获取Item + 所有SKU）
    ↓ item_detail + sku_ids
Stage 2: 并发调用（无依赖，4个服务）
    ├─ Inventory Service (SKU库存)
    ├─ Marketing Service (SKU营销)
    ├─ Review Service (评价汇总)
    └─ Recommendation Service (相关推荐)
    ↓
Stage 3: Pricing Service（依赖base_price + promo_info）
    ↓ final_price (每个SKU)
Stage 4: 聚合返回

总耗时：50 + max(30, 70, 40, 50) + 100 = 220ms
缓存命中：5ms（50%场景）
注：详情页关键数据（库存/价格）不降级
```

#### 降级策略矩阵

**搜索/列表场景（SearchOrchestrator）**

| 服务 | 是否核心依赖 | 失败处理 | 对用户的影响 |
|-----|------------|---------|-------------|
| **Search Service** | 是 | 返回错误 | 搜索不可用 |
| **Product Center** | 是 | 返回错误 | 搜索不可用 |
| **Inventory Service** | 否 | 降级（隐藏库存） | 不显示库存状态 |
| **Marketing Service** | 否 | 降级（无促销） | 只展示基础价 |
| **Pricing Service** | 否 | 降级（base_price） | 展示基础价，无促销价 |

**商品详情场景（DetailOrchestrator）**

| 服务 | 是否核心依赖 | 失败处理 | 对用户的影响 |
|-----|------------|---------|-------------|
| **Product Center** | 是 | 返回错误 | 详情页不可用 |
| **Inventory Service** | 是 | 返回错误 | 无法下单（库存是关键数据） |
| **Pricing Service** | 是 | 返回错误 | 无法下单（价格是关键数据） |
| **Marketing Service** | 是 | 返回错误 | 价格计算依赖营销规则 |
| **Review Service** | 否 | 降级（隐藏评价） | 无评价展示 |
| **Recommendation Service** | 否 | 降级（无推荐） | 无相关推荐

---

### 3.2 库存服务（Inventory Service）

#### 目录结构

```
inventory-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── domain/
│   │   ├── model/
│   │   │   ├── inventory.go          # 库存聚合根
│   │   │   ├── stock_unit.go         # 库存单元（券码/数量/时间）
│   │   │   └── reservation.go        # 预占记录
│   │   ├── value_object/
│   │   │   ├── management_type.go    # 管理类型（自营/供应商/无限）
│   │   │   ├── unit_type.go          # 单位类型
│   │   │   └── deduct_timing.go      # 扣减时机
│   │   ├── repository/
│   │   │   └── inventory_repo.go     # 仓储接口
│   │   └── service/
│   │       ├── stock_calculator.go   # 库存计算领域服务
│   │       └── reservation_manager.go # 预占管理
│   ├── application/
│   │   ├── dto/
│   │   │   ├── inventory_dto.go
│   │   │   └── reserve_dto.go
│   │   └── service/
│   │       ├── inventory_app_service.go   # 库存应用服务
│   │       ├── reserve_app_service.go     # 预占应用服务
│   │       ├── sync_app_service.go        # 同步应用服务
│   │       └── reconcile_app_service.go   # 对账应用服务
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   ├── mysql/
│   │   │   │   ├── inventory_repo_impl.go
│   │   │   │   └── migrations/
│   │   │   └── redis/
│   │   │       ├── inventory_cache.go
│   │   │       └── lua/
│   │   │           ├── reserve_stock.lua    # 原子预占脚本
│   │   │           └── release_stock.lua    # 原子释放脚本
│   │   ├── strategy/
│   │   │   ├── self_managed_strategy.go     # 自营库存策略
│   │   │   ├── supplier_strategy.go         # 供应商库存策略
│   │   │   └── unlimited_strategy.go        # 无限库存策略
│   │   ├── supplier/
│   │   │   ├── sync_adapter.go              # 供应商同步适配器
│   │   │   └── realtime_checker.go          # 实时库存查询
│   │   ├── event/
│   │   │   └── kafka_publisher.go           # Kafka事件发布
│   │   ├── rpc/
│   │   │   ├── product_client.go
│   │   │   └── supplier_client.go
│   │   └── job/
│   │       ├── cleanup_expired_reserves.go  # 清理过期预占
│   │       └── reconcile_job.go             # 库存对账任务
│   └── interfaces/
│       ├── grpc/
│       │   ├── inventory_handler.go
│       │   └── proto/
│       └── http/
│           └── inventory_controller.go
├── config/
│   └── config.yaml
└── go.mod
```

#### 核心领域模型

**库存聚合根（Inventory）**：

```go
type Inventory struct {
    ID              int64
    SKUID           int64
    TotalStock      int64   // 总库存
    AvailableStock  int64   // 可用库存
    ReservedStock   int64   // 预占库存
    SoldStock       int64   // 已售库存
    
    // 库存策略
    ManagementType  ManagementType  // 1:自营 2:供应商 3:无限
    UnitType        UnitType        // 1:券码 2:数量 3:时间
    DeductTiming    DeductTiming    // 1:下单扣 2:支付扣
    
    // 供应商相关
    SupplierID      *int64
    SyncStrategy    string  // realtime/scheduled/push
    LastSyncAt      *time.Time
    
    // 预占记录（聚合内存）
    reservations    []*Reservation
    
    // 版本控制
    Version         int
    UpdatedAt       time.Time
}

// Reserve 预占库存（领域方法）
func (inv *Inventory) Reserve(quantity int64, orderID string, userID int64) (*Reservation, error) {
    if inv.AvailableStock < quantity {
        return nil, ErrInsufficientStock
    }
    
    reservation := &Reservation{
        ID:        uuid.New().String(),
        SKUID:     inv.SKUID,
        OrderID:   orderID,
        UserID:    userID,
        Quantity:  quantity,
        Status:    ReservationStatusPending,
        ExpiresAt: time.Now().Add(15 * time.Minute),
        CreatedAt: time.Now(),
    }
    
    inv.AvailableStock -= quantity
    inv.ReservedStock += quantity
    inv.reservations = append(inv.reservations, reservation)
    
    // 发布领域事件
    inv.publishEvent(&StockReservedEvent{
        SKUID:    inv.SKUID,
        OrderID:  orderID,
        Quantity: quantity,
    })
    
    return reservation, nil
}
```

#### Redis原子操作（Lua脚本）

```lua
-- reserve_stock.lua (原子库存预占)
local sku_key = KEYS[1]              -- inventory:sku:12345
local reserve_key = KEYS[2]          -- reserve:uuid-xxx
local expire_zset_key = KEYS[3]      -- reserve:expiry

local quantity = tonumber(ARGV[1])
local reserve_data = ARGV[2]         -- JSON序列化的预占数据
local expires_at = tonumber(ARGV[3]) -- Unix时间戳
local ttl = tonumber(ARGV[4])        -- 15分钟

-- 检查库存
local available = tonumber(redis.call('HGET', sku_key, 'available'))
if not available or available < quantity then
    return {err = 'insufficient_stock'}
end

-- 扣减可用库存、增加预占库存
redis.call('HINCRBY', sku_key, 'available', -quantity)
redis.call('HINCRBY', sku_key, 'reserved', quantity)

-- 写入预占记录
redis.call('SET', reserve_key, reserve_data, 'EX', ttl)

-- 添加到过期索引（用于定时清理）
redis.call('ZADD', expire_zset_key, expires_at, reserve_key)

return {ok = 'success', available = available - quantity}
```

---

### 3.3 商品上架服务（Listing Service）

#### 目录结构

```
listing-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── domain/
│   │   ├── model/
│   │   │   ├── listing_task.go       # 上架任务聚合根
│   │   │   ├── audit_record.go       # 审核记录
│   │   │   └── publish_record.go     # 发布记录
│   │   ├── value_object/
│   │   │   ├── task_status.go        # 任务状态枚举
│   │   │   └── audit_result.go       # 审核结果
│   │   ├── repository/
│   │   │   └── listing_repo.go
│   │   └── service/
│   │       ├── state_machine.go      # 状态机领域服务
│   │       ├── audit_router.go       # 审核路由（按品类）
│   │       └── validator.go          # 业务规则校验
│   ├── application/
│   │   ├── dto/
│   │   │   ├── listing_dto.go
│   │   │   └── audit_dto.go
│   │   ├── service/
│   │   │   ├── listing_app_service.go    # 上架应用服务
│   │   │   ├── audit_app_service.go      # 审核应用服务
│   │   │   ├── publish_app_service.go    # 发布应用服务
│   │   │   └── batch_import_service.go   # 批量导入
│   │   └── saga/
│   │       └── publish_saga.go           # 发布Saga编排
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   ├── mysql/
│   │   │   └── redis/
│   │   ├── state_machine/
│   │   │   ├── config/
│   │   │   │   └── transitions.yaml     # 状态转换配置
│   │   │   ├── guard/                   # 状态转换守卫
│   │   │   └── handler/                 # 状态处理器
│   │   ├── audit/
│   │   │   └── strategy/
│   │   │       ├── manual_audit.go      # 人工审核
│   │   │       ├── auto_audit.go        # 自动审核
│   │   │       └── risk_audit.go        # 风控审核
│   │   ├── datasource/
│   │   │   ├── supplier_sync.go         # 供应商同步
│   │   │   ├── excel_import.go          # Excel导入
│   │   │   └── api_import.go            # API导入
│   │   ├── rpc/
│   │   │   ├── product_client.go
│   │   │   ├── inventory_client.go
│   │   │   ├── pricing_client.go
│   │   │   └── search_client.go
│   │   ├── event/
│   │   │   └── kafka_publisher.go
│   │   └── job/
│   │       ├── auto_publish_job.go      # 定时自动发布
│   │       └── expire_check_job.go      # 过期检查
│   └── interfaces/
│       ├── grpc/
│       └── http/
├── config/
│   ├── config.yaml
│   └── state_machine.yaml
└── go.mod
```

#### 状态机设计

```go
// StateMachine 状态机接口
type StateMachine interface {
    Transition(ctx context.Context, task *ListingTask, event Event) error
    CanTransition(currentStatus TaskStatus, event Event) bool
}

// StateMachineImpl 状态机实现
type StateMachineImpl struct {
    transitions map[TaskStatus]map[Event]TaskStatus  // 状态转换表
    guards      map[Event]Guard                      // 转换守卫
    handlers    map[TaskStatus]Handler               // 状态处理器
}

// Transition 状态转换
func (sm *StateMachineImpl) Transition(ctx context.Context, task *ListingTask, event Event) error {
    // 1. 检查转换合法性
    if !sm.CanTransition(task.Status, event) {
        return fmt.Errorf("invalid transition: %s -> %s", task.Status, event)
    }
    
    // 2. 执行守卫检查
    if guard, ok := sm.guards[event]; ok {
        if err := guard.Check(ctx, task); err != nil {
            return fmt.Errorf("guard check failed: %w", err)
        }
    }
    
    // 3. 获取目标状态
    targetStatus := sm.transitions[task.Status][event]
    
    // 4. 执行状态处理器
    if handler, ok := sm.handlers[targetStatus]; ok {
        if err := handler.Handle(ctx, task); err != nil {
            return fmt.Errorf("handler failed: %w", err)
        }
    }
    
    // 5. 更新状态
    oldStatus := task.Status
    task.Status = targetStatus
    task.UpdatedAt = time.Now()
    
    // 6. 发布领域事件
    task.PublishEvent(&StatusChangedEvent{
        TaskID:    task.ID,
        OldStatus: oldStatus,
        NewStatus: targetStatus,
        Event:     event,
    })
    
    return nil
}
```

#### Saga编排（发布流程）

```go
// PublishSaga 发布Saga编排器
type PublishSaga struct {
    productClient   rpc.ProductClient
    inventoryClient rpc.InventoryClient
    pricingClient   rpc.PricingClient
    searchClient    rpc.SearchClient
}

// Execute 执行发布Saga
func (s *PublishSaga) Execute(ctx context.Context, task *ListingTask) error {
    // 定义步骤
    steps := []SagaStep{
        &CreateProductStep{client: s.productClient},
        &InitInventoryStep{client: s.inventoryClient},
        &SetupPricingStep{client: s.pricingClient},
        &IndexSearchStep{client: s.searchClient},
    }
    
    executedSteps := make([]SagaStep, 0)
    
    // 顺序执行步骤
    for _, step := range steps {
        if err := step.Execute(ctx, task); err != nil {
            // 触发补偿
            s.compensate(ctx, task, executedSteps)
            return fmt.Errorf("saga step %s failed: %w", step.Name(), err)
        }
        executedSteps = append(executedSteps, step)
    }
    
    return nil
}

// compensate 补偿逻辑
func (s *PublishSaga) compensate(ctx context.Context, task *ListingTask, executedSteps []SagaStep) {
    // 逆序执行补偿
    for i := len(executedSteps) - 1; i >= 0; i-- {
        step := executedSteps[i]
        if err := step.Compensate(ctx, task); err != nil {
            // 记录补偿失败日志，进入人工处理队列
            log.Error("compensation failed", "step", step.Name(), "error", err)
        }
    }
}
```

---

### 3.4 供应商网关（Supplier Gateway）

#### 目录结构

```
supplier-gateway-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── domain/
│   │   ├── model/
│   │   │   ├── supplier.go
│   │   │   └── supplier_config.go
│   │   └── repository/
│   │       └── supplier_repo.go
│   ├── application/
│   │   ├── dto/
│   │   └── service/
│   │       ├── gateway_service.go
│   │       ├── health_check_service.go
│   │       └── metrics_service.go
│   ├── infrastructure/
│   │   ├── adapter/
│   │   │   ├── interface.go              # 供应商适配器接口
│   │   │   ├── base_adapter.go           # 基础适配器（模板方法）
│   │   │   ├── flight/
│   │   │   │   ├── supplier_a_adapter.go # 机票供应商A
│   │   │   │   └── supplier_b_adapter.go
│   │   │   ├── hotel/
│   │   │   │   ├── supplier_c_adapter.go
│   │   │   │   └── supplier_d_adapter.go
│   │   │   └── protocol/
│   │   │       ├── http_converter.go     # HTTP协议转换
│   │   │       ├── soap_converter.go     # SOAP协议转换
│   │   │       └── grpc_converter.go     # gRPC协议转换
│   │   ├── circuit_breaker/
│   │   │   └── hystrix_wrapper.go        # Hystrix熔断器
│   │   ├── rate_limiter/
│   │   │   ├── token_bucket.go           # 令牌桶算法
│   │   │   ├── sliding_window.go         # 滑动窗口算法
│   │   │   └── redis_limiter.go          # 分布式限流（Redis）
│   │   ├── retry/
│   │   │   ├── exponential_backoff.go    # 指数退避
│   │   │   └── fixed_backoff.go          # 固定间隔
│   │   ├── router/
│   │   │   ├── load_balancer.go          # 负载均衡（多供应商）
│   │   │   └── failover.go               # 故障切换
│   │   ├── monitor/
│   │   │   ├── metrics_collector.go      # 指标采集
│   │   │   ├── health_checker.go         # 健康检查
│   │   │   └── alerter.go                # 告警
│   │   └── cache/
│   │       ├── config_cache.go           # 配置缓存
│   │       └── response_cache.go         # 响应缓存（幂等）
│   └── interfaces/
│       └── grpc/
├── config/
│   ├── config.yaml
│   └── suppliers/                        # 供应商配置（外部化）
│       ├── flight/
│       │   ├── supplier_a.yaml
│       │   └── supplier_b.yaml
│       └── hotel/
│           ├── supplier_c.yaml
│           └── supplier_d.yaml
└── go.mod
```

#### 供应商适配器接口

```go
// SupplierAdapter 供应商适配器接口
type SupplierAdapter interface {
    // 查询商品库存
    QueryStock(ctx context.Context, req *StockQueryRequest) (*StockQueryResponse, error)
    
    // 创建订单
    CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
    
    // 查询订单状态
    QueryOrder(ctx context.Context, orderID string) (*OrderStatusResponse, error)
    
    // 取消订单
    CancelOrder(ctx context.Context, orderID string) error
    
    // 健康检查
    HealthCheck(ctx context.Context) error
}

// BaseAdapter 基础适配器（模板方法模式）
type BaseAdapter struct {
    config         *SupplierConfig
    httpClient     *http.Client
    circuitBreaker *hystrix.CircuitBreaker
    rateLimiter    RateLimiter
    retryPolicy    RetryPolicy
}

// Execute 执行模板方法（封装通用逻辑）
func (a *BaseAdapter) Execute(ctx context.Context, operation string, fn func() (interface{}, error)) (interface{}, error) {
    // 1. 限流检查
    if a.rateLimiter != nil && a.config.RateLimit.Enabled {
        if err := a.rateLimiter.Allow(ctx, a.config.SupplierID); err != nil {
            return nil, fmt.Errorf("rate limit exceeded: %w", err)
        }
    }
    
    // 2. 熔断执行
    if a.config.CircuitBreaker.Enabled {
        return a.circuitBreaker.Execute(func() (interface{}, error) {
            return a.executeWithRetry(ctx, operation, fn)
        })
    }
    
    return a.executeWithRetry(ctx, operation, fn)
}

// executeWithRetry 带重试的执行
func (a *BaseAdapter) executeWithRetry(ctx context.Context, operation string, fn func() (interface{}, error)) (interface{}, error) {
    var lastErr error
    
    for attempt := 1; attempt <= a.config.Retry.MaxAttempts; attempt++ {
        // 记录指标
        startTime := time.Now()
        
        result, err := fn()
        
        duration := time.Since(startTime)
        a.recordMetrics(operation, duration, err)
        
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // 判断是否可重试
        if !a.isRetryable(err) {
            break
        }
        
        // 退避等待
        if attempt < a.config.Retry.MaxAttempts {
            backoff := a.retryPolicy.NextBackoff(attempt)
            time.Sleep(backoff)
        }
    }
    
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

#### 供应商配置示例

```yaml
# config/suppliers/flight/supplier_a.yaml
supplier_id: "flight_supplier_a"
supplier_type: "flight"
supplier_name: "航空供应商A"
priority: "P0"  # P0:核心 P1:重要 P2:一般

# 协议配置
protocol: "HTTP"
base_url: "https://api.supplier-a.com"
timeout: 800ms
auth:
  type: "api_key"
  api_key: "${SUPPLIER_A_API_KEY}"  # 环境变量

# 熔断配置
circuit_breaker:
  enabled: true
  error_threshold: 60.0       # 错误率阈值60%
  min_requests: 20            # 最小请求数
  timeout: 1000               # 熔断超时1秒
  sleep_window: 5000          # 熔断后等待5秒进入半开状态

# 限流配置
rate_limit:
  enabled: true
  qps: 500                    # 每秒500次请求
  burst_size: 600             # 突发容量600
  time_window: 1              # 时间窗口1秒

# 重试配置
retry:
  enabled: true
  max_attempts: 3
  backoff_policy: "exponential"  # exponential/fixed
  initial_delay: 100             # 初始延迟100ms
  max_delay: 1000                # 最大延迟1秒
  retryable_errors:
    - "timeout"
    - "connection_refused"
    - "503"

# 降级配置
fallback:
  enabled: false
  fallback_data: null

# 监控配置
monitor:
  enabled: true
  alert_threshold:
    error_rate: 10.0      # 错误率>10%告警
    latency_p99: 2000     # P99延迟>2秒告警
```

---

## 四、数据流设计

### 4.1 同步 vs 异步

**分类原则**：

| 场景 | 调用方式 | 典型用例 |
|-----|---------|---------|
| **同步RPC** | 用户等待、需要立即返回结果 | 结算试算、库存查询、下单、支付 |
| **异步事件** | 非阻塞、最终一致性 | 订单状态变更通知、搜索索引更新、数据分析 |

### 4.2 用户搜索商品（导购场景）

```
场景：用户在首页搜索"无线耳机"

┌─────────────────────────────────────────────────────────────┐
│  搜索商品时序图（同步调用）                                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  [APP/Web]                                                  │
│      ↓ GET /search?keyword=无线耳机&page=1&size=20         │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ [API Gateway]                                           ││
│  │  • 鉴权：验证JWT Token（可选，支持游客搜索）            ││
│  │  • 限流：IP限流100次/分钟，防爬虫                       ││
│  │  • 路由：转发到 Aggregation Service                    ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 鉴权通过，转发请求                                   │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ [Aggregation Service] - 聚合服务（编排层）             ││
│  │  职责：编排多个微服务调用，聚合数据返回                 ││
│  │                                                          ││
│  │  Step 1: 缓存检查                                       ││
│  │      key = "search:无线耳机:page1:size20"               ││
│  │      ├─ 缓存命中 → 直接返回（5ms）✓                     ││
│  │      └─ 缓存未命中 → 继续查询                           ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 缓存未命中                                           │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 2: 调用Search Service查询ES                        ││
│  │  ┌─────────────────────────────────────────────────┐   ││
│  │  │ [Search Service] - 搜索服务                     │   ││
│  │  │  ↓ 调用 Elasticsearch                           │   ││
│  │  │  ┌─────────────────────────────────────────┐   │   ││
│  │  │  │ [Elasticsearch]                          │   │   ││
│  │  │  │  • 全文搜索："无线耳机"（中文分词）      │   │   ││
│  │  │  │  • 过滤条件：status=online               │   │   ││
│  │  │  │  • 排序规则：综合排序（销量+价格+评分）  │   │   ││
│  │  │  │  • 分页：from=0, size=20                │   │   ││
│  │  │  │  • 高亮：标题、描述中的关键词高亮        │   │   ││
│  │  │  │  • 聚合：品牌、价格区间、分类（筛选项）  │   │   ││
│  │  │  │  查询耗时：30-50ms                       │   │   ││
│  │  │  └─────────────────────────────────────────┘   │   ││
│  │  │      ↓ 返回                                      │   ││
│  │  │  {                                               │   ││
│  │  │    "sku_ids": [1001, 1002, ..., 1020],          │   ││
│  │  │    "total": 1230,                                │   ││
│  │  │    "filters": {...}  // 筛选项聚合               │   ││
│  │  │  }                                               │   ││
│  │  │  50ms                                            │   ││
│  │  └─────────────────────────────────────────────────┘   ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 获得 sku_ids: [1001, 1002, ..., 1020]               │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 3: 并发调用基础数据服务（无依赖关系）               ││
│  │  ┌────────────────────┐     ┌────────────────────┐      ││
│  │  │ [Product Center]   │     │ [Inventory Service]│      ││
│  │  │  RPC: BatchGet     │     │  RPC: BatchCheck   │      ││
│  │  │  Products          │     │  Stock             │      ││
│  │  │  (sku_ids)         │     │  (sku_ids)         │      ││
│  │  │  ↓                 │     │  ↓                 │      ││
│  │  │  返回：            │     │  返回：            │      ││
│  │  │  • title           │     │  • available_stock │      ││
│  │  │  • images          │     │  • stock_status    │      ││
│  │  │  • brand           │     │    (in_stock/      │      ││
│  │  │  • category        │     │     out_of_stock)  │      ││
│  │  │  • base_price ✓    │     │  • sold_count      │      ││
│  │  │  • attributes      │     │                    │      ││
│  │  │  50ms              │     │  30ms              │      ││
│  │  └────────────────────┘     └────────────────────┘      ││
│  │                                                          ││
│  │  并发调用，总耗时：max(50, 30) = 50ms                   ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 获得商品基础信息 + 基础价格 + 库存状态               │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 4: 调用Marketing Service获取营销信息               ││
│  │  ┌──────────────────────────────────────────────────┐  ││
│  │  │ [Marketing Service]                              │  ││
│  │  │  RPC: BatchGetPromotions(sku_ids, user_id)      │  ││
│  │  │  ↓                                                │  ││
│  │  │  返回每个SKU的营销活动：                          │  ││
│  │  │  • promo_id: 活动ID                              │  ││
│  │  │  • promo_type: 折扣/满减/限时购                  │  ││
│  │  │  • discount_rate: 0.9（九折）                    │  ││
│  │  │  • discount_amount: 400（满2000减400）           │  ││
│  │  │  • available_coupons: 可用优惠券列表             │  ││
│  │  │  70ms                                             │  ││
│  │  └──────────────────────────────────────────────────┘  ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 获得营销信息                                         │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 5: 调用Pricing Service计算最终价格                 ││
│  │  （依赖Step 3的base_price + Step 4的营销信息）          ││
│  │  ┌──────────────────────────────────────────────────┐  ││
│  │  │ [Pricing Service]                                │  ││
│  │  │  RPC: BatchCalculatePrice(items)                │  ││
│  │  │  输入：                                           │  ││
│  │  │  [                                               │  ││
│  │  │    {                                             │  ││
│  │  │      "sku_id": 1001,                             │  ││
│  │  │      "base_price": 2399.00,                      │  ││
│  │  │      "promo_id": "PROMO_001",                    │  ││
│  │  │      "discount_rate": 0.9,                       │  ││
│  │  │      "quantity": 1                               │  ││
│  │  │    },                                            │  ││
│  │  │    ...                                           │  ││
│  │  │  ]                                               │  ││
│  │  │  ↓                                                │  ││
│  │  │  内部计算流程：                                   │  ││
│  │  │  1. 应用促销折扣                                 │  ││
│  │  │     promo_price = base_price × discount_rate    │  ││
│  │  │                 = 2399 × 0.9 = 2159.1           │  ││
│  │  │  2. 查询Fee配置（服务费、税费）                  │  ││
│  │  │     fee = 0（部分商品免服务费）                  │  ││
│  │  │  3. 计算最终价格                                 │  ││
│  │  │     final_price = promo_price + fee             │  ││
│  │  │                 = 2159.1 + 0 = 2159.1           │  ││
│  │  │  ↓                                                │  ││
│  │  │  返回每个SKU的价格信息：                          │  ││
│  │  │  • original_price: 2399.00（原价）               │  ││
│  │  │  • promo_price: 2159.00（促销价）                │  ││
│  │  │  • discount_amount: 240.00（优惠金额）           │  ││
│  │  │  • fee: 0                                        │  ││
│  │  │  • final_price: 2159.00（最终价格）              │  ││
│  │  │  100ms                                            │  ││
│  │  └──────────────────────────────────────────────────┘  ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 获得最终价格                                         │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 6: 数据聚合与处理                                  ││
│  │  • 合并：商品信息 + 价格 + 库存 + 营销                  ││
│  │  • 无货商品置底或隐藏                                   ││
│  │  • 图片CDN地址拼接                                      ││
│  │  • 敏感信息过滤（成本价、供应商ID等）                   ││
│  │  • 个性化排序（已登录用户基于历史行为调整）             ││
│  │  处理耗时：20ms                                         ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 聚合完成                                             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 7: 写入缓存（异步，不阻塞返回）                     ││
│  │  ├─ Redis SET "search:无线耳机:page1:size20" result    ││
│  │  ├─ TTL: 10分钟（热门搜索词）                           ││
│  │  └─ 后台任务：记录搜索日志到Kafka（用于搜索分析）       ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓                                                      │
│  [APP/Web] ← 返回搜索结果                                   │
│      {                                                      │
│        "total": 1230,                                       │
│        "items": [                                           │
│          {                                                  │
│            "sku_id": 1001,                                  │
│            "title": "Sony无线<em>耳机</em> WH-1000XM5",     │
│            "brand": "Sony",                                 │
│            "image": "https://cdn.example.com/1001.jpg",    │
│            "price": {                                       │
│              "original": 2399.00,   // Product Center      │
│              "promo": 2159.00,      // Pricing Service     │
│              "discount": 240.00,    // 优惠金额            │
│              "promo_tag": "限时9折" // Marketing Service   │
│            },                                               │
│            "stock": {                                       │
│              "status": "in_stock",  // Inventory Service   │
│              "available": 450,                              │
│              "message": "现货充足"                          │
│            },                                               │
│            "sales": 12580           // Search Service ES   │
│          },                                                 │
│          ...                                                │
│        ],                                                   │
│        "filters": {  // 筛选项聚合结果（来自ES）            │
│          "brands": ["Sony", "Bose", "Apple", ...],         │
│          "price_ranges": ["0-500", "500-1000", ...]        │
│        }                                                    │
│      }                                                      │
│                                                             │
│  总耗时（实时聚合）：                                        │
│    ES查询(50ms)                                             │
│    + 并发调用Product+Inventory(50ms)                       │
│    + Marketing(70ms)                                        │
│    + Pricing(100ms)                                         │
│    + 数据聚合(20ms)                                         │
│    = 50 + 50 + 70 + 100 + 20 = 290ms                       │
│                                                             │
│  P95延迟：< 350ms（实时聚合完整流程）                       │
│  P95延迟：< 50ms（Redis缓存命中，80%+场景）                │
└─────────────────────────────────────────────────────────────┘

异步流程（不阻塞用户）：
┌─────────────────────────────────────────────────────────────┐
│  1. 用户搜索行为追踪                                         │
│  [Kafka Topic: search.query]                                │
│      {                                                      │
│        "user_id": 67890,                                    │
│        "keyword": "无线耳机",                               │
│        "result_count": 1230,                                │
│        "clicked_sku_ids": [],                               │
│        "timestamp": 1776138000                              │
│      }                                                      │
│      ↓                                                      │
│  订阅者：                                                    │
│  ├─→ [Analytics Service] 监听：搜索热词统计、转化率分析     │
│  ├─→ [Recommendation Service] 监听：更新用户画像           │
│  └─→ [Search Service] 监听：优化搜索排序算法（A/B Test）    │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  2. ES索引价格更新（保证搜索结果价格准确性）                 │
│  [Kafka Topic: price.updated]                               │
│      发布者：Pricing Service（定时任务或营销活动触发）      │
│      {                                                      │
│        "sku_ids": [1001, 1002, ...],                        │
│        "prices": [                                          │
│          {                                                  │
│            "sku_id": 1001,                                  │
│            "base_price": 2399.00,                           │
│            "promo_price": 1999.00,                          │
│            "discount_amount": 400.00,                       │
│            "discount_reason": "满2000减400",                │
│            "valid_until": 1776224400                        │
│          },                                                 │
│          ...                                                │
│        ],                                                   │
│        "timestamp": 1776138000                              │
│      }                                                      │
│      ↓                                                      │
│  订阅者：                                                    │
│  └─→ [Search Service] 监听：批量更新ES索引中的价格字段      │
│      UPDATE product_index SET                               │
│        base_price = ?,                                      │
│        promo_price = ?,                                     │
│        discount_amount = ?                                  │
│      WHERE sku_id IN (...)                                  │
│                                                             │
│  更新频率：                                                  │
│  • 定时任务：每5分钟全量更新（增量更新有变化的商品）        │
│  • 营销活动生效：实时推送（如限时折扣开始/结束）            │
└─────────────────────────────────────────────────────────────┘
```

**关键设计要点**：

1. **Aggregation Service（聚合服务）**：
   - 职责：编排多个微服务的调用顺序，聚合数据
   - 与BFF的区别：专注于搜索场景，不区分端（Web/App共用）
   - 为什么需要：
     - 解耦业务逻辑（Search Service只负责ES查询）
     - 统一编排（避免客户端多次RPC调用）
     - 便于扩展（新增数据源只需修改聚合服务）
   - 部署：独立微服务，6副本（QPS 6000）

2. **分阶段调用策略（关键）**：
   - 严格按依赖关系顺序调用，不能完全并发
   - **阶段1**：Search Service查询ES → 获得sku_ids（50ms）
   - **阶段2**：并发调用Product Center + Inventory Service（50ms）
     - 无依赖关系，可并发
     - Product返回商品信息 + **base_price**（关键）
   - **阶段3**：调用Marketing Service获取营销信息（70ms）
     - 需要sku_ids作为输入
   - **阶段4**：调用Pricing Service计算最终价格（100ms）
     - 依赖：base_price（来自Product） + 营销信息（来自Marketing）
     - 内部逻辑：应用折扣 + 计算Fee
   - **总耗时**：50 + 50 + 70 + 100 + 20(聚合) = 290ms

3. **数据依赖关系**：
   ```
   ┌──────────────────────────────────────────────┐
   │  服务调用依赖关系图                           │
   ├──────────────────────────────────────────────┤
   │                                              │
   │  Search Service (ES查询)                     │
   │      ↓ 提供 sku_ids                          │
   │  ┌────────────────┐   ┌────────────────┐    │
   │  │ Product Center │   │ Inventory Svc  │    │
   │  │ (并发)         │   │ (并发)         │    │
   │  └────────────────┘   └────────────────┘    │
   │      ↓ 提供 base_price                       │
   │  ┌────────────────┐                          │
   │  │ Marketing Svc  │                          │
   │  └────────────────┘                          │
   │      ↓ 提供 discount_info                    │
   │  ┌────────────────┐                          │
   │  │ Pricing Service│ ← 依赖 base_price +      │
   │  │                │   discount_info          │
   │  └────────────────┘                          │
   │      ↓ 返回 final_price                      │
   │  聚合返回                                     │
   └──────────────────────────────────────────────┘
   ```

4. **多级缓存**：
   - L1缓存（Redis）：完整搜索结果10分钟缓存，命中率80%+
   - L2缓存（ES本地缓存）：ES节点本地缓存
   - 缓存Key设计：`search:{keyword}:{page}:{size}:{user_id?}`
   - 个性化场景：已登录用户加user_id，未登录用户共享缓存

5. **降级策略**：
   - Marketing Service异常 → Pricing使用base_price，无折扣
   - Pricing Service异常 → 只展示base_price，标记"价格加载中"
   - Inventory Service异常 → 隐藏库存信息
   - Product Service异常 → 整个搜索失败（核心依赖）

6. **性能优化**：
   - 批量接口：所有RPC都使用BatchXXX批量接口
   - 超时控制：每个RPC 200ms超时，避免雪崩
   - 熔断保护：错误率>50%自动熔断
   - 限流保护：Aggregation Service QPS 6000

7. **性能指标**：
   - P50延迟：< 50ms（Redis缓存命中，80%场景）
   - P95延迟：< 350ms（实时聚合完整流程）
   - P99延迟：< 500ms
   - QPS峰值：6000（Aggregation Service需6副本）

---

### 4.3 查询商品详情

```
场景：用户从搜索结果点击进入商品详情页

┌─────────────────────────────────────────────────────────────┐
│  商品详情时序图（同步 + 缓存）                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  [APP/Web]                                                  │
│      ↓ GET /products/12345/detail                          │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ [API Gateway]                                           ││
│  │  • 鉴权：可选（游客可查看）                             ││
│  │  • 限流：用户限流30次/分钟（防刷）                      ││
│  │  • CDN：静态资源（图片、视频）CDN加速                   ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 路由到 Product Center                               │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ [Product Center] - 商品中心                             ││
│  │  Step 1: 三级缓存查询                                   ││
│  │  ┌─────────────────────────────────────────────────┐   ││
│  │  │ L1: 本地缓存（Ristretto）                       │   ││
│  │  │  key = "product:12345"                          │   ││
│  │  │  ├─ 命中 → 返回（<1ms）✓ 80%命中率             │   ││
│  │  │  └─ 未命中 → 查L2                               │   ││
│  │  │         ↓                                        │   ││
│  │  │ L2: Redis（分布式缓存）                          │   ││
│  │  │  key = "product:sku:12345"                      │   ││
│  │  │  ├─ 命中 → 返回（<5ms）✓ 95%命中率             │   ││
│  │  │  └─ 未命中 → 查L3                               │   ││
│  │  │         ↓                                        │   ││
│  │  │ L3: MySQL（权威数据）                            │   ││
│  │  │  SELECT * FROM product WHERE sku_id=12345       │   ││
│  │  │  └─ 返回（10-30ms）                             │   ││
│  │  └─────────────────────────────────────────────────┘   ││
│  │      ↓ 获得商品基础信息                                 ││
│  │      {                                                  ││
│  │        "sku_id": 12345,                                 ││
│  │        "spu_id": 5678,                                  ││
│  │        "title": "Sony WH-1000XM5 无线降噪耳机",         ││
│  │        "brand": "Sony",                                 ││
│  │        "category_id": 101,                              ││
│  │        "images": ["img1.jpg", "img2.jpg", ...],        ││
│  │        "attributes": {                                  ││
│  │          "color": "黑色",                               ││
│  │          "connectivity": "蓝牙5.2"                      ││
│  │        },                                               ││
│  │        "description": "...",                            ││
│  │        "status": "online"                               ││
│  │      }                                                  ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 并发查询其他维度数据（5个服务并发调用）              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 2: 并发调用多个微服务（扇出模式）                  ││
│  │                                                          ││
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐        ││
│  │  │ [Pricing]  │  │ [Inventory]│  │ [Marketing]│        ││
│  │  │  RPC:      │  │  RPC:      │  │  RPC:      │        ││
│  │  │  Calculate │  │  GetStock  │  │  GetProduct│        ││
│  │  │  Price     │  │  (sku_id)  │  │  Promotions│        ││
│  │  │  (sku_id)  │  │            │  │  (sku_id,  │        ││
│  │  │  ↓         │  │  ↓         │  │   user_id) │        ││
│  │  │  返回：    │  │  返回：    │  │  ↓         │        ││
│  │  │  base_price│  │  total: 500│  │  返回：    │        ││
│  │  │  2399.00   │  │  available │  │  • 可用券  │        ││
│  │  │  promo:    │  │  450       │  │  • 单品折扣│        ││
│  │  │  1999.00   │  │  reserved  │  │  • 跨商品  │        ││
│  │  │  100ms     │  │  50        │  │    促销    │        ││
│  │  │            │  │  30ms      │  │  • 组合优惠│        ││
│  │  │            │  │            │  │  80ms      │        ││
│  │  └────────────┘  └────────────┘  └────────────┘        ││
│  │                                                          ││
│  │  ┌────────────┐  ┌────────────┐                        ││
│  │  │ [Review]   │  │ [Recommend]│                        ││
│  │  │  RPC:      │  │  RPC:      │                        ││
│  │  │  GetReviews│  │  GetRelated│                        ││
│  │  │  (sku_id)  │  │  Products  │                        ││
│  │  │  ↓         │  │  (sku_id)  │                        ││
│  │  │  返回：    │  │  ↓         │                        ││
│  │  │  评分4.8   │  │  推荐商品  │                        ││
│  │  │  评论列表  │  │  [sku_id   │                        ││
│  │  │  (top 10)  │  │  list]     │                        ││
│  │  │  60ms      │  │  120ms     │                        ││
│  │  └────────────┘  └────────────┘                        ││
│  │                                                          ││
│  │  并发调用，总耗时：max(100,30,80,60,120) = 120ms       ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 数据聚合                                             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 3: 数据聚合与个性化处理                            ││
│  │  • 合并所有维度数据                                     ││
│  │  • 个性化推荐（已登录用户）                             ││
│  │  • 库存状态判断：                                       ││
│  │    - available >= 10 → "现货充足"                      ││
│  │    - available < 10 → "仅剩X件"                        ││
│  │    - available = 0 → "暂时缺货，到货通知"              ││
│  │  • 价格展示策略：                                       ││
│  │    - 有促销 → 显示划线价 + 促销价                      ││
│  │    - 有券 → 显示"券后价XXX元"                          ││
│  │  • 营销信息处理（吸引多买）：                           ││
│  │    - 按优先级排序促销活动（多买优惠 > 组合优惠）       ││
│  │    - 计算省钱提示："再买1件，可省XXX元"                ││
│  │    - 生成推荐购买数量（基于最优惠方案）                 ││
│  │    - 标记高价值促销（highlight: true）                 ││
│  │  • 敏感信息过滤（供应商信息、成本价等）                 ││
│  │  处理耗时：20ms                                         ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 聚合完成                                             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 4: 缓存回写 + 异步事件                             ││
│  │  • 写入Redis（完整商品详情）                            ││
│  │    key = "product:detail:12345"                         ││
│  │    TTL = 30分钟                                         ││
│  │  • 写入本地缓存（热点商品）                             ││
│  │  • 发布浏览事件到Kafka（用户行为分析）                  ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓                                                      │
│  [APP/Web] ← 返回商品详情                                   │
│      {                                                      │
│        "sku_id": 12345,                                     │
│        "title": "Sony WH-1000XM5 无线降噪耳机",             │
│        "images": ["https://cdn.example.com/img1.jpg", ...],│
│        "price": {                                           │
│          "original": 2399.00,                               │
│          "current": 1999.00,                                │
│          "coupon_available": true,                          │
│          "coupon_after": 1899.00                            │
│        },                                                   │
│        "stock": {                                           │
│          "status": "in_stock",                              │
│          "message": "仅剩12件"                              │
│        },                                                   │
│        "promotions": {                 // 营销信息（吸引多买）│
│          "active_promotions": [        // 当前生效的促销    │
│            {                                                │
│              "id": "PROMO_001",                             │
│              "type": "multi_buy",      // 多买优惠          │
│              "title": "买2件享9折",                         │
│              "description": "再买1件，立享9折优惠",         │
│              "conditions": {                                │
│                "min_quantity": 2,                           │
│                "discount_rate": 0.9                         │
│              },                                             │
│              "highlight": true,        // 前端高亮显示      │
│              "expires_at": "2026-04-20 23:59:59"           │
│            },                                               │
│            {                                                │
│              "id": "PROMO_002",                             │
│              "type": "bundle",         // 组合优惠          │
│              "title": "搭配充电器立减50元",                 │
│              "description": "购买耳机+充电器组合，减50元",  │
│              "bundle_products": [                           │
│                {                                            │
│                  "sku_id": 1005,                            │
│                  "title": "Sony快充充电器",                 │
│                  "price": 99.00,                            │
│                  "discount": 50.00                          │
│                }                                            │
│              ]                                              │
│            },                                               │
│            {                                                │
│              "id": "PROMO_003",                             │
│              "type": "category_discount", // 品类折扣      │
│              "title": "配件类满3件享8折",                   │
│              "description": "音频配件买满3件，享受8折优惠", │
│              "conditions": {                                │
│                "category": "音频配件",                      │
│                "min_quantity": 3,                           │
│                "discount_rate": 0.8                         │
│              }                                              │
│            }                                                │
│          ],                                                 │
│          "coupons": [                  // 可用优惠券        │
│            {                                                │
│              "code": "SAVE100",                             │
│              "title": "满2000减100",                        │
│              "threshold": 2000.00,                          │
│              "discount": 100.00,                            │
│              "expires_at": "2026-04-30"                     │
│            }                                                │
│          ],                                                 │
│          "saving_tips": {              // 省钱提示          │
│            "message": "再买1件，可享9折优惠，共省480元",    │
│            "recommended_quantity": 2,                       │
│            "total_savings": 480.00                          │
│          }                                                  │
│        },                                                   │
│        "reviews": {                                         │
│          "rating": 4.8,                                     │
│          "count": 12580,                                    │
│          "top_reviews": [...]                               │
│        },                                                   │
│        "related_products": [...],      // 相关推荐          │
│        "attributes": {...}             // 商品属性          │
│      }                                                      │
│                                                             │
│  总耗时：缓存查询(5ms) + 并发RPC(120ms) + 聚合(20ms)       │
│         = 145ms（P95 < 200ms，P99 < 300ms）                │
└─────────────────────────────────────────────────────────────┘

异步流程（用户行为追踪）：
┌─────────────────────────────────────────────────────────────┐
│  [Kafka Topic: user.behavior]                               │
│      消息内容：                                              │
│      {                                                      │
│        "user_id": 67890,                                    │
│        "event_type": "view_product",                        │
│        "sku_id": 12345,                                     │
│        "from_source": "search_result",  // 来源            │
│        "timestamp": 1776138000                              │
│      }                                                      │
│      ↓                                                      │
│  订阅者：                                                    │
│  ├─→ [Recommendation Service] 监听：更新用户兴趣标签        │
│  ├─→ [Analytics Service] 监听：漏斗分析（浏览→加购→下单）   │
│  ├─→ [Marketing Service] 监听：触发再营销（浏览未购买）     │
│  └─→ [Product Center] 监听：热度统计（更新商品热度排序）    │
└─────────────────────────────────────────────────────────────┘
```

**关键设计要点**：

1. **三级缓存策略**：
   - L1本地缓存：5分钟TTL，热点商品命中率80%+
   - L2 Redis缓存：30分钟TTL，整体命中率95%+
   - L3 MySQL：权威数据源

2. **并发调用优化**：
   - 5个微服务并发调用（扇出模式）
   - 使用超时控制（每个RPC 200ms超时）
   - 部分服务失败不影响主流程（降级）

3. **营销信息展示（促进多买）**：
   - **多买优惠**：展示"买2件享9折"，刺激用户增加购买数量
   - **组合优惠**：推荐搭配商品（如耳机+充电器），提升客单价
   - **品类促销**：展示"配件类满3件享8折"，引导用户购买同品类商品
   - **省钱提示**：明确告知"再买1件，可省480元"，量化优惠金额
   - **高亮显示**：重要促销信息前端高亮展示，提升转化率
   - **实时计算**：根据用户已选商品，动态计算最优优惠方案

4. **降级策略**：
   - Pricing异常 → 显示"价格加载中"或使用缓存价格
   - Inventory异常 → 隐藏库存信息或显示"请联系客服"
   - Marketing异常 → 隐藏营销模块（不影响购买）
   - Review异常 → 隐藏评论模块
   - Recommend异常 → 隐藏推荐商品或使用默认推荐

5. **性能指标**：
   - P50延迟：< 50ms（本地缓存命中）
   - P95延迟：< 200ms（Redis缓存命中）
   - P99延迟：< 300ms
   - QPS峰值：8000（Product Center需6副本）

6. **热点商品保护**：
   - 本地缓存前置（避免Redis热key）
   - 限流保护（单SKU QPS限制）
   - 降级开关（大促时关闭非核心功能如推荐）

7. **营销数据来源**：
   - Marketing Service统一管理所有促销规则
   - 支持A/B测试（不同用户展示不同促销）
   - 促销活动实时生效（无需重启服务）
   - 缓存TTL短（5分钟），确保促销信息及时更新

---

### 4.4 加购与试算（无购物车模式）

```
场景：用户选择多个商品（不同SKU），需要实时计算总价
适用于：快速结账场景（如电影票、充值卡等），无需持久化购物车

┌─────────────────────────────────────────────────────────────┐
│  加购与试算时序图（同步调用）                                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  用户操作流程：                                              │
│  1. 在商品详情页选择 SKU + 数量（前端临时存储）              │
│  2. 可添加多个商品（前端维护临时列表）                       │
│  3. 点击"结算"按钮，触发试算接口                             │
│                                                             │
│  [APP/Web] - 前端临时存储                                    │
│      selectedItems = [                                      │
│        { sku_id: 1001, quantity: 2, category: "耳机" },    │
│        { sku_id: 1005, quantity: 1, category: "充电器" },  │
│        { sku_id: 2003, quantity: 3, category: "数据线" }   │
│      ]                                                      │
│      ↓ POST /checkout/calculate                            │
│      {                                                      │
│        "user_id": 67890,                                    │
│        "items": [                                           │
│          { "sku_id": 1001, "quantity": 2 },                │
│          { "sku_id": 1005, "quantity": 1 },                │
│          { "sku_id": 2003, "quantity": 3 }                 │
│        ],                                                   │
│        "coupon_codes": ["SAVE50"]   // 用户选择的优惠券     │
│      }                                                      │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ [API Gateway]                                           ││
│  │  • 鉴权：必须登录（user_id验证）                        ││
│  │  • 限流：用户限流10次/分钟（防止恶意试算）              ││
│  │  • 参数校验：items数量≤20（防止超大订单）              ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 转发到 Checkout Service                             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ [Checkout Service] - 结算服务（编排层）                 ││
│  │  职责：编排多服务调用，计算订单总价                     ││
│  │                                                          ││
│  │  Step 1: 参数预处理与去重                               ││
│  │      • 合并相同SKU（quantity累加）                       ││
│  │      • 去除无效SKU（quantity≤0）                        ││
│  │      • 构建sku_ids列表：[1001, 1005, 2003]             ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓                                                      │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 2: 并发调用基础数据服务（2个服务）                 ││
│  │                                                          ││
│  │  ┌────────────────────┐  ┌────────────────────┐        ││
│  │  │ [Product Center]   │  │ [Inventory Service]│        ││
│  │  │  RPC: BatchGet     │  │  RPC: BatchCheck   │        ││
│  │  │  Products          │  │  Stock             │        ││
│  │  │  (sku_ids)         │  │  (sku_ids)         │        ││
│  │  │  ↓                 │  │  ↓                 │        ││
│  │  │  返回：            │  │  返回：            │        ││
│  │  │  • base_price      │  │  • available_stock │        ││
│  │  │  • title           │  │  • stock_status    │        ││
│  │  │  • category        │  │                    │        ││
│  │  │  50ms              │  │  30ms              │        ││
│  │  └────────────────────┘  └────────────────────┘        ││
│  │                                                          ││
│  │  并发调用，总耗时：max(50, 30) = 50ms                   ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 获得商品基础信息 + 基础价格 + 库存状态               │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 3: 库存校验（关键步骤，决定是否可下单）            ││
│  │  遍历每个SKU，检查库存：                                ││
│  │  • available_stock >= quantity → 可下单                ││
│  │  • available_stock < quantity → 返回错误                ││
│  │      错误信息："商品[XXX]库存不足，仅剩N件"            ││
│  │  • stock_status = "out_of_stock" → 返回错误            ││
│  │      错误信息："商品[XXX]已售罄"                       ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 库存校验通过                                         │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 4: 调用Marketing Service获取营销活动               ││
│  │  ┌──────────────────────────────────────────────────┐  ││
│  │  │ [Marketing Service]                              │  ││
│  │  │  RPC: CalculatePromotions(items, user_id)       │  ││
│  │  │  ↓                                                │  ││
│  │  │  输入：商品列表 + 用户ID                          │  ││
│  │  │  • 商品级别促销：                                 │  ││
│  │  │    - SKU 1001: 单件折扣9折                       │  ││
│  │  │    - SKU 1005: 限时购特价                        │  ││
│  │  │  • 跨商品促销（关键）：                           │  ││
│  │  │    - 满300减50（全场）                           │  ││
│  │  │    - 买3件打8折（同品类）                        │  ││
│  │  │    - 组合优惠：耳机+充电器减20元                  │  ││
│  │  │  • 用户优惠券：                                   │  ││
│  │  │    - 券码: SAVE50（满500减50）                   │  ││
│  │  │  ↓                                                │  ││
│  │  │  返回营销活动列表（按优先级排序）：               │  ││
│  │  │  [                                               │  ││
│  │  │    {                                             │  ││
│  │  │      "promo_id": "P001",                         │  ││
│  │  │      "type": "sku_discount",  // 商品级别折扣     │  ││
│  │  │      "sku_id": 1001,                             │  ││
│  │  │      "discount_rate": 0.9                        │  ││
│  │  │    },                                            │  ││
│  │  │    {                                             │  ││
│  │  │      "promo_id": "P002",                         │  ││
│  │  │      "type": "order_reduce",  // 订单级别满减    │  ││
│  │  │      "threshold": 300,                           │  ││
│  │  │      "reduce": 50                                │  ││
│  │  │    },                                            │  ││
│  │  │    {                                             │  ││
│  │  │      "promo_id": "P003",                         │  ││
│  │  │      "type": "category_discount", // 品类折扣    │  ││
│  │  │      "category": "配件",                         │  ││
│  │  │      "min_quantity": 3,                          │  ││
│  │  │      "discount_rate": 0.8                        │  ││
│  │  │    },                                            │  ││
│  │  │    {                                             │  ││
│  │  │      "promo_id": "P004",                         │  ││
│  │  │      "type": "coupon",        // 优惠券          │  ││
│  │  │      "code": "SAVE50",                           │  ││
│  │  │      "threshold": 500,                           │  ││
│  │  │      "reduce": 50                                │  ││
│  │  │    }                                             │  ││
│  │  │  ]                                               │  ││
│  │  │  80ms                                             │  ││
│  │  └──────────────────────────────────────────────────┘  ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 获得营销活动列表                                     │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 5: 调用Pricing Service计算最终价格                 ││
│  │  （复杂的价格计算逻辑，处理多层级优惠叠加）              ││
│  │  ┌──────────────────────────────────────────────────┐  ││
│  │  │ [Pricing Service]                                │  ││
│  │  │  RPC: CalculateFinalPrice(items, promos)        │  ││
│  │  │  ↓                                                │  ││
│  │  │  计算流程（4层架构）：                            │  ││
│  │  │                                                   │  ││
│  │  │  1. 商品原价计算                                  │  ││
│  │  │     SKU 1001: 299 × 2 = 598元                   │  ││
│  │  │     SKU 1005: 89 × 1 = 89元                     │  ││
│  │  │     SKU 2003: 19 × 3 = 57元                     │  ││
│  │  │     小计：744元                                   │  ││
│  │  │                                                   │  ││
│  │  │  2. 应用商品级别促销                              │  ││
│  │  │     SKU 1001: 598 × 0.9 = 538.2元（9折）        │  ││
│  │  │     SKU 1005: 89元（无促销）                     │  ││
│  │  │     SKU 2003: 57 × 0.8 = 45.6元（买3件8折）     │  ││
│  │  │     小计：672.8元                                 │  ││
│  │  │                                                   │  ││
│  │  │  3. 应用订单级别促销                              │  ││
│  │  │     满300减50：672.8 - 50 = 622.8元             │  ││
│  │  │                                                   │  ││
│  │  │  4. 应用优惠券                                    │  ││
│  │  │     满500减50：622.8 - 50 = 572.8元             │  ││
│  │  │                                                   │  ││
│  │  │  最终总价：572.8元                                │  ││
│  │  │  （注：运费、服务费等在确认下单时才计算）        │  ││
│  │  │  ↓                                                │  ││
│  │  │  返回详细价格明细：                               │  ││
│  │  │  {                                               │  ││
│  │  │    "items": [                                    │  ││
│  │  │      {                                           │  ││
│  │  │        "sku_id": 1001,                           │  ││
│  │  │        "quantity": 2,                            │  ││
│  │  │        "unit_price": 299.00,                     │  ││
│  │  │        "subtotal": 598.00,                       │  ││
│  │  │        "discount": 59.80,   // 9折优惠           │  ││
│  │  │        "final_price": 538.20                     │  ││
│  │  │      },                                          │  ││
│  │  │      // ... 其他商品                             │  ││
│  │  │    ],                                            │  ││
│  │  │    "subtotal": 744.00,          // 商品原价合计  │  ││
│  │  │    "item_discount": 71.20,      // 商品级别优惠  │  ││
│  │  │    "order_discount": 50.00,     // 订单级别优惠  │  ││
│  │  │    "coupon_discount": 50.00,    // 优惠券优惠    │  ││
│  │  │    "total": 572.80,             // 应付总额      │  ││
│  │  │    "saved": 171.20,             // 节省金额      │  ││
│  │  │    "promotions": [               // 已应用的促销  │  ││
│  │  │      { "id": "P001", "desc": "单件9折", "amount": 59.80 },│
│  │  │      { "id": "P002", "desc": "满300减50", "amount": 50.00 },│
│  │  │      { "id": "P003", "desc": "买3件8折", "amount": 11.40 },│
│  │  │      { "id": "P004", "desc": "优惠券SAVE50", "amount": 50.00 }│
│  │  │    ]                                             │  ││
│  │  │  }                                               │  ││
│  │  │  120ms                                            │  ││
│  │  └──────────────────────────────────────────────────┘  ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓ 获得最终价格明细                                     │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Step 6: 返回试算结果（不写入任何数据，纯计算）           ││
│  │  • 不创建订单                                           ││
│  │  • 不预占库存                                           ││
│  │  • 不扣券                                               ││
│  │  • 仅返回计算结果供用户确认                             ││
│  └─────────────────────────────────────────────────────────┘│
│      ↓                                                      │
│  [APP/Web] ← 返回试算结果                                   │
│      {                                                      │
│        "can_checkout": true,            // 是否可下单       │
│        "total": 572.80,                                     │
│        "saved": 171.20,                                     │
│        "items": [...],                  // 商品明细         │
│        "promotions": [...],             // 促销明细         │
│        "price_breakdown": {             // 价格分解         │
│          "subtotal": 744.00,            // 商品原价合计     │
│          "discount": 171.20,            // 总优惠金额       │
│          "final": 572.80                // 应付总额         │
│        }                                                    │
│      }                                                      │
│                                                             │
│  总耗时：并发查询(50ms) + 营销(80ms) + 计价(100ms) = 230ms │
│         （P95 < 300ms，P99 < 500ms）                       │
│                                                             │
│  说明：                                                     │
│  • 试算只计算商品价格和营销优惠                             │
│  • 运费、服务费在确认下单时再计算（需要地址信息）           │
│  • 简化计算流程，提升响应速度                               │
└─────────────────────────────────────────────────────────────┘
```

#### 关键设计要点

##### 1. 与购物车模式的对比

| 维度 | 购物车模式 | 无购物车模式（加购试算） |
|-----|-----------|---------------------|
| **数据持久化** | 需要（Redis/MySQL，保留7天） | 不需要（前端临时存储） |
| **适用场景** | 传统电商、需要跨设备同步 | 快速结账、单次性购买（票务、充值） |
| **用户操作** | 加购 → 进入购物车 → 修改 → 结算 | 选择商品 → 直接结算 |
| **后端服务** | Cart Service（CRUD操作） | Checkout Service（只有计算） |
| **数据一致性** | 需要处理购物车过期、失效 | 无需考虑（临时数据） |
| **系统复杂度** | 高（需要购物车同步、清理） | 低（无状态计算） |

##### 2. 跨商品营销规则处理

**关键挑战**：多个商品之间的营销规则互相影响，需要按优先级计算。

**营销规则分类**：

```
商品级别（Item-Level）
├─ 单品折扣（SKU Discount）：某个SKU享受9折
├─ 限时购（Flash Sale）：某个SKU特价
└─ 买N送M（Bundle）：买2送1

品类级别（Category-Level）
├─ 品类折扣（Category Discount）：配件类8折
└─ 品类满减（Category Reduce）：数码类满200减20

订单级别（Order-Level）
├─ 满减（Threshold Reduce）：满300减50
├─ 满折（Threshold Discount）：满500打9折
└─ 阶梯折扣（Tiered Discount）：满1000打8折

优惠券级别（Coupon-Level）
├─ 满减券：满500减50
├─ 折扣券：全场9折
└─ 品类券：数码类专用券
```

**优先级计算策略**（从上到下应用）：

```go
// PricingService中的计算流程（加购试算场景）
func (s *PricingService) CalculateFinalPrice(items []*Item, promos []*Promotion) *PriceDetail {
    // 1. 计算商品原价
    subtotal := calculateSubtotal(items)
    
    // 2. 应用商品级别促销（每个SKU独立计算）
    itemDiscount := applyItemLevelPromotions(items, promos)
    
    // 3. 应用品类级别促销（按品类分组计算）
    categoryDiscount := applyCategoryLevelPromotions(items, promos)
    
    // 4. 应用订单级别促销（全局计算）
    orderDiscount := applyOrderLevelPromotions(subtotal - itemDiscount - categoryDiscount, promos)
    
    // 5. 应用优惠券（最后应用，避免券叠加问题）
    couponDiscount := applyCouponPromotions(subtotal - itemDiscount - categoryDiscount - orderDiscount, promos)
    
    // 6. 最终总价（注：运费、服务费在确认下单时才计算）
    total := subtotal - itemDiscount - categoryDiscount - orderDiscount - couponDiscount
    
    return &PriceDetail{
        Subtotal:         subtotal,
        ItemDiscount:     itemDiscount,
        CategoryDiscount: categoryDiscount,
        OrderDiscount:    orderDiscount,
        CouponDiscount:   couponDiscount,
        Total:            total,
        Saved:            itemDiscount + categoryDiscount + orderDiscount + couponDiscount,
    }
}
```

##### 3. 实时试算 vs 下单确认的区别

| 阶段 | 试算（Calculate） | 确认下单（Confirm） |
|-----|------------------|-------------------|
| **API路径** | `POST /checkout/calculate` | `POST /checkout/confirm` |
| **计算内容** | 商品价格+营销优惠 | 商品价格+营销优惠+运费+服务费 |
| **库存操作** | 只查询，不预占 | 预占库存（Reserve） |
| **优惠券** | 只校验，不扣减 | 扣减优惠券 |
| **订单创建** | 不创建 | 创建订单 |
| **数据持久化** | 无 | 有 |
| **幂等性要求** | 无（纯计算） | 强（防重复下单） |
| **响应时间** | <500ms | <1s |
| **调用频率** | 高（用户多次试算） | 低（一次性操作） |

##### 4. 性能优化策略

**缓存策略**：
```
- 商品基础信息：L1+L2缓存，TTL 30分钟
- 营销规则：Redis缓存，TTL 5分钟（规则变化频繁）
- 优惠券信息：Redis缓存，TTL 10分钟
```

**并发优化**：
```
并发查询：Product + Inventory（2个服务）
串行查询：Marketing → Pricing（有数据依赖）
总耗时：并发(50ms) + 营销(80ms) + 计价(100ms) = 230ms
```

**降级策略**：
```
- 营销服务失败：只返回原价，不影响试算
- 库存服务失败：隐藏库存状态，但标记"库存待确认"
- 优惠券校验失败：移除该优惠券，继续计算
```

---

### 4.5 用户下单全链路数据流（先创单后支付模式）

> **核心设计模式**：预占-确认 两阶段提交（2PC）
> - Phase 1: 试算（性能优先，可用快照）
> - Phase 2: 创单（锁定资源，库存预占）
> - Phase 3: 支付（用户选择，渠道计费）
> - Phase 4: 确认（资源扣减，订单完成）
> - Phase 5: 超时（资源释放，订单取消）

#### API接口总览

| API接口 | 请求方 | 服务提供方 | 核心功能 | 关键操作 | 响应时间 | 调用时机 |
|---------|-------|-----------|---------|---------|---------|---------|
| **POST /checkout/calculate** | APP/Web | Checkout Service<br>Aggregation Service | 结算试算 | 1. 检查库存（不扣减）<br>2. 计算基础价格+营销优惠<br>3. 可使用快照数据 | 80-230ms | Phase 1<br>用户点击"去结算" |
| **POST /checkout/confirm** | APP/Web | Checkout Service<br>Order Service | 确认下单<br>创建订单 | 1. **库存预占**（CAS操作）<br>2. 实时查询商品+营销<br>3. 创建订单（PENDING_PAYMENT）<br>4. 发布order.created事件 | <500ms | Phase 2<br>用户点击"提交订单" |
| **POST /payment/calculate** | APP/Web | Payment Service | 支付前试算 | 1. 校验优惠券有效性<br>2. 计算Coin抵扣<br>3. 计算支付渠道费<br>4. 实时返回最终金额 | 100-200ms | Phase 3a<br>用户选择优惠券/Coin<br>（防抖100ms） |
| **POST /payment/create** | APP/Web | Payment Service<br>Payment Gateway | 创建支付 | 1. 后端重新计算金额（防篡改）<br>2. **预扣优惠券和Coin**<br>3. 创建支付记录<br>4. 调用支付网关（支付宝/微信） | 200-300ms | Phase 3b<br>用户点击"确认支付" |
| **POST /payment/callback** | 支付宝/微信 | Payment Service<br>Order Service | 支付成功回调 | 1. 幂等性校验<br>2. **确认库存扣减**<br>3. **确认优惠券/Coin扣减**<br>4. 更新订单状态（PAID）<br>5. 发布payment.paid事件 | <200ms | Phase 4<br>用户完成支付（异步通知） |

**内部RPC调用**：

| RPC接口 | 调用方 | 服务提供方 | 核心功能 | 关键操作 |
|---------|-------|-----------|---------|---------|
| **GetProducts()** | Checkout Service | Product Center | 查询商品信息 | 返回商品基础信息、价格 |
| **GetPromotions()** | Checkout Service | Marketing Service | 查询营销活动 | 返回当前有效的营销活动 |
| **CheckStock()** | Checkout Service | Inventory Service | 检查库存 | 查询可用库存（不扣减） |
| **ReserveStock()** | Checkout Service | Inventory Service | 库存预占 | Redis Lua原子操作，扣减可用库存，记录预占 |
| **ConfirmReserve()** | Payment Service | Inventory Service | 确认库存扣减 | 删除预占记录，确认扣减 |
| **ReleaseStock()** | Order Timeout Job | Inventory Service | 释放库存 | 恢复可用库存，删除预占记录 |
| **ValidateCoupon()** | Payment Service | Marketing Service | 校验优惠券 | 校验有效性、使用条件、适用范围 |
| **ReserveCoupon()** | Payment Service | Marketing Service | 预扣优惠券 | 状态：AVAILABLE → RESERVED |
| **ConfirmCoupon()** | Payment Service | Marketing Service | 确认扣减优惠券 | 状态：RESERVED → USED |
| **ReleaseCoupon()** | Order Timeout Job | Marketing Service | 回退优惠券 | 状态：RESERVED → AVAILABLE |
| **GetUserCoins()** | Payment Service | Marketing Service | 查询Coin余额 | 返回用户可用Coin数量 |
| **ReserveCoin()** | Payment Service | Marketing Service | 预扣Coin | available → reserved |
| **ConfirmCoin()** | Payment Service | Marketing Service | 确认扣减Coin | reserved → used |
| **ReleaseCoin()** | Order Timeout Job | Marketing Service | 回退Coin | reserved → available |
| **CalculateBasePrice()** | Checkout Service | Pricing Service | 计算基础价格 | 商品基础价格 + 营销优惠 |
| **CreateOrder()** | Checkout Service | Order Service | 创建订单 | 插入订单记录（PENDING_PAYMENT） |
| **UpdateOrderStatus()** | Payment Service | Order Service | 更新订单状态 | PENDING_PAYMENT → PAID → COMPLETED |

**Kafka事件**：

| Topic | 发布者 | 订阅者 | 触发时机 | 消息内容 |
|-------|-------|-------|---------|---------|
| **order.created** | Checkout Service | Cart, Search, Analytics, Notification | 订单创建成功 | order_id, user_id, items, amount, status |
| **payment.paid** | Payment Service | Order, Supplier Gateway, Notification, Analytics | 支付成功 | order_id, payment_id, amount, paid_at |
| **order.cancelled** | Order Timeout Job | Inventory, Marketing, Notification | 订单超时取消 | order_id, reason, cancelled_at |

---

```
用户操作流程：浏览商品 → 加购 → 试算 → 【创建订单】 → 【选择支付】 → 【完成支付】 → 履约

═══════════════════════════════════════════════════════════════
Phase 1: 试算计价（同步，性能优先，可用快照数据）
═══════════════════════════════════════════════════════════════

[APP/Web] 用户点击"去结算"
    ↓ POST /checkout/calculate
    ↓ {user_id, items: [{sku_id, quantity, snapshot}]}
    ↓
[Checkout Service/Aggregation Service]
    ↓
    ├─ Step 1: 判断快照是否过期
    │   ├─ 快照未过期 → 使用快照数据（商品信息、营销活动）
    │   └─ 快照过期 → 实时查询 Product + Marketing
    ↓
    ├─ Step 2: 实时查询库存（必须实时，不能用快照）
    │   └─→ [Inventory Service] RPC: CheckStock()
    │       └─ SELECT available FROM stock WHERE sku_id=?
    │       └─ 返回：available=10（检查不扣减）
    ↓
    ├─ Step 3: 调用计价服务（仅基础价格+营销）
    │   └─→ [Pricing Service] RPC: CalculateBasePrice()
    │       ├─ 商品基础价格：299.00
    │       ├─ 满减优惠：-30.00
    │       └─ 返回：base_price=299.00, discount=30.00
    ↓
    └─ Step 4: 返回试算结果
        ↓
[APP/Web] ← 返回试算详情（总耗时：80-230ms）
{
    "items": [...],
    "base_price": 299.00,
    "discount": 30.00,
    "amount_to_pay": 269.00,  // 待支付金额（未含支付渠道费）
    "available_coupons": [...] // 可用优惠券列表
}

═══════════════════════════════════════════════════════════════
Phase 2: 创建订单（同步 + 异步混合，锁定资源）
═══════════════════════════════════════════════════════════════

[APP/Web] 用户点击"提交订单"
    ↓ POST /checkout/confirm
    ↓ {user_id, items: [{sku_id, quantity}]}
    ↓
[Checkout Service] - Saga 协调者
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 1: 实时查询商品和营销（不使用快照）       │
    ├────────────────────────────────────────────────┤
    │ 并发调用：                                      │
    │   ├─→ [Product Center] RPC: GetProducts()     │
    │   │   └─ 返回：商品基础信息                    │
    │   └─→ [Marketing Service] RPC: GetPromotions()│
    │       └─ 返回：当前有效营销活动（实时校验）    │
    │ 耗时：100ms                                    │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 2: 库存预占（同步，必须成功）✅           │
    ├────────────────────────────────────────────────┤
    │ [Inventory Service] RPC: ReserveStock()        │
    │   ↓ Redis Lua 原子操作（CAS）                 │
    │   ↓ local available = redis.call('GET', key)  │
    │   ↓ if available >= quantity then             │
    │   ↓   redis.call('DECRBY', key, quantity)     │
    │   ↓   redis.call('SET', reserve_key, data, 'EX', 900) │
    │   ↓   return reserve_id                       │
    │   ↓ else return nil end                       │
    │   ↓                                            │
    │   └─ 返回：reserve_ids = ["rsv_001", ...]     │
    │   └─ 过期时间：15分钟（900秒）                 │
    │                                                │
    │ 库存变化：                                      │
    │   stock:available:1001 = 10 → 8（扣减2）      │
    │   stock:reserved:1001:rsv_001 = {              │
    │     quantity: 2,                               │
    │     order_id: null,                            │
    │     expires_at: 1744634100                     │
    │   } TTL=900秒                                  │
    │ 耗时：50ms                                     │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 3: 计算订单金额（基础价格+营销）          │
    ├────────────────────────────────────────────────┤
    │ [Pricing Service] RPC: CalculateBasePrice()    │
    │   ├─ 商品基础价格：299.00 × 2 = 598.00        │
    │   ├─ 满减优惠：-60.00                          │
    │   └─ 返回：base_price=598.00, discount=60.00  │
    │                                                │
    │ 注意：此时不计算支付渠道费（支付时计算）       │
    │ 耗时：80ms                                     │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 4: 创建订单（同步）✅                     │
    ├────────────────────────────────────────────────┤
    │ [Order Service] RPC: CreateOrder()             │
    │   ↓ INSERT INTO order_tab VALUES (            │
    │       order_id = 1001,                         │
    │       user_id = 67890,                         │
    │       status = 'PENDING_PAYMENT',  ← 关键状态  │
    │       base_price = 598.00,                     │
    │       discount = 60.00,                        │
    │       amount_to_pay = 538.00,                  │
    │       reserve_ids = '["rsv_001"]',             │
    │       pay_expire_at = NOW() + 15分钟,          │
    │       created_at = NOW()                       │
    │   )                                            │
    │   ↓                                            │
    │   └─ 返回：order_id = 1001                    │
    │ 耗时：100ms                                    │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 5: 发布 order.created 事件（异步）        │
    ├────────────────────────────────────────────────┤
    │ [Kafka Topic: order.created]                   │
    │   Payload: {                                   │
    │     order_id: 1001,                            │
    │     user_id: 67890,                            │
    │     items: [{sku_id, quantity}],               │
    │     amount: 538.00,                            │
    │     status: 'PENDING_PAYMENT',                 │
    │     reserve_ids: ["rsv_001"]                   │
    │   }                                            │
    │   ↓                                            │
    │   ├─→ [Cart Service] 监听：清理购物车          │
    │   ├─→ [Search Service] 监听：更新销量（+1）    │
    │   ├─→ [Analytics Service] 监听：订单漏斗分析   │
    │   └─→ [Notification Service] 监听：订单确认通知│
    └────────────────────────────────────────────────┘
    ↓
[APP/Web] ← 返回订单信息（总耗时：<500ms）
{
    "order_id": 1001,
    "base_price": 598.00,
    "discount": 60.00,
    "amount_to_pay": 538.00,
    "pay_expire_at": 1744634100,  // 15分钟后过期
    "status": "PENDING_PAYMENT",
    "reserved": true  // 库存已预占
}

═══════════════════════════════════════════════════════════════
Phase 3a: 支付页面试算（实时计算最终金额）✅ 关键
═══════════════════════════════════════════════════════════════

[APP/Web] 用户点击"去支付"，进入支付页面
    ↓
    ┌─────────────────────────────────────────────┐
    │ 支付页面展示：                               │
    │   • 订单金额：538.00                        │
    │   • 可用优惠券列表（实时查询）               │
    │   • 可用Coin余额：100                       │
    │   • 支付方式选择（支付宝、微信、银行卡）     │
    │   • 最终支付金额：待计算                    │
    └─────────────────────────────────────────────┘
    ↓
用户交互：选择/修改优惠券、Coin、支付渠道
    ↓ 每次选择变化时，触发实时试算（防抖100ms）
    ↓
    ↓ POST /payment/calculate（支付前试算）
    ↓ {
    ↓   order_id: 1001,
    ↓   coupon_id: "CPN001",        // 用户选择的优惠券
    ↓   coin_amount: 50,            // 用户选择使用50个Coin
    ↓   payment_channel: "alipay"   // 用户选择的支付渠道
    ↓ }
    ↓
[Payment Service]
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 1: 查询订单基础金额                       │
    ├────────────────────────────────────────────────┤
    │ [Order Service] RPC: GetOrder(1001)            │
    │   ↓ 校验订单状态                               │
    │   ├─ status == 'PENDING_PAYMENT' ✓            │
    │   ├─ pay_expire_at > NOW() ✓                  │
    │   └─ 返回：order {                            │
    │       order_id: 1001,                          │
    │       amount_to_pay: 538.00,  ← 订单基础金额   │
    │       items: [{sku_id, quantity, price}]       │
    │     }                                          │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 2: 校验并计算优惠券抵扣                   │
    ├────────────────────────────────────────────────┤
    │ if coupon_id != null {                         │
    │   [Marketing Service] RPC: ValidateCoupon()    │
    │     ↓ 校验优惠券：                             │
    │     ├─ 是否有效（未过期、未使用）              │
    │     ├─ 是否满足使用条件（满300减50）           │
    │     ├─ 是否适用当前订单（品类限制）            │
    │     └─ 返回：{                                │
    │         coupon_id: "CPN001",                   │
    │         type: "满减",                          │
    │         condition: 300,                        │
    │         discount: 50                           │
    │       }                                        │
    │                                                │
    │   计算抵扣金额：                                │
    │     if order.amount_to_pay >= 300 {            │
    │       coupon_discount = 50.00                  │
    │     }                                          │
    │ }                                              │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 3: 校验并计算Coin抵扣                     │
    ├────────────────────────────────────────────────┤
    │ if coin_amount > 0 {                           │
    │   [Marketing Service] RPC: GetUserCoins()      │
    │     ↓ 查询用户Coin余额                         │
    │     ├─ 可用余额：100                           │
    │     ├─ 本次使用：50（用户输入）                │
    │     └─ 校验：50 <= 100 ✓                      │
    │                                                │
    │   计算Coin抵扣金额：                            │
    │     coin_discount = coin_amount * 0.01         │
    │                   = 50 * 0.01 = 0.50           │
    │     // 通常1 Coin = 0.01元                    │
    │ }                                              │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 4: 计算支付渠道费                         │
    ├────────────────────────────────────────────────┤
    │ channel_fee = calculateChannelFee(              │
    │   payment_channel: 'alipay',                   │
    │   amount: 538.00                               │
    │ )                                              │
    │                                                │
    │ 渠道费率规则：                                  │
    │   • 支付宝/微信：0%                            │
    │   • 信用卡：1%                                 │
    │   • 花呗分期3期：2%                            │
    │   • 花呗分期6期：4%                            │
    │                                                │
    │ channel_fee = 0.00（支付宝无手续费）            │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 5: 计算最终支付金额（显示给用户）         │
    ├────────────────────────────────────────────────┤
    │ final_amount = order.amount_to_pay             │
    │              - coupon_discount                 │
    │              - coin_discount                   │
    │              + channel_fee                     │
    │              = 538.00 - 50.00 - 0.50 + 0.00    │
    │              = 487.50                          │
    │                                                │
    │ 价格明细：                                      │
    │   订单金额：538.00                             │
    │   优惠券：  -50.00                             │
    │   Coin抵扣：-0.50                              │
    │   渠道费：  +0.00                              │
    │   ──────────────                              │
    │   实付金额：487.50 ✅                          │
    └────────────────────────────────────────────────┘
    ↓
[APP/Web] ← 返回试算结果（实时更新页面）
{
    "order_id": 1001,
    "base_amount": 538.00,
    "coupon_discount": 50.00,
    "coin_discount": 0.50,
    "channel_fee": 0.00,
    "final_amount": 487.50,  ← 用户看到的最终金额
    "breakdown": {
        "订单金额": "¥538.00",
        "优惠券": "-¥50.00",
        "Coin抵扣": "-¥0.50",
        "支付渠道费": "+¥0.00"
    },
    "remaining_coin": 50  // 使用后剩余Coin
}

用户在支付页面看到实时更新的价格：
┌─────────────────────────────────────────────┐
│ 支付详情                                     │
│                                             │
│ 订单金额        ¥538.00                     │
│ 优惠券 CPN001  -¥50.00  ← 用户选择          │
│ Coin抵扣(50)   -¥0.50   ← 用户选择          │
│ 支付渠道费      ¥0.00   ← 自动计算          │
│ ─────────────────────                      │
│ 实付金额        ¥487.50  ← 实时更新 ✅       │
│                                             │
│ [确认支付] 按钮                              │
└─────────────────────────────────────────────┘

═══════════════════════════════════════════════════════════════
Phase 3b: 确认支付（用户点击"确认支付"）
═══════════════════════════════════════════════════════════════

[APP/Web] 用户点击"确认支付"
    ↓ POST /payment/create（创建支付记录）
    ↓ {
    ↓   order_id: 1001,
    ↓   payment_channel: "alipay",
    ↓   coupon_id: "CPN001",
    ↓   coin_amount: 50,
    ↓   expected_amount: 487.50  ← 前端试算的金额（防篡改校验）
    ↓ }
    ↓
[Payment Service]
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 1: 重新计算最终金额（后端校验）✅         │
    ├────────────────────────────────────────────────┤
    │ // 后端必须重新计算，不能信任前端传来的金额    │
    │ actual_amount = recalculate(                   │
    │   order_id, coupon_id, coin_amount, channel    │
    │ )                                              │
    │                                                │
    │ // 校验前端金额与后端计算是否一致              │
    │ if abs(actual_amount - expected_amount) > 0.01 {│
    │   return Error("价格已变化，请重新确认")        │
    │ }                                              │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 2: 预扣优惠券和Coin（支付成功后确认）     │
    ├────────────────────────────────────────────────┤
    │ if coupon_id != null {                         │
    │   [Marketing Service] RPC: ReserveCoupon()     │
    │     UPDATE coupon_user_log SET                 │
    │       status = 'RESERVED',  ← 预扣状态         │
    │       order_id = 1001,                         │
    │       reserved_at = NOW()                      │
    │ }                                              │
    │                                                │
    │ if coin_amount > 0 {                           │
    │   [Marketing Service] RPC: ReserveCoin()       │
    │     UPDATE user_coin SET                       │
    │       available = available - 50,  ← 预扣      │
    │       reserved = reserved + 50                 │
    │     WHERE user_id = 67890                      │
    │ }                                              │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 3: 创建支付记录                           │
    ├────────────────────────────────────────────────┤
    │ INSERT INTO payment VALUES (                   │
    │   payment_id = 2001,                           │
    │   order_id = 1001,                             │
    │   payment_channel = 'alipay',                  │
    │   base_amount = 538.00,                        │
    │   coupon_discount = 50.00,                     │
    │   coin_discount = 0.50,                        │
    │   channel_fee = 0.00,                          │
    │   final_amount = 487.50,  ← 最终金额           │
    │   status = 'PENDING',                          │
    │   created_at = NOW()                           │
    │ )                                              │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 4: 调用支付网关                           │
    ├────────────────────────────────────────────────┤
    │ [Payment Gateway] RPC: CreatePay()             │
    │   ↓ 调用支付宝/微信API                        │
    │   ↓ 参数：{                                   │
    │       amount: 487.50,  ← 最终支付金额          │
    │       order_no: "ORD1001",                     │
    │       callback_url: "https://api.com/callback" │
    │     }                                          │
    │   └─ 返回：pay_url（支付页面URL）             │
    └────────────────────────────────────────────────┘
    ↓
[APP/Web] ← 返回支付URL
{
    "payment_id": 2001,
    "pay_url": "https://alipay.com/...",
    "final_amount": 487.50,
    "qr_code": "data:image/png;base64,..."
}
    ↓
[APP/Web] 跳转到支付页面（或显示二维码）
    ↓
用户在支付宝/微信完成支付...

**关键设计说明**：
1. ✅ **支付前试算**：用户每次选择优惠券/Coin/支付渠道时，实时试算最终金额
2. ✅ **防抖优化**：试算接口100ms防抖，避免频繁调用
3. ✅ **后端校验**：创建支付时，后端必须重新计算金额，不信任前端
4. ✅ **预扣机制**：优惠券和Coin在创建支付时预扣，支付成功后确认扣减
5. ✅ **价格透明**：用户清楚看到每一项优惠的明细

═══════════════════════════════════════════════════════════════
Phase 4: 支付成功回调（异步通知，确认资源扣减）
═══════════════════════════════════════════════════════════════

[支付宝/微信] 支付成功
    ↓ POST /payment/callback（异步回调）
    ↓ {payment_id, trade_no, amount, ...}
    ↓
[Payment Service]
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 1: 幂等性校验                             │
    ├────────────────────────────────────────────────┤
    │ if isDuplicate(payment_id) {                   │
    │   return SUCCESS  // 重复回调，直接返回        │
    │ }                                              │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 2: 更新支付状态                           │
    ├────────────────────────────────────────────────┤
    │ UPDATE payment SET                             │
    │   status = 'PAID',                             │
    │   paid_at = NOW(),                             │
    │   trade_no = '支付宝流水号'                    │
    │ WHERE payment_id = 2001                        │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 3: 更新订单状态 ✅                        │
    ├────────────────────────────────────────────────┤
    │ [Order Service] RPC: UpdateOrderStatus()       │
    │   UPDATE order_tab SET                         │
    │     status = 'PAID',  ← 从 PENDING_PAYMENT 变更│
    │     paid_at = NOW()                            │
    │   WHERE order_id = 1001                        │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 4: 确认库存扣减（从预占转为实际扣减）✅   │
    ├────────────────────────────────────────────────┤
    │ [Inventory Service] RPC: ConfirmReserve()      │
    │   ↓ Redis 操作                                 │
    │   ↓ // 删除预占记录（已不需要）                │
    │   ↓ redis.call('DEL', 'stock:reserved:1001:rsv_001')│
    │   ↓                                            │
    │   ↓ // 库存已在预占时扣减，这里仅确认         │
    │   ↓ // stock:available:1001 = 8（已扣减）     │
    │   ↓                                            │
    │   └─ 返回：SUCCESS                            │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 5: 确认优惠券和Coin扣减（如果有）✅       │
    ├────────────────────────────────────────────────┤
    │ // 优惠券：从预扣状态转为确认扣减              │
    │ if coupon_id != null {                         │
    │   [Marketing Service] RPC: ConfirmCoupon()     │
    │     UPDATE coupon_user_log SET                 │
    │       status = 'USED',  ← 从RESERVED变为USED   │
    │       order_id = 1001,                         │
    │       used_at = NOW()                          │
    │     WHERE coupon_id = ? AND user_id = ?        │
    │ }                                              │
    │                                                │
    │ // Coin：从预扣状态转为确认扣减                │
    │ if coin_amount > 0 {                           │
    │   [Marketing Service] RPC: ConfirmCoin()       │
    │     UPDATE user_coin SET                       │
    │       reserved = reserved - 50,                │
    │       used = used + 50  ← 确认扣减             │
    │     WHERE user_id = 67890                      │
    │ }                                              │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 6: 发布 payment.paid 事件（异步）         │
    ├────────────────────────────────────────────────┤
    │ [Kafka Topic: payment.paid]                    │
    │   Payload: {                                   │
    │     order_id: 1001,                            │
    │     payment_id: 2001,                          │
    │     amount: 523.38,                            │
    │     paid_at: 1744633800                        │
    │   }                                            │
    │   ↓                                            │
    │   ├─→ [Supplier Gateway] 监听：提交供应商订单  │
    │   ├─→ [Notification Service] 监听：支付成功通知│
    │   ├─→ [Analytics Service] 监听：GMV统计        │
    │   └─→ [Risk Service] 监听：风控检测            │
    └────────────────────────────────────────────────┘

═══════════════════════════════════════════════════════════════
Phase 5: 供应商履约（异步）
═══════════════════════════════════════════════════════════════

[Supplier Gateway] 监听到 payment.paid 事件
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 1: 调用供应商API创建订单                  │
    ├────────────────────────────────────────────────┤
    │ [Supplier API] POST /create_order              │
    │   ↓ {sku_id, quantity, user_info, ...}        │
    │   ↓ 供应商侧处理...                           │
    │   └─ 返回：supplier_order_id = "S123456"      │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 2: 轮询供应商订单状态                     │
    ├────────────────────────────────────────────────┤
    │ 每隔10秒查询一次，最多查询60次（10分钟）       │
    │ [Supplier API] GET /query_order                │
    │   ↓ {supplier_order_id: "S123456"}            │
    │   └─ 返回：status = "SUCCESS", voucher_code   │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 3: 发布 order.fulfilled 事件              │
    ├────────────────────────────────────────────────┤
    │ [Kafka Topic: order.fulfilled]                 │
    │   Payload: {                                   │
    │     order_id: 1001,                            │
    │     supplier_order_id: "S123456",              │
    │     status: "COMPLETED",                       │
    │     voucher_code: "ABC123XYZ"                  │
    │   }                                            │
    │   ↓                                            │
    │   ├─→ [Order Service] 监听：更新订单状态→COMPLETED│
    │   ├─→ [Notification Service] 监听：发送券码/凭证│
    │   └─→ [Analytics Service] 监听：履约成功率统计 │
    └────────────────────────────────────────────────┘

═══════════════════════════════════════════════════════════════
Phase 6: 超时未支付处理（定时任务）
═══════════════════════════════════════════════════════════════

[Order Timeout Job] 定时扫描（每分钟执行一次）
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 1: 扫描超时未支付订单                     │
    ├────────────────────────────────────────────────┤
    │ SELECT * FROM order_tab WHERE                  │
    │   status = 'PENDING_PAYMENT'                   │
    │   AND pay_expire_at < NOW()                    │
    │ LIMIT 1000                                     │
    │   ↓                                            │
    │   └─ 返回：[order_1001, order_1002, ...]      │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 2: 释放库存（从预占状态释放）✅           │
    ├────────────────────────────────────────────────┤
    │ for each order {                               │
    │   [Inventory Service] RPC: ReleaseStock()      │
    │     ↓ Redis 操作                               │
    │     ↓ // 读取预占记录                         │
    │     ↓ reserve = redis.call('GET', reserve_key) │
    │     ↓ quantity = reserve.quantity              │
    │     ↓                                          │
    │     ↓ // 恢复可用库存                         │
    │     ↓ redis.call('INCRBY', available_key, quantity)│
    │     ↓ // stock:available:1001 = 8 → 10        │
    │     ↓                                          │
    │     ↓ // 删除预占记录                         │
    │     ↓ redis.call('DEL', reserve_key)           │
    │ }                                              │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 3: 回退优惠券和Coin（如果有预扣）         │
    ├────────────────────────────────────────────────┤
    │ // 查询订单关联的支付记录                      │
    │ payment = getPaymentByOrderID(order_id)        │
    │                                                │
    │ // 回退优惠券（从RESERVED回到AVAILABLE）       │
    │ if payment.coupon_id != null {                 │
    │   [Marketing Service] RPC: ReleaseCoupon()     │
    │     UPDATE coupon_user_log SET                 │
    │       status = 'AVAILABLE',  ← 回退预扣        │
    │       order_id = NULL,                         │
    │       reserved_at = NULL                       │
    │     WHERE coupon_id = ? AND user_id = ?        │
    │ }                                              │
    │                                                │
    │ // 回退Coin（从reserved回到available）         │
    │ if payment.coin_amount > 0 {                   │
    │   [Marketing Service] RPC: ReleaseCoin()       │
    │     UPDATE user_coin SET                       │
    │       available = available + 50,  ← 回退      │
    │       reserved = reserved - 50                 │
    │     WHERE user_id = 67890                      │
    │ }                                              │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 4: 更新订单状态                           │
    ├────────────────────────────────────────────────┤
    │ UPDATE order_tab SET                           │
    │   status = 'CANCELLED',  ← 从 PENDING_PAYMENT  │
    │   cancel_reason = '超时未支付',                 │
    │   cancelled_at = NOW()                         │
    │ WHERE order_id = ?                             │
    └────────────────────────────────────────────────┘
    ↓
    ┌────────────────────────────────────────────────┐
    │ Step 5: 发布 order.cancelled 事件              │
    ├────────────────────────────────────────────────┤
    │ [Kafka Topic: order.cancelled]                 │
    │   Payload: {                                   │
    │     order_id: 1001,                            │
    │     reason: '超时未支付',                       │
    │     cancelled_at: 1744634100                   │
    │   }                                            │
    │   ↓                                            │
    │   ├─→ [Analytics Service] 监听：订单漏斗分析   │
    │   └─→ [Notification Service] 监听：取消通知    │
    └────────────────────────────────────────────────┘
```

**关键数据流说明（完整资源状态变化）**：

| 阶段 | 操作 | 库存状态 | 订单状态 | 支付状态 | 优惠券/Coin | 耗时 |
|-----|------|---------|---------|---------|-----------|------|
| **Phase 1: 试算** | 检查库存（不扣减） | available=10 | - | - | 查询可用优惠券/Coin | 80-230ms |
| **Phase 2: 创单** | **库存预占** | available=8<br>reserved=2 | **PENDING_PAYMENT** | - | 未选择 | <500ms |
| **Phase 3a: 支付前试算** | 实时计算最终金额 | 不变 | 不变 | - | 查询并计算抵扣 | 100-200ms |
| **Phase 3b: 确认支付** | **预扣优惠券/Coin** | 不变 | 不变 | PENDING | coupon=RESERVED<br>coin: available=50, reserved=50 | 用户操作 |
| **Phase 4: 回调** | **确认扣减全部资源** | available=8（确认） | **PAID** | PAID | coupon=USED<br>coin: reserved=0, used=50 | <200ms |
| **Phase 5: 履约** | 供应商出票/出码 | 不变 | **COMPLETED** | 不变 | 不变 | 异步 |
| **Phase 6: 超时** | **释放全部资源** | available=10（回退+2） | **CANCELLED** | - | coupon=AVAILABLE<br>coin: available=100, reserved=0 | 定时任务 |

**核心设计亮点**：
1. ✅ **防止超卖**：创单时锁定库存（预占机制），其他用户无法下单
2. ✅ **用户体验好**：先锁定库存，再慢慢选择支付方式和优惠券
3. ✅ **实时试算**：支付页面实时展示最终金额（优惠券+Coin+渠道费）
4. ✅ **价格透明**：用户清楚看到每一项优惠的明细
5. ✅ **预占-确认机制**：库存、优惠券、Coin均采用"预占→确认"两阶段提交
6. ✅ **资源高效**：超时未支付自动释放全部资源（15分钟窗口）
7. ✅ **最终一致性**：通过 Kafka 事件驱动保证各系统状态一致

### 4.6 Kafka Topic设计

| Topic名称 | 发布者 | 订阅者 | 消息内容 | 分区数 | 副本数 | 保留时间 |
|----------|-------|-------|---------|-------|-------|---------|
| **order.created** | checkout-service | inventory, cart, search, analytics, notification | order_id, user_id, items, amount | 16 | 3 | 24h |
| **payment.paid** | payment-service | order, supplier-gateway, notification, analytics | order_id, payment_id, amount, paid_at | 16 | 3 | 24h |
| **inventory.reserved** | inventory-service | order, analytics | sku_id, reserve_id, quantity, expires_at | 8 | 3 | 24h |
| **inventory.deducted** | inventory-service | analytics, supplier-sync | sku_id, quantity, order_id | 8 | 3 | 24h |
| **order.fulfilled** | supplier-gateway | order, inventory, notification | order_id, supplier_order_id, status | 16 | 3 | 24h |
| **order.cancelled** | order-service | inventory, payment, notification | order_id, reason, cancelled_at | 8 | 3 | 24h |
| **product.updated** | product-center | search, listing, pricing | sku_id, update_type, data | 8 | 3 | 24h |
| **search.indexed** | search-service | analytics | sku_ids, indexed_at | 4 | 3 | 24h |
| **compensation.tasks** | checkout/order/inventory | compensation-worker | task_id, type, payload, retry_count | 4 | 3 | 7d |

---

## 五、关键技术决策

### 5.1 数据库设计

**分库分表策略（订单表）**：
- **分库**：8个库（按 `user_id % 8`）
- **分表**：32张表/库（按 `order_id % 32`）
- **总表数**：256张表
- **路由规则**：
  ```
  db_index = user_id % 8
  table_index = order_id % 32
  table_name = f"order_tab_{table_index}"
  ```

**订单主表设计**：

```sql
CREATE TABLE order_tab (
    order_id BIGINT PRIMARY KEY COMMENT '订单ID（雪花算法）',
    order_no VARCHAR(32) UNIQUE NOT NULL COMMENT '订单号',
    user_id BIGINT NOT NULL COMMENT '用户ID（分库键）',
    order_type VARCHAR(20) NOT NULL COMMENT 'flight/hotel/movie/topup',
    
    -- 金额
    total_amount DECIMAL(10,2) NOT NULL COMMENT '订单总额',
    discount_amount DECIMAL(10,2) DEFAULT 0 COMMENT '优惠金额',
    paid_amount DECIMAL(10,2) NOT NULL COMMENT '实付金额',
    
    -- 快照引用
    price_snapshot_id VARCHAR(64) NOT NULL COMMENT '价格快照ID',
    product_snapshot_id VARCHAR(64) NOT NULL COMMENT '商品快照ID',
    reserve_ids JSON COMMENT '库存预占ID列表',
    
    -- 状态
    order_status VARCHAR(20) NOT NULL COMMENT '订单状态',
    payment_status VARCHAR(20) NOT NULL COMMENT '支付状态',
    fulfillment_status VARCHAR(20) NOT NULL COMMENT '履约状态',
    
    -- 供应商信息
    supplier_id BIGINT COMMENT '供应商ID',
    supplier_order_id VARCHAR(64) COMMENT '供应商订单ID',
    
    -- 时间戳
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMP NULL,
    fulfilled_at TIMESTAMP NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 乐观锁
    version INT NOT NULL DEFAULT 0 COMMENT '版本号',
    
    INDEX idx_user_id (user_id),
    INDEX idx_order_status (order_status, created_at),
    INDEX idx_created_at (created_at),
    INDEX idx_supplier (supplier_id, supplier_order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单主表';
```

### 5.2 缓存架构（三级缓存）

| 层级 | 技术 | 容量 | TTL | 命中率 | 延迟 | 适用场景 |
|-----|------|------|-----|--------|------|---------|
| **L1本地缓存** | Ristretto | 500MB/实例 | 5分钟 | 80%+ | <1ms | 热点商品、配置 |
| **L2分布式缓存** | Redis Cluster | 1TB | 5-30分钟 | 95%+ | <5ms | 商品详情、价格、库存 |
| **L3数据库** | MySQL | - | - | - | 10-50ms | 权威数据源 |

**Redis数据结构设计**：

| 数据类型 | Redis结构 | Key示例 | TTL | 用途 |
|---------|-----------|---------|-----|------|
| 商品信息 | Hash | `product:sku:12345` | 30min | 商品详情 |
| 库存数据 | Hash | `inventory:sku:12345` | 5min | 库存实时查询 |
| 购物车 | Hash | `cart:user:67890` | 7天 | 购物车主存储 |
| 价格快照 | String | `snapshot:price:uuid-xxx` | 30min | 价格快照 |
| 库存预占 | String | `reserve:uuid-yyy` | 15min | 库存预占记录 |
| 预占索引 | ZSet | `reserve:expiry` | - | 按过期时间排序 |
| 幂等Key | String | `idempotent:checkout:token-zzz` | 5min | 防重复提交 |
| 限流计数 | String | `ratelimit:user:67890:checkout` | 1min | 用户限流 |

### 5.3 分布式事务（Saga模式）

**订单创建事务（跨3个服务）**：

```
[Checkout Service] - Saga协调者
    ↓
Step 1: 确认库存预占
    ├─ 调用 Inventory.ConfirmReserve(reserve_ids)
    ├─ 成功 → 继续
    └─ 失败 → 返回错误
    ↓
Step 2: 扣券
    ├─ 调用 Marketing.DeductCoupon(coupon_code)
    ├─ 成功 → 继续
    └─ 失败 → 补偿 Step 1（释放库存）
    ↓
Step 3: 创建订单
    ├─ 调用 Order.Create(order_data)
    ├─ 成功 → 返回 order_id
    └─ 失败 → 补偿 Step 1+2（释放库存+回退券）

补偿策略：
• 同步补偿：立即回滚（超时1s内）
• 异步补偿：写入Kafka补偿队列，定时任务重试
• 人工介入：3次重试失败，告警+人工处理
```

### 5.4 超卖防护

**核心品类（机票/酒店）- 零容忍**：
1. Redis预占（原子操作）
2. MySQL权威数据（15分钟后确认）
3. 供应商实时库存（下单时二次确认）
4. 定时对账（每5分钟）

**长尾品类（充值/礼品卡）- 可补偿**：
1. Redis预占
2. 真实超卖 → 人工补偿（补发券码、赔付现金、推荐替代商品）
3. 对账周期放宽（每小时）

### 5.5 幂等性设计

| 层级 | 方案 | 实现 |
|-----|------|------|
| **客户端幂等Token** | 前端生成UUID | Redis SET NX去重，5分钟TTL |
| **订单号唯一性** | 雪花算法 | DB唯一索引 `UNIQUE KEY (order_id)` |
| **支付回调幂等** | payment_id作为幂等Key | `UPDATE order SET status='PAID' WHERE order_id=? AND status='PENDING_PAYMENT'` |
| **营销扣减幂等** | 券扣减记录唯一索引 | `UNIQUE KEY (coupon_code, order_id)` + Redis原子操作 |

### 5.6 架构决策记录（ADR）

本节记录系统设计过程中的关键架构决策，包括决策背景、备选方案、最终决策及理由。

#### ADR-001: 计价中心数据输入方式

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：计价中心需要营销信息（促销规则、优惠券等）来计算最终价格，有两种方案：
- 方案1：计价中心自己调用Marketing Service获取营销信息
- 方案2：聚合服务获取营销信息后传递给计价中心

**决策**：采用**方案2**，由聚合服务获取营销信息后传递给计价中心。

**理由**：

1. **单一职责原则（SRP）**：
   - Pricing Service专注于价格计算逻辑（纯函数）
   - Aggregation Service负责数据编排和获取
   - 职责边界清晰，符合微服务设计原则

2. **依赖解耦**：
   ```
   方案1依赖链：Aggregation → Pricing → Marketing（传递性依赖）
   方案2依赖链：Aggregation → Pricing | Marketing（平行依赖）✓
   ```

3. **性能优化空间更大**：
   - 聚合层可以并发调用Marketing和其他服务（Product、Inventory）
   - Pricing变成纯计算，无IO等待
   - 减少网络调用层级（2层 vs 3层）

4. **易于测试**：
   ```go
   // 方案2：Pricing是纯函数，测试简单
   func TestCalculatePrice(t *testing.T) {
       priceItem := &PriceCalculateItem{
           SkuID:     1001,
           BasePrice: 2399.00,
           PromoInfo: &PromoInfo{DiscountRate: 0.9},  // Mock数据
       }
       result := pricingService.Calculate(priceItem)
       assert.Equal(t, 2159.10, result.FinalPrice)
   }
   ```

5. **统一降级处理**：
   - 聚合层统一处理各服务失败（Marketing、Product、Inventory）
   - Pricing Service无感知，始终收到完整输入数据
   - 降级逻辑不混入业务计算

6. **更好的可复用性**：
   - Pricing Service可被多个场景复用（搜索、详情、结算）
   - 输入参数标准化（base_price + promo_info）
   - 不依赖特定的服务调用链路

**代码示例**：

```go
// SearchOrchestrator（聚合服务）
func (o *SearchOrchestrator) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
    // Step 1: 获取sku_ids（从ES）
    // Step 2: 并发调用Product + Inventory
    // Step 3: 调用Marketing获取营销信息
    promos, err := o.marketingClient.BatchGetPromotions(ctx, skuIDs, req.UserID)
    if err != nil {
        promos = make(map[int64]*PromoInfo)  // 降级：空促销
    }
    
    // Step 4: 调用Pricing计算价格（传入营销信息）
    priceItems := buildPriceItems(basePriceMap, promos)  // 聚合层组装数据
    prices, err := o.pricingClient.BatchCalculatePrice(ctx, priceItems)
    
    return &SearchResponse{...}
}

// PricingService（计价中心）- 纯函数，只负责计算
func (s *PricingService) BatchCalculatePrice(ctx context.Context, items []*PriceItem) (map[int64]*Price, error) {
    results := make(map[int64]*Price)
    for _, item := range items {
        finalPrice := item.BasePrice  // 基础价格
        
        // 应用促销折扣（数据来自聚合层）
        if item.PromoInfo != nil {
            finalPrice = finalPrice * item.PromoInfo.DiscountRate
        }
        
        results[item.SkuID] = &Price{
            OriginalPrice: item.BasePrice,
            FinalPrice:    finalPrice,
            Discount:      item.BasePrice - finalPrice,
        }
    }
    return results, nil
}
```

**影响范围**：
- Aggregation Service：增加Marketing Service调用
- Pricing Service：接收PromoInfo作为输入参数
- Marketing Service：无影响

**后续行动**：
- ✓ 已实现：Aggregation Service编排逻辑
- ✓ 已实现：Pricing Service纯计算逻辑
- ✓ 已实现：Marketing Service RPC接口

---

#### ADR-002: 库存预占时机

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：在下单流程中，库存预占的时机有两种选择：
- 方案1：结算试算时预占（早期锁定）
- 方案2：确认下单时预占（延迟锁定）

**决策**：采用**方案2**，在确认下单时预占库存。

**理由**：

1. **减少无效预占**：
   - 用户在试算阶段可能多次修改商品、数量、优惠券
   - 早期预占会导致大量无效锁定（用户未真正下单）
   - 试算到下单的转化率通常只有20-30%

2. **提升库存利用率**：
   - 避免库存被长时间预占（用户可能犹豫、放弃）
   - 预占时长控制在15分钟内（支付超时自动释放）

3. **降低系统压力**：
   - 试算接口QPS高（用户多次试算），预占会导致Redis压力大
   - 确认下单QPS相对较低，预占操作更可控

4. **用户体验**：
   - 试算快速返回（不需要等待预占操作）
   - 确认下单时再预占，用户心理准备更充分

**权衡**：
- ✓ 优点：提升库存利用率、减少无效预占、降低系统压力
- ✗ 缺点：确认下单时可能库存不足（需要前端提示）

**降低缺点的措施**：
- 试算时展示实时库存状态（"仅剩N件"）
- 确认下单时二次校验库存，失败友好提示
- 热门商品提前告知"库存紧张，请尽快下单"

**代码示例**：

```go
// 试算接口：只查询库存，不预占
func (s *CheckoutService) Calculate(ctx context.Context, req *CalculateRequest) (*CalculateResponse, error) {
    // 查询库存状态（READ）
    stocks, _ := s.inventoryClient.BatchCheckStock(ctx, req.SkuIDs)
    
    // 计算价格...
    
    return &CalculateResponse{
        CanCheckout: allInStock(stocks, req.Items),  // 是否可下单
        Items:       items,
        Total:       total,
    }, nil
}

// 确认下单接口：预占库存
func (s *CheckoutService) Confirm(ctx context.Context, req *ConfirmRequest) (*ConfirmResponse, error) {
    // 预占库存（WRITE）
    reserveIDs, err := s.inventoryClient.ReserveStock(ctx, req.Items)
    if err != nil {
        return nil, fmt.Errorf("库存不足: %w", err)
    }
    
    // 创建订单...
    
    return &ConfirmResponse{OrderID: orderID}, nil
}
```

---

#### ADR-003: 聚合服务 vs BFF

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：在API Gateway和微服务之间，是使用BFF（Backend For Frontend）还是Aggregation Service？

**决策**：采用**Aggregation Service**，而不是传统BFF。

**理由**：

1. **业务导向 vs 端导向**：
   - BFF按端划分（Web BFF、App BFF、小程序 BFF）
   - Aggregation按业务场景划分（搜索聚合、详情聚合、结算聚合）✓
   - 本系统多个端（Web、App）的业务逻辑高度一致，按端拆分会导致重复代码

2. **代码复用**：
   ```
   BFF模式：
   ├─ Web BFF（搜索逻辑）
   ├─ App BFF（搜索逻辑）    ← 重复代码
   └─ 小程序 BFF（搜索逻辑） ← 重复代码
   
   Aggregation模式：✓
   ├─ Search Aggregation（Web/App/小程序共用）
   └─ Detail Aggregation（Web/App/小程序共用）
   ```

3. **维护成本**：
   - BFF需要维护多个端的代码一致性
   - Aggregation只需维护一套业务逻辑

4. **适配端差异的方式**：
   - API Gateway层处理端协议差异（HTTP、WebSocket、gRPC）
   - Aggregation返回标准数据格式，前端各端按需裁剪

**权衡**：
- ✓ 优点：减少重复代码、易于维护、业务逻辑统一
- ✗ 缺点：无法深度定制端特性（如App性能优化）

**适用场景**：
- ✓ 多端业务逻辑高度一致（如本系统）
- ✗ 不适用：各端业务逻辑差异大（如社交产品，Feed流算法不同）

---

#### ADR-004: 虚拟商品库存模型

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：虚拟商品（机票、充值卡、优惠券）的库存模型和实物商品差异大，应该如何设计？

**决策**：采用**二维库存模型**（ManagementType + UnitType）。

**库存管理类型（ManagementType）**：

| 类型 | 说明 | 典型品类 | 库存来源 |
|-----|------|---------|---------|
| **实时库存** | 强依赖供应商实时查询 | 机票、酒店 | 供应商API |
| **池化库存** | 自有库存，可超卖后补偿 | 充值卡、优惠券 | 平台采购 |
| **无限库存** | 虚拟商品，无库存限制 | SaaS服务、数字内容 | 无 |

**库存单位类型（UnitType）**：

| 类型 | 说明 | 典型品类 |
|-----|------|---------|
| **SKU级别** | 每个规格独立库存 | 充电器（颜色、规格） |
| **批次级别** | 按批次管理（有效期） | 优惠券、礼品卡 |
| **座位级别** | 唯一标识（座位号） | 机票、电影票 |

**理由**：
1. 不同品类的库存特性差异极大，无法用统一模型
2. 二维模型提供灵活性，支持策略模式动态选择
3. 便于扩展新品类（只需添加新策略）

**代码示例**：

```go
// 库存策略接口
type StockStrategy interface {
    Reserve(ctx context.Context, req *ReserveRequest) (*ReserveResponse, error)
    Deduct(ctx context.Context, req *DeductRequest) error
    Release(ctx context.Context, reserveID string) error
}

// 实时库存策略（机票、酒店）
type RealtimeStockStrategy struct {
    supplierClient rpc.SupplierClient
}

func (s *RealtimeStockStrategy) Reserve(ctx context.Context, req *ReserveRequest) (*ReserveResponse, error) {
    // 调用供应商实时预占接口
    return s.supplierClient.ReserveStock(ctx, req.SkuID, req.Quantity)
}

// 池化库存策略（充值卡、优惠券）
type PooledStockStrategy struct {
    redis redis.Client
}

func (s *PooledStockStrategy) Reserve(ctx context.Context, req *ReserveRequest) (*ReserveResponse, error) {
    // Redis原子扣减
    script := `
        local stock = redis.call('GET', KEYS[1])
        if tonumber(stock) >= tonumber(ARGV[1]) then
            redis.call('DECRBY', KEYS[1], ARGV[1])
            return 1
        else
            return 0
        end
    `
    return s.redis.Eval(ctx, script, []string{key}, req.Quantity).Result()
}
```

**影响范围**：
- Inventory Service：实现多种库存策略
- Supplier Gateway：对接供应商实时库存接口
- Product Center：商品配置中标记ManagementType

---

#### ADR-005: 同步 vs 异步数据流

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：下单流程中，哪些操作应该同步执行，哪些应该异步执行？

**决策**：采用**同步+异步混合模式**。

**同步操作（用户等待）**：
1. 库存预占（必须成功，否则无法下单）
2. 优惠券扣减（避免超发）
3. 订单创建（生成order_id）

**异步操作（Kafka事件）**：
1. 库存确认扣减（预占成功后，异步确认）
2. 搜索索引更新（销量、热度）
3. 购物车清理
4. 用户行为分析
5. 消息通知（订单确认、物流更新）

**理由**：

1. **用户体验**：
   - 同步操作<500ms，用户可接受
   - 非核心操作异步化，不阻塞下单

2. **系统解耦**：
   - 异步事件降低服务间强依赖
   - 消费者故障不影响下单流程

3. **性能优化**：
   - 减少下单接口响应时间
   - 异步操作可批量处理（提升吞吐）

4. **容错能力**：
   - 异步操作支持重试（Kafka消费者重试机制）
   - 同步操作失败可立即回滚（Saga模式）

**权衡**：
- ✓ 优点：高性能、高可用、易扩展
- ✗ 缺点：最终一致性（异步操作有延迟）

**一致性保障**：
- 订单状态机（状态流转确保一致性）
- 补偿机制（异步操作失败触发补偿）
- 定时对账（每小时对账，发现不一致）

---

#### ADR-006: 为什么引入聚合服务

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：为什么不直接在API Gateway调用多个微服务，而是引入聚合服务？

**决策**：引入**Aggregation Service**作为编排层。

**理由**：

1. **API Gateway职责单一**：
   - API Gateway只负责：鉴权、限流、路由、协议转换
   - 不应包含业务编排逻辑（违反SRP原则）

2. **复杂编排逻辑**：
   - 搜索场景：5步调用（ES → Product → Inventory → Marketing → Pricing）
   - 详情场景：6步调用（Product → 4个并发 → Pricing）
   - 结算场景：6步调用（Product → Inventory → Marketing → Pricing）
   - 如果放在Gateway，会导致Gateway代码膨胀

3. **数据依赖编排**：
   - 有些调用必须串行（Pricing依赖Marketing结果）
   - 有些调用可以并发（Product + Inventory）
   - 需要专门的编排层处理复杂依赖关系

4. **统一降级策略**：
   - 聚合层可以根据业务场景灵活降级
   - 搜索场景：Marketing失败可降级（只展示base_price）
   - 详情场景：Marketing失败不降级（必须返回错误）

5. **性能优化空间**：
   - 聚合层可以统一缓存聚合结果
   - 支持批量调用优化（BatchGet）
   - 支持超时控制和熔断

**架构对比**：

```
方案1（无聚合层）：
API Gateway → Product, Inventory, Marketing, Pricing（直接调用）
├─ Gateway需要处理复杂编排
├─ Gateway需要处理数据依赖
└─ Gateway代码膨胀，违反SRP

方案2（有聚合层）：✓
API Gateway → Aggregation → Product, Inventory, Marketing, Pricing
├─ Gateway职责单一（鉴权、限流、路由）
├─ Aggregation专注编排（数据获取、降级、聚合）
└─ 微服务职责单一（各司其职）
```

**影响范围**：
- 新增服务：Aggregation Service
- QPS估算：6000（正常）/ 30000（大促）
- 部署规模：6副本（正常）/ 18副本（大促）

---

#### ADR-007: MySQL分库分表策略

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：订单表、商品表数据量大（千万级），如何分库分表？

**决策**：
- **订单表**：按`user_id`分库（8库），按`create_time`分表（64表）
- **商品表**：按`category_id`分库（4库），不分表

**理由**：

**订单表分库策略**：
1. 按`user_id`分库：
   - ✓ 用户维度查询最频繁（"我的订单"）
   - ✓ 避免跨库查询，性能最优
   - ✗ 但按订单号查询会跨库（通过路由表解决）

2. 按`create_time`分表：
   - ✓ 历史订单按月归档（避免单表过大）
   - ✓ 查询"我的订单"时按时间范围查询（只查最近几个月）

**商品表分库策略**：
1. 按`category_id`分库：
   - ✓ 分类浏览场景性能最优
   - ✓ 不同品类的供应商同步逻辑隔离
   - ✗ 跨品类搜索需要聚合（通过ES解决）

2. 不分表：
   - 商品数据量可控（百万级），单表足够
   - 避免过度设计

**路由表设计**：

```sql
-- 订单路由表（解决按order_id查询的跨库问题）
CREATE TABLE order_route (
    order_id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    db_index TINYINT NOT NULL,  -- 库索引（0-7）
    table_index TINYINT NOT NULL,  -- 表索引（0-63）
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB;
```

**查询示例**：

```go
// 按order_id查询（需要路由表）
func (r *OrderRepo) GetByOrderID(orderID int64) (*Order, error) {
    // Step 1: 查询路由表
    route, err := r.routeTable.GetRoute(orderID)
    if err != nil {
        return nil, err
    }
    
    // Step 2: 根据路由信息查询目标分片
    db := r.shards[route.DBIndex]
    tableName := fmt.Sprintf("order_%d", route.TableIndex)
    
    return db.QueryOne("SELECT * FROM ? WHERE order_id = ?", tableName, orderID)
}

// 按user_id查询（不需要路由表）
func (r *OrderRepo) GetByUserID(userID int64, limit int) ([]*Order, error) {
    dbIndex := userID % 8  // 直接计算库索引
    db := r.shards[dbIndex]
    
    // 查询最近3个月的订单（最多3张表）
    tables := r.getRecentTables(3)  // ["order_62", "order_63", "order_0"]
    
    // 并发查询3张表，合并结果
    return db.QueryMultipleTables(tables, "SELECT * FROM ? WHERE user_id = ? ORDER BY create_time DESC LIMIT ?", userID, limit)
}
```

---

#### ADR-008: 试算接口是否复用详情页缓存数据

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：用户从商品详情页点击"结算"进入试算接口，能否复用详情页已缓存的商品信息（尤其是营销数据）来减少后端调用，还是需要重新查询Marketing Service？

**决策**：采用"快照ID（snapshot_id）+ 最终创单二次校验"方案

**核心设计思想**：
```
试算阶段：允许使用快照数据（性能优先）
创单阶段：强制实时校验（准确性优先）
```

**用户行为路径**：
```
详情页（生成快照）→ 点击结算（通常30秒内）→ 试算（使用快照）→ 确认下单（实时校验）
```

**三种方案对比**：

| 方案 | 实现方式 | 响应时间 | 数据准确性 | 后端压力 | 推荐 |
|-----|---------|---------|-----------|---------|-----|
| **A. 完全复用（无快照）** | 前端缓存→直接使用 | 50ms | ❌ 低（营销可能变化） | ✓ 最低 | ❌ |
| **B. 快照ID + 过期判断**✓ | 快照未过期→使用缓存<br>快照过期→重新查询 | 80ms | ✓ 高（最终创单二次校验） | ✓ 低 | ✅ |
| **C. 完全不复用** | 试算也重新查询所有服务 | 230ms | ✓ 最高 | ❌ 高 | ❌ |

**理由**：

**1. 各类数据的复用可行性分析**：

| 数据类型 | 变化频率 | 是否可复用 | 理由 |
|---------|---------|-----------|------|
| 商品基础信息（title/images） | 低（小时级） | ✅ 可复用 | 不常变化，复用安全 |
| 基础价格（base_price） | 低（天级） | ✅ 可复用 | 价格变化慢，可缓存 |
| 营销活动（promotions） | 中（分钟级） | ⚠️ 条件复用 | 快照5分钟有效期，创单时二次校验 |
| 库存状态（stock） | 高（秒级） | ❌ 必须实时查 | 防止超卖，不能使用缓存 |

**2. 性能优化显著**：
- 响应时间降低65%（80ms vs 230ms）
- Marketing Service QPS降低80%（命中率80%）
- 用户连续试算3次场景：240ms vs 690ms

**3. 数据一致性保证（关键）**：
- ✅ **试算阶段**：允许使用5分钟内的快照数据（性能优先，用户体验好）
- ✅ **创单阶段**：强制实时查询+校验营销活动（准确性优先，防止资损）
- ✅ 快照过期自动降级到重新查询（透明对用户）
- ✅ 最终创单时的二次校验是最后一道防线（核心安全保障）

**实现方案**：

**前端实现（传递快照数据）**：

```javascript
// 1. 详情页响应（附带快照ID和过期时间）
GET /products/12345/detail
{
  "sku_id": 12345,
  "price": {
    "base_price": 299.00,
    "final_price": 269.00
  },
  "promotions": {
    "active_promotions": [...],
    "coupons": [...]
  },
  "snapshot": {
    "snapshot_id": "snap:12345:1744633200",  // 快照ID（唯一标识）
    "created_at": 1744633200,                // 快照创建时间
    "expires_at": 1744633500,                // 快照过期时间（5分钟后）
    "ttl": 300                               // 快照有效期（秒）
  }
}

// 2. 前端存储详情页数据（包含快照信息）
const detailData = {
  sku_id: 12345,
  base_price: 299.00,
  promotions: {...},
  snapshot_id: "snap:12345:1744633200",
  snapshot_expires_at: 1744633500
};
localStorage.setItem('product_12345', JSON.stringify(detailData));

// 3. 试算接口（携带快照数据）
POST /checkout/calculate
{
  "user_id": 67890,
  "items": [
    {
      "sku_id": 1001,
      "quantity": 2,
      "snapshot": {                          // 前端缓存的快照数据
        "snapshot_id": "snap:1001:1744633200",
        "expires_at": 1744633500,
        "data": {
          "base_price": 299.00,
          "promotions": {...}
        }
      }
    },
    {
      "sku_id": 1005,
      "quantity": 1,
      "snapshot": null                       // 无详情页快照（需要查询）
    }
  ]
}
```

**后端实现（试算接口：使用快照数据）**：

```go
// CheckoutService.Calculate - 试算接口（性能优先）
func (s *CheckoutService) Calculate(ctx context.Context, req *CalculateRequest) (*CalculateResponse, error) {
    var (
        needQuerySKUs   []int64                       // 需要重新查询的SKU
        snapshotData    map[int64]*SnapshotData       // 可复用的快照数据
    )
    
    now := time.Now().Unix()
    
    // Step 1: 判断每个SKU的快照是否过期
    for _, item := range req.Items {
        if item.Snapshot != nil {
            // 快照未过期，直接使用
            if item.Snapshot.ExpiresAt > now {
                snapshotData[item.SkuID] = item.Snapshot.Data
                continue
            }
            // 快照已过期，需要重新查询
        }
        
        // 无快照或快照过期，需要查询
        needQuerySKUs = append(needQuerySKUs, item.SkuID)
    }
    
    // Step 2: 只查询未命中快照的SKU
    var products []*Product
    var promos []*Promotion
    if len(needQuerySKUs) > 0 {
        // 并发调用Product + Marketing
        products, _ = s.productClient.BatchGetProducts(ctx, needQuerySKUs)
        promos, _ = s.marketingClient.BatchGetPromotions(ctx, needQuerySKUs, req.UserID)
    }
    
    // Step 3: 合并快照数据和查询数据
    allProducts := s.mergeSnapshotAndQueried(snapshotData, products)
    
    // Step 4: 库存必须实时查询（不能使用快照）
    allSKUs := extractAllSKUs(req.Items)
    stocks, _ := s.inventoryClient.BatchCheckStock(ctx, allSKUs)
    
    // Step 5: 调用Pricing计算价格
    prices, _ := s.pricingClient.BatchCalculatePrice(ctx, allProducts)
    
    return &CalculateResponse{
        Items:         buildCalculateItems(allProducts, prices, stocks),
        TotalPrice:    calculateTotal(prices),
        SnapshotUsage: buildSnapshotUsageReport(snapshotData, needQuerySKUs), // 快照使用情况
    }, nil
}

// 合并快照数据和查询数据
func (s *CheckoutService) mergeSnapshotAndQueried(
    snapshotData map[int64]*SnapshotData,
    queriedProducts []*Product,
) []*ProductData {
    merged := make([]*ProductData, 0)
    
    // 添加快照数据
    for skuID, snapshot := range snapshotData {
        merged = append(merged, &ProductData{
            SkuID:      skuID,
            BasePrice:  snapshot.BasePrice,
            Promotions: snapshot.Promotions,
            Source:     "snapshot", // 标记数据来源
        })
    }
    
    // 添加查询数据
    for _, product := range queriedProducts {
        merged = append(merged, &ProductData{
            SkuID:      product.SkuID,
            BasePrice:  product.BasePrice,
            Promotions: product.Promotions,
            Source:     "realtime", // 标记数据来源
        })
    }
    
    return merged
}
```

**后端实现（确认下单接口：强制实时校验）**：

```go
// OrderService.CreateOrder - 确认下单接口（准确性优先）
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
    // 注意：确认下单时，不使用任何快照数据，全部实时查询
    
    // Step 1: 实时查询商品信息（不使用快照）
    products, err := s.productClient.BatchGetProducts(ctx, req.SkuIDs)
    if err != nil {
        return nil, fmt.Errorf("query products failed: %w", err)
    }
    
    // Step 2: 实时查询营销活动（强制最新数据）
    promos, err := s.marketingClient.BatchGetPromotions(ctx, req.SkuIDs, req.UserID)
    if err != nil {
        return nil, fmt.Errorf("query promotions failed: %w", err)
    }
    
    // Step 3: 校验营销活动有效性（关键：防止使用过期活动）
    for _, promo := range promos {
        if !s.validatePromotion(promo) {
            return nil, fmt.Errorf("promotion %s is invalid or expired", promo.ID)
        }
    }
    
    // Step 4: 预占库存（CAS操作，防止超卖）
    reserved, err := s.inventoryClient.ReserveStock(ctx, req.Items)
    if err != nil {
        return nil, fmt.Errorf("reserve stock failed: %w", err)
    }
    
    // Step 5: 实时计算价格（基于最新营销数据）
    price, err := s.pricingClient.CalculateFinalPrice(ctx, products, promos)
    if err != nil {
        // 回滚库存
        s.inventoryClient.ReleaseStock(ctx, reserved)
        return nil, fmt.Errorf("calculate price failed: %w", err)
    }
    
    // Step 6: 创建订单
    order := &Order{
        OrderID:    s.generateOrderID(),
        UserID:     req.UserID,
        Items:      req.Items,
        TotalPrice: price.FinalPrice,
        Status:     OrderStatusPending,
    }
    
    if err := s.orderRepo.Create(ctx, order); err != nil {
        // 回滚库存
        s.inventoryClient.ReleaseStock(ctx, reserved)
        return nil, fmt.Errorf("create order failed: %w", err)
    }
    
    return &CreateOrderResponse{
        OrderID:    order.OrderID,
        TotalPrice: price.FinalPrice,
    }, nil
}

// 校验营销活动有效性
func (s *OrderService) validatePromotion(promo *Promotion) bool {
    now := time.Now()
    
    // 1. 检查时间范围
    if now.Before(promo.StartTime) || now.After(promo.EndTime) {
        return false
    }
    
    // 2. 检查库存（如果是限量活动）
    if promo.StockLimit > 0 && promo.StockRemaining <= 0 {
        return false
    }
    
    // 3. 检查用户参与次数（如果有限制）
    if promo.UserLimit > 0 && promo.UserUsedCount >= promo.UserLimit {
        return false
    }
    
    return true
}
```

**快照机制设计**：

```
快照ID生成规则：
  snapshot_id = "snap:{sku_id}:{timestamp}"
  expires_at = created_at + 300（5分钟）

快照特性：
  1. 无需后端存储（前端自带快照数据）
  2. 简单过期判断（expires_at > now）
  3. 快照过期自动降级到实时查询
  4. 创单时强制实时校验（不使用快照）

快照生命周期：
  1. 详情页访问 → 生成快照（Aggregation Service）
  2. 点击结算 → 携带快照（前端传递）
  3. 试算计算 → 判断过期（Checkout Service）
     - 未过期：直接使用快照数据
     - 已过期：重新查询Marketing Service
  4. 确认下单 → 强制实时校验（Order Service）
     - 重新查询所有营销活动
     - 校验活动有效性（时间、库存、用户限制）
     - 防止使用过期/失效的营销活动
```

**试算 vs 创单的数据要求对比（关键）**：

| 维度 | 试算阶段（Calculate） | 创单阶段（CreateOrder） |
|-----|---------------------|----------------------|
| **目的** | 性能优先，快速预览价格 | 准确性优先，防止资损 |
| **商品数据** | 可使用快照（5分钟） | ✅ 强制实时查询 |
| **营销数据** | 可使用快照（5分钟） | ✅ 强制实时查询 + 活动有效性校验 |
| **库存数据** | ✅ 必须实时查询 | ✅ 必须实时查询 + CAS预占 |
| **价格计算** | 基于快照/实时混合数据 | ✅ 基于最新实时数据 |
| **数据一致性** | 最终一致性（允许5分钟延迟） | 强一致性（实时校验） |
| **性能要求** | P95 < 300ms | P95 < 500ms（可接受稍慢） |
| **安全保障** | 无资损风险（仅展示） | ✅ 多重校验（防止资损） |

**关键设计原则**：
- ✅ 试算允许使用快照（提升性能，用户体验好）
- ✅ 创单强制实时校验（保证准确性，防止资损）
- ✅ 快照过期自动降级（透明对用户）
- ✅ 最终创单是最后一道防线（即使试算用了过期快照，创单也会拦截）

**性能提升数据**：

**场景1：用户在详情页停留30秒后点击结算（快照未过期）**

| 指标 | 方案A（完全复用） | 方案B（快照ID）✓ | 方案C（不复用） |
|-----|-----------------|-----------------|----------------|
| 查询服务数 | 0个 | 1个（Inventory） | 3个（Product+Marketing+Inventory） |
| 响应时间 | 50ms | 80ms | 230ms |
| 快照命中率 | 100%（风险高） | 90%（安全） | 0% |
| 数据准确性 | ❌ 无保障 | ✅ 创单时二次校验 | ✅ 实时数据 |

**场景2：用户在详情页停留10分钟后点击结算（快照已过期）**

| 指标 | 方案B（快照ID）✓ | 方案C（不复用） |
|-----|-----------------|----------------|
| 查询服务数 | 2个（Product+Marketing） | 3个（Product+Marketing+Inventory） |
| 响应时间 | 180ms | 230ms |
| 快照命中率 | 0%（自动降级） | 0% |

**场景3：用户连续试算3次（调整数量、优惠券，快照未过期）**

| 指标 | 方案A | 方案B（快照ID）✓ | 方案C |
|-----|-------|-----------------|-------|
| 总查询次数 | 0次 | 3次（只查Inventory） | 9次 |
| 总响应时间 | 150ms | 240ms | 690ms |
| Marketing QPS | 0 | 0（使用快照） | 3 |

**权衡**：

| 维度 | 优势 | 劣势 |
|-----|------|------|
| **性能** | ✓ 快照命中时响应快65%（80ms vs 230ms）<br>✓ Marketing Service QPS降低90% | ⚠️ 快照过期时需重查（约10%场景） |
| **一致性** | ✓ 最终创单强制实时校验（零资损风险）<br>✓ 快照过期自动降级（透明） | ⚠️ 试算阶段允许5分钟延迟（可接受） |
| **复杂度** | ✓ 实现简单（无需Redis token）<br>✓ 前端存储快照，后端仅判断过期 | ⚠️ 需要前后端配合传递快照数据 |
| **安全性** | ✓ 创单时多重校验（活动有效性+库存）<br>✓ 试算用快照不影响最终准确性 | ✅ 无安全风险（创单是最后防线） |

**核心优势（相比cache_token方案）**：
1. ✅ **无需后端存储**：快照数据由前端携带，后端无需维护Redis token
2. ✅ **实现更简单**：仅需判断`expires_at > now`，无需复杂的token验证
3. ✅ **营销活动变更无影响**：无需监听营销变更事件更新token
4. ✅ **最终创单二次校验**：即使快照数据错误，创单阶段也会拦截

**影响范围**：
- **Aggregation Service**：商品详情接口返回`snapshot`字段（snapshot_id + expires_at + data）
- **Checkout Service**：
  - `Calculate`接口：判断快照是否过期，未过期则使用快照数据
  - `CreateOrder`接口：强制实时查询，不使用快照
- **Order Service**：确认下单时实时查询+校验营销活动有效性
- **前端（APP/Web）**：缓存详情页快照数据，试算时携带`snapshot`字段

**实施建议**：
1. **第一阶段**：试算接口支持快照数据（快照过期降级到实时查询）
2. **第二阶段**：优化快照TTL（根据用户行为分析，可能调整为3-10分钟）
3. **第三阶段**：增加快照命中率监控，优化用户体验

**监控指标**：
- **快照命中率**（目标90%，即90%用户在5分钟内点击结算）
- **试算接口P99响应时间**（目标<300ms）
- **创单接口营销校验失败率**（目标<1%，即99%的试算价格与创单价格一致）
- **Marketing Service QPS**（目标降低90%）

**关键设计亮点**：
> "试算用快照（性能优先），创单强制校验（准确性优先）"
> 
> 这是一个典型的"防御性设计"：即使试算阶段使用了过期快照，最终创单时的实时校验会拦截所有不一致的情况，用户最终支付的价格一定是准确的。

---

#### ADR-009: 创单与支付的时序关系（先创单后支付 vs 创单即支付）

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：在订单流程中，"创建订单"和"支付"这两个动作的时序关系有两种模式：
1. **创单即支付**：用户点击"立即购买"后，先支付，支付成功后再创建订单
2. **先创单后支付**：用户点击"提交订单"后，先创建订单（资源扣减），然后再选择支付方式、优惠券等，最后支付

**决策**：采用"先创单后支付"模式

**两种模式对比**：

| 维度 | 创单即支付（模式A） | 先创单后支付（模式B）✓ |
|-----|------------------|---------------------|
| **用户体验** | ⚠️ 需要先选择支付方式才能下单 | ✅ 先锁定库存，再慢慢支付 |
| **库存扣减时机** | 支付成功后扣减 | 创单时预占，支付成功后确认扣减 |
| **价格计算时机** | 支付前一次性计算 | 创单时计算基础价格，支付时计算支付渠道费 |
| **优惠券使用** | 支付前选择 | 创单时或支付时选择（更灵活） |
| **订单状态** | 仅两种：未支付、已支付 | 三种：待支付、已支付、已完成 |
| **超时未支付** | 不存在（支付后才创单） | 需要处理（释放预占库存） |
| **资损风险** | ⚠️ 高（库存未锁定，可能超卖） | ✅ 低（创单时锁定库存） |
| **复杂度** | 低 | 中（需要处理预占、超时释放） |

**理由**：

**1. 用户体验更好**：
- ✅ 用户点击"提交订单"后，订单立即生成，库存被锁定
- ✅ 用户可以慢慢选择支付方式（支付宝、微信、银行卡）
- ✅ 用户可以在支付环节选择优惠券、支付渠道优惠（如花呗立减）
- ✅ 用户可以先下单占位，稍后再支付（适合机票、酒店等场景）

**2. 防止超卖（关键）**：
```
【创单即支付模式的问题】：
1. 用户A看到库存=1
2. 用户B也看到库存=1
3. 用户A点击支付（此时库存未扣减）
4. 用户B也点击支付（库存仍未扣减）
5. 两人同时支付成功 → 超卖！

【先创单后支付模式的解决方案】：
1. 用户A点击"提交订单" → 库存预占：1 → 0（剩余可用）
2. 用户B点击"提交订单" → 库存不足，下单失败
3. 用户A有15分钟支付窗口
4. 如果用户A超时未支付 → 释放库存：0 → 1（其他人可下单）
```

**3. 价格计算灵活性**：
- 创单时计算：商品基础价格 + 营销优惠（折扣、满减）
- 支付时计算：支付渠道费（信用卡手续费、花呗分期费）+ 支付渠道优惠（立减）

**实现方案**：

**订单状态机设计**：

```
┌─────────────────────────────────────────────────────────────────┐
│                        订单状态流转                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  [1] PENDING_PAYMENT（待支付）                                  │
│      ↓ 创建订单时的初始状态                                    │
│      ↓ 库存已预占（15分钟TTL）                                 │
│      ↓ 优惠券可以在此阶段选择                                  │
│      ├───→ [超时未支付] → [CANCELLED]                         │
│      │      ↓ 释放库存                                         │
│      │      ↓ 回退优惠券                                       │
│      │                                                         │
│      └───→ [用户支付成功]                                      │
│             ↓                                                   │
│  [2] PAID（已支付）                                            │
│      ↓ 支付成功                                                │
│      ↓ 库存从"预占"转为"确认扣减"                              │
│      ↓ 优惠券从"预扣"转为"确认扣减"                            │
│      ↓ 发起供应商履约                                          │
│      ├───→ [供应商履约失败] → [REFUNDING]                     │
│      │      ↓ 发起退款                                         │
│      │                                                         │
│      └───→ [供应商履约成功]                                    │
│             ↓                                                   │
│  [3] COMPLETED（已完成）                                       │
│      ↓ 供应商出票/出码成功                                     │
│      ↓ 发送凭证给用户                                          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Phase 2: 确认下单（创建订单，不支付）**：

```go
// CheckoutService.ConfirmCheckout - 确认下单（不支付）
func (s *CheckoutService) ConfirmCheckout(ctx context.Context, req *ConfirmCheckoutRequest) (*ConfirmCheckoutResponse, error) {
    // Step 1: 实时查询商品和营销信息（不使用快照）
    products, _ := s.productClient.BatchGetProducts(ctx, req.SkuIDs)
    promos, _ := s.marketingClient.BatchGetPromotions(ctx, req.SkuIDs, req.UserID)
    
    // Step 2: 库存预占（Redis Lua原子操作，15分钟TTL）
    reserved, err := s.inventoryClient.ReserveStock(ctx, req.Items)
    if err != nil {
        return nil, fmt.Errorf("库存不足: %w", err)
    }
    
    // Step 3: 计算订单金额（基础价格 + 营销优惠，不含支付渠道费）
    basePrice, err := s.pricingClient.CalculateBasePrice(ctx, products, promos)
    if err != nil {
        s.inventoryClient.ReleaseStock(ctx, reserved) // 回滚库存
        return nil, fmt.Errorf("价格计算失败: %w", err)
    }
    
    // Step 4: 创建订单（状态：PENDING_PAYMENT）
    order := &Order{
        OrderID:       s.generateOrderID(),
        UserID:        req.UserID,
        Items:         req.Items,
        BasePrice:     basePrice.Amount,        // 商品总价
        DiscountPrice: basePrice.Discount,      // 营销优惠
        Status:        OrderStatusPendingPayment,
        PayExpireAt:   time.Now().Add(15 * time.Minute), // 15分钟支付窗口
        ReserveIDs:    reserved,                // 库存预占ID
    }
    
    if err := s.orderRepo.Create(ctx, order); err != nil {
        s.inventoryClient.ReleaseStock(ctx, reserved) // 回滚库存
        return nil, fmt.Errorf("创建订单失败: %w", err)
    }
    
    // Step 5: 发布 order.created 事件（异步）
    s.publishOrderCreatedEvent(order)
    
    return &ConfirmCheckoutResponse{
        OrderID:       order.OrderID,
        BasePrice:     basePrice.Amount,
        DiscountPrice: basePrice.Discount,
        AmountToPay:   basePrice.Amount - basePrice.Discount, // 待支付金额（未含支付渠道费）
        PayExpireAt:   order.PayExpireAt,
    }, nil
}
```

**Phase 3: 支付（计算支付渠道费 + 执行支付）**：

```go
// PaymentService.CreatePayment - 发起支付
func (s *PaymentService) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
    // Step 1: 查询订单（必须是待支付状态）
    order, err := s.orderClient.GetOrder(ctx, req.OrderID)
    if err != nil {
        return nil, err
    }
    if order.Status != OrderStatusPendingPayment {
        return nil, fmt.Errorf("订单状态不正确: %s", order.Status)
    }
    if time.Now().After(order.PayExpireAt) {
        return nil, fmt.Errorf("订单已过期")
    }
    
    // Step 2: 选择/校验优惠券（如果用户在支付时选择）
    var couponDiscount float64
    if req.CouponID != "" {
        coupon, err := s.marketingClient.ValidateCoupon(ctx, req.CouponID, req.UserID)
        if err != nil {
            return nil, fmt.Errorf("优惠券无效: %w", err)
        }
        couponDiscount = coupon.Amount
        
        // 预扣优惠券（支付成功后确认扣减）
        if err := s.marketingClient.ReserveCoupon(ctx, req.CouponID, req.UserID); err != nil {
            return nil, fmt.Errorf("优惠券预扣失败: %w", err)
        }
    }
    
    // Step 3: 计算支付渠道费（手续费、分期费等）
    channelFee := s.calculateChannelFee(req.PaymentChannel, order.AmountToPay)
    
    // Step 4: 计算最终支付金额
    finalAmount := order.AmountToPay - couponDiscount + channelFee
    
    // Step 5: 创建支付记录
    payment := &Payment{
        PaymentID:      s.generatePaymentID(),
        OrderID:        req.OrderID,
        UserID:         req.UserID,
        PaymentChannel: req.PaymentChannel,
        BaseAmount:     order.AmountToPay,     // 订单金额
        CouponDiscount: couponDiscount,        // 优惠券抵扣
        ChannelFee:     channelFee,            // 支付渠道费
        FinalAmount:    finalAmount,           // 最终支付金额
        Status:         PaymentStatusPending,
    }
    
    if err := s.paymentRepo.Create(ctx, payment); err != nil {
        // 回滚优惠券
        if req.CouponID != "" {
            s.marketingClient.ReleaseCoupon(ctx, req.CouponID, req.UserID)
        }
        return nil, err
    }
    
    // Step 6: 调用支付网关（支付宝、微信等）
    payURL, err := s.paymentGateway.CreatePay(ctx, payment)
    if err != nil {
        return nil, err
    }
    
    return &CreatePaymentResponse{
        PaymentID:   payment.PaymentID,
        PayURL:      payURL,
        FinalAmount: finalAmount,
    }, nil
}

// 支付回调（支付宝/微信异步通知）
func (s *PaymentService) PaymentCallback(ctx context.Context, req *PaymentCallbackRequest) error {
    // Step 1: 幂等性校验
    if s.isDuplicate(req.PaymentID) {
        return nil // 重复回调，直接返回成功
    }
    
    // Step 2: 更新支付状态
    payment, _ := s.paymentRepo.GetByID(ctx, req.PaymentID)
    payment.Status = PaymentStatusPaid
    payment.PaidAt = time.Now()
    s.paymentRepo.Update(ctx, payment)
    
    // Step 3: 更新订单状态：PENDING_PAYMENT → PAID
    s.orderClient.UpdateOrderStatus(ctx, payment.OrderID, OrderStatusPaid)
    
    // Step 4: 确认扣减库存（从预占转为实际扣减）
    s.inventoryClient.ConfirmReserve(ctx, payment.OrderID)
    
    // Step 5: 确认扣减优惠券
    if payment.CouponDiscount > 0 {
        s.marketingClient.ConfirmCoupon(ctx, payment.CouponID, payment.UserID)
    }
    
    // Step 6: 发布 payment.paid 事件
    s.publishPaymentPaidEvent(payment)
    
    return nil
}
```

**超时未支付处理（定时任务）**：

```go
// OrderTimeoutJob - 定时扫描超时未支付订单
func (j *OrderTimeoutJob) Run() {
    // 查询超时未支付订单（创建时间 > 15分钟，状态=PENDING_PAYMENT）
    expiredOrders := j.orderRepo.FindExpiredPendingPayment(15 * time.Minute)
    
    for _, order := range expiredOrders {
        // 1. 更新订单状态：PENDING_PAYMENT → CANCELLED
        order.Status = OrderStatusCancelled
        order.CancelReason = "超时未支付"
        j.orderRepo.Update(ctx, order)
        
        // 2. 释放库存（从预占状态释放）
        j.inventoryClient.ReleaseStock(ctx, order.ReserveIDs)
        
        // 3. 回退优惠券（如果有）
        if order.CouponID != "" {
            j.marketingClient.ReleaseCoupon(ctx, order.CouponID, order.UserID)
        }
        
        // 4. 发布 order.cancelled 事件
        j.publishOrderCancelledEvent(order)
    }
}
```

**权衡**：

| 维度 | 优势 | 劣势 |
|-----|------|------|
| **用户体验** | ✅ 先锁定库存，再支付<br>✅ 支付环节更灵活 | ⚠️ 15分钟内库存被占用 |
| **防止超卖** | ✅ 创单时锁定库存（零超卖） | ⚠️ 需要处理超时释放逻辑 |
| **价格灵活性** | ✅ 支付时可选优惠券、计算渠道费 | ⚠️ 支付时价格可能变化（需提示） |
| **系统复杂度** | ⚠️ 需要库存预占机制<br>⚠️ 需要超时释放定时任务 | ⚠️ 状态机更复杂（3种状态） |
| **库存利用率** | ⚠️ 预占库存可能被浪费（10-20%未支付率） | ✅ 可通过缩短支付窗口优化 |

**适用场景分析**：

| 场景 | 推荐模式 | 理由 |
|-----|---------|------|
| **高并发秒杀商品** | ✅ 先创单后支付 | 防止超卖，库存锁定是硬需求 |
| **机票、酒店预订** | ✅ 先创单后支付 | 用户需要时间确认行程、选支付方式 |
| **充值、话费** | 可选创单即支付 | 无库存限制，支付即充值 |
| **虚拟商品（券码）** | ✅ 先创单后支付 | 库存有限，需要锁定 |
| **线下优惠券** | ✅ 先创单后支付 | 优惠券数量有限，需要锁定 |

**影响范围**：
- **Order Service**：订单状态机增加 PENDING_PAYMENT 状态
- **Inventory Service**：增加库存预占（Reserve）和确认扣减（Confirm）接口
- **Marketing Service**：增加优惠券预扣（Reserve）和确认扣减（Confirm）接口
- **Payment Service**：支付时重新计算优惠券和渠道费
- **定时任务**：新增超时未支付订单扫描任务

**监控指标**：
- **未支付率**（目标<20%）：`未支付订单数 / 总创建订单数`
- **库存预占浪费率**（目标<15%）：`超时释放库存数 / 总预占库存数`
- **支付超时率**（目标<5%）：`超过15分钟未支付订单数 / 总创建订单数`
- **库存预占成功率**（目标>99%）：`库存预占成功数 / 库存预占请求数`

**关键设计亮点**：
> "先创单锁定资源，再支付执行扣减"
> 
> 这是一个典型的"预占-确认"两阶段提交模式（2PC思想），既保证了库存不超卖，又给用户留出了充足的支付选择时间，是电商高并发场景的标准做法。

---

## 六、部署架构（同城双活）

### 6.1 整体拓扑

```
                       互联网用户流量
                            ↓
              ┌─────────────────────────────┐
              │    全局DNS（GeoDNS）        │
              │  • 智能解析（就近接入）      │
              │  • 健康检查（故障切换）      │
              └─────────────────────────────┘
                     ↓          ↓
      ┌──────────────┴──────────┴──────────────┐
      ↓                                         ↓
┌─────────────────────────┐      ┌─────────────────────────┐
│   IDC-A（主机房）        │      │   IDC-B（备机房）        │
│   同城10km内            │◄─────│   同城10km内            │
│                         │ 双向 │                         │
│  • K8s Cluster (3M+50W) │ 同步 │  • K8s Cluster (3M+50W) │
│  • MySQL 主库（写）      │◄────►│  • MySQL 从库（读）      │
│  • Redis Cluster (8主8从)│      │  • Redis Cluster (8主8从)│
│  • Kafka (6 Broker)     │◄MM2─►│  • Kafka (6 Broker)     │
│  • Elasticsearch (6节点) │      │  • Elasticsearch (6节点) │
│                         │      │                         │
│  流量占比：60%          │      │  流量占比：40%          │
└─────────────────────────┘      └─────────────────────────┘
       网络延迟：< 2ms（专线连接）
```

### 6.2 MySQL双主部署

```
IDC-A                                 IDC-B
┌─────────────────────┐              ┌─────────────────────┐
│  MySQL Master A     │◄────双向────►│  MySQL Master B     │
│  • server_id: 1     │   binlog同步 │  • server_id: 2     │
│  • auto_increment_  │              │  • auto_increment_  │
│    offset: 1        │              │    offset: 2        │
│  • auto_increment_  │              │  • auto_increment_  │
│    increment: 2     │              │    increment: 2     │
│  • 承载60%写流量    │              │  • 承载40%写流量    │
└─────────────────────┘              └─────────────────────┘
         ↓                                     ↓
┌─────────────────────┐              ┌─────────────────────┐
│  MySQL Slave A1/A2  │              │  MySQL Slave B1/B2  │
│  • 只读（Read）     │              │  • 只读（Read）     │
│  • 延迟监控<1s      │              │  • 延迟监控<1s      │
└─────────────────────┘              └─────────────────────┘
```

**关键配置**：
- 半同步复制：保证数据不丢失（1秒超时降级为异步）
- GTID模式：简化主从切换
- 自增ID错开：避免主键冲突（offset+increment）

### 6.3 容灾切换SOP

**场景1：IDC-A计划性维护**
```
T-60min: 提前通知，确认IDC-B容量充足
T-30min: 检查双机房数据同步状态
T-15min: DNS权重调整 A:60%→30%，B:40%→70%
T-10min: 继续调整 A:30%→0%，B:70%→100%
T-5min:  确认IDC-A无流量，停止服务
T0:      开始维护
T+Xmin:  维护完成，逆向切回流量
```

**场景2：IDC-A突发故障**
```
T0:      监控发现IDC-A全量服务不可用
T+30s:   告警触发，呼叫值班人员
T+2min:  人工确认故障范围，决策切换
T+3min:  执行DNS强制切换：A:0%，B:100%
T+5min:  确认IDC-B承载全量流量，监控关键指标
T+10min: 检查数据一致性，启动补偿任务
```

---

## 七、稳定性保障体系

### 7.1 监控体系（四层监控）

| 层级 | 监控内容 | 工具 | 告警阈值示例 |
|-----|---------|------|-------------|
| **L1基础设施** | CPU、内存、磁盘、网络、数据库连接数、慢查询 | Node Exporter + cAdvisor | CPU>80%持续5分钟→P2 |
| **L2应用监控** | QPS、错误率、延迟(P50/P95/P99)、资源使用率 | Prometheus + Istio | 错误率>1%持续2分钟→P1 |
| **L3业务监控** | 下单量、支付率、库存充足率、供应商可用率 | 自定义埋点 | 下单量同比下降30%→P1 |
| **L4用户体验** | FCP、LCP、FID、CLS、前端错误率 | Sentry / DataDog RUM | LCP>4秒用户占比>10%→P2 |

### 7.2 告警分级与SLA

| 级别 | 触发条件 | 响应时间 | 恢复时间 | 告警方式 |
|-----|---------|---------|---------|---------|
| **P0致命** | 核心服务不可用、超卖 | <5分钟 | <30分钟 | 电话+短信+企微@all |
| **P1严重** | 核心接口错误率>5%、下单量骤降>50% | <10分钟 | <1小时 | 短信+企微+电话（3分钟未ACK） |
| **P2一般** | 非核心接口错误率>10%、P99延迟>5秒 | <30分钟 | <4小时 | 企微+邮件 |
| **P3预警** | 磁盘使用率>80%、Redis内存>85% | <1小时 | 当天内 | 企微（不@人） |

**核心服务SLA目标（年度）**：
- **Tier 1（订单/支付/结算/库存）**：可用性≥99.95%，P95<500ms，错误率<0.1%
- **Tier 2（商品/搜索/购物车/营销）**：可用性≥99.9%，P95<1s，错误率<0.5%
- **Tier 3（推荐/评价/日志）**：可用性≥99.5%，P95<2s，错误率<1%

### 7.3 限流策略（多层防护）

| 层级 | 工具 | 策略 |
|-----|------|------|
| **L1接入层** | APISIX | IP限流：100req/min；API限流：5000req/s |
| **L2用户维度** | Redis + Token Bucket | 下单：5次/分钟/用户；结算：10次/分钟/用户 |
| **L3服务维度** | Istio | order→inventory：3000req/s |
| **L4资源维度** | MySQL连接池 | 最大连接2000，等待超时3秒 |

### 7.4 全链路压测

**压测流程（大促前3周）**：
- **Week 1**：准备测试数据、扩容资源、配置压测标识（`X-Test-Flag: pressure-test`）
- **Week 2**：分层压测（接口单点→核心链路→全链路），目标峰值QPS 25万
- **Week 3**：瓶颈分析、扩容决策、降级预案验证

**压测指标**：
- QPS：能否达到目标值
- 响应时间：P95<500ms，P99<1s
- 错误率：<0.1%
- 资源使用：CPU<70%，内存<80%

### 7.5 故障演练（Chaos Engineering）

**演练频率**：每季度1次

**演练场景**：
1. **服务不可用**：随机删除Pod，验证K8s自愈能力
2. **数据库主从切换**：模拟主库宕机，验证MHA自动切换
3. **网络延迟**：注入2秒延迟（Istio Fault Injection），验证超时/熔断/降级
4. **机房故障**：模拟IDC-A整体不可用，验证DNS切换和数据一致性

**工具**：Chaos Mesh、Litmus、Istio Fault Injection

### 7.6 故障复盘流程

**触发条件**：P0/P1故障、用户影响>10000人、故障时长>30分钟、数据丢失/错误

**时间线**：
- **T+4h**：初步复盘（电话会议），梳理时间轴、影响范围
- **T+1day**：根因分析（5 Why、鱼骨图）
- **T+2day**：改进计划（短期Hotfix、中期监控、长期工具）
- **T+1week**：复盘文档归档、全员分享会（无责文化）

---

## 八、技术栈总结

| 技术领域 | 选型 | 版本 | 理由 |
|---------|------|------|------|
| **编程语言** | Go | 1.21+ | 高并发、部署简单 |
| **API网关** | APISIX | 3.x | 性能强、插件丰富 |
| **数据库** | MySQL | 8.0 | 事务、成熟度 |
| **缓存** | Redis Cluster | 7.x | 高性能、持久化 |
| **消息队列** | Kafka | 3.x | 高吞吐、持久化 |
| **搜索引擎** | Elasticsearch | 8.x | 全文搜索、聚合 |
| **RPC框架** | gRPC | 1.60+ | 高性能、跨语言 |
| **服务网格** | Istio | 1.20+ | 流量管理、可观测 |
| **配置中心** | Nacos | 2.x | 动态配置、服务发现 |
| **链路追踪** | Jaeger | 1.50+ | 分布式追踪 |
| **监控告警** | Prometheus + Grafana | - | 指标采集、可视化 |
| **日志** | ELK Stack | 8.x | 日志收集、分析 |
| **容器编排** | Kubernetes | 1.28+ | 容器调度、弹性伸缩 |

---

## 九、成本预估

**双机房总成本**：

| 资源类型 | 单机房月成本 | 双机房月成本 | 年成本 |
|---------|------------|------------|--------|
| 物理服务器（100台） | ¥300,000 | ¥600,000 | ¥7.2M |
| MySQL（24实例） | ¥120,000 | ¥240,000 | ¥2.88M |
| Redis（16实例） | ¥32,000 | ¥64,000 | ¥768K |
| Kafka（6 Broker） | ¥18,000 | ¥36,000 | ¥432K |
| Elasticsearch（6节点） | ¥24,000 | ¥48,000 | ¥576K |
| 网络带宽（10Gbps） | ¥50,000 | ¥100,000 | ¥1.2M |
| 负载均衡（F5） | ¥40,000 | ¥80,000 | ¥960K |
| 存储（500TB SSD） | ¥250,000 | ¥500,000 | ¥6M |
| 监控告警 | ¥20,000 | ¥40,000 | ¥480K |
| **合计** | **¥854,000** | **¥1.7M** | **¥20.5M** |

---

## 十、总结与展望

### 10.1 架构优势

1. **高可用**：同城双活+故障自动切换，核心服务SLA≥99.95%
2. **高性能**：三级缓存+分库分表+Redis原子操作，P99延迟<1秒
3. **高扩展**：微服务+K8s HPA，支持5-10倍弹性扩容
4. **容错性强**：供应商网关熔断降级+Saga补偿，单点故障不影响全局
5. **数据一致**：Saga+幂等+对账，保证订单/库存/资金强一致

### 10.2 技术挑战

1. **供应商接口复杂度**：50+供应商，需持续维护适配器插件
2. **数据一致性成本**：Saga补偿+对账任务，增加系统复杂度
3. **运维复杂度**：双机房部署+中间件集群，需专业SRE团队
4. **成本控制**：年成本2000万+，需持续优化资源使用率

### 10.3 未来演进方向

1. **异地多活**：从同城双活扩展到异地三中心（北京+上海+深圳）
2. **智能化运维**：引入AIOps，自动根因分析+自动扩缩容
3. **Serverless化**：边缘服务（推荐/评价）迁移到Serverless，降本增效
4. **全链路灰度**：基于流量染色的全链路灰度发布能力

---

## 参考资料

1. [Martin Fowler - Microservices Architecture](https://martinfowler.com/microservices/)
2. [Saga Pattern - Chris Richardson](https://microservices.io/patterns/data/saga.html)
3. [Google SRE Book](https://sre.google/sre-book/table-of-contents/)
4. [Alibaba技术 - 淘宝双11技术揭秘](https://developer.aliyun.com/topic/taobao1111)
5. [Redis官方文档](https://redis.io/docs/)
6. [Kubernetes官方文档](https://kubernetes.io/docs/home/)
7. [Istio官方文档](https://istio.io/latest/docs/)

---

**作者**：wxquare  
**日期**：2026-04-14  
**版本**：v1.0
