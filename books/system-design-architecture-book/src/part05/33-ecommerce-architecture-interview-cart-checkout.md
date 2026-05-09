# 35.3.2 购物车与结算题库

## 35.3.2 购物车与结算（15题）

#### 📊 题目1：购物车的数据存储设计

**问题描述**：
用户将商品加入购物车，需要跨设备同步（手机APP、Web、小程序）。如何设计购物车的存储方案？

**答案**：

**问题分析**：
购物车的核心要素：
1. 跨设备同步
2. 用户未登录也能加购
3. 数据持久化
4. 高并发读写

**方案一：Cookie存储**

核心思想：
购物车数据存储在浏览器Cookie。

优点：
- 无需服务器存储
- 减轻服务器压力

缺点：
- 不能跨设备
- Cookie大小限制（4KB）
- 不安全（可被篡改）

适用场景：
- 简单电商
- 临时购物车

**方案二：数据库存储（推荐）**

核心思想：
购物车存储在MySQL/Redis。

设计：
```sql
shopping_cart
├── cart_id
├── user_id
├── sku_id
├── quantity
├── selected（是否选中，用于结算）
├── added_at
└── updated_at

索引：
- PRIMARY KEY (cart_id)
- UNIQUE KEY (user_id, sku_id)
- INDEX (user_id)
```

优点：
- 跨设备同步
- 数据持久化
- 支持复杂操作

缺点：
- 服务器存储成本

**方案三：Redis+MySQL双写**

核心思想：
Redis提供高性能，MySQL保证持久化。

架构：
```text
写操作：
1. 写Redis（立即返回）
2. 异步写MySQL

读操作：
1. 优先读Redis
2. Redis不存在，读MySQL
3. 回写Redis
```

优点：
- 性能高
- 数据安全

缺点：
- 数据同步复杂

**推荐方案**：
采用**Redis+MySQL双写**。

实施要点：

1. **未登录用户**：
   ```
   未登录：
   - 生成临时cart_id（存Cookie）
   - 购物车数据存Redis
   - key: cart:temp:{cart_id}
   
   登录后：
   - 合并临时购物车到用户购物车
   - 删除临时购物车
   ```

2. **购物车合并**：
   ```java
   public void mergeCart(String tempCartId, Long userId) {
     List<CartItem> tempItems = getTempCart(tempCartId);
     List<CartItem> userItems = getUserCart(userId);
     
     for (CartItem temp : tempItems) {
       CartItem exist = findItem(userItems, temp.getSkuId());
       if (exist != null) {
         // 已存在，数量相加
         exist.setQuantity(exist.getQuantity() + temp.getQuantity());
       } else {
         // 不存在，添加
         userItems.add(temp);
       }
     }
     
     saveUserCart(userId, userItems);
     deleteTempCart(tempCartId);
   }
   ```

3. **失效商品处理**：
   ```
   商品失效场景：
   - 商品下架
   - 商品删除
   - 库存不足
   
   展示：
   - 失效商品置灰
   - 提示"商品已下架"
   - 提供"删除"或"移入收藏"选项
   ```

4. **购物车清理**：
   ```
   定时任务（每天凌晨）：
   - 删除90天未更新的购物车
   - 减少存储成本
   ```

5. **购物车同步**：
   ```
   跨设备同步：
   - 用户在APP加购 → 写Redis+MySQL
   - 用户在Web打开 → 读Redis → 显示购物车
   
   实时同步（WebSocket）：
   - 用户在设备A加购
   - 推送到设备B
   - 设备B实时更新购物车数量
   ```

**延伸思考**：
1. 购物车数量显示在导航栏，如何实时更新？
2. 如何处理购物车中的促销信息过期？
3. 购物车数据如何备份和恢复？

---

#### 🔧 题目2：购物车的价格计算

**问题描述**：
购物车中有多个商品，每个商品可能有不同促销（满减、折扣、优惠券）。如何设计购物车的实时价格计算？

**答案**：

**问题分析**：
购物车价格计算的复杂性：
1. 多商品组合
2. 多种促销叠加
3. 实时计算（用户修改数量即刻更新）
4. 价格明细展示

**推荐方案**：

价格计算引擎：
```java
public CartPrice calculateCart(Cart cart) {
  BigDecimal originalPrice = BigDecimal.ZERO;
  BigDecimal discountAmount = BigDecimal.ZERO;
  
  // 1. 计算商品级优惠
  for (CartItem item : cart.getItems()) {
    originalPrice = originalPrice.add(
      item.getPrice().multiply(new BigDecimal(item.getQuantity()))
    );
    
    // 商品折扣
    if (item.hasDiscount()) {
      BigDecimal itemDiscount = calculateItemDiscount(item);
      discountAmount = discountAmount.add(itemDiscount);
    }
  }
  
  // 2. 计算订单级优惠
  BigDecimal subtotal = originalPrice.subtract(discountAmount);
  
  // 满减
  BigDecimal fullReduceDiscount = calculateFullReduce(subtotal);
  discountAmount = discountAmount.add(fullReduceDiscount);
  
  // 优惠券
  if (cart.hasCoupon()) {
    BigDecimal couponDiscount = calculateCoupon(cart.getCoupon(), subtotal);
    discountAmount = discountAmount.add(couponDiscount);
  }
  
  // 3. 最终价格
  BigDecimal finalPrice = originalPrice.subtract(discountAmount);
  
  return new CartPrice(originalPrice, discountAmount, finalPrice);
}
```

实时计算触发：
```text
触发时机：
- 用户修改商品数量
- 用户选择/取消优惠券
- 用户勾选/取消商品
- 商品价格变动（后台推送）

性能优化：
- 防抖（用户停止操作500ms后计算）
- 缓存（相同购物车缓存5分钟）
```

**延伸思考**：
1. 购物车价格和下单后价格不一致如何处理？
2. 大促时购物车价格计算如何优化性能？

---

#### 💡 题目3：购物车的推荐功能

**问题描述**：
用户购物车中有商品A，如何推荐相关商品B，提升客单价？

**答案**：

**推荐策略**：

1. **关联推荐**：
   ```
   "买了还买"：
   - 统计购买商品A的用户还购买了哪些商品
   - 推荐高频商品
   
   示例：
   购物车有"iPhone 15" → 推荐"手机壳"、"钢化膜"、"充电器"
   ```

2. **凑单推荐**：
   ```
   购物车总价¥180
   满¥200减¥30
   
   推荐：再买¥20-30的商品，即可享受优惠
   ```

3. **替代推荐**：
   ```
   购物车中商品缺货 → 推荐同类商品
   ```

**延伸思考**：
1. 购物车推荐如何避免打扰用户？
2. 推荐商品点击率如何提升？

---

#### 📊 题目4：购物车的库存校验

**问题描述**：
用户加购物车时商品有货，结算时可能已无货。如何设计购物车的库存校验机制？

**答案**：

**校验时机**：

1. **加购时校验**：
   ```
   用户点击"加入购物车" → 检查库存
   库存充足 → 允许加购
   库存不足 → 提示"库存不足"
   ```

2. **结算时校验**：
   ```
   用户点击"去结算" → 
   1. 批量查询购物车所有商品库存
   2. 标记缺货商品
   3. 展示：
      - 有货商品（可结算）
      - 缺货商品（置灰，不可结算）
   ```

3. **实时推送**：
   ```
   商品库存变化（如售罄） → WebSocket推送
   前端实时更新购物车状态
   ```

**延伸思考**：
1. 购物车中的商品是否需要预占库存？
2. 库存不足时如何引导用户？

---

#### 🔧 题目5：购物车的性能优化

**问题描述**：
大促期间，购物车服务QPS达10万+，如何优化购物车性能？

**答案**：

**优化方案**：

1. **读写分离**：
   ```
   写操作（加购、删除）：
   - 写MySQL主库
   - 异步同步到Redis
   
   读操作（查询购物车）：
   - 读Redis（快）
   - 未命中读MySQL从库
   ```

2. **批量操作**：
   ```
   ❌ 单个加购：N次请求
   ✅ 批量加购：1次请求
   
   POST /api/cart/batch-add
   {
     "items": [
       {"skuId": "123", "quantity": 2},
       {"skuId": "456", "quantity": 1}
     ]
   }
   ```

3. **本地缓存**：
   ```
   热点用户购物车：
   - 加载到应用服务器内存
   - 减少Redis访问
   ```

4. **限流降级**：
   ```
   限流：
   - 单用户购物车操作频率限制（10次/分钟）
   
   降级：
   - Redis故障 → 降级到MySQL
   - MySQL故障 → 只读模式（不能加购）
   ```

**延伸思考**：
1. 购物车数据如何分片（sharding）？
2. 购物车服务如何实现高可用？

---

#### 📊 题目6：购物车商品失效的处理策略

**问题描述**：
用户购物车中的商品可能因为下架、删除、库存清零而失效。如何设计失效商品的处理策略，优化用户体验？

**答案**：

**问题分析**：
商品失效场景：
1. 商品下架（运营操作）
2. 商品删除（商品不再销售）
3. 库存售罄（暂时缺货）
4. 商品涨价（价格变动）
5. 促销过期（活动结束）

**方案一：定时批量检测**

核心思想：
定时任务扫描购物车，标记失效商品。

实现：
```text
定时任务（每小时）：
1. 查询所有购物车商品
2. 批量查询商品状态
3. 标记失效商品
4. 更新购物车
```

优点：
- 批量处理，效率高
- 服务器压力均匀

缺点：
- 实时性差（最长延迟1小时）
- 用户可能看到失效商品

**方案二：实时校验（推荐）**

核心思想：
用户打开购物车时，实时校验商品状态。

流程：
```text
用户打开购物车 →
1. 查询购物车商品列表
2. 批量查询商品最新状态（Redis缓存）
3. 分类展示：
   - 正常商品（可结算）
   - 失效商品（置灰，不可结算）
4. 标注失效原因
```

失效商品展示：
```text
[置灰显示]
iPhone 15 Pro 256GB
¥7999
状态：该商品已下架
操作：[删除] [移入收藏夹]
```

优点：
- 实时性好
- 用户体验清晰

缺点：
- 每次打开购物车都校验
- QPS增加

**方案三：消息推送**

核心思想：
商品状态变化时，主动推送更新购物车。

架构：
```text
商品下架 → 
发布事件（Kafka）→ 
购物车Worker消费 →
1. 查询包含该商品的购物车
2. 标记商品为失效
3. WebSocket推送用户（如果在线）
```

优点：
- 实时性最好
- 用户感知及时

缺点：
- 架构复杂
- 需要消息队列

**方案对比**：

| 方案 | 实时性 | 用户体验 | 实施难度 | 系统负载 |
|------|--------|---------|---------|---------|
| 定时检测 | ★★☆☆☆ | ★★★☆☆ | ★★★★★ | ★★★★☆ |
| 实时校验 | ★★★★☆ | ★★★★★ | ★★★★☆ | ★★★☆☆ |
| 消息推送 | ★★★★★ | ★★★★★ | ★★☆☆☆ | ★★★★☆ |

**推荐方案**：
采用**实时校验+消息推送**的组合。

实施要点：

1. **商品状态缓存**：
   ```
   Redis存储商品状态：
   key: product:status:{skuId}
   value: {
     "onSale": true,
     "stock": 100,
     "price": 7999,
     "promotionId": "xxx",
     "updatedAt": 1679800000
   }
   TTL: 10分钟
   
   商品变更时主动刷新
   ```

2. **批量校验优化**：
   ```java
   public Map<String, ProductStatus> batchCheckStatus(List<String> skuIds) {
     // 1. 批量查询Redis
     List<String> keys = skuIds.stream()
       .map(id -> "product:status:" + id)
       .collect(Collectors.toList());
     
     List<ProductStatus> cached = redis.mget(keys);
     
     // 2. 未命中的查数据库
     Set<String> missingIds = findMissingIds(cached);
     if (!missingIds.isEmpty()) {
       Map<String, ProductStatus> fromDB = queryFromDB(missingIds);
       // 写回Redis
       cacheToRedis(fromDB);
       cached.addAll(fromDB.values());
     }
     
     return toMap(cached);
   }
   ```

3. **失效商品操作**：
   ```
   用户操作：
   1. 删除：直接从购物车删除
   2. 移入收藏夹：
      - 加入收藏
      - 从购物车删除
      - 商品恢复上架时通知用户
   3. 查看替代品：
      - 推荐同类商品
      - 一键替换
   ```

4. **主动通知**：
   ```
   通知策略：
   - 商品下架 → App推送
     "您购物车中的【iPhone 15】已下架"
   - 商品降价 → App推送
     "您购物车中的【iPhone 15】降价了"
   - 库存恢复 → 收藏夹商品有货通知
   ```

5. **失效原因分类**：
   ```
   原因分类：
   - 已下架：运营下架
   - 已售罄：库存为0
   - 已删除：商品不存在
   - 已涨价：价格变动超过10%
   - 活动结束：促销过期
   
   针对性提示：
   - 已售罄 → "补货中，可先收藏"
   - 已涨价 → "当前价格¥xxx，加购时¥xxx"
   ```

**延伸思考**：
1. 如何设计购物车的自动清理（失效商品30天后自动删除）？
2. 失效商品是否计入购物车数量显示？
3. 如何处理部分失效（如只有某个规格缺货）？

---

#### 🔧 题目7：购物车的跨平台同步设计

**问题描述**：
用户在手机APP加购商品，打开电脑Web也能看到。如何实现购物车的跨平台实时同步？

**答案**：

**问题分析**：
跨平台同步的核心要素：
1. 数据一致性（同一购物车）
2. 实时性（秒级同步）
3. 冲突处理（同时操作）
4. 离线支持

**方案一：轮询同步**

核心思想：
客户端定时轮询服务器，获取最新购物车。

实现：
```javascript
// 前端定时轮询
setInterval(() => {
  fetch('/api/cart')
    .then(res => res.json())
    .then(cart => {
      if (cart.version > localVersion) {
        updateLocalCart(cart);
      }
    });
}, 5000); // 每5秒轮询一次
```

优点：
- 实现简单
- 兼容性好

缺点：
- 实时性差（5秒延迟）
- 浪费带宽（大部分请求无变化）
- 服务器压力大

**方案二：WebSocket推送（推荐）**

核心思想：
客户端与服务器建立长连接，服务器主动推送更新。

架构：
```text
用户A在APP加购 →
1. APP发送请求到服务器
2. 服务器更新购物车
3. 服务器通过WebSocket推送到用户A的所有设备
4. Web端接收推送，更新购物车显示

WebSocket消息格式：
{
  "type": "CART_UPDATE",
  "action": "ADD_ITEM",
  "data": {
    "skuId": "123",
    "quantity": 2
  },
  "version": 10,
  "timestamp": 1679800000
}
```

实现：
```java
// 服务端
@Service
public class CartService {
  @Autowired
  private WebSocketPushService pushService;
  
  public void addToCart(Long userId, String skuId, int quantity) {
    // 1. 更新购物车
    Cart cart = updateCart(userId, skuId, quantity);
    
    // 2. 推送到该用户所有在线设备
    CartUpdateMessage msg = new CartUpdateMessage(
      "ADD_ITEM", skuId, quantity, cart.getVersion()
    );
    pushService.pushToUser(userId, msg);
  }
}

// 客户端
websocket.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'CART_UPDATE') {
    // 更新本地购物车
    if (msg.version > localCartVersion) {
      applyCartUpdate(msg);
    }
  }
};
```

优点：
- 实时性好（秒级）
- 双向通信
- 节省带宽

缺点：
- 需要维护长连接
- 服务器成本高
- 需要心跳保活

**方案三：长轮询**

核心思想：
客户端发起请求，服务器hold住请求，有更新时返回。

实现：
```javascript
function longPoll() {
  fetch('/api/cart/poll?version=' + localVersion)
    .then(res => res.json())
    .then(cart => {
      if (cart.version > localVersion) {
        updateLocalCart(cart);
      }
      // 立即发起下一次轮询
      longPoll();
    })
    .catch(() => {
      // 失败后延迟重试
      setTimeout(longPoll, 5000);
    });
}
```

优点：
- 实时性较好
- 兼容性好（不需要WebSocket）

缺点：
- 服务器需要hold请求
- 连接可能超时

**方案对比**：

| 方案 | 实时性 | 服务器成本 | 兼容性 | 实施难度 |
|------|--------|-----------|--------|---------|
| 轮询 | ★★☆☆☆ | ★★☆☆☆ | ★★★★★ | ★★★★★ |
| WebSocket | ★★★★★ | ★★★☆☆ | ★★★★☆ | ★★★☆☆ |
| 长轮询 | ★★★★☆ | ★★☆☆☆ | ★★★★★ | ★★★★☆ |

**推荐方案**：
采用**WebSocket推送**（支持WebSocket）+ **轮询兜底**（不支持时降级）。

实施要点：

1. **连接管理**：
   ```java
   // 用户连接映射
   Map<Long, Set<WebSocketSession>> userSessions = new ConcurrentHashMap<>();
   
   // 用户连接时
   public void onConnect(Long userId, WebSocketSession session) {
     userSessions.computeIfAbsent(userId, k -> new ConcurrentHashSet<>())
       .add(session);
   }
   
   // 用户断开时
   public void onDisconnect(Long userId, WebSocketSession session) {
     Set<WebSocketSession> sessions = userSessions.get(userId);
     if (sessions != null) {
       sessions.remove(session);
     }
   }
   
   // 推送消息
   public void pushToUser(Long userId, Object message) {
     Set<WebSocketSession> sessions = userSessions.get(userId);
     if (sessions != null) {
       for (WebSocketSession session : sessions) {
         if (session.isOpen()) {
           session.sendMessage(new TextMessage(JSON.toJSONString(message)));
         }
       }
     }
   }
   ```

2. **版本控制**：
   ```
   购物车版本号：
   - 每次修改version+1
   - 客户端记录本地version
   - 接收推送时检查version
   - 如果本地version更新，忽略旧推送
   
   冲突解决：
   - 客户端操作携带version
   - 服务端CAS更新
   - 失败则拉取最新数据重试
   ```

3. **心跳保活**：
   ```javascript
   // 客户端定时发送心跳
   setInterval(() => {
     if (websocket.readyState === WebSocket.OPEN) {
       websocket.send(JSON.stringify({type: 'PING'}));
     }
   }, 30000); // 每30秒
   
   // 服务端响应心跳
   if (message.type === 'PING') {
     session.sendMessage(new TextMessage('{"type":"PONG"}'));
   }
   ```

4. **降级策略**：
   ```javascript
   // 检测WebSocket支持
   if ('WebSocket' in window) {
     connectWebSocket();
   } else {
     // 降级到轮询
     setInterval(pollCart, 10000);
   }
   
   // WebSocket断开时降级
   websocket.onclose = () => {
     console.log('WebSocket断开，降级到轮询');
     setInterval(pollCart, 10000);
   };
   ```

5. **离线支持**：
   ```
   离线操作：
   1. 用户离线时，操作保存到本地队列
   2. 用户上线后，批量同步到服务器
   3. 服务器合并操作，返回最终购物车
   
   冲突处理：
   - 添加：合并数量
   - 删除：以最新操作为准
   - 修改：以最新操作为准
   ```

**延伸思考**：
1. 如何处理网络不稳定导致的频繁重连？
2. 跨平台同步如何支持多账号（家庭共享）？
3. WebSocket服务如何实现横向扩展？

---

#### 💡 题目8：购物车推荐算法设计

**问题描述**：
用户购物车有"iPhone 15"，如何推荐相关商品（配件、保险、AppleCare）提升客单价？

**答案**：

**问题分析**：
购物车推荐的核心目标：
1. 提升客单价（关联销售）
2. 提升转化率（凑单满减）
3. 提升用户体验（需要的商品）

**推荐策略**：

1. **关联推荐（Frequently Bought Together）**：
   ```sql
   -- 统计商品关联
   SELECT b.sku_id, COUNT(*) as frequency
   FROM order_items a
   JOIN order_items b ON a.order_id = b.order_id
   WHERE a.sku_id = 'iPhone15' 
     AND b.sku_id != 'iPhone15'
   GROUP BY b.sku_id
   ORDER BY frequency DESC
   LIMIT 10;
   
   结果：
   - 手机壳（购买率80%）
   - 钢化膜（购买率70%）
   - 充电器（购买率60%）
   ```

2. **凑单推荐**：
   ```
   购物车总价：¥180
   满减活动：满¥200减¥30
   
   推荐策略：
   - 推荐价格在¥20-¥50的商品
   - 优先推荐与购物车商品相关的
   - 标注"再买¥20即享满减"
   ```

3. **类目互补推荐**：
   ```
   购物车有"相机" → 推荐：
   - 存储卡
   - 相机包
   - 三脚架
   
   购物车有"婴儿奶粉" → 推荐：
   - 奶瓶
   - 尿不湿
   - 湿巾
   ```

4. **个性化推荐**：
   ```
   基于用户历史：
   - 用户A经常买Apple产品
     → 推荐AppleCare+、AirPods
   - 用户B价格敏感
     → 推荐高性价比配件
   ```

**实施要点**：

1. **关联规则挖掘**：
   ```python
   # 使用Apriori算法
   from mlxtend.frequent_patterns import apriori, association_rules
   
   # 构建购物篮矩阵
   basket = orders.groupby(['order_id', 'sku_id'])['quantity'].sum().unstack().fillna(0)
   basket = basket.applymap(lambda x: 1 if x > 0 else 0)
   
   # 挖掘频繁项集
   frequent_itemsets = apriori(basket, min_support=0.01, use_colnames=True)
   
   # 生成关联规则
   rules = association_rules(frequent_itemsets, metric="confidence", min_threshold=0.5)
   
   # iPhone15 -> 手机壳 (confidence=0.8, lift=2.5)
   ```

2. **推荐展示位置**：
   ```
   位置1：购物车下方
   "买了还买"：展示3-5个商品
   
   位置2：结算页
   "凑单优惠"：满减差额商品
   
   位置3：加购弹窗
   用户加购商品A → 弹窗推荐配件B
   ```

3. **推荐排序**：
   ```
   score = w1 × 关联度 + 
           w2 × 利润率 + 
           w3 × 库存充足度 +
           w4 × 用户个性化得分
   
   w1=0.4, w2=0.3, w3=0.2, w4=0.1
   ```

4. **AB测试**：
   ```
   测试维度：
   - A组：展示3个推荐
   - B组：展示5个推荐
   - C组：不展示推荐
   
   评估指标：
   - 推荐点击率
   - 推荐加购率
   - 客单价提升
   ```

**延伸思考**：
1. 推荐商品如何避免干扰用户（显得推销）？
2. 推荐算法如何冷启动（新商品无关联数据）？
3. 推荐效果如何评估和持续优化？

---

#### 📊 题目9：购物车的结算流程设计

**问题描述**：
用户点击"去结算"，进入结算页面，需要选择地址、优惠券、支付方式。如何设计结算流程？

**答案**：

**问题分析**：
结算流程的核心环节：
1. 确认商品（数量、价格）
2. 选择收货地址
3. 选择配送方式
4. 应用优惠（优惠券、积分）
5. 选择支付方式
6. 提交订单

**方案一：单页结算**

核心思想：
所有信息在一个页面完成。

页面布局：
```text
结算页：
┌─────────────────┐
│ 1. 收货地址      │
│ [北京市朝阳区...] │
├─────────────────┤
│ 2. 商品清单      │
│ iPhone 15 × 1   │
│ ¥7999           │
├─────────────────┤
│ 3. 配送方式      │
│ ○ 标准配送（免费）│
│ ○ 次日达（¥10）  │
├─────────────────┤
│ 4. 优惠         │
│ 优惠券：¥30     │
│ 积分抵扣：¥10   │
├─────────────────┤
│ 5. 支付方式      │
│ ○ 支付宝        │
│ ○ 微信支付      │
├─────────────────┤
│ 总计：¥7959     │
│ [提交订单]       │
└─────────────────┘
```

优点：
- 流程简洁
- 一目了然
- 减少跳转

缺点：
- 页面信息多
- 移动端显示困难

**方案二：分步结算（推荐）**

核心思想：
分多个步骤完成结算。

流程：
```text
步骤1：选择地址
→ 步骤2：确认商品和配送
→ 步骤3：选择优惠
→ 步骤4：支付
```

优点：
- 逻辑清晰
- 移动端友好
- 可保存中间状态

缺点：
- 步骤多
- 可能流失

**推荐方案**：
PC端使用**单页结算**，移动端使用**分步结算**。

实施要点：

1. **结算前校验**：
   ```java
   public CheckoutResult preCheckout(Long userId) {
     // 1. 获取购物车
     Cart cart = getCart(userId);
     
     // 2. 校验商品状态
     List<String> invalidItems = new ArrayList<>();
     for (CartItem item : cart.getItems()) {
       Product product = productService.getProduct(item.getSkuId());
       if (!product.isOnSale()) {
         invalidItems.add(item.getSkuId() + "：已下架");
       } else if (product.getStock() < item.getQuantity()) {
         invalidItems.add(item.getSkuId() + "：库存不足");
       }
     }
     
     if (!invalidItems.isEmpty()) {
       return CheckoutResult.fail("部分商品无法结算", invalidItems);
     }
     
     // 3. 计算价格
     PriceDetail price = calculatePrice(cart);
     
     // 4. 返回结算信息
     return CheckoutResult.success(cart, price);
   }
   ```

2. **地址选择**：
   ```
   展示用户地址列表：
   - 默认地址（置顶）
   - 最近使用地址
   - 其他地址
   
   新增地址：
   - 省市区三级联动
   - 详细地址输入
   - 联系人和电话
   - 设为默认地址
   ```

3. **优惠券选择**：
   ```
   展示可用优惠券：
   - 按优惠力度排序
   - 标注"最优"推荐
   - 显示使用门槛
   
   自动选择：
   - 默认选择优惠最大的券
   - 用户可手动切换
   
   不可用优惠券：
   - 置灰显示
   - 标注不可用原因（如"不满足使用条件"）
   ```

4. **价格实时计算**：
   ```javascript
   // 监听用户操作
   onChange = () => {
     // 防抖：用户停止操作500ms后计算
     clearTimeout(this.timer);
     this.timer = setTimeout(() => {
       this.calculatePrice();
     }, 500);
   };
   
   calculatePrice = async () => {
     const params = {
       items: this.state.cartItems,
       addressId: this.state.selectedAddress,
       couponId: this.state.selectedCoupon,
       usePoints: this.state.usePoints
     };
     
     const result = await API.post('/api/order/calculate-price', params);
     this.setState({ priceDetail: result });
   };
   ```

5. **订单确认信息**：
   ```
   最终确认页展示：
   - 收货人：张三 138****1234
   - 收货地址：北京市朝阳区xxx
   - 商品清单：iPhone 15 × 1
   - 配送方式：标准配送（预计3天送达）
   - 优惠明细：
     * 商品折扣：-¥100
     * 满减优惠：-¥30
     * 优惠券：-¥20
   - 实付金额：¥7849
   
   用户确认无误后点击"提交订单"
   ```

**延伸思考**：
1. 如何设计结算页的防重复提交？
2. 结算过程中价格变动如何处理？
3. 结算流程如何优化转化率？

---

#### 🔧 题目10：购物车的分享功能设计

**问题描述**：
用户想分享购物车给朋友（如"帮我看看这些商品怎么样"），如何设计购物车分享功能？

**答案**：

**问题分析**：
购物车分享的核心场景：
1. 征求意见（送礼选择）
2. 代购（帮朋友买）
3. 拼单（一起买更便宜）

**方案一：生成分享链接**

核心思想：
生成唯一URL，包含购物车商品信息。

实现：
```text
生成分享：
1. 用户点击"分享购物车"
2. 服务端生成分享ID
3. 保存分享内容到数据库/Redis
4. 返回分享链接

分享链接：
https://example.com/cart/share/abc123

接收分享：
1. 朋友点击链接
2. 展示分享者的购物车商品
3. 可一键导入到自己购物车
```

数据设计：
```sql
cart_share
├── share_id（唯一ID）
├── user_id（分享者）
├── cart_snapshot（JSON，购物车快照）
├── expire_at（过期时间）
├── view_count（查看次数）
└── created_at
```

优点：
- 实现简单
- 支持任意平台

缺点：
- 链接可能泄露
- 分享内容是快照（不会实时更新）

**方案二：生成二维码**

核心思想：
生成二维码，扫码查看购物车。

实现：
```text
生成二维码：
1. 生成分享链接（同方案一）
2. 将链接转为二维码
3. 展示二维码供分享

扫码查看：
1. 扫描二维码
2. 跳转到分享页面
3. 展示商品列表
```

优点：
- 线下分享方便
- 移动端友好

缺点：
- 仍是快照

**方案三：实时共享购物车（推荐）**

核心思想：
创建共享购物车，多人实时协同。

实现：
```text
创建共享：
1. 用户创建共享购物车
2. 生成共享ID和密码（可选）
3. 邀请朋友加入

实时同步：
- 任何人添加/删除商品
- 通过WebSocket实时同步给所有成员
- 显示"张三添加了iPhone 15"

共享购物车表：
shared_cart
├── shared_cart_id
├── creator_id
├── name（如"周末采购清单"）
├── password（可选）
├── members（成员列表）
├── items（商品列表）
└── created_at
```

优点：
- 实时协同
- 支持多人编辑
- 适合家庭、团队采购

缺点：
- 实现复杂
- 需要冲突处理

**推荐方案**：
采用**分享链接+实时共享**的组合。

实施要点：

1. **分享类型**：
   ```
   类型1：只读分享
   - 生成分享链接
   - 朋友只能查看，不能修改
   - 可一键导入到自己购物车
   
   类型2：协同编辑
   - 创建共享购物车
   - 邀请成员
   - 成员可添加/删除商品
   ```

2. **分享页面设计**：
   ```
   分享页头部：
   "张三分享了购物车给你"
   
   商品列表：
   [展示所有商品]
   
   操作按钮：
   - [全部加入我的购物车]
   - [选择部分加入]
   - [保存为我的收藏清单]
   ```

3. **隐私控制**：
   ```
   隐私选项：
   - 公开：任何人都可查看
   - 仅好友：需要登录且是好友
   - 密码保护：需要输入密码
   
   敏感信息隐藏：
   - 不显示价格（可选）
   - 不显示数量（可选）
   ```

4. **分享统计**：
   ```
   统计指标：
   - 分享次数
   - 查看人数
   - 转化人数（查看后购买）
   - 传播路径（A分享给B，B分享给C）
   ```

5. **场景化推荐**：
   ```
   场景1：送礼征询
   "想送女朋友礼物，帮我选一个"
   → 展示多个候选商品
   → 朋友投票或评论
   
   场景2：拼单
   "一起买，更便宜"
   → 共享购物车
   → 凑满减金额
   → 分摊运费
   ```

**延伸思考**：
1. 如何设计购物车的协同冲突解决（同时删除同一商品）？
2. 分享购物车如何防止恶意刷单？
3. 共享购物车如何拆单结算（各付各的）？

---

#### 💡 题目11：购物车的满减凑单提示

**问题描述**：
购物车总价¥180，有满¥200减¥30活动。如何设计智能凑单提示，引导用户加购？

**答案**：

**推荐方案**：

1. **差额计算**：
   ```
   当前金额：¥180
   满减门槛：¥200
   差额：¥20
   
   提示："再买¥20，立减¥30"
   ```

2. **智能商品推荐**：
   ```
   推荐商品筛选条件：
   - 价格在¥20-¥50之间（差额附近）
   - 与购物车商品相关（配件、同类目）
   - 库存充足
   - 高评分
   
   排序：
   - 优先推荐价格接近差额的
   - 优先推荐关联度高的
   ```

3. **视觉引导**：
   ```
   进度条展示：
   [████████░░] 90% (¥180/¥200)
   "再买¥20，立减¥30，相当于打8.5折"
   
   推荐商品卡片：
   ┌───────────┐
   │ 手机壳     │
   │ ¥29       │
   │ [加入购物车]│
   └───────────┘
   ```

4. **多档位满减**：
   ```
   满减档位：
   - 满¥100减¥10（已达成✓）
   - 满¥200减¥30（差¥20）
   - 满¥500减¥100（差¥320）
   
   提示优先显示最接近的下一档
   ```

**延伸思考**：
1. 凑单推荐如何避免过度营销（让用户反感）？
2. 多个满减活动同时存在时如何提示？

---

#### 📊 题目12：购物车的批量操作设计

**问题描述**：
用户购物车有50个商品，想批量删除、批量加入收藏。如何设计批量操作功能？

**答案**：

**推荐方案**：

1. **批量选择**：
   ```
   界面设计：
   [全选] 已选0件
   
   ☑ 商品A  ¥100
   ☑ 商品B  ¥200
   ☐ 商品C  ¥300
   
   批量操作：
   [删除选中] [加入收藏] [移除失效商品]
   ```

2. **批量接口**：
   ```java
   POST /api/cart/batch-delete
   {
     "skuIds": ["123", "456", "789"]
   }
   
   POST /api/cart/batch-move-to-favorite
   {
     "skuIds": ["123", "456"]
   }
   ```

3. **事务处理**：
   ```
   批量操作的事务性：
   - 部分成功部分失败如何处理？
   
   方案A：全量事务
   - 全部成功才提交
   - 任一失败全部回滚
   
   方案B：部分成功（推荐）
   - 成功的操作提交
   - 失败的返回错误信息
   - 前端展示"成功X件，失败Y件"
   ```

4. **性能优化**：
   ```
   批量删除50个商品：
   ❌ for循环50次DELETE
   ✅ 一次DELETE WHERE sku_id IN (...)
   
   批量更新库存：
   ❌ 50次UPDATE
   ✅ 批量UPDATE CASE WHEN
   ```

**延伸思考**：
1. 批量操作如何支持撤销（Undo）？
2. 批量操作的进度如何展示？

---

#### 🔧 题目13：购物车的收藏夹联动

**问题描述**：
购物车和收藏夹如何联动？商品从购物车移入收藏，或从收藏加入购物车。

**答案**：

**推荐方案**：

1. **数据模型**：
   ```sql
   favorite
   ├── favorite_id
   ├── user_id
   ├── sku_id
   ├── source（CART/BROWSE）
   ├── added_at
   └── ...
   ```

2. **互相转换**：
   ```
   购物车 → 收藏夹：
   1. 用户点击"移入收藏"
   2. 加入收藏夹
   3. 从购物车删除
   4. 提示"已移入收藏夹"
   
   收藏夹 → 购物车：
   1. 用户点击"加入购物车"
   2. 加入购物车
   3. 保留在收藏夹（不删除）
   ```

3. **降价提醒**：
   ```
   收藏商品降价：
   - 监控收藏商品价格
   - 降价时推送通知
   - 引导用户加购
   ```

**延伸思考**：
1. 收藏夹和购物车的区别是什么？
2. 如何设计收藏夹的分组功能？

---

#### 💡 题目14：购物车的历史记录

**问题描述**：
用户删除了购物车商品，想恢复。如何设计购物车的历史记录功能？

**答案**：

**推荐方案**：

1. **软删除**：
   ```sql
   shopping_cart
   ├── ...
   ├── deleted_at（软删除标记）
   └── deleted（是否删除）
   
   查询购物车：
   SELECT * FROM shopping_cart 
   WHERE user_id=? AND deleted=0
   
   查询历史：
   SELECT * FROM shopping_cart 
   WHERE user_id=? AND deleted=1
   ORDER BY deleted_at DESC
   ```

2. **恢复功能**：
   ```
   历史记录页面：
   最近删除：
   - 商品A（3天前删除）[恢复]
   - 商品B（7天前删除）[恢复]
   
   恢复操作：
   UPDATE shopping_cart 
   SET deleted=0, deleted_at=NULL
   WHERE cart_id=?
   ```

3. **自动清理**：
   ```
   定时任务：
   - 删除30天后的历史记录
   - 减少存储成本
   ```

**延伸思考**：
1. 购物车历史记录是否需要版本控制（记录每次修改）？
2. 如何设计购物车的快照功能（保存多个购物清单）？

---

#### 📊 题目15：购物车的AB测试设计

**问题描述**：
想测试新的购物车布局对转化率的影响。如何设计购物车的AB测试？

**答案**：

**推荐方案**：

1. **分流策略**：
   ```java
   public String getCartVersion(Long userId) {
     // 基于用户ID哈希分流
     int hash = userId.hashCode();
     if (hash % 2 == 0) {
       return "A"; // 对照组
     } else {
       return "B"; // 实验组
     }
   }
   ```

2. **实验设计**：
   ```
   对照组A（50%用户）：
   - 旧购物车布局
   
   实验组B（50%用户）：
   - 新购物车布局（优化后）
   
   评估指标：
   - 加购率
   - 结算率
   - 转化率
   - 客单价
   ```

3. **数据埋点**：
   ```javascript
   // 购物车页面浏览
   track('cart_view', {
     version: 'A', // 或 'B'
     cartItemCount: 5
   });
   
   // 点击结算
   track('cart_checkout_click', {
     version: 'A',
     cartTotal: 1000
   });
   
   // 完成下单
   track('order_created', {
     version: 'A',
     orderAmount: 1000
   });
   ```

4. **结果分析**：
   ```
   结果对比：
   | 指标 | A组 | B组 | 提升 |
   |------|-----|-----|------|
   | 结算率 | 60% | 65% | +8.3% |
   | 转化率 | 40% | 45% | +12.5% |
   | 客单价 | ¥800 | ¥850 | +6.25% |
   
   结论：B组效果更好，全量发布
   ```

**延伸思考**：
1. AB测试如何保证结果的统计显著性？
2. 多个AB测试同时进行时如何隔离影响？
3. 如何设计购物车的渐进式发布（灰度发布）？

---
