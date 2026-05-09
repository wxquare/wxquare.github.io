# 35.3.3 订单系统题库

## 35.3.3 订单系统（15题）

#### 📊 题目1：订单状态机的设计

**问题描述**：
订单从创建到完成，经历多个状态（待支付、待发货、待收货、已完成）。如何设计订单状态机，保证状态流转的正确性？

**答案**：

**问题分析**：
订单状态流转的核心要素：
1. 状态定义清晰
2. 流转规则明确
3. 防止非法跳转
4. 支持异常流程（取消、退款）

**状态定义**：
```text
正向流程：
PENDING_PAYMENT（待支付）
→ PAID（已支付/待发货）
→ SHIPPED（已发货/待收货）
→ RECEIVED（已收货/待评价）
→ COMPLETED（已完成）

逆向流程：
CANCELLED（已取消）
REFUNDING（退款中）
REFUNDED（已退款）

特殊状态：
TIMEOUT（超时关闭）
```

**状态机实现**：

方案一：If-Else判断
```java
public void updateOrderStatus(Order order, OrderStatus newStatus) {
  OrderStatus currentStatus = order.getStatus();
  
  if (currentStatus == PENDING_PAYMENT) {
    if (newStatus == PAID || newStatus == CANCELLED || newStatus == TIMEOUT) {
      order.setStatus(newStatus);
    } else {
      throw new IllegalStateException("非法状态转换");
    }
  } else if (currentStatus == PAID) {
    if (newStatus == SHIPPED || newStatus == REFUNDING) {
      order.setStatus(newStatus);
    } else {
      throw new IllegalStateException("非法状态转换");
    }
  }
  // ... 更多判断
}
```

缺点：
- 代码冗长
- 难以维护
- 状态多时复杂度爆炸

方案二：状态转换表（推荐）
```java
// 定义状态转换规则
private static final Map<OrderStatus, Set<OrderStatus>> TRANSITIONS = Map.of(
  PENDING_PAYMENT, Set.of(PAID, CANCELLED, TIMEOUT),
  PAID, Set.of(SHIPPED, REFUNDING),
  SHIPPED, Set.of(RECEIVED, REFUNDING),
  RECEIVED, Set.of(COMPLETED, REFUNDING),
  REFUNDING, Set.of(REFUNDED)
);

public void updateOrderStatus(Order order, OrderStatus newStatus) {
  OrderStatus currentStatus = order.getStatus();
  
  Set<OrderStatus> allowedTransitions = TRANSITIONS.get(currentStatus);
  if (allowedTransitions == null || !allowedTransitions.contains(newStatus)) {
    throw new IllegalStateException(
      String.format("不允许从%s转换到%s", currentStatus, newStatus)
    );
  }
  
  // 记录状态变更历史
  OrderStatusHistory history = new OrderStatusHistory();
  history.setOrderId(order.getId());
  history.setFromStatus(currentStatus);
  history.setToStatus(newStatus);
  history.setOperator(getCurrentUser());
  history.setReason(reason);
  historyRepository.save(history);
  
  // 更新订单状态
  order.setStatus(newStatus);
  orderRepository.save(order);
  
  // 发布状态变更事件
  eventPublisher.publish(new OrderStatusChangedEvent(order, currentStatus, newStatus));
}
```

优点：
- 规则清晰
- 易于维护
- 可扩展

**状态流转图**：
```text
                    ┌─> CANCELLED
                    │
PENDING_PAYMENT ──┬─┴─> PAID ───> SHIPPED ───> RECEIVED ───> COMPLETED
                  │                  │            │
                  └─> TIMEOUT        │            │
                                     │            │
                                     └─> REFUNDING <─┘
                                            │
                                            └─> REFUNDED
```

**延伸思考**：
1. 如何设计订单的子状态（如待发货细分为待拣货、待打包、待出库）？
2. 订单状态变更如何触发后续操作（如发货后通知物流）？
3. 如何处理状态流转的并发冲突？

---

#### 🔧 题目2：订单号生成规则

**问题描述**：
订单号需要唯一、有序、不易被猜测。如何设计订单号生成规则？

**答案**：

**订单号设计要求**：
1. 全局唯一
2. 趋势递增（便于分库分表）
3. 信息可读（包含时间、业务类型）
4. 安全性（不易被遍历）
5. 长度适中（15-20位）

**方案一：数据库自增ID**

优点：
- 简单
- 唯一

缺点：
- 连续，易被猜测
- 分布式环境难实现
- 信息量少

**方案二：UUID**

优点：
- 全局唯一
- 无需中心化

缺点：
- 无序（影响索引性能）
- 长度太长（36位）
- 无业务含义

**方案三：Snowflake算法（推荐）**

结构：
```text
64位Long型：
1位符号位 + 41位时间戳 + 10位机器ID + 12位序列号

示例：
0 - 00000000000000000000000000000000000000000 - 0000000000 - 000000000000
│   └─────────────41位时间戳─────────────────┘   └10位机器┘   └12位序列┘
符号位

生成的订单号：1234567890123456789（19位）
```

实现：
```java
public class SnowflakeIdGenerator {
  // 起始时间戳（2020-01-01）
  private final long epoch = 1577836800000L;
  
  // 机器ID（数据中心ID + 机器ID）
  private final long workerId;
  
  // 序列号
  private long sequence = 0L;
  
  // 上次生成ID的时间戳
  private long lastTimestamp = -1L;
  
  public synchronized long nextId() {
    long timestamp = System.currentTimeMillis();
    
    // 时钟回拨检测
    if (timestamp < lastTimestamp) {
      throw new RuntimeException("时钟回拨");
    }
    
    // 同一毫秒内
    if (timestamp == lastTimestamp) {
      sequence = (sequence + 1) & 4095; // 4095=2^12-1
      if (sequence == 0) {
        // 序列号用完，等待下一毫秒
        timestamp = waitNextMillis(lastTimestamp);
      }
    } else {
      sequence = 0;
    }
    
    lastTimestamp = timestamp;
    
    // 组装ID
    return ((timestamp - epoch) << 22) 
         | (workerId << 12) 
         | sequence;
  }
}
```

优点：
- 趋势递增
- 高性能
- 分布式友好

缺点：
- 依赖机器时钟
- 机器ID需要管理

**方案四：业务规则拼接**

结构：
```text
订单号格式：业务前缀 + 日期 + 随机数

示例：
OR20260418123456789
│  └────┘└───────┘
│   日期   随机数
业务前缀（OR=Order）

生成：
String orderId = "OR" 
               + LocalDate.now().format(DateTimeFormatter.BASIC_ISO_DATE)
               + RandomStringUtils.randomNumeric(9);
```

优点：
- 可读性强
- 包含业务信息
- 可自定义

缺点：
- 需要保证随机数不重复
- 长度较长

**推荐方案**：
使用**Snowflake算法**生成基础ID，再转为业务订单号。

实现：
```java
public String generateOrderNo() {
  long snowflakeId = idGenerator.nextId();
  
  // 转为订单号（添加业务前缀）
  return "OR" + snowflakeId;
}
```

**延伸思考**：
1. 如何设计订单号的校验规则（防止伪造）？
2. 订单号如何支持多业务类型（普通订单、预售订单、拼团订单）？
3. 分库分表场景下订单号如何设计路由键？

---

#### 💡 题目3：订单超时自动取消

**问题描述**：
用户下单30分钟未支付，订单自动关闭并释放库存。如何实现订单超时自动取消？

**答案**：

**方案一：定时任务扫描**

核心思想：
定时任务定期扫描超时订单。

实现：
```java
@Scheduled(fixedDelay = 60000) // 每分钟执行
public void cancelTimeoutOrders() {
  // 查询超时未支付订单
  List<Order> timeoutOrders = orderRepository.findByStatusAndCreateTimeBefore(
    OrderStatus.PENDING_PAYMENT,
    LocalDateTime.now().minus(30, ChronoUnit.MINUTES)
  );
  
  for (Order order : timeoutOrders) {
    try {
      // 取消订单
      orderService.cancel(order.getId(), "超时未支付自动取消");
      
      // 释放库存
      inventoryService.release(order.getItems());
      
      // 通知用户
      notificationService.send(order.getUserId(), "订单已超时关闭");
    } catch (Exception e) {
      log.error("取消订单失败", e);
    }
  }
}
```

优点：
- 实现简单
- 可靠性高

缺点：
- 实时性差（最长延迟1分钟）
- 数据库扫描压力大
- 定时任务单点故障

**方案二：延迟队列（推荐）**

核心思想：
订单创建时发送延迟消息，30分钟后消费取消订单。

使用RabbitMQ延迟队列：
```java
// 创建订单时
public void createOrder(Order order) {
  // 1. 保存订单
  orderRepository.save(order);
  
  // 2. 发送延迟消息（30分钟后）
  rabbitTemplate.convertAndSend(
    "order.cancel.exchange",
    "order.cancel.routing.key",
    order.getId(),
    message -> {
      message.getMessageProperties().setDelay(30 * 60 * 1000); // 30分钟
      return message;
    }
  );
}

// 消费延迟消息
@RabbitListener(queues = "order.cancel.queue")
public void handleOrderCancel(Long orderId) {
  Order order = orderRepository.findById(orderId);
  
  // 检查订单状态
  if (order.getStatus() == OrderStatus.PENDING_PAYMENT) {
    // 仍未支付，取消订单
    orderService.cancel(orderId, "超时未支付自动取消");
    inventoryService.release(order.getItems());
  }
  // 如果已支付，忽略
}
```

使用Redis实现延迟队列：
```java
// 创建订单时
public void createOrder(Order order) {
  orderRepository.save(order);
  
  // 添加到Redis有序集合（Sorted Set）
  long expireTime = System.currentTimeMillis() + 30 * 60 * 1000;
  redis.zadd("order:timeout", expireTime, order.getId());
}

// 定时消费
@Scheduled(fixedDelay = 1000) // 每秒执行
public void processTimeoutOrders() {
  long now = System.currentTimeMillis();
  
  // 获取已到期的订单ID
  Set<String> orderIds = redis.zrangeByScore("order:timeout", 0, now);
  
  for (String orderId : orderIds) {
    try {
      // 处理超时订单
      processTimeoutOrder(Long.parseLong(orderId));
      
      // 从集合中移除
      redis.zrem("order:timeout", orderId);
    } catch (Exception e) {
      log.error("处理超时订单失败", e);
    }
  }
}
```

优点：
- 准确到秒
- 分布式友好
- 性能好

缺点：
- 依赖消息队列
- 需要处理消息丢失

**方案三：时间轮算法**

核心思想：
使用时间轮数据结构管理超时任务。

实现（Netty HashedWheelTimer）：
```java
private final HashedWheelTimer timer = new HashedWheelTimer(
  1, TimeUnit.SECONDS,  // 每秒tick一次
  60                     // 60个槽位
);

public void createOrder(Order order) {
  orderRepository.save(order);
  
  // 添加超时任务
  timer.newTimeout(timeout -> {
    Order latestOrder = orderRepository.findById(order.getId());
    if (latestOrder.getStatus() == OrderStatus.PENDING_PAYMENT) {
      orderService.cancel(order.getId(), "超时未支付自动取消");
    }
  }, 30, TimeUnit.MINUTES);
}
```

优点：
- 高性能
- 精确度高

缺点：
- 内存占用（任务在内存）
- 单机方案（不支持分布式）
- 服务重启任务丢失

**方案对比**：

| 方案 | 实时性 | 可靠性 | 分布式 | 实施难度 |
|------|--------|--------|--------|---------|
| 定时扫描 | ★★☆☆☆ | ★★★★★ | ★★★★☆ | ★★★★★ |
| 延迟队列 | ★★★★★ | ★★★★☆ | ★★★★★ | ★★★☆☆ |
| 时间轮 | ★★★★★ | ★★★☆☆ | ★★☆☆☆ | ★★★☆☆ |

**推荐方案**：
采用**延迟队列（RabbitMQ或Redis）**。

实施要点：

1. **幂等性保证**：
   ```java
   @Transactional
   public void cancel(Long orderId, String reason) {
     Order order = orderRepository.findById(orderId);
     
     // 检查当前状态
     if (order.getStatus() != OrderStatus.PENDING_PAYMENT) {
       log.warn("订单{}状态不是待支付，跳过取消", orderId);
       return; // 已被其他线程处理
     }
     
     // CAS更新状态
     int updated = orderRepository.updateStatus(
       orderId, 
       OrderStatus.CANCELLED,
       OrderStatus.PENDING_PAYMENT // 期望的旧状态
     );
     
     if (updated == 0) {
       log.warn("订单{}取消失败，可能已被处理", orderId);
       return;
     }
     
     // 释放库存
     inventoryService.release(order.getItems());
   }
   ```

2. **异常重试**：
   ```
   取消失败的处理：
   - 消息重新入队，稍后重试
   - 最多重试3次
   - 仍失败则记录告警，人工处理
   ```

3. **监控告警**：
   ```
   监控指标：
   - 超时订单数量
   - 取消成功率
   - 延迟队列堆积量
   
   告警：
   - 取消失败率 > 1%
   - 延迟队列堆积 > 10000
   ```

**延伸思考**：
1. 如何设计不同订单类型的不同超时时间（普通30分钟，秒杀10分钟）？
2. 订单超时取消如何通知用户？
3. 大促期间超时订单激增如何处理？

---

#### 📊 题目4：订单拆单与合单策略

**问题描述**：
用户购买多个商品，可能来自不同仓库或不同商家。如何设计订单拆单与合单策略？

**答案**：

**问题分析**：
拆单场景：
1. 多仓库发货（就近发货）
2. 多商家发货（平台+第三方卖家）
3. 预售+现货（发货时间不同）
4. 自营+跨境（清关时间不同）

合单场景：
1. 同一地址多笔订单（节省运费）
2. 同一商家商品（方便发货）

**方案一：用户下单时拆单**

核心思想：
用户提交订单时，系统自动拆分为多个子订单。

流程：
```text
用户购物车：
- 商品A（北京仓）
- 商品B（上海仓）
- 商品C（北京仓）

拆单规则：
按仓库拆分：
→ 子订单1：商品A + C（北京仓）
→ 子订单2：商品B（上海仓）

数据结构：
parent_order（父订单）
├── parent_order_id
├── user_id
├── total_amount
└── status

sub_order（子订单）
├── sub_order_id
├── parent_order_id
├── warehouse_id
├── items
└── status
```

用户支付：
```text
用户支付父订单 → 分配金额到各子订单
子订单独立发货、收货
```

优点：
- 逻辑清晰
- 用户感知明确

缺点：
- 用户体验复杂（多个运单号）
- 退款复杂（部分退款）

**方案二：后台自动拆单（推荐）**

核心思想：
用户下单时是一个订单，后台根据规则自动拆分为多个发货单。

流程：
```text
用户下单：创建订单（单个）
↓
订单支付成功
↓
订单中心分析：需要拆单
↓
创建多个发货单（shipment）
- 发货单1：商品A+C → 北京仓
- 发货单2：商品B → 上海仓
↓
各仓库独立发货
```

数据结构：
```sql
order（订单）
├── order_id
├── user_id
├── total_amount
└── status

shipment（发货单）
├── shipment_id
├── order_id
├── warehouse_id
├── items（发货商品）
├── tracking_number（运单号）
└── status
```

优点：
- 用户无感知（看到的是一个订单）
- 退款简单（按订单退）
- 灵活（可随时调整拆单规则）

缺点：
- 实现复杂
- 需要维护订单和发货单的关系

**拆单规则**：

1. **按仓库拆分**：
   ```java
   public List<Shipment> splitByWarehouse(Order order) {
     // 1. 为每个商品选择最优仓库
     Map<String, Warehouse> itemWarehouse = new HashMap<>();
     for (OrderItem item : order.getItems()) {
       Warehouse warehouse = selectWarehouse(item.getSkuId(), order.getAddress());
       itemWarehouse.put(item.getSkuId(), warehouse);
     }
     
     // 2. 按仓库分组
     Map<Warehouse, List<OrderItem>> grouped = order.getItems().stream()
       .collect(Collectors.groupBy(item -> itemWarehouse.get(item.getSkuId())));
     
     // 3. 生成发货单
     List<Shipment> shipments = new ArrayList<>();
     for (Map.Entry<Warehouse, List<OrderItem>> entry : grouped.entrySet()) {
       Shipment shipment = new Shipment();
       shipment.setOrderId(order.getId());
       shipment.setWarehouseId(entry.getKey().getId());
       shipment.setItems(entry.getValue());
       shipments.add(shipment);
     }
     
     return shipments;
   }
   ```

2. **按商家拆分**：
   ```
   平台订单包含：
   - 自营商品（平台发货）
   - 第三方商品（商家发货）
   
   拆分：
   - 子订单1：自营商品
   - 子订单2：商家A的商品
   - 子订单3：商家B的商品
   ```

3. **按发货时间拆分**：
   ```
   订单包含：
   - 现货商品（立即发货）
   - 预售商品（15天后发货）
   
   拆分：
   - 发货单1：现货（立即发）
   - 发货单2：预售（延迟发）
   ```

**合单策略**：

1. **同地址合并**：
   ```
   用户A在1小时内下了3笔订单：
   - 订单1：商品A（北京仓）
   - 订单2：商品B（北京仓）
   - 订单3：商品C（上海仓）
   
   合单：
   - 发货单1：订单1+订单2的商品（北京仓合并发货）
   - 发货单2：订单3的商品（上海仓单独发货）
   
   好处：
   - 节省运费
   - 减少包裹数量
   ```

2. **运费优化**：
   ```
   规则：
   - 同一仓库、同一地址、24小时内的订单
   - 自动合并发货
   - 运费退还到用户余额
   ```

**推荐方案**：
采用**后台自动拆单**。

实施要点：

1. **拆单时机**：
   ```
   时机选择：
   - 订单支付后立即拆单（推荐）
   - 发货前拆单（更灵活）
   ```

2. **用户展示**：
   ```
   订单详情页：
   订单号：OR123456
   总金额：¥1000
   
   发货信息：
   - 包裹1：商品A+B（运单号：SF123）
     状态：已发货
   - 包裹2：商品C（运单号：SF456）
     状态：待发货
   ```

3. **退款处理**：
   ```
   部分商品退款：
   - 用户申请退商品A
   - 计算退款金额（商品价 + 分摊运费）
   - 只退部分金额
   - 其他商品正常履约
   ```

**延伸思考**：
1. 如何设计拆单的运费分摊规则？
2. 拆单后如何保证库存一致性？
3. 跨境订单的拆单有何特殊性？

---

#### 🔧 题目5：订单的并发创建与幂等性

**问题描述**：
用户可能重复点击"提交订单"按钮，导致创建多个订单。如何保证订单创建的幂等性？

**答案**：

**问题分析**：
重复下单的原因：
1. 用户重复点击
2. 网络超时重试
3. 前端未防抖
4. 恶意刷单

**方案一：前端防抖**

核心思想：
前端限制用户短时间内多次点击。

实现：
```javascript
let submitting = false;

function submitOrder() {
  if (submitting) {
    return; // 正在提交中，忽略
  }
  
  submitting = true;
  
  fetch('/api/order/create', {
    method: 'POST',
    body: JSON.stringify(orderData)
  })
  .then(res => {
    // 处理结果
  })
  .finally(() => {
    submitting = false; // 完成后恢复
  });
}
```

优点：
- 简单有效

缺点：
- 仅防止前端重复
- 无法防止恶意绕过前端

**方案二：唯一索引（推荐）**

核心思想：
数据库层面保证唯一性。

实现：
```sql
CREATE TABLE orders (
  order_id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  idempotent_key VARCHAR(64) UNIQUE, -- 幂等键
  ...
);

CREATE UNIQUE INDEX uk_user_idempotent ON orders(user_id, idempotent_key);
```

创建订单：
```java
@Transactional
public Order createOrder(OrderRequest request, String idempotentKey) {
  try {
    // 1. 构建订单
    Order order = new Order();
    order.setUserId(request.getUserId());
    order.setIdempotentKey(idempotentKey);
    order.setItems(request.getItems());
    // ...
    
    // 2. 保存订单（唯一索引保证幂等）
    orderRepository.save(order);
    
    // 3. 扣减库存
    inventoryService.deduct(order.getItems());
    
    return order;
  } catch (DuplicateKeyException e) {
    // 幂等键重复，说明订单已创建
    return orderRepository.findByIdempotentKey(idempotentKey);
  }
}
```

幂等键生成：
```java
// 方案1：前端生成UUID
String idempotentKey = UUID.randomUUID().toString();

// 方案2：后端生成（基于购物车内容）
String idempotentKey = DigestUtils.md5Hex(
  userId + ":" + cartItems.toString() + ":" + timestamp
);
```

优点：
- 数据库层面保证
- 可靠性高

缺点：
- 依赖唯一索引
- 需要生成幂等键

**方案三：分布式锁**

核心思想：
使用Redis分布式锁，同一用户同时只能创建一个订单。

实现：
```java
public Order createOrder(OrderRequest request) {
  String lockKey = "order:create:" + request.getUserId();
  
  // 尝试获取锁
  boolean locked = redisLock.tryLock(lockKey, 10, TimeUnit.SECONDS);
  if (!locked) {
    throw new BizException("正在创建订单，请勿重复提交");
  }
  
  try {
    // 创建订单
    Order order = doCreateOrder(request);
    return order;
  } finally {
    // 释放锁
    redisLock.unlock(lockKey);
  }
}
```

优点：
- 防止并发创建
- 灵活控制

缺点：
- 依赖Redis
- 锁超时需要处理

**方案四：Token机制**

核心思想：
用户进入结算页时，服务端生成唯一Token，提交订单时校验Token。

流程：
```text
1. 用户进入结算页
   → 请求服务端生成Token
   → 服务端生成Token并存Redis
   → 返回Token给前端

2. 用户提交订单
   → 携带Token
   → 服务端校验Token是否存在
   → 存在则删除Token，创建订单
   → 不存在则拒绝（重复提交）
```

实现：
```java
// 生成Token
public String generateOrderToken(Long userId) {
  String token = UUID.randomUUID().toString();
  String key = "order:token:" + token;
  redis.setex(key, 300, userId.toString()); // 5分钟有效
  return token;
}

// 创建订单（校验Token）
@Transactional
public Order createOrder(OrderRequest request, String token) {
  String key = "order:token:" + token;
  
  // 检查Token是否存在
  String userId = redis.get(key);
  if (userId == null) {
    throw new BizException("订单Token无效或已使用");
  }
  
  // 验证Token归属
  if (!userId.equals(request.getUserId().toString())) {
    throw new BizException("订单Token不匹配");
  }
  
  // 删除Token（保证一次性）
  redis.del(key);
  
  // 创建订单
  return doCreateOrder(request);
}
```

优点：
- 防止重复提交
- 安全性高（Token一次性）

缺点：
- 需要多次交互
- Token过期需要重新获取

**方案对比**：

| 方案 | 可靠性 | 易用性 | 性能 | 适用场景 |
|------|--------|--------|------|----------|
| 前端防抖 | ★★☆☆☆ | ★★★★★ | ★★★★★ | 辅助手段 |
| 唯一索引 | ★★★★★ | ★★★★☆ | ★★★★☆ | 通用 |
| 分布式锁 | ★★★★☆ | ★★★☆☆ | ★★★☆☆ | 高并发 |
| Token机制 | ★★★★★ | ★★★☆☆ | ★★★★☆ | 安全性要求高 |

**推荐方案**：
采用**唯一索引+Token机制**的组合。

实施要点：

1. **多层防护**：
   ```
   L1：前端防抖（用户体验）
   L2：Token机制（防恶意）
   L3：唯一索引（最后防线）
   ```

2. **幂等键设计**：
   ```
   幂等键组成：
   userId + cartVersion + timestamp
   
   例如：
   123_v10_1679800000
   
   说明：
   - userId：用户ID
   - cartVersion：购物车版本（购物车内容变化版本号+1）
   - timestamp：提交时间戳（精确到秒）
   ```

3. **异常处理**：
   ```java
   try {
     return createOrder(request, token);
   } catch (DuplicateKeyException e) {
     // 唯一索引冲突，查询已存在的订单
     Order existingOrder = findByIdempotentKey(idempotentKey);
     return existingOrder;
   } catch (BizException e) {
     // Token无效等业务异常
     throw e;
   }
   ```

**延伸思考**：
1. 如何设计订单创建的限流（防止刷单）？
2. 订单创建失败如何回滚库存？
3. 分布式事务下如何保证订单创建的一致性？

---

#### 📊 题目6：订单的分布式事务设计（Saga模式）

**问题描述**：
订单创建涉及多个服务（订单服务、库存服务、优惠券服务、积分服务）。如何使用Saga模式保证分布式事务一致性？

**答案**：

**问题分析**：
订单创建的分布式事务流程：
1. 扣减库存（库存服务）
2. 核销优惠券（营销服务）
3. 扣减积分（会员服务）
4. 创建订单（订单服务）

任一环节失败，已执行的操作需要回滚。

**Saga模式实现**（使用Go）：

```go
package saga

import (
	"context"
	"fmt"
)

// SagaStep 定义Saga步骤
type SagaStep struct {
	Name         string
	Execute      func(ctx context.Context, data interface{}) error
	Compensate   func(ctx context.Context, data interface{}) error
}

// SagaOrchestrator Saga编排器
type SagaOrchestrator struct {
	steps []SagaStep
}

// Execute 执行Saga
func (s *SagaOrchestrator) Execute(ctx context.Context, data interface{}) error {
	executedSteps := make([]int, 0)
	
	// 正向执行
	for i, step := range s.steps {
		if err := step.Execute(ctx, data); err != nil {
			// 执行失败，触发补偿
			s.compensate(ctx, data, executedSteps)
			return fmt.Errorf("步骤 %s 执行失败: %w", step.Name, err)
		}
		executedSteps = append(executedSteps, i)
	}
	
	return nil
}

// compensate 执行补偿
func (s *SagaOrchestrator) compensate(ctx context.Context, data interface{}, executedSteps []int) {
	// 反向补偿
	for i := len(executedSteps) - 1; i >= 0; i-- {
		stepIndex := executedSteps[i]
		step := s.steps[stepIndex]
		
		if err := step.Compensate(ctx, data); err != nil {
			// 补偿失败，记录日志，转人工处理
			log.Errorf("步骤 %s 补偿失败: %v", step.Name, err)
		}
	}
}

// 订单创建Saga示例
func CreateOrderSaga(orderReq *CreateOrderRequest) error {
	saga := &SagaOrchestrator{
		steps: []SagaStep{
			// 步骤1：扣减库存
			{
				Name: "DeductInventory",
				Execute: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					return inventoryService.Deduct(ctx, req.Items)
				},
				Compensate: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					return inventoryService.Release(ctx, req.Items)
				},
			},
			// 步骤2：核销优惠券
			{
				Name: "UseCoupon",
				Execute: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					if req.CouponID == "" {
						return nil // 无优惠券，跳过
					}
					return couponService.Use(ctx, req.UserID, req.CouponID)
				},
				Compensate: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					if req.CouponID == "" {
						return nil
					}
					return couponService.Release(ctx, req.UserID, req.CouponID)
				},
			},
			// 步骤3：扣减积分
			{
				Name: "DeductPoints",
				Execute: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					if req.PointsToUse == 0 {
						return nil
					}
					return pointsService.Deduct(ctx, req.UserID, req.PointsToUse)
				},
				Compensate: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					if req.PointsToUse == 0 {
						return nil
					}
					return pointsService.Refund(ctx, req.UserID, req.PointsToUse)
				},
			},
			// 步骤4：创建订单
			{
				Name: "CreateOrder",
				Execute: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					order := &Order{
						OrderID:   generateOrderID(),
						UserID:    req.UserID,
						Items:     req.Items,
						Status:    OrderStatusPending,
					}
					return orderRepo.Create(ctx, order)
				},
				Compensate: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					// 订单创建失败不需要补偿（未持久化）
					return nil
				},
			},
		},
	}
	
	return saga.Execute(context.Background(), orderReq)
}
```

**优点**：
- 逻辑清晰（正向+补偿）
- 解耦各服务
- 支持长事务

**缺点**：
- 实现复杂
- 补偿可能失败（需要人工介入）
- 中间状态可见（不是强一致性）

**延伸思考**：
1. Saga补偿失败如何处理？
2. 如何设计Saga的可视化监控？
3. Saga vs 2PC（两阶段提交）如何选择？

---

#### 🔧 题目7：订单数据的分库分表设计

**问题描述**：
订单表数据量达到亿级，单表查询性能下降。如何设计订单的分库分表方案？

**答案**：

**问题分析**：
订单分库分表的核心要素：
1. 分片键选择（user_id还是order_id）
2. 分片数量（16、32、64、128）
3. 跨片查询（如运营查询某时间段订单）
4. 数据扩容

**方案一：按user_id分片（推荐）**

核心思想：
同一用户的订单存储在同一分片。

分片规则：
```go
// 分片数量
const ShardCount = 64

// 计算分片
func GetShardIndex(userID int64) int {
	return int(userID % ShardCount)
}

// 路由到数据源
func GetDataSource(userID int64) *sql.DB {
	shardIndex := GetShardIndex(userID)
	return dataSources[shardIndex]
}
```

表结构：
```sql
-- 64个库，每个库有orders表
database_00.orders
database_01.orders
...
database_63.orders

订单ID生成：
order_id = snowflake_id
不包含分片信息（通过user_id路由）
```

优点：
- 用户维度查询高效（"我的订单"）
- 单用户订单聚合容易
- 避免跨库JOIN

缺点：
- 按订单ID查询需要广播（查所有分片）
- 数据可能不均匀（大客户订单多）

**方案二：按order_id分片**

核心思想：
按订单ID散列分片。

分片规则：
```go
func GetShardIndex(orderID int64) int {
	return int(orderID % ShardCount)
}
```

订单ID生成（包含分片信息）：
```go
// 订单ID结构：分片位 + Snowflake ID
// 前6位：分片号（0-63）
// 后13位：Snowflake ID

func GenerateOrderID(userID int64) int64 {
	shardIndex := GetShardIndex(userID)
	snowflakeID := snowflake.Generate()
	
	// 组装：分片号（6位） + snowflake（13位）
	return int64(shardIndex)*1e13 + snowflakeID
}

// 解析分片
func ParseShard(orderID int64) int {
	return int(orderID / 1e13)
}
```

优点：
- 按订单ID查询高效（直接定位分片）
- 数据均匀

缺点：
- 用户维度查询需要广播
- "我的订单"查询慢

**方案三：复合分片**

核心思想：
主表按user_id分片，建立order_id到分片的映射表。

设计：
```text
主表（按user_id分片）：
shard_00.orders
shard_01.orders

映射表（不分片，单独集群）：
order_routing
├── order_id（主键）
├── shard_index（分片号）
└── user_id

查询流程：
1. 按订单ID查询：
   - 查询order_routing获取分片号
   - 路由到对应分片查询

2. 按用户ID查询：
   - 直接路由到用户分片
```

优点：
- 支持多种查询方式
- 灵活

缺点：
- 映射表是单点
- 实现复杂

**方案对比**：

| 方案 | 用户查询 | 订单查询 | 数据均匀度 | 实施难度 |
|------|---------|---------|-----------|---------|
| 按user_id | ★★★★★ | ★★☆☆☆ | ★★★☆☆ | ★★★★☆ |
| 按order_id | ★★☆☆☆ | ★★★★★ | ★★★★★ | ★★★★☆ |
| 复合分片 | ★★★★★ | ★★★★★ | ★★★★★ | ★★☆☆☆ |

**推荐方案**：
采用**按user_id分片**。

实施要点（Go实现）：

1. **分片路由中间件**：
   ```go
   package sharding
   
   import (
   	"context"
   	"database/sql"
   )
   
   // ShardingManager 分片管理器
   type ShardingManager struct {
   	dataSources []*sql.DB
   	shardCount  int
   }
   
   // NewShardingManager 创建分片管理器
   func NewShardingManager(dsns []string) (*ShardingManager, error) {
   	dbs := make([]*sql.DB, len(dsns))
   	for i, dsn := range dsns {
   		db, err := sql.Open("mysql", dsn)
   		if err != nil {
   			return nil, err
   		}
   		dbs[i] = db
   	}
   	
   	return &ShardingManager{
   		dataSources: dbs,
   		shardCount:  len(dsns),
   	}, nil
   }
   
   // GetDB 根据用户ID获取数据库连接
   func (sm *ShardingManager) GetDB(userID int64) *sql.DB {
   	shardIndex := userID % int64(sm.shardCount)
   	return sm.dataSources[shardIndex]
   }
   
   // ExecuteOnShard 在指定分片执行查询
   func (sm *ShardingManager) ExecuteOnShard(ctx context.Context, userID int64, 
   	fn func(*sql.DB) error) error {
   	db := sm.GetDB(userID)
   	return fn(db)
   }
   
   // Broadcast 广播到所有分片执行
   func (sm *ShardingManager) Broadcast(ctx context.Context, 
   	fn func(*sql.DB) error) []error {
   	errors := make([]error, 0)
   	for _, db := range sm.dataSources {
   		if err := fn(db); err != nil {
   			errors = append(errors, err)
   		}
   	}
   	return errors
   }
   ```

2. **订单Repository实现**：
   ```go
   type OrderRepository struct {
   	shardingMgr *ShardingManager
   }
   
   // Create 创建订单
   func (r *OrderRepository) Create(ctx context.Context, order *Order) error {
   	return r.shardingMgr.ExecuteOnShard(ctx, order.UserID, func(db *sql.DB) error {
   		query := `INSERT INTO orders (order_id, user_id, total_amount, status, created_at)
   		          VALUES (?, ?, ?, ?, ?)`
   		_, err := db.ExecContext(ctx, query, 
   			order.OrderID, order.UserID, order.TotalAmount, 
   			order.Status, time.Now())
   		return err
   	})
   }
   
   // FindByUserID 查询用户订单（单分片）
   func (r *OrderRepository) FindByUserID(ctx context.Context, userID int64, 
   	page, size int) ([]*Order, error) {
   	var orders []*Order
   	
   	err := r.shardingMgr.ExecuteOnShard(ctx, userID, func(db *sql.DB) error {
   		query := `SELECT * FROM orders 
   		          WHERE user_id=? 
   		          ORDER BY created_at DESC 
   		          LIMIT ? OFFSET ?`
   		rows, err := db.QueryContext(ctx, query, userID, size, (page-1)*size)
   		if err != nil {
   			return err
   		}
   		defer rows.Close()
   		
   		for rows.Next() {
   			order := &Order{}
   			// 扫描数据...
   			orders = append(orders, order)
   		}
   		return nil
   	})
   	
   	return orders, err
   }
   
   // FindByOrderID 按订单ID查询（需要广播）
   func (r *OrderRepository) FindByOrderID(ctx context.Context, orderID int64) (*Order, error) {
   	// 方案1：广播到所有分片查询（慢）
   	for _, db := range r.shardingMgr.dataSources {
   		order, err := queryFromDB(db, orderID)
   		if err == nil && order != nil {
   			return order, nil
   		}
   	}
   	return nil, ErrOrderNotFound
   	
   	// 方案2：维护order_id -> user_id映射（推荐）
   	// userID := r.getOrderUserMapping(orderID)
   	// return r.FindByUserAndOrderID(ctx, userID, orderID)
   }
   ```

3. **订单ID包含分片信息**：
   ```go
   // 订单ID结构：6位分片号 + 13位Snowflake
   
   func GenerateOrderIDWithShard(userID int64) int64 {
   	shardIndex := userID % ShardCount
   	snowflakeID := snowflake.NextID()
   	
   	// 组装：前6位是分片号
   	return shardIndex*1e13 + snowflakeID
   }
   
   // 解析分片号
   func ParseShardFromOrderID(orderID int64) int {
   	return int(orderID / 1e13)
   }
   
   // 直接定位查询
   func (r *OrderRepository) FindByOrderIDFast(ctx context.Context, orderID int64) (*Order, error) {
   	shardIndex := ParseShardFromOrderID(orderID)
   	db := r.shardingMgr.dataSources[shardIndex]
   	
   	query := `SELECT * FROM orders WHERE order_id=?`
   	row := db.QueryRowContext(ctx, query, orderID)
   	
   	order := &Order{}
   	err := row.Scan(&order.OrderID, &order.UserID, ...) 
   	return order, err
   }
   ```

4. **扩容方案**：
   ```
   扩容策略（64 → 128分片）：
   
   方案A：双写期
   1. 新建64个分片（总共128个）
   2. 新订单写入新分片规则
   3. 老订单保留在老分片
   4. 查询时先查新分片，未命中再查老分片
   
   方案B：一致性哈希
   1. 使用一致性哈希算法
   2. 扩容时只需迁移部分数据
   3. 数据迁移期间双写
   ```

**延伸思考**：
1. 如何设计分库分表的全局查询（如运营后台）？
2. 订单归档如何设计（冷热数据分离）？
3. 分库分表如何支持跨库JOIN？

---

#### 💡 题目8：订单履约流程的编排

**问题描述**：
订单支付成功后，需要依次执行：分配仓库、创建拣货单、打包、出库、创建运单、发货。如何设计订单履约流程的编排？

**答案**：

**推荐方案**：事件驱动+状态机

架构（Go实现）：
```go
package fulfillment

import (
	"context"
)

// FulfillmentEvent 履约事件
type FulfillmentEvent struct {
	OrderID   int64
	EventType string
	Data      map[string]interface{}
}

// FulfillmentOrchestrator 履约编排器
type FulfillmentOrchestrator struct {
	eventBus EventBus
}

// OnOrderPaid 订单支付事件处理
func (o *FulfillmentOrchestrator) OnOrderPaid(ctx context.Context, orderID int64) error {
	// 1. 分配仓库
	warehouse, err := o.allocateWarehouse(ctx, orderID)
	if err != nil {
		return err
	}
	
	// 2. 创建拣货单
	pickingOrder, err := o.createPickingOrder(ctx, orderID, warehouse.ID)
	if err != nil {
		return err
	}
	
	// 3. 发布拣货事件
	o.eventBus.Publish(&FulfillmentEvent{
		OrderID:   orderID,
		EventType: "PickingOrderCreated",
		Data: map[string]interface{}{
			"pickingOrderID": pickingOrder.ID,
			"warehouseID":    warehouse.ID,
		},
	})
	
	return nil
}

// OnPickingCompleted 拣货完成事件处理
func (o *FulfillmentOrchestrator) OnPickingCompleted(ctx context.Context, event *FulfillmentEvent) error {
	orderID := event.OrderID
	
	// 1. 打包
	if err := o.pack(ctx, orderID); err != nil {
		return err
	}
	
	// 2. 出库
	if err := o.outbound(ctx, orderID); err != nil {
		return err
	}
	
	// 3. 创建物流运单
	trackingNumber, err := o.createShipment(ctx, orderID)
	if err != nil {
		return err
	}
	
	// 4. 发布发货事件
	o.eventBus.Publish(&FulfillmentEvent{
		OrderID:   orderID,
		EventType: "OrderShipped",
		Data: map[string]interface{}{
			"trackingNumber": trackingNumber,
		},
	})
	
	return nil
}

// 事件监听器
func (o *FulfillmentOrchestrator) Start() {
	o.eventBus.Subscribe("OrderPaid", o.OnOrderPaid)
	o.eventBus.Subscribe("PickingCompleted", o.OnPickingCompleted)
	o.eventBus.Subscribe("PackingCompleted", o.OnPackingCompleted)
	// ...
}
```

**履约状态机**：
```go
type FulfillmentStatus int

const (
	FulfillmentPending      FulfillmentStatus = 0  // 待履约
	FulfillmentWarehouseAllocated FulfillmentStatus = 1  // 已分配仓库
	FulfillmentPicking      FulfillmentStatus = 2  // 拣货中
	FulfillmentPacked       FulfillmentStatus = 3  // 已打包
	FulfillmentOutbound     FulfillmentStatus = 4  // 已出库
	FulfillmentShipped      FulfillmentStatus = 5  // 已发货
	FulfillmentReceived     FulfillmentStatus = 6  // 已签收
)

// 状态流转规则
var fulfillmentTransitions = map[FulfillmentStatus][]FulfillmentStatus{
	FulfillmentPending:            {FulfillmentWarehouseAllocated},
	FulfillmentWarehouseAllocated: {FulfillmentPicking},
	FulfillmentPicking:            {FulfillmentPacked},
	FulfillmentPacked:             {FulfillmentOutbound},
	FulfillmentOutbound:           {FulfillmentShipped},
	FulfillmentShipped:            {FulfillmentReceived},
}

// UpdateStatus 更新履约状态
func (o *FulfillmentOrchestrator) UpdateStatus(ctx context.Context, 
	orderID int64, newStatus FulfillmentStatus) error {
	// 1. 查询当前状态
	currentStatus, err := o.getStatus(ctx, orderID)
	if err != nil {
		return err
	}
	
	// 2. 检查状态流转是否合法
	allowedTransitions := fulfillmentTransitions[currentStatus]
	if !contains(allowedTransitions, newStatus) {
		return fmt.Errorf("不允许从%v转换到%v", currentStatus, newStatus)
	}
	
	// 3. 更新状态
	return o.updateStatusInDB(ctx, orderID, newStatus)
}
```

**延伸思考**：
1. 履约流程如何支持异常处理（缺货、商品损坏）？
2. 多个发货单如何协调履约进度？
3. 履约时效如何监控和告警？

---

#### 📊 题目9：订单的退款和售后流程设计

**问题描述**：
用户申请退款（仅退款、退货退款），如何设计售后流程，保证资金安全和用户体验？

**答案**：

**退款场景**：
1. 仅退款（未发货）
2. 退货退款（已发货）
3. 部分退款（退部分商品）
4. 售后退款（商品质量问题）

**推荐方案**（Go实现）：

退款状态机：
```go
type RefundStatus int

const (
	RefundPending   RefundStatus = 0  // 待审核
	RefundApproved  RefundStatus = 1  // 已同意
	RefundRejected  RefundStatus = 2  // 已拒绝
	RefundReturning RefundStatus = 3  // 退货中
	RefundReturned  RefundStatus = 4  // 已退货
	RefundCompleted RefundStatus = 5  // 已退款
)

// Refund 退款单
type Refund struct {
	RefundID     int64
	OrderID      int64
	UserID       int64
	RefundType   string  // REFUND_ONLY, RETURN_REFUND
	RefundAmount decimal.Decimal
	Reason       string
	Status       RefundStatus
	CreatedAt    time.Time
}

// RefundService 退款服务
type RefundService struct {
	orderRepo   OrderRepository
	paymentSvc  PaymentService
	inventorySvc InventoryService
}

// CreateRefund 创建退款申请
func (s *RefundService) CreateRefund(ctx context.Context, req *RefundRequest) (*Refund, error) {
	// 1. 校验订单状态
	order, err := s.orderRepo.FindByID(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}
	
	if order.Status != OrderStatusPaid && order.Status != OrderStatusShipped {
		return nil, errors.New("订单状态不允许退款")
	}
	
	// 2. 校验退款金额
	if req.RefundAmount.GreaterThan(order.PaidAmount) {
		return nil, errors.New("退款金额超过实付金额")
	}
	
	// 3. 创建退款单
	refund := &Refund{
		RefundID:     generateRefundID(),
		OrderID:      req.OrderID,
		UserID:       req.UserID,
		RefundType:   req.RefundType,
		RefundAmount: req.RefundAmount,
		Reason:       req.Reason,
		Status:       RefundPending,
		CreatedAt:    time.Now(),
	}
	
	if err := s.refundRepo.Create(ctx, refund); err != nil {
		return nil, err
	}
	
	// 4. 自动审核（部分场景）
	if s.shouldAutoApprove(refund) {
		return s.Approve(ctx, refund.RefundID)
	}
	
	return refund, nil
}

// Approve 审核通过退款
func (s *RefundService) Approve(ctx context.Context, refundID int64) (*Refund, error) {
	refund, err := s.refundRepo.FindByID(ctx, refundID)
	if err != nil {
		return nil, err
	}
	
	// 1. 更新退款状态
	refund.Status = RefundApproved
	if err := s.refundRepo.Update(ctx, refund); err != nil {
		return nil, err
	}
	
	// 2. 根据退款类型处理
	if refund.RefundType == "REFUND_ONLY" {
		// 仅退款：直接退款
		return s.processRefund(ctx, refund)
	} else {
		// 退货退款：等待用户退货
		refund.Status = RefundReturning
		s.refundRepo.Update(ctx, refund)
		// 生成退货地址和快递单号
		s.generateReturnLabel(ctx, refund)
		return refund, nil
	}
}

// processRefund 执行退款
func (s *RefundService) processRefund(ctx context.Context, refund *Refund) (*Refund, error) {
	// 1. 调用支付服务退款
	if err := s.paymentSvc.Refund(ctx, refund.OrderID, refund.RefundAmount); err != nil {
		return nil, fmt.Errorf("退款失败: %w", err)
	}
	
	// 2. 回补库存
	order, _ := s.orderRepo.FindByID(ctx, refund.OrderID)
	if err := s.inventorySvc.Return(ctx, order.Items); err != nil {
		log.Errorf("回补库存失败: %v", err)
		// 不阻塞退款流程，记录异常任务
		s.createCompensationTask(ctx, "ReturnInventory", refund.RefundID)
	}
	
	// 3. 更新退款状态
	refund.Status = RefundCompleted
	if err := s.refundRepo.Update(ctx, refund); err != nil {
		return nil, err
	}
	
	// 4. 更新订单状态
	s.orderRepo.UpdateStatus(ctx, refund.OrderID, OrderStatusRefunded)
	
	// 5. 发送通知
	s.notifySvc.Send(ctx, refund.UserID, "退款已到账")
	
	return refund, nil
}
```

**自动审核规则**：
```go
func (s *RefundService) shouldAutoApprove(refund *Refund) bool {
	// 自动同意条件：
	// 1. 订单未发货
	// 2. 退款金额 < 500元
	// 3. 用户信用良好
	
	order, _ := s.orderRepo.FindByID(context.Background(), refund.OrderID)
	
	if order.Status == OrderStatusPaid &&
		refund.RefundAmount.LessThan(decimal.NewFromInt(500)) &&
		s.userSvc.IsTrusted(refund.UserID) {
		return true
	}
	
	return false
}
```

**延伸思考**：
1. 退款失败如何重试和补偿？
2. 恶意退款如何识别和防范？
3. 部分退款如何计算退款金额（商品价+运费分摊）？

---

#### 🔧 题目10：订单的异常处理（缺货、地址错误）

**问题描述**：
订单履约过程中可能出现异常（缺货、地址无法送达、商品损坏）。如何设计异常处理流程？

**答案**：

**异常场景及处理方案**：

1. **库存不足（超卖）**：
   ```go
   // 发现超卖
   func (s *FulfillmentService) HandleOutOfStock(ctx context.Context, orderID int64) error {
   	// 1. 联系用户
   	s.notifySvc.Send(ctx, order.UserID, "商品暂时缺货，为您申请退款")
   	
   	// 2. 创建退款
   	refund := &Refund{
   		OrderID:      orderID,
   		RefundType:   "OUT_OF_STOCK",
   		RefundAmount: order.PaidAmount,
   		AutoApprove:  true,
   	}
   	return s.refundSvc.CreateRefund(ctx, refund)
   }
   ```

2. **地址无法送达**：
   ```go
   func (s *FulfillmentService) HandleUndeliverableAddress(ctx context.Context, 
   	orderID int64) error {
   	// 1. 通知用户修改地址
   	s.notifySvc.Send(ctx, order.UserID, "收货地址无法送达，请修改地址")
   	
   	// 2. 订单挂起
   	s.orderRepo.UpdateStatus(ctx, orderID, OrderStatusAddressError)
   	
   	// 3. 用户修改地址后重新履约
   	// 或超时自动退款
   	s.scheduleAutoRefund(ctx, orderID, 48*time.Hour)
   	
   	return nil
   }
   ```

3. **商品损坏**：
   ```go
   func (s *FulfillmentService) HandleDamaged(ctx context.Context, 
   	orderID int64, itemID string) error {
   	// 1. 记录损坏
   	s.logDamage(ctx, orderID, itemID)
   	
   	// 2. 检查是否有替代品
   	if hasReplace, err := s.inventorySvc.CheckStock(ctx, itemID); err == nil && hasReplace {
   		// 有替代品，重新拣货
   		return s.repick(ctx, orderID, itemID)
   	}
   	
   	// 3. 无替代品，部分退款
   	item := s.getOrderItem(ctx, orderID, itemID)
   	return s.refundSvc.CreatePartialRefund(ctx, orderID, item.Amount)
   }
   ```

**延伸思考**：
1. 异常订单如何统计和分析？
2. 如何设计异常的自动化处理规则？

---

#### 💡 题目11：订单的搜索和查询优化

**问题描述**：
用户需要查询历史订单（按时间、状态、商品筛选），运营需要查询全部订单。如何设计订单查询系统？

**答案**：

**方案一：主从分离**

用户查询（读从库）：
```go
// 查询我的订单
func (r *OrderRepository) FindUserOrders(ctx context.Context, 
	userID int64, filter *OrderFilter) ([]*Order, error) {
	// 路由到从库
	db := r.shardingMgr.GetReadDB(userID)
	
	query := `SELECT * FROM orders WHERE user_id=?`
	args := []interface{}{userID}
	
	// 添加筛选条件
	if filter.Status != "" {
		query += ` AND status=?`
		args = append(args, filter.Status)
	}
	
	if !filter.StartTime.IsZero() {
		query += ` AND created_at >= ?`
		args = append(args, filter.StartTime)
	}
	
	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, filter.PageSize, filter.Offset)
	
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return scanOrders(rows)
}
```

**方案二：ES同步（推荐）**

架构：
```text
订单创建/更新 → Kafka → 同步Worker → Elasticsearch

ES索引设计：
{
  "order_id": "123",
  "user_id": 456,
  "status": "PAID",
  "total_amount": 1000,
  "created_at": "2024-04-18T10:00:00Z",
  "items": [
    {"sku_id": "789", "title": "iPhone 15"}
  ]
}
```

查询实现：
```go
// 复杂查询用ES
func (r *OrderRepository) SearchOrders(ctx context.Context, 
	query *OrderSearchQuery) (*SearchResult, error) {
	esQuery := elastic.NewBoolQuery()
	
	// 用户维度
	if query.UserID > 0 {
		esQuery.Must(elastic.NewTermQuery("user_id", query.UserID))
	}
	
	// 订单号
	if query.OrderID != "" {
		esQuery.Must(elastic.NewTermQuery("order_id", query.OrderID))
	}
	
	// 状态
	if len(query.Statuses) > 0 {
		esQuery.Must(elastic.NewTermsQuery("status", query.Statuses...))
	}
	
	// 时间范围
	if !query.StartTime.IsZero() || !query.EndTime.IsZero() {
		rangeQuery := elastic.NewRangeQuery("created_at")
		if !query.StartTime.IsZero() {
			rangeQuery.Gte(query.StartTime)
		}
		if !query.EndTime.IsZero() {
			rangeQuery.Lte(query.EndTime)
		}
		esQuery.Must(rangeQuery)
	}
	
	// 商品筛选（嵌套查询）
	if query.SkuID != "" {
		esQuery.Must(elastic.NewNestedQuery("items",
			elastic.NewTermQuery("items.sku_id", query.SkuID)))
	}
	
	// 执行查询
	searchResult, err := r.esClient.Search().
		Index("orders").
		Query(esQuery).
		From(query.From).
		Size(query.Size).
		Sort("created_at", false).
		Do(ctx)
	
	if err != nil {
		return nil, err
	}
	
	return parseESResult(searchResult), nil
}
```

**延伸思考**：
1. 订单数据如何归档（如1年前的订单）？
2. 分库分表+ES同步如何保证一致性？

---

#### 📊 题目12：订单的消息通知设计

**问题描述**：
订单状态变化时需要通知用户（下单成功、发货、签收）。如何设计消息通知系统？

**答案**：

**通知渠道**：
1. App推送
2. 短信
3. 微信公众号/服务号
4. 站内信
5. 邮件

**推荐方案**（Go实现）：

```go
package notification

import (
	"context"
)

// NotificationService 通知服务
type NotificationService struct {
	pushSvc     PushService     // App推送
	smsSvc      SMSService      // 短信
	wechatSvc   WechatService   // 微信
	emailSvc    EmailService    // 邮件
	inboxSvc    InboxService    // 站内信
}

// NotifyOrderStatusChanged 订单状态变更通知
func (s *NotificationService) NotifyOrderStatusChanged(ctx context.Context, 
	order *Order, oldStatus, newStatus OrderStatus) error {
	
	// 根据状态确定通知内容
	template := s.getTemplate(newStatus)
	
	// 并行发送多渠道通知
	errChan := make(chan error, 5)
	
	// 1. App推送（必发）
	go func() {
		errChan <- s.pushSvc.Push(ctx, order.UserID, PushMessage{
			Title:   template.Title,
			Content: template.Content,
			Data:    map[string]interface{}{"order_id": order.OrderID},
		})
	}()
	
	// 2. 短信（重要状态才发）
	if s.shouldSendSMS(newStatus) {
		go func() {
			phone := s.getUserPhone(ctx, order.UserID)
			errChan <- s.smsSvc.Send(ctx, phone, template.SMSContent)
		}()
	} else {
		errChan <- nil
	}
	
	// 3. 微信（用户已绑定才发）
	go func() {
		if openID := s.getUserWechatOpenID(ctx, order.UserID); openID != "" {
			errChan <- s.wechatSvc.SendTemplateMessage(ctx, openID, template.WechatTemplate)
		} else {
			errChan <- nil
		}
	}()
	
	// 4. 站内信（必发）
	go func() {
		errChan <- s.inboxSvc.Create(ctx, &InboxMessage{
			UserID:  order.UserID,
			Title:   template.Title,
			Content: template.Content,
			Type:    "ORDER_UPDATE",
		})
	}()
	
	// 5. 邮件（用户订阅才发）
	go func() {
		if s.userHasEmailSubscription(ctx, order.UserID) {
			email := s.getUserEmail(ctx, order.UserID)
			errChan <- s.emailSvc.Send(ctx, email, template.EmailContent)
		} else {
			errChan <- nil
		}
	}()
	
	// 收集结果（至少一个渠道成功即可）
	successCount := 0
	for i := 0; i < 5; i++ {
		if err := <-errChan; err == nil {
			successCount++
		}
	}
	
	if successCount == 0 {
		return errors.New("所有通知渠道都失败")
	}
	
	return nil
}

// 通知模板
func (s *NotificationService) getTemplate(status OrderStatus) *NotificationTemplate {
	templates := map[OrderStatus]*NotificationTemplate{
		OrderStatusPaid: {
			Title:       "订单支付成功",
			Content:     "您的订单已支付成功，我们将尽快为您发货",
			SMSContent:  "【京东】您的订单已支付成功，预计3天内送达",
		},
		OrderStatusShipped: {
			Title:       "订单已发货",
			Content:     "您的订单已发货，快递单号：SF1234567890",
			SMSContent:  "【京东】您的订单已发货，单号SF1234567890",
		},
		OrderStatusReceived: {
			Title:       "订单已签收",
			Content:     "您的订单已签收，期待您的评价",
		},
	}
	
	return templates[status]
}

// 是否发送短信
func (s *NotificationService) shouldSendSMS(status OrderStatus) bool {
	// 只有关键状态发短信（控制成本）
	importantStatuses := []OrderStatus{
		OrderStatusPaid,
		OrderStatusShipped,
		OrderStatusRefunded,
	}
	
	for _, s := range importantStatuses {
		if s == status {
			return true
		}
	}
	return false
}
```

**延伸思考**：
1. 通知失败如何重试？
2. 如何设计通知的用户偏好设置（关闭某些通知）？
3. 大批量通知如何限流（避免骚扰）？

---

#### 🔧 题目13：订单数据的冷热分离

**问题描述**：
订单数据90天后很少查询，但占用大量存储。如何设计订单数据的冷热分离？

**答案**：

**推荐方案**：

```go
// 冷热分离策略
type OrderArchiveService struct {
	hotDB  *sql.DB  // 热数据库（MySQL）
	coldDB *sql.DB  // 冷数据库（可以是低成本存储）
	ossClient OSSClient // 对象存储
}

// 归档策略
func (s *OrderArchiveService) ArchiveOrders(ctx context.Context) error {
	// 1. 查询90天前已完成的订单
	cutoffTime := time.Now().AddDate(0, 0, -90)
	
	query := `SELECT * FROM orders 
	          WHERE status IN ('COMPLETED', 'CANCELLED', 'REFUNDED')
	          AND updated_at < ?
	          LIMIT 1000`
	
	rows, err := s.hotDB.QueryContext(ctx, query, cutoffTime)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	orders := make([]*Order, 0)
	for rows.Next() {
		order := &Order{}
		// 扫描数据...
		orders = append(orders, order)
	}
	
	// 2. 写入冷库
	for _, order := range orders {
		if err := s.writeToArchive(ctx, order); err != nil {
			log.Errorf("归档订单%d失败: %v", order.OrderID, err)
			continue
		}
		
		// 3. 删除热库数据
		if err := s.deleteFromHot(ctx, order.OrderID); err != nil {
			log.Errorf("删除热库订单%d失败: %v", order.OrderID, err)
		}
	}
	
	return nil
}

// 查询时智能路由
func (s *OrderArchiveService) FindByID(ctx context.Context, orderID int64) (*Order, error) {
	// 1. 先查热库
	order, err := s.queryFromHot(ctx, orderID)
	if err == nil && order != nil {
		return order, nil
	}
	
	// 2. 查冷库
	order, err = s.queryFromArchive(ctx, orderID)
	if err == nil && order != nil {
		return order, nil
	}
	
	return nil, ErrOrderNotFound
}
```

**延伸思考**：
1. 归档订单如何支持查询？
2. 冷数据恢复到热库的策略？

---

#### 💡 题目14：订单的限流和防刷

**问题描述**：
恶意用户频繁下单不支付，占用库存和系统资源。如何设计订单的限流和防刷机制？

**答案**：

**推荐方案**（Go实现）：

```go
package ratelimit

import (
	"context"
	"fmt"
	"time"
	
	"github.com/go-redis/redis/v8"
)

// OrderRateLimiter 订单限流器
type OrderRateLimiter struct {
	rdb *redis.Client
}

// CheckLimit 检查用户是否超过限流
func (l *OrderRateLimiter) CheckLimit(ctx context.Context, userID int64) error {
	// 限流规则：
	// 1. 每分钟最多下单5次
	// 2. 每小时最多下单20次
	// 3. 每天最多50个待支付订单
	
	// 规则1：每分钟限流
	key1 := fmt.Sprintf("order:limit:min:%d:%s", userID, time.Now().Format("200601021504"))
	count1, err := l.rdb.Incr(ctx, key1).Result()
	if err != nil {
		return err
	}
	if count1 == 1 {
		l.rdb.Expire(ctx, key1, time.Minute)
	}
	if count1 > 5 {
		return errors.New("下单太频繁，请稍后再试")
	}
	
	// 规则2：每小时限流
	key2 := fmt.Sprintf("order:limit:hour:%d:%s", userID, time.Now().Format("2006010215"))
	count2, err := l.rdb.Incr(ctx, key2).Result()
	if err != nil {
		return err
	}
	if count2 == 1 {
		l.rdb.Expire(ctx, key2, time.Hour)
	}
	if count2 > 20 {
		return errors.New("您今天下单次数过多，请明天再试")
	}
	
	// 规则3：待支付订单数量限制
	pendingCount, err := l.getPendingOrderCount(ctx, userID)
	if err != nil {
		return err
	}
	if pendingCount >= 50 {
		return errors.New("您有过多待支付订单，请先完成支付")
	}
	
	return nil
}

// 用户信用评分
type UserCreditService struct {
	repo UserCreditRepository
}

func (s *UserCreditService) CheckCredit(ctx context.Context, userID int64) error {
	credit := s.repo.GetCredit(ctx, userID)
	
	// 信用分低于60分，禁止下单
	if credit.Score < 60 {
		return errors.New("您的信用分过低，暂时无法下单")
	}
	
	return nil
}

// 信用分扣减规则
func (s *UserCreditService) UpdateCredit(ctx context.Context, userID int64, behavior string) {
	switch behavior {
	case "ORDER_TIMEOUT":
		// 订单超时未支付：-5分
		s.repo.DeductCredit(ctx, userID, 5, "订单超时未支付")
	case "MALICIOUS_REFUND":
		// 恶意退款：-10分
		s.repo.DeductCredit(ctx, userID, 10, "恶意退款")
	case "ORDER_COMPLETED":
		// 订单完成：+1分
		s.repo.AddCredit(ctx, userID, 1, "订单完成")
	}
}
```

**延伸思考**：
1. 如何识别黄牛和恶意用户？
2. 限流策略如何针对不同用户等级差异化？

---

#### 📊 题目15：订单的实时数据统计

**问题描述**：
运营大盘需要实时显示订单量、GMV、转化率。如何设计订单的实时统计系统？

**答案**：

**推荐方案**：Flink流式计算

```go
// 实时统计指标
type OrderMetrics struct {
	Timestamp      time.Time
	OrderCount     int64           // 订单数
	GMV            decimal.Decimal // 交易额
	PaidOrderCount int64           // 已支付订单数
	AvgOrderAmount decimal.Decimal // 客单价
}

// 指标计算Worker（消费Kafka）
func ConsumeOrderEvents(ctx context.Context) {
	consumer := kafka.NewConsumer(...)
	
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			continue
		}
		
		event := parseOrderEvent(msg.Value)
		
		switch event.Type {
		case "OrderCreated":
			// 订单数+1
			metrics.IncrOrderCount()
			
		case "OrderPaid":
			// 已支付订单数+1
			metrics.IncrPaidOrderCount()
			// GMV累加
			metrics.AddGMV(event.Order.PaidAmount)
			
		case "OrderCancelled":
			// 订单数-1（或单独统计取消数）
			metrics.IncrCancelledOrderCount()
		}
		
		// 定期刷新到Redis
		if time.Now().Unix()%10 == 0 {
			metrics.FlushToRedis()
		}
	}
}

// 实时大盘查询
func GetRealTimeMetrics(ctx context.Context) (*OrderMetrics, error) {
	// 从Redis读取实时指标
	rdb := redis.NewClient(...)
	
	orderCount, _ := rdb.Get(ctx, "metrics:order:count").Int64()
	gmv, _ := rdb.Get(ctx, "metrics:order:gmv").Float64()
	paidCount, _ := rdb.Get(ctx, "metrics:order:paid_count").Int64()
	
	return &OrderMetrics{
		Timestamp:      time.Now(),
		OrderCount:     orderCount,
		GMV:            decimal.NewFromFloat(gmv),
		PaidOrderCount: paidCount,
		AvgOrderAmount: decimal.NewFromFloat(gmv).Div(decimal.NewFromInt(paidCount)),
	}, nil
}
```

**延伸思考**：
1. 实时统计如何保证准确性（与离线对账）？
2. 多维度统计（按类目、品牌）如何设计？

---
