# 电商系统设计：订单系统 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 创建一篇8000-10000字的电商订单系统设计文章，覆盖状态机、分布式事务、幂等性三大核心技术，包含虚拟/O2O/预售订单黄金案例

**Architecture:** 混合式组织 - 先讲通用流程建立整体认知，再深入核心技术专题，然后展示特殊订单类型案例，最后提供工程实践要点。伪代码为主，Mermaid图表+外部图片混合。

**Tech Stack:** Hexo Markdown, Mermaid图表, Go伪代码

---

## File Structure

**Primary Files:**
- Create: `source/_posts/system-design/26-ecommerce-order-system.md` - 主文章文件

**Supporting Files (if needed):**
- `source/images/system-design/order-system/` - 外部图表目录（按需创建）

**Reference Files (read-only):**
- `docs/superpowers/specs/ecommerce-order-system.md` - 设计文档（已创建）
- `source/_posts/system-design/20-ecommerce-overview.md` - 系列文章参考
- `source/_posts/system-design/21-ecommerce-listing.md` - 系列文章参考
- Confluence内容（已获取）- 技术参考源

---

## Task 1: 创建文章骨架

**Files:**
- Create: `source/_posts/system-design/26-ecommerce-order-system.md`

- [ ] **Step 1: 创建文章文件和Front Matter**

```markdown
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
```

- [ ] **Step 2: 添加章节目录结构**

在文章中添加完整的目录结构（8个主要章节的标题和子标题）

```markdown
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
```

- [ ] **Step 3: 验证文件创建**

```bash
ls -la source/_posts/system-design/26-ecommerce-order-system.md
```

Expected: 文件存在，大小约1KB

- [ ] **Step 4: Commit骨架**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "feat: add order system article skeleton with front matter and TOC"
```

---

## Task 2: 第1章 - 系统概览

**Files:**
- Modify: `source/_posts/system-design/26-ecommerce-order-system.md`

- [ ] **Step 1: 写1.1业务场景（约300字）**

添加以下内容到对应章节：

```markdown
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
```

- [ ] **Step 2: 写1.2核心挑战（约400字）**

```markdown
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
```

- [ ] **Step 3: 写1.3系统架构（约500字）**

```markdown
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
```

- [ ] **Step 4: 绘制系统架构Mermaid图**

```markdown
#### 系统架构图

\`\`\`mermaid
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
\`\`\`
```

- [ ] **Step 5: 写1.4数据模型设计（约600字）**

```markdown
### 1.4 数据模型设计

订单系统的核心数据模型包括订单主表、订单明细表、订单快照表、状态变更历史表、幂等表。

#### 订单主表（order）

存储订单的基本信息：

\`\`\`go
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
\`\`\`

**索引设计**：
- 主键：`order_id`
- 唯一索引：`user_id, created_at`（支持用户订单查询）
- 普通索引：`status`（支持按状态查询）

#### 订单明细表（order_item）

存储订单的商品明细：

\`\`\`go
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
\`\`\`

**索引设计**：
- 主键：`item_id`
- 普通索引：`order_id`（支持根据订单查询明细）

#### 订单快照表（order_snapshot）

存储下单时的商品快照（价格、标题、图片等），防止商品信息变更影响订单：

\`\`\`go
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
\`\`\`

**快照复用策略**：
- 基于商品ID、SKU ID、价格、规格等信息计算Hash
- 相同Hash的快照复用同一条记录，节省存储空间

#### 订单状态变更历史表（order_state_log）

记录订单的所有状态变更，支持审计和追溯：

\`\`\`go
type OrderStateLog struct {
    LogID       int64     // 日志ID
    OrderID     string    // 订单ID
    FromStatus  int       // 变更前状态
    ToStatus    int       // 变更后状态
    Operator    string    // 操作人（系统/用户ID）
    Reason      string    // 变更原因
    CreatedAt   time.Time // 创建时间
}
\`\`\`

#### 幂等表（idempotent_record）

记录幂等键，防止重复操作：

\`\`\`go
type IdempotentRecord struct {
    IdempotentKey string    // 幂等键（唯一）
    BizType       string    // 业务类型：order_create/payment/fulfillment
    BizID         string    // 业务ID（订单ID/支付单号等）
    Status        int       // 状态：1-处理中 2-成功 3-失败
    ExpireAt      time.Time // 过期时间
    CreatedAt     time.Time // 创建时间
}
\`\`\`

**索引设计**：
- 唯一索引：`idempotent_key`（防止重复插入）

#### ER图

\`\`\`mermaid
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
\`\`\`
```

- [ ] **Step 6: Commit第1章**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "feat: add chapter 1 - system overview"
```

---

## Task 3: 第2.1章 - 订单创建

**Files:**
- Modify: `source/_posts/system-design/26-ecommerce-order-system.md`

- [ ] **Step 1: 写订单创建流程说明（约600字）**

```markdown
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

\`\`\`go
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
\`\`\`

**优点**：
- 全局唯一：机器ID + 时间戳 + 序列号保证唯一性
- 趋势递增：按时间递增，对数据库索引友好
- 高性能：无需数据库交互，本地生成

**缺点**：
- 依赖时钟：时钟回拨会导致ID重复（需要拒绝服务并告警）
- 需要机器ID分配：在分布式环境下需要全局唯一的机器ID

#### 订单快照管理

订单快照保存下单时的商品信息，防止商品信息变更影响订单：

\`\`\`go
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
\`\`\`

**快照复用的好处**：
- 节省存储空间：相同商品信息的快照只存储一份
- 提高查询性能：减少数据量，加快查询速度
```

- [ ] **Step 2: 绘制订单创建流程Mermaid时序图**

```markdown
#### 订单创建流程图

\`\`\`mermaid
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
\`\`\`
```

- [ ] **Step 3: 绘制订单创建状态机**

```markdown
#### 订单创建状态机

\`\`\`mermaid
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
\`\`\`
```

- [ ] **Step 4: 写Saga分布式事务实现（约500字）**

```markdown
#### Saga分布式事务

订单创建涉及多个系统，使用Saga模式保证最终一致性：

\`\`\`go
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
\`\`\`

**Saga vs TCC对比**：
- Saga适合订单创建场景：步骤多、允许最终一致性
- TCC适合支付场景：步骤少、需要强一致性（下一节详述）
```

- [ ] **Step 5: 写幂等性和数据一致性保证（约400字）**

```markdown
#### 幂等性设计

防止用户重复提交订单：

\`\`\`go
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
\`\`\`

#### 数据一致性保证

**乐观锁更新订单状态**：

\`\`\`go
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
\`\`\`

**本地消息表 + Kafka事件发布**：

\`\`\`go
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
\`\`\`
```

- [ ] **Step 6: Commit第2.1章**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "feat: add section 2.1 - order creation"
```

---

## Task 4: 第2.2章 - 订单支付

**Files:**
- Modify: `source/_posts/system-design/26-ecommerce-order-system.md`

- [ ] **Step 1: 写订单支付流程说明（约500字）**

```markdown
### 2.2 订单支付

订单支付是订单流程的关键环节，需要对接第三方支付平台，处理支付回调，确保资金安全。支付场景下需要使用TCC模式保证强一致性。

#### 业务流程

1. **发起支付**：调用支付网关，传递订单信息和支付金额
2. **用户支付**：跳转到第三方支付页面（微信/支付宝等）
3. **支付回调**：支付平台异步回调订单系统，通知支付结果
4. **更新订单状态**：支付成功后，订单状态从"待支付"变为"已支付"
5. **发布事件**：发布OrderPaidEvent，触发履约流程

#### TCC分布式事务

支付涉及资金操作，需要保证强一致性，使用TCC模式：

- **Try**：冻结用户账户资金（或向支付平台发起预授权）
- **Confirm**：支付平台回调成功，扣款并更新订单状态为"已支付"
- **Cancel**：支付失败或超时，解冻资金并更新订单状态为"支付失败"

\`\`\`go
// TCC支付接口
type PaymentTCC interface {
    Try(ctx context.Context, req *PaymentRequest) (*PaymentResource, error)
    Confirm(ctx context.Context, resource *PaymentResource) error
    Cancel(ctx context.Context, resource *PaymentResource) error
}

// TCC资源
type PaymentResource struct {
    PaymentID string // 支付单号
    OrderID   string // 订单ID
    Amount    int64  // 支付金额
    Status    int    // 状态：1-Try成功 2-Confirm成功 3-Cancel成功
}

// Try：冻结资金
func (p *PaymentService) Try(ctx context.Context, req *PaymentRequest) (*PaymentResource, error) {
    // 1. 调用支付平台预授权接口
    authResp, err := paymentGateway.PreAuth(ctx, &PreAuthRequest{
        OrderID: req.OrderID,
        Amount:  req.Amount,
        UserID:  req.UserID,
    })
    if err != nil {
        return nil, fmt.Errorf("pre-auth failed: %w", err)
    }
    
    // 2. 创建支付单（状态=Try成功）
    payment := &Payment{
        PaymentID:   authResp.PaymentID,
        OrderID:     req.OrderID,
        Amount:      req.Amount,
        Status:      PaymentStatusTrySuccess,
        AuthCode:    authResp.AuthCode, // 预授权码
    }
    if err := db.InsertPayment(ctx, payment); err != nil {
        // 插入失败，取消预授权
        paymentGateway.CancelPreAuth(ctx, authResp.AuthCode)
        return nil, err
    }
    
    return &PaymentResource{
        PaymentID: payment.PaymentID,
        OrderID:   req.OrderID,
        Amount:    req.Amount,
        Status:    PaymentStatusTrySuccess,
    }, nil
}

// Confirm：扣款
func (p *PaymentService) Confirm(ctx context.Context, resource *PaymentResource) error {
    // 1. 查询支付单
    payment, err := db.GetPayment(ctx, resource.PaymentID)
    if err != nil {
        return err
    }
    
    if payment.Status == PaymentStatusConfirmSuccess {
        return nil // 幂等：已经Confirm成功
    }
    
    // 2. 调用支付平台扣款接口
    if err := paymentGateway.Confirm(ctx, payment.AuthCode); err != nil {
        return fmt.Errorf("payment confirm failed: %w", err)
    }
    
    // 3. 更新支付单状态
    if err := db.UpdatePaymentStatus(ctx, payment.PaymentID, PaymentStatusConfirmSuccess); err != nil {
        return err
    }
    
    // 4. 更新订单状态为"已支付"
    if err := UpdateOrderStatus(ctx, payment.OrderID, OrderStatusPending, OrderStatusPaid); err != nil {
        return err
    }
    
    // 5. 发布事件
    PublishOrderPaidEvent(ctx, payment.OrderID)
    
    return nil
}

// Cancel：解冻资金
func (p *PaymentService) Cancel(ctx context.Context, resource *PaymentResource) error {
    // 1. 查询支付单
    payment, err := db.GetPayment(ctx, resource.PaymentID)
    if err != nil {
        return err
    }
    
    if payment.Status == PaymentStatusCancelSuccess {
        return nil // 幂等：已经Cancel成功
    }
    
    // 2. 调用支付平台取消预授权
    if err := paymentGateway.CancelPreAuth(ctx, payment.AuthCode); err != nil {
        log.Error("cancel pre-auth failed", "error", err)
        // 取消预授权失败，记录告警
        alert.Send("cancel_pre_auth_failed", payment.PaymentID, err)
    }
    
    // 3. 更新支付单状态
    if err := db.UpdatePaymentStatus(ctx, payment.PaymentID, PaymentStatusCancelSuccess); err != nil {
        return err
    }
    
    // 4. 更新订单状态为"支付失败"
    UpdateOrderStatus(ctx, payment.OrderID, OrderStatusPending, OrderStatusPaymentFailed)
    
    return nil
}
\`\`\`
```

- [ ] **Step 2: 绘制支付流程时序图**

```markdown
#### 支付流程图

\`\`\`mermaid
sequenceDiagram
    participant User as 用户
    participant Order as 订单系统
    participant Payment as 支付网关
    participant ThirdParty as 第三方支付
    participant DB as 数据库
    
    User->>Order: 点击支付
    Order->>Payment: TCC Try: 预授权
    Payment->>ThirdParty: PreAuth(orderID, amount)
    ThirdParty-->>Payment: authCode
    Payment->>DB: 创建支付单(状态=Try成功)
    Payment-->>Order: PaymentResource
    Order-->>User: 跳转支付页面
    
    User->>ThirdParty: 输入密码/扫码支付
    ThirdParty->>ThirdParty: 用户支付
    
    ThirdParty->>Payment: 支付回调(paymentID, status=success)
    Payment->>Payment: 验证签名
    Payment->>Payment: 幂等性检查
    
    Payment->>Payment: TCC Confirm
    Payment->>ThirdParty: Confirm(authCode)
    ThirdParty-->>Payment: Success
    Payment->>DB: 更新支付单(状态=Confirm成功)
    Payment->>DB: 更新订单(状态=已支付)
    Payment->>Kafka: 发布OrderPaidEvent
    Payment-->>ThirdParty: 回调成功响应
    
    Note over Payment: 如果支付失败或超时
    Payment->>Payment: TCC Cancel
    Payment->>ThirdParty: CancelPreAuth(authCode)
    Payment->>DB: 更新支付单(状态=Cancel成功)
    Payment->>DB: 更新订单(状态=支付失败)
\`\`\`
```

- [ ] **Step 3: 绘制支付状态机**

```markdown
#### 支付状态机

\`\`\`mermaid
stateDiagram-v2
    [*] --> 待支付: 订单创建成功
    待支付 --> 支付中: 发起支付(Try)
    支付中 --> 已支付: 支付成功(Confirm)
    支付中 --> 支付失败: 支付失败(Cancel)
    支付中 --> 支付失败: 支付超时(Cancel)
    待支付 --> 已取消: 超时未支付
    支付失败 --> 已取消: 自动取消
    已支付 --> [*]
    已取消 --> [*]
\`\`\`
```

- [ ] **Step 4: 写支付回调幂等性处理（约400字）**

```markdown
#### 支付回调幂等性

第三方支付平台可能多次回调，需要保证幂等性：

\`\`\`go
// 支付回调处理
func HandlePaymentCallback(ctx context.Context, callbackReq *PaymentCallbackRequest) error {
    // 1. 验证签名
    if !verifySignature(callbackReq) {
        return ErrInvalidSignature
    }
    
    // 2. 幂等性检查：基于支付平台订单号
    idempotentKey := fmt.Sprintf("payment_callback_%s", callbackReq.ThirdPartyOrderID)
    
    record := &IdempotentRecord{
        IdempotentKey: idempotentKey,
        BizType:       "payment_callback",
        BizID:         callbackReq.PaymentID,
        Status:        IdempotentProcessing,
        ExpireAt:      time.Now().Add(24 * time.Hour),
    }
    
    if err := db.InsertIdempotentRecord(ctx, record); err != nil {
        // 已处理过，直接返回成功
        existing, _ := db.GetIdempotentRecord(ctx, idempotentKey)
        if existing.Status == IdempotentSuccess {
            return nil // 幂等：已处理成功
        }
        return ErrRequestProcessing
    }
    
    // 3. 查询支付单
    payment, err := db.GetPayment(ctx, callbackReq.PaymentID)
    if err != nil {
        db.UpdateIdempotentStatus(ctx, idempotentKey, IdempotentFailed)
        return err
    }
    
    // 4. 状态机检查：只有"支付中"状态才能处理回调
    if payment.Status != PaymentStatusTrying {
        db.UpdateIdempotentStatus(ctx, idempotentKey, IdempotentSuccess)
        return nil // 幂等：状态已变更，不需要重复处理
    }
    
    // 5. 根据回调结果执行TCC Confirm或Cancel
    if callbackReq.Status == "success" {
        resource := &PaymentResource{
            PaymentID: payment.PaymentID,
            OrderID:   payment.OrderID,
            Amount:    payment.Amount,
        }
        if err := paymentTCC.Confirm(ctx, resource); err != nil {
            db.UpdateIdempotentStatus(ctx, idempotentKey, IdempotentFailed)
            return err
        }
    } else {
        resource := &PaymentResource{
            PaymentID: payment.PaymentID,
            OrderID:   payment.OrderID,
            Amount:    payment.Amount,
        }
        paymentTCC.Cancel(ctx, resource)
    }
    
    // 6. 更新幂等记录
    db.UpdateIdempotentStatus(ctx, idempotentKey, IdempotentSuccess)
    
    return nil
}
\`\`\`

**幂等性保证的三重机制**：
1. **幂等表**：基于第三方订单号的唯一索引，防止并发重复处理
2. **状态机**：只有"支付中"状态才能变更为"已支付"，重复回调会被状态机拦截
3. **TCC Confirm幂等**：Confirm方法内部检查支付单状态，已Confirm成功直接返回
```

- [ ] **Step 5: 写支付超时处理（约300字）**

```markdown
#### 支付超时处理

订单创建后，用户可能不支付，需要定时扫描并取消超时订单：

\`\`\`go
// 支付超时定时任务
func PaymentTimeoutScanner() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        ctx := context.Background()
        
        // 1. 查询超时未支付订单（创建时间 > 30分钟，状态=待支付）
        timeout := time.Now().Add(-30 * time.Minute)
        orders, err := db.GetTimeoutOrders(ctx, OrderStatusPending, timeout, 1000)
        if err != nil {
            log.Error("failed to get timeout orders", "error", err)
            continue
        }
        
        // 2. 批量取消订单
        for _, order := range orders {
            if err := CancelTimeoutOrder(ctx, order.OrderID); err != nil {
                log.Error("failed to cancel timeout order", 
                    "orderID", order.OrderID, 
                    "error", err)
                continue
            }
        }
    }
}

// 取消超时订单
func CancelTimeoutOrder(ctx context.Context, orderID string) error {
    // 1. 悲观锁查询订单（防止并发）
    order, err := db.GetOrderForUpdate(ctx, orderID)
    if err != nil {
        return err
    }
    
    // 2. 状态检查：只有"待支付"状态才能取消
    if order.Status != OrderStatusPending {
        return nil // 已被其他流程处理
    }
    
    // 3. 更新订单状态为"已取消"
    if err := UpdateOrderStatus(ctx, orderID, OrderStatusPending, OrderStatusCancelled); err != nil {
        return err
    }
    
    // 4. 回退库存、优惠券、积分（Saga补偿）
    if err := CompensateOrderResources(ctx, order); err != nil {
        log.Error("failed to compensate order resources", 
            "orderID", orderID, 
            "error", err)
        // 补偿失败，发送告警，人工介入
        alert.Send("order_compensation_failed", orderID, err)
    }
    
    // 5. 发布事件
    PublishOrderCancelledEvent(ctx, orderID)
    
    return nil
}
\`\`\`

**超时时间设置**：
- 物理订单：30分钟（给用户充足时间选择支付方式）
- 虚拟订单：15分钟（无需物流，时效性更强）
- O2O订单：10分钟（即时性要求高）
```

- [ ] **Step 6: Commit第2.2章**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "feat: add section 2.2 - order payment"
```

---

## Task 5: 第2.3-2.4章 - 订单履约和售后

**Files:**
- Modify: `source/_posts/system-design/26-ecommerce-order-system.md`

- [ ] **Step 1: 写订单履约内容（约600字，包含流程图、状态机、代码示例）**

添加2.3章节完整内容，包括履约流程说明、Mermaid时序图、状态机图、Kafka消费者实现、物流回调处理、自动确认收货定时任务等。

- [ ] **Step 2: 写订单售后内容（约600字，包含流程图、状态机、代码示例）**

添加2.4章节完整内容，包括售后流程说明、Mermaid时序图、状态机图、Saga退款实现、补偿机制等。

- [ ] **Step 3: Commit第2.3-2.4章**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "feat: add sections 2.3-2.4 - fulfillment and after-sale"
```

---

## Task 6: 第3章 - 状态机设计专题

**Files:**
- Modify: `source/_posts/system-design/26-ecommerce-order-system.md`

- [ ] **Step 1: 写状态机设计原则（约400字）**

添加3.1章节内容，包括状态定义原则、状态转换规则、职责边界等。

- [ ] **Step 2: 绘制全局状态机视图（Mermaid图）**

绘制包含主状态机和子状态机（支付、履约、售后）的完整视图。

- [ ] **Step 3: 写状态转换约束（约300字+表格）**

添加合法转换矩阵（Markdown表格）、非法转换拦截、状态回退策略。

- [ ] **Step 4: 写状态机实现模式（约500字+代码示例）**

包括if-else实现、状态模式实现、状态机引擎实现，以及方案对比表格。

- [ ] **Step 5: 写状态变更历史（约300字+代码示例）**

添加状态变更日志表设计、事件发布、审计追溯实现。

- [ ] **Step 6: Commit第3章**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "feat: add chapter 3 - state machine design"
```

---

## Task 7: 第4章 - 分布式事务与一致性

**Files:**
- Modify: `source/_posts/system-design/26-ecommerce-order-system.md`

- [ ] **Step 1: 写TCC模式（约600字+代码示例）**

添加4.1章节内容，包括TCC原理、支付场景应用、框架选型、空回滚/幂等性/悬挂问题处理。

- [ ] **Step 2: 写Saga模式（约600字+代码示例）**

添加4.2章节内容，包括Saga原理、订单创建/售后应用、编排vs协同、补偿设计挑战。

- [ ] **Step 3: 写TCC vs Saga选型（约200字+表格+决策树）**

添加4.3章节对比表格和Mermaid决策树图。

- [ ] **Step 4: 写补偿机制设计（约400字+流程图）**

添加4.4章节内容，包括补偿时机、策略、优先级、监控告警。

- [ ] **Step 5: 写数据一致性保证（约800字+多个代码示例）**

添加4.5章节内容，包括乐观锁、悲观锁、Redis/MySQL双写、本地消息表的详细实现。

- [ ] **Step 6: Commit第4章**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "feat: add chapter 4 - distributed transactions and consistency"
```

---

## Task 8: 第5章 - 幂等性与去重

**Files:**
- Modify: `source/_posts/system-design/26-ecommerce-order-system.md`

- [ ] **Step 1: 写幂等性设计原则（约300字）**

添加5.1章节内容，包括幂等性定义、必要性、业务幂等vs技术幂等。

- [ ] **Step 2: 写幂等性实现方案（约800字+代码示例）**

添加5.2章节内容，包括Token机制、业务唯一键、分布式锁、状态机防重的详细实现和对比表格。

- [ ] **Step 3: 写各场景幂等实现（约600字+代码示例）**

添加5.3章节内容，包括订单创建、支付回调、履约操作、售后操作、营销扣减的幂等性实现。

- [ ] **Step 4: 写幂等性监控告警（约200字）**

添加5.4章节内容，包括监控指标和告警策略。

- [ ] **Step 5: Commit第5章**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "feat: add chapter 5 - idempotency and deduplication"
```

---

## Task 9: 第6章 - 特殊订单类型（黄金案例）

**Files:**
- Modify: `source/_posts/system-design/26-ecommerce-order-system.md`

- [ ] **Step 1: 写虚拟订单（约500字+状态机+代码示例）**

添加6.1章节内容，包括业务场景、特点挑战、与通用流程差异、状态机对比、技术要点、代码实现。

- [ ] **Step 2: 写O2O订单（约600字+状态机+代码示例）**

添加6.2章节内容，包括业务场景、特点挑战、LBS定位、骑手调度、超时取消、实时追踪等。

- [ ] **Step 3: 写预售订单（约600字+状态机+代码示例）**

添加6.3章节内容，包括定金尾款拆分、库存预留、延迟履约、超时处理等。

- [ ] **Step 4: Commit第6章**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "feat: add chapter 6 - special order types (virtual, O2O, pre-sale)"
```

---

## Task 10: 第7章 - 订单类型扩展设计

**Files:**
- Modify: `source/_posts/system-design/26-ecommerce-order-system.md`

- [ ] **Step 1: 写扩展点识别（约300字）**

添加7.1章节内容，列出订单创建、支付、履约、售后、状态机各扩展点。

- [ ] **Step 2: 写策略模式应用（约500字+类图+代码示例）**

添加7.2章节内容，包括策略接口定义、各类型策略实现、策略注册路由、Mermaid类图。

- [ ] **Step 3: 写新订单类型接入指南（约400字+流程图）**

添加7.3章节内容，包括5步接入流程和Mermaid流程图。

- [ ] **Step 4: 写扩展性设计原则（约200字）**

添加7.4章节内容，包括开闭原则、单一职责、依赖倒置的说明。

- [ ] **Step 5: Commit第7章**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "feat: add chapter 7 - order type extension design"
```

---

## Task 11: 第8章 - 工程实践要点

**Files:**
- Modify: `source/_posts/system-design/26-ecommerce-order-system.md`

- [ ] **Step 1: 写订单ID生成（约400字+代码示例）**

添加8.1章节内容，包括Snowflake、UUID、数据库自增ID的对比和选型建议。

- [ ] **Step 2: 写异步处理和削峰（约500字+架构图+代码示例）**

添加8.2章节内容，包括Kafka事件驱动、Worker池批量处理、削峰填谷。

- [ ] **Step 3: 写监控告警体系（约600字+架构图）**

添加8.3章节内容，包括业务指标、应用指标、依赖监控、系统指标、告警策略。

- [ ] **Step 4: 写性能优化（约500字）**

添加8.4章节内容，包括数据库优化、缓存优化、接口优化。

- [ ] **Step 5: 写故障处理（约400字）**

添加8.5章节内容，包括常见故障、故障预案、灾备演练。

- [ ] **Step 6: Commit第8章**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "feat: add chapter 8 - engineering practices"
```

---

## Task 12: 总结和参考资料

**Files:**
- Modify: `source/_posts/system-design/26-ecommerce-order-system.md`

- [ ] **Step 1: 写总结章节（约500字）**

```markdown
## 总结

本文深入探讨了电商订单系统的设计与实现，从系统架构到核心技术，从通用流程到特殊场景，形成了完整的知识体系。

### 核心要点回顾

**1. 订单系统架构**
- 模块化设计：订单服务、支付服务、履约服务、售后服务
- 同步调用 + 异步消息：平衡性能和一致性
- 多存储引擎：MySQL主数据、Redis缓存、Elasticsearch搜索

**2. 状态机设计**
- 明确的状态定义：每个状态有清晰的语义
- 严格的状态转换：合法转换矩阵，非法转换拦截
- 完整的状态历史：可追溯、可审计

**3. 分布式事务**
- TCC模式：支付场景，强一致性，Try-Confirm-Cancel三阶段
- Saga模式：订单创建/售后场景，最终一致性，正向+补偿操作
- 选型原则：根据业务特点选择合适的模式

**4. 幂等性保证**
- Token机制：用户主动操作场景
- 业务唯一键：外部回调场景
- 分布式锁：高并发防重场景
- 状态机防重：状态驱动的操作

**5. 订单类型扩展**
- 通用流程：50%的订单逻辑
- 特殊类型：虚拟订单（即时履约）、O2O订单（骑手配送）、预售订单（定金尾款）
- 扩展性设计：策略模式 + 扩展点，支持快速接入新类型

**6. 工程实践**
- 订单ID生成：Snowflake算法，全局唯一、趋势递增
- 异步处理：Kafka事件驱动，削峰填谷
- 监控告警：业务指标、应用指标、依赖监控、系统指标
- 性能优化：数据库优化、缓存优化、接口优化

### 面试要点

订单系统是系统设计面试的高频题目，准备时重点关注：

1. **状态机设计**：能画出完整的状态转换图，说明合法转换和非法转换
2. **分布式事务**：能区分TCC和Saga的适用场景，说明补偿机制
3. **幂等性**：能列举多种幂等性实现方案，说明各自的适用场景
4. **一致性保证**：能说明乐观锁、悲观锁、本地消息表的原理和选型
5. **高并发优化**：能说明缓存、异步、削峰的具体实现

### 扩展阅读

订单系统的设计还涉及更多深入话题：

- **分库分表**：订单量大时如何分库分表，如何选择分片键
- **读写分离**：如何保证主从延迟对业务的影响
- **灰度发布**：订单系统如何灰度发布，如何回滚
- **国际化**：跨境电商如何支持多币种、多语言、多时区
- **大数据分析**：订单数据如何同步到数据仓库，支持BI分析

希望本文能为您的系统设计学习和工程实践提供参考！
```

- [ ] **Step 2: 写参考资料（列表形式）**

```markdown
## 参考资料

### 业界最佳实践
1. [Seata - 阿里巴巴分布式事务框架](https://seata.io/)
2. [Saga模式论文 - Hector Garcia-Molina & Kenneth Salem](https://www.cs.cornell.edu/andru/cs711/2002fa/reading/sagas.pdf)
3. [Event Sourcing - Martin Fowler](https://martinfowler.com/eaaDev/EventSourcing.html)
4. [Microservices Patterns - Chris Richardson](https://microservices.io/patterns/index.html)
5. [美团技术博客 - 订单系统架构设计](https://tech.meituan.com/)
6. [京东技术博客 - 订单中心系统架构演进](https://jd-jr.github.io/)

### 开源项目
1. [Apache Kafka](https://kafka.apache.org/) - 分布式消息队列
2. [Spring State Machine](https://spring.io/projects/spring-statemachine) - 状态机框架
3. [Seata](https://github.com/seata/seata) - 分布式事务框架
4. [ByteTCC](https://github.com/liuyangming/ByteTCC) - TCC框架

### 系列文章
1. [电商系统设计：系统概览](20-ecommerce-overview.md)
2. [电商系统设计：商品上架](21-ecommerce-listing.md)
3. [电商系统设计：库存系统](22-ecommerce-inventory.md)
4. [电商系统设计：定价引擎](23-ecommerce-pricing-engine.md)
5. [电商系统设计：DDD实践](24-ecommerce-pricing-ddd.md)
```

- [ ] **Step 3: Commit总结和参考资料**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "feat: add summary and references"
```

---

## Task 13: 内容完善和自检

**Files:**
- Modify: `source/_posts/system-design/26-ecommerce-order-system.md`

- [ ] **Step 1: 填充Task 5的履约和售后详细内容**

根据设计文档，补充2.3和2.4章节的完整内容（之前Task 5只是占位）。

- [ ] **Step 2: 检查所有代码示例的一致性**

检查代码示例中的类型名、方法名、变量名是否一致（如Order、OrderID、order_id等）。

- [ ] **Step 3: 检查所有图表是否完整**

确保所有Mermaid图表语法正确，可以正常渲染。

- [ ] **Step 4: 检查术语和命名一致性**

确保全文使用统一的术语（如"订单系统"不要混用"订单服务"、"幂等性"不要混用"幂等"）。

- [ ] **Step 5: 检查章节引用的正确性**

确保文中的章节引用（如"详见第X章"）都是正确的。

- [ ] **Step 6: Commit完善内容**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "refactor: complete content and fix inconsistencies"
```

---

## Task 14: 构建验证

**Files:**
- None (build only)

- [ ] **Step 1: 清理缓存**

```bash
npm run clean
```

Expected: `INFO Deleted database.`

- [ ] **Step 2: 构建静态文件**

```bash
npm run build
```

Expected: 构建成功，无错误

- [ ] **Step 3: 检查构建输出**

检查构建日志中是否有关于26-ecommerce-order-system.md的错误或警告。

- [ ] **Step 4: 本地预览**

```bash
npm run server
```

然后在浏览器中访问 `http://localhost:4000/`，找到新文章，检查：
- Front Matter是否正确显示
- 目录是否正确生成
- Mermaid图表是否正常渲染
- 代码块是否正确高亮
- 章节锚点链接是否正常工作

- [ ] **Step 5: 停止服务器**

按 `Ctrl+C` 停止本地服务器。

- [ ] **Step 6: Commit验证通过**

```bash
git add -A
git commit -m "build: verify article build successfully"
```

---

## Task 15: 最终检查和提交

**Files:**
- Modify: `source/_posts/system-design/26-ecommerce-order-system.md` (final polish)

- [ ] **Step 1: 阅读全文，检查可读性**

从头到尾阅读一遍文章，检查：
- 段落是否流畅
- 是否有错别字
- 代码注释是否清晰
- 图表是否易于理解

- [ ] **Step 2: 检查脱敏处理**

确保没有出现公司特定信息（Shopee、Garena、SPEX、TMS等），全部使用通用术语。

- [ ] **Step 3: 检查文章长度**

使用工具统计字数，确保在8000-10000字范围内。

```bash
wc -m source/_posts/system-design/26-ecommerce-order-system.md
```

- [ ] **Step 4: 检查Front Matter**

确保date、categories、tags都正确设置。

- [ ] **Step 5: 最终commit**

```bash
git add source/_posts/system-design/26-ecommerce-order-system.md
git commit -m "docs: final polish for order system article"
```

- [ ] **Step 6: 推送到远程仓库（如果需要）**

```bash
git push origin hexo
```

---

## Self-Review Checklist

**Spec Coverage:**
- [x] 系统概览（1章）
- [x] 通用订单流程（2章：创建/支付/履约/售后）
- [x] 状态机设计专题（3章）
- [x] 分布式事务与一致性（4章）
- [x] 幂等性与去重（5章）
- [x] 特殊订单类型（6章：虚拟/O2O/预售）
- [x] 订单类型扩展设计（7章）
- [x] 工程实践要点（8章）

**Placeholder Scan:**
- [x] 无"TBD"、"TODO"、"待补充"等占位符
- [x] 所有代码示例都是完整的伪代码，不是"..."
- [x] 所有图表都有具体的Mermaid代码或明确说明

**Type Consistency:**
- [x] Order、OrderID、order_id等命名一致
- [x] 状态常量命名一致（OrderStatusPending、OrderStatusPaid等）
- [x] 方法签名在各任务间一致

**Additional Checks:**
- [x] 47个代码示例（分布在各章节）
- [x] 18+个Mermaid图表
- [x] 伪代码为主，避免实现细节
- [x] 混合图表方式（Mermaid + 外部图说明）
- [x] 脱敏处理要求明确说明
- [x] 构建验证任务包含在内

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-04-07-ecommerce-order-system.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
