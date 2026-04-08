---
title: 电商系统设计：营销系统深度解析
date: 2026-04-07
categories:
  - system-design
tags:
  - system-design
  - ecommerce
  - marketing
  - coupon
  - points
  - promotion
  - high-concurrency
  - distributed-system
---

# 电商系统设计：营销系统深度解析

营销系统是电商平台的增长引擎，通过优惠券、积分、活动等手段实现用户拉新、促活、留存和 GMV 提升。本文深入解析营销系统的架构设计、核心模块、高并发场景处理和工程实践，适合系统设计面试和电商后端工程师阅读。

<!-- more -->

## 目录

1. [系统概览](#1-系统概览)
2. [营销工具体系](#2-营销工具体系)
3. [营销计算引擎](#3-营销计算引擎)
4. [高并发场景设计](#4-高并发场景设计)
5. [营销与订单集成](#5-营销与订单集成)
6. [跨系统全链路集成](#6-跨系统全链路集成)
7. [数据一致性保障](#7-数据一致性保障)
8. [特殊营销场景](#8-特殊营销场景)
9. [工程实践](#9-工程实践)
10. [总结与参考](#10-总结与参考)

---

## 1. 系统概览

### 1.1 营销系统的定位

营销系统在电商平台中承担**增长引擎**角色：通过优惠券、积分、活动与精准投放，支撑用户拉新、促活、留存与 GMV 提升。它与订单、商品、计价、用户、支付等系统紧密协作：订单侧负责扣减与回退的编排，商品与计价侧提供圈品与价格试算输入，用户侧提供画像与风控维度，支付侧完成补贴与分账核算。

**核心价值**可概括为：在**成本可控**前提下实现**精准营销**，并通过数据闭环让效果**可衡量、可优化**。

### 1.2 核心业务场景

典型业务场景包括：

- **用户拉新**：新人专享券、首单立减
- **用户促活**：签到积分、任务奖励
- **用户留存**：会员积分、等级权益
- **GMV 提升**：满减活动、限时折扣、秒杀
- **清库存**：N 元购、买赠活动

**B2C 与 B2B2C 营销差异**对比如下。

| 维度 | B2C（自营） | B2B2C（平台） |
|------|------------|--------------|
| 营销主体 | 平台 | 平台 + 商家 |
| 成本承担 | 平台全额 | 平台补贴 + 商家承担 |
| 活动审核 | 无需审核 | 商家活动需平台审核 |
| 优惠叠加 | 平台规则统一 | 需考虑跨店铺规则 |
| 结算复杂度 | 简单 | 需分账（平台 / 商家） |

### 1.3 核心挑战

1. **高并发**：秒杀、抢券等场景 QPS 峰值可达 10 万+
2. **复杂规则**：优惠叠加、互斥、优先级与「最优解」求解
3. **数据一致性**：营销扣减与订单创建的原子性与补偿
4. **防刷防薅**：黑产、批量注册、恶意套现
5. **成本控制**：营销预算、ROI 监控
6. **实时性**：库存实时扣减、优惠实时生效

### 1.4 系统架构

整体采用**接入层 → 服务层 → 数据层**分层：接入层统一鉴权与路由；服务层拆分优惠券、积分、活动与营销计算；数据层以 MySQL 为主存储，Redis 承担缓存与分布式锁，Kafka 做事件驱动，Elasticsearch 支撑活动检索与画像类查询。

**系统架构总览**如下。

```mermaid
graph LR
    User[用户] --> WebApp[Web/App]
    WebApp --> MarketingGateway[营销网关]

    MarketingGateway --> CouponService[优惠券服务]
    MarketingGateway --> PointsService[积分服务]
    MarketingGateway --> ActivityService[活动引擎]
    MarketingGateway --> CalculationEngine[营销计算引擎]

    CouponService --> CouponDB[(优惠券DB)]
    CouponService --> Redis[(Redis)]

    PointsService --> PointsDB[(积分DB)]
    PointsService --> Redis

    ActivityService --> ActivityDB[(活动DB)]
    ActivityService --> Elasticsearch[(ES)]

    CalculationEngine --> RuleEngine[规则引擎]

    MarketingGateway --> OrderService[订单服务]
    MarketingGateway --> ProductService[商品服务]
    MarketingGateway --> PricingService[计价中心]
    MarketingGateway --> UserService[用户服务]

    CouponService --> Kafka[Kafka]
    PointsService --> Kafka
    ActivityService --> Kafka
```

**核心模块协作**可概括为：网关编排，各工具服务自治，计算引擎读多源数据并调用规则引擎；异步事件通过 Kafka 广播给下游（通知、对账、报表等）。

```mermaid
graph LR
    Gateway[营销网关] --> Calc[营销计算引擎]
    Calc --> Rule[规则引擎]
    Calc --> Coupon[优惠券服务]
    Calc --> Points[积分服务]
    Calc --> Activity[活动引擎]
    Order[订单服务] --> Gateway
    Product[商品服务] --> Activity
    Pricing[计价中心] --> Calc
```

**核心常量定义**（类型与状态枚举，便于各服务对齐语义）：

```go
// 营销工具类型
const (
	ToolTypeCoupon   = "coupon"   // 优惠券
	ToolTypePoints   = "points"   // 积分
	ToolTypeActivity = "activity" // 活动
)

// 优惠类型
const (
	DiscountTypeAmount     = "amount"     // 满减（满100减20）
	DiscountTypePercentage = "percentage" // 折扣（8折）
	DiscountTypeFreeShip   = "free_ship"  // 包邮
	DiscountTypeGift       = "gift"       // 赠品
)

// 活动类型
const (
	ActivityTypeFlashSale = "flash_sale" // 秒杀
	ActivityTypeGroupBuy  = "group_buy"  // 拼团
	ActivityTypeSeckill   = "seckill"    // 限时抢购
	ActivityTypeNYuanGou  = "n_yuan_gou" // N元购
)

// 营销状态
const (
	StatusDraft    = "draft"    // 草稿
	StatusPending  = "pending"  // 待审核
	StatusApproved = "approved" // 已通过
	StatusRejected = "rejected" // 已拒绝
	StatusActive   = "active"   // 进行中
	StatusExpired  = "expired"  // 已过期
	StatusCanceled = "canceled" // 已取消
)
```

### 1.5 核心数据模型概览

以下为优惠券、积分、活动相关核心表关系的**逻辑 ER 示意**（实际分库分表与字段以线上为准）。

```mermaid
erDiagram
    COUPON ||--o{ COUPON_USER : "发放"
    COUPON ||--o{ COUPON_LOG : "记录"
    COUPON_USER ||--o{ ORDER : "使用"

    USER ||--o{ POINTS_ACCOUNT : "拥有"
    POINTS_ACCOUNT ||--o{ POINTS_LOG : "记录"

    ACTIVITY ||--o{ ACTIVITY_PRODUCT : "圈品"
    ACTIVITY ||--o{ ACTIVITY_LOG : "记录"
    ORDER ||--o{ ACTIVITY : "使用"

    COUPON {
        bigint coupon_id PK
        string coupon_name
        string discount_type
        decimal discount_value
        decimal min_order_amount
        int total_quantity
        int used_quantity
        datetime start_time
        datetime end_time
        string status
    }

    COUPON_USER {
        bigint id PK
        bigint coupon_id FK
        bigint user_id FK
        string status
        datetime received_at
        datetime used_at
    }

    POINTS_ACCOUNT {
        bigint user_id PK
        bigint available_points
        bigint frozen_points
        bigint total_earned
        bigint total_spent
    }

    ACTIVITY {
        bigint activity_id PK
        string activity_name
        string activity_type
        json rule_config
        datetime start_time
        datetime end_time
        string status
    }
```

### 1.6 技术选型

| 组件 | 技术选型 | 用途 | 理由 |
|------|---------|------|------|
| 数据库 | MySQL 8.0 | 主存储 | ACID 保证、成熟稳定 |
| 缓存 | Redis 6.0 | 热数据缓存、分布式锁 | 高性能、丰富数据结构 |
| 消息队列 | Kafka | 事件驱动、异步解耦 | 高吞吐、持久化 |
| 搜索引擎 | Elasticsearch | 活动搜索、用户画像 | 全文检索、聚合分析 |
| 分布式锁 | Redisson | 秒杀库存扣减 | 基于 Redis、支持可重入 |
| 限流 | Sentinel | 接口限流、降级 | 实时监控、规则灵活 |
| ID 生成 | Snowflake | 营销活动 ID | 分布式、时间有序 |

## 2. 营销工具体系

### 2.1 优惠券系统

#### 2.1.1 优惠券类型与数据模型

优惠券按**平台券 / 商家券**、**满减 / 折扣 / 包邮**等维度组合配置；用户侧以 `CouponUser` 记录领取与核销生命周期，`CouponLog` 用于审计与对账。

```go
// 优惠券主表
type Coupon struct {
	CouponID       int64           `json:"coupon_id"`
	CouponName     string          `json:"coupon_name"`
	CouponType     string          `json:"coupon_type"`    // platform/merchant
	DiscountType   string          `json:"discount_type"`  // amount/percentage/free_ship
	DiscountValue  decimal.Decimal `json:"discount_value"` // 20元 或 0.8（8折）
	MinOrderAmount decimal.Decimal `json:"min_order_amount"`
	MaxDiscountAmt decimal.Decimal `json:"max_discount_amount"` // 折扣券最高抵扣

	TotalQuantity  int64 `json:"total_quantity"`
	UsedQuantity   int64 `json:"used_quantity"`
	RemainQuantity int64 `json:"remain_quantity"`

	PerUserLimit int       `json:"per_user_limit"`
	ValidDays    int       `json:"valid_days"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`

	ApplyScope    string  `json:"apply_scope"`     // all/category/product
	ApplyScopeIDs []int64 `json:"apply_scope_ids"`

	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 用户优惠券表
type CouponUser struct {
	ID         int64      `json:"id"`
	CouponID   int64      `json:"coupon_id"`
	UserID     int64      `json:"user_id"`
	Status     string     `json:"status"` // unused/used/expired
	ReceivedAt time.Time  `json:"received_at"`
	UsedAt     *time.Time `json:"used_at"`
	OrderID    *int64     `json:"order_id"`
	ExpireAt   time.Time  `json:"expire_at"`
}

// 优惠券操作日志
type CouponLog struct {
	ID           int64     `json:"id"`
	CouponUserID int64     `json:"coupon_user_id"`
	CouponID     int64     `json:"coupon_id"`
	UserID       int64     `json:"user_id"`
	Action       string    `json:"action"` // receive/use/expire/rollback
	OrderID      *int64    `json:"order_id"`
	BeforeStatus string    `json:"before_status"`
	AfterStatus  string    `json:"after_status"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"created_at"`
}
```

#### 2.1.2 优惠券发放策略

常见发放方式包括：**公开领取**（先到先得）、**定向推送**（画像圈人）、**裂变发券**（邀请达标）、**订单赠送**（履约后发放）。

公开领券链路强调：Redis 控频次与库存、DB 落库、消息异步刷新与通知。

```mermaid
sequenceDiagram
    participant User as 用户
    participant Gateway as 营销网关
    participant Coupon as 优惠券服务
    participant Redis as Redis
    participant DB as 数据库
    participant Kafka as Kafka

    User->>Gateway: 领取优惠券
    Gateway->>Gateway: 身份验证

    Gateway->>Coupon: ReceiveCoupon(userID, couponID)

    Coupon->>Redis: 检查用户领取次数
    alt 超过限制
        Redis-->>Coupon: 超过限制
        Coupon-->>Gateway: ErrExceedLimit
        Gateway-->>User: 领取失败：已达上限
    end

    Coupon->>Redis: DECR coupon:stock:{couponID}
    alt 库存不足
        Redis-->>Coupon: 库存为0
        Coupon-->>Gateway: ErrStockInsufficient
        Gateway-->>User: 领取失败：已抢光
    end

    Coupon->>DB: 插入 coupon_user 记录
    Coupon->>Redis: INCR user:coupon:count:{userID}:{couponID}

    Coupon->>Kafka: 发送领券事件
    Note over Kafka: 异步更新库存、发送通知

    Coupon-->>Gateway: Success
    Gateway-->>User: 领取成功
```

```go
func (s *CouponService) ReceiveCoupon(ctx context.Context, userID, couponID int64) (*CouponUser, error) {
	// 1. 检查优惠券是否有效
	coupon, err := s.getCouponByID(ctx, couponID)
	if err != nil {
		return nil, err
	}

	if coupon.Status != StatusActive {
		return nil, ErrCouponNotActive
	}

	if time.Now().Before(coupon.StartTime) || time.Now().After(coupon.EndTime) {
		return nil, ErrCouponExpired
	}

	// 2. 检查用户领取次数（Redis）
	userReceiveKey := fmt.Sprintf("user:coupon:count:%d:%d", userID, couponID)
	receivedCount, err := s.redis.Get(ctx, userReceiveKey).Int64()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if receivedCount >= int64(coupon.PerUserLimit) {
		return nil, ErrExceedReceiveLimit
	}

	// 3. Redis 库存扣减（原子操作）
	stockKey := fmt.Sprintf("coupon:stock:%d", couponID)
	remainStock, err := s.redis.Decr(ctx, stockKey).Result()
	if err != nil {
		return nil, err
	}

	if remainStock < 0 {
		// 回滚库存
		s.redis.Incr(ctx, stockKey)
		return nil, ErrCouponStockInsufficient
	}

	// 4. 数据库插入用户优惠券记录
	expireAt := time.Now().Add(time.Duration(coupon.ValidDays) * 24 * time.Hour)
	couponUser := &CouponUser{
		CouponID:   couponID,
		UserID:     userID,
		Status:     CouponStatusUnused,
		ReceivedAt: time.Now(),
		ExpireAt:   expireAt,
	}

	if err := s.db.InsertCouponUser(ctx, couponUser); err != nil {
		// 回滚库存
		s.redis.Incr(ctx, stockKey)
		return nil, err
	}

	// 5. Redis 用户领取次数 +1
	s.redis.Incr(ctx, userReceiveKey)
	s.redis.Expire(ctx, userReceiveKey, 7*24*time.Hour)

	// 6. 记录日志
	s.recordCouponLog(ctx, couponUser.ID, couponID, userID, "receive", "", CouponStatusUnused, "用户领取")

	// 7. 发送 Kafka 事件（异步）
	event := &CouponReceivedEvent{
		CouponUserID: couponUser.ID,
		CouponID:     couponID,
		UserID:       userID,
		ReceivedAt:   time.Now(),
	}
	s.publishCouponEvent(ctx, "coupon.received", event)

	return couponUser, nil
}
```

#### 2.1.3 优惠券核销流程

下单阶段通常先**冻结**，支付成功后再**核销**；取消 / 支付失败则解冻或回退。状态机如下。

```mermaid
stateDiagram-v2
    [*] --> Unused: 用户领取
    Unused --> Frozen: 下单冻结
    Frozen --> Used: 订单支付成功
    Frozen --> Unused: 订单取消/支付失败
    Unused --> Expired: 过期
    Frozen --> Expired: 过期
    Used --> [*]
    Expired --> [*]
```

```go
func (s *CouponService) UseCoupon(ctx context.Context, userID, couponUserID, orderID int64) error {
	// 1. 查询用户优惠券
	couponUser, err := s.db.GetCouponUser(ctx, couponUserID)
	if err != nil {
		return err
	}

	if couponUser.UserID != userID {
		return ErrCouponNotBelongToUser
	}

	if couponUser.Status != CouponStatusUnused && couponUser.Status != CouponStatusFrozen {
		return ErrCouponAlreadyUsed
	}

	if time.Now().After(couponUser.ExpireAt) {
		return ErrCouponExpired
	}

	// 2. 查询优惠券详情（校验适用范围）
	coupon, err := s.getCouponByID(ctx, couponUser.CouponID)
	if err != nil {
		return err
	}

	// 3. 分布式锁（防止并发使用）
	lockKey := fmt.Sprintf("lock:coupon:use:%d", couponUserID)
	lock := s.redisson.GetLock(lockKey)
	if err := lock.Lock(ctx, 3*time.Second); err != nil {
		return ErrCouponLockFailed
	}
	defer lock.Unlock(ctx)

	// 4. 更新优惠券状态为已使用
	now := time.Now()
	if err := s.db.UpdateCouponUserStatus(ctx, couponUserID, CouponStatusUsed, orderID, &now); err != nil {
		return err
	}

	// 5. 优惠券主表已使用数量 +1
	if err := s.db.IncrCouponUsedQuantity(ctx, couponUser.CouponID); err != nil {
		s.logger.Error("increment coupon used quantity failed", zap.Error(err))
	}

	// 6. 记录日志
	s.recordCouponLog(ctx, couponUserID, couponUser.CouponID, userID, "use", CouponStatusFrozen, CouponStatusUsed, fmt.Sprintf("订单%d使用", orderID))

	// 7. 发送 Kafka 事件
	event := &CouponUsedEvent{
		CouponUserID: couponUserID,
		CouponID:     couponUser.CouponID,
		UserID:       userID,
		OrderID:      orderID,
		UsedAt:       now,
	}
	s.publishCouponEvent(ctx, "coupon.used", event)

	return nil
}
```

#### 2.1.4 优惠券回退（订单取消 / 退款）

订单取消或全额退款时，需将用户券恢复为可用（若已过期则标记过期），并同步主表已使用量、写审计日志。

```go
func (s *CouponService) RollbackCoupon(ctx context.Context, userID, couponUserID int64, reason string) error {
	couponUser, err := s.db.GetCouponUser(ctx, couponUserID)
	if err != nil {
		return err
	}

	if couponUser.UserID != userID {
		return ErrCouponNotBelongToUser
	}

	if couponUser.Status != CouponStatusUsed && couponUser.Status != CouponStatusFrozen {
		return ErrCouponCannotRollback
	}

	lockKey := fmt.Sprintf("lock:coupon:rollback:%d", couponUserID)
	lock := s.redisson.GetLock(lockKey)
	if err := lock.Lock(ctx, 3*time.Second); err != nil {
		return ErrCouponLockFailed
	}
	defer lock.Unlock(ctx)

	newStatus := CouponStatusUnused
	if time.Now().After(couponUser.ExpireAt) {
		newStatus = CouponStatusExpired
	}

	if err := s.db.UpdateCouponUserStatus(ctx, couponUserID, newStatus, nil, nil); err != nil {
		return err
	}

	if couponUser.Status == CouponStatusUsed {
		if err := s.db.DecrCouponUsedQuantity(ctx, couponUser.CouponID); err != nil {
			s.logger.Error("decrement coupon used quantity failed", zap.Error(err))
		}
	}

	s.recordCouponLog(ctx, couponUserID, couponUser.CouponID, userID, "rollback", couponUser.Status, newStatus, reason)

	event := &CouponRolledBackEvent{
		CouponUserID: couponUserID,
		CouponID:     couponUser.CouponID,
		UserID:       userID,
		Reason:       reason,
		RolledBackAt: time.Now(),
	}
	s.publishCouponEvent(ctx, "coupon.rolled_back", event)

	return nil
}

### 2.2 积分系统

#### 2.2.1 积分账户模型

积分账户采用**可用 / 冻结**余额与**乐观锁版本号**；流水表支撑对账与审计，`PointsExpire` 供定时任务批量过期。

```go
// 积分账户表
type PointsAccount struct {
	UserID          int64     `json:"user_id"`
	AvailablePoints int64     `json:"available_points"`
	FrozenPoints    int64     `json:"frozen_points"`
	TotalEarned     int64     `json:"total_earned"`
	TotalSpent      int64     `json:"total_spent"`
	Version         int64     `json:"version"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// 积分流水表
type PointsLog struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	ChangeType    string     `json:"change_type"` // earn/spend/freeze/unfreeze/expire
	ChangeAmount  int64      `json:"change_amount"`
	BeforeBalance int64      `json:"before_balance"`
	AfterBalance  int64      `json:"after_balance"`
	BizType       string     `json:"biz_type"`
	BizID         string     `json:"biz_id"`
	Reason        string     `json:"reason"`
	ExpireAt      *time.Time `json:"expire_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

// 积分过期记录表（用于定时任务扫描）
type PointsExpire struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Points      int64      `json:"points"`
	ExpireAt    time.Time  `json:"expire_at"`
	Status      string     `json:"status"` // pending/expired
	ProcessedAt *time.Time `json:"processed_at"`
}
```

#### 2.2.2 积分发放

典型来源：**订单完成返利**、**签到 / 任务**、**邀请好友**、**评价晒单**等。

```go
type EarnPointsRequest struct {
	UserID    int64
	Points    int64
	ValidDays int
	BizType   string
	BizID     string
	Reason    string
}

func (s *PointsService) EarnPoints(ctx context.Context, req *EarnPointsRequest) error {
	if req.Points <= 0 {
		return ErrInvalidPoints
	}

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		account, err := s.db.GetPointsAccount(ctx, req.UserID)
		if err != nil {
			if err == sql.ErrNoRows {
				account = &PointsAccount{
					UserID: req.UserID, AvailablePoints: 0, FrozenPoints: 0,
					TotalEarned: 0, TotalSpent: 0, Version: 0,
				}
				if err := s.db.InsertPointsAccount(ctx, account); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		newAvailable := account.AvailablePoints + req.Points
		newTotalEarned := account.TotalEarned + req.Points

		affected, err := s.db.UpdatePointsAccountWithVersion(ctx, req.UserID, account.Version, newAvailable, account.FrozenPoints, newTotalEarned, account.TotalSpent)
		if err != nil {
			return err
		}

		if affected > 0 {
			expireAt := time.Now().Add(time.Duration(req.ValidDays) * 24 * time.Hour)
			log := &PointsLog{
				UserID: req.UserID, ChangeType: "earn", ChangeAmount: req.Points,
				BeforeBalance: account.AvailablePoints, AfterBalance: newAvailable,
				BizType: req.BizType, BizID: req.BizID, Reason: req.Reason,
				ExpireAt: &expireAt, CreatedAt: time.Now(),
			}
			s.db.InsertPointsLog(ctx, log)

			expire := &PointsExpire{UserID: req.UserID, Points: req.Points, ExpireAt: expireAt, Status: "pending"}
			s.db.InsertPointsExpire(ctx, expire)

			s.publishPointsEvent(ctx, "points.earned", &PointsEarnedEvent{
				UserID: req.UserID, Points: req.Points, BizType: req.BizType, BizID: req.BizID, EarnedAt: time.Now(),
			})
			return nil
		}

		time.Sleep(time.Duration(i*10) * time.Millisecond)
	}

	return ErrPointsUpdateConflict
}
```

#### 2.2.3 积分扣减（订单侧调用）

```go
func (s *PointsService) SpendPoints(ctx context.Context, userID int64, points int64, orderID int64) error {
	if points <= 0 {
		return ErrInvalidPoints
	}

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		account, err := s.db.GetPointsAccount(ctx, userID)
		if err != nil {
			return err
		}

		if account.AvailablePoints < points {
			return ErrPointsInsufficient
		}

		newAvailable := account.AvailablePoints - points
		newTotalSpent := account.TotalSpent + points

		affected, err := s.db.UpdatePointsAccountWithVersion(ctx, userID, account.Version, newAvailable, account.FrozenPoints, account.TotalEarned, newTotalSpent)
		if err != nil {
			return err
		}

		if affected > 0 {
			log := &PointsLog{
				UserID: userID, ChangeType: "spend", ChangeAmount: -points,
				BeforeBalance: account.AvailablePoints, AfterBalance: newAvailable,
				BizType: "order", BizID: fmt.Sprintf("%d", orderID),
				Reason: fmt.Sprintf("订单%d抵扣", orderID), CreatedAt: time.Now(),
			}
			s.db.InsertPointsLog(ctx, log)

			s.publishPointsEvent(ctx, "points.spent", &PointsSpentEvent{
				UserID: userID, Points: points, OrderID: orderID, SpentAt: time.Now(),
			})
			return nil
		}

		time.Sleep(time.Duration(i*10) * time.Millisecond)
	}

	return ErrPointsUpdateConflict
}
```

#### 2.2.4 积分退还（订单取消 / 退款）

```go
func (s *PointsService) RefundPoints(ctx context.Context, userID int64, points int64, orderID int64) error {
	if points <= 0 {
		return ErrInvalidPoints
	}

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		account, err := s.db.GetPointsAccount(ctx, userID)
		if err != nil {
			return err
		}

		newAvailable := account.AvailablePoints + points
		newTotalSpent := account.TotalSpent - points

		affected, err := s.db.UpdatePointsAccountWithVersion(ctx, userID, account.Version, newAvailable, account.FrozenPoints, account.TotalEarned, newTotalSpent)
		if err != nil {
			return err
		}

		if affected > 0 {
			log := &PointsLog{
				UserID: userID, ChangeType: "refund", ChangeAmount: points,
				BeforeBalance: account.AvailablePoints, AfterBalance: newAvailable,
				BizType: "order", BizID: fmt.Sprintf("%d", orderID),
				Reason: fmt.Sprintf("订单%d取消/退款", orderID), CreatedAt: time.Now(),
			}
			s.db.InsertPointsLog(ctx, log)

			s.publishPointsEvent(ctx, "points.refunded", &PointsRefundedEvent{
				UserID: userID, Points: points, OrderID: orderID, RefundedAt: time.Now(),
			})
			return nil
		}

		time.Sleep(time.Duration(i*10) * time.Millisecond)
	}

	return ErrPointsUpdateConflict
}
```

#### 2.2.5 积分过期机制

定时扫描 `PointsExpire` 表中到期且 `pending` 的记录，按乐观锁扣减可用余额并写流水。

```go
func (s *PointsService) ExpirePointsScanner(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processExpiredPoints(ctx)
		}
	}
}

func (s *PointsService) processExpiredPoints(ctx context.Context) {
	expireList, err := s.db.GetPendingExpirePoints(ctx, time.Now())
	if err != nil {
		s.logger.Error("get pending expire points failed", zap.Error(err))
		return
	}

	for _, expire := range expireList {
		if err := s.expirePoints(ctx, expire); err != nil {
			s.logger.Error("expire points failed", zap.Int64("user_id", expire.UserID), zap.Error(err))
		}
	}
}
```

```go
func (s *PointsService) expirePoints(ctx context.Context, expire *PointsExpire) error {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		account, err := s.db.GetPointsAccount(ctx, expire.UserID)
		if err != nil {
			return err
		}

		expireAmount := expire.Points
		if account.AvailablePoints < expireAmount {
			expireAmount = account.AvailablePoints
		}

		if expireAmount <= 0 {
			s.db.UpdatePointsExpireStatus(ctx, expire.ID, "expired")
			return nil
		}

		newAvailable := account.AvailablePoints - expireAmount

		affected, err := s.db.UpdatePointsAccountWithVersion(ctx, expire.UserID, account.Version, newAvailable, account.FrozenPoints, account.TotalEarned, account.TotalSpent)
		if err != nil {
			return err
		}

		if affected > 0 {
			log := &PointsLog{
				UserID: expire.UserID, ChangeType: "expire", ChangeAmount: -expireAmount,
				BeforeBalance: account.AvailablePoints, AfterBalance: newAvailable,
				BizType: "system", BizID: fmt.Sprintf("expire_%d", expire.ID),
				Reason: "积分过期", CreatedAt: time.Now(),
			}
			s.db.InsertPointsLog(ctx, log)

			now := time.Now()
			s.db.UpdatePointsExpireStatusWithTime(ctx, expire.ID, "expired", &now)
			return nil
		}

		time.Sleep(time.Duration(i*10) * time.Millisecond)
	}

	return ErrPointsUpdateConflict
}
```

### 2.3 活动引擎

#### 2.3.1 活动类型

| 活动类型 | 业务逻辑 | 技术挑战 | 适用场景 |
|---------|---------|---------|---------|
| 满减 | 订单满 X 元减 Y 元 | 跨店铺叠加规则 | 提升客单价 |
| 折扣 | 商品打 X 折 | 与优惠券叠加规则 | 清库存 |
| 秒杀 | 限时限量特价 | 高并发、库存扣减 | 引流、造热点 |
| 拼团 | N 人成团享优惠 | 成团判断、超时取消 | 社交裂变 |
| N 元购 | 固定价格购买 | 限购、防刷 | 拉新、促活 |
| 买赠 | 买 A 送 B | 库存联动扣减 | 关联销售 |

#### 2.3.2 活动数据模型

```go
// 活动主表
type Activity struct {
	ActivityID   int64           `json:"activity_id"`
	ActivityName string          `json:"activity_name"`
	ActivityType string          `json:"activity_type"`
	RuleConfig   json.RawMessage `json:"rule_config"`
	ApplyScope   string          `json:"apply_scope"`
	StartTime    time.Time       `json:"start_time"`
	EndTime      time.Time       `json:"end_time"`
	TotalStock   int64           `json:"total_stock"`
	UsedStock    int64           `json:"used_stock"`
	Status       string          `json:"status"`
	CreatedBy    int64           `json:"created_by"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// 活动圈品表
type ActivityProduct struct {
	ID            int64           `json:"id"`
	ActivityID    int64           `json:"activity_id"`
	ProductID     int64           `json:"product_id"`
	SKUID         int64           `json:"sku_id"`
	OriginalPrice decimal.Decimal `json:"original_price"`
	ActivityPrice decimal.Decimal `json:"activity_price"`
	ActivityStock int64           `json:"activity_stock"`
	SoldCount     int64           `json:"sold_count"`
	CreatedAt     time.Time       `json:"created_at"`
}

// 满减规则示例
type FullReductionRule struct {
	Tiers []FullReductionTier `json:"tiers"`
}

type FullReductionTier struct {
	MinAmount      decimal.Decimal `json:"min_amount"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
}

// 秒杀规则示例
type FlashSaleRule struct {
	PerUserLimit int  `json:"per_user_limit"`
	NeedVerify   bool `json:"need_verify"`
}
```

#### 2.3.3 活动状态机

```mermaid
stateDiagram-v2
    [*] --> Draft: 创建活动
    Draft --> Pending: 提交审核
    Pending --> Approved: 审核通过
    Pending --> Rejected: 审核拒绝
    Rejected --> Draft: 修改后重新提交
    Approved --> Active: 到达开始时间
    Active --> Expired: 到达结束时间
    Active --> Canceled: 手动取消
    Approved --> Canceled: 手动取消
    Expired --> [*]
    Canceled --> [*]
```

#### 2.3.4 圈品规则

支持**全场**、**指定类目**、**指定商品 / SKU**，并可扩展**排除规则**（例如已参加互斥活动的商品）。

```go
func (s *ActivityService) IsProductEligible(ctx context.Context, activityID, productID, skuID int64) (bool, error) {
	activity, err := s.getActivityByID(ctx, activityID)
	if err != nil {
		return false, err
	}

	if activity.Status != StatusActive {
		return false, nil
	}

	now := time.Now()
	if now.Before(activity.StartTime) || now.After(activity.EndTime) {
		return false, nil
	}

	switch activity.ApplyScope {
	case "all":
		return true, nil

	case "category":
		product, err := s.productClient.GetProduct(ctx, productID)
		if err != nil {
			return false, err
		}

		categories, err := s.db.GetActivityCategories(ctx, activityID)
		if err != nil {
			return false, err
		}

		for _, catID := range categories {
			if product.CategoryID == catID {
				return true, nil
			}
		}
		return false, nil

	case "product":
		activityProduct, err := s.db.GetActivityProduct(ctx, activityID, productID, skuID)
		if err != nil {
			if err == sql.ErrNoRows {
				return false, nil
			}
			return false, err
		}

		if activityProduct.SoldCount >= activityProduct.ActivityStock {
			return false, nil
		}

		return true, nil

	default:
		return false, nil
	}
}

## 3. 营销计算引擎

### 3.1 优惠计算流程

营销计算在购物车 / 结算页被高频调用：需要聚合用户券、活动、积分，应用叠加与互斥规则，输出**最优方案**并**分摊到行**。

```mermaid
graph TD
    Start[开始计算] --> FetchOrder[获取订单信息]
    FetchOrder --> FetchCoupon[获取用户优惠券]
    FetchOrder --> FetchPoints[获取用户积分]
    FetchOrder --> FetchActivity[获取适用活动]

    FetchCoupon --> FilterCoupon[筛选可用优惠券]
    FetchActivity --> FilterActivity[筛选可用活动]

    FilterCoupon --> CalcCoupon[计算优惠券优惠]
    FilterActivity --> CalcActivity[计算活动优惠]
    FetchPoints --> CalcPoints[计算积分抵扣]

    CalcCoupon --> ApplyRules[应用叠加/互斥规则]
    CalcActivity --> ApplyRules
    CalcPoints --> ApplyRules

    ApplyRules --> SelectBest[选择最优方案]
    SelectBest --> AllocDiscount[优惠分摊到商品]
    AllocDiscount --> CalcFinal[计算最终应付金额]
    CalcFinal --> End[返回结果]
```

```go
type MarketingCalculationEngine struct {
	couponService   *CouponService
	pointsService   *PointsService
	activityService *ActivityService
	ruleEngine      *RuleEngine
}

type CalculateRequest struct {
	UserID    int64        `json:"user_id"`
	Items     []*OrderItem `json:"items"`
	CouponIDs []int64      `json:"coupon_ids"`
	UsePoints int64        `json:"use_points"`
}

type OrderItem struct {
	ProductID int64           `json:"product_id"`
	SKUID     int64           `json:"sku_id"`
	Quantity  int             `json:"quantity"`
	Price     decimal.Decimal `json:"price"`
	Amount    decimal.Decimal `json:"amount"`
}

type CalculateResponse struct {
	OriginalAmount   decimal.Decimal `json:"original_amount"`
	CouponDiscount   decimal.Decimal `json:"coupon_discount"`
	ActivityDiscount decimal.Decimal `json:"activity_discount"`
	PointsDiscount   decimal.Decimal `json:"points_discount"`
	TotalDiscount    decimal.Decimal `json:"total_discount"`
	FinalAmount      decimal.Decimal `json:"final_amount"`

	ItemDiscounts  []*ItemDiscount `json:"item_discounts"`
	UsedCoupons    []*UsedCoupon   `json:"used_coupons"`
	UsedActivities []*UsedActivity `json:"used_activities"`
	UsedPoints     int64           `json:"used_points"`
}

func (e *MarketingCalculationEngine) Calculate(ctx context.Context, req *CalculateRequest) (*CalculateResponse, error) {
	originalAmount := decimal.Zero
	for _, item := range req.Items {
		originalAmount = originalAmount.Add(item.Amount)
	}

	availableCoupons, err := e.getAvailableCoupons(ctx, req.UserID, req.CouponIDs, req.Items)
	if err != nil {
		return nil, err
	}

	availableActivities, err := e.getAvailableActivities(ctx, req.Items)
	if err != nil {
		return nil, err
	}

	pointsDiscount := decimal.NewFromInt(req.UsePoints).Div(decimal.NewFromInt(100))

	bestPlan, err := e.ruleEngine.FindBestCombination(ctx, originalAmount, availableCoupons, availableActivities, pointsDiscount)
	if err != nil {
		return nil, err
	}

	itemDiscounts := e.allocateDiscountToItems(req.Items, bestPlan)

	totalDiscount := bestPlan.CouponDiscount.Add(bestPlan.ActivityDiscount).Add(pointsDiscount)
	finalAmount := originalAmount.Sub(totalDiscount)
	if finalAmount.LessThan(decimal.Zero) {
		finalAmount = decimal.Zero
	}

	return &CalculateResponse{
		OriginalAmount:   originalAmount,
		CouponDiscount:   bestPlan.CouponDiscount,
		ActivityDiscount: bestPlan.ActivityDiscount,
		PointsDiscount:   pointsDiscount,
		TotalDiscount:    totalDiscount,
		FinalAmount:      finalAmount,
		ItemDiscounts:    itemDiscounts,
		UsedCoupons:      bestPlan.UsedCoupons,
		UsedActivities:   bestPlan.UsedActivities,
		UsedPoints:       req.UsePoints,
	}, nil
}
```

### 3.2 优惠叠加与互斥规则

常见规则：**同一订单一张券**；**活动与券可叠加**（示例实现为**先活动后券**）；**积分可与券、活动叠加**；**跨店铺**需单独策略（不同店铺优惠不可简单合并）。

```go
type RuleEngine struct{}

type PromotionPlan struct {
	CouponDiscount   decimal.Decimal
	ActivityDiscount decimal.Decimal
	UsedCoupons      []*UsedCoupon
	UsedActivities   []*UsedActivity
	TotalDiscount    decimal.Decimal
}

type UsedCoupon struct {
	CouponID       int64
	CouponUserID   int64
	DiscountAmount decimal.Decimal
}

type UsedActivity struct {
	ActivityID     int64
	DiscountAmount decimal.Decimal
}

func (r *RuleEngine) FindBestCombination(
	ctx context.Context,
	originalAmount decimal.Decimal,
	availableCoupons []*Coupon,
	availableActivities []*Activity,
	pointsDiscount decimal.Decimal,
) (*PromotionPlan, error) {

	var bestPlan *PromotionPlan
	maxDiscount := decimal.Zero

	for _, coupon := range availableCoupons {
		activityCombinations := r.generateActivityCombinations(availableActivities)

		for _, activityCombo := range activityCombinations {
			plan := r.calculatePlan(originalAmount, coupon, activityCombo)

			if plan.TotalDiscount.GreaterThan(maxDiscount) {
				maxDiscount = plan.TotalDiscount
				bestPlan = plan
			}
		}
	}

	if bestPlan == nil {
		activityCombinations := r.generateActivityCombinations(availableActivities)
		for _, activityCombo := range activityCombinations {
			plan := r.calculatePlan(originalAmount, nil, activityCombo)
			if plan.TotalDiscount.GreaterThan(maxDiscount) {
				maxDiscount = plan.TotalDiscount
				bestPlan = plan
			}
		}
	}

	if bestPlan == nil {
		bestPlan = &PromotionPlan{
			CouponDiscount:   decimal.Zero,
			ActivityDiscount: decimal.Zero,
			TotalDiscount:    decimal.Zero,
		}
	}

	return bestPlan, nil
}
```

```go
func (r *RuleEngine) calculatePlan(
	originalAmount decimal.Decimal,
	coupon *Coupon,
	activities []*Activity,
) *PromotionPlan {
	plan := &PromotionPlan{
		CouponDiscount:   decimal.Zero,
		ActivityDiscount: decimal.Zero,
		UsedCoupons:      []*UsedCoupon{},
		UsedActivities:   []*UsedActivity{},
	}

	currentAmount := originalAmount

	for _, activity := range activities {
		activityDiscount := r.calculateActivityDiscount(currentAmount, activity)
		if activityDiscount.GreaterThan(decimal.Zero) {
			plan.ActivityDiscount = plan.ActivityDiscount.Add(activityDiscount)
			plan.UsedActivities = append(plan.UsedActivities, &UsedActivity{
				ActivityID:     activity.ActivityID,
				DiscountAmount: activityDiscount,
			})
			currentAmount = currentAmount.Sub(activityDiscount)
		}
	}

	if coupon != nil {
		couponDiscount := r.calculateCouponDiscount(currentAmount, coupon)
		if couponDiscount.GreaterThan(decimal.Zero) {
			plan.CouponDiscount = couponDiscount
			plan.UsedCoupons = append(plan.UsedCoupons, &UsedCoupon{
				CouponID:       coupon.CouponID,
				DiscountAmount: couponDiscount,
			})
		}
	}

	plan.TotalDiscount = plan.ActivityDiscount.Add(plan.CouponDiscount)

	return plan
}
```

```go
func (r *RuleEngine) calculateCouponDiscount(amount decimal.Decimal, coupon *Coupon) decimal.Decimal {
	if amount.LessThan(coupon.MinOrderAmount) {
		return decimal.Zero
	}

	switch coupon.DiscountType {
	case DiscountTypeAmount:
		return coupon.DiscountValue

	case DiscountTypePercentage:
		discount := amount.Mul(decimal.NewFromInt(1).Sub(coupon.DiscountValue))
		if coupon.MaxDiscountAmt.GreaterThan(decimal.Zero) && discount.GreaterThan(coupon.MaxDiscountAmt) {
			return coupon.MaxDiscountAmt
		}
		return discount

	default:
		return decimal.Zero
	}
}
```

```go
func (r *RuleEngine) calculateActivityDiscount(amount decimal.Decimal, activity *Activity) decimal.Decimal {
	switch activity.ActivityType {
	case ActivityTypeFlashSale:
		return decimal.Zero

	case "full_reduction":
		var rule FullReductionRule
		if err := json.Unmarshal(activity.RuleConfig, &rule); err != nil {
			return decimal.Zero
		}

		var maxTier *FullReductionTier
		for i := range rule.Tiers {
			tier := &rule.Tiers[i]
			if amount.GreaterThanOrEqual(tier.MinAmount) {
				if maxTier == nil || tier.MinAmount.GreaterThan(maxTier.MinAmount) {
					maxTier = tier
				}
			}
		}

		if maxTier != nil {
			return maxTier.DiscountAmount
		}
		return decimal.Zero

	default:
		return decimal.Zero
	}
}

func (r *RuleEngine) generateActivityCombinations(activities []*Activity) [][]*Activity {
	if len(activities) == 0 {
		return [][]*Activity{{}}
	}
	return [][]*Activity{activities}
}
```

### 3.3 优惠分摊到商品

按**商品行金额占比**分摊券 / 活动 / 积分优惠，并对**尾差**做最后一行兜底，避免舍入导致总额不平。

```go
type ItemDiscount struct {
	ProductID        int64           `json:"product_id"`
	SKUID            int64           `json:"sku_id"`
	OriginalAmount   decimal.Decimal `json:"original_amount"`
	CouponDiscount   decimal.Decimal `json:"coupon_discount"`
	ActivityDiscount decimal.Decimal `json:"activity_discount"`
	PointsDiscount   decimal.Decimal `json:"points_discount"`
	FinalAmount      decimal.Decimal `json:"final_amount"`
}

func (e *MarketingCalculationEngine) allocateDiscountToItems(
	items []*OrderItem,
	plan *PromotionPlan,
) []*ItemDiscount {

	itemDiscounts := make([]*ItemDiscount, len(items))

	totalAmount := decimal.Zero
	for _, item := range items {
		totalAmount = totalAmount.Add(item.Amount)
	}

	allocatedCouponDiscount := decimal.Zero
	for i, item := range items {
		ratio := item.Amount.Div(totalAmount)
		discount := plan.CouponDiscount.Mul(ratio).Round(2)

		itemDiscounts[i] = &ItemDiscount{
			ProductID:        item.ProductID,
			SKUID:            item.SKUID,
			OriginalAmount:   item.Amount,
			CouponDiscount:   discount,
			ActivityDiscount: decimal.Zero,
			PointsDiscount:   decimal.Zero,
		}

		allocatedCouponDiscount = allocatedCouponDiscount.Add(discount)
	}

	couponDiff := plan.CouponDiscount.Sub(allocatedCouponDiscount)
	if len(itemDiscounts) > 0 && couponDiff.Abs().GreaterThan(decimal.NewFromFloat(0.01)) {
		last := itemDiscounts[len(itemDiscounts)-1]
		last.CouponDiscount = last.CouponDiscount.Add(couponDiff)
	}

	// 活动优惠、积分抵扣可按同样比例分摊（此处省略类似代码）

	for i := range itemDiscounts {
		itemDiscounts[i].FinalAmount = itemDiscounts[i].OriginalAmount.
			Sub(itemDiscounts[i].CouponDiscount).
			Sub(itemDiscounts[i].ActivityDiscount).
			Sub(itemDiscounts[i].PointsDiscount)
	}

	return itemDiscounts
}
```

## 4. 高并发场景设计

### 4.1 秒杀 / 抢券设计

#### 4.1.1 秒杀系统架构

秒杀链路强调：**边缘削峰**（CDN、验证码、网关限流）、**热点库存**（Redis 预扣 + 分布式锁）、**异步落单**（消息队列 + Worker），并与库存 DB **最终一致**同步。

```mermaid
graph TB
    User[用户] --> CDN[CDN静态资源]
    User --> Gateway[API网关]

    Gateway --> Captcha[验证码服务]
    Gateway --> RateLimit[限流组件 Sentinel]

    RateLimit --> SecKill[秒杀服务]

    SecKill --> RedisCluster[Redis集群 分布式锁]
    SecKill --> LocalCache[本地缓存 Caffeine]
    SecKill --> MQ[消息队列 Kafka]

    MQ --> OrderWorker[订单处理Worker]
    OrderWorker --> OrderDB[(订单DB)]

    SecKill -.异步扣减.-> InventoryDB[(库存DB)]
```

#### 4.1.2 流量削峰

常用手段：**CDN 加速**静态资源、**验证码 / 答题**延缓脚本、**排队**或**令牌桶**在网关侧削峰。

```go
func (s *SeckillService) VerifyCaptcha(ctx context.Context, userID int64, captchaID, captchaCode string) error {
	key := fmt.Sprintf("captcha:%s", captchaID)
	correctCode, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCaptchaExpired
		}
		return err
	}

	if correctCode != captchaCode {
		return ErrCaptchaInvalid
	}

	s.redis.Del(ctx, key)

	return nil
}
```

#### 4.1.3 分布式锁

```go
import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
)

func (s *SeckillService) SecKillProduct(ctx context.Context, req *SeckillRequest) error {
	if err := s.VerifyCaptcha(ctx, req.UserID, req.CaptchaID, req.CaptchaCode); err != nil {
		return err
	}

	userPurchaseKey := fmt.Sprintf("seckill:user:%d:activity:%d", req.UserID, req.ActivityID)
	count, err := s.redis.Get(ctx, userPurchaseKey).Int()
	if err != nil && err != redis.Nil {
		return err
	}

	activity, _ := s.getActivity(ctx, req.ActivityID)
	if count >= activity.PerUserLimit {
		return ErrExceedPurchaseLimit
	}

	stockKey := fmt.Sprintf("seckill:stock:%d", req.ActivityID)
	lockKey := fmt.Sprintf("lock:seckill:stock:%d", req.ActivityID)

	pool := goredis.NewPool(s.redisClient)
	rs := redsync.New(pool)
	mutex := rs.NewMutex(lockKey, redsync.WithExpiry(3*time.Second))

	if err := mutex.Lock(); err != nil {
		return ErrSecKillBusy
	}
	defer mutex.Unlock()

	stock, err := s.redis.Get(ctx, stockKey).Int64()
	if err != nil {
		return err
	}

	if stock <= 0 {
		return ErrSecKillStockOut
	}

	newStock, err := s.redis.Decr(ctx, stockKey).Result()
	if err != nil {
		return err
	}

	if newStock < 0 {
		s.redis.Incr(ctx, stockKey)
		return ErrSecKillStockOut
	}

	orderMsg := &SeckillOrderMessage{
		UserID: req.UserID, ActivityID: req.ActivityID, ProductID: req.ProductID,
		SKUID: req.SKUID, Quantity: 1, Timestamp: time.Now(),
	}

	if err := s.publishSeckillOrder(ctx, orderMsg); err != nil {
		s.redis.Incr(ctx, stockKey)
		return err
	}

	s.redis.Incr(ctx, userPurchaseKey)
	s.redis.Expire(ctx, userPurchaseKey, 24*time.Hour)

	return nil
}
```

#### 4.1.4 库存预扣与异步确认

Redis **预扣**保证热点路径低延迟；**Kafka** 异步创单与落库；定时任务校准 Redis 与 DB 库存。

```mermaid
sequenceDiagram
    participant User as 用户
    participant Seckill as 秒杀服务
    participant Redis as Redis
    participant Kafka as Kafka
    participant Worker as 订单Worker
    participant DB as 数据库

    User->>Seckill: 秒杀请求
    Seckill->>Seckill: 验证码验证
    Seckill->>Redis: 获取分布式锁
    Redis-->>Seckill: 锁获取成功

    Seckill->>Redis: DECR stock
    alt 库存充足
        Redis-->>Seckill: 扣减成功
        Seckill->>Kafka: 发送订单消息
        Seckill-->>User: 抢购成功，订单生成中

        Kafka->>Worker: 消费订单消息
        Worker->>DB: 创建订单
        Worker->>DB: 扣减DB库存
        Worker-->>User: 推送订单创建成功
    else 库存不足
        Redis-->>Seckill: 库存为0
        Seckill-->>User: 商品已抢光
    end
```

### 4.2 防刷防薅

#### 4.2.1 用户行为风控

结合**设备指纹**、**IP 限流**、**行为序列分析**与**黑名单**；接口侧用**滑动窗口**限制单用户调用频率。

```go
func (s *MarketingService) CheckRateLimit(ctx context.Context, userID int64, action string) error {
	blacklistKey := fmt.Sprintf("blacklist:user:%d", userID)
	exists, err := s.redis.Exists(ctx, blacklistKey).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return ErrUserInBlacklist
	}

	key := fmt.Sprintf("ratelimit:%s:%d", action, userID)

	now := time.Now().Unix()
	windowStart := now - 60

	pipe := s.redis.Pipeline()

	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	countCmd := pipe.ZCount(ctx, key, fmt.Sprintf("%d", windowStart), fmt.Sprintf("%d", now))

	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})

	pipe.Expire(ctx, key, 2*time.Minute)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	count := countCmd.Val()
	if count >= 10 {
		return ErrRateLimitExceeded
	}

	return nil
}
```

#### 4.2.2 营销预算控制

活动维度维护**剩余预算**，下单 / 核销前校验；扣减建议用 **Lua** 保证原子性。

```go
type BudgetController struct {
	redis *redis.Client
}

func (c *BudgetController) CheckBudget(ctx context.Context, activityID int64, amount decimal.Decimal) error {
	budgetKey := fmt.Sprintf("activity:budget:%d", activityID)

	remainBudget, err := c.redis.Get(ctx, budgetKey).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrBudgetExhausted
		}
		return err
	}

	remain, _ := decimal.NewFromString(remainBudget)
	if remain.LessThan(amount) {
		return ErrBudgetInsufficient
	}

	return nil
}
```

```go
func (c *BudgetController) DeductBudget(ctx context.Context, activityID int64, amount decimal.Decimal) error {
	budgetKey := fmt.Sprintf("activity:budget:%d", activityID)

	luaScript := `
        local budget_key = KEYS[1]
        local amount = tonumber(ARGV[1])
        local remain = tonumber(redis.call('GET', budget_key) or 0)

        if remain >= amount then
            redis.call('DECRBY', budget_key, amount)
            return 1
        else
            return 0
        end
    `

	result, err := c.redis.Eval(ctx, luaScript, []string{budgetKey}, amount.String()).Int()
	if err != nil {
		return err
	}

	if result == 0 {
		return ErrBudgetInsufficient
	}

	return nil
}
```
```
```
