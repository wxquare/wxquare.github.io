---
title: 电商系统设计（八）：支付系统深度解析
date: 2025-08-07
categories:
  - 电商系统设计
tags:
  - payment
  - distributed-transaction
  - idempotency
  - state-machine
  - saga
  - tcc
  - reconciliation
---

> **电商系统设计系列**（篇次与[（一）推荐阅读顺序](/system-design/20-ecommerce-overview/)一致）
> - [（一）全景概览与领域划分](/system-design/20-ecommerce-overview/)
> - [（二）商品中心系统](/system-design/21-ecommerce-product-center/)
> - [（三）库存系统](/system-design/22-ecommerce-inventory/)
> - [（四）营销系统深度解析](/system-design/23-ecommerce-marketing-system/)
> - [（五）计价引擎](/system-design/24-ecommerce-pricing-engine/)
> - [（六）计价系统 DDD 实践](/system-design/25-ecommerce-pricing-ddd/)
> - [（七）订单系统](/system-design/26-ecommerce-order-system/)
> - **（八）支付系统深度解析**（本文）
> - [（九）商品上架系统](/system-design/28-ecommerce-listing/)
> - [（十）B 端运营系统](/system-design/29-ecommerce-b-side-ops/)

## 引言

支付系统是电商平台的资金流枢纽，连接用户、平台、商家、第三方支付等多方角色。本文从系统设计面试的角度，深入解析支付系统的核心流程、状态机设计、分布式事务等高频考点。

**适合读者**：准备系统设计面试的候选人

**阅读时长**：30-40 分钟

**核心内容**：
- 支付系统整体架构
- 支付和退款流程
- 状态机设计
- 分布式事务（Saga/TCC）
- 幂等性设计
- 一致性保证

## 一、业务背景与挑战

### 1.1 支付系统的定位

支付系统是电商平台的**资金流枢纽**，承担以下职责：

- **C 端**：用户支付、退款、余额管理
- **B 端**：商家结算、提现、对账
- **平台**：分账、风控、审计

支付系统与其他系统的协作关系：

```mermaid
graph LR
    A[用户] -->|支付| B[支付系统]
    B -->|状态同步| C[订单系统]
    B -->|优惠计算| D[营销系统]
    B -->|扣减库存| E[库存系统]
    B -->|触发发货| F[物流系统]
```

### 1.2 核心业务场景

| 场景 | 说明 | 关键点 |
|-----|------|-------|
| 标准支付 | 用户使用余额、微信、支付宝支付订单 | 组合支付、渠道路由 |
| 退款 | 全额退款、部分退款、营销优惠退款 | 可退金额计算、多次退款 |
| 清结算 | 平台佣金、商家收益、营销补贴分账 | T+N 结算、提现管理 |
| 对账 | 交易对账、资金对账、差错处理 | 长款、短款、人工复核 |

### 1.3 核心挑战

支付系统面临的核心挑战及对应技术方案：

| 挑战维度 | 具体问题 | 技术方案 |
|---------|---------|---------|
| **资金安全** | 强一致性、防重防篡改、审计追溯 | 分布式事务、幂等性、操作日志 |
| **高并发** | 大促期间支付峰值（如双 11） | Redis 缓存、限流、降级、异步化 |
| **分布式事务** | 支付成功后订单状态同步 | Saga、TCC、本地消息表 |
| **多渠道接入** | 微信、支付宝、银行等渠道差异 | 支付网关、策略模式、适配器模式 |
| **对账复杂度** | 多方对账、差错处理 | 定时任务、对账算法、人工复核 |

**面试重点**：

在系统设计面试中，面试官通常会从以下角度考察：

1. **如何保证支付与订单的最终一致性？** → 分布式事务
2. **如何防止用户重复支付？** → 幂等性设计
3. **支付系统如何应对高并发？** → 缓存、限流、降级
4. **第三方支付回调失败怎么办？** → 重试机制、补偿

## 二、整体架构设计

### 2.1 分层架构

支付系统采用经典的分层架构，每层职责清晰：

```mermaid
graph TB
    subgraph 接入层
        A1[C 端小程序/APP]
        A2[B 端运营后台]
        A3[Open API]
    end
    
    subgraph 应用服务层
        B1[订单服务]
        B2[支付服务]
        B3[对账服务]
    end
    
    subgraph 核心业务层
        C1[账户系统]
        C2[支付网关]
        C3[清结算引擎]
        C4[风控系统]
    end
    
    subgraph 基础设施层
        D1[MySQL]
        D2[Redis]
        D3[Kafka]
        D4[XXL-Job]
    end
    
    subgraph 第三方服务层
        E1[微信支付]
        E2[支付宝]
        E3[银行网关]
    end
    
    A1 --> B2
    A2 --> B2
    A3 --> B2
    B1 --> C2
    B2 --> C1
    B2 --> C2
    B2 --> C3
    B2 --> C4
    C1 --> D1
    C1 --> D2
    C2 --> D3
    C3 --> D4
    C2 --> E1
    C2 --> E2
    C2 --> E3
```

### 2.2 核心子系统

| 子系统 | 核心职责 | 关键技术 |
|-------|---------|---------|
| **账户系统** | 管理用户账户、商家账户、平台账户<br>- 余额查询、冻结/解冻<br>- 充值、提现<br>- 账户流水 | - Redis + MySQL 双写<br>- 账户流水表<br>- 定时对账 |
| **支付网关** | 统一支付入口，屏蔽第三方差异<br>- 渠道抽象（适配器模式）<br>- 路由策略（余额优先、组合支付）<br>- 重试补偿 | - 策略模式<br>- 适配器模式<br>- 异步回调 |
| **清结算引擎** | 分账计算和结算管理<br>- 分账规则（平台佣金、商家收益）<br>- 结算周期（T+1、T+7）<br>- 提现管理 | - 定时任务<br>- 分账算法<br>- 限额控制 |
| **风控系统** | 保障资金安全<br>- 支付密码验证<br>- 异常交易监控<br>- 限额控制 | - 规则引擎<br>- 实时监控<br>- 黑名单 |

### 2.3 设计原则

在架构设计中，遵循以下原则：

**1. 领域边界清晰**

账户、支付、清结算、对账等各司其职，通过明确的接口交互。

**2. 事件驱动解耦**

系统间通过 Kafka 事件解耦，支付服务发布"支付成功"事件，订单服务订阅并更新状态。

**3. 可扩展性**

- 支持新支付渠道接入（如数字货币）
- 支持新支付方式（如分期付款）
- 支持新业务场景（如预售、拼团）

**4. 可观测性**

- 结构化日志（JSON 格式）
- 全链路追踪（TraceID）
- 实时监控告警（Prometheus + Grafana）

**面试加分项**：

能够在白板上快速画出以上架构图，并说明各层职责，会给面试官留下深刻印象。

## 三、核心业务流程

### 3.1 支付流程

#### 3.1.1 时序图

```mermaid
sequenceDiagram
    participant 用户
    participant 订单服务
    participant 支付服务
    participant 支付网关
    participant 第三方支付

    用户->>订单服务: 1. 下单
    订单服务->>支付服务: 2. 创建支付单
    支付服务->>支付服务: 3. 幂等性检查
    支付服务->>支付网关: 4. 渠道路由
    支付网关->>第三方支付: 5. 调用支付接口
    第三方支付-->>支付网关: 6. 同步返回
    支付网关-->>用户: 7. 返回支付页面
    用户->>第三方支付: 8. 完成支付
    第三方支付->>支付网关: 9. 异步回调
    支付网关->>支付服务: 10. 更新支付状态
    支付服务->>订单服务: 11. 通知订单状态变更
    支付服务->>用户: 12. 推送支付成功通知
```

#### 3.1.2 关键步骤

**Step 1: 创建支付单（幂等性保证）**

```go
// 支付单创建接口
func CreatePaymentOrder(req *CreatePaymentRequest) (*PaymentOrder, error) {
    // 1. 生成幂等键（order_id + user_id）
    idempotencyKey := fmt.Sprintf("%d_%d", req.OrderID, req.UserID)
    
    // 2. Redis 分布式锁（防并发）
    lock := redis.Lock(idempotencyKey, 10*time.Second)
    if !lock.TryLock() {
        return nil, errors.New("concurrent request, please retry")
    }
    defer lock.Unlock()
    
    // 3. 检查是否已存在
    existing := queryPaymentByIdempotencyKey(idempotencyKey)
    if existing != nil {
        return existing, nil  // 幂等返回
    }
    
    // 4. 创建支付单
    payment := &PaymentOrder{
        PaymentID:      snowflake.Generate(),
        OrderID:        req.OrderID,
        UserID:         req.UserID,
        PaymentAmount:  req.Amount,
        PaymentStatus:  "PENDING",
        IdempotencyKey: idempotencyKey,
    }
    
    // 5. 插入数据库（唯一索引保证幂等）
    if err := db.Insert(payment); err != nil {
        if isDuplicateKeyError(err) {
            return queryPaymentByIdempotencyKey(idempotencyKey), nil
        }
        return nil, err
    }
    
    return payment, nil
}
```

**Step 2: 渠道路由**

支付网关根据策略选择支付渠道：

1. **余额优先策略**：用户余额足够则优先使用余额
2. **组合支付**：余额不足时，余额 + 第三方支付
3. **渠道降级**：主渠道不可用时切换备用渠道

**Step 3: 异步回调处理（幂等性保证）**

```go
// 处理第三方支付回调
func HandlePaymentCallback(callbackData *CallbackData) error {
    // 1. 验签
    if !verifySignature(callbackData) {
        return errors.New("invalid signature")
    }
    
    // 2. 幂等检查（第三方交易号）
    existing := queryPaymentByChannelTradeNo(callbackData.ChannelTradeNo)
    if existing != nil && existing.PaymentStatus == "SUCCESS" {
        return nil  // 已处理过，直接返回
    }
    
    // 3. 更新支付状态（乐观锁）
    affected := db.Exec(`
        UPDATE payment_order 
        SET payment_status = 'SUCCESS',
            channel_trade_no = ?,
            callback_time = ?,
            version = version + 1
        WHERE payment_id = ? AND version = ?
    `, callbackData.ChannelTradeNo, time.Now(), 
       callbackData.PaymentID, callbackData.Version)
    
    if affected == 0 {
        return errors.New("concurrent update conflict")
    }
    
    // 4. 发布支付成功事件（Saga）
    publishPaymentSuccessEvent(callbackData.PaymentID, callbackData.OrderID)
    
    return nil
}
```

**Step 4: 重试机制**

第三方回调可能失败，需要重试机制：

- **主动重试**：最多 3 次，指数退避（1s, 2s, 4s）
- **定时补偿**：每分钟扫描超时支付单，主动查询第三方状态
- **人工介入**：超过重试次数，进入人工复核

### 3.2 退款流程

#### 3.2.1 时序图

```mermaid
sequenceDiagram
    participant 用户
    participant 订单服务
    participant 支付服务
    participant 第三方支付
    participant 账户系统

    用户->>订单服务: 1. 申请退款
    订单服务->>支付服务: 2. 创建退款单
    支付服务->>支付服务: 3. 校验可退金额
    支付服务->>第三方支付: 4. 调用退款接口
    第三方支付-->>支付服务: 5. 退款成功
    支付服务->>账户系统: 6. 账户入账
    支付服务->>订单服务: 7. 通知退款成功
    支付服务->>用户: 8. 推送退款通知
```

#### 3.2.2 可退金额计算

```go
// 计算可退金额
func CalculateRefundableAmount(paymentID int64) (*RefundableAmount, error) {
    // 1. 查询支付单
    payment := queryPaymentOrder(paymentID)
    if payment.PaymentStatus != "SUCCESS" {
        return nil, errors.New("payment not success")
    }
    
    // 2. 查询已退款金额
    refundedAmount := sumRefundedAmount(paymentID)
    
    // 3. 可退金额 = 实付金额 - 已退金额
    refundable := payment.PaymentAmount.Sub(refundedAmount)
    if refundable.LessThanOrEqual(decimal.Zero) {
        return nil, errors.New("no refundable amount")
    }
    
    // 4. 营销优惠按比例退款
    // 例如：实付 100（原价 120，优惠 20），退款 50
    // 则退实付 50，营销优惠退 10
    result := &RefundableAmount{
        TotalRefundable: refundable,
    }
    
    if payment.PromotionAmount.GreaterThan(decimal.Zero) {
        ratio := refundable.Div(payment.PaymentAmount)
        result.PromotionRefund = payment.PromotionAmount.Mul(ratio)
        result.ActualRefund = refundable.Sub(result.PromotionRefund)
    } else {
        result.ActualRefund = refundable
    }
    
    return result, nil
}
```

#### 3.2.3 部分退款

支持多次部分退款：

- **累计限制**：所有退款金额之和 ≤ 实付金额
- **退款记录**：每次退款都生成独立的退款单
- **状态联动**：全额退款后，支付单状态变为 `REFUNDED`

### 3.3 异步回调处理

```mermaid
graph LR
    A[第三方回调] --> B{签名校验}
    B -->|失败| C[拒绝并记录]
    B -->|成功| D{幂等检查}
    D -->|已处理| E[直接返回 SUCCESS]
    D -->|未处理| F[更新支付状态]
    F --> G[发布事件到 Kafka]
    G --> H[返回 SUCCESS]
```

**回调要点**：

1. **签名校验**：防止伪造回调
2. **幂等检查**：第三方交易号去重
3. **乐观锁**：version 字段防止并发
4. **事件驱动**：发布到 Kafka，解耦订单服务

## 四、状态机设计

状态机是支付系统的核心设计之一，清晰的状态定义和转换规则能够保证系统的稳定性。

### 4.1 支付单状态机

#### 4.1.1 状态定义

```mermaid
stateDiagram-v2
    [*] --> PENDING: 创建支付单
    PENDING --> PAYING: 用户发起支付
    PAYING --> SUCCESS: 第三方回调成功
    PAYING --> FAILED: 第三方回调失败
    PAYING --> CANCELED: 用户取消支付
    SUCCESS --> REFUNDING: 用户申请退款
    REFUNDING --> REFUNDED: 退款成功
    FAILED --> [*]
    CANCELED --> [*]
    REFUNDED --> [*]
```

#### 4.1.2 状态转换表

| 当前状态 | 允许的下一状态 | 触发条件 | 备注 |
|---------|--------------|---------|-----|
| PENDING | PAYING | 用户发起支付 | - |
| PAYING | SUCCESS | 第三方回调成功 | - |
| PAYING | FAILED | 第三方回调失败 | 可重新发起支付 |
| PAYING | CANCELED | 用户取消支付 | 超时自动取消 |
| SUCCESS | REFUNDING | 用户申请退款 | - |
| REFUNDING | REFUNDED | 退款成功 | 支持部分退款 |

**非法状态转换示例**：

- PENDING → SUCCESS（跳过 PAYING 状态）
- FAILED → REFUNDING（失败的支付单不能退款）
- REFUNDED → SUCCESS（已退款不能恢复）

### 4.2 退款单状态机

#### 4.2.1 状态定义

```mermaid
stateDiagram-v2
    [*] --> PENDING: 创建退款单
    PENDING --> PROCESSING: 调用第三方退款
    PROCESSING --> SUCCESS: 退款成功
    PROCESSING --> FAILED: 退款失败
    SUCCESS --> [*]
    FAILED --> [*]
```

#### 4.2.2 与支付单状态联动

- **前置条件**：支付单必须是 `SUCCESS` 状态才能发起退款
- **状态同步**：退款成功后，支付单状态变为 `REFUNDED`
- **部分退款**：第一次退款成功后，支付单状态变为 `PARTIAL_REFUNDED`

### 4.3 状态机实现

#### 4.3.1 状态转换校验

```go
// 状态转换规则
var stateTransitionRules = map[string][]string{
    "PENDING":   {"PAYING"},
    "PAYING":    {"SUCCESS", "FAILED", "CANCELED"},
    "SUCCESS":   {"REFUNDING"},
    "REFUNDING": {"REFUNDED"},
}

// 校验状态转换是否合法
func isValidTransition(currentState, targetState string) bool {
    allowedStates, exists := stateTransitionRules[currentState]
    if !exists {
        return false
    }
    
    for _, state := range allowedStates {
        if state == targetState {
            return true
        }
    }
    return false
}
```

#### 4.3.2 状态转换（乐观锁）

```go
// 状态转换函数
func TransitState(paymentID int64, targetState string, version int) error {
    // 1. 查询当前状态
    current := queryPaymentOrder(paymentID)
    
    // 2. 校验状态转换是否合法
    if !isValidTransition(current.PaymentStatus, targetState) {
        return errors.New(fmt.Sprintf(
            "invalid state transition: %s -> %s",
            current.PaymentStatus, targetState))
    }
    
    // 3. 乐观锁更新
    affected := db.Exec(`
        UPDATE payment_order 
        SET payment_status = ?,
            updated_at = ?,
            version = version + 1
        WHERE payment_id = ? AND version = ?
    `, targetState, time.Now(), paymentID, version)
    
    if affected == 0 {
        return errors.New("concurrent update conflict, please retry")
    }
    
    // 4. 发布状态变更事件
    publishStateChangeEvent(paymentID, current.PaymentStatus, targetState)
    
    return nil
}
```

#### 4.3.3 状态变更事件

```go
// 状态变更事件
type PaymentStateChangeEvent struct {
    PaymentID    int64  `json:"payment_id"`
    OrderID      int64  `json:"order_id"`
    OldState     string `json:"old_state"`
    NewState     string `json:"new_state"`
    ChangeTime   time.Time `json:"change_time"`
}

// 发布事件到 Kafka
func publishStateChangeEvent(paymentID int64, oldState, newState string) {
    event := &PaymentStateChangeEvent{
        PaymentID:  paymentID,
        OrderID:    getOrderIDByPaymentID(paymentID),
        OldState:   oldState,
        NewState:   newState,
        ChangeTime: time.Now(),
    }
    
    kafka.Publish("payment_state_change", event)
}
```

**面试要点**：

1. **为什么需要状态机**：确保状态转换的合法性，防止业务逻辑错误
2. **如何保证并发安全**：乐观锁（version 字段）
3. **如何与其他系统协作**：通过 Kafka 事件驱动

## 五、高频考点深入

### 5.1 分布式事务

#### 5.1.1 问题场景

**面试问题**：支付成功后，如何保证订单状态同步更新？

这是一个典型的分布式事务问题：

- 支付服务：更新支付单状态为 `SUCCESS`
- 订单服务：更新订单状态为 `PAID`

两个操作在不同的服务和数据库中，无法使用传统的 ACID 事务保证一致性。

#### 5.1.2 Saga 模式

Saga 模式是微服务架构下常用的分布式事务解决方案，通过一系列本地事务和补偿机制实现最终一致性。

**Saga 流程图**：

```mermaid
sequenceDiagram
    participant 订单服务
    participant 支付服务
    participant 本地消息表
    participant Kafka

    订单服务->>订单服务: 1. 创建订单
    订单服务->>支付服务: 2. 请求支付
    支付服务->>支付服务: 3. 更新支付状态 SUCCESS
    支付服务->>本地消息表: 4. 插入"订单已支付"事件
    支付服务->>支付服务: 5. 提交本地事务
    本地消息表->>Kafka: 6. 定时任务扫描并发送
    Kafka->>订单服务: 7. 订单服务消费事件
    订单服务->>订单服务: 8. 更新订单状态 PAID
```

**本地消息表实现**：

```go
// 本地消息表结构
type LocalMessage struct {
    ID          int64     `db:"id"`
    OrderID     int64     `db:"order_id"`
    PaymentID   int64     `db:"payment_id"`
    Event       string    `db:"event"`
    Status      string    `db:"status"`  // PENDING, SENT, FAILED
    RetryCount  int       `db:"retry_count"`
    CreatedAt   time.Time `db:"created_at"`
}

// 支付成功后插入本地消息表
func OnPaymentSuccess(paymentID int64, orderID int64) error {
    // 开启本地事务
    tx, _ := db.Begin()
    
    // 1. 更新支付状态
    tx.Exec(`
        UPDATE payment_order 
        SET payment_status = 'SUCCESS'
        WHERE payment_id = ?
    `, paymentID)
    
    // 2. 插入本地消息表
    tx.Exec(`
        INSERT INTO local_message (order_id, payment_id, event, status, retry_count)
        VALUES (?, ?, 'ORDER_PAID', 'PENDING', 0)
    `, orderID, paymentID)
    
    // 3. 提交本地事务
    return tx.Commit()
}

// 定时任务扫描并发送消息
func ScanAndSendMessages() {
    messages := db.Query(`
        SELECT * FROM local_message 
        WHERE status = 'PENDING' AND retry_count < 3
        ORDER BY created_at ASC
        LIMIT 100
    `)
    
    for _, msg := range messages {
        // 发送到 Kafka
        if err := kafka.Publish("order_paid", msg); err == nil {
            db.Exec(`
                UPDATE local_message 
                SET status = 'SENT' 
                WHERE id = ?
            `, msg.ID)
        } else {
            db.Exec(`
                UPDATE local_message 
                SET retry_count = retry_count + 1 
                WHERE id = ?
            `, msg.ID)
        }
    }
}
```

**补偿机制**：

如果支付失败，订单服务需要取消订单：

```go
// 支付失败补偿
func OnPaymentFailed(paymentID int64, orderID int64) error {
    // 1. 更新支付状态
    db.Exec(`UPDATE payment_order SET payment_status = 'FAILED' WHERE payment_id = ?`, paymentID)
    
    // 2. 发布支付失败事件
    kafka.Publish("payment_failed", &PaymentFailedEvent{
        PaymentID: paymentID,
        OrderID:   orderID,
    })
    
    // 订单服务消费事件并取消订单
    return nil
}
```

#### 5.1.3 TCC 模式

TCC（Try-Confirm-Cancel）模式适用于对一致性要求较高的场景，如退款流程。

**TCC 三阶段**：

```mermaid
sequenceDiagram
    participant 订单服务
    participant 支付服务
    participant 账户系统

    订单服务->>账户系统: Try: 冻结账户余额
    订单服务->>支付服务: Try: 创建退款单
    
    alt 全部成功
        订单服务->>账户系统: Confirm: 解冻并扣减
        订单服务->>支付服务: Confirm: 退款成功
    else 任一失败
        订单服务->>账户系统: Cancel: 直接解冻
        订单服务->>支付服务: Cancel: 取消退款
    end
```

**TCC 实现示例**：

```go
// Try 阶段：冻结账户余额
func TryFreezeBalance(userID int64, amount decimal.Decimal) (string, error) {
    // 生成冻结记录 ID
    freezeID := uuid.New().String()
    
    // 冻结余额
    affected := db.Exec(`
        UPDATE account 
        SET balance = balance - ?,
            frozen_balance = frozen_balance + ?
        WHERE user_id = ? AND balance >= ?
    `, amount, amount, userID, amount)
    
    if affected == 0 {
        return "", errors.New("insufficient balance")
    }
    
    // 记录冻结记录
    db.Exec(`
        INSERT INTO account_freeze (freeze_id, user_id, amount, status)
        VALUES (?, ?, ?, 'FROZEN')
    `, freezeID, userID, amount)
    
    return freezeID, nil
}

// Confirm 阶段：解冻并扣减
func ConfirmFreezeBalance(freezeID string) error {
    // 查询冻结记录
    freeze := db.QueryRow(`
        SELECT user_id, amount FROM account_freeze 
        WHERE freeze_id = ? AND status = 'FROZEN'
    `, freezeID)
    
    // 更新冻结记录状态
    db.Exec(`
        UPDATE account_freeze 
        SET status = 'CONFIRMED' 
        WHERE freeze_id = ?
    `, freezeID)
    
    // 扣减冻结余额
    db.Exec(`
        UPDATE account 
        SET frozen_balance = frozen_balance - ?
        WHERE user_id = ?
    `, freeze.Amount, freeze.UserID)
    
    return nil
}

// Cancel 阶段：直接解冻
func CancelFreezeBalance(freezeID string) error {
    // 查询冻结记录
    freeze := db.QueryRow(`
        SELECT user_id, amount FROM account_freeze 
        WHERE freeze_id = ? AND status = 'FROZEN'
    `, freezeID)
    
    // 解冻余额
    db.Exec(`
        UPDATE account 
        SET balance = balance + ?,
            frozen_balance = frozen_balance - ?
        WHERE user_id = ?
    `, freeze.Amount, freeze.Amount, freeze.UserID)
    
    // 更新冻结记录状态
    db.Exec(`
        UPDATE account_freeze 
        SET status = 'CANCELED' 
        WHERE freeze_id = ?
    `, freezeID)
    
    return nil
}
```

#### 5.1.4 Saga vs TCC

| 维度 | Saga | TCC |
|-----|------|-----|
| **一致性** | 最终一致性 | 强一致性 |
| **实现复杂度** | 低 | 高 |
| **性能** | 高（异步） | 中（同步） |
| **适用场景** | 支付流程 | 退款流程 |
| **回滚方式** | 补偿操作 | Cancel 操作 |

**面试建议**：

- Saga 适合长事务、跨服务场景
- TCC 适合短事务、强一致性场景

### 5.2 幂等性设计

#### 5.2.1 为什么需要幂等

- **第三方回调重复**：网络抖动导致重复回调
- **用户重复点击**：前端未防抖，用户多次点击
- **系统重试**：超时重试导致重复请求

#### 5.2.2 三个关键场景

**场景 1: 支付单创建幂等**

```go
// 幂等键：order_id + user_id
func CreatePaymentOrder(orderID int64, userID int64, amount decimal.Decimal) (*PaymentOrder, error) {
    idempotencyKey := fmt.Sprintf("payment_%d_%d", orderID, userID)
    
    // Redis 分布式锁（防并发）
    lock := redis.Lock(idempotencyKey, 10*time.Second)
    if !lock.TryLock() {
        return nil, errors.New("concurrent request")
    }
    defer lock.Unlock()
    
    // 检查是否已存在
    existing := db.QueryRow(`
        SELECT * FROM payment_order 
        WHERE order_id = ? AND user_id = ?
    `, orderID, userID)
    
    if existing != nil {
        return existing, nil  // 幂等返回
    }
    
    // 创建支付单（数据库唯一索引保证幂等）
    payment := &PaymentOrder{
        PaymentID: snowflake.Generate(),
        OrderID:   orderID,
        UserID:    userID,
        Amount:    amount,
    }
    
    if err := db.Insert(payment); err != nil {
        if isDuplicateKeyError(err) {
            return db.QueryRow(`SELECT * FROM payment_order WHERE order_id = ? AND user_id = ?`, orderID, userID), nil
        }
        return nil, err
    }
    
    return payment, nil
}
```

**场景 2: 支付回调幂等**

```go
// 幂等键：channel_trade_no（第三方交易号）
func HandlePaymentCallback(channelTradeNo string, status string) error {
    // 检查是否已处理
    existing := db.QueryRow(`
        SELECT * FROM payment_order 
        WHERE channel_trade_no = ?
    `, channelTradeNo)
    
    if existing != nil && existing.PaymentStatus == "SUCCESS" {
        return nil  // 已处理，直接返回
    }
    
    // 更新支付状态
    db.Exec(`
        UPDATE payment_order 
        SET payment_status = ?, channel_trade_no = ?
        WHERE payment_id = ?
    `, status, channelTradeNo, existing.PaymentID)
    
    return nil
}
```

**场景 3: 退款幂等**

```go
// 幂等键：refund_id
func CreateRefundOrder(paymentID int64, amount decimal.Decimal) (*RefundOrder, error) {
    refundID := snowflake.Generate()
    
    // 检查是否已存在
    existing := db.QueryRow(`
        SELECT * FROM refund_order 
        WHERE payment_id = ? AND amount = ?
    `, paymentID, amount)
    
    if existing != nil {
        return existing, nil  // 幂等返回
    }
    
    // 创建退款单
    refund := &RefundOrder{
        RefundID:  refundID,
        PaymentID: paymentID,
        Amount:    amount,
        Status:    "PENDING",
    }
    
    db.Insert(refund)
    return refund, nil
}
```

#### 5.2.3 实现手段总结

| 手段 | 使用场景 | 优点 | 缺点 |
|-----|---------|------|-----|
| **Redis 分布式锁** | 高并发场景 | 性能好，防并发 | 需要考虑锁超时 |
| **数据库唯一索引** | 防重复插入 | 可靠，数据库保证 | 性能略低 |
| **乐观锁** | 防并发更新 | 无锁开销 | 需要重试 |

### 5.3 一致性保证

#### 5.3.1 账户余额一致性

**问题**：Redis 缓存和 MySQL 数据不一致

**方案**：Redis + MySQL 双写 + Lua 脚本 + 定时对账

```lua
-- Redis Lua 脚本（原子操作）
local balance_key = KEYS[1]
local amount = tonumber(ARGV[1])

local balance = tonumber(redis.call('GET', balance_key) or '0')

if balance >= amount then
    redis.call('DECRBY', balance_key, amount)
    return 1
else
    return 0
end
```

```go
// 扣减余额
func DeductBalance(userID int64, amount decimal.Decimal) error {
    // 1. Redis 扣减（Lua 脚本保证原子性）
    result := redis.Eval(luaScript, []string{fmt.Sprintf("balance:%d", userID)}, amount.String())
    if result == 0 {
        return errors.New("insufficient balance")
    }
    
    // 2. MySQL 扣减
    affected := db.Exec(`
        UPDATE account 
        SET balance = balance - ?
        WHERE user_id = ? AND balance >= ?
    `, amount, userID, amount)
    
    if affected == 0 {
        // 回滚 Redis
        redis.IncrBy(fmt.Sprintf("balance:%d", userID), amount.IntPart())
        return errors.New("insufficient balance")
    }
    
    // 3. 记录流水
    db.Exec(`
        INSERT INTO account_transaction (user_id, amount, type, created_at)
        VALUES (?, ?, 'DEDUCT', NOW())
    `, userID, amount)
    
    return nil
}

// 定时对账任务
func ReconcileAccountBalance() {
    users := db.Query("SELECT user_id FROM account")
    
    for _, user := range users {
        // 查询 MySQL 余额
        mysqlBalance := db.QueryRow(`SELECT balance FROM account WHERE user_id = ?`, user.ID)
        
        // 查询 Redis 余额
        redisBalance := redis.Get(fmt.Sprintf("balance:%d", user.ID))
        
        // 对比
        if mysqlBalance != redisBalance {
            log.Error("balance mismatch", "user_id", user.ID, "mysql", mysqlBalance, "redis", redisBalance)
            
            // 以 MySQL 为准，修复 Redis
            redis.Set(fmt.Sprintf("balance:%d", user.ID), mysqlBalance)
        }
    }
}
```

#### 5.3.2 支付流水一致性

**问题**：支付单金额与流水表汇总金额不一致

**方案**：

1. 每笔支付/退款都记录流水
2. 流水表不可更新，只能插入
3. 定时任务校验：支付单金额 = 流水表汇总金额

```go
// 定时对账任务
func ReconcilePaymentTransaction() {
    payments := db.Query("SELECT payment_id, payment_amount FROM payment_order WHERE payment_status = 'SUCCESS'")
    
    for _, payment := range payments {
        // 查询流水表汇总金额
        transactionAmount := db.QueryRow(`
            SELECT SUM(amount) FROM payment_transaction 
            WHERE payment_id = ?
        `, payment.PaymentID)
        
        // 对比
        if payment.PaymentAmount != transactionAmount {
            log.Error("amount mismatch", "payment_id", payment.PaymentID, 
                      "payment", payment.PaymentAmount, "transaction", transactionAmount)
            
            // 人工介入
            notifyAdmin(payment.PaymentID)
        }
    }
}
```

**面试要点**：

1. **一致性方案**: Redis + MySQL 双写 + 定时对账
2. **原子性保证**: Lua 脚本
3. **数据源优先级**: MySQL 为准，Redis 为辅

## 六、扩展模块

本章简要介绍支付系统的扩展模块，这些模块在面试中通常不是重点，但了解基本概念有助于建立完整的知识体系。

### 6.1 清结算系统

#### 6.1.1 核心概念

清结算系统负责将支付金额按规则分配给平台、商家、营销补贴等各方。

**分账规则示例**：

假设一笔订单实付 100 元，分账如下：

| 角色 | 比例 | 金额 |
|-----|------|-----|
| 平台佣金 | 5% | 5 元 |
| 商家收益 | 95% | 95 元 |
| 营销补贴 | - | 由平台承担 |

#### 6.1.2 结算周期

- **T+1**：次日结算（快速到账，适合小商家）
- **T+7**：7 日后结算（标准周期）
- **账期结算**：按月结算（企业客户）

#### 6.1.3 提现管理

```go
// 提现限额控制
func WithdrawRequest(merchantID int64, amount decimal.Decimal) error {
    // 1. 单笔限额检查
    if amount.GreaterThan(decimal.NewFromInt(50000)) {
        return errors.New("single withdraw limit exceeded: 50,000")
    }
    
    // 2. 每日限额检查
    todayWithdrawn := db.QueryRow(`
        SELECT SUM(amount) FROM withdraw_order 
        WHERE merchant_id = ? AND DATE(created_at) = CURDATE()
    `, merchantID)
    
    if todayWithdrawn.Add(amount).GreaterThan(decimal.NewFromInt(200000)) {
        return errors.New("daily withdraw limit exceeded: 200,000")
    }
    
    // 3. 风控校验
    if isRiskMerchant(merchantID) {
        return errors.New("risk merchant, withdraw suspended")
    }
    
    // 4. 创建提现单
    withdraw := &WithdrawOrder{
        WithdrawID: snowflake.Generate(),
        MerchantID: merchantID,
        Amount:     amount,
        Status:     "PENDING",
    }
    
    db.Insert(withdraw)
    return nil
}
```

#### 6.1.4 清结算流程

```mermaid
graph LR
    A[T 日交易] --> B[T+1 日凌晨定时任务]
    B --> C[计算分账金额]
    C --> D[生成结算单]
    D --> E[商家确认]
    E --> F[提现申请]
    F --> G[打款到银行账户]
    G --> H[更新账户余额]
```

### 6.2 对账系统

#### 6.2.1 为什么需要对账

系统与第三方的交易数据可能存在以下问题：

- **长款**：第三方有，本地无（用户已支付，但系统未记录）
- **短款**：本地有，第三方无（系统记录已支付，但第三方未收到）
- **金额不符**：订单号一致，但金额不一致

#### 6.2.2 对账维度

| 对账维度 | 对账内容 | 数据来源 |
|---------|---------|---------|
| **交易对账** | 订单号、金额、状态 | 第三方对账文件 vs 本地支付流水 |
| **资金对账** | 入账金额、手续费 | 第三方结算单 vs 本地账户流水 |

#### 6.2.3 对账流程

```mermaid
graph LR
    A[每日凌晨 2:00] --> B[拉取第三方对账文件]
    B --> C[解析对账文件]
    C --> D[与本地流水对比]
    D --> E[生成差错报告]
    E --> F{有差错?}
    F -->|是| G[人工复核]
    F -->|否| H[记录对账结果]
    G --> I[差错处理]
```

#### 6.2.4 差错处理

```go
// 对账差错分类
type ReconciliationError struct {
    Type          string // LONG_PAYMENT, SHORT_PAYMENT, AMOUNT_MISMATCH
    OrderID       int64
    LocalAmount   decimal.Decimal
    RemoteAmount  decimal.Decimal
    Description   string
}

// 对账任务
func ReconcileTransactions(date time.Time) ([]*ReconciliationError, error) {
    var errors []*ReconciliationError
    
    // 1. 拉取第三方对账文件
    remoteRecords := fetchRemoteReconciliationFile(date)
    
    // 2. 查询本地流水
    localRecords := db.Query(`
        SELECT order_id, amount FROM payment_transaction 
        WHERE DATE(created_at) = ?
    `, date)
    
    // 3. 对比（本地有，第三方无 - 短款）
    for _, local := range localRecords {
        if !existsInRemote(local.OrderID, remoteRecords) {
            errors = append(errors, &ReconciliationError{
                Type:        "SHORT_PAYMENT",
                OrderID:     local.OrderID,
                LocalAmount: local.Amount,
                Description: "本地有记录，第三方无",
            })
        }
    }
    
    // 4. 对比（第三方有，本地无 - 长款）
    for _, remote := range remoteRecords {
        if !existsInLocal(remote.OrderID, localRecords) {
            errors = append(errors, &ReconciliationError{
                Type:         "LONG_PAYMENT",
                OrderID:      remote.OrderID,
                RemoteAmount: remote.Amount,
                Description:  "第三方有记录，本地无",
            })
        }
    }
    
    // 5. 对比（金额不符）
    for _, local := range localRecords {
        remote := findRemoteRecord(local.OrderID, remoteRecords)
        if remote != nil && !local.Amount.Equal(remote.Amount) {
            errors = append(errors, &ReconciliationError{
                Type:         "AMOUNT_MISMATCH",
                OrderID:      local.OrderID,
                LocalAmount:  local.Amount,
                RemoteAmount: remote.Amount,
                Description:  "金额不一致",
            })
        }
    }
    
    return errors, nil
}
```

### 6.3 风控系统

#### 6.3.1 三个阶段

| 阶段 | 风控手段 | 示例 |
|-----|---------|-----|
| **事前风控** | 支付前验证 | 支付密码、指纹、人脸识别 |
| **事中风控** | 交易实时监控 | 短时间大额、异地登录、限额控制 |
| **事后风控** | 对账差错分析 | 资金流向追踪、异常模式识别 |

#### 6.3.2 常见风控规则

```go
// 风控规则配置
var riskRules = []RiskRule{
    {
        Name:        "单笔大额",
        Condition:   "amount > 10000",
        Action:      "二次验证",
        Description: "单笔支付超过 10,000 元需要二次验证",
    },
    {
        Name:        "短时间高频",
        Condition:   "count_in_1h > 5",
        Action:      "触发风控审核",
        Description: "1 小时内支付超过 5 次",
    },
    {
        Name:        "异地登录",
        Condition:   "ip_city_change",
        Action:      "短信验证码",
        Description: "IP 地址城市变更",
    },
}

// 风控检查
func RiskCheck(userID int64, amount decimal.Decimal) error {
    // 1. 单笔大额检查
    if amount.GreaterThan(decimal.NewFromInt(10000)) {
        return errors.New("need second verification for large amount")
    }
    
    // 2. 短时间高频检查
    count := redis.Incr(fmt.Sprintf("payment_count:%d", userID))
    redis.Expire(fmt.Sprintf("payment_count:%d", userID), 1*time.Hour)
    
    if count > 5 {
        return errors.New("too many payments in 1 hour")
    }
    
    // 3. 异地登录检查
    lastCity := redis.Get(fmt.Sprintf("last_city:%d", userID))
    currentCity := getIPCity()
    
    if lastCity != "" && lastCity != currentCity {
        return errors.New("city changed, need SMS verification")
    }
    
    redis.Set(fmt.Sprintf("last_city:%d", userID), currentCity)
    
    return nil
}
```

#### 6.3.3 黑名单机制

```go
// 黑名单检查
func IsBlacklisted(userID int64) bool {
    return redis.SIsMember("user_blacklist", userID)
}

// 加入黑名单
func AddToBlacklist(userID int64, reason string) {
    redis.SAdd("user_blacklist", userID)
    
    // 记录黑名单原因
    db.Exec(`
        INSERT INTO blacklist_record (user_id, reason, created_at)
        VALUES (?, ?, NOW())
    `, userID, reason)
}
```

**面试要点**：

- 清结算：分账规则、结算周期、提现限额
- 对账：长款、短款、差错处理
- 风控：事前、事中、事后三阶段

## 七、面试问答锦囊

本章总结支付系统面试中最常被问到的问题，并给出简洁的回答要点，便于考前快速复习。

### 7.1 架构设计类

#### Q1: 如何设计一个支付系统？⭐⭐⭐

**回答框架**（3 分钟讲清楚）：

1. **分层架构**：
   - 接入层（C 端/B 端/Open API）
   - 应用服务层（订单、支付、对账服务）
   - 核心业务层（账户、支付网关、清结算、风控）
   - 基础设施层（MySQL、Redis、Kafka、XXL-Job）
   - 第三方服务层（微信支付、支付宝、银行网关）

2. **核心子系统**：
   - 账户系统：余额管理、冻结/解冻
   - 支付网关：渠道抽象、路由策略
   - 清结算引擎：分账计算、结算管理
   - 风控系统：支付验证、异常监控

3. **关键技术**：
   - 分布式事务：Saga/TCC
   - 幂等性设计：分布式锁 + 唯一索引
   - 状态机：支付单/退款单状态流转

**白板画图要点**：

```
┌─────────────────────────────────┐
│    接入层（C/B 端）              │
├─────────────────────────────────┤
│    应用层（订单/支付/对账）      │
├─────────────────────────────────┤
│    核心层（账户/网关/清结算）    │
├─────────────────────────────────┤
│    基础设施（MySQL/Redis/Kafka） │
├─────────────────────────────────┤
│    第三方（微信/支付宝）         │
└─────────────────────────────────┘
```

**详见**：第 2 章

---

#### Q2: 支付系统如何保证高可用？⭐⭐

**回答要点**（5 个维度）：

1. **多机房部署**：异地容灾，主备切换
2. **Redis 主从 + 哨兵**：缓存高可用，故障自动切换
3. **Kafka 集群**：消息高可用，分区副本机制
4. **限流降级**：大促期间保护核心服务
5. **熔断机制**：第三方支付故障时切换备用渠道

**详见**：第 1.3 章

---

### 7.2 技术方案类

#### Q3: 支付成功后如何保证订单状态同步更新？⭐⭐⭐

**回答要点**：

1. **Saga 模式** + **本地消息表** + **最终一致性**

2. **实现步骤**：
   - 支付服务更新支付状态 + 插入本地消息表（同一事务）
   - 定时任务扫描本地消息表，发送到 Kafka
   - 订单服务消费 Kafka 消息，更新订单状态

3. **补偿机制**：
   - 支付失败 → 订单服务取消订单

**详见**：第 5.1.2 章

---

#### Q4: 如何保证支付幂等性？⭐⭐⭐

**回答要点**（三个场景）：

| 场景 | 幂等键 | 实现手段 |
|-----|-------|---------|
| 支付单创建 | order_id + user_id | Redis 分布式锁 + DB 唯一索引 |
| 支付回调 | channel_trade_no | DB 唯一索引 |
| 退款 | refund_id | DB 唯一索引 |

**核心思想**：

1. **防并发**：Redis 分布式锁
2. **防重复**：数据库唯一索引
3. **防并发更新**：乐观锁（version 字段）

**详见**：第 5.2 章

---

#### Q5: 第三方支付回调失败怎么办？⭐⭐

**回答要点**（三层保障）：

1. **主动重试**：最多 3 次，指数退避（1s, 2s, 4s）
2. **定时补偿**：每分钟扫描超时支付单，主动查询第三方状态
3. **人工介入**：超过重试次数，进入人工复核队列

**详见**：第 3.3 章

---

#### Q6: 如何处理部分退款？⭐⭐

**回答要点**：

1. **可退金额计算**：实付金额 - 已退金额
2. **多次退款支持**：累计退款金额 ≤ 实付金额
3. **营销优惠处理**：按比例退款

**示例**：
- 实付 100 元（原价 120，优惠 20）
- 退款 50 元 → 退实付 41.67，营销优惠退 8.33

**详见**：第 3.2.2 章

---

### 7.3 场景题类

#### Q7: 大促期间支付峰值如何应对？⭐⭐

**回答要点**（4 个维度）：

| 维度 | 方案 | 说明 |
|-----|------|-----|
| **缓存** | Redis 缓存热点数据 | 账户余额、支付单状态 |
| **限流** | 令牌桶限流 | 单用户 QPS 限制 |
| **降级** | 关闭非核心功能 | 账单查询、历史记录 |
| **异步化** | Kafka 异步处理 | 支付回调、状态同步 |

**详见**：第 1.3 章

---

#### Q8: 如何防止恶意刷单？⭐

**回答要点**（风控规则）：

| 规则 | 阈值 | 动作 |
|-----|------|-----|
| 短时间大额 | 单笔 > 10,000 元 | 二次验证 |
| 短时间高频 | 1 小时内 > 5 笔 | 触发风控审核 |
| 异地登录 | IP 城市变更 | 短信验证码 |
| 限额控制 | 单日 > 200,000 元 | 暂停支付 |

**详见**：第 6.3.2 章

---

### 7.4 面试技巧

**1. 架构设计题**

- 快速画出分层架构图（30 秒）
- 说明核心子系统职责（1 分钟）
- 讲解关键技术方案（2 分钟）

**2. 技术方案题**

- 先说核心思想（10 秒）
- 再讲实现步骤（1 分钟）
- 最后补充边界情况（30 秒）

**3. 场景题**

- 先识别核心问题（10 秒）
- 提出多个解决方案（1 分钟）
- 对比优劣并推荐（30 秒）

**4. 常见追问**

- "如果第三方支付挂了怎么办？" → 渠道降级、备用渠道
- "如何保证资金安全？" → 分布式事务、幂等性、审计日志
- "支付系统的瓶颈在哪？" → 数据库写入、第三方支付 QPS

---

## 总结

本文从系统设计面试的角度，深入解析了电商支付系统的核心知识点：

1. **整体架构**：分层架构、核心子系统、设计原则
2. **核心流程**：支付流程、退款流程、异步回调
3. **状态机设计**：支付单/退款单状态流转、状态转换规则
4. **高频考点**：
   - 分布式事务（Saga/TCC）
   - 幂等性设计（三个场景）
   - 一致性保证（账户余额、支付流水）
5. **扩展模块**：清结算、对账、风控
6. **面试锦囊**：8 个高频面试题及回答要点

**面试建议**：

- 熟练掌握分层架构图，能够在白板上快速画出
- 深入理解分布式事务（Saga/TCC）的实现原理
- 掌握幂等性设计的三个关键场景
- 了解清结算、对账、风控的基本概念

**参考资料**：

- [26-电商订单系统](./26-ecommerce-order-system.md)
- [22-电商库存系统](./22-ecommerce-inventory.md)
- [28-电商营销系统](./28-ecommerce-marketing-system.md)

---

**全文完**
