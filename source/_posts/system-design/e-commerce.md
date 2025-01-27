---
title: 互联网业务系统 - 电商系统设计
categories:
- 系统设计
---


https://axureboutique.com/blogs/product-design/understanding-the-structure-of-e-commerce-products


## 电商系统业务流程 (business process)
<p align="center">
  <img src="/images/E-commerce-whole-business-process.webp" width=1200 height=700>
  <br/>
  <strong><a href="https://axureboutique.com/blogs/product-design/understanding-the-structure-of-e-commerce-products">E-commerce process</a></strong>
</p>

## 电商系统系统(system process)
<p align="center">
  <img src="/images/E-commerce-whole-system-process.webp" width=1200 height=700>
  <br/>
  <strong><a href="https://axureboutique.com/blogs/product-design/understanding-the-structure-of-e-commerce-products">E-commerce whole process of system</a></strong>
</p>


## 电商系统整体产品架构 (Product Structure)
<p align="center">
  <img src="/images/E-commerce-product-structure.webp" width=1200 height=700>
  <br/>
  <strong><a href="https://axureboutique.com/blogs/product-design/understanding-the-structure-of-e-commerce-products">E-commerce product structure</a></strong>
</p>


<p align="center">
  <img src="/images/e-commerce-system.png" width=800 height=1300>
</p>



## 电商Product Center
### 类目属性管理：classification/category/brand/attribute Management
<p align="center">
  <img src="/images/E-commerce-category-brand-product.webp" width=1000 height=400>
  <br/>
  <strong><a href="https://axureboutique.com/blogs/product-design/build-an-e-commerce-product-center-from-scratch">E-commerce product center</a></strong>
</p>

### 商品管理：Product management
<p align="center">
  <img src="/images/E-commerce-product-management.webp" width=1000 height=400>
  <br/>
  <strong><a href="https://axureboutique.com/blogs/product-design/build-an-e-commerce-product-center-from-scratch">E-commerce product center</a></strong>
</p>

### 商品管理（SPU和SKU）

方案一：同时创建多个SKU，并同步生成关联的SPU。整体方案是直接创建SKU，并维护多个不同的属性；该方案适用于大多数C2C综合电商平台（例如，阿里巴巴就是采用这种方式创建商品）。

方案二：先创建SPU，再根据SPU创建SKU。整体方案是由平台的主数据团队负责维护SPU，商家（包括自营和POP）根据SPU维护SKU。在创建SKU时，首先选择SPU（SPU中的基本属性由数据团队维护），然后基于SPU维护销售属性和物流属性，最后生成SKU；该方案适用于高度专业化的垂直B2B行业，如汽车、医药等。

这两种方案的原因是：垂直B2B平台上的业务（传统行业、年长的商家）操作能力有限，维护产品属性的错误率远高于C2C平台，同时平台对产品结构控制的要求较高。为了避免同一产品被不同商家维护成多个不同的属性（例如，汽车轮胎的胎面宽度、尺寸等属性），平台通常选择专门的数据团队来维护产品的基本属性，即维护SPU。

此外，B2B垂直电商的品类较少，SKU数量相对较小，品类标准化程度高，平台统一维护的可行性较高。

对于拥有成千上万品类的综合电商平台，依靠平台数据团队的统一维护是不现实的，或者像服装这样非标准化的品类对商品结构化管理的要求较低。因此，综合平台（阿里巴巴和亚马逊）的设计方向与垂直平台有所不同。

实际上，即使对于综合平台，不同的品类也会有不同的设计方法。一些品类具有垂直深度，因此也采用平台维护SPU和商家创建SKU的方式


### 模型1:
<p align="center">
  <img src="/images/E-commerce-product-management-ER.jpg" width=1000 height=500>
</p>

### 模型2:
<p align="center">
  <img src="/images/item-sku-er.png" width=600 height=1000>
</p>

- 为什么需要fe_category 和 be_category，提供前台运营的灵活性
- sku_item,sku_model。一个item是一系列sku的集合

<p align="center">
  <img src="/images/item-example.png" width=8000 height=400>
  <br>
  <span style="color: blue; font-weight: bold;">item</span>
</p>

<p align="center">
  <img src="/images/item-sku-example.png" width=600 height=400>
  <br>
  <span style="color: blue; font-weight: bold;">item-SKU</span>
</p>



## 电商Order Transaction Center
<p align="center">
  <img src="/images/order_er.png" width=800 height=600>
</p>

### 支付订单表（pay_order_tab）：主要用于记录用户的支付信息。主键为 pay_order_id，标识唯一的支付订单。
  - user_id：用户ID，标识支付的用户。
  - payment_method：支付方式，如信用卡、支付宝等。
  - payment_status：支付状态，如已支付、未支付等。
  - pay_amount、cash_amount、coin_amount、voucher_amount：支付金额、现金支付金额、代币支付金额、优惠券使用金额。
  - 时间戳字段包括创建时间、初始化时间和更新时间

### 订单表（order_tab）：记录用户的购买订单信息。主键为 order_id。
  - pay_order_id：支付订单ID，作为外键关联支付订单。
  - user_id：用户ID，标识购买订单的用户。
  - total_amount：订单的总金额。
  - order_status：订单状态，如已完成、已取消等。
  - payment_status：支付状态，与支付订单相关。
  - fulfillment_status：履约状态，表示订单的配送或服务状态。
  - refund_status：退款状态，用于标识订单是否有退款

### 订单商品表（order_item_tab：记录订单中具体商品的信息。主键为 order_item_id。
  - order_id：订单ID，作为外键关联订单。
  - item_id：商品ID，表示订单中的商品。
  - item_snapshot_id：商品快照ID，记录当时购买时的商品信息快照。
  - item_status：商品状态，如已发货、退货等。
  - quantity：购买数量。
  - price：商品单价。
  - discount：商品折扣金额

#### 退款表（refund_tab）：记录订单或订单项的退款信息。主键为 refund_id。
  - order_id：订单ID，作为外键关联订单。
  - order_item_id：订单项ID，标识具体商品的退款。
  - refund_amount：退款金额。
  - reason：退款原因。
  - quantity：退款的商品数量。
  - refund_status：退款状态。
  - refund_time：退款操作时间。

### 实体间关系：
### 支付订单与订单：
- 一个支付订单可能关联多个购买订单，形成 一对多 关系。
例如，用户可以通过一次支付购买多个不同的订单。
### 订单与订单商品：
一个订单可以包含多个订单项，形成 一对多 关系。
订单项代表订单中所购买的每个商品的详细信息。
### 订单与退款：
  - 一个订单可能包含多个退款，形成 一对多 关系。
  - 退款可以是针对订单整体，也可以针对订单中的某个商品


### 订单状态机
<p align="center">
  <img src="/images/order_state_machine.png" width=800 height=800>
</p>


### 核心业务流
### B 端
#### 首页运营和维护
#### 批量商品上传、编辑商品信息、价格、库存、状态
- mass/single upload
- mass/single edit
- verify，upload
- item sync fetch pull
- openapi

### APP端
#### 首页获取
#### 商品搜索
#### 
要求：
- 海量的数据，亿级别的商品量；
- 高并发查询，日 PV 过亿；
- 请求需要快速响应
特点：
- 商品数据已经结构化，但散布在商品、库存、价格、促销、仓储等多个系统
- 召回率要求高，保证每一个正常的商品均能够被搜索到
- 为保证用户体验，商品信息变更（比如价格、库存的变化）实时性要求高，导致更新量大，每天的更新量为千万级别
- 较强的个性化需求，由于是一个相对垂直的搜索领域，需要满足用户的个性化搜索意图，比如用户搜索“小说”有的用户希望找言情小说有的人需要找武侠小说有的人希望找到励志小说

#### 商品（商品详情）

#### 创单核心逻辑
- 用户校验
- 商品信息获取和校验
- 价格校验
- 营销活动校验
- antifraud
- 库存校验
- 生成payorderid和orderid
- 库存扣减和返还
- 营销活动扣减和返还
- 构建订单信息，插入DB
- 不同类型的创单逻辑会不同，这里通过接口定义基础的创单逻辑，后续不同类型的定义机遇这个逻辑扩展
```Go
package orderserver

// OrderRequest 包含创建订单所需的参数
type OrderRequest struct {
	UserID    string
	ProductID string
	Quantity  int
}

// OrderResponse 表示创建订单的响应
type OrderResponse struct {
	OrderID string
	Message string // 返回的信息，例如错误信息
}

// OrderServer 接口定义了创建订单的功能
type OrderServer interface {
	// validate
	ValidateUser(userID string) errors.ErrorCode
	GetProductInfo(productID string) (ProductInfo, errors.ErrorCode)
	ValidateProduct(productID string) errors.ErrorCode
	ValidatePrice(productID string) errors.ErrorCode
	ValidateInventory(productID string, quantity int) errors.ErrorCode
	ValidatePromotionCode(promoCode string) errors.ErrorCode
	CheckFraud() errors.ErrorCode

	GeneratePayOrderID() (string, errors.ErrorCode)
	GenerateOrderID() (string, errors.ErrorCode)

	DeductInventory(productID string, quantity int) errors.ErrorCode
	ReturnInventory(productID string, quantity int) errors.ErrorCode

	DeductPromotion(promoCode string) errors.ErrorCode
	ReturnPromotion(promoCode string) errors.ErrorCode

  BuildDBModels() errors.ErrorCode
	InsertOrder(order OrderRequest) (OrderResponse, error)
	LogOperation(orderID string, userID string) error

  PushOrderCreateEvent() errors.ErrorCode
}

// BaseOrderService 实现 OrderServer 接口
type BaseOrderService struct {
	// 可以添加数据库连接或其他依赖项
	req  OrderRequest
	resp OrderResponse

	OrderModel      *order.Model
  PayOrderModel   *order.PayModel
	OrderItemModels []*item.OrderItemModel
}

func (bos *BaseOrderService) ValidateUser(userID string) ErrorCode {
	// 简单的用户验证逻辑（示例）
	return Success
}
....
```
#### 订单支付和支付结果回调
<p align="center">
  <img src="/images/order_pay.png" width=500 height=1000>
</p>

```Go
  type OrderPayRequest struct {
      UserID      string
      OrderID     string
      PaymentMethod string // 支付方式，例如信用卡、支付宝等
      EVoucherCode string // 可选的电子券代码
      Coins        int     // 可用积分
  }

  // OrderPayResponse 表示支付请求的响应
  type OrderPayResponse struct {
      Success     bool
      Message     string
      PaymentID   string // 支付订单ID
  }

  // OrderPayService 接口定义了支付相关的功能
  type OrderPayService interface {
      ValidateUser(userID string) ErrorCode
      ValidateOrderStatus(orderID string) ErrorCode
      ValidatePrice(orderID string) ErrorCode
      ValidatePromotionCode(promoCode string) ErrorCode
      ValidateEVoucher(evoucherCode string) ErrorCode
      ValidateCoins(coins int) ErrorCode
      ValidatePaymentMethod(paymentMethod string) ErrorCode
      ValidatePaymentFees(orderID string, paymentMethod string) ErrorCode

      RedeemEVoucher(evoucherCode string) ErrorCode
      DeductCoins(coins int) ErrorCode

      InitializePayment(orderID string, paymentMethod string) (OrderPayResponse, ErrorCode)
      ConstructPaymentRequest(orderID string) (OrderPayResponse, ErrorCode)
      UpdateOrderStatus(orderID string, status string) ErrorCode
      HandleError(orderID string, err error) ErrorCode
  }
```

#### 订单履约和履约结果回调
<p align="center">
  <img src="/images/order_fulfillment.png" width=500 height=1000>
</p>

```Go
package fulfillmentserver

// OrderFulfillmentRequest 包含履行请求所需的参数
type OrderFulfillmentRequest struct {
    OrderID  string
    UserID   string
    Quantity int
}

// OrderFulfillmentResponse 表示履行请求的响应
type OrderFulfillmentResponse struct {
    Success   bool
    Message   string
    TrackingID string // 物流跟踪ID
}

// FulfillmentService 接口定义了订单履行相关的功能
type FulfillmentService interface {
    ValidateStock(orderID string, quantity int) ErrorCode
    ProcessOrder(request OrderFulfillmentRequest) (OrderFulfillmentResponse, ErrorCode)
    UpdateOrderStatus(orderID string, status string) ErrorCode
    HandleDelivery(orderID string) ErrorCode
    HandleError(orderID string, err error) ErrorCode
}
```
#### return & refund
<p align="center">
  <img src="/images/return&refund.png" width=500 height=1000>
</p>

- UserRefundOrderService、AdminRefundOrderService、FailedFulfillmentRefundOrderService
- Return
- refund

##### RefundPlaceOrder
```Go
package refundservice

// RefundOrderRequest 包含退款请求所需的参数
type RefundOrderRequest struct {
    OrderID  string
    UserID   string
    Amount   float64
}

// RefundOrderResponse 表示退款请求的响应
type RefundOrderResponse struct {
    Success   bool
    Message   string
    RefundID  string // 退款ID
}

// RefundOrderService 接口定义了退款相关的功能
type RefundOrderService interface {
    ValidateRefund(request RefundOrderRequest) ErrorCode
    PlaceOrder(request RefundOrderRequest) (RefundOrderResponse, ErrorCode)
}

// BaseRefundOrderService 实现 RefundOrderService 接口
type BaseRefundOrderService struct{}

func (bros *BaseRefundOrderService) ValidateRefund(request RefundOrderRequest) ErrorCode {
    if request.Amount <= 0 {
        return ErrInvalidAmount
    }
    if !orderExists(request.OrderID) {
        return ErrOrderNotFound
    }
    return Success
}
// 假设的辅助函数
func orderExists(orderID string) bool {
    // 检查订单是否存在的逻辑
    return true // 示例返回
}

func initiateRefund(orderID string, amount float64) string {
    // 处理退款并返回退款ID的逻辑
    return "refund123" // 示例返回
}


// UserRefundOrderService 实现用户退款订单的逻辑
type UserRefundOrderService struct {
    baseService *BaseRefundOrderService
}

func (uros *UserRefundOrderService) ValidateRefund(request RefundOrderRequest) ErrorCode {
    return uros.baseService.ValidateRefund(request)
}

func (uros *UserRefundOrderService) PlaceOrder(request RefundOrderRequest) (RefundOrderResponse, ErrorCode) {
    if errCode := uros.ValidateRefund(request); errCode != Success {
        return RefundOrderResponse{}, errCode
    }

    // 处理用户创建退款订单的逻辑
    refundID := initiateRefund(request.OrderID, request.Amount)

    return RefundOrderResponse{
        Success:  true,
        Message:  "User refund order created successfully.",
        RefundID: refundID,
    }, Success
}

// AdminRefundOrderService 实现管理员退款订单的逻辑
type AdminRefundOrderService struct {
    baseService *BaseRefundOrderService
}

func (aros *AdminRefundOrderService) ValidateRefund(request RefundOrderRequest) ErrorCode {
    return aros.baseService.ValidateRefund(request)
}

func (aros *AdminRefundOrderService) PlaceOrder(request RefundOrderRequest) (RefundOrderResponse, ErrorCode) {
    if errCode := aros.ValidateRefund(request); errCode != Success {
        return RefundOrderResponse{}, errCode
    }

    // 处理管理员创建退款订单的逻辑
    refundID := initiateRefund(request.OrderID, request.Amount)

    return RefundOrderResponse{
        Success:  true,
        Message:  "Admin refund order created successfully.",
        RefundID: refundID,
    }, Success
}

// FailedDeliveryRefundOrderService 实现发货失败退款订单的逻辑
type FailedFulfillmentRefundOrderService struct {
    baseService *BaseRefundOrderService
}

func (fdros *FailedFulfillmentRefundOrderService) HandleFailedDelivery(orderID string) (RefundOrderResponse, ErrorCode) {
    // 假设处理发货失败的逻辑
    refundRequest := RefundOrderRequest{
        OrderID: orderID,
        UserID:  "system", // 系统自动处理
        Amount:  0.0,      // 假设金额为0，具体金额需要根据业务逻辑设置
    }

    // 处理退款
    refundID := initiateRefund(refundRequest.OrderID, refundRequest.Amount)

    return RefundOrderResponse{
        Success:  true,
        Message:  "Refund order created due to failed delivery.",
        RefundID: refundID,
    }, Success
}
```

##### RefundApproveService
```Go
// RefundApproveRequest 包含退款审批请求所需的参数
type RefundApproveRequest struct {
    RefundID string
    Approve  bool
}

// RefundApproveResponse 表示退款审批的响应
type RefundApproveResponse struct {
    Success bool
    Message string
}

// RefundApproveService 接口定义了退款审批相关的功能
type RefundApproveService interface {
    ApproveRefund(request RefundApproveRequest) (RefundApproveResponse, ErrorCode)
}
```

##### ReturnPurchaseService
```Go
// ReturnPurchaseRequest 包含退货请求所需的参数
type ReturnPurchaseRequest struct {
    OrderID  string
    UserID   string
    Reason   string
    Amount   float64
}

// ReturnPurchaseResponse 表示退货请求的响应
type ReturnPurchaseResponse struct {
    Success   bool
    Message   string
    ReturnID  string // 退货ID
}
// ReturnPurchaseService 接口定义了退货相关的功能
type ReturnPurchaseService interface {
    ValidateReturn(request ReturnPurchaseRequest) ErrorCode
    ProcessReturn(request ReturnPurchaseRequest) (ReturnPurchaseResponse, ErrorCode)
}
```

##### RefundService
```Go
package refundservice

// RefundRequest 包含退款请求所需的参数
type RefundRequest struct {
    OrderID string
    UserID  string
    Amount  float64
}

// RefundResponse 表示退款请求的响应
type RefundResponse struct {
    Success  bool
    Message  string
    RefundID string // 退款ID
}

// RefundService 接口定义了退款相关的功能
type RefundService interface {
    ValidateRefund(request RefundRequest) ErrorCode
    ProcessRefund(request RefundRequest) (RefundResponse, ErrorCode)
}
```
#### 订单详情查询
## 系统挑战
### 如何维护订单状态的最终一致性？
<p align="center">
  <img src="/images/order_final_consistency_activity.png" width=600 height=600>
</p>

- 状态机一定要设计好，只有特定的原始状态 + 特定的事件才可以推进到指定的状态。
- 并发更新数据库前，要用乐观锁或者悲观锁，先使用select for update进行锁行记录，同时在更新时判断版本号是否是之前取出来的版本号，更新成功就结束，更新失败就组成消息发到消息队列，后面再消费。
- 通过补偿机制兜底，比如查询补单。
- 通过上述三个步骤，正常情况下，最终的数据状态一定是正确的。除非是某个系统有异常，比如外部渠道开始返回支付成功，然后又返回支付失败，说明依赖的外部系统已经异常，这样只能进人工差错处理流程。


### 商品信息缓存和数据一致性
<p align="center">
  <img src="/images/item-info-cache.png" width=600 height=500>
</p>

### 主从架构中如何获取最新的数据，避免因为主从延时导致获得脏数据
<p align="center">
  <img src="/images/master-slave-get-latest-data.png" width=500 height=400>
</p>

| **策略**     | **优点**                                                                         | **缺点**                                                                        |
|-------------|----------------------------------------------------------------------------------|--------------------------------------------------------------------------------|
| **1. 直接读取主库** | - **一致性:** 始终获取最新的数据。                | - **性能:** 增加主库的负载，可能导致性能瓶颈。                                                                     |
|                   | - **简单性:** 实现简单直接，因为它直接查询可信的源。 | - **可扩展性:** 主库可能成为瓶颈，限制系统在高读流量下有效扩展的能力。                                                 |
|                   |                                               |                                                                                                             |
| **2. 使用VersionCache与从库** | - **性能:** 分散读取负载到从库，减少主库的压力。  | - **复杂性:** 实现更加复杂，需要进行缓存管理并处理潜在的不一致性问题。                                   |
|                             | - **可扩展性:** 通过将大部分读取操作卸载到从库，实现更好的扩展性。      | - **缓存管理:** 需要进行适当的缓存失效处理和同步，以确保数据的一致性。                 |
|                             | - **一致性:** 通过比较版本并在必要时回退到主库，提供确保最新数据的机制。| - **潜在延迟:** 从库的数据可能仍然存在不同步的可能性，导致数据更新前有轻微延迟。         |




## 参考:

- [订单状态机的设计和实现](https://www.cnblogs.com/wanglifeng717/p/16214122.html)
- 