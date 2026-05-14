# 多品类购买链路统一设计与接口盘点

日期：2026-05-13  
范围：`shopee-api-server/prehandler` 中充值/账单、OTA、电影票、券类商品相关 C 端入口  
目标：先完成现有接口到标准购买阶段的映射，再确定低风险改造批次

## 背景

当前 C 端购买链路按品类逐步演进，入口散落在通用 `item/order/payment/refund/aggregation`、`local-service`、`evoucher/giftcard`、`movie`、`flight/bus/train/ferry/hotel` 等路由组里。搜索展示、详情聚合、资源确认、支付履约和退款规则都存在品类差异：

- 充值/账单偏虚拟商品，核心是账号校验、账单查询、创单快照、支付和履约回调。
- OTA 偏实时资源，核心是查价查库存、锁资源、预订状态轮询、支付超时释放、行程类退款规则。
- 电影票偏座位资源，核心是场次/座位图、占座、座位释放、支付后出票。
- 券类商品偏库存与核销，核心是券搜索/详情、库存、下单、支付后发券/核销、售后约束。

更大的背景是 DP 要从单个业务系统走向平台化。平台化不是简单支持更多品类，而是把 Search、Detail、Booking、Order、Checkout、Payment、Fulfillment、Refund 这些购买能力沉淀成可复用的平台能力。如果每个品类继续独立实现一套购买链路，新品类接入、支付履约复用、退款规则、监控排障和补偿治理都会被品类分支拖住。

因此，多品类流程统一是 DP 平台化的前置条件。统一的目标不是把所有前端接口合成一个大 API，而是统一购买生命周期、状态语义、订单快照、幂等、补偿和观测。前端品类入口可以保留差异，后端必须把共性的阶段能力收敛出来，让新品类从“重新搭链路”变成“实现阶段 adapter”。

统一链路不要求一次性替换老接口。第一阶段应保留现有 URL 和响应结构，先在后端建立标准阶段模型、路由标签、品类适配器和统一观测。

## 面试叙事材料

### 一句话概括

这个项目要解决的是多品类扩张以后，充值/账单、酒店/机票、电影票、券类商品各自长出一套购买流程，导致接口割裂、重复建设、状态不一致、履约和售后难排查的问题。我的做法是先不强行统一对外 API，而是把 C 端购买链路抽象成 `Search / Detail / Cart / Booking / Order / Checkout / Payment / Fulfillment / Refund` 标准阶段，再通过品类 adapter 保留差异、统一生命周期。

从平台化角度看，这件事的本质是把 DP 从“按品类堆业务逻辑”升级为“按购买阶段提供平台能力”。新品类不再复制一套完整购买链路，而是接入已有阶段：Search 负责召回，Detail 负责商品和规则，Booking 负责资源锁定，Order 固化快照，Checkout/Payment 负责收银，Fulfillment/Refund 负责履约售后闭环。

### 出发点

最早数字商品链路主要服务充值、账单这类虚拟商品，链路相对简单：查商品、校验账号、创单、支付、履约。随着业务扩展到酒店、机票、电影票、券类商品，链路复杂度明显上升：

- 充值/账单关注账号校验、账单查询、支付后充值履约。
- 酒店/机票关注实时价格、库存、供应商预订、超时释放。
- 电影票关注场次、座位图、占座、锁座过期。
- 券类商品关注库存、发券、核销、退款规则。

如果每个品类继续在 `prehandler` 和 handler 里单独堆接口，短期上线很快，但长期会拖慢新品类接入，并且让支付、退款、履约这些核心链路越来越难统一排查。因此出发点不是单纯重构，而是支撑业务多品类扩张，让新品类从“重新搭一套购买链路”变成“接入标准阶段能力”。

所以这个项目放在 DP 平台化背景下看，是平台治理的一部分：把品类差异留在 adapter，把通用能力沉淀到阶段模型、状态机、快照、幂等和补偿任务里。这样平台团队可以持续增强公共能力，而不是每次新品类上线都改一遍支付、履约、退款和监控。

### 核心挑战

第一个挑战是业务模型差异大。充值/账单没有强资源锁定，电影票要锁座，机票/酒店要实时查价查库存，券类商品又是库存和核销模型。如果直接强行统一一个大接口，很容易把所有特殊字段塞到一个复杂请求里。

应对方式是先统一阶段语义，而不是先统一 API 形态。比如 Booking 阶段在电影票里是占座，在机票里是创建供应商 booking，在酒店里短期是 pre-check/锁房，在充值账单里可以是空实现或轻量校验。这样统一的是生命周期，不抹平品类差异。

第二个挑战是老接口和前端强依赖。现有路由里已经有 `/item/flight/search`、`/item/movie/seats`、`/order/create`、`/aggregation/checkout`、`/refund/place` 等大量历史入口，直接换成一套新 API 风险很高。

应对方式是渐进式改造：老 URL 和响应结构不动，先做接口盘点和阶段映射，再补统一阶段标签、统一上下文和 adapter。这样可以在不影响前端的前提下收敛后端链路。

第三个挑战是价格、资源和售后规则必须固化。机票、酒店、电影票的价格和库存会变化，如果 Detail 阶段看一次规则，Order 阶段再查一次，Refund 阶段又重算一次，容易出现前后不一致。

应对方式是引入 `OrderSnapshot`：创单时固化商品快照、价格快照、资源快照、履约快照和售后快照。后续 Checkout、Payment、Fulfillment、Refund 优先基于订单快照推进，而不是回头依赖实时详情。

第四个挑战是回调乱序、重复和失败补偿。支付回调、供应商 booking 回调、履约回调、退款回调都可能重复、乱序或超时。尤其资源型商品，如果支付失败但资源没有释放，或者支付成功但履约失败，会造成用户体验和资源占用问题。

应对方式是用统一状态机、幂等键和补偿任务推进闭环。回调不直接覆盖状态，而是根据当前状态做合法迁移；补偿任务负责 booking 超时释放、支付失败释放资源、履约失败重试、退款状态回查。

### 推进方式

第一步是接口盘点，把现有 `prehandler` 里的入口按标准阶段归类：Search 负责召回和导购，Detail 负责详情和实时规则，Booking 负责查价查库存和资源锁定，Order 负责固化快照，Checkout 负责支付页聚合，Payment 负责支付请求，Fulfillment 负责履约推进，Refund 负责售后。

第二步是定义统一阶段模型，例如 `PurchaseContext`、`ProductDetailSnapshot`、`BookingSession`、`OrderSnapshot`、`CheckoutContext`、`RefundContext`。这些模型先作为后端内部语义，不直接要求前端使用。

第三步是确定改造顺序。没有先改最复杂的 Search/Detail，而是优先改 Booking，因为多品类差异最大、风险最高的是资源锁定、过期释放和查询一致性。推荐顺序是电影票、轮渡/巴士/火车、机票、酒店，最后再收敛充值/账单和券类的 Detail、OrderSnapshot、Fulfillment。

第四步是复用已有能力。`aggregation/checkout` 已经在做订单详情、支付渠道、促销、voucher、coin、价格计算的聚合，因此它应作为统一 Checkout 阶段的现有基础，只补阶段标签和输入快照，不重写。

### 收益

业务收益：

- 新品类接入从“重新搭一套购买链路”变成“实现标准阶段 adapter”，接入成本降低。
- DP 从品类业务系统演进为购买平台，平台能力可以被充值/账单、OTA、电影票、券类商品复用。
- C 端购买体验更一致，前端理解和接入成本降低。
- 支付、退款、履约规则更统一，减少用户感知不一致。

工程收益：

- `prehandler` 不再继续无限堆品类流程，路由层只负责注册和打标签。
- 通用能力沉淀到 Search、Detail、Booking、Order、Checkout 等阶段，减少重复代码。
- 订单快照统一后，Checkout、Payment、Refund 不需要重复回查易变详情，逻辑更稳定。

稳定性收益：

- Booking 超时释放、支付失败释放资源、履约失败重试、退款状态回查都有统一补偿路径。
- 回调链路通过状态机和幂等推进，减少重复回调、乱序回调造成的问题。
- 通过 `purchase_stage`、`purchase_domain`、`capability` 做统一监控，排查问题时可以快速定位是 Search 慢、Booking 失败、Payment 异常还是 Fulfillment 卡住。

面试中可以强调：这个项目的关键不是做一个很大的统一接口，而是先统一生命周期和状态语义，再用 adapter 保留品类差异，最后逐步沉淀快照、幂等、补偿和观测能力。

## 标准阶段定义

| 阶段 | 边界 | 必须固化的信息 | 当前常见来源 |
| --- | --- | --- | --- |
| Search | 商品召回、筛选排序、导购价格、首页/推荐/关键词 | `search_context`、候选商品、展示价、库存/可售粗粒度标记 | item/list、flight/search、hotel/search、local-service/search、movie now-showing |
| Detail | 商品详情、实时价格/库存、营销、履约/退款规则、表单配置 | `product_detail_snapshot`、实时价/库存、限制规则、售后规则 | item/detail、bill_info、flight form-config/addons、hotel details/extra-info、movie seat-map |
| Cart | 购物车/结算前价格计算、可用券、购买限制、重复单预检 | `cart_snapshot`、优惠试算、限购、金额结构 | pricing/cart/calculate、final-price、duplicate/precheck、purchase_limit |
| Booking | 供应商查价查库存、锁房/占座/订票确认、预订状态、超时释放 | `booking_id`、`booking_status`、`expire_time`、资源锁定快照、幂等键 | flight/bus/ferry/train/movie booking、hotel pre-check |
| Order | 创单并固化商品、价格、资源、履约、售后快照 | `order_id`、order item snapshot、booking/resource ref、refund policy snapshot | order/create、dsn-create、receive |
| Checkout | checkout 聚合、支付渠道、渠道优惠、voucher/coin 自动推荐、变价 | `checkout_context`、默认支付渠道、优惠计算、最终应付 | aggregation/checkout、price_change、auto_apply_voucher |
| Payment | 支付请求、支付状态、合单支付、独立收银台 | `payment_attempt`、channel、final_price、payment_status | order/pay、merge_pay、payment/channels |
| Fulfillment | 支付后履约、出票/发券/充值、补偿重试、履约状态查询 | `fulfillment_status`、供应商履约号、发货/出票明细 | order/complete、fulfill/retry、last-fulfillment、redeem |
| Refund | 可退规则、退款物品、退款原因、退款申请、退款阻断 | `refund_policy_snapshot`、return_items、refund_id、block_info | refund/place、flight/bus/hotel return info、refund tickets |

## 现有接口盘点

### 通用购买能力

这些接口已经跨品类复用，是统一链路的第一层底座。

| 阶段 | 路由 | Handler | 说明 |
| --- | --- | --- | --- |
| Cart | `POST /digital-product/api/pricing/cart/calculate` | `pricing.GetCartPricing` | PDP/Cart 价格计算入口 |
| Detail/Cart | `POST /digital-product/api/item/final-price` | `item.FinalPriceCalculate` | 商品最终价格试算 |
| Checkout | `POST /digital-product/api/aggregation/checkout` | `aggregation.CheckoutAggregation` | 当前最接近统一 checkout 编排的实现，已按阶段并发聚合订单、支付渠道、促销、coin、voucher |
| Checkout | `POST /digital-product/api/aggregation/price_change` | `aggregation.PriceChangeAggregation` | 支付页变价与优惠重算 |
| Checkout | `POST /digital-product/api/aggregation/auto_apply_voucher` | `autoApplyVoucherAggregation` | 自动用券与预选 |
| Order/Checkout | `POST /digital-product/api/aggregation/one_click_order` | `aggregation.OneClickOrder` | 一键下单聚合入口，适合沉淀为统一 Order adapter 的先行样本 |
| Payment | `GET /digital-product/api/payment/channels` | `order.GetPaymentChannels` | 老支付渠道入口 |
| Payment | `POST /digital-product/api/payment/v2/channels` | `order.GetPaymentChannelsV2` | 新支付渠道入口 |
| Payment | `GET /digital-product/api/payment/handling-fee` | `order.GetPaymentChannelHandlingFee` | 支付手续费 |
| Payment | `GET /digital-product/api/cashier/is-independent-payment` | `cashier.IsSupportedIndependentPayment` | 独立收银台能力判断 |
| Order | `POST /digital-product/api/order/create` | `order.Create` | 通用创单，当前承载大量品类差异参数 |
| Order | `POST /digital-product/api/order/dsn-create` | `order.DsnCreate` | DSN 创单 |
| Payment | `POST /digital-product/api/order/pay` | `order.Pay` | 通用支付 |
| Payment | `POST /digital-product/api/order/merge_pay` | `order.MergePay` | 合单支付 |
| Order | `GET /digital-product/api/order/detail` / `GET /v2/detail` | `order.GetDetail` / `order.GetDetailV2` | 订单详情 |
| Order | `POST /digital-product/api/order/cancel` | `order.Cancel` | 订单取消 |
| Fulfillment | `POST /digital-product/api/order/complete` | `order.CompleteOrder` | 订单完成回调/推进 |
| Fulfillment | `GET /digital-product/api/order/fulfill/retry` | `order.FulfillRetry` | 履约重试 |
| Fulfillment | `GET /digital-product/api/order/last-fulfillment/info` | `order.LastFulfillmentInfo` | 最近履约信息 |
| Refund | `POST /digital-product/api/refund/place` | `order.PlaceRefundOrder` | 通用退款申请 |
| Refund | `GET /digital-product/api/refund/block_info` | `order.RefundBlockInfo` | 退款阻断信息 |
| Refund | `GET /digital-product/api/order/refund/tickets` | `refund.GetTickets` | 用户可退票据列表 |
| Refund | `GET /digital-product/api/order/refund/ticket-detail` | `refund.GetDetail` | 退款票据详情 |

统一落点：这些接口不应被拆散，而应作为标准阶段的 shared adapter。短期先给它们补阶段标签和统一上下文，长期把品类分支从 `order.Create` / `order.Pay` 参数校验中迁到品类 adapter。

### 充值/账单

| 阶段 | 路由 | Handler | 现状判断 |
| --- | --- | --- | --- |
| Search | `GET /digital-product/api/item/list` / `/v2/list` | `handler.GetItemList` / `GetItemListV2` | 通用商品列表，用于充值/账单商品召回 |
| Search | `GET /digital-product/api/item/top-up/list` | `handler.GetTopUpItemList` | top-up 专用列表 |
| Search | `GET /digital-product/api/item/top-up/subscriber/profile/list` | `handler.GetTopUpItemListWithSubscriberProfile` | 召回时携带用户订阅档案 |
| Search | `GET /digital-product/api/item/carrier/search` | `handler.GetCarrierSearch` | carrier 搜索 |
| Search/Detail | `GET /digital-product/api/entrance/bill/config` / `v2/bill/config` | `handler.GetBillConfig` / `GetBillConfigV2` | 账单筛选与入口配置 |
| Detail | `GET /digital-product/api/item/brief` | `handler.GetItemBrief` | 商品简要信息 |
| Detail | `GET /digital-product/api/item/status` | `handler.GetItemStatus` | 商品状态 |
| Detail | `GET /digital-product/api/item/minimum-spend` | `handler.GetMinimumSpend` | 最小消费限制 |
| Detail | `GET /digital-product/api/item/bill_info` | `handler.GetBillInfo` | 老账单查询 |
| Detail | `GET /digital-product/api/item/v2/bill_info` | `bill.QueryBill` | 新账单查询 |
| Detail | `GET /digital-product/api/item/bill/flow/query_bill` | `bill.FlowQueryBill` | Bill flow 查询 |
| Detail | `GET /digital-product/api/item/pln-prepaid/inquiry` | `bill.PLNPrepaidInquiry` | PLN prepaid PDP 并行查询客户信息 |
| Detail | `GET /digital-product/api/item/emoney-info` | `handler.GetEMoneyInfo` | e-money 商品信息 |
| Detail | `GET /digital-product/api/item/account/validation` | `account.ValidateAccount` | 通用账号校验，适合纳入 Detail validation |
| Detail | `GET /digital-product/api/item/topup-account/validation` | `handler.ValidateTopupAccount` | 老 top-up 账号校验 |
| Cart | `POST /digital-product/api/order/duplicate_order_precheck` | `order.DuplicateOrderPrecheck` | 重复单预检 |
| Cart | `POST /digital-product/api/order/purchase_limit_precheck` | `order.PurchaseLimitPrecheck` | 限购预检 |
| Order | `POST /digital-product/api/order/create` | `order.Create` | 充值/账单通用创单 |
| Payment | `POST /digital-product/api/order/pay` | `order.Pay` | 充值/账单通用支付 |
| Fulfillment | `POST /digital-product/api/order/e-money/bca/fulfillment/create` | `order.EmoneyFulfillmentCreate` | e-money 履约创建 |
| Fulfillment | `POST /digital-product/api/order/e-money/bca/topup` | `order.TopupEMoney` | e-money 充值履约 |
| Fulfillment | `POST /digital-product/api/order/e-money/bca/fulfillment_result/notify` | `order.NotifyEMoneyFulfillmentResult` | e-money 履约结果通知 |
| Fulfillment | `POST /digital-product/api/order/e-money/bca/reverse` | `order.ReverseEMoneyTopup` | e-money 冲正 |
| Fulfillment | `POST /digital-product/api/topuptransfer/place` | `order.RequestTransfer` | top-up transfer |
| Refund | `POST /digital-product/api/refund/place` | `order.PlaceRefundOrder` | 通用退款 |
| Refund | `GET /digital-product/api/refund/block_info` | `order.RefundBlockInfo` | 退款阻断 |

Bill Center 更像账单域的用户资产和订阅管理，不直接等同购买阶段，但会影响 Search/Detail/Checkout：

| 能力 | 路由 | 映射阶段 |
| --- | --- | --- |
| 支持品类/账号/新用户区 | `/digital-product/api/bill-center/support_category_list`、`/account/list`、`/new-user-zone` | Search/Detail 输入 |
| 账单历史/账单信息 | `/bill-record`、`/account/bill-info/list` | Detail/Order 辅助 |
| 订阅管理 | `/subscription/list`、`/subscription/detail`、`/subscription/preview`、`/subscription/auth-pay` | Cart/Checkout/Payment 辅助 |

统一判断：充值/账单可以最早接入统一链路，因为资源锁定弱。Booking 阶段通常为空实现或轻量 validation/query_bill，不应强制要求供应商预订。

### OTA：机票、酒店、巴士、火车、轮渡

#### Flight

| 阶段 | 路由 | Handler | 现状判断 |
| --- | --- | --- | --- |
| Search | `GET /digital-product/api/item/flight/keywords` | `flight.GetKeywords` | 关键字召回 |
| Search | `GET /digital-product/api/item/flight/v2/keywords` | `flight.GetKeywordsV2` | v2 关键字召回 |
| Search | `POST /digital-product/api/item/flight/search` | `flight.Search` | 老航班搜索 |
| Search | `POST /digital-product/api/item/flight/v2/search` | `flight.SearchV2` | 新航班搜索 |
| Search | `GET /digital-product/api/item/flight/lowest-price-calendar` | `flight.GetLowestPriceCalendar` | 导购低价日历 |
| Detail | `POST /digital-product/api/item/flight/form-config` | `flight.GetFormConfig` | 乘机人表单规则 |
| Detail | `POST /digital-product/api/item/flight/addons` | `flight.GetAddOnOptions` | 行李等增值项 |
| Detail | `GET /digital-product/api/item/flight/booking/rules` | `flight.GetBookingRules` | 预订规则 |
| Detail/Refund | `POST /digital-product/api/item/flight/refund/policies` | `flight.GetRefundPolicy` | 航班退款政策 |
| Booking | `POST /digital-product/api/item/flight/create-booking` | `flight.CreateBooking` | 供应商预订/锁资源 |
| Booking | `POST /digital-product/api/item/flight/query-booking` | `flight.QueryBookingStatus` | 预订状态查询 |
| Detail/Cart | `POST /digital-product/api/item/flight/insurance` | `flight.QueryInsuranceItems` | 保险商品查询 |
| Detail/Cart | `POST /digital-product/api/item/flight/verify-insurance` | `flight.VerifyInsuranceItems` | 保险校验 |
| Fulfillment | `GET /digital-product/api/order/flight-change/log` | `order.GetFlightChangeLog` | 航变日志 |
| Fulfillment | `GET /digital-product/api/order/flight/insurance/detail` | `order.GetFlightInsuranceDetail` | 保险履约详情 |
| Refund | `GET /digital-product/api/refund/flight/return_items` | `order.GetFlightReturnItems` | 可退航段/乘客 |
| Refund | `POST /digital-product/api/refund/flight/static/upload` | `handler.UploadAttachmentToS3` | 航班退款附件 |

#### Hotel

| 阶段 | 路由 | Handler | 现状判断 |
| --- | --- | --- | --- |
| Search | `GET /digital-product/api/item/hotel/current-city` | `hotel.GetCitySearchKey` | 当前城市 |
| Search | `GET /digital-product/api/item/hotel/keywords` | `hotel.GetKeywords` | 酒店关键词 |
| Search | `POST /digital-product/api/item/hotel/search` | `hotel.SearchHotel` | 酒店搜索 |
| Search | `GET /digital-product/api/item/hotel/hot-cities` | `hotel.GetHotCities` | 热门城市 |
| Search | `POST /digital-product/api/item/hotel/search-filter-resource` | `hotel.GetHotelSearchFilterResource` | 筛选资源 |
| Detail | `POST /digital-product/api/item/hotel/details` | `hotel.Details` | 酒店详情 |
| Detail | `POST /digital-product/api/item/hotel/extra-info` | `hotel.ExtraInfo` | 房型/价格/规则补充 |
| Detail | `GET /digital-product/api/item/hotel/gallery` | `hotel.GetGallery` | 图集 |
| Detail | `GET /digital-product/api/item/hotel/multilingual-info` | `hotel.MultilingualInfo` | 多语言信息 |
| Detail | `POST /digital-product/api/item/hotel/reviews` | `hotel.Reviews` | 评论 |
| Booking | `POST /digital-product/api/item/hotel/pre-check` | `hotel.PreCheck` | 下单前查价查库存/预检 |
| Booking | `GET /digital-product/api/order/hotel/duplication-check` | `hotel.BookingHotelDupCheck` | 重复预订检查 |
| Refund | `GET /digital-product/api/refund/hotel/return_info` | `handler.GetHotelReturnInfo` | 酒店售后退改信息 |

酒店目前没有在 `prehandler` 暴露独立 booking create，预订资源应在 PreCheck 和 Order server 创单链路中完成。统一时应把 `pre-check` 作为 Booking adapter 的第一阶段，后续再识别 Order server 里的实际锁房点。

#### Bus

| 阶段 | 路由 | Handler | 现状判断 |
| --- | --- | --- | --- |
| Search | `GET /digital-product/api/item/bus/search` | `bus.Search` | 老巴士搜索 |
| Search | `POST /digital-product/api/item/bus/v2/search` | `busv2.Search` | 新巴士搜索 |
| Search | `POST /digital-product/api/item/bus/v2/keywords` | `busv2.SearchKeywords` | 关键词 |
| Search | `POST /digital-product/api/item/bus/v2/filter/statistics` | `busv2.FilterStatistics` | 筛选统计 |
| Detail | `GET /digital-product/api/item/bus/seat-map` | `bus.GetSeatMap` | 座位图 |
| Detail | `POST /digital-product/api/item/bus/v2/detail` | `busv2.Detail` | 行程详情 |
| Detail | `POST /digital-product/api/item/bus/v2/seatmap` | `busv2.SeatMap` | v2 座位图 |
| Detail/Refund | `GET /digital-product/api/item/bus/refund-policy` | `bus.GetRefundPolicy` | 巴士退款政策 |
| Booking | `POST /digital-product/api/item/bus/create-booking` | `bus.CreateBooking` | 老巴士预订 |
| Booking | `POST /digital-product/api/item/v2/bus/create-booking` | `bus.CreateBookingV2` | v2 巴士预订 |
| Booking | `GET /digital-product/api/item/bus/query-booking` | `bus.QueryBooking` | 预订查询 |
| Booking | `POST /digital-product/api/item/bus/cancel-booking` | `bus.CancelBooking` | 释放预订 |
| Refund | `GET /digital-product/api/refund/bus/return_items` | `order.GetBusReturnItems` | 可退行程 |
| Refund | `GET /digital-product/api/refund/bus/return_reason` | `order.GetBusReturnReason` | 退款原因 |

#### Train

| 阶段 | 路由 | Handler | 现状判断 |
| --- | --- | --- | --- |
| Search | `GET /digital-product/api/item/train/station` | `handler.GetStationList` | 车站列表 |
| Search | `GET /digital-product/api/item/train/list` | `handler.GetSectorTrainList` | 老火车列表 |
| Detail | `POST /digital-product/api/item/train/select_train` | `handler.GetSelectedTrainInfo` | 老火车车次详情 |
| Detail | `GET /digital-product/api/item/train/seat_map` | `handler.GetTrainSeatMaps` | 老座位图 |
| Booking | `POST /digital-product/api/item/train/select_seat` | `handler.SelectSeatHandler` | 老选座/占座 |
| Booking | `POST /digital-product/api/item/train/cancel_seat` | `handler.CancelTrainSeatHandler` | 老取消占座 |
| Search | `GET /digital-product/api/item/train/v2/hot-keywords` | `handler.HotKeywords` | v2 热词 |
| Search | `GET /digital-product/api/item/train/v2/keywords` | `handler.GetTrainKeywords` | v2 关键词 |
| Search | `POST /digital-product/api/item/train/v2/list` | `handler.GetSectorTrainListV2` | v2 车次列表 |
| Detail | `POST /digital-product/api/item/train/v2/select_train` | `handler.GetSelectedTrainInfoV2` | v2 车次详情 |
| Detail | `GET /digital-product/api/item/train/v2/seat_map` | `railway.GetWagonSeatMap` | VN v2 座位图 |
| Booking | `POST /digital-product/api/item/train/v2/select_seat` | `railway.SelectSeat` | VN v2 选座 |
| Booking | `POST /digital-product/api/item/train/v2/booking` | `railway.CreateBooking` | VN v2 订票确认 |
| Booking | `GET /digital-product/api/item/train/v2/booking_result` | `railway.GetBookingResult` | VN v2 预订结果 |
| Fulfillment | `GET /digital-product/api/order/train/fulfillment/additional_info` | `order.GetTrainFulfillmentInfo` | 火车履约附加信息 |

#### Ferry

| 阶段 | 路由 | Handler | 现状判断 |
| --- | --- | --- | --- |
| Search | `GET /digital-product/api/item/ferry/config/metadata` | `ferry.GetMetadata` | 港口/船型/乘客类型等 metadata |
| Search | `GET /digital-product/api/item/ferry/schedules/daily` | `ferry.GetDailySchedules` | 日期班次 |
| Detail/Booking | `POST /digital-product/api/item/ferry/quote` | `ferry.Quote` | 询价/可用 offer，介于 Detail 与 Booking |
| Booking | `POST /digital-product/api/item/ferry/booking` | `ferry.CreateBooking` | 预订并返回 `order_plan_id`、状态、支付截止时间 |
| Booking | `GET /digital-product/api/item/ferry/booking` | `ferry.GetBookingStatus` | 查询预订状态 |

OTA 统一判断：OTA 需要优先抽 Booking adapter，因为资源锁定、超时释放和查询结果一致性是最大差异点。Flight/Bus/Ferry/Train 已有显式 booking；Hotel 应从 `pre-check` 进入，再补 Order server 实际锁房链路盘点。

### 电影票

| 阶段 | 路由 | Handler | 现状判断 |
| --- | --- | --- | --- |
| Search | `GET /digital-product/api/item/movie/film/now-showing` | `movie.NowShowingFilmHandler` | 热映电影 |
| Search | `GET /digital-product/api/item/movie/film/coming-soon` | `movie.ComingSoonFilmHandler` | 即将上映 |
| Search | `GET /digital-product/api/item/movie/cinemas` | `movie.CinemasHandler` | 影院列表/搜索 |
| Search | `GET /digital-product/api/item/movie/film/cinemas` | `movie.CinemasHandler` | 影片下影院 |
| Search | `GET /digital-product/api/item/movie/film/dates` | `movie.FilmDatesHandler` | 可售日期 |
| Search | `POST /digital-product/api/item/movie/cinema-list` | `movie.CinemaList` | 影院列表 |
| Detail | `GET /digital-product/api/item/movie/film` | `movie.FilmDetailHandler` | 影片详情 |
| Detail | `GET /digital-product/api/item/movie/film/sessions` | `movie.FilmSessionsHandler` | 场次 |
| Detail | `GET /digital-product/api/item/movie/seat-map` / `/seat-map/v2` | `movie.SeatMap` / `SeatMapV2` | 座位图 |
| Detail | `GET /digital-product/api/item/movie/cinema-snacks` | `movie.GetCinemaSnackList` | 影院小食 |
| Detail | `GET /digital-product/api/item/movie/session-snacks` | `movie.GetSessionSnackList` | 场次小食 |
| Detail | `GET /digital-product/api/item/movie/snack-info` | `movie.GetSnackInfo` | 小食详情 |
| Booking | `POST /digital-product/api/item/movie/seats` | `movie.SelectSeat` | 占座，返回 `booking_id`、金额、过期时间 |
| Booking | `DELETE /digital-product/api/item/movie/seats` | `movie.CancelSeat` | 释放占座 |
| Booking | `POST /digital-product/api/order/movie/seats` | `movie.Booking` | 订单域电影 booking |
| Booking | `DELETE /digital-product/api/order/movie/seats` | `movie.CancelBooking` | 订单域取消 booking |
| Order | `GET /digital-product/api/order/success-movie` | `order.GetSuccessOrders` | 成功订单 |
| Fulfillment | `PUT /digital-product/api/order/redeem/movie/snack` | `movie.UpdateSnackStatus` | 小食核销/状态更新 |
| Checkout | `POST /digital-product/api/aggregation/checkout` | `aggregation.CheckoutAggregation` | 已有电影票优惠适配，包含 movie ticket 到 promotion ticket 转换 |

统一判断：电影票应作为 Booking adapter 的样板之一。它有明确 `booking_id` 和 `expired_time`，并且占座/释放链路已暴露在 item 和 order 两个域，适合先定义统一资源锁生命周期。

### 券类商品

券类商品在当前路由里主要包括 `evoucher`、`local-service` 的 OPV/deal voucher、`giftcard`，以及通用 promotion voucher。

#### Evoucher

| 阶段 | 路由 | Handler | 现状判断 |
| --- | --- | --- | --- |
| Search | `POST /digital-product/api/item/evoucher/hot-merchants` | `evoucher.HotMerchantsHandler` | 热门商户 |
| Search | `POST /digital-product/api/item/evoucher/v3/merchants` | `evoucher.MerchantListV3Handler` | 商户列表 |
| Search | `GET /digital-product/api/item/evoucher/merchants/search` | `evoucher.MerchantsSearchHandler` | 商户搜索 |
| Search | `POST /digital-product/api/item/evoucher/v3/deals` | `evoucher.DealListV3Handler` | 券 deal 列表 |
| Search | `POST /digital-product/api/item/evoucher/v4/deals` | `evoucher.DealListV4Handler` | v4 deal 列表 |
| Search | `GET /digital-product/api/item/evoucher/v4/search` | `evoucher.SearchV4Handler` | v4 搜索 |
| Search | `GET /digital-product/api/item/evoucher/trending-searches` | `evoucher.HotSearchKeywords` | 热词 |
| Detail | `GET /digital-product/api/item/evoucher/v3/merchant` | `evoucher.MerchantV3Handler` | 商户详情 |
| Detail | `GET /digital-product/api/item/evoucher/v3/deal` | `evoucher.DealV3Handler` | deal 详情 |
| Detail | `GET /digital-product/api/item/evoucher/v4/deal` | `evoucher.DealV4Handler` | v4 deal 详情 |
| Detail | `GET /digital-product/api/item/evoucher/v4/deal/denomination/promotion` | `evoucher.DealDenominationPromotionHandler` | 面额促销 |
| Detail | `GET /digital-product/api/item/evoucher/redeem-outlets` | `evoucher.GetOutletList` | 可核销门店 |
| Detail/Cart | `GET /digital-product/api/item/evoucher/check-stock` | `evoucher.CheckStockHandler` | 库存检查 |
| Fulfillment | `POST /digital-product/api/order/redeem/evoucher` | `redeem.Evoucher` | 券码核销 |
| Fulfillment | `POST /digital-product/api/order/redeem/evoucher/code` | `redeem.UpdateEvoucherCodeStatus` | 券码状态更新 |
| Detail/Order | `GET /digital-product/api/user-info/evoucher/user-details` | `uinfo.GetEvoucherUserDetails` | 用户填写资料 |

#### Local Service OPV / Deal Voucher

| 阶段 | 路由 | Handler | 现状判断 |
| --- | --- | --- | --- |
| Search | `GET /local-service/api/search` / `/v2/search` | `localService.SearchV2` | 本地服务搜索 |
| Search | `POST /local-service/api/v2/search` | `localService.GlobalSearch` | 全局搜索和无结果推荐 |
| Search | `GET /local-service/api/trending-searches` | `localService.HotSearchKeywords` | 热词 |
| Search | `GET /local-service/api/payment-vouchers` | `localService.PaymentVouchers` | OPV 列表 |
| Search | `GET /local-service/api/deal-vouchers` | `localService.DealVouchers` | deal voucher 列表 |
| Search | `GET /local-service/api/search/deal` | `localService.SearchDealVouchers` | deal voucher 搜索 |
| Search | `GET /local-service/api/search/payment-vouchers` | `localService.SearchPaymentVouchers` | OPV 搜索 |
| Search | `GET /local-service/api/brand/payment-vouchers` | `localService.BrandPaymentVouchers` | 品牌下 OPV |
| Search | `GET /local-service/api/outlet/payment-vouchers` | `localService.OutletPaymentVouchers` | 门店下 OPV |
| Detail | `GET /local-service/api/outlet/detail` | `localService.OutletDetail` | 门店详情 |
| Detail | `GET /local-service/api/payment-voucher/detail` / `/v2/payment-voucher/detail` | `localService.PaymentVoucherDetail` / `PaymentVoucherDetailV2` | OPV 详情 |
| Detail | `GET /local-service/api/deal-voucher/detail` / `/v2/deal-voucher/detail` | `localService.DealVoucherDetail` / `DealVoucherDetailV2` | Deal voucher 详情 |
| Detail | `GET /local-service/api/evoucher/redeem-outlets` | `localService.RedeemOutlets` | 可核销门店 |
| Detail/Cart | `GET /local-service/api/evoucher/stock` | `localService.VoucherStock` | 库存 |
| Cart | `GET /local-service/api/purchase-limit` | `localService.GetUserPromotionPurchaseLimit` | 限购 |
| Order | `POST /local-service/api/order/create` | `order.Create` | LS 共用通用创单 |
| Payment | `POST /local-service/api/order/pay` | `order.Pay` | LS 共用通用支付 |
| Payment | `GET /local-service/api/payment/channels` | `order.GetPaymentChannels` | LS 支付渠道 |
| Fulfillment | `GET /local-service/api/order/my-vouchers/v2` / `/v3` / `/v4` | `localService.GetUserVouchersV2/V3/V4` | 用户已购券 |
| Fulfillment | `GET /local-service/api/order/my-vouchers/code` | `localService.GetUserVoucherCode` | 用户券码 |
| Fulfillment | `PUT /local-service/api/order/redeem` | `localService.CodeRedeemStatus` | 核销状态 |
| Refund | `POST /local-service/api/refund/place` | `order.PlaceRefundOrder` | LS 退款 |
| Refund | `GET /local-service/api/order/refund/ticket-detail` | `refund.GetDetail` | LS 退款详情 |

#### Gift Card 与 Promotion Voucher

| 阶段 | 路由 | Handler | 现状判断 |
| --- | --- | --- | --- |
| Search | `GET /digital-product/api/giftcard/theme-hot-skins` | `giftcard.GetThemeHotSkins` | 礼品卡主题皮肤 |
| Search | `GET /digital-product/api/giftcard/hot-skins` | `giftcard.GetGiftCardHotSkins` | 热门皮肤 |
| Detail | `GET /digital-product/api/giftcard/detail` | `giftcard.GetGiftCardDetail` | 礼品卡详情 |
| Detail | `GET /digital-product/api/giftcard/code/shared/detail` | `giftcard.GetGiftCardSharedDetail` | 分享详情 |
| Checkout | `GET /digital-product/api/promotion/voucher/list` / `/v2/list` | `voucher.GetVoucherList` / `GetVoucherListV2` | 可用券列表 |
| Checkout | `POST /digital-product/api/promotion/voucher/auto-apply` | `voucher.AutoApplyVoucherV2` | Checkout 自动用券 |
| Checkout | `POST /digital-product/api/promotion/voucher/validation` | `voucher.ValidateVoucher` | 券校验 |
| Checkout | `POST /digital-product/api/promotion/voucher/select` | `voucher.SelectVoucher` | 选券 |
| Checkout | `POST /digital-product/api/promotion/voucher/prompt` | `voucher.PromptVoucher` | 用券提示 |
| Checkout | `POST /digital-product/api/promotion/voucher/claim` | `voucher.ClaimVoucher` | 领券 |

券类统一判断：券类 Search/Detail 最分散，但 Order/Payment/Refund 已较多复用通用链路。应先统一商品详情和库存语义，再接统一订单快照。

## 目标架构

### 一层：路由阶段标签

在不改变 URL 的前提下，为 `prehandler` 中目标接口补充统一阶段标签：

- `purchase_domain`：`bill_topup`、`ota_flight`、`ota_hotel`、`ota_bus`、`ota_train`、`ota_ferry`、`movie_ticket`、`evoucher`、`local_service_voucher`、`giftcard`。
- `purchase_stage`：`search`、`detail`、`cart`、`booking`、`order`、`checkout`、`payment`、`fulfillment`、`refund`。
- `capability`：`query_bill`、`quote`、`seat_hold`、`booking_create`、`booking_query`、`refund_policy`、`voucher_stock` 等细粒度动作。

标签先用于日志、监控、限流分组和迁移开关，不改变请求响应。

### 二层：标准阶段上下文

新增统一上下文模型，作为 adapter 之间传递的稳定语义，不直接要求前端使用：

| 模型 | 字段示例 | 说明 |
| --- | --- | --- |
| `PurchaseContext` | user、partner、region、client、trace、source、publish_id | 贯穿所有阶段 |
| `ProductCandidate` | product_key、category_id、carrier_id、display_price、availability_hint | Search 输出 |
| `ProductDetailSnapshot` | item、real_time_price、stock、marketing、fulfillment_rule、refund_rule | Detail 输出 |
| `CartSnapshot` | items、quantity、price_breakdown、promotion_preview、limit_result | Cart 输出 |
| `BookingSession` | booking_id/order_plan_id、status、expire_time、resource_snapshot、idempotency_key | Booking 输出 |
| `OrderSnapshot` | order_id、item_snapshot、price_snapshot、resource_snapshot、aftersale_snapshot | Order 输出 |
| `CheckoutContext` | order_id、pay_amount、channel_options、voucher/coin、risk_context | Checkout 输出 |
| `PaymentAttempt` | payment_id、order_id、channel、final_price、status | Payment 输出 |
| `FulfillmentSnapshot` | fulfillment_id、status、provider_ref、retry_policy | Fulfillment 输出 |
| `RefundContext` | refund_id、return_items、policy_snapshot、block_info、reason | Refund 输出 |

### 三层：品类 adapter

这里的 adapter 是一层后端适配代码，不是新的 HTTP 接口，也不是新的供应商 client。它的职责是把统一购买阶段的输入转换成当前品类已有的 handler/RPC 请求，再把返回结果转换成统一阶段模型。

换句话说，adapter 解决的是“统一阶段语义”和“品类历史实现”之间的翻译问题：

```text
统一阶段输入 -> 品类 adapter -> 现有 handler/service/RPC -> 品类 adapter -> 统一阶段输出
```

adapter 应该做的事情：

- 参数转换：把 `BookingInput` 转成 `cmd.MovieSelectSeatReq`、`cmd.FlightBookingReq`、`cmd.FerryCreateBooking_Request` 等已有请求。
- 结果归一：把不同品类返回的 `booking_id`、`order_plan_id`、`expired_time`、`payment_deadline` 统一成 `BookingSession`。
- 状态映射：把供应商或品类内部状态映射成统一状态，例如 `Booked`、`BookingExpired`、`Released`。
- 幂等处理：统一读取和传递 `client_request_token`、booking ref、order ref。
- 错误转换：把品类错误码转换成阶段错误，例如参数错误、资源不可用、价格变化、库存不足、供应商超时。
- 观测打点：统一上报 `purchase_domain`、`purchase_stage`、`capability`、耗时、错误码。

adapter 不应该做的事情：

- 不改变老接口响应结构，避免影响前端。
- 不把所有品类特殊逻辑塞进一个大 switch。
- 不替代 Order server、Item server、Promotion、Pricing 等已有服务。
- 不在 `prehandler` 里实现业务编排，`prehandler` 只负责路由注册和通用中间件。

具体例子：

- `MovieBookingAdapter`：接收统一 `BookingInput`，组装现有 `MovieSelectSeatReq`，调用电影票占座能力，返回统一 `BookingSession{booking_id, status, expire_time, resource_snapshot}`。
- `FerryBookingAdapter`：把 `quotation_id`、`selected_offers`、乘客/车辆信息转成 `FerryCreateBooking_Request`，调用已有 ferry booking RPC，返回统一 `BookingSession{order_plan_id, booking_status, payment_deadline}`。
- `FlightBookingAdapter`：把航班、乘机人、联系人、保险/行李等信息转成 `FlightBookingReq`，调用 `FlightBookTicket` 和 `FlightQueryBookingStatus`，同时保留供应商原始 booking status。
- `BillTopupBookingAdapter`：充值/账单没有真实资源锁定，可以实现为空 booking 或轻量 validation，把账号校验/账单查询结果作为 Detail/Cart 快照传给 Order。
- `VoucherDetailAdapter`：把 evoucher、local-service OPV/deal voucher、giftcard 的不同详情和库存接口统一成 `ProductDetailSnapshot`，但不要求它们使用相同的老响应字段。

每个品类 adapter 只实现自己真正需要的阶段：

| 品类 | 必选 adapter | 可空 adapter |
| --- | --- | --- |
| 充值/账单 | Search、Detail、Cart、Order、Checkout、Payment、Fulfillment、Refund | Booking 可为空实现 |
| Flight/Bus/Train/Ferry | Search、Detail、Booking、Order、Checkout、Payment、Fulfillment、Refund | Cart 可轻量化 |
| Hotel | Search、Detail、Booking(pre-check)、Order、Checkout、Payment、Refund | Fulfillment 依赖订单详情和供应商回调 |
| 电影票 | Search、Detail、Booking(seat hold)、Order、Checkout、Payment、Fulfillment、Refund | Cart 可复用 Detail/Checkout |
| 券类商品 | Search、Detail、Cart、Order、Checkout、Payment、Fulfillment、Refund | Booking 通常为空实现 |

接口形态建议：

```go
type StageAdapter interface {
    Domain() PurchaseDomain
    Stage() PurchaseStage
    Execute(ctx context.Context, in StageInput) (StageOutput, error)
}

type BookingAdapter interface {
    Create(ctx context.Context, in BookingInput) (BookingSession, error)
    Query(ctx context.Context, ref BookingRef) (BookingSession, error)
    Release(ctx context.Context, ref BookingRef) error
}
```

### 具体落地实施蓝图

这一节把方案落到当前代码结构里。原则是三条：

1. 老接口 URL、请求、响应先不变。
2. `prehandler` 只补注册、标签和灰度，不承载业务编排。
3. 新代码先以内部 package 的方式沉淀，逐步让旧 handler 调用。

#### 目录和文件规划

建议新增一个 `purchase` 目录放统一购买链路的内部抽象，避免继续把逻辑堆到 `prehandler` 或某个品类 handler。

| 文件 | 类型 | 责任 |
| --- | --- | --- |
| `shopee-api-server/purchase/stage/types.go` | 新增 | 定义 `PurchaseDomain`、`PurchaseStage`、`Capability`、`StageMeta` |
| `shopee-api-server/purchase/stage/registry.go` | 新增 | 维护 method + path 到阶段标签的映射 |
| `shopee-api-server/purchase/stage/registry_test.go` | 新增 | 测试核心路由能映射到正确阶段 |
| `shopee-api-server/middleware/purchase_stage_tag.go` | 新增 | 从 registry 读取阶段标签，写入 gin context 和日志上下文 |
| `shopee-api-server/middleware/purchase_stage_tag_test.go` | 新增 | 测试 middleware 对命中/未命中路由的行为 |
| `shopee-api-server/purchase/model/context.go` | 新增 | 定义 `PurchaseContext`，封装 user、partner、region、client、trace |
| `shopee-api-server/purchase/model/booking.go` | 新增 | 定义 `BookingInput`、`BookingRef`、`BookingSession`、统一 booking status |
| `shopee-api-server/purchase/model/snapshot.go` | 新增 | 定义 `ProductDetailSnapshot`、`OrderSnapshot`、`RefundContext` |
| `shopee-api-server/purchase/booking/adapter.go` | 新增 | 定义 `BookingAdapter` 接口和 adapter factory |
| `shopee-api-server/purchase/booking/movie_adapter.go` | 新增 | 电影票占座/释放适配 |
| `shopee-api-server/purchase/booking/ferry_adapter.go` | 新增 | 轮渡 quote/booking/query 适配 |
| `shopee-api-server/purchase/booking/bill_topup_adapter.go` | 新增 | 充值/账单空 booking 或轻量校验适配 |
| `shopee-api-server/purchase/booking/adapter_test.go` | 新增 | 测试 adapter factory 和状态映射 |
| `shopee-api-server/handler/movie/seat.go` | 修改 | 第二批灰度时接入 `MovieBookingAdapter` |
| `shopee-api-server/handler/ferry/booking.go` | 修改 | 第二批灰度时接入 `FerryBookingAdapter` |
| `shopee-api-server/prehandler/router.go` | 修改 | 注入 `PurchaseStageTagMiddleWare`，保持原路由不变 |

第一批只需要新增 `stage`、middleware 和测试，业务 handler 不改。第二批再接入 booking adapter。

#### 阶段标签 registry

先落一个静态 registry，覆盖目标范围内的核心路由。它不替代 gin 路由，只为观测和灰度提供稳定标签。

建议结构：

```go
package stage

type PurchaseDomain string
type PurchaseStage string
type Capability string

const (
    DomainBillTopup          PurchaseDomain = "bill_topup"
    DomainOTAFlight          PurchaseDomain = "ota_flight"
    DomainOTAFerry           PurchaseDomain = "ota_ferry"
    DomainMovieTicket        PurchaseDomain = "movie_ticket"
    DomainLocalServiceVoucher PurchaseDomain = "local_service_voucher"

    StageSearch      PurchaseStage = "search"
    StageDetail      PurchaseStage = "detail"
    StageCart        PurchaseStage = "cart"
    StageBooking     PurchaseStage = "booking"
    StageOrder       PurchaseStage = "order"
    StageCheckout    PurchaseStage = "checkout"
    StagePayment     PurchaseStage = "payment"
    StageFulfillment PurchaseStage = "fulfillment"
    StageRefund      PurchaseStage = "refund"
)

type StageMeta struct {
    Domain     PurchaseDomain
    Stage      PurchaseStage
    Capability Capability
}

type RouteKey struct {
    Method string
    Path   string
}
```

第一批 registry 先覆盖高价值路由：

| Method | Path | Domain | Stage | Capability |
| --- | --- | --- | --- | --- |
| `GET` | `/digital-product/api/item/v2/bill_info` | `bill_topup` | `detail` | `query_bill` |
| `GET` | `/digital-product/api/item/account/validation` | `bill_topup` | `detail` | `account_validation` |
| `POST` | `/digital-product/api/order/create` | `bill_topup` | `order` | `create_order` |
| `POST` | `/digital-product/api/order/pay` | `bill_topup` | `payment` | `pay_order` |
| `POST` | `/digital-product/api/item/movie/seats` | `movie_ticket` | `booking` | `seat_hold` |
| `DELETE` | `/digital-product/api/item/movie/seats` | `movie_ticket` | `booking` | `seat_release` |
| `POST` | `/digital-product/api/item/ferry/quote` | `ota_ferry` | `booking` | `quote` |
| `POST` | `/digital-product/api/item/ferry/booking` | `ota_ferry` | `booking` | `booking_create` |
| `GET` | `/digital-product/api/item/ferry/booking` | `ota_ferry` | `booking` | `booking_query` |
| `POST` | `/digital-product/api/aggregation/checkout` | `bill_topup` | `checkout` | `checkout_aggregation` |
| `POST` | `/digital-product/api/refund/place` | `bill_topup` | `refund` | `place_refund` |
| `GET` | `/local-service/api/payment-voucher/detail` | `local_service_voucher` | `detail` | `voucher_detail` |
| `GET` | `/local-service/api/evoucher/stock` | `local_service_voucher` | `detail` | `voucher_stock` |

`/order/create`、`/order/pay`、`/aggregation/checkout`、`/refund/place` 是跨品类共享路由，第一批可以先标成 `bill_topup` 作为默认值，后续通过 request 中的 `category_id`、`carrier_id`、order detail 再二次修正 domain。这样不会阻塞观测落地。

#### middleware 接入方式

新增 `PurchaseStageTagMiddleWare` 后，在 `RouterInit` 里放在 `RequestTagMiddleWare` 后、限流前：

```go
engine.Use(middleware.RequestTagMiddleWare)
engine.Use(middleware.PurchaseStageTagMiddleWare)
engine.Use(middleware.URLRateLimitMiddleWare)
```

middleware 行为：

1. 用 `c.Request.Method` 和 `c.Request.URL.Path` 查 registry。
2. 命中时写入 gin context：`purchase_domain`、`purchase_stage`、`purchase_capability`。
3. 命中时在日志里输出统一字段。
4. 未命中时不报错、不阻断请求、不改变响应。

建议 context key 固定为：

```go
const (
    ContextPurchaseDomain     = "purchase_domain"
    ContextPurchaseStage      = "purchase_stage"
    ContextPurchaseCapability = "purchase_capability"
)
```

验收标准：

- 命中 `/digital-product/api/item/movie/seats` 时，context 中 stage 为 `booking`。
- 未命中普通健康检查时，不写阶段标签。
- 中间件不改变 HTTP status 和 response body。

#### Booking adapter 第一批落地

第一批 adapter 不直接替换旧逻辑，先做 shadow conversion：旧 handler 仍然按原方式调用 RPC，成功后把旧响应转换成统一 `BookingSession`，只打日志和指标，不改变返回。

统一模型：

```go
type BookingStatus string

const (
    BookingStatusCreated BookingStatus = "created"
    BookingStatusHeld    BookingStatus = "held"
    BookingStatusExpired BookingStatus = "expired"
    BookingStatusReleased BookingStatus = "released"
    BookingStatusFailed  BookingStatus = "failed"
)

type BookingSession struct {
    Domain         stage.PurchaseDomain
    BookingID      string
    OrderPlanID    string
    Status         BookingStatus
    ExpireTime     int64
    PaymentDeadline int64
    IdempotencyKey string
    ResourceSnapshot map[string]interface{}
    ProviderStatus string
}
```

电影票 shadow 接入点：

- 文件：`shopee-api-server/handler/movie/seat.go`
- 函数：`SelectSeat`
- 旧返回：`booking_id`、`total_amount`、`expired_time`
- 转换后：`BookingSession{Domain: movie_ticket, BookingID: resp.GetBookingId(), Status: held, ExpireTime: resp.GetExpiredTime()}`

轮渡 shadow 接入点：

- 文件：`shopee-api-server/handler/ferry/booking.go`
- 函数：`CreateBooking`
- 旧返回：`order_plan_id`、`booking_status`、`payment_deadline`
- 转换后：`BookingSession{Domain: ota_ferry, OrderPlanID: order_plan_id, ProviderStatus: booking_status, PaymentDeadline: payment_deadline}`

充值/账单 shadow 接入点：

- 文件：`shopee-api-server/handler/bill/query_bill.go` 或现有账单查询 handler 文件
- 函数：`QueryBill` / `FlowQueryBill`
- 转换后：不生成真实 booking，输出 `BookingSession{Domain: bill_topup, Status: created}` 或跳过 booking 指标，只输出 Detail snapshot。

切换到真实 adapter 调用时再加灰度开关：

```go
if apolloConfig.EnablePurchaseBookingAdapter("movie_ticket") {
    session, err := movieBookingAdapter.Create(ctx, input)
    // 转回旧 response，保持 API 兼容
} else {
    // 原逻辑
}
```

#### Booking Response 契约

前端入口不建议第一阶段强行统一成一个 `/booking/create`。电影票和轮渡的入参差异很大：

- 电影票需要 `cinema_code`、`session_code`、`seats`、`tickets`、`carrier_id`。
- 轮渡需要 `quotation_id`、`service_id`、`ship_class_id`、`selected_offers`、`passengers`、`vehicles`、`contact`。

因此第一阶段保持品类入口不变：

```text
/digital-product/api/item/movie/seats   -> MovieBookingAdapter -> BookingSession
/digital-product/api/item/ferry/booking -> FerryBookingAdapter -> BookingSession
```

但 Booking 成功后的响应语义要逐步统一成“统一外壳 + 品类 detail”。统一外壳给后续 Order/Checkout/Payment 使用，品类 detail 保留前端当前页面需要展示的差异字段。

推荐响应模型：

```go
type BookingResponse struct {
    Category     string          `json:"category"`
    BookingToken string          `json:"booking_token"`
    Status       string          `json:"status"`
    Amount       int64           `json:"amount,omitempty"`
    ExpireAt     int64           `json:"expire_at"`
    NextAction   string          `json:"next_action"`
    Detail       json.RawMessage `json:"detail"`
}
```

`detail` 是必填字段，但它应该是 JSON object，不应该是 JSON string。

正确返回：

```json
{
  "category": "ferry",
  "booking_token": "987654321012345678",
  "status": "held",
  "expire_at": 1715600000,
  "next_action": "create_order",
  "detail": {
    "order_plan_id": "987654321012345678",
    "quotation_id": "123456789012345678",
    "service_id": 1,
    "ship_class_id": 2,
    "passenger_count": 2
  }
}
```

不推荐返回：

```json
{
  "category": "ferry",
  "booking_token": "987654321012345678",
  "detail": "{\"order_plan_id\":\"987654321012345678\",\"quotation_id\":\"123456789012345678\"}"
}
```

区别是：JSON object 是结构化字段，前端可以直接按对象访问；JSON string 是被转义过的文本，前端还要二次 `JSON.parse`，接口文档、字段校验、日志检索和灰度兼容都会变弱。

后端实现时可以用 `json.RawMessage` 承载品类 detail。`json.RawMessage` 不会把 detail 再转义成字符串，而是把已经 marshal 好的 JSON bytes 原样嵌入响应：

```go
type FerryBookingDetail struct {
    OrderPlanID    string `json:"order_plan_id"`
    QuotationID    string `json:"quotation_id"`
    ServiceID      int64  `json:"service_id"`
    ShipClassID    int64  `json:"ship_class_id"`
    PassengerCount int    `json:"passenger_count"`
}

detail, err := json.Marshal(FerryBookingDetail{
    OrderPlanID:    orderPlanID,
    QuotationID:    req.QuotationID,
    ServiceID:      req.ServiceID,
    ShipClassID:    req.ShipClassID,
    PassengerCount: len(req.Passengers),
})
if err != nil {
    return BookingResponse{}, err
}

return BookingResponse{
    Category:     "ferry",
    BookingToken: orderPlanID,
    Status:       "held",
    ExpireAt:     booking.GetPaymentDeadline(),
    NextAction:   "create_order",
    Detail:       detail,
}, nil
```

如果 `Detail` 定义成 `string`，则会变成 JSON string：

```go
type BadBookingResponse struct {
    Detail string `json:"detail"`
}

BadBookingResponse{
    Detail: string(detail),
}
```

输出会变成：

```json
"detail": "{\"order_plan_id\":\"987654321012345678\"}"
```

精度约束：

- `json.RawMessage` 本身不会制造数值精度问题，它只负责原样嵌入 JSON bytes。
- 真正的风险来自前端 JavaScript `number` 的安全整数上限是 `9007199254740991`，以及后端如果把 JSON 反序列化到 `map[string]interface{}`，数字默认会变成 `float64`。
- 因此 `booking_token`、`order_plan_id`、`quotation_id`、`product_id`、`offer_id` 这类大 ID 一律返回 string。
- 金额、数量、过期时间如果确认不会超过 JS 安全整数可以继续用 number；如果是毫秒级超大时间戳或外部供应商大整数，也应返回 string。

电影票 detail 示例：

```go
type MovieBookingDetail struct {
    BookingID   string   `json:"booking_id"`
    CinemaCode  string   `json:"cinema_code"`
    SessionCode string   `json:"session_code"`
    SeatCodes   []string `json:"seat_codes"`
}
```

轮渡 detail 示例：

```go
type FerryBookingDetail struct {
    OrderPlanID    string `json:"order_plan_id"`
    QuotationID    string `json:"quotation_id"`
    ServiceID      int64  `json:"service_id"`
    ShipClassID    int64  `json:"ship_class_id"`
    PassengerCount int    `json:"passenger_count"`
}
```

前端 TypeScript 可以按 `category` 做联合类型：

```ts
type BookingBase = {
  category: 'movie' | 'ferry'
  booking_token: string
  status: 'held' | 'failed' | 'expired' | 'released'
  amount?: number
  expire_at: number
  next_action: 'create_order'
}

type MovieBookingResponse = BookingBase & {
  category: 'movie'
  detail: {
    booking_id: string
    cinema_code: string
    session_code: string
    seat_codes: string[]
  }
}

type FerryBookingResponse = BookingBase & {
  category: 'ferry'
  detail: {
    order_plan_id: string
    quotation_id: string
    service_id: number
    ship_class_id: number
    passenger_count: number
  }
}

type BookingResponse = MovieBookingResponse | FerryBookingResponse
```

这样前端消费统一字段时只看 `booking_token/status/amount/expire_at/next_action`，需要品类展示时再根据 `category` 读取 `detail`。后端也避免用裸 `map[string]interface{}` 作为长期接口契约，`map` 只用于内部 shadow、日志或临时扩展。

#### OrderSnapshot 落地方式

OrderSnapshot 不建议第一批就改 proto。分两步：

第一步是 API 层构造 shadow snapshot，只用于日志、监控和一致性校验：

```go
type OrderSnapshot struct {
    OrderID            string
    Domain             stage.PurchaseDomain
    CategoryID         int64
    CarrierID          int64
    ItemSnapshot       map[string]interface{}
    PriceSnapshot      map[string]interface{}
    ResourceSnapshot   map[string]interface{}
    FulfillmentSnapshot map[string]interface{}
    RefundPolicySnapshot map[string]interface{}
}
```

shadow snapshot 的来源：

- `order.Create` 入参中的 item、total amount、category、carrier。
- Booking 阶段产出的 `booking_id`、`order_plan_id`、座位/航班/车次/轮渡 offer。
- Detail 阶段查到的退款政策、履约规则。

第二步再推动 Order server 和 common proto 增加结构化字段。字段进入订单存储后，Checkout、Payment、Fulfillment、Refund 才切换为优先读取 snapshot。

这样做的原因是：第一步不影响下游存储和 proto 发布，能先验证 snapshot 字段是否足够；第二步再做跨服务改造，风险更可控。

#### 状态机落地方式

状态机不要先做成一个大平台。第一阶段只做状态枚举、合法迁移表和幂等校验工具，先给 Booking/Payment/Fulfillment/Refund 用。

建议新增：

| 文件 | 责任 |
| --- | --- |
| `shopee-api-server/purchase/statemachine/status.go` | 定义统一购买状态 |
| `shopee-api-server/purchase/statemachine/transition.go` | 定义合法迁移 |
| `shopee-api-server/purchase/statemachine/transition_test.go` | 测试合法/非法迁移 |

第一批合法迁移：

| From | Event | To |
| --- | --- | --- |
| `detailed` | `booking_created` | `booked` |
| `booked` | `order_created` | `ordered` |
| `ordered` | `checkout_ready` | `checkout_ready` |
| `checkout_ready` | `payment_started` | `paying` |
| `paying` | `payment_success` | `paid` |
| `paying` | `payment_failed` | `payment_failed` |
| `payment_failed` | `booking_released` | `closed` |
| `paid` | `fulfillment_started` | `fulfilling` |
| `fulfilling` | `fulfillment_success` | `fulfilled` |
| `fulfilled` | `refund_requested` | `refund_requested` |
| `refund_requested` | `refund_accepted` | `refund_processing` |
| `refund_processing` | `refund_success` | `refunded` |

幂等键建议：

| 阶段 | 幂等键 |
| --- | --- |
| Booking create | `domain + user_id + client_request_token` |
| Booking query | `domain + booking_id/order_plan_id` |
| Order create | `user_id + client_order_token/order_id` |
| Payment callback | `payment_transaction_id + callback_status` |
| Fulfillment callback | `order_id + provider_fulfillment_id + provider_status` |
| Refund callback | `refund_id + provider_refund_id + provider_status` |

#### 灰度和开关

每一批都需要独立开关，避免一个统一链路开关影响所有品类。

建议开关维度：

| 开关 | 作用 |
| --- | --- |
| `enable_purchase_stage_tag` | 是否启用阶段标签 middleware |
| `enable_purchase_booking_shadow` | 是否记录 BookingSession shadow 输出 |
| `enable_purchase_booking_adapter_movie` | 电影票是否真实走 BookingAdapter |
| `enable_purchase_booking_adapter_ferry` | 轮渡是否真实走 BookingAdapter |
| `enable_purchase_order_snapshot_shadow` | 是否构造 OrderSnapshot shadow |
| `enable_purchase_state_machine_shadow` | 是否做状态机 shadow 校验 |

灰度顺序：

1. `stage_tag` 全量打开，因为不改变业务行为。
2. `booking_shadow` 对 movie/ferry 打开，只记录转换结果。
3. `state_machine_shadow` 对 movie/ferry 打开，只校验迁移，不阻断。
4. `booking_adapter_movie` 小流量打开，旧响应保持兼容。
5. `booking_adapter_ferry` 小流量打开。
6. 稳定后再接 bus/train/flight/hotel。

#### 测试计划

单元测试：

- `stage/registry_test.go`：覆盖核心路由映射。
- `middleware/purchase_stage_tag_test.go`：覆盖命中路由、未命中路由、context 写入。
- `purchase/booking/adapter_test.go`：覆盖 domain 到 adapter 的 factory。
- `purchase/booking/movie_adapter_test.go`：覆盖 movie booking response 到 `BookingSession` 的转换。
- `purchase/booking/ferry_adapter_test.go`：覆盖 ferry order plan 到 `BookingSession` 的转换。
- `purchase/statemachine/transition_test.go`：覆盖合法迁移和非法迁移。

回归测试命令：

```bash
go test ./shopee-api-server/purchase/...
go test ./shopee-api-server/middleware -run TestPurchaseStageTag
go test ./shopee-api-server/handler/movie -run Test
go test ./shopee-api-server/handler/ferry -run Test
go test ./shopee-api-server/handler/aggregation -run TestCheckoutAggregation
```

接口回归：

- 电影票占座：请求 `POST /digital-product/api/item/movie/seats`，确认响应仍包含旧字段 `booking_id`、`total_amount`、`expired_time`。
- 轮渡 booking：请求 `POST /digital-product/api/item/ferry/booking`，确认响应仍包含旧字段 `order_plan_id`、`booking_status`、`payment_deadline`。
- Checkout：请求 `POST /digital-product/api/aggregation/checkout`，确认支付渠道、voucher、coin 字段不变。
- Refund：请求 `POST /digital-product/api/refund/place`，确认错误码和 data 结构不变。

#### 验收指标

第一批阶段标签验收：

- 目标路由阶段标签命中率大于 99%。
- 阶段标签 middleware 不引入 5xx。
- 未命中 registry 的路由不受影响。

第二批 Booking shadow 验收：

- movie/ferry booking 成功请求中，shadow `BookingSession` 构造成功率大于 99%。
- shadow 状态和旧响应状态无明显不一致。
- 日志可以按 `purchase_domain=movie_ticket`、`purchase_stage=booking` 检索到完整链路。

第三批真实 adapter 灰度验收：

- 电影票/轮渡 booking 成功率不低于改造前基线。
- booking 平均耗时和 P99 不高于改造前基线的 5%。
- 支付失败后资源释放补偿有记录，重复释放不造成业务错误。
- 旧前端响应字段完全兼容。

#### 任务拆解

批次 1：阶段标签和观测

1. 新增 `purchase/stage` 类型和 registry。
2. 为 bill、movie、ferry、local-service voucher、aggregation、refund 核心路由配置 `StageMeta`。
3. 新增 `PurchaseStageTagMiddleWare`。
4. 在 `RouterInit` 中接入 middleware。
5. 补 registry 和 middleware 单元测试。
6. 灰度打开 `enable_purchase_stage_tag`。

批次 2：Booking shadow

1. 新增 `purchase/model/booking.go`。
2. 新增 `purchase/booking` adapter 接口。
3. 实现 movie response 到 `BookingSession` 的 converter。
4. 实现 ferry response 到 `BookingSession` 的 converter。
5. 在 movie/ferry handler 成功返回前增加 shadow conversion 和日志。
6. 补 movie/ferry converter 单元测试。
7. 灰度打开 `enable_purchase_booking_shadow`。

批次 3：Booking adapter 真实调用

1. 在 movie adapter 中封装现有 `MovieSelectSeat` / `MovieCancelSeat` 调用。
2. 在 ferry adapter 中封装现有 `FerryCreateBooking` / `FerryGetBookingDetail` 调用。
3. 在 handler 中通过品类开关选择 adapter 或旧逻辑。
4. 确保 adapter 输出再转换回旧 response。
5. 小流量打开 movie adapter。
6. 小流量打开 ferry adapter。

批次 4：OrderSnapshot shadow

1. 新增 `purchase/model/snapshot.go`。
2. 在 `order.Create` 成功路径构造 `OrderSnapshot` shadow。
3. 从 BookingSession 补充 resource snapshot。
4. 从 create order req 补充 item/price snapshot。
5. 记录 snapshot 字段完整率。
6. 根据完整率决定 common proto 和 Order server 改造字段。

批次 5：状态机 shadow 和补偿闭环

1. 新增 `purchase/statemachine`。
2. 定义核心状态和迁移表。
3. 在 booking/payment/fulfillment/refund 成功路径做 shadow transition 校验。
4. 对非法迁移只打日志，不阻断。
5. 收集非法迁移样本，反推缺失状态或重复回调。
6. 再把 booking 超时释放、支付失败释放、履约失败重试纳入补偿任务。

### 四层：状态机与补偿任务

统一状态机应覆盖跨品类共有状态，不替代品类内部供应商状态：

```text
Searched -> Detailed -> Carted -> Booked -> Ordered -> CheckoutReady -> Paying
Paying -> Paid -> Fulfilling -> Fulfilled
Paying -> PaymentFailed -> BookingReleased
Booked -> BookingExpired -> Released
Paid -> RefundRequested -> RefundProcessing -> Refunded
Any -> Compensating -> Closed
```

关键约束：

- Booking 必须带 `expire_time`，电影票/巴士/火车/轮渡/机票必须可查询或释放；酒店先以 `pre-check` 作为短期 Booking。
- Order 必须固化 `price_snapshot`、`resource_snapshot`、`fulfillment_snapshot`、`refund_policy_snapshot`，后续支付页、履约、退款不能回读易变详情重算关键规则。
- Payment 回调、供应商回调、履约重试必须通过幂等键推进状态机，不能直接假设当前状态。
- 补偿任务按阶段定义：booking 超时释放、支付失败释放资源、支付成功但履约失败重试、退款申请后回查供应商退款状态。

## 改造批次

### 批次 0：接口盘点与阶段基线

目标：只做文档、路由标签方案、指标命名，不改业务行为。

交付：

- 完成本文档中四类商品的阶段映射。
- 给每个路由定义 `purchase_domain`、`purchase_stage`、`capability` 映射表。
- 确认需要迁移的入口和不纳入范围的运营/用户资产接口。

验收：

- 新增路由时必须能归入一个标准阶段。
- 每个目标品类能画出 Search -> Detail -> Booking/Order -> Checkout -> Payment -> Fulfillment/Refund 的实际路径。

### 批次 1：统一观测与灰度开关

目标：在 `prehandler` 或通用 middleware 注入阶段标签，打通监控。

交付：

- 路由到阶段的 registry。
- 统一日志字段：`purchase_domain`、`purchase_stage`、`capability`、`order_id`、`booking_id`、`resource_id`、`idempotency_key`。
- 阶段级耗时、错误码、业务错误率、超时率。

优先品类：

1. 电影票 Booking：链路短，`booking_id` 与过期时间明确。
2. Ferry Booking：新代码里已经有 `client_request_token` 和 `order_plan_id`，适合验证幂等观测。
3. 账单查询：高频但无资源锁，可验证 Detail/Cart 标签。

### 批次 2：标准 Booking adapter

目标：先统一资源锁生命周期，而不是先重写所有 Search/Detail。

交付：

- `BookingSession` 标准输出。
- `Create/Query/Release` 三类能力。
- 超时释放补偿任务的统一接口。

迁移顺序：

1. Movie：`item/movie/seats`、`order/movie/seats`。
2. Ferry：`ferry/quote`、`ferry/booking`、`GET ferry/booking`。
3. Bus/Train：显式 create/query/cancel/select seat。
4. Flight：create/query booking，补齐释放语义。
5. Hotel：先接 `pre-check`，再补 Order server 内部锁房点。

### 批次 3：Order snapshot 标准化

目标：把 `order.Create` 中品类差异从散落参数迁到标准快照。

交付：

- 创单请求增加或内部构造 `resource_snapshot`、`refund_policy_snapshot`、`fulfillment_snapshot`。
- 订单详情统一透出标准快照，老字段继续兼容。
- Checkout 只依赖 OrderSnapshot，不再回查 Detail 阶段易变字段。

优先品类：

1. 电影票和 OTA，因为资源价格、座位/行程、退款规则最需要固化。
2. 券类商品，因为库存、核销规则和售后限制要固化。
3. 充值/账单，最后清理 `order.Create` 中大量 category/carrier 分支。

### 批次 4：Search/Detail Facade

目标：统一 C 端导购和 PDP 聚合语义，同时保留老接口。

交付：

- 新增内部 Facade：`SearchProduct`、`GetProductDetail`。
- 老 handler 调用 Facade 或 Facade 调用老 handler 背后的 service，按品类逐步迁移。
- 输出统一 `ProductCandidate` 和 `ProductDetailSnapshot`。

迁移顺序：

1. Evoucher/local-service voucher：接口多但资源锁弱，适合先统一 Search/Detail schema。
2. 充值/账单：统一 bill query、account validation、item list。
3. Movie：统一电影、影院、场次、座位图。
4. OTA：保留供应商专用搜索结构，只统一 wrapper metadata 和价格/库存/规则字段。

### 批次 5：Refund policy 与售后闭环

目标：统一退款规则快照、可退 item、退款原因和退款申请。

交付：

- `RefundContext` 标准结构。
- Flight/Bus/Hotel 的 return items/policy 归一。
- 券类、电影票、充值/账单按可退/不可退、自动退/人工退归入统一规则。
- 退款状态机接入补偿任务和幂等回调。

## 风险与约束

| 风险 | 说明 | 缓解 |
| --- | --- | --- |
| 老接口响应结构被前端依赖 | 现有 RN/H5 可能依赖字段细节 | 第一阶段只新增内部上下文和标签，不改响应 |
| `order.Create` 已承载大量品类差异 | 直接改会影响所有品类 | 先在外层构造 snapshot，逐品类灰度 |
| OTA 供应商状态不一致 | booking status、释放能力、超时语义不同 | Booking adapter 保留 provider 原始状态，同时映射标准状态 |
| Hotel booking 边界不清 | prehandler 只有 pre-check，没有显式 create booking | 批次 0 继续向 Order server 追踪锁房点，短期把 pre-check 作为 Booking |
| Checkout 已有复杂聚合 | 直接拆改风险高 | `aggregation/checkout` 先作为统一 Checkout 实现，补领域标签和输入快照 |
| 退款规则易变 | Detail 阶段实时查的规则可能和退款时不一致 | Order 阶段固化 refund policy snapshot |

## 推荐落地结论

1. 不先新增统一对外 API。先保留老路由，建立内部标准阶段和观测。
2. 先统一 Booking，因为多品类最大差异在资源锁定、超时释放、幂等查询。
3. 以 `aggregation/checkout` 作为 Checkout 标准阶段的现有基础，不重写。
4. 以 OrderSnapshot 作为阶段闭环核心，所有后续 Checkout、Payment、Fulfillment、Refund 都优先读订单快照。
5. 路由层只负责注册和打标签，业务差异下沉到品类 adapter，避免 `prehandler` 继续堆品类流程。

## 自检结果

- 本文档没有待填项或占位符。
- 阶段定义与接口映射一致：每个目标品类至少覆盖 Search、Detail、Order、Checkout、Payment、Refund；资源型品类覆盖 Booking。
- 改造批次遵循低风险原则：先观测和 adapter，再改快照和 facade。
- Hotel 的 booking 边界已显式标注为短期假设，需要后续在 Order server 继续追踪实际锁房点。
