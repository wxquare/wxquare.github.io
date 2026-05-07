# 35.2.2 库存系统题库

## 35.2.2 库存系统（17题）

#### 🔧 题目0扩展：库存是怎么创建出来的？

**问题描述**：
很多库存系统只讲扣减、预占和释放，但真实业务里库存首先要被创建出来。有的 SKU 只是简单数量，有的需要券码池，有的是系统自己生成券码，有的还和门店、日期、时段有关。如何设计库存创建链路？

**答案**：

库存创建不是简单 `insert stock=100`，而是把商品中心的销售契约物化成库存域可扣减、可对账、可恢复的实例。推荐把库存创建做成独立命令和任务：

```text
ProductPublished / OpsImportSubmitted / SupplierSnapshotReady
  → InventoryCreateCommand
  → inventory_create_task
  → InventoryInitWorker
  → inventory_config / inventory_balance / inventory_code_pool_XX
  → Redis 热视图预热
  → InventoryReady / InventoryCreateFailed
```

创建命令要表达清楚：

```text
sku_id / offer_id
management_type：平台自管 / 供应商管理 / 无限库存
unit_type：数量 / 券码 / 时间 / 座位 / 组合
scope_type / scope_id：GLOBAL / STORE / CITY / WAREHOUSE / DATE / CHANNEL
batch_id：券码批次或货品批次
calendar_date / time_slot：日期或时段
initial_quantity：初始数量
code_source：IMPORTED / SYSTEM_GENERATED / SUPPLIER_GENERATED
idempotency_key：防重复创建
```

不同库存类型的创建方式不同：

| 类型 | 创建方式 | 关键点 |
|------|----------|--------|
| 简单数量库存 | 创建 `inventory_config` 和一行 `inventory_balance` | 写 `INIT/INBOUND` 流水，不能绕过账本直接改 stock |
| 门店数量库存 | 按 `sku_id + store_id` 创建库存行 | 门店上下线要支持锁定、迁移和审计 |
| 日期 / 时段库存 | 按 `sku_id + store_id + date + slot` 创建切片 | 高流量品类提前物化，长尾门店懒创建 |
| 导入券码库存 | 创建 `inventory_code_batch`，逐行写 `inventory_code_pool_XX` | 加密存储、哈希去重、Redis LIST 只预热 `code_id` |
| 系统生成券码 | 预生成批次，或按订单幂等生成后落库 | 返回给用户前必须先有 MySQL 权威行 |
| 供应商库存 | 创建供应商映射和本地快照 | 本地快照不是最终承诺，下单前需要强刷或预订 |

面试时可以强调三个原则：

1. **库存创建要任务化**：商品发布事务不应该同步创建海量券码或未来 365 天日历库存，否则发布链路会被库存写放大拖垮。
2. **库存创建要幂等**：同一个发布版本、导入批次或供应商快照重复投递时，不能重复入库或重复生成券码。
3. **库存创建要能解释来源**：每一次初始化、导入、补货、系统生码都要有任务、批次和账本流水，否则后续对账只能看到“库存变了”，无法解释为什么变。

对于券码制，最容易踩坑的是把 Redis 当成码池权威。正确做法是：

```text
导入或生成券码
  → 加密写入 inventory_code_pool_XX
  → status=AVAILABLE
  → Redis LIST 只灌入 code_id
  → 下单时弹出 code_id
  → MySQL CAS: AVAILABLE -> BOOKING
```

只有 MySQL 状态机更新成功，才算真正锁码成功。Redis 可以丢、可以重建，但不能成为唯一账本。

**延伸思考**：
1. 库存创建任务部分成功时，哪些数据可以继续保留，哪些必须回滚？
2. 系统生成券码如何防止被猜测和批量撞库？
3. 酒店或门店预约类库存，未来多久的日历切片应该提前物化？

---

#### 🔧 题目0扩展B：库存如何和商品供给运营平台、商品生命周期联动？

**问题描述**：
作为一个长期做电商平台的工程师，不能只讲库存扣减。商品从供给入口进入平台、经过审核发布、上线、下架、结束销售、售后核销，库存系统应该如何和商品供给运营平台以及商品生命周期联动？

**答案**：

核心判断是：**商品发布不等于商品可售，审核通过也不等于库存 ready**。

三层职责要分开：

| 层 | 负责什么 | 不能做什么 |
|----|----------|------------|
| 商品供给运营平台 | Draft、Staging、QC、Diff、风险审核、发布任务 | 直接写库存余额和券码池 |
| 商品生命周期 | `ONLINE/OFFLINE/ENDED/BANNED/ARCHIVED`、销售时间、发布版本 | 直接判断库存扣减是否成功 |
| 库存系统 | 库存配置、数量、码池、门店 / 日期切片、预占、账本 | 决定商品标题、类目、审核结果 |
| 营销系统 | 活动、券、补贴、预算、营销库存、优惠规则 | 直接改商品生命周期和库存账本 |
| 可售投影 | 合成商品、库存、价格、营销、履约、渠道、风控状态 | 不能替代库存权威账本 |

推荐链路：

```text
供给入口 / 运营编辑 / 供应商同步
  → Draft / Staging / QC / Diff
  → Publish Transaction
      写正式商品、publish_version、交易契约、Outbox
  → InventoryCreateCommand / InventoryAdjustCommand
  → 库存任务创建或调整库存实例
  → InventoryReady / InventoryChanged / InventoryFailed
  → Marketing Command / Eligibility Event
  → Availability Projector 合成可售状态
  → 搜索、缓存、详情页、运营看板刷新
```

生命周期和库存动作可以这样对应：

| 商品生命周期动作 | 库存系统动作 | 可售影响 |
|------------------|--------------|----------|
| Draft / Staging | 只做配置校验，不创建 C 端可用库存 | 不可见、不可售 |
| QC 通过 | 可以预创建库存任务，但不开放 Reserve | 仍不可售 |
| Publish 成功 | 消费 Outbox，创建 `inventory_config`、数量行、码池或时间切片 | 等待 InventoryReady |
| ONLINE 生效 | 若库存 ready 且未锁定，允许 Reserve | 可售 |
| 运营补货 | 走 `AdjustInventory/ImportCodeBatch/GenerateCodeBatch` | 可售水位变化 |
| OFFLINE 下架 | 停止新 Reserve，保留历史预占和已售记录 | 不可下单 |
| ENDED 销售结束 | 锁定剩余库存，过期未售券码，停止供应商 booking | 不可售，只保留售后 |
| BANNED 风控封禁 | 立即冻结新预占，必要时锁定码池 | 不可售，人工处理 |

成熟平台通常会单独做可售投影：

```text
Sellable =
  product_status == ONLINE
  AND now in sale_time_window
  AND inventory_status in READY/AVAILABLE
  AND price_status == READY
  AND fulfillment_status == READY
  AND channel_policy allows current channel
  AND risk_status not in BLOCKED
```

这样运营后台可以解释商品为什么不能卖：

```text
商品已发布，但不可售：
- 库存创建任务失败：券码文件有重复码
- 门店 1001 未配置营业时段
- 供应商 external_sku_id 映射缺失
- 搜索索引刷新失败，等待 Outbox 重试
```

要避免的反模式：

1. 供给后台直接改 `stock` 字段，绕过库存账本；
2. 商品 `ONLINE` 后默认可卖，忽略库存、价格、履约和搜索刷新状态；
3. 下架时删除库存行，导致历史订单、售后和券码核销不可追溯；
4. 供应商同步直接覆盖运营手工修复的库存策略；
5. 库存系统直接改商品生命周期，绕过审核和发布版本。

一句话总结：

> 供给平台治理变更，生命周期控制线上状态，库存系统提供可承诺资源，可售投影把这些状态合成用户能否下单。它们通过命令、事件、版本和幂等键协作，而不是互相直接改库。

**延伸思考**：
1. 商品已发布但库存初始化失败，是否允许展示“售罄”？
2. 运营手工补货和供应商同步库存冲突时，字段主导权怎么判定？
3. 下架后已有预占订单是否继续履约，谁来仲裁？

---

#### 📊 题目1：设计防止库存超卖的方案

**问题描述**：
电商大促时，热门商品库存100件，但短时间涌入1000个订单。如何设计库存扣减方案，防止超卖？

**答案**：

**问题分析**：
库存超卖的核心原因：
1. 并发扣减：多个请求同时扣减库存
2. 分布式环境：库存分散在多个节点
3. 缓存不一致：Redis和DB库存不同步
4. 库存回滚：订单取消后库存未释放

**方案一：数据库悲观锁**

核心思想：
使用数据库行锁保证原子性。

实现：
```sql
-- 查询并锁定
SELECT stock FROM inventory 
WHERE sku_id='123' 
FOR UPDATE;

-- 检查库存
if (stock >= quantity) {
  -- 扣减库存
  UPDATE inventory 
  SET stock = stock - quantity
  WHERE sku_id='123';
  
  COMMIT;
} else {
  ROLLBACK;
  throw new OutOfStockException();
}
```

优点：
- 强一致性
- 不会超卖
- 实现简单

缺点：
- 性能差（锁冲突）
- 并发度低
- 可能死锁

适用场景：
- 并发不高（QPS<1000）
- 小规模系统

**方案二：数据库乐观锁**

核心思想：
使用版本号，更新失败时重试。

实现：
```sql
-- 查询库存和版本号
SELECT stock, version FROM inventory WHERE sku_id='123';

-- 扣减库存（带版本号）
affected = UPDATE inventory 
SET stock = stock - quantity, version = version + 1
WHERE sku_id='123' AND version = {oldVersion} AND stock >= quantity;

if (affected == 0) {
  // 更新失败，重试
  retry();
}
```

优点：
- 无锁，性能好
- 不会超卖

缺点：
- 高并发时重试多
- 用户体验差（重试慢）

适用场景：
- 中等并发（QPS 1000-5000）
- 普通商品

**方案三：Redis原子操作（推荐）**

核心思想：
使用Redis的DECR原子操作扣减库存。

实现：
```lua
-- Lua脚本（原子执行）
local stock = redis.call('GET', KEYS[1])
if tonumber(stock) >= tonumber(ARGV[1]) then
  redis.call('DECRBY', KEYS[1], ARGV[1])
  return 1
else
  return 0
end

调用：
String key = "stock:sku:123";
Long result = redis.eval(luaScript, 
                         Collections.singletonList(key), 
                         Collections.singletonList(quantity));

if (result == 1) {
  // 扣减成功，异步同步到DB
  createOrder();
} else {
  // 库存不足
  throw new OutOfStockException();
}

异步同步DB：
定时任务（每10秒）：
1. 收集Redis库存变更
2. 批量更新MySQL
3. 对账纠偏
```

优点：
- 性能极高（Redis内存操作）
- 支持高并发（10万+ QPS）
- 不会超卖

缺点：
- Redis和DB最终一致性
- Redis故障风险
- 需要对账

**方案对比**：

| 方案 | 性能 | 一致性 | 并发度 | 适用场景 |
|------|------|--------|--------|----------|
| 悲观锁 | ★★☆☆☆ | 强一致 | ★★☆☆☆ | 低并发 |
| 乐观锁 | ★★★☆☆ | 强一致 | ★★★☆☆ | 中并发 |
| Redis原子 | ★★★★★ | 最终一致 | ★★★★★ | 高并发 |

**推荐方案**：
采用**Redis原子操作+异步同步DB**。

实施要点：

1. **双层库存设计**：
   ```
   Redis（实时库存）：
   - 用于扣减判断
   - 高性能
   - 可能丢失
   
   MySQL（权威库存）：
   - 定期同步Redis
   - 数据持久化
   - 对账基准
   ```

2. **库存同步**：
   ```
   Redis → MySQL：
   - 定时任务（每10秒）
   - 批量更新（减少DB压力）
   - 增量同步（只同步变更的SKU）
   
   MySQL → Redis：
   - 商品上架时初始化Redis
   - 运营调整库存时更新Redis
   - Redis故障恢复时从MySQL加载
   ```

3. **库存预热**：
   ```
   大促前：
   1. 识别热门商品（预测销量）
   2. 提前加载到Redis
   3. 设置永不过期
   4. 多副本（主从）
   ```

4. **降级方案**：
   ```
   Redis故障：
   - 降级到MySQL悲观锁
   - 限流（降低并发度）
   - 提示用户（商品火爆）
   ```

5. **监控告警**：
   ```
   指标：
   - Redis和MySQL库存差异
   - 库存扣减QPS
   - 库存不足次数
   - 超卖告警（库存为负）
   
   告警：
   - 库存差异 > 100
   - 超卖发生
   - Redis同步延迟 > 1分钟
   ```

**延伸思考**：
1. 秒杀场景如何进一步优化（如库存分段、令牌桶）？
2. Redis故障导致库存丢失如何恢复？
3. 如何处理订单取消后的库存回补？

---

#### 🔧 题目2：如何设计分布式库存系统？

**问题描述**：
电商平台有多个仓库（北京、上海、深圳），商品在不同仓库有不同库存。如何设计分布式库存系统？

**答案**：

**问题分析**：
分布式库存的核心挑战：
1. 库存分布：如何在多仓库间分配库存
2. 库存查询：如何快速查询总库存
3. 库存分配：用户下单时选择哪个仓库发货
4. 库存调拨：仓库间库存转移

**方案一：集中式库存**

核心思想：
所有仓库库存汇总到一个中心库存池。

设计：
```sql
inventory
├── sku_id
├── total_stock（总库存 = sum(所有仓库)）
├── reserved_stock（预占库存）
└── available_stock（可售库存）

warehouse_inventory（仓库库存明细）
├── sku_id
├── warehouse_id
├── stock
└── reserved_stock

库存扣减：
1. 扣减total_stock（集中判断）
2. 分配仓库（路由算法）
3. 扣减warehouse_inventory
```

优点：
- 逻辑简单
- 总库存查询快
- 不会出现"有总库存但无仓库可发"

缺点：
- 集中式瓶颈
- 仓库分配逻辑复杂

**方案二：分布式库存（独立核算）**

核心思想：
每个仓库独立管理库存，用户下单时路由到最优仓库。

设计：
```sql
warehouse_inventory
├── sku_id
├── warehouse_id
├── stock
├── reserved_stock
└── available_stock

用户下单流程：
1. 根据用户地址选择就近仓库
2. 查询该仓库库存
3. 如果有货，扣减该仓库库存
4. 如果无货，选择次近仓库
```

仓库路由策略：
```text
策略1：就近原则
- 北京用户 → 北京仓
- 上海用户 → 上海仓

策略2：库存优先
- 查询所有仓库库存
- 优先选择库存最多的仓库

策略3：成本优先
- 考虑运费、配送时效
- 选择性价比最高的仓库
```

优点：
- 分布式，无单点
- 性能好
- 仓库自治

缺点：
- 总库存需要聚合
- 仓库间库存不均
- 路由策略复杂

**方案三：虚拟库存池（推荐）**

核心思想：
前台展示虚拟总库存，后台按规则分配实际仓库。

设计：
```text
前台层（用户可见）：
inventory_view
├── sku_id
├── total_available（虚拟总库存）
    = sum(warehouse_inventory.available_stock)

后台层（实际库存）：
warehouse_inventory
├── sku_id
├── warehouse_id
├── physical_stock（实际库存）
├── reserved_stock（预占）
├── safety_stock（安全库存）
└── available_stock = physical_stock - reserved_stock - safety_stock

用户下单：
1. 检查虚拟总库存（快速判断）
2. 预占总库存（防止超卖）
3. 路由算法选择仓库
4. 扣减仓库库存
5. 如果仓库分配失败，尝试其他仓库
```

路由算法：
```text
优先级：
1. 就近仓库（配送快）
2. 库存充足仓库（避免缺货）
3. 成本低仓库（运费低）

加权打分：
score = w1 * distance_score + w2 * stock_score + w3 * cost_score
选择score最高的仓库
```

优点：
- 用户体验好（总库存可见）
- 灵活分配（后台优化）
- 支持复杂路由

缺点：
- 实现复杂
- 需要智能分配算法

**方案对比**：

| 维度 | 集中式 | 分布式 | 虚拟池 |
|------|--------|--------|--------|
| 用户体验 | ★★★★★ | ★★★☆☆ | ★★★★★ |
| 性能 | ★★★☆☆ | ★★★★★ | ★★★★☆ |
| 库存利用率 | ★★★★★ | ★★★☆☆ | ★★★★★ |
| 实施难度 | ★★★★☆ | ★★★★☆ | ★★★☆☆ |

**推荐方案**：
采用**虚拟库存池**。

实施要点：

1. **库存聚合**：
   ```
   实时聚合（Redis）：
   total_stock:sku:123 = 
     stock:warehouse:1:sku:123 + 
     stock:warehouse:2:sku:123 + 
     stock:warehouse:3:sku:123
   
   更新触发：
   - 仓库库存变更 → 更新总库存
   - 使用Redis Pipeline批量更新
   ```

2. **仓库选择算法**：
   ```java
   public Warehouse selectWarehouse(
     String userId, Address address, String skuId, int quantity
   ) {
     // 1. 筛选有货仓库
     List<Warehouse> candidates = warehouses.stream()
       .filter(w -> w.getStock(skuId) >= quantity)
       .collect(Collectors.toList());
     
     // 2. 计算每个仓库的得分
     return candidates.stream()
       .map(w -> new ScoredWarehouse(w, calculateScore(w, address)))
       .max(Comparator.comparing(ScoredWarehouse::getScore))
       .map(ScoredWarehouse::getWarehouse)
       .orElseThrow(OutOfStockException::new);
   }
   
   private double calculateScore(Warehouse w, Address addr) {
     double distanceScore = 1.0 / distance(w, addr);  // 距离越近越高
     double stockScore = w.getStock() / 100.0;         // 库存越多越高
     double costScore = 1.0 / w.getShippingCost();    // 成本越低越高
     
     return 0.5 * distanceScore + 0.3 * stockScore + 0.2 * costScore;
   }
   ```

3. **库存预占**：
   ```
   预占流程：
   1. 用户下单 → 预占库存（reserved_stock +quantity）
   2. 用户支付 → 确认扣减（stock -quantity, reserved_stock -quantity）
   3. 用户取消 → 释放库存（reserved_stock -quantity）
   
   超时释放：
   - 未支付订单30分钟后自动取消
   - 定时任务扫描超时预占，自动释放
   ```

4. **库存调拨**：
   ```
   场景：
   - 北京仓库存100，上海仓库存0
   - 上海用户下单，需要从北京调拨
   
   调拨流程：
   1. 创建调拨单
   2. 北京仓库：stock -10
   3. 运输中...
   4. 上海仓库：stock +10
   ```

5. **安全库存**：
   ```
   设计：
   available_stock = physical_stock - reserved_stock - safety_stock
   
   作用：
   - 预留库存应对盘点误差
   - 预留库存应对损坏、丢失
   - 建议：safety_stock = physical_stock * 5%
   ```

**延伸思考**：
1. 如何设计库存预警机制（库存不足提醒）？
2. 多仓库场景下如何最优化运费成本？
3. 如何处理商品跨仓拆单（一单多仓发货）？

---

#### 💡 题目3：大促场景下的库存预热和削峰方案

**问题描述**：
双11大促，预计订单量是平时的100倍。如何对库存系统进行预热和削峰，保证不超卖且性能可控？

**答案**：

**问题分析**：
大促库存的核心挑战：
1. 瞬时流量暴增（平时1000 QPS → 10万 QPS）
2. 热点商品集中（TOP 100商品占80%流量）
3. Redis/DB压力大
4. 需要防止库存击穿

**方案一：库存分段+令牌桶**

核心思想：
将库存分为多段，每段独立扣减，最后汇总。

设计：
```text
库存分段：
总库存10000件，分为10段：
segment_1: 1000件
segment_2: 1000件
...
segment_10: 1000件

Redis存储：
stock:sku:123:segment:1 = 1000
stock:sku:123:segment:2 = 1000
...

扣减逻辑：
1. 随机选择一个segment
2. 尝试扣减该segment库存
3. 如果成功，返回
4. 如果失败（库存不足），重试其他segment
5. 所有segment都不足，返回无货
```

优点：
- 降低Redis单key热点
- 提高并发度
- 不会超卖

缺点：
- 可能出现库存碎片（某段有货但其他段无货）
- 需要定期平衡segment

**方案二：本地库存+定期同步**

核心思想：
将库存预分配到应用服务器本地内存，减少Redis压力。

设计：
```text
初始化（大促前）：
1. 总库存10000件
2. 分配到100台服务器
3. 每台服务器本地内存：100件

扣减流程：
1. 用户请求到服务器A
2. 扣减服务器A本地库存（内存操作，极快）
3. 本地库存不足时，向Redis申请补货
4. Redis库存不足，返回无货

补货机制：
if (local_stock < 10) {
  申请补货100件
  Redis扣减100件
  local_stock += 100
}
```

优点：
- 性能极高（内存操作）
- 减轻Redis压力
- 支持极高并发

缺点：
- 服务器重启库存丢失（需要归还Redis）
- 库存分散，利用率低
- 需要补货机制

**方案三：队列削峰+异步扣减（推荐）**

核心思想：
请求进队列，消费端限速扣减，流量削峰。

设计：
```text
用户下单 
→ 请求入队（Kafka）
→ 库存扣减Worker（限速消费）
→ 扣减成功/失败
→ 通知用户（WebSocket/轮询）

限速策略：
1. 设置消费速率：5000 TPS
2. 队列堆积：允许100万消息堆积
3. 超时处理：队列中超过5分钟的请求自动取消

用户体验：
1. 提交订单立即返回"排队中"
2. 显示排队位置（前面还有XXX人）
3. 扣减成功后通知用户
4. 扣减失败（无货）通知用户
```

优点：
- 削峰效果好
- 库存系统压力可控
- 用户体验可接受（秒杀场景）

缺点：
- 用户等待时间长
- 需要排队机制
- 实现复杂

**方案对比**：

| 方案 | 性能 | 削峰效果 | 用户体验 | 实施难度 |
|------|------|---------|---------|---------|
| 库存分段 | ★★★★☆ | ★★★☆☆ | ★★★★★ | ★★★☆☆ |
| 本地库存 | ★★★★★ | ★★★★★ | ★★★★★ | ★★★☆☆ |
| 队列削峰 | ★★★☆☆ | ★★★★★ | ★★★☆☆ | ★★☆☆☆ |

**推荐方案**：
采用**库存分段+本地库存**的组合。

实施要点：

1. **库存预热**：
   ```
   大促前3天：
   1. 识别热销商品（TOP 1000）
   2. Redis预加载：
      - 库存数据
      - 商品信息
      - 价格信息
   3. 本地缓存预加载
   4. 压测验证
   ```

2. **分段策略**：
   ```
   分段数量 = max(库存数量 / 100, 服务器数量)
   
   示例：库存10000，服务器100台
   → 分段数 = max(10000/100, 100) = 100段
   → 每段100件
   
   优点：
   - 降低单key热度
   - 并发度=分段数
   ```

3. **本地库存管理**：
   ```java
   public class LocalInventory {
     private final ConcurrentHashMap<String, AtomicInteger> localStock;
     
     public boolean tryDeduct(String skuId, int quantity) {
       AtomicInteger stock = localStock.computeIfAbsent(
         skuId, k -> new AtomicInteger(0)
       );
       
       // 乐观尝试扣减
       int current = stock.get();
       if (current >= quantity) {
         if (stock.compareAndSet(current, current - quantity)) {
           return true;
         }
       }
       
       // 本地库存不足，申请补货
       if (requestRecharge(skuId, 100)) {
         return tryDeduct(skuId, quantity); // 重试
       }
       
       return false;
     }
   }
   ```

4. **监控大盘**：
   ```
   实时监控：
   - 总库存水位
   - 扣减QPS
   - 成功率
   - Redis热key
   - 本地库存分布
   
   告警：
   - 库存水位 < 20%
   - 扣减失败率 > 5%
   - Redis单key QPS > 10万
   ```

**延伸思考**：
1. 秒杀开始前如何预热（避免冷启动）？
2. 大促结束后如何回收本地库存？
3. 如何应对恶意刷单占用库存？

---

#### 📊 题目4：库存预占与释放的设计

**问题描述**：
用户加入购物车或进入结算页时，需要预占库存，防止其他用户抢走。但如果用户不支付，需要释放库存。如何设计库存预占机制？

**答案**：

**问题分析**：
库存预占的核心挑战：
1. 预占时机：什么时候预占（加购、结算、下单）
2. 预占时长：预占多久（太短影响支付，太长占用库存）
3. 超时释放：如何自动释放超时预占
4. 并发安全：多个请求同时预占

**方案一：下单时预占**

核心思想：
用户下单时才预占库存，加购和结算不预占。

设计：
```text
加购物车：不预占库存
进入结算页：不预占库存
提交订单：预占库存
  → 成功：进入支付流程
  → 失败：提示库存不足

预占超时：30分钟
支付成功：确认扣减
订单取消：释放库存
```

优点：
- 库存利用率高
- 实现简单

缺点：
- 用户结算时可能无货（体验差）
- 无法保证结算页的库存

适用场景：
- 普通商品
- 库存充足

**方案二：结算时预占（推荐）**

核心思想：
用户进入结算页时预占库存，支付成功确认，超时释放。

设计：
```sql
inventory
├── sku_id
├── total_stock（总库存）
├── reserved_stock（预占库存）
├── sold_stock（已售库存）
└── available_stock = total_stock - reserved_stock - sold_stock

预占记录表：
reservation
├── reservation_id
├── sku_id
├── order_id
├── quantity
├── status（RESERVED/CONFIRMED/RELEASED）
├── expire_at（过期时间）
└── created_at
```

流程：
```text
1. 进入结算页：
   BEGIN TRANSACTION
     UPDATE inventory 
     SET reserved_stock = reserved_stock + quantity
     WHERE sku_id=? AND available_stock >= quantity;
     
     INSERT INTO reservation (sku_id, order_id, quantity, expire_at)
     VALUES (?, ?, ?, NOW() + INTERVAL 15 MINUTE);
   COMMIT

2. 支付成功：
   UPDATE inventory 
   SET reserved_stock = reserved_stock - quantity,
       sold_stock = sold_stock + quantity
   WHERE sku_id=?;
   
   UPDATE reservation SET status='CONFIRMED' WHERE reservation_id=?;

3. 超时释放（定时任务）：
   SELECT * FROM reservation 
   WHERE status='RESERVED' AND expire_at < NOW();
   
   For each expired:
     UPDATE inventory 
     SET reserved_stock = reserved_stock - quantity;
     
     UPDATE reservation SET status='RELEASED';
```

优点：
- 保证结算页库存
- 用户体验好
- 防止超卖

缺点：
- 预占时间内库存被占用
- 需要定时任务释放

**方案三：分级预占**

核心思想：
根据用户等级和商品类型，设置不同的预占时长。

设计：
```text
预占时长策略：
VIP用户：30分钟
普通用户：15分钟
新用户：10分钟

热门商品：10分钟（快速流转）
普通商品：15分钟
冷门商品：30分钟（不占用热门商品库存）

动态调整：
if (available_stock < 10% * total_stock) {
  // 库存紧张，缩短预占时间
  expire_time = 5分钟
} else {
  expire_time = 15分钟
}
```

优点：
- 差异化服务
- 库存利用率高
- 灵活调整

缺点：
- 规则复杂
- 实现成本高

**方案对比**：

| 方案 | 用户体验 | 库存利用率 | 超卖风险 | 实施难度 |
|------|---------|-----------|---------|---------|
| 下单预占 | ★★★☆☆ | ★★★★★ | ★★★☆☆ | ★★★★★ |
| 结算预占 | ★★★★★ | ★★★★☆ | ★★★★★ | ★★★☆☆ |
| 分级预占 | ★★★★★ | ★★★★★ | ★★★★★ | ★★☆☆☆ |

**推荐方案**：
采用**结算时预占**。

实施要点：

1. **预占时长设置**：
   ```
   考虑因素：
   - 支付流程耗时（通常2-3分钟）
   - 用户犹豫时间（5-10分钟）
   - 库存周转率（紧俏商品缩短）
   
   建议：
   - 默认15分钟
   - 库存<10%时缩短到5分钟
   - VIP用户延长到30分钟
   ```

2. **预占幂等性**：
   ```
   使用order_id作为幂等键：
   INSERT INTO reservation (reservation_id, order_id, ...)
   ON DUPLICATE KEY UPDATE updated_at=NOW();
   
   防止重复预占：
   - 同一订单多次预占，使用相同reservation记录
   - 延长expire_at即可
   ```

3. **超时释放优化**：
   ```
   方案A：定时任务扫描
   - 每分钟扫描一次
   - 查询expire_at < NOW()
   - 批量释放
   
   方案B：延迟队列（推荐）
   - 预占时发送延迟消息（延迟15分钟）
   - 消息到期时检查状态
   - 如果未支付，释放库存
   
   优点：精确释放，无需轮询
   ```

4. **库存保护**：
   ```
   最大预占比例：
   - 允许预占库存 <= total_stock * 90%
   - 保留10%库存应对预占释放后的瞬时需求
   
   预占限流：
   - 单用户最多预占5个订单
   - 单商品最多被预占total_stock * 80%
   ```

**延伸思考**：
1. 用户在结算页停留很久不支付，如何处理？
2. 预占释放后其他用户如何得知库存恢复？
3. 如何设计库存预占的监控指标？

---

#### 🔧 题目5：如何设计库存的分级管理（前台可售vs仓库实际）？

**问题描述**：
仓库实际库存100件，但前台可售库存只有80件（预留20件应对售后、损耗）。如何设计库存的分级管理？

**答案**：

**问题分析**：
库存分级的核心挑战：
1. 不同层级库存含义不同
2. 层级间库存同步
3. 安全库存设置
4. 库存占用追踪

**方案一：单一库存（简化版）**

核心思想：
只维护一个库存字段，不区分层级。

设计：
```sql
inventory
├── sku_id
├── stock（唯一库存字段）
└── reserved_stock（预占）
```

优点：
- 实现简单
- 无需同步

缺点：
- 无法预留安全库存
- 无法应对损耗

**方案二：多级库存（推荐）**

核心思想：
区分物理库存、可售库存、预占库存、已售库存。

设计：
```sql
inventory
├── sku_id
├── physical_stock（物理库存，仓库实际数量）
├── reserved_stock（预占库存，待支付订单）
├── sold_stock（已售库存，已支付待发货）
├── safety_stock（安全库存，预留）
├── available_stock（可售库存，计算得出）
    = physical_stock - reserved_stock - sold_stock - safety_stock
└── version

库存关系：
physical_stock（100）
  - safety_stock（10，安全库存）
  - sold_stock（20，已售待发货）
  - reserved_stock（15，预占待支付）
  = available_stock（55，可售）
```

库存流转：
```text
用户下单：
available_stock -10, reserved_stock +10

用户支付：
reserved_stock -10, sold_stock +10

商品发货：
sold_stock -10, physical_stock -10

订单取消：
reserved_stock -10, available_stock +10

售后退货：
physical_stock +10, available_stock +10
```

优点：
- 库存含义清晰
- 支持安全库存
- 易于追踪

缺点：
- 字段多，维护成本高
- 同步逻辑复杂

**方案三：占用日志模式**

核心思想：
只维护物理库存，所有占用记录在日志表。

设计：
```sql
inventory
├── sku_id
└── physical_stock

inventory_occupation（库存占用日志）
├── occupation_id
├── sku_id
├── occupation_type（RESERVED/SOLD/SAFETY）
├── quantity
├── reference_id（order_id/warehouse_id）
├── status（ACTIVE/RELEASED）
└── expire_at

可售库存计算：
available_stock = physical_stock - sum(active_occupations)
```

优点：
- 灵活，支持多种占用类型
- 可追溯所有占用历史
- 易于扩展

缺点：
- 查询需要聚合计算
- 性能较差

**方案对比**：

| 维度 | 单一库存 | 多级库存 | 占用日志 |
|------|---------|---------|----------|
| 清晰度 | ★★☆☆☆ | ★★★★★ | ★★★★☆ |
| 性能 | ★★★★★ | ★★★★☆ | ★★★☆☆ |
| 灵活性 | ★★☆☆☆ | ★★★☆☆ | ★★★★★ |
| 实施难度 | ★★★★★ | ★★★☆☆ | ★★☆☆☆ |

**推荐方案**：
采用**多级库存**。

实施要点：

1. **安全库存设置**：
   ```
   策略：
   - 标准：safety_stock = 5% * physical_stock
   - 易损商品：safety_stock = 10% * physical_stock
   - 高价商品：safety_stock = 2% * physical_stock
   
   动态调整：
   - 根据历史损耗率调整
   - 旺季增加，淡季减少
   ```

2. **库存同步检查**：
   ```
   不变量检查：
   physical_stock = 
     available_stock + 
     reserved_stock + 
     sold_stock + 
     safety_stock
   
   定期对账：
   如果不等式不成立，说明库存有问题
   ```

3. **库存调整接口**：
   ```
   运营调整物理库存：
   adjustPhysicalStock(skuId, delta, reason)
   
   自动调整安全库存：
   adjustSafetyStock(skuId, percentage)
   ```

4. **库存报表**：
   ```
   库存健康度：
   - 库存周转率 = 销量 / 平均库存
   - 滞销率 = 30天未售商品数 / 总商品数
   - 缺货率 = 用户下单失败次数 / 总下单次数
   ```

**延伸思考**：
1. 如何设计库存盘点功能（盘点期间库存锁定）？
2. 安全库存不足时如何处理？
3. 已售库存发货后如何核减？

---

#### 💡 题目6：库存扣减失败的补偿机制

**问题描述**：
在订单创建流程中，扣减库存可能失败（并发冲突、网络超时、服务故障）。如何设计补偿机制，保证数据一致性？

**答案**：

**问题分析**：
库存扣减失败的核心场景：
1. 网络超时：不知道是否扣减成功
2. 服务故障：库存服务不可用
3. 并发冲突：乐观锁更新失败
4. 数据不一致：订单已创建但库存未扣减

**方案一：同步重试**

核心思想：
扣减失败时立即重试，最多重试3次。

实现：
```java
public void deductInventory(String skuId, int quantity) {
  int maxRetries = 3;
  for (int i = 0; i < maxRetries; i++) {
    try {
      inventoryService.deduct(skuId, quantity);
      return; // 成功
    } catch (ConcurrentModificationException e) {
      if (i == maxRetries - 1) {
        throw e; // 最后一次重试失败，抛出异常
      }
      Thread.sleep(100 * (i + 1)); // 指数退避
    }
  }
}
```

优点：
- 实现简单
- 实时性好

缺点：
- 重试占用用户等待时间
- 多次重试可能仍失败
- 影响用户体验

**方案二：异步补偿**

核心思想：
扣减失败时订单标记为待处理，后台异步补偿。

流程：
```text
1. 订单创建：
   if (扣减库存失败) {
     订单状态 = PENDING_INVENTORY
     记录补偿任务
   }

2. 补偿Worker：
   定时扫描PENDING_INVENTORY订单
   重试扣减库存
   成功 → 更新订单状态CONFIRMED
   失败 → 继续重试或人工介入

3. 补偿任务表：
   compensation_task
   ├── task_id
   ├── order_id
   ├── task_type（DEDUCT_INVENTORY）
   ├── payload（JSON）
   ├── status（PENDING/SUCCESS/FAILED）
   ├── retry_count
   └── next_retry_at
```

优点：
- 不阻塞用户
- 支持多次重试
- 可人工介入

缺点：
- 最终一致性
- 用户可能看到"处理中"状态
- 实现复杂

**方案三：补偿+对账（推荐）**

核心思想：
结合同步重试和异步补偿，再加对账兜底。

流程：
```text
1. 扣减库存：
   try {
     inventoryService.deduct(skuId, quantity);
   } catch (Exception e) {
     // 同步重试1次
     retry once
     if (still failed) {
       // 记录补偿任务
       compensationService.record(orderId, "DEDUCT_INVENTORY");
     }
   }

2. 补偿Worker（每分钟）：
   查询补偿任务
   重试执行
   成功 → 标记完成
   失败 → retry_count +1

3. 对账任务（每小时）：
   查询已支付订单
   检查库存是否已扣减
   未扣减 → 创建补偿任务

4. 人工兜底：
   - retry_count > 5次仍失败
   - 转人工处理
   - 排查根本原因
```

优点：
- 多层保障
- 可靠性高
- 覆盖各种异常

缺点：
- 实现最复杂

**方案对比**：

| 方案 | 实时性 | 可靠性 | 用户体验 | 实施难度 |
|------|--------|--------|---------|---------|
| 同步重试 | ★★★★★ | ★★★☆☆ | ★★★☆☆ | ★★★★★ |
| 异步补偿 | ★★★☆☆ | ★★★★☆ | ★★★★☆ | ★★★☆☆ |
| 补偿+对账 | ★★★☆☆ | ★★★★★ | ★★★★☆ | ★★☆☆☆ |

**推荐方案**：
采用**补偿+对账**。

实施要点：

1. **幂等性保证**：
   ```java
   public void deductInventory(DeductRequest req) {
     // 使用orderId作为幂等键
     if (isAlreadyDeducted(req.getOrderId())) {
       return; // 已扣减，直接返回
     }
     
     // 执行扣减
     doDeduct(req);
     
     // 记录已扣减
     markDeducted(req.getOrderId());
   }
   ```

2. **补偿任务重试策略**：
   ```
   指数退避：
   第1次：立即重试
   第2次：1分钟后
   第3次：5分钟后
   第4次：15分钟后
   第5次：1小时后
   
   超过5次 → 转人工
   ```

3. **补偿任务优先级**：
   ```
   P0：已支付订单（优先处理）
   P1：待支付订单
   P2：其他
   ```

4. **对账规则**：
   ```
   检查项：
   1. 订单状态=PAID → 库存必须已扣减
   2. 订单金额 = 商品价格 × 数量
   3. 库存不能为负数
   
   差异处理：
   - 自动补偿（低风险）
   - 人工介入（高风险）
   ```

**延伸思考**：
1. 如果补偿重试多次仍失败，如何处理？
2. 补偿过程中订单状态如何展示给用户？
3. 如何监控补偿任务的执行情况？

---

#### 📊 题目7：设计库存盘点系统

**问题描述**：
仓库需要定期盘点库存，核对系统库存和实际库存是否一致。如何设计库存盘点系统？

**答案**：

**问题分析**：
库存盘点的核心挑战：
1. 盘点期间如何处理库存变更
2. 盘点差异如何调整
3. 大规模商品盘点效率
4. 盘点结果审核

**方案一：冻结盘点**

核心思想：
盘点期间冻结库存，禁止出入库。

流程：
```text
1. 创建盘点任务：
   - 选择仓库
   - 选择商品范围（全部/部分）
   - 冻结库存（禁止扣减和补货）

2. 仓库人员盘点：
   - 扫描商品条码
   - 录入实际数量

3. 生成盘点报告：
   - 系统库存 vs 实际库存
   - 差异清单

4. 审核调整：
   - 审核员确认差异
   - 调整系统库存
   - 解冻库存
```

优点：
- 准确性高
- 实现简单

缺点：
- 盘点期间影响业务
- 效率低
- 用户体验差

**方案二：动态盘点（推荐）**

核心思想：
盘点期间不冻结，记录盘点时间段的出入库，最后计算差异。

流程：
```text
1. 开始盘点：
   记录盘点开始时间T1
   快照当前系统库存S1

2. 盘点期间：
   正常出入库
   记录所有库存变更日志

3. 结束盘点：
   记录盘点结束时间T2
   记录实际库存数量P

4. 计算差异：
   期间出库：delta_out = sum(T1到T2的出库)
   期间入库：delta_in = sum(T1到T2的入库)
   
   理论库存：S2 = S1 - delta_out + delta_in
   实际库存：P
   差异：diff = P - S2

5. 调整库存：
   if (diff != 0) {
     inventory.physical_stock += diff
     记录盘点调整日志
   }
```

优点：
- 不影响业务
- 准确性高
- 可并行盘点

缺点：
- 计算复杂
- 需要完整的出入库日志

**方案三：循环盘点**

核心思想：
不是一次性盘点所有商品，而是每天盘点一部分。

流程：
```text
将商品分为ABC类：
A类（高价值，20%）：每月盘点
B类（中价值，30%）：每季度盘点
C类（低价值，50%）：每年盘点

每日盘点：
1. 系统自动生成今日盘点任务
2. 仓库人员按任务盘点
3. 异常差异及时调整
4. 正常差异汇总报告
```

优点：
- 分散盘点，效率高
- 重点商品关注度高
- 不影响业务

缺点：
- 需要分类管理
- 全盘点周期长

**方案对比**：

| 方案 | 对业务影响 | 准确性 | 效率 | 适用场景 |
|------|-----------|--------|------|----------|
| 冻结盘点 | ★★☆☆☆ | ★★★★★ | ★★☆☆☆ | 小仓库 |
| 动态盘点 | ★★★★★ | ★★★★★ | ★★★★☆ | 大仓库 |
| 循环盘点 | ★★★★★ | ★★★★☆ | ★★★★★ | 商品多 |

**推荐方案**：
采用**动态盘点+循环盘点**的组合。

实施要点：

1. **盘点任务生成**：
   ```
   创建盘点单：
   inventory_check
   ├── check_id
   ├── warehouse_id
   ├── check_type（FULL/PARTIAL/CYCLE）
   ├── status（PENDING/CHECKING/COMPLETED）
   ├── start_snapshot_id（开始时库存快照）
   ├── start_at
   ├── end_at
   └── operator
   
   盘点明细：
   check_detail
   ├── check_id
   ├── sku_id
   ├── system_stock（系统库存）
   ├── actual_stock（实际库存）
   ├── diff（差异）
   ├── reason（差异原因）
   └── adjusted（是否已调整）
   ```

2. **盘点APP设计**：
   ```
   功能：
   - 扫码盘点（扫条码自动录入）
   - 语音录入（解放双手）
   - 拍照记录（有问题的商品拍照）
   - 离线模式（网络不好时）
   
   优化：
   - 按货架号排序（减少走动）
   - 实时同步（避免数据丢失）
   ```

3. **差异分析**：
   ```
   差异原因分类：
   - 损耗（DAMAGE）：商品破损
   - 丢失（LOSS）：商品丢失
   - 错发（WRONG_SHIP）：发错货
   - 漏记（MISSING_RECORD）：出入库漏记
   - 系统bug（SYSTEM_ERROR）
   
   自动调整规则：
   - diff < 5% → 自动调整
   - diff >= 5% → 需要审核
   - diff > 20% → 必须复盘（可能系统bug）
   ```

4. **盘点报告**：
   ```
   报告内容：
   - 盘点汇总：总商品数、差异数、差异金额
   - 差异TOP 10：差异最大的商品
   - 差异原因分布：损耗X件、丢失Y件
   - 仓库对比：各仓库差异率
   ```

**延伸思考**：
1. 如何设计盘点的权限控制（防止作弊）？
2. 盘点差异过大时如何追责？
3. 如何设计移动盘点的离线模式？

---

#### 🔧 题目8：如何处理库存的并发更新？

**问题描述**：
多个订单同时扣减同一商品库存，如何处理并发冲突，保证库存不超卖？

**答案**：

**问题分析**：
并发更新的核心场景：
1. 秒杀场景：1万人抢100件商品
2. 正常场景：多个用户同时下单
3. 分布式场景：多个服务器同时扣减

**方案一：数据库行锁（FOR UPDATE）**

实现：
```sql
BEGIN TRANSACTION;

-- 锁定行
SELECT stock FROM inventory 
WHERE sku_id='123' FOR UPDATE;

-- 检查库存
if (stock >= quantity) {
  UPDATE inventory SET stock = stock - quantity;
  COMMIT;
} else {
  ROLLBACK;
}
```

优点：
- 强一致性
- 不会超卖

缺点：
- 锁冲突，性能差
- 并发度低
- 长事务风险

吞吐量：约1000 TPS

**方案二：乐观锁（CAS）**

实现：
```sql
-- 查询当前库存
SELECT stock, version FROM inventory WHERE sku_id='123';

-- 尝试更新（CAS）
affected = UPDATE inventory 
SET stock = stock - quantity, version = version + 1
WHERE sku_id='123' 
  AND version = oldVersion 
  AND stock >= quantity;

if (affected == 0) {
  // 更新失败，重试
  retry with exponential backoff
}
```

优点：
- 无锁，性能好
- 并发度高

缺点：
- 高并发时重试多，成功率低
- 可能饿死（一直重试失败）

吞吐量：约5000-10000 TPS

**方案三：Redis+Lua脚本（推荐）**

实现：
```lua
-- Lua脚本（Redis原子执行）
local stock_key = KEYS[1]
local quantity = tonumber(ARGV[1])

local stock = tonumber(redis.call('GET', stock_key) or "0")

if stock >= quantity then
  redis.call('DECRBY', stock_key, quantity)
  return 1
else
  return 0
end
```

调用：
```java
String key = "inventory:sku:123";
Long result = redis.eval(luaScript, 
                         Arrays.asList(key), 
                         Arrays.asList(String.valueOf(quantity)));

if (result == 1) {
  // 扣减成功
  createOrder();
  // 异步同步到MySQL
  asyncSyncToMySQL(skuId, -quantity);
} else {
  throw new OutOfStockException();
}
```

优点：
- 性能极高（内存操作）
- 原子性（Lua脚本）
- 支持极高并发

缺点：
- Redis和MySQL最终一致
- Redis故障风险
- 需要对账机制

吞吐量：约10万+ TPS

**方案对比**：

| 方案 | TPS | 超卖风险 | 一致性 | 复杂度 |
|------|-----|---------|--------|--------|
| 行锁 | 1K | 无 | 强一致 | ★★★★☆ |
| 乐观锁 | 5-10K | 无 | 强一致 | ★★★☆☆ |
| Redis+Lua | 100K+ | 无 | 最终一致 | ★★★☆☆ |

**推荐方案**：
根据场景选择：
- **普通商品**：乐观锁（MySQL）
- **秒杀商品**：Redis+Lua
- **低并发**：悲观锁

实施要点：

1. **Redis高可用**：
   ```
   - Redis主从+哨兵
   - 双机房部署
   - 持久化：AOF every second
   ```

2. **库存同步**：
   ```
   Redis → MySQL：
   - 定时任务（每10秒）
   - 批量更新（减少DB压力）
   - 对账纠偏（每小时）
   ```

3. **降级方案**：
   ```
   Redis故障 → 降级到MySQL乐观锁
   MySQL故障 → 停止扣减，返回系统繁忙
   ```

4. **监控**：
   ```
   - Redis和MySQL库存差异
   - 扣减成功率
   - 扣减耗时P99
   - 并发冲突次数
   ```

**延伸思考**：
1. 如何设计秒杀的库存扣减（更极端的高并发）？
2. 分库分表场景下如何扣减库存？
3. Redis和MySQL数据不一致如何恢复？

---

#### 💡 题目9：虚拟库存vs实物库存的差异

**问题描述**：
实物商品有物理库存限制，虚拟商品（如充值卡、游戏币）可以无限生成。两者在库存设计上有什么差异？

**答案**：

**问题分析**：
虚拟库存的核心特点：
1. 可按需生成（理论无限）
2. 实际受限于供应商配额
3. 卡密池管理（有卡密才能售卖）
4. 即时发货（无需物流）

**方案一：无限库存模式**

核心思想：
虚拟商品库存设为无限大，不限制购买。

设计：
```sql
product
├── product_id
├── product_type（PHYSICAL/VIRTUAL）
└── unlimited_stock（布尔，是否无限库存）

扣减逻辑：
if (product.unlimitedStock) {
  // 虚拟商品，不扣减库存
  return true;
} else {
  // 实物商品，正常扣减
  return deductStock(skuId, quantity);
}
```

优点：
- 实现最简单
- 用户体验好（永不缺货）

缺点：
- 不适合卡密类商品（卡密有限）
- 无法控制销售节奏
- 可能超过供应商配额

适用场景：
- 可按需生成的虚拟商品（游戏币、积分）

**方案二：卡密 / 券码池模式（推荐）**

核心思想：
维护卡密 / 券码池，库存=可用卡密或券码数量。

设计：
```sql
virtual_product
├── product_id
├── supplier_id（供应商）
├── card_type（充值卡类型）
└── face_value（面值）

card_pool（卡密 / 券码池）
├── card_id
├── product_id
├── card_no（卡号）
├── card_pwd（密码，加密存储）
├── status（AVAILABLE/BOOKING/SOLD/LOCKED/EXPIRED/INVALID）
├── booked_at
├── reservation_id / order_id
├── sold_at
└── sold_order_id

库存计算：
available_stock = COUNT(*) WHERE status='AVAILABLE'
reserved_stock = COUNT(*) WHERE status='BOOKING'
```

生产级设计里，不建议只用一张简单 `card_pool` 表，更推荐把库存域的券码池收敛成 `inventory_code_pool_XX` 分表：

```text
inventory_code_pool_XX
├── code_id（全局唯一，Redis LIST 只缓存这个 ID）
├── batch_id / inventory_key / sku_id（批次、库存项、SKU）
├── code_cipher（加密后的券码或卡密）
├── code_hash（去重和排查，不保存明文）
├── status（AVAILABLE/BOOKING/SOLD/LOCKED/EXPIRED/INVALID）
├── reservation_id / order_id / user_id
├── booked_at / sold_at / expire_at
└── version（CAS 与幂等控制）
```

面试时要特别强调：**Redis LIST 不是权威库存，只是 `code_id` 热队列**。下单时可以先从 Redis 弹出 `code_id`，但必须再执行 MySQL CAS：

```sql
UPDATE inventory_code_pool_XX
SET status='BOOKING',
    reservation_id=?,
    order_id=?,
    booked_at=NOW(),
    version=version+1
WHERE code_id=? AND status='AVAILABLE';
```

只有这条更新成功，才算真正锁码成功。支付成功后 `BOOKING -> SOLD`；订单取消或超时后 `BOOKING -> AVAILABLE`，再通过 Outbox 或补偿任务把 `code_id` 回填到 Redis。已经交付或核销链路可见的 `SOLD` 码，不应直接回到可售池，退款要走售后和履约规则。

这个设计的价值是：
- 防止 Redis 丢数据导致无法追溯；
- 避免 LIST 存明文券码造成泄漏；
- 用状态机防止重复发码和并发超卖；
- Redis 故障后可以从 MySQL `AVAILABLE` 状态重建热队列；
- 对账时能按订单、批次、供应商和码状态逐行追踪。

库存流转：
```text
用户下单：
1. SELECT * FROM card_pool 
   WHERE product_id=? AND status='AVAILABLE' 
   LIMIT 1 FOR UPDATE;
   
2. UPDATE card_pool 
   SET status='BOOKING', booked_at=NOW(), order_id=?
   WHERE card_id=?;

用户支付：
UPDATE card_pool 
SET status='SOLD', sold_at=NOW(), sold_order_id=?
WHERE card_id=? AND status='BOOKING';

订单取消：
UPDATE card_pool 
SET status='AVAILABLE', order_id=NULL
WHERE card_id=? AND status='BOOKING';
```

优点：
- 库存真实（有卡密才能售）
- 支持卡密管理
- 防止超卖

缺点：
- 需要维护卡密池
- 卡密补货

**方案三：配额模式**

核心思想：
供应商给定配额，按配额售卖。

设计：
```sql
supplier_quota（供应商配额）
├── supplier_id
├── product_id
├── total_quota（总配额）
├── used_quota（已使用）
├── remaining_quota（剩余）
└── validity_period（有效期）

扣减逻辑：
1. 检查剩余配额
2. 扣减配额
3. 订单成功后，向供应商申请实际卡密
4. 发货给用户
```

优点：
- 无需提前准备卡密
- 按需申请
- 库存灵活

缺点：
- 实时性依赖供应商
- 供应商故障风险

**方案对比**：

| 方案 | 准确性 | 供应商依赖 | 实施难度 | 适用场景 |
|------|--------|-----------|---------|----------|
| 无限库存 | ★★☆☆☆ | ★★★★★ | ★★★★★ | 可生成虚拟品 |
| 卡密池 | ★★★★★ | ★★★☆☆ | ★★★☆☆ | 充值卡、券码 |
| 配额模式 | ★★★★☆ | ★★☆☆☆ | ★★★☆☆ | 供应商直连 |

**推荐方案**：
根据虚拟商品类型选择：
- **可生成**（游戏币、积分）：无限库存
- **卡密类**（充值卡、激活码）：卡密池
- **供应商直连**（机票、酒店）：配额模式

实施要点：

1. **卡密安全**：
   ```
   - 卡密加密存储（AES-256）
   - 卡密传输加密（HTTPS）
   - 卡密脱敏展示（**** **** **** 1234）
   - 限制查询频率（防止批量获取）
   ```

2. **卡密补货**：
   ```
   补货触发：
   - 可用卡密 < 安全阈值（如1000张）
   - 自动告警
   
   补货方式：
   - 供应商API自动拉取
   - 或人工Excel导入
   ```

3. **卡密有效期**：
   ```
   过期处理：
   - 定时任务扫描过期卡密
   - 状态更新为INVALID
   - 库存减少（不可售）
   - 向供应商申请补卡
   ```

4. **虚拟发货**：
   ```
   自动发货：
   - 支付成功 → 立即分配卡密
   - 推送给用户（短信/App）
   - 订单状态 → COMPLETED
   
   发货耗时：< 30秒
   ```

**延伸思考**：
1. 卡密被盗用如何防范？
2. 虚拟商品是否需要支持退款？
3. 供应商配额不足时如何处理？

---

#### 📊 题目10：多仓库场景下的库存分配策略

**问题描述**：
电商平台有5个仓库（华北、华东、华南、西南、西北），用户下单时如何选择仓库发货？请设计库存分配策略。

**答案**：

**问题分析**：
仓库选择的核心考量：
1. 配送时效：就近仓库配送快
2. 运费成本：距离影响运费
3. 库存充足度：优先选择库存多的仓库
4. 仓库负载：避免单仓库压力过大

**方案一：就近原则**

核心思想：
根据用户地址，选择最近的仓库。

设计：
```text
仓库覆盖范围：
- 北京仓：北京、天津、河北
- 上海仓：上海、江苏、浙江
- 深圳仓：广东、广西、福建
- 成都仓：四川、重庆、云南
- 西安仓：陕西、甘肃、新疆

路由逻辑：
1. 解析用户收货地址的省份
2. 查找覆盖该省份的仓库
3. 检查库存
4. 有货 → 该仓库发货
5. 无货 → 选择次近仓库
```

优点：
- 配送快
- 用户体验好
- 运费低

缺点：
- 库存可能不均衡
- 跨区发货增加成本

**方案二：智能调度（推荐）**

核心思想：
综合考虑配送时效、库存、成本，动态选择最优仓库。

设计：
```text
评分模型：
score = w1 * distance_score + 
        w2 * stock_score + 
        w3 * cost_score +
        w4 * load_score

各项得分计算：
1. distance_score（距离）：
   = 1.0 / (distance_km + 100)
   距离越近分越高

2. stock_score（库存）：
   = warehouse_stock / max_stock
   库存越多分越高

3. cost_score（成本）：
   = 1.0 / shipping_cost
   运费越低分越高

4. load_score（负载）：
   = 1.0 - (current_orders / capacity)
   当前订单越少分越高

权重设置：
- 普通商品：w1=0.5, w2=0.3, w3=0.1, w4=0.1
- 秒杀商品：w1=0.3, w2=0.5, w3=0.1, w4=0.1（库存优先）
- 大件商品：w1=0.4, w2=0.2, w3=0.3, w4=0.1（成本优先）
```

优点：
- 全局最优
- 灵活可配置
- 支持多种策略

缺点：
- 计算复杂
- 需要实时数据（各仓库负载）

**方案三：库存均衡策略**

核心思想：
主动调配库存，保持各仓库库存均衡。

设计：
```text
库存均衡算法：
1. 计算各仓库库存偏离度
   deviation = (warehouse_stock - avg_stock) / avg_stock

2. 如果偏离度 > 30%，触发调拨
   从库存多的仓库调拨到库存少的仓库

3. 调拨优先级：
   - 距离近优先
   - 库存差距大优先

调拨执行：
1. 创建调拨单
2. 源仓库出库
3. 物流运输
4. 目标仓库入库
```

优点：
- 库存均衡，利用率高
- 减少缺货
- 优化全局

缺点：
- 调拨成本高
- 调拨周期长（天级）
- 需要预测算法

**方案对比**：

| 方案 | 配送时效 | 成本 | 库存利用率 | 复杂度 |
|------|---------|------|-----------|--------|
| 就近原则 | ★★★★★ | ★★★★☆ | ★★★☆☆ | ★★★★★ |
| 智能调度 | ★★★★☆ | ★★★★★ | ★★★★☆ | ★★★☆☆ |
| 均衡策略 | ★★★☆☆ | ★★★☆☆ | ★★★★★ | ★★☆☆☆ |

**推荐方案**：
采用**智能调度**。

实施要点：

1. **仓库路由服务**：
   ```java
   public interface WarehouseRouter {
     // 选择单个仓库
     Warehouse route(Order order);
     
     // 多商品拆单（可能分多仓库发货）
     Map<Warehouse, List<OrderItem>> routeMulti(Order order);
   }
   
   实现：
   public Warehouse route(Order order) {
     List<Warehouse> candidates = getCandidateWarehouses(order);
     
     return candidates.stream()
       .filter(w -> hasStock(w, order))
       .map(w -> new ScoredWarehouse(w, calculateScore(w, order)))
       .max(Comparator.comparing(ScoredWarehouse::getScore))
       .map(ScoredWarehouse::getWarehouse)
       .orElseThrow(OutOfStockException::new);
   }
   ```

2. **拆单策略**：
   ```
   场景：用户购买商品A、B、C
   - 商品A：北京仓有货
   - 商品B：上海仓有货
   - 商品C：两个仓库都有货
   
   策略1：优先合单
   - 查找能满足所有商品的仓库
   - 减少拆单，降低运费
   
   策略2：就近发货
   - 每个商品从最近仓库发货
   - 可能拆多单，但配送快
   
   策略3：混合
   - 大件商品就近发货
   - 小件商品合单发货
   ```

3. **库存预测**：
   ```
   预测模型：
   - 输入：历史销量、季节、促销活动
   - 输出：未来7天各仓库销量预测
   
   预分配：
   - 根据预测提前调拨库存
   - 避免大促时调拨来不及
   ```

4. **负载均衡**：
   ```
   仓库容量管理：
   - 每个仓库设置日处理能力（如1万单/天）
   - 接近容量时降低选择权重
   - 超过容量时停止分配
   
   动态调整：
   - 实时监控各仓库订单量
   - 动态调整路由权重
   ```

**延伸思考**：
1. 如何处理跨仓拆单的运费计算？
2. 用户能否指定发货仓库？
3. 仓库之间如何协同（库存调拨、应急支援）？

---

#### 🔧 题目11：如何设计库存安全水位和补货机制？

**问题描述**：
电商系统需要设置库存安全水位，当库存低于安全水位时自动触发补货。如何设计这套机制？

**答案**：

**问题分析**：
库存安全水位的核心要素：
1. 安全水位如何设置（太高占用资金，太低容易缺货）
2. 补货时机和数量
3. 补货周期（供应商交付时间）
4. 多SKU的补货优先级

**方案一：固定安全水位**

核心思想：
为每个SKU设置固定的安全库存数量。

设计：
```sql
inventory
├── sku_id
├── stock
├── safety_stock（安全库存，人工设置）
└── reorder_point（补货点 = safety_stock + lead_time_demand）

补货触发：
if (stock <= reorder_point) {
  创建补货单
  补货数量 = (max_stock - current_stock)
}
```

优点：
- 实现简单
- 易于理解

缺点：
- 不够灵活
- 无法应对销量波动
- 需要人工调整

**方案二：动态安全水位（推荐）**

核心思想：
根据销量预测动态调整安全水位。

设计：
```text
销量预测：
avg_daily_sales = sum(last_30_days_sales) / 30

前置时间：
lead_time = 供应商交付周期（如7天）

安全库存：
safety_stock = avg_daily_sales * lead_time * safety_factor

其中：
- safety_factor = 1.5（安全系数，应对波动）
- 旺季调高到2.0
- 淡季调低到1.2

补货点：
reorder_point = safety_stock + lead_time * avg_daily_sales

补货数量（EOQ经济订货批量）：
order_quantity = sqrt((2 * annual_demand * order_cost) / holding_cost)
```

优点：
- 动态调整
- 科学合理
- 节省成本

缺点：
- 依赖销量预测准确性
- 计算复杂

**方案三：ABC分类管理**

核心思想：
将商品分为ABC类，采用不同的补货策略。

分类标准：
```text
A类商品（20%商品，80%销售额）：
- 高价值，严格管理
- 低安全库存（减少资金占用）
- 频繁补货（每周）
- 精准预测

B类商品（30%商品，15%销售额）：
- 中等价值，常规管理
- 中等安全库存
- 定期补货（每月）
- 简单预测

C类商品（50%商品，5%销售额）：
- 低价值，粗放管理
- 高安全库存（减少缺货）
- 批量补货（每季度）
- 不预测
```

优点：
- 差异化管理
- 资源聚焦
- 效率高

缺点：
- 需要定期重分类
- ABC边界商品难处理

**方案对比**：

| 方案 | 准确性 | 资金占用 | 维护成本 | 适用规模 |
|------|--------|---------|---------|----------|
| 固定水位 | ★★★☆☆ | ★★☆☆☆ | ★★★★★ | 小规模 |
| 动态水位 | ★★★★★ | ★★★★☆ | ★★★☆☆ | 大规模 |
| ABC管理 | ★★★★☆ | ★★★★★ | ★★☆☆☆ | 超大规模 |

**推荐方案**：
采用**动态水位+ABC分类**。

实施要点：

1. **销量预测模型**：
   ```
   简单移动平均：
   avg_sales = sum(last_N_days) / N
   
   加权移动平均：
   avg_sales = sum(sales[i] * weight[i])
   权重：最近的销量权重更高
   
   指数平滑：
   forecast[t] = α * actual[t-1] + (1-α) * forecast[t-1]
   α = 0.3（平滑系数）
   
   时间序列模型（高级）：
   - ARIMA
   - Prophet（Facebook开源）
   - 考虑季节性、趋势、促销影响
   ```

2. **补货决策表**：
   ```sql
   replenishment_rule
   ├── sku_id
   ├── category（ABC分类）
   ├── safety_stock
   ├── reorder_point
   ├── lead_time（补货周期）
   ├── order_quantity（建议补货量）
   ├── max_stock（最大库存）
   └── updated_at
   ```

3. **自动补货流程**：
   ```
   定时任务（每天凌晨）：
   1. 扫描所有SKU库存
   2. 识别低于补货点的SKU
   3. 生成补货建议单
   4. 采购员审核
   5. 自动下单给供应商（或人工）
   
   补货单：
   purchase_order
   ├── po_id
   ├── supplier_id
   ├── sku_id
   ├── quantity
   ├── expected_delivery_date
   ├── status（PENDING/CONFIRMED/SHIPPED/RECEIVED）
   └── created_at
   ```

4. **补货优先级**：
   ```
   优先级计算：
   priority = w1 * shortage_ratio + 
              w2 * sales_velocity + 
              w3 * profit_margin
   
   shortage_ratio = (reorder_point - current_stock) / reorder_point
   sales_velocity = daily_sales
   profit_margin = (price - cost) / price
   
   优先补货：
   - 严重缺货（shortage_ratio > 0.5）
   - 高销量
   - 高利润
   ```

5. **监控告警**：
   ```
   告警条件：
   - 库存 < 安全库存 → 缺货预警
   - 库存 > 最大库存 * 1.5 → 积压告警
   - 补货单超期未到货 → 交付延迟告警
   
   报表：
   - 缺货率（SKU缺货天数 / 总天数）
   - 库存周转率（销量 / 平均库存）
   - 补货及时率（按时到货 / 总补货单）
   ```

**延伸思考**：
1. 促销活动前如何调整补货策略？
2. 供应商交付不稳定如何应对？
3. 新品如何设置安全库存（无历史数据）？

---

#### 💡 题目12：库存快照在订单中的应用

**问题描述**：
订单下单时需要记录当时的库存状态，用于售后和数据分析。如何设计库存快照机制？

**答案**：

**问题分析**：
库存快照的核心目的：
1. 售后分析（为何超卖、缺货）
2. 数据审计（库存变更追溯）
3. 报表统计（某时刻库存状态）
4. 性能要求（不能影响下单）

**方案一：订单表冗余库存字段**

核心思想：
在订单表记录下单时的库存数量。

设计：
```sql
order_item
├── order_id
├── sku_id
├── quantity（购买数量）
├── stock_at_order（下单时库存，快照）
└── ...
```

优点：
- 实现最简单
- 查询方便

缺点：
- 快照信息有限
- 无法追溯详细变更

适用场景：
- 简单记录，不需要详细分析

**方案二：库存变更日志**

核心思想：
记录所有库存变更，按需查询历史状态。

设计：
```sql
inventory_change_log
├── log_id
├── sku_id
├── change_type（ORDER/CANCEL/REPLENISH/ADJUST）
├── quantity_delta（变更量，±）
├── stock_before（变更前库存）
├── stock_after（变更后库存）
├── reference_id（关联ID：order_id/po_id）
├── operator
└── created_at

查询某时刻库存：
1. 获取当前库存
2. 反向应用change_log（created_at > target_time）
3. 得到目标时刻库存
```

优点：
- 完整追溯
- 支持任意时刻查询
- 审计能力强

缺点：
- 查询需要计算
- 存储成本高

**方案三：定期快照+增量日志（推荐）**

核心思想：
定期保存全量快照，中间记录增量日志。

设计：
```sql
inventory_snapshot（快照，每小时）
├── snapshot_id
├── sku_id
├── stock
├── reserved_stock
├── snapshot_time
└── created_at

inventory_change_log（增量日志）
├── log_id
├── sku_id
├── change_type
├── quantity_delta
├── stock_after
├── reference_id
└── created_at

查询某时刻库存：
1. 找到目标时刻之前最近的快照
2. 应用快照之后的增量日志
3. 得到目标时刻库存

示例：
查询2024-04-18 15:30的库存
→ 找到15:00的快照（stock=100）
→ 应用15:00-15:30的日志（-5, -3, -2）
→ 结果：100 - 5 - 3 - 2 = 90
```

优点：
- 平衡性能和存储
- 快照恢复快
- 审计能力强

缺点：
- 实现复杂度中等

**方案对比**：

| 方案 | 查询性能 | 存储成本 | 审计能力 | 实施难度 |
|------|---------|---------|---------|---------|
| 冗余字段 | ★★★★★ | ★★★★★ | ★★☆☆☆ | ★★★★★ |
| 变更日志 | ★★★☆☆ | ★★☆☆☆ | ★★★★★ | ★★★☆☆ |
| 快照+日志 | ★★★★☆ | ★★★★☆ | ★★★★★ | ★★★☆☆ |

**推荐方案**：
采用**定期快照+增量日志**。

实施要点：

1. **快照生成策略**：
   ```
   定时快照：
   - 每小时生成一次快照
   - 或库存变更超过1000次时生成
   
   快照内容：
   - SKU ID
   - 物理库存
   - 预占库存
   - 已售库存
   - 可售库存
   - 快照时间
   ```

2. **变更日志记录**：
   ```java
   @Aspect
   public class InventoryChangeLogger {
     @Around("execution(* InventoryService.deduct*(..))")
     public Object logChange(ProceedingJoinPoint pjp) {
       // 记录变更前库存
       int stockBefore = getStock(skuId);
       
       // 执行扣减
       Object result = pjp.proceed();
       
       // 记录变更后库存
       int stockAfter = getStock(skuId);
       
       // 保存日志
       InventoryChangeLog log = new InventoryChangeLog();
       log.setSkuId(skuId);
       log.setChangeType("ORDER");
       log.setQuantityDelta(stockBefore - stockAfter);
       log.setStockBefore(stockBefore);
       log.setStockAfter(stockAfter);
       log.setReferenceId(orderId);
       logRepository.save(log);
       
       return result;
     }
   }
   ```

3. **历史库存查询API**：
   ```
   GET /api/inventory/{skuId}/history?time=2024-04-18T15:30:00
   
   响应：
   {
     "skuId": "123",
     "stock": 90,
     "reserved": 10,
     "available": 80,
     "snapshotTime": "2024-04-18T15:30:00"
   }
   ```

4. **数据归档**：
   ```
   归档策略：
   - 变更日志保留90天
   - 90天后归档到对象存储（OSS）
   - 快照保留1年
   - 1年后删除（保留年度快照）
   ```

5. **应用场景**：
   ```
   场景1：售后分析
   用户投诉超卖 → 查询下单时库存 → 分析扣减日志 → 定位问题
   
   场景2：数据对账
   每日对账：今日库存 = 昨日库存 + 今日入库 - 今日出库
   不一致 → 查询变更日志 → 找出差异
   
   场景3：报表统计
   生成"每日库存报表" → 查询每日0点快照 → 生成报表
   ```

**延伸思考**：
1. 如何设计库存变更的审计流程？
2. 变更日志如何支持回滚操作？
3. 大批量商品的快照如何优化存储？

---

#### 📊 题目13：库存的实时性vs一致性权衡

**问题描述**：
库存系统中，Redis提供高性能但可能丢失数据，MySQL提供强一致但性能较低。如何在实时性和一致性之间权衡？

**答案**：

**问题分析**：
实时性vs一致性的核心矛盾：
1. 用户期望实时看到库存
2. 系统要保证不超卖
3. 高并发下性能压力大
4. 数据一致性难保证

**方案一：强一致性优先（MySQL为准）**

核心思想：
所有库存操作直接读写MySQL，放弃Redis。

设计：
```sql
-- 使用悲观锁
BEGIN;
SELECT stock FROM inventory WHERE sku_id=? FOR UPDATE;
UPDATE inventory SET stock = stock - ? WHERE sku_id=?;
COMMIT;
```

CAP理论选择：
- C（一致性）：强一致性
- A（可用性）：可用性一般（锁冲突）
- P（分区容错）：单机MySQL，不支持分区

优点：
- 绝对一致性
- 不会超卖
- 不会丢数据

缺点：
- 性能差（1000-5000 TPS）
- 无法支持秒杀
- 并发度低

适用场景：
- 库存量少的高价商品（奢侈品）
- 对一致性要求极高的场景

**方案二：最终一致性（Redis为主）**

核心思想：
库存扣减在Redis，异步同步到MySQL。

设计：
```text
扣减流程：
1. Redis DECR扣减
2. 扣减成功，创建订单
3. 异步同步到MySQL

同步策略：
- 定时任务（每10秒）批量同步
- 或消息队列异步同步

数据恢复：
- Redis故障 → 从MySQL加载
- 对账任务（每小时）纠正差异
```

CAP理论选择：
- C（一致性）：最终一致性
- A（可用性）：高可用
- P（分区容错）：支持分区

优点：
- 性能极高（10万+ TPS）
- 支持高并发
- 用户体验好

缺点：
- Redis和MySQL可能不一致
- Redis故障可能丢数据
- 需要对账机制

适用场景：
- 秒杀场景
- 高并发场景
- 普通商品

**方案三：分层一致性（推荐）**

核心思想：
根据商品类型和场景，采用不同一致性策略。

设计：
```text
商品分类：
1. 高价商品（>10000元）：
   - 使用MySQL悲观锁
   - 强一致性
   - 不追求性能
   
2. 秒杀商品：
   - 使用Redis+Lua
   - 最终一致性
   - 极致性能
   
3. 普通商品：
   - 使用MySQL乐观锁
   - 强一致性
   - 中等性能

扣减逻辑：
if (product.type == HIGH_VALUE) {
  return deductWithPessimisticLock();
} else if (product.type == SECKILL) {
  return deductWithRedis();
} else {
  return deductWithOptimisticLock();
}
```

优点：
- 灵活权衡
- 性能和一致性兼顾
- 差异化服务

缺点：
- 实现复杂
- 需要商品分类

**方案对比**：

| 方案 | 一致性 | 性能 | 实现难度 | 适用场景 |
|------|--------|------|---------|----------|
| 强一致 | ★★★★★ | ★★☆☆☆ | ★★★★☆ | 高价商品 |
| 最终一致 | ★★★☆☆ | ★★★★★ | ★★★☆☆ | 秒杀 |
| 分层一致 | ★★★★☆ | ★★★★☆ | ★★☆☆☆ | 综合场景 |

**推荐方案**：
采用**分层一致性**。

实施要点：

1. **一致性级别定义**：
   ```
   强一致（Strong Consistency）：
   - MySQL事务
   - 悲观锁或串行化
   - 实时一致
   
   最终一致（Eventual Consistency）：
   - Redis扣减 + 异步同步
   - 秒级延迟
   - 需要对账
   
   因果一致（Causal Consistency）：
   - 同一用户操作有序
   - 不同用户可能看到不同状态
   ```

2. **降级策略**：
   ```
   正常模式：
   - 秒杀商品：Redis（最终一致）
   - 普通商品：MySQL乐观锁（强一致）
   
   降级模式（Redis故障）：
   - 秒杀商品：暂停售卖或限流到MySQL
   - 普通商品：MySQL悲观锁
   
   极端模式（MySQL故障）：
   - 只读Redis，禁止扣减
   - 提示用户稍后再试
   ```

3. **一致性检查**：
   ```
   实时检查：
   - 扣减后检查Redis和MySQL差异
   - 差异 > 阈值（如100）→ 告警
   
   定期对账：
   - 每小时全量对账
   - 自动纠正小差异（< 5）
   - 大差异（> 10）→ 人工介入
   ```

4. **监控指标**：
   ```
   一致性指标：
   - Redis-MySQL差异数量
   - 差异持续时间
   - 对账修复次数
   
   性能指标：
   - 扣减TPS
   - 扣减耗时P99
   - Redis命中率
   ```

**延伸思考**：
1. 如何设计Redis的持久化策略（AOF/RDB）？
2. 分布式场景下如何保证Redis和MySQL一致性？
3. CAP理论在库存系统中如何权衡？

---

#### 🔧 题目14：库存回滚机制的设计

**问题描述**：
用户下单后未支付，或者订单取消，需要回滚库存。如何设计库存回滚机制，保证幂等性和正确性？

**答案**：

**问题分析**：
库存回滚的核心场景：
1. 订单取消（用户主动取消）
2. 超时未支付（30分钟自动取消）
3. 支付失败（扣款失败）
4. 售后退货（订单完成后退货）

核心挑战：
1. 幂等性：重复回滚不能多加库存
2. 并发安全：多个回滚请求同时执行
3. 部分回滚：一单多商品部分退货
4. 补偿机制：回滚失败如何处理

**方案一：直接加库存**

核心思想：
取消订单时直接增加库存。

实现：
```sql
-- 订单取消
UPDATE inventory 
SET stock = stock + quantity
WHERE sku_id = ?;

-- 更新订单状态
UPDATE orders 
SET status = 'CANCELLED'
WHERE order_id = ?;
```

优点：
- 实现简单

缺点：
- 无法保证幂等性（重复调用会多加库存）
- 并发不安全

**方案二：基于订单状态回滚**

核心思想：
检查订单状态，只有首次取消才回滚库存。

实现：
```sql
-- 原子更新订单状态
UPDATE orders 
SET status = 'CANCELLED'
WHERE order_id = ? AND status = 'PENDING';

if (affected_rows == 1) {
  // 状态更新成功，说明是首次取消
  UPDATE inventory 
  SET reserved_stock = reserved_stock - quantity,
      available_stock = available_stock + quantity
  WHERE sku_id = ?;
}
```

优点：
- 保证幂等性
- 并发安全

缺点：
- 需要精确的状态流转
- 状态机复杂

**方案三：回滚记录表（推荐）**

核心思想：
维护库存回滚记录，保证幂等性和可追溯。

设计：
```sql
inventory_rollback
├── rollback_id
├── order_id
├── sku_id
├── quantity
├── rollback_type（CANCEL/REFUND/TIMEOUT）
├── status（PENDING/SUCCESS/FAILED）
├── retry_count
├── created_at
└── updated_at

回滚流程：
1. 创建回滚记录（唯一约束：order_id + sku_id）
2. 执行回滚：
   UPDATE inventory 
   SET reserved_stock = reserved_stock - quantity
   WHERE sku_id = ?;
   
3. 更新回滚记录状态为SUCCESS
4. 如果失败，标记为FAILED，后台重试

幂等性保证：
INSERT INTO inventory_rollback (order_id, sku_id, quantity)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE updated_at = NOW();

if (affected_rows == 1) {
  // 首次插入，执行回滚
  doRollback();
}
```

优点：
- 幂等性强
- 可追溯
- 支持重试
- 审计友好

缺点：
- 实现复杂度高
- 需要额外表

**方案对比**：

| 方案 | 幂等性 | 并发安全 | 可追溯 | 实施难度 |
|------|--------|---------|--------|---------|
| 直接加库存 | ★☆☆☆☆ | ★★☆☆☆ | ★☆☆☆☆ | ★★★★★ |
| 基于状态 | ★★★★☆ | ★★★★☆ | ★★★☆☆ | ★★★☆☆ |
| 回滚记录 | ★★★★★ | ★★★★★ | ★★★★★ | ★★☆☆☆ |

**推荐方案**：
采用**回滚记录表**。

实施要点：

1. **回滚类型设计**：
   ```
   CANCEL：订单取消
   - 释放预占库存
   - 回补可售库存
   
   REFUND：售后退货
   - 增加物理库存
   - 增加可售库存
   
   TIMEOUT：超时未支付
   - 释放预占库存
   
   ADJUST：库存调整（人工）
   ```

2. **回滚执行逻辑**：
   ```java
   @Transactional
   public void rollbackInventory(String orderId) {
     // 1. 创建回滚记录（幂等键）
     RollbackRecord record = new RollbackRecord();
     record.setOrderId(orderId);
     record.setSkuId(skuId);
     record.setQuantity(quantity);
     record.setStatus("PENDING");
     
     try {
       rollbackRepository.insert(record);
     } catch (DuplicateKeyException e) {
       // 已存在回滚记录，直接返回
       return;
     }
     
     // 2. 执行库存回滚
     try {
       inventoryService.release(skuId, quantity);
       record.setStatus("SUCCESS");
     } catch (Exception e) {
       record.setStatus("FAILED");
       record.setRetryCount(record.getRetryCount() + 1);
       throw e;
     } finally {
       rollbackRepository.update(record);
     }
   }
   ```

3. **部分退货处理**：
   ```
   场景：用户购买3件商品，退货1件
   
   处理：
   1. 创建部分回滚记录
   2. 回滚数量 = 退货数量（1件）
   3. 更新订单项状态（2件已发货，1件已退货）
   ```

4. **失败重试**：
   ```
   补偿Worker：
   1. 定时扫描FAILED状态的回滚记录
   2. 重试执行回滚
   3. 最多重试5次
   4. 仍失败 → 转人工处理
   ```

5. **监控告警**：
   ```
   指标：
   - 回滚成功率
   - 回滚延迟（下单到回滚的时间）
   - 失败回滚数量
   
   告警：
   - 回滚成功率 < 99%
   - 失败回滚 > 100条
   ```

**延伸思考**：
1. 如何防止恶意下单占用库存？
2. 库存回滚失败如何人工介入？
3. 大批量订单取消如何优化回滚性能？

---

#### 💡 题目15：跨境电商的库存管理（多国库存）

**问题描述**：
跨境电商在中国、美国、欧洲都有仓库，同一商品在不同地区有库存。如何设计全球库存管理系统？

**答案**：

**问题分析**：
跨境库存的核心挑战：
1. 时区差异（中国和美国相差12小时）
2. 币种不同（人民币、美元、欧元）
3. 清关周期长（跨境物流10-30天）
4. 库存调拨困难

**方案一：独立库存池**

核心思想：
每个国家/地区独立管理库存，互不共享。

设计：
```sql
inventory
├── sku_id
├── country_code（US/CN/EU）
├── warehouse_id
├── stock
└── currency

用户购买：
1. 根据用户IP或选择的站点确定国家
2. 查询该国家的库存
3. 扣减该国家库存
4. 不跨国发货
```

优点：
- 实现简单
- 各国独立运营
- 无跨境调拨

缺点：
- 库存利用率低（美国有货但中国无货）
- 用户体验差（本地无货无法购买）

**方案二：全球库存池（虚拟统一）**

核心思想：
虚拟层展示全球总库存，实际按地区分配。

设计：
```text
虚拟层：
global_inventory
├── sku_id
├── total_stock = sum(所有国家库存)

实际层：
regional_inventory
├── sku_id
├── region_code
├── stock

用户下单：
1. 展示全球总库存（用户可见）
2. 选择发货国家（就近优先）
3. 扣减该国库存
4. 跨境发货（如果本地无货）
```

优点：
- 用户体验好（看到全球库存）
- 库存利用率高
- 支持跨境发货

缺点：
- 跨境物流慢、贵
- 复杂的库存分配

**方案三：混合模式（推荐）**

核心思想：
优先本地发货，支持跨境应急。

设计：
```text
库存层级：
1. 本地库存（Local Stock）：
   - 用户所在国家的库存
   - 优先扣减
   - 配送快（2-3天）

2. 区域库存（Regional Stock）：
   - 相邻国家的库存
   - 次优选择
   - 配送中等（5-7天）

3. 全球库存（Global Stock）：
   - 其他国家的库存
   - 最后选择
   - 配送慢（10-30天）

路由策略：
1. 查询本地库存
   - 有货 → 本地发货
2. 查询区域库存
   - 有货 → 跨境发货（用户确认）
3. 查询全球库存
   - 有货 → 全球发货（用户确认）
4. 都无货 → 缺货
```

优点：
- 平衡速度和成本
- 灵活
- 用户可选

缺点：
- 需要智能路由
- 用户决策成本

**方案对比**：

| 方案 | 库存利用率 | 配送速度 | 用户体验 | 实施难度 |
|------|-----------|---------|---------|---------|
| 独立池 | ★★☆☆☆ | ★★★★★ | ★★★☆☆ | ★★★★★ |
| 全球池 | ★★★★★ | ★★☆☆☆ | ★★★★★ | ★★★☆☆ |
| 混合模式 | ★★★★☆ | ★★★★☆ | ★★★★☆ | ★★☆☆☆ |

**推荐方案**：
采用**混合模式**。

实施要点：

1. **库存数据结构**：
   ```sql
   global_inventory
   ├── sku_id
   ├── region_code（US/CN/EU/JP）
   ├── warehouse_id
   ├── stock
   ├── currency
   ├── local_price（本地售价）
   └── shipping_cost_to_other（跨境运费）
   ```

2. **库存分配策略**：
   ```
   初始分配（新品上架）：
   - 根据各地区历史销量预测
   - US: 40%, EU: 30%, CN: 20%, JP: 10%
   
   动态调整（运营中）：
   - 每周根据销量调整
   - 滞销地区调拨到热销地区
   ```

3. **跨境发货流程**：
   ```
   用户下单：
   1. 显示配送选项：
      - 本地发货（2-3天，免运费）
      - 跨境发货（10-15天，运费$20）
   
   2. 用户选择跨境发货
   
   3. 扣减源国库存
   
   4. 清关、物流
   
   5. 配送到用户
   ```

4. **币种和价格**：
   ```
   价格策略：
   - 每个地区独立定价（考虑关税、运费）
   - 实时汇率转换
   
   示例：
   商品成本：$100
   - 美国售价：$150（含税15%，利润$35）
   - 中国售价：¥1200（含税13%，利润约$40）
   - 欧洲售价：€140（含税20%，利润约$30）
   ```

5. **库存同步**：
   ```
   同步机制：
   - 各地区库存独立数据库
   - 聚合到全球视图（Redis缓存）
   - 更新延迟 < 1秒
   
   时区处理：
   - 所有时间戳使用UTC
   - 本地展示转换为用户时区
   ```

**延伸思考**：
1. 如何设计跨境库存调拨的审批流程？
2. 清关失败如何处理库存回滚？
3. 不同国家的退货政策如何影响库存管理？

---
