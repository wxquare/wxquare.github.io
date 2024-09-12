---
title: 互联网业务系统 - 电商系统后台
categories:
- 系统设计
---

<p align="center">
  <img src="/images/e-commerce-system.png" width=800 height=1300>
</p>


## 通用商品模型
<p align="center">
  <img src="/images/item-sku.png" width=600 height=1000>
</p>


## 通用商品缓存架构
<p align="center">
  <img src="/images/item-info-cache.png" width=600 height=500>
</p>

## 通用订单模型
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

### 订单项表（order_item_tab：记录订单中具体商品的信息。主键为 order_item_id。
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


## 通用订单状态机
<p align="center">
  <img src="/images/order_state_machine.png" width=800 height=800>
</p>


## 主从架构中如何获取最新的数据，避免因为主从延时导致获得脏数据
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


## 订单支付核心逻辑
<p align="center">
  <img src="/images/order_pay.png" width=500 height=1000>
</p>


## 订单履约核心逻辑
<p align="center">
  <img src="/images/order_fulfillment.png" width=500 height=1000>
</p>
