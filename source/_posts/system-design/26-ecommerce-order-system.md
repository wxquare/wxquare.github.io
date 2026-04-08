---
title: 电商系统设计：订单系统
date: 2026-04-07
categories:
  - system-design
  - e-commerce
tags:
  - order-system
  - state-machine
  - distributed-transaction
  - idempotency
  - saga
  - tcc
  - consistency
  - e-commerce
---

# 电商系统设计：订单系统

订单系统是电商平台的核心，承载着从下单到履约的完整业务流程。本文将深入探讨订单系统的设计与实现，重点讲解状态机、分布式事务、幂等性三大核心技术，并通过虚拟订单、O2O订单、预售订单三个黄金案例，展示如何设计可扩展的订单系统。

本文既适合系统设计面试准备，也适合工程实践参考。

## 目录

- [1. 系统概览](#1-系统概览)
  - [1.1 业务场景](#11-业务场景)
  - [1.2 核心挑战](#12-核心挑战)
  - [1.3 系统架构](#13-系统架构)
  - [1.4 数据模型设计](#14-数据模型设计)
- [2. 通用订单流程](#2-通用订单流程)
  - [2.1 订单创建](#21-订单创建)
  - [2.2 订单支付](#22-订单支付)
  - [2.3 订单履约](#23-订单履约)
  - [2.4 订单售后](#24-订单售后)
- [3. 状态机设计专题](#3-状态机设计专题)
- [4. 分布式事务与一致性](#4-分布式事务与一致性)
- [5. 幂等性与去重](#5-幂等性与去重)
- [6. 特殊订单类型](#6-特殊订单类型)
  - [6.1 虚拟订单](#61-虚拟订单)
  - [6.2 O2O订单](#62-o2o订单)
  - [6.3 预售订单](#63-预售订单)
- [7. 订单类型扩展设计](#7-订单类型扩展设计)
- [8. 工程实践要点](#8-工程实践要点)
- [总结](#总结)
- [参考资料](#参考资料)

## 1. 系统概览

### 1.1 业务场景

订单系统是电商平台的核心枢纽，连接用户、商品、库存、支付、物流、营销等多个子系统。它的主要职责包括：

- **订单创建**：接收用户下单请求，协调库存扣减、优惠计算、积分扣减等操作
- **订单支付**：对接支付系统，处理支付回调，更新订单状态
- **订单履约**：对接物流系统，跟踪物流状态，自动确认收货
- **订单售后**：处理退款退货，协调库存回补、优惠退还等逆向操作

订单系统的职责边界：
- **负责**：订单状态管理、订单数据持久化、订单流程编排
- **不负责**：具体的库存扣减逻辑（由库存系统负责）、具体的支付逻辑（由支付系统负责）

与其他系统的交互：
- **商品系统**：获取商品信息，创建订单快照
- **库存系统**：扣减库存、回补库存
- **支付系统**：发起支付、接收支付回调
- **物流系统**：创建物流单、接收物流状态更新
- **营销系统**：扣减优惠券、扣减积分、回退优惠

### 1.2 核心挑战

订单系统面临以下核心技术挑战：

**1. 高并发**
- 大促期间订单创建QPS可达百万级
- 需要支持数据库分库分表、缓存、消息队列削峰
- 需要合理的限流和熔断策略

**2. 强一致性**
- 订单创建涉及库存、优惠、积分等多个系统，需要保证事务一致性
- 支付回调需要防止重复扣款
- 库存扣减和订单创建需要原子性

**3. 状态复杂**
- 订单生命周期涉及多个状态：待支付、已支付、待发货、已发货、运输中、已送达、已完成、已取消、售后中等
- 状态转换需要严格控制，防止非法转换
- 需要记录完整的状态变更历史

**4. 类型多样**
- 物理订单：需要物流配送
- 虚拟订单：无需物流，即时履约
- O2O订单：需要商家接单、骑手配送
- 预售订单：定金尾款分期支付、延迟履约
- 每种订单类型的状态机和业务逻辑都有差异

**5. 幂等性**
- 支付回调可能重复：同一笔支付可能收到多次回调
- 物流回调可能重复：同一个物流状态可能上报多次
- 用户重复点击：用户可能多次点击支付按钮
- 需要在订单创建、支付、履约、售后等各个环节保证幂等性

**6. 可追溯**
- 需要保存订单快照：商品信息、价格、优惠信息在下单时的状态
- 需要记录完整的状态变更历史：谁在什么时间做了什么操作
- 需要支持订单审计和数据对账

### 1.3 系统架构

#### 整体架构

订单系统在电商平台中处于核心位置，通过同步API和异步消息与其他系统交互：

- **同步调用**：订单创建时同步调用库存系统、营销系统（需要立即返回结果）
- **异步消息**：订单支付成功后发布事件，履约系统异步消费（允许延迟处理）

#### 模块划分

订单系统内部分为以下核心模块：

**1. Order Service（订单核心服务）**
- 订单创建：接收下单请求，编排分布式事务
- 订单查询：提供订单查询API
- 订单状态管理：状态机驱动的状态转换

**2. Payment Service（支付服务）**
- 支付发起：调用第三方支付平台
- 支付回调：处理支付平台回调，更新订单状态
- 支付对账：定期与支付平台对账

**3. Fulfillment Service（履约服务）**
- 履约编排：订单支付成功后触发履约流程
- 物流对接：创建物流单，跟踪物流状态
- 自动确认：超时自动确认收货

**4. After-sale Service（售后服务）**
- 售后申请：用户发起退款退货
- 售后审核：人工或自动审核
- 退款处理：调用支付系统退款，回退库存和优惠

#### 技术栈

**存储层**
- **MySQL**：订单主数据存储，支持ACID事务
- **Redis**：订单缓存，提高查询性能
- **Elasticsearch**：订单搜索，支持复杂查询

**消息队列**
- **Kafka**：事件驱动架构，发布订单事件（OrderCreatedEvent、OrderPaidEvent等）

**分布式事务**
- **TCC框架**：支付场景，强一致性
- **Saga框架**：订单创建、售后场景，最终一致性

#### 系统架构图

```mermaid
graph TB
    User[用户] --> OrderAPI[订单API]
    OrderAPI --> OrderService[订单服务]
    
    OrderService --> ProductService[商品服务]
    OrderService --> InventoryService[库存服务]
    OrderService --> MarketingService[营销服务]
    OrderService --> PaymentGateway[支付网关]
    
    OrderService --> MySQL[(MySQL<br/>订单主数据)]
    OrderService --> Redis[(Redis<br/>缓存)]
    OrderService --> ES[(Elasticsearch<br/>订单搜索)]
    
    OrderService --> Kafka[Kafka消息队列]
    Kafka --> FulfillmentWorker[履约Worker]
    FulfillmentWorker --> LogisticsService[物流服务]
    
    PaymentGateway --> ThirdPartyPay[第三方支付]
    ThirdPartyPay -.支付回调.-> PaymentCallback[支付回调]
    PaymentCallback --> OrderService
    
    LogisticsService -.物流状态回调.-> LogisticsCallback[物流回调]
    LogisticsCallback --> FulfillmentWorker

    style OrderService fill:#e1f5ff
    style MySQL fill:#ffe1e1
    style Redis fill:#ffe1e1
    style Kafka fill:#e1ffe1
```

### 1.4 数据模型设计

订单系统的核心数据模型包括订单主表、订单明细表、订单快照表、状态变更历史表、幂等表。

#### 订单主表（order）

存储订单的基本信息：

```go
type Order struct {
    OrderID       string    // 订单ID（Snowflake生成）
    UserID        int64     // 用户ID
    OrderType     int       // 订单类型：1-物理订单 2-虚拟订单 3-O2O订单 4-预售订单
    Status        int       // 订单状态：1-待支付 2-已支付 3-待发货 4-已发货 ...
    TotalAmount   int64     // 订单总金额（分）
    PaymentAmount int64     // 实付金额（分）
    DiscountAmount int64    // 优惠金额（分）
    CASVersion    int64     // 乐观锁版本号
    CreatedAt     time.Time // 创建时间
    UpdatedAt     time.Time // 更新时间
}
```

**索引设计**：
- 主键：`order_id`
- 唯一索引：`user_id, created_at`（支持用户订单查询）
- 普通索引：`status`（支持按状态查询）

#### 订单明细表（order_item）

存储订单的商品明细：

```go
type OrderItem struct {
    ItemID      int64     // 明细ID
    OrderID     string    // 订单ID
    ProductID   int64     // 商品ID
    SkuID       int64     // SKU ID
    Quantity    int       // 数量
    Price       int64     // 单价（分）
    SnapshotID  string    // 快照ID
    CreatedAt   time.Time // 创建时间
}
```

**索引设计**：
- 主键：`item_id`
- 普通索引：`order_id`（支持根据订单查询明细）

#### 订单快照表（order_snapshot）

存储下单时的商品快照（价格、标题、图片等），防止商品信息变更影响订单：

```go
type OrderSnapshot struct {
    SnapshotID    string    // 快照ID（Hash生成，支持复用）
    ProductID     int64     // 商品ID
    SkuID         int64     // SKU ID
    Title         string    // 商品标题
    Image         string    // 商品图片
    Price         int64     // 商品价格（分）
    Specifications string   // 规格信息（JSON）
    CreatedAt     time.Time // 创建时间
}
```

**快照复用策略**：
- 基于商品ID、SKU ID、价格、规格等信息计算Hash
- 相同Hash的快照复用同一条记录，节省存储空间

#### 订单状态变更历史表（order_state_log）

记录订单的所有状态变更，支持审计和追溯：

```go
type OrderStateLog struct {
    LogID       int64     // 日志ID
    OrderID     string    // 订单ID
    FromStatus  int       // 变更前状态
    ToStatus    int       // 变更后状态
    Operator    string    // 操作人（系统/用户ID）
    Reason      string    // 变更原因
    CreatedAt   time.Time // 创建时间
}
```

#### 幂等表（idempotent_record）

记录幂等键，防止重复操作：

```go
type IdempotentRecord struct {
    IdempotentKey string    // 幂等键（唯一）
    BizType       string    // 业务类型：order_create/payment/fulfillment
    BizID         string    // 业务ID（订单ID/支付单号等）
    Status        int       // 状态：1-处理中 2-成功 3-失败
    ExpireAt      time.Time // 过期时间
    CreatedAt     time.Time // 创建时间
}
```

**索引设计**：
- 唯一索引：`idempotent_key`（防止重复插入）

#### ER图

```mermaid
erDiagram
    ORDER ||--o{ ORDER_ITEM : contains
    ORDER ||--o{ ORDER_STATE_LOG : tracks
    ORDER_ITEM ||--|| ORDER_SNAPSHOT : references
    ORDER ||--o| IDEMPOTENT_RECORD : protected_by
    
    ORDER {
        string order_id PK
        int64 user_id
        int order_type
        int status
        int64 total_amount
        int64 payment_amount
        int64 cas_version
        timestamp created_at
    }
    
    ORDER_ITEM {
        int64 item_id PK
        string order_id FK
        int64 product_id
        int64 sku_id
        int quantity
        int64 price
        string snapshot_id FK
    }
    
    ORDER_SNAPSHOT {
        string snapshot_id PK
        int64 product_id
        int64 sku_id
        string title
        string image
        int64 price
        string specifications
    }
    
    ORDER_STATE_LOG {
        int64 log_id PK
        string order_id FK
        int from_status
        int to_status
        string operator
        string reason
        timestamp created_at
    }
    
    IDEMPOTENT_RECORD {
        string idempotent_key PK
        string biz_type
        string biz_id
        int status
        timestamp expire_at
    }
```

## 2. 通用订单流程

本章讲解标准订单从创建到完成的完整流程，覆盖大部分订单类型的通用逻辑（约占50%内容）。特殊订单类型的差异将在第6章详细讲解。

### 2.1 订单创建

订单创建是整个订单流程的起点，需要协调多个系统完成库存扣减、优惠计算、积分扣减等操作。由于涉及多个系统，需要使用分布式事务保证一致性。

#### 业务流程

1. **参数校验**：验证商品是否存在、库存是否充足、优惠券是否可用等
2. **生成订单ID**：使用Snowflake算法生成全局唯一的订单ID
3. **创建订单快照**：保存下单时的商品信息（价格、标题、图片等）
4. **分布式事务编排**：
   - Try阶段：预扣库存、预扣优惠券、预扣积分、创建订单（状态为"草稿"）
   - Confirm阶段：所有Try成功后，订单状态变为"待支付"
   - Cancel阶段：任何Try失败，执行补偿回滚

#### 订单ID生成策略

使用Snowflake算法生成订单ID：

```go
// Snowflake算法：64位long型
// [0] [1-41时间戳] [42-51机器ID] [52-63序列号]

type SnowflakeGenerator struct {
    machineID int64  // 机器ID（10位，0-1023）
    sequence  int64  // 序列号（12位，0-4095）
    lastTime  int64  // 上次生成时间戳
    mu        sync.Mutex
}

func (s *SnowflakeGenerator) NextID() string {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    now := time.Now().UnixMilli()
    
    // 如果时间戳相同，序列号递增
    if now == s.lastTime {
        s.sequence = (s.sequence + 1) & 4095
        if s.sequence == 0 {
            // 序列号溢出，等待下一毫秒
            for now <= s.lastTime {
                now = time.Now().UnixMilli()
            }
        }
    } else {
        s.sequence = 0
    }
    
    s.lastTime = now
    
    // 组装：时间戳(41位) + 机器ID(10位) + 序列号(12位)
    id := ((now - 1640995200000) << 22) | (s.machineID << 12) | s.sequence
    return strconv.FormatInt(id, 10)
}
```

**优点**：
- 全局唯一：机器ID + 时间戳 + 序列号保证唯一性
- 趋势递增：按时间递增，对数据库索引友好
- 高性能：无需数据库交互，本地生成

**缺点**：
- 依赖时钟：时钟回拨会导致ID重复（需要拒绝服务并告警）
- 需要机器ID分配：在分布式环境下需要全局唯一的机器ID

#### 订单快照管理

订单快照保存下单时的商品信息，防止商品信息变更影响订单：

```go
// 生成快照ID（基于Hash，支持复用）
func GenerateSnapshotID(productID, skuID int64, price int64, spec string) string {
    data := fmt.Sprintf("%d_%d_%d_%s", productID, skuID, price, spec)
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:16]) // 使用前16字节
}

// 创建或复用快照
func CreateOrReuseSnapshot(snapshot *OrderSnapshot) (string, error) {
    snapshotID := GenerateSnapshotID(
        snapshot.ProductID,
        snapshot.SkuID,
        snapshot.Price,
        snapshot.Specifications,
    )
    
    // 尝试查询是否已存在
    existing, err := db.GetSnapshotByID(snapshotID)
    if err == nil && existing != nil {
        return snapshotID, nil // 复用现有快照
    }
    
    // 不存在，创建新快照
    snapshot.SnapshotID = snapshotID
    if err := db.InsertSnapshot(snapshot); err != nil {
        return "", err
    }
    
    return snapshotID, nil
}
```

**快照复用的好处**：
- 节省存储空间：相同商品信息的快照只存储一份
- 提高查询性能：减少数据量，加快查询速度

#### 订单创建流程图

```mermaid
sequenceDiagram
    participant User as 用户
    participant API as 订单API
    participant Order as 订单服务
    participant Inventory as 库存服务
    participant Marketing as 营销服务
    participant DB as 数据库
    participant Kafka as Kafka
    
    User->>API: 提交订单
    API->>Order: CreateOrder(request)
    
    Note over Order: 1. 参数校验
    Order->>Order: ValidateRequest()
    
    Note over Order: 2. 生成订单ID
    Order->>Order: GenerateOrderID()
    
    Note over Order: 3. 创建快照
    Order->>Order: CreateSnapshot()
    
    Note over Order: 4. Saga事务开始
    Order->>Inventory: Try: 预扣库存
    Inventory-->>Order: Success
    
    Order->>Marketing: Try: 预扣优惠券
    Marketing-->>Order: Success
    
    Order->>Marketing: Try: 预扣积分
    Marketing-->>Order: Success
    
    Order->>DB: 创建订单（状态=草稿）
    DB-->>Order: Success
    
    Note over Order: 5. 所有Try成功，Confirm
    Order->>DB: 更新订单状态=待支付
    DB-->>Order: Success
    
    Order->>DB: 写入本地消息表
    Order->>Kafka: 发布OrderCreatedEvent
    
    Order-->>API: Success(orderID)
    API-->>User: 订单创建成功
    
    Note over Order: 如果任一步骤失败
    Order->>Inventory: Cancel: 回滚库存
    Order->>Marketing: Cancel: 回滚优惠券
    Order->>Marketing: Cancel: 回滚积分
    Order-->>API: Failure
    API-->>User: 订单创建失败
```

#### 订单创建状态机

```mermaid
stateDiagram-v2
    [*] --> 草稿: 创建订单
    草稿 --> 待支付: Try全部成功
    草稿 --> 已取消: Try失败/超时
    待支付 --> 支付中: 发起支付
    待支付 --> 已取消: 超时未支付
    支付中 --> 已支付: 支付成功
    支付中 --> 支付失败: 支付失败
    支付失败 --> 已取消: 自动取消
    已取消 --> [*]
```

#### Saga分布式事务

订单创建涉及多个系统，使用Saga模式保证最终一致性：

```go
// Saga步骤定义
type SagaStep struct {
    Name       string
    TryFunc    func(ctx context.Context) error    // 正向操作
    CancelFunc func(ctx context.Context) error    // 补偿操作
}

// Saga协调器
type SagaOrchestrator struct {
    steps []*SagaStep
}

func (s *SagaOrchestrator) Execute(ctx context.Context) error {
    executed := make([]*SagaStep, 0)
    
    // 顺序执行Try
    for _, step := range s.steps {
        if err := step.TryFunc(ctx); err != nil {
            // 失败，执行补偿
            s.compensate(ctx, executed)
            return fmt.Errorf("saga step %s failed: %w", step.Name, err)
        }
        executed = append(executed, step)
    }
    
    return nil
}

func (s *SagaOrchestrator) compensate(ctx context.Context, executed []*SagaStep) {
    // 逆序执行Cancel
    for i := len(executed) - 1; i >= 0; i-- {
        step := executed[i]
        if err := step.CancelFunc(ctx); err != nil {
            // 补偿失败，记录日志并告警
            log.Error("saga compensation failed", 
                "step", step.Name, 
                "error", err)
            // 发送告警，人工介入
            alert.Send("saga_compensation_failed", step.Name, err)
        }
    }
}

// 订单创建Saga
func CreateOrderSaga(ctx context.Context, req *CreateOrderRequest) error {
    saga := &SagaOrchestrator{
        steps: []*SagaStep{
            {
                Name: "扣减库存",
                TryFunc: func(ctx context.Context) error {
                    return inventoryClient.DeductStock(ctx, req.Items)
                },
                CancelFunc: func(ctx context.Context) error {
                    return inventoryClient.RollbackStock(ctx, req.Items)
                },
            },
            {
                Name: "扣减优惠券",
                TryFunc: func(ctx context.Context) error {
                    return marketingClient.DeductCoupon(ctx, req.CouponID)
                },
                CancelFunc: func(ctx context.Context) error {
                    return marketingClient.RollbackCoupon(ctx, req.CouponID)
                },
            },
            {
                Name: "扣减积分",
                TryFunc: func(ctx context.Context) error {
                    return marketingClient.DeductPoints(ctx, req.UserID, req.Points)
                },
                CancelFunc: func(ctx context.Context) error {
                    return marketingClient.RollbackPoints(ctx, req.UserID, req.Points)
                },
            },
            {
                Name: "创建订单",
                TryFunc: func(ctx context.Context) error {
                    order := buildOrder(req)
                    order.Status = OrderStatusDraft // 草稿状态
                    return db.InsertOrder(ctx, order)
                },
                CancelFunc: func(ctx context.Context) error {
                    return db.DeleteOrder(ctx, req.OrderID)
                },
            },
        },
    }
    
    // 执行Saga
    if err := saga.Execute(ctx); err != nil {
        return err
    }
    
    // 所有Try成功，更新订单状态为"待支付"
    return db.UpdateOrderStatus(ctx, req.OrderID, OrderStatusPending)
}
```

**Saga vs TCC对比**：
- Saga适合订单创建场景：步骤多、允许最终一致性
- TCC适合支付场景：步骤少、需要强一致性（下一节详述）

#### 幂等性设计

防止用户重复提交订单：

```go
// 基于请求ID的幂等性
func CreateOrderIdempotent(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    // 1. 幂等键：用户ID + 请求ID
    idempotentKey := fmt.Sprintf("order_create_%d_%s", req.UserID, req.RequestID)
    
    // 2. 尝试插入幂等记录（唯一索引保证原子性）
    record := &IdempotentRecord{
        IdempotentKey: idempotentKey,
        BizType:       "order_create",
        Status:        IdempotentProcessing,
        ExpireAt:      time.Now().Add(24 * time.Hour),
    }
    
    if err := db.InsertIdempotentRecord(ctx, record); err != nil {
        // 插入失败，说明已经处理过
        existing, _ := db.GetIdempotentRecord(ctx, idempotentKey)
        if existing.Status == IdempotentSuccess {
            // 已成功，返回之前的订单
            order, _ := db.GetOrder(ctx, existing.BizID)
            return order, nil
        }
        // 处理中，返回错误提示稍后重试
        return nil, ErrRequestProcessing
    }
    
    // 3. 执行订单创建
    order, err := CreateOrderSaga(ctx, req)
    if err != nil {
        // 失败，更新幂等记录状态
        db.UpdateIdempotentStatus(ctx, idempotentKey, IdempotentFailed)
        return nil, err
    }
    
    // 4. 成功，更新幂等记录
    db.UpdateIdempotentRecord(ctx, idempotentKey, IdempotentSuccess, order.OrderID)
    
    return order, nil
}
```

#### 数据一致性保证

**乐观锁更新订单状态**：

```go
// 使用CAS版本号防止并发冲突
func UpdateOrderStatus(ctx context.Context, orderID string, oldStatus, newStatus int) error {
    query := `
        UPDATE orders 
        SET status = ?, cas_version = cas_version + 1, updated_at = ?
        WHERE order_id = ? AND status = ? AND cas_version = ?
    `
    
    // 先查询当前版本号
    order, err := db.GetOrder(ctx, orderID)
    if err != nil {
        return err
    }
    
    if order.Status != oldStatus {
        return ErrInvalidStatusTransition
    }
    
    // CAS更新
    result, err := db.Exec(ctx, query, newStatus, time.Now(), orderID, oldStatus, order.CASVersion)
    if err != nil {
        return err
    }
    
    if result.RowsAffected == 0 {
        // 更新失败，可能被其他请求修改了
        return ErrConcurrentUpdate
    }
    
    return nil
}
```

**本地消息表 + Kafka事件发布**：

```go
// Outbox Pattern：保证消息一定发送
func PublishOrderCreatedEvent(ctx context.Context, order *Order) error {
    // 1. 在同一事务中：插入订单 + 插入本地消息
    tx, err := db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // 插入订单
    if err := tx.InsertOrder(ctx, order); err != nil {
        return err
    }
    
    // 插入本地消息
    msg := &OutboxMessage{
        MessageID: uuid.New().String(),
        Topic:     "order.created",
        Payload:   json.Marshal(order),
        Status:    OutboxPending,
    }
    if err := tx.InsertOutboxMessage(ctx, msg); err != nil {
        return err
    }
    
    // 提交事务
    if err := tx.Commit(); err != nil {
        return err
    }
    
    // 2. 异步发送Kafka消息（定时任务扫描本地消息表）
    // 这里只是插入，实际发送由后台任务完成
    return nil
}

// 后台任务：扫描本地消息表并发送
func OutboxMessageSender() {
    ticker := time.NewTicker(1 * time.Second)
    for range ticker.C {
        messages, _ := db.GetPendingOutboxMessages(context.Background(), 100)
        for _, msg := range messages {
            if err := kafkaProducer.Send(msg.Topic, msg.Payload); err != nil {
                log.Error("failed to send kafka message", "error", err)
                continue
            }
            // 发送成功，更新状态
            db.UpdateOutboxMessageStatus(context.Background(), msg.MessageID, OutboxSent)
        }
    }
}
```
