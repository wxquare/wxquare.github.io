---
title: 计价中心项目面试材料
date: 2026-05-13
---

场景一：618 / 双11 运营“定时批量变价与控价”
如果没有 staging：
商家提前 3 天把 1000 款衣服的大促价（从 100 元改成 49 元）提交送审。QC 小二在 61月2日 审核通过了。如果只有 draft，审核通过就必须立刻合流到 item_tab，结果大促还没开始，前台直接按 49 元低价开卖，产生巨大的价格资损。

引入 staging 后：
QC 小二在 6月2日 审核通过后，数据原子合流到 staging_tab（预发暂存表），打上生效时间戳 effect_time = 2026-06-03 24:00:00。
线上核心表 item_tab 依然保持 100 元正常售卖。等到 24:00 钟声敲响，全网秒杀启动，商品中心利用高并发定时调度器（如 XXL-JOB 分片），在毫秒级执行一个极轻量的本地事务，将 staging 数据瞬间激活（Flip）到 item_tab，并秒级双删 Redis 缓存。全网变价在 50ms 内完成，没有一丁点审核延迟。



# 计价中心项目面试材料

本文是计价中心项目的最终单文件版，已合并并重组以下三份材料的有效内容：

- `pricing-engine-interview-questions.md`
- `project-pricing-engine-interview.md`
- `project-pricing-engine-model-optimized.md`

目标不是把原文机械拼接，而是整理成一份能用于面试复盘、项目介绍、白板讲解和追问回答的材料。面试时建议先讲业务问题和系统价值，再根据追问展开模型、快照、灰度、性能、测试、项目管理和具体 Bug。

---

## 一、项目一句话

在 Shopee Digital Purchase 数字电商平台中，建设统一计价中心，将商品基础价、营销优惠、平台费用、券/积分抵扣等分散在商品、营销、订单、支付等系统中的计价逻辑收敛为场景化计算能力，支撑列表导购价、详情确认价、加购试算价、订单锁定价和收银台支付价，提升多品类接入效率、价格可解释性和交易金额一致性。

---

## 二、30 秒项目介绍

Digital Purchase 同时支持充值、账单缴费、电子券、礼品卡、酒店、机票、电影票等多种数字商品。早期价格逻辑分散在多个服务和多个品类实现里，列表价、详情价、订单价、支付价的口径不完全统一，新增变价因素需要改多个系统，价格问题也容易演变成资损风险。

我参与建设统一计价中心，按 PDP、Cart、Order、Checkout、Payment 等场景拆分计价职责，沉淀基础价、促销、费用、优惠抵扣的分层模型，并通过品类策略、价格明细、订单/支付快照、空跑比对和灰度迁移保证正确性。项目支撑 10+ 数字商品品类和日均 5000 万+ 高频调用，关键接口 P99 从 500ms 优化到 200ms 以内，项目上线后未造成资损事故。

---

## 三、2 分钟项目介绍

这个项目的背景是数字电商的价格计算越来越复杂。一方面业务覆盖充值、账单、券类、酒店、机票、电影票等多品类，不同品类有不同的基础价、促销规则、手续费、供应商加价和库存实时性要求；另一方面价格计算贯穿完整购买链路，用户在列表、详情、加购、创单、收银台、支付看到或使用的价格并不完全一样。

早期系统的问题主要有四类：

1. **计价逻辑分散**：商品、营销、订单、支付、前端都有价格相关逻辑，新增一个变价因素往往要改多个服务。
2. **价格术语混乱**：市场价、折扣价、促销价、售卖价、应付金额等概念在不同团队和代码字段中口径不一致。
3. **品类差异大**：券类、充值、酒店、电影票等品类各自实现计价规则，代码复用率低，新品类接入成本高。
4. **金额一致性风险**：展示价、订单价、支付价如果重算口径不同，容易出现用户投诉、支付金额不一致甚至资损。

解决思路是把计价中心设计成场景驱动的统一计算能力，而不是一个万能大接口。PDP 和列表侧强调性能，只计算基础价和促销展示价；Cart 做加购试算；Order 阶段实时计算并固化订单价格；Checkout 阶段结合支付渠道、券、积分、手续费生成支付快照；Payment 阶段不再重新计算价格，只做快照和金额校验。

在模型上，我们将价格拆成基础价格、促销价格、费用、优惠抵扣四个业务层，并保留 Final 汇总层。对于多品类差异，通用逻辑放在 BaseCalService 中，Home Service、Event Ticket、EVoucher 等复杂品类通过专用 Calculator 承载。为了让价格可解释和可追溯，我们设计 PriceBreakdown 记录每一层计算结果、命中的促销、费用明细和最终金额。

上线过程中，价格计算属于高风险链路，所以采用空跑比对和灰度迁移：新老逻辑并行运行，先用老逻辑返回结果，新逻辑只做差异比对；差异收敛后再按品类、地区、用户比例逐步切流。项目最终将新增变价因素的上线周期从 2-3 周缩短到 3-5 天，P99 从 500ms 优化到 200ms 以内，并且项目上线后未造成资损事故。

---

## 四、一页速记卡

### 业务问题

- 多品类：充值、账单、券类、酒店、机票、电影票、活动票等。
- 多场景：Search/List、PDP、Cart、Booking/Order、Checkout、Payment、Refund。
- 多变价因素：基础价、促销价、手续费、服务费、供应商加价、优惠券、积分、支付渠道优惠。
- 主要风险：价格口径不统一、重复计算、展示价和实扣价不一致、灰度迁移导致资损。

### 核心方案

- **场景驱动 API**：PDP、Cart、Order、Checkout、Payment 分开设计。
- **四层计价模型**：Base Price → Promotion → Fee → Discount/Voucher → Final。
- **品类策略扩展**：BaseCalService + Category Calculator，复杂品类独立实现。
- **价格快照**：订单快照锁定创单金额，支付快照锁定收银台金额。
- **空跑灰度**：新老逻辑并行比对，按品类/地区/用户灰度切流。
- **可观测与对账**：差异率、错误率、P99、缓存命中率、异常金额监控。

### 可以强调的结果

- 支撑 10+ 数字商品品类。
- 日均 5000 万+ 高频调用。
- 关键接口 P99 从 500ms 优化到 <200ms。
- 新增变价因素从 2-3 周缩短到 3-5 天。
- 项目上线后未造成资损事故。

---

## 五、为什么要做统一计价中心

### 1. 价格逻辑分散

早期价格计算分散在多个系统中：

- 商品服务维护基础价格、折扣价、商品费用。
- 营销服务维护活动价格、限购、Bundle、新用户价。
- 订单服务负责创单总价、商品明细和价格校验。
- 支付/收银台负责券、积分、手续费和支付渠道相关金额。
- 前端也保留了一部分展示和试算逻辑。

这会导致新增变价因素时需要同时改多个服务。例如新增一个服务费或新的促销规则，需要前端展示、详情试算、订单创建、支付确认都同步调整，开发、测试和灰度成本都很高。

### 2. 价格术语不统一

价格相关概念非常多：

- 市场价、原价、划线价。
- 折扣价、日常销售价。
- 促销价、活动价、新用户价、秒杀价。
- 计价金额、售卖价、订单总额、支付金额。
- 服务费、手续费、供应商加价、券抵扣、积分抵扣。

如果没有统一语言，产品、前端、后端、QA 对同一个词可能有不同理解。面试时可以举例：

> 产品说“售价”，前端可能理解为详情页展示价，订单侧可能理解为创单金额，支付侧可能理解为实扣金额。如果字段和语义不统一，需求评审和线上排查都会非常痛苦。

### 3. 多品类差异大

不同品类的计价逻辑差异很大：

| 品类 | 典型计价特点 |
| --- | --- |
| 充值/账单 | 面额、账单金额、手续费、部分支付规则 |
| EVoucher/GiftCard | 券面额、库存、Bundle、购买限制 |
| Movie/Event | 场次、座位、套餐、新用户价、秒杀、限购 |
| Hotel/Flight | 供应商实时价格、加价、库存变化、Booking Token |
| Home Service | 最低交易金额、部分商品是否计入最低交易金额 |

如果每个品类各写一套计价逻辑，短期看上线快，长期会造成重复实现、测试困难和逻辑不一致。

### 4. 金额一致性要求高

价格链路和资金直接相关。列表价可以短暂不准，但订单价和支付价必须一致。计价中心要解决的是：

- 用户看到的价格为什么是这个数。
- 下单时金额是否和后端计算一致。
- 收银台选择券、积分、支付渠道后，最终支付价是否可追溯。
- 支付时是否允许重新计算，还是只能验证快照。
- 退款和客服解释时，能否还原当时的价格明细。

---

## 六、总体架构

### 1. 场景驱动，而不是万能接口

价格计算贯穿完整购买链路，不同场景对性能、准确性和计算内容要求不同：

| 场景 | 目标 | 计算内容 | 缓存策略 | 结果定位 |
| --- | --- | --- | --- | --- |
| Search/List | 快速导购 | 候选商品价格、最低价、可售标记 | 高缓存 | 导购价格 |
| PDP/Detail | 展示准确 | 基础价、促销价、部分实时库存/规则 | 中高缓存 | 详情确认价 |
| Cart | 试算 | 数量、促销、预估优惠、可售校验 | 中缓存 | 加购试算价 |
| Booking/Order | 创单安全 | 实时价格、库存/资源确认、费用 | 不缓存 | 订单锁定价 |
| Checkout | 收银台确认 | 券、积分、手续费、支付渠道 | 不缓存 | 收银台支付价 |
| Payment | 防篡改 | 快照验证，原则上零重算 | 不缓存 | 实扣金额 |

面试表达：

> 我们没有设计一个“大而全”的价格 API，而是按购买链路拆场景。展示场景可以缓存和降级，交易场景必须实时计算和固化快照，支付场景尽量不重算，只校验快照，这样才能同时满足性能和资金安全。

### 2. 四层业务模型加 Final 汇总层

计价模型可以这样讲：

```text
Layer 1: Base Price
  market_price / discount_price / original_discount_price

Layer 2: Promotion
  flash_sale / new_user_price / bundle_price / scheduling

Layer 3: Fee
  hub_admin_fee / additional_charge / service_fee / handling_fee / mark_up

Layer 4: Discount
  voucher_redeemed / coins_redeemed / payment_discount

Final:
  order_total_amount / payment_amount / price_breakdown
```

为什么这样分层：

- 符合业务顺序：先有商品价格，再有营销，再加费用，再做订单级优惠。
- 易于解释：每一层都有输入、输出和明细。
- 易于扩展：新增费用进 Fee Layer，新增券抵扣进 Discount Layer。
- 易于测试：每一层可以独立构造测试用例。

### 3. 品类策略模型

设计上使用通用计算器加品类计算器：

```go
type Calculator interface {
    CalPDPPrice(req *CalPDPPriceRequest) (*CalPDPPriceResponse, error)
    CalOrderPrice(req *CalOrderPriceRequest) (*CalOrderPriceResponse, error)
    CalPayPrice(req *CalPayPriceRequest) (*CalPayPriceResponse, error)
}
```

通用逻辑放在 BaseCalService：

- 无促销时使用折扣价乘数量。
- 有订单促销时使用促销结果。
- 填充 PricingAmount 和 SellingPriceInfo。
- 做基础金额校验。

复杂品类单独实现：

- Home Service：最低交易金额和 `sub_charge`。
- Event Ticket：FlashSale、BundlePrice、NewUserPrice、限购。
- EVoucher：券码库存、Bundle、购买限制。
- Hotel/Flight：供应商报价、加价、Booking Token。

面试表达：

> 我的原则是通用逻辑尽量沉到 BaseCalService，品类差异收敛到 Calculator。这样既避免通用逻辑里堆满 category if-else，也不会为了统一而牺牲品类表达能力。

---

## 七、核心设计一：场景化计价 API

### 1. PDP/Detail 计价

目标是支撑详情页展示和用户调整数量时的价格试算。

特点：

- 主要计算基础价和促销价。
- 支持一定缓存。
- 不做最终券/积分抵扣。
- 不锁定资源。
- 返回前端需要展示的多维价格。

典型返回维度：

```go
type PDPPriceDisplay struct {
    UnitPrice struct {
        Market    int64 `json:"market"`
        Discount  int64 `json:"discount"`
        Promotion int64 `json:"promotion"`
    } `json:"unit_price"`

    TotalPrice struct {
        BeforePromotion int64 `json:"before_promotion"`
        AfterPromotion  int64 `json:"after_promotion"`
    } `json:"total_price"`

    Savings struct {
        Amount     int64   `json:"amount"`
        Percentage float64 `json:"percentage"`
    } `json:"savings"`

    Quantity int64 `json:"quantity"`
}
```

### 2. Order 计价

目标是创建订单时锁定订单金额。

特点：

- 不使用缓存。
- 必须做商品、促销、费用的实时计算。
- 需要填充 `PricingAmount`、`SellingPriceInfo`、`ExtraInfo`。
- 生成订单价格快照。
- 和前端提交金额做一致性校验。

### 3. Checkout/Payment 计价

Checkout 是收银台阶段，用户选择支付渠道、券、积分、手续费规则后，生成最终支付金额。

Payment 阶段尽量不重新计算，只校验支付快照：

```go
func ValidatePaymentSnapshot(snapshot *PaymentSnapshot, req *PayRequest) error {
    if snapshot == nil {
        return ErrSnapshotExpired
    }
    if snapshot.OrderID != req.OrderID {
        return ErrOrderMismatch
    }
    if snapshot.Status == SnapshotUsed {
        return ErrDuplicatePayment
    }
    if snapshot.FinalAmount != req.PayAmount {
        return ErrAmountMismatch
    }
    return nil
}
```

面试追问：

- 为什么 Payment 不重新计算？
- 如果收银台快照过期怎么办？
- 用户换支付渠道后金额变了怎么办？
- Order 快照和 Payment 快照分别解决什么问题？

回答重点：

> Payment 重新计算会引入规则变更、券状态变化、手续费变化等不确定性。支付阶段应该验证用户即将支付的金额是否等于收银台快照，而不是重新推导一次价格。

---

## 八、核心设计二：价格术语和领域模型

### 1. 核心术语表

| 层级 | 中文术语 | 英文字段 | 定义 | 典型场景 |
| --- | --- | --- | --- | --- |
| Base | 市场原价 | `market_price` | 商品标价，常用于划线价 | PDP |
| Base | 折扣价 | `discount_price` | 无促销时的日常销售价 | PDP/Order |
| Base | 原始折扣价 | `original_discount_price` | 参与计算的兜底价 | Order |
| Promotion | 促销价 | `promotion_price` | 秒杀、新用户价、Bundle 后价格 | PDP/Cart |
| Promotion | 促销优惠金额 | `promotion_discount_amount` | 折扣价和促销价差额 | PDP |
| Fee | 服务费 | `service_fee` | 商品或订单级服务费用 | Order/Checkout |
| Fee | 手续费 | `handling_fee` | 支付渠道或收银台费用 | Checkout |
| Discount | 券抵扣 | `voucher_redeemed` | 优惠券抵扣金额 | Checkout |
| Discount | 积分抵扣 | `coins_redeemed` | 积分抵扣金额 | Checkout |
| Final | 订单总额 | `order_total_amount` | 订单最终金额 | Order/Payment |

### 2. 单价、商品总价、订单总价

面试时一定要讲清楚价格维度：

- **单价**：一个商品单位的价格，例如 `market_price`。
- **商品总价**：单价乘数量，例如 `promotion_price * quantity`。
- **订单总价**：多个商品汇总后，再加费用、减订单级优惠。

很多价格 Bug 的根源是把单价、商品总价和订单总价混在一起。

### 3. PriceBreakdown

PriceBreakdown 的价值是可解释和可追溯：

```go
type PriceBreakdown struct {
    ItemID     int64
    CategoryID int64
    Quantity   int64

    MarketPrice           int64
    DiscountPrice         int64
    OriginalDiscountPrice int64

    PromotionType       string
    PromotionPrice      int64
    PurchaseLimit       int64
    PromotionAppliedQty int64
    PromotionNormalQty  int64

    DiscountTotalPrice      int64
    PricingAmount           int64
    PromotionDiscountAmount int64

    HubAdminFee        int64
    DpAdditionalCharge int64
    MarkUp             int64

    VoucherRedeemed int64
    CoinsRedeemed   int64
    FinalAmount     int64

    Scene     string
    RequestID string
}
```

应用场景：

- 客服解释价格。
- 退款金额追溯。
- 新老逻辑差异对比。
- 价格审计和资损排查。
- QA 构造测试用例。

面试表达：

> 我们不只返回一个 final price，而是返回每层价格明细。这样当用户问为什么支付这个金额，或者新老逻辑差异时，可以快速定位是基础价、促销、费用还是券抵扣导致的。

---

## 九、核心设计三：快照和金额一致性

### 1. 为什么需要快照

价格在多个阶段都会变化：

- 商品基础价可能被运营修改。
- 促销可能结束或限购被消耗。
- 供应商价格和库存可能变化。
- 券、积分、手续费可能因用户选择变化。

如果每个阶段都重新计算，就会出现：

- 展示价和订单价不一致。
- 订单价和支付价不一致。
- 用户支付后难以解释当时为什么是这个金额。
- 退款时无法还原原始优惠分摊。

### 2. 双快照设计

```text
Order Snapshot
  生成时机：Order 创建阶段
  内容：商品信息、基础价、促销价、费用、资源确认结果、履约/售后规则
  用途：锁定订单价格和交易上下文

Payment Snapshot
  生成时机：Checkout 收银台阶段
  内容：订单金额、券/积分抵扣、支付渠道、手续费、最终支付金额
  用途：支付发起和支付回调时校验金额
```

### 3. 防资损链路

可以总结为五道防线：

1. **统一计算入口**：减少多个服务重复计算。
2. **金额校验**：前端提交金额和后端计算结果做一致性校验。
3. **订单/支付快照**：交易阶段使用快照，减少重算风险。
4. **空跑比对**：新老逻辑并行，差异告警。
5. **灰度和异常拦截**：差异超过阈值停止放量或回滚。

面试追问：

- 如果订单快照和支付快照金额不同怎么办？
- 快照过期后是否允许重新计算？
- 支付回调金额和快照金额不一致怎么办？
- 快照表如何分库分表？

回答重点：

> 快照不是为了阻止业务变化，而是为了锁定某一次交易的上下文。价格可以继续变化，但已经进入订单和支付阶段的交易必须基于当时确认的价格推进。

---

## 十、核心设计四：多品类差异处理

### 1. 策略模式

通用结构：

```go
type BaseCalService struct {
    ExtraInfo        map[string]string
    PdpOtherPriceMap map[string]int64
}

type HomeServiceCalService struct {
    *BaseCalService
}

type EventTicketCalService struct {
    *BaseCalService
}
```

选择逻辑：

```go
func SelectCalculator(categoryID int64) Calculator {
    if calculator, ok := categoryCalculators[categoryID]; ok {
        return calculator
    }
    return baseCalculator
}
```

设计原则：

- 简单品类走通用逻辑。
- 复杂品类用独立 Calculator。
- 品类差异不要污染 BaseCalService。
- 通用模型必须能表达单价、总价、费用、优惠、快照。

### 2. Home Service 最低交易金额

业务规则：

- 部分商品计入最低交易金额。
- 部分商品不计入最低交易金额。
- 计入部分不足最低交易金额时，需要补 `sub_charge`。

公式：

```go
includedAmount := int64(0)
excludedAmount := int64(0)

for _, item := range items {
    amount := item.Price * item.Quantity
    if item.ExcludeMinimumValue {
        excludedAmount += amount
    } else {
        includedAmount += amount
    }
}

subCharge := max(0, minTransaction - includedAmount)
total := excludedAmount + max(minTransaction, includedAmount)
```

示例：

```text
最低交易金额：60000
商品 A，计入最低交易金额：35000
商品 B，排除最低交易金额：30000

sub_charge = max(0, 60000 - 35000) = 25000
total = 30000 + max(60000, 35000) = 90000
```

面试可以讲的 Bug：

> 曾经 CalPDPPrice 和 CalOrderPrice 对最低交易金额的公式不一致，导致多商品混合场景下 `sub_charge` 表达不一致。修复方式是将两个场景统一为 `excludedAmount + max(minTransaction, includedAmount)`，并补充全部计入、全部排除、等于最低金额、超过最低金额等边界测试。

### 3. Event Ticket 促销计价

支持多种促销：

- FlashSale：促销价，受限购约束。
- BundlePrice：满 N 件折扣或固定金额优惠。
- NewUserPrice：新用户价，通常限购 1。
- Scheduling：普通活动价。

通用逻辑：

```go
func calculateEventFinalPrice(promotion Promotion, item PriceEntity) (originalAmount, finalAmount int64) {
    quantity := item.Quantity

    originalDiscountPrice := item.OriginalDiscountPrice
    if originalDiscountPrice == 0 {
        originalDiscountPrice = item.DiscountPrice
    }

    originalAmount = originalDiscountPrice * quantity

    switch promotion.Type {
    case BundlePrice:
        return originalAmount, calculateBundlePrice(promotion, item)
    case FlashSale, Scheduling, NewUserPrice:
        return originalAmount, calculateLimitedPromotionPrice(promotion, item)
    default:
        return originalAmount, originalAmount
    }
}
```

面试可以讲的 Bug：

> 有一次线上场景中 `OriginalDiscountPrice` 没有传，Protobuf 默认值是 0，计价逻辑直接使用这个字段导致金额为 0。修复方式是在所有入口统一做兜底：`OriginalDiscountPrice == 0` 时使用 `DiscountPrice`，并补充字段缺失、数量为 0、无促销等测试。

### 4. BundlePrice

典型问题：满 N 件享折扣，购买数量可能不是 N 的整数倍，且存在最大优惠轮数。

示例：

```text
商品单价：50
活动：买 3 件 8 折，最多 5 轮
用户购买：10 件

满足轮数：10 / 3 = 3 轮，剩余 1 件
每轮优惠：50 * 3 * 20% = 30
最终金额：50 * 10 - 30 * 3 = 410
```

关键点：

- 金额统一使用整数，避免浮点误差。
- 处理购买数量小于最小捆绑数。
- 处理超过最大优惠轮数。
- 处理尾差，通常最后一件或最后一个分摊对象承接尾差。
- 不同地区可能有不同精度规则。

### 5. 复杂计价场景拆解

面试官如果追问“计价中心最复杂的场景是什么”，不要只回答“促销很多”或者“品类很多”。更好的回答方式是把复杂性拆成：**价格来源复杂、促销叠加复杂、费用复杂、优惠分摊复杂、交易阶段复杂、逆向退款复杂**。

下面这些场景可以作为面试展开材料。

#### 场景一：商品级促销 + 订单级优惠 + 支付手续费

这是最典型的电商订单计价场景。

```text
订单包含 2 个商品：

商品 A：
  原价：1000
  日常折扣价：900
  秒杀价：800
  数量：2
  商品服务费：20

商品 B：
  原价：500
  日常折扣价：480
  无商品级促销
  数量：1
  商品服务费：10

订单级优惠：
  优惠券：100
  积分抵扣：50

支付费用：
  支付手续费：30
```

计算过程：

```text
商品 A：
  促销后商品金额 = 800 * 2 = 1600
  商品小计 = 1600 + 20 = 1620

商品 B：
  商品金额 = 480 * 1 = 480
  商品小计 = 480 + 10 = 490

订单商品总额 = 1620 + 490 = 2110
优惠后金额 = 2110 - 100 - 50 = 1960
最终支付金额 = 1960 + 30 = 1990
```

这个场景的关键不是公式，而是**每个金额属于哪一层**：

- 秒杀价属于商品级促销。
- 商品服务费属于 Fee Layer。
- 优惠券和积分属于订单级优惠。
- 支付手续费属于收银台阶段费用。
- Payment 阶段应使用 Checkout 快照校验 `1990`，而不是重新计算。

面试追问：

- 如果用户换支付渠道，手续费变化怎么办？
- 如果优惠券在支付前过期怎么办？
- 如果支付阶段重新算出 1980，系统应该相信哪个金额？

回答重点：

> Checkout 阶段会基于用户选择的券、积分和支付渠道生成支付快照。用户更换支付渠道时，需要重新生成支付快照；Payment 阶段只校验快照金额，不重新推导。

#### 场景二：BundlePrice + 限购 + 超限原价

这类场景复杂在：优惠不是简单单价替换，而是和购买数量、轮数、限购有关。

```text
商品单价：100
活动规则：满 3 件减 60，最多 2 轮
用户购买：8 件
```

计算过程：

```text
可享受优惠轮数 = min(8 / 3, 2) = 2
优惠金额 = 60 * 2 = 120
原始金额 = 100 * 8 = 800
最终金额 = 800 - 120 = 680
```

如果再叠加“促销库存只剩 5 件”，则需要继续拆：

```text
促销可用数量：5
用户购买数量：8
促销部分：5 件
原价部分：3 件

如果规则是每 3 件减 60：
  促销部分可形成 1 轮
  优惠金额 = 60
  最终金额 = 100 * 8 - 60 = 740
```

关键问题：

- 轮数按购买数量算，还是按促销可用数量算？
- 超过限购的部分使用什么价格？
- 如果购买数量不是 N 的整数倍，剩余部分怎么处理？
- 百分比折扣如何做精度和取整？

回答重点：

> BundlePrice 不能只看促销单价，要同时考虑 minQuantity、maxRound、purchaseLimit、promotion stock 和地区精度。我们会在 PriceBreakdown 中记录促销适用数量、原价数量和总优惠金额，方便退款和排查。

#### 场景三：新用户价 + 购买多件

NewUserPrice 看起来简单，但很容易被误算。

```text
原价：500
新用户价：350
新用户限购：1
用户购买：3 件
```

正确计算：

```text
优惠价部分 = 350 * 1
原价部分 = 500 * 2
最终金额 = 1350
```

错误做法：

```text
350 * 3 = 1050
```

这个错误会直接造成资损，因为把新用户优惠应用到了所有购买数量上。

面试回答重点：

> 对限购类促销，我会把数量拆成 `promotionAppliedQty` 和 `promotionNormalQty`。前者使用促销价，后者回退到原始折扣价，并在明细中记录，避免退款和审计时说不清楚。

#### 场景四：最低交易金额 + 排除商品

Home Service 的难点在于：有些商品计入最低交易金额，有些商品不计入。

```text
最低交易金额：600

商品 A：安装服务，350，计入最低交易金额
商品 B：材料费，300，不计入最低交易金额
```

正确计算：

```text
计入金额 included = 350
排除金额 excluded = 300
补足金额 sub_charge = max(0, 600 - 350) = 250

最终金额 = excluded + max(600, included)
        = 300 + 600
        = 900
```

容易出错的公式：

```text
max(included + excluded, minTransaction)
= max(350 + 300, 600)
= 650
```

这个公式错在把排除商品也拿去满足最低交易金额了。

面试回答重点：

> 这个场景的本质是业务集合划分，不是数学公式。先分清哪些商品可以用来满足最低交易金额，再计算补足金额。PDP 和 Order 必须共用同一套公式，否则用户看到的补足金额和创单金额会不一致。

#### 场景五：供应商实时价格 + Booking 锁价

酒店、机票、船票这类资源确认型商品，价格和库存由供应商实时决定。

```text
列表页：
  展示缓存价格 1000

详情页：
  刷新供应商价格 1020

Booking：
  供应商确认价 1030
  返回 booking token，有效期 10 分钟

Order：
  使用 booking token 和确认价创建订单快照

Checkout：
  加上手续费、券、积分，生成支付快照

Payment：
  校验支付快照，不重新查供应商价格
```

这个场景的难点：

- 列表价、详情价、Booking 价可能不同。
- 供应商查价可能超时或返回库存不足。
- Booking Token 有有效期。
- 用户支付超时后要释放资源。
- 支付成功后履约可能失败，需要退款或补偿。

面试回答重点：

> 对供应商品类，我们不追求列表价绝对准确，而是明确“列表快、详情准、创单安全”。真正进入交易前，在 Booking 阶段实时确认价格和资源，并把确认结果写入订单快照。

#### 场景六：合并支付 + 分单履约 + 部分退款

这个场景复杂在订单模型，而不只是价格模型。

```text
用户一次支付两个商品：

商品 A：话费充值，金额 100
商品 B：碎屏险，金额 20

订单级优惠券：10
最终支付金额：110
```

优惠分摊：

```text
商品 A 金额占比 = 100 / 120
商品 B 金额占比 = 20 / 120

商品 A 分摊优惠 = 8.33，按最小货币单位取整
商品 B 分摊优惠 = 1.67
尾差由最后一个商品承接
```

如果商品 B 履约失败，需要部分退款：

```text
退款金额不能直接退 20
应该退商品 B 的实付金额：
  商品 B 原金额 - 商品 B 分摊优惠
```

关键问题：

- 一个支付单对应多个订单。
- 每个订单可能有自己的履约状态。
- 订单级优惠要分摊到订单项。
- 部分退款必须基于分摊后的实付金额，而不是原价。
- 分摊结果要落库，不能退款时重新算。

面试回答重点：

> 只要支持合并支付和部分退款，优惠分摊就必须在正向交易时完成并落库。否则逆向退款时规则变了、券过期了、商品价格变了，都无法还原当时的实付金额。

#### 场景七：券、积分和支付渠道优惠的叠加顺序

收银台阶段常见问题是多个优惠来源叠加：

```text
订单商品金额：1000
平台券：100
积分抵扣：50
支付渠道优惠：20
支付手续费：10
```

一种明确的计算顺序：

```text
订单金额 = 1000
券后金额 = 1000 - 100 = 900
积分后金额 = 900 - 50 = 850
支付渠道优惠后 = 850 - 20 = 830
最终支付金额 = 830 + 10 = 840
```

复杂点：

- 券和积分是否都允许叠加。
- 支付渠道优惠是优惠还是费用减免。
- 手续费是在优惠前加，还是优惠后加。
- 最终金额不能小于 0。
- 出资方不同，财务对账口径不同。

面试回答重点：

> 优惠叠加顺序必须产品、财务、研发一起定义清楚，并固化在计价模型和 PriceBreakdown 中。否则同样是 20 元支付优惠，不同系统可能理解成减少用户支付、平台补贴或手续费减免，对账会出问题。

#### 场景八：逆向退款金额计算

退款金额不是简单退原价，而要看正向交易时用户实际支付了多少。

```text
商品 A：500
商品 B：300
商品 C：200
订单级满减：100
用户实付：900
```

按比例分摊：

```text
A 分摊优惠：50，实付 450
B 分摊优惠：30，实付 270
C 分摊优惠：20，实付 180
```

用户只退商品 B：

```text
退款金额 = 270
```

进一步复杂的情况：

- 退后订单不再满足满减门槛，是否追回优惠？
- 支付渠道优惠是否可退？
- 积分抵扣是退积分还是退现金？
- 手续费是否退？
- 部分履约失败是自动退款还是人工审核？

面试回答重点：

> 退款金额应该基于正向订单的价格明细和分摊结果，而不是根据当前规则重新计算。计价中心需要提供可追溯的 PriceBreakdown，订单/退款系统基于快照推进逆向流程。

#### 场景总结

复杂计价的本质可以总结成四句话：

- **展示价可以缓存，交易价必须确认。**
- **商品级优惠先算，订单级优惠要分摊。**
- **正向金额要落快照，逆向退款要按快照。**
- **所有金额都要能解释、能追溯、能对账。**

---

## 十一、核心设计五：性能优化

### 1. 场景化缓存

不是所有场景都缓存：

| 场景 | 缓存策略 | 原因 |
| --- | --- | --- |
| Search/List | 强缓存，多级缓存 | 高 QPS，允许短暂不一致 |
| PDP/Detail | 多级缓存，短 TTL | 展示价需要快，同时尽量准 |
| Cart | 短 TTL，必要时刷新 | 试算要兼顾性能和准确性 |
| Order | 不缓存 | 创单价格必须准确 |
| Checkout | 不缓存 | 支付价必须准确 |
| Payment | 快照验证 | 防止支付阶段重算 |

### 2. 多级缓存

```text
L1 本地缓存
  延迟低，适合热点商品和规则

L2 Redis 缓存
  跨实例共享，适合促销、商品基础价、配置

L3 DB / 依赖服务
  权威数据源
```

优化手段：

- 本地缓存 + Redis 缓存。
- Redis Pipeline / MGet 批量获取。
- DB 批量查询，避免 N 次 RPC。
- 热点数据预热。
- TTL 加随机偏移，降低缓存雪崩。
- Singleflight 合并同 key 并发请求。
- 依赖服务并发查询。

### 3. 批量计算

列表页或推荐页可能一次计算几十到上百个商品价格。

优化思路：

- 批量拉取商品、促销、库存和配置。
- 控制并发度，例如 10 个 worker。
- 限制单次请求最大 item 数。
- 对失败 item 做局部降级，避免拖垮整批。

面试表达：

> 计价性能优化不是一味加缓存，而是按场景区分。展示场景可以缓存，交易场景不能缓存。列表页主要靠批量拉取和多级缓存，Checkout 主要靠并发查询和快照减少重算。

---

## 十二、核心设计六：灰度迁移和空跑比对

### 1. 为什么迁移风险高

计价系统和资金强相关，重构不能只靠单元测试保证正确性。历史数据、品类差异、隐藏规则、前端旧逻辑都可能导致新老结果不同。

### 2. 空跑比对

空跑期的原则是：老逻辑返回给用户，新逻辑只用于比对。

```go
func CalculateWithDryRun(ctx context.Context, req *PricingRequest) (*PricingResponse, error) {
    oldResp := oldEngine.Calculate(ctx, req)

    go func() {
        newResp := newEngine.Calculate(ctx, req)
        diff := ComparePrice(oldResp, newResp)
        if diff.HasDiff {
            ReportPriceDiff(req, oldResp, newResp, diff)
        }
    }()

    return oldResp, nil
}
```

比对维度：

- 最终金额是否一致。
- 每个商品 `pricing_amount` 是否一致。
- 促销命中是否一致。
- 费用和优惠明细是否一致。
- 差异金额和差异比例是否超过阈值。

### 3. 灰度发布

灰度维度：

- 品类：先简单品类，后复杂品类。
- 地区：先单市场，后多市场。
- 用户：基于 user_id hash 稳定分流。
- 场景：先 PDP/Cart，再 Order/Checkout。

灰度节奏：

```text
0% 空跑比对
1% 白名单/小流量
10% 单品类单地区
50% 扩大品类和地区
100% 全量
```

回滚条件：

- 错误率超过阈值。
- 价格差异率超过阈值。
- 资损风险指标异常。
- P99 延迟明显劣化。
- 客诉或下单转化异常。

---

## 十三、测试策略

### 1. 单元测试

覆盖：

- 无促销。
- FlashSale。
- BundlePrice。
- NewUserPrice。
- 数量为 0。
- `OriginalDiscountPrice` 缺失。
- 超过限购。
- 多商品订单。
- 最低交易金额。
- 费用和优惠叠加。

### 2. 一致性测试

重点验证：

- CalPDPPrice 和 CalOrderPrice 在共同字段上的一致性。
- 新老逻辑结果一致。
- 前端展示用字段和后端金额字段口径一致。
- Payment 使用快照，不重新计算。

### 3. 压测和监控

指标：

- 按场景统计 QPS、P95、P99。
- 缓存命中率。
- 依赖服务耗时。
- 错误率。
- 价格差异率。
- 异常金额分布。

---

## 十四、面试高频追问与回答提纲

### Q1：为什么不直接在订单服务里做计价？

订单服务只适合承接交易结果，不适合承载所有价格规则。价格计算横跨商品、营销、费用、券、积分、支付渠道，如果都放在订单服务里，会导致订单服务过重，并且 PDP、Cart、Checkout 还会继续复制逻辑。

更好的方式是把价格计算独立成计价中心，订单服务在创单时调用计价中心获取订单锁定价和价格明细，并固化快照。

### Q2：PDP 价格和 Order 价格不一致怎么办？

先区分业务允许的不一致和系统错误：

- PDP 是展示价，可以有短暂缓存。
- Order 是交易价，必须实时确认。
- 如果差异来自促销过期、库存变化或供应商实时价变化，需要给用户明确提示。
- 如果差异来自计算口径不同，就是系统问题，需要通过统一计价中心、价格字典和空跑比对解决。

### Q3：Payment 阶段为什么不重新计算？

因为支付阶段重新计算会引入新的不确定性：券状态、积分状态、手续费配置、营销规则都可能变化。Payment 应该验证用户在 Checkout 阶段确认的支付快照，保证实扣金额和收银台金额一致。

### Q4：如何防止价格被前端篡改？

- 前端金额只作为校验输入，不能作为最终可信金额。
- 后端根据商品、营销、费用、优惠重新计算。
- 如果前端提交金额和后端计算金额不一致，拒绝创单或提示刷新。
- Order/Payment 阶段使用快照和状态校验。

### Q5：如何处理优惠分摊？

订单级优惠需要分摊到商品级，主要用于退款和财务对账。

常用方式是按商品金额比例分摊，前 N-1 个商品向下取整，最后一个商品承接尾差，保证分摊总额等于订单优惠总额。

```text
商品 A：500
商品 B：300
商品 C：200
订单级优惠：100

A 分摊：50
B 分摊：30
C 分摊：20
```

### Q6：如何保证金额计算精度？

- 金额统一使用最小货币单位的整数，例如分。
- 避免用 float 做核心金额计算。
- 百分比折扣使用固定精度整数，例如 `discount / 100000`。
- 对尾差有明确归属规则。
- 退款和分摊基于落库明细，不重新推导。

### Q7：如何支持新品类？

流程：

1. 梳理新品类价格来源、费用、促销、库存和履约规则。
2. 判断能否走 BaseCalService。
3. 如果存在特殊逻辑，实现专用 Calculator。
4. 复用统一 PriceEntity、PriceBreakdown、Snapshot。
5. 增加场景测试和空跑比对。
6. 按品类灰度上线。

### Q8：这个设计有什么缺点？

可以坦诚回答：

- 模型变统一后，初期学习成本更高。
- 场景拆分多，接口边界必须定义清楚。
- PriceBreakdown 字段较多，需要控制返回体大小。
- 对旧品类迁移有成本，需要空跑和灰度周期。
- 如果抽象过度，可能影响简单品类开发效率。

改进方向：

- 对简单品类提供更轻量的默认实现。
- 对复杂品类提供配置化能力，但避免过度规则引擎化。
- 对价格快照和明细做冷热分离。
- 引入更完善的自动化差异分析工具。

---

## 十五、三个可展开案例

### 案例一：前后端价格逻辑不一致

**背景**：部分促销逻辑最早在前端实现，后端订单侧也有一套类似逻辑，导致展示价和订单价可能不一致。

**问题**：

- 前端和后端规则各自演进。
- 一些字段前端有，后端没有。
- 复杂促销如 BundlePrice 容易出现边界差异。

**方案**：

- 将前端试算逻辑迁移到后端计价中心。
- 前端只负责展示，不再作为价格权威。
- 通过单元测试覆盖前端历史用例。
- 用空跑比对验证新老逻辑差异。

**效果**：

- 展示和创单口径收敛。
- 价格问题定位路径缩短。
- 后续新增变价因素只需在计价中心扩展。

### 案例二：Home Service 最低交易金额

**背景**：Home Service 存在最低交易金额，不同商品是否计入最低金额由 SKU 属性决定。

**问题**：

- CalPDPPrice 和 CalOrderPrice 公式不一致。
- 多商品混合场景下，`sub_charge` 容易计算错误。
- `sub_charge` 在不同响应结构中的类型和位置不一致。

**方案**：

- 统一公式：`excludedAmount + max(minTransaction, includedAmount)`。
- 明确 `sub_charge = max(0, minTransaction - includedAmount)`。
- 增加全部计入、全部排除、混合、等于最低金额、超过最低金额等测试。

**可讲收获**：

> 这类问题不是算法复杂，而是业务语义容易误解。最重要的是把计入和排除两个集合定义清楚，再让 PDP 和 Order 共享同一套公式。

### 案例三：OriginalDiscountPrice 缺失导致金额异常

**背景**：Event Ticket 某些请求没有传 `OriginalDiscountPrice`。

**问题**：

- Protobuf 默认值为 0。
- 计价逻辑直接使用 `OriginalDiscountPrice * quantity`。
- 导致无促销或促销超限部分金额异常。

**方案**：

- 在所有相关计算函数中增加兜底：`OriginalDiscountPrice == 0` 时使用 `DiscountPrice`。
- 增加字段缺失测试。
- 对价格关键字段增加合法性校验。
- 在 PriceBreakdown 中记录实际使用的 base price，方便排查。

**可讲收获**：

> 金额字段不能把 0 当普通默认值处理。对于价格系统，字段缺失和真实 0 元是两个完全不同的语义。

---

## 十六、白板讲解模板

### 1. 总体架构图

```text
Client
  |
  | Search / Detail / Cart / Order / Checkout / Payment
  v
Pricing Center
  |
  +-- Scene Router
  |     +-- PDP Calculator
  |     +-- Cart Calculator
  |     +-- Order Calculator
  |     +-- Checkout Calculator
  |     +-- Payment Snapshot Validator
  |
  +-- Pricing Engine
  |     +-- Base Price Layer
  |     +-- Promotion Layer
  |     +-- Fee Layer
  |     +-- Discount Layer
  |     +-- Final Aggregation
  |
  +-- Category Calculator
  |     +-- BaseCalService
  |     +-- HomeServiceCalService
  |     +-- EventTicketCalService
  |     +-- EVoucherCalService
  |
  +-- Snapshot / Breakdown / Compare
        +-- Order Snapshot
        +-- Payment Snapshot
        +-- PriceBreakdown
        +-- Dry-run Compare
```

### 2. 价格计算流程

```text
Input
  item / sku / quantity / user / scene / category / promotion / fee / voucher

Normalize
  统一字段，处理默认值和兜底

Select Calculator
  根据 category 和 scene 选择通用或专用计算器

Calculate
  Base Price → Promotion → Fee → Discount → Final

Validate
  非负校验、金额校验、折扣阈值、前后端一致性

Persist
  PriceBreakdown / Order Snapshot / Payment Snapshot

Monitor
  latency / error / diff / abnormal amount
```

---

## 十七、回答时的注意点

### 1. 不要只讲设计模式

不要把项目讲成“我用了责任链、策略模式、DDD”。面试官更关心：

- 为什么需要这些设计。
- 它解决了什么业务问题。
- 有没有线上收益。
- 有没有真实坑和取舍。

推荐表达：

> 我们不是为了用 DDD 而 DDD，而是价格术语和计算口径太混乱，需要一个统一语言和可追溯模型。

### 2. 不要把 PDP、Order、Payment 混在一起

这几个场景是面试官最容易追问的：

- PDP 可以缓存，Order 不应缓存。
- Order 生成订单快照，Checkout 生成支付快照。
- Payment 原则上只验证快照，不重新计算。

### 3. 不要过度承诺“所有资损都解决”

更稳妥的表达是：

> 这个项目上线后未造成资损事故，同时通过快照、校验、空跑比对和监控降低了金额不一致风险。

### 4. 准备好讲一个具体 Bug

建议优先讲：

- `OriginalDiscountPrice` 缺失导致金额异常。
- Home Service 最低交易金额公式不一致。
- BundlePrice 精度和尾差问题。

这些比泛泛讲架构更容易体现深度。

---

## 十八、简历可用表达

如果要压缩成简历 bullet，可以写：

> **计价中心与金额一致性**：针对商品、营销、订单、支付多处价格逻辑分散，以及列表价、详情价、订单价、收银台支付价口径不一致的问题，主导统一计价中心建设；按 PDP / Cart / Order / Checkout / Payment 拆分场景化计算能力，沉淀 Base Price → Promotion → Fee → Discount 四层模型和品类 Calculator 扩展机制，通过 PriceBreakdown、订单/支付快照、空跑比对、灰度迁移和异常拦截保障价格可解释与金额一致，项目上线后未造成资损事故。

如果希望更偏架构：

> **场景化计价引擎**：将分散在商品、营销、订单和支付系统中的计价逻辑收敛为统一服务，按展示、试算、创单、收银台、支付校验拆分计算深度和缓存策略；通过分层计价模型、品类策略、价格明细和快照机制支撑 10+ 数字商品品类，新增变价因素从 2-3 周缩短到 3-5 天，关键接口 P99 优化至 <200ms。

---

## 十九、最终复习清单

面试前重点准备这 10 个问题：

1. 为什么要做统一计价中心？
2. PDP、Order、Checkout、Payment 的计价边界是什么？
3. 四层计价模型怎么设计？为什么是这个顺序？
4. 订单快照和支付快照分别存什么？
5. 如何保证展示价、订单价、支付价一致？
6. 多品类差异怎么扩展？
7. BundlePrice、NewUserPrice、最低交易金额如何计算？
8. 如何处理金额精度、尾差和优惠分摊？
9. 空跑比对和灰度迁移怎么做？
10. 项目中遇到过哪些真实 Bug？怎么修复和预防？

---

## 二十、推荐面试讲法

推荐顺序：

1. 先讲背景：多品类、多场景、多变价因素、金额一致性风险。
2. 再讲方案：场景化 API、四层模型、品类 Calculator、快照。
3. 再讲工程保障：空跑比对、灰度、监控、测试。
4. 最后讲具体案例：Home Service、Event Ticket 或 BundlePrice。

一段完整回答可以这样收尾：

> 这个项目对我最大的价值是，我不只是做了一个价格计算函数，而是把价格这个高风险领域变成了一个可解释、可扩展、可灰度、可追溯的平台能力。它的难点不在公式本身，而在多品类差异、场景边界、金额一致性和安全迁移。

---

## 二十一、30 道面试题索引

这一节保留原 `pricing-engine-interview-questions.md` 的 30 题结构，方便面试前按题目快速复习。详细回答可以回到前文对应章节展开。

### 项目背景与架构设计

| 题目 | 回答重点 |
| --- | --- |
| 1. 为什么要构建统一计价中心？ | 价格逻辑分散、术语混乱、多品类重复实现、金额一致性风险；统一计价中心把价格变成平台能力。 |
| 2. 计价引擎采用什么分层架构？ | Base Price → Promotion → Fee → Discount → Final；每层职责清晰，便于扩展、测试和审计。 |
| 3. 如何支持 EVoucher、GiftCard、MovieTicket 等多品类差异？ | BaseCalService 承载通用逻辑，复杂品类实现 Calculator；避免通用逻辑被 category if-else 污染。 |
| 4. 如何保证展示价、订单价、支付价一致？ | 场景化计算、订单快照、支付快照、金额校验、空跑比对、灰度迁移。 |
| 5. 价格快照如何防资损？ | Order Snapshot 锁定订单上下文，Payment Snapshot 锁定收银台实付金额，Payment 阶段只验证不重算。 |
| 6. 如何在计价领域应用 DDD？ | 统一语言、Price 聚合根、PriceComponent、PriceBreakdown、领域服务、ACL 隔离外部依赖。 |
| 7. 如何设计 PDP、Cart、Order、Checkout API？ | 场景驱动，不做万能 API；展示场景重性能，交易场景重准确性，支付场景重快照校验。 |
| 8. 如何处理自营品类和供应商品类差异？ | 自营品类以本地价格/库存为主，供应商品类需要实时查价查库存、Booking Token、超时释放和补偿。 |

### 复杂业务场景处理

| 题目 | 回答重点 |
| --- | --- |
| 9. EVoucher BundlePrice 如何计算？ | 满 N 件折扣、最大轮数、超限原价、整数金额、尾差处理。 |
| 10. FlashSale 和 NewUserPrice 如何处理优先级和互斥？ | 促销优先级、互斥规则、限购、显式 selected promotion，避免默认回退导致金额不一致。 |
| 11. 订单级优惠如何分摊？ | 按商品金额比例分摊，前 N-1 项取整，最后一项承接尾差；退款基于分摊后实付金额。 |
| 12. 如何保证价格计算确定性？ | 统一输入、整数计算、快照、Breakdown、幂等校验、空跑比对、监控告警。 |
| 13. Hotel/Flight 如何处理实时价格和 BookingToken？ | 列表缓存、详情刷新、Booking 实时确认、Order 快照、Payment 快照、超时释放。 |
| 14. 如何处理前后端计价逻辑统一？ | 前端不做价格权威，后端统一计算；历史前端用例迁移为后端测试；新老逻辑空跑比对。 |
| 15. 最复杂的计价场景是什么？ | 可以讲 Bundle + 限购 + 超限原价，或供应商 Booking + 支付快照，或合并支付 + 部分退款。 |
| 16. 如何设计 PriceBreakdown？ | 记录基础价、促销、费用、优惠、最终价、规则 ID、请求 ID、场景，支撑客服、审计、退款和排查。 |

### 性能优化与高并发

| 题目 | 回答重点 |
| --- | --- |
| 17. 大促期间如何保证高性能？ | 场景化缓存、批量查询、并发拉取依赖、热点预热、限流、熔断和降级。 |
| 18. 如何平衡缓存和准确性？ | Search/PDP 可缓存，Order/Checkout 不缓存，Payment 用快照；“列表快、详情准、创单安全”。 |
| 19. 列表页 100 个商品批量计价如何优化？ | 批量 RPC、Redis Pipeline、DB IN 查询、worker pool、控制并发和请求上限。 |
| 20. Redis 挂了怎么办？ | 本地缓存兜底、降级 DB、限流保护 DB、熔断、过期缓存、监控告警。 |
| 21. 如何设计监控告警？ | 按场景/品类/地区看 QPS、成功率、P99、差异率、异常金额、依赖耗时、缓存命中率。 |
| 22. 如何做限流和熔断？ | 应用层限流、依赖服务熔断、非核心能力降级、供应商查询超时兜底。 |

### 项目管理与团队协作

| 题目 | 回答重点 |
| --- | --- |
| 23. 项目从立项到上线经历哪些阶段？ | 调研、建模、核心开发、迁移、空跑、灰度、全量、复盘。 |
| 24. 如何做灰度发布和空跑比对？ | 老逻辑返回，新逻辑比对；按品类、地区、用户比例逐步切流，异常回滚。 |
| 25. 你在项目中承担什么角色？ | 架构设计、核心模型开发、品类迁移、性能优化、灰度与问题排查。 |
| 26. 最大挑战是什么？ | 前后端/新老逻辑一致性、多品类差异抽象、安全迁移、资损防控。 |

### 技术深度与扩展

| 题目 | 回答重点 |
| --- | --- |
| 27. 如何支持 C2C 竞价模式？ | 新增 BiddingCalculator，维护起拍价、当前价、Buy Now 价、竞价锁、WebSocket 推送和风控阈值。 |
| 28. 如何支持跨境汇率实时转换？ | 汇率服务、汇率缓存、订单汇率快照、汇率波动告警、结算汇率锁定。 |
| 29. 如何支持算法动态定价？ | Feature Store、模型推理、价格上下限、A/B 测试、异常回滚和人工审核。 |
| 30. 如果重构，会做哪些改进？ | 控制抽象层级、加强性能基准、快照冷热分离、自动化差异分析、Fuzzing 和配置治理。 |

---

## 二十二、项目管理、角色与指标

### 项目阶段

| 阶段 | 重点工作 | 风险点 |
| --- | --- | --- |
| 调研阶段 | 梳理多品类、多服务、多场景价格逻辑，识别重复逻辑和口径差异。 | 隐藏规则多，历史逻辑分散。 |
| 建模阶段 | 统一价格术语，设计四层计价模型、场景 API、PriceBreakdown 和快照。 | 抽象过度或抽象不足。 |
| 核心开发 | 实现 BaseCalService、品类 Calculator、快照、空跑比对和监控。 | 影响核心交易链路，必须兼容旧逻辑。 |
| 迁移阶段 | 按品类迁移前端/订单/支付侧计价逻辑，补单元测试和差异比对。 | 新老逻辑细节不一致。 |
| 空跑阶段 | 老逻辑返回，新逻辑异步比对，收敛差异。 | 差异归因和误报治理。 |
| 灰度阶段 | 按品类、地区、用户比例切流，观察金额差异、P99 和错误率。 | 资损风险和体验风险。 |
| 全量复盘 | 下线老逻辑，沉淀文档、监控、SOP 和新人指南。 | 遗留逻辑清理不彻底。 |

### 我的职责

- 主导或参与计价中心的整体架构设计，拆分场景 API 和分层计价模型。
- 参与核心领域模型建设，包括 PriceEntity、PriceBreakdown、PriceSnapshot 和品类 Calculator。
- 推动价格术语统一，澄清市场价、折扣价、促销价、计价金额、支付金额等关键概念。
- 处理复杂品类计价，包括 BundlePrice、NewUserPrice、最低交易金额、供应商实时价格和支付手续费。
- 设计空跑比对、灰度迁移和异常告警，降低核心交易链路迁移风险。
- 参与性能优化，通过缓存、批量查询、并发拉取和场景降级优化高频接口 P99。

### 技术栈

| 类型 | 技术 |
| --- | --- |
| 语言 | Go |
| 服务框架 | 自研 RPC / HTTP 服务框架 |
| 存储 | MySQL、Redis |
| 消息与异步 | Kafka、Worker、定时任务 |
| 可观测性 | Prometheus、Grafana、日志、链路追踪 |
| 工程方法 | DDD、策略模式、责任链、灰度、空跑比对、快照、对账 |

### 指标速查

| 指标 | 改进前 | 改进后 |
| --- | --- | --- |
| 价格逻辑维护 | 分散在商品、营销、订单、支付等多处 | 收敛到计价中心和品类 Calculator |
| 新增变价因素 | 2-3 周 | 3-5 天 |
| 高频接口 P99 | 500ms 左右 | <200ms |
| PDP 缓存命中率 | 约 70% | 90%+ |
| 品类接入 | 多品类重复实现 | 复用统一模型和策略扩展 |
| 项目资损 | 高风险区域 | 项目上线后未造成资损事故 |

---

## 二十三、命名规范与模型检查清单

### 推荐字段命名

| 场景 | 推荐命名 | 避免命名 | 说明 |
| --- | --- | --- | --- |
| 市场价 | `market_price` | `original_price`、`list_price` | 用于划线价和原始标价。 |
| 折扣价 | `discount_price` | `sale_price` | 无促销时的日常销售价。 |
| 原始折扣价 | `original_discount_price` | `base_price` | 参与计算的兜底字段，缺失时回退到 `discount_price`。 |
| 促销价 | `promotion_price` | `activity_price` | 秒杀、新用户价、Bundle 后价格。 |
| 计价金额 | `pricing_amount` | `final_price` | 单个商品当前阶段的计价结果。 |
| 售卖价信息 | `selling_price_info` | `sell_price` | Order 场景用于固化商品售卖价格结构。 |
| 最低交易金额 | `min_transaction` | `min_order` | Home Service 等特殊品类使用。 |
| 附加费用 | `additional_charge`、`sub_charge` | `extra_fee` | `sub_charge` 专指补足最低交易金额。 |
| 支付金额 | `payment_amount`、`final_amount` | `total` | 支付阶段必须语义明确。 |

### 注释规范

```go
// 市场原价（单价）
marketPrice := int64(1000)

// 折扣价（单价）
discountPrice := int64(980)

// 促销前总价 = 折扣价 * 数量
discountTotalPrice := discountPrice * quantity

// 促销后总价 = 促销价 * 数量
promotionTotalPrice := promotionPrice * quantity

// 促销优惠金额 = 促销前总价 - 促销后总价
promotionDiscountAmount := discountTotalPrice - promotionTotalPrice
```

避免：

```go
price := int64(1000)      // 什么价格？
finalPrice := int64(980)  // 是商品最终价、订单价，还是支付价？
total := price * qty      // 是促销前还是促销后？
```

### Protobuf 字段建议

```protobuf
message PriceEntity {
  optional int64 item_id = 1;
  optional int64 sku_id = 2;
  optional int64 category_id = 3;
  optional int64 quantity = 4;

  optional int64 market_price = 10;
  optional int64 discount_price = 11;
  optional int64 original_discount_price = 12;

  optional int64 pricing_amount = 20;
  optional SellingPriceInfo selling_price_info = 21;
}

message CalPDPPriceRequest {
  repeated PriceEntity price_entities = 1;
  optional int64 min_transaction = 2;
}

message CalOrderPriceResponse {
  optional int64 pricing_total_amount = 1;
  repeated PriceEntity price_entities = 2;
  map<string, string> extra_info = 3;
}
```

### 设计检查清单

- 是否区分了单价、商品总价、订单总价？
- 是否区分了 PDP、Order、Checkout、Payment 的计算职责？
- 是否避免 Payment 阶段重新计算价格？
- 价格关键字段为 0 时是否有明确语义和兜底逻辑？
- 订单级优惠是否完成商品级分摊并落库？
- PriceBreakdown 是否足够解释客服、退款和对账问题？
- 新品类是否只实现 Calculator，而不是修改通用逻辑？
- 新老逻辑迁移是否有空跑比对和灰度开关？
- 缓存是否只用于展示/试算场景，而不污染交易价？
- 金额计算是否统一使用整数和明确取整规则？

---

## 二十四、测试覆盖与质量保障

### 单元测试覆盖建议

| 品类/模块 | 必测场景 |
| --- | --- |
| BaseCalService | 无促销、有促销、数量为 0、折扣价缺失、SellingPriceInfo 填充。 |
| Home Service | 全部计入最低交易金额、全部排除、混合、等于最低金额、超过最低金额、SkuInfo 缺失。 |
| Event Ticket | 无促销、FlashSale、BundlePrice、Scheduling、NewUserPrice、限购、OriginalDiscountPrice 缺失。 |
| EVoucher | Bundle、券码库存、购买数量边界、最大轮数、超限原价。 |
| Checkout | 券、积分、手续费、支付渠道优惠、最终金额非负校验。 |
| Snapshot | 过期、重复使用、订单不匹配、金额不匹配、状态不合法。 |

### 一致性测试

- CalPDPPrice 和 CalOrderPrice 在共同金额字段上是否一致。
- 前端历史价格用例迁移到后端后是否一致。
- 新老计价逻辑在空跑期的最终金额和明细是否一致。
- PriceBreakdown 中各层金额加总是否等于最终金额。
- 正向订单分摊结果和逆向退款金额是否一致。

### 性能测试

- 单商品 PDP 计价 P99。
- 列表 100 个商品批量计价 P99。
- Checkout 并发查询依赖服务的耗时。
- Redis 命中、Redis 失效、DB 兜底三种路径耗时。
- 大促热点商品缓存穿透和 singleflight 效果。

### Fuzzing / 边界测试方向

- `quantity = 0`、`quantity = 1`、极大数量。
- `market_price = 0`、`discount_price = 0`、`original_discount_price = 0`。
- 优惠金额大于商品金额。
- 多个优惠叠加后最终金额小于 0。
- Bundle 的 `min_quantity = 0`、`max_round = 0`。
- 分摊尾差无法整除。

---

## 二十五、扩展设计题

### 1. 支持 C2C 竞价模式

竞价模式和普通电商定价不同，价格不是平台直接计算出的固定值，而是由竞价过程决定。

核心模型：

- 起拍价 `start_price`
- 当前最高价 `current_price`
- 加价幅度 `bid_step`
- 一口价 `buy_now_price`
- 截止时间 `end_time`
- 自动出价策略

设计思路：

- 新增 `BiddingCalculator`，作为特殊品类 Calculator。
- 当前价存 Redis，并通过乐观锁或 Lua 保证并发出价原子性。
- 出价成功后通过 WebSocket 或消息推送更新前端。
- Order 阶段固化最终成交价和出价记录快照。
- 增加风控限制，避免异常低价、高频刷价和恶意抬价。

面试回答重点：

> 竞价模式可以复用计价中心的 PriceBreakdown 和快照能力，但价格来源从商品基础价变成竞价状态，因此需要把竞价状态作为 Base Price Layer 的输入，并对并发出价做强一致控制。

### 2. 支持跨境汇率实时转换

跨境场景需要处理展示币种、支付币种、结算币种和汇率波动。

设计思路：

- 独立汇率服务维护实时汇率和缓存。
- PDP 可使用短 TTL 汇率展示预估价格。
- Order/Checkout 阶段锁定汇率快照。
- Payment 和 Refund 使用订单汇率快照，不使用当前汇率重算。
- 汇率大幅波动时触发告警或暂停高风险品类交易。

需要落库：

- 原币种金额。
- 目标币种金额。
- 汇率值。
- 汇率来源。
- 汇率生效时间。
- 汇率快照 ID。

### 3. 支持算法动态定价

算法定价的关键不是直接让模型决定最终价格，而是让模型输出建议价，并由规则和风控兜底。

输入特征：

- 历史成交价。
- 当前库存。
- 竞争价格。
- 用户转化率。
- 时间、节假日、地区。
- 供应商价格波动。

设计思路：

- 模型服务输出建议价。
- 计价中心在 Base Price 或 Promotion Layer 接入建议价。
- 设置价格上下限和单次变动幅度。
- 通过 A/B 测试观察 GMV、转化率、利润率和客诉。
- 异常价格自动回滚到规则价。

面试回答重点：

> 算法定价不能绕过计价中心直接改最终价。它应该作为一个可解释的价格输入，并被价格上下限、快照、灰度和监控约束。

### 4. 未来重构方向

- **控制抽象层级**：避免调用链过深，保持模型清晰。
- **增强差异分析**：空跑差异自动归因到基础价、促销、费用或优惠。
- **快照冷热分离**：热订单快照保留在线库，历史快照归档。
- **配置治理**：促销优先级、互斥规则、费用规则配置化，但避免无约束脚本化。
- **自动化测试**：增加 Fuzzing、属性测试、性能基准和线上回放测试。
- **可解释性增强**：PriceBreakdown 面向客服、财务、研发提供不同视图。

---

## 二十六、最终保留文件说明

本文件已经覆盖并重组原三份材料的主要内容：

- 原 `pricing-engine-interview-questions.md` 的 30 道题已整理到“30 道面试题索引”和对应正文。
- 原 `project-pricing-engine-interview.md` 的项目背景、痛点、方案、挑战、职责、收益和常见问答已合并到正文与附录。
- 原 `project-pricing-engine-model-optimized.md` 的术语、模型、品类计价、Bug、公式、命名规范、测试覆盖和最佳实践已合并到正文与检查清单。

后续只维护这一份材料即可。
