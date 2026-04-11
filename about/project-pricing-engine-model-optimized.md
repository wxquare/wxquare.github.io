# 电商价格计算引擎 - 统一价格模型（基于实际项目优化）

> 本文档基于 Digital Purchase Service 实际项目代码优化，统一价格术语和计算模型。

---

## 一、价格术语统一体系

### 1.1 核心问题

**实际项目中的术语混乱**：
- 代码中：`MarketPrice`、`DiscountPrice`、`OriginalDiscountPrice`、`PricingAmount`、`SellingPrice`
- 产品文档：市场价、原价、划线价、折扣价、活动价、售卖价、应付金额
- 各团队理解不一致，导致需求理解偏差 20%+

### 1.2 统一价格术语表（基于实际代码）

| 层级 | 术语（中文） | 术语（英文） | 字段名（代码） | 定义 | 适用场景 |
|-----|------------|------------|--------------|------|---------|
| **Layer 1: 基础价格** | | | | | |
| | 市场原价 | Market Price | `market_price` | 商品的标价，通常用于划线展示 | PDP、Cart |
| | 日常折扣价 | Discount Price | `discount_price` | 商品的日常销售价，无促销时的售价 | PDP、Cart、Order |
| | 原始折扣价 | Original Discount Price | `original_discount_price` | 用于计算的原始折扣价（兜底字段） | PDP、Order |
| **Layer 2: 促销价格** | | | | | |
| | 促销价 | Promotion Price | `promotion_price` | 应用促销后的价格（FlashSale/NewUserPrice/BundlePrice） | PDP、Cart、Order |
| | 促销优惠金额 | Promotion Discount Amount | `promotion_discount_amount` | 促销节省的金额 = 折扣价 - 促销价 | PDP展示 |
| | 计价金额 | Pricing Amount | `pricing_amount` | 商品最终计价金额（含促销，可能含费用） | Order、Payment |
| **Layer 3: 费用** | | | | | |
| | 平台服务费 | Hub Admin Fee | `hub_admin_fee` | 平台收取的服务费 | Order、Payment |
| | 附加费用 | Additional Charge | `dp_additional_charge` | 数字商品附加费用 | Order、Payment |
| | 服务费 | Service Fee | `service_fee` | 服务手续费 | Payment |
| | 手续费 | Handling Fee | `handling_fee` | 支付手续费 | Payment |
| | 加价费 | Mark Up Fee | `mark_up` | 供应商加价 | Order、Payment |
| | 附加费（Home Service） | Sub Charge | `sub_charge` | 最低交易金额补足费用 | Order（Home Service） |
| **Layer 4: 优惠抵扣** | | | | | |
| | 优惠券抵扣 | Voucher Redeemed | `voucher_redeemed` | 优惠券抵扣金额 | Checkout、Payment |
| | 积分抵扣 | Coins Redeemed | `coins_redeemed` | 积分抵扣金额 | Checkout、Payment |
| **Final: 最终价格** | | | | | |
| | 订单总额 | Order Total Amount | `order_total_amount` | 用户最终需要支付的金额 | Order、Checkout、Payment |
| | 售卖价格信息 | Selling Price Info | `selling_price_info` | 售卖价格结构体（包含selling_price等） | Order |

### 1.3 价格维度说明

**三个核心维度**：

#### 维度1: 单价 vs 总价

| 维度 | 说明 | 代码表现 | 适用场景 |
|-----|------|---------|---------|
| **单价（Unit Price）** | 单个商品的价格 | `market_price`（不含quantity） | PDP展示单价 |
| **商品总价（Item Total）** | 单价 × 数量 | `pricing_amount = price × quantity` | Cart、Order |
| **订单总价（Order Total）** | 所有商品总价 + 费用 - 优惠 | `sum(pricing_amount) + fees - discounts` | Order、Checkout |

#### 维度2: 计算阶段

| 阶段 | 说明 | 代码表现 | 适用场景 |
|-----|------|---------|---------|
| **促销前（Before Promotion）** | 使用 `discount_price` | `discount_price × quantity` | PDP对比展示 |
| **促销后（After Promotion）** | 使用 `promotion_price` | `promotion_price × quantity` | PDP、Cart、Order |
| **加费用后（After Fee）** | 促销价 + 各类费用 | `pricing_amount + fees` | Order |
| **优惠后（After Discount）** | 减去券/积分 | `pricing_amount - voucher - coins` | Checkout、Payment |

#### 维度3: 场景差异

| 场景 | 设置字段 | 计算内容 | 缓存策略 |
|-----|---------|---------|---------|
| **CalPDPPrice** | `pricing_amount`<br>`pdp_other_price_map` | 基础价 + 促销价<br>（不含SellingPriceInfo） | 高缓存 |
| **CalOrderPrice** | `pricing_amount`<br>`selling_price_info`<br>`extra_info` | 基础价 + 促销价 + 费用<br>（含SellingPriceInfo） | 不缓存 |
| **CalPayPrice** | 无（使用快照） | 快照验证（零计算） | 不缓存 |

---

## 二、实际项目中的价格计算模型

### 2.1 通用四层计价模型

基于实际代码提炼的统一模型：

```go
// ========== Layer 1: 基础价格 ==========
marketPrice := item.GetMarketPrice()           // 市场原价
discountPrice := item.GetDiscountPrice()       // 日常折扣价
originalDiscountPrice := item.GetOriginalDiscountPrice()  // 原始折扣价（兜底）
if originalDiscountPrice == 0 {
    originalDiscountPrice = discountPrice       // 兜底逻辑（重要！）
}

// ========== Layer 2: 促销价格 ==========
var promotionPrice int64 = discountPrice       // 默认使用折扣价

// 2.1 FlashSale（秒杀）
if hasFlashSale(item) {
    promotionPrice = getFlashSalePrice(item)
    if quantity > purchaseLimit {
        // 超限部分使用原价
        pricingAmount = promotionPrice * purchaseLimit + 
                       originalDiscountPrice * (quantity - purchaseLimit)
    } else {
        pricingAmount = promotionPrice * quantity
    }
}

// 2.2 BundlePrice（套餐价）
if hasBundlePrice(item) {
    // 满 N 件享折扣
    rounds := quantity / minQuantity
    discountPerRound := calculateDiscount(price, minQuantity, discount)
    totalDiscount := discountPerRound * rounds
    pricingAmount = (price * quantity) - totalDiscount
}

// 2.3 NewUserPrice（新用户价）
if hasNewUserPrice(item) {
    promotionPrice = getNewUserPrice(item)
    purchaseLimit := 1  // 硬编码限购1（实际代码）
    if quantity > purchaseLimit {
        pricingAmount = promotionPrice * purchaseLimit + 
                       originalDiscountPrice * (quantity - purchaseLimit)
    } else {
        pricingAmount = promotionPrice * quantity
    }
}

// 2.4 无促销
if !hasPromotion(item) {
    pricingAmount = originalDiscountPrice * quantity
}

// ========== Layer 3: 费用（CalOrderPrice 场景） ==========
// 注意：CalPDPPrice 不计算 SellingPriceInfo
if scene == SceneCalOrderPrice {
    FillSellingPriceInfo(item, hubAdminFee, dpAdditionalCharge, markUp)
    // SellingPriceInfo 包含：
    // - selling_price: 售卖单价
    // - original_discount_price: 原始折扣价
    // - dp_additional_charge: 附加费用
    // - mark_up: 加价
}

// ========== Layer 4: 订单级优惠（Checkout 场景） ==========
orderTotal := sum(all_items.pricing_amount)
orderTotal = orderTotal - voucherRedeemed - coinsRedeemed

// 价格保护
if orderTotal < 0 {
    orderTotal = 0
}
```

### 2.2 Home Service 特殊计价模型

Home Service 品类有特殊的**最低交易金额（MinTransaction）**机制：

```go
// Home Service 独有逻辑
func (c *HomeServiceCalService) CalPDPPrice(req *pricing.CalPDPPriceRequest) {
    IncludedMinTransactionAmount := int64(0)   // 计入最低交易金额的商品
    ExcludeMinTransactionAmount := int64(0)    // 排除最低交易金额的商品
    
    for _, priceEntity := range req.PriceEntities {
        pricingAmount := priceEntity.GetMarketPrice() * priceEntity.GetQuantity()
        
        // 判断是否计入最低交易金额
        if priceEntity.GetSkuInfo().GetHomeServiceSkuInfo().GetExcludeMinimumValue() {
            ExcludeMinTransactionAmount += pricingAmount  // 排除
        } else {
            IncludedMinTransactionAmount += pricingAmount  // 计入
        }
    }
    
    // 计算 sub_charge（附加费用）
    subCharge := max(0, req.GetMinTransaction() - IncludedMinTransactionAmount)
    c.PdpOtherPriceMap["sub_charge"] = subCharge
    
    // 最终总价 = 排除商品 + max(最低交易金额, 计入商品)
    pricingTotalAmount := ExcludeMinTransactionAmount + 
                         max(req.GetMinTransaction(), IncludedMinTransactionAmount)
    
    return pricingTotalAmount, req.PriceEntities, errors.Ok
}
```

**Home Service 专有术语**：

| 术语（中文） | 术语（英文） | 字段名 | 定义 |
|------------|------------|-------|------|
| 最低交易金额 | Minimum Transaction | `min_transaction` | 订单必须达到的最低金额 |
| 计入最低交易金额 | Included in Min Transaction | `!exclude_minimum_value` | 该商品计入最低交易金额 |
| 排除最低交易金额 | Excluded from Min Transaction | `exclude_minimum_value=true` | 该商品不计入最低交易金额 |
| 附加费用 | Sub Charge | `sub_charge` | 为达到最低交易金额需要补足的金额 |

**计算示例**：

```go
// 场景：Home Service 订单，最低交易金额 60000
// 商品1（计入）: 35000 × 1 = 35000
// 商品2（排除）: 15000 × 2 = 30000

IncludedMinTransactionAmount = 35000
ExcludeMinTransactionAmount = 30000
minTransaction = 60000

// sub_charge = max(0, 60000 - 35000) = 25000
sub_charge = max(0, minTransaction - IncludedMinTransactionAmount)

// 最终总价 = 30000 + max(60000, 35000) = 30000 + 60000 = 90000
pricingTotalAmount = ExcludeMinTransactionAmount + 
                    max(minTransaction, IncludedMinTransactionAmount)
```

### 2.3 Event Ticket 促销计价模型

Event Ticket 支持多种促销类型，计算逻辑复杂：

```go
// Event Ticket 促销计算流程
func calculateEventFinalPrice(promotion, priceEntity) (originalAmount, finalAmount int64) {
    quantity := priceEntity.GetQuantity()
    
    // 兜底逻辑（重要！防止线上Bug）
    originalDiscountPrice := priceEntity.GetOriginalDiscountPrice()
    if originalDiscountPrice == 0 {
        originalDiscountPrice = priceEntity.GetDiscountPrice()
    }
    
    originalAmount := originalDiscountPrice * quantity
    
    // 无促销
    if promotion == nil || promotion.Type == UnknownActivityType {
        return originalAmount, originalAmount
    }
    
    // BundlePrice（套餐价）
    if promotion.Type == BundlePrice {
        return calculateEventBundlePrice(promotion, priceEntity)
    }
    
    // FlashSale/Scheduling/NewUserPrice（限购促销）
    if promotion.Type in [FlashSale, Scheduling, NewUserPrice] {
        return calculateEventPromoPrice(promotion, priceEntity)
    }
    
    return originalAmount, originalAmount
}
```

**Event Ticket 促销类型术语**：

| 促销类型 | 英文 | Proto 枚举 | 计算逻辑 | 限购逻辑 |
|---------|------|-----------|---------|---------|
| 秒杀 | Flash Sale | `PromotionActivityType_FlashSale` | 促销价 × 数量（超限部分原价） | 从 `PurchaseLimitInfo.AvailablePurchaseLimit` 获取 |
| 套餐价 | Bundle Price | `PromotionActivityType_BundlePrice` | 满N件享折扣（百分比或固定金额） | 考虑库存和限购 |
| 新用户价 | New User Price | `PromotionActivityType_NewUserPrice` | 新用户专享价（**硬编码限购1**） | 硬编码 `return 1` |
| 普通促销 | Scheduling | `PromotionActivityType_Scheduling` | 促销价 × 数量（超限部分原价） | 从 `PurchaseLimitInfo` 获取 |

**BundlePrice 计算细节**：

```go
// 套餐价：满3件20%折扣，购买6件
price := 1000000         // 单价
quantity := 6            // 数量
minQuantity := 3         // 满3件
discount := 20000        // 20%折扣（20000/100000）

// 印尼地区特殊处理（precision=0）
region := env.Region()
precision := getPrecision(region)  // 印尼: 0, 其他: 2

// 计算每轮折扣
beforeDiscountPrice := price * quantity  // 6000000
rounds := quantity / minQuantity         // 2 轮
pricePerRound := price * minQuantity     // 3000000

if rewardType == PercentageDiscount {
    discountAmount := (pricePerRound * discount) / 100000
    // 印尼: floor(600000 / 100000) × 100000 = 600000
    discountPerRound := mathFloorWithPrecision(discountAmount, precision)
    totalDiscount := discountPerRound * rounds  // 1200000
}

finalAmount := beforeDiscountPrice - totalDiscount  // 4800000
```

### 2.4 IEvent Ticket 简化计价模型

IEvent 品类最简单，**不支持任何促销活动**：

```go
// IEvent CalPDPPrice - 简化逻辑
func (c *IEventTicketCalService) CalPDPPrice(req *pricing.CalPDPPriceRequest) {
    pricingTotalAmount := int64(0)
    for _, priceEntity := range req.PriceEntities {
        // 直接使用市场价
        pricingAmount := priceEntity.GetMarketPrice() * priceEntity.GetQuantity()
        priceEntity.PricingAmount = proto.Int64(pricingAmount)
        pricingTotalAmount += pricingAmount
    }
    return pricingTotalAmount, req.PriceEntities, errors.Ok
}

// IEvent CalOrderPrice - 委托给 BaseCalService
func (c *IEventTicketCalService) CalOrderPrice(req *pricing.CalOrderPriceRequest) {
    return c.BaseCalService.CalOrderPrice(req)  // 使用 DiscountPrice
}
```

**IEvent 特点**：
- ✅ 无促销支持
- ✅ 无最低交易金额
- ✅ 无复杂规则
- ✅ CalPDPPrice 使用 `MarketPrice`
- ✅ CalOrderPrice 使用 `DiscountPrice`（委托给 BaseCalService）

---

## 三、场景驱动的计价逻辑

### 3.1 场景对比表（基于实际代码）

| 场景 | 方法名 | 计算内容 | 设置字段 | 返回价格 | 适用品类 |
|-----|--------|---------|---------|---------|---------|
| **PDP（商品详情页）** | `CalPDPPrice` | 基础价 + 促销价 | `pricing_amount`<br>`pdp_other_price_map` | `pricingTotalAmount` | 全部 |
| **Order（创建订单）** | `CalOrderPrice` | 基础价 + 促销价 + 费用 | `pricing_amount`<br>`selling_price_info`<br>`extra_info` | `pricingTotalAmount` | 全部 |
| **Payment（支付）** | `CalPayPrice` | 快照验证（零计算） | 无 | `finalPrice` | 全部 |

### 3.2 CalPDPPrice vs CalOrderPrice 差异

**实际代码差异总结**：

| 特性 | CalPDPPrice | CalOrderPrice |
|-----|------------|---------------|
| **计算范围** | 基础价 + 促销价 | 基础价 + 促销价 + 费用 |
| **设置 PricingAmount** | ✅ 设置 | ✅ 设置 |
| **设置 SellingPriceInfo** | ❌ 不设置 | ✅ 设置 |
| **sub_charge 存储位置** | `PdpOtherPriceMap` | `ExtraInfo` |
| **sub_charge 类型** | `int64` | `string` |
| **促销支持** | 品类相关 | 品类相关 |
| **缓存策略** | 高缓存（5-30min） | 不缓存（实时计算） |
| **性能要求** | P99 < 100ms | P99 < 300ms |

**示例对比（Home Service）**：

```go
// ========== CalPDPPrice ==========
service.CalPDPPrice(req)
// 返回：
// - pricingTotalAmount: 90000
// - PdpOtherPriceMap["sub_charge"]: 25000 (int64)
// - priceEntity.PricingAmount: 已设置
// - priceEntity.SellingPriceInfo: nil（不设置）

// ========== CalOrderPrice ==========
service.CalOrderPrice(req)
// 返回：
// - pricingTotalAmount: 90000
// - ExtraInfo["sub_charge"]: "25000" (string)
// - priceEntity.PricingAmount: 已设置
// - priceEntity.SellingPriceInfo: 已填充（含 selling_price, dp_additional_charge 等）
```

---

## 四、品类差异化计价策略

### 4.1 品类计价特性对比

| 品类 | 促销支持 | 最低交易金额 | 特殊逻辑 | 复杂度 |
|-----|---------|-------------|---------|-------|
| **Home Service** | ❌ 不支持 | ✅ 支持 | `ExcludeMinimumValue`<br>`sub_charge` 计算 | ⭐⭐⭐ |
| **Event Ticket** | ✅ 全支持 | ❌ 不支持 | FlashSale/BundlePrice/NewUserPrice<br>购买限制/库存管理 | ⭐⭐⭐⭐⭐ |
| **IEvent Ticket** | ❌ 不支持 | ❌ 不支持 | 无特殊逻辑 | ⭐ |
| **Movie Ticket** | ✅ 部分支持 | ❌ 不支持 | 多种定价模式（single/bundle） | ⭐⭐⭐⭐ |

### 4.2 品类计算器接口设计

基于实际代码结构，设计通用接口：

```go
// Calculator 品类计算器接口
type Calculator interface {
    // CalPDPPrice PDP 场景计价
    CalPDPPrice(req *pricing.CalPDPPriceRequest) (int64, []*pricing.PriceEntity, errors.ErrorCode)
    
    // CalOrderPrice 订单场景计价
    CalOrderPrice(req *pricing.CalOrderPriceRequest) (int64, []*pricing.PriceEntity, errors.ErrorCode)
    
    // CalPayPrice 支付场景计价
    CalPayPrice(req *pricing.CalPayPriceRequest, sumHandlingFee, serviceFee int64) (int64, errors.ErrorCode)
}

// 实际实现示例
type HomeServiceCalService struct {
    *base.BaseCalService
}

type EventTicketCalService struct {
    *base.BaseCalService
}

type IEventTicketCalService struct {
    *base.BaseCalService
}
```

### 4.3 BaseCalService 通用逻辑

```go
// BaseCalService 提供通用计价逻辑（适用于简单品类）
func (b *BaseCalService) CalOrderPrice(req *pricing.CalOrderPriceRequest) (int64, []*pricing.PriceEntity, errors.ErrorCode) {
    pricingTotalAmount := int64(0)
    for _, priceEntity := range req.PriceEntities {
        var pricingAmount int64
        
        // 有促销时使用促销价
        if priceEntity.GetOrderItemPromotionInfo() != nil && 
           priceEntity.GetOrderItemPromotionInfo().GetPromotionInfo().GetPromotionActivityId() > 0 {
            pricingAmount = priceEntity.GetOrderItemPromotionInfo().GetTotalAmount()
        } else {
            // 无促销时使用折扣价
            pricingAmount = priceEntity.GetDiscountPrice() * priceEntity.GetQuantity()
        }
        
        priceEntity.PricingAmount = proto.Int64(pricingAmount)
        FillSellingPriceInfo(priceEntity, req.GetHubAdminFee(), req.GetDpAdditionalCharge(), 0)
        pricingTotalAmount += pricingAmount
    }
    
    return pricingTotalAmount, req.PriceEntities, errors.Ok
}
```

---

## 五、实际项目中的关键 Bug 与修复

### 5.1 Bug 1: Home Service CalOrderPrice 逻辑不一致

**问题**：
```go
// CalPDPPrice（正确）
pricingTotalAmount = ExcludeMinTransactionAmount + 
                    max(minTransaction, IncludedMinTransactionAmount)

// CalOrderPrice（错误）- 修复前
pricingTotalAmount = max(pricingTotalAmount, minTransaction)
```

**影响**：多商品混合场景下，`sub_charge` 无法正确体现在最终总价中。

**修复**：
```go
// CalOrderPrice（正确）- 修复后
pricingTotalAmount = ExcludeMinTransactionAmount + 
                    max(minTransaction, IncludedMinTransactionAmount)
```

**测试覆盖**：新增 4 个边界测试
- `TestCalOrderPrice_MultipleItems_AllIncluded`
- `TestCalOrderPrice_MultipleItems_AllExcluded`
- `TestCalOrderPrice_IncludedEqualsMinTransaction`
- `TestCalOrderPrice_IncludedExceedsMinTransaction`

---

### 5.2 Bug 2: Event Ticket OriginalDiscountPrice 为 0 导致计价为 0

**线上问题**（TraceID: `3c45b2d14bdf73eed4a7b33bb509d200`）：
```
输入: market_price=1500000, discount_price=1500000, quantity=1
输出: pricing_amount=0, pricingTotalAmount=0  ← 严重错误！
```

**根本原因**：
```go
// 修复前
func calculateEventFinalPrice(promotion, priceEntity) (int64, int64) {
    originalDiscountAmount := priceEntity.GetOriginalDiscountPrice() * quantity
    // 当 OriginalDiscountPrice 未传递（为0）时，计算结果为 0
}
```

**修复方案**：
```go
// 修复后（添加兜底逻辑）
func calculateEventFinalPrice(promotion, priceEntity) (int64, int64) {
    originalDiscountPrice := priceEntity.GetOriginalDiscountPrice()
    if originalDiscountPrice == 0 {
        originalDiscountPrice = priceEntity.GetDiscountPrice()  // 兜底
    }
    originalDiscountAmount := originalDiscountPrice * quantity
}
```

**关键教训**：
- ⚠️ **兜底逻辑必须到处添加**：`calculateEventFinalPrice`、`calculateEventBundlePrice`、`calculateEventPromoPrice` 都需要
- ⚠️ **Protobuf 默认值为 0**：未传递的字段默认为 0，必须做 0 值检查
- ⚠️ **关键字段不能为 0**：价格相关的关键字段（如 `OriginalDiscountPrice`）为 0 会导致计算错误

**测试覆盖**：
```go
func TestCalPDPPrice_WithoutOriginalDiscountPrice(t *testing.T) {
    // 构建 PriceEntity - 不设置 OriginalDiscountPrice
    priceEntity := &pricing.PriceEntity{
        MarketPrice:   proto.Int64(1500000),
        DiscountPrice: proto.Int64(1500000),
        // OriginalDiscountPrice 不设置（模拟线上场景）
        Quantity: proto.Int64(1),
    }
    
    // 预期：应该使用 DiscountPrice 兜底
    expectedTotal := int64(1500000)
    assert.Equal(t, expectedTotal, totalAmount)
}
```

---

## 六、统一领域模型设计（基于实际项目优化）

### 6.1 PriceEntity（价格实体）

基于实际 Protobuf 定义优化：

```go
// PriceEntity 价格实体（单个商品的价格信息）
type PriceEntity struct {
    // ========== 基础信息 ==========
    ItemID      int64  // 商品ID
    SkuID       int64  // SKU ID
    CategoryID  int64  // 品类ID
    CarrierID   int64  // 供应商ID
    Quantity    int64  // 数量
    
    // ========== Layer 1: 基础价格 ==========
    MarketPrice           int64  // 市场原价（单价）
    DiscountPrice         int64  // 日常折扣价（单价）
    OriginalDiscountPrice int64  // 原始折扣价（兜底字段，重要！）
    
    // ========== Layer 2: 促销信息 ==========
    // PDP 场景使用（CalPDPPrice）
    ActivityItemPriorityPromotionInfo *ActivityItemPriorityPromotionInfo
    // Order 场景使用（CalOrderPrice）
    OrderItemPromotionInfo            *OrderItemPromotionInfo
    // 购买限制信息
    PurchaseLimitInfo                 *PromotionItemPurchaseLimitInfo
    
    // ========== 计算结果 ==========
    PricingAmount      int64             // 计价金额（最终价格）
    SellingPriceInfo   *SellingPriceInfo // 售卖价格信息（CalOrderPrice设置）
    
    // ========== 品类特殊字段 ==========
    SkuInfo *SkuInfo  // 品类特殊信息（如 HomeServiceSkuInfo）
}

// SellingPriceInfo 售卖价格信息（CalOrderPrice 场景）
type SellingPriceInfo struct {
    SellingPrice           int64  // 售卖价（单价）
    OriginalDiscountPrice  int64  // 原始折扣价
    DpAdditionalCharge     int64  // 附加费用
    MarkUp                 int64  // 加价
}

// HomeServiceSkuInfo Home Service 特殊信息
type HomeServiceSkuInfo struct {
    ServiceId           int64  // 服务ID（与 PriceListId 共同唯一标识服务项目）
    PriceListId         int64  // 价格列表ID
    ExcludeMinimumValue bool   // 是否排除最低交易金额
}
```

### 6.2 PricingRequest 和 PricingResponse

```go
// CalPDPPriceRequest PDP 计价请求
type CalPDPPriceRequest struct {
    RequestId      string          // 请求ID
    PartnerId      int64           // 合作方ID
    CategoryId     int64           // 品类ID
    PriceEntities  []*PriceEntity  // 商品列表
    MinTransaction int64           // 最低交易金额（Home Service）
}

// CalPDPPriceResponse PDP 计价响应
type CalPDPPriceResponse struct {
    PricingTotalAmount          int64               // 总计价金额
    PricingDpAdditionalCharge   int64               // 附加费用总额
    PriceEntities               []*PriceEntity      // 商品列表（含 PricingAmount）
    OtherPrice                  map[string]int64    // 其他价格（如 totalOriginalPrice, sub_charge）
}

// CalOrderPriceRequest 订单计价请求
type CalOrderPriceRequest struct {
    CategoryId          int64           // 品类ID
    PriceEntities       []*PriceEntity  // 商品列表
    MinTransaction      int64           // 最低交易金额（Home Service）
    HubAdminFee         int64           // 平台服务费
    DpAdditionalCharge  int64           // 附加费用
}

// CalOrderPriceResponse 订单计价响应
type CalOrderPriceResponse struct {
    PricingTotalAmount  int64               // 总计价金额
    PriceEntities       []*PriceEntity      // 商品列表（含 PricingAmount 和 SellingPriceInfo）
    ExtraInfo           map[string]string   // 额外信息（如 sub_charge）
}
```

---

## 七、计价公式速查表

### 7.1 通用计价公式

```go
// ========== 单个商品计价 ==========

// Step 1: 获取基础价格（兜底逻辑）
basePrice := priceEntity.GetOriginalDiscountPrice()
if basePrice == 0 {
    basePrice = priceEntity.GetDiscountPrice()  // 兜底（重要！）
}

// Step 2: 应用促销（根据促销类型）
var promotionPrice int64
switch promotionType {
case FlashSale:
    promotionPrice = getFlashSalePrice()
case BundlePrice:
    promotionPrice = calculateBundlePrice()  // 复杂计算
case NewUserPrice:
    promotionPrice = getNewUserPrice()
default:
    promotionPrice = basePrice  // 无促销
}

// Step 3: 考虑购买限制
var pricingAmount int64
if quantity > purchaseLimit {
    // 超限部分使用原价
    pricingAmount = promotionPrice * purchaseLimit + basePrice * (quantity - purchaseLimit)
} else {
    pricingAmount = promotionPrice * quantity
}

// Step 4: CalOrderPrice 场景 - 填充 SellingPriceInfo
if scene == CalOrderPrice {
    sellingPrice := promotionPrice  // 售卖单价 = 促销单价
    FillSellingPriceInfo(priceEntity, hubAdminFee, dpAdditionalCharge, markUp)
}

// ========== 订单汇总（CalOrderPrice） ==========
pricingTotalAmount := sum(all_items.pricing_amount)
```

### 7.2 Home Service 专用公式

```go
// Home Service CalPDPPrice / CalOrderPrice（逻辑一致）

// Step 1: 分类汇总
IncludedMinTransactionAmount := int64(0)
ExcludeMinTransactionAmount := int64(0)

for _, priceEntity := range req.PriceEntities {
    pricingAmount := priceEntity.GetMarketPrice() * priceEntity.GetQuantity()
    
    if priceEntity.GetSkuInfo().GetHomeServiceSkuInfo().GetExcludeMinimumValue() {
        ExcludeMinTransactionAmount += pricingAmount
    } else {
        IncludedMinTransactionAmount += pricingAmount
    }
}

// Step 2: 计算 sub_charge
subCharge := max(0, req.GetMinTransaction() - IncludedMinTransactionAmount)

// CalPDPPrice 存储
c.PdpOtherPriceMap["sub_charge"] = subCharge  // int64

// CalOrderPrice 存储
c.ExtraInfo["sub_charge"] = strconv.FormatInt(subCharge, 10)  // string

// Step 3: 计算最终总价
pricingTotalAmount := ExcludeMinTransactionAmount + 
                     max(req.GetMinTransaction(), IncludedMinTransactionAmount)
```

**关键公式理解**：

```
场景：最低交易金额 60000
商品1（计入）: 35000
商品2（排除）: 30000

计算步骤：
1. IncludedMinTransactionAmount = 35000
2. ExcludeMinTransactionAmount = 30000
3. sub_charge = max(0, 60000 - 35000) = 25000
4. pricingTotalAmount = 30000 + max(60000, 35000)
                      = 30000 + 60000
                      = 90000

逻辑解释：
- 排除商品（30000）按实际价格计入
- 计入商品（35000）需补足到最低交易金额（60000）
- sub_charge（25000）就是需要补足的金额
```

### 7.3 Event Ticket 促销公式

**FlashSale / Scheduling / NewUserPrice（限购促销）**：

```go
promotionPrice := promotion.GetPrice()           // 促销单价
purchaseLimit := getPurchaseLimit(priceEntity)   // 获取限购数量

// 特殊：NewUserPrice 硬编码限购1
if promotionType == NewUserPrice {
    purchaseLimit = 1
}

// 计算
if quantity > purchaseLimit {
    finalAmount = promotionPrice * purchaseLimit + 
                 originalDiscountPrice * (quantity - purchaseLimit)
} else {
    finalAmount = promotionPrice * quantity
}

// 示例：原价 500, 促销价 350, 限购 1, 购买 3
// finalAmount = 350 × 1 + 500 × 2 = 1350
```

**BundlePrice（套餐价）**：

```go
// 套餐价：满 minQuantity 件享 discount 折扣
price := priceEntity.GetDiscountPrice()
quantity := priceEntity.GetQuantity()
minQuantity := promotion.GetBundlePriceMinQuantity()
discount := promotion.GetBundlePriceDiscount()

// 计算轮数
rounds := quantity / minQuantity

// 计算每轮折扣
pricePerRound := price * minQuantity

if rewardType == PercentageDiscount {
    // 百分比折扣：20% → 20000/100000
    discountAmount := (pricePerRound * discount) / 100000
    
    // 印尼地区特殊处理（precision=0）
    precision := getPrecision(env.Region())
    discountPerRound := mathFloorWithPrecision(discountAmount, precision)
} else if rewardType == FixedAmountDiscount {
    // 固定金额折扣
    discountPerRound := discount
}

// 总折扣
totalDiscount := discountPerRound * rounds

// 最终价格
finalAmount := (price * quantity) - totalDiscount

// 示例：单价 1000000, 数量 6, 满3件20%折扣, 印尼地区
// rounds = 6 / 3 = 2
// pricePerRound = 1000000 × 3 = 3000000
// discountAmount = 3000000 × 20000 / 100000 = 600000
// discountPerRound = floor(600000 / 100000) × 100000 = 600000（印尼precision=0）
// totalDiscount = 600000 × 2 = 1200000
// finalAmount = 6000000 - 1200000 = 4800000
```

---

## 八、最佳实践与设计原则

### 8.1 兜底逻辑原则

**原则**: 所有涉及价格计算的关键字段，必须做 0 值检查和兜底。

```go
// ✅ 正确示例 1: OriginalDiscountPrice 兜底
originalDiscountPrice := priceEntity.GetOriginalDiscountPrice()
if originalDiscountPrice == 0 {
    originalDiscountPrice = priceEntity.GetDiscountPrice()
}

// ✅ 正确示例 2: Quantity 为 0 的保护
if priceEntity.GetQuantity() == 0 {
    priceEntity.Quantity = proto.Int64(1)  // 默认为1
}

// ✅ 正确示例 3: 无限大值的模拟
stock := promotion.GetStock()
if stock == 0 {
    stock = InfinityValue  // 999999999 模拟无限库存
}

// ❌ 错误示例：直接使用可能为0的字段
originalDiscountAmount := priceEntity.GetOriginalDiscountPrice() * quantity  // 可能为0！
```

### 8.2 CalPDPPrice vs CalOrderPrice 设计原则

**原则**: 两个方法的职责不同，必须区分清楚。

| 原则 | CalPDPPrice | CalOrderPrice |
|-----|------------|---------------|
| **职责** | PDP展示价格 | 订单确认价格 |
| **计算范围** | 基础价 + 促销价 | 基础价 + 促销价 + 费用 |
| **设置字段** | `pricing_amount` | `pricing_amount`<br>`selling_price_info` |
| **缓存策略** | 高缓存（5-30min） | 不缓存（实时） |
| **性能要求** | P99 < 100ms | P99 < 300ms |
| **准确性要求** | 允许轻微延迟 | 必须准确实时 |

**示例代码**：

```go
// ========== CalPDPPrice ==========
func (c *Service) CalPDPPrice(req *pricing.CalPDPPriceRequest) {
    for _, priceEntity := range req.PriceEntities {
        // 计算 PricingAmount
        pricingAmount := calculateWithPromotion(priceEntity)
        priceEntity.PricingAmount = proto.Int64(pricingAmount)
        
        // ❌ 不设置 SellingPriceInfo
        // priceEntity.SellingPriceInfo = ...  // 不在 CalPDPPrice 设置
    }
    
    // 存储额外信息到 PdpOtherPriceMap
    c.PdpOtherPriceMap["sub_charge"] = subCharge  // int64
    c.PdpOtherPriceMap["totalOriginalPrice"] = totalOriginalPrice
}

// ========== CalOrderPrice ==========
func (c *Service) CalOrderPrice(req *pricing.CalOrderPriceRequest) {
    for _, priceEntity := range req.PriceEntities {
        // 计算 PricingAmount
        pricingAmount := calculateWithPromotion(priceEntity)
        priceEntity.PricingAmount = proto.Int64(pricingAmount)
        
        // ✅ 必须填充 SellingPriceInfo
        FillSellingPriceInfo(priceEntity, hubAdminFee, dpAdditionalCharge, markUp)
    }
    
    // 存储额外信息到 ExtraInfo
    c.ExtraInfo["sub_charge"] = strconv.FormatInt(subCharge, 10)  // string
}
```

### 8.3 促销类型判断原则

**原则**: 促销类型判断必须严格，避免走错分支。

```go
// ✅ 正确：严格判断促销类型
func getActivePromotion(info *ActivityItemPriorityPromotionInfo) *PriorityPromotionInfo {
    if info == nil {
        return &PriorityPromotionInfo{
            PromotionActivityType: PromotionActivityType_UnknownActivityType,
        }
    }
    
    promotion := info.GetPriorityPromotionInfo()
    if promotion == nil || promotion.GetPromotionActivityType() == UnknownActivityType {
        return &PriorityPromotionInfo{
            PromotionActivityType: PromotionActivityType_UnknownActivityType,
        }
    }
    
    return promotion
}

// ✅ 正确：按优先级判断促销类型
func calculateFinalPrice(promotion, priceEntity) (int64, int64) {
    // 无促销
    if promotion.Type == UnknownActivityType {
        return originalAmount, originalAmount
    }
    
    // BundlePrice（特殊逻辑）
    if promotion.Type == BundlePrice {
        return calculateBundlePrice(promotion, priceEntity)
    }
    
    // 限购促销（FlashSale/Scheduling/NewUserPrice）
    if promotion.Type in [FlashSale, Scheduling, NewUserPrice] {
        return calculatePromoPrice(promotion, priceEntity)
    }
    
    return originalAmount, originalAmount
}
```

### 8.4 品类差异处理原则

**原则**: 品类特殊逻辑封装在专用 Calculator 中，不污染通用逻辑。

```go
// ✅ 正确：品类特殊逻辑封装
type HomeServiceCalService struct {
    *base.BaseCalService
}

func (c *HomeServiceCalService) CalPDPPrice(req) {
    // Home Service 特有的最低交易金额逻辑
    pricingTotalAmount := ExcludeMinTransactionAmount + 
                         max(minTransaction, IncludedMinTransactionAmount)
}

// ✅ 正确：简单品类委托给 BaseCalService
type IEventTicketCalService struct {
    *base.BaseCalService
}

func (c *IEventTicketCalService) CalOrderPrice(req) {
    return c.BaseCalService.CalOrderPrice(req)  // 委托
}

// ❌ 错误：在通用逻辑中处理品类差异
func (b *BaseCalService) CalOrderPrice(req) {
    if req.CategoryId == HomeService {
        // Home Service 特殊逻辑（不应该在这里！）
    }
}
```

---

## 九、价格追溯与调试

### 9.1 PriceBreakdown（价格拆解明细）

用于价格审计和调试：

```go
// PriceBreakdown 价格拆解明细（基于实际项目优化）
type PriceBreakdown struct {
    // ========== 基础信息 ==========
    ItemID     int64  // 商品ID
    CategoryID int64  // 品类ID
    Quantity   int64  // 数量
    
    // ========== Layer 1: 基础价格 ==========
    MarketPrice           int64  // 市场原价（单价）
    DiscountPrice         int64  // 折扣价（单价）
    OriginalDiscountPrice int64  // 原始折扣价（实际使用）
    
    // ========== Layer 2: 促销价格 ==========
    PromotionType         string // 促销类型（FlashSale/BundlePrice/NewUserPrice）
    PromotionPrice        int64  // 促销单价
    PurchaseLimit         int64  // 购买限制
    PromotionAppliedQty   int64  // 实际享受促销的数量
    PromotionNormalQty    int64  // 原价购买的数量
    
    // ========== 价格计算结果 ==========
    // 促销前
    DiscountTotalPrice    int64  // 折扣价总额 = discount_price × quantity
    // 促销后
    PricingAmount         int64  // 计价金额（最终价格）
    PromotionDiscountAmount int64 // 促销优惠金额
    
    // ========== Layer 3: 费用（CalOrderPrice场景） ==========
    SellingPrice          int64  // 售卖单价
    HubAdminFee           int64  // 平台服务费
    DpAdditionalCharge    int64  // 附加费用
    MarkUp                int64  // 加价
    
    // ========== Home Service 专有 ==========
    MinTransaction                int64  // 最低交易金额
    IncludedMinTransactionAmount  int64  // 计入最低交易金额的商品总价
    ExcludeMinTransactionAmount   int64  // 排除最低交易金额的商品总价
    SubCharge                     int64  // 附加费用（补足最低交易金额）
    
    // ========== 追溯信息 ==========
    CalculateTime  int64   // 计算时间戳
    Scene          string  // 计算场景（PDP/Order/Payment）
    RequestID      string  // 请求ID（用于追溯）
}

// ToLogString 生成日志字符串（用于调试）
func (b *PriceBreakdown) ToLogString() string {
    var logs []string
    
    logs = append(logs, fmt.Sprintf("ItemID=%d, Quantity=%d", b.ItemID, b.Quantity))
    logs = append(logs, fmt.Sprintf("Layer1: market=%d, discount=%d, original=%d", 
        b.MarketPrice, b.DiscountPrice, b.OriginalDiscountPrice))
    
    if b.PromotionType != "" {
        logs = append(logs, fmt.Sprintf("Layer2: type=%s, price=%d, limit=%d, applied=%d, normal=%d", 
            b.PromotionType, b.PromotionPrice, b.PurchaseLimit, 
            b.PromotionAppliedQty, b.PromotionNormalQty))
    }
    
    logs = append(logs, fmt.Sprintf("Result: discount_total=%d, pricing_amount=%d, saved=%d", 
        b.DiscountTotalPrice, b.PricingAmount, b.PromotionDiscountAmount))
    
    if b.SubCharge > 0 {
        logs = append(logs, fmt.Sprintf("HomeService: min_transaction=%d, included=%d, excluded=%d, sub_charge=%d", 
            b.MinTransaction, b.IncludedMinTransactionAmount, 
            b.ExcludeMinTransactionAmount, b.SubCharge))
    }
    
    return strings.Join(logs, " | ")
}

// 输出示例：
// ItemID=600001, Quantity=3 | Layer1: market=500, discount=500, original=500 | 
// Layer2: type=NewUserPrice, price=350, limit=1, applied=1, normal=2 | 
// Result: discount_total=1500, pricing_amount=1350, saved=150
```

---

## 十、实际项目测试覆盖总结

### 10.1 测试用例统计

| 品类 | CalPDPPrice | CalOrderPrice | 特殊场景 | 总计 | 通过率 |
|-----|------------|--------------|---------|------|-------|
| **Home Service** | 3 | 7 | 2（辅助方法） | 12 | 100% ✅ |
| **IEvent Ticket** | 3 | 3 | 0 | 6 | 100% ✅ |
| **Event Ticket** | 7 | 1 | 1（Bug修复） | 9 | 100% ✅ |
| **总计** | **13** | **11** | **3** | **27** | **100%** ✅ |

### 10.2 覆盖场景总结

**Home Service 覆盖**：
- ✅ 基础场景（单商品、多商品）
- ✅ 最低交易金额场景（计入、排除、混合）
- ✅ 边界条件（等于、超过最低交易金额）
- ✅ 错误处理（SkuInfo为nil）
- ✅ CalPDPPrice vs CalOrderPrice 逻辑一致性验证

**Event Ticket 覆盖**：
- ✅ 无促销场景
- ✅ FlashSale（限购）
- ✅ BundlePrice（套餐价，百分比和固定金额）
- ✅ Scheduling（普通促销）
- ✅ NewUserPrice（新用户价，硬编码限购1）
- ✅ 混合促销（多商品不同促销类型）
- ✅ 边界条件（数量为0）
- ✅ **线上Bug修复**（OriginalDiscountPrice为0）

**IEvent Ticket 覆盖**：
- ✅ 基础场景（单商品、多商品）
- ✅ 边界条件（数量为0）
- ✅ PDP 和 Order 场景

---

## 十一、关键代码片段（实际项目）

### 11.1 Home Service 完整计价流程

```go
func (c *HomeServiceCalService) CalOrderPrice(req *pricing.CalOrderPriceRequest) (int64, []*pricing.PriceEntity, errors.ErrorCode) {
    IncludedMinTransactionAmount := int64(0)
    ExcludeMinTransactionAmount := int64(0)
    
    for _, priceEntity := range req.PriceEntities {
        pricingAmount := priceEntity.GetMarketPrice() * priceEntity.GetQuantity()
        priceEntity.PricingAmount = proto.Int64(pricingAmount)
        
        skuInfo := priceEntity.GetSkuInfo()
        if skuInfo == nil {
            return 0, nil, errors.ErrorParams
        }
        
        homeServiceSkuInfo := skuInfo.GetHomeServiceSkuInfo()
        if homeServiceSkuInfo != nil && !homeServiceSkuInfo.GetExcludeMinimumValue() {
            IncludedMinTransactionAmount += pricingAmount
        } else {
            ExcludeMinTransactionAmount += pricingAmount
        }
    }
    
    // 计算 sub_charge
    if c.ExtraInfo == nil {
        c.ExtraInfo = make(map[string]string)
    }
    c.ExtraInfo["sub_charge"] = strconv.FormatInt(
        max(0, req.GetMinTransaction() - IncludedMinTransactionAmount), 10)
    
    // 计算最终总价（关键逻辑）
    pricingTotalAmount := ExcludeMinTransactionAmount + 
                         max(req.GetMinTransaction(), IncludedMinTransactionAmount)
    
    // 填充 SellingPriceInfo
    c.fillSellingPriceInfo(req.PriceEntities)
    
    return pricingTotalAmount, req.PriceEntities, errors.Ok
}
```

### 11.2 Event Ticket 促销计算流程

```go
func (c *EventTicketCalService) CalPDPPrice(req *pricing.CalPDPPriceRequest) (int64, []*pricing.PriceEntity, errors.ErrorCode) {
    pricingTotalAmount := int64(0)
    pricingTotalOriginalAmount := int64(0)
    
    for _, priceEntity := range req.PriceEntities {
        // 数量为0保护
        if priceEntity.GetQuantity() == 0 {
            priceEntity.Quantity = proto.Int64(1)
        }
        
        // 获取有效促销
        promotion := getActivePromotionForEvent(
            priceEntity.GetActivityItemPriorityPromotionInfo())
        
        // 计算最终价格
        originalDiscountAmount, finalAmount := calculateEventFinalPrice(promotion, priceEntity)
        
        priceEntity.PricingAmount = proto.Int64(finalAmount)
        pricingTotalAmount += finalAmount
        pricingTotalOriginalAmount += originalDiscountAmount
    }
    
    // 存储折扣前价格（用于UI显示节省金额）
    c.PdpOtherPriceMap["totalOriginalPrice"] = pricingTotalOriginalAmount
    
    return pricingTotalAmount, req.PriceEntities, errors.Ok
}

func calculateEventFinalPrice(promotion, priceEntity) (int64, int64) {
    quantity := priceEntity.GetQuantity()
    
    // 兜底逻辑（防止线上Bug）
    originalDiscountPrice := priceEntity.GetOriginalDiscountPrice()
    if originalDiscountPrice == 0 {
        originalDiscountPrice = priceEntity.GetDiscountPrice()
    }
    originalDiscountAmount := originalDiscountPrice * quantity
    
    // 无促销
    if promotion.Type == UnknownActivityType {
        return originalDiscountAmount, originalDiscountAmount
    }
    
    // BundlePrice（套餐价）
    if promotion.Type == BundlePrice {
        return calculateEventBundlePrice(promotion, priceEntity)
    }
    
    // 限购促销
    if promotion.Type in [FlashSale, Scheduling, NewUserPrice] {
        return calculateEventPromoPrice(promotion, priceEntity)
    }
    
    return originalDiscountAmount, originalDiscountAmount
}
```

### 11.3 BaseCalService 通用逻辑

```go
func (b *BaseCalService) CalOrderPrice(req *pricing.CalOrderPriceRequest) (int64, []*pricing.PriceEntity, errors.ErrorCode) {
    pricingTotalAmount := int64(0)
    
    for _, priceEntity := range req.PriceEntities {
        var pricingAmount int64
        
        // 有促销时使用 OrderItemPromotionInfo.TotalAmount
        if priceEntity.GetOrderItemPromotionInfo() != nil && 
           priceEntity.GetOrderItemPromotionInfo().GetPromotionInfo().GetPromotionActivityId() > 0 {
            pricingAmount = priceEntity.GetOrderItemPromotionInfo().GetTotalAmount()
        } else {
            // 无促销时使用折扣价
            pricingAmount = priceEntity.GetDiscountPrice() * priceEntity.GetQuantity()
        }
        
        priceEntity.PricingAmount = proto.Int64(pricingAmount)
        FillSellingPriceInfo(priceEntity, req.GetHubAdminFee(), req.GetDpAdditionalCharge(), 0)
        pricingTotalAmount += pricingAmount
    }
    
    return pricingTotalAmount, req.PriceEntities, errors.Ok
}
```

---

## 十二、术语统一检查清单

### 12.1 代码命名规范

**✅ 推荐使用**：

| 场景 | 推荐命名 | 避免使用 |
|-----|---------|---------|
| 市场价 | `market_price` | `original_price`, `list_price` |
| 折扣价 | `discount_price` | `sale_price`, `regular_price` |
| 促销价 | `promotion_price` | `activity_price`, `promo_price` |
| 计价金额 | `pricing_amount` | `final_price`, `total_price` |
| 售卖价 | `selling_price` | `sell_price` |
| 最低交易金额 | `min_transaction` | `minimum_amount`, `min_order` |
| 附加费用 | `sub_charge` (Home Service)<br>`additional_charge` (通用) | `extra_fee`, `surcharge` |

### 12.2 注释规范

```go
// ✅ 正确：清晰标注价格维度和计算阶段
var (
    marketPrice   int64 = 1000  // 市场原价（单价）
    discountPrice int64 = 980   // 折扣价（单价）
    quantity      int64 = 2     // 数量
)

// 促销前总价（营销前）
discountTotalPrice := discountPrice * quantity  // 1960

// 促销后总价（营销后）
promotionTotalPrice := promotionPrice * quantity  // 1760

// 优惠金额
savingsAmount := discountTotalPrice - promotionTotalPrice  // 200

// ❌ 错误：术语不清晰
var (
    price      int64 = 1000  // 什么价？
    finalPrice int64 = 980   // 最终价？
)
total := price * qty  // 促销前还是促销后？
```

### 12.3 Protobuf 字段命名

```protobuf
// ✅ 正确：使用统一术语
message PriceEntity {
    optional int64 market_price = 1;              // 市场原价
    optional int64 discount_price = 2;            // 折扣价
    optional int64 original_discount_price = 3;   // 原始折扣价
    optional int64 pricing_amount = 4;            // 计价金额
    optional int64 quantity = 5;                  // 数量
}

message CalPDPPriceRequest {
    optional int64 min_transaction = 1;           // 最低交易金额（Home Service）
    repeated PriceEntity price_entities = 2;      // 商品列表
}

// ❌ 错误：术语不一致
message PriceEntity {
    optional int64 original_price = 1;   // 避免使用，容易混淆
    optional int64 sale_price = 2;       // 避免使用
    optional int64 final_price = 3;      // 避免使用（不够精确）
}
```

---

## 十三、面试问题准备（基于实际项目）

### Q1: 你们项目中遇到过哪些价格计算的线上 Bug？

**A**: 我们遇到过两个严重的价格计算 Bug：

**Bug 1: Event Ticket 计价为 0（严重资损风险）**
- **现象**: 线上日志显示 `market_price=1500000`, `pricing_amount=0`
- **原因**: `OriginalDiscountPrice` 字段未传递（默认为0），计算时直接使用导致结果为0
- **影响**: 所有未传 `OriginalDiscountPrice` 的请求都会计价为0，严重资损风险
- **修复**: 添加兜底逻辑 `if originalDiscountPrice == 0 { use discountPrice }`
- **预防**: 新增专门测试 `TestCalPDPPrice_WithoutOriginalDiscountPrice`

**Bug 2: Home Service CalOrderPrice 逻辑不一致**
- **现象**: `CalPDPPrice` 和 `CalOrderPrice` 在多商品混合场景下计算结果不一致
- **原因**: `CalOrderPrice` 使用错误公式 `max(allItems, minTransaction)`
- **影响**: `sub_charge` 无法正确体现在最终总价中，用户看到的附加费用与实际不符
- **修复**: 统一使用 `ExcludeAmount + max(minTransaction, IncludedAmount)`
- **预防**: 新增 4 个边界测试覆盖各种组合

**关键教训**：
1. **兜底逻辑必不可少**: Protobuf 字段未传递时默认为0，关键字段必须做0值检查
2. **CalPDPPrice 和 CalOrderPrice 逻辑必须一致**: 两个方法应该返回相同的价格（不考虑 SellingPriceInfo）
3. **单元测试要覆盖边界**: 特别是字段为0、超限购、混合场景等边界条件

### Q2: Home Service 的最低交易金额是如何设计的？

**A**: Home Service 是一个特殊品类，有"最低交易金额"机制，且部分商品可以"排除"最低交易金额：

**业务场景**：
- 某些服务（如安装费）计入最低交易金额
- 某些商品（如材料费）不计入最低交易金额
- 订单必须达到最低交易金额才能下单

**数据结构**：
```go
type HomeServiceSkuInfo struct {
    ServiceId           int64  // 服务ID
    PriceListId         int64  // 价格列表ID
    ExcludeMinimumValue bool   // 是否排除最低交易金额
}
```

**计算逻辑**：
```go
// Step 1: 分类汇总
for _, item := range items {
    amount := item.price * item.quantity
    if item.ExcludeMinimumValue {
        ExcludeMinTransactionAmount += amount  // 不计入
    } else {
        IncludedMinTransactionAmount += amount  // 计入
    }
}

// Step 2: 计算需要补足的金额
sub_charge = max(0, min_transaction - IncludedMinTransactionAmount)

// Step 3: 计算最终总价
// 关键：排除商品按实际价格，计入商品需补足到最低交易金额
total = ExcludeMinTransactionAmount + max(min_transaction, IncludedMinTransactionAmount)
```

**示例**：
```
最低交易金额: 60000
商品1（安装费，计入）: 35000
商品2（材料费，排除）: 30000

计算：
sub_charge = max(0, 60000 - 35000) = 25000
total = 30000 + max(60000, 35000) = 90000

解释：材料费30000按实际计入，安装费35000需补足到60000
```

### Q3: Event Ticket 的 NewUserPrice 为什么硬编码限购1？

**A**: 这是业务规则和技术实现的权衡：

**业务规则**：
- NewUserPrice 是专门给新用户的优惠价，目的是降低新用户首单门槛
- 为了控制成本，新用户只能以优惠价购买1件，超过部分按原价

**技术实现**：
```go
func getUserPurchaseLimit(priceEntity, promotionType) int64 {
    if promotionType == PromotionActivityType_NewUserPrice {
        return 1  // 硬编码（业务规则）
    }
    
    // 其他促销从 PurchaseLimitInfo 动态获取
    return priceEntity.GetPurchaseLimitInfo().GetAvailablePurchaseLimit()
}
```

**为什么硬编码？**
- 业务规则固定，不需要动态配置
- 简化逻辑，避免从营销服务额外查询
- 性能考虑，减少一次 RPC 调用

**测试覆盖**：
```go
// 测试用例明确标注硬编码逻辑
purchaseLimit := 1  // 硬编码限购1（代码中硬编码）
expectedFinalPrice := newUserPrice * 1 + originalPrice * (quantity - 1)
```

---

## 十四、未来优化方向

### 14.1 模型优化

**1. 引入更精细的价格维度**：
```go
type EnhancedPriceEntity struct {
    // 当前只有单价，未来可以区分
    UnitPrice struct {
        Market    int64  // 市场单价
        Discount  int64  // 折扣单价
        Promotion int64  // 促销单价
    }
    
    TotalPrice struct {
        BeforePromotion int64  // 促销前总价
        AfterPromotion  int64  // 促销后总价
    }
    
    Quantity int64
}
```

**2. 统一 sub_charge 类型**：
```go
// 当前不一致：
// CalPDPPrice: PdpOtherPriceMap["sub_charge"] = int64
// CalOrderPrice: ExtraInfo["sub_charge"] = string

// 优化为：统一使用 int64
type PricingResponse struct {
    SubCharge int64  // 统一类型
}
```

**3. 统一 SellingPriceInfo 设置逻辑**：
```go
// 当前：CalPDPPrice 不设置，CalOrderPrice 设置
// 优化：两个方法都设置，但 CalPDPPrice 可以选择性返回

type PricingOptions struct {
    FillSellingPriceInfo bool  // 是否填充 SellingPriceInfo
    CalculateSubCharge   bool  // 是否计算 sub_charge
}
```

### 14.2 测试优化

**1. 增加性能测试**：
```go
func BenchmarkCalPDPPrice(b *testing.B) {
    // 测试不同商品数量的性能
    for n := 0; n < b.N; n++ {
        service.CalPDPPrice(req)
    }
}
```

**2. 增加集成测试**：
```go
func TestCalPDPPriceToCalOrderPrice_Consistency(t *testing.T) {
    // 验证 CalPDPPrice 和 CalOrderPrice 价格一致性
    pdpTotal, _, _ := service.CalPDPPrice(pdpReq)
    orderTotal, _, _ := service.CalOrderPrice(orderReq)
    
    // 价格应该一致（不考虑 SellingPriceInfo）
    assert.Equal(t, pdpTotal, orderTotal)
}
```

**3. 增加模糊测试**：
```go
func FuzzCalPDPPrice(f *testing.F) {
    // 模糊测试各种边界输入
    f.Fuzz(func(t *testing.T, price int64, quantity int64) {
        // 确保不会 panic 或返回负数
    })
}
```

---

## 十五、总结

### 15.1 统一术语的价值

基于实际项目的术语统一工作带来：
- ✅ **沟通效率提升**: 团队不再混淆单价和总价、促销前和促销后
- ✅ **Bug 减少**: 明确的术语减少理解偏差，发现了 2 个严重 Bug
- ✅ **代码可读性**: 统一命名规范，新人上手速度提升
- ✅ **测试覆盖**: 明确的模型指导测试用例设计，覆盖率达到 100%

### 15.2 实际项目经验总结

**1. 兜底逻辑的重要性**：
- `OriginalDiscountPrice` 为 0 导致线上计价为 0
- 所有关键字段必须做 0 值检查

**2. CalPDPPrice 和 CalOrderPrice 必须一致**：
- Home Service 的逻辑不一致导致 `sub_charge` 计算错误
- 两个方法除了 `SellingPriceInfo` 外，价格应该相同

**3. 品类差异要明确**：
- Home Service: 最低交易金额
- Event Ticket: 多种促销
- IEvent Ticket: 简化逻辑

**4. 单元测试的价值**：
- 27 个测试用例覆盖了所有核心场景和边界条件
- 发现并修复了 2 个严重 Bug
- 为未来重构提供了安全保障

---

**本文档基于 Digital Purchase Service 实际项目代码编写，所有术语、公式、示例均来自真实生产代码。**
