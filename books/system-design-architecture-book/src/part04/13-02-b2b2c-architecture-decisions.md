# 32.4-32.5 整体架构与技术选型

## 32.4 整体架构设计（Application Architecture - 设计结果）

基于32.3节识别的12个限界上下文和上下文映射关系，本节展示如何将它们落地为具体的架构设计：四层架构、微服务拆分、服务依赖关系、数据流转模式。

**32.3 → 32.4的映射关系**：

```text
32.3 限界上下文           →    32.4 微服务
├─ 订单上下文             →    Order Service
├─ 商品上下文             →    Product Center
├─ 库存上下文             →    Inventory Service
├─ 计价上下文             →    Pricing Service
├─ 营销上下文             →    Marketing Service
├─ 支付上下文             →    Payment Service
├─ 搜索上下文             →    Search Service
└─ 供应商上下文           →    Supplier Gateway

32.3 上下文映射           →    32.4 服务集成
├─ Customer-Supplier      →    同步RPC调用
├─ Anti-Corruption Layer  →    适配器模式
└─ Published Language     →    Kafka事件
```

### 32.4.1 分层架构

采用经典的四层架构，确保职责清晰、易于维护。

基于前面识别的限界上下文和映射关系，本节通过实际案例展示如何划分边界、重构边界。

**案例1：计价系统的边界重构**

**初始问题**：
- 价格计算逻辑分散在订单、营销、商品三个域
- 购物车、订单创建、支付确认三处价格计算不一致
- 无法支持"PDP加购试算"场景

**重构方案**：
1. **新建计价上下文**：职责是提供统一的试算接口
2. **定义边界**：
   - 计价上下文**不拥有**商品基础价、营销规则、订单状态
   - 对外提供 `Calculate(items, promotions, context) -> PriceBreakdown`
   - 各场景通过统一接口获取价格
3. **收益**：
   - 价格一致性得到保证
   - 营销规则变更只需在营销域发布事件
   - 支持了试算、价格预览、价格审计等新需求

**案例2：库存预占的归属**

**争议**：库存预占应该放在订单域还是库存域？

**决策**：放在库存域

**理由**：
- 库存域拥有库存数据所有权
- 预占是库存的一种状态（可售 → 预占 → 扣减）
- 订单域只需调用库存域的 `Reserve` 接口
- 降低耦合：订单域不需要了解库存的存储结构

### 32.4.4 集成模式选择

| 集成场景 | 模式 | 理由 |
|---------|------|------|
| 订单 → 商品 | 同步RPC | 需要实时获取商品信息，延迟<100ms |
| 订单 → 库存 | 同步RPC | 库存预占是核心路径，必须同步 |
| 订单 → 支付 | 同步RPC | 支付创建需要同步返回支付URL |
| 订单成功 → 搜索 | 异步事件 | 销量更新非核心路径，可最终一致 |
| 订单成功 → 积分 | 异步事件 | 积分增加非核心路径 |

**事件驱动示例**：

```go
// 订单域发布事件
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    // 创建订单...
    order := &Order{...}
    s.repo.Save(ctx, order)
    
    // 发布事件（Outbox模式）
    event := &OrderCreatedEvent{
        OrderID:    order.ID,
        UserID:     order.UserID,
        TotalPrice: order.TotalPrice,
        Items:      order.Items,
    }
    s.outbox.Publish(ctx, "order-events", event)
    
    return order, nil
}

// 搜索域订阅事件
func (s *SearchService) HandleOrderCreated(ctx context.Context, event *OrderCreatedEvent) error {
    // 更新商品销量（用于排序）
    for _, item := range event.Items {
        s.incrementSales(ctx, item.SkuID, item.Quantity)
    }
    return nil
}
```

### 32.4.5 跨系统事务处理

**Saga模式（编排）**：

```go
// 订单创建Saga
type CreateOrderSaga struct {
    inventoryClient rpc.InventoryClient
    marketingClient rpc.MarketingClient
    orderRepo       *OrderRepo
}

func (s *CreateOrderSaga) Execute(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    var reserveID string
    var couponLockID string
    
    // Step 1: 库存预占
    reserve, err := s.inventoryClient.ReserveStock(ctx, req.Items)
    if err != nil {
        return nil, fmt.Errorf("库存预占失败: %w", err)
    }
    reserveID = reserve.ReserveID
    defer func() {
        if err != nil {
            // 补偿：释放库存
            s.inventoryClient.ReleaseStock(ctx, reserveID)
        }
    }()
    
    // Step 2: 优惠券锁定
    couponLock, err := s.marketingClient.LockCoupon(ctx, req.CouponCode, req.UserID)
    if err != nil {
        return nil, fmt.Errorf("优惠券锁定失败: %w", err)
    }
    couponLockID = couponLock.LockID
    defer func() {
        if err != nil {
            // 补偿：释放优惠券
            s.marketingClient.UnlockCoupon(ctx, couponLockID)
        }
    }()
    
    // Step 3: 创建订单
    order := &Order{
        ID:           generateOrderID(),
        UserID:       req.UserID,
        Items:        req.Items,
        ReserveID:    reserveID,
        CouponLockID: couponLockID,
        Status:       StatusPendingPayment,
    }
    err = s.orderRepo.Save(ctx, order)
    if err != nil {
        return nil, fmt.Errorf("订单创建失败: %w", err)
    }
    
    return order, nil
}
```

### 32.4.6 防腐层设计

**防腐层（Anti-Corruption Layer）**：

```go
// 供应商响应模型（外部）
type SupplierFlightResponse struct {
    Code    string  `json:"code"`
    Message string  `json:"message"`
    Data    struct {
        FlightNo  string  `json:"flight_no"`
        Available int     `json:"available"`
        Price     float64 `json:"price"`
    } `json:"data"`
}

// 平台库存模型（内部）
type StockResponse struct {
    Available bool
    Quantity  int
    Message   string
}

// 防腐层：翻译外部模型 → 内部模型
func (a *FlightSupplierACL) TranslateStock(supplierResp *SupplierFlightResponse) *StockResponse {
    return &StockResponse{
        Available: supplierResp.Code == "SUCCESS" && supplierResp.Data.Available > 0,
        Quantity:  supplierResp.Data.Available,
        Message:   supplierResp.Message,
    }
}
```

**收益**：
- 领域层不被供应商模型污染
- 供应商接口变更时，修改集中在ACL
- 测试时可以使用Fake实现替代真实供应商

---

## 32.5 技术选型决策（Technology Architecture）

### 32.5.1 选型原则

**原则1：成熟度优先**
- 优先选择生产级成熟技术（避免踩坑）
- 社区活跃、文档完善、案例丰富
- 避免使用 alpha/beta 版本

**原则2：团队能力匹配**
- 技术栈与团队技能对齐
- 学习曲线可控（新技术培训 < 1个月）
- 有内部专家支持

**原则3：生态完整性**
- 工具链完善（测试、监控、部署）
- 第三方库丰富
- 云服务支持（AWS/GCP/阿里云）

**原则4：成本可控**
- 开源优先（降低License成本）
- 云服务按需使用（避免自建中间件）
- 运维成本可接受

### 32.5.2 Go生态选型

**语言选择：Go**

| 维度 | Go | Java | 理由 |
|------|-----|------|------|
| 性能 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 协程模型，高并发性能优异 |
| 开发效率 | ⭐⭐⭐⭐ | ⭐⭐⭐ | 编译快，部署简单（单一二进制） |
| 学习曲线 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | 语法简洁，容易上手 |
| 生态 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 微服务生态完善（gRPC/Consul/Envoy） |
| 团队能力 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | 团队有Go经验 |

**Web框架：Gin**
```go
// 理由：
// 1. 性能优异（httprouter，零内存分配）
// 2. 中间件丰富（鉴权、限流、日志）
// 3. 社区活跃（GitHub 70k+ stars）

router := gin.Default()
router.Use(middleware.Auth())
router.Use(middleware.RateLimit(1000))
router.GET("/products/:id", handler.GetProduct)
```

**ORM：GORM**
```go
// 理由：
// 1. 支持MySQL、PostgreSQL、SQLite
// 2. 关联查询、预加载、Hook机制完善
// 3. 自动迁移（开发环境）

type Product struct {
    ID       int64  `gorm:"primaryKey"`
    Title    string `gorm:"size:255;not null"`
    Price    int64  `gorm:"not null"`
}
```

**RPC：gRPC + Protobuf**
```go
// 理由：
// 1. 二进制序列化（性能优于JSON）
// 2. 强类型（编译期检查）
// 3. 支持流式调用（双向流）

service ProductService {
    rpc GetProduct(GetProductRequest) returns (GetProductResponse);
    rpc BatchGetProduct(BatchGetProductRequest) returns (stream Product);
}
```

**依赖注入：Google Wire**
```go
// 理由：
// 1. 编译时生成（无反射，性能高）
// 2. 类型安全（编译期检查依赖）
// 3. 官方支持（Google开源）

//go:generate wire
func InitializeApp() (*App, error) {
    wire.Build(
        NewDB,
        NewRedis,
        NewProductRepo,
        NewProductService,
        NewApp,
    )
    return nil, nil
}
```

### 32.4.3 数据库选型

**MySQL（主库）**

| 场景 | 选择理由 | 配置 |
|------|---------|------|
| 订单表 | ACID保证、事务支持 | InnoDB，8分库64表 |
| 商品表 | 关联查询、JOIN支持 | InnoDB，4分库 |
| 支付表 | 强一致性、金融级可靠性 | InnoDB，双主互备 |

**Redis（缓存 + 库存）**

| 场景 | 数据结构 | TTL |
|------|---------|-----|
| 商品详情 | Hash | 30分钟 |
| 库存数量 | String（Lua原子扣减） | 永久 |
| 券码池热队列 | List（只存 `code_id`，MySQL CAS 后才算锁码成功） | 可从 MySQL 重建 |
| 用户Session | String | 2小时 |

**Elasticsearch（搜索 + 日志）**

| 场景 | 索引设计 | 刷新间隔 |
|------|---------|---------|
| 商品搜索 | product_index（标题、类目、属性） | 30秒 |
| 订单查询 | order_index（订单号、用户ID、状态） | 1分钟 |
| 日志搜索 | log-{date}（按日分索引） | 5秒 |

### 32.4.4 中间件选型

**Kafka（消息队列）**

| 场景 | Topic | Partition | Replication |
|------|-------|-----------|-------------|
| 订单事件 | order-events | 16 | 3 |
| 库存事件 | inventory-events | 8 | 3 |
| 日志采集 | logs | 32 | 2 |

**Consul（服务发现）**
- 健康检查：HTTP/TCP/gRPC
- 配置中心：动态配置热更新
- KV存储：Feature Flag

**Envoy（Service Mesh）**
- 流量管理：灰度发布、A/B测试
- 可观测性：自动生成Trace
- 安全：mTLS加密

---
