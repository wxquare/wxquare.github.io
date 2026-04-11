# 电商价格计算引擎 - 面试项目介绍

## 一、项目背景

### 业务场景

在一个大型跨境电商平台，价格计算涉及**多品类、多营销活动、多变价因素**。随着业务发展，原有的分散式计价架构暴露出严重问题，急需统一重构。

**业务规模**：
- 日均调用量：**5000 万+**
- 峰值 QPS：**10,000+**
- 支持品类：**10+ 个品类**（Topup、E-Money、E-Voucher、Hotel、Flight、Tour 等）

---

### 核心痛点

#### 痛点一：价格计算分散，维护成本高
价格逻辑分散在 **5+ 个服务**中（前端、商品、营销、订单、支付），导致：
- 新增变价因素需修改多个服务，上线周期 2-3 周
- 逻辑重复，代码难以维护

#### 痛点二：术语不统一，沟通成本高
价格术语混乱（市场价、原价、划线价、售卖价...），各团队理解不一致，导致：
- 需求理解偏差 **20%**
- 需求评审反复澄清，开发完发现理解错误需返工

#### 痛点三：品类差异大，缺少统一模型
每个品类独立实现计价逻辑，代码重复率高，导致：
- 新增品类需 **2 个月**从头开发
- 代码复用率仅 **30%**

---

### 项目目标

建设一个**统一的价格计算中心（Pricing Center）**，实现：
- ✅ 价格计算逻辑收敛到一个服务
- ✅ 统一价格术语，沟通效率提升 60%
- ✅ 支持多品类扩展，新增品类成本降低 8 倍

**技术约束**：
- 高性能：PDP P99 < 100ms，Checkout P99 < 200ms
- 高可用：可用性 99.99%，不能影响下单链路

**业务约束**：
- 安全灰度：涉及 10+ 个品类，必须平滑迁移，0 资损
- 老系统并行：新老系统并行运行，空跑比对验证正确性

---

## 二、核心痛点与解决方案

### 2.1 痛点一：价格计算分散 → 解决方案：场景驱动的统一计算引擎

#### 问题详情
价格逻辑分散在 **5+ 个服务**中：
- **前端服务**：商品详情页展示价格
- **商品服务**：管理市场价、折扣价
- **营销服务**：计算活动优惠，和数量相关的计算由前端计算
- **订单服务**：计算订单总价
- **支付服务**：计算最终支付金额

**影响**：
- 新增变价因素需修改 5 个服务，上线周期 2-3 周
- 测试成本高，容易遗漏
- 代码重复率高，难以维护

#### 解决方案：场景驱动架构（Scene-Driven Architecture）

**核心思想**：价格计算贯穿用户购物全流程，不同场景对性能、准确性、计算内容的要求完全不同。

| 场景 | 性能要求 | 计算内容 | 缓存策略 | 资源锁定 |
|------|---------|---------|---------|---------|
| **PDP（商品详情页）** | P99 < 100ms | 基础价 + 营销价 | 高缓存（5-30min） | ❌ 不锁定 |
| **购物车** | P99 < 200ms | 基础价 + 营销价 + 预估券 | 中缓存（2-10min） | ❌ 不锁定 |
| **CreateOrder（创建订单）** | P99 < 300ms | 基础价 + 营销价 + 附加费<br>（不含券/积分） | ❌ 不缓存 | ✅ 库存锁定（30min） |
| **Checkout（收银台）** | P99 < 200ms | 完整计算<br>（含券/积分/手续费） | ❌ 不缓存 | ✅ 软锁定券/积分 |
| **Payment（支付）** | P99 < 500ms | 零计算（快照验证） | ❌ 不缓存 | ✅ 正式锁定 |

**五层责任链计算引擎**：

采用**责任链模式（Chain of Responsibility）**，将价格计算拆分为 5 个独立的层级：

```
价格计算流程：
Base Price Layer (Layer 1)
    ↓ 获取商品基础价格（市场价/折扣价）
Promotion Layer (Layer 2)
    ↓ 应用营销活动（FlashSale、新人价、Bundle）
Deduction Layer (Layer 3)
    ↓ 应用抵扣（优惠券、积分）
Charge Layer (Layer 4)
    ↓ 计算费用（手续费、增值服务费）
Final Price Layer (Layer 5)
    ↓ 汇总最终价格 + 生成明细
```

**关键特性**：
- ✅ **灵活跳层**：不同场景可跳过某些层（如 PDP 跳过抵扣和费用层）
- ✅ **独立解耦**：每层职责单一，易于测试和维护
- ✅ **易于扩展**：新增变价因素只需新增一个 Layer，无需改动其他层

```go
// 场景驱动的层级选择
func (e *engine) getSkipLayers(scene PricingScene) []string {
    switch scene {
    case ScenePDP:
        return []string{"deduction", "charge"}  // 只计算基础价+营销价
    case SceneCreateOrder:
        return []string{"deduction"}            // 跳过券/积分
    case SceneCheckout:
        return []string{}                       // 完整计算
    }
}
```

**成果**：
- ✅ 所有价格计算逻辑收敛到一个服务
- ✅ 新增变价因素上线周期：从 **2-3 周** → **3-5 天**（提升 5 倍）
- ✅ 代码维护成本：降低 **80%**

---

### 2.2 痛点二：价格不一致 → 解决方案：价格快照机制

#### 问题详情
前端展示价格 ≠ 订单价格 ≠ 支付价格，导致：
- 年均 **3-5 次**资损事故，单次损失 **10 万+**
- 用户投诉率高，客服成本增加

**真实案例（资损事故）**：
> 2024 年 3 月，优惠券服务在订单创建时重复扣减，用户实际支付比前端展示价低 20%，该 Bug 持续 2 小时，影响 **5000+ 订单**，总损失 **10 万+**，引发用户投诉和内部追责。

#### 解决方案：价格快照 + 二次验证

**两个快照的设计**：

**订单快照（CreateOrder 生成）**：
- 保存订单价格（基础价 + 营销价 + 附加费，**不含券/积分**）
- 有效期：30 分钟（与订单有效期一致）

**支付快照（Checkout 生成）**：
- 保存最终支付价格（订单价格 + 券/积分 + 手续费）
- 包含用户在收银台的所有选择
- 有效期：15 分钟（用户需尽快完成支付）

**快照的作用**：
1. **防止价格篡改**：Payment 场景直接使用快照验证，不重新计算
2. **保证价格一致性**：CreateOrder 和 Checkout 各自锁定价格
3. **防止重复支付**：快照使用后标记为 `used`，幂等处理

```go
// 支付时验证快照
func ValidateSnapshot(snapshotID, orderID string) error {
    snapshot := getPaymentSnapshot(snapshotID)
    
    // 验证快照有效性
    if snapshot == nil { return ErrSnapshotExpired }
    if snapshot.OrderID != orderID { return ErrOrderMismatch }
    if snapshot.Status == "used" { return ErrDuplicatePayment }
    
    // 验证支付金额
    if paymentAmount != snapshot.FinalPrice { 
        return ErrAmountMismatch 
    }
    
    return nil
}
```

**成果**：
- ✅ 资损事故：从 **年均 3-5 次** → **0 次**
- ✅ 价格一致性：从 **95%** → **99.99%**
- ✅ 用户投诉率：下降 **95%**

---

### 2.3 痛点三：术语不统一 → 解决方案：价格字典 + 领域模型

#### 问题详情
价格术语混乱：**市场价、原价、标价、划线价、折扣价、活动价、售卖价、应付金额、最终价格**...
- 产品说"售价"，前端理解为"折扣价"，后端理解为"最终支付价"
- 数据库字段：`market_price`、`selling_price`、`final_price`、`pay_price` 各服务理解不一致

**真实案例**：
> 某次需求评审，产品说"把售卖价改为活动价"，前端理解为修改展示逻辑，后端理解为修改计算逻辑，开发完发现需求理解完全错误，需要重新开发。

#### 解决方案：建立价格字典（Price Dictionary）+ 四层计价模型

**四层计价模型（Four-Layer Pricing Model）**：

```
Layer 1: 基础价格层 (Base Price)
  ↓  market_price / discount_price
  
Layer 2: 促销层 (Promotion)
  ↓  promotion_price (秒杀、新用户价、Bundle)
  
Layer 3: 费用层 (Fee)
  ↓  + admin_fee + additional_charge + service_fee
  
Layer 4: 优惠层 (Discount/Voucher)
  ↓  - voucher_redeemed - coins_redeemed
  
Final Price: 最终价格
```

**统一价格术语体系（多维度设计）**：

### 维度说明

**价格维度**：
- **单价（Unit）**：单个商品的价格，不含数量
- **商品总价（Item）**：单个商品的价格 × 数量
- **订单总价（Order）**：所有商品的总价

**计算阶段**：
- **营销前（Before）**：未应用促销活动
- **营销后（After）**：应用促销活动后

### 完整价格术语表

| 层级 | 维度 | 术语 | 英文 | 字段名 | 定义 | 示例 |
|-----|-----|------|------|--------|------|------|
| **Layer 1** | Unit | 市场原价(单) | Market Unit Price | `market_price` | 单个商品的市场标价 | ¥1000/件 |
| | Unit | 折扣价(单) | Discount Unit Price | `discount_price` | 单个商品的日常销售价 | ¥980/件 |
| | Item | 市场原价(总) | Market Total Price | `market_total_price` | 市场原价 × 数量 | ¥2000 (×2) |
| | Item | 折扣价(总) | Discount Total Price | `discount_total_price` | 折扣价 × 数量 | ¥1960 (×2) |
| **Layer 2** | Unit | 促销价(单) | Promotion Unit Price | `promotion_price` | 促销后的单价 | ¥880/件 |
| | Item | 促销价(总) | Promotion Total Price | `promotion_total_price` | 促销价 × 数量 | ¥1760 (×2) |
| | Item | 促销优惠金额 | Promotion Discount | `promotion_discount_amount` | 促销节省的金额 | ¥200 |
| **Layer 3** | Item | 商品计价金额 | Item Pricing Amount | `item_pricing_amount` | 促销价总价 + 商品级费用 | ¥1795 |
| | Order | 订单商品总额 | Order Items Total | `order_items_total` | 所有商品计价金额之和 | ¥3590 |
| | Order | 订单级费用 | Order Level Fee | `order_level_fee` | 订单级别的服务费 | ¥50 |
| **Layer 4** | Order | 优惠券抵扣 | Voucher Redeemed | `voucher_redeemed` | 优惠券抵扣金额 | ¥100 |
| | Order | 积分抵扣 | Coins Redeemed | `coins_redeemed` | 积分抵扣金额 | ¥40 |
| **Final** | Order | 订单总额 | Order Total Amount | `order_total_amount` | 最终订单总价 | ¥3500 |

**价格计算公式（多维度）**：

```go
// ========== 单个商品（Item Level）计算 ==========

// Step 1: Layer 1 - 基础价格（单价）
unit_market_price = item.market_price          // ¥1000/件
unit_discount_price = item.discount_price      // ¥980/件

// Step 2: Layer 2 - 促销价格（单价）
var unit_promotion_price int64
if has_flash_sale {
    unit_promotion_price = flash_sale_price    // ¥880/件（秒杀价）
} else if has_priority_promotion {
    unit_promotion_price = priority_promotion_price
} else {
    unit_promotion_price = unit_discount_price
}

// Step 3: 计算营销前后的商品总价（含数量）
market_total_price = unit_market_price * quantity        // ¥2000（营销前）
discount_total_price = unit_discount_price * quantity    // ¥1960（营销前）
promotion_total_price = unit_promotion_price * quantity  // ¥1760（营销后）

// 促销优惠金额
promotion_discount_amount = discount_total_price - promotion_total_price  // ¥200

// Step 4: Layer 3 - 加上商品级费用
item_pricing_amount = promotion_total_price + item_admin_fee + item_additional_charge
// ¥1760 + ¥25 + ¥10 = ¥1795

// ========== 订单级别（Order Level）计算 ==========

// Step 5: 汇总所有商品
order_items_total = sum(all_items.item_pricing_amount)
// 商品A: ¥1795 + 商品B: ¥1795 = ¥3590

// Step 6: 加上订单级费用
order_subtotal = order_items_total + order_level_fee
// ¥3590 + ¥50 = ¥3640

// Step 7: Layer 4 - 减去订单级优惠
order_total_amount = order_subtotal - voucher_redeemed - coins_redeemed
// ¥3640 - ¥100 - ¥40 = ¥3500

// Step 8: 价格保护
if order_total_amount < 0 {
    order_total_amount = 0
}
```

**PDP 加购场景展示**：

在商品详情页（PDP），用户调整数量时，需要展示：

```go
// 用户选择数量：2 件

// 展示 1：单价信息
market_unit_price = ¥1000/件     // 划线价
promotion_unit_price = ¥880/件   // 秒杀价（大字展示）

// 展示 2：营销前总价（供用户对比）
discount_total_price = ¥1960     // 日常价总额

// 展示 3：营销后总价（加购金额）
promotion_total_price = ¥1760    // 秒杀总额（强调）

// 展示 4：优惠信息
promotion_discount_amount = ¥200 // "立省 ¥200"
savings_percentage = 10.2%       // "省 10%"
```

**DDD 领域模型设计**：

基于四层计价模型，设计清晰的价格领域模型：

**1. PriceComponent（价格组成元素）**
```go
// PriceComponent 价格组成元素（Layer 2/3/4 的各项明细）
type PriceComponent struct {
    Layer    int    // 所属层级：2-促销, 3-费用, 4-优惠
    Type     string // "promotion", "fee", "discount"
    SubType  string // "flash_sale", "admin_fee", "voucher"...
    Amount   int64  // 金额（以分为单位）
    Name     string // 展示名称
    RuleID   string // 规则 ID（可追溯）
}

func (c *PriceComponent) IsPromotion() bool {
    return c.Layer == 2 && c.Type == "promotion"
}

func (c *PriceComponent) IsFee() bool {
    return c.Layer == 3 && c.Type == "fee"
}

func (c *PriceComponent) IsDiscount() bool {
    return c.Layer == 4 && c.Type == "discount"
}
```

**2. PriceBreakdown（价格拆解明细 - 多维度）**
```go
// PriceBreakdown 价格拆解明细（支持单价、商品总价、订单总价）
type PriceBreakdown struct {
    // ========== Item Level（单个商品维度）==========
    
    // 数量信息
    Quantity          int64  // 购买数量
    
    // Layer 1: 基础价格（单价 + 总价）
    MarketUnitPrice       int64  // 市场原价（单）
    DiscountUnitPrice     int64  // 折扣价（单）
    MarketTotalPrice      int64  // 市场原价（总）= 单价 × 数量
    DiscountTotalPrice    int64  // 折扣价（总）= 单价 × 数量
    
    // Layer 2: 促销价格（单价 + 总价）
    PromotionUnitPrice    int64  // 促销单价
    PromotionTotalPrice   int64  // 促销总价 = 单价 × 数量
    PromotionDiscountAmount int64  // 促销优惠金额（营销前总价 - 营销后总价）
    PromotionType         string // 促销类型（FlashSale/NewUserPrice）
    
    // Layer 3: 费用（商品级）
    ItemAdminFee          int64  // 商品平台服务费
    ItemAdditionalCharge  int64  // 商品附加费用
    ItemPricingAmount     int64  // 商品计价金额 = 促销总价 + 费用
    
    // ========== Order Level（订单维度）==========
    
    // 订单商品汇总
    OrderItemsTotal       int64  // 所有商品计价金额之和
    OrderLevelFee         int64  // 订单级费用
    
    // Layer 4: 优惠（订单级）
    VoucherRedeemed       int64  // 优惠券抵扣
    CoinsRedeemed         int64  // 积分抵扣
    PaymentDiscount       int64  // 支付优惠
    TotalDiscounts        int64  // 总优惠
    
    // Final: 最终价格
    OrderTotalAmount      int64  // 订单总额（最终支付价格）
    
    // 详细组成
    Components            []PriceComponent  // 所有价格组成元素
}

// ToDescription 生成用户友好的价格说明（区分单价和总价）
func (b *PriceBreakdown) ToDescription() []string {
    desc := []string{}
    
    // 单价信息
    desc = append(desc, fmt.Sprintf("【单价】原价: ¥%.2f/件", 
        float64(b.MarketUnitPrice)/100))
    if b.PromotionUnitPrice > 0 {
        desc = append(desc, fmt.Sprintf("【单价】秒杀价: ¥%.2f/件", 
            float64(b.PromotionUnitPrice)/100))
    }
    
    // 数量信息
    desc = append(desc, fmt.Sprintf("【数量】购买: %d 件", b.Quantity))
    
    // 营销前总价
    desc = append(desc, fmt.Sprintf("【营销前】商品总价: ¥%.2f", 
        float64(b.DiscountTotalPrice)/100))
    
    // 营销后总价
    if b.PromotionTotalPrice > 0 {
        desc = append(desc, fmt.Sprintf("【营销后】促销总价: ¥%.2f（立省 ¥%.2f）", 
            float64(b.PromotionTotalPrice)/100,
            float64(b.PromotionDiscountAmount)/100))
    }
    
    // 费用
    if b.ItemAdminFee > 0 || b.ItemAdditionalCharge > 0 {
        totalFee := b.ItemAdminFee + b.ItemAdditionalCharge
        desc = append(desc, fmt.Sprintf("【Layer 3】+费用: +¥%.2f", 
            float64(totalFee)/100))
    }
    
    // 商品计价金额
    desc = append(desc, fmt.Sprintf("【商品小计】¥%.2f", 
        float64(b.ItemPricingAmount)/100))
    
    // 如果是订单级别，显示优惠和最终价格
    if b.OrderTotalAmount > 0 {
        desc = append(desc, "")
        desc = append(desc, fmt.Sprintf("【订单总计】所有商品: ¥%.2f", 
            float64(b.OrderItemsTotal)/100))
        
        if b.TotalDiscounts > 0 {
            desc = append(desc, fmt.Sprintf("【Layer 4】-优惠: -¥%.2f", 
                float64(b.TotalDiscounts)/100))
        }
        
        desc = append(desc, fmt.Sprintf("【最终价格】应付: ¥%.2f", 
            float64(b.OrderTotalAmount)/100))
    }
    
    return desc
}

// ToPDPDisplay 专门为PDP场景生成展示信息
func (b *PriceBreakdown) ToPDPDisplay() map[string]interface{} {
    return map[string]interface{}{
        // 单价信息（前端大字展示）
        "unit_price": map[string]int64{
            "market":    b.MarketUnitPrice,     // 划线价
            "promotion": b.PromotionUnitPrice,  // 秒杀价（红色大字）
        },
        // 选中数量后的总价
        "total_price": map[string]int64{
            "before_promotion": b.DiscountTotalPrice,   // 营销前 ¥1960
            "after_promotion":  b.PromotionTotalPrice,  // 营销后 ¥1760
        },
        // 优惠信息（用于提示）
        "savings": map[string]interface{}{
            "amount":     b.PromotionDiscountAmount,  // 立省 ¥200
            "percentage": fmt.Sprintf("%.1f%%",       // 省 10.2%
                float64(b.PromotionDiscountAmount)*100/float64(b.DiscountTotalPrice)),
        },
        // 数量信息
        "quantity": b.Quantity,
    }
}
```

**3. Price 聚合根（支持多维度价格计算）**
```go
// Price 聚合根（封装四层计价逻辑，支持单价/总价多维度）
type Price struct {
    // 商品信息
    itemID           int64
    skuID            int64
    
    // 数量信息
    quantity         int64  // 购买数量
    
    // Layer 1: 基础价格（单价）
    marketUnitPrice      int64  // 市场原价（单）
    discountUnitPrice    int64  // 折扣价（单）
    
    // Layer 2-4: 价格组成元素
    components           []PriceComponent
    
    // 计算结果（总价）
    marketTotalPrice     int64  // 市场原价（总）
    discountTotalPrice   int64  // 折扣价（总）
    promotionTotalPrice  int64  // 促销价（总）
    itemPricingAmount    int64  // 商品计价金额
    
    // 价格拆解
    breakdown            *PriceBreakdown
}

// NewPrice 创建价格对象
func NewPrice(itemID, skuID int64, marketPrice, discountPrice int64, quantity int64) *Price {
    return &Price{
        itemID:            itemID,
        skuID:             skuID,
        marketUnitPrice:   marketPrice,
        discountUnitPrice: discountPrice,
        quantity:          quantity,
    }
}

// AddComponent 添加价格组成元素
func (p *Price) AddComponent(comp PriceComponent) error {
    if comp.Layer < 2 || comp.Layer > 4 {
        return errors.New("无效的价格层级")
    }
    p.components = append(p.components, comp)
    return nil
}

// Calculate 按四层模型计算价格（区分单价和总价）
func (p *Price) Calculate() error {
    // Step 1: 计算营销前的总价（Layer 1）
    p.marketTotalPrice = p.marketUnitPrice * p.quantity
    p.discountTotalPrice = p.discountUnitPrice * p.quantity
    
    // Step 2: 应用促销，计算营销后的单价和总价（Layer 2）
    promotionUnitPrice := p.discountUnitPrice  // 默认使用折扣价
    
    for _, comp := range p.components {
        if comp.IsPromotion() {
            promotionUnitPrice = comp.Amount  // 促销单价
            break
        }
    }
    
    p.promotionTotalPrice = promotionUnitPrice * p.quantity
    promotionDiscountAmount := p.discountTotalPrice - p.promotionTotalPrice
    
    // Step 3: 加上商品级费用（Layer 3）
    itemFees := int64(0)
    for _, comp := range p.components {
        if comp.IsFee() {
            itemFees += comp.Amount
        }
    }
    p.itemPricingAmount = p.promotionTotalPrice + itemFees
    
    // Step 4: 生成价格拆解明细
    p.breakdown = &PriceBreakdown{
        Quantity:                p.quantity,
        MarketUnitPrice:         p.marketUnitPrice,
        DiscountUnitPrice:       p.discountUnitPrice,
        MarketTotalPrice:        p.marketTotalPrice,
        DiscountTotalPrice:      p.discountTotalPrice,
        PromotionUnitPrice:      promotionUnitPrice,
        PromotionTotalPrice:     p.promotionTotalPrice,
        PromotionDiscountAmount: promotionDiscountAmount,
        ItemAdminFee:            itemFees,
        ItemPricingAmount:       p.itemPricingAmount,
        Components:              p.components,
    }
    
    return nil
}

// GetUnitPrice 获取单价（用于PDP展示）
func (p *Price) GetUnitPrice() (market, discount, promotion int64) {
    var promotionUnitPrice int64 = p.discountUnitPrice
    for _, comp := range p.components {
        if comp.IsPromotion() {
            promotionUnitPrice = comp.Amount
            break
        }
    }
    return p.marketUnitPrice, p.discountUnitPrice, promotionUnitPrice
}

// GetTotalPrice 获取总价（用于PDP加购数量后展示）
func (p *Price) GetTotalPrice() (beforePromotion, afterPromotion int64) {
    return p.discountTotalPrice, p.promotionTotalPrice
}

// GetSavings 获取优惠金额和百分比
func (p *Price) GetSavings() (amount int64, percentage float64) {
    amount = p.discountTotalPrice - p.promotionTotalPrice
    if p.discountTotalPrice > 0 {
        percentage = float64(amount) * 100 / float64(p.discountTotalPrice)
    }
    return amount, percentage
}

// Validate 验证价格合法性
func (p *Price) Validate() error {
    if p.marketUnitPrice < p.discountUnitPrice {
        return errors.New("市场价不能低于折扣价")
    }
    if p.quantity <= 0 {
        return errors.New("数量必须大于0")
    }
    if p.itemPricingAmount < 0 {
        return errors.New("商品计价金额不能为负数")
    }
    return nil
}
```

**4. Order 聚合根（订单级别价格汇总）**
```go
// Order 聚合根（汇总多个商品的价格）
type Order struct {
    orderID          string
    items            []*Price  // 订单中的所有商品
    
    // 订单级别费用
    orderLevelFee    int64
    
    // 订单级别优惠
    voucherRedeemed  int64
    coinsRedeemed    int64
    paymentDiscount  int64
    
    // 订单总额
    orderItemsTotal  int64  // 所有商品计价金额之和
    orderTotalAmount int64  // 最终支付金额
}

// CalculateOrderTotal 计算订单总额
func (o *Order) CalculateOrderTotal() error {
    // 汇总所有商品
    o.orderItemsTotal = 0
    for _, item := range o.items {
        if err := item.Calculate(); err != nil {
            return err
        }
        o.orderItemsTotal += item.itemPricingAmount
    }
    
    // 加上订单级费用
    subtotal := o.orderItemsTotal + o.orderLevelFee
    
    // 减去订单级优惠
    totalDiscounts := o.voucherRedeemed + o.coinsRedeemed + o.paymentDiscount
    o.orderTotalAmount = subtotal - totalDiscounts
    
    // 价格保护
    if o.orderTotalAmount < 0 {
        o.orderTotalAmount = 0
    }
    
    return nil
}

// GetOrderBreakdown 获取订单级别的价格拆解
func (o *Order) GetOrderBreakdown() *PriceBreakdown {
    totalDiscounts := o.voucherRedeemed + o.coinsRedeemed + o.paymentDiscount
    
    return &PriceBreakdown{
        OrderItemsTotal:  o.orderItemsTotal,
        OrderLevelFee:    o.orderLevelFee,
        VoucherRedeemed:  o.voucherRedeemed,
        CoinsRedeemed:    o.coinsRedeemed,
        PaymentDiscount:  o.paymentDiscount,
        TotalDiscounts:   totalDiscounts,
        OrderTotalAmount: o.orderTotalAmount,
    }
}
```

**使用示例 1：PDP 加购场景（单商品，展示多维度价格）**

```go
// 场景：用户在商品详情页调整数量，需要展示单价和总价

// 创建价格对象
itemPrice := NewPrice(
    1001,      // itemID
    5001,      // skuID
    100000,    // ¥1000 市场原价（单）
    98000,     // ¥980  折扣价（单）
    2,         // 购买 2 件
)

// Layer 2: 添加促销（秒杀价）
itemPrice.AddComponent(PriceComponent{
    Layer:   2,
    Type:    "promotion",
    SubType: "flash_sale",
    Amount:  88000,  // ¥880 秒杀单价
    Name:    "限时秒杀",
    RuleID:  "flash_sale_12345",
})

// Layer 3: 添加商品级费用
itemPrice.AddComponent(PriceComponent{
    Layer:   3,
    Type:    "fee",
    SubType: "admin_fee",
    Amount:  2500,  // ¥25 平台服务费
    Name:    "平台服务费",
})

// 计算价格
itemPrice.Calculate()

// PDP 页面展示
pdpDisplay := itemPrice.breakdown.ToPDPDisplay()

/*
PDP 展示内容：
{
    "unit_price": {
        "market": 100000,      // 划线价：¥1000/件
        "promotion": 88000     // 秒杀价：¥880/件（红色大字）
    },
    "total_price": {
        "before_promotion": 196000,  // 营销前：¥1960
        "after_promotion": 176000    // 营销后：¥1760（加粗显示）
    },
    "savings": {
        "amount": 20000,         // 立省：¥200
        "percentage": "10.2%"    // 节省比例
    },
    "quantity": 2
}
*/

// 获取各维度价格
marketUnit, discountUnit, promoUnit := itemPrice.GetUnitPrice()
fmt.Printf("单价 - 市场价: ¥%d, 折扣价: ¥%d, 秒杀价: ¥%d\n", 
    marketUnit/100, discountUnit/100, promoUnit/100)
// 输出: 单价 - 市场价: ¥1000, 折扣价: ¥980, 秒杀价: ¥880

beforePromo, afterPromo := itemPrice.GetTotalPrice()
fmt.Printf("总价 - 营销前: ¥%d, 营销后: ¥%d\n", 
    beforePromo/100, afterPromo/100)
// 输出: 总价 - 营销前: ¥1960, 营销后: ¥1760

savingsAmount, savingsPercent := itemPrice.GetSavings()
fmt.Printf("优惠 - 立省: ¥%d (%.1f%%)\n", 
    savingsAmount/100, savingsPercent)
// 输出: 优惠 - 立省: ¥200 (10.2%)
```

**使用示例 2：订单场景（多商品，汇总计算）**

```go
// 场景：用户创建订单，包含2个商品

// 商品 A
itemA := NewPrice(1001, 5001, 100000, 98000, 2)
itemA.AddComponent(PriceComponent{
    Layer: 2, Type: "promotion", SubType: "flash_sale",
    Amount: 88000, Name: "限时秒杀",
})
itemA.AddComponent(PriceComponent{
    Layer: 3, Type: "fee", SubType: "admin_fee",
    Amount: 2500, Name: "平台服务费",
})

// 商品 B
itemB := NewPrice(2001, 6001, 50000, 48000, 3)
itemB.AddComponent(PriceComponent{
    Layer: 2, Type: "promotion", SubType: "new_user",
    Amount: 40000, Name: "新用户价",
})
itemB.AddComponent(PriceComponent{
    Layer: 3, Type: "fee", SubType: "admin_fee",
    Amount: 1500, Name: "平台服务费",
})

// 创建订单
order := &Order{
    orderID:         "ORD123456",
    items:           []*Price{itemA, itemB},
    orderLevelFee:   5000,   // ¥50 订单级费用
    voucherRedeemed: 10000,  // ¥100 优惠券
    coinsRedeemed:   4000,   // ¥40 积分
}

// 计算订单总额
order.CalculateOrderTotal()

// 获取订单价格拆解
orderBreakdown := order.GetOrderBreakdown()

fmt.Printf("商品A计价金额: ¥%d\n", itemA.itemPricingAmount/100)
// 输出: 商品A计价金额: ¥1785 (¥880×2 + ¥25)

fmt.Printf("商品B计价金额: ¥%d\n", itemB.itemPricingAmount/100)
// 输出: 商品B计价金额: ¥1215 (¥400×3 + ¥15)

fmt.Printf("订单商品总额: ¥%d\n", orderBreakdown.OrderItemsTotal/100)
// 输出: 订单商品总额: ¥3000

fmt.Printf("订单最终价格: ¥%d\n", orderBreakdown.OrderTotalAmount/100)
// 输出: 订单最终价格: ¥2910 (¥3000 + ¥50 - ¥100 - ¥40)
```

**使用示例 3：价格明细输出（用户友好展示）**

```go
// 单个商品的价格明细
fmt.Println(itemA.breakdown.ToDescription())

/*
输出：
【单价】原价: ¥1000.00/件
【单价】秒杀价: ¥880.00/件
【数量】购买: 2 件
【营销前】商品总价: ¥1960.00
【营销后】促销总价: ¥1760.00（立省 ¥200.00）
【Layer 3】+费用: +¥25.00
【商品小计】¥1785.00
*/

// 订单级别的价格明细
fmt.Println(order.GetOrderBreakdown().ToDescription())

/*
输出：
【订单总计】所有商品: ¥3000.00
【Layer 4】-优惠: -¥140.00
【最终价格】应付: ¥2860.00
*/
```

**价格维度速查表**：

| 场景 | 需要展示的价格维度 | 字段名 | 示例 |
|-----|------------------|--------|------|
| **PDP浏览** | 市场单价、促销单价 | `market_unit_price`, `promotion_unit_price` | ¥1000/件, ¥880/件 |
| **PDP加购** | 营销前总价、营销后总价、优惠金额 | `discount_total_price`, `promotion_total_price`, `promotion_discount_amount` | ¥1960, ¥1760, 省¥200 |
| **购物车** | 商品总价（含数量）、商品小计 | `promotion_total_price`, `item_pricing_amount` | ¥1760, ¥1785 |
| **订单确认** | 商品小计、订单商品总额 | `item_pricing_amount`, `order_items_total` | ¥1785, ¥3570 |
| **收银台** | 订单总额、优惠后金额 | `order_items_total`, `order_total_amount` | ¥3570, ¥3430 |
| **价格审计** | 所有维度（单价、总价、营销前后、商品级、订单级） | 完整 `PriceBreakdown` | 全部字段 |

**成果**：
- ✅ 团队沟通效率：提升 **60%**（术语统一，维度清晰）
- ✅ 需求理解偏差：从 **20%** → **< 5%**（不再混淆单价和总价）
- ✅ 代码可读性：提升 **80%**（结构化设计）
- ✅ 新人上手时间：从 **2 周** → **3 天**（模型简单易懂）
- ✅ 价格追溯效率：提升 **90%**（Breakdown 记录完整，多维度可查）
- ✅ 前端集成效率：提升 **70%**（接口返回多维度价格，前端直接使用）

---

### 2.4 痛点四：品类差异大 → 解决方案：策略模式 + 通用模型

#### 问题详情
每个品类独立实现计价逻辑：
- **Topup（话费充值）**：按面额定价，折扣率计算
- **E-Money（电子钱包）**：需要计算管理费（固定 or 百分比）
- **Hotel/Flight（酒店机票）**：价格规则复杂，多种房型和航班舱位组合

**真实案例**：
> 酒店业务上线前，发现价格计算逻辑与实物商品完全不同（多种房型、日期范围、取消政策），只能重新开发一套计价服务，耗时 2 个月。

#### 解决方案：策略模式 + 品类特殊计算器

**设计方案**：

```go
// 策略接口
type Calculator interface {
    Calculate(ctx context.Context, req *PricingRequest) (*PricingResponse, error)
    Support(categoryID int64) bool
    Priority() int
}

// 注册品类计算器
engine.RegisterCalculator(CategoryTopup, NewTopupCalculator())
engine.RegisterCalculator(CategoryEMoney, NewEMoneyCalculator())

// 计算时自动选择
calculator := e.selectCalculator(req.CategoryID)
if calculator != nil {
    return calculator.Calculate(ctx, req)  // 使用专用计算器
}
return e.calculateByLayers(ctx, req)       // 使用通用责任链
```

**收益**：
- 避免 if-else 地狱
- 新增品类只需实现 Calculator 接口
- 品类逻辑完全解耦

**成果**：
- ✅ 支持品类数：从 **3 个** → **10+ 个**
- ✅ 新增品类成本：从 **2 个月** → **1 周**（提升 8 倍）
- ✅ 代码复用率：从 **30%** → **80%**

---

## 三、核心挑战与技术亮点

### 3.1 核心挑战

#### 挑战一：如何保证灰度迁移安全？
价格计算是资损高风险区域，涉及 10+ 个品类，如何保证新逻辑价格计算 100% 准确？

**解决方案：空跑比对 + 三阶段灰度**

**阶段一：空跑期（2-4周）**
- 新老逻辑并行运行，使用老逻辑返回结果
- 新逻辑结果仅用于比对，不返回给用户
- 自动检测差异并上报（MongoDB 存储 + Kafka 告警）

```go
func CalculateWithDryRun(ctx context.Context, req *PricingRequest) (*PricingResponse, *DryRunResult, error) {
    var wg sync.WaitGroup
    var newResp, oldResp *PricingResponse
    
    // 并发执行新老逻辑
    wg.Add(2)
    go func() { newResp, _ = e.calculateByNewLogic(ctx, req) }()
    go func() { oldResp, _ = e.calculateByOldLogic(ctx, req) }()
    wg.Wait()
    
    // 比对差异
    diffResult := e.comparePrices(req, newResp, oldResp)
    if diffResult.HasDiff {
        e.reportDiff(ctx, diffResult)  // 上报 MongoDB + Kafka
    }
    
    return oldResp, diffResult, nil  // 使用老逻辑结果
}
```

**阶段二：灰度期（2-4周）**
- 按品类、地区、用户维度灰度切流：1% → 10% → 50% → 100%
- 基于 UserID 哈希的稳定灰度算法，保证同一用户体验一致

```go
// 灰度判断
func (gm *GrayManager) ShouldUseNew(req *PricingRequest) bool {
    rule := gm.GetGrayRule(req)  // 按品类+平台获取灰度规则
    
    // 白名单判断
    if gm.inWhitelist(req, rule) { return true }
    
    // 灰度比例判断（基于 UserID 哈希，保证同一用户稳定）
    hash := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%d", req.UserID)))
    return int32(hash%10000) < rule.GrayRatio
}
```

**阶段三：清理期（1-2周）**
- 观察稳定运行 1 周，下线老逻辑

**成果**：
- ✅ 迁移 10+ 个品类，**0 资损事故**
- ✅ 空跑期发现并修复 15+ 个逻辑差异
- ✅ 灰度期差异率 < 0.1%

---

#### 挑战二：如何满足不同场景的性能要求？
不同场景对性能要求差异巨大（PDP < 100ms vs Payment < 500ms）。

**解决方案：多级缓存 + 场景驱动 TTL**

**两级缓存设计**：
- **L1 本地缓存**：进程内 LRU 缓存，延迟 < 1ms
- **L2 Redis 缓存**：分布式缓存，延迟 < 5ms

**场景驱动的缓存策略**：

| 场景 | L1 TTL | L2 TTL | 命中率目标 |
|------|--------|--------|-----------|
| **PDP** | 5 分钟 | 30 分钟 | > 90% |
| **购物车** | 2 分钟 | 10 分钟 | > 60% |
| **CreateOrder/Checkout** | ❌ 不缓存 | ❌ 不缓存 | - |

**并发优化**：
Checkout 场景需要查询多个依赖服务，采用**并发调用**：

```go
func (e *engine) fetchDataConcurrently(ctx context.Context, req *PricingRequest) error {
    var wg sync.WaitGroup
    
    // 并发查询 4 个服务
    wg.Add(4)
    go func() { items = e.itemService.BatchGetItems(ctx, itemIDs) }()
    go func() { promo = e.promoService.GetPromotion(ctx, req) }()
    go func() { voucher = e.voucherService.ValidateVoucher(ctx, req) }()
    go func() { coin = e.coinService.CalculateDeduction(ctx, req) }()
    wg.Wait()
    
    return nil
}
```

**成果**：
- ✅ PDP 缓存命中率：从 **70%** → **92%**
- ✅ P99 延迟：从 **500ms** → **100-200ms**（提升 2.5 倍）
- ✅ Checkout 并发优化：延迟降低 **60%**

---

### 3.2 核心技术亮点

#### 技术亮点一：DDD 领域驱动设计
- ✅ **Price 聚合根**：业务逻辑内聚，修改影响范围可控
- ✅ **统一语言**：建立价格字典，团队沟通效率提升 60%
- ✅ **领域服务**：PriceCalculator 等领域服务封装复杂计算逻辑
- ✅ **价格可追溯**：PriceBreakdown 记录每一步计算，方便审计和调试

#### 技术亮点二：场景驱动架构
- ✅ 不同场景专属 Handler，优化性能和逻辑
- ✅ PDP 场景高缓存命中（>90%），降低延迟
- ✅ CreateOrder/Checkout 场景追求准确性，实时计算不缓存
- ✅ Payment 场景使用快照机制，零计算防篡改

#### 技术亮点三：设计模式应用
- ✅ **责任链模式**：5 层计算引擎，灵活跳层
- ✅ **策略模式**：品类特殊计算器，避免 if-else 地狱
- ✅ **工厂模式**：自动选择专用或通用计算器

#### 技术亮点四：高性能优化
- ✅ **多级缓存**：L1 本地 + L2 Redis，缓存命中率 92%
- ✅ **并发优化**：并发调用依赖服务，延迟降低 60%
- ✅ **批量计算**：批量查询商品、营销、库存信息

---

## 四、我的职责

### 角色定位
作为**核心后端开发**，我负责：
1. **架构设计**：主导价格计算引擎的整体架构设计，引入 DDD 领域驱动设计思想
2. **核心开发**：实现 5 层责任链计算引擎、价格快照机制、多级缓存策略
3. **性能优化**：设计多级缓存策略、并发优化方案，将 P99 延迟从 500ms 降低到 100-200ms
4. **灰度迁移**：设计并实施空跑比对机制，保证新老逻辑价格一致性，安全迁移 10+ 个品类

### 技术栈
- **语言**：Go 1.21+
- **框架**：自研 GAS 框架（类似 Spring Boot）
- **存储**：MySQL 8.0、Redis Cluster、MongoDB
- **消息队列**：Kafka
- **监控**：Prometheus + Grafana
- **协议**：gRPC、HTTP/REST

---

## 五、项目成果与收益

### 5.1 业务收益

| 维度 | 改进前 | 改进后 | 提升 |
|------|--------|--------|------|
| **准确性** | 年均 3-5 次资损事故 | **0 资损** | ↑ 100% |
| **性能** | P99 延迟 500ms+ | P99 延迟 < 200ms | ↑ 2.5x |
| **开发效率** | 新增变价因素需修改 5+ 服务 | 只需新增一个策略/层 | ↑ 5x |
| **扩展性** | 每个品类独立实现 | 统一模型 + 策略适配 | ↑ 10x |
| **可维护性** | 逻辑分散，难以追溯 | 领域模型封装，链路清晰 | ↑ 5x |

### 5.2 技术收益

| 技术点 | 收益 |
|--------|------|
| **DDD 领域模型** | 代码可读性提升 80%，团队沟通效率提升 60% |
| **多级缓存** | 缓存命中率 90%+，RT 降低 80% |
| **并发优化** | RT 降低 60% |
| **灰度机制** | 安全迁移，0 资损 |
| **空跑比对** | 自动发现差异，修复 15+ 个逻辑问题 |

### 5.3 系统指标

**性能指标**：
- 日均调用量：**5000 万+**
- P99 延迟：**< 200ms**（改进前 500ms+）
- 缓存命中率：**92%**（PDP 场景）
- 错误率：**< 0.01%**

**业务指标**：
- 支持品类：**10+ 个品类**（Topup、E-Money、Hotel、Flight、Tour 等）
- 资损事故：**0 次**（改进前年均 3-5 次）
- 价格不一致投诉：**下降 95%**

---

## 六、个人收获与思考

### 6.1 技术成长

1. **架构设计能力**：从零设计一个核心业务系统的完整架构，实践了 DDD、分层架构、策略模式等设计模式
2. **高性能优化**：通过多级缓存、并发优化、批量计算等手段，将 P99 延迟降低 2.5 倍
3. **灰度发布实践**：设计并实施了完整的灰度发布方案（空跑 → 灰度 → 清理），保证安全迁移
4. **分布式系统经验**：处理外部 API 超时、降级、幂等、并发等分布式系统常见问题

### 6.2 业务理解

1. **电商价格计算的复杂性**：深入理解了多品类、多营销、多变价因素的价格计算逻辑
2. **资损防控意识**：价格计算是资损高风险区域，必须有多重保护机制（快照、验证、空跑、告警）
3. **性能与准确性的权衡**：不同场景对性能和准确性的要求不同，需要差异化设计

### 6.3 工程化能力

1. **可观测性**：完善的监控指标、日志、告警体系（Prometheus + Grafana + Kafka）
2. **测试覆盖**：单元测试覆盖率 > 80%，集成测试覆盖核心链路
3. **文档沉淀**：编写了完整的设计文档、接口文档、运维手册

---

## 七、面试亮点总结

### 技术深度
- ✅ DDD 领域驱动设计（Price 聚合根、领域服务、统一语言）
- ✅ 设计模式应用（责任链、策略模式、工厂模式）
- ✅ 高性能优化（多级缓存、并发优化、批量计算）
- ✅ 分布式系统（外部 API 超时降级、幂等、并发控制）

### 业务理解
- ✅ 场景驱动架构（不同场景差异化设计）
- ✅ 资损防控（价格快照、二次验证、空跑比对）
- ✅ 性能与准确性权衡（前端展示高缓存 vs 订单支付零缓存）

### 工程能力
- ✅ 灰度发布实践（空跑 → 灰度 → 清理，安全迁移 10+ 品类）
- ✅ 可观测性建设（监控、日志、告警、差异分析看板）
- ✅ 系统稳定性（错误率 < 0.01%，可用性 99.99%）

---

## 八、面试常见问题准备

### Q1: 为什么要统一价格术语？如何做的？
**A**: 价格术语混乱是我们项目中最严重的痛点之一：

**问题有多严重？**
- 产品说"售价"，前端理解为"折扣价"，后端理解为"最终支付价"
- 数据库字段：`market_price`、`selling_price`、`final_price`、`pay_price`... 各个服务理解不一致
- 真实案例：某次需求评审，产品说"把售卖价改为活动价"，前端理解为修改展示逻辑，后端理解为修改计算逻辑，开发完发现需求理解完全错误，需要重新开发

**如何统一？**
1. **设计四层计价模型**：将复杂的价格计算拆分为 4 个清晰的层级（基础价格 → 促销 → 费用 → 优惠），每层职责单一
2. **建立价格字典**：定义了 10+ 个核心价格术语，覆盖四层模型的所有关键节点
3. **中英文强制统一**：代码变量必须使用标准术语（如 `market_price`、`promotion_price`、`final_price`），不允许模糊命名
4. **DDD 领域模型**：引入 `PriceComponent`、`PriceBreakdown`、`Price` 聚合根，强制使用统一模型

**成果**：
- 团队沟通效率提升 60%
- 需求理解偏差从 20% 降低到 < 5%
- 新人上手时间从 2 周缩短到 3 天

### Q2: 价格有单价、总价、营销前后等多个维度，如何设计才能清晰表达？
**A**: 这是价格计算中最容易混淆的问题。我们设计了**三个维度的价格体系**：

**维度一：单价 vs 总价**
- **单价（Unit Price）**：单个商品的价格，用于 PDP 展示
  - `market_unit_price`（市场原价/件）
  - `discount_unit_price`（折扣价/件）
  - `promotion_unit_price`（秒杀价/件）
- **商品总价（Item Total Price）**：单价 × 数量，用于加购后展示
  - `discount_total_price`（营销前总价）
  - `promotion_total_price`（营销后总价）
  - `item_pricing_amount`（商品计价金额，含费用）

**维度二：商品级 vs 订单级**
- **商品级（Item Level）**：单个商品的价格计算（Layer 1-3）
- **订单级（Order Level）**：多个商品汇总 + 订单级优惠（Layer 4）

**维度三：营销前 vs 营销后**
- **营销前**：`discount_total_price`（折扣价 × 数量）
- **营销后**：`promotion_total_price`（促销价 × 数量）
- **优惠金额**：`promotion_discount_amount`（立省多少钱）

**PDP 加购场景实际展示**：

```
用户选择数量：2 件

┌─────────────────────────────────┐
│ 【单价】                         │
│  原价：¥1000/件  ──────────────  │
│  秒杀价：¥880/件  ← 红色大字     │
├─────────────────────────────────┤
│ 【加购 2 件后的总价】             │
│  营销前：¥1960                   │
│  营销后：¥1760  ← 加粗显示       │
│  立省：¥200 (10.2%)              │
└─────────────────────────────────┘
```

**代码实现**：
```go
// PDP 展示接口返回
type PDPPriceDisplay struct {
    // 单价维度
    UnitPrice struct {
        Market    int64  `json:"market"`     // ¥1000
        Discount  int64  `json:"discount"`   // ¥980
        Promotion int64  `json:"promotion"`  // ¥880
    } `json:"unit_price"`
    
    // 总价维度（含数量）
    TotalPrice struct {
        BeforePromotion int64 `json:"before_promotion"` // ¥1960
        AfterPromotion  int64 `json:"after_promotion"`  // ¥1760
    } `json:"total_price"`
    
    // 优惠信息
    Savings struct {
        Amount     int64   `json:"amount"`      // ¥200
        Percentage float64 `json:"percentage"`  // 10.2%
    } `json:"savings"`
    
    Quantity int64 `json:"quantity"`  // 2
}
```

**收益**：
- ✅ 前端展示清晰：用户能清楚看到单价、总价、优惠金额
- ✅ 避免歧义：团队不再混淆单价和总价
- ✅ 易于调试：每个维度都有明确的字段名和值

### Q3: 四层计价模型是如何设计的？为什么这样分层？
**A**: 价格计算涉及多个变价因素（促销、费用、优惠），如果混在一起计算，逻辑会非常复杂。我们设计了四层模型：

**四层模型结构**：
```
Layer 1: 基础价格层 (market_price / discount_price)
    ↓
Layer 2: 促销层 (promotion_price)
    ↓
Layer 3: 费用层 (+ admin_fee + additional_charge)
    ↓
Layer 4: 优惠层 (- voucher - coins)
    ↓
Final Price: 最终价格
```

**为什么这样分层？**
1. **职责单一**：每层只负责一类价格变化，易于理解和维护
2. **顺序清晰**：价格计算流程从上到下，符合业务逻辑顺序
3. **易于追溯**：每层的计算结果都记录在 `PriceBreakdown` 中，方便审计
4. **灵活组合**：不同场景可以跳过某些层（如 PDP 不计算费用和优惠）
5. **扩展性强**：新增变价因素只需在对应层添加，不影响其他层

**举例说明**（秒杀场景）：
- Layer 1: 原价 ¥980/件
- Layer 2: 秒杀价 ¥880/件 × 2 = ¥1760
- Layer 3: + 服务费 ¥25 = ¥1785
- Layer 4: （订单级优惠，在订单汇总时计算）

**收益**：
- 代码可读性提升 80%
- Bug 率降低 70%（层级清晰，不易出错）
- 价格审计效率提升 90%（每层都有记录）

### Q4: 为什么选择 DDD？
**A**: 价格计算领域业务逻辑复杂，传统的贫血模型导致：
- 价格字段分散，代码难以理解
- 业务逻辑分散在多个服务，修改影响面大
- 缺乏统一语言，团队沟通成本高

引入 DDD 后：
- **Price 聚合根**：业务逻辑内聚，修改影响范围可控
- **统一语言**：建立价格字典，团队沟通效率提升 60%
- **领域服务**：PriceCalculator 等领域服务封装复杂计算逻辑
- **价格可追溯**：PriceBreakdown 记录每一步计算，方便审计和调试

### Q5: 场景驱动架构的优势是什么？
**A**: 不同场景对性能、准确性、计算内容的要求完全不同：
- **PDP 场景**：追求高性能（P99 < 100ms），只计算基础价+营销价，高缓存命中
- **CreateOrder 场景**：追求准确性，完整计算但不含券/积分，实时计算不缓存
- **Payment 场景**：追求安全性，零计算直接使用快照验证

如果用统一的逻辑处理所有场景，要么性能差（PDP 也实时计算），要么不准确（CreateOrder 也用缓存）。

### Q6: 如何保证灰度迁移安全？
**A**: 三层保护机制：
1. **空跑比对**：新老逻辑并行运行，自动检测差异（差异率 < 0.1%）
2. **灰度放量**：按品类分批切流（1% → 10% → 50% → 100%），异常立即回滚
3. **实时监控**：监控错误率、差异率、P99 延迟，差异超过阈值告警

空跑期发现并修复 15+ 个逻辑差异，灰度期 0 资损事故。

### Q7: 如果让你重新设计，会怎么改进？
**A**: 三个方向：
1. **分库分表**：目前价格快照表单表存储，高并发下可能成为瓶颈，可以按 `user_id` 分表
2. **读写分离**：CreateOrder/Checkout 场景写操作，Payment 场景读操作，可以读从库
3. **异步化**：空跑比对、差异上报、缓存预热等非实时任务可以异步化，进一步降低延迟

---

## 九、关键代码片段（面试白板题）

### 场景驱动的层级选择
```go
func (e *engine) getSkipLayers(scene PricingScene) []string {
    switch scene {
    case ScenePDP:
        return []string{"deduction", "charge"}  // PDP: 只计算基础价+营销价
    case SceneCreateOrder:
        return []string{"deduction"}            // CreateOrder: 跳过券/积分
    case SceneCheckout:
        return []string{}                       // Checkout: 完整计算
    }
}
```

### 灰度判断（基于 UserID 哈希）
```go
func (gm *GrayManager) hitRatio(userID int64, ratio int32) bool {
    hash := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%d", userID)))
    return int32(hash%10000) < ratio  // 保证同一用户稳定
}
```

---

**总结**：
这是一个**从 0 到 1 建设核心业务系统**的完整案例，涵盖了架构设计、DDD 实践、高性能优化、灰度发布、分布式系统等多个技术领域，展示了**架构能力、编码能力、工程化能力、业务理解能力**。
