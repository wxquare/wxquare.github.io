---
title: 高频系统设计面试题速查手册
date: 2025-06-25
categories:
- 系统设计
tags:
- 面试
- 系统设计
- 架构
- 高并发
toc: true
---

<!-- toc -->

## 一、高并发与流量治理

### 1. 秒杀系统设计

**核心挑战**：瞬时流量巨大、库存超卖、恶意脚本。

**架构分层**：

| 层级 | 策略 |
|------|------|
| **客户端/CDN** | 静态资源缓存；按钮置灰+答题验证（削峰防刷） |
| **网关层** | 令牌桶/漏桶限流；黑名单拦截；设备指纹识别 |
| **服务层** | 库存预热到 Redis；MQ 异步扣减 DB 库存；非核心服务降级 |

**防超卖**（核心）：
- **Redis Lua 脚本**原子扣减：`if redis.call('get', key) > 0 then redis.call('decr', key) ...`
- **DB 乐观锁兜底**：`UPDATE stock SET num = num - 1 WHERE id = ? AND num > 0`

**防黄牛/脚本**：
- 滑块验证 / 人机识别
- 设备指纹 + 行为分析（点击间隔、轨迹）
- 实名认证 + 限购（身份证/手机号去重）

---

### 2. 分布式限流

**算法对比**：

| 算法 | 优点 | 缺点 |
|------|------|------|
| **固定窗口计数器** | 实现简单 | 临界突发：窗口交界处可能 2 倍流量 |
| **滑动窗口** | 解决临界突发 | 内存开销大（需存每个请求时间戳） |
| **漏桶** | 平滑输出 | 无法应对合理突发 |
| **令牌桶** | 允许突发 | 实现稍复杂 |

**分布式实现**：Redis + Lua（ZSet 滑动窗口 / Token Bucket）。

**动态限流**：基于 CPU、RT、错误率自适应调整阈值（Sentinel / Hystrix）。

---

### 3. 热点发现与隔离

**场景**：秒杀商品、热搜词、突发事件导致单个 Key 流量爆炸。

**方案**：
1. **探测**：实时统计 QPS，自动识别热点 Key。
2. **本地缓存**：热点 Key 复制到 JVM 内存（Caffeine），直接拦截。
3. **分散压力**：Key 后缀加随机值（`key_1 ~ key_N`），分散到多个 Redis 分片。
4. **隔离**：热点请求走独立线程池 + 独立缓存节点，不影响普通流量。

---

### 4. 熔断、降级、限流的区别

| 手段 | 目标 | 触发条件 |
|------|------|----------|
| **限流** | 控制入口流量 | QPS 超阈值 |
| **熔断** | 切断对下游的调用 | 下游错误率/超时率过高 |
| **降级** | 关闭非核心功能 | 系统负载高、人工/自动触发 |
| **兜底** | 给用户默认响应 | 降级后的补偿策略 |

**口诀**：限流防激增，熔断防雪崩，降级保核心，兜底提体验。

---

### 5. AI Agent 高并发架构

**挑战**：LLM 推理慢（秒级）、显存/线程池易耗尽、Token 成本高。

**优化策略**：
- **全异步化**：请求 → MQ → Agent 消费 → 结果存储 → 前端 SSE 推送。
- **流式输出 (SSE)**：Token 级返回，降低首屏感知延迟。
- **语义缓存 (Semantic Cache)**：向量相似度匹配高频问题，直接返回缓存。
- **成本优化**：模型蒸馏（小模型处理简单请求）；KV Cache 复用；请求批处理 (Batching)。
- **限流熔断**：严格限制 Agent 调用内部工具接口的频率，防止 AI 攻击内部系统。

---

## 二、海量数据与存储

### 1. 40亿数据去重（1GB 内存限制）

**方案对比**：

| 方案 | 空间 | 精确度 | 支持删除 |
|------|------|--------|----------|
| **Bitmap** | 40亿 ≈ 500MB | 精确 | 否 |
| **Bloom Filter** | 极小（几十 MB） | 有误判 | 否（Counting BF 可以，但空间 ×4） |
| **HyperLogLog** | 12KB | 误差 0.81% | 否 |

**最佳回答**：
- 40亿 QQ 号（unsigned int 范围 0~2^32）→ **Bitmap**，约 512MB 可精确去重。
- 若内存更紧张或允许少量误判 → **Bloom Filter**。
- 只需统计基数（不需要知道具体哪些重复）→ **HyperLogLog**。

---

### 2. 1亿玩家实时排行榜

**Redis ZSet 方案**：

```
ZADD rank 5000 "player_1"
ZREVRANGE rank 0 9  -- Top 10
ZRANK rank "player_1" -- 查排名
```

**陷阱**：ZSet 元素超过千万级 → 大 Key 阻塞主线程。

**解决方案（分桶 + 聚合）**：
1. 按玩家 ID 模 N 分到 N 个 ZSet：`rank_0, rank_1 ... rank_N`。
2. 每个桶取 Top K。
3. 应用层归并 N 个桶的 Top K，得到全局 Top K。

**分页优化**：
- `ZRANGE` 深分页性能差（O(logN + M)）。
- **游标分页**：记录上一页最后的 `(score, member_id)`，下一页从该位置继续查。
- **快照分页**：定时 dump 排行到 DB，前端查快照。

---

### 3. 海量数据排序（100GB 数据，8GB 内存）

1. **分块读入**：每次读入 8GB → 内存快排 → 写出有序文件。
2. **多路归并**：用小顶堆同时从 13 个有序文件中取最小值，输出全局有序文件。
3. **分布式**：MapReduce / Spark 分布式排序。

---

### 4. 10亿用户在线状态

**Bitmap**：1 bit 表示 1 个用户的在线/离线。1亿用户仅 12MB，10 亿用户约 120MB。

```
SETBIT online 123456 1   -- 用户123456上线
GETBIT online 123456     -- 查询是否在线
BITCOUNT online          -- 统计在线人数
```

---

## 三、典型业务场景设计

### 1. 订单超时自动取消

**场景**：下单 30 分钟未支付自动关闭。

| 方案 | 优点 | 缺点 |
|------|------|------|
| **定时任务扫表** | 实现简单 | 数据量大时效率低，延迟高 |
| **Redis 过期监听** | 简单 | 不可靠（不保证触发），**不推荐** |
| **Redis ZSet 轮询** | 精度高 | 需维护消费者 |
| **RocketMQ 延迟消息** | 可靠、可扩展 | 延迟级别有限 |
| **RabbitMQ TTL + DLX** | 灵活 | 架构复杂 |
| **时间轮 (Time Wheel)** | 高吞吐、内存高效 | 适合固定延迟场景 |

**最佳回答**：
- **短延迟 + 高吞吐**（如 <5 min）：时间轮。
- **长延迟 + 高可靠**（如 30 min 关单）：RocketMQ 延迟消息或 Redis ZSet。
- **千万级订单**：定时任务扫表无法胜任，必须用延迟队列。

---

### 2. 分布式 ID 生成器

| 方案 | 有序性 | 性能 | 问题 |
|------|--------|------|------|
| **UUID** | 无序 | 高 | 太长（128bit），B+ 树索引性能差 |
| **数据库号段** | 趋势递增 | 高 | 批量取号，DB 宕机有号段浪费 |
| **Snowflake** | 趋势递增 | 高 | 依赖时钟，回拨会重复 |
| **Redis INCR** | 递增 | 高 | 持久化风险，单点问题 |

**Snowflake 结构**：1 位符号 + 41 位时间戳（69 年）+ 10 位机器 ID + 12 位序列号（4096/ms）。

**容器化环境机器 ID 唯一**：
- Pod Name / IP 哈希取模。
- 启动时向 Etcd/ZooKeeper 注册获取唯一 ID。
- Redis INCR 动态分配 workerID。

---

### 3. 短链接系统

**生成策略**：
- **发号器 + Base62**：分布式 ID → 62 进制编码（a-z, A-Z, 0-9），6 位可表示 $62^6$ ≈ 568 亿。
- Hash（MD5/Murmur）取前 N 位：简单但需处理冲突。

**重定向选择**：
- **301 永久重定向**：浏览器缓存，无法统计点击数。
- **302 临时重定向**：每次经过服务端，可统计 UA、IP、Referer 等点击来源。

**点击统计**：302 重定向时解析 UA/IP/渠道 → 异步写入日志 → Flink 聚合 → ClickHouse 存储。

---

### 4. Feed 流系统

| 模式 | 读性能 | 写性能 | 适用场景 |
|------|--------|--------|----------|
| **推 (Write-fanout)** | 快 | 慢（写 N 个粉丝收件箱） | 普通用户 |
| **拉 (Read-fanout)** | 慢（聚合 N 个关注人） | 快 | 大 V |
| **推拉结合** | 均衡 | 均衡 | **业界主流** |

**推拉结合策略**：
- 活跃用户 / 普通博主：推模式。
- 大 V / 僵尸粉：拉模式。

**已读去重**：用户维度维护 RoaringBitmap，推送前 `if (!bitmap.contains(postId)) push()`。

---

### 5. 评论系统（B站/抖音盖楼）

**存储模型对比**：

| 模型 | 原理 | 优点 | 缺点 |
|------|------|------|------|
| **邻接表** | `id, parent_id` | 简单 | 查子树需递归，性能差 |
| **路径枚举** | `id, path="1/2/5"` | 前缀查询方便 | 路径长度受限 |
| **闭包表** | 单独表存所有祖先-后代 | 查询极快 | 写入量大 |

**业界主流（两层结构）**：
- **一级评论**：按热度/时间排序（Redis ZSet 或 DB 索引）。
- **二级回复**：扁平化存储。`parent_id` 指向一级评论，`reply_to_id` 指向被回复的人。不做无限嵌套。

**防灌水**：发言频率限制 → 敏感词过滤（AC 自动机）→ 举报+审核队列 → 新用户评论需审核（信任分体系）。

---

### 6. 红包算法

**二倍均值法**：`amount = random(1, remain / remain_count * 2)`，数学上保证期望恒定。

**高并发实现**：
- **预分配**：发红包时一次性算好所有金额，存入 Redis List。
- **抢红包**：`LPOP` 原子弹出，天然串行化。

---

### 7. 支付系统设计

**核心链路**：下单 → 锁库存 → 创建支付单 → 调第三方支付 → 异步回调 → 扣库存 → 发货。

**关键设计点**：
- **幂等**：`transaction_id` 唯一索引，重复回调不重复处理。
- **签名验签**：防篡改请求金额。
- **对账系统**：每日与第三方支付平台账单核对，发现差异报警。
- **事务消息**：RocketMQ 半消息保证扣库存与支付状态一致。

---

### 8. 库存系统深度设计

#### Q1：如何设计一个统一库存系统，支持电商、虚拟商品、本地生活等多品类？

**核心洞察**：不同品类库存差异巨大，需要抽象出通用模型。

**两个正交维度分类**：

```
维度一：谁管库存？
  - 自管理 (SelfManaged)：平台维护（Deal、OPV）
  - 供应商管理 (SupplierManaged)：第三方维护（酒店、机票）
  - 无限库存 (Unlimited)：无需管理（话费充值）

维度二：库存形态是什么？
  - 券码制 (CodeBased)：每个库存是唯一券码（电子券、Giftcard）
  - 数量制 (QuantityBased)：库存是一个数字（虚拟服务券）
  - 时间维度 (TimeBased)：按日期/时段管理（酒店、票务）
  - 组合型 (BundleBased)：多子项联动扣减（套餐）
```

**品类分类矩阵示例**：

| 品类 | 管理类型 | 单元类型 | 扣减时机 |
|------|----------|----------|----------|
| 电子券 | Self | Code | 下单 |
| 虚拟服务券 | Self | Quantity | 下单 |
| 酒店 | Supplier | Time | 支付 |
| 礼品卡(实时生成) | Supplier | Code | 支付 |

**架构设计（策略模式）**：

```
业务层 (Order Service)
    ↓
库存管理器 (InventoryManager)
    ↓
策略路由器 (根据 inventory_config 选策略)
    ↓
具体策略: SelfManagedStrategy / SupplierManagedStrategy / UnlimitedStrategy
    ↓
存储层: Redis (Hot) + MySQL (Cold) + Kafka (Async)
```

**核心优势**：
- ✅ 新品类接入只需写配置，无需改代码。
- ✅ 每个策略独立实现，复杂度隔离。

---

#### Q2：券码制库存（如电子券）如何实现高并发扣减？

**Redis 存储结构**：

```
Key:   inventory:code:pool:{itemID}:{skuID}:{batchID}
Type:  LIST
Value: [codeID_1, codeID_2, ...]

Key:   inventory:code:cursor:{itemID}:{skuID}:{batchID}
Value: "lastCodeID:lockCount"  (补货游标)

Key:   inventory:empty:{itemID}:{skuID}:{batchID}
TTL:   1h  (库存空标志，避免重复查 DB)
```

**出货流程**（核心）：

```
1. 检查库存空标志 → 命中则直接返回缺货
2. Redis LIST 原子出货 (Lua: LRANGE + LTRIM)
3. 如果库存不足 → 补货 (从 MySQL 查 3000 个可用券码 → RPUSH 到 Redis)
4. 更新 MySQL 券码状态: AVAILABLE → BOOKING
5. 同步更新 inventory 表: booking_stock += quantity
6. 发送 Kafka 事件异步记录日志
```

**Lua 脚本（原子性保证）**：

```lua
local result = redis.call('LRANGE', KEYS[1], 0, ARGV[1] - 1)
redis.call('LTRIM', KEYS[1], ARGV[1], -1)
return result
```

**关键设计**：
- **Lazy Loading**：按需补货，避免一次性加载全量券码到 Redis（节省内存）。
- **分布式锁**：补货时加锁，防止并发补货导致重复。
- **库存空标志**：DB 无库存后，1小时内拦截所有请求，避免反复查 DB。

---

#### Q3：数量制库存（如虚拟服务券）如何支持营销活动动态库存？

**Redis HASH 设计**：

```
Key:   inventory:qty:stock:{itemID}:{skuID}
Type:  HASH
Fields:
  "available"   : 10000       # 普通可售库存
  "booking"     : 50          # 预订中
  "issued"      : 5000        # 已售
  "{promotionID}": 500        # 营销活动独立库存（动态字段）
```

**预订 Lua 脚本（支持营销库存）**：

```lua
-- 1. 获取普通库存和营销库存
local available = tonumber(redis.call('HGET', key, 'available') or 0)
local promo = tonumber(redis.call('HGET', key, promotion_id) or 0)
local total = available + promo

-- 2. 检查库存
if book_num > total then return -1 end

-- 3. 优先扣营销库存，不足时扣普通库存
if promo >= book_num then
    redis.call('HINCRBY', key, promotion_id, -book_num)
else
    redis.call('HSET', key, promotion_id, 0)
    redis.call('HINCRBY', key, 'available', -(book_num - promo))
end

-- 4. 增加预订数
redis.call('HINCRBY', key, 'booking', book_num)
```

**亮点**：动态字段设计，无需提前建表，营销活动 ID 直接作为 HASH field。

---

#### Q4：供应商管理的库存（如酒店、机票）如何同步？

**三种同步策略**：

| 策略 | 适用场景 | 实时性 | 实现 |
|------|----------|--------|------|
| **实时查询** | 库存变化快（机票） | 高 | 每次请求调 API（30s 缓存） |
| **定时同步** | 变化中等（酒店） | 中 | 定时任务每 5 分钟拉取 |
| **Webhook 推送** | 供应商主动推送 | 高 | 接收推送更新本地缓存 |

**实时查询流程**：

```go
func CheckStock() {
    // 1. 查 Redis 缓存（30s TTL）
    if stock := redis.Get(cacheKey); stock != nil {
        return stock  // 命中缓存
    }
    
    // 2. 缓存未命中，调供应商 API
    stock := supplierAPI.QueryStock(itemID, date)
    
    // 3. 写入 Redis（30s）+ 异步写快照表（用于对账）
    redis.Set(cacheKey, stock, 30*time.Second)
    go saveSnapshot(itemID, stock, "api")
    
    return stock
}
```

**预订时**：调供应商预订接口 → 保存供应商订单号映射 → 更新本地 booking_stock。

---

#### Q5：如何保证 Redis 与 MySQL 库存数据一致性？

**双写策略**：

| 操作 | Redis | MySQL | 一致性 |
|------|-------|-------|--------|
| 预订 (Book) | 同步扣减（Lua） | Kafka 异步更新 | 最终一致 |
| 支付 (Sell) | 同步更新 | Kafka 异步更新 | 最终一致 |
| 营销锁定 (Lock) | 同步 | 同步（DB 事务） | 强一致 |

**核心原则**：
- **Redis 是热路径**：所有高频操作走 Redis（毫秒级响应）。
- **MySQL 是权威数据源**：故障恢复时以 MySQL 为准。
- **Kafka 异步持久化**：不阻塞主流程。

**定时对账（每小时）**：

```go
redisStock := getRedisAvailable(itemID)
mysqlStock := getMySQLAvailable(itemID)
diff := redisStock - mysqlStock

// 校验库存恒等式: total = available + booking + locked + sold
if mysqlTotal != mysqlAvailable + mysqlBooking + mysqlLocked + mysqlSold {
    alert("MySQL 数据不一致")
}

// Redis vs MySQL 差异
if abs(diff) > 100 || abs(diff) > mysqlStock*0.1 {
    alert("库存差异过大")
    syncRedisFromMySQL(itemID)  // 自动修复
}
```

---

#### Q6：Redis 宕机了，库存系统如何降级？

**降级方案**：

```
Redis 可用
  ↓
正常走 Redis（< 10ms）

Redis 不可用
  ↓
降级到 MySQL 直接操作（~100ms，性能下降但业务不中断）
  ↓
券码制: SELECT ... FOR UPDATE + UPDATE status
数量制: UPDATE available_stock = available_stock - ? WHERE available_stock >= ?
  ↓
记录降级日志，Redis 恢复后从 MySQL 全量同步
```

**注意**：
- 降级期间性能下降约 10 倍，需配合限流。
- MySQL 需提前规划好容量，支持降级时的流量。

---

#### Q7：Giftcard 实时生成卡密，供应商 API 超时怎么办？

**问题**：支付成功后调供应商 API 生成卡密，超时会导致用户等待。

**解决方案（异步生成 + 重试补偿）**：

```
支付成功
  ↓
1. 订单状态更新为"处理中"
  ↓
2. 发送到 MQ 异步队列 (giftcard.generate)
  ↓
3. 用户先看到"卡密生成中，稍后通知"

异步消费者：
  ↓
调用供应商 API 生成卡密
  ↓
失败？→ 指数退避重试 (1s, 2s, 4s)
  ↓
3 次仍失败？→ 人工补发 + 告警
  ↓
成功：保存卡密 → 推送通知用户
```

**卡密安全**：
- 存储时 AES-256 加密。
- 管理后台脱敏显示（`XXXX-XXXX-XXXX-1234`）。
- 所有访问记录审计日志。

---

#### Q8：时间维度库存（酒店/票务）与普通库存有什么不同？

**差异**：

| 维度 | 普通库存 | 时间维度库存 |
|------|----------|-------------|
| **库存粒度** | SKU 级别 | SKU + 日期 |
| **存储** | 单条记录 | 每个日期一条记录 |
| **查询** | 按 item_id + sku_id | 按 item_id + sku_id + date |
| **TTL** | 永久 | Redis 缓存 7 天 |

**Redis 设计**：

```
Key:   inventory:time:stock:{itemID}:{skuID}:{date}
Type:  HASH
Fields:
  "total"     : 100
  "available" : 80
  "booking"   : 15
  "sold"      : 5
TTL: 7天（历史日期自动过期，节省内存）
```

**挑战**：
- 酒店 1 个月有 30 条记录，查询"未来 7 天房态"需扫描 7 个 Key。
- 优化：批量 `MGET` + 并行查询。

---

#### Q9：如何支持"秒杀活动锁定 1000 件库存"？

**场景**：运营配置秒杀活动，需从总库存中锁定 1000 件，活动结束释放。

**Lua 脚本（营销锁定）**：

```lua
local available = tonumber(redis.call('HGET', key, 'available') or 0)
local promo_stock = tonumber(redis.call('HGET', key, promotion_id) or 0)

-- 检查库存
if lock_num > available then return -1 end

-- 从普通库存转移到营销库存
redis.call('HINCRBY', key, 'available', -lock_num)
redis.call('HSET', key, promotion_id, lock_num)
```

**数据库同步**：

```sql
UPDATE inventory 
SET available_stock = available_stock - ?,
    locked_stock = locked_stock + ?
WHERE item_id = ?
```

**活动结束解锁**：反向操作，营销库存 → 普通库存。

---

#### Q10：新接入一个品类"演唱会门票"，如何快速支持？

**三步接入**：

```go
// 1. 评估分类
// 演唱会门票 → 供应商管理 + 时间维度（按场次） + 支付成功扣减

// 2. 写配置
INSERT INTO inventory_config (item_id, management_type, unit_type, deduct_timing, supplier_id, sync_strategy)
VALUES (900001, 2, 3, 2, 700001, 2);

// 3. 调用统一接口（无需改代码）
inventoryManager.BookStock(ctx, &BookStockReq{
    ItemID:       900001,
    SKUID:        0,
    Quantity:     2,
    OrderID:      orderID,
    CalendarDate: "2025-08-15",  // 场次日期
})
```

**亮点**：配置驱动，零代码接入。

---

#### 面试追问点（高级）

**Q：为什么券码制库存不一次性加载全量到 Redis，而是按需补货？**
- **内存成本**：百万张券码全量加载需要几百 MB 内存，大部分可能永远用不到。
- **Lazy Loading**：按需补货，每次补 3000 个，节省内存。
- **补货游标**：记录上次补到哪个 codeID，避免重复查询。

**Q：库存对账发现 Redis 比 MySQL 多 500 个，怎么办？**
- **可能原因**：
  - Kafka 消息积压，MySQL 异步更新延迟。
  - Redis 补货后，MySQL 更新失败。
  - 存在未完成的预订订单（booking 状态）。
- **处理**：
  - 检查 Kafka 消费 lag。
  - 以 **MySQL 为准**，用 MySQL 数据覆盖 Redis（权威数据源原则）。
  - 人工核查异常订单。

**Q：多平台（Shopee、ShopeePay）如何独立统计库存？**
- Redis HASH 中增加 `booking_shopee`、`booking_shopeepay` 字段。
- 扣减时根据 `platform` 参数路由到不同字段。
- DB 也冗余存储 `booking_stock` 和 `spp_booking_stock`。

**Q：库存扣减后支付失败，如何归还库存？**
- **订单超时未支付**：延迟队列（30min）→ 触发 UnbookStock。
  - 券码制：code status BOOKING → AVAILABLE，RPUSH 回 Redis LIST。
  - 数量制：Redis `HINCRBY booking -1, HINCRBY available +1`。
- **支付明确失败**：立即同步释放。

---

---

## 四、分布式一致性与事务

### 1. 分布式事务

| 方案 | 一致性 | 性能 | 侵入性 | 适用场景 |
|------|--------|------|--------|----------|
| **2PC (XA)** | 强一致 | 差（阻塞） | 低 | 单体拆分初期 |
| **TCC** | 最终一致 | 中 | 高（需写 Try/Confirm/Cancel） | 金融转账 |
| **本地消息表** | 最终一致 | 高 | 中 | 通用场景 |
| **事务消息 (RocketMQ)** | 最终一致 | 高 | 低 | 电商下单 |
| **Saga** | 最终一致 | 高 | 中 | 长事务（跨多个服务） |

**TCC 追问：Confirm/Cancel 失败怎么办？**
- 必须保证幂等 + 重试。
- 设置最大重试次数，超过后记录悬挂事务，人工补偿。

---

### 2. Redis 与 MySQL 双写一致性

| 方案 | 流程 | 优缺点 |
|------|------|--------|
| **Cache Aside（推荐）** | 先更新 DB → 再删 Cache | 简单，极端并发下有短暂不一致 |
| **延迟双删** | 删 Cache → 更 DB → sleep → 再删 Cache | 减少脏读窗口，sleep 时间难定 |
| **Canal 订阅 Binlog** | 更 DB → Canal 监听 → 异步删/更新 Cache | 最终一致性好，架构复杂 |

**追问：先删缓存再更新 DB 有什么问题？**
- 删缓存后，另一个请求读到旧 DB 数据并回填缓存 → 脏数据长期存在。
- **正确顺序**：先更新 DB，再删缓存。即使删失败，下次读取时缓存 Miss 会加载最新数据。

---

### 3. 接口幂等性

**场景**：网络抖动重复提交、支付回调重复通知。

| 方案 | 实现 | 适用 |
|------|------|------|
| **数据库唯一索引** | `INSERT IGNORE` 或 `UNIQUE KEY` | 写操作去重 |
| **Token 机制** | 请求前获取 Token，提交时 Redis Lua 原子校验+删除 | 表单防重复提交 |
| **状态机** | `UPDATE SET status='PAID' WHERE id=? AND status='UNPAID'` | 状态流转类 |

**支付回调幂等**：支付平台传唯一 `transaction_id` → `INSERT IGNORE INTO payment_records (txn_id)`，已存在则跳过。

---

## 五、并发编程

### 1. 线程池设计

**线程数设置**：
- **CPU 密集型**：`N + 1`（N = CPU 核数）。
- **IO 密集型**：`N × (1 + Wait/Compute)` 或简化为 `2N`。

**量化估算**：
> 核心接口 RT = 500ms，目标 1 万 QPS。
> 单线程 QPS = 1000/500 = 2。
> 单机需线程数 = 10000 / 2 = 5000 → 不现实。
> → 需 **多台机器**：如 10 台，每台承担 1000 QPS，每台 500 线程。

**共享 vs 独享**：
- **独享**：核心业务（支付、下单），防止被边缘业务拖垮。
- **共享**：非核心业务共用 Common 线程池。

**监控**：暴露 `activeCount, queueSize, completedTaskCount`，队列 >80% 告警。

---

### 2. 异步并行优化

**场景**：接口串行调用 A（用户信息）、B（积分）、C（优惠券），总耗时 T = Ta + Tb + Tc。

**优化**：`CompletableFuture` (Java) / `errgroup` (Go) 并行调用，T = max(Ta, Tb, Tc)。

**风险与应对**：
- 并行度过高 → 下游瞬时压力倍增 → 配合限流和熔断。
- 部分失败 → 降级返回默认值（如积分返回 0）。
- 长尾超时 → `orTimeout(500ms)` 强制超时。

---

## 六、中间件选型与原理

### 1. 消息队列选型

| 维度 | Kafka | RocketMQ | RabbitMQ |
|------|-------|----------|----------|
| **吞吐量** | 极高（百万级 TPS） | 高（十万级） | 中（万级） |
| **延迟** | ms 级 | ms 级 | us 级 |
| **事务消息** | 不支持 | 支持 | 不支持 |
| **延迟队列** | 不原生 | 支持 | TTL + DLX |
| **适用场景** | 日志、大数据 | 金融、电商 | 中小规模、复杂路由 |

**为什么用 MQ？**
- **解耦**：上游不需要知道有几个下游消费者。
- **异步**：主流程快速返回，耗时操作后台处理。
- **削峰**：MQ 缓冲突发流量，消费者匀速消费，保护 DB。

**消息不丢失（三环节保障）**：
1. **生产者**：同步发送 + 失败重试。
2. **Broker**：同步刷盘 (`SYNC_FLUSH`) + 主从同步。
3. **消费者**：处理成功后再手动 ACK。

**消息重复**：消费端做幂等（唯一索引/状态机）。

**消息积压**：先扩容消费者 → 排查消费阻塞原因 → 必要时跳过非关键消息。

---

### 2. Redis 核心问题

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| **缓存穿透** | 查不存在的数据 | Bloom Filter / 缓存空值 |
| **缓存击穿** | 热点 Key 过期 | 互斥锁(Mutex) / 逻辑过期 |
| **缓存雪崩** | 大量 Key 同时过期 | 随机过期时间 / 多级缓存 |
| **Big Key** | 阻塞主线程 | 拆分 / `UNLINK` 异步删除 |

**Key 过期内存释放**：
- **惰性删除**：访问时才检查是否过期。
- **定期删除**：每秒随机抽取 20 个 Key 检查。
- **陷阱**：Redis 并非过期立即释放。从库不主动删，等主库发 DEL 命令 → 可能出现"主库内存正常，从库爆满"。

---

### 3. MySQL 分库分表

**拆分策略**：
- **垂直拆分**：按业务拆库（用户库、订单库），按字段拆表（大字段独立）。
- **水平拆分**：按 `Hash(UserID)` 或 `Range(Time)` 分散数据行。

**核心难题**：
- **分布式 ID**：Snowflake / 号段模式。
- **跨库 Join**：应用层组装，或宽表冗余。
- **非 Sharding Key 查询**：按 UserID 分片后，商家查订单（MerchantID）怎么办？→ **异构索引表**，另建一套按 MerchantID 分片的表（或同步到 ES）。
- **在线扩容**：双写迁移 → Canal 同步增量 → 灰度切读 → 切写。

**索引高频考点**：最左前缀、回表与覆盖索引、索引失效（函数/隐式转换/`!=`/`LIKE '%xx'`）、深分页优化（`WHERE id > last_id LIMIT 10`）。

---

### 4. Elasticsearch 架构

**日增 1TB 场景设计**：
- **冷热分离**：Hot（SSD，最近 3-7 天）→ Warm/Cold（HDD，历史数据）。
- **分片**：单分片 30-50GB，主分片创建后不可修改。
- **Rollover**：按时间/大小自动滚动创建新索引。

**查询优化**：
- 避免 `wildcard`，改用 `ngram` 分词器。
- 精确匹配用 `keyword` 类型。
- 深分页用 `search_after` 替代 `from + size`。

---

### 5. ClickHouse

- **适用**：日志分析、报表、OLAP 大屏、用户行为分析。
- **快的原因**：列式存储 + 数据有序 + 向量化执行。
- **不适合**：高并发单行查询、频繁 UPDATE。

---

## 七、安全

### 1. 密码存储

**问题**：为什么只能重置密码，不能找回原密码？

**回答**：密码存储的是 `bcrypt(password + salt)` 的**不可逆哈希值**。即使数据库泄露，攻击者也无法还原明文。

- **Salt（盐）**：随机字符串，防彩虹表。即使两人密码相同，Hash 也不同。
- **为什么用 bcrypt 而非 SHA256？** bcrypt 是**慢哈希**，故意设计得慢（可调 cost 参数），暴力破解成本极高。SHA256 太快，GPU 每秒可算数十亿次。

### 2. 常见攻防

| 攻击 | 防御 |
|------|------|
| **XSS** | 输出转义、CSP 头、HttpOnly Cookie |
| **CSRF** | CSRF Token、SameSite Cookie |
| **SQL 注入** | 预编译（`#{}` 而非 `${}`） |
| **重放攻击** | 签名 + 时间戳 + nonce + 设备指纹 |

### 3. HTTPS 握手

1. 服务端下发证书（含公钥）。
2. 客户端验证证书合法性。
3. 客户端生成随机对称密钥，用公钥加密传给服务端。
4. 后续通信使用对称加密。

**一句话**：非对称加密传密钥，对称加密传数据。

---

## 八、可观测性

### 1. 三大支柱

| 支柱 | 工具 | 核心 |
|------|------|------|
| **Logging** | Filebeat → Kafka → ES → Kibana | 结构化 JSON 日志，含 trace_id |
| **Metrics** | Prometheus + Grafana | 黄金信号：延迟、流量、错误率、饱和度 |
| **Tracing** | Jaeger / Zipkin | TraceID 串联全链路 |

### 2. "接口突然变慢"排查套路

1. **看链路 (Tracing)**：哪一跳耗时突增？
2. **看指标 (Metrics)**：DB CPU 飙升？MQ 积压？线程池满？
3. **看日志 (Logging)**：是否有异常堆栈？
4. **对比变更**：最近是否上线/扩容/配置变更？

**止血第一**：先回滚或切流量，再定位根因。

---

## 九、云原生与弹性架构

### 1. 服务网格 (Service Mesh)
- Istio + Envoy Sidecar：实现熔断、限流、灰度发布，**无代码侵入**。

### 2. 弹性伸缩
- **HPA**：基于 CPU/内存/QPS 自动扩缩容。
- **KEDA**：基于事件驱动（如 MQ 积压量）扩缩容。

### 3. Serverless
- 适用：突发流量、定时任务、Webhook。
- 限制：冷启动延迟（秒级）、执行时长上限。

---

## 十、计算机基础

### 1. 为什么 0.1 + 0.2 != 0.3？
- **IEEE 754**：二进制无法精确表示 0.1 和 0.2（无限循环小数），相加后精度丢失。
- **0.1 + 0.1 == 0.2**：两次相同的舍入误差在低位恰好抵消。
- **解决**：金额计算必须用 `Decimal` 类型（定点数）或转为整数（分）计算。

### 2. TCP 三次握手为什么不能两次？
- 两次握手无法防止**历史连接初始化**：旧的 SYN 包延迟到达，服务端误建连接，浪费资源。
- 三次握手确保双方都确认对方的收发能力正常。

---

## 十一、面试灵魂拷问

**Q：系统瓶颈在哪？怎么优化？**
先定位（DB？Redis？MQ？外部接口？）→ 再给方案（索引/分库/缓存/异步/并行/批量化）。

**Q：流量突增 10 倍怎么扛？**
限流（挡住超量）→ 扩容（水平加机器）→ 缓存（减少穿透）→ 异步（削峰填谷）→ 降级（保核心）。

**Q：线上故障排查流程？**
止血（回滚/切流）→ 看监控 → 看日志 → 看调用链 → 定位根因 → 修复 → 复盘。

**Q：分布式系统最难的是什么？**
网络不可靠、时钟不一致、节点随时会挂。核心矛盾是 **CAP 取舍**：金融选 CP（强一致），互联网选 AP（最终一致）。

**Q：方案有什么副作用？**
面试加分项——主动说出 trade-off。例如："虽然异步解耦了，但增加了链路追踪的复杂度和排查成本。"

---

## 参考

- [大厂面试真题 - Fox爱分享](https://juejin.cn/column/7566818477114490926)
