---
title: 电商系统设计（五）：计价系统 DDD 实践
date: 2026-03-14
categories:
- 系统设计
tags:
- e-commerce
- system-design
- ddd
- hexagonal-architecture
- aggregate-root
toc: true
---

<!-- toc -->

> **电商系统设计系列**
> - [（一）全景概览与领域划分](/system-design/20-ecommerce-overview/)
> - [（二）商品上架系统](/system-design/21-ecommerce-listing/)
> - [（三）库存系统](/system-design/22-ecommerce-inventory/)
> - [（四）计价引擎](/system-design/23-ecommerce-pricing-engine/)
> - **（五）计价系统 DDD 实践**（本文）
> - [（六）B 端运营系统](/system-design/25-ecommerce-b-side-ops/)

本文是电商系统设计系列的第五篇，是[（四）计价引擎](/system-design/23-ecommerce-pricing-engine/)的姊妹篇，从 DDD 视角重新审视计价系统的建模。

> 本文是计价引擎系列的方法论篇，聚焦 DDD 在计价系统中的战略/战术设计实践。系统架构与实现细节详见：[电商系统价格计算引擎设计与实现](./24-pricing-engine-design.md)。

## 一、背景与挑战

在构建电商计价系统的过程中，价格计算并非简单的"标价"，而是由基础价格、营销折扣、平台费用、用户抵扣等多层因素叠加而成。随着业务规模扩大，我们面临着三大核心挑战：

**1. 隐晦性（Obscurity）**

- **抽象层面的隐晦**：同一个"价格"概念，在不同场景下含义不同
  - 商品详情页展示价：用户看到的价格
  - 订单价：创建订单时的价格快照
  - 支付价：最终扣款价格
  
- **实现层面的隐晦**：代码中的术语混乱
  - 有人叫`originalPrice`，有人叫`marketPrice`
  - 有人叫`salePrice`，有人叫`discountPrice`
  - 业务人员和技术人员理解不一致

**2. 耦合性（Coupling）**

- **代码层面**：价格计算逻辑散落在各处
  - 商品服务有一套计算
  - 订单服务又重复计算
  - 支付服务再计算一次
  
- **模块层面**：计价依赖多个外部服务
  - 促销服务（获取活动信息）
  - 商品服务（获取基础价格）
  - 用户服务（判断用户类型）
  
- **系统层面**：前后端价格不一致导致资损

**3. 变化性（Variability）**

- **业务需求频繁变化**
  - 促销规则每周调整（双11期间优先级变化）
  - 新增促销类型（买赠、满减、阶梯价）
  - 不同地区有不同定价策略
  
- **品类扩展需求**
  - 实物商品、虚拟商品、服务类商品
  - 每种品类有特殊的计价规则

### 1.3 初期设计的问题

最初的实现方式是**面向过程的事务脚本**：

```go
// ❌ 问题代码示例
func CalculatePrice(itemID, userID int64, quantity int) (int64, error) {
    // 1. 获取商品基础信息
    item := getItemFromDB(itemID)
    basePrice := item.Price
    
    // 2. 检查是否有秒杀
    if flashSale := getFlashSale(itemID); flashSale != nil {
        basePrice = flashSale.Price
    }
    
    // 3. 检查新用户
    if isNewUser(userID) {
        if newUserPrice := getNewUserPrice(itemID); newUserPrice < basePrice {
            basePrice = newUserPrice
        }
    }
    
    // 4. 计算数量价格
    totalPrice := basePrice * quantity
    
    // 5. 加上服务费
    if fee := getAdminFee(itemID); fee > 0 {
        totalPrice += fee
    }
    
    // 6. 减去优惠券
    if voucher := getUserVoucher(userID); voucher != nil {
        totalPrice -= voucher.Amount
    }
    
    return totalPrice, nil
}
```

**核心问题**：
- ❌ 业务逻辑分散在各个函数中，难以理解整体流程
- ❌ 缺乏业务概念的抽象，只有数据获取和计算
- ❌ 新增促销类型需要修改核心计算逻辑
- ❌ 无法支持复杂的业务规则（如买N件享M折）
- ❌ 测试困难，需要Mock大量外部依赖

---

## 二、DDD核心概念

在介绍计价系统的DDD实践之前，先回顾一下DDD的核心概念。

### 2.1 什么是领域？

领域由三部分组成：

```
┌─────────────────────────────────────────┐
│              领域（Domain）              │
├─────────────────────────────────────────┤
│                                          │
│  涉众域 (Stakeholders)                   │
│    └─ 用户：商家、运营、消费者、财务     │
│                                          │
│  问题域 (Problem Space)                  │
│    └─ 业务价值：如何定价？如何促销？     │
│                                          │
│  解决方案域 (Solution Space)             │
│    └─ 解决方案：四层计价模型             │
│                                          │
└─────────────────────────────────────────┘
```

**计价领域示例**：
- **涉众域**：商家（定价）、运营（促销）、消费者（购买）、财务（结算）
- **问题域**：如何准确计算价格？如何支持多种促销？如何保证一致性？
- **解决方案域**：统一的计价模型、规则引擎、价格快照

### 2.2 什么是领域驱动设计？

> 针对特定业务领域，用户在面对业务问题时有对应的解决方案，这些问题与方案构成了领域知识。领域驱动设计就是围绕这些知识来设计系统。

**计价领域知识**：
- **流程**：商品展示 → 加入购物车 → 创建订单 → 支付结算
- **规则**：促销优先级、费用计算规则、优惠抵扣规则
- **方法**：四层计价模型、价格快照机制

---

## 三、战略设计实践

### 3.1 确定用例

我们使用**用例图**来表达用户与系统的交互：

```
┌─────────────────────────────────────────────────────┐
│              计价系统用例图                          │
├─────────────────────────────────────────────────────┤
│                                                      │
│  商家 (Merchant)                                     │
│    ├─ 设置商品价格                                   │
│    ├─ 配置促销活动                                   │
│    └─ 查看销售数据                                   │
│                                                      │
│  运营 (Operator)                                     │
│    ├─ 创建营销活动                                   │
│    ├─ 配置优惠券                                     │
│    └─ 调整费用规则                                   │
│                                                      │
│  消费者 (Customer)                                   │
│    ├─ 查看商品价格                                   │
│    ├─ 创建订单                                       │
│    └─ 支付结算                                       │
│                                                      │
│  财务 (Finance)                                      │
│    ├─ 查看结算明细                                   │
│    └─ 对账                                          │
│                                                      │
└─────────────────────────────────────────────────────┘
```

### 3.2 统一语言（Ubiquitous Language）

从用例中抽取概念，建立**统一语言**。这是DDD最关键的一步。

#### 基础价格术语

| 中文术语 | 英文术语 | Term | 含义 |
|---------|---------|------|------|
| 市场原价 | Market Price | `market_price` | 商品的市场标价，来自供应商 |
| 折扣价 | Discount Price | `discount_price` | 平台日常销售价 |
| 划线价 | Listed Price | `listed_price` | 用于展示的对比价格 |

#### 促销术语

| 中文术语 | 英文术语 | Term | 含义 |
|---------|---------|------|------|
| 促销价 | Promotion Price | `promotion_price` | 参与促销活动后的价格 |
| 秒杀价 | Flash Sale Price | `flash_sale_price` | 限时秒杀活动价格 |
| 新用户价 | New User Price | `new_user_price` | 新用户专享价格 |
| 满减价 | Threshold Price | `threshold_price` | 满XX减XX后的价格 |

#### 费用术语

| 中文术语 | 英文术语 | Term | 含义 |
|---------|---------|------|------|
| 平台服务费 | Platform Fee | `platform_fee` | 平台收取的服务费 |
| 配送费 | Delivery Fee | `delivery_fee` | 物流配送费用 |
| 手续费 | Handling Fee | `handling_fee` | 支付渠道手续费 |

#### 最终价格术语

| 中文术语 | 英文术语 | Term | 含义 |
|---------|---------|------|------|
| 计价金额 | Pricing Amount | `pricing_amount` | 单个SKU的计价金额 |
| 最终价格 | Final Price | `final_price` | 用户最终支付价格 |
| 结算金额 | Settlement Amount | `settlement_amount` | 商家结算金额 |

**统一语言的重要性**：

```
案例：某电商平台的混乱

改造前：
- 技术团队：MarketPrice、DiscountPrice、ActualPrice
- 产品团队：原价、活动价、实付价
- 运营团队：建议零售价、会员价、到手价
- 客服团队：标价、优惠后价格、支付价

结果：
- 沟通成本高（每次对话都要先对齐概念）
- 需求理解错误（产品要改"活动价"，技术改了"优惠后价格"）
- 价格bug频发（前端显示"实付价"，后端计算的是"到手价"）

改造后：
- 所有团队统一使用：市场原价、折扣价、促销价、最终价格
- 文档、代码、会议统一使用这些术语
- 新人一周即可理解价格体系
```

### 3.3 概念模型（Concept Model）

基于统一语言，建立**概念模型**，明确概念之间的关系：

```
┌───────────────────────────────────────────────────────────────┐
│                     计价领域概念模型                           │
└───────────────────────────────────────────────────────────────┘

                    ┌─────────────────┐
                    │   PriceEntity   │
                    │   (价格实体)     │
                    └────────┬────────┘
                             │
                    ┌────────┴────────┐
                    │                 │
         ┌──────────▼──────────┐  ┌──▼──────────────┐
         │    BasePrice        │  │   Promotion     │
         │    (基础价格)        │  │   (促销)        │
         │  • MarketPrice      │  │  • FlashSale    │
         │  • DiscountPrice    │  │  • NewUserPrice │
         │  • ListedPrice      │  │  • ThresholdDiscount │
         └─────────────────────┘  └─────────────────┘
                    │
         ┌──────────┴──────────┐
         │                     │
    ┌────▼────────┐      ┌────▼──────────┐
    │    Fee      │      │   Discount    │
    │   (费用)     │      │   (优惠)      │
    │ • Platform  │      │ • Voucher     │
    │ • Delivery  │      │ • Points      │
    │ • Handling  │      │ • Payment     │
    └─────────────┘      └───────────────┘
         │                     │
         └──────────┬──────────┘
                    │
              ┌─────▼─────┐
              │FinalPrice │
              │ (最终价格) │
              └───────────┘
```

**概念关系说明**：
- **PriceEntity** 包含 (1:1) **BasePrice** - 每个商品有且只有一个基础价格
- **PriceEntity** 可能有 (1:0..N) **Promotion** - 可以参与多个促销（但只能选一个）
- **Order** 可能有 (1:0..N) **Fee** - 订单可能产生多种费用
- **Order** 可能有 (1:0..N) **Discount** - 订单可能使用多种优惠
- **Order** 产生 (1:1) **FinalPrice** - 最终计算出唯一的支付价格

### 3.4 子域划分（Subdomain）

将复杂问题拆解为多个简单问题，我们基于**问题域**进行拆分：

```
┌─────────────────────────────────────────────────────────┐
│              计价领域 (Pricing Domain)                   │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  核心子域 (Core Subdomain)                               │
│  ┌───────────────────────────────────────────────┐     │
│  │  定价子域 (Pricing Subdomain)                  │     │
│  │  • 问题：如何准确计算价格？                    │     │
│  │  • 方案：四层计价模型                          │     │
│  │  • 职责：价格计算、校验、快照                  │     │
│  └───────────────────────────────────────────────┘     │
│                                                          │
│  支撑子域 (Supporting Subdomain)                         │
│  ┌──────────────────┐  ┌──────────────────┐           │
│  │  促销子域         │  │  商品子域         │           │
│  │  (Promotion)     │  │  (Item)          │           │
│  │  • 促销规则管理   │  │  • 商品信息      │           │
│  │  • 活动配置      │  │  • 库存管理      │           │
│  └──────────────────┘  └──────────────────┘           │
│                                                          │
│  ┌──────────────────┐  ┌──────────────────┐           │
│  │  用户子域         │  │  支付子域         │           │
│  │  (User)          │  │  (Payment)       │           │
│  │  • 用户信息      │  │  • 支付方式      │           │
│  │  • 用户分群      │  │  • 优惠券        │           │
│  └──────────────────┘  └──────────────────┘           │
│                                                          │
│  通用子域 (Generic Subdomain)                            │
│  ┌──────────────────┐  ┌──────────────────┐           │
│  │  缓存子域         │  │  配置子域         │           │
│  │  (Cache)         │  │  (Config)        │           │
│  └──────────────────┘  └──────────────────┘           │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

**拆分原则**：

1. **定价域（核心）**：专注价格计算逻辑，这是业务的核心竞争力
2. **促销域（支撑）**：管理促销规则，为定价提供数据支持
3. **商品域（支撑）**：提供商品基础信息
4. **用户域（支撑）**：提供用户信息和分群数据
5. **支付域（支撑）**：处理支付和优惠券
6. **缓存/配置域（通用）**：技术基础设施

**为什么这样拆分？**

```
问题：为什么不把所有逻辑都放在一个"定价域"？

答案：
1. 业务职责分离
   - 促销规则（运营负责）
   - 商品定价（商家负责）
   - 用户分群（市场负责）
   
2. 团队分工
   - 定价团队：核心计算逻辑
   - 促销团队：活动和规则
   - 商品团队：商品信息
   
3. 变化频率不同
   - 定价逻辑：相对稳定
   - 促销规则：频繁变化
   - 商品信息：偶尔变化
```

### 3.5 上下文映射（Context Mapping）

定义子域之间的协作关系：

```
┌─────────────────────────────────────────────────────────────┐
│              上下文映射关系                                  │
└─────────────────────────────────────────────────────────────┘

        ┌──────────────────┐
        │   Pricing        │
        │   Context        │
        │   (定价上下文)    │
        └────────┬─────────┘
                 │
    ┌────────────┼────────────┬────────────┐
    │            │            │            │
    │ ACL        │ ACL        │ ACL        │ ACL: Anti-Corruption Layer
    │            │            │            │      (防腐层)
    ▼            ▼            ▼            ▼
┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐
│Promotion│  │  Item  │  │  User  │  │Payment │
│Context  │  │Context │  │Context │  │Context │
└────────┘  └────────┘  └────────┘  └────────┘
```

**防腐层的作用**：

防腐层（Anti-Corruption Layer）保护领域模型不被外部系统污染。

```go
// ❌ 错误：直接依赖外部服务的数据结构
type PricingService struct {
    promotionClient *external.PromotionClient
}

func (s *PricingService) GetPromotionPrice(itemID int64) int64 {
    // 直接使用外部结构，耦合到外部系统
    promoData := s.promotionClient.GetPromotion(itemID)
    return promoData.ActivityPrice  // 如果外部改字段名，这里就挂了
}

// ✅ 正确：通过防腐层转换
type PricingService struct {
    promotionAdapter PromotionAdapter  // 防腐层接口
}

// 防腐层接口（定价域定义）
type PromotionAdapter interface {
    GetPromotionInfo(itemID int64) *PromotionInfo
}

// 定价域的促销信息（领域模型）
type PromotionInfo struct {
    ActivityID int64
    Price      int64
    Type       PromotionType
}

// 防腐层实现（基础设施层）
type PromotionAdapterImpl struct {
    externalClient *external.PromotionClient
}

func (a *PromotionAdapterImpl) GetPromotionInfo(itemID int64) *PromotionInfo {
    // 外部数据转换为领域模型
    externalData := a.externalClient.GetPromotion(itemID)
    
    return &PromotionInfo{
        ActivityID: externalData.ActivityId,
        Price:      externalData.ActivityPrice,
        Type:       convertType(externalData.ActivityType),
    }
}
```

**收益**：
- ✅ 外部系统变化不影响领域模型
- ✅ 保持领域模型的纯粹性
- ✅ 易于切换外部服务实现

---

## 四、战术设计实践

战略设计得到了概念模型和子域划分，战术设计则是将概念模型映射为代码模型。

### 4.1 实体（Entity）与值对象（Value Object）

#### 实体：有唯一标识和生命周期

```go
// ✅ 实体：PriceEntity（价格实体）
type PriceEntity struct {
    // 唯一标识
    ItemID int64
    SkuID  int64
    
    // 生命周期状态
    Status PriceStatus  // Draft(草稿)、Active(生效)、Expired(过期)
    
    // 属性（使用值对象）
    basePrice    *BasePrice
    promotions   []Promotion
    fees         []Fee
    
    // 时间戳
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 实体的行为（封装业务规则）
func (e *PriceEntity) ApplyPromotion(promo Promotion) error {
    // 业务规则1：促销价不能高于折扣价
    if promo.Price > e.basePrice.DiscountPrice {
        return errors.New("promotion price cannot exceed discount price")
    }
    
    // 业务规则2：同一类型的促销只能有一个
    for _, existing := range e.promotions {
        if existing.Type == promo.Type {
            return errors.New("promotion type already exists")
        }
    }
    
    e.promotions = append(e.promotions, promo)
    e.UpdatedAt = time.Now()
    return nil
}

func (e *PriceEntity) Activate() error {
    // 业务规则：只有草稿状态可以激活
    if e.Status != Draft {
        return errors.New("only draft price can be activated")
    }
    
    e.Status = Active
    e.UpdatedAt = time.Now()
    return nil
}
```

#### 值对象：无唯一标识，不可变

```go
// ✅ 值对象：Price（价格）
type Price struct {
    amount   int64   // 金额（分为单位）
    currency string  // 货币类型
}

// 不可变：所有操作返回新对象
func NewPrice(amount int64, currency string) (Price, error) {
    if amount < 0 {
        return Price{}, errors.New("price cannot be negative")
    }
    return Price{amount: amount, currency: currency}, nil
}

// 值对象的行为（返回新对象）
func (p Price) Add(other Price) (Price, error) {
    if p.currency != other.currency {
        return Price{}, errors.New("currency mismatch")
    }
    return Price{
        amount:   p.amount + other.amount,
        currency: p.currency,
    }, nil
}

func (p Price) Multiply(factor int64) Price {
    return Price{
        amount:   p.amount * factor,
        currency: p.currency,
    }
}

func (p Price) IsZero() bool {
    return p.amount == 0
}

// ✅ 值对象：Promotion（促销）
type Promotion struct {
    activityID int64
    name       string
    price      Price
    startTime  time.Time
    endTime    time.Time
}

// 值对象的行为
func (p Promotion) IsActive() bool {
    now := time.Now()
    return now.After(p.startTime) && now.Before(p.endTime)
}

func (p Promotion) IsExpired() bool {
    return time.Now().After(p.endTime)
}
```

**实体 vs 值对象的判断标准**：

```
问题：某个概念应该是实体还是值对象？

判断标准：
1. 是否需要追踪其变化历史？
   - 需要 → 实体
   - 不需要 → 值对象

2. 是否关心"哪一个"？
   - 关心 → 实体（如：哪个商品）
   - 不关心 → 值对象（如：100元就是100元）

3. 是否有生命周期？
   - 有 → 实体（如：价格实体从创建到生效到过期）
   - 无 → 值对象（如：金额没有生命周期）

案例：
- 商品价格：实体（需要知道是哪个商品的价格）
- 100元：值对象（不关心是哪张100元钞票）
- 订单：实体（需要追踪订单状态变化）
- 收货地址：值对象（相同地址信息是等价的）
```

### 4.2 聚合根（Aggregate Root）

聚合根是一组相关对象的入口，保证业务规则的一致性。

**聚合根设计原则**：
1. 满足业务一致性（促销、费用、价格必须一致）
2. 满足数据完整性（不存在没有基础价格的价格实体）
3. 考虑技术限制（避免加载过大数据）

```go
// ✅ 聚合根：PricingAggregate
type PricingAggregate struct {
    // 聚合根ID
    id string
    
    // 实体
    priceEntity *PriceEntity
    
    // 值对象
    context *PricingContext
    
    // 领域事件
    events []DomainEvent
    
    // 版本号（乐观锁）
    version int64
}

// 聚合根的行为（封装业务规则）
func (a *PricingAggregate) CalculatePrice() (*PricingResult, error) {
    // 步骤1：业务规则校验
    if err := a.validate(); err != nil {
        return nil, err
    }
    
    // 步骤2：选择促销（业务规则）
    promotion := a.selectPromotion()
    
    // 步骤3：计算基础价格
    baseAmount := a.calculateBaseAmount(promotion)
    
    // 步骤4：计算费用
    feeAmount := a.calculateFees()
    
    // 步骤5：应用优惠
    discountAmount := a.applyDiscounts()
    
    // 步骤6：计算最终价格
    finalPrice := baseAmount + feeAmount - discountAmount
    if finalPrice < 0 {
        finalPrice = 0  // 价格保护
    }
    
    // 步骤7：生成领域事件
    a.addEvent(&PriceCalculatedEvent{
        AggregateID: a.id,
        FinalPrice:  finalPrice,
        Timestamp:   time.Now(),
    })
    
    return &PricingResult{
        FinalPrice: finalPrice,
        Breakdown:  a.buildBreakdown(baseAmount, feeAmount, discountAmount),
    }, nil
}

// 业务规则：选择促销（优先级规则）
func (a *PricingAggregate) selectPromotion() *Promotion {
    promotions := a.priceEntity.promotions
    
    // 规则1：秒杀优先（优先级最高）
    for _, promo := range promotions {
        if promo.Type == FlashSale && promo.IsActive() {
            return &promo
        }
    }
    
    // 规则2：新用户价（次优先级）
    if a.context.IsNewUser {
        for _, promo := range promotions {
            if promo.Type == NewUserPrice && promo.IsActive() {
                return &promo
            }
        }
    }
    
    // 规则3：默认折扣价
    return nil  // 使用基础折扣价
}

// 业务规则校验
func (a *PricingAggregate) validate() error {
    // 规则1：市场价必须大于0
    if a.priceEntity.basePrice.MarketPrice <= 0 {
        return errors.New("market price must be positive")
    }
    
    // 规则2：折扣价不能大于市场价
    if a.priceEntity.basePrice.DiscountPrice > a.priceEntity.basePrice.MarketPrice {
        return errors.New("discount price cannot exceed market price")
    }
    
    // 规则3：数量必须大于0
    if a.context.Quantity <= 0 {
        return errors.New("quantity must be positive")
    }
    
    return nil
}

// 领域事件管理
func (a *PricingAggregate) addEvent(event DomainEvent) {
    a.events = append(a.events, event)
}

func (a *PricingAggregate) GetEvents() []DomainEvent {
    return a.events
}

func (a *PricingAggregate) ClearEvents() {
    a.events = []DomainEvent{}
}
```

**聚合根边界的确定**：

```
问题：什么应该放在聚合根内？什么应该放在聚合根外？

判断标准：
1. 是否需要保证事务一致性？
   - 需要 → 放在聚合根内
   - 不需要 → 放在聚合根外

2. 是否需要同时修改？
   - 需要 → 放在聚合根内
   - 不需要 → 放在聚合根外

3. 是否影响聚合根的状态？
   - 影响 → 放在聚合根内
   - 不影响 → 放在聚合根外

案例：
定价聚合根内：
- 基础价格（必须同时存在）
- 促销信息（影响最终价格）
- 费用信息（影响最终价格）

定价聚合根外：
- 用户信息（只是查询，不修改）
- 商品库存（独立的聚合根）
- 订单信息（独立的聚合根）
```

### 4.3 领域服务（Domain Service）

**什么时候使用领域服务？**

不适合放在聚合根里的领域逻辑，可以放在领域服务中：

```go
// ❌ 不适合放在聚合根：跨聚合根的逻辑
// 例：从多个促销活动中选择最优的一个

// ✅ 使用领域服务
type PromotionSelectionService struct {
    rules []SelectionRule
}

// 领域服务：选择最优促销
func (s *PromotionSelectionService) SelectBestPromotion(
    promotions []*Promotion,
    context *PricingContext,
) *Promotion {
    var bestPromotion *Promotion
    lowestPrice := int64(math.MaxInt64)
    
    for _, promo := range promotions {
        // 检查是否适用
        if !s.isApplicable(promo, context) {
            continue
        }
        
        // 选择价格最低的
        if promo.Price < lowestPrice {
            lowestPrice = promo.Price
            bestPromotion = promo
        }
    }
    
    return bestPromotion
}

func (s *PromotionSelectionService) isApplicable(
    promo *Promotion,
    context *PricingContext,
) bool {
    // 检查时间有效性
    if !promo.IsActive() {
        return false
    }
    
    // 检查用户类型
    if promo.Type == NewUserPrice && !context.IsNewUser {
        return false
    }
    
    // 检查购买数量
    if context.Quantity < promo.MinQuantity {
        return false
    }
    
    return true
}

// 领域服务：复杂价格计算（如：买N件享M折）
type BundlePriceCalculator struct{}

func (c *BundlePriceCalculator) Calculate(
    basePrice int64,
    quantity int64,
    config *BundleConfig,
) int64 {
    // 计算可享受优惠的轮数
    rounds := quantity / config.MinQuantity
    effectiveRounds := min(rounds, config.MaxRounds)
    
    // 计算每轮优惠
    var discountPerRound int64
    switch config.DiscountType {
    case PercentageDiscount:
        // 百分比折扣：basePrice * minQty * (1 - discount%)
        discountPerRound = basePrice * config.MinQuantity * config.Discount / 10000
    case FixedDiscount:
        // 固定金额折扣
        discountPerRound = config.Discount
    case FixedPrice:
        // 固定总价：原价 - 固定价
        discountPerRound = basePrice * config.MinQuantity - config.Discount
    }
    
    // 总价 = 原价 * 数量 - 优惠 * 有效轮数
    return basePrice * quantity - discountPerRound * effectiveRounds
}
```

**领域服务 vs 应用服务**：

```
领域服务（Domain Service）：
- 包含业务逻辑
- 操作领域对象
- 无状态
- 例：促销选择、复杂价格计算

应用服务（Application Service）：
- 编排领域对象
- 处理事务
- 协调外部服务
- 例：处理HTTP请求、管理数据库事务
```

### 4.4 贫血模型 vs 充血模型

**我们的选择：混合模式**

```go
// ✅ 核心领域逻辑：充血模型
type PricingAggregate struct {
    id          string
    priceEntity *PriceEntity
    
    // 富含业务逻辑的方法
}

func (a *PricingAggregate) CalculatePrice() (*PricingResult, error) {
    // 封装复杂的业务规则
    // 包含促销选择、价格计算、优惠应用等逻辑
}

func (a *PricingAggregate) ApplyPromotion(promo *Promotion) error {
    // 封装促销应用的业务规则
}

// ✅ 简单CRUD：贫血模型
type PriceSnapshot struct {
    ID         int64
    OrderID    int64
    Price      int64
    CreatedAt  time.Time
}

// 简单的数据访问对象，没有业务逻辑
type PriceSnapshotRepository interface {
    Save(snapshot *PriceSnapshot) error
    FindByOrderID(orderID int64) (*PriceSnapshot, error)
}
```

**选择标准**：

```
充血模型（适用场景）：
- 核心业务逻辑复杂
- 业务规则频繁变化
- 需要封装业务不变性

贫血模型（适用场景）：
- 简单CRUD操作
- 数据传输对象（DTO）
- 持久化对象（PO）

混合使用：
- 领域层：充血模型（封装业务逻辑）
- 应用层：贫血模型（DTO）
- 基础设施层：贫血模型（PO）
```

### 4.5 体现业务语义的代码

代码应该体现业务含义，让非技术人员也能理解：

```go
// ❌ 错误：没有业务含义
func (a *PricingAggregate) UpdateStatus(status int) error {
    a.status = status  // 什么业务操作？为什么要这样做？
    return nil
}

// ✅ 正确：清晰的业务语义
func (a *PricingAggregate) SubmitForReview() error {
    // 业务规则：只有草稿状态可以提交审核
    if a.status != Draft {
        return errors.New("only draft pricing can be submitted for review")
    }
    
    // 业务操作：提交审核
    a.status = PendingReview
    a.submittedAt = time.Now()
    
    // 发布领域事件
    a.addEvent(&PricingSubmittedEvent{
        AggregateID: a.id,
        SubmittedAt: time.Now(),
    })
    
    return nil
}

func (a *PricingAggregate) Approve(approver string, comment string) error {
    // 业务规则：只有待审核状态可以审批通过
    if a.status != PendingReview {
        return errors.New("only pending pricing can be approved")
    }
    
    // 业务操作：审批通过
    a.status = Approved
    a.approver = approver
    a.approvalComment = comment
    a.approvedAt = time.Now()
    
    // 发布领域事件
    a.addEvent(&PricingApprovedEvent{
        AggregateID: a.id,
        Approver:    approver,
        ApprovedAt:  time.Now(),
    })
    
    return nil
}

func (a *PricingAggregate) Reject(reviewer string, reason string) error {
    // 业务规则：只有待审核状态可以拒绝
    if a.status != PendingReview {
        return errors.New("only pending pricing can be rejected")
    }
    
    // 业务操作：拒绝
    a.status = Rejected
    a.reviewer = reviewer
    a.rejectionReason = reason
    a.rejectedAt = time.Now()
    
    // 发布领域事件
    a.addEvent(&PricingRejectedEvent{
        AggregateID: a.id,
        Reason:      reason,
        RejectedAt:  time.Now(),
    })
    
    return nil
}
```

**收益**：
- ✅ 代码即文档（看方法名就知道做什么）
- ✅ 业务规则显式化（不需要深入代码才能理解）
- ✅ 易于沟通（产品和技术可以用同样的语言）

### 4.6 价格快照与一致性保障

#### 4.6.1 业务场景与挑战

在电商系统中，用户从浏览商品（PDP）到最终下单，价格可能发生变化，这是一个非常常见且重要的问题。

**典型场景**：
```
用户路径：
PDP展示价格 → 加入购物车 → 创建订单
   100元          105元          ???

价格变化的原因：
1. 促销活动已结束（时间到期）
2. 促销库存已用完
3. 缓存未更新（PDP用了旧缓存）
4. 用户身份变化（新用户期限过期）
5. 前后端计算逻辑不一致
6. 价格规则版本不同
```

**核心挑战**：
- ❌ 价格不一致导致用户投诉
- ❌ 可能造成资损风险
- ❌ 影响用户购买体验
- ❌ 需要平衡准确性和用户体验

#### 4.6.2 价格快照机制设计

**核心思路**：在PDP阶段生成价格快照，用户加购/创单时验证快照有效性。

```go
// 价格快照值对象
type PriceSnapshot struct {
    snapshotID    string        // 快照ID
    itemID        int64         // 商品ID
    userID        int64         // 用户ID
    displayPrice  int64         // 展示价格
    promotionID   int64         // 促销活动ID
    snapshotTime  time.Time     // 快照时间
    expireAt      time.Time     // 过期时间
    ruleVersion   string        // 规则版本
}

// 快照不可变
func NewPriceSnapshot(
    itemID int64,
    userID int64,
    priceResult *PricingResult,
    ttl time.Duration,
) *PriceSnapshot {
    return &PriceSnapshot{
        snapshotID:   generateSnapshotID(),
        itemID:       itemID,
        userID:       userID,
        displayPrice: priceResult.FinalPrice,
        promotionID:  priceResult.PromotionID,
        snapshotTime: time.Now(),
        expireAt:     time.Now().Add(ttl),
        ruleVersion:  priceResult.RuleVersion,
    }
}

// 快照是否有效
func (s *PriceSnapshot) IsValid() bool {
    return time.Now().Before(s.expireAt)
}

// 快照是否即将过期
func (s *PriceSnapshot) IsExpiringSoon(threshold time.Duration) bool {
    return time.Until(s.expireAt) < threshold
}
```

#### 4.6.3 完整实现方案

**方案1：PDP阶段生成快照**

```go
// 领域服务：价格快照管理
type PriceSnapshotService struct {
    snapshotRepo SnapshotRepository
    cache        CacheService
}

// PDP阶段：生成价格快照
func (s *PriceSnapshotService) CreateSnapshot(
    itemID int64,
    userID int64,
    priceResult *PricingResult,
) (*PriceSnapshot, error) {
    // 1. 创建快照（10分钟有效期）
    snapshot := NewPriceSnapshot(
        itemID,
        userID,
        priceResult,
        10*time.Minute,
    )
    
    // 2. 存储到Redis（快速访问）
    snapshotKey := fmt.Sprintf("price_snapshot:%d:%d", userID, itemID)
    err := s.cache.SetEx(snapshotKey, snapshot, 10*time.Minute)
    if err != nil {
        return nil, fmt.Errorf("failed to cache snapshot: %w", err)
    }
    
    // 3. 异步持久化（用于审计）
    go s.snapshotRepo.Save(snapshot)
    
    return snapshot, nil
}

// 创单阶段：验证价格快照
func (s *PriceSnapshotService) ValidateSnapshot(
    itemID int64,
    userID int64,
    expectedPrice int64,
) (*SnapshotValidationResult, error) {
    // 1. 获取快照
    snapshotKey := fmt.Sprintf("price_snapshot:%d:%d", userID, itemID)
    snapshot, err := s.cache.Get(snapshotKey)
    
    if err != nil || snapshot == nil {
        // 快照不存在或已过期
        return &SnapshotValidationResult{
            Status:  SnapshotExpired,
            Message: "价格快照已过期，请刷新后重试",
        }, nil
    }
    
    // 2. 检查快照是否即将过期
    if snapshot.IsExpiringSoon(2 * time.Minute) {
        return &SnapshotValidationResult{
            Status:  SnapshotExpiringSoon,
            Message: "价格快照即将过期，建议尽快下单",
        }, nil
    }
    
    // 3. 价格比对（容忍度±1元）
    priceDiff := abs(expectedPrice - snapshot.displayPrice)
    tolerance := int64(100) // ±1元
    
    if priceDiff <= tolerance {
        // 价格一致或差异在容忍范围内
        return &SnapshotValidationResult{
            Status:       SnapshotValid,
            SnapshotID:   snapshot.snapshotID,
            ActualPrice:  expectedPrice,
            Message:      "价格验证通过",
        }, nil
    }
    
    // 4. 价格差异较大，需要用户确认
    return &SnapshotValidationResult{
        Status:       PriceChanged,
        SnapshotID:   snapshot.snapshotID,
        OldPrice:     snapshot.displayPrice,
        NewPrice:     expectedPrice,
        PriceDiff:    priceDiff,
        Message:      fmt.Sprintf("价格已变动%+.2f元，请确认后继续", float64(priceDiff)/100),
        ChangeReason: s.detectPriceChangeReason(snapshot),
    }, nil
}

// 检测价格变化原因
func (s *PriceSnapshotService) detectPriceChangeReason(
    snapshot *PriceSnapshot,
) string {
    // 1. 检查促销是否结束
    promotion := s.promotionService.GetPromotion(snapshot.promotionID)
    if promotion == nil || promotion.IsExpired() {
        return "促销活动已结束"
    }
    
    // 2. 检查促销库存
    if promotion.Stock <= 0 {
        return "促销库存已售罄"
    }
    
    // 3. 检查用户资格
    if promotion.Type == NewUserPromotion {
        user := s.userService.GetUser(snapshot.userID)
        if !user.IsNewUser() {
            return "新用户专享活动已结束"
        }
    }
    
    return "商品价格已调整"
}
```

**方案2：活动有效期前置校验**

```go
// 领域服务：促销有效期管理
type PromotionExpirationService struct {
    warningThreshold time.Duration // 预警阈值（如15分钟）
}

// 检查促销是否即将过期
func (s *PromotionExpirationService) CheckExpiration(
    promotion *Promotion,
) *ExpirationWarning {
    if promotion == nil {
        return nil
    }
    
    timeLeft := time.Until(promotion.EndTime)
    
    // 活动剩余时间 < 预警阈值
    if timeLeft > 0 && timeLeft < s.warningThreshold {
        return &ExpirationWarning{
            PromotionID:  promotion.ActivityID,
            TimeLeft:     timeLeft,
            Message:      fmt.Sprintf("活动即将结束（剩余%d分钟），请尽快下单", int(timeLeft.Minutes())),
            ActivityEndTime: promotion.EndTime,
            Urgency:      s.calculateUrgency(timeLeft),
        }
    }
    
    return nil
}

// 计算紧急程度
func (s *PromotionExpirationService) calculateUrgency(timeLeft time.Duration) UrgencyLevel {
    switch {
    case timeLeft < 2*time.Minute:
        return UrgencyCritical  // 紧急：倒计时显示
    case timeLeft < 5*time.Minute:
        return UrgencyHigh      // 高：红色提示
    case timeLeft < 15*time.Minute:
        return UrgencyMedium    // 中：黄色提示
    default:
        return UrgencyLow       // 低：无需提示
    }
}

// PDP价格计算时集成过期检查
func (a *PricingAggregate) CalculatePDPPrice() (*PricingResult, error) {
    // 正常计算价格
    result := a.calculatePrice()
    
    // 检查促销过期
    if result.Promotion != nil {
        warning := s.expirationService.CheckExpiration(result.Promotion)
        if warning != nil {
            result.ExpirationWarning = warning
        }
    }
    
    return result, nil
}
```

**方案3：库存预锁定**

```go
// 领域服务：促销库存管理
type PromotionStockService struct {
    stockRepo StockRepository
    lockCache CacheService
}

// 加购时预锁定促销库存
func (s *PromotionStockService) ReserveStock(
    itemID int64,
    userID int64,
    quantity int64,
    ttl time.Duration,
) (*StockReservation, error) {
    // 1. 检查促销库存
    promotion := s.promotionRepo.GetPromotion(itemID)
    if promotion == nil {
        return nil, errors.New("promotion not found")
    }
    
    // 2. 尝试锁定库存（使用Redis分布式锁）
    lockKey := fmt.Sprintf("stock_lock:%d:%d", itemID, userID)
    locked, err := s.lockCache.SetNX(lockKey, quantity, ttl)
    
    if !locked || err != nil {
        return nil, errors.New("failed to reserve stock")
    }
    
    // 3. 扣减库存（乐观锁）
    success := s.stockRepo.DeductStock(itemID, quantity, promotion.Version)
    if !success {
        // 回滚锁
        s.lockCache.Delete(lockKey)
        return nil, errors.New("stock not available")
    }
    
    // 4. 创建预订记录
    reservation := &StockReservation{
        ReservationID: generateReservationID(),
        ItemID:        itemID,
        UserID:        userID,
        Quantity:      quantity,
        LockedUntil:   time.Now().Add(ttl),
        Status:        ReservationActive,
    }
    
    s.stockRepo.SaveReservation(reservation)
    
    return reservation, nil
}

// 释放库存（超时或取消订单）
func (s *PromotionStockService) ReleaseStock(reservationID string) error {
    reservation := s.stockRepo.GetReservation(reservationID)
    if reservation == nil {
        return errors.New("reservation not found")
    }
    
    // 回补库存
    s.stockRepo.IncreaseStock(reservation.ItemID, reservation.Quantity)
    
    // 删除锁
    lockKey := fmt.Sprintf("stock_lock:%d:%d", reservation.ItemID, reservation.UserID)
    s.lockCache.Delete(lockKey)
    
    // 更新预订状态
    reservation.Status = ReservationReleased
    s.stockRepo.UpdateReservation(reservation)
    
    return nil
}
```

#### 4.6.4 用户体验优化

**价格变动提示策略**：

```go
// 应用服务：订单创建（集成价格验证）
func (s *OrderApplicationService) CreateOrder(
    req *CreateOrderRequest,
) (*OrderResult, error) {
    // 1. 验证价格快照
    validation, err := s.snapshotService.ValidateSnapshot(
        req.ItemID,
        req.UserID,
        req.ExpectedPrice,
    )
    if err != nil {
        return nil, err
    }
    
    // 2. 处理不同验证结果
    switch validation.Status {
    case SnapshotValid:
        // 价格一致，正常创建订单
        return s.createOrderNormally(req, validation.ActualPrice)
        
    case SnapshotExpired:
        // 快照过期，重新计算价格并返回
        newPrice := s.pricingService.CalculatePrice(req)
        return &OrderResult{
            Status:   OrderPriceRecalculated,
            Message:  "价格已更新，请确认",
            OldPrice: req.ExpectedPrice,
            NewPrice: newPrice.FinalPrice,
            RequireConfirmation: true,
        }, nil
        
    case PriceChanged:
        // 价格变动，需要用户确认
        return &OrderResult{
            Status:   OrderPriceChanged,
            Message:  validation.Message,
            OldPrice: validation.OldPrice,
            NewPrice: validation.NewPrice,
            PriceDiff: validation.PriceDiff,
            ChangeReason: validation.ChangeReason,
            RequireConfirmation: true,
        }, nil
        
    case SnapshotExpiringSoon:
        // 即将过期，提示但允许创建
        order, err := s.createOrderNormally(req, validation.ActualPrice)
        if err != nil {
            return nil, err
        }
        order.Warning = validation.Message
        return order, nil
        
    default:
        return nil, errors.New("unknown validation status")
    }
}

// 价格变动确认后创建订单
func (s *OrderApplicationService) CreateOrderWithPriceConfirmation(
    req *CreateOrderRequest,
    confirmedPrice int64,
) (*OrderResult, error) {
    // 用户已确认价格变动，使用新价格创建订单
    req.ExpectedPrice = confirmedPrice
    req.PriceConfirmed = true
    
    return s.createOrderNormally(req, confirmedPrice)
}
```

**前端UI交互示例**：

```javascript
// 前端处理价格变动
async function createOrder(items) {
    const response = await api.createOrder({
        items: items,
        expectedPrice: getTotalPrice(items),
        snapshotID: getSnapshotID(items),
    });
    
    // 处理价格变动
    if (response.status === 'price_changed') {
        const confirmed = await showPriceChangeDialog({
            title: '价格变动提示',
            oldPrice: response.oldPrice,
            newPrice: response.newPrice,
            priceDiff: response.priceDiff,
            reason: response.changeReason,
            message: response.message,
        });
        
        if (confirmed) {
            // 用户确认，使用新价格创建订单
            return await api.createOrder({
                items: items,
                expectedPrice: response.newPrice,
                priceConfirmed: true,
            });
        } else {
            // 用户取消
            return null;
        }
    }
    
    // 处理价格重新计算
    if (response.status === 'price_recalculated') {
        const confirmed = await showPriceRecalculatedDialog({
            message: '价格已更新，请确认',
            oldPrice: response.oldPrice,
            newPrice: response.newPrice,
        });
        
        if (confirmed) {
            return await createOrder(items); // 重试
        }
    }
    
    return response;
}
```

#### 4.6.5 监控与告警

```go
// 领域服务：价格一致性监控
type PriceConsistencyMonitor struct {
    metrics MetricsService
    alerter AlertService
}

// 记录价格差异
func (m *PriceConsistencyMonitor) RecordPriceDifference(
    itemID int64,
    snapshotPrice int64,
    actualPrice int64,
    reason string,
) {
    diff := abs(actualPrice - snapshotPrice)
    diffPercent := float64(diff) / float64(snapshotPrice) * 100
    
    // 记录指标
    m.metrics.RecordPriceDiff(itemID, diff, diffPercent)
    
    // 差异过大告警
    if diffPercent > 10 {
        m.alerter.Send(&Alert{
            Level:   AlertLevelHigh,
            Title:   "价格差异过大",
            Message: fmt.Sprintf("商品%d价格差异%.2f%%", itemID, diffPercent),
            Reason:  reason,
        })
    }
}

// 监控快照过期率
func (m *PriceConsistencyMonitor) RecordSnapshotExpiration(
    itemID int64,
    expiredAt time.Time,
) {
    m.metrics.IncSnapshotExpirationCount(itemID)
    
    // 快照过期率过高告警
    expirationRate := m.metrics.GetSnapshotExpirationRate(time.Hour)
    if expirationRate > 0.2 { // 20%
        m.alerter.Send(&Alert{
            Level:   AlertLevelMedium,
            Title:   "快照过期率过高",
            Message: fmt.Sprintf("过去1小时快照过期率%.2f%%", expirationRate*100),
        })
    }
}
```

#### 4.6.6 最佳实践总结

| 措施 | 说明 | 优先级 |
|------|------|--------|
| **价格快照** | PDP生成快照（10分钟），加购/创单时验证 | P0 |
| **价格校验** | 前端传入期望价格，后端验证（容忍度±1元） | P0 |
| **活动预警** | 活动剩余时间<15分钟时前置提示 | P1 |
| **库存预锁** | 加购时预锁定促销库存（5分钟） | P1 |
| **降级策略** | 促销失效时自动降级到原价 | P0 |
| **用户提示** | 价格变动时明确告知原因并二次确认 | P0 |
| **监控告警** | 价格差异率、快照过期率监控 | P1 |

**关键设计原则**：
1. **快照不可变**：价格快照创建后不可修改，保证一致性
2. **短期有效**：快照有效期10-15分钟，平衡准确性和体验
3. **容忍小差异**：±1元差异可接受，避免频繁提示
4. **明确告知**：价格变动时必须告知原因和差异金额
5. **用户确认**：价格上涨时必须用户二次确认
6. **降级保护**：促销失效时自动降级到原价，不阻断流程

---

## 五、代码架构实践

### 5.1 六边形架构（Hexagonal Architecture）

我们采用六边形架构（也称为端口和适配器架构）：

```
┌────────────────────────────────────────────────────────┐
│                   六边形架构                            │
└────────────────────────────────────────────────────────┘

                      外部世界
                         ↓
    ┌─────────────────────────────────────────────┐
    │          Adapters (适配器层)                 │
    │  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
    │  │ HTTP     │  │ gRPC     │  │ Message  │  │
    │  │ Handler  │  │ Handler  │  │ Consumer │  │
    │  └──────────┘  └──────────┘  └──────────┘  │
    └──────────────────┬──────────────────────────┘
                       │ Port (端口)
    ┌──────────────────▼──────────────────────────┐
    │          Application Layer (应用层)          │
    │  ┌──────────────────────────────────────┐   │
    │  │  PricingApplicationService           │   │
    │  │  • CalculateDisplayPrice()           │   │
    │  │  • CalculateOrderPrice()             │   │
    │  │  • CalculatePaymentPrice()           │   │
    │  └──────────────────────────────────────┘   │
    └──────────────────┬──────────────────────────┘
                       │
    ┌──────────────────▼──────────────────────────┐
    │          Domain Layer (领域层)               │
    │  ┌──────────────────────────────────────┐   │
    │  │  PricingAggregate (聚合根)           │   │
    │  │  • CalculatePrice()                  │   │
    │  │  • ApplyPromotion()                  │   │
    │  └──────────────────────────────────────┘   │
    │  ┌──────────────────────────────────────┐   │
    │  │  PromotionSelectionService (领域服务)│   │
    │  │  BundlePriceCalculator (领域服务)    │   │
    │  └──────────────────────────────────────┘   │
    └──────────────────┬──────────────────────────┘
                       │ Port (端口)
    ┌──────────────────▼──────────────────────────┐
    │       Infrastructure Layer (基础设施层)      │
    │  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
    │  │ Database │  │ Cache    │  │ External │  │
    │  │ Repo     │  │ Service  │  │ API      │  │
    │  └──────────┘  └──────────┘  └──────────┘  │
    └─────────────────────────────────────────────┘
```

**核心原则**：
1. 领域层是核心，不依赖外部
2. 外层依赖内层（依赖倒置）
3. 通过端口（接口）隔离

### 5.2 实际代码结构

```
pricing-service/
├── cmd/
│   └── main.go                          # 启动入口
│
├── internal/
│   ├── adapter/                         # 适配器层（外层）
│   │   ├── http/                        # HTTP适配器
│   │   │   └── handler.go
│   │   ├── grpc/                        # gRPC适配器
│   │   │   └── server.go
│   │   └── event/                       # 事件适配器
│   │       └── consumer.go
│   │
│   ├── application/                     # 应用层
│   │   ├── service/
│   │   │   ├── pricing_service.go       # 应用服务
│   │   │   └── pricing_service_test.go
│   │   ├── dto/
│   │   │   ├── request.go              # 请求DTO
│   │   │   └── response.go             # 响应DTO
│   │   └── port/
│   │       ├── inbound.go              # 入站端口（接口）
│   │       └── outbound.go             # 出站端口（接口）
│   │
│   ├── domain/                          # 领域层（核心）
│   │   ├── model/
│   │   │   ├── aggregate/              # 聚合根
│   │   │   │   └── pricing_aggregate.go
│   │   │   ├── entity/                 # 实体
│   │   │   │   └── price_entity.go
│   │   │   └── valueobject/            # 值对象
│   │   │       ├── price.go
│   │   │       ├── promotion.go
│   │   │       └── price_snapshot.go   # 价格快照（值对象）
│   │   ├── service/                    # 领域服务
│   │   │   ├── promotion_selector.go
│   │   │   ├── bundle_calculator.go
│   │   │   ├── snapshot_service.go     # 快照管理服务
│   │   │   └── consistency_monitor.go  # 一致性监控服务
│   │   ├── repository/                 # 仓储接口（领域层定义）
│   │   │   ├── pricing_repository.go
│   │   │   └── snapshot_repository.go  # 快照仓储
│   │   └── event/                      # 领域事件
│   │       ├── price_calculated.go
│   │       └── price_changed.go        # 价格变动事件
│   │
│   └── infrastructure/                  # 基础设施层（外层）
│       ├── persistence/                 # 持久化实现
│       │   ├── mysql/
│       │   │   └── pricing_repo_impl.go
│       │   └── redis/
│       │       └── cache_impl.go
│       ├── external/                    # 外部服务适配器
│       │   ├── promotion_adapter.go
│       │   └── item_adapter.go
│       └── config/
│           └── config.go
│
└── pkg/                                 # 共享代码
    ├── errors/
    └── utils/
```

### 5.3 依赖方向

```
核心原则：依赖倒置原则（Dependency Inversion Principle）

┌────────────────────────────────────────────────────┐
│  依赖方向：外层 ───> 内层                           │
├────────────────────────────────────────────────────┤
│                                                     │
│  HTTP Handler (适配器层)                           │
│      ↓ 依赖                                        │
│  PricingApplicationService (应用层)                │
│      ↓ 依赖                                        │
│  PricingAggregate (领域层) ← 核心                  │
│      ↓ 依赖接口（不依赖实现）                      │
│  PricingRepository interface (领域层定义接口)     │
│      ↑ 实现接口                                    │
│  MySQLPricingRepository (基础设施层)               │
│                                                     │
└────────────────────────────────────────────────────┘
```

### 5.4 实际代码示例

#### 领域层：定义接口

```go
// domain/repository/pricing_repository.go

package repository

// 仓储接口（由领域层定义，基础设施层实现）
type PricingRepository interface {
    Save(aggregate *aggregate.PricingAggregate) error
    FindByID(id string) (*aggregate.PricingAggregate, error)
    FindByItemID(itemID int64) (*aggregate.PricingAggregate, error)
}

// 促销适配器接口（防腐层）
type PromotionAdapter interface {
    GetPromotionInfo(itemID int64) (*valueobject.Promotion, error)
    ListActivePromotions(itemIDs []int64) ([]*valueobject.Promotion, error)
}
```

#### 基础设施层：实现接口

```go
// infrastructure/persistence/mysql/pricing_repo_impl.go

package mysql

type MySQLPricingRepository struct {
    db *sql.DB
}

// 实现领域层定义的接口
func (r *MySQLPricingRepository) Save(
    agg *aggregate.PricingAggregate,
) error {
    // 将聚合根转换为数据库模型
    dbModel := r.toDBModel(agg)
    
    // 保存到数据库
    query := "INSERT INTO pricing (id, item_id, price, status) VALUES (?, ?, ?, ?)"
    _, err := r.db.Exec(query, dbModel.ID, dbModel.ItemID, dbModel.Price, dbModel.Status)
    return err
}

func (r *MySQLPricingRepository) FindByID(id string) (*aggregate.PricingAggregate, error) {
    // 从数据库查询
    query := "SELECT * FROM pricing WHERE id = ?"
    row := r.db.QueryRow(query, id)
    
    // 转换为聚合根
    return r.toAggregate(row)
}
```

#### 应用层：协调各层

```go
// application/service/pricing_service.go

package service

type PricingApplicationService struct {
    // 依赖领域层接口（不是实现）
    pricingRepo      repository.PricingRepository
    promotionAdapter repository.PromotionAdapter
    
    // 领域服务
    promotionSelector *domainservice.PromotionSelectionService
}

func (s *PricingApplicationService) CalculateDisplayPrice(
    req *dto.CalculateDisplayPriceRequest,
) (*dto.PriceResponse, error) {
    // 1. 获取促销信息（通过适配器）
    promotions, err := s.promotionAdapter.ListActivePromotions(req.ItemIDs)
    if err != nil {
        return nil, fmt.Errorf("failed to get promotions: %w", err)
    }
    
    // 2. 构建定价上下文
    context := &domain.PricingContext{
        UserID:    req.UserID,
        IsNewUser: req.IsNewUser,
        Quantity:  req.Quantity,
    }
    
    // 3. 构建聚合根
    aggregate := aggregate.NewPricingAggregate(req.ItemID)
    aggregate.SetContext(context)
    aggregate.SetPromotions(promotions)
    
    // 4. 执行领域逻辑
    result, err := aggregate.CalculatePrice()
    if err != nil {
        return nil, fmt.Errorf("failed to calculate price: %w", err)
    }
    
    // 5. 转换为DTO返回
    return &dto.PriceResponse{
        ItemID:     req.ItemID,
        FinalPrice: result.FinalPrice,
        Breakdown:  s.toBreakdownDTO(result.Breakdown),
    }, nil
}
```

#### 适配器层：处理HTTP请求

```go
// adapter/http/handler.go

package http

type PricingHandler struct {
    pricingService *service.PricingApplicationService
}

func (h *PricingHandler) CalculateDisplayPrice(c *gin.Context) {
    // 1. 解析HTTP请求
    var req struct {
        ItemIDs   []int64 `json:"item_ids"`
        UserID    int64   `json:"user_id"`
        Quantity  int64   `json:"quantity"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }
    
    // 2. 转换为应用层DTO
    appReq := &dto.CalculateDisplayPriceRequest{
        ItemIDs:  req.ItemIDs,
        UserID:   req.UserID,
        Quantity: req.Quantity,
    }
    
    // 3. 调用应用服务
    resp, err := h.pricingService.CalculateDisplayPrice(appReq)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // 4. 返回HTTP响应
    c.JSON(200, resp)
}
```

---

## 六、DDD带来的收益

### 6.1 量化收益

根据实际项目经验，DDD带来的量化收益：

| 指标 | 改造前 | 改造后 | 改善 |
|------|--------|--------|------|
| **代码重复率** | 40% | 15% | -62% |
| **新功能开发时间** | 2周 | 3天 | -86% |
| **单元测试覆盖率** | 45% | 90% | +100% |
| **Bug密度** | 3.5/KLOC | 0.8/KLOC | -77% |
| **需求变更响应时间** | 3天 | 0.5天 | -83% |
| **新人上手时间** | 2周 | 3天 | -85% |
| **价格不一致投诉** | 2-3起/月 | <0.01% | -99% |
| **价格变动处理时间** | 人工处理2小时 | 自动处理秒级 | -99.9% |

### 6.2 质量收益

**1. 概念清晰，沟通顺畅**

```
改造前：
- 技术、产品、业务各说各话
- "价格"有10种不同理解
- 需求讨论会经常吵架
- 新人需要2周才能理解业务

改造后：
- 统一语言，所有人说同一种话
- 文档和代码术语一致
- 需求讨论高效，快速达成共识
- 新人3天上手
```

**2. 业务规则显式化**

```go
// ❌ 改造前：隐藏在代码中
if price < 100 && user.days <= 7 && user.orders == 0 {
    // 什么规则？为什么这么做？
}

// ✅ 改造后：显式的业务规则
func (s *PromotionSelector) IsNewUserEligible(user *User) bool {
    return user.DaysSinceRegister <= NewUserMaxDays &&
           user.TotalOrders == 0
}
```

**3. 易于测试**

```go
// ✅ 聚合根可以独立测试
func TestPricingAggregate_CalculatePrice(t *testing.T) {
    // Arrange
    aggregate := NewPricingAggregate("test-001")
    aggregate.SetBasePrice(NewPrice(100, "USD"))
    aggregate.AddPromotion(NewFlashSalePromotion(80))
    
    // Act
    result, err := aggregate.CalculatePrice()
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, int64(80), result.FinalPrice)
}

// ✅ 价格快照可以独立测试
func TestPriceSnapshotService_ValidateSnapshot(t *testing.T) {
    // Arrange
    service := NewPriceSnapshotService(mockRepo, mockCache)
    snapshot := NewPriceSnapshot(
        itemID: 123,
        userID: 456,
        price:  10000, // 100元
        ttl:    10*time.Minute,
    )
    
    // Act
    result, err := service.ValidateSnapshot(123, 456, 10000)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, SnapshotValid, result.Status)
}

// ✅ 价格变动场景测试
func TestPriceSnapshot_PriceChanged(t *testing.T) {
    tests := []struct {
        name          string
        snapshotPrice int64
        actualPrice   int64
        expectedStatus SnapshotStatus
    }{
        {"价格未变", 10000, 10000, SnapshotValid},
        {"小幅上涨+50分", 10000, 10050, SnapshotValid},    // ±1元内可接受
        {"大幅上涨+500分", 10000, 10500, PriceChanged},    // >1元需确认
        {"价格下降-200分", 10000, 9800, PriceChanged},     // 降价也需提示
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validatePrice(tt.snapshotPrice, tt.actualPrice)
            assert.Equal(t, tt.expectedStatus, result.Status)
        })
    }
}
```

**4. 价格一致性保障**

通过价格快照机制，DDD帮助我们实现了价格一致性保障：

```
改造前：
- PDP展示价、订单价、支付价各自计算
- 价格不一致导致用户投诉（每月2-3起）
- 无价格变动提示，用户体验差
- 促销结束后仍展示促销价，导致资损

改造后：
- PDP生成价格快照（10分钟有效）
- 加购/创单时验证快照有效性
- 价格变动时明确提示原因（活动结束/库存售罄）
- 用户二次确认机制，降低投诉
- 价格不一致投诉率<0.01%
- 资损事件降至0起
```

**具体案例**：

```
案例1：促销活动结束
- 用户在PDP看到秒杀价100元
- 5分钟后促销结束
- 用户创建订单时：
  ✅ 系统提示"促销活动已结束，当前价格120元"
  ✅ 用户可以选择继续或取消
  ✅ 避免了"看到100元，支付120元"的投诉

案例2：促销库存售罄
- 用户在PDP看到促销价80元
- 库存在浏览期间售罄
- 用户加购时：
  ✅ 系统提示"促销库存已售罄，当前价格100元"
  ✅ 可选择等待补货或购买原价
  ✅ 透明的价格变动处理

案例3：新用户期限过期
- 新用户在注册7天内享受85元新人价
- 用户第8天创建订单
- 系统处理：
  ✅ 自动降级到折扣价90元
  ✅ 提示"新用户专享活动已结束"
  ✅ 差异5元，用户可接受
```

### 6.3 维护性收益

**1. 修改影响范围可控**

```
需求变更：秒杀优先级临时提升

改造前：
- 修改核心计算函数（影响所有品类）
- 需要回归测试所有场景
- 风险高，不敢改

改造后：
- 修改PromotionSelector（只影响选择逻辑）
- 单元测试覆盖
- 风险可控，放心改
```

**2. 业务变化适应性强**

```
新需求：增加VIP专享价

改造前：
- 修改多个if-else分支
- 容易遗漏某个场景
- 容易引入bug

改造后：
- 新增VIPPromotion类
- 实现Promotion接口
- 注册到规则引擎
- 不影响现有代码
```

---

## 七、常见误区与最佳实践

### 7.1 常见误区

#### 误区1：深陷DDD概念

```go
// ❌ 过度设计：生搬硬套概念
type PriceValueObject struct {
    value Price
}

type PriceEntity struct {
    id              string
    priceVO         *PriceValueObject      // 过度抽象
    domainEvents    []DomainEvent          // 不必要的复杂性
    aggregateRoot   *AggregateRoot         // 概念混乱
}

// ✅ 简单实用：根据需要选择
type Price struct {
    amount   int64
    currency string
}

type PriceEntity struct {
    itemID    int64
    price     Price
    createdAt time.Time
}
```

#### 误区2：试图一次性设计完美

```
❌ 错误做法：
- 花3个月设计"完美"的领域模型
- 考虑所有可能的场景
- 抽象所有可能的概念
- 结果：过度设计，无法落地

✅ 正确做法：
- Week 1-2: 基础模型（覆盖80%场景）
- Week 3-4: 迭代优化
- Week 5-6: 新增场景
- 持续迭代，逐步完善
```

#### 误区3：忽视统一语言

```
❌ 问题：
- 技术：MarketPrice、DiscountPrice
- 产品：原价、折扣价、活动价
- 业务：市场价、会员价、到手价
- 客服：标价、优惠价、实付价

结果：
- 每次对话都要先对齐概念
- 需求理解错误
- 代码和文档脱节

✅ 解决：
- 建立统一语言表
- 所有文档、代码、讨论统一使用
- 定期Review和更新
```

#### 误区4：过度使用领域事件

```go
// ❌ 过度使用：为事件而事件
type PriceUpdatedEvent struct { ... }
type PriceCreatedEvent struct { ... }
type PriceDeletedEvent struct { ... }
type PriceValidatedEvent struct { ... }  // 不必要
type PriceCalculatedEvent struct { ... } // 不必要

// ✅ 适度使用：只在需要异步通知时使用
type PriceApprovedEvent struct {
    // 需要通知其他服务：价格已审批通过
}
```

### 7.2 最佳实践

#### 1. 从业务出发

```
✅ 正确顺序：
1. 理解业务（用例分析）
2. 统一语言（概念抽取）
3. 建立模型（概念模型）
4. 映射代码（代码模型）

❌ 错误顺序：
1. 看到需求
2. 开始设计表结构
3. 写CRUD代码
4. 发现业务规则
5. 用if-else实现
```

#### 2. 持续迭代

```
领域模型不是一次性设计出来的

Week 1-2: 基础模型
  └─ 支持核心场景

Week 3-4: 发现问题
  └─ 调整模型

Week 5-6: 新需求
  └─ 扩展模型

Week 7-8: 重构
  └─ 优化模型
```

#### 3. 团队协作

```
┌──────────────────────────────────────┐
│     领域模型是团队的共同成果          │
├──────────────────────────────────────┤
│  产品：提供业务视角和需求            │
│  开发：提供技术实现和约束            │
│  测试：提供边界场景和异常情况        │
│  运营：提供实际问题和反馈            │
└──────────────────────────────────────┘
```

#### 4. 适度抽象

```go
// ✅ 简单场景：简单实现
type PriceHistory struct {
    OrderID   int64
    Price     int64
    CreatedAt time.Time
}

func SaveHistory(history *PriceHistory) error {
    // 简单CRUD，不需要复杂建模
}

// ✅ 复杂场景：领域模型
type PricingAggregate struct {
    // 复杂业务规则
}

func (a *PricingAggregate) CalculatePrice() {
    // 封装复杂的计算逻辑
}
```

#### 5. 价格快照机制

价格快照是DDD在计价系统中的重要实践：

```
核心设计：
1. 快照是值对象（不可变）
2. 快照管理是领域服务（PriceSnapshotService）
3. 快照验证是应用服务的职责
4. 快照存储通过基础设施层（Redis + DB）

最佳实践：
✅ 快照有效期：10-15分钟（平衡准确性和体验）
✅ 价格容忍度：±1元（避免频繁提示）
✅ 变动提示：明确告知原因和差异
✅ 二次确认：价格上涨时必须确认
✅ 降级保护：促销失效自动降级到原价
✅ 监控告警：差异率>1%告警
```

---

## 八、总结

### 8.1 DDD的核心价值

1. **统一语言**：消除沟通障碍，提升协作效率
2. **领域模型**：业务知识显式化，代码即文档
3. **分层架构**：职责清晰，易于维护和扩展
4. **持续迭代**：适应业务变化，拥抱需求变更

### 8.2 实施要点

```
战略设计：
  ├─ 用例分析（理解业务玩法）
  ├─ 统一语言（概念抽取和定义）
  ├─ 概念模型（关系梳理）
  └─ 子域划分（化繁为简）

战术设计：
  ├─ 实体与值对象（映射概念）
  ├─ 聚合根（封装业务规则）
  ├─ 领域服务（处理跨聚合逻辑）
  └─ 业务语义代码（代码可读）

代码架构：
  ├─ 六边形架构（依赖倒置）
  ├─ 分层清晰（职责单一）
  └─ 接口隔离（易于测试）
```

### 8.3 给后来者的建议

**1. 不要害怕DDD的概念体系**

从简单开始，逐步深入：
- 第1周：建立统一语言
- 第2周：绘制概念模型
- 第3周：实现简单聚合根
- 第4周：逐步完善

**2. 重视统一语言**

> 没有统一语言就没有概念模型，没有概念模型就没有好的代码

投入时间在统一语言上，回报率最高

**3. 持续迭代**

领域模型是演进出来的，不是设计出来的：
- 业务理解加深 → 模型调整
- 抽象角度变化 → 模型重构
- 业务需求变化 → 模型扩展

**4. 团队协作**

DDD是团队工作，不是个人英雄主义：
- 产品提供业务视角
- 开发提供技术实现
- 测试提供边界场景
- 运营提供实际反馈

**5. 重视价格一致性**

价格一致性是电商系统的生命线：
- 第一时间建立价格快照机制
- 明确定义价格变动处理策略
- 充分的用户提示和二次确认
- 实时监控价格差异率
- 建立价格变动审计日志

```
价格一致性的三个关键：
1. 快照锁定（PDP生成，创单验证）
2. 差异提示（明确告知原因和差异）
3. 用户确认（价格上涨必须二次确认）
```

### 8.4 适用场景

**✅ 适合使用DDD的场景**：
- 业务逻辑复杂（计价、促销、订单、风控）
- 需求频繁变化
- 需要长期维护
- 团队规模较大（5人以上）

**❌ 不适合使用DDD的场景**：
- 简单CRUD系统
- 技术型系统（日志、监控）
- 短期项目（< 3个月）
- 单人开发

### 8.5 最后的话

> DDD不是银弹，但它是应对业务复杂性的有效方法。
>
> 最重要的不是掌握所有DDD概念，而是学会从业务出发，建立清晰的领域模型，让代码真正反映业务。

---

## 九、参考资料

### 经典书籍
1. 《领域驱动设计》 Eric Evans
2. 《实现领域驱动设计》 Vaughn Vernon
3. 《企业应用架构模式》 Martin Fowler

### 在线资源
1. Martin Fowler 的博客：https://martinfowler.com/
2. Domain-Driven Design Community：https://dddcommunity.org/
3. DDD Reference：http://domainlanguage.com/ddd/reference/

### 相关文章
1. [电商系统价格计算引擎设计与实现](./24-pricing-engine-design.md) — 系统架构、场景分析、核心实现
2. The Clean Architecture - Robert C. Martin
3. Hexagonal Architecture - Alistair Cockburn
4. Bounded Context - Martin Fowler

---

**写于 2026年3月14日**  
**作者：后端架构师**

*领域驱动设计：让软件真正反映业务*

---

> **系列导航**
> 计价引擎的工程实现细节（多级缓存、降级策略等），详见[（四）计价引擎](/system-design/23-ecommerce-pricing-engine/)。
