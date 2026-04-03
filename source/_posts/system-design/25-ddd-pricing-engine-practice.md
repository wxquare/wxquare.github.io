---
title: 领域驱动设计DDD在计价引擎中的实践
date: 2026-03-14
categories:
- 系统设计
tags:
- DDD
- 领域驱动设计
- 计价引擎
- 架构设计
toc: true
---

<!-- toc -->

## 一、背景与挑战

### 1.1 业务背景

在电商平台中，价格计算是最核心的业务能力之一。一个商品的最终价格并非简单的"标价"，而是由多个因素叠加计算而成：

```
最终支付价格 = 基础价格
              - 营销活动折扣（秒杀、新用户价、捆绑价）
              + 平台服务费用（服务费、手续费、附加费）
              - 用户优惠抵扣（优惠券、积分、支付优惠）
```

### 1.2 系统复杂性挑战

在构建数字商品计价系统的过程中，我们面临着与美团营销系统类似的挑战：

**1. 隐晦性**

- **抽象层面**：同一个"价格"概念，在不同场景下含义不同
  - PDP展示价：用户看到的价格
  - 订单价：创建订单时的价格
  - 支付价：最终扣款价格
- **实现层面**：代码中的`MarketPrice`、`DiscountPrice`、`PromotionPrice`、`FinalPrice` 等概念混乱

**2. 耦合性**

- **代码层面**：价格计算逻辑散落在8个品类（EVoucher、GiftCard、MovieTicket等）的225个文件中
- **模块层面**：计价依赖促销服务、商品服务、用户服务
- **系统层面**：前端展示价、订单价、支付价不一致导致资损

**3. 变化性**

- **业务需求**：促销规则频繁变化（每周调整优先级、新增促销类型）
- **品类扩展**：新品类接入需要2周时间
- **地区差异**：不同地区有不同的定价策略

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

**问题**：
- ❌ 业务逻辑分散在各个函数中，难以理解整体流程
- ❌ 缺乏业务概念的抽象，只有数据和操作
- ❌ 新增促销类型需要修改核心计算逻辑
- ❌ 无法支持复杂的业务规则（如BundlePrice）

---

## 二、DDD核心概念

在介绍计价引擎的DDD实践之前，先回顾一下DDD的核心概念。

### 2.1 什么是领域？

领域由三部分组成：

```
┌─────────────────────────────────────────┐
│              领域（Domain）              │
├─────────────────────────────────────────┤
│                                          │
│  涉众域 (Stakeholders)                   │
│    └─ 用户：商家、运营、用户             │
│                                          │
│  问题域 (Problem Space)                  │
│    └─ 业务价值：如何定价？如何促销？     │
│                                          │
│  解决方案域 (Solution Space)             │
│    └─ 解决方案：四层计价模型             │
│                                          │
└─────────────────────────────────────────┘
```

**计价领域**：
- **涉众域**：商家（定价）、运营（促销）、用户（消费）、财务（结算）
- **问题域**：如何准确计算价格？如何支持多种促销？如何保证前后端一致？
- **解决方案域**：统一的计价模型、规则引擎、价格快照

### 2.2 什么是领域驱动设计？

> 针对特定业务领域，用户在面对业务问题时有对应的解决方案，这些问题与方案构成了领域知识，它包含流程、规则以及处理问题的方法。领域驱动设计就是围绕这些知识来设计系统。

**计价领域知识**：
- **流程**：PDP展示 → 订单创建 → 支付结算
- **规则**：促销优先级、费用计算、优惠抵扣
- **方法**：四层计价模型、价格快照

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
│    └─ 查看销售报表                                   │
│                                                      │
│  运营 (Operator)                                     │
│    ├─ 创建促销活动                                   │
│    ├─ 配置优惠券                                     │
│    └─ 调整费用规则                                   │
│                                                      │
│  用户 (Customer)                                     │
│    ├─ 查看商品价格 (PDP)                             │
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

从用例中抽取概念，建立统一语言：

#### 基础价格术语

| 中文术语 | 英文术语 | Term | 含义 |
|---------|---------|------|------|
| 市场原价 | Market Price | `market_price` | 商品的市场标价，来自Hub/Supplier |
| 折扣价 | Discount Price | `discount_price` | 平台日常销售价，可能包含长期折扣 |
| 原始折扣价 | Original Discount Price | `original_discount_price` | 用于活动降级时的备份价格 |

#### 促销术语

| 中文术语 | 英文术语 | Term | 含义 |
|---------|---------|------|------|
| 促销价 | Promotion Price | `promotion_price` | 参与促销活动后的价格 |
| 秒杀价 | Flash Sale Price | `flash_sale_price` | 限时秒杀活动价格 |
| 新用户价 | New User Price | `new_user_price` | 新用户专享价格 |
| 捆绑价 | Bundle Price | `bundle_price` | 买N件享优惠的价格 |

#### 费用术语

| 中文术语 | 英文术语 | Term | 含义 |
|---------|---------|------|------|
| 平台服务费 | Admin Fee | `admin_fee` | Hub平台收取的服务费 |
| 附加费用 | Additional Charge | `additional_charge` | DP平台收取的额外费用 |
| 手续费 | Handling Fee | `handling_fee` | 支付渠道手续费 |

#### 最终价格术语

| 中文术语 | 英文术语 | Term | 含义 |
|---------|---------|------|------|
| 计价金额 | Pricing Amount | `pricing_amount` | 单个SKU的计价金额（含数量） |
| 最终价格 | Final Price | `final_price` | 用户最终支付价格 |
| 结算金额 | Checkout Amount | `checkout_amount` | 结算时的金额 |

**统一语言的应用**：

```go
// ✅ 正确：使用统一语言
type PriceEntity struct {
    MarketPrice           int64  // 市场原价
    DiscountPrice         int64  // 折扣价
    PromotionPrice        int64  // 促销价
    AdminFee              int64  // 平台服务费
    FinalPrice            int64  // 最终价格
}

// ❌ 错误：术语混乱
type Price struct {
    OriginalPrice  int64  // 原价？市场价？
    SalePrice      int64  // 售价？折扣价？促销价？
    ActualPrice    int64  // 实际价格？最终价格？
    PlatformFee    int64  // 平台费？服务费？
    TotalPrice     int64  // 总价？最终价？
}
```

### 3.3 概念模型（Concept Model）

基于统一语言，建立概念模型：

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
         └─────────────────────┘  │  • BundlePrice  │
                                  └─────────────────┘
                    │
         ┌──────────┴──────────┐
         │                     │
    ┌────▼────────┐      ┌────▼──────────┐
    │    Fee      │      │   Discount    │
    │   (费用)     │      │   (优惠)      │
    │ • AdminFee  │      │ • Voucher     │
    │ • Handling  │      │ • Coins       │
    └─────────────┘      └───────────────┘
         │                     │
         └──────────┬──────────┘
                    │
              ┌─────▼─────┐
              │FinalPrice │
              │ (最终价格) │
              └───────────┘
```

**概念关系**：
- **PriceEntity** 包含 (1:1) **BasePrice**
- **PriceEntity** 可能包含 (1:0..N) **Promotion**
- **PriceEntity** 可能包含 (1:0..N) **Fee**
- **Order** 可能包含 (1:0..N) **Discount**
- **PriceEntity** 产生 (1:1) **FinalPrice**

### 3.4 子域划分（Subdomain）

参考美团的拆分思路，我们基于**问题域**进行拆分：

```
┌─────────────────────────────────────────────────────────┐
│              计价领域 (Pricing Domain)                   │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  核心子域 (Core Subdomain)                               │
│  ┌───────────────────────────────────────────────┐     │
│  │  定价域 (Pricing Subdomain)                    │     │
│  │  • 问题：如何计算价格？                        │     │
│  │  • 方案：四层计价模型                          │     │
│  │  • 职责：价格计算、校验、快照                  │     │
│  └───────────────────────────────────────────────┘     │
│                                                          │
│  支撑子域 (Supporting Subdomain)                         │
│  ┌──────────────────┐  ┌──────────────────┐           │
│  │  促销域           │  │  商品域           │           │
│  │  (Promotion)     │  │  (Item)          │           │
│  │  • 促销规则管理   │  │  • 商品信息      │           │
│  │  • 活动配置      │  │  • 库存管理      │           │
│  └──────────────────┘  └──────────────────┘           │
│                                                          │
│  ┌──────────────────┐  ┌──────────────────┐           │
│  │  用户域           │  │  支付域           │           │
│  │  (User)          │  │  (Payment)       │           │
│  │  • 用户信息      │  │  • 支付方式      │           │
│  │  • 用户分群      │  │  • 优惠券        │           │
│  └──────────────────┘  └──────────────────┘           │
│                                                          │
│  通用子域 (Generic Subdomain)                            │
│  ┌──────────────────┐  ┌──────────────────┐           │
│  │  缓存域           │  │  配置域           │           │
│  │  (Cache)         │  │  (Config)        │           │
│  └──────────────────┘  └──────────────────┘           │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

**拆分原则**：

1. **定价域（核心）**：专注价格计算逻辑，这是业务的核心竞争力
2. **促销域（支撑）**：管理促销规则，为定价提供数据
3. **商品域（支撑）**：提供商品基础信息
4. **用户域（支撑）**：提供用户信息和分群
5. **支付域（支撑）**：处理支付和优惠券
6. **缓存/配置域（通用）**：技术基础设施

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

**防腐层示例**：

```go
// 定价域不直接依赖促销域的数据结构
// 通过防腐层转换

// ❌ 错误：直接依赖外部服务的数据结构
type PricingService struct {
    promotionClient *promotion.Client
}

func (s *PricingService) GetPromotionPrice(itemID int64) int64 {
    // 直接使用外部结构
    promoData := s.promotionClient.GetPromotion(itemID)
    return promoData.Price  // 耦合到外部结构
}

// ✅ 正确：通过防腐层转换
type PricingService struct {
    promotionAdapter PromotionAdapter  // 防腐层
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

// 防腐层实现
type PromotionAdapterImpl struct {
    promotionClient *promotion.Client
}

func (a *PromotionAdapterImpl) GetPromotionInfo(itemID int64) *PromotionInfo {
    // 外部数据转换为领域模型
    externalData := a.promotionClient.GetPromotion(itemID)
    
    return &PromotionInfo{
        ActivityID: externalData.PromotionActivityId,
        Price:      externalData.RewardPrice,
        Type:       convertPromotionType(externalData.Type),
    }
}
```

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
    Status PriceStatus  // Draft, Active, Expired
    
    // 属性
    basePrice    *BasePrice    // 值对象
    promotions   []Promotion   // 值对象集合
    fees         []Fee         // 值对象集合
    
    // 行为
    ApplyPromotion(promo Promotion) error
    AddFee(fee Fee) error
    Calculate() int64
}

// 实体的行为
func (e *PriceEntity) ApplyPromotion(promo Promotion) error {
    // 业务规则：促销价不能高于折扣价
    if promo.Price > e.basePrice.DiscountPrice {
        return errors.New("promotion price cannot exceed discount price")
    }
    
    e.promotions = append(e.promotions, promo)
    return nil
}
```

#### 值对象：无唯一标识，不可变

```go
// ✅ 值对象：Price（价格）
type Price struct {
    amount   int64
    currency string
}

// 不可变：所有操作返回新对象
func NewPrice(amount int64, currency string) Price {
    if amount < 0 {
        panic("price cannot be negative")
    }
    return Price{amount: amount, currency: currency}
}

func (p Price) Add(other Price) Price {
    if p.currency != other.currency {
        panic("currency mismatch")
    }
    return Price{
        amount:   p.amount + other.amount,
        currency: p.currency,
    }
}

func (p Price) Multiply(factor int64) Price {
    return Price{
        amount:   p.amount * factor,
        currency: p.currency,
    }
}

// ✅ 值对象：Promotion（促销）
type Promotion struct {
    activityID int64
    name       string
    price      Price
    rewardType RewardType
}

// 值对象是不可变的
func (p Promotion) WithPrice(newPrice Price) Promotion {
    return Promotion{
        activityID: p.activityID,
        name:       p.name,
        price:      newPrice,  // 返回新对象
        rewardType: p.rewardType,
    }
}
```

### 4.2 聚合根（Aggregate Root）

**聚合根设计原则**：
1. 满足业务一致性（促销、费用、价格必须一致）
2. 满足数据完整性（不存在没有基础价格的实体）
3. 技术限制（避免加载过大数据）

#### 计价聚合根设计

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
}

// 聚合根的行为（封装业务规则）
func (a *PricingAggregate) CalculatePrice() (*PricingResult, error) {
    // 1. 业务规则校验
    if err := a.validate(); err != nil {
        return nil, err
    }
    
    // 2. 选择促销
    promotion := a.selectPromotion()
    
    // 3. 计算基础价格
    baseAmount := a.calculateBaseAmount(promotion)
    
    // 4. 计算费用
    feeAmount := a.calculateFees()
    
    // 5. 应用优惠
    discountAmount := a.applyDiscounts()
    
    // 6. 计算最终价格
    finalPrice := baseAmount + feeAmount - discountAmount
    
    // 7. 生成领域事件
    a.addEvent(&PriceCalculatedEvent{
        AggregateID: a.id,
        FinalPrice:  finalPrice,
        Timestamp:   time.Now(),
    })
    
    return &PricingResult{
        FinalPrice: finalPrice,
        Breakdown:  a.buildBreakdown(),
    }, nil
}

// 业务规则：选择促销（优先级）
func (a *PricingAggregate) selectPromotion() *Promotion {
    // 规则1：秒杀优先
    if flashSale := a.getFlashSale(); flashSale != nil {
        return flashSale
    }
    
    // 规则2：新用户价
    if newUserPrice := a.getNewUserPrice(); newUserPrice != nil {
        return newUserPrice
    }
    
    // 规则3：默认折扣价
    return a.getDefaultPrice()
}

// 业务规则校验
func (a *PricingAggregate) validate() error {
    // 规则1：市场价 > 0
    if a.priceEntity.basePrice.MarketPrice <= 0 {
        return errors.New("market price must be positive")
    }
    
    // 规则2：折扣价 <= 市场价
    if a.priceEntity.basePrice.DiscountPrice > a.priceEntity.basePrice.MarketPrice {
        return errors.New("discount price cannot exceed market price")
    }
    
    return nil
}
```

### 4.3 领域服务（Domain Service）

**什么时候使用领域服务**？

不适合放在聚合根里的领域逻辑，可以放在领域服务中：

```go
// ❌ 不适合放在聚合根：跨聚合根的逻辑
// 例：选择优先级最高的促销活动

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
    highestPriority := 999
    
    for _, promo := range promotions {
        if !s.isApplicable(promo, context) {
            continue
        }
        
        priority := s.getPromotionPriority(promo)
        if priority < highestPriority {
            highestPriority = priority
            bestPromotion = promo
        }
    }
    
    return bestPromotion
}

// 领域服务：BundlePrice计算（EVoucher专用）
type BundlePriceCalculator struct{}

func (c *BundlePriceCalculator) Calculate(
    basePrice int64,
    quantity int64,
    config *BundleConfig,
) int64 {
    curRound := quantity / config.MinQuantity
    effectiveRounds := min(config.MaxRound, curRound)
    
    var discountPerRound int64
    switch config.RewardType {
    case DiscountPercentage:
        discountPerRound = basePrice * config.MinQuantity * config.Discount / 100000
    case ReductionPrice:
        discountPerRound = config.Discount
    case BundlePriceFixedAmount:
        discountPerRound = basePrice * config.MinQuantity - config.Discount
    }
    
    return basePrice * quantity - discountPerRound * effectiveRounds
}
```

### 4.4 贫血模型 vs 充血模型

**我们的选择：混合模式**

```go
// ✅ 核心领域逻辑：充血模型
type PricingAggregate struct {
    id          string
    priceEntity *PriceEntity
    
    // 富含业务逻辑的方法
    func (a *PricingAggregate) CalculatePrice() (*PricingResult, error)
    func (a *PricingAggregate) ApplyPromotion(promo *Promotion) error
    func (a *PricingAggregate) Validate() error
}

// ✅ 简单CRUD：贫血模型
type PriceSnapshot struct {
    ID         int64
    OrderID    int64
    Price      int64
    CreatedAt  time.Time
}

// 简单的数据访问对象
type PriceSnapshotRepository interface {
    Save(snapshot *PriceSnapshot) error
    FindByOrderID(orderID int64) (*PriceSnapshot, error)
}
```

**选择标准**：
- 核心业务逻辑（价格计算）→ 充血模型
- 简单CRUD（数据存储）→ 贫血模型
- 避免过度设计

### 4.5 体现业务语义的代码

```go
// ❌ 错误：没有业务含义
func (a *PricingAggregate) UpdateStatus(status int) error {
    a.status = status
    return nil
}

// ✅ 正确：清晰的业务语义
func (a *PricingAggregate) SubmitForReview() error {
    if a.status != Draft {
        return errors.New("only draft pricing can be submitted")
    }
    
    a.status = PendingReview
    a.addEvent(&PricingSubmittedEvent{...})
    return nil
}

func (a *PricingAggregate) Approve() error {
    if a.status != PendingReview {
        return errors.New("only pending pricing can be approved")
    }
    
    a.status = Approved
    a.addEvent(&PricingApprovedEvent{...})
    return nil
}

func (a *PricingAggregate) Reject(reason string) error {
    if a.status != PendingReview {
        return errors.New("only pending pricing can be rejected")
    }
    
    a.status = Rejected
    a.rejectionReason = reason
    a.addEvent(&PricingRejectedEvent{Reason: reason})
    return nil
}
```

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
    │  │  • CalPDPPrice()                     │   │
    │  │  • CalOrderPrice()                   │   │
    │  │  • CalPayPrice()                     │   │
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
    │  │ MySQL    │  │ Redis    │  │ RPC      │  │
    │  │ Repo     │  │ Cache    │  │ Client   │  │
    │  └──────────┘  └──────────┘  └──────────┘  │
    └─────────────────────────────────────────────┘
```

### 5.2 实际代码结构

```
pricing-server/
├── cmd/
│   └── main.go                          # 启动入口
│
├── internal/
│   ├── adapter/                         # 适配器层
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
│   ├── domain/                          # 领域层
│   │   ├── model/
│   │   │   ├── aggregate/              # 聚合根
│   │   │   │   └── pricing_aggregate.go
│   │   │   ├── entity/                 # 实体
│   │   │   │   └── price_entity.go
│   │   │   └── valueobject/            # 值对象
│   │   │       ├── price.go
│   │   │       └── promotion.go
│   │   ├── service/                    # 领域服务
│   │   │   ├── promotion_selector.go
│   │   │   └── bundle_calculator.go
│   │   ├── repository/                 # 仓储接口
│   │   │   └── pricing_repository.go
│   │   └── event/                      # 领域事件
│   │       └── price_calculated.go
│   │
│   └── infrastructure/                  # 基础设施层
│       ├── persistence/                 # 持久化
│       │   ├── mysql/
│       │   │   └── pricing_repo_impl.go
│       │   └── redis/
│       │       └── cache_impl.go
│       ├── rpc/                        # RPC客户端
│       │   ├── promotion_client.go
│       │   └── item_client.go
│       └── config/
│           └── config.go
│
└── common/                              # 共享代码
    ├── errors/
    └── utils/
```

### 5.3 依赖方向

```
核心原则：依赖倒置（Dependency Inversion Principle）

外层 ───依赖───> 内层
适配器层 ───> 应用层 ───> 领域层 <─── 基础设施层
                              (通过接口)

具体实现：

┌────────────────────────────────────────────────────┐
│  HTTP Handler (适配器层)                           │
│    ↓ 调用                                          │
│  PricingApplicationService (应用层)                │
│    ↓ 调用                                          │
│  PricingAggregate (领域层)                         │
│    ↓ 依赖接口                                      │
│  PricingRepository interface (领域层定义)          │
│    ↑ 实现                                          │
│  MySQLPricingRepository (基础设施层)               │
└────────────────────────────────────────────────────┘
```

### 5.4 实际代码示例

#### 领域层：定义接口

```go
// domain/repository/pricing_repository.go

package repository

// 仓储接口（由领域层定义）
type PricingRepository interface {
    Save(aggregate *aggregate.PricingAggregate) error
    FindByID(id string) (*aggregate.PricingAggregate, error)
    FindByItemID(itemID int64) (*aggregate.PricingAggregate, error)
}

// 促销适配器接口（防腐层）
type PromotionAdapter interface {
    GetPromotionInfo(itemID int64) (*valueobject.Promotion, error)
    ListPromotions(itemIDs []int64) ([]*valueobject.Promotion, error)
}
```

#### 基础设施层：实现接口

```go
// infrastructure/persistence/mysql/pricing_repo_impl.go

package mysql

import "pricing-server/internal/domain/repository"

type MySQLPricingRepository struct {
    db *sql.DB
}

// 实现领域层定义的接口
func (r *MySQLPricingRepository) Save(
    aggregate *aggregate.PricingAggregate,
) error {
    // 将聚合根转换为数据库模型
    dbModel := r.aggregateToModel(aggregate)
    
    // 保存到数据库
    _, err := r.db.Exec("INSERT INTO pricing ...", dbModel)
    return err
}

func (r *MySQLPricingRepository) FindByID(id string) (*aggregate.PricingAggregate, error) {
    // 从数据库查询
    row := r.db.QueryRow("SELECT * FROM pricing WHERE id = ?", id)
    
    // 转换为聚合根
    return r.modelToAggregate(row)
}
```

#### 应用层：协调

```go
// application/service/pricing_service.go

package service

type PricingApplicationService struct {
    // 依赖领域层接口
    pricingRepo      repository.PricingRepository
    promotionAdapter repository.PromotionAdapter
    
    // 领域服务
    promotionSelector *service.PromotionSelectionService
}

func (s *PricingApplicationService) CalPDPPrice(
    req *dto.CalPDPPriceRequest,
) (*dto.CalPDPPriceResponse, error) {
    // 1. 获取促销信息（通过适配器）
    promotions, err := s.promotionAdapter.ListPromotions(req.ItemIDs)
    if err != nil {
        return nil, err
    }
    
    // 2. 构建聚合根
    aggregate := aggregate.NewPricingAggregate(req.ItemID)
    aggregate.SetPromotions(promotions)
    
    // 3. 执行领域逻辑
    result, err := aggregate.CalculatePrice()
    if err != nil {
        return nil, err
    }
    
    // 4. 转换为DTO返回
    return s.toDTO(result), nil
}
```

#### 适配器层：处理请求

```go
// adapter/http/handler.go

package http

type PricingHandler struct {
    pricingService *service.PricingApplicationService
}

func (h *PricingHandler) CalPDPPrice(c *gin.Context) {
    // 1. 解析HTTP请求
    var httpReq struct {
        ItemIDs []int64 `json:"item_ids"`
        UserID  int64   `json:"user_id"`
    }
    if err := c.ShouldBindJSON(&httpReq); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // 2. 转换为应用层DTO
    appReq := &dto.CalPDPPriceRequest{
        ItemIDs: httpReq.ItemIDs,
        UserID:  httpReq.UserID,
    }
    
    // 3. 调用应用服务
    resp, err := h.pricingService.CalPDPPrice(appReq)
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

| 指标 | DDD之前 | DDD之后 | 改善幅度 |
|------|---------|---------|---------|
| **代码重复率** | 40% | 15% | -62% |
| **新品类接入时间** | 2周 | 3天 | -86% |
| **单元测试覆盖率** | 45% | 90% | +100% |
| **Bug密度** | 3.5/KLOC | 0.8/KLOC | -77% |
| **需求变更响应时间** | 3天 | 0.5天 | -83% |

### 6.2 质量收益

**1. 概念清晰**

```
改造前：
- 10种不同的"价格"命名
- 业务和技术对话困难
- 新人需要1周理解代码

改造后：
- 统一的术语标准（50+术语）
- 业务文档和代码一致
- 新人3天上手
```

**2. 业务规则显式化**

```go
// ❌ 改造前：隐藏在代码中的规则
if price < 100 && user.registerDays <= 7 {
    // 什么规则？为什么这么做？
}

// ✅ 改造后：显式的业务规则
func (s *PromotionSelectionService) IsNewUserEligible(
    user *User,
) bool {
    return user.RegisterDays <= NewUserMaxDays &&
           user.OrderCount == 0
}
```

**3. 易于测试**

```go
// ✅ 聚合根可以独立测试
func TestPricingAggregate_CalculatePrice(t *testing.T) {
    aggregate := NewPricingAggregate("test-001")
    aggregate.SetBasePrice(NewPrice(100000, "IDR"))
    aggregate.AddPromotion(NewFlashSalePromotion(80000))
    
    result, err := aggregate.CalculatePrice()
    
    assert.NoError(t, err)
    assert.Equal(t, int64(80000), result.FinalPrice)
}
```

### 6.3 维护性收益

**1. 修改影响范围可控**

```
需求变更：秒杀优先级临时提升

改造前：
- 修改base_service.go（核心逻辑）
- 影响所有品类
- 测试所有场景
- 风险高

改造后：
- 修改promotion_selector.go（领域服务）
- 只影响促销选择逻辑
- 单元测试覆盖
- 风险低
```

**2. 业务变化适应性强**

```
新需求：增加VIP专享价

改造前：
- 修改多个if-else分支
- 影响现有促销逻辑
- 容易引入bug

改造后：
- 新增VIPPriceRule类
- 注册到规则引擎
- 不影响现有规则
```

---

## 七、常见误区与最佳实践

### 7.1 常见误区

#### 误区1：深陷DDD概念

```go
// ❌ 过度设计：生搬硬套概念
type PriceValueObject struct {
    value int64
}

type PriceEntity struct {
    id           string
    priceVO      *PriceValueObject  // 过度抽象
    domainEvents []DomainEvent      // 不必要的复杂性
}

// ✅ 简单实用：根据需要选择
type Price struct {
    amount   int64
    currency string
}
```

#### 误区2：试图一次性设计完美

```
❌ 错误：花3个月设计"完美"的领域模型

✅ 正确：
- Week 1-2: 基础模型（80%场景）
- Week 3-4: 迭代优化
- Week 5-6: 新增场景
- 持续迭代
```

#### 误区3：忽视统一语言

```
❌ 错误：
- 技术：MarketPrice、DiscountPrice
- 产品：原价、折扣价
- 业务：市场价、会员价

✅ 正确：
- 所有人：市场原价、折扣价
- 文档、代码、讨论统一使用
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
```

#### 2. 持续迭代

```
领域模型不是一次性设计出来的

Week 1-2: 基础模型
  └─ 支持80%场景

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
│  产品：提供业务视角                   │
│  开发：提供技术实现                   │
│  测试：提供边界场景                   │
│  运营：提供实际问题                   │
└──────────────────────────────────────┘
```

#### 4. 适度抽象

```go
// ✅ 简单场景：简单实现
type PriceSnapshot struct {
    OrderID int64
    Price   int64
}

func SaveSnapshot(snapshot *PriceSnapshot) error {
    // 简单CRUD
}

// ✅ 复杂场景：领域模型
type PricingAggregate struct {
    // 复杂业务逻辑
}

func (a *PricingAggregate) CalculatePrice() {
    // 封装复杂规则
}
```

---

## 八、总结

### 8.1 DDD的核心价值

1. **统一语言**：消除沟通障碍
2. **领域模型**：业务知识显式化
3. **分层架构**：职责清晰，易于维护
4. **持续迭代**：适应业务变化

### 8.2 实施要点

```
战略设计：
  ├─ 用例分析（理解业务）
  ├─ 统一语言（概念抽取）
  ├─ 概念模型（关系梳理）
  └─ 子域划分（化繁为简）

战术设计：
  ├─ 实体与值对象（映射概念）
  ├─ 聚合根（封装规则）
  ├─ 领域服务（跨聚合逻辑）
  └─ 业务语义（代码可读）

代码架构：
  ├─ 六边形架构（依赖倒置）
  ├─ 分层清晰（职责单一）
  └─ 接口隔离（易于测试）
```

### 8.3 给后来者的建议

**1. 不要害怕DDD的概念体系**

```
从简单开始：
- 第1周：统一语言
- 第2周：概念模型
- 第3周：简单聚合根
- 第4周：逐步完善
```

**2. 重视统一语言**

```
没有统一语言 → 没有概念模型 → 没有好的代码

投入时间在统一语言上，回报率最高
```

**3. 持续迭代**

```
模型是相对稳定的，但不是一成不变的

业务理解加深 → 模型调整
抽象角度变化 → 模型重构
业务需求变化 → 模型扩展
```

**4. 团队协作**

```
领域驱动设计是团队工作

产品 + 开发 + 测试 + 运营
    = 完整的领域知识
```

### 8.4 适用场景

**✅ 适合使用DDD的场景**：
- 业务逻辑复杂（计价、促销、订单）
- 需求频繁变化
- 需要长期维护
- 团队规模较大

**❌ 不适合使用DDD的场景**：
- 简单CRUD系统
- 技术型系统（缓存、日志）
- 短期项目
- 单人开发

---

## 九、参考资料

### 书籍
1. 《领域驱动设计》 Eric Evans
2. 《实现领域驱动设计》 Vaughn Vernon
3. 《企业应用架构模式》 Martin Fowler

### 文章
1. 美团技术团队：《领域驱动设计在营销系统的实践》
2. Martin Fowler：《BoundedContext》
3. Robert C. Martin：《The Clean Architecture》

### 项目实践
1. [计价引擎设计文档](./24-pricing-engine-design.md)
2. [计价引擎项目复盘](./24-pricing-engine-retrospective.md)
3. [业务规则引擎设计](./16-rule-engine-design.md)

---

**写于 2026年3月14日**  
**作者：资深后端工程师，25年经验**  
**项目周期：2024年10月 - 2025年5月（7个月）**

*领域驱动设计不是银弹，但它是应对业务复杂性的有效方法。*
