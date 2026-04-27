
**导航**：[书籍主页](../README.md) | [完整目录](../SUMMARY.md) | [上一章：第16章](../part2/transaction/chapter15.md)

---

# 第17章 B2B2C平台完整架构

> **综合案例**：一个中大型B2B2C电商平台的完整架构设计，从品类分析到技术选型，从系统设计到团队协作，覆盖200+人团队、日订单200万级的实战经验与架构决策。

---

## 16.1 项目背景与业务约束（Business Context）

本章讨论的是一个中大型 B2B2C 聚合电商平台。它不是传统实物电商，也不是单一供应商商城，而是连接多类外部供应商与自营虚拟商品的平台型系统。平台侧负责商品组织、搜索导购、价格试算、营销、下单、支付和履约编排；供应商侧负责实际资源确认与数字履约。

这个背景很重要。因为系统的核心复杂度不来自物流仓配，而来自**多品类、多供应商、强实时交易和不一致外部接口**的叠加：机票和酒店要求零超卖，充值和礼品卡允许失败后补偿，电子券又依赖本地券码池发放。不同品类背后的库存、价格、履约模型完全不同，直接决定后续的领域划分、服务边界和一致性策略。

### 16.1.1 业务定位

平台采用“聚合供应商 + 自营虚拟商品”的 B2B2C 模式，连接航司/GDS、酒店 OTA/PMS、运营商、院线、券码供应商等外部系统，同时保留部分自营业务能力。

**业务范围**：

| 业务类型 | 典型品类 | 履约方式 | 关键特征 |
|---------|---------|---------|---------|
| **供应商聚合** | 机票、酒店、充值、电影票 | 调用供应商 API 完成出票、确认、充值、锁座 | 接口差异大，实时性和可用性依赖外部系统 |
| **平台自营** | 优惠券、线下券、礼品卡 | 平台本地发券码或调用内部发码系统 | 可控性更强，但需要券码池、核销和过期管理 |
| **数字履约** | 所有品类 | API 调用、异步确认、电子凭证发放 | 无物流链路，但交易状态和补偿链路更复杂 |

**业务全景图**：

![B2B2C 数字商品平台业务全景](../../images/b2b2c-business-overview.png)

这张图展示了平台业务链路的四个关键视角：供应商供给侧负责资源供给与数字履约；平台核心能力负责商品、库存、价格、营销、订单、支付、履约和售后编排；本地商家与平台运营负责商品录入、审核、维护、促销和上下架；C 端用户则围绕搜索导购、结算下单、支付、履约结果和退款售后形成完整交易闭环。

本平台的一个关键前提是**无物流场景**。所有商品都是虚拟数字商品，履约不经过仓库、分拣、配送，而是通过 API 调用、电子票、确认单、充值结果或券码完成。因此，本章不会讨论仓配、物流轨迹、签收等实物电商问题，而是聚焦数字商品平台中更核心的四类问题：

1. **供应商差异**：50+ 外部供应商接口形态不一致，可能同时存在实时查询、定时同步和事件推送。
2. **库存差异**：机票/酒店依赖供应商实时库存，电子券依赖本地券码池，充值类商品近似无限库存。
3. **价格差异**：机票动态定价、酒店日历价、充值固定面额、优惠券固定折扣价并存。
4. **履约差异**：有的品类同步返回结果，有的品类需要异步确认，有的品类需要本地分配唯一券码。

### 16.1.2 业务与技术目标

平台的核心目标可以概括为四句话：

1. **交易链路要稳**：订单创建、支付、履约不能丢数据，核心链路故障要快速恢复。
2. **导购链路要快**：搜索、详情、结算要在高并发下保持低延迟，并允许适度降级。
3. **供给运营要可控**：商品上架、运营编辑、供应商同步、促销配置和上下架要有明确流程、审核机制和可追溯记录。
4. **品类接入要快**：新增品类和供应商不能反复改造主流程。
5. **团队协作要顺**：服务边界、API 契约和事件契约必须足够清晰，支撑多人多团队并行开发。

**性能目标**：

| 指标 | 正常值 | 大促峰值 | 设计含义 |
|------|--------|---------|---------|
| 日订单量 | 200 万 | 1000 万 | 交易链路需要支持 5 倍峰值弹性 |
| 搜索 QPS | 3000 | 15000 | 搜索与聚合层需要缓存、批量查询和降级能力 |
| 详情页 QPS | 5000 | 25000 | 商品、库存、计价、营销服务需要并发编排 |
| 下单 QPS | 1000 | 5000 | 库存预占、价格校验、订单写入必须控制事务边界 |
| P99 延迟 | 200ms | 500ms | 大促期间允许部分非核心能力降级 |

**可用性目标**：

| 目标 | 要求 | 说明 |
|------|------|------|
| 核心链路 SLA | 99.95% | 覆盖订单创建、支付、履约等交易动作 |
| 搜索/详情 SLA | 99.9% | 可通过缓存、兜底价格、隐藏营销信息等方式降级 |
| RTO | < 5 分钟 | 故障后需要在 5 分钟内恢复核心能力 |
| RPO | 0 | 核心交易数据不允许丢失 |

**扩展性目标**：

| 扩展场景 | 目标 | 关键依赖 |
|---------|------|---------|
| 新品类接入 | < 2 周 | 品类策略、库存策略、履约策略可插拔 |
| 新供应商接入 | < 1 周 | 供应商适配器、防腐层、统一错误模型 |
| 新营销玩法 | < 3 天 | 规则引擎、计价输入标准化、活动配置平台 |

### 16.1.3 本章的核心架构命题

在上述业务背景下，架构设计的难点不是“要不要拆微服务”，而是**如何在品类差异、供应商不确定性和交易一致性之间找到可演进的边界**。

本章后续会围绕四个问题展开：

1. **如何理解品类差异**：机票、酒店、充值、电子券为什么不能用同一套库存和履约模型。
2. **如何划分系统边界**：订单、商品、库存、计价、营销、支付、供应商网关各自拥有怎样的数据和职责。
3. **如何组织交易链路**：搜索、详情、结算、下单、支付如何在性能与准确性之间取舍。
4. **如何沉淀架构决策**：通过 ADR 记录关键取舍，避免团队在同一问题上反复争论。

---

## 16.2 品类业务模型分析（Business Architecture）

不同品类的业务模型存在显著差异，直接影响架构设计决策。理解这些差异是系统设计的基础。

### 16.2.1 机票业务模型

**业务特点**：
```
• 库存模型：实时库存（供应商侧），强依赖供应商实时查询
• 价格模型：动态定价，实时波动（可能秒级变化）
• SKU复杂度：极高（航司+航班号+舱位+日期+...组合）
• 库存单位：座位数量（不可超卖）
• 扣减时机：创单前向供应商实时确认并占座 → 创单待支付 → 支付确认 → 出票
• 履约流程：查询报价/库存 → 占座/预订 → 创建订单 → 支付 → 出票（调用GDS/供应商API）→ 发送电子票
```

**架构影响**：
- ✓ 必须支持实时库存查询（高频调用供应商API）
- ✓ 价格快照必须精确到秒级，防止价格变动纠纷
- ✓ 超卖零容忍 → 创建订单前必须完成供应商侧占座/预订
- ✓ 供应商故障需快速切换到备用供应商
- ✓ 订单状态复杂（待出票、出票中、出票失败、已出票）

**技术要点**：
```go
// 机票库存查询策略
type FlightStockStrategy struct {
    supplierClient rpc.SupplierClient
    redis          redis.Client
    config         *FlightConfig
}

func (s *FlightStockStrategy) CheckStock(ctx context.Context, req *StockRequest) (*StockResponse, error) {
    // Step 1: 尝试从Redis获取缓存（TTL=5分钟）
    cacheKey := fmt.Sprintf("flight:stock:%s:%s", req.FlightNo, req.Date)
    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        return parseStockFromCache(cached), nil
    }
    
    // Step 2: 缓存未命中，调用供应商实时查询
    ctx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)  // 800ms超时
    defer cancel()
    
    stock, err := s.supplierClient.QueryStock(ctx, req)
    if err != nil {
        // 供应商故障，切换备用供应商
        return s.fallbackToSecondarySupplier(ctx, req)
    }
    
    // Step 3: 缓存结果（短TTL，机票价格变化快）
    s.redis.Set(ctx, cacheKey, marshal(stock), 5*time.Minute)
    
    return stock, nil
}
```

**监控指标**：
- 供应商调用超时率：< 1%
- 缓存命中率：> 70%
- 出票成功率：> 99.5%
- 出票平均时长：< 30秒

### 16.2.2 酒店业务模型

**业务特点**：
```
• 库存模型：房间数量（按日期维度管理）
• 价格模型：日历房价（每个日期不同价格）
• SKU复杂度：高（酒店ID+房型+日期范围+早餐+...）
• 库存单位：房间数/间夜数
• 扣减时机：下单预占 → 支付确认 → 供应商确认
• 履约流程：下单 → 支付 → 提交供应商 → 确认单 → 入住凭证
```

**架构影响**：
- ✓ 支持日期范围查询（check-in到check-out）
- ✓ 日历价格存储（每个日期一条记录）
- ✓ 库存按日期维度管理（某天无房不影响其他日期）
- ✓ 支持"担保"模式（先占房，入住时结算）
- ✓ 需处理"确认单延迟"（供应商异步确认）

**数据模型**：

```go
// 酒店日历价格表（宽表存储）
type HotelCalendarPrice struct {
    HotelID      int64     `gorm:"primaryKey"`
    RoomTypeID   int64     `gorm:"primaryKey"`
    Date         time.Time `gorm:"primaryKey;index"`  // 日期维度
    BasePrice    int64     // 基础价格（分）
    WeekendPrice int64     // 周末价格
    Stock        int       // 当日库存
    Status       string    // 可售状态（AVAILABLE/SOLD_OUT/CLOSED）
}

// 查询日期范围内的价格与库存
func (r *HotelRepo) GetCalendarPrice(hotelID, roomTypeID int64, checkIn, checkOut time.Time) ([]*HotelCalendarPrice, error) {
    var prices []*HotelCalendarPrice
    err := r.db.Where("hotel_id = ? AND room_type_id = ? AND date >= ? AND date < ?",
        hotelID, roomTypeID, checkIn, checkOut).
        Order("date ASC").
        Find(&prices).Error
    return prices, err
}
```

**缓存策略**：
- 热门酒店：30分钟缓存
- 长尾酒店：1小时缓存
- 价格变更：主动失效缓存

### 16.2.3 充值业务模型

**业务特点**：
```
• 库存模型：无限库存（供应商侧无限制）
• 价格模型：固定面额（10元、50元、100元）
• SKU复杂度：低（运营商+面额）
• 库存单位：无限
• 扣减时机：支付后
• 履约流程：下单 → 支付 → 调用供应商API → 充值成功/失败
```

**架构影响**：
- ✓ 无需库存管理（库存类型=无限）
- ✓ 价格简单（基础价+平台服务费）
- ✓ 超卖可接受（事后补偿）
- ✓ 供应商调用简单（同步API，3秒内返回）
- ✓ 失败重试友好（幂等性强）

**技术要点**：
```go
// 充值库存策略（无限库存）
type RechargeStockStrategy struct{}

func (s *RechargeStockStrategy) CheckStock(ctx context.Context, req *StockRequest) (*StockResponse, error) {
    // 充值类商品无需检查库存，直接返回"可售"
    return &StockResponse{
        Available: true,
        Quantity:  999999,  // 虚拟无限库存
        Message:   "充值类商品，库存充足",
    }, nil
}

func (s *RechargeStockStrategy) Reserve(ctx context.Context, req *ReserveRequest) (*ReserveResponse, error) {
    // 充值类商品无需预占，直接返回成功
    return &ReserveResponse{
        ReserveID: "",  // 无预占ID
        Success:   true,
    }, nil
}
```

### 16.2.4 电子券业务模型

**业务特点**：
```
• 库存模型：固定库存（券码池）
• 价格模型：固定折扣价
• SKU复杂度：中（商户+门店+商品+...）
• 库存单位：券码（一券一码）
• 扣减时机：支付后
• 履约流程：下单 → 支付 → 发券码 → 到店核销
```

**架构影响**：
- ✓ 券码池管理（预生成10万个券码）
- ✓ 券码发放（支付后随机分配）
- ✓ 核销系统（商户扫码核销）
- ✓ 过期管理（券有效期7天-180天）
- ✓ 退款逻辑（未核销可退，已核销不可退）

**技术要点**：
```go
// 券码池管理（Redis实现）
type VoucherCodePool struct {
    redis redis.Client
}

func (p *VoucherCodePool) AssignCode(ctx context.Context, skuID int64, orderID int64) (string, error) {
    // Step 1: 从Redis Set中原子弹出一个未使用的券码
    poolKey := fmt.Sprintf("voucher:pool:%d", skuID)
    code, err := p.redis.SPop(ctx, poolKey).Result()
    if err == redis.Nil {
        return "", errors.New("券码已售罄")
    }
    
    // Step 2: 记录券码分配关系（券码 → 订单号）
    assignKey := fmt.Sprintf("voucher:assign:%s", code)
    p.redis.Set(ctx, assignKey, orderID, 0)  // 永久存储
    
    // Step 3: 设置券码有效期（ZSet按过期时间排序）
    expiresAt := time.Now().Add(90 * 24 * time.Hour)  // 90天有效期
    expiryKey := fmt.Sprintf("voucher:expiry:%d", skuID)
    p.redis.ZAdd(ctx, expiryKey, redis.Z{
        Score:  float64(expiresAt.Unix()),
        Member: code,
    })
    
    return code, nil
}
```

### 16.2.5 差异化设计策略

通过上述品类分析，我们提炼出三个核心设计维度：

**维度1：库存管理类型**

| 类型 | 典型品类 | 库存来源 | 预占策略 |
|------|---------|---------|---------|
| **实时库存** | 机票、酒店、电影票 | 供应商实时查询 | 创单前确认资源，订单超时释放 |
| **池化库存** | 优惠券、礼品卡 | 平台自有（券码池） | 支付后扣减 |
| **无限库存** | 充值、SaaS服务 | 无库存概念 | 无需预占 |

**维度2：价格模型**

| 类型 | 典型品类 | 缓存策略 | 快照策略 |
|------|---------|---------|---------|
| **动态定价** | 机票 | 5分钟TTL | 秒级快照 |
| **日历定价** | 酒店 | 30分钟TTL | 日期维度快照 |
| **固定定价** | 充值、礼品卡 | 1小时TTL | 简单快照 |

**维度3：履约模式**

| 类型 | 典型品类 | 调用方式 | 失败处理 |
|------|---------|---------|---------|
| **同步履约** | 充值 | 同步API（3秒超时） | 立即重试3次 |
| **异步履约** | 机票、酒店 | 异步轮询（30秒/次） | 补偿任务 |
| **券码发放** | 优惠券 | 本地分配（无外部调用） | 券码池补充 |

**统一抽象**：

```go
// 品类策略接口（策略模式）
type CategoryStrategy interface {
    // 库存检查
    CheckStock(ctx context.Context, req *StockRequest) (*StockResponse, error)
    // 库存预占
    ReserveStock(ctx context.Context, req *ReserveRequest) (*ReserveResponse, error)
    // 价格计算
    CalculatePrice(ctx context.Context, req *PriceRequest) (*PriceResponse, error)
    // 订单履约
    Fulfill(ctx context.Context, order *Order) (*FulfillResult, error)
}

// 策略工厂（根据品类选择策略）
type CategoryStrategyFactory struct {
    strategies map[CategoryType]CategoryStrategy
}

func (f *CategoryStrategyFactory) GetStrategy(categoryType CategoryType) CategoryStrategy {
    return f.strategies[categoryType]
}
```

**设计原则**：
1. **策略模式**：每个品类一个策略实现，避免 if-else 地狱
2. **适配器模式**：统一供应商接口差异，降低耦合
3. **模板方法**：下单流程统一，具体步骤由策略实现
4. **可扩展性**：新增品类只需新增策略，不影响主流程

---

## 16.3 DDD战略设计与系统边界（Application Architecture - 设计过程）

基于16.2的品类业务分析，本节展示如何运用DDD战略设计方法，从业务领域识别限界上下文、划分系统边界、设计服务间集成方式，最终形成16.4的整体架构全貌。

### 16.3.1 限界上下文识别

**限界上下文是DDD战略设计的核心概念**，它定义了一个模型的适用边界。本系统通过事件风暴识别出12个核心限界上下文。

**识别过程**（事件风暴Workshop）：

```
第1步：领域事件识别（橙色便签）
• OrderCreated（订单创建）
• ProductOnShelf（商品上架）
• StockReserved（库存预占）
• PaymentPaid（支付成功）
• PromotionApplied（促销应用）
...

第2步：聚合命令（蓝色便签）
• CreateOrder（创建订单）
• ReserveStock（预占库存）
• CalculatePrice（计算价格）
• ApplyPromotion（应用促销）
...

第3步：聚合实体（黄色便签）
• Order（订单）
• Product（商品）
• Stock（库存）
• Payment（支付）
• Promotion（促销）
...

第4步：限界上下文识别（用绳子圈起相关的实体/命令/事件）
• 订单上下文：Order + CreateOrder + OrderCreated
• 商品上下文：Product + OnShelfProduct + ProductOnShelf
• 库存上下文：Stock + ReserveStock + StockReserved
...
```

**识别出的12个限界上下文**：

| 限界上下文 | 核心聚合根 | 核心职责 | 数据所有权 |
|---------|---------|---------|-----------|
| **订单上下文** | Order | 订单创建、状态机、履约 | orders、order_items |
| **商品上下文** | Product | 商品信息、类目、属性 | products、categories |
| **库存上下文** | Stock | 库存管理、预占、扣减 | stocks、stock_logs |
| **计价上下文** | Price | 价格计算、试算、快照 | price_snapshots |
| **营销上下文** | Promotion | 营销规则、优惠券、活动 | promotions、coupons |
| **支付上下文** | Payment | 支付、退款、对账 | payments、refunds |
| **搜索上下文** | ProductIndex | 商品搜索、筛选、排序 | ES索引 |
| **用户上下文** | User | 用户信息、登录、权限 | users、roles |
| **供应商上下文** | Supplier | 供应商对接、适配、熔断 | suppliers、supplier_products |
| **购物车上下文** | Cart | 购物车管理、合并 | carts |
| **评价上下文** | Review | 用户评价、晒单 | reviews |
| **消息上下文** | Notification | 消息通知、推送 | notifications |

**为什么这样划分？**

1. **订单与商品分离**：
   - 订单关注"交易流程"（下单、支付、履约）
   - 商品关注"商品信息"（SPU/SKU、类目、属性）
   - 分离原因：变化速度不同（订单频繁变更，商品相对稳定）

2. **库存独立**：
   - 库存是"资源"，订单/商品都依赖它
   - 库存有独立的生命周期（预占 → 扣减 → 释放）
   - 独立原因：单一职责，避免库存逻辑分散

3. **计价独立**：
   - 价格计算涉及多个维度（基础价、营销、优惠券、Coin）
   - 多个场景需要试算（详情页、购物车、结算页）
   - 独立原因：统一计价逻辑，避免不一致

4. **营销独立**：
   - 营销规则复杂（满减、折扣、买赠、限时秒杀）
   - 营销活动变化频繁
   - 独立原因：灵活支持新玩法，不影响主流程

**上下文大小原则**：

```
过小：每个实体一个上下文 ❌
• 导致上下文过多，通信成本高
• 事务边界不清晰

合适：一个聚合根（或紧密相关的聚合根）一个上下文 ✅
• 订单上下文：Order + OrderItem
• 商品上下文：Product + Category

过大：多个不相关的聚合根在一个上下文 ❌
• 导致上下文职责不清晰
• 团队协作困难
```

### 16.3.2 上下文映射关系

**上下文映射是限界上下文之间的关系**，定义了它们如何协作、如何通信、谁主导谁跟随。

**本系统的上下文映射图**：

```mermaid
graph LR
    Order[订单上下文<br/>Order Context] 
    Product[商品上下文<br/>Product Context]
    Inventory[库存上下文<br/>Inventory Context]
    Pricing[计价上下文<br/>Pricing Context]
    Marketing[营销上下文<br/>Marketing Context]
    Payment[支付上下文<br/>Payment Context]
    Search[搜索上下文<br/>Search Context]
    Supplier[供应商上下文<br/>Supplier Context]
    
    Order -->|Customer-Supplier| Product
    Order -->|Customer-Supplier| Inventory
    Order -->|Customer-Supplier| Pricing
    Order -->|Customer-Supplier| Marketing
    Order -->|Customer-Supplier| Payment
    
    Search -->|Conformist| Product
    Search -->|Open Host Service| Product
    
    Inventory -->|Anti-Corruption Layer| Supplier
    Product -->|Anti-Corruption Layer| Supplier
    
    Pricing -->|Shared Kernel| Marketing
```

**映射关系类型**：

| 关系类型 | 说明 | 本系统示例 | 实现方式 |
|---------|------|-----------|---------|
| **Customer-Supplier** | 下游（客户）依赖上游（供应商） | 订单 → 商品<br/>订单 → 库存 | 同步RPC调用 |
| **Conformist** | 下游完全遵循上游模型 | 搜索 → 商品 | 搜索直接使用商品模型 |
| **Anti-Corruption Layer** | 下游用防腐层保护自己 | 库存 → 供应商 | 适配器翻译外部模型 |
| **Open Host Service** | 上游提供公开服务 | 商品 → 搜索 | RESTful API + Events |
| **Shared Kernel** | 两个上下文共享部分模型 | 计价 ⇄ 营销 | 共享折扣计算规则 |
| **Published Language** | 上游定义标准数据格式 | 订单事件（Kafka） | Protobuf/JSON Schema |

**关键决策解析**：

**决策1：订单 → 商品（Customer-Supplier）**

```
为什么不是Conformist（遵奉者）？
• 订单需要保存商品快照（商品模型可能变化）
• 订单不应该被商品模型变更影响
• 订单有自己的领域模型（OrderItem vs Product）

为什么是Customer-Supplier？
• 订单依赖商品（下游依赖上游）
• 商品提供稳定的API（上游为下游服务）
• 变更需要协商（商品API变更需通知订单团队）
```

**决策2：库存 → 供应商（Anti-Corruption Layer）**

```
为什么需要防腐层？
• 供应商模型不稳定（50+供应商，接口各不相同）
• 防止供应商模型污染库存域
• 便于切换供应商（ACL隔离变化）

防腐层职责：
• 翻译外部模型 → 内部模型
• 统一异常处理
• 适配器模式（每个供应商一个适配器）
```

**决策3：计价 ⇄ 营销（Shared Kernel）**

```
为什么是Shared Kernel？
• 折扣计算规则在两个上下文都需要
• 规则变更需要两个上下文同步
• 共享折扣计算代码（避免重复）

Shared Kernel范围：
• DiscountRule（折扣规则接口）
• PriceBreakdown（价格明细结构）
• 仅共享"计算规则"，不共享"数据存储"
```

**上下文通信机制**：

| 场景 | 通信方式 | 协议 | 示例 |
|------|---------|------|------|
| **同步查询** | RPC | gRPC + Protobuf | 订单查询商品信息 |
| **同步操作** | RPC | gRPC + Protobuf | 订单预占库存 |
| **异步事件** | 消息队列 | Kafka + Protobuf | 订单创建 → 搜索更新销量 |
| **批量查询** | RPC | gRPC + Stream | 批量查询商品价格 |

### 16.3.3 边界划分实践案例

```
┌──────────────────────────────────────────────────────┐
│              接入层（API Gateway）                    │
│  • 鉴权、限流、路由、协议转换                         │
│  • Web/App/小程序统一接入                            │
└──────────────────────────────────────────────────────┘
                          ↓
┌──────────────────────────────────────────────────────┐
│             聚合层（Aggregation Service）             │
│  • 数据编排：并发调用多个微服务                       │
│  • 降级策略：服务故障时的降级处理                     │
│  • 缓存优化：聚合结果缓存                            │
└──────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│                   业务服务层（Microservices）                │
│  ┌────────┬────────┬────────┬────────┬────────┬────────┐   │
│  │ Product│Inventory│ Pricing│Marketing│ Order │ Payment│   │
│  │  商品  │  库存  │  计价  │  营销  │  订单 │  支付  │   │
│  └────────┴────────┴────────┴────────┴────────┴────────┘   │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌──────────────────────────────────────────────────────┐
│           基础设施层（Infrastructure）                │
│  • MySQL、Redis、Elasticsearch、Kafka               │
│  • 服务发现（Consul）、服务网格（Envoy）             │
│  • 监控告警（Prometheus、Grafana、Jaeger）          │
└──────────────────────────────────────────────────────┘
```

**分层职责**：

| 层级 | 服务 | 职责 | 不负责 |
|------|------|------|--------|
| **接入层** | API Gateway | 鉴权、限流、路由 | 业务逻辑、数据编排 |
| **聚合层** | Aggregation | 数据获取、编排、降级 | 具体业务计算 |
| **业务层** | Microservices | 单一业务领域逻辑 | 跨域数据获取 |
| **基础层** | Infra | 存储、消息、监控 | 业务规则 |

### 16.4.2 微服务拆分

**拆分原则**：
1. **按业务能力拆分**（而非技术层次）
2. **单一职责**：每个服务只负责一个限界上下文
3. **数据所有权**：每个服务拥有自己的数据库
4. **API优先**：服务间只通过API或事件通信

**核心服务清单**：

| 服务名称 | 职责 | 数据库 | QPS（峰值） | 团队规模 |
|---------|------|--------|------------|---------|
| **Product Center** | 商品信息、类目、属性 | MySQL（4分库） | 20000 | 12人 |
| **Inventory Service** | 库存管理、预占、扣减 | MySQL+Redis | 8000 | 10人 |
| **Pricing Service** | 价格计算、试算、快照 | MySQL | 15000 | 8人 |
| **Marketing Service** | 营销规则、优惠券、活动 | MySQL+Redis | 10000 | 12人 |
| **Order Service** | 订单创建、状态机、履约 | MySQL（8分库64表） | 5000 | 15人 |
| **Payment Service** | 支付、退款、对账 | MySQL | 6000 | 10人 |
| **Search Service** | 商品搜索、筛选、排序 | Elasticsearch | 15000 | 8人 |
| **User Service** | 用户信息、登录、权限 | MySQL | 8000 | 6人 |
| **Supplier Gateway** | 供应商对接、适配、熔断 | MySQL+Redis | 12000 | 15人 |

**聚合服务**：

| 服务 | 职责 | 依赖服务 |
|------|------|---------|
| **Search Aggregation** | 搜索结果聚合 | Search + Product + Inventory + Pricing |
| **Detail Aggregation** | 详情页聚合 | Product + Inventory + Pricing + Marketing |
| **Checkout Aggregation** | 结算页聚合 | Product + Inventory + Pricing + Marketing |

### 16.4.3 服务依赖关系

```mermaid
graph TB
    subgraph 接入层
        Gateway[API Gateway]
    end
    
    subgraph 聚合层
        SearchAgg[搜索聚合]
        DetailAgg[详情聚合]
        CheckoutAgg[结算聚合]
    end
    
    subgraph 业务服务层
        Product[商品中心]
        Inventory[库存服务]
        Pricing[计价服务]
        Marketing[营销服务]
        Order[订单服务]
        Payment[支付服务]
        Search[搜索服务]
    end
    
    subgraph 基础服务
        Supplier[供应商网关]
        User[用户服务]
    end
    
    Gateway --> SearchAgg
    Gateway --> DetailAgg
    Gateway --> CheckoutAgg
    Gateway --> Order
    
    SearchAgg --> Search
    SearchAgg --> Product
    SearchAgg --> Inventory
    SearchAgg --> Pricing
    
    DetailAgg --> Product
    DetailAgg --> Inventory
    DetailAgg --> Pricing
    DetailAgg --> Marketing
    
    CheckoutAgg --> Product
    CheckoutAgg --> Inventory
    CheckoutAgg --> Pricing
    CheckoutAgg --> Marketing
    
    Order --> Inventory
    Order --> Payment
    Order --> Supplier
    
    Inventory --> Supplier
    Product --> Supplier
```

**依赖原则**：
1. **上游 → 下游**：聚合层调用业务层，不反向依赖
2. **避免循环依赖**：严格禁止服务间循环调用
3. **异步解耦**：非核心路径使用Kafka事件异步
4. **降级友好**：下游故障不影响上游核心功能

### 16.4.4 数据流转

**同步数据流（关键路径）**：

```
用户搜索商品：
API Gateway → Search Aggregation 
            → Search Service（ES查询）
            → Product Service（批量获取基础信息）
            → Inventory Service（批量查库存）
            → Pricing Service（批量计算价格）
            ← 返回聚合结果

响应时间：< 200ms（P99）
```

**异步数据流（非关键路径）**：

```
订单创建成功 → Kafka Event：OrderCreated
            → 订阅者1：Inventory Service（确认扣减）
            → 订阅者2：Search Service（更新销量）
            → 订阅者3：User Service（积分增加）
            → 订阅者4：Data Team（数据分析）

最终一致性：< 5秒
```

**16.4小结**：

以上展示了系统的整体架构全貌：四层架构、12个核心微服务、服务依赖关系、数据流转模式。这些是16.3战略设计的具体落地——12个限界上下文对应12个微服务，上下文映射关系决定了服务间的集成方式。

接下来16.5节将讨论技术选型决策，16.6节将深入各个系统的详细设计。

---

## 16.4 整体架构设计（Application Architecture - 设计结果）

基于16.3节识别的12个限界上下文和上下文映射关系，本节展示如何将它们落地为具体的架构设计：四层架构、微服务拆分、服务依赖关系、数据流转模式。

**16.3 → 16.4的映射关系**：

```
16.3 限界上下文           →    16.4 微服务
├─ 订单上下文             →    Order Service
├─ 商品上下文             →    Product Center
├─ 库存上下文             →    Inventory Service
├─ 计价上下文             →    Pricing Service
├─ 营销上下文             →    Marketing Service
├─ 支付上下文             →    Payment Service
├─ 搜索上下文             →    Search Service
└─ 供应商上下文           →    Supplier Gateway

16.3 上下文映射           →    16.4 服务集成
├─ Customer-Supplier      →    同步RPC调用
├─ Anti-Corruption Layer  →    适配器模式
└─ Published Language     →    Kafka事件
```

### 16.4.1 分层架构

采用经典的四层架构，确保职责清晰、易于维护。

基于前面识别的限界上下文和映射关系，本节通过实际案例展示如何划分边界、重构边界。

**案例1：计价系统的边界重构**

**初始问题**：
- 价格计算逻辑分散在订单、营销、商品三个域
- 购物车、订单创建、支付确认三处价格计算不一致
- 无法支持"PDP加购试算"场景

**重构方案**：
1. **新建计价上下文**：职责是提供统一的试算接口
2. **定义边界**：
   - 计价上下文**不拥有**商品基础价、营销规则、订单状态
   - 对外提供 `Calculate(items, promotions, context) -> PriceBreakdown`
   - 各场景通过统一接口获取价格
3. **收益**：
   - 价格一致性得到保证
   - 营销规则变更只需在营销域发布事件
   - 支持了试算、价格预览、价格审计等新需求

**案例2：库存预占的归属**

**争议**：库存预占应该放在订单域还是库存域？

**决策**：放在库存域

**理由**：
- 库存域拥有库存数据所有权
- 预占是库存的一种状态（可售 → 预占 → 扣减）
- 订单域只需调用库存域的 `Reserve` 接口
- 降低耦合：订单域不需要了解库存的存储结构

### 16.4.4 集成模式选择

| 集成场景 | 模式 | 理由 |
|---------|------|------|
| 订单 → 商品 | 同步RPC | 需要实时获取商品信息，延迟<100ms |
| 订单 → 库存 | 同步RPC | 库存预占是核心路径，必须同步 |
| 订单 → 支付 | 同步RPC | 支付创建需要同步返回支付URL |
| 订单成功 → 搜索 | 异步事件 | 销量更新非核心路径，可最终一致 |
| 订单成功 → 积分 | 异步事件 | 积分增加非核心路径 |

**事件驱动示例**：

```go
// 订单域发布事件
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    // 创建订单...
    order := &Order{...}
    s.repo.Save(ctx, order)
    
    // 发布事件（Outbox模式）
    event := &OrderCreatedEvent{
        OrderID:    order.ID,
        UserID:     order.UserID,
        TotalPrice: order.TotalPrice,
        Items:      order.Items,
    }
    s.outbox.Publish(ctx, "order-events", event)
    
    return order, nil
}

// 搜索域订阅事件
func (s *SearchService) HandleOrderCreated(ctx context.Context, event *OrderCreatedEvent) error {
    // 更新商品销量（用于排序）
    for _, item := range event.Items {
        s.incrementSales(ctx, item.SkuID, item.Quantity)
    }
    return nil
}
```

### 16.4.5 跨系统事务处理

**Saga模式（编排）**：

```go
// 订单创建Saga
type CreateOrderSaga struct {
    inventoryClient rpc.InventoryClient
    marketingClient rpc.MarketingClient
    orderRepo       *OrderRepo
}

func (s *CreateOrderSaga) Execute(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    var reserveID string
    var couponLockID string
    
    // Step 1: 库存预占
    reserve, err := s.inventoryClient.ReserveStock(ctx, req.Items)
    if err != nil {
        return nil, fmt.Errorf("库存预占失败: %w", err)
    }
    reserveID = reserve.ReserveID
    defer func() {
        if err != nil {
            // 补偿：释放库存
            s.inventoryClient.ReleaseStock(ctx, reserveID)
        }
    }()
    
    // Step 2: 优惠券锁定
    couponLock, err := s.marketingClient.LockCoupon(ctx, req.CouponCode, req.UserID)
    if err != nil {
        return nil, fmt.Errorf("优惠券锁定失败: %w", err)
    }
    couponLockID = couponLock.LockID
    defer func() {
        if err != nil {
            // 补偿：释放优惠券
            s.marketingClient.UnlockCoupon(ctx, couponLockID)
        }
    }()
    
    // Step 3: 创建订单
    order := &Order{
        ID:           generateOrderID(),
        UserID:       req.UserID,
        Items:        req.Items,
        ReserveID:    reserveID,
        CouponLockID: couponLockID,
        Status:       StatusPendingPayment,
    }
    err = s.orderRepo.Save(ctx, order)
    if err != nil {
        return nil, fmt.Errorf("订单创建失败: %w", err)
    }
    
    return order, nil
}
```

### 16.4.6 防腐层设计

**防腐层（Anti-Corruption Layer）**：

```go
// 供应商响应模型（外部）
type SupplierFlightResponse struct {
    Code    string  `json:"code"`
    Message string  `json:"message"`
    Data    struct {
        FlightNo  string  `json:"flight_no"`
        Available int     `json:"available"`
        Price     float64 `json:"price"`
    } `json:"data"`
}

// 平台库存模型（内部）
type StockResponse struct {
    Available bool
    Quantity  int
    Message   string
}

// 防腐层：翻译外部模型 → 内部模型
func (a *FlightSupplierACL) TranslateStock(supplierResp *SupplierFlightResponse) *StockResponse {
    return &StockResponse{
        Available: supplierResp.Code == "SUCCESS" && supplierResp.Data.Available > 0,
        Quantity:  supplierResp.Data.Available,
        Message:   supplierResp.Message,
    }
}
```

**收益**：
- 领域层不被供应商模型污染
- 供应商接口变更时，修改集中在ACL
- 测试时可以使用Fake实现替代真实供应商

---

## 16.5 技术选型决策（Technology Architecture）

### 16.5.1 选型原则

**原则1：成熟度优先**
- 优先选择生产级成熟技术（避免踩坑）
- 社区活跃、文档完善、案例丰富
- 避免使用 alpha/beta 版本

**原则2：团队能力匹配**
- 技术栈与团队技能对齐
- 学习曲线可控（新技术培训 < 1个月）
- 有内部专家支持

**原则3：生态完整性**
- 工具链完善（测试、监控、部署）
- 第三方库丰富
- 云服务支持（AWS/GCP/阿里云）

**原则4：成本可控**
- 开源优先（降低License成本）
- 云服务按需使用（避免自建中间件）
- 运维成本可接受

### 16.5.2 Go生态选型

**语言选择：Go**

| 维度 | Go | Java | 理由 |
|------|-----|------|------|
| 性能 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 协程模型，高并发性能优异 |
| 开发效率 | ⭐⭐⭐⭐ | ⭐⭐⭐ | 编译快，部署简单（单一二进制） |
| 学习曲线 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | 语法简洁，容易上手 |
| 生态 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 微服务生态完善（gRPC/Consul/Envoy） |
| 团队能力 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | 团队有Go经验 |

**Web框架：Gin**
```go
// 理由：
// 1. 性能优异（httprouter，零内存分配）
// 2. 中间件丰富（鉴权、限流、日志）
// 3. 社区活跃（GitHub 70k+ stars）

router := gin.Default()
router.Use(middleware.Auth())
router.Use(middleware.RateLimit(1000))
router.GET("/products/:id", handler.GetProduct)
```

**ORM：GORM**
```go
// 理由：
// 1. 支持MySQL、PostgreSQL、SQLite
// 2. 关联查询、预加载、Hook机制完善
// 3. 自动迁移（开发环境）

type Product struct {
    ID       int64  `gorm:"primaryKey"`
    Title    string `gorm:"size:255;not null"`
    Price    int64  `gorm:"not null"`
}
```

**RPC：gRPC + Protobuf**
```go
// 理由：
// 1. 二进制序列化（性能优于JSON）
// 2. 强类型（编译期检查）
// 3. 支持流式调用（双向流）

service ProductService {
    rpc GetProduct(GetProductRequest) returns (GetProductResponse);
    rpc BatchGetProduct(BatchGetProductRequest) returns (stream Product);
}
```

**依赖注入：Google Wire**
```go
// 理由：
// 1. 编译时生成（无反射，性能高）
// 2. 类型安全（编译期检查依赖）
// 3. 官方支持（Google开源）

//go:generate wire
func InitializeApp() (*App, error) {
    wire.Build(
        NewDB,
        NewRedis,
        NewProductRepo,
        NewProductService,
        NewApp,
    )
    return nil, nil
}
```

### 16.4.3 数据库选型

**MySQL（主库）**

| 场景 | 选择理由 | 配置 |
|------|---------|------|
| 订单表 | ACID保证、事务支持 | InnoDB，8分库64表 |
| 商品表 | 关联查询、JOIN支持 | InnoDB，4分库 |
| 支付表 | 强一致性、金融级可靠性 | InnoDB，双主互备 |

**Redis（缓存 + 库存）**

| 场景 | 数据结构 | TTL |
|------|---------|-----|
| 商品详情 | Hash | 30分钟 |
| 库存数量 | String（Lua原子扣减） | 永久 |
| 券码池 | Set（SPOP原子弹出） | 永久 |
| 用户Session | String | 2小时 |

**Elasticsearch（搜索 + 日志）**

| 场景 | 索引设计 | 刷新间隔 |
|------|---------|---------|
| 商品搜索 | product_index（标题、类目、属性） | 30秒 |
| 订单查询 | order_index（订单号、用户ID、状态） | 1分钟 |
| 日志搜索 | log-{date}（按日分索引） | 5秒 |

### 16.4.4 中间件选型

**Kafka（消息队列）**

| 场景 | Topic | Partition | Replication |
|------|-------|-----------|-------------|
| 订单事件 | order-events | 16 | 3 |
| 库存事件 | inventory-events | 8 | 3 |
| 日志采集 | logs | 32 | 2 |

**Consul（服务发现）**
- 健康检查：HTTP/TCP/gRPC
- 配置中心：动态配置热更新
- KV存储：Feature Flag

**Envoy（Service Mesh）**
- 流量管理：灰度发布、A/B测试
- 可观测性：自动生成Trace
- 安全：mTLS加密

---

## 16.6 核心系统设计（Application + Data Architecture详细设计）

基于16.4的整体架构，本节深入每个核心系统的详细设计，包括应用层的业务逻辑设计和数据层的模型设计。

### 16.6.1 商品中心设计

#### 16.6.1.1 服务定位与职责

一句话概括，**商品中心 = 商品主数据 + 供给运营 + 库存管理 + 搜索导购中心**。

商品中心处在供应商供给、平台运营、C 端导购和交易系统之间。它不是简单的商品表 CRUD，也不是只维护标题、图片、类目和上下架状态的 PIM 系统；在数字商品平台里，商品中心还要承接商品如何进入平台、如何被运营维护、如何保持供应商数据新鲜、如何判断可售、如何支撑搜索列表和详情页，以及如何在下单前给订单系统提供稳定的商品快照和库存校验结果。

由于团队规模和系统演进阶段限制，本平台没有独立拆分库存中心和搜索中心，库存能力与搜索导购能力都由商品中心内部模块承接。这里需要特别说明：**库存和搜索归商品中心，不代表商品主数据、库存状态、搜索索引混在一起**。商品中心内部仍然按六个域拆分，分别管理不同的数据模型、生命周期和对外契约，避免商品定义、库存状态、搜索索引、供应商模型和交易状态互相污染。

从业务链路看，商品中心主要承接三类问题：

1. **供给侧问题**：商品从哪里来，如何上传、审核、同步、修正和下架。
2. **导购侧问题**：用户如何在首页、列表页、详情页看到正确、可搜索、可筛选、可展示的商品。
3. **交易前问题**：商品是否存在、是否上架、是否可售、库存是否满足、是否需要供应商实时确认。

因此，商品中心内部可以拆成六个稳定的职责域：

| 职责域 | 解决的问题 | 关键输出 |
|-------|-----------|---------|
| **商品主数据域** | 定义商品是什么，包括类目、SPU/SKU、属性、素材、业务实体和商品状态 | 标准商品模型、类目属性、商品详情、商品快照 |
| **商品供给与运营域** | 管理商品如何进入平台、如何审核、如何编辑、如何上下架 | Listing Task、审核结果、发布事件、变更日志 |
| **供应商商品同步域** | 管理外部供应商商品如何映射、同步、刷新和补偿 | 供应商映射、同步任务、数据完整性报告 |
| **库存与可售域** | 判断商品是否能卖，统一无限库存、池化库存和实时库存差异 | 库存查询结果、预占结果、可售状态、券码发放结果 |
| **搜索与导购域** | 支撑首页、列表页、详情页的召回、筛选、排序、Hydrate 和缓存 | ES 索引、搜索结果、详情页聚合数据、降级结果 |
| **系统集成与事件域** | 向营销、计价、订单、履约、供应商网关和数据平台输出稳定契约 | 查询 API、领域事件、CDC、质量监控数据 |

**商品中心内部模块划分**：

```
商品中心 Product Center
├─ 1. 商品主数据域（Product Master Data）
│  ├─ 类目：前台类目、后台类目、类目层级、类目属性模板
│  ├─ SPU/SKU：标准商品、销售单元、组合 SKU
│  ├─ 商品属性：基础属性、业务属性、动态属性、扩展属性
│  ├─ 业务实体：运营商、银行、航司、酒店、影院、商户、门店
│  ├─ 商品素材：标题、描述、图片、Icon、展示标签
│  └─ 商品状态：草稿、待审核、已上架、已下架、已归档
│
├─ 2. 商品供给与运营域（Supply & Operation）
│  ├─ 商品供给：人工上传、批量上传、模板下载、供应商导入
│  ├─ 数据校验：字段校验、类目校验、属性校验、价格/库存预检
│  ├─ 审核发布：新商品审核、编辑审核、高风险变更审核
│  ├─ 商品运营：编辑、批量编辑、上下架、排序、入口配置
│  ├─ 质量治理：缺字段检查、异常价格检查、库存异常检查
│  └─ 操作追踪：Listing Task、审核日志、变更日志、状态流水
│
├─ 3. 供应商商品同步域（Supplier Sync）
│  ├─ 商品映射：平台 SKU 与供应商 SKU、外部资源 ID、业务实体 ID 映射
│  ├─ 静态同步：酒店基础信息、影院信息、商户门店、票务基础数据
│  ├─ 动态同步：可缓存价格、库存水位、上下架状态
│  ├─ 同步任务：全量同步、增量同步、供应商 Push、接入层 Push
│  └─ 同步治理：重试、补偿、告警、数据完整性巡检
│
├─ 4. 库存与可售域（Stock & Sellable）
│  ├─ 库存模型：无限库存、池化库存、实时库存
│  ├─ 库存来源：本地 DB、券码池、供应商接入层 API、供应商 API
│  ├─ 交易动作：查询、预占、释放、扣减、回补
│  ├─ 券码管理：券码池、发码、核销状态、过期管理
│  └─ 可售判断：商品状态、库存状态、供应商可用性、业务规则
│
├─ 5. 搜索与导购域（Search & Discovery）
│  ├─ 搜索索引：ES 索引构建、索引刷新、索引回滚
│  ├─ 召回筛选：关键词、类目、品牌/Carrier、城市、商户、标签
│  ├─ 排序展示：运营排序、销量、价格、活动标签、库存状态
│  ├─ Hydrate：补齐商品详情、库存状态、展示价、营销标签
│  ├─ 页面能力：首页入口、列表页、详情页、商品缓存
│  └─ 降级策略：缓存兜底、隐藏营销标签、库存弱展示
│
└─ 6. 系统集成与事件域（Integration & Event）
   ├─ 对营销：类目、Tag、商品范围、圈品能力、可营销状态
   ├─ 对计价：基础价、类目、属性、库存上下文、能力配置
   ├─ 对订单：商品快照、上下架状态、可售校验、库存预占结果
   ├─ 对履约：履约类型、供应商映射、发货/出票/充值能力配置
   ├─ 对供应商网关：查价、查库存、同步任务、履约参数映射
   ├─ 对数据平台：CDC、商品变更日志、质量监控、经营分析
   └─ 事件机制：商品创建、商品更新、上下架、库存变化、同步失败
```

**与其他系统的边界**：

| 系统 | 商品中心提供 | 对方负责 |
|------|-------------|---------|
| **营销系统** | 类目、Tag、业务实体、商品范围、可营销状态 | 活动配置、圈品、优惠券、满减/折扣规则 |
| **计价中心** | 基础价、类目、属性、库存上下文、能力配置 | PDP 价格、结算页试算价、下单价、支付价、结算价 |
| **订单系统** | 商品详情、商品快照、上下架状态、库存可售性、预占结果 | 订单创建、订单状态机、支付前后流转 |
| **履约系统** | 履约类型、供应商映射、商品能力配置 | 出票、预订确认、充值、发券、销账、履约补偿 |
| **供应商网关/供应商接入层** | 平台 SKU、供应商映射、同步任务、商品能力配置 | 外部 API 适配、供应商查价查库存、供应商履约调用 |

因此，商品中心的定位不是“商品表 CRUD 服务”，而是数字商品平台交易前链路的核心系统。它统一商品定义、库存能力和搜索导购能力，屏蔽供应商商品模型差异，对外稳定输出商品、库存、搜索结果和能力配置，并通过事件机制驱动营销、计价、订单和履约系统协同。

---

#### 16.6.1.2 核心设计挑战：异构商品模型

数字商品平台的商品中心，最大的难点不是“字段很多”，而是**不同品类对交易对象的定义并不相同**。实物电商的交易对象通常比较稳定：用户买的是一个 SKU，平台围绕 SKU 管库存、价格、物流和售后即可。但 OTA、O2O 和虚拟商品不是这样。它们卖的可能是一次账户余额变更、一个数字凭证、一项到店服务权益、一个特定时间窗口内的资源确认权，或者一次供应商实时返回的临时报价。

所以，这里的“商品”不能简单理解为 `Product + SKU`。更准确的说法是：**商品中心要统一的是交易前的经营表达，而不是所有品类的实时交易状态**。

**1. 不同品类卖的不是同一种东西**

| 品类类型 | 用户实际购买的是什么 | 典型品类 | 核心复杂度 |
|---------|------------------|---------|-----------|
| **账户变更型** | 给外部账户充值、销账或开通权益 | Topup、账单缴费、流量包 | 下单前要校验账户，支付后要确认外部账户状态变化 |
| **数字凭证型** | 一个可兑换、可消费或可核销的凭证 | Gift Card、E-Voucher、Payment Voucher | 商品定义与券码库存必须隔离，发码后状态不可随意回滚 |
| **到店服务型** | 某商户或门店的一次服务权益 | Local Service、Deal Voucher | 商户、门店、核销、过期、退款规则比 SKU 字段更重要 |
| **资源确认型** | 某个时间窗口下的稀缺资源确认权 | Flight、Hotel、Movie、Train、Bus | 价格和库存高度实时，通常需要供应商确认或锁定 |
| **组合套餐型** | 多个权益或资源的组合 | 电影 + 小食、酒店 + 活动券 | 需要处理组合价、组合库存、部分履约和部分退款 |

如果用实物电商的思路强行套这些品类，会遇到五类问题：

1. **SKU 爆炸**：把机票、酒店、电影票的每次报价都沉淀成 SKU，会产生海量临时 SKU，而且很快过期。
2. **字段污染**：把所有品类字段都放进一张商品宽表，最后会变成大量空字段、重复字段和语义不清的扩展字段。
3. **实时性误判**：把动态价格和实时库存当成商品主数据保存，会导致列表页、详情页和下单价频繁不一致。
4. **流程耦合**：把账号校验、账单查询、锁座、房态确认、券码发放都写进商品 CRUD，会让商品中心变成交易系统和履约系统的混合体。
5. **售后规则丢失**：OTA/O2O 的退改签、取消政策、核销后不可退等规则不是普通展示字段，而是影响订单状态机和资损风险的交易规则。

**2. 三种常见解决方案**

面对异构商品，业界通常会经历三种建模方案。它们不是简单的“谁对谁错”，而是适用于不同阶段、不同品类复杂度。

**方案A：标准 SPU/SKU + EAV/ExtInfo 扩展**

这是最接近传统电商商品中心的方案。核心模型是：

```text
Category
  → SPU
  → SKU
  → Attribute / EAV
  → ExtInfo JSON
```

所有商品尽量表达为 SPU/SKU。固定字段放主表，可搜索、可筛选字段放属性表，品类专属展示字段放 `ExtInfo JSON`。

| 维度 | 评价 |
|------|------|
| **优点** | 简单直观，运营后台容易实现，适合 Topup、Gift Card、E-Voucher、Local Service 等固定面额或固定券模板商品 |
| **缺点** | 难以表达 Flight、Hotel、Movie 这类实时供给；如果把日期、舱位、房态、座位都 SKU 化，会造成 SKU 爆炸 |
| **适用阶段** | 平台早期、品类较少、以固定数字商品为主 |

这套方案的问题在于：它容易让团队误以为“所有东西都应该变成 SKU”。一旦把实时报价、房态、座位图、账单金额都塞进 SKU，商品中心就会从主数据系统滑向交易结果缓存系统。

**方案B：资源中心化模型**

OTA 和 O2O 平台常见的另一种做法，是先把业务资源标准化，再在资源上包装可售商品。

```text
Resource
  → Product Package
  → Offer / Rate Plan
  → Availability
```

这里的 Resource 可以是酒店、房型、城市、机场、车站、影院、影厅、影片、商户、门店、账单机构等。SPU/SKU 不再是唯一核心，而是资源上的销售包装。

| 维度 | 评价 |
|------|------|
| **优点** | 适合酒店、电影、本地服务、交通票务；避免把所有资源组合都沉淀成 SKU；供应商资源映射更清晰 |
| **缺点** | 模型理解成本更高；只解决资源建模，还不能完整表达用户输入、预订锁定、履约和售后 |
| **适用阶段** | 平台开始接入 OTA/O2O 品类，资源、门店、城市、场次、房型等成为核心数据 |

资源中心化模型能解决“商品背后是什么资源”的问题，但还没有完整回答“这个资源在交易链路里怎么报价、怎么锁定、怎么履约、怎么退款”。

**方案C：商品交易契约模型（推荐）**

更完整的做法是把商品中心从“商品字段存储系统”升级为“交易前契约系统”。它不是推翻 SPU/SKU，也不是单纯资源化，而是把两者组合起来：

```text
SPU/SKU               表达平台商品定义
Resource              表达商品背后的业务资源
Offer / Rate Plan     表达售卖条件和报价规则
Capability Matrix     表达类目能力差异
Runtime Context       表达交易前运行时上下文
```

这套方案的核心思想是：

> 商品中心统一的是经营表达和交易前契约，不是所有品类的实时资源状态。

| 维度 | 评价 |
|------|------|
| **优点** | 解释力强，能同时覆盖固定 SKU、资源型商品和实时供给商品；边界清晰，适合书籍总结和面试表达 |
| **缺点** | 初期理解成本较高，需要治理“哪些数据稳定、哪些数据实时、哪些数据只进快照” |
| **适用阶段** | 多品类平台，尤其是同时覆盖 Topup、Bill、Voucher、Hotel、Flight、Movie、Local Service 的平台 |

因此，本章采用方案 C 作为推荐方案。它吸收方案 A 的 SPU/SKU 基础能力，也吸收方案 B 的 Resource 建模能力，再通过能力矩阵和运行时上下文把不同品类的交易差异显式表达出来。

**3. 八层商品交易模型**

对 OTA、O2O 和虚拟商品来说，一个可交易商品通常可以拆成八层。不是每个品类都完整使用八层，但这个分层能帮助我们判断“什么应该进商品中心，什么应该留给库存、计价、订单和履约系统”。

| 层次 | 解决的问题 | 商品中心是否负责 | 示例 |
|------|-----------|----------------|------|
| **Product Definition** | 平台如何运营和展示这个商品 | 负责 | 类目、标题、图片、品牌/实体、基础描述、上下架状态 |
| **Resource** | 商品背后的资源是什么 | 负责稳定部分 | 酒店、房型、影院、场次、商户、门店、账单机构、城市站点 |
| **Offer / Rate Plan** | 在某个上下文下如何报价 | 负责配置，不负责所有实时结果 | 面额、套餐、日历价规则、供应商报价计划、活动价输入 |
| **Availability** | 当前是否可买 | 负责统一入口和可售判断 | 券码池库存、供应商实时库存、房态、座位、通道可用性 |
| **Input Schema** | 下单前需要用户提供什么 | 负责配置 | 手机号、账单号、乘客证件、入住人、座位选择、邮箱 |
| **Booking / Lock** | 支付前是否需要锁定资源 | 只负责能力配置和结果引用 | 占座、锁房、锁券码、锁账单金额、锁场次座位 |
| **Fulfillment Contract** | 支付后如何交付 | 负责履约能力配置 | 充值、销账、发码、出票、预订确认、到店核销 |
| **Refund / After-sale Rule** | 失败或退款时如何处理 | 负责规则配置和快照输入 | 未核销可退、已出票退改签、取消政策、失败自动退款 |

这八层之间的关系可以理解为：

```text
Product Definition  定义平台卖什么
  → Resource        指向背后的资源
  → Offer           生成可展示或可购买的报价
  → Availability    判断当前是否可买
  → Input Schema    收集交易所需信息
  → Booking / Lock  锁定稀缺资源或金额
  → Fulfillment     支付后完成数字交付
  → Refund Rule     失败或售后时决定如何回滚
```

这套分层的价值在于：**它不要求所有品类长得一样，但要求所有品类在交易链路里说清楚自己处在哪一层、依赖哪些层、哪些数据需要实时确认**。

**4. 典型品类的八层映射**

| 品类 | Product Definition | Resource | Offer / Rate Plan | Availability | Input Schema | Booking / Lock | Fulfillment | Refund / After-sale |
|------|-------------------|----------|-------------------|--------------|--------------|----------------|-------------|---------------------|
| **Topup** | 运营商、面额、套餐说明 | 手机号账户、区域 | 固定面额/套餐价 | 供应商通道可用性 | 手机号、区域 | 通常无需锁定 | 充值到账 | 失败退款，成功后通常不可退 |
| **Bill** | 账单机构、账单类型 | 用户账单账户 | 查账后生成金额 | 账单是否可缴 | 账单号、用户标识 | 可锁定账单金额或查询流水 | 代缴销账 | 已销账通常不可退 |
| **Gift Card** | 品牌、面额、有效期、使用说明 | 券码池 | 固定面额/折扣价 | 未分配券码数量 | 收件账号/邮箱 | 可在支付后分配，也可提前锁码 | 发码 | 未发码可退，已发码受限 |
| **E-Voucher / Local Service** | 商户、门店、券模板、核销规则 | 门店服务能力、券码池 | 券售价/活动价 | 本地库存/券码数量 | 门店、购买数量 | 可锁库存或支付后扣减 | 发券 + 到店核销 | 未核销可退，已核销不可退 |
| **Flight / Train / Bus** | 城市、站点、承运方、基础运营配置 | 班次、舱位/座位 | 供应商实时报价 | 实时座位/占座结果 | 乘客、证件、行李等 | 创单前占座/预订 | 出票 | 退改签规则复杂 |
| **Hotel** | 酒店、房型、设施、地理位置、政策 | 房型 + 日期范围 | Rate Plan / 日历价 / 动态价 | 房态确认 | 入住人、日期、人数 | 预订确认或担保锁房 | 预订确认 | 受取消政策约束 |
| **Movie** | 影片、影院、影厅、场次基础信息 | 场次 + 座位 | 场次价/套餐价 | 座位图和锁座状态 | 座位、手机号 | 锁座 | 出票/取票码 | 通常不可退或限时退 |

这张表说明了一个关键事实：**同样叫商品，但不同品类的“可售单元”可能位于不同层次**。Gift Card 的可售单元很接近 SKU；Hotel 的可售单元是“房型 + 日期范围 + Rate Plan”；Flight 的可售单元是一次实时报价和占座结果；Bill 的可售单元甚至要在用户输入账单号之后才形成。

**5. 商品中心的职责边界**

基于八层模型，商品中心应该重点负责交易前可复用、可运营、可搜索、可配置的部分：

```text
商品中心负责：
  Product Definition：类目、SPU/SKU、标题、图片、状态、Tag
  Resource 稳定部分：酒店、房型、影院、商户、门店、城市、站点、账单机构
  Offer 配置：基础价、面额、套餐、价格规则输入、供应商报价映射
  Availability 入口：库存类型、库存来源、可售规则、查询/预占能力
  Input Schema：手机号、账单号、乘客、入住人、座位等表单配置
  Contract 配置：履约类型、退款规则、供应商映射、能力开关

商品中心不负责：
  实时航班报价、实时房态房价、座位锁定状态
  用户账单金额、支付结果、订单履约状态、售后处理结果
```

这样划分之后，商品中心不会因为接入一个新品类就不断膨胀。新增品类时，优先回答八个问题：

1. 它的稳定商品定义是什么？
2. 它依赖哪些资源？
3. 报价是平台配置还是供应商实时返回？
4. 可用性由谁确认？
5. 用户下单前需要输入什么？
6. 是否需要预订、锁定或占用资源？
7. 支付后如何履约？
8. 失败、取消、退款时遵循什么规则？

这八个问题回答清楚，商品中心、计价、库存、订单、履约和售后之间的边界也就清楚了。

**6. 建模原则**

最终的建模原则可以总结为六句话：

1. **不要用一个 SKU 表硬套所有品类**：SKU 是稳定可售单元，不是所有实时报价和资源组合的容器。
2. **静态资源和动态资源分离**：酒店资料、影院资料可以同步；房态、座位、报价必须按时效处理。
3. **商品定义和交易结果分离**：商品中心保存能力和规则，订单/履约系统保存每笔交易的状态。
4. **用户输入配置化**：不同品类的表单和校验规则要通过 Input Schema 表达，避免写死在交易代码里。
5. **履约和售后契约前置**：商品中心要告诉订单系统“这个商品怎么履约、怎么退”，但不处理具体订单的履约状态。
6. **供应商差异通过映射和防腐层隔离**：商品中心只保留平台模型和供应商映射，不让供应商字段污染主模型。

这一节的核心结论是：**商品中心真正统一的是经营表达和交易前契约，而不是统一所有品类的实时资源状态**。这是数字商品平台避免商品模型失控的关键。

---

#### 16.6.1.3 统一商品模型设计

商品中心的统一模型目标不是把所有品类强行压成同一种 SKU，而是提供一个稳定的“商品表达框架”，让不同品类都能被运营、搜索、计价、下单和履约系统理解。

**核心模型分层**：

| 模型 | 作用 | 示例 |
|------|------|------|
| **Category** | 统一品类层级和能力开关 | `40102` 机票、`10102` 话费充值、`70101` Deal Voucher |
| **Resource** | 表达商品背后的稳定业务资源 | 酒店、房型、城市、机场、影院、商户、门店、账单机构 |
| **Carrier / Brand** | 统一业务实体、品牌、运营商、机构 | 某运营商、某礼品卡品牌、某酒店品牌、某银行、某影院 |
| **SPU** | 表达平台商品或商品族 | 某品牌礼品卡、某酒店商品页、某商户套餐 |
| **SKU** | 表达稳定可售单元或销售模板 | 100 元礼品卡、某券模板、某充值面额、某房型 + Rate Plan |
| **Offer / Rate Plan** | 表达报价和售卖条件 | 固定面额、套餐价、含早/无早、可取消/不可取消 |
| **Attribute / EAV** | 支持可搜索、可筛选、可分析属性 | 面额、有效期、城市、商户类型、是否支持退款 |
| **ExtInfo JSON** | 承接低频、展示型、品类专属字段 | 酒店设施、券使用说明、账单字段配置 |
| **Supplier Mapping** | 连接平台商品与外部供应商资源 | 平台 SKU ↔ 供应商 SKU / 外部资源 ID / 业务实体 ID |
| **Category Capability** | 表达类目在交易链路中的能力差异 | 是否实时查价、是否需要输入账号、是否需要锁座/锁房 |
| **Runtime Context** | 为列表、详情、结算、创单组装交易前上下文 | 商品定义 + 资源 + 报价 + 可售 + 输入 + 履约 + 售后 |

这套模型可以理解为三层：

```text
第一层：稳定主数据
  Category + Resource + SPU + SKU + Attribute

第二层：交易前契约
  Offer / Rate Plan + Capability + Input Schema + Fulfillment Rule + Refund Rule

第三层：运行时上下文
  Runtime Context = 稳定主数据 + 交易前契约 + 实时查询结果
```

其中第一层主要持久化在商品中心；第二层也是商品中心负责维护，但会被计价、订单、履约和售后系统消费；第三层通常不是永久主数据，而是在搜索、详情、结算、创单等场景下按需组装，并在订单创建时形成订单快照。

**设计原则**：

1. **类目表达业务类型**：不要再额外引入 `product_type` 与 `category_id` 互相重叠，类目编码本身可以表达一级、二级、三级业务含义。
2. **Resource 表达稳定业务资源**：酒店、房型、影院、商户、门店、城市、机场、车站等不是普通 SKU 字段，而是可以被多个商品、多个供应商和多个场景复用的资源。
3. **Carrier / Brand 表达业务实体**：运营商、银行、航司、影院、酒店品牌、商户都可以归入业务实体模型，再通过实体类型区分。
4. **SPU/SKU 表达可运营商品**：Gift Card、Voucher、Topup 面额适合沉淀 SKU；Flight 搜索结果不适合沉淀完整 SKU，只沉淀城市、站点、航司等基础资源。
5. **Offer / Rate Plan 表达售卖条件**：酒店的含早/无早、可取消/不可取消，电影套餐，礼品卡面额，都应该从“商品是什么”里拆出来，作为“如何售卖”的配置。
6. **EAV 只放可检索属性**：需要筛选、搜索、分析的字段进入属性表；仅用于展示的复杂结构进入 `ExtInfo`。
7. **动态价格和实时库存不进商品主表**：商品主表保存稳定定义，动态报价、房态、座位库存通过缓存、供应商查询和订单快照处理。

---

#### 16.6.1.4 商品中心核心表设计

商品中心的表设计要支撑三件事：稳定主数据、灵活品类差异、可追溯运营链路。下面是核心表的定位，不要求所有字段一次性设计到位，但边界要清晰。

| 表名 | 定位 | 关键内容 |
|------|------|---------|
| `category_tab` | 类目树 | 类目编码、父类目、层级、名称、排序、状态、能力开关 |
| `carrier_brand_tab` | 业务实体/品牌/运营商 | 实体类型、名称、Logo、国家/地区、状态、扩展配置 |
| `resource_tab` | 统一业务资源 | 资源类型、资源编码、名称、父资源、国家/城市、状态、通用属性 |
| `resource_ext_*_tab` | 高频资源扩展表 | 酒店、房型、门店、影院、航线等高频字段，避免全部塞进 JSON |
| `product_spu_tab` | 标准商品或商品族 | SPU Code、类目、品牌/实体、标题、状态、素材 |
| `product_sku_tab` | 可售单元 | SKU Code、SPU ID、基础价、销售状态、库存类型、履约类型 |
| `product_resource_mapping_tab` | 商品与资源关系 | SKU/SPU 与酒店、房型、门店、影院、城市等资源的关系 |
| `product_offer_tab` | 商品报价配置 | 固定价、套餐价、展示价、报价来源、价格生效范围 |
| `rate_plan_tab` | 售卖条件计划 | 早餐、取消政策、支付方式、入住人数、供应商报价计划 |
| `product_attr_definition_tab` | 属性定义 | 属性 Code、属性类型、适用类目、是否可搜索/筛选 |
| `product_attr_value_tab` | 商品属性值 | SKU/SPU 与属性值绑定，用于筛选、搜索、分析 |
| `product_ext_info_tab` | 品类扩展信息 | JSON 结构，保存低频展示型、品类专属字段 |
| `supplier_product_mapping_tab` | 供应商商品映射 | 平台 SKU/SPU 与供应商 SKU、外部资源 ID、业务实体 ID 的映射 |
| `category_capability_tab` | 类目能力矩阵 | 商品模型类型、报价类型、库存类型、输入 Schema、锁定模式、履约类型、售后规则 |
| `input_schema_tab` | 用户输入表单配置 | 手机号、账单号、乘客、入住人、邮箱、座位选择等输入字段 |
| `fulfillment_rule_tab` | 履约契约配置 | 充值、销账、发码、出票、预订确认、核销等履约模式 |
| `refund_rule_tab` | 售后规则配置 | 是否可退、是否人工审核、取消政策、核销后限制、供应商退改规则 |
| `product_stock_tab` | 库存与可售状态 | 库存类型、库存来源、可售状态、库存数量、更新时间 |
| `voucher_code_pool_tab` | 券码池 | 券码、SKU、状态、有效期、分配订单、核销状态 |
| `product_supply_task` | 商品供给任务 | 导入批次、任务状态、操作人、成功/失败数量、错误文件 |
| `product_audit_log_tab` | 审核日志 | 审核对象、变更内容、审核结论、审核人、驳回原因 |
| `product_change_log_tab` | 变更日志 | 商品字段变更前后值、操作来源、TraceID、操作人 |
| `product_search_index_task_tab` | 搜索索引任务 | 索引动作、目标 SKU、任务状态、重试次数、失败原因 |

**方案 3 的核心 ER 关系**：

```mermaid
erDiagram
    CATEGORY_TAB ||--o{ PRODUCT_SPU_TAB : contains
    PRODUCT_SPU_TAB ||--o{ PRODUCT_SKU_TAB : contains

    CATEGORY_TAB ||--|| CATEGORY_CAPABILITY_TAB : configures
    CATEGORY_TAB ||--o{ RATE_PLAN_TAB : defines
    CATEGORY_TAB ||--o{ INPUT_SCHEMA_TAB : defines
    CATEGORY_TAB ||--o{ FULFILLMENT_RULE_TAB : defines
    CATEGORY_TAB ||--o{ REFUND_RULE_TAB : defines

    PRODUCT_SKU_TAB ||--o{ PRODUCT_OFFER_TAB : has
    RATE_PLAN_TAB ||--o{ PRODUCT_OFFER_TAB : applies_to

    PRODUCT_SKU_TAB ||--|| PRODUCT_STOCK_TAB : has
    PRODUCT_SKU_TAB ||--o{ VOUCHER_CODE_POOL_TAB : allocates

    PRODUCT_SPU_TAB ||--o{ PRODUCT_RESOURCE_MAPPING_TAB : maps
    PRODUCT_SKU_TAB ||--o{ PRODUCT_RESOURCE_MAPPING_TAB : maps
    RESOURCE_TAB ||--o{ PRODUCT_RESOURCE_MAPPING_TAB : referenced_by

    RESOURCE_TAB ||--o{ RESOURCE_RELATION_TAB : from_resource
    RESOURCE_TAB ||--o{ RESOURCE_RELATION_TAB : to_resource

    RESOURCE_TAB ||--o{ SUPPLIER_RESOURCE_MAPPING_TAB : maps_to_supplier
    PRODUCT_SKU_TAB ||--o{ SUPPLIER_PRODUCT_MAPPING_TAB : maps_to_supplier
    PRODUCT_OFFER_TAB ||--o{ SUPPLIER_PRODUCT_MAPPING_TAB : maps_to_supplier

    CATEGORY_TAB {
        bigint category_id PK
        varchar category_code
        bigint parent_id
        int level
        varchar name
        varchar status
    }

    CATEGORY_CAPABILITY_TAB {
        bigint capability_id PK
        bigint category_id FK
        varchar product_model_type
        varchar offer_type
        varchar availability_type
        varchar booking_mode
        varchar fulfillment_type
        varchar refund_rule_id
        varchar supplier_dependency
    }

    PRODUCT_SPU_TAB {
        bigint spu_id PK
        bigint category_id FK
        varchar spu_code
        varchar title
        bigint brand_id
        varchar status
        json ext_info
    }

    PRODUCT_SKU_TAB {
        bigint sku_id PK
        bigint spu_id FK
        varchar sku_code
        varchar sku_name
        bigint base_price
        varchar inventory_type
        varchar fulfillment_type
        varchar status
        json ext_info
    }

    RESOURCE_TAB {
        bigint resource_id PK
        varchar resource_type
        varchar resource_code
        varchar name
        bigint parent_resource_id
        varchar country_code
        varchar city_code
        varchar status
        json attributes
    }

    PRODUCT_RESOURCE_MAPPING_TAB {
        bigint id PK
        bigint spu_id FK
        bigint sku_id FK
        bigint resource_id FK
        varchar relation_type
        int priority
        varchar status
    }

    RESOURCE_RELATION_TAB {
        bigint id PK
        bigint from_resource_id FK
        bigint to_resource_id FK
        varchar relation_type
        varchar status
    }

    PRODUCT_OFFER_TAB {
        bigint offer_id PK
        bigint sku_id FK
        bigint rate_plan_id FK
        varchar offer_type
        bigint price
        varchar currency
        varchar price_rule
        datetime valid_from
        datetime valid_to
        varchar status
    }

    RATE_PLAN_TAB {
        bigint rate_plan_id PK
        bigint category_id FK
        varchar plan_code
        varchar meal_type
        varchar cancel_policy
        varchar payment_type
        json constraints
        varchar status
    }

    PRODUCT_STOCK_TAB {
        bigint stock_id PK
        bigint sku_id FK
        varchar stock_type
        varchar source_type
        int quantity
        boolean sellable
        datetime updated_at
    }

    VOUCHER_CODE_POOL_TAB {
        bigint code_id PK
        bigint sku_id FK
        varchar code_status
        datetime expire_at
        bigint assigned_order_id
        varchar redeem_status
    }

    INPUT_SCHEMA_TAB {
        bigint schema_id PK
        bigint category_id FK
        varchar scene
        json fields
        json validation_rules
        varchar status
    }

    FULFILLMENT_RULE_TAB {
        bigint rule_id PK
        bigint category_id FK
        varchar fulfillment_type
        varchar mode
        int timeout_sec
        json params
        varchar status
    }

    REFUND_RULE_TAB {
        bigint refund_rule_id PK
        bigint category_id FK
        boolean refundable
        boolean need_review
        json policy
        varchar status
    }

    SUPPLIER_RESOURCE_MAPPING_TAB {
        bigint id PK
        bigint resource_id FK
        bigint supplier_id
        varchar supplier_resource_code
        varchar supplier_resource_type
        varchar status
    }

    SUPPLIER_PRODUCT_MAPPING_TAB {
        bigint id PK
        bigint sku_id FK
        bigint offer_id FK
        bigint supplier_id
        varchar supplier_product_code
        varchar mapping_status
        json ext_ref
    }
```

一个重要经验是：**表结构要承认异构，而不是掩盖异构**。主表保持稳定，资源表承接业务资源，Offer/Rate Plan 表承接售卖条件，能力矩阵承接流程差异，映射表连接供应商，日志表保证可追溯。这样既不会让主表无限膨胀，也不会每新增一个品类就新建一整套孤立模型。

资源表建议采用“统一资源表 + 高频扩展表”的方式：

```text
resource_tab
  保存资源身份、资源类型、名称、父子关系、状态、国家/城市等通用字段

resource_ext_hotel_tab / resource_ext_room_type_tab
  保存酒店星级、地址、经纬度、设施、房型面积、床型等高频字段

resource_ext_merchant_tab / resource_ext_outlet_tab
  保存商户、门店、营业时间、地理位置、核销能力等高频字段

resource_ext_cinema_tab / resource_ext_route_tab
  保存影院、影厅、城市站点、航线/车线等高频字段
```

这样设计的原因是：不是所有资源都值得单独建完整模型，但高频检索、高频展示、高频排序的资源字段不能长期躲在 JSON 里。统一 `resource_tab` 负责身份和关系，扩展表负责高频业务字段。

---

#### 16.6.1.5 不同品类的数据存储样例

不同品类进入商品中心时，关键是判断“哪些信息稳定，哪些信息动态，哪些信息不应该沉淀”。下面按典型品类说明。

| 品类 | SPU/SKU 存什么 | Resource 存什么 | Offer / Rate Plan 存什么 | 动态数据在哪里 |
|------|---------------|----------------|--------------------------|---------------|
| **Topup** | 运营商面额 SKU、套餐说明、基础价、可售状态 | 运营商、国家/地区、号码归属规则 | 固定面额、套餐价、手续费规则 | 手机号校验、供应商通道状态实时查询或短缓存 |
| **Bill** | 账单机构商品、账单类型、缴费入口 | 账单机构、账单地区、账单账号类型 | 手续费规则、是否支持部分支付、滞纳金规则 | 用户账单金额、欠费明细、账单可缴状态实时查询 |
| **Gift Card** | 品牌 + 面额 SKU、有效期、使用说明 | 礼品卡品牌、券码池资源 | 固定面额、折扣价、发码规则 | 券码分配结果、用户核销状态进入履约/核销链路 |
| **E-Voucher / Local Service** | 券模板 SKU、购买限制、可用时间、核销规则摘要 | 商户、门店、服务项目、券码池 | 券售价、活动价、门店适用范围 | 发券、核销、过期、退款状态进入履约/售后链路 |
| **Flight / Train / Bus** | 通常不沉淀完整行程 SKU，只存基础运营配置 | 城市、机场/车站、航司/车司、线路 | 供应商报价计划、服务费、加价规则 | 航班/班次报价、座位、舱位、占座结果实时查询 |
| **Hotel** | 酒店 SPU、房型销售模板 SKU、上下架、展示素材 | 酒店、房型、品牌、城市、商圈、设施 | Rate Plan：含早/无早、可取消/不可取消、支付方式 | 某日期房态房价、税费、供应商确认结果走缓存或实时查询 |
| **Movie** | 影片/影院/套餐商品、基础场次配置 | 影片、影院、影厅、座位区域 | 场次价、套餐价、服务费规则 | 实时座位图、锁座状态、出票结果走供应商查询 |

这种划分的核心标准是：**稳定内容进商品中心，动态资源走缓存/供应商，交易结果进订单/履约/售后快照**。

以酒店为例，酒店本身是 Resource，平台上售卖的不是“酒店这一条记录”，而是围绕酒店资源包装出来的商品和售卖条件：

```text
resource_tab
  HOTEL: Bangkok Central Hotel
  ROOM_TYPE: Deluxe King Room

product_spu_tab
  SPU-HOTEL-90001: 平台上的 Bangkok Central Hotel 商品页

product_sku_tab
  SKU-HOTEL-90001-ROOM90002-BF-RF:
    Deluxe King Room + 含早 + 可取消

rate_plan_tab
  BF_RF:
    breakfast = included
    cancel_policy = free_cancel_before_deadline
    payment_type = prepay

实时查询 / 报价缓存
  check_in = 2026-05-01
  nights = 2
  adult = 2
  price = 实时返回
  availability = 实时确认
```

这个例子说明：`resource_tab` 回答“资源是什么”，`product_spu_tab` 回答“平台是否运营这个资源”，`product_sku_tab` 回答“卖哪个稳定销售模板”，`rate_plan_tab` 回答“用什么售卖条件”，实时查询回答“这个日期和人数下是否真的可以买”。

再以账单缴费为例，账单机构和缴费入口可以沉淀在商品中心，但用户账单金额不能沉淀成商品主数据：

```text
resource_tab
  BILLER: 某电力公司

product_spu_tab
  SPU-BILL-ELECTRICITY: 电费缴费

product_sku_tab
  SKU-BILL-ELECTRICITY-REGION-A: A 地区电费缴费入口

input_schema_tab
  account_no: 必填，数字，长度 10-16

实时查账
  account_no = 用户输入
  bill_amount = 实时返回
  due_date = 实时返回
```

这类商品的可售单元不是固定面额，而是“用户输入账号后形成的一次账单支付上下文”。因此商品中心只保存账单机构、输入规则、手续费规则和履约契约，具体账单金额进入计价上下文和订单快照。

---

#### 16.6.1.6 商品供给与运营链路

商品供给与运营链路解决的是“商品如何进入平台，以及上线后如何被持续维护”的问题。供应商同步本质上属于供给链路，但它不是唯一入口。更准确的划分是：

```text
商品供给与运营治理平台
  ├─ 人工创建/上传：运营/商家从 0 到 1 创建商品
  ├─ 批量导入：文件、模板、批量任务导入商品和配置
  ├─ 运营编辑：标题、图片、类目、价格、库存、上下架、退款规则变更
  └─ 供应商同步：外部供给数据全量/增量/Push/刷新
```

这四类入口应该进入同一个“供给治理控制面”，共享任务模型、校验、审核、发布版本、Outbox、补偿和可观测性；但供应商同步因为存在长任务、Checkpoint、Raw Snapshot、Worker 租约、DLQ、数据新鲜度等复杂问题，可以单独展开成 `16.6.1.7` 和附录案例。

面试时可以先给出这个判断：

> 供应商同步是商品供给链路的一种入口，但不能把商品供给系统等同于供应商同步系统。商品供给平台要统一承接人工创建、批量导入、运营编辑和供应商同步；其中供应商同步是最复杂的自动化供给分支，需要独立的同步任务、Checkpoint 和数据治理链路。

如果这条链路设计不好，问题会很快暴露到 C 端交易链路：列表页搜不到、详情页价格错误、下单时库存不可用、券码发放失败、供应商映射缺失、退款规则不完整。商品供给链路的核心不是“把商品写进数据库”，而是“让一个商品从供给入口到可被搜索、可被下单、可被履约、可被追溯”。

这条链路的系统难点和解决方法可以先这样收敛：

| 难点 | 典型表现 | 解决方法 |
|------|----------|----------|
| **入口多且语义不同** | 人工创建、批量导入、运营编辑、供应商同步都在改商品 | 统一进入 Supply Task 和 Staging，但按 `task_type` 路由不同策略 |
| **半成品污染线上** | 草稿、导入半成品、同步脏数据直接写正式表 | Draft / Staging 与正式表隔离，只有发布事务能写线上版本 |
| **品类差异大** | 酒店、话费、账单、礼品卡、电影票字段和交易规则完全不同 | 类目模板 + 能力矩阵 + Schema 驱动表单、校验和发布规则 |
| **运营误操作风险高** | 批量改价、退款规则变更、类目迁移导致资损或投诉 | 字段级 Diff、风险评分、强审核、版本回滚和灰度发布 |
| **供应商和运营冲突** | 供应商同步覆盖运营修正字段，线上数据反复抖动 | 字段主导权、人工覆盖保护期、冲突日志和巡检 |
| **发布不一致** | 商品库成功，ES、缓存、营销、计价没有刷新 | 发布事务 + Outbox + 异步刷新 + 补偿重试 |
| **失败不可运营** | 错误只在日志里，运营不知道哪一行失败、怎么修 | 行级明细、错误文件、MySQL DLQ、修复建议和重新投递 |
| **历史订单被新配置影响** | 商品改价或退款规则变更后影响旧订单解释 | 创单保存商品快照、报价快照、履约契约和退款规则快照 |

**1. 三种方案对比**

| 方案 | 核心思路 | 优点 | 缺点 | 适用阶段 |
|------|---------|------|------|---------|
| **方案A：后台 CRUD + 简单审核** | 运营直接编辑商品正式表，审核通过后上架 | 实现简单，适合固定 SKU、低规模自营商品 | 无法支撑批量导入、错误隔离、版本回滚、下游一致性和事故追溯 | 早期平台 |
| **方案B：任务化上架系统** | 用 Listing Task 承接人工上传、批量导入和运营编辑 | 支持异步处理、进度追踪、错误文件、失败重试 | 只解决“任务怎么跑”，没有完整的质量治理、风险分级和发布一致性 | 中期平台 |
| **方案C：供给治理平台** | 在任务化基础上加入暂存区、标准化、质量校验、Diff、差异化审核、发布快照、Outbox、补偿和巡检 | 能支撑 OTA、O2O、虚拟商品、多供应商、多运营角色长期演进 | 设计复杂度更高，需要明确模型边界和流程状态机 | 多品类、多来源、强运营平台 |

本系统选择 **方案C：供给治理平台**。它不是把所有入口混成一条大流程，而是提供统一控制面，让不同入口走不同策略、共享同一套发布和治理能力。

**2. 推荐架构：供给治理控制面**

```text
供给入口层
  → Draft / Staging 暂存区
  → Listing Task / Batch / Item
  → 标准化与类目模板适配
  → 多层质量校验
  → Diff 与风险识别
  → 审核流 / 自动准入
  → 发布事务：主数据 + 交易契约 + Outbox
  → 搜索索引 / 缓存 / 营销 / 计价 / 订单上下文刷新
  → DLQ / 补偿 / 巡检 / 质量报表
```

各层职责如下：

| 层级 | 职责 | 关键产物 |
|------|------|----------|
| **供给入口层** | 接收运营表单、批量文件、供应商同步、商家 API | 原始输入、操作者、来源、TraceID |
| **暂存层** | 保存未发布数据，避免污染线上正式表 | Draft、Staging Snapshot、Import Row |
| **任务编排层** | 把一次供给动作变成可恢复任务 | `product_supply_task`、`product_supply_task_item`、进度和失败明细 |
| **标准化层** | 把入口数据转换成平台 Resource/SPU/SKU/Offer/Rule 模型 | 标准化模型、字段来源、数据 hash |
| **校验层** | 检查字段、主数据、交易契约、可售规则 | 校验结果、错误码、质量分 |
| **风险审核层** | 识别高风险变更并路由审核 | Diff、风险等级、审核单 |
| **发布层** | 写正式表、生成版本、写 Outbox | `publish_version`、商品快照、事件 |
| **下游刷新层** | 刷新搜索、缓存、营销、计价、数据平台 | 索引任务、缓存失效任务、补偿任务 |
| **治理层** | 巡检、补偿、报表、审计 | DLQ、质量报告、操作日志 |

这里最重要的边界是：**所有入口都不要直接写商品正式表**。人工表单、Excel 导入、供应商同步和运营批量编辑都先写暂存区和任务表，经过校验、审核和发布后，再写入正式主数据和交易契约表。

**3. 四类入口的差异化设计**

| 入口 | 典型场景 | 主流程 | 关键风险 | 处理策略 |
|------|----------|--------|----------|----------|
| **人工创建** | 运营创建本地生活券、礼品卡、话费套餐、账单缴费入口 | 表单草稿 → 实时校验 → 提交审核 → 发布 | 字段漏填、类目选错、履约规则不完整 | 表单配置化、类目模板、强校验、完整审核 |
| **批量导入** | 大促前批量创建套餐、门店、价格计划、券码池 | 上传文件 → 预校验 → 异步解析 → 分批处理 → 错误文件 | 大批量错误、重复导入、局部失败 | 任务化、行级状态、部分成功、失败文件、幂等 key |
| **运营编辑** | 改标题、图片、价格、库存、退款规则、上下架 | 读取线上版本 → 创建变更单 → Diff → 风险审核 → 发布 | 误操作、批量事故、覆盖供应商数据 | 字段主导权、版本锁、风险阈值、回滚 |
| **供应商同步** | 酒店、影院、活动、票务等外部数据同步 | 同步任务 → Raw Snapshot → 标准化 → 映射 → Diff → 发布 | 接口不稳定、模型不一致、新鲜度、长任务失败 | 独立同步链路，见 `16.6.1.7` |

这四类入口共享最终发布模型，但入口策略不同。人工创建强调“完整性和可解释”；批量导入强调“吞吐和错误隔离”；运营编辑强调“Diff、权限和风险”；供应商同步强调“可恢复、可追溯和自动化治理”。

**4. 核心数据模型**

供给治理平台的表设计不要从“一张商品表”出发，而要围绕“未发布隔离、任务可恢复、行级可定位、校验可解释、变更可审核、发布可追溯、失败可补偿”来组织。核心可以分成八组：

| 表组 | 典型表 | 作用 |
|------|--------|------|
| **Draft 草稿表** | `product_supply_draft`、`product_supply_draft_version` | 保存单商品创建和编辑过程中的草稿，允许反复保存，不影响线上 |
| **Task 任务表** | `product_supply_task` | 记录一次供给动作，如人工创建、批量导入、运营编辑、供应商同步后的商品变更接入 |
| **Task Item 明细表** | `product_supply_task_item` | 记录每一行、每个商品、每个 Offer 或每条规则的处理状态，是失败定位单元 |
| **Staging 暂存表** | `product_supply_staging`、`product_supply_staging_snapshot` | 保存已经提交、已经标准化、但未发布到正式表的数据 |
| **Validation 校验表** | `product_validation_result` | 保存 Schema、类目模板、主数据、商品模型、交易契约、风险规则的校验结果 |
| **Change / Audit 表** | `product_change_request`、`product_audit_log`、`product_field_ownership` | 保存字段 Diff、风险等级、审核策略、审核动作和字段主导权 |
| **Publish / Snapshot 表** | `product_publish_record`、`product_publish_snapshot`、`product_change_log` | 保存发布批次、线上版本快照和正式变更日志，支持追溯和回滚 |
| **Outbox / DLQ / Compensation 表** | `product_outbox_event`、`product_supply_dead_letter`、`product_compensation_task`、`product_quality_issue` | 保证下游最终一致，承接失败补偿、人工修复和质量巡检 |

第一期不一定把所有可选表都建齐，最小闭环建议包括：

```text
product_supply_draft
product_supply_task
product_supply_task_item
product_supply_staging
product_validation_result
product_change_request
product_audit_log
product_publish_snapshot
product_change_log
product_outbox_event
product_supply_dead_letter
```

供应商同步执行层可以独立使用 `supplier_sync_task`、`supplier_sync_batch`、`supplier_sync_snapshot` 和 `supplier_sync_dead_letter`，但标准化之后要进入供给平台的 `product_supply_staging`、`product_validation_result`、`product_change_request` 和统一发布链路。

`product_supply_task` 记录一次供给动作：

```sql
CREATE TABLE product_supply_task (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id VARCHAR(64) NOT NULL,
    task_type VARCHAR(32) NOT NULL COMMENT 'MANUAL_CREATE/BATCH_IMPORT/OPS_EDIT/SUPPLIER_SYNC',
    source_type VARCHAR(32) NOT NULL COMMENT 'OPS/MERCHANT/SUPPLIER/SYSTEM',
    source_id VARCHAR(64) DEFAULT NULL,
    category_code VARCHAR(32) NOT NULL,
    operator_id VARCHAR(64) DEFAULT NULL,
    trigger_id VARCHAR(64) DEFAULT NULL COMMENT '外部幂等 ID',
    status VARCHAR(32) NOT NULL COMMENT 'DRAFT/VALIDATING/REVIEWING/APPROVED/PUBLISHING/PUBLISHED/PARTIAL_FAILED/REJECTED/FAILED/CANCELLED',
    total_count INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    failed_count INT NOT NULL DEFAULT 0,
    skipped_count INT NOT NULL DEFAULT 0,
    current_stage VARCHAR(64) DEFAULT NULL,
    error_file_ref VARCHAR(512) DEFAULT NULL,
    publish_version BIGINT DEFAULT NULL,
    created_at DATETIME NOT NULL,
    started_at DATETIME DEFAULT NULL,
    finished_at DATETIME DEFAULT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_task_id (task_id),
    UNIQUE KEY uk_trigger (task_type, trigger_id),
    KEY idx_status (status),
    KEY idx_category_status (category_code, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品供给任务';
```

`product_supply_task_item` 记录每个商品、资源或 Offer 的处理结果：

```sql
CREATE TABLE product_supply_task_item (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id VARCHAR(64) NOT NULL,
    item_no VARCHAR(64) NOT NULL COMMENT '文件行号或外部对象序号',
    item_type VARCHAR(32) NOT NULL COMMENT 'RESOURCE/SPU/SKU/OFFER/STOCK/RULE',
    idempotency_key VARCHAR(128) NOT NULL,
    platform_resource_id BIGINT DEFAULT NULL,
    spu_id BIGINT DEFAULT NULL,
    sku_id BIGINT DEFAULT NULL,
    offer_id BIGINT DEFAULT NULL,
    status VARCHAR(32) NOT NULL COMMENT 'PENDING/VALIDATING/REVIEWING/PUBLISHING/SUCCESS/FAILED/SKIPPED',
    risk_level VARCHAR(32) DEFAULT NULL COMMENT 'LOW/MEDIUM/HIGH',
    error_code VARCHAR(128) DEFAULT NULL,
    error_message VARCHAR(1024) DEFAULT NULL,
    draft_ref VARCHAR(512) DEFAULT NULL,
    normalized_ref VARCHAR(512) DEFAULT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_task_item (task_id, item_no),
    UNIQUE KEY uk_task_idempotency (task_id, idempotency_key),
    KEY idx_task_status (task_id, status),
    KEY idx_platform_object (platform_resource_id, spu_id, sku_id, offer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品供给任务明细';
```

暂存区保存未发布的数据，不直接影响线上：

```sql
CREATE TABLE product_supply_staging (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    staging_id VARCHAR(64) NOT NULL,
    task_id VARCHAR(64) NOT NULL,
    item_no VARCHAR(64) NOT NULL,
    object_type VARCHAR(32) NOT NULL COMMENT 'RESOURCE/SPU/SKU/OFFER/RATE_PLAN/STOCK/RULE',
    object_key VARCHAR(128) NOT NULL,
    source_type VARCHAR(32) NOT NULL,
    raw_payload_ref VARCHAR(512) DEFAULT NULL,
    normalized_payload JSON NOT NULL,
    payload_hash VARCHAR(64) NOT NULL,
    base_publish_version BIGINT DEFAULT NULL,
    status VARCHAR(32) NOT NULL COMMENT 'DRAFT/VALIDATED/REVIEWING/APPROVED/PUBLISHED/REJECTED',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_staging_id (staging_id),
    UNIQUE KEY uk_task_object (task_id, object_type, object_key),
    KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品供给暂存数据';
```

变更日志保存 Diff、风险和审核依据：

```sql
CREATE TABLE product_change_request (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    change_id VARCHAR(64) NOT NULL,
    task_id VARCHAR(64) NOT NULL,
    object_type VARCHAR(32) NOT NULL,
    object_id BIGINT DEFAULT NULL,
    old_publish_version BIGINT DEFAULT NULL,
    new_staging_id VARCHAR(64) NOT NULL,
    changed_fields JSON NOT NULL,
    risk_level VARCHAR(32) NOT NULL,
    review_policy VARCHAR(32) NOT NULL COMMENT 'AUTO_APPROVE/MANUAL_REVIEW/BLOCK',
    status VARCHAR(32) NOT NULL COMMENT 'PENDING/APPROVED/REJECTED/PUBLISHED',
    reviewer_id VARCHAR(64) DEFAULT NULL,
    review_note VARCHAR(1024) DEFAULT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_change_id (change_id),
    KEY idx_task (task_id),
    KEY idx_status_risk (status, risk_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品供给变更单';
```

**5. 人工创建链路**

人工创建不是简单表单提交。它要把“类目模板、交易契约、履约契约、退款规则”一次性收齐，否则商品看似创建成功，交易时会失败。

```text
选择类目
  → 加载类目模板和能力矩阵
  → 填写 Resource / SPU / SKU / Offer / Rule
  → 前端实时校验 + 后端强校验
  → 保存 Draft
  → 提交 Listing Task
  → 生成 Staging Snapshot
  → 质量校验
  → 新商品审核
  → 发布正式表
```

人工创建的关键设计：

| 设计点 | 说明 |
|--------|------|
| **类目模板驱动表单** | 不同品类展示不同字段，例如酒店要地址和坐标，充值要号码规则，账单缴费要 Input Schema |
| **Draft 与正式表隔离** | 草稿允许反复保存，不影响线上商品 |
| **交易契约强校验** | Offer、库存来源、履约规则、退款规则、Input Schema 不完整时不能提交 |
| **审核证据完整** | 审核员看到的是标准化后的商品快照、字段来源、风险命中和历史版本 |
| **创建后不等于上线** | 发布成功后还要等待库存初始化、索引刷新、可售校验通过 |

面试时可以强调：人工创建链路最容易被低估。真正难点不是页面表单，而是类目差异、交易契约完整性、审核证据和发布一致性。

**6. 批量导入链路**

批量导入要按“任务 + 行级明细 + 暂存快照 + 错误文件”设计，不能把整个 Excel 读进内存后循环写正式表。

```text
下载模板
  → 上传文件
  → 文件格式预检
  → 创建 product_supply_task(status=PENDING)
  → 流式解析文件
  → 每行生成 product_supply_task_item
  → 分批标准化和校验
  → 成功项进入发布/审核
  → 失败项生成错误文件
  → 任务状态汇总为 PUBLISHED / PARTIAL_FAILED / FAILED
```

批量导入的关键设计：

| 设计点 | 说明 |
|--------|------|
| **模板版本化** | 模板字段随类目演进，导入文件必须记录 `template_version` |
| **行级幂等** | 用 `task_id + row_no` 和业务幂等 key 防止重复导入 |
| **部分成功** | 10000 行中 9800 行成功、200 行失败时，不应该整批回滚 |
| **错误文件下载** | 失败行要带 `error_code`、`error_message`、原始值和建议修复方式 |
| **背压与限流** | 大文件分片处理，避免压垮商品库、库存系统和搜索刷新 |
| **批量事故防护** | 高风险批量变更必须二次确认或抽样审核 |

批量异步链路要拆成多个 Worker 阶段，而不是一个 Worker 从解析文件一路写到正式表：

```text
上传文件 / 批量提交
  → 创建 product_supply_task(status=PENDING)
  → Parser Worker 抢占任务并流式解析文件
  → 批量写入 product_supply_task_item
  → Item Worker 分批处理 item
  → 标准化 / 校验 / Staging / Diff
  → 低风险自动发布，高风险进入审核
  → Publish Worker 发布正式表并写 Outbox
  → 生成错误文件 / DLQ / 质量报告
```

**Parser Worker 只负责解析，不负责发布**。它校验文件 hash、模板版本和列结构，按行流式读取文件，每 N 行批量插入 `product_supply_task_item`，并持续更新 `parsed_count` 和 `parse_checkpoint`。如果 Worker 中途退出，下次从 checkpoint 继续；如果重复解析上一小批数据，通过 `task_id + item_no` 和 `task_id + idempotency_key` 唯一键去重。

```json
{
  "sheet": "Sheet1",
  "row_no": 12000,
  "byte_offset": 8842211
}
```

**`product_supply_task_item` 是真正的问题定位单元**。Task 只表示一次批量任务，Item 表示每一行、每一个商品对象或每一个 Offer 的处理状态。

```text
PENDING
  → NORMALIZING
  → VALIDATING
  → STAGING
  → DIFFING
  → REVIEWING
  → PUBLISHING
  → SUCCESS

失败分支：
NORMALIZING / VALIDATING / STAGING / DIFFING / PUBLISHING
  → FAILED / DLQ / SKIPPED
```

Item Worker 不按文件整批处理，而是扫描小批量 item：

```sql
SELECT *
FROM product_supply_task_item
WHERE task_id = ?
  AND status IN ('PENDING', 'FAILED')
  AND next_retry_at <= NOW()
ORDER BY item_no ASC
LIMIT 500;
```

每个 item 或小批次独立事务，流程是：读取原始行 → 按类目模板标准化 → 写 `normalized_ref` → 执行 Schema、主数据、交易契约校验 → 写 `product_supply_staging` → 与线上 `publish_version` 做 Diff → 生成 `product_change_request` → 按风险等级进入自动发布或人工审核。

Publish Worker 只处理已经通过审核或自动准入的变更：

```text
读取 APPROVED change
  → 开启发布事务
  → 写 Resource / SPU / SKU / Offer / Rule
  → 写 publish_snapshot 和 change_log
  → 写 outbox_event
  → 提交事务
  → item.status = SUCCESS
```

Task 状态由 item 统计汇总，而不是 Worker 主观判断：

| Item 汇总结果 | Task 状态 |
|---------------|-----------|
| 全部 `SUCCESS` | `PUBLISHED` |
| 部分 `SUCCESS`，部分 `FAILED/DLQ` | `PARTIAL_FAILED` |
| 全部失败 | `FAILED` |
| 存在 `REVIEWING` | `REVIEWING` |
| 存在 `PUBLISHING` | `PUBLISHING` |

这里的关键原则是：**Parser Worker 只解析，Item Worker 推进行级状态，Publish Worker 只做已审核发布；Task 管整体进度，Item 管失败定位，Staging 管线上隔离，Outbox 管下游一致性**。

错误文件示例：

```text
row_no, sku_code, field, error_code, error_message
12, SKU_001, price, PRICE_TOO_LOW, price is lower than category floor price
25, SKU_014, refund_rule, REFUND_RULE_MISSING, refund rule is required for hotel offer
31, SKU_020, city_code, CITY_NOT_FOUND, city code cannot map to platform city
```

**7. 运营编辑链路**

运营编辑不是“打开商品详情页直接保存”。一个线上商品可能同时被供应商同步、运营编辑、库存系统、风控系统影响。运营编辑必须明确字段主导权、版本锁和风险审核。

```text
读取当前 publish_version
  → 创建编辑草稿
  → 修改字段
  → 与线上版本做 Diff
  → 判断字段主导权和风险等级
  → 自动通过 / 人工审核 / 阻断
  → 发布新 publish_version
  → Outbox 通知搜索、缓存、营销、计价、订单
```

字段主导权可以这样定义：

| 字段 | 主导方 | 供应商同步能否覆盖 | 运营编辑策略 |
|------|--------|-------------------|--------------|
| 酒店名称、地址、设施 | 供应商/平台治理 | 低风险可覆盖，高风险审核 | 可人工修正并加保护期 |
| 标题、卖点、活动标签 | 平台运营 | 不能直接覆盖 | 运营编辑为准 |
| 基础价格、Rate Plan | 供应商/计价 | 取决于品类 | 超阈值审核 |
| 库存水位、可售状态 | 库存域/供应商 | 可覆盖 | 异常告警，不建议人工长期覆盖 |
| 退款规则、履约规则 | 平台/供应商契约 | 高风险覆盖 | 强制审核 |
| 类目、Resource 映射 | 平台治理 | 不能自动覆盖 | 强制审核和巡检 |

高风险运营编辑必须具备三个能力：

1. **Diff 可读**：审核员看到字段级变化，而不是整段 JSON。
2. **版本可回滚**：发布新版本后出现事故，可以回滚到上一个 `publish_version`。
3. **覆盖可解释**：如果运营字段覆盖了供应商字段，要记录覆盖原因、有效期和责任人。

**8. 标准化校验与风险审核**

数字商品供给不能只做字段必填校验，还要校验交易前契约是否完整。

| 校验层 | 检查内容 | 示例 | 失败处理 |
|--------|----------|------|----------|
| **Schema 校验** | 字段类型、必填、枚举、长度、格式 | 图片 URL、手机号规则、账单号长度 | 行级失败 |
| **类目模板校验** | 类目要求的属性、能力、扩展字段是否完整 | Gift Card 必须有面额和有效期 | 阻断提交 |
| **主数据校验** | Resource、Brand、Carrier、城市、商户是否存在 | 酒店必须有关联城市 | 进入人工映射 |
| **商品模型校验** | SPU/SKU/Offer/Rate Plan 关系是否成立 | SKU 不能缺 Offer | 阻断发布 |
| **交易契约校验** | 库存来源、Input Schema、履约规则、退款规则是否完整 | Voucher 券码池为空不能发布 | 阻断发布或告警 |
| **风险校验** | 价格、类目、履约、退款、映射是否高风险 | 价格大幅变化、退款规则变严 | 人工审核 |

审核策略应该差异化，而不是所有变更都人工审核：

| 变更类型 | 风险等级 | 策略 |
|----------|----------|------|
| 标题、描述、普通图片修改 | 低 | 自动通过，记录变更日志 |
| 库存水位、供应商可售状态 | 中 | 自动校验，通过后发布，异常告警 |
| 展示价、Offer 规则、活动标签 | 中高 | 超过阈值进入人工审核 |
| 类目、履约类型、退款规则 | 高 | 强制人工审核 |
| 供应商映射、Resource ID、SPU/SKU 结构 | 高 | 强制审核，并触发巡检 |

风险规则要配置化：

```text
risk_score =
  field_weight
  + change_ratio_weight
  + category_weight
  + operator_risk_weight
  + product_heat_weight
```

例如同样是改价，长尾商品小幅调价可以自动通过，热门酒店或高销量礼品卡大幅降价必须人工复核。

**9. 发布一致性设计**

审核通过不代表商品已经可售。真正发布时，要保证商品主数据、资源映射、交易契约、库存可售、搜索缓存和下游系统最终一致。

```text
审核通过
  → 开启发布事务
  → 写入 Resource / SPU / SKU / Offer / Rate Plan
  → 写入 Stock Config / Sellable Rule
  → 写入 Input Schema / Fulfillment Rule / Refund Rule
  → 生成 publish_version 和 product_snapshot
  → 写入 product_change_log
  → 写入 Outbox 事件
  → 提交事务
  → 异步刷新搜索、缓存、营销、计价、数据平台
```

发布事务里只做商品中心必须强一致的事情；ES 刷新、缓存失效、营销圈品、计价上下文刷新都通过 Outbox 异步执行。

| 设计点 | 说明 |
|--------|------|
| **正式表与暂存表分离** | 任务处理中的半成品不能污染线上 |
| **发布版本化** | 每次发布生成 `publish_version`，支持回滚、对账和排查 |
| **Outbox 同事务** | 商品变更与事件写入同事务，避免“商品变了但下游不知道” |
| **下游刷新可重试** | ES、缓存、营销、计价刷新失败进入补偿任务 |
| **订单只信快照** | 创单保存商品快照、报价快照、履约契约和退款规则快照 |

Outbox 事件示例：

```text
ProductPublished
ProductContentChanged
OfferChanged
SellableRuleChanged
FulfillmentRuleChanged
SearchIndexRefreshRequired
ProductCacheInvalidationRequired
```

**10. DLQ、补偿与质量巡检**

人工供给和运营编辑也需要 DLQ。它们的失败不一定来自供应商接口，更多来自文件格式、字段错误、审核驳回、发布失败和下游刷新失败。

| 失败类型 | 示例 | 处理 |
|----------|------|------|
| **输入失败** | Excel 字段非法、必填缺失 | 生成错误文件，运营修复后重新提交 |
| **映射失败** | 城市、商户、品牌、Resource 找不到 | 进入人工映射队列 |
| **审核失败** | 高风险变更被驳回 | 回到草稿，保留驳回原因 |
| **发布失败** | DB 写入冲突、版本过期 | 重试或要求基于最新版本重新编辑 |
| **下游失败** | ES 刷新失败、缓存失效失败 | Outbox 补偿 |
| **质量失败** | 缺图、缺价、无库存、不可履约 | 质量巡检下架或告警 |

补偿任务包括：

1. 失败行重新投递。
2. 审核通过但发布失败重试。
3. 搜索索引重建。
4. 商品缓存失效重试。
5. 发布版本与 ES 索引一致性校验。
6. 商品质量日报。
7. 运营覆盖字段到期巡检。
8. 无库存、无价格、无履约规则商品巡检。

**11. 可观测性指标**

供给运营链路需要可观测，否则运营会遇到“上传了但不知道失败在哪里”“审核通过但前台搜不到”“商品发布了但不能下单”等问题。

| 指标 | 说明 | 目标 |
|------|------|------|
| **任务成功率** | 成功任务 / 总任务 | 按入口拆分统计 |
| **行级成功率** | 成功 item / 总 item | 批量导入核心指标 |
| **任务完成耗时** | 从创建到发布完成 | P95 可控 |
| **自动审核占比** | 自动通过 / 总审核 | 持续提升，但高风险不追求自动化 |
| **审核驳回率** | 驳回 / 审核提交 | 反映输入质量和规则合理性 |
| **发布失败率** | 发布失败 / 发布任务 | < 1% |
| **索引刷新成功率** | ES 刷新成功 / 总刷新 | > 99% |
| **缓存失效成功率** | 缓存失效成功 / 总失效 | > 99% |
| **商品质量缺陷率** | 缺图、缺价、无库存、映射缺失商品占比 | 持续下降 |
| **人工修复耗时** | 从失败到修复完成 | 按错误类型统计 |

运营后台至少要能看到：

```text
任务进度：总数、成功、失败、跳过、当前阶段
失败原因：错误码、错误字段、建议修复方式、错误文件
审核队列：风险等级、命中规则、Diff、责任人
发布结果：publish_version、Outbox 状态、索引/缓存刷新状态
质量看板：缺图、缺价、无库存、无履约规则、映射缺失
```

**12. 与供应商同步链路的关系**

供应商同步不是被排除在供给链路之外，而是供给链路中自动化程度最高、数据治理要求最强的入口。

```text
统一供给治理平台
  → 统一发布模型：Resource / SPU / SKU / Offer / Rule
  → 统一治理能力：校验 / Diff / 审核 / 发布 / Outbox / 补偿
  → 统一观测能力：任务进度 / 失败明细 / 质量指标

供应商同步专项链路
  → Raw Snapshot
  → Checkpoint
  → Worker Lease
  → Sync Batch Version
  → Supplier Mapping
  → 数据新鲜度
```

所以本章采用“主链路 + 专项链路”的写法：`16.6.1.6` 讲统一商品供给与运营治理平台，完整设计见[附录G：商品供给与运营治理平台](../appendix/product-supply-ops.md)；`16.6.1.7` 专门讲供应商同步，因为它有长任务恢复、外部数据追溯和供应商质量治理等额外复杂度。

**面试总结**：

> 我不会把商品供给设计成后台 CRUD，也不会把供应商同步和人工运营割裂成两套互不相干的系统。我的设计是统一供给治理平台：人工创建、批量导入、运营编辑和供应商同步都先进入任务和暂存区，通过标准化、质量校验、Diff、差异化审核、版本化发布、Outbox、补偿和巡检后，再写入正式商品主数据和交易契约。供应商同步属于供给链路，但因为它涉及 Raw Snapshot、Checkpoint、Worker 租约、DLQ 和数据新鲜度，所以作为专项链路单独展开。这样既能保证入口统一，又能处理不同来源的复杂度差异。

---

#### 16.6.1.7 供应商商品同步链路

供应商同步链路解决的是“外部供给数据如何进入平台，并持续保持可用、可信、足够新鲜”的问题。它不是简单的定时任务，也不是把供应商字段原样搬进商品表，而是一个完整的数据治理链路：接入外部数据、适配不同协议、完成平台模型映射、校验数据质量、生成发布版本、刷新搜索缓存、通知下游系统，并在失败时可追踪、可补偿、可人工修复。

从系统边界上看，它属于 `16.6.1.6` 里的供应商供给入口，但执行层要单独设计。统一供给平台负责发布模型、审核、Outbox 和质量治理；供应商同步专项链路负责外部协议适配、Raw Snapshot、Checkpoint、租约、批次版本、新鲜度和供应商质量治理。

在数字商品平台中，供应商同步的复杂度来自四个方面：

1. **接口不稳定**：供应商可能超时、限流、重复推送、乱序推送，也可能临时修改字段含义。
2. **模型不一致**：供应商有自己的酒店、房型、套餐、面额、场次、票种模型，平台则使用 Resource、SPU、SKU、Offer、Rate Plan 等统一抽象。
3. **新鲜度不同**：酒店地址和设施可以小时级更新，酒店最低价需要分钟级刷新，机票报价和下单前房态房价必须实时确认。
4. **交易风险高**：列表页展示可以允许轻微过期，但创单前如果使用过期价格或库存，就会带来资损、投诉和履约失败。

因此，供应商同步链路的核心目标不是“同步成功”，而是：**正确映射、变化可追溯、错误可隔离、数据可验证、过期可感知、失败可补偿**。

核心难点和解决方法如下：

| 难点 | 典型表现 | 解决方法 |
|------|----------|----------|
| **长任务易中断** | 100 万酒店全量同步跑 10 小时，发布、重启、OOM 都可能中断 | Batch + Page/Cursor Checkpoint + Worker Lease |
| **外部数据不可控** | 字段缺失、枚举变化、分页游标失效、重复 Push | Adapter 防腐层、Schema 校验、幂等 key、指数退避和熔断 |
| **模型映射复杂** | 供应商酒店/房型/套餐无法直接对应平台 Resource/SPU/SKU/Offer | supplier mapping 表、标准化快照、映射失败进入人工修复 |
| **同步成功不等于可发布** | 拉到了数据但城市映射失败、价格异常、坐标漂移 | 质量校验、Diff、风险分级、低风险自动发布，高风险审核或 DLQ |
| **数据新鲜度不一致** | 静态信息小时级即可，房态房价下单前必须实时 | 按数据类型和交易阶段分层 TTL，列表缓存、详情刷新、创单确认 |
| **失败需要可追溯** | 线上价格异常时不知道供应商当时返回什么 | Raw Snapshot / Normalized Snapshot / Diff / publish_version 分离 |
| **下游最终一致** | DB 更新成功但 ES、缓存、营销、计价没有同步 | Outbox 同事务写入，索引和缓存刷新失败进入补偿 |

**供应商同步架构图与 Data Flow Diagram**：

![供应商数据同步链路架构图](../../images/supplier-sync-architecture.png)

![供应商数据同步 Data Flow Diagram](../../images/supplier-sync-data-flow.png)

完整的任务模型、Checkpoint、Worker 租约、DLQ、监控指标和面试题设计，见[附录F：供应商数据同步链路](../appendix/supplier-sync.md)。

图中可以看到，供应商数据进入平台后会经过五个阶段：

```text
供应商数据源
  → 接入适配与同步任务
  → 标准化、质量校验、平台模型映射
  → Resource / SPU / SKU / Offer / Mapping / Stock Snapshot 落库
  → 搜索索引、缓存、营销、计价、订单、数据平台刷新
  → 失败补偿、监控告警、数据巡检
```

**1. 同步对象分层**

供应商同步首先要分清楚“同步的到底是什么”。不同数据的生命周期、新鲜度和交易风险完全不同，不能放在同一张表、使用同一个刷新策略。

| 数据层 | 示例 | 平台承接模型 | 同步特点 |
|-------|------|-------------|---------|
| **资源数据** | 城市、机场、车站、酒店、影院、商户、门店 | `resource_tab` | 相对稳定，适合全量 + 增量同步 |
| **商品主数据** | 标题、图片、类目、属性、可售范围 | `product_spu_tab`、`product_sku_tab` | 变化频率中等，需要审核与发布版本 |
| **销售配置** | 面额、套餐、房价计划、票种、售卖规则 | `product_offer_tab`、`rate_plan_tab` | 直接影响展示价和可售性 |
| **动态交易数据** | 价格、库存、座位图、房态、可售状态 | `product_stock_tab`、缓存、实时查询 | 变化快，需要 TTL 和交易前确认 |
| **供应商映射** | 供应商酒店 ID、房型 ID、套餐 ID、票种 ID | `supplier_product_mapping_tab` | 是履约、查价、查库存的关键桥梁 |

这个分层决定了同步策略：静态资源可以沉淀，动态报价可以缓存，强交易数据必须实时确认。商品中心不能为了统一而把所有数据都持久化成 SKU，也不能为了灵活而完全不沉淀基础资源。

**2. 同步模式设计**

供应商同步通常要同时支持五种模式：

| 同步模式 | 适用场景 | 设计重点 |
|---------|---------|---------|
| **全量同步** | 新供应商接入、数据修复、周期校准 | 分片、断点续跑、批次版本、失败明细 |
| **增量同步** | 日常商品、资源、状态变化 | 游标、更新时间、水位记录、乱序处理 |
| **供应商 Push** | 供应商主动推送价格、库存、上下架变化 | 幂等、签名校验、重复消息去重 |
| **平台主动刷新** | 热门酒店、热门影片、热门面额、活动商品 | 根据曝光、点击、转化、变价率动态调频 |
| **交易前实时确认** | Flight、Hotel、Movie 等强实时品类 | 下单前查价、查库存、锁资源或确认可售 |

实际系统中这五种模式会同时存在。比如酒店静态信息来自全量和增量同步，列表页最低价来自定时刷新，详情页房态房价来自短 TTL 缓存，下单前必须实时向供应商确认。

**3. 幂等设计**

供应商同步的幂等要覆盖三层：

| 层级 | 幂等对象 | 幂等 Key | 目的 |
|------|---------|---------|------|
| **接入层** | 一次供应商 Push 或同步消息 | `supplier_id + event_id` 或 payload hash | 防止重复消费 |
| **映射层** | 一个外部资源或商品 | `supplier_id + supplier_resource_code + supplier_product_code` | 防止重复创建 Resource/SPU/SKU |
| **发布层** | 一次平台商品变更 | `sync_batch_id + platform_product_id + data_hash` | 防止重复发布、重复刷新索引 |

其中 `supplier_resource_code` 表示供应商侧稳定资源，例如酒店 ID、影院 ID、商户 ID、机场/车站代码；`supplier_product_code` 表示供应商侧可售对象，例如房型、套餐、面额、场次、票种。

```text
供应商原始数据
  → 生成 source_hash
  → 查询 supplier_mapping
  → 已存在：比较 data_hash，变化才更新
  → 不存在：创建平台 Resource / SPU / SKU / Offer
  → 写入 mapping，保证后续同步可定位
```

这里最容易踩坑的是供应商编码不稳定。有些供应商会复用商品编码、合并资源、拆分资源，甚至换供应商后编码体系完全变化。因此平台不能直接把供应商编码当成平台主键，而要维护独立的 `platform_resource_id`、`spu_id`、`sku_id`，供应商编码只作为映射关系存在。

**4. 版本设计**

版本要分清楚三类，不能混在一起：

| 版本 | 含义 | 用途 |
|------|------|------|
| `sync_batch_version` | 本次同步任务版本 | 排查“哪次同步带来了变化” |
| `data_snapshot_version` | 原始数据和标准化数据快照版本 | 支持回放、diff、回滚 |
| `publish_version` | 平台正式发布版本 | 控制搜索、缓存、下游事件一致性 |

推荐链路如下：

```text
Sync Batch v102
  → Raw Snapshot v102.1
  → Normalized Snapshot v102.1
  → Diff: price changed / room name changed / offer disabled
  → Publish Version p5688
  → ProductUpdated / OfferChanged / StockChanged Event
```

版本设计的关键是：**同步版本不等于发布版本**。供应商同步可能只是拉到了数据，但经过校验后发现字段缺失，不应该发布；也可能一次同步中有 10 万条数据，只有 300 条真正变化。平台需要把“同步到了什么”和“发布了什么”分开记录。

**5. 质量校验设计**

质量校验不能只做字段非空，而要分成五层：

| 校验层 | 校验内容 | 失败处理 |
|-------|---------|---------|
| **Schema 校验** | 必填字段、类型、枚举、时间格式、货币单位 | 直接拦截，进入失败明细 |
| **主数据校验** | 城市、机场、酒店、商户、品牌、类目是否存在 | 进入待映射或人工修复 |
| **模型校验** | 是否能映射到 Resource/SPU/SKU/Offer/Rate Plan | 阻断发布 |
| **交易校验** | 价格是否异常、库存是否为负、可售状态是否矛盾 | 高风险拦截或降级 |
| **业务规则校验** | 是否允许该站点、渠道、品类售卖，是否需要审核 | 进入审核或灰度发布 |

质量校验要支持“部分成功”。例如酒店全量同步 100 万条房型数据，不能因为 100 条数据失败就整批失败。更合理的处理方式是：可处理数据继续写入，失败明细单独记录 `error_code`、`error_message`、`raw_payload_ref`，高风险数据不发布，进入人工修复或补偿队列。

面试时可以强调：**供应商同步系统不是追求 100% 同步成功，而是要做到失败可定位、可隔离、可修复、可补偿**。

**6. 新鲜度设计**

不同数据的 TTL 不一样，不能使用统一缓存时间。

| 数据类型 | 示例 | 新鲜度要求 | 策略 |
|---------|------|-----------|------|
| **静态资源** | 酒店名称、地址、设施、机场、车站 | 小时级或天级 | 全量 + 增量同步 |
| **半动态数据** | 酒店最低价、可售状态、热门库存水位 | 分钟级 | 定时刷新 + 热门加频 |
| **强动态数据** | 机票报价、座位图、下单前房态房价 | 秒级或实时 | 搜索缓存，详情刷新，下单实时确认 |
| **交易契约** | 退款规则、履约参数、供应商映射 | 强一致倾向 | 发布版本控制，不随意覆盖 |

新鲜度可以按三个维度决策：

```text
TTL = f(category, popularity, transaction_stage)
```

示例：

| 场景 | TTL 策略 |
|------|---------|
| Hotel 列表页最低价 | 热门酒店 10 分钟刷新，长尾酒店 1-6 小时 |
| Hotel 详情页房态房价 | 用户进入详情页时刷新或短 TTL 缓存 |
| Hotel 下单前确认 | 必须实时查供应商 |
| Flight 搜索报价 | 实时查或极短 TTL |
| Topup 面额配置 | 小时级缓存即可 |
| Bill 账单金额 | 用户输入账单号后实时查询 |

这里的原则是：**L 页可以快，D 页要准，创单必须安全**。列表页价格允许作为导购参考，详情页价格要尽量接近实时，创单价格必须基于最新供应商状态确认。

**7. 补偿设计**

补偿不是“失败后重试三次”这么简单，而要按失败类型分类处理。

| 失败类型 | 示例 | 处理方式 |
|---------|------|---------|
| **临时失败** | 网络超时、供应商 5xx、限流 | 指数退避重试 |
| **数据失败** | 字段缺失、枚举非法、价格异常 | 不盲目重试，进入修复队列 |
| **映射失败** | 找不到城市、酒店、影院、商户映射 | 进入人工映射或规则匹配 |
| **发布失败** | DB 成功但 ES 刷新失败 | Outbox 重试，索引补偿 |
| **一致性失败** | 平台数据与供应商数据长期不一致 | 对账任务 + 差异修复 |

推荐处理链路：

```text
Sync Failed
  → 判断错误类型
  → Retryable：延迟重试
  → NonRetryable：进入失败明细
  → MappingRequired：进入人工修复队列
  → PublishFailed：Outbox 补偿
  → StaleData：巡检任务重新拉取
```

死信队列中不要只存错误信息，要存完整上下文：

```text
supplier_id
sync_batch_id
supplier_resource_code
supplier_product_code
error_code
error_message
raw_payload_ref
retry_count
next_retry_time
owner_team
```

否则线上排查时只能看到“同步失败”，不知道失败的是哪个供应商、哪个资源、哪条商品。

**8. 死信队列落地设计**

死信队列（Dead Letter Queue, DLQ）不要只理解成“失败消息丢到一个 MQ Topic”。在供应商商品同步场景里，失败往往不是单纯的消息消费失败，而是字段缺失、映射失败、价格异常、发布失败、索引刷新失败等需要人工修复、状态流转和审计的问题。因此，推荐设计成 **MySQL 为主的可运营 DLQ + MQ/Redis 做调度辅助**。

| 组件 | 职责 |
|------|------|
| **MySQL DLQ 表** | 权威问题单，支持查询、筛选、人工修复、状态流转、审计和报表 |
| **Kafka / MQ** | 可选，用于失败事件的短期缓冲和异步投递 |
| **Redis ZSet** | 可选，用于延迟重试调度，按 `next_retry_at` 排序 |
| **Raw Snapshot / 对象存储** | 保存大体积原始 payload，DLQ 表只保存引用 |

判断一条失败是否进入 DLQ，需要先做错误分类：

| 失败类型 | 是否进入 DLQ | 处理方式 |
|---------|-------------|---------|
| 网络超时、供应商 5xx | 不一定 | 先自动重试，超过次数后进入 DLQ |
| 供应商限流 | 不一定 | 延迟重试、降速、熔断 |
| 字段缺失、枚举非法 | 是 | 需要人工或规则修复 |
| 城市、酒店、影院、商户映射失败 | 是 | 需要补映射 |
| 价格异常、库存异常 | 是 | 高风险数据拦截 |
| DB 写入成功但 ES 刷新失败 | 是 | 索引补偿 |
| 同步成功但发布失败 | 是 | 发布补偿 |
| 重复消息 | 否 | 幂等丢弃即可 |

推荐处理架构如下：

```text
同步任务失败
  → 错误分类
  → Retryable：进入延迟重试队列
  → 达到最大重试次数：写入 MySQL DLQ
  → NonRetryable：直接写入 MySQL DLQ
  → 人工修复 / 规则修复 / 定时补偿
  → 重新投递同步任务
  → 成功后标记 RESOLVED
```

DLQ 主表可以这样设计：

```sql
CREATE TABLE supplier_sync_dead_letter (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,

    -- 定位同步批次
    sync_batch_id VARCHAR(64) NOT NULL,
    sync_task_id VARCHAR(64) NOT NULL,
    sync_mode VARCHAR(32) NOT NULL COMMENT 'FULL/INCREMENTAL/PUSH/REFRESH',
    category_code VARCHAR(32) NOT NULL,

    -- 定位供应商和外部对象
    supplier_id BIGINT NOT NULL,
    supplier_resource_code VARCHAR(128) DEFAULT NULL,
    supplier_product_code VARCHAR(128) DEFAULT NULL,

    -- 平台侧映射，可为空，因为很多失败发生在映射前
    platform_resource_id BIGINT DEFAULT NULL,
    spu_id BIGINT DEFAULT NULL,
    sku_id BIGINT DEFAULT NULL,
    offer_id BIGINT DEFAULT NULL,

    -- 错误分类
    error_stage VARCHAR(64) NOT NULL COMMENT 'ADAPTER/VALIDATION/MAPPING/PUBLISH/INDEX',
    error_type VARCHAR(64) NOT NULL COMMENT 'RETRYABLE/NON_RETRYABLE/MAPPING_REQUIRED/RISK_BLOCKED',
    error_code VARCHAR(128) NOT NULL,
    error_message VARCHAR(1024) NOT NULL,

    -- Payload 不建议大字段直接塞满主表
    raw_payload_ref VARCHAR(512) DEFAULT NULL,
    raw_payload_hash VARCHAR(64) DEFAULT NULL,
    normalized_payload_ref VARCHAR(512) DEFAULT NULL,

    -- 重试与状态
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING'
        COMMENT 'PENDING/RETRYING/MANUAL_FIX/RESOLVED/IGNORED/FAILED',
    retry_count INT NOT NULL DEFAULT 0,
    max_retry_count INT NOT NULL DEFAULT 5,
    next_retry_at DATETIME DEFAULT NULL,
    last_retry_at DATETIME DEFAULT NULL,

    -- 人工处理
    owner_team VARCHAR(64) DEFAULT NULL,
    assignee VARCHAR(64) DEFAULT NULL,
    fix_note VARCHAR(1024) DEFAULT NULL,

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    resolved_at DATETIME DEFAULT NULL,

    UNIQUE KEY uk_dedup (
        sync_batch_id,
        supplier_id,
        supplier_resource_code,
        supplier_product_code,
        error_stage,
        raw_payload_hash
    ),
    KEY idx_status_next_retry (status, next_retry_at),
    KEY idx_supplier_status (supplier_id, status),
    KEY idx_category_status (category_code, status),
    KEY idx_task (sync_task_id),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商同步死信队列';
```

这个表有几个关键设计点：

1. **定位外部对象**：`supplier_resource_code` 和 `supplier_product_code` 用来定位供应商侧资源和可售对象。
2. **平台映射允许为空**：很多失败发生在映射前，所以 `platform_resource_id`、`sku_id`、`offer_id` 都不能强制非空。
3. **Payload 存引用**：供应商原始数据可能很大，尤其是酒店图片、设施、电影座位图，不建议全部放在 DLQ 主表。
4. **唯一键去重**：`uk_dedup` 防止同一条错误反复写入 DLQ。
5. **补偿扫描索引**：`idx_status_next_retry` 支持补偿 Job 按状态和下次重试时间扫描。

如果不想引入对象存储，也可以把 payload 放到单独快照表：

```sql
CREATE TABLE supplier_sync_payload_snapshot (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    payload_ref VARCHAR(128) NOT NULL,
    payload_type VARCHAR(32) NOT NULL COMMENT 'RAW/NORMALIZED',
    payload_json JSON NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_payload_ref (payload_ref)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商同步载荷快照';
```

DLQ 状态机建议保持简单：

```text
PENDING
  → RETRYING
  → RESOLVED

PENDING
  → MANUAL_FIX
  → RETRYING
  → RESOLVED

PENDING
  → IGNORED

RETRYING
  → FAILED
```

| 状态 | 含义 |
|------|------|
| `PENDING` | 等待系统或人工处理 |
| `RETRYING` | 正在补偿重试 |
| `MANUAL_FIX` | 需要人工补映射、修字段、确认风险 |
| `RESOLVED` | 已修复成功 |
| `IGNORED` | 确认无需处理，例如供应商下架或数据已废弃 |
| `FAILED` | 多次补偿仍失败，需要升级 |

补偿 Job 可以按 `status` 和 `next_retry_at` 扫描：

```sql
SELECT *
FROM supplier_sync_dead_letter
WHERE status IN ('PENDING', 'FAILED')
  AND next_retry_at <= NOW()
  AND retry_count < max_retry_count
ORDER BY next_retry_at ASC
LIMIT 100;
```

处理时要按错误类型走不同分支：

```go
func ProcessDeadLetter(ctx context.Context, dlq *DeadLetter) error {
    if !tryLock(dlq.ID) {
        return nil
    }

    switch dlq.ErrorType {
    case "RETRYABLE":
        return retryOriginalSync(ctx, dlq)
    case "MAPPING_REQUIRED":
        if !mappingFixed(dlq) {
            markManualFix(dlq)
            return nil
        }
        return retryOriginalSync(ctx, dlq)
    case "RISK_BLOCKED":
        return waitManualApproval(ctx, dlq)
    case "PUBLISH_FAILED":
        return retryPublish(ctx, dlq)
    default:
        markManualFix(dlq)
        return nil
    }
}
```

重试时间建议使用指数退避，避免供应商故障时补偿任务反复打爆外部接口：

```text
next_retry_at = now + min(2^retry_count minutes, 1 hour)
```

为什么不只用 Kafka DLQ？因为 Kafka 更适合保留失败消息，不适合作为运营治理主存储。

| 能力 | Kafka DLQ | MySQL DLQ |
|------|-----------|-----------|
| 保留失败消息 | 好 | 好 |
| 按供应商、品类、错误码查询 | 弱 | 强 |
| 人工修复状态流转 | 弱 | 强 |
| 审计和报表 | 弱 | 强 |
| 定时补偿扫描 | 一般 | 强 |
| 高吞吐消息暂存 | 强 | 一般 |

所以更推荐采用：

```text
Kafka DLQ：短期消息缓冲，可选
MySQL DLQ：权威问题单和补偿状态
```

面试时可以这样总结：

> 我会把供应商同步 DLQ 设计成 MySQL 主存储，而不是只依赖 Kafka 死信 Topic。因为同步失败往往不是单纯消息消费失败，而是字段缺失、映射失败、价格异常、发布失败这些需要人工修复、状态流转和审计的问题。Kafka 可以作为失败消息的缓冲层，但真正的死信治理要落 MySQL，记录供应商、外部编码、平台映射、错误阶段、payload 引用、重试次数、下次重试时间和处理状态。这样问题才能被查询、补偿、统计和运营化处理。

**9. 监控设计**

供应商同步监控要分成技术指标、数据质量指标和业务影响指标。

| 指标类型 | 指标 | 说明 |
|---------|------|------|
| **技术指标** | 同步成功率、失败率、平均耗时、P99 耗时、重试次数 | 看任务是否健康 |
| **数据质量** | 字段缺失率、映射失败率、重复数据率、异常价格率 | 看数据是否可信 |
| **新鲜度** | 数据延迟、过期数据比例、热门商品刷新延迟 | 看数据是否足够新 |
| **交易影响** | L-D 变价率、D-B 不可售率、下单前确认失败率 | 看同步对转化和交易的影响 |
| **供应商维度** | 每个供应商成功率、超时率、字段错误率 | 支持供应商治理 |
| **品类维度** | 每个品类同步量、失败率、变价率 | 支持品类策略优化 |

核心指标可以这样定义：

```text
同步成功率 = 成功处理 item 数 / 总 item 数
映射失败率 = 映射失败 item 数 / 总 item 数
字段缺失率 = 缺失关键字段 item 数 / 总 item 数
数据新鲜度延迟 = now - last_success_sync_time
L-D 变价率 = 详情页价格 != 列表页价格 的访问占比
D-B 不可售率 = 下单前确认不可售 / 详情页可售点击
```

这些指标不仅用于技术告警，也应该反馈到运营和供应商治理。比如某个供应商字段缺失率长期高，说明不是偶发故障，而是供应商数据质量问题；某个品类 L-D 变价率长期高，说明列表页缓存刷新策略需要调整；某个热门酒店 D-B 不可售率高，说明详情页房态刷新或下单前确认策略存在问题。

**10. 不同品类的同步策略**

| 品类 | 平台沉淀什么 | 实时获取什么 | 同步重点 |
|------|-------------|-------------|---------|
| **Flight / Train / Bus** | 城市、机场、车站、航司/车司、基础线路 | 报价、余票、座位、退改规则确认 | 少沉淀 SKU，搜索和下单前强实时 |
| **Hotel** | 酒店、房型、设施、地理位置、图片、品牌 | 房态、房价、取消规则最终确认 | 静态资源沉淀，动态价格库存按热度刷新 |
| **Topup** | 运营商、国家/地区、面额、套餐、号码规则 | 账号可用性、供应商可用性 | 商品配置稳定，重点是账号校验和供应商状态 |
| **Bill** | 账单机构、账单类型、输入字段、支付规则 | 账单金额、欠费状态、是否可缴 | 低代码表单 + 实时查账单 |
| **Movie / Event** | 影片、影院、活动、场次、票种、套餐 | 座位图、最终票态、锁座结果 | 半同步半实时，座位相关必须实时确认 |
| **Voucher / Gift Card** | 商户、品牌、面额、有效期、核销规则 | 本地券码池库存或供应商券码状态 | 更偏平台自营库存，重点是券码池和核销状态 |

**面试总结**：

> 我不会把供应商同步理解成“写几个定时任务拉数据”。在 OTA、O2O 和虚拟商品平台里，供应商同步本质上是一套外部供给数据治理体系。它要解决幂等映射、版本回溯、质量校验、新鲜度控制、失败补偿和监控治理。列表页可以使用缓存提升性能，详情页要更接近实时，创单前必须实时确认。这样才能在供应商接口不稳定、商品模型不一致、价格库存频繁变化的情况下，仍然保证平台商品数据可信、交易链路安全、问题可追踪。

---

#### 16.6.1.8 库存与可售设计

在本平台中库存没有独立拆服务，而是由商品中心内部的库存与可售域承接。这里的关键不是把库存字段放进商品主表，而是建立统一的库存抽象，屏蔽不同品类的库存来源差异。

| 库存类型 | 典型品类 | 库存来源 | 处理方式 |
|---------|---------|---------|---------|
| **无限库存** | Topup、Bill | 无明确库存或供应商容量足够 | 只做可售规则与供应商可用性校验 |
| **池化库存** | Voucher、Gift Card | 平台券码池或本地库存 | 支付后分配券码，库存不足时停止售卖 |
| **实时库存** | Flight、Hotel、Movie | 供应商接入层 / 外部供应商实时查询 | 搜索展示可缓存，下单前必须实时确认 |

**统一可售判断**：

```text
可售 = 商品状态可售
    + 类目/站点/渠道可售
    + 库存满足
    + 供应商可用
    + 风控/业务规则允许
```

库存相关动作包括查询、预占、释放、扣减、回补和对账。对于 Flight/Hotel/Movie 这类实时库存，商品中心更多是统一入口和状态判断，真实资源确认仍然要通过供应商网关或供应商接入层完成。

---

#### 16.6.1.9 搜索与导购设计

由于搜索也由商品中心负责，商品中心不仅要管理商品数据，还要负责把商品组织成用户可浏览、可搜索、可筛选、可点击的前台体验。

**搜索导购链路**：

```text
首页入口/类目导航
  → 列表页搜索与筛选
  → ES 召回
  → 排序与过滤
  → Hydrate 商品信息、库存、展示价、营销标签
  → 返回列表页
  → 详情页聚合
```

**关键设计点**：

| 能力 | 设计重点 |
|------|---------|
| 首页入口 | 入口配置、类目分组、排序、发布快照、CDN/Redis 缓存 |
| ES 索引 | 商品主数据、类目、实体、Tag、可售状态、运营排序字段 |
| Hydrate | 搜索只返回候选 ID，详情信息、库存、价格、营销标签统一补齐 |
| 缓存 | 热门商品、本地缓存、Redis、索引快照结合使用 |
| 降级 | 价格不可用时展示基础价，营销不可用时隐藏标签，库存不可用时弱提示 |
| 新鲜度 | L 页允许缓存，D 页更接近实时，创单必须实时校验 |

面试时可以这样总结：**搜索导购不是单纯 ES 查询，而是“召回 + 排序 + 商品补齐 + 库存价格营销融合 + 降级”的完整读链路**。

---

#### 16.6.1.10 跨系统集成与事件设计

商品中心位于交易前链路，需要向多个系统输出稳定契约。

| 下游系统 | 商品中心输出 | 对方使用方式 |
|---------|-------------|-------------|
| **营销系统** | 类目、Tag、商品范围、业务实体、可营销状态 | 圈品、活动配置、券可用范围 |
| **计价中心** | 基础价、类目、属性、库存上下文、能力配置 | PDP 价格、结算试算价、下单价、结算价 |
| **订单系统** | 商品快照、上下架状态、可售校验、库存预占结果 | 创建订单前校验与订单快照落库 |
| **履约系统** | 履约类型、供应商映射、履约参数 | 出票、充值、发券、预订确认 |
| **供应商网关/供应商接入层** | 平台 SKU、供应商映射、同步任务上下文 | 查价、查库存、同步、履约调用 |
| **数据平台** | CDC、商品变更日志、质量监控数据 | 经营分析、质量报表、异常发现 |

**核心事件**：

| 事件 | 触发时机 | 典型消费者 |
|------|---------|------------|
| `ProductCreated` | 商品创建成功 | 搜索索引、营销、数据平台 |
| `ProductUpdated` | 商品字段变化 | 搜索索引、缓存刷新、质量监控 |
| `ProductOnShelf` | 商品上架 | 搜索、营销、推荐 |
| `ProductOffShelf` | 商品下架 | 搜索、订单前校验、运营看板 |
| `StockChanged` | 库存变化 | 搜索导购、告警、数据平台 |
| `SupplierSyncFailed` | 供应商同步失败 | 告警、运营后台、任务补偿 |

事件发布建议使用 Outbox 或可靠消息机制，避免“商品已更新但事件丢失”导致搜索、营销、计价数据不一致。

---

#### 16.6.1.11 架构取舍与经验总结

**1. 为什么没有独立拆库存中心和搜索中心？**

这是团队规模和演进阶段下的取舍。库存和搜索都与商品强相关，早期拆成独立服务会增加团队协作、接口维护和数据一致性成本。把它们放在商品中心内部，可以减少跨服务调用，提高迭代效率。但内部必须按域隔离，避免主数据、库存状态和搜索索引混成一团。

**2. 为什么不用一张大宽表？**

大宽表短期开发快，但会很快变成 80+ 字段、15+ 品类混杂、字段语义不清的“大泥球”。更好的做法是：稳定字段进主表，可检索字段进属性表，展示型差异进 ExtInfo，高频品类能力进入扩展表。

**3. ExtInfo JSON、EAV、扩展表怎么取舍？**

| 方案 | 适合场景 | 不适合场景 |
|------|---------|-----------|
| **ExtInfo JSON** | 低频展示字段、品类专属配置 | 高频查询、筛选、排序 |
| **EAV 属性表** | 可搜索、可筛选、可分析属性 | 强事务、高频更新字段 |
| **扩展表** | 高频访问的品类核心字段 | 低频字段和一次性配置 |

**4. 为什么 Flight 不沉淀完整 SKU？**

机票报价由日期、航线、航班、舱位、乘客类型、供应商策略共同决定，价格和库存变化太快。如果把每次报价都沉淀成 SKU，会产生巨量临时 SKU，且数据很快过期。更合理的是商品中心维护城市、机场、航司等基础资源，搜索和创单实时请求供应商。

**5. 为什么 Hotel 要静态信息和动态房态房价分离？**

酒店名称、地址、设施、图片相对稳定，适合同步到商品中心并进入搜索索引；房态和房价按日期、间夜、人数实时变化，适合缓存刷新和下单前实时确认。两者分离可以兼顾搜索性能和交易准确性。

**6. 如何避免商品中心变成“大泥球”？**

关键是三条边界：内部按六个域拆分，数据库按主表/属性/扩展/映射/日志拆分，对外用 API 和事件契约隔离。即便物理上是一个商品中心，逻辑上也要保持清晰边界，为未来拆分成商品中台、库存中心、搜索中心留下演进空间。

---

#### 16.6.1.12 实现落地参考：DDD 分层与代码组织

前面 16.6.1.2 到 16.6.1.11 讨论的是商品中心的业务架构和数据架构。本节保留一个简化版 DDD 代码落地参考，用于说明 Product 聚合根、Repository、接口层、事件订阅和缓存如何组织。它不是上述完整商品中心的全量实现，而是帮助读者理解“架构如何落到代码”的最小示例。

##### 16.6.1.12.1 八层模型的工程落地方式

八层商品交易模型不建议直接落成八个服务或八组强耦合表。更合理的落地方式是：**用八层模型识别品类差异，用品类能力矩阵沉淀差异，用 Runtime Context 输出交易前上下文，用 Category Strategy 执行动态差异，用供应商适配器隔离外部接口差异**。

```text
八层商品交易模型
  → Category Capability Matrix  品类能力矩阵
  → Product Master Model        商品主模型
  → ProductRuntimeContext       交易前运行时上下文
  → CategoryStrategy            品类策略
  → Supplier Adapter / ACL      供应商适配器/防腐层
```

**1. Category Capability Matrix：把八层模型变成品类配置**

能力矩阵描述每个类目在八层模型上的行为。新增品类时，先补齐能力矩阵，再判断是否需要新增策略代码。

```text
category_capability
├─ category_id
├─ product_model_type       // SINGLE_SKU / RESOURCE_BASED / REALTIME_OFFER / ACCOUNT_BASED
├─ resource_type            // NONE / HOTEL / FLIGHT / MOVIE / MERCHANT / BILLER
├─ offer_type               // FIXED_PRICE / RATE_PLAN / REALTIME_QUOTE / BILL_QUERY
├─ availability_type        // UNLIMITED / LOCAL_POOL / REALTIME_SUPPLIER / SEATMAP
├─ input_schema_id
├─ booking_mode             // NONE / PRE_LOCK / PAY_THEN_LOCK / CONFIRM_AFTER_PAY
├─ fulfillment_type         // TOPUP / BILL_PAY / ISSUE_CODE / TICKET / BOOKING_CONFIRM
├─ refund_rule_id
└─ supplier_dependency      // LOW / MEDIUM / HIGH
```

| 品类 | 商品模型 | 报价 | 可用性 | 输入 | 锁定 | 履约 |
|------|---------|------|--------|------|------|------|
| Topup | `SINGLE_SKU` | `FIXED_PRICE` | `SUPPLIER_CHANNEL` | 手机号 | `NONE` | 充值 |
| Bill | `ACCOUNT_BASED` | `BILL_QUERY` | `BILL_PAYABLE` | 账单号 | `LOCK_AMOUNT` | 销账 |
| Gift Card | `SINGLE_SKU` | `FIXED_PRICE` | `LOCAL_POOL` | 邮箱/账号 | `LOCK_CODE` | 发码 |
| Hotel | `RESOURCE_BASED` | `RATE_PLAN` | `REALTIME_SUPPLIER` | 入住人/日期 | `CONFIRM_BOOKING` | 预订确认 |
| Flight | `REALTIME_OFFER` | `REALTIME_QUOTE` | `REALTIME_SUPPLIER` | 乘客证件 | `PRE_LOCK` | 出票 |

**2. ProductRuntimeContext：给交易前链路统一输出**

搜索、详情、结算、创单都需要商品信息，但需要的深度不同。因此商品中心可以输出统一的运行时上下文，再由调用方按场景读取需要的部分。

```go
type ProductRuntimeContext struct {
    ProductDefinition   ProductDefinition
    ResourceContext     ResourceContext
    OfferContext        OfferContext
    Availability        AvailabilityContext
    InputSchema         InputSchema
    BookingRequirement  BookingRequirement
    FulfillmentContract FulfillmentContract
    RefundRule          RefundRule
}
```

| 场景 | 需要的上下文 |
|------|-------------|
| 首页 | `ProductDefinition` |
| 列表页 | `ProductDefinition + Offer 展示价 + Availability 弱状态` |
| 详情页 | `Product + Resource + Offer + Availability + RefundRule` |
| 结算页 | `Product + Offer + Availability + InputSchema` |
| 创单 | `Product + 实时 Offer + 实时 Availability + Booking + Fulfillment + RefundRule` |

**3. CategoryStrategy：让主流程不感知品类差异**

主流程不应该写大量 `if category == flight`、`if category == hotel`。品类差异应该进入策略接口。

```go
type CategoryStrategy interface {
    BuildProductContext(ctx context.Context, req *RuntimeRequest) (*ProductDefinition, error)
    ResolveOffer(ctx context.Context, req *RuntimeRequest) (*OfferContext, error)
    CheckAvailability(ctx context.Context, req *RuntimeRequest) (*AvailabilityContext, error)
    ValidateInput(ctx context.Context, req *RuntimeRequest) error
    PrepareBooking(ctx context.Context, req *RuntimeRequest) (*BookingRequirement, error)
    BuildFulfillmentContract(ctx context.Context, req *RuntimeRequest) (*FulfillmentContract, error)
    BuildRefundRule(ctx context.Context, req *RuntimeRequest) (*RefundRule, error)
}
```

不同品类实现不同策略：

```text
TopupStrategy
BillStrategy
GiftCardStrategy
LocalServiceStrategy
HotelStrategy
FlightStrategy
MovieStrategy
```

**4. Supplier Adapter / ACL：隔离供应商接口差异**

供应商请求参数、响应字段、错误码和超时策略都不一致，不能让这些差异污染商品中心主模型。供应商适配器负责请求转换、响应转换、错误码统一、超时重试、熔断、幂等键和 TraceID 透传。

```text
平台统一请求
  → Supplier Adapter
  → 供应商 A/B/C 私有协议
  → Supplier Adapter 归一化响应
  → OfferContext / AvailabilityContext
```

**5. 创单场景下的完整运行流程**

```text
1. 订单系统请求商品中心构造 RuntimeContext
2. 商品中心读取 category_capability
3. 根据 category_id 找到对应 CategoryStrategy
4. Strategy 读取商品主数据、资源、输入配置、履约配置
5. Strategy 调供应商适配器获取实时 Offer / Availability
6. Strategy 校验用户输入
7. 如果需要 Booking，则执行预订、占座或锁定
8. 商品中心返回 ProductRuntimeContext
9. 订单系统保存商品快照、报价快照、履约契约、售后规则快照
10. 支付成功后，履约系统按 FulfillmentContract 执行交付
```

这套落地方式的关键是：**商品中心保存能力与规则，RuntimeContext 输出交易前上下文，订单保存交易快照，履约系统执行交付，售后系统执行退款规则**。

**领域模型设计思想**：商品域的特点是"树形结构+读多写少"，与订单域的"复杂状态机+高并发写"完全不同。

##### 16.6.1.12.2 Product 聚合根示例

```go
// Product聚合根（SKU维度）
type Product struct {
    // 聚合根ID
    skuID SKU_ID  // 值对象
    
    // SPU信息（实体引用）
    spu *SPU
    
    // SKU规格（值对象）
    specs Specifications
    
    // 基础价格（值对象）
    basePrice Price
    
    // 状态（值对象）
    status ProductStatus
    
    // 多媒体素材
    images []ImageURL
    
    // 时间戳
    createdAt time.Time
    updatedAt time.Time
    
    // 领域事件（未提交）
    domainEvents []DomainEvent
}

// 值对象：SKU_ID
type SKU_ID struct {
    value int64
}

func NewSKU_ID(id int64) SKU_ID {
    return SKU_ID{value: id}
}

func (id SKU_ID) Int64() int64 {
    return id.value
}

// 值对象：Price（基础价格，单位：分）
type Price struct {
    amount int64  // 分为单位
}

func NewPrice(amount int64) (Price, error) {
    if amount < 0 {
        return Price{}, errors.New("价格不能为负数")
    }
    if amount > 100000000 { // 100万元上限
        return Price{}, errors.New("价格超过上限")
    }
    return Price{amount: amount}, nil
}

func (p Price) Amount() int64 {
    return p.amount
}

func (p Price) Yuan() float64 {
    return float64(p.amount) / 100.0
}

// 值对象：Specifications（SKU规格）
type Specifications struct {
    attributes map[string]string  // {"颜色":"红色","尺寸":"L"}
}

func NewSpecifications(attrs map[string]string) Specifications {
    return Specifications{attributes: attrs}
}

func (s Specifications) Get(key string) string {
    return s.attributes[key]
}

func (s Specifications) ToJSON() string {
    data, _ := json.Marshal(s.attributes)
    return string(data)
}

// 值对象：ProductStatus
type ProductStatus string

const (
    ProductDraft     ProductStatus = "DRAFT"      // 草稿
    ProductOnShelf   ProductStatus = "ON_SHELF"   // 在架
    ProductOffShelf  ProductStatus = "OFF_SHELF"  // 下架
)

// 实体：SPU（标准产品单元）
type SPU struct {
    id         SPU_ID
    title      string
    categoryID int64
    brandID    int64
    attributes map[string][]string  // 属性模板{"颜色":["红","蓝"],"尺寸":["S","M","L"]}
    description string
    
    // SPU下的所有SKU（聚合内实体集合）
    skus []*Product
}

func (spu *SPU) ID() SPU_ID {
    return spu.id
}

func (spu *SPU) Title() string {
    return spu.title
}

func (spu *SPU) AddSKU(sku *Product) error {
    // 不变量检查：SKU规格必须符合SPU属性模板
    if !spu.isValidSpecs(sku.specs) {
        return errors.New("SKU规格不符合SPU属性模板")
    }
    spu.skus = append(spu.skus, sku)
    return nil
}

func (spu *SPU) isValidSpecs(specs Specifications) bool {
    // 检查SKU的规格是否都在SPU的属性模板中
    for key, value := range specs.attributes {
        allowedValues, exists := spu.attributes[key]
        if !exists {
            return false
        }
        if !contains(allowedValues, value) {
            return false
        }
    }
    return true
}
```

##### 16.6.1.12.3 聚合根方法

```go
// 上架（状态转换）
func (p *Product) OnShelf() error {
    if p.status == ProductOnShelf {
        return errors.New("商品已在架")
    }
    
    // 不变量检查：必须有基础价格
    if p.basePrice.Amount() == 0 {
        return errors.New("商品未设置价格，不能上架")
    }
    
    // 不变量检查：必须有商品图片
    if len(p.images) == 0 {
        return errors.New("商品未上传图片，不能上架")
    }
    
    oldStatus := p.status
    p.status = ProductOnShelf
    p.updatedAt = time.Now()
    
    // 发布领域事件
    p.addDomainEvent(&ProductOnShelfEvent{
        SKUID:      p.skuID,
        SPUID:      p.spu.id,
        OnShelfTime: p.updatedAt,
    })
    
    return nil
}

// 下架
func (p *Product) OffShelf(reason string) error {
    if p.status == ProductOffShelf {
        return errors.New("商品已下架")
    }
    
    oldStatus := p.status
    p.status = ProductOffShelf
    p.updatedAt = time.Now()
    
    // 发布领域事件
    p.addDomainEvent(&ProductOffShelfEvent{
        SKUID:       p.skuID,
        Reason:      reason,
        OffShelfTime: p.updatedAt,
    })
    
    return nil
}

// 更新基础价格
func (p *Product) UpdateBasePrice(newPrice Price) error {
    if newPrice.Amount() == p.basePrice.Amount() {
        return nil  // 价格未变化
    }
    
    oldPrice := p.basePrice
    p.basePrice = newPrice
    p.updatedAt = time.Now()
    
    // 发布领域事件
    p.addDomainEvent(&PriceChangedEvent{
        SKUID:    p.skuID,
        OldPrice: oldPrice.Amount(),
        NewPrice: newPrice.Amount(),
        ChangedAt: p.updatedAt,
    })
    
    return nil
}

// 领域事件管理
func (p *Product) addDomainEvent(event DomainEvent) {
    p.domainEvents = append(p.domainEvents, event)
}

func (p *Product) DomainEvents() []DomainEvent {
    return p.domainEvents
}

func (p *Product) ClearDomainEvents() {
    p.domainEvents = nil
}

// 查询方法
func (p *Product) IsOnShelf() bool {
    return p.status == ProductOnShelf
}

func (p *Product) BasePrice() Price {
    return p.basePrice
}

func (p *Product) Specs() Specifications {
    return p.specs
}
```

##### 16.6.1.12.4 Repository 模式（防腐层）

```go
// ProductRepository接口（领域层定义）
type ProductRepository interface {
    // 查询
    FindBySKUID(ctx context.Context, skuID SKU_ID) (*Product, error)
    FindBySPUID(ctx context.Context, spuID SPU_ID) ([]*Product, error)
    BatchFindBySKUIDs(ctx context.Context, skuIDs []SKU_ID) ([]*Product, error)
    
    // 保存
    Save(ctx context.Context, product *Product) error
    Update(ctx context.Context, product *Product) error
    
    // 删除
    Delete(ctx context.Context, skuID SKU_ID) error
}

// ProductRepositoryImpl实现（基础设施层）
type ProductRepositoryImpl struct {
    db             *gorm.DB
    cache          cache.Cache
    eventPublisher EventPublisher
    sharding       ShardingStrategy
}

func (r *ProductRepositoryImpl) FindBySKUID(ctx context.Context, skuID SKU_ID) (*Product, error) {
    // Step 1: 查询L1本地缓存
    cacheKey := fmt.Sprintf("product:%d", skuID.Int64())
    if cached, found := r.cache.GetLocal(cacheKey); found {
        return cached.(*Product), nil
    }
    
    // Step 2: 查询L2 Redis缓存
    if cached, err := r.cache.Get(ctx, cacheKey); err == nil {
        product := r.unmarshalProduct(cached)
        r.cache.SetLocal(cacheKey, product, 1*time.Minute)
        return product, nil
    }
    
    // Step 3: 查询MySQL
    productDO, err := r.queryFromDB(ctx, skuID)
    if err != nil {
        return nil, err
    }
    
    // Step 4: 转换DO → Domain Model
    product := r.toDomain(productDO)
    
    // Step 5: 回写缓存
    r.cache.Set(ctx, cacheKey, r.marshalProduct(product), 30*time.Minute)
    r.cache.SetLocal(cacheKey, product, 1*time.Minute)
    
    return product, nil
}

func (r *ProductRepositoryImpl) Save(ctx context.Context, product *Product) error {
    // Step 1: 转换Domain Model → DO
    productDO := r.toDataObject(product)
    
    // Step 2: 分库路由
    db := r.sharding.Route(product.spu.categoryID)
    
    // Step 3: 保存到数据库
    if err := db.WithContext(ctx).Create(productDO).Error; err != nil {
        return fmt.Errorf("save product failed: %w", err)
    }
    
    // Step 4: 发布领域事件（事务提交后）
    for _, event := range product.DomainEvents() {
        if err := r.eventPublisher.Publish(ctx, event); err != nil {
            log.Errorf("publish event failed: %v", err)
        }
    }
    product.ClearDomainEvents()
    
    // Step 5: 清除缓存
    cacheKey := fmt.Sprintf("product:%d", product.skuID.Int64())
    r.cache.Delete(ctx, cacheKey)
    
    return nil
}

func (r *ProductRepositoryImpl) BatchFindBySKUIDs(ctx context.Context, skuIDs []SKU_ID) ([]*Product, error) {
    products := make([]*Product, 0, len(skuIDs))
    
    // 批量查询优化：分离缓存命中和未命中
    var missedIDs []SKU_ID
    
    for _, skuID := range skuIDs {
        cacheKey := fmt.Sprintf("product:%d", skuID.Int64())
        if cached, err := r.cache.Get(ctx, cacheKey); err == nil {
            products = append(products, r.unmarshalProduct(cached))
        } else {
            missedIDs = append(missedIDs, skuID)
        }
    }
    
    // 批量查询数据库（未命中的）
    if len(missedIDs) > 0 {
        missedProducts, err := r.batchQueryFromDB(ctx, missedIDs)
        if err != nil {
            return nil, err
        }
        
        // 回写缓存
        for _, product := range missedProducts {
            cacheKey := fmt.Sprintf("product:%d", product.skuID.Int64())
            r.cache.Set(ctx, cacheKey, r.marshalProduct(product), 30*time.Minute)
        }
        
        products = append(products, missedProducts...)
    }
    
    return products, nil
}
```

---

##### 16.6.1.12.5 基础设施层（Infrastructure Layer）

###### 16.6.1.12.5.1 核心存储设计

**表结构设计**：

```sql
-- SPU表（标准产品单元）
CREATE TABLE product_spu (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL COMMENT '商品标题',
    category_id BIGINT NOT NULL COMMENT '类目ID',
    brand_id BIGINT COMMENT '品牌ID',
    attributes JSON COMMENT '属性模板',
    description TEXT COMMENT '商品描述',
    status VARCHAR(20) DEFAULT 'DRAFT' COMMENT '状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category_id),
    INDEX idx_brand (brand_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SPU表';

-- SKU表（库存保持单元）
CREATE TABLE product_sku (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    spu_id BIGINT NOT NULL COMMENT 'SPU ID',
    sku_code VARCHAR(100) UNIQUE NOT NULL COMMENT 'SKU编码',
    specs JSON COMMENT '规格值',
    base_price BIGINT NOT NULL COMMENT '基础价格（分）',
    images JSON COMMENT '商品图片',
    status VARCHAR(20) DEFAULT 'DRAFT' COMMENT '状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_spu (spu_id),
    INDEX idx_code (sku_code),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SKU表';

-- 类目表
CREATE TABLE product_category (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    parent_id BIGINT DEFAULT 0 COMMENT '父类目ID',
    level INT DEFAULT 1 COMMENT '层级',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status VARCHAR(20) DEFAULT 'ACTIVE',
    INDEX idx_parent (parent_id),
    INDEX idx_level (level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品类目表';
```

**分库分表策略**：

```sql
-- 按 category_id 分4库
-- 理由：同品类商品通常一起查询（搜索、推荐）
db_index = category_id % 4

-- 单表不分表
-- 理由：单品类商品数量可控（< 100万），查询模式简单
```

**索引策略**：

| 索引名 | 字段 | 类型 | 用途 |
|-------|------|------|------|
| PRIMARY | id | 主键 | 主键查询 |
| idx_category | category_id | 普通 | 类目查询 |
| idx_brand | brand_id | 普通 | 品牌查询 |
| idx_status | status | 普通 | 状态筛选 |
| idx_spu | spu_id | 普通 | SPU查SKU |
| idx_code | sku_code | 唯一 | SKU编码查询 |

---

###### 16.6.1.12.5.2 缓存策略

详见下文"16.6.1.12.8 三级缓存实现"。

---

###### 16.6.1.12.5.3 消息中间件（Messaging）⭐️

**职责划分**：

| 组件 | 层级 | 职责 | 示例 |
|-----|------|------|------|
| **Kafka Producer** | Infrastructure | 事件发布（技术实现） | 发送消息到Kafka Topic |
| **Kafka Consumer** | Infrastructure | 事件消费（技术实现） | 订阅Topic、接收消息、路由 |
| **Event Handler** | Interface | 协议适配 | Kafka消息 → DTO |

**Kafka Producer（事件发布）**：

```go
// internal/infrastructure/messaging/kafka_producer.go

type KafkaProducer struct {
    producer *kafka.Producer
}

func (p *KafkaProducer) Publish(ctx context.Context, event domain.DomainEvent) error {
    topic := p.getTopicByEventType(event.EventType())
    
    // 序列化事件
    data, _ := json.Marshal(event)
    
    // 发送到Kafka
    return p.producer.Produce(&kafka.Message{
        TopicPartition: kafka.TopicPartition{
            Topic:     &topic,
            Partition: kafka.PartitionAny,
        },
        Key:   []byte(event.EventType()),
        Value: data,
        Headers: []kafka.Header{
            {Key: "event_type", Value: []byte(event.EventType())},
            {Key: "timestamp", Value: []byte(fmt.Sprint(event.OccurredAt().Unix()))},
        },
    }, nil)
}
```

**Kafka Consumer（事件消费）**：

```go
// internal/infrastructure/messaging/kafka_consumer.go

type KafkaConsumer struct {
    consumer     *kafka.Consumer
    eventHandler *event.ProductEventHandler  // 注入Interface Layer的Handler
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
    // 订阅Topic
    c.consumer.SubscribeTopics([]string{
        "supplier-product-events",
        "pricing-events",
    }, nil)
    
    // 消费循环
    for {
        msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
        if err != nil {
            continue
        }
        
        // ⭐️ 路由到Interface Layer的Event Handler
        messageType := string(msg.Key)
        if err := c.eventHandler.HandleMessage(ctx, messageType, msg.Value); err != nil {
            log.Errorf("Handle message failed: %v", err)
        } else {
            c.consumer.CommitMessage(msg)  // 手动提交offset
        }
    }
}
```

**Topic设计**：

| Topic | 生产者 | 消费者 | 用途 |
|-------|--------|--------|------|
| `product-domain-events` | Product Service | Search Service, Marketing Service | 商品领域事件（商品创建、上架、价格变更） |
| `supplier-product-events` | Supplier Service | Product Service | 供应商商品事件（供应商创建商品） |
| `pricing-events` | Pricing Service | Product Service | 定价事件（价格计算完成） |

详见 `16.6.1.12.7.4 事件订阅者的分层设计`。

---

##### 16.6.1.12.6 接口层（Interface Layer）

###### 16.6.1.12.6.1 gRPC 接口定义

**核心接口**（product.proto）：

```protobuf
// ProductService商品服务
service ProductService {
    // 查询单个商品
    rpc GetProduct(GetProductRequest) returns (GetProductResponse);
    
    // 批量查询商品
    rpc BatchGetProducts(BatchGetProductsRequest) returns (BatchGetProductsResponse);
    
    // 创建商品
    rpc CreateProduct(CreateProductRequest) returns (CreateProductResponse);
    
    // 更新基础价格
    rpc UpdateBasePrice(UpdateBasePriceRequest) returns (UpdateBasePriceResponse);
    
    // 上架
    rpc OnShelf(OnShelfRequest) returns (OnShelfResponse);
    
    // 下架
    rpc OffShelf(OffShelfRequest) returns (OffShelfResponse);
}

message GetProductRequest {
    int64 sku_id = 1;
}

message GetProductResponse {
    ProductInfo product = 1;
}

message BatchGetProductsRequest {
    repeated int64 sku_ids = 1;  // 最多100个
}

message BatchGetProductsResponse {
    repeated ProductInfo products = 1;
}

message ProductInfo {
    int64 sku_id = 1;
    int64 spu_id = 2;
    string sku_code = 3;
    string sku_name = 4;
    Price base_price = 5;
    Specifications specs = 6;
    ProductStatus status = 7;
}

message Price {
    int64 amount = 1;  // 金额（分）
    string currency = 2;  // 货币（CNY）
}

message Specifications {
    string color = 1;
    string size = 2;
    map<string, string> attrs = 3;  // 其他属性
}

enum ProductStatus {
    DRAFT = 0;      // 草稿
    ON_SHELF = 1;   // 上架
    OFF_SHELF = 2;  // 下架
}
```

###### 16.6.1.12.6.2 HTTP 接口（可选）

```go
// HTTP接口（供运营后台使用）
GET    /api/v1/products/:sku_id           # 查询商品
POST   /api/v1/products                    # 创建商品
PUT    /api/v1/products/:sku_id           # 更新商品
POST   /api/v1/products/:sku_id/on-shelf  # 上架
POST   /api/v1/products/:sku_id/off-shelf # 下架
```

###### 16.6.1.12.6.3 Event 接口（异步）⭐️

**事件订阅接口**（接收外部服务事件）：

```go
// ProductEventHandler 商品事件处理器（接口层）
// 职责：适配外部事件消息 → 调用Application Service
type ProductEventHandler struct {
    productService *service.ProductService
}

// 处理消息入口
func (h *ProductEventHandler) HandleMessage(ctx context.Context, messageType string, data []byte) error

// 订阅的事件类型
const (
    SupplierProductCreated = "supplier.product.created"  // 供应商商品创建
    PricingPriceChanged    = "pricing.price_changed"     // 定价变更
)
```

**与HTTP/gRPC的区别**：
- **同步接口**（HTTP/gRPC）：客户端等待响应
- **异步接口**（Event）：消息队列异步触发，无响应

**职责**：
- ✅ 协议适配（Kafka消息 → DTO）
- ✅ 调用Application Service
- ❌ 不负责Kafka连接（由Infrastructure Layer的Kafka Consumer负责）

详见 `16.6.1.12.7.4 事件订阅者的分层设计`。

---

##### 16.6.1.12.7 应用服务层（Application Layer）

###### 16.6.1.12.7.1 核心代码结构

```
product-service/
├── cmd/
│   └── main.go                          # 服务入口
├── internal/
│   ├── domain/                          # 领域模型层
│   │   ├── product.go                   # Product聚合根
│   │   ├── spu.go                       # SPU实体
│   │   ├── value_objects.go             # 值对象（SKU_ID, Price, Specifications）
│   │   ├── events.go                    # 领域事件
│   │   └── repository.go                # Repository接口
│   ├── application/                     # 应用服务层
│   │   ├── dto/
│   │   │   ├── product_request.go       # 请求DTO
│   │   │   └── product_response.go      # 响应DTO
│   │   └── service/
│   │       ├── product_service.go       # 商品应用服务
│   │       └── product_query_service.go # 查询服务（CQRS）
│   ├── infrastructure/                  # 基础设施层
│   │   ├── persistence/
│   │   │   ├── product_repository.go    # Repository实现
│   │   │   ├── data_object.go           # 数据对象（DO）
│   │   │   └── sharding.go              # 分库路由
│   │   ├── cache/
│   │   │   ├── redis_cache.go           # Redis缓存
│   │   │   └── local_cache.go           # 本地缓存
│   │   └── messaging/                   # 消息中间件 ⭐️
│   │       ├── kafka_producer.go        # Kafka生产者（事件发布）
│   │       └── kafka_consumer.go        # Kafka消费者（技术实现）
│   └── interfaces/                      # 接口层
│       ├── grpc/
│       │   ├── product_handler.go       # gRPC处理器
│       │   └── proto/
│       │       └── product.proto        # Protobuf定义
│       ├── http/
│       │   └── product_handler.go       # HTTP处理器（可选）
│       └── event/ ⭐️                     # 事件接口（异步）
│           └── product_event_handler.go # Event Handler
├── config/
│   └── config.yaml                      # 配置文件
├── migrations/                          # 数据库迁移
│   └── 001_create_product_tables.sql
└── go.mod
```

---

###### 16.6.1.12.7.2 核心应用服务实现

**应用服务层**（product_service.go）：

```go
type ProductService struct {
    repo           domain.ProductRepository
    eventPublisher EventPublisher
}

// GetProduct 查询商品（三级缓存）
func (s *ProductService) GetProduct(ctx context.Context, skuID int64) (*dto.ProductResponse, error) {
    // Step 1: 通过Repository查询（Repository内部实现三级缓存）
    product, err := s.repo.FindBySKUID(ctx, domain.NewSKU_ID(skuID))
    if err != nil {
        return nil, fmt.Errorf("product not found: %w", err)
    }
    
    // Step 2: Domain Model → DTO
    return s.toDTO(product), nil
}

// BatchGetProducts 批量查询商品
func (s *ProductService) BatchGetProducts(ctx context.Context, skuIDs []int64) ([]*dto.ProductResponse, error) {
    // 参数校验：限制批量大小
    if len(skuIDs) > 100 {
        return nil, errors.New("批量查询最多100个")
    }
    
    // 转换为值对象
    domainIDs := make([]domain.SKU_ID, len(skuIDs))
    for i, id := range skuIDs {
        domainIDs[i] = domain.NewSKU_ID(id)
    }
    
    // 批量查询
    products, err := s.repo.BatchFindBySKUIDs(ctx, domainIDs)
    if err != nil {
        return nil, err
    }
    
    // 转换为DTO
    dtos := make([]*dto.ProductResponse, len(products))
    for i, p := range products {
        dtos[i] = s.toDTO(p)
    }
    
    return dtos, nil
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
    // Step 1: DTO → Domain Model
    product, err := s.buildProduct(req)
    if err != nil {
        return nil, fmt.Errorf("build product failed: %w", err)
    }
    
    // Step 2: 保存（Repository内部发布领域事件）
    if err := s.repo.Save(ctx, product); err != nil {
        return nil, fmt.Errorf("save product failed: %w", err)
    }
    
    return s.toDTO(product), nil
}

// OnShelf 商品上架
func (s *ProductService) OnShelf(ctx context.Context, skuID int64) error {
    // Step 1: 查询聚合根
    product, err := s.repo.FindBySKUID(ctx, domain.NewSKU_ID(skuID))
    if err != nil {
        return err
    }
    
    // Step 2: 执行领域逻辑（状态转换）
    if err := product.OnShelf(); err != nil {
        return err
    }
    
    // Step 3: 保存聚合根（自动发布领域事件）
    return s.repo.Update(ctx, product)
}

// UpdateBasePrice 更新基础价格
func (s *ProductService) UpdateBasePrice(ctx context.Context, skuID int64, newPrice int64) error {
    // Step 1: 查询聚合根
    product, err := s.repo.FindBySKUID(ctx, domain.NewSKU_ID(skuID))
    if err != nil {
        return err
    }
    
    // Step 2: 创建价格值对象（带校验）
    price, err := domain.NewPrice(newPrice)
    if err != nil {
        return fmt.Errorf("invalid price: %w", err)
    }
    
    // Step 3: 执行领域逻辑
    if err := product.UpdateBasePrice(price); err != nil {
        return err
    }
    
    // Step 4: 保存聚合根（自动发布PriceChangedEvent）
    return s.repo.Update(ctx, product)
}
```

---

###### 16.6.1.12.7.3 领域事件

| 事件名 | 触发时机 | 事件数据 | 消费方 | Topic | 用途 |
|-------|---------|---------|--------|-------|------|
| **ProductCreated** | 商品创建成功 | sku_id, spu_id, title, category_id, base_price | Search Service, Recommendation | product-events | 同步到ES索引 |
| **ProductUpdated** | 商品信息更新 | sku_id, changed_fields | Search Service, Cache Invalidation | product-events | 更新ES、清缓存 |
| **ProductOnShelf** | 商品上架 | sku_id, spu_id, on_shelf_time | Search Service, Marketing | product-events | 上架通知、活动关联 |
| **ProductOffShelf** | 商品下架 | sku_id, reason, off_shelf_time | Search Service, Order Service | product-events | 从ES移除、停止接单 |
| **PriceChanged** | 基础价格变更 | sku_id, old_price, new_price | Pricing Service, Analytics | product-events | 重新计算售价、价格分析 |

**事件结构定义**：

```go
// ProductCreatedEvent 商品创建事件
type ProductCreatedEvent struct {
    SKUID      int64     `json:"sku_id"`
    SPUID      int64     `json:"spu_id"`
    Title      string    `json:"title"`
    CategoryID int64     `json:"category_id"`
    BasePrice  int64     `json:"base_price"`
    CreatedAt  time.Time `json:"created_at"`
}

func (e *ProductCreatedEvent) Type() string {
    return "product.created"
}

// ProductOnShelfEvent 商品上架事件
type ProductOnShelfEvent struct {
    SKUID       int64     `json:"sku_id"`
    SPUID       int64     `json:"spu_id"`
    OnShelfTime time.Time `json:"on_shelf_time"`
}

func (e *ProductOnShelfEvent) Type() string {
    return "product.on_shelf"
}

// PriceChangedEvent 价格变更事件
type PriceChangedEvent struct {
    SKUID     int64     `json:"sku_id"`
    OldPrice  int64     `json:"old_price"`
    NewPrice  int64     `json:"new_price"`
    ChangedAt time.Time `json:"changed_at"`
}

func (e *PriceChangedEvent) Type() string {
    return "product.price_changed"
}
```

---

###### 16.6.1.12.7.4 事件订阅者的分层设计 ⭐️

**核心问题**：DDD架构中，事件订阅者（Event Subscriber）应该放在哪一层？

这是微服务架构中的经典设计问题。不同的分层方案会影响代码的复用性、可测试性和职责清晰度。

###### 方案对比

**方案A：Interface Layer（推荐）⭐️**

事件订阅者是"异步接口"，与HTTP/gRPC同级。

```
┌─────────────────────────────────────────────────────────────┐
│              Interface Layer (接口层)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │ HTTP Handler │  │ gRPC Handler │  │Event Subscriber│     │
│  │  (同步接口)   │  │  (同步接口)   │  │  (异步接口) ⭐️│     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘       │
│         │                  │                  │               │
└─────────┼──────────────────┼──────────────────┼─────────────┘
          ↓                  ↓                  ↓
┌─────────────────────────────────────────────────────────────┐
│    同一个 Application Service (ProductService)                │
│    - GetProduct()     (查询)                                 │
│    - CreateProduct()  (命令)                                 │
│    - OnShelf()        (命令)                                 │
│    - UpdatePrice()    (命令)                                 │
└─────────────────────────────────────────────────────────────┘
```

**方案B：Infrastructure Layer（不推荐）**

Kafka Consumer直接调用Application Service。

问题：
- ❌ 违反依赖倒置原则（Infrastructure依赖Application）
- ❌ 职责不清晰（Infrastructure既是实现层又是入口层）
- ❌ 难以替换消息队列（从Kafka切换到RabbitMQ需要大量修改）

**推荐方案A的原因**：

1. **对称性**：HTTP、gRPC、Event都是外部触发源，应该同级
2. **复用性**：Application Service被所有接口复用，业务逻辑只写一次
3. **职责清晰**：Interface负责协议适配，Infrastructure负责技术实现
4. **易于测试**：可以直接测试Application Service，不依赖Kafka
5. **易于替换**：更换消息队列只需修改Infrastructure Layer

---

###### 完整调用链路

**HTTP同步调用**：

```
Client
  ↓ HTTP Request
┌─────────────────────────────────┐
│ Interface Layer - HTTP Handler  │ ← 解析HTTP请求
│ product_handler.go              │
└──────────────┬──────────────────┘
               ↓ DTO
┌─────────────────────────────────┐
│ Application Layer               │ ← 业务编排
│ product_service.go              │
└──────────────┬──────────────────┘
               ↓ Domain Model
┌─────────────────────────────────┐
│ Domain Layer                    │ ← 业务规则
│ product.go                      │
└──────────────┬──────────────────┘
               ↓ Repository
┌─────────────────────────────────┐
│ Infrastructure Layer            │ ← 数据持久化
│ product_repository.go           │
└─────────────────────────────────┘
```

**Kafka异步调用**：

```
Kafka Topic (supplier-product-events)
  ↓ 异步消息
┌─────────────────────────────────┐
│ Infrastructure Layer            │ ← 技术实现（Kafka连接、消息接收）
│ kafka_consumer.go               │
└──────────────┬──────────────────┘
               ↓ 消息路由
┌─────────────────────────────────┐
│ Interface Layer - Event Handler │ ← 协议适配（Kafka消息 → DTO）
│ product_event_handler.go        │
└──────────────┬──────────────────┘
               ↓ DTO
┌─────────────────────────────────┐
│ Application Layer               │ ← 业务编排（复用同一个Service）
│ product_service.go              │
└──────────────┬──────────────────┘
               ↓ Domain Model
┌─────────────────────────────────┐
│ Domain Layer                    │ ← 业务规则
│ product.go                      │
└──────────────┬──────────────────┘
               ↓ Repository
┌─────────────────────────────────┐
│ Infrastructure Layer            │ ← 数据持久化
│ product_repository.go           │
└─────────────────────────────────┘
```

**关键差异**：
- HTTP: `Interface → Application`
- Event: `Infrastructure → Interface → Application`（多一层技术实现）

---

###### 目录结构调整

```diff
product-service/
├── internal/
│   ├── interfaces/                      # 接口层
│   │   ├── http/
│   │   │   └── product_handler.go       # HTTP Handler（同步）
│   │   ├── grpc/
│   │   │   ├── proto/product.proto      
│   │   │   └── product_handler.go       # gRPC Handler（同步）
+│   │   └── event/ ⭐️                    # 新增：事件接口
+│   │       └── product_event_handler.go # Event Handler（异步接口）
│   │
│   ├── application/                     
│   │   └── service/
│   │       └── product_service.go       # 应用服务（被所有接口复用）
│   │
│   ├── infrastructure/                  
│   │   ├── persistence/
│   │   │   └── product_repository.go    
-│   │   └── event/
-│   │       └── kafka_publisher.go      # 事件发布
+│   │   └── messaging/ ⭐️                # 重命名：消息中间件
+│   │       ├── kafka_producer.go       # Kafka生产者（事件发布）
+│   │       └── kafka_consumer.go       # Kafka消费者（技术实现）
```

---

###### Interface Layer - Event Handler实现

**职责**：协议适配（Kafka消息 → DTO → 调用Application Service）

```go
// internal/interfaces/event/product_event_handler.go

package event

import (
    "context"
    "encoding/json"
    "fmt"

    "product-service/internal/application/dto"
    "product-service/internal/application/service"
)

// ProductEventHandler 商品事件处理器（接口层）
// 职责：适配外部事件消息 → 调用Application Service
// 与HTTP/gRPC Handler同级，是"异步接口"
type ProductEventHandler struct {
    productService *service.ProductService
}

func NewProductEventHandler(productService *service.ProductService) *ProductEventHandler {
    return &ProductEventHandler{
        productService: productService,
    }
}

// HandleMessage 统一的消息处理入口
// 由Infrastructure Layer的Kafka Consumer调用
func (h *ProductEventHandler) HandleMessage(ctx context.Context, messageType string, data []byte) error {
    fmt.Printf("🔔 [Interface Layer - Event] Received message: %s\n", messageType)

    switch messageType {
    case "supplier.product.created":
        return h.handleSupplierProductCreated(ctx, data)

    case "pricing.price_changed":
        return h.handlePriceChanged(ctx, data)

    default:
        fmt.Printf("⚠️  Unknown message type: %s\n", messageType)
        return nil
    }
}

// handleSupplierProductCreated 处理供应商商品创建事件
// 场景：供应商服务创建新商品后，通过Kafka通知商品服务同步
func (h *ProductEventHandler) handleSupplierProductCreated(ctx context.Context, data []byte) error {
    // Step 1: 反序列化Kafka消息
    var kafkaEvent struct {
        SupplierID  int64  `json:"supplier_id"`
        SupplierSKU string `json:"supplier_sku"`
        Title       string `json:"title"`
        BasePrice   int64  `json:"base_price"`
        CategoryID  int64  `json:"category_id"`
    }
    if err := json.Unmarshal(data, &kafkaEvent); err != nil {
        return fmt.Errorf("反序列化失败: %w", err)
    }

    // Step 2: Kafka消息 → DTO（协议适配）
    req := &dto.CreateProductRequest{
        SupplierID:  kafkaEvent.SupplierID,
        SupplierSKU: kafkaEvent.SupplierSKU,
        Title:       kafkaEvent.Title,
        BasePrice:   kafkaEvent.BasePrice,
        CategoryID:  kafkaEvent.CategoryID,
    }

    // Step 3: 调用应用服务（与HTTP/gRPC调用同一个方法！）
    resp, err := h.productService.CreateProduct(ctx, req)
    if err != nil {
        return fmt.Errorf("创建商品失败: %w", err)
    }

    fmt.Printf("✅ [Interface Layer - Event] Product created, SKUID=%d\n", resp.SKUID)
    return nil
}

// handlePriceChanged 处理价格变更事件
// 场景：定价服务计算出新价格后，通知商品服务更新基础价格
func (h *ProductEventHandler) handlePriceChanged(ctx context.Context, data []byte) error {
    var kafkaEvent struct {
        SKUID    int64 `json:"sku_id"`
        NewPrice int64 `json:"new_price"`
    }
    if err := json.Unmarshal(data, &kafkaEvent); err != nil {
        return fmt.Errorf("反序列化失败: %w", err)
    }

    // Kafka消息 → DTO
    req := &dto.UpdatePriceRequest{
        SKUID:    kafkaEvent.SKUID,
        NewPrice: kafkaEvent.NewPrice,
    }

    // 调用应用服务（复用业务逻辑）
    _, err := h.productService.UpdateBasePrice(ctx, req)
    if err != nil {
        return fmt.Errorf("更新价格失败: %w", err)
    }

    fmt.Printf("✅ [Interface Layer - Event] Price updated, SKUID=%d\n", kafkaEvent.SKUID)
    return nil
}
```

**关键设计点**：

1. **协议适配**：将Kafka特有的消息格式转换为通用的DTO
2. **复用Service**：调用 `productService.CreateProduct()`（与HTTP/gRPC同一个方法）
3. **错误处理**：返回错误给Kafka Consumer，由Consumer决定重试或DLQ
4. **日志跟踪**：打印日志便于调试异步流程

---

###### Infrastructure Layer - Kafka Consumer实现

**职责**：技术实现（Kafka连接、订阅Topic、接收消息、路由到Interface Layer）

```go
// internal/infrastructure/messaging/kafka_consumer.go

package messaging

import (
    "context"
    "fmt"
    "time"

    eventHandler "product-service/internal/interfaces/event"
    "github.com/confluentinc/confluent-kafka-go/kafka"
)

// KafkaConsumer Kafka消费者（基础设施层）
// 职责：管理Kafka连接、订阅Topic、接收消息、路由到Interface Layer
type KafkaConsumer struct {
    consumer     *kafka.Consumer
    eventHandler *eventHandler.ProductEventHandler
    topics       []string
}

func NewKafkaConsumer(eventHandler *eventHandler.ProductEventHandler) *KafkaConsumer {
    return &KafkaConsumer{
        eventHandler: eventHandler,
        topics: []string{
            "supplier-product-events",  // 供应商事件
            "pricing-events",           // 定价事件
        },
    }
}

// Start 启动消费者（阻塞）
func (c *KafkaConsumer) Start(ctx context.Context) error {
    fmt.Printf("📡 [Infrastructure - Kafka Consumer] Starting...\n")
    
    // 初始化Kafka Consumer
    consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
        "bootstrap.servers": "localhost:9092",
        "group.id":          "product-service-group",
        "auto.offset.reset": "earliest",
        "enable.auto.commit": false,  // 手动提交offset（保证at-least-once）
    })
    if err != nil {
        return fmt.Errorf("创建Kafka Consumer失败: %w", err)
    }
    c.consumer = consumer

    // 订阅Topic
    if err := c.consumer.SubscribeTopics(c.topics, nil); err != nil {
        return fmt.Errorf("订阅Topic失败: %w", err)
    }
    
    fmt.Printf("📡 [Infrastructure - Kafka Consumer] Subscribed to: %v\n", c.topics)

    // 消费循环
    for {
        select {
        case <-ctx.Done():
            fmt.Println("📡 [Infrastructure - Kafka Consumer] Stopping...")
            return ctx.Err()
            
        default:
            // 读取消息（100ms超时）
            msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
            if err != nil {
                continue  // 超时或临时错误，继续循环
            }

            // ⭐️ 路由消息到Interface Layer的Handler
            messageType := string(msg.Key)
            if err := c.routeMessage(ctx, messageType, msg.Value); err != nil {
                fmt.Printf("❌ [Infrastructure - Kafka Consumer] Handle error: %v\n", err)
                // 错误处理：重试或发送到DLQ (Dead Letter Queue)
            } else {
                // ⭐️ 手动提交offset（保证at-least-once）
                c.consumer.CommitMessage(msg)
            }
        }
    }
}

// routeMessage 路由消息到Interface Layer的Handler
func (c *KafkaConsumer) routeMessage(ctx context.Context, messageType string, data []byte) error {
    fmt.Printf("📬 [Infrastructure - Kafka Consumer] Routing: %s\n", messageType)

    // ⭐️ 调用Interface Layer的Handler
    if err := c.eventHandler.HandleMessage(ctx, messageType, data); err != nil {
        return fmt.Errorf("处理消息失败: %w", err)
    }

    return nil
}

// Stop 停止消费者
func (c *KafkaConsumer) Stop() error {
    if c.consumer != nil {
        return c.consumer.Close()
    }
    return nil
}
```

**关键设计点**：

1. **技术实现**：负责Kafka连接、消息接收等技术细节
2. **消息路由**：根据消息类型路由到Interface Layer的Handler
3. **手动提交Offset**：保证at-least-once语义（消息处理成功后才提交）
4. **错误处理**：失败消息可以重试或发送到DLQ
5. **优雅停机**：通过Context取消消费循环

---

###### 依赖注入与启动

**main.go**：

```go
package main

import (
    "context"
    "product-service/internal/application/service"
    "product-service/internal/infrastructure/messaging"
    "product-service/internal/infrastructure/persistence"
    eventHandler "product-service/internal/interfaces/event"
    httpHandler "product-service/internal/interfaces/http"
)

func main() {
    // Infrastructure Layer - 持久化
    repo := persistence.NewProductRepository(...)

    // Infrastructure Layer - 消息发布
    eventPublisher := messaging.NewKafkaProducer()

    // Application Layer
    productService := service.NewProductService(repo, eventPublisher)

    // Interface Layer - HTTP
    handler := httpHandler.NewProductHandler(productService)

    // ⭐️ Interface Layer - Event (异步接口)
    evtHandler := eventHandler.NewProductEventHandler(productService)

    // ⭐️ Infrastructure Layer - Kafka Consumer (技术实现)
    kafkaConsumer := messaging.NewKafkaConsumer(evtHandler)

    // 启动HTTP服务器
    go startHTTPServer(handler)

    // ⭐️ 启动Kafka Consumer
    ctx := context.Background()
    if err := kafkaConsumer.Start(ctx); err != nil {
        log.Fatalf("Kafka Consumer error: %v", err)
    }
}
```

**依赖关系**：

```
Infrastructure (KafkaConsumer)
    ↓ 调用
Interface (ProductEventHandler)
    ↓ 调用
Application (ProductService)
    ↓ 调用
Domain (Product)
```

✅ 依赖方向：外层 → 内层（符合依赖倒置原则）

---

###### 实际项目优化

**1. 服务分离部署**

```
product-service-api (处理HTTP/gRPC)
├── Deployment: 10副本
├── 职责：对外接口
└── 扩容依据：QPS

product-service-consumer (处理Kafka事件)
├── Deployment: 3副本
├── Consumer Group: product-service-group
├── 职责：消费事件
└── 扩容依据：消息堆积

共享：
├── Application Service
├── Domain Model
└── Repository
```

**好处**：
- API服务可以独立扩容（根据QPS）
- Consumer服务可以独立扩容（根据消息堆积）
- Consumer故障不影响API可用性

**2. Outbox Pattern（保证一致性）**

```sql
-- 事件表（保证事务一致性）
CREATE TABLE domain_event_outbox (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    aggregate_type VARCHAR(50) NOT NULL COMMENT '聚合类型',
    aggregate_id BIGINT NOT NULL COMMENT '聚合ID',
    event_type VARCHAR(100) NOT NULL COMMENT '事件类型',
    event_data JSON NOT NULL COMMENT '事件数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP NULL COMMENT '发布时间',
    status ENUM('PENDING', 'PUBLISHED', 'FAILED') DEFAULT 'PENDING',
    retry_count INT DEFAULT 0,
    INDEX idx_status_created (status, created_at)
) ENGINE=InnoDB COMMENT='领域事件发件箱';
```

**Outbox发送器**：

```go
// 定时扫描未发布的事件
func (s *OutboxSender) SendPendingEvents(ctx context.Context) error {
    events, err := s.repo.FindPendingEvents(ctx, limit)
    if err != nil {
        return err
    }

    for _, event := range events {
        // 发布到Kafka
        if err := s.kafkaProducer.Publish(ctx, event); err != nil {
            s.repo.MarkFailed(ctx, event.ID)
            continue
        }

        // 标记为已发布
        s.repo.MarkPublished(ctx, event.ID)
    }

    return nil
}
```

**3. 事件幂等性处理**

```go
// Event Handler中的幂等性检查
func (h *ProductEventHandler) handlePriceChanged(ctx context.Context, data []byte) error {
    var event PriceChangedEvent
    json.Unmarshal(data, &event)

    // ⭐️ 检查是否已处理（根据event_id或业务维度）
    if h.isEventProcessed(ctx, event.EventID) {
        fmt.Println("⏭️  Event already processed, skipping...")
        return nil  // 幂等性：已处理，直接返回成功
    }

    // 处理业务逻辑
    if err := h.productService.UpdateBasePrice(ctx, ...); err != nil {
        return err
    }

    // 记录已处理
    h.markEventProcessed(ctx, event.EventID)

    return nil
}

// 使用Redis或DB记录已处理的事件
func (h *ProductEventHandler) isEventProcessed(ctx context.Context, eventID string) bool {
    exists, _ := h.redis.Exists(ctx, "processed:event:"+eventID).Result()
    return exists > 0
}

func (h *ProductEventHandler) markEventProcessed(ctx context.Context, eventID string) {
    // 保存24小时（TTL根据业务重试窗口决定）
    h.redis.SetEX(ctx, "processed:event:"+eventID, "1", 24*time.Hour)
}
```

---

###### 总结

**事件订阅者应该放在Interface Layer**，原因：

1. ✅ **对称性**：HTTP、gRPC、Event都是外部触发源，应该同级
2. ✅ **复用性**：Application Service被所有接口复用，业务逻辑只写一次
3. ✅ **职责清晰**：Interface负责协议适配，Infrastructure负责技术实现
4. ✅ **易于测试**：可以直接测试Application Service，不依赖Kafka
5. ✅ **易于替换**：更换消息队列只需修改Infrastructure Layer

**完整调用链路**：

```
Kafka Topic 
  → Infrastructure (接收消息)
  → Interface (协议适配) 
  → Application (业务编排)
  → Domain (业务规则)
  → Infrastructure (持久化)
```

**示例代码**：参见 `/ecommerce-book/example-codes/product-service/` 完整 Demo

---

##### 16.6.1.12.8 三级缓存实现（Infrastructure Layer 详细实现）

**三级缓存架构**：

```go
// L1: 本地缓存（1分钟）
// 优点：延迟最低（<1ms），适合热点商品
// 缺点：容量有限，多实例不一致
localCache.Set("product:"+skuID, product, 1*time.Minute)

// L2: Redis缓存（30分钟）
// 优点：容量大，多实例共享
// 缺点：网络开销（1-5ms）
redis.Set("product:"+skuID, marshal(product), 30*time.Minute)

// L3: MySQL（源数据）
// 优点：数据权威、一致
// 缺点：延迟最高（10-50ms）
db.QueryOne("SELECT * FROM product_sku WHERE id = ?", skuID)
```

**缓存更新策略**：

1. **商品更新时**：主动删除缓存（Cache Aside模式）
2. **上下架时**：删除L1+L2缓存，强制下次查询走DB
3. **价格变更时**：删除缓存 + 发布PriceChangedEvent通知Pricing Service

**缓存Key设计**：

```
product:{sku_id}                    # 单个商品
product:spu:{spu_id}                # SPU下所有SKU（Hash结构）
product:category:{category_id}      # 类目商品列表（Set结构）
```

---

##### 16.6.1.12.9 完整示例代码 ⭐️

本章节的完整代码实现（包含 DDD 四层架构、事件发布订阅、三级缓存等）详见：

**示例代码路径**：`/ecommerce-book/example-codes/product-service/`

这部分示例代码不是完整生产系统，而是为了帮助读者把前面的架构设计映射到工程结构中。它重点覆盖商品中心的 DDD 分层、聚合根、Repository、缓存、事件发布订阅和接口适配。前面讨论的供给运营、供应商同步、库存可售、搜索导购等能力，在真实项目中会继续扩展为更多应用服务和任务模块；示例工程先保留最小可读骨架，避免把主线淹没在实现细节里。

**目录结构**：

```text
example-codes/product-service/
├── README.md                               # 项目说明
├── QUICKSTART.md                           # 快速开始指南
├── EIGHT_LAYER_MODEL.md                    # 八层商品交易模型说明
├── EVENT_PATTERN.md                        # 事件模式说明
├── EVENT_SUBSCRIBER_LAYER.md               # 事件订阅者分层设计
├── RESTORE_GUIDE.md                        # 示例恢复与排查指南
├── go.mod
├── cmd/
│   └── main.go                             # 程序入口
└── internal/
    ├── domain/                             # 领域层
    │   ├── product.go                      # Product 聚合根
    │   ├── spu.go                          # SPU 实体
    │   ├── value_objects.go                # 值对象
    │   ├── events.go                       # 领域事件
    │   ├── repository.go                   # Repository 接口
    │   ├── category_capability.go          # 品类能力矩阵
    │   ├── runtime_context.go              # 八层运行时上下文
    │   ├── category_strategy.go            # 品类策略接口
    │   └── strategy/                       # Topup/GiftCard/Flight/Hotel策略
    ├── application/                        # 应用层
    │   ├── dto/
    │   │   ├── product_dto.go              # 商品 DTO
    │   │   ├── runtime_context_dto.go      # 八层上下文 DTO
    │   │   └── category_action_dto.go      # 品类动作/垂直搜索 DTO
    │   └── service/
    │       ├── product_service.go          # 商品应用服务
    │       ├── runtime_context_service.go  # 八层上下文应用服务
    │       └── category_action_service.go  # Topup校验/Flight搜索应用服务
    ├── infrastructure/                     # 基础设施层
    │   ├── cache/cache.go                  # 三级缓存
    │   ├── event/
    │   │   ├── event_handlers.go           # 领域事件处理
    │   │   └── event_publisher.go          # 事件发布抽象
    │   ├── messaging/
    │   │   ├── kafka_producer.go           # Kafka 生产者
    │   │   └── kafka_consumer.go           # Kafka 消费者
    │   └── persistence/
    │       ├── product_repository.go       # Repository 实现
    │       ├── data_object.go              # 数据对象
    │       └── capability_repository.go    # 品类能力与示例数据
    └── interfaces/                         # 接口层
        ├── http/product_handler.go         # HTTP 接口
        ├── http/runtime_context_handler.go # 八层上下文 HTTP 接口
        ├── http/category_action_handler.go # 品类动作/垂直搜索 HTTP 接口
        ├── grpc/product_handler.go         # gRPC 接口
        └── event/product_event_handler.go  # Event 接口
```

**章节内容与示例代码的对应关系**：

| 本节设计点 | 示例代码位置 | 说明 |
|-----------|-------------|------|
| DDD 分层结构 | `internal/domain`、`internal/application`、`internal/infrastructure`、`internal/interfaces` | 展示领域层、应用层、基础设施层、接口层的依赖方向 |
| Product 聚合根 | `internal/domain/product.go` | 承载商品状态、基础价格、上下架行为和领域事件 |
| SPU 与值对象 | `internal/domain/spu.go`、`internal/domain/value_objects.go` | 展示 SPU/SKU、价格、规格等核心模型 |
| 八层商品交易模型 | `internal/domain/runtime_context.go`、`internal/domain/category_capability.go` | 展示 Product Definition、Resource、Offer、Availability、Input Schema、Booking、Fulfillment、Refund Rule |
| 品类差异策略 | `internal/domain/strategy` | 展示 Topup、Gift Card、Flight、Hotel 如何分别构建八层上下文 |
| 交易前上下文编排 | `internal/application/service/runtime_context_service.go` | 根据 SKU、类目和场景选择策略并组装 `ProductRuntimeContext` |
| 品类动作与垂直搜索 | `internal/application/service/category_action_service.go`、`internal/interfaces/http/category_action_handler.go` | 展示 Topup 账号校验和 Flight 实时搜索如何复用品类能力，但保留独立场景接口 |
| Repository 抽象 | `internal/domain/repository.go` | 在领域层定义仓储接口，避免领域模型依赖数据库实现 |
| Repository 实现 | `internal/infrastructure/persistence/product_repository.go` | 负责 DO 与 Domain Model 转换、缓存回写和持久化 |
| 品类能力示例数据 | `internal/infrastructure/persistence/capability_repository.go` | 用内存数据模拟品类能力矩阵和商品运行时数据 |
| 缓存策略 | `internal/infrastructure/cache/cache.go` | 演示本地缓存、Redis 缓存和源数据查询的组合方式 |
| HTTP/gRPC 接口 | `internal/interfaces/http`、`internal/interfaces/grpc` | 同步接口入口，负责请求协议到应用服务 DTO 的转换 |
| Event 接口 | `internal/interfaces/event/product_event_handler.go` | 异步接口入口，把外部事件转换为应用服务调用 |
| Kafka 技术实现 | `internal/infrastructure/messaging` | 负责消息队列连接、消费、生产等基础设施细节 |
| 事件模式说明 | `EVENT_PATTERN.md`、`EVENT_SUBSCRIBER_LAYER.md` | 解释事件发布、订阅者分层和接口层定位 |

**运行 Demo**：

```bash
cd ecommerce-book/example-codes/product-service
go run cmd/main.go
```

**学习要点**：

1. **DDD分层架构**：Domain、Application、Infrastructure、Interface四层
2. **事件订阅者分层**：Interface Layer作为异步接口（推荐阅读 `EVENT_SUBSCRIBER_LAYER.md`）
3. **三级缓存实现**：L1本地缓存 → L2 Redis → L3 MySQL
4. **领域事件发布订阅**：Domain产生事件 → Application发布 → Infrastructure投递 → Interface消费
5. **八层商品交易模型**：通过 `ProductRuntimeContext` 展示 Topup、Gift Card、Flight、Hotel 的商品模型差异

---

### 16.6.2 库存系统设计

**二维库存模型**（参考16.2.5）：

```go
// 库存策略接口
type StockStrategy interface {
    CheckStock(ctx context.Context, req *StockRequest) (*StockResponse, error)
    Reserve(ctx context.Context, req *ReserveRequest) (*ReserveResponse, error)
    Deduct(ctx context.Context, req *DeductRequest) error
    Release(ctx context.Context, reserveID string) error
}

// 策略工厂
func NewStockStrategy(managementType ManagementType) StockStrategy {
    switch managementType {
    case Realtime:
        return &RealtimeStockStrategy{}  // 机票、酒店
    case Pooled:
        return &PooledStockStrategy{}    // 优惠券
    case Unlimited:
        return &UnlimitedStockStrategy{} // 充值
    }
}
```

**预占机制**：

```go
// Redis Lua脚本（原子预占）
const reserveScript = `
local stock_key = KEYS[1]
local reserve_key = KEYS[2]
local qty = tonumber(ARGV[1])
local ttl = tonumber(ARGV[2])

local stock = tonumber(redis.call('GET', stock_key) or 0)
if stock >= qty then
    redis.call('DECRBY', stock_key, qty)
    redis.call('SET', reserve_key, qty, 'EX', ttl)
    return 1
else
    return 0
end
`

func (r *StockRepo) Reserve(ctx context.Context, skuID int64, qty int, ttl time.Duration) (string, error) {
    reserveID := generateReserveID()
    stockKey := fmt.Sprintf("stock:%d", skuID)
    reserveKey := fmt.Sprintf("reserve:%s", reserveID)
    
    result, err := r.redis.Eval(ctx, reserveScript, 
        []string{stockKey, reserveKey}, 
        qty, int(ttl.Seconds())).Result()
    
    if result == int64(1) {
        return reserveID, nil
    }
    return "", errors.New("库存不足")
}
```

### 16.6.3 订单系统设计

**状态机**：

```go
type OrderStatus string

const (
    StatusCreated          OrderStatus = "CREATED"           // 已创建
    StatusPendingPayment   OrderStatus = "PENDING_PAYMENT"   // 待支付
    StatusPaid             OrderStatus = "PAID"              // 已支付
    StatusFulfilling       OrderStatus = "FULFILLING"        // 履约中
    StatusFulfilled        OrderStatus = "FULFILLED"         // 已履约
    StatusCanceled         OrderStatus = "CANCELED"          // 已取消
    StatusRefunded         OrderStatus = "REFUNDED"          // 已退款
)

// 状态转换规则
var transitions = map[OrderStatus][]OrderStatus{
    StatusCreated:        {StatusPendingPayment, StatusCanceled},
    StatusPendingPayment: {StatusPaid, StatusCanceled},
    StatusPaid:           {StatusFulfilling, StatusRefunded},
    StatusFulfilling:     {StatusFulfilled, StatusRefunded},
    StatusFulfilled:      {StatusRefunded},  // 已履约可申请退款
}

func (o *Order) TransitionTo(newStatus OrderStatus) error {
    allowed, ok := transitions[o.Status]
    if !ok || !contains(allowed, newStatus) {
        return fmt.Errorf("不允许从 %s 转换到 %s", o.Status, newStatus)
    }
    o.Status = newStatus
    return nil
}
```

**分库分表**（参考ADR-007）：

```
• 分库：按 user_id % 8（用户维度查询最频繁）
• 分表：按 create_time 分表（按月归档，64表）
• 路由表：order_route（order_id → db_index, table_index）
```

### 16.6.4 支付系统设计

**支付流程**：

```go
// Step 1: 创建支付单
func (s *PaymentService) CreatePayment(ctx context.Context, orderID int64, amount int64) (*Payment, error) {
    payment := &Payment{
        ID:      generatePaymentID(),
        OrderID: orderID,
        Amount:  amount,
        Status:  PaymentStatusCreated,
    }
    s.repo.Save(ctx, payment)
    return payment, nil
}

// Step 2: 调用支付渠道（支付宝/微信）
func (s *PaymentService) Pay(ctx context.Context, paymentID int64, channel string) (*PayURL, error) {
    gateway := s.gatewayFactory.Get(channel)
    payURL, err := gateway.CreateOrder(ctx, payment)
    return payURL, err
}

// Step 3: 接收支付回调（幂等处理）
func (s *PaymentService) HandleCallback(ctx context.Context, callbackData *CallbackData) error {
    // 幂等性检查
    payment, err := s.repo.GetByPaymentID(ctx, callbackData.PaymentID)
    if payment.Status == PaymentStatusPaid {
        return nil  // 已处理，幂等返回
    }
    
    // 验签
    if !s.verifySign(callbackData) {
        return errors.New("签名验证失败")
    }
    
    // 更新支付状态（乐观锁）
    affected, err := s.repo.UpdateStatus(ctx, callbackData.PaymentID, 
        PaymentStatusCreated, PaymentStatusPaid)
    if affected == 0 {
        return errors.New("支付单状态已变更")
    }
    
    // 发布支付成功事件
    s.eventPublisher.Publish(ctx, &PaymentPaidEvent{
        OrderID:   payment.OrderID,
        PaymentID: payment.ID,
        Amount:    payment.Amount,
    })
    
    return nil
}
```

**对账流程**：

```go
// 每小时对账任务
func (s *PaymentService) ReconcileHourly(ctx context.Context, hour time.Time) error {
    // Step 1: 获取本地支付记录
    localPayments, _ := s.repo.GetByHour(ctx, hour)
    
    // Step 2: 获取支付渠道对账单
    remotePayments, _ := s.gatewayClient.DownloadBill(ctx, hour)
    
    // Step 3: 比对差异
    diff := s.compare(localPayments, remotePayments)
    
    // Step 4: 处理差异
    for _, d := range diff {
        if d.Type == Missing {
            // 本地有，渠道无 → 可能是渠道延迟
            s.alertService.Alert("支付对账差异", d)
        } else if d.Type == Extra {
            // 本地无，渠道有 → 可能是回调丢失
            s.补单处理(d)
        }
    }
    
    return nil
}
```

### 16.6.5 供应商集成设计

**适配器模式**：

```go
// 供应商接口（统一抽象）
type SupplierAdapter interface {
    QueryStock(ctx context.Context, req *StockQueryRequest) (*StockQueryResponse, error)
    ReserveStock(ctx context.Context, req *ReserveRequest) (*ReserveResponse, error)
    CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
    QueryOrderStatus(ctx context.Context, orderID string) (*OrderStatus, error)
}

// 机票供应商适配器
type FlightSupplierAdapter struct {
    client *FlightSupplierClient
    config *Config
}

func (a *FlightSupplierAdapter) QueryStock(ctx context.Context, req *StockQueryRequest) (*StockQueryResponse, error) {
    // Step 1: 参数转换（平台模型 → 供应商模型）
    supplierReq := a.transformRequest(req)
    
    // Step 2: 调用供应商API（熔断保护）
    supplierResp, err := a.client.QueryAvailability(ctx, supplierReq)
    if err != nil {
        return nil, fmt.Errorf("供应商调用失败: %w", err)
    }
    
    // Step 3: 响应转换（供应商模型 → 平台模型）
    resp := a.transformResponse(supplierResp)
    return resp, nil
}
```

**熔断机制**：

```go
import "github.com/sony/gobreaker"

func NewSupplierClientWithCircuitBreaker(client *http.Client) *SupplierClient {
    cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "SupplierAPI",
        MaxRequests: 3,
        Interval:    10 * time.Second,
        Timeout:     30 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 3 && failureRatio >= 0.5
        },
        OnStateChange: func(name string, from, to gobreaker.State) {
            log.Printf("熔断器 %s 状态变更: %s -> %s", name, from, to)
        },
    })
    
    return &SupplierClient{
        client: client,
        cb:     cb,
    }
}

func (c *SupplierClient) QueryStock(ctx context.Context, req *Request) (*Response, error) {
    result, err := c.cb.Execute(func() (interface{}, error) {
        return c.client.Do(buildHTTPRequest(req))
    })
    if err != nil {
        return nil, err
    }
    return parseResponse(result), nil
}
```

---

## 16.7 完整业务链路（系统集成与数据流）

**从子系统到完整链路**：前面章节展示了各个子系统的内部设计（商品中心、库存、订单、支付、供应商集成），本章展示这些子系统如何协作，形成端到端的业务链路。

**两条关键链路**：
- **B端链路**（供应商 → 运营 → 平台）：商品生命周期管理，决定"商品如何进入、如何管理、如何退出"
- **C端链路**（用户 → 交易 → 履约）：交易流完整路径，决定"用户如何发现、如何下单、如何完成支付"

**集成视角的关键点**：
- **数据流转**：跨系统的数据传递（事件驱动、同步调用、异步任务）
- **状态同步**：多系统间的状态一致性（商品状态、库存状态、订单状态）
- **异常处理**：跨系统的容错与补偿（Saga、重试、降级）

---

### 16.7.1 B端商品生命周期完整链路

**B端商品生命周期是平台运营的核心能力**，决定了"商品如何进入、如何管理、如何退出"。本节展示从商品录入到下架归档的完整链路，涵盖供应商、运营、系统三方协作。

**完整生命周期（7个阶段）**：

```
阶段1：商品录入（手动/批量/API）
   ↓ 录入成功率 > 95%
阶段2：审核发布（人工/自动）
   ↓ 审核通过率 > 85%
阶段3：供应商同步（实时/定时）
   ↓ 同步成功率 > 98%
阶段4：库存管理（同步/监控/对账）
   ↓ 库存准确率 > 99.5%
阶段5：日常维护（单品/批量编辑）
   ↓ 编辑成功率 > 99%
阶段6：促销配置（活动关联/价格设置）
   ↓ 生效准时率 > 99.9%
阶段7：下架归档（临时/永久）
   ↓ 归档成功率 > 99%
```

**与C端链路的对比**：

| 维度 | B端链路 | C端链路 |
|-----|--------|--------|
| **阶段数量** | 7个阶段 | 5个阶段 |
| **时间跨度** | 数天到数月（商品生命周期） | 数分钟（单次购物流程） |
| **参与角色** | 供应商、运营、系统 | 用户、系统 |
| **核心关注** | 数据准确性、流程合规性 | 用户体验、转化率 |
| **关键技术** | 幂等性、状态机、异步任务 | 聚合编排、快照、Saga |

---

#### 阶段1：商品录入

**业务场景**：
- **手动录入**：运营人员通过后台表单录入新商品（小批量、高质量）
- **批量导入**：商家/运营通过Excel批量导入（大促前、品类扩展）
- **API推送**：供应商通过OpenAPI实时推送新品（自动化、规模化）

**技术难点**：
- **幂等性保证**：防止重复提交（网络超时、用户重试）
- **异步解耦**：批量导入通过异步任务处理，避免阻塞用户
- **数据校验**：必填字段、格式校验、业务规则校验

**核心设计**：

```go
// 上架任务状态机
type ListingStatus string

const (
    ListingDraft      ListingStatus = "DRAFT"       // 草稿
    ListingPending    ListingStatus = "PENDING"     // 待审核
    ListingApproved   ListingStatus = "APPROVED"    // 审核通过
    ListingRejected   ListingStatus = "REJECTED"    // 审核驳回
    ListingPublished  ListingStatus = "PUBLISHED"   // 已发布
)

// 上架任务
type ListingTask struct {
    TaskCode    string        // 幂等性标识
    ItemInfo    ItemInfo      // 商品信息
    SupplierID  int64         // 供应商ID
    Status      ListingStatus
    ReviewerID  int64         // 审核人
    RejectReason string       // 驳回原因
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// 创建上架任务（幂等性保证）
func (s *ListingService) CreateListingTask(ctx context.Context, req *ListingRequest) (*ListingTask, error) {
    // Step 1: 生成幂等性标识符
    taskCode := s.generateTaskCode(req)
    
    // Step 2: FirstOrCreate（幂等性）
    task := &ListingTask{
        TaskCode:   taskCode,
        ItemInfo:   req.ItemInfo,
        SupplierID: req.SupplierID,
        Status:     ListingDraft,
    }
    
    result := s.db.Where("task_code = ?", taskCode).FirstOrCreate(task)
    if result.RowsAffected > 0 {
        // 首次创建，发布事件
        s.eventPublisher.Publish(ctx, &ListingTaskCreatedEvent{
            TaskCode: taskCode,
            ItemInfo: req.ItemInfo,
        })
    }
    
    return task, nil
}

// 提交审核
func (s *ListingService) SubmitForReview(ctx context.Context, taskCode string) error {
    // 状态转换：DRAFT → PENDING
    return s.updateStatus(ctx, taskCode, ListingDraft, ListingPending)
}

// 审核通过
func (s *ListingService) Approve(ctx context.Context, taskCode string, reviewerID int64) error {
    // Step 1: 状态转换：PENDING → APPROVED
    if err := s.updateStatus(ctx, taskCode, ListingPending, ListingApproved); err != nil {
        return err
    }
    
    // Step 2: 创建商品记录（写入商品中心）
    task, _ := s.getTask(ctx, taskCode)
    itemID, err := s.productCenter.CreateProduct(ctx, &CreateProductRequest{
        ItemInfo:   task.ItemInfo,
        SupplierID: task.SupplierID,
    })
    if err != nil {
        return fmt.Errorf("create product failed: %w", err)
    }
    
    // Step 3: 初始化库存（调用库存服务）
    s.inventoryClient.InitStock(ctx, itemID, task.ItemInfo.InitStock)
    
    // Step 4: 初始化价格（调用计价服务）
    s.pricingClient.InitPrice(ctx, itemID, task.ItemInfo.BasePrice)
    
    // Step 5: 更新搜索索引（异步）
    s.eventPublisher.Publish(ctx, &ProductCreatedEvent{
        ItemID:     itemID,
        ItemInfo:   task.ItemInfo,
        SupplierID: task.SupplierID,
    })
    
    return nil
}
```

**批量导入**：

```go
// 批量导入服务
type BatchImportService struct {
    taskRepo     TaskRepository
    taskQueue    TaskQueue
    validator    ItemValidator
    fileParser   FileParser
}

// 批量导入（异步）
func (s *BatchImportService) BatchImport(ctx context.Context, file io.Reader, operatorID int64) (*BatchImportTask, error) {
    // Step 1: 解析文件（支持Excel/CSV）
    items, parseErr := s.fileParser.Parse(file)
    if parseErr != nil {
        return nil, fmt.Errorf("文件解析失败: %w", parseErr)
    }
    
    // Step 2: 数据校验（预检）
    validItems := make([]*ItemInfo, 0, len(items))
    invalidItems := make([]*ValidationError, 0)
    
    for _, item := range items {
        if err := s.validator.Validate(item); err != nil {
            invalidItems = append(invalidItems, &ValidationError{
                Item:  item,
                Error: err.Error(),
            })
        } else {
            validItems = append(validItems, item)
        }
    }
    
    // Step 3: 创建批量任务
    batchTask := &BatchImportTask{
        TaskID:       generateTaskID(),
        TotalCount:   len(items),
        ValidCount:   len(validItems),
        InvalidCount: len(invalidItems),
        Status:       "PENDING",
        OperatorID:   operatorID,
        CreatedAt:    time.Now(),
    }
    
    if err := s.taskRepo.Save(ctx, batchTask); err != nil {
        return nil, err
    }
    
    // Step 4: 发送到任务队列（异步处理）
    s.taskQueue.Publish(ctx, &BatchImportEvent{
        TaskID:       batchTask.TaskID,
        Items:        validItems,
        InvalidItems: invalidItems,
    })
    
    return batchTask, nil
}

// 批量任务处理（Consumer）
func (s *BatchImportService) ProcessBatchTask(ctx context.Context, event *BatchImportEvent) error {
    successCount := 0
    failedItems := make([]*ImportFailure, 0)
    
    // 逐条处理（控制并发度）
    semaphore := make(chan struct{}, 10) // 限制10并发
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for _, item := range event.Items {
        wg.Add(1)
        semaphore <- struct{}{}
        
        go func(item *ItemInfo) {
            defer wg.Done()
            defer func() { <-semaphore }()
            
            // 创建上架任务
            if _, err := s.listingService.CreateListingTask(ctx, &ListingRequest{
                ItemInfo:   item,
                SupplierID: item.SupplierID,
            }); err != nil {
                mu.Lock()
                failedItems = append(failedItems, &ImportFailure{
                    Item:  item,
                    Error: err.Error(),
                })
                mu.Unlock()
            } else {
                mu.Lock()
                successCount++
                mu.Unlock()
            }
        }(item)
    }
    
    wg.Wait()
    
    // 更新任务状态
    s.taskRepo.Update(ctx, event.TaskID, &BatchImportResult{
        Status:       "COMPLETED",
        SuccessCount: successCount,
        FailedCount:  len(failedItems),
        FailedItems:  failedItems,
        CompletedAt:  time.Now(),
    })
    
    return nil
}
```

**API推送**：

```go
// OpenAPI Service（供应商接口）
type OpenAPIService struct {
    listingService *ListingService
    rateLimiter    RateLimiter      // 限流器
    authService    AuthService      // 鉴权
}

// API推送商品
func (s *OpenAPIService) PushProduct(ctx context.Context, req *PushProductRequest) (*PushProductResponse, error) {
    // Step 1: 鉴权（API Key + Signature）
    supplierID, err := s.authService.Authenticate(ctx, req.APIKey, req.Signature)
    if err != nil {
        return nil, fmt.Errorf("鉴权失败: %w", err)
    }
    
    // Step 2: 限流（防止刷接口）
    if !s.rateLimiter.Allow(supplierID) {
        return nil, fmt.Errorf("请求过于频繁，请稍后重试")
    }
    
    // Step 3: 参数校验
    if err := s.validatePushRequest(req); err != nil {
        return nil, fmt.Errorf("参数校验失败: %w", err)
    }
    
    // Step 4: 创建上架任务（幂等性）
    task, err := s.listingService.CreateListingTask(ctx, &ListingRequest{
        ItemInfo:   req.ItemInfo,
        SupplierID: supplierID,
    })
    if err != nil {
        return nil, fmt.Errorf("创建任务失败: %w", err)
    }
    
    // Step 5: 自动提交审核（API推送默认进入审核）
    if err := s.listingService.SubmitForReview(ctx, task.TaskCode); err != nil {
        return nil, err
    }
    
    return &PushProductResponse{
        TaskCode: task.TaskCode,
        Status:   string(task.Status),
        Message:  "商品已提交审核，预计1-3小时内完成",
    }, nil
}
```

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **录入成功率** | > 95% | 按录入方式（手动/批量/API） |
| **批量导入耗时** | < 10s/千条 | 按文件大小 |
| **API响应时间** | P99 < 500ms | 按供应商 |
| **幂等性命中率** | < 5% | 按操作类型 |

---

#### 阶段2：审核发布

**业务场景**：
- **人工审核**：高风险商品（特定类目、新供应商）需人工审核
- **自动审核**：低风险商品通过规则引擎自动审核通过
- **审核驳回**：不合规商品驳回并通知修改

**技术难点**：
- **规则引擎**：多维度规则组合（合规性、完整性、准确性）
- **审核SLA**：自动审核秒级响应，人工审核小时级
- **审核日志**：完整记录审核决策，支持溯源

**自动审核规则引擎**：

```go
// 审核引擎
type ReviewEngine struct {
    rules []ReviewRule
}

// 审核规则接口
type ReviewRule interface {
    Check(ctx context.Context, task *ListingTask) *ReviewResult
}

// 自动审核
func (e *ReviewEngine) AutoReview(ctx context.Context, task *ListingTask) (*ReviewResult, error) {
    results := make([]*ReviewResult, 0, len(e.rules))
    
    // 执行所有规则
    for _, rule := range e.rules {
        result := rule.Check(ctx, task)
        results = append(results, result)
        
        // 任何规则拒绝则直接返回
        if result.Decision == ReviewReject {
            return result, nil
        }
    }
    
    // 所有规则通过
    return &ReviewResult{
        Decision: ReviewApprove,
        Score:    e.calculateScore(results),
    }, nil
}

// 规则1: 合规性检查
type ComplianceRule struct {
    sensitiveWords []string
    bannedCategories []int64
}

func (r *ComplianceRule) Check(ctx context.Context, task *ListingTask) *ReviewResult {
    // 违禁词检测
    for _, word := range r.sensitiveWords {
        if strings.Contains(task.ItemInfo.Title, word) || 
           strings.Contains(task.ItemInfo.Description, word) {
            return &ReviewResult{
                Decision: ReviewReject,
                Reason:   fmt.Sprintf("包含违禁词: %s", word),
            }
        }
    }
    
    // 禁售类目检测
    for _, category := range r.bannedCategories {
        if task.ItemInfo.CategoryID == category {
            return &ReviewResult{
                Decision: ReviewReject,
                Reason:   "该类目禁止销售",
            }
        }
    }
    
    return &ReviewResult{Decision: ReviewApprove}
}

// 规则2: 完整性检查
type CompletenessRule struct{}

func (r *CompletenessRule) Check(ctx context.Context, task *ListingTask) *ReviewResult {
    item := task.ItemInfo
    
    // 必填字段检查
    if item.Title == "" || item.Description == "" || item.BasePrice <= 0 {
        return &ReviewResult{
            Decision: ReviewReject,
            Reason:   "缺少必填字段",
        }
    }
    
    // 图片数量检查
    if len(item.Images) < 3 {
        return &ReviewResult{
            Decision: ReviewReject,
            Reason:   "商品图片不足3张",
        }
    }
    
    return &ReviewResult{Decision: ReviewApprove}
}

// 规则3: 准确性检查
type AccuracyRule struct{}

func (r *AccuracyRule) Check(ctx context.Context, task *ListingTask) *ReviewResult {
    item := task.ItemInfo
    
    // 价格合理性检查
    if item.BasePrice < 1 || item.BasePrice > 1000000 {
        return &ReviewResult{
            Decision: ReviewManual, // 转人工审核
            Reason:   "价格异常，需人工确认",
        }
    }
    
    // 类目匹配检查（示例：通过标题关键词）
    if !r.isCategoryMatched(item.Title, item.CategoryID) {
        return &ReviewResult{
            Decision: ReviewManual,
            Reason:   "类目与标题不匹配，需人工确认",
        }
    }
    
    return &ReviewResult{Decision: ReviewApprove}
}
```

**审核策略**：

| 审核维度 | 检查项 | 风险等级 | 处理策略 |
|---------|--------|---------|---------|
| **合规性** | 违禁词检测、敏感内容 | 高 | 自动拒绝 |
| **完整性** | 必填字段、图片数量 | 中 | 自动拒绝 |
| **准确性** | 价格合理性、类目匹配 | 中 | 转人工审核 |
| **一致性** | SPU/SKU关系、属性匹配 | 低 | 自动通过 |

**审核流程**：

```go
// 审核服务
func (s *ListingService) ProcessReview(ctx context.Context, taskCode string) error {
    task, _ := s.getTask(ctx, taskCode)
    
    // Step 1: 自动审核
    autoResult, err := s.reviewEngine.AutoReview(ctx, task)
    if err != nil {
        return err
    }
    
    switch autoResult.Decision {
    case ReviewApprove:
        // 自动通过 → 直接发布
        return s.Approve(ctx, taskCode, SystemReviewerID)
        
    case ReviewReject:
        // 自动拒绝 → 驳回
        return s.Reject(ctx, taskCode, autoResult.Reason)
        
    case ReviewManual:
        // 转人工审核 → 进入审核队列
        return s.assignToReviewer(ctx, taskCode)
    }
    
    return nil
}
```

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **审核通过率** | > 85% | 按供应商、按类目 |
| **自动审核占比** | > 70% | 按审核决策 |
| **人工审核SLA** | < 4h | P99耗时 |
| **审核驳回率** | < 15% | 按驳回原因 |

---

#### 阶段3：供应商同步

**业务场景**：
- 供应商定时推送商品数据（每小时/每天）
- 供应商实时推送价格/库存变更
- 供应商商品可能已存在，也可能不存在

**核心挑战**：Upsert语义

```
如果商品存在 → 更新
如果商品不存在 → 创建
```

**实现方案**：

```go
// 供应商同步任务
type SyncTask struct {
    SyncID       string    // 同步批次ID
    SupplierID   int64     // 供应商ID
    SupplierSkuID string   // 供应商SKU ID
    SyncData     SyncData  // 同步数据
    SyncType     string    // FULL/INCREMENTAL
    Status       string    // PENDING/SUCCESS/FAILED
}

// Upsert处理（幂等性保证）
func (s *SyncService) UpsertProduct(ctx context.Context, req *SyncRequest) error {
    // Step 1: 根据供应商SKU ID查询平台商品ID
    mapping, err := s.repo.GetMapping(ctx, req.SupplierID, req.SupplierSkuID)
    
    if err == ErrNotFound {
        // 场景1：商品不存在 → 创建（走上架流程）
        return s.createNewProduct(ctx, req)
    } else {
        // 场景2：商品存在 → 更新（走同步流程）
        return s.updateExistingProduct(ctx, mapping.ItemID, req)
    }
}

// 创建新商品（供应商同步触发的上架）
func (s *SyncService) createNewProduct(ctx context.Context, req *SyncRequest) error {
    // Step 1: 创建上架任务
    task, err := s.listingService.CreateListingTask(ctx, &ListingRequest{
        ItemInfo:   transformToItemInfo(req.SyncData),
        SupplierID: req.SupplierID,
        Source:     "SUPPLIER_SYNC",  // 标记来源
    })
    
    // Step 2: 根据供应商信用等级，决定是否需要审核
    if s.needReview(req.SupplierID) {
        // 低信用供应商：需要人工审核
        task.Status = ListingPending
    } else {
        // 高信用供应商：自动通过
        task.Status = ListingApproved
        s.listingService.Approve(ctx, task.TaskCode, SYSTEM_REVIEWER_ID)
    }
    
    return nil
}

// 更新现有商品（供应商同步）
func (s *SyncService) updateExistingProduct(ctx context.Context, itemID int64, req *SyncRequest) error {
    // Step 1: 对比差异
    existing, _ := s.productCenter.GetProduct(ctx, itemID)
    diff := s.compareDiff(existing, req.SyncData)
    
    // Step 2: 根据差异类型决定是否需要审核
    if diff.HasHighRiskChange() {
        // 高风险变更（价格变化>50%、类目变更）→ 需要审核
        return s.createReviewTask(ctx, itemID, diff)
    } else {
        // 低风险变更（库存、图片）→ 直接更新
        return s.productCenter.UpdateProduct(ctx, itemID, diff)
    }
}

// 判断供应商是否需要审核
func (s *SyncService) needReview(supplierID int64) bool {
    supplier, _ := s.supplierRepo.Get(ctx, supplierID)
    
    // 根据供应商信用等级和历史表现决定
    return supplier.CreditLevel < 3 || supplier.RejectRate > 0.1
}
```

**差异化审核策略**：

| 变更类型 | 变更范围 | 审核策略 | 理由 |
|---------|---------|---------|------|
| **价格变更** | < 10% | 自动通过 | 正常波动 |
| **价格变更** | 10-50% | 需要审核 | 防止错误 |
| **价格变更** | > 50% | 必须审核 + 告警 | 高风险 |
| **库存变更** | 任意 | 自动通过 | 实时性要求高 |
| **标题变更** | 轻微修改 | 自动通过 | 低风险 |
| **类目变更** | 任意 | 必须审核 | 影响搜索 |
| **图片变更** | 任意 | 自动通过 | 低风险 |

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **同步成功率** | > 98% | 按供应商、按同步类型 |
| **同步耗时** | P99 < 5s | 按数据大小 |
| **差异化审核命中率** | 10-20% | 按变更类型 |
| **供应商数据质量** | 错误率 < 5% | 按供应商 |

---

#### 阶段4：库存管理

**业务场景**：
- **实时同步**：热卖商品库存通过WebHook实时推送（减库存事件）
- **定时同步**：长尾商品库存通过定时任务批量拉取（每小时/每天）
- **水位监控**：库存低于阈值时触发告警，通知供应商补货
- **日终对账**：每日对账供应商库存与平台库存，发现差异自动修正

**技术难点**：
- **同步策略**：实时 vs 定时的平衡（成本 vs 准确性）
- **库存准确性**：多方数据源（供应商、订单系统、售后退款）一致性
- **并发控制**：高并发扣减库存时的原子性保证
- **对账修正**：发现差异后的自动修正 vs 人工介入

**核心设计**：

```go
// 库存同步策略（策略模式）
type StockSyncStrategy interface {
    Sync(ctx context.Context, skuID int64) (*SyncResult, error)
}

// 实时同步策略（高价值商品）
type RealtimeStockSyncStrategy struct {
    supplierClient SupplierClient
    inventoryRepo  InventoryRepository
    cache          *redis.Client
}

func (s *RealtimeStockSyncStrategy) Sync(ctx context.Context, skuID int64) (*SyncResult, error) {
    // Step 1: 调用供应商API实时查询库存
    supplierStock, err := s.supplierClient.GetStock(ctx, skuID)
    if err != nil {
        return nil, fmt.Errorf("供应商库存查询失败: %w", err)
    }
    
    // Step 2: 更新库存（数据库 + 缓存）
    if err := s.inventoryRepo.Update(ctx, skuID, supplierStock); err != nil {
        return nil, err
    }
    
    // Step 3: 更新缓存（防止穿透）
    s.cache.Set(ctx, fmt.Sprintf("stock:%d", skuID), supplierStock, 5*time.Minute)
    
    return &SyncResult{
        SKUID:         skuID,
        OldStock:      0, // 旧库存
        NewStock:      supplierStock,
        SyncTime:      time.Now(),
        SyncType:      "REALTIME",
    }, nil
}

// 定时同步策略（长尾商品）
type ScheduledStockSyncStrategy struct {
    supplierClient SupplierClient
    inventoryRepo  InventoryRepository
}

func (s *ScheduledStockSyncStrategy) Sync(ctx context.Context, skuID int64) (*SyncResult, error) {
    // Step 1: 批量拉取供应商库存（减少API调用）
    supplierStocks, err := s.supplierClient.BatchGetStock(ctx, []int64{skuID})
    if err != nil {
        return nil, err
    }
    
    // Step 2: 批量更新数据库
    updates := make(map[int64]int32)
    for _, stock := range supplierStocks {
        updates[stock.SKUID] = stock.Quantity
    }
    
    if err := s.inventoryRepo.BatchUpdate(ctx, updates); err != nil {
        return nil, err
    }
    
    return &SyncResult{
        SKUID:    skuID,
        NewStock: supplierStocks[skuID].Quantity,
        SyncTime: time.Now(),
        SyncType: "SCHEDULED",
    }, nil
}

// 库存同步服务（根据商品分级选择策略）
type StockSyncService struct {
    realtimeStrategy  *RealtimeStockSyncStrategy
    scheduledStrategy *ScheduledStockSyncStrategy
    productRepo       ProductRepository
}

func (s *StockSyncService) SyncStock(ctx context.Context, skuID int64) error {
    // 根据商品热度选择同步策略
    product, _ := s.productRepo.Get(ctx, skuID)
    
    var strategy StockSyncStrategy
    if product.Hotness > 80 { // 热卖商品
        strategy = s.realtimeStrategy
    } else {
        strategy = s.scheduledStrategy
    }
    
    _, err := strategy.Sync(ctx, skuID)
    return err
}
```

**库存水位监控**：

```go
// 库存水位监控器
type StockWatermarkMonitor struct {
    inventoryRepo InventoryRepository
    alertService  AlertService
    supplierClient SupplierClient
}

// 检查库存水位（定时任务，每5分钟执行）
func (m *StockWatermarkMonitor) CheckWatermark(ctx context.Context, skuID int64) error {
    // Step 1: 查询当前库存
    stock, err := m.inventoryRepo.GetStock(ctx, skuID)
    if err != nil {
        return err
    }
    
    // Step 2: 计算水位线（根据历史销量）
    watermark := m.calculateWatermark(ctx, skuID)
    
    // Step 3: 库存低于水位线 → 告警
    if stock.Available < watermark {
        m.alertService.Send(ctx, &Alert{
            Level:   "WARNING",
            Type:    "LOW_STOCK",
            SKUID:   skuID,
            Message: fmt.Sprintf("库存低于水位线（当前: %d, 水位: %d）", stock.Available, watermark),
        })
        
        // Step 4: 通知供应商补货
        m.supplierClient.RequestReplenishment(ctx, &ReplenishmentRequest{
            SKUID:        skuID,
            RequestQty:   watermark * 2, // 建议补货量
            UrgencyLevel: "NORMAL",
        })
    }
    
    return nil
}

// 计算水位线（基于历史销量）
func (m *StockWatermarkMonitor) calculateWatermark(ctx context.Context, skuID int64) int32 {
    // 查询最近7天日均销量
    avgDailySales := m.inventoryRepo.GetAvgDailySales(ctx, skuID, 7)
    
    // 水位线 = 3天销量（安全库存）
    return avgDailySales * 3
}
```

**库存对账**：

```go
// 库存对账任务（每日凌晨执行）
type StockReconciliationJob struct {
    inventoryRepo  InventoryRepository
    supplierClient SupplierClient
    diffRepo       DiffRepository
}

func (j *StockReconciliationJob) Run(ctx context.Context, date time.Time) error {
    // Step 1: 批量拉取所有SKU的供应商库存
    supplierStocks, err := j.supplierClient.GetAllStocks(ctx)
    if err != nil {
        return err
    }
    
    // Step 2: 批量查询平台库存
    platformStocks, err := j.inventoryRepo.GetAllStocks(ctx)
    if err != nil {
        return err
    }
    
    // Step 3: 对比差异
    diffs := make([]*StockDiff, 0)
    for skuID, supplierQty := range supplierStocks {
        platformQty := platformStocks[skuID]
        
        if supplierQty != platformQty {
            diff := &StockDiff{
                SKUID:        skuID,
                SupplierQty:  supplierQty,
                PlatformQty:  platformQty,
                Difference:   supplierQty - platformQty,
                ReconcileDate: date,
            }
            diffs = append(diffs, diff)
        }
    }
    
    // Step 4: 记录差异
    if err := j.diffRepo.BatchSave(ctx, diffs); err != nil {
        return err
    }
    
    // Step 5: 自动修正（差异 < 10% 自动修正，> 10% 人工介入）
    for _, diff := range diffs {
        if math.Abs(float64(diff.Difference)/float64(diff.PlatformQty)) < 0.1 {
            // 小差异：自动修正
            j.inventoryRepo.Update(ctx, diff.SKUID, diff.SupplierQty)
        } else {
            // 大差异：告警 + 人工介入
            j.alertService.Send(ctx, &Alert{
                Level:   "ERROR",
                Type:    "STOCK_MISMATCH",
                SKUID:   diff.SKUID,
                Message: fmt.Sprintf("库存差异过大（供应商: %d, 平台: %d）", diff.SupplierQty, diff.PlatformQty),
            })
        }
    }
    
    return nil
}
```

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **库存准确率** | > 99.5% | 按SKU、按供应商 |
| **实时同步成功率** | > 98% | 按供应商API可用性 |
| **定时同步耗时** | < 10min/批次 | 按SKU数量 |
| **水位告警响应时间** | < 5min | P99 |
| **日终对账差异率** | < 2% | 按供应商 |
| **自动修正覆盖率** | > 80% | 按差异范围 |

---

#### 阶段5：日常维护

**业务场景**：
- 单品编辑（修改标题、描述、图片）
- 批量编辑（批量调价、批量上下架）
- 批量导入导出（Excel操作）

**核心设计**：

```go
// 运营编辑任务
type EditTask struct {
    TaskID      string       // 任务ID
    ItemIDs     []int64      // 商品ID列表（支持批量）
    EditType    string       // SINGLE/BATCH
    Changes     []Change     // 变更内容
    Status      string       // PENDING/EXECUTING/SUCCESS/FAILED
    Progress    int          // 进度（0-100）
    TotalCount  int          // 总数
    SuccessCount int         // 成功数
    FailedCount int          // 失败数
}

// 批量编辑（异步任务）
func (s *EditService) BatchEdit(ctx context.Context, req *BatchEditRequest) (*EditTask, error) {
    // Step 1: 创建批量编辑任务
    task := &EditTask{
        TaskID:     generateTaskID(),
        ItemIDs:    req.ItemIDs,
        EditType:   "BATCH",
        Changes:    req.Changes,
        Status:     "PENDING",
        TotalCount: len(req.ItemIDs),
    }
    s.taskRepo.Save(ctx, task)
    
    // Step 2: 发布异步任务
    s.taskQueue.Publish(ctx, &BatchEditTaskEvent{
        TaskID: task.TaskID,
    })
    
    return task, nil
}

// 批量编辑执行器（异步）
func (w *BatchEditWorker) Execute(ctx context.Context, taskID string) error {
    task, _ := w.taskRepo.Get(ctx, taskID)
    
    // 逐个处理商品
    for i, itemID := range task.ItemIDs {
        err := w.editSingleItem(ctx, itemID, task.Changes)
        
        if err == nil {
            task.SuccessCount++
        } else {
            task.FailedCount++
            log.Errorf("edit item %d failed: %v", itemID, err)
        }
        
        // 更新进度
        task.Progress = (i + 1) * 100 / task.TotalCount
        w.taskRepo.Update(ctx, task)
    }
    
    // 更新任务状态
    if task.FailedCount == 0 {
        task.Status = "SUCCESS"
    } else if task.SuccessCount == 0 {
        task.Status = "FAILED"
    } else {
        task.Status = "PARTIAL_SUCCESS"
    }
    
    return nil
}
```

**进度追踪**：

```go
// 查询任务进度
func (s *EditService) GetTaskProgress(ctx context.Context, taskID string) (*TaskProgress, error) {
    task, _ := s.taskRepo.Get(ctx, taskID)
    
    return &TaskProgress{
        TaskID:       task.TaskID,
        Status:       task.Status,
        Progress:     task.Progress,
        TotalCount:   task.TotalCount,
        SuccessCount: task.SuccessCount,
        FailedCount:  task.FailedCount,
        EstimateLeft: s.estimateTimeLeft(task),
    }, nil
}
```

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **编辑成功率** | > 99% | 按编辑类型（单品/批量） |
| **批量编辑耗时** | < 5s/千条 | 按数据大小 |
| **进度更新频率** | 每秒 | 任务执行期间 |
| **部分成功占比** | < 10% | 按失败原因 |

---

#### 阶段6：促销配置

**业务场景**：
- **活动关联**：将商品关联到大促活动（618、双11）
- **价格设置**：配置活动价、满减、折扣券
- **定时生效**：活动开始时自动生效，结束时自动失效
- **价格验证**：确保活动价 < 原价，防止虚假促销

**技术难点**：
- **定时生效**：活动开始/结束时间精确到秒，需要定时任务支持
- **价格一致性**：促销价变更后需同步到商品中心、搜索、缓存
- **并发控制**：大促开始时大量商品同时生效，避免雪崩

**核心设计**：

```go
// 促销配置服务
type PromotionConfigService struct {
    productRepo    ProductRepository
    pricingClient  PricingClient
    promotionRepo  PromotionRepository
    cache          *redis.Client
}

// 配置促销（运营人员）
func (s *PromotionConfigService) ConfigPromotion(ctx context.Context, req *ConfigPromotionRequest) error {
    // Step 1: 参数校验
    if err := s.validatePromotionConfig(req); err != nil {
        return fmt.Errorf("配置校验失败: %w", err)
    }
    
    // Step 2: 价格验证（活动价 < 原价）
    product, _ := s.productRepo.Get(ctx, req.SKUID)
    if req.PromotionPrice >= product.BasePrice {
        return fmt.Errorf("促销价必须低于原价")
    }
    
    // Step 3: 创建促销配置
    promotionConfig := &PromotionConfig{
        ConfigID:       generateConfigID(),
        SKUID:          req.SKUID,
        ActivityID:     req.ActivityID,
        PromotionPrice: req.PromotionPrice,
        PromotionType:  req.PromotionType, // DISCOUNT/COUPON/FULL_REDUCTION
        StartTime:      req.StartTime,
        EndTime:        req.EndTime,
        Status:         "PENDING", // 待生效
        CreatedBy:      req.OperatorID,
    }
    
    if err := s.promotionRepo.Save(ctx, promotionConfig); err != nil {
        return err
    }
    
    // Step 4: 注册定时任务（生效/失效）
    s.scheduleActivation(ctx, promotionConfig)
    s.scheduleDeactivation(ctx, promotionConfig)
    
    return nil
}

// 批量配置促销（大促场景）
func (s *PromotionConfigService) BatchConfigPromotion(ctx context.Context, req *BatchConfigRequest) (*BatchConfigTask, error) {
    // Step 1: 创建批量任务
    task := &BatchConfigTask{
        TaskID:     generateTaskID(),
        ActivityID: req.ActivityID,
        TotalCount: len(req.Configs),
        Status:     "PENDING",
    }
    
    s.taskRepo.Save(ctx, task)
    
    // Step 2: 异步处理
    s.taskQueue.Publish(ctx, &BatchConfigEvent{
        TaskID:  task.TaskID,
        Configs: req.Configs,
    })
    
    return task, nil
}

// 定时任务：促销生效
type PromotionActivationJob struct {
    promotionRepo  PromotionRepository
    pricingClient  PricingClient
    searchClient   SearchClient
    cache          *redis.Client
}

func (j *PromotionActivationJob) Run(ctx context.Context) {
    // Step 1: 查询即将生效的促销（未来5分钟）
    now := time.Now()
    upcoming := j.promotionRepo.FindUpcoming(ctx, now, now.Add(5*time.Minute))
    
    // Step 2: 逐个生效
    for _, config := range upcoming {
        if time.Now().After(config.StartTime) {
            j.activatePromotion(ctx, config)
        }
    }
}

func (j *PromotionActivationJob) activatePromotion(ctx context.Context, config *PromotionConfig) error {
    // Step 1: 更新价格服务（促销价生效）
    if err := j.pricingClient.UpdatePromotionPrice(ctx, &UpdatePriceRequest{
        SKUID:          config.SKUID,
        PromotionPrice: config.PromotionPrice,
        ValidUntil:     config.EndTime,
    }); err != nil {
        return err
    }
    
    // Step 2: 更新搜索索引（展示促销标签）
    j.searchClient.UpdatePromotionTag(ctx, config.SKUID, config.ActivityID)
    
    // Step 3: 清理缓存（强制刷新）
    j.cache.Del(ctx, fmt.Sprintf("product:%d", config.SKUID))
    j.cache.Del(ctx, fmt.Sprintf("price:%d", config.SKUID))
    
    // Step 4: 更新促销配置状态
    config.Status = "ACTIVE"
    j.promotionRepo.Update(ctx, config)
    
    // Step 5: 发布促销生效事件
    j.eventPublisher.Publish(ctx, &PromotionActivatedEvent{
        SKUID:      config.SKUID,
        ActivityID: config.ActivityID,
        ActiveTime: time.Now(),
    })
    
    return nil
}

// 定时任务：促销失效
type PromotionDeactivationJob struct {
    promotionRepo  PromotionRepository
    pricingClient  PricingClient
    searchClient   SearchClient
    cache          *redis.Client
}

func (j *PromotionDeactivationJob) Run(ctx context.Context) {
    // 查询已过期的促销
    expired := j.promotionRepo.FindExpired(ctx, time.Now())
    
    for _, config := range expired {
        j.deactivatePromotion(ctx, config)
    }
}

func (j *PromotionDeactivationJob) deactivatePromotion(ctx context.Context, config *PromotionConfig) error {
    // Step 1: 恢复原价
    j.pricingClient.RestoreOriginalPrice(ctx, config.SKUID)
    
    // Step 2: 移除促销标签
    j.searchClient.RemovePromotionTag(ctx, config.SKUID)
    
    // Step 3: 清理缓存
    j.cache.Del(ctx, fmt.Sprintf("product:%d", config.SKUID))
    j.cache.Del(ctx, fmt.Sprintf("price:%d", config.SKUID))
    
    // Step 4: 更新状态
    config.Status = "EXPIRED"
    j.promotionRepo.Update(ctx, config)
    
    return nil
}
```

**价格验证策略**：

```go
// 价格验证器
type PriceValidator struct {
    productRepo ProductRepository
}

func (v *PriceValidator) Validate(ctx context.Context, req *ConfigPromotionRequest) error {
    product, _ := v.productRepo.Get(ctx, req.SKUID)
    
    // 规则1: 促销价 < 原价
    if req.PromotionPrice >= product.BasePrice {
        return fmt.Errorf("促销价必须低于原价")
    }
    
    // 规则2: 折扣不能过低（防止价格战）
    discount := float64(product.BasePrice-req.PromotionPrice) / float64(product.BasePrice)
    if discount > 0.7 {
        return fmt.Errorf("折扣过大（> 70%%），需审批")
    }
    
    // 规则3: 价格必须为整数（避免定价错误）
    if req.PromotionPrice%100 != 0 {
        return fmt.Errorf("价格必须为整数（单位：分）")
    }
    
    return nil
}
```

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **生效准时率** | > 99.9% | 按活动、按SKU |
| **失效准时率** | > 99.9% | 按活动、按SKU |
| **价格一致性** | 100% | 商品中心、搜索、缓存 |
| **配置错误率** | < 1% | 按配置类型 |

---

#### 阶段7：下架归档

**业务场景**：
- **临时下架**：商品缺货、质量问题临时下架，后续可恢复
- **永久下架**：商品停产、违规下架，不可恢复
- **订单检查**：下架前检查是否有进行中的订单，避免影响用户
- **历史归档**：永久下架商品归档到历史库，释放主库空间

**技术难点**：
- **订单安全**：下架前必须检查订单状态，防止影响履约
- **数据一致性**：下架需同步到商品中心、搜索、库存、价格
- **归档策略**：历史数据归档到冷存储，降低成本

**核心设计**：

```go
// 下架服务
type OffShelfService struct {
    productRepo   ProductRepository
    orderRepo     OrderRepository
    searchClient  SearchClient
    inventoryClient InventoryClient
    pricingClient PricingClient
}

// 下架商品
func (s *OffShelfService) OffShelf(ctx context.Context, req *OffShelfRequest) error {
    // Step 1: 检查进行中的订单
    activeOrders, err := s.orderRepo.FindActiveOrdersBySKU(ctx, req.SKUID)
    if err != nil {
        return err
    }
    
    if len(activeOrders) > 0 && req.OffShelfType == "PERMANENT" {
        return fmt.Errorf("存在进行中的订单（%d个），暂不能永久下架", len(activeOrders))
    }
    
    // Step 2: 更新商品状态
    product, _ := s.productRepo.Get(ctx, req.SKUID)
    
    if req.OffShelfType == "TEMPORARY" {
        // 临时下架 → OFF_SHELF
        product.Status = "OFF_SHELF"
        product.OffShelfReason = req.Reason
        product.OffShelfTime = time.Now()
    } else {
        // 永久下架 → ARCHIVED
        product.Status = "ARCHIVED"
        product.ArchivedReason = req.Reason
        product.ArchivedTime = time.Now()
    }
    
    s.productRepo.Update(ctx, product)
    
    // Step 3: 从搜索索引中移除
    s.searchClient.RemoveProduct(ctx, req.SKUID)
    
    // Step 4: 冻结库存（防止误售）
    s.inventoryClient.FreezeStock(ctx, req.SKUID)
    
    // Step 5: 移除促销配置
    s.pricingClient.RemovePromotions(ctx, req.SKUID)
    
    // Step 6: 发布下架事件
    s.eventPublisher.Publish(ctx, &ProductOffShelfEvent{
        SKUID:         req.SKUID,
        OffShelfType:  req.OffShelfType,
        Reason:        req.Reason,
        OffShelfTime:  time.Now(),
    })
    
    return nil
}

// 恢复上架（仅临时下架可恢复）
func (s *OffShelfService) RestoreOnShelf(ctx context.Context, skuID int64) error {
    product, _ := s.productRepo.Get(ctx, skuID)
    
    // 只有临时下架可恢复
    if product.Status != "OFF_SHELF" {
        return fmt.Errorf("商品状态不支持恢复（当前状态: %s）", product.Status)
    }
    
    // Step 1: 恢复商品状态
    product.Status = "ON_SHELF"
    s.productRepo.Update(ctx, product)
    
    // Step 2: 恢复搜索索引
    s.searchClient.AddProduct(ctx, product)
    
    // Step 3: 解冻库存
    s.inventoryClient.UnfreezeStock(ctx, skuID)
    
    return nil
}

// 归档服务（永久下架商品）
type ArchiveService struct {
    productRepo     ProductRepository
    archiveRepo     ArchiveRepository
    orderRepo       OrderRepository
    inventoryClient InventoryClient
}

// 归档商品（异步任务，每日凌晨执行）
func (s *ArchiveService) ArchiveProduct(ctx context.Context, skuID int64) error {
    product, _ := s.productRepo.Get(ctx, skuID)
    
    // 只归档永久下架的商品
    if product.Status != "ARCHIVED" {
        return nil
    }
    
    // Step 1: 检查是否有未完成的订单（防御性检查）
    activeOrders, _ := s.orderRepo.FindActiveOrdersBySKU(ctx, skuID)
    if len(activeOrders) > 0 {
        return fmt.Errorf("仍有进行中的订单，暂不归档")
    }
    
    // Step 2: 归档商品数据（写入历史库）
    archiveData := &ArchivedProduct{
        SKUID:        skuID,
        ProductData:  product.ToJSON(),
        ArchivedTime: time.Now(),
    }
    s.archiveRepo.Save(ctx, archiveData)
    
    // Step 3: 归档订单数据
    historicalOrders, _ := s.orderRepo.FindAllOrdersBySKU(ctx, skuID)
    for _, order := range historicalOrders {
        s.archiveRepo.SaveOrder(ctx, &ArchivedOrder{
            OrderID:      order.OrderID,
            SKUID:        skuID,
            OrderData:    order.ToJSON(),
            ArchivedTime: time.Now(),
        })
    }
    
    // Step 4: 删除主库数据（释放空间）
    s.productRepo.Delete(ctx, skuID)
    s.inventoryClient.DeleteStock(ctx, skuID)
    
    return nil
}
```

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **下架成功率** | > 99% | 按下架类型（临时/永久） |
| **订单冲突率** | < 1% | 永久下架前的订单检查 |
| **恢复成功率** | 100% | 临时下架恢复 |
| **归档耗时** | < 5s/商品 | 按数据大小 |
| **归档完整性** | 100% | 商品数据、订单数据 |

---

### 16.7.2 C端交易流完整链路

**交易流是电商的核心价值链**，从用户搜索到完成支付的完整路径。本节展示五个阶段的设计与集成。

**与B端链路的对比**：

| 维度 | B端链路（16.7.1） | C端链路（16.7.2） |
|-----|-----------------|-----------------|
| **参与方** | 供应商、运营、系统 | 用户、系统 |
| **时间跨度** | 数天到数月（商品生命周期） | 数分钟（单次购物流程） |
| **关键技术** | 幂等性、状态机、异步任务 | 聚合编排、快照、Saga |
| **核心关注** | 数据准确性、流程合规性 | 用户体验、转化率 |

---

#### 阶段1：搜索与导购

**业务场景**：用户搜索"iPhone 15"

**系统架构**：

```
用户输入关键词
    ↓
API Gateway → Search Aggregation
    ↓
Query理解（分词、纠错、意图识别）
    ↓
Elasticsearch召回（相关性排序）
    ↓
Hydrate编排（并发调用多个服务）
    ├─ Product Service（商品信息）
    ├─ Inventory Service（库存状态）
    ├─ Pricing Service（价格计算）
    └─ Marketing Service（活动标签）
    ↓
返回搜索结果
```

**核心代码**：

```go
// 搜索聚合服务
type SearchAggregation struct {
    esClient        *elasticsearch.Client
    productClient   rpc.ProductClient
    inventoryClient rpc.InventoryClient
    pricingClient   rpc.PricingClient
    marketingClient rpc.MarketingClient
}

func (a *SearchAggregation) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
    // Step 1: Query理解（分词、意图识别）
    query := a.parseQuery(req.Keyword)
    
    // Step 2: ES召回（按相关性排序）
    hits, err := a.esClient.Search(ctx, query)
    if err != nil {
        return nil, err
    }
    
    skuIDs := extractSkuIDs(hits)
    
    // Step 3: Hydrate编排（并发调用）
    var products map[int64]*Product
    var stocks map[int64]*Stock
    var prices map[int64]*Price
    var promos map[int64]*PromoInfo
    
    g, ctx := errgroup.WithContext(ctx)
    
    // 并发调用4个服务
    g.Go(func() error {
        products, _ = a.productClient.BatchGet(ctx, skuIDs)
        return nil
    })
    g.Go(func() error {
        stocks, _ = a.inventoryClient.BatchCheck(ctx, skuIDs)
        return nil
    })
    g.Go(func() error {
        priceItems := buildPriceItems(skuIDs)
        prices, _ = a.pricingClient.BatchCalculate(ctx, priceItems)
        return nil
    })
    g.Go(func() error {
        promos, _ = a.marketingClient.BatchGet(ctx, skuIDs, req.UserID)
        // 降级：Marketing故障时使用空促销
        if promos == nil {
            promos = make(map[int64]*PromoInfo)
        }
        return nil
    })
    
    g.Wait()
    
    // Step 4: 聚合结果
    return a.buildSearchResponse(hits, products, stocks, prices, promos), nil
}
```

**性能优化**：
- ES查询：P99 < 50ms
- Hydrate并发：4个服务并发调用，总耗时 < 200ms
- 缓存策略：热门搜索词缓存5分钟

---

#### 阶段2：商品详情页（PDP）

**业务场景**：用户点击商品进入详情页

**核心设计**：

```go
// 详情页聚合服务
func (a *DetailAggregation) GetDetail(ctx context.Context, skuID int64, userID int64) (*DetailResponse, error) {
    // 并发调用5个服务
    var product *Product
    var stock *Stock
    var price *Price
    var promos []*Promotion
    var reviews []*Review
    
    g, ctx := errgroup.WithContext(ctx)
    
    g.Go(func() error {
        product, _ = a.productClient.Get(ctx, skuID)
        return nil
    })
    g.Go(func() error {
        stock, _ = a.inventoryClient.Check(ctx, skuID)
        return nil
    })
    g.Go(func() error {
        price, _ = a.pricingClient.Calculate(ctx, skuID, userID)
        return nil
    })
    g.Go(func() error {
        promos, _ = a.marketingClient.GetPromotions(ctx, skuID, userID)
        return nil
    })
    g.Go(func() error {
        reviews, _ = a.reviewClient.GetTopReviews(ctx, skuID, 5)
        return nil
    })
    
    g.Wait()
    
    // 生成快照（用于后续试算）
    snapshot := a.generateSnapshot(product, price, promos)
    
    return &DetailResponse{
        Product:   product,
        Stock:     stock,
        Price:     price,
        Promos:    promos,
        Reviews:   reviews,
        Snapshot:  snapshot,  // 快照ID，5分钟有效
    }, nil
}
```

---

#### 阶段3：购物车

**业务场景**：用户加购商品

**未登录加购**：

```go
// 未登录用户（Cookie存储）
func (c *CartService) AddToCartAnonymous(ctx context.Context, req *AddCartRequest) error {
    // Step 1: 获取匿名cartID（存储在Cookie）
    cartID := req.AnonymousCartID
    if cartID == "" {
        cartID = generateCartID()
    }
    
    // Step 2: 存储到Redis（TTL=7天）
    cartKey := fmt.Sprintf("cart:anon:%s", cartID)
    cartData, _ := c.redis.Get(ctx, cartKey).Result()
    
    cart := parseCart(cartData)
    cart.AddItem(req.SkuID, req.Quantity)
    
    c.redis.Set(ctx, cartKey, marshal(cart), 7*24*time.Hour)
    
    return nil
}
```

**登录后合并**：

```go
// 用户登录后合并购物车
func (c *CartService) MergeCartOnLogin(ctx context.Context, userID int64, anonymousCartID string) error {
    // Step 1: 获取匿名购物车
    anonCartKey := fmt.Sprintf("cart:anon:%s", anonymousCartID)
    anonCart, _ := c.redis.Get(ctx, anonCartKey).Result()
    
    // Step 2: 获取用户购物车
    userCartKey := fmt.Sprintf("cart:user:%d", userID)
    userCart, _ := c.redis.Get(ctx, userCartKey).Result()
    
    // Step 3: 合并（相同商品累加数量）
    merged := mergeCarts(parseCart(anonCart), parseCart(userCart))
    
    // Step 4: 保存到用户购物车
    c.redis.Set(ctx, userCartKey, marshal(merged), 0)  // 永久存储
    
    // Step 5: 删除匿名购物车
    c.redis.Del(ctx, anonCartKey)
    
    // Step 6: 异步持久化到MySQL（防止Redis丢失）
    c.eventPublisher.Publish(ctx, &CartMergedEvent{
        UserID: userID,
        Items:  merged.Items,
    })
    
    return nil
}
```

---

#### 阶段4：结算页试算

**业务场景**：用户点击"去结算"

**核心设计**：

```go
// 结算页聚合服务
func (a *CheckoutAggregation) Calculate(ctx context.Context, req *CalculateRequest) (*CalculateResponse, error) {
    // Step 1: 判断是否使用快照（ADR-008）
    var products []*Product
    var promos []*Promotion
    
    if req.Snapshot != nil && !req.Snapshot.IsExpired() {
        // 快照未过期，使用快照数据（性能优先）
        products = req.Snapshot.Products
        promos = req.Snapshot.Promos
    } else {
        // 快照过期，实时查询
        products, _ = a.productClient.BatchGet(ctx, req.SkuIDs)
        promos, _ = a.marketingClient.GetPromotions(ctx, req.SkuIDs, req.UserID)
    }
    
    // Step 2: 实时查询库存（不能用快照）
    stocks, _ := a.inventoryClient.BatchCheck(ctx, req.SkuIDs)
    
    // Step 3: 计算价格
    prices, _ := a.pricingClient.BatchCalculate(ctx, products, promos)
    
    // Step 4: 检查可下单性
    canCheckout := a.checkCanCheckout(stocks, req.Items)
    
    return &CalculateResponse{
        Items:       buildItems(products, stocks, prices),
        TotalPrice:  calculateTotal(prices),
        CanCheckout: canCheckout,
        Warnings:    a.generateWarnings(stocks, promos),
    }, nil
}
```

---

#### 阶段5：下单与支付

**完整下单流程**（Saga模式）：

```go
// 订单创建Saga（编排多个服务调用）
type CreateOrderSaga struct {
    productClient   rpc.ProductClient
    inventoryClient rpc.InventoryClient
    pricingClient   rpc.PricingClient
    marketingClient rpc.MarketingClient
    orderRepo       *OrderRepo
}

func (s *CreateOrderSaga) Execute(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    var err error
    var reserved *ReserveResult
    var couponLock *CouponLock
    
    // Step 1: 实时查询商品信息（ADR-009：不使用快照）
    products, err := s.productClient.BatchGet(ctx, req.SkuIDs)
    if err != nil {
        return nil, fmt.Errorf("query products failed: %w", err)
    }
    
    // Step 2: 实时查询营销信息
    promos, err := s.marketingClient.GetPromotions(ctx, req.SkuIDs, req.UserID)
    if err != nil {
        return nil, fmt.Errorf("query promotions failed: %w", err)
    }
    
    // Step 3: 校验营销活动有效性
    for _, promo := range promos {
        if !s.validatePromotion(promo) {
            return nil, fmt.Errorf("promotion %s expired", promo.ID)
        }
    }
    
    // Step 4: 库存预占（CAS操作）
    reserved, err = s.inventoryClient.Reserve(ctx, req.Items)
    if err != nil {
        return nil, fmt.Errorf("库存不足: %w", err)
    }
    defer func() {
        if err != nil {
            // 补偿：释放库存
            s.inventoryClient.Release(ctx, reserved.ReserveID)
        }
    }()
    
    // Step 5: 优惠券锁定
    if req.CouponCode != "" {
        couponLock, err = s.marketingClient.LockCoupon(ctx, req.CouponCode, req.UserID)
        if err != nil {
            return nil, fmt.Errorf("优惠券锁定失败: %w", err)
        }
        defer func() {
            if err != nil {
                // 补偿：释放优惠券
                s.marketingClient.UnlockCoupon(ctx, couponLock.LockID)
            }
        }()
    }
    
    // Step 6: 实时计算价格
    price, err := s.pricingClient.Calculate(ctx, products, promos)
    if err != nil {
        return nil, fmt.Errorf("价格计算失败: %w", err)
    }
    
    // Step 7: 价格校验（ADR-011）
    if req.ExpectedPrice > 0 {
        if err := s.validatePriceChange(req.ExpectedPrice, price.FinalPrice); err != nil {
            return nil, err
        }
    }
    
    // Step 8: 生成商品快照
    snapshot := s.generateProductSnapshot(products, promos, price)
    
    // Step 9: 创建订单
    order := &Order{
        OrderID:         s.generateOrderID(),
        UserID:          req.UserID,
        Items:           req.Items,
        TotalPrice:      price.FinalPrice,
        ProductSnapshot: marshal(snapshot),
        ReserveID:       reserved.ReserveID,
        CouponLockID:    couponLock.LockID,
        Status:          StatusPendingPayment,
        ExpireTime:      time.Now().Add(15 * time.Minute),
    }
    
    err = s.orderRepo.Create(ctx, order)
    if err != nil {
        return nil, fmt.Errorf("订单创建失败: %w", err)
    }
    
    // Step 10: 发布订单创建事件（异步）
    s.eventPublisher.Publish(ctx, &OrderCreatedEvent{
        OrderID: order.OrderID,
        UserID:  order.UserID,
        Items:   order.Items,
    })
    
    return order, nil
}
```

**交易流监控**：

| 阶段 | 关键指标 | 目标值 |
|------|---------|--------|
| **搜索** | 搜索→点击转化率 | > 15% |
| **详情页** | 详情→加购转化率 | > 8% |
| **购物车** | 加购→结算转化率 | > 30% |
| **结算页** | 结算→下单转化率 | > 60% |
| **支付** | 下单→支付转化率 | > 85% |
| **整体** | 搜索→支付转化率 | > 2% |

---

## 16.8 DDD战术设计实践

**领域模型是系统设计的核心**。本节展示如何在订单域应用DDD战术模式。

### 聚合设计：Order聚合根

```go
// Order聚合根
type Order struct {
    // 聚合根ID
    orderID OrderID  // 值对象
    
    // 基本信息
    userID    int64
    shopID    int64
    
    // 订单明细（实体集合）
    items []*OrderItem
    
    // 价格信息（值对象）
    pricing *OrderPricing
    
    // 状态（值对象）
    status OrderStatus
    
    // 时间戳
    createdAt time.Time
    updatedAt time.Time
    
    // 领域事件（未提交）
    domainEvents []DomainEvent
}

// 值对象：OrderID
type OrderID struct {
    value string
}

func NewOrderID() OrderID {
    return OrderID{value: generateSnowflakeID()}
}

func (id OrderID) String() string {
    return id.value
}

// 值对象：OrderPricing
type OrderPricing struct {
    subtotal       int64  // 商品总价
    discount       int64  // 折扣金额
    couponDiscount int64  // 优惠券
    payableAmount  int64  // 应付金额
}

func (p *OrderPricing) Calculate() int64 {
    return p.subtotal - p.discount - p.couponDiscount
}

// 实体：OrderItem
type OrderItem struct {
    itemID    int64
    skuID     int64
    quantity  int
    unitPrice int64
    
    // 快照
    snapshot *ItemSnapshot
}

// 聚合根方法：状态转换
func (o *Order) TransitionTo(newStatus OrderStatus) error {
    // 检查状态转换是否合法
    if !o.status.CanTransitionTo(newStatus) {
        return fmt.Errorf("不允许从 %s 转换到 %s", o.status, newStatus)
    }
    
    oldStatus := o.status
    o.status = newStatus
    o.updatedAt = time.Now()
    
    // 发布领域事件
    o.addDomainEvent(&OrderStatusChangedEvent{
        OrderID:   o.orderID,
        OldStatus: oldStatus,
        NewStatus: newStatus,
        ChangedAt: o.updatedAt,
    })
    
    return nil
}

// 聚合根方法：添加商品项
func (o *Order) AddItem(item *OrderItem) error {
    // 不变量检查：订单金额不能超过限额
    if o.calculateTotal()+item.Total() > MAX_ORDER_AMOUNT {
        return errors.New("订单金额超过限额")
    }
    
    o.items = append(o.items, item)
    
    // 发布领域事件
    o.addDomainEvent(&OrderItemAddedEvent{
        OrderID: o.orderID,
        Item:    item,
    })
    
    return nil
}

// 不变量：订单金额 = 所有商品项之和
func (o *Order) calculateTotal() int64 {
    total := int64(0)
    for _, item := range o.items {
        total += item.Total()
    }
    return total
}

// 领域事件管理
func (o *Order) addDomainEvent(event DomainEvent) {
    o.domainEvents = append(o.domainEvents, event)
}

func (o *Order) DomainEvents() []DomainEvent {
    return o.domainEvents
}

func (o *Order) ClearDomainEvents() {
    o.domainEvents = nil
}
```

### Repository模式

```go
// OrderRepository接口（领域层定义）
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, orderID OrderID) (*Order, error)
    FindByUserID(ctx context.Context, userID int64, limit int) ([]*Order, error)
}

// OrderRepositoryImpl实现（基础设施层）
type OrderRepositoryImpl struct {
    db            *gorm.DB
    eventPublisher EventPublisher
}

func (r *OrderRepositoryImpl) Save(ctx context.Context, order *Order) error {
    // Step 1: 转换聚合根 → 数据模型
    orderDO := r.toDataObject(order)
    
    // Step 2: 保存到数据库
    err := r.db.Transaction(func(tx *gorm.DB) error {
        // 保存订单主表
        if err := tx.Create(orderDO).Error; err != nil {
            return err
        }
        
        // 保存订单明细表
        for _, item := range order.Items() {
            itemDO := r.toItemDataObject(item, orderDO.ID)
            if err := tx.Create(itemDO).Error; err != nil {
                return err
            }
        }
        
        return nil
    })
    
    if err != nil {
        return err
    }
    
    // Step 3: 发布领域事件（事务提交后）
    for _, event := range order.DomainEvents() {
        r.eventPublisher.Publish(ctx, event)
    }
    order.ClearDomainEvents()
    
    return nil
}
```

### 领域事件与Outbox模式

```go
// Outbox表（确保事件必达）
type Outbox struct {
    ID          int64
    EventType   string
    EventData   string  // JSON
    Status      string  // PENDING/PUBLISHED/FAILED
    RetryCount  int
    CreatedAt   time.Time
}

// 发布领域事件（Outbox模式）
func (p *EventPublisher) Publish(ctx context.Context, event DomainEvent) error {
    // Step 1: 序列化事件
    eventData, _ := json.Marshal(event)
    
    // Step 2: 写入Outbox表（与业务在同一事务）
    outbox := &Outbox{
        EventType: event.Type(),
        EventData: string(eventData),
        Status:    "PENDING",
        CreatedAt: time.Now(),
    }
    
    return p.db.Create(outbox).Error
}

// Outbox轮询器（定时扫描未发布的事件）
func (w *OutboxWorker) Run() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        // Step 1: 查询待发布事件（PENDING状态）
        var outboxes []*Outbox
        w.db.Where("status = ? AND retry_count < ?", "PENDING", 3).
            Limit(100).
            Find(&outboxes)
        
        // Step 2: 发布到Kafka
        for _, outbox := range outboxes {
            err := w.kafkaProducer.Send(outbox.EventType, outbox.EventData)
            
            if err == nil {
                // 发布成功，标记为PUBLISHED
                w.db.Model(outbox).Update("status", "PUBLISHED")
            } else {
                // 发布失败，重试计数+1
                w.db.Model(outbox).Updates(map[string]interface{}{
                    "retry_count": gorm.Expr("retry_count + 1"),
                    "status":      "FAILED",
                })
            }
        }
    }
}
```

---

## 16.9 架构决策记录（ADR）

本节记录系统设计过程中的关键架构决策，包括决策背景、备选方案、最终决策及理由。**ADR是架构演进的重要资产，帮助团队理解「为什么这样设计」，避免重复讨论。**

### ADR-001: 计价中心数据输入方式

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：计价中心需要营销信息（促销规则、优惠券等）来计算最终价格，有两种方案：
- 方案1：计价中心自己调用Marketing Service获取营销信息
- 方案2：聚合服务获取营销信息后传递给计价中心

**决策**：采用**方案2**，由聚合服务获取营销信息后传递给计价中心。

**理由**：

1. **单一职责原则（SRP）**：
   - Pricing Service专注于价格计算逻辑（纯函数）
   - Aggregation Service负责数据编排和获取
   - 职责边界清晰，符合微服务设计原则

2. **依赖解耦**：
   ```
   方案1依赖链：Aggregation → Pricing → Marketing（传递性依赖）
   方案2依赖链：Aggregation → Pricing | Marketing（平行依赖）✓
   ```

3. **性能优化空间更大**：
   - 聚合层可以并发调用Marketing和其他服务（Product、Inventory）
   - Pricing变成纯计算，无IO等待
   - 减少网络调用层级（2层 vs 3层）

4. **易于测试**：
   ```go
   // 方案2：Pricing是纯函数，测试简单
   func TestCalculatePrice(t *testing.T) {
       priceItem := &PriceCalculateItem{
           SkuID:     1001,
           BasePrice: 2399.00,
           PromoInfo: &PromoInfo{DiscountRate: 0.9},  // Mock数据
       }
       result := pricingService.Calculate(priceItem)
       assert.Equal(t, 2159.10, result.FinalPrice)
   }
   ```

5. **统一降级处理**：
   - 聚合层统一处理各服务失败（Marketing、Product、Inventory）
   - Pricing Service无感知，始终收到完整输入数据
   - 降级逻辑不混入业务计算

**代码示例**：

```go
// SearchOrchestrator（聚合服务）
func (o *SearchOrchestrator) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
    // Step 1: 获取sku_ids（从ES）
    skuIDs, _ := o.searchClient.QuerySkuIDs(ctx, req.Keyword)
    
    // Step 2: 并发调用Product + Inventory + Marketing
    var products []*Product
    var stocks []*Stock
    var promos map[int64]*PromoInfo
    
    g, ctx := errgroup.WithContext(ctx)
    g.Go(func() error {
        products, _ = o.productClient.BatchGet(ctx, skuIDs)
        return nil
    })
    g.Go(func() error {
        stocks, _ = o.inventoryClient.BatchCheck(ctx, skuIDs)
        return nil
    })
    g.Go(func() error {
        promos, _ = o.marketingClient.BatchGet(ctx, skuIDs, req.UserID)
        // 降级：Marketing故障时使用空促销
        if promos == nil {
            promos = make(map[int64]*PromoInfo)
        }
        return nil
    })
    g.Wait()
    
    // Step 3: 调用Pricing计算价格（传入营销信息）
    priceItems := buildPriceItems(products, promos)
    prices, _ := o.pricingClient.BatchCalculate(ctx, priceItems)
    
    return buildSearchResponse(products, stocks, prices), nil
}

// PricingService（计价中心）- 纯函数，只负责计算
func (s *PricingService) Calculate(item *PriceItem) *PriceResult {
    finalPrice := item.BasePrice
    
    // 应用促销折扣（数据来自聚合层）
    if item.PromoInfo != nil {
        finalPrice = finalPrice * item.PromoInfo.DiscountRate
    }
    
    return &PriceResult{
        OriginalPrice: item.BasePrice,
        FinalPrice:    finalPrice,
        Discount:      item.BasePrice - finalPrice,
    }
}
```

**影响范围**：
- Aggregation Service：增加Marketing Service调用
- Pricing Service：接收PromoInfo作为输入参数
- Marketing Service：无影响

---

### ADR-002: 库存预占时机

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：在下单流程中，库存预占的时机有两种选择：
- 方案1：结算试算时预占（早期锁定）
- 方案2：确认下单时预占（延迟锁定）

**决策**：采用**方案2**，在确认下单时预占库存。

**理由**：

1. **减少无效预占**：
   - 用户在试算阶段可能多次修改商品、数量、优惠券
   - 早期预占会导致大量无效锁定（用户未真正下单）
   - 试算到下单的转化率通常只有20-30%

2. **提升库存利用率**：
   - 避免库存被长时间预占（用户可能犹豫、放弃）
   - 预占时长控制在15分钟内（支付超时自动释放）

3. **降低系统压力**：
   - 试算接口QPS高（用户多次试算），预占会导致Redis压力大
   - 确认下单QPS相对较低，预占操作更可控

4. **用户体验**：
   - 试算快速返回（不需要等待预占操作）
   - 确认下单时再预占，用户心理准备更充分

**权衡**：
- ✓ 优点：提升库存利用率、减少无效预占、降低系统压力
- ✗ 缺点：确认下单时可能库存不足（需要前端提示）

**降低缺点的措施**：
- 试算时展示实时库存状态（"仅剩N件"）
- 确认下单时二次校验库存，失败友好提示
- 热门商品提前告知"库存紧张，请尽快下单"

---

### ADR-003: 聚合服务 vs BFF

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：在API Gateway和微服务之间，是使用BFF（Backend For Frontend）还是Aggregation Service？

**决策**：采用**Aggregation Service**，而不是传统BFF。

**理由**：

1. **业务导向 vs 端导向**：
   - BFF按端划分（Web BFF、App BFF、小程序 BFF）
   - Aggregation按业务场景划分（搜索聚合、详情聚合、结算聚合）✓
   - 本系统多个端（Web、App）的业务逻辑高度一致，按端拆分会导致重复代码

2. **代码复用**：
   ```
   BFF模式：
   ├─ Web BFF（搜索逻辑）
   ├─ App BFF（搜索逻辑）    ← 重复代码
   └─ 小程序 BFF（搜索逻辑） ← 重复代码
   
   Aggregation模式：✓
   ├─ Search Aggregation（Web/App/小程序共用）
   └─ Detail Aggregation（Web/App/小程序共用）
   ```

3. **维护成本**：
   - BFF需要维护多个端的代码一致性
   - Aggregation只需维护一套业务逻辑

4. **适配端差异的方式**：
   - API Gateway层处理端协议差异（HTTP、WebSocket、gRPC）
   - Aggregation返回标准数据格式，前端各端按需裁剪

**适用场景**：
- ✓ 多端业务逻辑高度一致（如本系统）
- ✗ 不适用：各端业务逻辑差异大（如社交产品，Feed流算法不同）

---

### ADR-004: 虚拟商品库存模型

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：虚拟商品（机票、充值卡、优惠券）的库存模型和实物商品差异大，应该如何设计？

**决策**：采用**二维库存模型**（ManagementType + UnitType）。

**库存管理类型（ManagementType）**：

| 类型 | 说明 | 典型品类 | 库存来源 |
|-----|------|---------|---------|
| **实时库存** | 强依赖供应商实时查询 | 机票、酒店 | 供应商API |
| **池化库存** | 自有库存，可超卖后补偿 | 充值卡、优惠券 | 平台采购 |
| **无限库存** | 虚拟商品，无库存限制 | SaaS服务、数字内容 | 无 |

**库存单位类型（UnitType）**：

| 类型 | 说明 | 典型品类 |
|-----|------|---------|
| **SKU级别** | 每个规格独立库存 | 充电器（颜色、规格） |
| **批次级别** | 按批次管理（有效期） | 优惠券、礼品卡 |
| **座位级别** | 唯一标识（座位号） | 机票、电影票 |

**理由**：
1. 不同品类的库存特性差异极大，无法用统一模型
2. 二维模型提供灵活性，支持策略模式动态选择
3. 便于扩展新品类（只需添加新策略）

---

### ADR-005: 同步 vs 异步数据流

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：下单流程中，哪些操作应该同步执行，哪些应该异步执行？

**决策**：采用**同步+异步混合模式**。

**同步操作（用户等待）**：
1. 库存预占（必须成功，否则无法下单）
2. 优惠券扣减（避免超发）
3. 订单创建（生成order_id）

**异步操作（Kafka事件）**：
1. 库存确认扣减（预占成功后，异步确认）
2. 搜索索引更新（销量、热度）
3. 购物车清理
4. 用户行为分析
5. 消息通知（订单确认、物流更新）

**理由**：

1. **用户体验**：
   - 同步操作<500ms，用户可接受
   - 非核心操作异步化，不阻塞下单

2. **系统解耦**：
   - 异步事件降低服务间强依赖
   - 消费者故障不影响下单流程

3. **性能优化**：
   - 减少下单接口响应时间
   - 异步操作可批量处理（提升吞吐）

4. **容错能力**：
   - 异步操作支持重试（Kafka消费者重试机制）
   - 同步操作失败可立即回滚（Saga模式）

---

### ADR-009: 创单时是否使用快照数据（核心安全决策）

**决策日期**：2026-04-15  
**状态**：已采纳 ✓

**问题描述**：用户从详情页到提交订单期间，前端已经缓存了商品信息、价格、活动等快照数据。在用户点击"提交订单"创建订单时，后端是否可以使用这些快照数据来提升性能，避免重复查询？

**备选方案**：

| 方案 | 描述 | 优点 | 缺点 |
|------|------|------|------|
| **方案A：使用快照** | 创单时直接使用前端传递的快照数据 | ✅ 性能好（无需查询）<br>✅ 响应快（200ms → 50ms） | ❌ 安全风险高（快照可能被篡改）<br>❌ 资损风险 |
| **方案B：强制实时查询** ✓ | 创单时强制调用商品服务、营销服务查询最新数据 | ✅ 数据绝对准确<br>✅ 安全性高（防篡改）<br>✅ 无资损风险 | ❌ 性能稍差（多次RPC调用）<br>❌ RT增加100-200ms |
| **方案C：混合模式** | 普通商品用快照，营销商品强制查询 | ⚠️ 复杂度高<br>⚠️ 容易出错 | ❌ 维护成本高<br>❌ 边界不清晰 |

**决策**：采用**方案B（强制实时查询）**

**决策理由**：

1. **安全性优先于性能**
   ```
   风险分析：
   - 如果用快照，活动结束但快照未更新 → 用户用秒杀价下单 → 资损
   - 如果用快照，用户篡改价格 → 恶意低价下单 → 资损
   - 性能损失：100-200ms
   - 资损风险：每单可能损失数百至数千元
   
   结论：100ms的性能代价 << 资损风险
   ```

2. **涉及资金的操作必须实时校验**
   ```
   创单 = 锁定库存 + 锁定价格 + 准备扣款
   → 必须基于最新、最准确的数据
   → 不能因为性能优化而妥协安全性
   ```

3. **防止恶意篡改**
   ```
   场景：黑产抓包修改快照数据
   快照：{"expected_payable": 799900}  // 原价 ¥7,999
   篡改：{"expected_payable": 1}       // 改成 ¥0.01
   
   如果后端使用快照：
   → 按 ¥0.01 创单 → 公司巨额损失！
   
   强制实时查询：
   → 后端查到实际价格 ¥7,999
   → 对比快照 ¥0.01 vs 实际 ¥7,999
   → 差异巨大，拒绝创单！
   ```

4. **活动可能随时变化**
   ```
   10:00  秒杀价 ¥7,999，生成快照
   10:04  秒杀活动提前结束（库存售罄）
   10:05  用户提交订单
   
   如果用快照：
   → 按 ¥7,999 创单（活动已结束！）
   → 资损
   
   强制查询：
   → 查到活动已结束，价格 ¥8,999
   → 提示用户价格变化
   → 避免资损
   ```

**实现方案**：

```go
// OrderService.CreateOrder - 确认下单接口（准确性优先）
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    // ⚠️ 关键：创单时不使用任何前端传递的快照数据，全部实时查询
    
    // Step 1: 实时查询商品信息（不使用前端快照）
    products, err := s.productClient.BatchGetProducts(ctx, req.SkuIDs)
    if err != nil {
        return nil, fmt.Errorf("query products failed: %w", err)
    }
    
    // Step 2: 实时查询营销活动（强制最新数据）
    promos, err := s.marketingClient.BatchGetPromotions(ctx, req.SkuIDs, req.UserID)
    if err != nil {
        return nil, fmt.Errorf("query promotions failed: %w", err)
    }
    
    // Step 3: 校验营销活动有效性（关键：防止使用过期活动）
    for _, promo := range promos {
        if !s.validatePromotion(promo) {
            return nil, fmt.Errorf("promotion %s is invalid or expired", promo.ID)
        }
    }
    
    // Step 4: 实时计算价格（基于最新营销数据）
    price, err := s.pricingClient.CalculateFinalPrice(ctx, products, promos)
    if err != nil {
        return nil, fmt.Errorf("calculate price failed: %w", err)
    }
    
    // Step 5: 价格校验（对比前端传递的期望价格）
    if req.ExpectedPrice > 0 {
        if err := s.validatePriceChange(req.ExpectedPrice, price.FinalPrice); err != nil {
            return nil, err  // 价格变化过大，拒绝创单
        }
    }
    
    // Step 6: 预占库存
    reserved, err := s.inventoryClient.ReserveStock(ctx, req.Items)
    if err != nil {
        return nil, fmt.Errorf("reserve stock failed: %w", err)
    }
    
    // Step 7: 生成商品快照（基于实时查询的数据）
    snapshot := s.generateProductSnapshot(products, promos, price)
    
    // Step 8: 创建订单（保存快照）
    order := &Order{
        OrderID:         s.generateOrderID(),
        UserID:          req.UserID,
        Items:           req.Items,
        TotalPrice:      price.FinalPrice,
        ProductSnapshot: marshal(snapshot),  // 💾 保存商品快照
        Status:          OrderStatusPendingPayment,
        ExpireTime:      time.Now().Add(15 * time.Minute),
        ReserveIDs:      reserved,
    }
    
    return s.orderRepo.Create(ctx, order)
}

// 价格校验逻辑（防止用户感知差）
func (s *OrderService) validatePriceChange(expected, actual int64) error {
    diff := actual - expected
    diffPercent := float64(diff) / float64(expected) * 100
    
    // 场景1: 价格降低 → 允许（对用户有利）
    if diff < 0 {
        return nil
    }
    
    // 场景2: 价格上涨 < 1元 → 允许（误差容忍）
    if diff <= 100 { // 100分 = 1元
        return nil
    }
    
    // 场景3: 价格上涨 >= 1元 且 < 5% → 允许但记录日志
    if diffPercent < 5.0 {
        log.Warnf("price increased: expected=%d, actual=%d", expected, actual)
        return nil
    }
    
    // 场景4: 价格上涨 >= 5% → 拒绝，要求用户重新确认
    return &PriceChangedError{
        Expected: expected,
        Actual:   actual,
        Message:  fmt.Sprintf("价格已变化，请重新确认"),
    }
}
```

**核心原则**：
```
┌────────────────────────────────────────────────────────┐
│ 试算阶段：性能优先 → 可用快照（5分钟缓存）              │
│ 创单阶段：准确性优先 → 强制实时查询                     │
│ 历史查询：可追溯性 → 保存快照到订单表                   │
└────────────────────────────────────────────────────────┘
```

---

### ADR-010: 创单与支付的时序关系

**决策日期**：2026-04-14  
**状态**：已采纳 ✓

**问题描述**：在订单流程中，"创建订单"和"支付"这两个动作的时序关系有两种模式：
1. **创单即支付**：用户点击"立即购买"后，先支付，支付成功后再创建订单
2. **先创单后支付**：用户点击"提交订单"后，先创建订单（资源扣减），然后再支付

**决策**：采用"先创单后支付"模式

**理由**：

**1. 防止超卖（关键）**：
```
【创单即支付模式的问题】：
1. 用户A看到库存=1
2. 用户B也看到库存=1
3. 用户A点击支付（此时库存未扣减）
4. 用户B也点击支付（库存仍未扣减）
5. 两人同时支付成功 → 超卖！

【先创单后支付模式的解决方案】：
1. 用户A点击"提交订单" → 库存预占：1 → 0（剩余可用）
2. 用户B点击"提交订单" → 库存不足，下单失败
3. 用户A有15分钟支付窗口
4. 如果用户A超时未支付 → 释放库存：0 → 1（其他人可下单）
```

**2. 用户体验更好**：
- ✅ 用户点击"提交订单"后，订单立即生成，库存被锁定
- ✅ 用户可以慢慢选择支付方式（支付宝、微信、银行卡）
- ✅ 用户可以在支付环节选择优惠券、支付渠道优惠
- ✅ 用户可以先下单占位，稍后再支付（适合机票、酒店）

**3. 价格计算灵活性**：
- 创单时计算：商品基础价格 + 营销优惠（折扣、满减）
- 支付时计算：支付渠道费（信用卡手续费、花呗分期费）+ 支付渠道优惠

**权衡**：

| 维度 | 优势 | 劣势 |
|-----|------|------|
| **用户体验** | ✅ 先锁定库存，再支付<br>✅ 支付环节更灵活 | ⚠️ 15分钟内库存被占用 |
| **防止超卖** | ✅ 创单时锁定库存（零超卖） | ⚠️ 需要处理超时释放逻辑 |
| **库存利用率** | ⚠️ 预占库存可能被浪费（10-20%未支付率） | ✅ 可通过缩短支付窗口优化 |
| **系统复杂度** | ⚠️ 需要库存预占机制<br>⚠️ 需要超时释放定时任务 | ⚠️ 状态机更复杂 |

**超时未支付处理**：

```go
// OrderTimeoutJob - 定时扫描超时未支付订单
func (j *OrderTimeoutJob) Run() {
    // 查询超时订单（创建时间 > 15分钟，状态=PENDING_PAYMENT）
    expiredOrders := j.orderRepo.FindExpiredPendingPayment(15 * time.Minute)
    
    for _, order := range expiredOrders {
        // 1. 更新订单状态：PENDING_PAYMENT → CANCELLED
        order.Status = OrderStatusCancelled
        order.CancelReason = "超时未支付"
        j.orderRepo.Update(ctx, order)
        
        // 2. 释放库存
        j.inventoryClient.ReleaseStock(ctx, order.ReserveIDs)
        
        // 3. 回退优惠券
        if order.CouponID != "" {
            j.marketingClient.ReleaseCoupon(ctx, order.CouponID, order.UserID)
        }
        
        // 4. 发布订单取消事件
        j.eventPublisher.Publish(ctx, &OrderCancelledEvent{
            OrderID: order.OrderID,
            Reason:  "超时未支付",
        })
    }
}
```

---

### ADR-011: 创单时前后端价格校验策略

**决策日期**：2026-04-15  
**状态**：已采纳 ✓

**问题描述**：创单时后端实时查询得到的价格，可能和前端展示的价格不一致（活动变化、价格调整）。应该如何处理这种差异？

**决策**：采用**差异容忍 + 提示机制**

**价格对比规则**：

| 场景 | 差异情况 | 处理策略 | 理由 |
|------|---------|---------|------|
| **场景1** | 价格降低 | ✅ 直接通过 | 对用户有利 |
| **场景2** | 价格上涨 < 1元 | ✅ 允许（容忍误差） | 微小差异，可接受 |
| **场景3** | 价格上涨 >= 1元 且 < 5% | ✅ 允许但记录日志 | 合理波动范围 |
| **场景4** | 价格上涨 >= 5% | ❌ 拒绝，要求重新确认 | 差异过大，影响用户决策 |

**实现代码**：

```go
func (s *OrderService) validatePriceChange(expected, actual int64) error {
    diff := actual - expected
    diffPercent := float64(diff) / float64(expected) * 100
    
    // 场景1: 价格降低 → 允许（对用户有利）
    if diff < 0 {
        return nil
    }
    
    // 场景2: 价格上涨 < 1元 → 允许
    if diff <= 100 {
        return nil
    }
    
    // 场景3: 价格上涨 < 5% → 允许但记录
    if diffPercent < 5.0 {
        log.Warnf("price increased: expected=%d, actual=%d, diff=%d", 
            expected, actual, diff)
        return nil
    }
    
    // 场景4: 价格上涨 >= 5% → 拒绝
    return &PriceChangedError{
        Expected: expected,
        Actual:   actual,
        Message:  fmt.Sprintf("价格已变化：原价%.2f元，现价%.2f元", 
            float64(expected)/100, float64(actual)/100),
    }
}
```

**前端交互**：

```javascript
// 前端处理价格变化错误
try {
    const order = await api.createOrder(orderData);
} catch (error) {
    if (error.code === 'PRICE_CHANGED') {
        // 弹窗提示用户
        showConfirmDialog({
            title: '价格已变化',
            message: error.message,
            confirm: '接受新价格并下单',
            cancel: '返回重新选择'
        }).then((confirmed) => {
            if (confirmed) {
                // 用户接受新价格，使用新价格重新下单
                api.createOrder({
                    ...orderData,
                    acceptNewPrice: true,
                    expectedPrice: error.actualPrice
                });
            }
        });
    }
}
```

---

### ADR-012: 试算价格计算与创单价格计算的统一与差异

**决策日期**：2026-04-15  
**状态**：已采纳 ✓

**问题描述**：试算接口（`/checkout/calculate`）和创单接口（`/order/create`）都需要计算价格，两者的价格计算逻辑应该如何设计？

**决策**：**统一计价服务 + 差异化数据输入**

**核心设计**：

| 接口 | 数据输入 | 计算逻辑 | 快照策略 |
|------|---------|---------|---------|
| **试算接口** | 可使用快照（5分钟） | 调用统一计价服务 | 允许快照数据 |
| **创单接口** | 强制实时查询 | 调用统一计价服务 | 禁止快照数据 |

**理由**：

1. **计价逻辑统一**：
   - 试算和创单使用同一个 `PricingService.Calculate`
   - 避免"试算价格"与"订单价格"不一致
   - 营销规则变更只需更新一处

2. **数据输入差异化**：
   - 试算：允许使用缓存/快照数据（性能优先）
   - 创单：强制实时查询（准确性优先）

3. **最终一致性保证**：
   - 试算阶段可能使用过期快照
   - 创单阶段的实时查询是最后防线
   - 价格差异会被拦截并提示用户

**架构图**：

```mermaid
graph TB
    subgraph 试算接口
        A1[Checkout.Calculate]
        A2[使用快照数据<br/>性能优先]
    end
    
    subgraph 创单接口
        B1[Order.Create]
        B2[强制实时查询<br/>准确性优先]
    end
    
    subgraph 计价服务
        C[PricingService.Calculate<br/>统一计算逻辑]
    end
    
    A1 --> A2
    A2 --> C
    B1 --> B2
    B2 --> C
```

---

### ADR-013: 价格在整个交易链路中的流转与计算策略

**决策日期**：2026-04-15  
**状态**：已采纳 ✓

**问题描述**：从用户搜索商品到最终支付，价格会经历多个阶段（搜索列表 → 商品详情 → 加购试算 → 创单 → 支付）。每个阶段的价格计算范围、数据来源、系统交互都不同。需要一个全局视角来理解价格是如何流转的，以及各阶段的相同点和不同点。

**核心挑战**：
```
业务困惑：
• 为什么搜索列表的价格和详情页不一样？
• 详情页显示的价格和试算价格能保证一致吗？
• 试算价格和最终支付价格可能不同吗？
• 每个阶段都要调用Pricing Service吗？
• 基础价格、营销折扣、优惠券、Coin、支付渠道费分别在哪个阶段计算？
```

**决策**：采用**"分阶段计算 + 逐步扩展价格维度 + 最终强制校验"**策略

---

**价格流转全局图**：

```
用户旅程：搜索 → 详情 → 试算 → 创单 → 支付
           ↓      ↓      ↓      ↓      ↓
价格计算： 基础价  +营销  +营销  +营销  +Coin+Voucher+渠道费
           ↓      ↓      ↓      ↓      ↓
数据来源： ES缓存  实时   快照   强制   强制实时
                         (可选) 实时
           ↓      ↓      ↓      ↓      ↓
性能目标： 30ms   150ms  230ms  500ms  200ms
```

**五个阶段对比**：

| 阶段 | 价格维度 | 数据来源 | 性能目标 | 计算复杂度 | 资损风险 |
|------|---------|---------|---------|-----------|---------|
| **搜索列表** | 基础价（最低价） | ES缓存（延迟1-5分钟） | P95 < 30ms | 低（只查ES） | 无 |
| **商品详情** | 基础价 + 营销折扣 | 实时查询 + 生成快照 | P95 < 150ms | 中（3个服务） | 无 |
| **结算试算** | 基础价 + 营销 + 数量 | 快照 OR 实时查询 | P95 < 230ms | 中（可能3个服务） | 无 |
| **确认下单** | 基础价 + 营销 + 数量 + 券 | 强制实时查询 | P95 < 500ms | 高（4个服务 + 预占） | 高 |
| **支付确认** | 上述 + Coin + Voucher + 渠道费 | 强制实时查询 | P95 < 200ms | 高（多维度计算） | 极高 |

**核心设计原则**：

1. **逐步扩展价格维度**
   ```
   搜索：最低价（吸引用户）
   详情：折扣价（展示营销）
   试算：总价（含数量、券）
   创单：锁定价（预占资源）
   支付：最终价（含所有优惠与费用）
   ```

2. **数据来源分级**
   ```
   搜索/详情：允许缓存（性能优先）
   试算：允许快照（性能与准确性平衡）
   创单/支付：强制实时（安全优先）
   ```

3. **多道防线保证准确性**
   ```
   详情页：生成快照（用于试算）
   试算：对比快照与实时（发现变化）
   创单：强制实时 + 价格校验（最后防线）
   支付：二次校验 + Coin/Voucher锁定（终极防线）
   ```

**监控指标**：
- 各阶段P95响应时间
- 快照命中率（目标 > 80%）
- 价格差异率（试算vs创单，目标 < 5%）
- 价格变化拦截率（创单价格校验触发频率）

---

## 16.10 高可用与性能优化（Infrastructure & Operations）

### 16.7.1 高可用设计

**服务多副本部署**：

| 服务 | 正常副本 | 大促副本 | 扩容策略 |
|------|---------|---------|---------|
| Product Center | 6 | 18 | CPU > 70% 自动扩容 |
| Inventory | 6 | 18 | QPS > 5000 扩容 |
| Order | 8 | 24 | QPS > 3000 扩容 |
| Payment | 4 | 12 | QPS > 2000 扩容 |

**数据库高可用**：

```
MySQL：
• 主从复制（1主2从）
• 双主互备（支付库）
• 自动故障转移（MHA）

Redis：
• Sentinel模式（1主2从3哨兵）
• 自动故障转移

Kafka：
• 3副本
• ISR机制
```

**熔断与降级**：

```go
// 熔断配置
type CircuitBreakerConfig struct {
    MaxRequests       uint32        // 半开状态最大请求数
    Interval          time.Duration // 统计窗口
    Timeout           time.Duration // 熔断超时时间
    FailureThreshold  float64       // 失败率阈值（0-1）
}

// 降级策略
func (s *SearchAggregation) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
    // 尝试调用Marketing Service
    promos, err := s.marketingClient.GetPromotions(ctx, req.SkuIDs)
    if err != nil {
        // 降级：使用基础价格（不展示营销信息）
        log.Warn("Marketing Service故障，降级为基础价格")
        promos = make(map[int64]*PromoInfo)  // 空促销
    }
    
    // 继续后续流程...
    return s.buildResponse(products, promos)
}
```

### 16.7.2 性能优化

**缓存策略**（多级缓存）：

```go
// L1: 本地缓存（进程内）
type LocalCache struct {
    cache *bigcache.BigCache
}

func (c *LocalCache) Get(key string) (interface{}, error) {
    data, err := c.cache.Get(key)
    if err == nil {
        return unmarshal(data), nil
    }
    return nil, err
}

// L2: Redis缓存
// L3: MySQL数据库

func (s *ProductService) GetProduct(ctx context.Context, skuID int64) (*Product, error) {
    // L1: 本地缓存
    if product, err := s.localCache.Get(skuID); err == nil {
        return product, nil
    }
    
    // L2: Redis缓存
    if product, err := s.redis.Get(ctx, fmt.Sprintf("product:%d", skuID)); err == nil {
        s.localCache.Set(skuID, product)  // 回填L1
        return product, nil
    }
    
    // L3: MySQL数据库
    product, err := s.repo.GetByID(ctx, skuID)
    if err != nil {
        return nil, err
    }
    
    // 回填缓存
    s.redis.Set(ctx, fmt.Sprintf("product:%d", skuID), product, 30*time.Minute)
    s.localCache.Set(skuID, product)
    
    return product, nil
}
```

**批量查询优化**：

```go
// 批量获取商品信息（减少RPC调用）
func (s *ProductService) BatchGetProducts(ctx context.Context, skuIDs []int64) (map[int64]*Product, error) {
    // Step 1: 尝试从缓存批量获取
    cached := s.redis.MGet(ctx, toCacheKeys(skuIDs))
    
    // Step 2: 找出缺失的ID
    missingIDs := findMissing(skuIDs, cached)
    
    // Step 3: 批量查询数据库（IN查询）
    if len(missingIDs) > 0 {
        missing, _ := s.repo.GetByIDs(ctx, missingIDs)
        // 回填缓存
        s.redis.MSet(ctx, missing, 30*time.Minute)
        cached = merge(cached, missing)
    }
    
    return cached, nil
}
```

**数据库优化**：

```sql
-- 索引优化
CREATE INDEX idx_order_user_create ON `order` (user_id, create_time DESC);
CREATE INDEX idx_order_status ON `order` (status, create_time DESC);

-- 避免SELECT *（只查询需要的字段）
SELECT order_id, status, total_price FROM `order` WHERE user_id = ?;

-- 分页优化（使用索引覆盖）
SELECT order_id FROM `order` 
WHERE user_id = ? AND create_time > ?
ORDER BY create_time DESC
LIMIT 20;
```

### 16.7.3 容灾与降级

**多机房部署**：

```
Region A（主）：
• 写流量：100%
• 读流量：70%

Region B（备）：
• 写流量：0%（只读副本）
• 读流量：30%

灾难切换：
• 自动故障检测（3秒）
• 流量切换到Region B（30秒）
• RTO：< 2分钟
```

**降级开关**：

```go
// Feature Flag控制降级
func (s *CheckoutService) Calculate(ctx context.Context, req *CalculateRequest) (*CalculateResponse, error) {
    // 检查Feature Flag
    if s.featureFlag.IsEnabled(ctx, "marketing.enabled") {
        // 正常逻辑：调用Marketing Service
        promos, _ := s.marketingClient.GetPromotions(ctx, req)
        return s.calculateWithPromos(req, promos)
    } else {
        // 降级逻辑：不使用营销信息
        return s.calculateBasic(req)
    }
}
```

---

## 16.11 团队组织与协作（Organization & Governance）

### 16.11.1 团队结构

**康威定律实践**：系统架构反映组织沟通结构。

```
订单团队（15人）
├─ 订单核心（5人）：订单创建、状态机
├─ 订单查询（3人）：我的订单、订单详情
├─ 履约对接（4人）：供应商履约、异常处理
└─ 测试（3人）

商品团队（12人）
├─ 商品中心（6人）：SPU/SKU管理
├─ 类目属性（3人）：类目树、属性模板
└─ 测试（3人）

库存团队（10人）
├─ 库存核心（5人）：预占、扣减、释放
├─ 供应商同步（3人）：实时查询、定时同步
└─ 测试（2人）
```

**跨团队协作**：

| 场景 | 协作方式 | 工具 |
|------|---------|------|
| API契约 | OpenAPI/Proto定义 | Swagger、Buf |
| 事件契约 | Schema Registry | Confluent Schema Registry |
| 联调测试 | 契约测试 | Pact |
| 故障处理 | On-call轮值 | PagerDuty |

### 16.8.2 协作流程

**需求评审流程**：

```
1. 产品提需求（PRD）
   ↓
2. 技术评审（架构师+各团队Lead）
   • 是否需要新增服务？
   • 是否需要修改API契约？
   • 是否需要数据库迁移？
   ↓
3. API契约评审（上下游团队）
   • 定义Request/Response
   • 明确超时、重试策略
   • 确认降级方案
   ↓
4. 开发排期
   • 各团队独立开发
   • 契约测试通过后联调
   ↓
5. 集成测试
   • 端到端测试
   • 性能测试
   ↓
6. 灰度发布
   • 5% → 20% → 50% → 100%
```

**实际案例：新增"拼团"功能的完整协作流程**

**第1周：需求评审与技术方案**

```
【产品需求】
- 用户发起拼团（3人成团，24小时有效）
- 拼团价格比正常价格低20%
- 成团后统一发货，不成团退款

【技术评审会议】（2小时，架构师+6个团队Lead）
问题1：拼团功能是否需要新增服务？
  → 决策：新增"拼团服务"（GroupBuy Service）
  → 理由：拼团逻辑复杂（成团判断、超时处理），独立服务便于维护

问题2：拼团价格如何计算？
  → 决策：在Pricing Service中新增"拼团价格策略"
  → 理由：价格计算逻辑应该统一管理

问题3：拼团成功后如何扣减库存？
  → 决策：拼团成功时批量预占库存（3人份）
  → 理由：避免成团后库存不足

【输出物】
- 技术方案文档（15页）
- 服务依赖图（Mermaid图）
- 数据库设计（ER图）
- 时序图（成团流程、超时处理）
```

**第2周：API契约评审**

```
【API契约】
// 创建拼团
POST /groupbuy/create
Request:
{
  "sku_id": 1001,
  "original_price": 299.00,
  "groupbuy_price": 239.00,  // 8折
  "required_count": 3,        // 3人成团
  "expire_hours": 24          // 24小时有效
}
Response:
{
  "groupbuy_id": "GB20260501123456",
  "status": "waiting",        // 等待中
  "current_count": 1,         // 当前人数
  "required_count": 3,
  "expires_at": 1744633200
}

// 参与拼团
POST /groupbuy/join
Request:
{
  "groupbuy_id": "GB20260501123456",
  "user_id": 67890
}
Response:
{
  "status": "success",        // 成功 or 团满
  "order_id": "ORD123456",    // 如果成团，返回订单号
  "current_count": 3
}

【契约测试】
- 上游：前端团队（Web、App）
- 下游：Pricing Service、Inventory Service、Order Service
- 测试工具：Pact
- 测试覆盖：100%（所有API）
```

**第3-4周：并行开发**

```
【团队分工】
拼团团队（5人）：
  - 拼团服务核心逻辑
  - 超时任务（15分钟扫描一次）
  - 数据库表设计（groupbuy、groupbuy_participant）

计价团队（2人）：
  - 新增拼团价格策略
  - 拼团价格校验

库存团队（2人）：
  - 批量预占库存接口

订单团队（3人）：
  - 拼团成团后批量创建订单
  - 拼团失败后退款

前端团队（4人）：
  - 拼团页面（发起、参与、分享）
  - 倒计时组件

【每日站会】（15分钟）
- 各团队汇报进度
- 识别阻塞点
- 协调资源

【契约测试通过率】
- 第3周末：70%
- 第4周末：100% ✅
```

**第5周：集成测试**

```
【测试场景】
场景1：正常成团
  1. 用户A发起拼团（3人成团）
  2. 用户B、C参与拼团
  3. 成团 → 创建3个订单 → 预占库存（3份）
  4. 用户A、B、C支付 → 确认扣减库存

场景2：超时未成团
  1. 用户A发起拼团（3人成团）
  2. 只有用户B参与（2人）
  3. 24小时后超时 → 标记拼团失败 → 退款

场景3：库存不足
  1. 用户A发起拼团（3人成团）
  2. 用户B、C参与拼团
  3. 成团时库存不足（只剩2个）→ 拼团失败 → 退款

【性能测试】
- 并发创建拼团：1000 TPS
- 并发参与拼团：5000 TPS
- 超时扫描任务：1000个拼团/秒
- P99延迟：< 300ms ✅
```

**第6周：灰度发布**

```
【灰度策略】
阶段1（5%）：内部员工 + 白名单用户（1000人）
  → 观察1天：成团率、退款率、投诉数

阶段2（20%）：北京、上海用户
  → 观察3天：性能指标、业务指标

阶段3（50%）：全国用户
  → 观察1周

阶段4（100%）：全量发布
  → 持续监控1个月

【关键指标】
- 成团率：65%（目标 > 60%）✅
- 退款率：5%（目标 < 10%）✅
- 用户投诉：3起/天（目标 < 10起）✅
- P99延迟：280ms（目标 < 300ms）✅
```

**协作关键点**：

| 阶段 | 关键协作点 | 工具/机制 |
|------|----------|---------|
| **需求评审** | 架构师+各团队Lead对齐技术方案 | 会议+文档 |
| **API契约** | 上下游团队明确接口定义 | OpenAPI + Pact |
| **并行开发** | 各团队独立开发，通过契约测试联调 | Pact + Mock Server |
| **集成测试** | 端到端测试，验证完整流程 | 自动化测试平台 |
| **灰度发布** | 分阶段发布，持续监控 | Feature Flag + Grafana |

**变更管理**：

```go
// ADR（Architecture Decision Record）
// 记录重大架构决策

## ADR-014: 拼团功能是否复用订单服务

**决策日期**：2026-05-01
**状态**：已采纳 ✓

**问题描述**：
拼团功能需要创建订单，是在订单服务中新增拼团逻辑，还是新建拼团服务？

**备选方案**：
A. 在订单服务中新增拼团逻辑
   ✓ 复用订单创建逻辑
   ✗ 订单服务变得臃肿
   ✗ 拼团逻辑与订单逻辑耦合

B. 新建拼团服务
   ✓ 拼团逻辑独立，便于维护
   ✓ 订单服务保持单一职责
   ✗ 需要新建服务（增加运维成本）

**决策**：采用方案B，新建拼团服务

**理由**：
1. 拼团逻辑复杂（成团判断、超时处理、退款逻辑）
2. 拼团是营销活动，不是订单核心流程
3. 未来可能有"砍价""秒杀"等类似活动，独立服务便于扩展

**影响范围**：
- 新增服务：GroupBuy Service
- QPS估算：2000（正常）/ 10000（大促）
- 部署规模：4副本（正常）/ 12副本（大促）

**后续行动**：
- ✓ 已完成：GroupBuy Service开发
- ✓ 已完成：与订单服务集成
- ✓ 已完成：灰度上线
```

### 16.8.3 技术治理

**代码评审清单**：

- [ ] 是否符合分层架构（依赖方向正确）
- [ ] 是否有单元测试（覆盖率 > 80%）
- [ ] 是否有集成测试（核心路径）
- [ ] 是否有性能测试（Benchmark）
- [ ] 是否有监控指标（Prometheus Metrics）
- [ ] 是否有日志（结构化日志）
- [ ] 是否有文档（API文档、设计文档）
- [ ] 是否考虑降级方案

**技术债管理**：

```markdown
## 技术债清单

| 优先级 | 类型 | 描述 | 负责人 | 预计工作量 |
|-------|------|------|--------|-----------|
| P0 | 性能 | 订单查询慢查询优化 | @张三 | 2天 |
| P1 | 安全 | 支付回调签名验证 | @李四 | 1天 |
| P2 | 代码 | 商品中心重复代码重构 | @王五 | 3天 |
```

---

## 16.12 上线与演进（Deployment & Evolution）

### 16.12.1 上线策略

**分阶段上线**：

```
阶段1：基础功能（2周）
• 商品中心、库存服务、订单服务
• 支持机票、酒店两个品类
• 单机房部署

阶段2：营销功能（2周）
• 营销服务、计价服务
• 支持优惠券、活动

阶段3：新品类（每周1个）
• 充值、电影票、优惠券、礼品卡

阶段4：多机房（4周）
• 双机房部署
• 流量灰度切换
```

### 16.9.2 灰度发布

**灰度策略**：

```go
// 灰度规则
type GrayReleaseRule struct {
    Version    string   // 新版本号
    Percentage int      // 流量比例（0-100）
    Whitelist  []int64  // 白名单用户ID
    Regions    []string // 灰度地区
}

func (r *GrayRouter) Route(userID int64, region string) string {
    // 白名单用户直接路由到新版本
    if contains(r.rule.Whitelist, userID) {
        return r.rule.Version
    }
    
    // 按地区灰度
    if !contains(r.rule.Regions, region) {
        return "stable"  // 老版本
    }
    
    // 按百分比灰度
    if hash(userID) % 100 < r.rule.Percentage {
        return r.rule.Version  // 新版本
    }
    
    return "stable"  // 老版本
}
```

**灰度步骤**：

```
1. 5%流量（白名单用户 + 内部员工）
   观察1小时：错误率、延迟、业务指标

2. 20%流量（特定地区）
   观察2小时

3. 50%流量
   观察4小时

4. 100%流量（全量发布）
   观察24小时

5. 下线老版本
```

### 16.9.3 监控告警

**三级监控体系**：

| 层级 | 监控对象 | 工具 | 告警阈值 |
|------|---------|------|---------|
| **业务监控** | 订单量、GMV、转化率 | Grafana + ClickHouse | 同比下降20% |
| **应用监控** | QPS、延迟、错误率 | Prometheus + Grafana | P99延迟>500ms |
| **基础设施监控** | CPU、内存、磁盘、网络 | Prometheus + Node Exporter | CPU>80% |

**核心指标**：

```go
// Prometheus Metrics
package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
    // 业务指标
    orderCreatedTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "order_created_total",
            Help: "订单创建总数",
        },
        []string{"category", "status"},  // 标签：品类、状态
    )
    
    orderCreatedLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "order_created_latency_seconds",
            Help:    "订单创建延迟",
            Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
        },
        []string{"category"},
    )
    
    orderGMV = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "order_gmv_total",
            Help: "订单GMV（元）",
        },
        []string{"date"},
    )
    
    // 系统指标
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP请求延迟",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_request_total",
            Help: "HTTP请求总数",
        },
        []string{"method", "endpoint", "status"},
    )
    
    // 依赖服务指标
    rpcCallDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "rpc_call_duration_seconds",
            Help:    "RPC调用延迟",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
        },
        []string{"service", "method", "status"},
    )
    
    rpcCallTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "rpc_call_total",
            Help: "RPC调用总数",
        },
        []string{"service", "method", "status"},
    )
    
    // 数据库指标
    dbQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "数据库查询延迟",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"query_type", "table"},
    )
    
    // Redis指标
    redisCommandDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "redis_command_duration_seconds",
            Help:    "Redis命令延迟",
            Buckets: []float64{.0001, .0005, .001, .005, .01, .025, .05, .1},
        },
        []string{"command"},
    )
)

// 使用示例
func CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    startTime := time.Now()
    
    // 业务逻辑...
    order, err := createOrderInternal(ctx, req)
    
    // 记录指标
    duration := time.Since(startTime).Seconds()
    category := req.Category
    status := "success"
    if err != nil {
        status = "failed"
    }
    
    // 记录订单创建总数
    orderCreatedTotal.WithLabelValues(category, status).Inc()
    
    // 记录订单创建延迟
    orderCreatedLatency.WithLabelValues(category).Observe(duration)
    
    // 记录GMV
    if err == nil {
        orderGMV.WithLabelValues(time.Now().Format("2006-01-02")).Add(float64(order.TotalPrice))
    }
    
    return order, err
}
```

**告警规则**：

```yaml
# Prometheus AlertManager规则
groups:
  - name: order-service-alerts
    rules:
      # P99延迟告警
      - alert: OrderCreateLatencyHigh
        expr: histogram_quantile(0.99, order_created_latency_seconds) > 1
        for: 5m
        labels:
          severity: warning
          service: order-service
        annotations:
          summary: "订单创建延迟过高"
          description: "P99延迟 {{ $value }}s > 1s（持续5分钟）"
          dashboard: "https://grafana.example.com/d/order-service"
      
      # 错误率告警
      - alert: OrderCreateErrorRateHigh
        expr: |
          sum(rate(order_created_total{status="failed"}[5m])) 
          / sum(rate(order_created_total[5m])) > 0.01
        for: 5m
        labels:
          severity: critical
          service: order-service
        annotations:
          summary: "订单创建失败率过高"
          description: "失败率 {{ $value | humanizePercentage }} > 1%"
          runbook: "https://wiki.example.com/runbook/order-create-error"
      
      # QPS下降告警（业务异常）
      - alert: OrderCreateQPSDrop
        expr: |
          (sum(rate(order_created_total[5m])) 
          / sum(rate(order_created_total[5m] offset 1h))) < 0.5
        for: 10m
        labels:
          severity: warning
          service: order-service
        annotations:
          summary: "订单创建QPS骤降"
          description: "当前QPS {{ $value }}，比1小时前下降50%以上"
      
      # GMV下降告警（业务异常）
      - alert: OrderGMVDrop
        expr: |
          (sum(rate(order_gmv_total[1h])) 
          / sum(rate(order_gmv_total[1h] offset 24h))) < 0.8
        for: 30m
        labels:
          severity: critical
          service: order-service
        annotations:
          summary: "订单GMV大幅下降"
          description: "当前GMV {{ $value }}，比昨天同期下降20%以上"
      
      # RPC调用失败率告警
      - alert: RPCCallErrorRateHigh
        expr: |
          sum(rate(rpc_call_total{status!="success"}[5m])) by (service) 
          / sum(rate(rpc_call_total[5m])) by (service) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "RPC调用失败率过高：{{ $labels.service }}"
          description: "失败率 {{ $value | humanizePercentage }} > 5%"
      
      # 数据库慢查询告警
      - alert: DBSlowQuery
        expr: histogram_quantile(0.99, db_query_duration_seconds) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "数据库慢查询"
          description: "P99延迟 {{ $value }}s > 100ms"
      
      # Redis延迟告警
      - alert: RedisLatencyHigh
        expr: histogram_quantile(0.99, redis_command_duration_seconds) > 0.01
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Redis延迟过高"
          description: "P99延迟 {{ $value }}s > 10ms"
      
      # 服务实例Down告警
      - alert: ServiceInstanceDown
        expr: up{job="order-service"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "服务实例宕机"
          description: "实例 {{ $labels.instance }} 已宕机超过1分钟"
```

**告警分级与处理**：

| 级别 | 触发条件 | 通知方式 | 响应时间 | 处理人 |
|------|---------|---------|---------|--------|
| **P0（紧急）** | GMV下降>20%、服务全部宕机 | 电话+短信+企业微信 | < 5分钟 | On-call工程师+经理 |
| **P1（严重）** | 错误率>1%、P99延迟>1s | 企业微信+短信 | < 15分钟 | On-call工程师 |
| **P2（警告）** | QPS下降>50%、数据库慢查询 | 企业微信 | < 30分钟 | 值班工程师 |
| **P3（提示）** | 磁盘使用>80%、内存使用>80% | 邮件 | < 2小时 | 运维团队 |

**监控大屏**：

```
┌─────────────────────────────────────────────────────────┐
│                   订单服务实时监控大屏                    │
├─────────────────────────────────────────────────────────┤
│  今日订单量: 1,234,567  ↑ 12.3%   今日GMV: ¥456,789,012  │
│  当前QPS: 2,345         P99延迟: 234ms    错误率: 0.12%  │
├─────────────────────────────────────────────────────────┤
│  订单创建趋势（24小时）             QPS & P99延迟         │
│  ███████████████████████████████   ███████████████████   │
│  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓   ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒     │
│  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░   ░░░░░░░░░░░░░░░░░     │
├─────────────────────────────────────────────────────────┤
│  品类分布               服务依赖健康度                   │
│  机票: 45%  ████████   Product Service:   ✅ 正常        │
│  酒店: 30%  ██████     Inventory Service: ✅ 正常        │
│  充值: 15%  ███        Pricing Service:   ⚠️  延迟高     │
│  其他: 10%  ██         Marketing Service: ✅ 正常        │
├─────────────────────────────────────────────────────────┤
│  活跃告警（3条）                                         │
│  ⚠️  P1 Pricing Service P99延迟>500ms（持续10分钟）      │
│  📊 P2 订单QPS比昨天同期下降15%                          │
│  💾 P3 MySQL主库连接数>80%                              │
└─────────────────────────────────────────────────────────┘
```

**On-call值班机制**：

```
【值班表】（7x24小时）
周一：张三（订单团队）
周二：李四（商品团队）
周三：王五（库存团队）
...

【值班职责】
1. 响应P0/P1告警（5分钟内）
2. 排查问题根因（15分钟内定位）
3. 协调资源修复（30分钟内恢复）
4. 事后复盘（24小时内）

【升级机制】
On-call工程师无法处理 → 升级到Team Lead
Team Lead无法处理 → 升级到架构师
架构师无法处理 → 升级到CTO
```

### 16.9.4 系统演进路径

**已完成**：
- ✅ 基础架构搭建（微服务、服务发现、监控）
- ✅ 核心品类上线（机票、酒店、充值）
- ✅ 营销系统（优惠券、活动）
- ✅ 双机房部署

**进行中**：
- 🚧 性能优化（P99延迟 < 200ms）
- 🚧 新品类接入（电影票、礼品卡）
- 🚧 供应商扩展（50+ → 100+）

**规划中**：
- 📅 国际化（多语言、多币种）
- 📅 推荐系统（AI推荐）
- 📅 智能客服（NLP）
- 📅 区块链溯源（高端商品）

---

## 16.13 经验总结（Lessons Learned）

### 16.13.1 成功经验

**1. 架构决策记录（ADR）制度**

价值：
- 重大决策留痕，新人可快速了解背景
- 避免重复讨论已解决的问题
- 架构演进有据可查

建议：
- 每个ADR包含：问题、决策、理由、权衡、影响范围
- 定期Review（每季度）
- 与代码一起版本管理

**2. 品类差异化设计**

价值：
- 避免"一刀切"架构（机票与充值差异大）
- 策略模式让新品类接入成本降低80%
- 适配器模式让供应商集成周期从4周缩短到1周

建议：
- 先分析业务模型差异，再设计技术方案
- 抽象共性，策略处理差异
- 避免过度抽象（YAGNI原则）

**3. 聚合层编排模式**

价值：
- API Gateway职责单一（鉴权、限流、路由）
- 业务编排集中在聚合层，易于优化
- 降级策略统一管理

建议：
- 聚合层只做数据获取与编排，不做业务计算
- 支持并发调用（提升性能）
- 统一降级策略（Marketing故障降级为基础价）

**4. 多级缓存策略**

价值：
- P99延迟从500ms降低到200ms
- Redis QPS降低60%（本地缓存命中率30%）
- 大促期间扛住5倍流量

建议：
- L1（本地）：热点数据，1分钟TTL
- L2（Redis）：通用数据，30分钟TTL
- L3（MySQL）：源数据
- 缓存失效策略：主动失效 + TTL兜底

**5. 契约测试**

价值：
- 上下游团队并行开发（不等联调）
- API变更影响提前发现
- 集成测试成本降低70%

建议：
- 使用Pact等契约测试工具
- API契约与代码一起版本管理
- CI自动运行契约测试

### 16.10.2 踩过的坑

**坑1：过早引入Event Sourcing**

**问题**：
- 初期为了"追求架构完美"引入Event Sourcing
- 团队对ES理解不足，查询复杂，运维困难
- 投影重建耗时长（大促后修复bug需要重建投影，耗时4小时）

**教训**：
- Event Sourcing不是银弹，适用于审计要求极高的场景
- 对于大部分电商场景，CQRS（不带ES）足够
- 先用简单方案（CRUD），待确认瓶颈后再演进

**坑2：供应商接口未做熔断**

**问题**：
- 某供应商故障，接口超时（30秒）
- 大量请求堆积，线程池耗尽
- 整个订单服务不可用（影响其他供应商）

**教训**：
- 所有外部调用必须熔断（gobreaker）
- 超时时间合理设置（不超过1秒）
- 故障隔离（某个供应商故障不影响其他）

**坑3：分库分表过早**

**问题**：
- 订单量100万时就分库分表（8库64表）
- 运维复杂度激增（扩容、迁移、对账）
- 跨库查询需要路由表，增加延迟

**教训**：
- 单表500万以下不分表（MySQL性能足够）
- 单库3000万以下不分库
- 分库分表需要充分评估成本收益

**坑4：忽视数据一致性对账**

**问题**：
- 库存预占后未释放（代码bug）
- 累积1个月后，库存数据严重不准确
- 影响用户体验（明明有库存却提示"已售罄"）

**教训**：
- 异步操作必须有对账机制（每小时/每天）
- 对账发现差异要有自动补偿
- 监控库存准确率（定期抽查）

**坑5：缓存穿透导致雪崩**

**问题**：
- 恶意请求查询不存在的商品（skuID=0）
- 缓存未命中，直接打到数据库
- 数据库连接池耗尽，服务雪崩

**教训**：
- 布隆过滤器（Bloom Filter）拦截不存在的Key
- 缓存空值（TTL=1分钟）
- 请求参数校验（前置拦截非法请求）

### 16.10.3 改进方向

**短期改进（3个月内）**：

1. **性能优化**
   - **目标**：P99延迟从200ms降低到150ms
   - **措施**：
     - 热点数据预加载：大促前提前加载10万+热门商品到Redis
     - 数据库慢查询优化：全部慢查询(<50ms)，添加复合索引
     - 连接池优化：MySQL连接池从100提升到500
     - 批量查询优化：单次查询支持100+商品（原50个）
   - **预期收益**：QPS提升30%，响应时间降低25%

2. **稳定性提升**
   - **混沌工程实践**：
     - 每周定期故障演练（随机Kill Pod、网络延迟、数据库主从切换）
     - 自动化故障注入工具（Chaos Mesh）
     - 故障恢复时间目标：< 3分钟
   - **降级开关完善**：
     - 所有非核心功能支持降级（营销、推荐、评论）
     - Feature Flag平台（实时开关，无需重启）
     - 降级决策自动化（根据错误率自动降级）
   - **容量规划**：
     - 提前3个月预估资源需求（基于历史数据+增长率）
     - 大促前1个月进行压测（验证容量）
     - 弹性扩容策略（CPU > 70%自动扩容）

3. **开发效率**
   - **统一脚手架**：
     - 一键创建新服务（包含标准目录结构、配置文件、CI/CD）
     - 内置最佳实践（监控、日志、链路追踪）
     - 代码生成工具（Proto → Go代码自动生成）
   - **自动化测试**：
     - 单元测试覆盖率 > 90%（核心业务逻辑100%覆盖）
     - 集成测试自动化（每次提交自动运行）
     - 性能测试定期执行（每周一次，P99延迟不能退化）
   - **CI/CD优化**：
     - 构建时间 < 5分钟（并行构建、增量构建、缓存优化）
     - 自动化部署（合并到main分支自动部署到生产）
     - 灰度发布流程标准化（5% → 20% → 50% → 100%）

**中期改进（6-12个月）**：

1. **智能化**
   - **推荐系统**：
     - 协同过滤（基于用户行为相似度）
     - 深度学习模型（基于用户画像+商品属性）
     - 实时推荐（用户浏览行为实时调整推荐结果）
     - A/B测试（对比推荐效果，持续优化）
     - 预期提升：点击率+15%，转化率+10%
   
   - **动态定价**：
     - 根据供需关系自动调价（库存少+需求高 → 涨价）
     - 竞品价格监控（爬虫+算法，自动调整价格）
     - 用户画像定价（VIP用户优惠力度更大）
     - 时段定价（早上价格高，晚上价格低）
     - 预期提升：毛利率+8%，订单量+12%
   
   - **智能客服**：
     - FAQ自动回复（NLP模型识别用户问题）
     - 订单查询自动化（用户输入订单号，自动查询状态）
     - 售后自动化（退款、换货流程自动化）
     - 人工客服辅助（AI推荐回复话术）
     - 预期收益：客服成本降低40%，响应速度提升50%

2. **国际化**
   - **多语言支持（i18n）**：
     - 支持英语、中文、日语、韩语、泰语
     - 翻译管理平台（统一管理翻译资源）
     - 动态语言切换（用户可随时切换语言）
     - 本地化适配（日期格式、货币符号、文化差异）
   
   - **多币种支持**：
     - 支持USD、EUR、JPY、CNY等10+币种
     - 汇率实时转换（接入外汇API，每分钟更新）
     - 价格展示优化（根据用户地区自动选择币种）
     - 结算币种选择（支持多币种支付）
   
   - **跨境支付**：
     - 接入PayPal、Stripe（国际信用卡）
     - 本地化支付（日本：Pay-easy，韩国：KakaoPay）
     - 外汇结算（自动结汇，降低汇率风险）

3. **数据驱动**
   - **实时数据大屏**：
     - GMV实时展示（今日/本周/本月）
     - 订单量、转化率、客单价实时监控
     - 品类TOP10、商品TOP100
     - 地域分布、用户画像
     - 技术栈：Flink + ClickHouse + Grafana
   
   - **A/B测试平台**：
     - 灰度实验（新功能A/B测试）
     - 流量分配（按用户ID哈希，保证一致性）
     - 效果评估（点击率、转化率、收入对比）
     - 自动化决策（效果好的方案自动全量）
   
   - **用户画像**：
     - 行为标签（浏览、加购、下单、复购）
     - 偏好标签（品类偏好、价格敏感度、优惠敏感度）
     - 生命周期标签（新用户、活跃用户、流失用户）
     - 精准营销（根据画像推送个性化优惠）

**长期愿景（1-3年）**：

1. **平台化**
   - **开放API**：
     - 商品API（第三方接入商品数据）
     - 订单API（第三方接入订单流程）
     - 支付API（第三方接入支付能力）
     - API网关（统一鉴权、限流、监控）
     - 预期收益：生态规模扩大3倍
   
   - **SaaS化**：
     - 中小企业独立部署（提供SaaS服务）
     - 多租户隔离（数据隔离、资源隔离）
     - 按需付费（按订单量或GMV收费）
     - 自助配置（商家自助配置商品、营销）
   
   - **生态建设**：
     - 开发者社区（技术文档、SDK、Demo）
     - 第三方插件市场（营销插件、支付插件）
     - 合作伙伴计划（供应商、物流商、支付商）

2. **技术创新**
   - **Serverless架构**：
     - 函数计算（FaaS）替代部分微服务
     - 按需计费（降低运维成本50%）
     - 自动扩容（无需手动扩容）
     - 适用场景：短信通知、数据清洗、报表生成
   
   - **Edge Computing**：
     - CDN边缘计算（静态资源、动态渲染）
     - 边缘缓存（用户就近访问，降低延迟）
     - 边缘函数（简单业务逻辑在边缘执行）
     - 预期收益：首屏加载时间降低60%
   
   - **区块链溯源**：
     - 高端商品防伪（奢侈品、珠宝）
     - 全链路追溯（生产、流通、销售）
     - 不可篡改（区块链存证）
     - 增强用户信任

**改进路线图**：

```mermaid
gantt
    title 系统改进路线图
    dateFormat YYYY-MM
    section 短期（3个月）
    性能优化           :2026-05, 3M
    稳定性提升         :2026-05, 3M
    开发效率           :2026-05, 3M
    
    section 中期（6-12个月）
    智能化             :2026-08, 12M
    国际化             :2026-08, 12M
    数据驱动           :2026-08, 12M
    
    section 长期（1-3年）
    平台化             :2027-08, 24M
    技术创新           :2027-08, 24M
```

**关键里程碑**：

| 时间 | 里程碑 | 成功标准 |
|------|-------|---------|
| **2026-08** | 性能优化完成 | P99延迟 < 150ms，QPS提升30% |
| **2026-11** | 稳定性提升完成 | 故障恢复时间 < 3分钟，可用性 > 99.99% |
| **2027-02** | 智能化上线 | 推荐点击率+15%，动态定价毛利率+8% |
| **2027-05** | 国际化完成 | 支持5种语言，10种币种，海外订单占比20% |
| **2027-08** | 数据驱动成熟 | A/B测试平台日活10万+，用户画像覆盖率100% |
| **2028-08** | 平台化初步完成 | 开放API日调用100万+，接入第三方100+ |
| **2029-08** | 技术创新落地 | Serverless占比30%，边缘计算覆盖80%流量 |

---

## 16.14 本章小结（Chapter Summary）

本章通过一个中大型B2B2C电商平台的完整案例，展示了从业务分析到技术落地的全过程，**是全书知识点的综合实践验证**。本章不仅覆盖了架构方法论（第1-6章），还深入展示了**供给运营系统（第11章）**和**C端核心交易流（第12-16章）**的完整实现，真正做到了"理论→实践→落地"的闭环。

---

### 核心要点回顾

**1. 品类差异化设计是关键**

不同品类的业务模型存在本质差异，这是架构设计的基础：

| 品类 | 库存模型 | 价格模型 | 履约模式 | 超卖容忍度 |
|------|---------|---------|---------|-----------|
| **机票** | 实时库存（供应商） | 动态定价 | 异步出票 | 零容忍 |
| **酒店** | 日历库存 | 日历定价 | 异步确认 | 零容忍 |
| **充值** | 无限库存 | 固定面额 | 同步充值 | 可补偿 |
| **优惠券** | 券码池 | 固定折扣 | 即时发放 | 可补偿 |

**设计启示**：
- ✅ 使用策略模式处理品类差异（避免 if-else 地狱）
- ✅ 使用适配器模式统一供应商接口（降低耦合）
- ✅ 模板方法定义统一流程（具体步骤由策略实现）
- ❌ 避免"一刀切"架构（机票与充值差异巨大，不能用同一套逻辑）

**2. 聚合层解决跨服务编排问题**

```
API Gateway（职责单一）
   ↓ 鉴权、限流、路由
Aggregation Service（编排层）
   ↓ 并发调用、数据聚合、降级处理
Business Services（业务层）
   ↓ 单一职责、独立部署
Infrastructure（基础设施层）
```

**为什么需要聚合层？**
- ✅ API Gateway保持职责单一（鉴权、限流、路由）
- ✅ 复杂编排逻辑集中管理（搜索场景：ES → Product → Inventory → Marketing → Pricing）
- ✅ 统一降级策略（Marketing故障降级为基础价）
- ✅ 性能优化空间大（并发调用、批量查询、缓存聚合结果）

**3. 架构决策记录（ADR）是宝贵资产**

本章记录了13个关键ADR决策：

| ADR编号 | 决策主题 | 核心价值 |
|---------|---------|---------|
| **ADR-001** | 计价中心数据输入方式 | 聚合层传入 vs 计价层自己调用 |
| **ADR-002** | 库存预占时机 | 试算 vs 创单 |
| **ADR-003** | 聚合服务 vs BFF | 按业务场景 vs 按端 |
| **ADR-004** | 虚拟商品库存模型 | 二维模型（ManagementType + UnitType） |
| **ADR-005** | 同步 vs 异步数据流 | 核心路径同步，非核心异步 |
| **ADR-009** | 创单时是否使用快照 | 强制实时查询（安全优先） |
| **ADR-010** | 创单与支付的时序 | 先创单后支付（防止超卖） |
| **ADR-011** | 前后端价格校验策略 | 差异容忍 + 提示机制 |
| **ADR-012** | 试算与创单价格计算 | 统一引擎 + 差异化数据来源 |
| **ADR-013** | 价格流转全局策略 | 分阶段计算 + 逐步扩展维度 |

**ADR的价值**：
- ✅ 记录决策背景（新人快速了解"为什么这样设计"）
- ✅ 避免重复讨论（已解决的问题有文档可查）
- ✅ 架构演进有据可查（回顾历史决策，持续优化）
- ✅ 与代码一起版本管理（决策与实现同步演进）

**4. 系统边界清晰至关重要**

**案例1：计价系统的边界重构**
- **问题**：价格计算逻辑分散在订单、营销、商品三个域
- **重构**：新建计价上下文，提供统一试算接口
- **收益**：价格一致性得到保证，营销规则变更只需在营销域发布事件

**案例2：库存预占的归属**
- **争议**：库存预占应该放在订单域还是库存域？
- **决策**：放在库存域
- **理由**：库存域拥有库存数据所有权，预占是库存的一种状态，订单域只需调用库存域的 Reserve 接口

**案例3：防腐层保护领域模型**
```go
// 供应商响应模型（外部）
type SupplierFlightResponse struct {
    Code    string
    Message string
    Data    struct {...}
}

// 平台库存模型（内部）
type StockResponse struct {
    Available bool
    Quantity  int
    Message   string
}

// 防腐层：翻译外部模型 → 内部模型
func (a *FlightSupplierACL) TranslateStock(supplierResp) *StockResponse {
    // 领域层不被供应商模型污染
}
```

**5. 高可用需要多层防护**

| 层级 | 措施 | 工具/技术 |
|------|------|---------|
| **应用层** | 服务多副本、自动扩容 | Kubernetes HPA |
| **接口层** | 熔断、降级、限流 | gobreaker、Feature Flag |
| **缓存层** | 多级缓存（本地+Redis+DB） | BigCache + Redis |
| **数据层** | 主从复制、读写分离 | MySQL Replication |
| **机房层** | 多机房部署、灰度发布 | Multi-Region + Canary |

**稳定性三板斧**：
- ✅ **熔断**：供应商调用失败率>50%，熔断10秒
- ✅ **降级**：Marketing Service故障，降级为基础价
- ✅ **限流**：令牌桶算法，QPS=500

**6. 供给运营是平台的核心能力**（新增16.5.6）

**三种核心场景**：

| 场景 | 业务语义 | 处理逻辑 | 审核策略 |
|------|---------|---------|---------|
| **商品上架** | 新商品首次进入平台 | Create | 完整审核流程 |
| **供应商同步** | 供应商数据变更 | Upsert | 差异化审核 |
| **运营编辑** | 已上线商品维护 | Update | 差异化审核 |

**设计要点**：
- ✅ **幂等性保证**：task_code唯一索引（上架）、sync_id唯一索引（同步）
- ✅ **差异化审核**：高风险变更（价格变化>50%、类目变更）必须审核
- ✅ **批量操作**：异步任务 + 进度追踪（100+ SKU批量编辑）
- ✅ **状态机**：DRAFT → PENDING → APPROVED → PUBLISHED
- ✅ **与商品中心集成**：审核通过后写入商品中心、初始化库存/价格

**7. C端交易流贯穿整个业务链路**（新增16.5.7）

**五个阶段完整设计**：

```
搜索（Query理解+ES召回+Hydrate）
   ↓ 转化率 > 15%
详情页（多服务聚合+快照生成）
   ↓ 转化率 > 8%
购物车（未登录加购+登录合并+双写）
   ↓ 转化率 > 30%
结算页（价格试算+库存检查+优惠校验）
   ↓ 转化率 > 60%
下单支付（Saga编排+实时查询+价格校验）
   ↓ 转化率 > 85%
```

**关键技术**：
- ✅ **Hydrate编排**：并发调用4-5个服务（Product、Inventory、Pricing、Marketing）
- ✅ **快照机制**：详情页生成快照（5分钟TTL），结算页可选使用（性能优先）
- ✅ **购物车合并**：未登录Redis存储，登录后合并到用户购物车
- ✅ **Saga编排**：下单时依次执行库存预占、优惠券锁定、价格计算、订单创建
- ✅ **强制实时查询**：创单时不使用任何快照（ADR-009，安全优先）

**8. DDD战术设计落地实践**（新增16.5.8）

**Order聚合根设计**：
```go
// 聚合根
type Order struct {
    orderID OrderID          // 值对象（聚合根ID）
    items   []*OrderItem     // 实体集合
    pricing *OrderPricing    // 值对象
    status  OrderStatus      // 值对象
    domainEvents []DomainEvent // 领域事件
}

// 值对象：OrderID（不可变）
// 值对象：OrderPricing（无ID，通过属性比较相等性）
// 实体：OrderItem（有ID，可变）
```

**Repository + Outbox模式**：
- ✅ **Repository接口在领域层定义**（不依赖基础设施）
- ✅ **领域事件与业务在同一事务**（Outbox表）
- ✅ **Outbox轮询器**：定时扫描未发布事件，发布到Kafka

**领域事件**：
```go
// OrderStatusChangedEvent（订单状态变更）
// OrderItemAddedEvent（商品项添加）
// OrderCreatedEvent（订单创建）
```

**9. 团队协作与技术治理同等重要**

**康威定律实践**：
```
订单团队（15人）→ 订单服务
商品团队（12人）→ 商品中心
库存团队（10人）→ 库存服务
...
```

**契约测试加速并行开发**：
- ✅ 上下游团队定义API契约（OpenAPI/Proto）
- ✅ 消费者编写契约测试（Pact）
- ✅ 提供者验证契约（契约测试通过后联调）
- ✅ 契约变更影响提前发现（CI自动运行）

**技术治理机制**：
- ✅ ADR记录重大决策
- ✅ 代码评审清单（架构、设计、代码、测试）
- ✅ 技术债管理（优先级、负责人、工作量）
- ✅ 定期架构Review（每季度）

---

### 实战价值

本章不是空洞的理论，而是200+人团队、日订单200万级的真实实践总结：

**成功经验**（值得借鉴）：
1. **ADR制度**：让架构演进有据可查，新人快速上手
2. **品类差异化**：策略模式让新品类接入成本降低80%
3. **聚合编排**：API Gateway职责单一，性能优化空间大
4. **多级缓存**：P99延迟从500ms降低到200ms
5. **契约测试**：团队并行开发，集成测试成本降低70%

**踩过的坑**（避坑指南）：
1. **过早引入Event Sourcing**：团队理解不足，查询复杂，运维困难
2. **供应商接口未做熔断**：某供应商故障，整个订单服务不可用
3. **分库分表过早**：订单量100万就分库分表，运维复杂度激增
4. **忽视数据一致性对账**：库存预占后未释放，累积1个月后严重不准确
5. **缓存穿透导致雪崩**：恶意请求查询不存在的商品，数据库连接池耗尽

**改进方向**（持续演进）：
- **短期**（3个月）：性能优化、稳定性提升、开发效率
- **中期**（6-12个月）：智能化、国际化、数据驱动
- **长期**（1-3年）：平台化、技术创新

---

### 与其他章节的关系

本章是全书知识点的综合应用与实践验证：

| 前置章节 | 在本章的应用 |
|---------|------------|
| **第1章**（架构方法论） | Clean Architecture分层、DDD战略设计（16.5.8）、CQRS读写分离 |
| **第2章**（领域驱动设计） | 12个限界上下文划分、上下文映射、防腐层（16.6.4） |
| **第3章**（代码整洁） | 策略模式（品类策略）、适配器模式（供应商集成）、SOLID原则 |
| **第4章**（质量保障） | ADR（13个决策）、代码评审清单、测试策略 |
| **第8章**（商品中心） | SPU/SKU模型、类目属性、商品快照 |
| **第9章**（库存系统） | 二维库存模型、预占机制、超时释放（16.5.2） |
| **第10章**（营销系统） | 营销规则引擎、优惠券锁定、最优解求解 |
| **第11章**（供给运营） | **商品上架、供应商同步、运营编辑（16.5.6 新增）** |
| **第12章**（计价系统） | 四层价格模型、试算接口、快照生成 |
| **第13章**（搜索导购） | **Query→Recall→Rank→Hydrate链路（16.5.7 新增）** |
| **第14章**（购物车结算） | **未登录加购、登录合并、Saga编排（16.5.7 新增）** |
| **第15章**（订单系统） | **状态机、Saga模式、幂等性（16.5.7、16.5.8 新增）** |
| **第16章**（支付系统） | 支付创建、回调处理、对账流程 |

**后续演进提示**：如果后续继续扩展本书，可以在本章基础上继续展开系统演进与重构、团队协作与工程实践等主题。

---

### 给读者的建议

1. **不要盲目照搬架构**
   - 根据团队规模调整（10人团队不需要12个微服务）
   - 根据业务特点优化（B2C和B2B2C差异大）
   - 根据发展阶段选择（初创期先单体，成熟期再拆分）

2. **架构是演进出来的**
   - 先简单方案（单体应用、MySQL单表）
   - 再根据瓶颈优化（QPS瓶颈→缓存，数据量瓶颈→分库分表）
   - 避免过度设计（YAGNI原则：You Aren't Gonna Need It）

3. **ADR是宝贵财富**
   - 记录决策过程（不只是结果）
   - 记录备选方案（为什么不选A而选B）
   - 记录权衡取舍（有什么优点和缺点）
   - 定期Review（每季度回顾，持续优化）

4. **从错误中学习**
   - 本章的"踩过的坑"是避坑指南
   - 建立错误知识库（每个错误都是学习机会）
   - 持续改进（错误 → 规则 → 自动化检查）

5. **关注业务价值**
   - 技术服务于业务（不是为了炫技）
   - 优先解决业务痛点（性能瓶颈、稳定性问题）
   - 量化技术收益（P99延迟降低、QPS提升、成本节省）

---

### 关键数据回顾

| 指标 | 数值 | 说明 |
|------|------|------|
| **团队规模** | 200+人 | 前台60、中台80、基础设施30、数据20、测试10 |
| **日订单量** | 200万（正常）/ 1000万（大促） | 大促5倍流量 |
| **服务数量** | 12个核心服务 + 3个聚合服务 | 按业务能力拆分，单一职责 |
| **ADR数量** | 13个 | 记录重大架构决策 |
| **响应时间** | P99 < 200ms（正常）/ 500ms（大促） | 多级缓存优化 |
| **可用性** | 99.95%（核心链路） | 多层防护 |
| **代码覆盖率** | > 80% | 单元测试 + 集成测试 |

---

**导航**：[返回目录](../SUMMARY.md) | [上一章：第16章](../part2/transaction/chapter15.md) | [书籍主页](../README.md)
