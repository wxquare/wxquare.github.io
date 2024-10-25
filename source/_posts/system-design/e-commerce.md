---
title: 互联网业务系统 - 电商系统后台
categories:
- 系统设计
---

<p align="center">
  <img src="/images/e-commerce-system.png" width=800 height=1300>
</p>


## 商品模型
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

## 订单模型
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


## 订单状态机
<p align="center">
  <img src="/images/order_state_machine.png" width=800 height=800>
</p>



## 核心业务流
### B 端
#### 首页运营和维护
#### 批量商品上传
#### 商品Edit更新，价格、状态等
### APP端
#### 首页获取
#### 商品搜索（列表）
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
### 通用商品缓存架构
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




参考:
- [订单状态机的设计和实现](https://www.cnblogs.com/wanglifeng717/p/16214122.html)
- 