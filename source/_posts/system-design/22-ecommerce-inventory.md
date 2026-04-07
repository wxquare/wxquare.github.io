---
title: 多品类统一库存系统设计：电商·虚拟商品·本地生活
date: 2025-06-28
categories:
- 系统设计
tags:
- 库存系统
- 电商
- 系统设计
- 高并发
toc: true
---

<!-- toc -->

## 一、背景与挑战

### 1.1 多品类库存差异

在数字电商/本地生活平台中，不同品类的库存特性差异极大：

| 品类 | 库存特点 | 扣减时机 | 典型示例 |
|------|----------|----------|----------|
| **电子券 (Deal)** | 券码制，每个券码唯一 | 下单预订 | 星巴克电子券 |
| **虚拟服务券 (OPV)** | 数量制，分平台统计 | 下单预订 | 美甲/按摩服务券 |
| **酒店** | 时间维度，按日期管理 | 支付成功 | Agoda 酒店房间 |
| **机票/票务** | 座位/场次制 | 支付成功 | 航班座位、电影票 |
| **礼品卡 (Giftcard)** | 实时生成或预采购卡密 | 支付成功 | Google Play 充值卡 |
| **话费充值 (TopUp)** | 无限库存 | 无需扣减 | 手机话费 |
| **本地生活套餐** | 组合型，多子项联动 | 下单预订 | 火锅双人套餐 |

### 1.2 核心痛点

1. **模型割裂**：每个品类独立设计库存逻辑，无法复用。
2. **数据不一致**：Redis 与 MySQL 之间、预订数量 (booking) 与实际状态脱节。
3. **供应商策略不统一**：有的实时查询，有的定时同步，有的无需管理。
4. **缺乏统一服务**：各业务方直接操作 DB/Redis，维护成本高。
5. **监控缺失**：超卖、库存差异、供应商同步延迟难以发现。

### 1.3 设计目标

| 目标 | 说明 | 优先级 |
|------|------|--------|
| **统一模型** | 多品类共用一套库存模型 | P0 |
| **高性能** | 支持万级 QPS 秒杀场景 | P0 |
| **灵活扩展** | 新品类接入无需修改核心代码 | P0 |
| **最终一致** | Redis 与 MySQL 数据最终一致 | P0 |
| **供应商集成** | 支持实时/定时/推送多种同步策略 | P1 |

---

## 二、库存分类体系

### 2.1 两个核心维度

设计统一库存模型的关键是将所有品类抽象为 **两个正交维度**：

**维度一：谁管库存？（Management Type）**

```go
const (
    SelfManaged     = 1 // 自管理：平台维护库存数据
    SupplierManaged = 2 // 供应商管理：第三方维护，平台定期同步
    Unlimited       = 3 // 无限库存：无需库存管理
)
```

**维度二：库存长什么样？（Unit Type）**

```go
const (
    CodeBased     = 1 // 券码制：每个库存是唯一券码（Deal、Giftcard）
    QuantityBased = 2 // 数量制：库存是一个数字（OPV、本地服务）
    TimeBased     = 3 // 时间维度：按日期/时段管理（酒店、票务）
    BundleBased   = 4 // 组合型：多子项联动扣减（套餐）
)
```

### 2.2 品类分类矩阵

| 品类 | 管理类型 | 单元类型 | 扣减时机 |
|------|----------|----------|----------|
| 电子券 (Deal) | Self | Code | 下单 |
| 虚拟服务券 (OPV) | Self | Quantity | 下单 |
| 本地服务 | Self | Quantity | 下单 |
| 酒店 | Supplier | Time | 支付 |
| 机票 | Supplier | Quantity | 支付 |
| 话费充值 | Unlimited | - | 无 |
| 礼品卡(预采购) | Self | Code | 下单 |
| 礼品卡(实时生成) | Supplier | Code | 支付 |
| 套餐组合 | Self | Bundle | 下单 |

> **核心洞察**：任何新品类接入时，只需确定它属于哪个 `(ManagementType, UnitType)` 组合，即可复用对应的库存策略，无需修改核心代码。

---

## 三、统一数据模型

### 3.1 库存配置表（inventory_config）

每个 SKU 一条配置，决定该商品使用哪种库存策略：

```sql
CREATE TABLE inventory_config (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  item_id         BIGINT NOT NULL,
  sku_id          BIGINT NOT NULL DEFAULT 0,
  
  -- 库存分类（核心）
  management_type INT NOT NULL COMMENT '1=自管理,2=供应商,3=无限',
  unit_type       INT NOT NULL COMMENT '1=券码,2=数量,3=时间,4=组合',
  deduct_timing   INT NOT NULL DEFAULT 1 COMMENT '1=下单,2=支付,3=发货',
  
  -- 供应商配置
  supplier_id     BIGINT NOT NULL DEFAULT 0,
  sync_strategy   INT NOT NULL DEFAULT 0 COMMENT '1=定时,2=实时,3=推送',
  sync_interval   INT NOT NULL DEFAULT 300 COMMENT '同步间隔(秒)',
  
  -- 风控配置
  oversell_allowed TINYINT NOT NULL DEFAULT 0,
  low_stock_threshold INT NOT NULL DEFAULT 100,
  
  UNIQUE KEY uk_item_sku (item_id, sku_id)
);
```

### 3.2 核心库存表（inventory）

所有品类共用一张库存表，通过不同字段组合适配不同场景：

```sql
CREATE TABLE inventory (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  item_id         BIGINT NOT NULL,
  sku_id          BIGINT NOT NULL,
  batch_id        BIGINT NOT NULL DEFAULT 0 COMMENT '批次(券码制)',
  calendar_date   DATE DEFAULT NULL COMMENT '日期(时间维度)',
  
  -- 核心库存字段
  total_stock     INT NOT NULL DEFAULT 0 COMMENT '总库存',
  available_stock INT NOT NULL DEFAULT 0 COMMENT '可售库存',
  booking_stock   INT NOT NULL DEFAULT 0 COMMENT '预订(已下单未支付)',
  locked_stock    INT NOT NULL DEFAULT 0 COMMENT '锁定(营销活动)',
  sold_stock      INT NOT NULL DEFAULT 0 COMMENT '已售',
  
  -- 供应商同步
  supplier_stock     INT NOT NULL DEFAULT 0,
  supplier_sync_time BIGINT NOT NULL DEFAULT 0,
  
  status INT NOT NULL DEFAULT 1 COMMENT '1=正常,2=缺货,3=停售',
  
  UNIQUE KEY uk_sku_batch_date (sku_id, batch_id, calendar_date)
);
```

**库存恒等式**：

```
total_stock = available_stock + booking_stock + locked_stock + sold_stock
```

**可售库存计算**（不同管理类型计算方式不同）：

```go
func CalcAvailable(inv *Inventory, cfg *Config) int32 {
    switch cfg.ManagementType {
    case SelfManaged:
        return inv.TotalStock - inv.SoldStock - inv.BookingStock - inv.LockedStock
    case SupplierManaged:
        return inv.SupplierStock - inv.BookingStock - inv.LockedStock
    case Unlimited:
        return 999999
    }
    return 0
}
```

### 3.3 券码池表（inventory_code_pool，分 100 张表）

仅用于券码制商品（Deal、Giftcard 预采购模式）：

```sql
CREATE TABLE inventory_code_pool_00 (
  id           BIGINT PRIMARY KEY COMMENT '雪花算法',
  item_id      BIGINT NOT NULL,
  sku_id       BIGINT NOT NULL,
  batch_id     BIGINT NOT NULL,
  
  code         VARCHAR(255) NOT NULL COMMENT '券码(唯一)',
  serial_number VARCHAR(255) DEFAULT '' COMMENT '序列号/PIN',
  code_url     VARCHAR(500) DEFAULT '' COMMENT '兑换链接',
  
  status       INT NOT NULL DEFAULT 1 COMMENT '1=可用,2=预订,3=已售,4=已核销,5=退款,6=过期',
  order_id     BIGINT NOT NULL DEFAULT 0,
  
  booking_time  BIGINT DEFAULT 0,
  purchase_time BIGINT DEFAULT 0,
  expire_time   BIGINT DEFAULT 0,
  
  UNIQUE KEY uk_code (code),
  KEY idx_status (status)
);
-- 分表规则：item_id % 100
```

### 3.4 库存操作日志表（inventory_operation_log）

所有库存变更留痕，用于对账和审计：

```sql
CREATE TABLE inventory_operation_log (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  item_id         BIGINT NOT NULL,
  sku_id          BIGINT NOT NULL,
  operation_type  VARCHAR(50) NOT NULL COMMENT 'book/unbook/sell/refund/lock/unlock',
  quantity        INT NOT NULL,
  order_id        BIGINT NOT NULL DEFAULT 0,
  before_available INT NOT NULL DEFAULT 0,
  after_available  INT NOT NULL DEFAULT 0,
  create_time     BIGINT NOT NULL DEFAULT 0,
  
  KEY idx_order_id (order_id),
  KEY idx_create_time (create_time)
);
```

---

## 四、策略模式：核心架构

### 4.1 整体架构

```
┌─────────────────────────────────────────────────┐
│  业务层 (Order Service / Promotion Service)     │
└──────────────────┬──────────────────────────────┘
                   ▼
┌─────────────────────────────────────────────────┐
│  统一库存管理器 (InventoryManager)               │
│  BookStock / UnbookStock / SellStock / Refund   │
└──────────────────┬──────────────────────────────┘
                   ▼
┌─────────────────────────────────────────────────┐
│  策略路由器 (StrategyRouter)                     │
│  根据 inventory_config 选择策略                  │
├────────┬────────┬────────┬──────────────────────┤
│ Self   │Supplier│Unlimit │Estimated             │
│Managed │Managed │Strategy│Strategy              │
│Strategy│Strategy│        │                      │
└────┬───┴────┬───┴────┬───┴──────────────────────┘
     ▼        ▼        ▼
┌─────────┐┌─────────┐┌──────────────┐
│  Redis  ││  MySQL  ││ Kafka Events │
│  (Hot)  ││  (Cold) ││   (Async)    │
└─────────┘└─────────┘└──────────────┘
```

### 4.2 策略接口定义

```go
// InventoryStrategy 库存管理策略接口
type InventoryStrategy interface {
    CheckStock(ctx context.Context, req *CheckStockReq) (*CheckStockResp, error)
    BookStock(ctx context.Context, req *BookStockReq) (*BookStockResp, error)
    UnbookStock(ctx context.Context, req *UnbookStockReq) error
    SellStock(ctx context.Context, req *SellStockReq) error
    RefundStock(ctx context.Context, req *RefundStockReq) error
}

// StrategyFactory 策略工厂
func GetStrategy(mgmtType int) InventoryStrategy {
    switch mgmtType {
    case SelfManaged:
        return &SelfManagedStrategy{}
    case SupplierManaged:
        return &SupplierManagedStrategy{}
    case Unlimited:
        return &UnlimitedStrategy{}
    default:
        return &UnlimitedStrategy{}
    }
}
```

---

## 五、自管理策略：券码制（Deal / Giftcard）

### 5.1 Redis 存储结构

```
Key:   inventory:code:pool:{itemID}:{skuID}:{batchID}
Type:  LIST
Value: [codeID_1, codeID_2, codeID_3, ...]
说明:  券码池，LPOP 出货，RPUSH 补货/退还

Key:   inventory:code:cursor:{itemID}:{skuID}:{batchID}
Type:  STRING
Value: "lastCodeID:lockCount"
说明:  补货游标，记录上次补到哪里

Key:   inventory:empty:{itemID}:{skuID}:{batchID}
Type:  STRING (TTL 1h)
Value: "1"
说明:  库存空标志，避免重复查库
```

### 5.2 核心流程：出货 + 补货

```
用户下单
  │
  ▼
1. 检查库存空标志 ──── 命中 → 返回缺货
  │ 未命中
  ▼
2. 从 Redis LIST 原子出货 (Lua: LRANGE + LTRIM)
  │
  ├── 出货成功 → 步骤 4
  │
  └── 库存不足 → 3. 补货 (从 MySQL 查可用券码 → RPUSH 到 Redis)
                       │
                       ├── 补货成功 → 再次出货 → 步骤 4
                       └── DB 也无库存 → 设置空标志(1h) → 返回缺货
  │
  ▼
4. 更新 MySQL 券码状态: AVAILABLE → BOOKING (绑定 order_id)
  │
  ▼
5. 同步更新 MySQL inventory 表: booking_stock += quantity
  │
  ▼
6. 发送 Kafka 事件 (异步)
```

**出货 Lua 脚本**（原子性保证）：

```lua
-- 原子取出 N 个券码
local result = redis.call('LRANGE', KEYS[1], 0, ARGV[1] - 1)
redis.call('LTRIM', KEYS[1], ARGV[1], -1)
return result
```

**补货流程**（加分布式锁防并发）：

```go
func (s *SelfManagedStrategy) replenish(ctx context.Context, itemID, skuID, batchID uint64) error {
    // 1. 获取分布式锁（10s 超时）
    lockKey := fmt.Sprintf("inventory:lock:replenish:%d:%d:%d", itemID, skuID, batchID)
    if !acquireLock(lockKey, 10*time.Second) {
        return nil // 其他进程正在补货，等待即可
    }
    defer releaseLock(lockKey)

    // 2. 读取补货游标（上次补到哪个 codeID）
    lastCodeID := getCursor(itemID, skuID, batchID)

    // 3. 从 MySQL 查 3000 个可用券码
    codes, err := db.Query(`
        SELECT id FROM inventory_code_pool_xx
        WHERE item_id=? AND sku_id=? AND batch_id=? AND status=1 AND id > ?
        ORDER BY id LIMIT 3000
    `, itemID, skuID, batchID, lastCodeID)

    if len(codes) == 0 {
        // DB 也无库存，设置空标志
        redis.Set(emptyKey, "1", 1*time.Hour)
        return ErrStockNotEnough
    }

    // 4. 原子写入 Redis LIST + 更新游标
    redis.Eval(replenishScript, stockKey, cursorKey, codeIDs, newCursor)
    return nil
}
```

---

## 六、自管理策略：数量制（OPV / 本地服务）

### 6.1 Redis 存储结构

```
Key:   inventory:qty:stock:{itemID}:{skuID}
Type:  HASH
Fields:
  "available"   : 10000       # 可售库存
  "booking"     : 50          # Shopee 预订中
  "issued"      : 5000        # 已售
  "locked"      : 500         # 营销锁定
  "{promotionID}": 200        # 营销活动独立库存（动态字段）
```

### 6.2 预订 Lua 脚本

```lua
local key = KEYS[1]
local book_num = tonumber(ARGV[1])
local promotion_id = ARGV[2]  -- 空字符串表示普通库存

-- 1. 获取可用库存
local available = tonumber(redis.call('HGET', key, 'available') or 0)

-- 2. 如果有营销活动，合并计算
local promo_stock = 0
if promotion_id ~= '' then
    promo_stock = tonumber(redis.call('HGET', key, promotion_id) or 0)
end
local total_available = available + promo_stock

-- 3. 检查库存
if book_num > total_available then
    return -1  -- 库存不足
end

-- 4. 优先扣营销库存，不足时扣普通库存
if promo_stock > 0 then
    if book_num <= promo_stock then
        redis.call('HINCRBY', key, promotion_id, -book_num)
    else
        redis.call('HSET', key, promotion_id, 0)
        redis.call('HINCRBY', key, 'available', -(book_num - promo_stock))
    end
else
    redis.call('HINCRBY', key, 'available', -book_num)
end

-- 5. 增加预订数
redis.call('HINCRBY', key, 'booking', book_num)

return total_available - book_num
```

### 6.3 支付成功 / 取消订单

```lua
-- 支付成功：booking → issued
local booking = tonumber(redis.call('HGET', key, 'booking') or 0)
if stock > booking then return -1 end  -- 异常保护
redis.call('HINCRBY', key, 'booking', -stock)
redis.call('HINCRBY', key, 'issued', stock)

-- 取消订单：booking → available
redis.call('HINCRBY', key, 'booking', -stock)
redis.call('HINCRBY', key, 'available', stock)
```

---

## 七、供应商管理策略（酒店 / 机票）

### 7.1 同步策略

| 策略 | 适用场景 | 实时性 | 实现方式 |
|------|----------|--------|----------|
| **实时查询** | 库存变化快（机票） | 高 | 每次请求调供应商 API（30s 缓存） |
| **定时同步** | 库存变化中等（酒店） | 中 | 定时任务每 5 分钟拉取 |
| **Webhook** | 供应商主动推送 | 高 | 接收推送更新本地缓存 |

### 7.2 实时查询流程

```go
func (s *SupplierManagedStrategy) CheckStock(ctx context.Context, req *CheckStockReq) (*CheckStockResp, error) {
    // 1. 查 Redis 缓存（30s TTL）
    cacheKey := fmt.Sprintf("inventory:supplier:%d:%d:%s", req.ItemID, req.SKUID, req.Date)
    if stock, err := redis.Get(cacheKey).Int(); err == nil {
        return &CheckStockResp{Available: stock, FromCache: true}, nil
    }

    // 2. 缓存未命中，调供应商 API
    resp, err := supplierClient.QueryStock(ctx, req.SupplierID, req.ProductID, req.Date)
    if err != nil {
        return nil, err
    }

    // 3. 写入 Redis 缓存（30s）+ 异步写快照表
    redis.Set(cacheKey, resp.Stock, 30*time.Second)
    go saveSnapshot(req.ItemID, resp.Stock, "api")

    return &CheckStockResp{Available: resp.Stock, FromCache: false}, nil
}
```

### 7.3 预订流程（供应商管理）

#### 7.3.1 同步预订（理想情况）

供应商 API 质量好，预订接口同步返回结果：

```go
func (s *SupplierManagedStrategy) BookStock(ctx context.Context, req *BookStockReq) (*BookStockResp, error) {
    // 1. 调供应商预订接口（同步返回成功/失败）
    resp, err := supplierClient.Book(ctx, req.SupplierID, req.ProductID, req.OrderID)
    if err != nil {
        return nil, err
    }

    // 2. 保存供应商订单号映射
    saveOrderMapping(req.OrderID, resp.SupplierOrderID)

    // 3. 更新本地库存表（记录 booking）
    updateInventoryBooking(req.ItemID, req.SKUID, req.Quantity, +1)

    // 4. 发送事件
    publishEvent(&InventoryEvent{Type: "book", OrderID: req.OrderID, SupplierOrderID: resp.SupplierOrderID})

    return &BookStockResp{Success: true, SupplierOrderID: resp.SupplierOrderID}, nil
}
```

---

#### 7.3.2 异步预订（供应商系统较差）

**场景**：部分供应商系统不稳定，预订流程为：
1. 创建 booking 单 → 立即返回 `booking_id`（状态 `PENDING`）
2. 轮询查询 booking 状态 → 最终返回 `CONFIRMED` / `FAILED`
3. 只有 `CONFIRMED` 后才能继续下单

**挑战**：
- 用户不能等待轮询完成（可能需要 10-30 秒）。
- 需要异步处理 + 状态机 + 补偿机制。

**状态机设计**：

```
用户下单
  ↓
BOOKING_INIT (初始化)
  ↓
调供应商创建 booking → 返回 booking_id
  ↓
BOOKING_PENDING (等待确认)
  ↓
异步轮询 booking 状态（每 2s 查询一次，最多 30s）
  ↓
  ├─ CONFIRMED → BOOKING_SUCCESS
  ├─ FAILED    → BOOKING_FAILED (释放库存)
  └─ TIMEOUT   → BOOKING_TIMEOUT (人工介入)
```

**数据库表设计**：

```sql
CREATE TABLE supplier_booking (
  id                BIGINT PRIMARY KEY AUTO_INCREMENT,
  order_id          BIGINT NOT NULL COMMENT '平台订单ID',
  item_id           BIGINT NOT NULL,
  supplier_id       BIGINT NOT NULL,
  
  booking_id        VARCHAR(100) NOT NULL COMMENT '供应商 booking ID',
  booking_status    VARCHAR(50) NOT NULL COMMENT 'PENDING/CONFIRMED/FAILED/TIMEOUT',
  
  create_time       BIGINT NOT NULL,
  confirm_time      BIGINT DEFAULT 0,
  query_count       INT DEFAULT 0 COMMENT '轮询次数',
  last_query_time   BIGINT DEFAULT 0,
  
  error_msg         TEXT,
  
  KEY idx_order_id (order_id),
  KEY idx_booking_id (booking_id),
  KEY idx_status_time (booking_status, create_time)
);
```

**实现流程**：

```go
// 1. 用户下单时：创建 booking 单，立即返回"处理中"
func (s *SupplierManagedStrategy) BookStock(ctx context.Context, req *BookStockReq) (*BookStockResp, error) {
    // 调供应商创建 booking
    resp, err := supplierClient.CreateBooking(ctx, req.SupplierID, req.ProductID)
    if err != nil {
        return nil, err
    }

    // 保存 booking 记录（状态 PENDING）
    saveSupplierBooking(&SupplierBooking{
        OrderID:       req.OrderID,
        ItemID:        req.ItemID,
        SupplierID:    req.SupplierID,
        BookingID:     resp.BookingID,
        BookingStatus: "PENDING",
        CreateTime:    time.Now().Unix(),
    })

    // 发送到 MQ 异步轮询
    publishToMQ(&BookingPollTask{
        OrderID:    req.OrderID,
        BookingID:  resp.BookingID,
        SupplierID: req.SupplierID,
    })

    // 立即返回给用户（告知"预订处理中"）
    return &BookStockResp{
        Success:       false, // 尚未确认
        Status:        "PROCESSING",
        BookingID:     resp.BookingID,
        EstimateTime:  30, // 预计 30 秒内确认
    }, nil
}

// 2. 异步 Consumer：轮询 booking 状态
func PollBookingStatus(task *BookingPollTask) {
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()
    
    timeout := time.After(30 * time.Second)
    queryCount := 0

    for {
        select {
        case <-ticker.C:
            queryCount++
            
            // 调供应商查询接口
            status, err := supplierClient.QueryBookingStatus(task.SupplierID, task.BookingID)
            
            updateQueryRecord(task.OrderID, queryCount, time.Now().Unix())
            
            if err != nil {
                log.Error("query booking failed", err)
                continue
            }

            switch status {
            case "CONFIRMED":
                // 预订成功
                handleBookingSuccess(task.OrderID, task.BookingID)
                return
                
            case "FAILED":
                // 预订失败
                handleBookingFailed(task.OrderID, task.BookingID, "supplier rejected")
                return
                
            case "PENDING":
                // 继续等待
                continue
            }

        case <-timeout:
            // 超时未确认
            handleBookingTimeout(task.OrderID, task.BookingID)
            return
        }
    }
}

// 3. 预订成功回调
func handleBookingSuccess(orderID uint64, bookingID string) {
    // 更新状态
    updateSupplierBooking(orderID, "CONFIRMED", time.Now().Unix())
    
    // 更新本地库存
    updateInventoryBooking(orderID, +1)
    
    // 通知用户（Push / SMS / Email）
    notifyUser(orderID, "您的订单预订成功，请尽快支付")
    
    // 设置支付超时（15 分钟）
    setPaymentTimeout(orderID, 15*time.Minute)
}

// 4. 预订失败回调
func handleBookingFailed(orderID uint64, bookingID string, reason string) {
    updateSupplierBooking(orderID, "FAILED", time.Now().Unix())
    
    // 释放本地库存（如果有预扣）
    releaseInventoryBooking(orderID)
    
    // 关闭订单
    closeOrder(orderID, "supplier booking failed: " + reason)
    
    // 通知用户
    notifyUser(orderID, "抱歉，预订失败，请重新下单")
}

// 5. 预订超时回调
func handleBookingTimeout(orderID uint64, bookingID string) {
    updateSupplierBooking(orderID, "TIMEOUT", time.Now().Unix())
    
    // 记录异常，人工介入
    alert("Booking timeout: order=%d, booking=%s", orderID, bookingID)
    
    // 继续在后台轮询（降低频率：每 1 分钟查询一次，最多 24 小时）
    scheduleBackgroundPoll(orderID, bookingID)
    
    // 暂不关闭订单，等待人工处理
}
```

---

#### 7.3.3 用户体验优化

**问题**：用户下单后看到"预订处理中"，体验不佳。

**优化方案**：

1. **前端轮询展示进度**：

```javascript
// 用户下单后，前端每 2 秒轮询订单状态
function pollOrderStatus(orderId) {
    const interval = setInterval(async () => {
        const resp = await fetch(`/api/order/${orderId}/status`);
        const data = await resp.json();
        
        if (data.status === 'BOOKING_SUCCESS') {
            showMessage('预订成功！请在 15 分钟内完成支付');
            clearInterval(interval);
            redirectToPayment(orderId);
        } else if (data.status === 'BOOKING_FAILED') {
            showMessage('预订失败，库存不足');
            clearInterval(interval);
        } else {
            // 继续等待
            updateProgress(data.queryCount, 15); // 进度条：已查询 X/15 次
        }
    }, 2000);
    
    // 30 秒后停止轮询
    setTimeout(() => clearInterval(interval), 30000);
}
```

2. **WebSocket / SSE 推送**：

```go
// 服务端：booking 确认后推送消息
func handleBookingSuccess(orderID uint64) {
    // ... 更新状态 ...
    
    // 推送给前端
    websocketHub.Push(orderID, &Message{
        Type: "BOOKING_CONFIRMED",
        Data: map[string]interface{}{
            "order_id": orderID,
            "status":   "SUCCESS",
        },
    })
}
```

3. **短信/Push 通知**：

```go
// 预订成功后 1 分钟内发送通知
notifyUser(orderID, "您的【泰国普吉岛酒店】预订成功，请尽快支付")
```

---

#### 7.3.4 异常场景处理

**场景 1：轮询期间用户取消订单**

```go
func PollBookingStatus(task *BookingPollTask) {
    for {
        // 每次轮询前检查订单状态
        order := getOrder(task.OrderID)
        if order.Status == "CANCELLED" {
            // 调供应商取消接口
            supplierClient.CancelBooking(task.SupplierID, task.BookingID)
            return
        }
        
        // ... 继续轮询 ...
    }
}
```

**场景 2：供应商 API 持续超时**

```go
// 连续 3 次查询超时 → 降级到人工处理
if queryCount >= 3 && allTimeout {
    handleBookingTimeout(task.OrderID, task.BookingID)
    
    // 发送企业微信/钉钉告警
    alertOps("供应商 API 异常: supplier_id=%d, booking_id=%s", 
             task.SupplierID, task.BookingID)
}
```

**场景 3：供应商确认后用户未支付**

```go
// booking 成功后设置 15 分钟支付超时
func handleBookingSuccess(orderID uint64, bookingID string) {
    // ... 更新状态 ...
    
    // 15 分钟后自动取消
    scheduleTask(&CancelBookingTask{
        OrderID:   orderID,
        BookingID: bookingID,
        Delay:     15 * time.Minute,
    })
}

// 超时后调供应商取消接口
func CancelExpiredBooking(task *CancelBookingTask) {
    order := getOrder(task.OrderID)
    if order.Status != "PAID" {
        // 调供应商取消
        supplierClient.CancelBooking(order.SupplierID, task.BookingID)
        
        // 关闭订单
        closeOrder(task.OrderID, "payment timeout")
    }
}
```

---

#### 7.3.5 监控指标

| 指标 | 阈值 | 说明 |
|------|------|------|
| **booking 成功率** | > 95% | 供应商库存准确性 |
| **平均确认时长** | < 10s | P99 < 30s |
| **超时率** | < 1% | 需要人工介入的比例 |
| **取消率** | < 5% | 用户等待期间取消订单 |

```go
// Prometheus Metrics
bookingConfirmDuration := prometheus.NewHistogram(...)
bookingSuccessRate := prometheus.NewCounter(...)
bookingTimeoutCount := prometheus.NewCounter(...)
```

---

## 八、无限库存策略（TopUp / 保险）

最简单的策略，只记录操作日志：

```go
type UnlimitedStrategy struct{}

func (s *UnlimitedStrategy) CheckStock(ctx context.Context, req *CheckStockReq) (*CheckStockResp, error) {
    return &CheckStockResp{Available: 999999, IsUnlimited: true}, nil
}

func (s *UnlimitedStrategy) BookStock(ctx context.Context, req *BookStockReq) (*BookStockResp, error) {
    // 仅记录日志（用于统计销量）
    logOperation("book", req.ItemID, req.Quantity, req.OrderID)
    return &BookStockResp{Success: true}, nil
}

func (s *UnlimitedStrategy) UnbookStock(ctx context.Context, req *UnbookStockReq) error {
    return nil // 无操作
}
```

---

## 九、核心流程汇总

### 9.1 统一预订流程

```
用户下单
  │
  ▼
1. 查 inventory_config → 获取 management_type + unit_type
  │
  ▼
2. StrategyFactory.GetStrategy(management_type)
  │
  ├─ Self + Code     → 券码出货 (Redis LIST LPOP)
  ├─ Self + Quantity  → 数量扣减 (Redis HASH Lua 原子)
  ├─ Supplier + Time  → 调供应商预订 API
  ├─ Unlimited        → 直接成功
  └─ Self + Bundle    → 遍历子项，逐一扣减
  │
  ▼
3. 更新 inventory 表: booking_stock += quantity
  │
  ▼
4. 发送 Kafka 事件 → 异步消费写操作日志
  │
  ▼
5. 返回结果（券码制返回 codeIDs，供应商返回 supplierOrderID）
```

### 9.2 支付成功流程

```
支付回调
  │
  ▼
1. 路由到对应策略
  │
  ├─ 券码制: code status BOOKING → SOLD, 设置 purchase_time/expire_time
  ├─ 数量制: Redis booking--, sold++
  ├─ 供应商: 调供应商确认接口（可选）
  └─ Giftcard(实时生成): 调供应商 API 生成卡密 → 保存到 code_pool
  │
  ▼
2. 更新 inventory: booking_stock -= qty, sold_stock += qty
  │
  ▼
3. 发送事件
```

### 9.3 取消/超时释放流程

```
订单取消 / 超时未支付
  │
  ▼
1. 路由到对应策略
  │
  ├─ 券码制: code status BOOKING → AVAILABLE, RPUSH 回 Redis LIST
  ├─ 数量制: Redis booking--, available++
  └─ 供应商: 调供应商取消接口
  │
  ▼
2. 更新 inventory: booking_stock -= qty, available_stock += qty
```

---

## 十、数据一致性保障

### 10.1 Redis 与 MySQL 双写策略

| 操作 | Redis | MySQL | 一致性保障 |
|------|-------|-------|-----------|
| **预订 (Book)** | 同步扣减（Lua 原子） | Kafka 异步更新 | 最终一致 |
| **支付 (Sell)** | 同步更新 | Kafka 异步更新 | 最终一致 |
| **营销锁定 (Lock)** | 同步 | 同步（DB 事务） | 强一致 |
| **补货 (Replenish)** | 同步写入 | 不变 | - |

**核心原则**：
- **Redis 是热路径**：所有高频读写走 Redis，保证毫秒级响应。
- **MySQL 是权威数据源**：故障恢复以 MySQL 为准。
- **Kafka 异步持久化**：Book/Sell 等操作通过 MQ 异步落库，不阻塞主流程。

### 10.2 定时对账（每小时）

```go
func Reconcile() {
    configs := queryAllSelfManagedConfigs()
    for _, cfg := range configs {
        redisStock := getRedisAvailable(cfg.ItemID, cfg.SKUID)
        mysqlStock := getMySQLAvailable(cfg.ItemID, cfg.SKUID)
        diff := redisStock - mysqlStock

        // 恒等式校验：total = available + booking + locked + sold
        mysqlTotal := getMySQLTotal(cfg.ItemID, cfg.SKUID)
        mysqlCalc := mysqlStock + mysqlBooking + mysqlLocked + mysqlSold
        if mysqlCalc != mysqlTotal {
            alert("MySQL 数据不一致: item=%d, total=%d, calc=%d", cfg.ItemID, mysqlTotal, mysqlCalc)
        }

        // Redis vs MySQL 差异
        if abs(diff) > 100 || abs(diff) > mysqlStock/10 {
            alert("库存差异过大: item=%d, redis=%d, mysql=%d", cfg.ItemID, redisStock, mysqlStock)
        }

        // 自动修复（可选，以 MySQL 为准）
        if cfg.AutoReconcile {
            syncRedisFromMySQL(cfg.ItemID, cfg.SKUID)
        }
    }
}
```

### 10.3 降级方案

```
Redis 可用 → 正常读写 Redis
     │
Redis 不可用
     │
     ▼
降级到 MySQL 直接操作（性能下降但业务不中断）
     │
     ├─ 券码制: SELECT ... FOR UPDATE + 状态更新
     ├─ 数量制: UPDATE available_stock = available_stock - ? WHERE available_stock >= ?
     └─ 记录降级日志，Redis 恢复后全量同步
```

---

## 十一、Kafka 事件设计

```protobuf
message InventoryEvent {
    string event_id    = 1;  // UUID
    string event_type  = 2;  // book/unbook/sell/refund/lock/sync
    int64  timestamp   = 3;

    int64  item_id     = 10;
    int64  sku_id      = 11;
    int64  batch_id    = 12;
    string calendar_date = 13; // 时间维度库存

    int32  quantity    = 20;
    repeated int64 code_ids = 21; // 券码制

    int64  order_id    = 30;
    string supplier_order_id = 31;

    int32  before_available = 40; // 操作前快照
    int32  after_available  = 41; // 操作后快照
}
```

**Topic 设计**：
- `inventory.book` — 预订
- `inventory.unbook` — 释放
- `inventory.sell` — 售出
- `inventory.refund` — 退款
- `inventory.sync` — 供应商同步

---

## 十二、Giftcard 特殊设计

Giftcard 横跨三种库存模式，是统一模型的最佳验证：

| 模式 | 管理类型 | 流程 | 适用场景 |
|------|----------|------|----------|
| **预采购卡密** | Self + Code | 批量导入 → Redis 出货 | 高频热销卡 |
| **实时生成** | Supplier + Code | 支付成功 → 调 API 生成 → 存入 code_pool | 长尾低频卡 |
| **无限库存** | Unlimited | 直接成功 | 供应商保证库存 |

**卡密安全**：
- 存储时 AES-256 加密卡号和 PIN。
- 管理后台脱敏显示（`XXXX-XXXX-XXXX-1234`）。
- 所有访问记录审计日志。

**供应商 API 超时处理**：
- 支付成功后异步生成，完成后推送通知用户。
- 指数退避重试（1s, 2s, 4s），3 次失败后人工补发。

---

## 十三、监控与告警

### 13.1 关键指标

| 指标 | 阈值 | 告警级别 |
|------|------|----------|
| **超卖次数** | > 0 | P0 |
| **Redis vs MySQL 差异** | > 100 | P1 |
| **库存服务错误率** | > 1% | P1 |
| **库存扣减 P99** | > 200ms | P2 |
| **补货失败率** | > 5% | P2 |
| **供应商同步延迟** | > 10min | P2 |
| **低库存商品数** | > 100 | P3 |

### 13.2 Prometheus Metrics

```
# 操作计数
inventory_operation_total{op="book|sell|refund", mgmt="self|supplier", status="ok|fail"}

# 操作延迟
inventory_operation_duration_seconds{op="book|sell"}

# 库存差异
inventory_reconcile_diff{item_id, sku_id}

# 缺货次数
inventory_out_of_stock_total{item_id}
```

---

## 十四、新品类接入指南

**三步接入**：

1. **评估分类**：确定 `(ManagementType, UnitType, DeductTiming)`。
2. **写配置**：在 `inventory_config` 表插入一条记录。
3. **调接口**：使用统一 `InventoryManager.BookStock()` 即可。

```go
// 示例：接入新品类"演唱会门票"
// 1. 评估：供应商管理 + 时间维度 + 支付成功扣减
// 2. 写配置
INSERT INTO inventory_config (item_id, management_type, unit_type, deduct_timing, supplier_id, sync_strategy)
VALUES (900001, 2, 3, 2, 700001, 2);

// 3. 调用统一接口
inventoryManager.BookStock(ctx, &BookStockReq{
    ItemID:   900001,
    SKUID:    0,
    Quantity: 2,
    OrderID:  orderID,
    CalendarDate: "2025-08-15",
})
```

---

## 十五、生产环境实战数据

### 15.1 业务规模

| 指标 | 数值 | 说明 |
|------|------|------|
| **秒杀峰值 QPS** | 20,000 | 单个爆款商品，持续 5-10 分钟 |
| **日均 QPS** | 50 | 常态流量 |
| **日均订单量** | 2,000,000 | 支付成功订单 |
| **日均库存扣减** | 6,700,000 | 含预订、支付、取消等操作 |
| **峰值/日均比** | 870:1 | 流量极度不均匀 |

**容量规划推算**：

```
日均订单 2M / 86400s ≈ 23 TPS
秒杀峰值 20k QPS = 日均的 870 倍

假设订单转化率 30%（下单 → 支付成功）
日均扣减请求 = 2M / 0.3 ≈ 6.7M 次
Kafka 异步落库 MySQL TPS = 6.7M / 86400 ≈ 80 TPS（日均）
秒杀峰值 MySQL TPS ≈ 300-500 TPS（批量写入优化后）
```

---

### 15.2 集群配置

#### Redis 集群

```
拓扑: Redis Cluster (3 主 3 从)
分片: 按 item_id Hash 分片
单分片配置: 32GB 内存, 16 核
持久化: AOF + RDB 混合模式
```

**容量规划**：
- 券码池：100 万张券码 × 8 字节 ≈ **8 MB**（单商品）
- 热点商品预热：10 个商品 × 8MB = **80 MB**
- 数量制商品：1 万个 SKU × 1 KB ≈ **10 MB**
- 总计：**< 200 MB**（核心数据），32GB 绰绰有余

#### 应用服务

```
实例数: 10 台 (Kubernetes Pod)
单实例配置: 4 核 8GB
线程池: 每实例 500 线程（IO 密集型，2N 配置）
单实例承载: 2,000 QPS（秒杀峰值）
```

**为什么 10 台能抗 2w QPS？**
- Redis 操作 RT < 5ms，单线程 QPS = 1000/5 = 200
- 500 线程 × 200 QPS = 100k QPS 理论上限（实际 2k QPS，留足余量）

#### MySQL 集群

```
架构: 1 主 2 从（半同步复制）
主库配置: 16 核 64GB, SSD
从库: 读流量（对账、报表）
分表: inventory_code_pool 分 100 张表
```

**容量规划**：
- 券码池：1 亿张券码 × 500 字节 ≈ **50 GB**（分 100 张表，单表 500 MB）
- inventory 表：10 万条记录 × 1 KB ≈ **100 MB**
- operation_log：日增 670 万条 × 200 字节 ≈ **1.3 GB/天**（保留 30 天 ≈ 40 GB）

#### Kafka 集群

```
Broker: 3 台
Topic: inventory.events (6 分区)
消费者组: 6 个 Consumer（并发消费）
```

**吞吐量验证**：
- 秒杀峰值写入 Kafka: 20k TPS × 500 字节 = **10 MB/s**
- Kafka 单分区吞吐 > 50 MB/s，6 分区 = **300 MB/s** 理论上限
- 实际使用 **< 5%**，非常充裕

---

### 15.3 性能指标实测

| 操作 | P50 | P99 | P999 | 备注 |
|------|-----|-----|------|------|
| **券码制预订** | 15ms | 50ms | 150ms | 含 Redis + MySQL 同步更新 |
| **数量制预订** | 8ms | 30ms | 100ms | 仅 Redis Lua 脚本 |
| **供应商库存查询** | 200ms | 500ms | 2s | 第三方 API，30s 缓存 |
| **Redis 单次操作** | 1ms | 5ms | 10ms | LIST/HASH 操作 |
| **MySQL 券码状态更新** | 10ms | 50ms | 200ms | 主库写入 |
| **Kafka 异步消费延迟** | 50ms | 200ms | 1s | 非秒杀场景 |

**秒杀场景优化后**：
- 券码**提前预热**到 Redis（活动前 1 小时）
- P99 降至 **30ms**（无 DB 补货开销）

---

### 15.4 真实案例与优化

#### 案例 1：秒杀 2w QPS 热点 Key 瓶颈

**问题**：
- 单个爆款商品，所有请求打到同一个 Redis Key。
- Redis 单线程模型，QPS 上限 **10 万**（理论值），但 **网卡带宽** 先打满。
- 实测单 Key 极限 **5 万 QPS**（1KB 数据 × 5w = 50 MB/s，接近千兆网卡上限）。

**解决方案**：
1. **本地缓存 (Caffeine)**：
   - 应用层缓存库存数（非强一致，允许轻微超卖）。
   - 本地缓存拦截 80% 读请求，Redis 只承担 **4k QPS**。

2. **Key 分散**（适用于读多写少）：
   - 将热点 Key 复制 10 份：`stock:item_123:0 ~ stock:item_123:9`。
   - 读请求随机路由，写请求同步更新所有副本。

3. **限流前置**：
   - 网关层按 `item_id` 限流，单商品最大 **2.5w QPS**（留 20% 余量）。
   - 超出部分直接返回"繁忙"，避免击穿 Redis。

---

#### 案例 2：券码补货锁超时

**问题**：
- 补货时加分布式锁（10s 超时），从 MySQL 查 3000 张券码。
- DB 慢查询导致补货耗时 **12s**，锁提前过期。
- 另一个进程拿到锁，重复补货，导致 **券码重复出货**。

**根因**：
- MySQL `inventory_code_pool_xx` 表数据量大（千万级），`status=1` 索引选择性差。
- 执行计划走了全表扫描。

**解决方案**：
1. **优化 SQL**：
   ```sql
   -- 增加复合索引
   KEY idx_item_status_id (item_id, status, id)
   
   -- 查询改为游标分页
   SELECT id FROM inventory_code_pool_xx
   WHERE item_id=? AND status=1 AND id > ?
   ORDER BY id LIMIT 3000
   ```
   耗时从 12s 降至 **50ms**。

2. **锁续期**：
   - 补货时启动守护线程，每 5s 检查锁是否需要续期。
   - 避免长事务导致锁过期。

3. **异步补货**：
   - 检测库存低于阈值（1000 张）时，**提前异步补货**。
   - 避免用户请求阻塞在补货逻辑。

---

#### 案例 3：Kafka 消费积压

**问题**：
- 秒杀活动结束后，Kafka 积压 **50 万条消息**（2.5 万 QPS × 20s）。
- 6 个 Consumer 消费速度跟不上，MySQL 写入成为瓶颈。

**瓶颈分析**：
- Consumer 逐条更新 MySQL：`UPDATE inventory SET booking_stock = booking_stock + 1`
- MySQL 单线程提交，TPS **< 5000**（主从半同步复制延迟）。

**解决方案**：
1. **批量写入**：
   ```go
   // 攒批 100 条，批量 INSERT
   INSERT INTO inventory_operation_log (item_id, operation_type, quantity, ...)
   VALUES (?, ?, ?), (?, ?, ?), ...  -- 100 rows
   ```
   TPS 从 5k 提升至 **8 万**（提升 16 倍）。

2. **降低一致性要求**：
   - `inventory_operation_log` 日志表改为**异步从库写入**。
   - 主库只更新 `inventory` 核心表。

3. **削峰**：
   - Kafka 设置 `linger.ms=100ms`，Producer 端攒批发送。
   - 减少消息数量。

---

#### 案例 4：对账发现的典型问题

**统计数据**（3 个月）：
- 对账次数：**2160 次**（每小时 1 次）
- 发现差异：**87 次**（4% 频率）
- 差异 > 100：**3 次**（严重）

**主要根因**：

| 原因 | 占比 | 说明 |
|------|------|------|
| Kafka 消费延迟 | 60% | 秒杀后消费积压，MySQL 未及时更新 |
| Redis 补货未同步 MySQL | 25% | 券码补货只更新 Redis，DB 未记录 |
| 人工后台操作 | 10% | 运营手动修改 DB 库存 |
| Redis 重启丢数据 | 5% | AOF 未及时刷盘（`appendfsync everysec`）|

**优化措施**：
1. **Kafka 消费延迟告警**：lag > 1000 立即告警。
2. **Redis 补货同步**：补货时同步更新 MySQL `total_stock`。
3. **后台操作审计**：所有库存修改必须通过 API，禁止直接改 DB。
4. **Redis 持久化增强**：改为 `appendfsync always`（性能下降 30%，换取强一致）。

---

### 15.5 成本分析

| 资源 | 配置 | 数量 | 月成本（美元） |
|------|------|------|--------------|
| Redis Cluster | 32GB × 6 节点 | 1 套 | $800 |
| MySQL | 64GB 主库 + 32GB × 2 从库 | 1 套 | $1,200 |
| 应用服务 | 4C8G Pod | 10 台 | $600 |
| Kafka | 8C16G Broker | 3 台 | $900 |
| **总计** | - | - | **$3,500/月** |

**日均订单成本**：$3,500 / 2,000,000 = **$0.00175/单**（0.175 美分）

---

### 15.6 核心设计决策

| 决策 | 选择 | 原因 |
|------|------|------|
| **统一 vs 独立** | 统一模型 + 策略模式 | 复用逻辑，新品类零代码接入 |
| **Redis vs MySQL** | Redis 优先，MySQL 持久化 | 高并发性能 + 数据可靠 |
| **同步 vs 异步** | 扣减同步，落库异步 | 热路径极速，冷路径可靠 |
| **券码出货方式** | Lazy Loading（按需补货） | 节省内存，避免一次性加载全量 |
| **对账策略** | 每小时自动对账，MySQL 为准 | 兜底一致性 |
| **降级策略** | Redis 宕机切 MySQL | 性能下降 10 倍，但业务不中断 |

---

### 15.7 业界对比

| 维度 | 淘宝/京东 | Amazon | 本设计 |
|------|-----------|--------|--------|
| 库存单元 | SKU 数量 | ASIN + FBA | SKU + 批次/日期 |
| 扣减时机 | 下单预订 | 支付成功 | **可配置** |
| 虚拟商品 | 部分支持 | 完善 | **核心场景** |
| 时间维度 | 不支持 | 不支持 | **支持** |
| 券码管理 | 部分 | 完善 | **核心能力** |
| 供应商集成 | 少量 | FBA 模式 | **多策略** |
| 峰值 QPS | 100 万+ | 50 万+ | **2 万**（中型平台）|
