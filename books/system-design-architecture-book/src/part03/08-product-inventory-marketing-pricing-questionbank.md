# 第 40 章 电商架构面试题精选（二）：商品、库存、营销与计价

> 本章是电商架构面试题库的第二部分，题库使用说明与面试官导航见[第 43 章](./03-ecommerce-architecture-interview.md)。

## 40.1 商品、库存、营销与计价题库

本专题聚焦电商供给与转化基础能力，按题型拆成三个子章节：

- [35.2.1 商品中心系统](./03-ecommerce-architecture-interview.md)
- [35.2.2 库存系统](./03-ecommerce-architecture-interview.md)
- [35.2.3 营销与计价系统](./03-ecommerce-architecture-interview.md)

建议先读商品中心，确认主数据与快照边界，再读库存和营销计价。

---

### 40.1.1 商品中心系统（16题）

##### 📊 题目1：设计支持多品类的SPU/SKU数据模型

**问题描述**：
电商平台需要支持实物商品（服装、3C）、虚拟商品（充值卡、会员）、服务类商品（保险、课程）。如何设计一个统一且可扩展的商品数据模型？

**答案**：

**问题分析**：
多品类商品模型的核心挑战：
1. 不同品类属性差异巨大（服装有尺码颜色，充值卡有卡密）
2. 需要支持灵活的属性扩展，避免频繁加字段
3. 查询性能要求高（详情页、列表页高并发）
4. 需要支持类目体系和属性继承

**方案一：EAV（实体-属性-值）模式**

核心思想：
将商品属性拆分为独立的键值对存储。

表结构：
```sql
product（商品主表）
├── product_id
├── spu_code
├── category_id
├── name
└── status

product_attribute（属性表）
├── product_id
├── attribute_key
├── attribute_value
└── attribute_type

category_template（类目模板）
├── category_id
├── attribute_definitions（JSON）
└── validation_rules
```

优点：
- 扩展性极强，加属性不需要改表结构
- 适合属性差异大的场景
- 灵活度高

缺点：
- 查询性能差（需要多次JOIN）
- 难以建立索引
- 类型校验在应用层
- SQL复杂

**方案二：宽表+JSON扩展字段**

核心思想：
核心字段固定，扩展字段用JSON存储。

表结构：
```sql
product
├── id, spu, name, category
├── common_attrs（固定字段：brand、主图等）
└── ext_attrs（JSONB：类目特有属性）

sku
├── sku_code, spu_id
├── spec_attrs（JSONB：颜色、尺码等规格）
└── ext_attrs（JSONB：其他扩展）
```

优点：
- 查询性能好（单表查询）
- PostgreSQL的JSONB支持索引
- 平衡灵活性和性能

缺点：
- JSON字段查询能力有限
- 需要应用层解析和校验
- 不同数据库支持程度不同

**方案三：混合模式（推荐）**

核心设计：
1. **主表存储通用字段**：product_core（id, spu, name, category, status）
2. **类目模板定义属性规范**：attribute_meta（属性元数据、类型、校验规则）
3. **分层存储**：
   - product_common_attr：高频查询字段（品牌、价格区间）
   - product_ext_attr：JSONB，低频字段
   - product_spec：SKU规格，单独表
4. **搜索侧异步构建宽表**：ES文档包含所有筛选字段

数据流：
- **写入**：商品创建 → 按模板校验 → 分表存储 → 事件发布 → ES同步
- **读取详情页**：主表+扩展表（缓存）
- **读取列表页**：直接查ES
- **后台管理**：全量字段（可接受慢查询）

优点：
- 扩展性强
- 查询性能好
- 支持复杂筛选（通过ES）
- 核心字段有索引

缺点：
- 架构复杂度中等
- 需要维护ES同步
- 最终一致性

**方案对比**：

| 维度 | EAV | 宽表+JSON | 混合模式 |
|------|-----|-----------|----------|
| 扩展性 | ★★★★★ | ★★★★☆ | ★★★★★ |
| 查询性能 | ★★☆☆☆ | ★★★★☆ | ★★★★★ |
| 开发复杂度 | ★★★☆☆ | ★★★★☆ | ★★★☆☆ |
| 类型安全 | ★★☆☆☆ | ★★★☆☆ | ★★★★☆ |

**推荐方案**：
采用**混合模式**。

实施要点：
1. **核心字段晋升机制**：高频查询字段从JSON移到固定列
2. **JSONB索引**：PostgreSQL建立GIN索引
3. **ES映射模板**：自动从类目模板生成
4. **缓存策略**：L1进程内 + L2 Redis，TTL分层设置
5. **属性校验**：类目模板定义规则，运行时校验

虚拟商品特殊处理：
- 充值卡：卡密存储加密、核销记录独立表
- 会员服务：有效期、权益包用JSON存储
- 服务类：预约时间、服务人员信息扩展字段

**延伸思考**：
1. 如何处理类目属性变更（模板升级）？
2. 历史订单中的商品快照如何存储？
3. 跨类目搜索时如何统一属性映射？

---

##### 🔧 题目2：商品详情页的缓存架构设计

**问题描述**：
商品详情页是电商系统访问量最大的页面，QPS可达百万级。请设计商品详情页的缓存架构，保证高性能和数据一致性。

**答案**：

**问题分析**：
详情页缓存的核心挑战：
1. 流量巨大，需要多级缓存
2. 数据来源多（商品、价格、库存、营销），聚合复杂
3. 数据更新频繁，缓存一致性难保证
4. 热点商品流量集中

**方案一：纯CDN缓存**

核心思想：
详情页直接缓存在CDN，用户请求直接命中CDN。

设计：
```text
用户 → CDN → 源站

CDN配置：
- 缓存时间：5分钟
- 缓存键：/product/{productId}
- 回源：CDN未命中时请求源站

更新策略：
- 商品信息变更 → 主动刷新CDN
- 或等待TTL过期自然更新
```

优点：
- 性能极高（边缘节点响应）
- 减轻源站压力
- 成本低

缺点：
- 实时性差（分钟级延迟）
- 个性化内容难处理（如用户登录状态）
- 价格库存等动态信息不适合

适用场景：
- 纯静态内容（商品图文）
- 对实时性要求不高

**方案二：多级缓存（推荐）**

核心思想：
L1本地缓存 + L2 Redis + L3数据库。

架构：
```text
用户 → 应用服务器
       ├→ L1: 本地缓存（Caffeine/Guava）
       ├→ L2: Redis（集中式）
       └→ L3: MySQL（源数据）

缓存策略：
L1: 热点数据，容量1000条，TTL 30秒
L2: 全量数据，TTL 5分钟
L3: 源数据

查询流程：
1. 查L1，命中返回
2. L1未命中，查L2，写入L1，返回
3. L2未命中，查L3，写入L2和L1，返回
```

详情页数据聚合：
```text
详情页数据：
- 商品基本信息（商品中心）→ 缓存5分钟
- 价格信息（计价系统）→ 缓存1分钟
- 库存信息（库存系统）→ 不缓存或缓存10秒
- 营销信息（营销系统）→ 缓存1分钟
- 推荐商品（推荐系统）→ 缓存30分钟

聚合策略：
// 并行调用
Future<Product> product = getProductAsync(productId);
Future<Price> price = getPriceAsync(productId);
Future<Stock> stock = getStockAsync(productId);
Future<Promotion> promo = getPromotionAsync(productId);

// 等待所有结果
ProductDetail detail = new ProductDetail(
  product.get(500, MILLISECONDS),
  price.get(300, MILLISECONDS),
  stock.get(200, MILLISECONDS),
  promo.get(300, MILLISECONDS)
);
```

优点：
- 性能好（多级缓存）
- 灵活度高（可针对不同数据设置不同TTL）
- 支持个性化

缺点：
- 架构复杂度中等
- 缓存一致性需要处理
- 多级缓存增加运维成本

**方案三：缓存+预热+旁路**

核心思想：
提前预热热点数据，冷数据旁路查询。

设计：
```text
1. 预热：
   - 大促前：提前加载热销商品
   - 运营后台：手动预热重点商品
   - 定时任务：每小时预热TOP 1000热门商品

2. 热点识别：
   - 实时统计访问频率
   - 超过阈值的商品加入热点列表
   - 热点商品缓存时间更长

3. 旁路加载：
   - 热点商品：L1+L2缓存
   - 普通商品：L2缓存
   - 长尾商品：直接查数据库

4. 缓存更新：
   - 商品信息变更 → 发布事件 → 主动失效缓存
   - 或使用版本号：缓存键包含版本号
```

优点：
- 热点商品性能极高
- 资源利用率高
- 大促效果好

缺点：
- 预热逻辑复杂
- 热点识别有延迟
- 需要实时监控

**方案对比**：

| 维度 | 纯CDN | 多级缓存 | 缓存+预热 |
|------|-------|----------|----------|
| 性能 | ★★★★★ | ★★★★☆ | ★★★★★ |
| 实时性 | ★★☆☆☆ | ★★★★☆ | ★★★★☆ |
| 个性化 | ★☆☆☆☆ | ★★★★★ | ★★★★★ |
| 复杂度 | ★★★★★ | ★★★☆☆ | ★★☆☆☆ |

**推荐方案**：
采用**多级缓存+热点预热**的组合。

实施要点：

1. **缓存分层**：
   ```
   L1（本地缓存）：
   - 容量：1000条
   - TTL：30秒
   - 淘汰策略：LRU
   - 适用：超热门商品（TOP 100）
   
   L2（Redis）：
   - 容量：100万条
   - TTL：5分钟
   - 集群部署：主从+哨兵
   - 适用：热门+普通商品
   
   L3（数据库）：
   - 全量数据
   - 读写分离
   ```

2. **缓存键设计**：
   ```
   方案1：不带版本号
   Key: product:detail:{productId}
   Value: JSON
   更新：商品变更时主动删除key
   
   方案2：带版本号（推荐）
   Key: product:detail:{productId}:{version}
   Value: JSON
   更新：版本号+1，旧key自然过期
   ```

3. **缓存更新策略**：
   ```
   Cache Aside模式：
   1. 读取：先查缓存，未命中再查DB，写入缓存
   2. 更新：先更新DB，再删除缓存
   
   Write Through模式：
   1. 更新：同时更新DB和缓存
   2. 读取：直接读缓存
   ```

4. **热点治理**：
   ```
   识别热点：
   - 实时统计访问频率（滑动窗口）
   - 超过阈值（如10000 QPS）标记为热点
   
   热点处理：
   - 本地缓存延长TTL（30秒 → 5分钟）
   - Redis分片存储（product:123:1, product:123:2...）
   - 限流保护（单商品限流）
   ```

5. **缓存穿透/击穿/雪崩**：
   ```
   穿透（查询不存在的数据）：
   - 布隆过滤器预判
   - 空值缓存（TTL短，如1分钟）
   
   击穿（热点key过期）：
   - 互斥锁（只有一个请求回源）
   - 热点key永不过期（后台异步更新）
   
   雪崩（大量key同时过期）：
   - TTL加随机值（5分钟±30秒）
   - 缓存预热
   - 降级方案（返回旧数据）
   ```

**延伸思考**：
1. 缓存和数据库数据不一致如何处理？
2. 如何设计缓存的监控指标？
3. 大促时如何做缓存容量规划？

---

##### 💡 题目3：如何解决商品信息变更后搜索不一致问题？

**问题描述**：
运营修改了商品标题和价格，但搜索结果中仍然显示旧信息。这是典型的最终一致性问题。如何设计商品到搜索的数据同步方案？

**答案**：

**问题分析**：
商品搜索一致性的核心挑战：
1. 数据变更频繁（价格调整、库存变化）
2. 搜索索引构建有延迟
3. 用户期望实时看到最新信息
4. 大量商品同步对ES集群压力大

**方案一：实时同步（强一致性）**

核心思想：
商品信息变更时，同步更新ES索引。

设计：
```text
1. 运营后台：修改商品信息
2. 商品服务：
   BEGIN TRANSACTION
     UPDATE products SET title=?, price=?
     // 同步更新ES
     esClient.update(productId, {title, price})
   COMMIT
3. 用户搜索：立即看到最新数据
```

优点：
- 实时一致性
- 用户体验好

缺点：
- ES更新慢（可能超时）
- 影响商品更新性能
- ES故障影响商品服务

适用场景：
- 对一致性要求极高
- 变更频率低

**方案二：异步同步（最终一致性）**

核心思想：
通过消息队列异步同步，保证最终一致性。

设计：
```text
1. 商品服务：
   BEGIN TRANSACTION
     UPDATE products SET title=?, price=?, version=version+1
     INSERT INTO outbox_events (
       event_type='ProductUpdated',
       payload={productId, title, price, version}
     )
   COMMIT

2. 事件发布器：
   扫描outbox_events → 发送到Kafka

3. 搜索同步Worker：
   监听Kafka ProductUpdated事件
   更新ES索引
   
4. 幂等处理：
   根据version判断是否需要更新
   if (event.version > es_doc.version) {
     update ES
   }
```

优点：
- 解耦，不影响商品服务性能
- ES故障不影响商品更新
- 支持重试和补偿

缺点：
- 最终一致性（秒级延迟）
- 实现复杂度中等

适用场景：
- 大部分场景
- 可接受秒级延迟

**方案三：双写+对账**

核心思想：
同时写MySQL和ES，对账纠正不一致。

设计：
```text
1. 商品服务写入：
   // 双写（并行）
   Future<Void> f1 = mysqlClient.update(...)
   Future<Void> f2 = esClient.update(...)
   
   // 等待两个都成功
   f1.get()
   f2.get()

2. 对账任务（每小时）：
   - 查询MySQL最近变更的商品
   - 与ES中的数据对比
   - 发现不一致，重新同步

3. 增量同步（每分钟）：
   - 基于updated_at增量同步
   - 作为对账的补充
```

优点：
- 接近实时
- 有补偿机制

缺点：
- 双写失败处理复杂
- 两个数据源可能不一致
- 实现复杂

**方案对比**：

| 维度 | 实时同步 | 异步同步 | 双写+对账 |
|------|---------|---------|-----------|
| 实时性 | ★★★★★ | ★★★★☆ | ★★★★☆ |
| 系统解耦 | ★★☆☆☆ | ★★★★★ | ★★★☆☆ |
| 一致性保证 | ★★★★☆ | ★★★★★ | ★★★★★ |
| 实施难度 | ★★★★☆ | ★★★☆☆ | ★★☆☆☆ |

**推荐方案**：
采用**异步同步+对账**。

实施要点：

1. **事件设计**：
   ```
   ProductCreated：商品创建
   ProductUpdated：商品信息变更（title、desc、images）
   ProductPriceChanged：价格变更
   ProductStatusChanged：上下架
   ProductDeleted：删除
   ```

2. **同步Worker设计**：
   ```
   消费逻辑：
   1. 从Kafka消费ProductUpdated事件
   2. 根据productId查询完整商品信息
   3. 构建ES文档
   4. 批量更新ES（bulk API，提高吞吐）
   5. 提交offset
   
   批量优化：
   - 攒批：100条或1秒批量提交
   - 去重：同一商品多次变更只保留最新
   - 合并：多个字段变更合并为一次更新
   ```

3. **幂等处理**：
   ```
   ES文档设计：
   {
     "productId": "123",
     "title": "iPhone 15",
     "price": 5999,
     "version": 10,  // 版本号
     "updatedAt": 1679800000
   }
   
   更新逻辑：
   if (event.version > doc.version) {
     update ES
   } else {
     skip (乱序消息)
   }
   ```

4. **对账机制**：
   ```
   对账任务（每小时）：
   SELECT product_id, version, updated_at 
   FROM products 
   WHERE updated_at >= NOW() - INTERVAL 2 HOUR
   
   对每个商品：
   - 查询ES中的version
   - 如果MySQL.version > ES.version
   - 发送补偿事件到Kafka
   ```

5. **监控告警**：
   ```
   指标：
   - 同步延迟（消息产生到ES更新完成的时间）
   - 失败率（同步失败的比例）
   - 对账差异数（MySQL和ES不一致的商品数）
   
   告警：
   - 同步延迟 > 10秒
   - 失败率 > 1%
   - 对账差异 > 100条
   ```

**延伸思考**：
1. 如果ES集群故障，搜索如何降级？
2. 商品删除后ES索引如何处理？
3. 大批量商品导入如何优化ES同步性能？

---

##### 🔧 题目3 扩展：直接订阅 Binlog 同步 ES 的弊端是什么？如果不同变更之间存在依赖关系，应该怎么处理？

**问题描述**：
一些电商系统会通过 Binlog / CDC 捕获商品表变更，然后由 ES Synchronizer 消费消息并更新搜索索引。例如商品主表、SKU 表、Offer 表、类目映射表、供应商映射表发生变更后，同步服务根据表名和字段变化去更新 ES 文档。这种方式有什么弊端？如果一个 ES 文档依赖多张表，不同变更之间存在先后关系和依赖关系，应该如何设计？

**答案**：

**问题分析**：

直接订阅 Binlog 同步 ES 的本质是：

```text
数据库表级变化
  → 触发 ES 文档更新
```

而商品搜索索引的本质通常是：

```text
多张业务表
  → 聚合成一个商品搜索宽文档
```

两者粒度不一致。Binlog 看到的是“某张表某一行变了”，ES 需要的是“某个商品聚合视图应该变成什么样”。这会带来几个典型问题：

1. **业务语义弱**：Binlog 只表达 `insert/update/delete`，不表达 `ProductPublished`、`ProductOffline`、`OfferChanged`、`RefundRuleChanged`。
2. **强依赖表结构**：字段新增、删除、顺序变化、JSON 结构变化，都可能影响同步逻辑。
3. **跨表依赖复杂**：一个 ES 商品文档可能依赖 item、spu、sku、offer、resource、category、stock config、refund rule 等多张表。
4. **顺序不稳定**：同一业务发布可能写多张表，Binlog 事件到达不同 consumer 时不一定按业务语义有序。
5. **并发覆盖风险**：两个表变更同时 patch 同一个 ES doc，可能出现后写基于旧 doc 覆盖前写结果。
6. **版本语义不足**：Binlog timestamp 或 position 不等价于商品业务版本，难以判断旧事件是否应该覆盖新事件。
7. **失败补偿困难**：失败消息只知道表和字段，不一定知道影响哪个商品、哪个发布版本、是否可以安全重建。

**典型错误做法：按每条 Binlog 直接 patch ES**

```text
carrier_tab update
  → 查询旧 ES doc
  → 修改 carrier 基础字段
  → update ES

mapping_tab update
  → 查询旧 ES doc
  → 修改 support category / entrance
  → update ES
```

这种做法的问题是：两个 handler 都可能先读取旧 ES doc，再各自修改一部分字段，最后谁后写谁赢。如果后写的 doc 是基于旧版本读出来的，就可能把前一个变更覆盖掉。

**方案一：继续直接 Binlog Patch**

核心思想：
每张表的 Binlog handler 只更新 ES 文档中自己负责的字段。

优点：
- 实现直观。
- 延迟低。
- 不需要改上游业务系统。

缺点：
- 依赖关系散落在多个 handler 中。
- 跨表顺序难保证。
- 多个 handler patch 同一个 doc 时容易覆盖字段。
- 表结构变化会影响同步逻辑。
- 出问题后难以判断 ES doc 应该重建成什么样。

适用场景：
- ES 文档和 DB 表几乎一对一。
- 变更字段简单，没有跨表依赖。
- 对一致性要求不高。

**方案二：Binlog 只标记 Dirty Doc，再重建完整 ES 文档**

核心思想：
Binlog 不直接写 ES，而是只负责发现“哪个聚合根脏了”。

```text
Binlog Event
  → 解析影响对象
  → mark dirty(doc_type, doc_id)
  → Index Worker 从 DB 读取最新数据
  → rebuild full ES doc
  → versioned upsert ES
```

例如：

```text
product_offer_tab changed
  → affected item_id = item_80001
  → mark dirty: product_doc / item_80001

refund_rule_tab changed
  → affected item_id = item_80001
  → mark dirty: product_doc / item_80001

category_mapping_tab changed
  → affected item_id list
  → mark dirty for each item
```

Dirty Doc 表可以这样设计：

```sql
CREATE TABLE es_sync_dirty_doc (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    doc_type VARCHAR(64) NOT NULL,
    doc_id VARCHAR(128) NOT NULL,
    source_table VARCHAR(128) NOT NULL,
    source_event_id VARCHAR(128) DEFAULT NULL,
    source_version BIGINT DEFAULT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING'
        COMMENT 'PENDING/RUNNING/SUCCESS/FAILED/DLQ',
    retry_count INT NOT NULL DEFAULT 0,
    next_retry_at DATETIME DEFAULT NULL,
    last_error_message VARCHAR(1024) DEFAULT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_doc (doc_type, doc_id),
    KEY idx_status_retry (status, next_retry_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ES 同步脏文档队列';
```

同一个 doc 在短时间内多次变化，只保留一条 dirty 记录：

```text
item update
offer update
refund rule update
  → 合并成 item_80001 的一次 rebuild
```

重建逻辑：

```text
读取 item_id
  → 查询 item 最新状态
  → 查询 spu / sku / offer
  → 查询类目、资源、库存配置、履约规则、退款规则
  → 判断是否应该被索引
      是：upsert ES doc
      否：delete ES doc
```

优点：
- 不依赖 Binlog 到达顺序。
- 不会因为局部 patch 覆盖字段。
- ES 文档构建逻辑集中。
- 可以合并多次变更，降低 ES 写入压力。
- 失败后可以按 `doc_type + doc_id` 重试和补偿。

缺点：
- 延迟比直接 patch 略高。
- 每次重建需要回查 DB，DB 压力更大。
- 需要维护 dependency mapping。

适用场景：
- ES 文档是多表聚合宽文档。
- 商品、Offer、类目、规则之间存在依赖。
- 搜索一致性和可恢复性比毫秒级延迟更重要。

**方案三：业务事件 + Outbox + 快照重建**

核心思想：
不要让 ES Synchronizer 从表级 Binlog 里猜业务含义，而是让商品发布链路明确发出业务事件。

```text
Publish Transaction
  → 写商品正式表
  → 写 publish_version
  → 写 product_snapshot
  → 写 product_outbox_event(ProductPublished)
  → ES Synchronizer 消费 ProductPublished
  → 按 item_id + publish_version 读取快照
  → rebuild ES doc
```

事件示例：

```json
{
  "event_id": "evt_20260428_000001",
  "event_type": "ProductPublished",
  "item_id": "item_80001",
  "publish_version": 4,
  "publish_id": "pub_20001",
  "snapshot_id": "snap_90001",
  "changed_fields": ["title", "offer", "refund_rule"]
}
```

ES 写入时带版本：

```text
if event.publish_version < es_doc.publish_version:
    ignore
else:
    upsert
```

优点：
- 业务语义清晰。
- 下游不依赖内部表结构。
- `publish_version` 可以防乱序。
- 可以基于发布快照构建 ES，结果更稳定。
- 排查问题时能回到一次发布动作，而不是一堆表变更。

缺点：
- 需要上游商品中心或供给平台改造。
- 需要设计事件契约和 Outbox。
- 对存量 Binlog 同步系统需要渐进迁移。

适用场景：
- 商品发布、上下架、封禁、回滚等核心业务链路。
- 多系统依赖商品变更通知。
- 搜索、缓存、计价上下文和营销资格消费者都需要一致理解商品版本。

**方案对比**：

| 维度 | 直接 Binlog Patch | Dirty Doc 重建 | 业务事件 + Outbox |
|------|-------------------|----------------|-------------------|
| 实现成本 | 低 | 中 | 中高 |
| 业务语义 | 弱 | 中 | 强 |
| 跨表依赖处理 | 差 | 好 | 很好 |
| 防并发覆盖 | 差 | 好 | 很好 |
| 防乱序能力 | 弱 | 中 | 强 |
| 对表结构耦合 | 强 | 中 | 弱 |
| 故障补偿 | 弱 | 好 | 很好 |
| 适合场景 | 简单索引 | 多表聚合索引 | 核心商品发布链路 |

**推荐方案**：

短期采用 **Binlog → Dirty Doc Queue → Full Rebuild ES Doc**，中长期演进到 **业务事件 + Outbox + 商品快照重建 ES**。

推荐落地路径：

1. **定义 ES doc 聚合根**

   ```text
   product index:
     doc_id = item_id

   carrier index:
     doc_id = carrier_id

   event index:
     doc_id = event_id
   ```

2. **维护依赖映射**

   ```text
   product_item_tab              → item_id
   product_offer_tab             → item_id
   product_refund_rule_tab       → item_id
   resource_tab                  → affected item_id list
   supplier_product_mapping_tab  → item_id
   category_mapping_tab          → affected item_id list
   ```

3. **Binlog handler 只做 mark dirty**

   ```text
   onBinlog(table, row):
       doc_ids = resolveAffectedDocIds(table, row)
       for doc_id in doc_ids:
           upsert es_sync_dirty_doc(doc_type, doc_id)
   ```

4. **Index Worker 串行处理同一个 doc**

   ```text
   SELECT *
   FROM es_sync_dirty_doc
   WHERE status = 'PENDING'
   ORDER BY updated_at ASC
   LIMIT 100;
   ```

   同一个 `doc_type + doc_id` 通过唯一键合并，Worker 抢占后重建完整文档。

5. **重建时读取 DB 最新状态**

   ```text
   buildProductDoc(item_id):
       item = query item
       offers = query offers
       rules = query fulfillment / refund rules
       if item is not indexable:
           delete ES doc
       else:
           upsert full doc
   ```

6. **写 ES 带版本**

   商品类索引用 `publish_version`；没有业务版本的对象至少使用 `updated_at`、`rebuild_seq` 或 `source_version`。

7. **失败进入 DLQ 和补偿**

   失败时记录：

   ```text
   doc_type
   doc_id
   source_table
   error_code
   retry_count
   next_retry_at
   ```

8. **定期 full sync 和对账**

   ```text
   DB latest hash != ES doc hash
     → mark dirty
   ```

   对于全量重建，建议使用新索引 + alias switch，避免重建期间影响线上查询。

**面试总结**：

直接订阅 Binlog 同步 ES 不是不能用，而是要清楚它的边界：

> Binlog 是表级数据变化，ES index 是业务聚合视图。两者粒度不一致，直接 patch ES 会在跨表依赖、事件顺序、并发覆盖、版本防乱序和失败补偿上变复杂。

更稳的设计是：

```text
短期：
  Binlog 只负责发现哪个 doc 脏了
  Dirty Queue 合并变更
  Worker 从 DB 重建完整 ES doc

长期：
  商品发布事务写 Outbox 业务事件
  ES Synchronizer 消费 ProductPublished / ProductOffline
  按 item_id + publish_version 读取商品快照
  versioned upsert ES
```

这样 ES 同步消费的是“商品版本已发布”这个业务事实，而不是从一堆表级 Binlog 里猜商品到底发生了什么。

**延伸思考**：

1. 如何设计 `resolveAffectedDocIds`，避免一张配置表变更导致全量商品都被标脏？
2. ES 写入使用 external version 有什么限制？
3. Dirty Queue 堆积时，如何区分高优先级商品和普通商品？
4. 全量重建和增量同步同时发生时，如何避免旧增量写到新索引？

---

##### 📊 题目4：设计商品类目体系和属性管理

**问题描述**：
电商平台有上千个类目（如手机、服装、食品），每个类目有不同的属性（手机有内存、颜色，服装有尺码、材质）。如何设计类目体系和属性管理系统？

**答案**：

**问题分析**：
类目属性管理的核心挑战：
1. 类目层级深（最多5-6级）
2. 属性类型多样（文本、数值、枚举、多选）
3. 属性继承和覆盖
4. 属性校验规则复杂

**方案一：树形类目+固定属性**

核心思想：
类目按树形组织，每个类目预定义固定属性。

设计：
```sql
category（类目表）
├── category_id
├── parent_id
├── name
├── level
├── path（/1/10/100/，便于查询祖先）
└── leaf（是否叶子节点）

category_attribute（类目属性定义）
├── category_id
├── attribute_id
├── required（是否必填）
└── display_order

attribute_definition（属性定义）
├── attribute_id
├── name
├── input_type（text/number/enum/multi_enum）
├── validation_rule（JSON）
└── options（枚举值）
```

优点：
- 结构清晰
- 属性定义规范
- 易于校验

缺点：
- 属性变更需要改表结构
- 不够灵活
- 类目迁移困难

**方案二：动态属性模板**

核心思想：
类目关联属性模板，属性模板可复用和继承。

设计：
```sql
category
├── category_id
├── parent_id
├── attribute_template_id（属性模板）
└── inherit_parent（是否继承父类目属性）

attribute_template（属性模板）
├── template_id
├── name
└── description

template_attribute（模板属性关联）
├── template_id
├── attribute_id
├── required
├── display_order
└── default_value

attribute_meta（属性元数据）
├── attribute_id
├── name
├── code（唯一标识，如"screen_size"）
├── data_type（string/int/decimal/enum/boolean）
├── input_type（input/select/checkbox/radio）
├── validation_rule（JSON：min/max/regex/enum_values）
└── searchable（是否可搜索）
```

继承规则：
```text
示例：手机 → 智能手机 → iPhone

手机类目（一级）：
- 品牌、型号、屏幕尺寸、操作系统

智能手机（二级）：
- 继承手机的所有属性
- 新增：前置摄像头、后置摄像头、电池容量

iPhone（三级）：
- 继承智能手机的所有属性
- 新增：Face ID、MagSafe
- 覆盖：操作系统固定为"iOS"
```

优点：
- 高度灵活
- 支持继承和复用
- 属性可动态添加

缺点：
- 实现复杂
- 继承逻辑复杂
- 性能有一定影响

**方案三：属性分组+扩展字段**

核心思想：
将属性分为核心属性（固定字段）和扩展属性（JSON）。

设计：
```sql
product
├── 核心属性（固定字段）：
│   brand_id, price, weight, status
└── 扩展属性（JSONB）：
    ext_attrs: {
      "screen_size": "6.1英寸",
      "memory": "256GB",
      "color": "深空黑"
    }

category_attr_group（属性分组）
├── category_id
├── group_name（基本信息/规格参数/包装清单）
└── attributes（JSON数组）
```

优点：
- 平衡性能和灵活性
- 核心属性有索引
- 扩展属性灵活

缺点：
- JSON查询能力有限
- 属性分组需要人工维护

**方案对比**：

| 维度 | 固定属性 | 动态模板 | 分组+扩展 |
|------|---------|---------|-----------|
| 灵活性 | ★★☆☆☆ | ★★★★★ | ★★★★☆ |
| 性能 | ★★★★★ | ★★★☆☆ | ★★★★☆ |
| 实施难度 | ★★★★★ | ★★☆☆☆ | ★★★☆☆ |
| 可维护性 | ★★★☆☆ | ★★★★☆ | ★★★★☆ |

**推荐方案**：
采用**动态属性模板+继承**。

实施要点：

1. **类目层级设计**：
   ```
   建议：不超过4级
   L1：大类（手机、服装、食品）
   L2：中类（智能手机、T恤、零食）
   L3：小类（iPhone、圆领T恤、膨化食品）
   L4：细分类（iPhone 15系列）
   ```

2. **属性校验**：
   ```java
   public void validateProduct(Product product, Category category) {
     // 1. 获取类目属性模板
     List<AttributeMeta> attrs = getAttributesByCategory(category);
     
     // 2. 检查必填属性
     for (AttributeMeta attr : attrs) {
       if (attr.isRequired() && !product.hasAttribute(attr.getCode())) {
         throw new ValidationException("缺少必填属性: " + attr.getName());
       }
     }
     
     // 3. 校验属性值
     for (ProductAttribute attr : product.getAttributes()) {
       AttributeMeta meta = getAttributeMeta(attr.getCode());
       meta.validate(attr.getValue()); // 类型、范围、枚举值校验
     }
   }
   ```

3. **属性搜索支持**：
   ```
   ES映射自动生成：
   {
     "mappings": {
       "properties": {
         "productId": {"type": "keyword"},
         "title": {"type": "text", "analyzer": "ik_max_word"},
         "category_id": {"type": "long"},
         "brand_id": {"type": "long"},
         // 动态属性
         "attrs": {
           "type": "nested",
           "properties": {
             "code": {"type": "keyword"},
             "value": {"type": "keyword"}
           }
         }
       }
     }
   }
   ```

4. **属性演进**：
   ```
   新增属性：
   1. 在attribute_meta表添加属性定义
   2. 关联到类目模板
   3. 存量商品渐进补齐（批量任务或人工）
   
   弃用属性：
   1. 标记为deprecated
   2. 新商品不展示该属性
   3. 老商品保留（不删除）
   ```

5. **多语言支持**：
   ```sql
   attribute_i18n（属性国际化）
   ├── attribute_id
   ├── locale（zh_CN/en_US）
   ├── name
   └── description
   ```

**延伸思考**：
1. 如何处理类目合并和拆分？
2. 属性过多时如何优化详情页加载性能？
3. 跨类目搜索时属性如何映射？

---

##### 🔧 题目5：商品图片的存储和CDN方案

**问题描述**：
电商平台商品图片数量巨大（百万级），每天上传图片数万张。如何设计图片存储和CDN方案，保证加载速度和成本可控？

**答案**：

**问题分析**：
图片存储的核心挑战：
1. 存储成本高（TB级数据）
2. 访问量大（详情页、列表页都需要图片）
3. 需要支持多种尺寸（缩略图、中图、大图）
4. 图片上传和审核流程

**方案一：自建存储+Nginx**

核心思想：
图片存储在自有服务器，通过Nginx提供静态服务。

设计：
```text
上传流程：
1. 应用服务器接收图片
2. 保存到本地磁盘：/data/images/{年}/{月}/{日}/{uuid}.jpg
3. 返回URL：http://img.example.com/2026/04/18/xxx.jpg

访问流程：
用户 → Nginx → 本地磁盘

多尺寸处理：
- 上传时生成多个尺寸
- 或使用Nginx image_filter模块动态缩放
```

优点：
- 完全可控
- 无外部依赖
- 成本可控

缺点：
- 带宽成本高
- 跨地域访问慢
- 需要自己做高可用
- 缺少图片处理能力

**方案二：对象存储OSS + CDN（推荐）**

核心思想：
图片存储在云厂商对象存储，通过CDN加速访问。

设计：
```text
上传流程：
1. 客户端 → 应用服务器申请上传凭证
2. 应用服务器 → OSS生成临时上传URL（STS）
3. 客户端 → 直传OSS
4. OSS → 回调应用服务器（上传成功）
5. 应用服务器 → 保存图片URL到数据库

访问流程：
用户 → CDN → OSS

图片处理：
URL参数控制：
- 缩放：?x-oss-process=image/resize,w_800
- 裁剪：?x-oss-process=image/crop,w_200,h_200
- 水印：?x-oss-process=image/watermark,text_xxx
- 格式转换：?x-oss-process=image/format,webp
```

优点：
- 性能好（CDN加速）
- 可靠性高（99.999999999%）
- 图片处理能力强
- 无需运维

缺点：
- 成本较高（按量付费）
- 被云厂商锁定
- 数据外传

**方案三：分层存储**

核心思想：
热图片存储在SSD+CDN，冷图片存储在归档存储。

设计：
```text
热存储（最近30天）：
- OSS标准存储 + CDN
- 访问速度快
- 成本高

冷存储（30天以上）：
- OSS归档存储
- 访问需要解冻（分钟级）
- 成本低（1/10）

智能分层：
- 根据访问频率自动迁移
- 热点商品图片永久在热存储
```

优点：
- 成本优化
- 性能保证

缺点：
- 归档解冻有延迟
- 分层逻辑复杂

**方案对比**：

| 维度 | 自建 | OSS+CDN | 分层存储 |
|------|------|---------|----------|
| 性能 | ★★★☆☆ | ★★★★★ | ★★★★☆ |
| 成本 | ★★★☆☆ | ★★★☆☆ | ★★★★☆ |
| 运维成本 | ★★☆☆☆ | ★★★★★ | ★★★☆☆ |
| 功能丰富度 | ★★☆☆☆ | ★★★★★ | ★★★★☆ |

**推荐方案**：
采用**OSS+CDN**。

实施要点：

1. **图片命名规范**：
   ```
   {bucket}/{年}/{月}/{日}/{category}/{uuid}.{ext}
   
   示例：
   product-images/2026/04/18/phone/550e8400-e29b-41d4-a716-446655440000.jpg
   ```

2. **多尺寸策略**：
   ```
   方案A：上传时生成（推荐）
   - 上传1张原图
   - 后台异步生成：缩略图(100x100)、小图(400x400)、中图(800x800)
   - 分别存储：{uuid}_thumb.jpg, {uuid}_small.jpg, {uuid}_medium.jpg
   
   方案B：访问时生成
   - 只存储原图
   - 通过OSS图片处理参数动态生成
   - URL：{url}?x-oss-process=image/resize,w_400
   ```

3. **CDN配置**：
   ```
   缓存策略：
   - 原图：缓存7天
   - 缩略图：缓存30天
   - 回源策略：304协商缓存
   
   防盗链：
   - Referer白名单
   - 签名URL（临时访问）
   - IP黑名单
   ```

4. **图片审核**：
   ```
   流程：
   1. 上传到临时bucket
   2. 触发审核（内容安全API）
   3. 审核通过 → 移动到正式bucket
   4. 审核不通过 → 标记为违规，删除
   
   审核内容：
   - 色情识别
   - 暴恐识别
   - 二维码识别
   - 文字OCR+敏感词
   ```

5. **性能优化**：
   ```
   图片格式：
   - 优先WebP（体积小30%）
   - 降级JPEG/PNG（老浏览器）
   
   懒加载：
   - 首屏图片优先加载
   - 下方图片懒加载
   - 占位图优化体验
   
   压缩：
   - JPEG质量80%（肉眼无感知）
   - PNG使用TinyPNG压缩
   ```

**延伸思考**：
1. 如何防止图片盗链？
2. 商家上传违规图片如何处理？
3. 图片存储成本如何优化？

---

##### 💡 题目6：虚拟商品vs实物商品的设计差异

**问题描述**：
实物商品需要物流配送，虚拟商品（如充值卡、会员）是即时发货。两者在系统设计上有哪些差异？

**答案**：

**问题分析**：
虚拟商品的核心差异：
1. 无需物流，履约方式不同
2. 库存是卡密池，不是物理库存
3. 发货是推送卡密，不是创建运单
4. 支持自动发货

**方案一：统一建模，类型区分**

核心思想：
实物和虚拟商品共用一套模型，通过类型字段区分。

设计：
```sql
product
├── product_id
├── product_type（PHYSICAL/VIRTUAL/SERVICE）
├── fulfillment_type（LOGISTICS/INSTANT/APPOINTMENT）
└── 其他通用字段

订单履约流程：
if (product_type == PHYSICAL) {
  创建运单 → 发货 → 签收
} else if (product_type == VIRTUAL) {
  分配卡密 → 推送用户 → 确认收货
} else if (product_type == SERVICE) {
  预约 → 服务 → 评价
}
```

优点：
- 模型统一，代码复用
- 易于扩展新类型
- 适合混合场景（一单既有实物又有虚拟）

缺点：
- 需要大量if/else判断
- 虚拟商品的特殊字段无法体现

**方案二：拆分建模，独立系统**

核心思想：
实物商品和虚拟商品拆分为两个系统。

设计：
```text
实物商品系统：
- product, sku（标准商品模型）
- order, order_item
- logistics（物流）

虚拟商品系统：
- virtual_product（虚拟商品）
  ├── card_type（充值卡类型）
  ├── face_value（面值）
  └── validity_period（有效期）
- card_pool / inventory_code_pool_XX（卡密 / 券码池）
  ├── card_no
  ├── card_pwd
  ├── status（AVAILABLE/BOOKING/SOLD/LOCKED/EXPIRED/INVALID）
  └── order_id
- virtual_order（虚拟订单）
```

优点：
- 模型清晰，职责分明
- 可针对性优化
- 团队独立

缺点：
- 系统重复（订单、支付）
- 混合订单难处理
- 用户体验割裂

**方案三：统一订单，差异化履约**

核心思想：
订单系统统一，履约环节根据商品类型路由到不同履约系统。

设计：
```text
订单系统（统一）：
- 统一的订单模型
- 统一的下单流程
- 统一的支付流程

履约路由：
if (orderItem.productType == PHYSICAL) {
  route to LogisticsService
} else if (orderItem.productType == VIRTUAL) {
  route to CardDistributionService
} else if (orderItem.productType == SERVICE) {
  route to AppointmentService
}

卡密分配服务：
1. 从卡密池分配未使用的卡密
2. 绑定到订单
3. 推送给用户（短信/App）
4. 标记卡密为已分配
```

优点：
- 订单模型统一
- 支持混合订单
- 履约解耦

缺点：
- 履约系统复杂度增加

**方案对比**：

| 维度 | 统一建模 | 拆分系统 | 统一订单+差异履约 |
|------|---------|---------|-------------------|
| 模型清晰度 | ★★★☆☆ | ★★★★★ | ★★★★☆ |
| 混合订单 | ★★★★★ | ★★☆☆☆ | ★★★★★ |
| 实施难度 | ★★★★☆ | ★★☆☆☆ | ★★★☆☆ |
| 用户体验 | ★★★★★ | ★★★☆☆ | ★★★★★ |

**推荐方案**：
采用**统一订单+差异化履约**。

实施要点：

1. **虚拟商品特殊字段**：
   ```sql
   virtual_product_ext
   ├── product_id
   ├── card_type（MOBILE_CHARGE/VIP_CARD/GAME_COIN）
   ├── face_value（面值）
   ├── validity_days（有效天数）
   └── auto_deliver（是否自动发货）
   ```

2. **卡密池设计**：
   ```sql
   card_pool
   ├── card_id
   ├── product_id
   ├── card_no
   ├── card_pwd（加密存储）
   ├── status（AVAILABLE/LOCKED/USED/INVALID）
   ├── locked_at（预占时间）
   ├── order_id
   ├── used_at
   └── expire_at
   
   预占机制：
   1. 下单时：status=LOCKED, locked_at=NOW()
   2. 支付成功：status=USED, order_id=xxx
   3. 超时未支付：定时任务释放（status=AVAILABLE）
   ```

3. **自动发货**：
   ```
   触发条件：
   - 支付成功事件
   - 商品类型=虚拟
   - auto_deliver=true
   
   发货流程：
   1. 从卡密池分配卡密
   2. 更新订单状态=COMPLETED
   3. 推送卡密给用户（短信/App推送）
   4. 记录发货日志
   ```

4. **卡密补货**：
   ```
   监控：
   - 可用卡密数量 < 1000 → 告警
   
   补货：
   - 供应商批量导入
   - 或系统自动生成（如游戏币）
   ```

5. **安全控制**：
   ```
   - 卡密加密存储（AES）
   - 卡密脱敏展示（只显示后4位）
   - 限制查询频率（防止爬虫）
   - 异常查询告警
   ```

**延伸思考**：
1. 如何防止卡密被盗刷？
2. 卡密分配失败如何处理？
3. 虚拟商品是否需要支持退款？

---

##### 📊 题目7：商品上架流程的工作流设计

**问题描述**：
商品从创建到上架需要经过多个环节（信息录入、图片上传、价格设置、审核）。请设计商品上架的工作流系统。

**答案**：

**问题分析**：
商品上架工作流的核心挑战：
1. 流程长，涉及多个环节和角色
2. 需要支持驳回和重新提交
3. 审核规则复杂（机审+人审）
4. 大批量商品上架性能

**方案一：状态机模式**

核心思想：
商品的状态流转按状态机管理。

状态定义：
```text
DRAFT（草稿）
→ PENDING_REVIEW（待审核）
  → APPROVED（审核通过）
    → ONLINE（已上架）
    → OFFLINE（已下架）
  → REJECTED（审核拒绝）→ DRAFT（重新编辑）
```

状态表：
```sql
product
├── product_id
├── status（当前状态）
├── review_status（审核状态：PENDING/PASS/REJECT）
└── reject_reason

product_status_history（状态流水）
├── product_id
├── from_status
├── to_status
├── operator
├── reason
└── created_at
```

优点：
- 简单直观
- 状态清晰

缺点：
- 复杂流程表达力不足
- 难以支持并行审核

**方案二：工作流引擎**

核心思想：
使用工作流引擎（如Activiti、Camunda）编排流程。

流程定义（BPMN）：
```text
开始 → 填写基本信息 → 上传图片 → 设置价格 
    → 提交审核 → 
      [机器审核] → 通过？
        → YES → [人工审核] → 通过？
          → YES → 上架成功
          → NO → 驳回
        → NO → 驳回
```

工作流表：
```sql
workflow_instance（流程实例）
├── instance_id
├── business_id（product_id）
├── workflow_def_id（流程定义ID）
├── current_node（当前节点）
├── status（RUNNING/COMPLETED/TERMINATED）
└── variables（流程变量，JSON）

workflow_task（任务）
├── task_id
├── instance_id
├── assignee（处理人）
├── status（PENDING/COMPLETED）
└── completed_at
```

优点：
- 流程可视化（BPMN图）
- 支持复杂流程（并行、分支、子流程）
- 易于调整流程

缺点：
- 引入工作流引擎，学习成本
- 重量级方案
- 调试困难

**方案三：轻量级流程引擎**

核心思想：
自己实现简化版工作流引擎，满足基本需求。

设计：
```java
// 流程定义（代码配置）
WorkflowDefinition productOnboard = new WorkflowDefinition()
  .addNode("FILL_INFO", new FillInfoNode())
  .addNode("UPLOAD_IMAGE", new UploadImageNode())
  .addNode("SET_PRICE", new SetPriceNode())
  .addNode("MACHINE_REVIEW", new MachineReviewNode())
  .addNode("MANUAL_REVIEW", new ManualReviewNode())
  .addTransition("FILL_INFO", "UPLOAD_IMAGE")
  .addTransition("UPLOAD_IMAGE", "SET_PRICE")
  .addTransition("SET_PRICE", "MACHINE_REVIEW")
  .addTransition("MACHINE_REVIEW", "MANUAL_REVIEW", condition="pass")
  .addTransition("MACHINE_REVIEW", "FILL_INFO", condition="reject")
  .addTransition("MANUAL_REVIEW", "ONLINE", condition="pass")
  .addTransition("MANUAL_REVIEW", "FILL_INFO", condition="reject");

// 流程执行引擎
public class WorkflowEngine {
  public void execute(String instanceId) {
    WorkflowInstance instance = getInstances(instanceId);
    Node currentNode = instance.getCurrentNode();
    
    // 执行当前节点
    NodeResult result = currentNode.execute(instance.getContext());
    
    // 根据结果流转到下一节点
    Node nextNode = getNextNode(currentNode, result);
    instance.setCurrentNode(nextNode);
    
    // 保存状态
    saveInstance(instance);
  }
}
```

优点：
- 轻量级，无外部依赖
- 代码即文档
- 易于调试和定制

缺点：
- 功能相对简单
- 不支持BPMN可视化
- 需要自己维护

**方案对比**：

| 维度 | 状态机 | 工作流引擎 | 轻量引擎 |
|------|--------|-----------|----------|
| 实施难度 | ★★★★★ | ★★☆☆☆ | ★★★★☆ |
| 流程表达力 | ★★☆☆☆ | ★★★★★ | ★★★★☆ |
| 维护成本 | ★★★★☆ | ★★★☆☆ | ★★★★☆ |
| 适用场景 | 简单流程 | 复杂流程 | 中等流程 |

**推荐方案**：
对于商品上架，推荐**轻量级流程引擎**。

实施要点：

1. **审核规则设计**：
   ```
   机器审核：
   - 图片审核（色情、暴恐）
   - 标题敏感词检测
   - 价格合理性检测（异常低价）
   - 类目属性完整性检测
   
   人工审核：
   - 机器审核不通过 → 必须人审
   - 高风险类目（药品、食品） → 必须人审
   - 新商家首批商品 → 必须人审
   - 其他商品 → 机审通过直接上架
   ```

2. **批量上架优化**：
   ```
   单个上架：
   - 提交 → 立即审核 → 立即上架
   
   批量上架：
   - 提交100个商品
   - 异步审核（队列）
   - 审核完成后批量回调
   - 生成审核报告
   ```

3. **驳回重审**：
   ```
   驳回原因分类：
   - 图片问题（重新上传图片即可）
   - 价格问题（重新设置价格）
   - 类目错误（重新选择类目，属性重填）
   
   重审流程：
   - 修改后自动重新提审
   - 或需要人工重新提交
   ```

4. **工作流监控**：
   ```
   指标：
   - 待审核商品数量
   - 平均审核时长
   - 审核通过率
   - 驳回原因分布
   
   告警：
   - 待审核积压 > 1000
   - 审核通过率 < 80%
   ```

**延伸思考**：
1. 如何设计商品的定时上架功能？
2. 批量上架如何保证事务性？
3. 审核规则如何动态配置？

---

##### 🔧 题目8：如何支持商品的多规格选择（颜色、尺码等）？

**问题描述**：
服装类商品有多个规格（颜色、尺码），用户需要先选择规格再下单。如何设计商品规格和SKU的选择逻辑？

**答案**：

**问题分析**：
多规格选择的核心挑战：
1. 规格组合爆炸（3个颜色×5个尺码=15个SKU）
2. 无效组合处理（某颜色没有某尺码）
3. 库存关联（每个SKU独立库存）
4. 价格差异（不同规格价格不同）

**方案一：预生成所有SKU**

核心思想：
商品创建时生成所有可能的规格组合。

设计：
```sql
spu（商品）
├── spu_id
├── title
└── spec_definitions（规格定义）
    {
      "color": ["黑色", "白色", "蓝色"],
      "size": ["S", "M", "L", "XL"]
    }

sku（商品SKU）
├── sku_id
├── spu_id
├── spec_values（规格取值）
    {"color": "黑色", "size": "M"}
├── price
├── stock
└── status（可售/售罄/下架）

生成逻辑：
笛卡尔积：3颜色 × 4尺码 = 12个SKU
```

前端逻辑：
```text
1. 用户选择颜色"黑色"
   → 查询：黑色有哪些尺码可选
   → 禁用无货尺码

2. 用户选择尺码"M"
   → 确定SKU：{color:黑色, size:M}
   → 显示价格、库存
   → 加入购物车（记录sku_id）
```

优点：
- 逻辑简单
- 查询性能好（直接查SKU表）
- 库存价格独立管理

缺点：
- SKU数量多（组合爆炸）
- 无效组合浪费存储
- 规格变更需要重新生成

**方案二：动态组合**

核心思想：
不预生成SKU，用户选择时动态计算。

设计：
```text
spu表：
只存储SPU和规格定义，不生成SKU

规格库存表：
spec_stock
├── spu_id
├── spec_hash（规格组合hash）
    MD5("color:黑色,size:M")
├── stock
└── price

查询逻辑：
1. 用户选择规格 → 计算spec_hash
2. 查询spec_stock表获取库存价格
3. 下单时记录spec_hash
```

优点：
- 灵活，规格可动态调整
- 不会产生无效SKU
- 节省存储

缺点：
- 查询复杂（需要计算hash）
- 订单记录不直观（spec_hash）
- 难以支持SKU级别的运营（如促销、限购）

**方案三：混合模式（主流+无效过滤）**

核心思想：
预生成SKU，但只生成有效组合。

设计：
```sql
sku_constraint（无效组合）
├── spu_id
├── constraint_type（DENY/ALLOW）
├── constraint_rule（JSON）
    {"color": "黑色", "size": "XL"}  // 黑色没有XL

SKU生成逻辑：
1. 计算笛卡尔积
2. 过滤无效组合（根据constraint规则）
3. 生成有效SKU

前端逻辑：
1. 查询所有有效的规格组合
2. 根据用户已选规格，计算可选项
3. 禁用无货或无效的选项
```

优点：
- 灵活性和性能兼顾
- 支持无效组合
- SKU数量合理

缺点：
- 需要维护约束规则
- 生成逻辑复杂

**方案对比**：

| 维度 | 预生成所有 | 动态组合 | 混合模式 |
|------|-----------|---------|----------|
| SKU数量 | 多 | 无 | 适中 |
| 查询性能 | ★★★★★ | ★★★☆☆ | ★★★★☆ |
| 灵活性 | ★★☆☆☆ | ★★★★★ | ★★★★☆ |
| 运营友好 | ★★★★★ | ★★☆☆☆ | ★★★★☆ |

**推荐方案**：
采用**混合模式（预生成+无效过滤）**。

实施要点：

1. **前端规格选择组件**：
   ```
   逻辑：
   1. 加载所有有效SKU
   2. 构建规格树
   3. 根据已选规格，计算可选项
   4. 禁用无货或无效选项
   
   示例（用户已选"黑色"）：
   可选尺码 = 筛选(所有SKU, color="黑色" && stock>0)
   禁用尺码 = 筛选(所有SKU, color="黑色" && stock=0)
   ```

2. **规格约束表达**：
   ```
   方案A：黑名单
   "不存在黑色XL"
   
   方案B：白名单
   "只有这些组合：黑色+M, 黑色+L, 白色+S, ..."
   
   推荐：黑名单（灵活）
   ```

3. **SKU图片**：
   ```
   商品主图：展示默认规格
   规格图：每个颜色独立图片
   
   用户选择颜色 → 切换主图
   ```

4. **性能优化**：
   ```
   缓存：
   - 缓存商品的所有SKU（减少查询）
   - 缓存规格树（减少计算）
   
   压缩：
   - 规格数据压缩传输
   ```

**延伸思考**：
1. 如何支持规格变更（新增颜色、下架尺码）？
2. 用户加购时记录SKU还是规格组合？
3. 如何优化规格选择的用户体验？

---

##### 💡 题目9：商品快照在订单中的应用

**问题描述**：
用户下单后，商家可能修改商品标题、价格、图片。为了避免纠纷，需要在订单中保存商品快照。请设计商品快照方案。

**答案**：

**问题分析**：
商品快照的核心挑战：
1. 快照内容：保存哪些字段
2. 存储成本：每个订单都存快照，数据量大
3. 快照时机：下单时还是支付时
4. 快照更新：商品变更后订单快照是否更新

**方案一：订单表冗余字段**

核心思想：
在订单明细表中冗余商品关键字段。

设计：
```sql
order_item
├── order_id
├── product_id
├── sku_id
├── product_title（快照）
├── product_image（快照）
├── price（快照）
├── quantity
└── total_amount
```

优点：
- 查询方便
- 无需JOIN

缺点：
- 字段冗余
- 快照内容有限
- 表结构膨胀

**方案二：独立快照表**

核心思想：
商品快照存储在独立表，订单引用快照ID。

设计：
```sql
product_snapshot
├── snapshot_id
├── product_id
├── sku_id
├── snapshot_data（JSON）
    {
      "title": "iPhone 15 Pro",
      "price": 7999,
      "images": ["url1", "url2"],
      "specs": {"color": "黑色", "storage": "256GB"},
      "brand": "Apple",
      "attributes": {...}
    }
├── content_hash（MD5，去重）
├── version
└── created_at

order_item
├── order_id
├── snapshot_id（引用快照）
├── quantity
└── total_amount
```

快照生成时机：
```text
时机1：用户下单时
- 优点：反映下单时的商品信息
- 缺点：未支付订单占用存储

时机2：用户支付时
- 优点：反映支付时的商品信息，更准确
- 缺点：支付时商品可能已下架

推荐：下单时生成，支付时校验
```

优点：
- 快照完整（可存储任意字段）
- 去重优化（相同快照共享）
- 订单表轻量

缺点：
- 需要JOIN查询
- 存储成本高

**方案三：按需快照+延迟生成**

核心思想：
下单时不生成快照，只有在需要时（如退货纠纷）才生成。

设计：
```text
order_item
├── product_id
├── sku_id
├── snapshot_id（初始为NULL）
└── snapshot_at（快照生成时间）

生成时机：
1. 用户申请退货
2. 商家纠纷
3. 定时任务（订单完成后30天生成快照）

生成逻辑：
1. 根据product_id查询当前商品信息
2. 生成快照（尽力而为）
3. 如果商品已删除，快照为空
```

优点：
- 存储成本低
- 按需生成

缺点：
- 延迟生成可能获取不到准确信息
- 商品删除后无法生成

**方案对比**：

| 维度 | 冗余字段 | 独立快照表 | 按需快照 |
|------|---------|-----------|----------|
| 快照完整性 | ★★☆☆☆ | ★★★★★ | ★★★☆☆ |
| 存储成本 | ★★★☆☆ | ★★☆☆☆ | ★★★★★ |
| 查询性能 | ★★★★★ | ★★★★☆ | ★★★☆☆ |
| 准确性 | ★★★★★ | ★★★★★ | ★★★☆☆ |

**推荐方案**：
采用**独立快照表+去重优化**。

实施要点：

1. **快照内容设计**：
   ```
   必须包含：
   - 商品标题、主图
   - SKU规格、价格
   - 品牌、类目
   
   可选包含：
   - 商品详情图（占用空间大）
   - 营销信息（优惠券、满减）
   - 服务承诺（七天无理由退货）
   ```

2. **快照去重**：
   ```
   生成流程：
   1. 计算快照内容的MD5: content_hash
   2. 查询是否已存在相同hash的快照
   3. 如果存在，复用snapshot_id
   4. 如果不存在，创建新快照
   
   收益：
   - 相同商品的订单共享快照
   - 存储成本降低50%+
   ```

3. **快照压缩**：
   ```
   JSON压缩：
   - 使用gzip压缩snapshot_data
   - 读取时解压
   
   字段裁剪：
   - 只保留关键字段
   - 详情图等大字段不保存
   ```

4. **快照过期清理**：
   ```
   策略：
   - 订单完成后保留2年（法律要求）
   - 2年后匿名化处理（删除用户信息，保留快照）
   - 5年后归档到对象存储
   ```

5. **快照版本化**：
   ```
   快照schema版本：
   V1: {title, price, image}
   V2: {title, price, images[], brand, specs}
   
   读取时兼容：
   if (snapshot.version == 1) {
     return convertV1ToV2(snapshot)
   }
   ```

**延伸思考**：
1. 商品快照如何支持营销信息（如"限时折扣"）？
2. 快照生成失败如何处理？
3. 如何设计快照的版本兼容？

---

##### 📊 题目10：设计商品推荐系统的架构

**问题描述**：
电商平台需要在详情页、列表页、首页展示个性化推荐商品。请设计商品推荐系统的架构。

**答案**：

**问题分析**：
推荐系统的核心挑战：
1. 推荐算法复杂（协同过滤、深度学习）
2. 实时性要求（用户行为实时影响推荐）
3. 冷启动问题（新用户、新商品）
4. 性能要求高（毫秒级响应）

**方案一：基于规则的推荐**

核心思想：
使用人工配置的规则进行推荐。

规则示例：
```text
规则1：看了还看
- 用户浏览商品A
- 推荐：浏览过A的用户还浏览了哪些商品

规则2：相似商品
- 用户浏览iPhone 15
- 推荐：同类目、相似价格的商品

规则3：热门商品
- 推荐：该类目下销量TOP 10

规则4：运营配置
- 推荐：运营手动配置的商品（大促主推）
```

优点：
- 实现简单
- 可控性强
- 无需算法团队

缺点：
- 推荐效果一般
- 不支持个性化
- 规则难以维护

**方案二：离线推荐+在线召回**

核心思想：
离线计算推荐结果，在线实时召回。

架构：
```text
离线计算（T+1）：
1. 收集用户行为数据（浏览、加购、购买）
2. 训练推荐模型（协同过滤、矩阵分解）
3. 计算用户-商品推荐矩阵
4. 存储到Redis：user:123:rec → [prod1, prod2, ...]

在线召回：
1. 用户请求推荐
2. 从Redis查询预计算结果
3. 过滤下架/无货商品
4. 返回推荐列表

实时反馈：
用户点击推荐 → 记录日志 → 下次离线计算时使用
```

优点：
- 支持复杂算法
- 性能好（在线只查询）
- 推荐效果好

缺点：
- 实时性差（T+1）
- 冷启动问题
- 存储成本高

**方案三：实时推荐（流式计算）**

核心思想：
使用流式计算（Flink）实时更新推荐结果。

架构：
```text
用户行为 → Kafka → Flink流式计算 → 更新Redis推荐结果

Flink计算逻辑：
1. 实时聚合用户行为（滑动窗口）
2. 更新用户画像（兴趣标签）
3. 实时计算推荐（基于规则或轻量模型）
4. 更新Redis

在线服务：
查询Redis获取实时推荐结果
```

优点：
- 实时性好（秒级）
- 支持个性化
- 反馈快

缺点：
- 架构复杂
- 成本高
- 算法受限（不能用复杂模型）

**方案对比**：

| 维度 | 规则推荐 | 离线+在线 | 实时推荐 |
|------|---------|-----------|----------|
| 推荐效果 | ★★☆☆☆ | ★★★★☆ | ★★★★★ |
| 实时性 | ★★★★★ | ★★☆☆☆ | ★★★★★ |
| 实施难度 | ★★★★★ | ★★★☆☆ | ★★☆☆☆ |
| 成本 | ★★★★★ | ★★★☆☆ | ★★☆☆☆ |

**推荐方案**：
采用**离线推荐+实时规则补充**的混合方案。

实施要点：

1. **推荐场景分类**：
   ```
   首页推荐：
   - 个性化推荐（基于用户画像）
   - 热门推荐（兜底）
   
   详情页推荐：
   - 看了还看（基于商品相似度）
   - 买了还买（基于订单关联）
   
   购物车推荐：
   - 凑单推荐（基于购物车商品关联）
   - 优惠推荐（基于满减规则）
   ```

2. **推荐召回链路**：
   ```
   第一层：个性化召回（离线计算）
   - 协同过滤召回
   - 内容召回（基于用户兴趣标签）
   
   第二层：规则召回（在线计算）
   - 热门商品
   - 运营配置
   
   第三层：排序
   - 点击率预估
   - 转化率预估
   - 业务规则调权（如新品扶持）
   
   第四层：过滤
   - 去重
   - 过滤下架/无货商品
   - 多样性（不全是同一类目）
   ```

3. **冷启动处理**：
   ```
   新用户：
   - 展示热门商品
   - 根据注册信息推断兴趣（地域、年龄）
   - 引导用户选择兴趣标签
   
   新商品：
   - 基于类目和属性推荐给相关用户
   - 运营人工推送给种子用户
   - 根据早期反馈调整推荐策略
   ```

4. **A/B测试**：
   ```
   实验：
   - 对照组：规则推荐
   - 实验组：算法推荐
   
   指标：
   - 点击率（CTR）
   - 转化率（CVR）
   - 人均订单金额
   ```

5. **监控指标**：
   ```
   业务指标：
   - 推荐位点击率
   - 推荐商品转化率
   - 推荐覆盖度（多少用户有推荐）
   
   技术指标：
   - 推荐响应时间
   - 推荐服务可用性
   - 离线计算任务成功率
   ```

**延伸思考**：
1. 如何评估推荐系统的效果？
2. 推荐系统如何防止马太效应（热门更热，冷门更冷）？
3. 如何保护用户隐私（不过度使用用户数据）？

---

##### 🔧 题目11：商品搜索的倒排索引设计

**问题描述**：
搜索引擎的核心是倒排索引。请说明电商商品搜索的倒排索引如何设计，包括分词、索引结构、查询优化等。

**答案**：

**问题分析**：
倒排索引的核心要点：
1. 分词策略（中文分词难点）
2. 索引字段选择（哪些字段需要索引）
3. 相关性打分（如何排序）
4. 性能优化（索引大小、查询速度）

**方案一：基于Elasticsearch标准分词**

核心思想：
使用ES内置的standard分词器。

配置：
```json
{
  "mappings": {
    "properties": {
      "title": {
        "type": "text",
        "analyzer": "standard"
      }
    }
  }
}

倒排索引示例：
商品标题："Apple iPhone 15 Pro 256GB 黑色"
分词结果：[Apple, iPhone, 15, Pro, 256GB, 黑色]

倒排索引：
Apple → [doc1, doc3, doc8]
iPhone → [doc1, doc2, doc3]
15 → [doc1, doc5]
Pro → [doc1, doc4]
```

优点：
- 实现简单
- 无需额外配置

缺点：
- 中文分词效果差
- 不支持同义词
- 相关性一般

**方案二：基于IK分词器（推荐）**

核心思想：
使用中文分词器（IK Analyzer），支持智能分词。

配置：
```json
{
  "mappings": {
    "properties": {
      "title": {
        "type": "text",
        "analyzer": "ik_max_word",      // 索引时：最细粒度分词
        "search_analyzer": "ik_smart"   // 搜索时：智能分词
      },
      "brand": {
        "type": "keyword"  // 不分词
      },
      "category": {
        "type": "keyword"
      },
      "price": {
        "type": "double"
      },
      "sales": {
        "type": "long"
      }
    }
  }
}

分词示例：
标题："小米手机13 Ultra 5G智能手机"
ik_max_word：[小米, 米手, 手机, 小米手机, 13, Ultra, 5G, 智能, 智能手机]
ik_smart：[小米, 手机, 13, Ultra, 5G, 智能手机]
```

优点：
- 中文分词准确
- 支持自定义词典
- 搜索效果好

缺点：
- 需要安装插件
- 词典需要维护

**方案三：多字段+权重**

核心思想：
对不同字段建立索引，搜索时设置不同权重。

配置：
```json
{
  "mappings": {
    "properties": {
      "title": {
        "type": "text",
        "analyzer": "ik_max_word",
        "boost": 3.0  // 标题权重最高
      },
      "brand": {
        "type": "keyword",
        "boost": 2.0  // 品牌权重次之
      },
      "category": {
        "type": "keyword",
        "boost": 1.5
      },
      "description": {
        "type": "text",
        "analyzer": "ik_max_word",
        "boost": 1.0  // 描述权重最低
      }
    }
  }
}

查询：
{
  "query": {
    "multi_match": {
      "query": "小米手机",
      "fields": ["title^3", "brand^2", "description"]
    }
  }
}
```

优点：
- 相关性更准确
- 可调整权重
- 支持多字段搜索

缺点：
- 查询复杂度增加
- 权重调优需要经验

**方案对比**：

| 维度 | 标准分词 | IK分词 | 多字段+权重 |
|------|---------|--------|-------------|
| 中文效果 | ★★☆☆☆ | ★★★★☆ | ★★★★★ |
| 实施难度 | ★★★★★ | ★★★★☆ | ★★★☆☆ |
| 相关性 | ★★★☆☆ | ★★★★☆ | ★★★★★ |
| 性能 | ★★★★☆ | ★★★★☆ | ★★★☆☆ |

**推荐方案**：
采用**IK分词+多字段权重**。

实施要点：

1. **自定义词典**：
   ```
   品牌词：小米、iPhone、华为
   型号词：13Ultra、15Pro、Mate60
   行业词：闪充、快充、护眼屏
   
   维护：
   - 定期更新词典
   - 新品牌/新词及时添加
   ```

2. **同义词处理**：
   ```json
   {
     "filter": {
       "synonym_filter": {
         "type": "synonym",
         "synonyms": [
           "手机,移动电话",
           "充电器,充电头",
           "iPhone,苹果手机"
         ]
       }
     }
   }
   ```

3. **拼音搜索**：
   ```
   支持拼音搜索：
   "xiaomi" → 小米
   "pingguo" → 苹果
   
   实现：
   - 使用pinyin分词插件
   - 或维护拼音映射表
   ```

4. **搜索建议（suggest）**：
   ```
   输入"xiao"  → 建议：[小米, 小天才, 小度]
   输入"iphone" → 建议：[iPhone 15, iPhone 14, iPhone 13]
   
   实现：
   - 使用ES的completion suggester
   - 基于前缀匹配
   ```

5. **性能优化**：
   ```
   索引优化：
   - 只索引需要搜索的字段
   - 使用doc_values减少内存占用
   - 定期合并段（segment merge）
   
   查询优化：
   - 结果分页（from+size < 10000）
   - 深度分页用scroll或search_after
   - 热门查询结果缓存
   ```

**延伸思考**：
1. 如何实现搜索纠错（"小米手及" → "小米手机"）？
2. 如何优化长尾查询的性能？
3. 搜索结果如何排序（相关性、销量、价格）？

---

##### 💡 题目12：如何处理商品数据的历史版本？

**问题描述**：
商品信息会不断变更（价格调整、标题修改、图片更换）。为了审计和纠纷处理，需要保留商品的历史版本。如何设计商品版本管理？

**答案**：

**问题分析**：
商品版本管理的核心挑战：
1. 版本数据量大（每次变更都存储）
2. 查询历史版本（某个时间点的商品信息）
3. 版本对比（对比两个版本的差异）
4. 存储成本

**方案一：全量版本存储**

核心思想：
每次变更都保存完整的商品数据。

设计：
```sql
product（当前版本）
├── product_id
├── title
├── price
├── version（当前版本号）
└── updated_at

product_history（历史版本）
├── history_id
├── product_id
├── version
├── title
├── price
├── changed_fields（变更字段）
├── operator（操作人）
└── created_at
```

查询历史：
```sql
-- 查询商品在2024-01-15的版本
SELECT * FROM product_history
WHERE product_id='123' 
  AND created_at <= '2024-01-15'
ORDER BY created_at DESC
LIMIT 1
```

优点：
- 查询简单
- 可完整恢复任意版本

缺点：
- 存储成本高（每次变更都全量存储）
- 字段冗余

**方案二：增量版本存储**

核心思想：
只保存变更的字段（diff）。

设计：
```sql
product_version
├── version_id
├── product_id
├── version_no
├── changed_fields（JSON）
    {
      "title": {"old": "iPhone 14", "new": "iPhone 15"},
      "price": {"old": 5999, "new": 7999}
    }
├── operator
└── created_at
```

恢复历史版本：
```text
1. 查询当前版本
2. 查询所有版本变更记录（按时间倒序）
3. 依次应用反向变更
4. 得到目标时间点的版本
```

优点：
- 存储成本低
- 可追踪变更内容

缺点：
- 查询复杂（需要计算）
- 版本恢复慢

**方案三：混合模式（快照+增量）**

核心思想：
定期保存全量快照，中间保存增量。

设计：
```text
product_snapshot（快照，每周保存）
├── snapshot_id
├── product_id
├── snapshot_data（JSON，完整数据）
├── snapshot_version
└── created_at

product_changelog（变更日志）
├── change_id
├── product_id
├── version
├── changed_fields（JSON）
└── created_at

查询策略：
1. 找到目标时间点之前最近的快照
2. 应用快照之后的变更日志
3. 得到目标版本
```

优点：
- 平衡存储和查询性能
- 快照恢复快
- 增量节省空间

缺点：
- 实现复杂度中等

**方案对比**：

| 维度 | 全量版本 | 增量版本 | 混合模式 |
|------|---------|---------|----------|
| 存储成本 | ★★☆☆☆ | ★★★★★ | ★★★★☆ |
| 查询性能 | ★★★★★ | ★★☆☆☆ | ★★★★☆ |
| 实施难度 | ★★★★★ | ★★★☆☆ | ★★★☆☆ |
| 审计能力 | ★★★★★ | ★★★★★ | ★★★★★ |

**推荐方案**：
对于电商系统，推荐**混合模式**。

实施要点：

1. **快照策略**：
   ```
   触发快照的时机：
   - 商品上架时（V1）
   - 每周日凌晨（定期快照）
   - 重大变更时（价格变动>20%）
   ```

2. **变更日志记录**：
   ```java
   public void updateProduct(Product product, ProductUpdate update) {
     Product old = getProduct(product.getId());
     
     // 1. 更新商品
     product.apply(update);
     product.setVersion(old.getVersion() + 1);
     productRepository.save(product);
     
     // 2. 记录变更日志
     ChangeLog log = new ChangeLog();
     log.setProductId(product.getId());
     log.setVersion(product.getVersion());
     log.setChangedFields(diff(old, product));  // 计算diff
     log.setOperator(getCurrentUser());
     changeLogRepository.save(log);
   }
   ```

3. **版本查询API**：
   ```
   GET /api/products/{productId}/versions
   → 返回所有版本列表
   
   GET /api/products/{productId}/versions/{version}
   → 返回指定版本数据
   
   GET /api/products/{productId}/diff?from=10&to=12
   → 返回版本差异
   ```

4. **存储优化**：
   ```
   - 快照使用压缩存储（gzip）
   - 超过1年的版本归档到对象存储
   - 变更日志保留2年（法律要求）
   ```

**延伸思考**：
1. 如何支持版本回滚（恢复到历史版本）？
2. 版本数据如何支持跨表查询（如关联订单）？
3. 大批量商品版本查询如何优化？

---

##### 📊 题目13：多租户场景下的商品数据隔离

**问题描述**：
在B2B2C平台中，多个商家共用一套系统。如何设计商品数据的租户隔离，保证数据安全和性能？

**答案**：

**问题分析**：
多租户隔离的核心挑战：
1. 数据隔离：商家A看不到商家B的商品
2. 性能隔离：商家A的流量不影响商家B
3. 成本优化：共享基础设施降低成本
4. 个性化：支持商家自定义配置

**方案一：独立数据库（物理隔离）**

核心思想：
每个租户独立数据库。

设计：
```text
租户A → 数据库A → product_a, order_a
租户B → 数据库B → product_b, order_b
租户C → 数据库C → product_c, order_c

路由逻辑：
public DataSource getDataSource(String tenantId) {
  return dataSourceMap.get(tenantId);
}
```

优点：
- 隔离性强（物理隔离）
- 性能互不影响
- 支持定制化schema
- 数据迁移方便

缺点：
- 成本高（每个租户一个数据库）
- 运维复杂（管理多个数据库）
- 跨租户查询困难

适用场景：
- 大租户（数据量大、QPS高）
- 对隔离要求极高

**方案二：共享数据库+tenant_id字段（逻辑隔离）**

核心思想：
所有租户共享一个数据库，通过tenant_id字段隔离。

设计：
```sql
product
├── product_id
├── tenant_id（租户ID）
├── title
├── price
└── ...
INDEX idx_tenant_product (tenant_id, product_id)

查询：
SELECT * FROM product 
WHERE tenant_id='tenant_001' AND product_id='123'
```

Row-Level Security（PostgreSQL）：
```sql
CREATE POLICY tenant_isolation ON product
  USING (tenant_id = current_setting('app.current_tenant')::text);

-- 应用层设置
SET app.current_tenant = 'tenant_001';
```

优点：
- 成本低（共享资源）
- 运维简单（一个数据库）
- 跨租户查询方便

缺点：
- 隔离性弱（逻辑隔离）
- 性能互相影响
- 数据量大时性能下降
- 误删风险（忘记加tenant_id条件）

适用场景：
- 小租户（数据量小、QPS低）
- 成本敏感

**方案三：分库分表（混合隔离）**

核心思想：
大租户独立数据库，小租户共享分片。

设计：
```text
大租户（VIP）：
tenant_001 → database_001
tenant_002 → database_002

小租户（普通）：
tenant_101, tenant_102, ... → database_shared_01
tenant_201, tenant_202, ... → database_shared_02

路由策略：
if (isVIPTenant(tenantId)) {
  return getDedicatedDataSource(tenantId);
} else {
  int shardId = hash(tenantId) % 8;
  return getSharedDataSource(shardId);
}
```

优点：
- 成本优化（大租户独享，小租户共享）
- 性能隔离（大租户独立）
- 灵活（可动态迁移）

缺点：
- 架构复杂
- 租户迁移成本

**方案对比**：

| 维度 | 独立数据库 | 共享+tenant_id | 混合隔离 |
|------|-----------|---------------|----------|
| 隔离性 | ★★★★★ | ★★☆☆☆ | ★★★★☆ |
| 成本 | ★★☆☆☆ | ★★★★★ | ★★★★☆ |
| 运维复杂度 | ★★☆☆☆ | ★★★★★ | ★★★☆☆ |
| 扩展性 | ★★★★★ | ★★★☆☆ | ★★★★☆ |

**推荐方案**：
采用**混合隔离（分库分表）**。

实施要点：

1. **租户分级**：
   ```
   VIP租户（月GMV>1000万）：
   - 独立数据库
   - 独立Redis
   - 独立ES索引
   
   普通租户：
   - 共享分片数据库
   - 共享Redis（按tenant_id前缀隔离）
   - 共享ES索引（按tenant_id过滤）
   ```

2. **数据源路由**：
   ```java
   @Aspect
   public class TenantDataSourceAspect {
     @Around("execution(* com.example..*Repository.*(..))")
     public Object route(ProceedingJoinPoint pjp) {
       String tenantId = TenantContext.get();
       DataSource ds = getDataSource(tenantId);
       // 切换数据源
       DynamicDataSourceHolder.set(ds);
       return pjp.proceed();
     }
   }
   ```

3. **租户升降级**：
   ```
   普通→VIP（升级）：
   1. 创建独立数据库
   2. 数据迁移（双写验证）
   3. 切换路由
   4. 清理旧数据
   
   VIP→普通（降级）：
   1. 迁移到共享分片
   2. 切换路由
   3. 删除独立数据库
   ```

4. **安全控制**：
   ```
   - 强制tenant_id过滤（ORM拦截器）
   - 禁止跨租户查询
   - API鉴权（JWT包含tenant_id）
   - 审计日志（记录租户操作）
   ```

**延伸思考**：
1. 如何防止误查询跨租户数据（ORM层面）？
2. 租户数据如何备份和恢复？
3. 如何支持租户级别的功能开关？

---

##### 🔧 题目14：商品导入的批量处理优化

**问题描述**：
商家需要批量导入商品（一次导入1000-10000个）。如何设计批量导入功能，保证性能和数据正确性？

**答案**：

**问题分析**：
批量导入的核心挑战：
1. 数据量大，处理时间长
2. 需要校验每个商品（格式、必填项、业务规则）
3. 部分成功部分失败如何处理
4. 导入进度如何实时反馈

**方案一：同步导入**

核心思想：
用户上传文件，服务端同步处理，处理完返回结果。

流程：
```text
1. 用户上传Excel/CSV文件
2. 服务端解析文件
3. 逐行校验和插入数据库
4. 返回导入结果（成功X条，失败Y条）
```

优点：
- 实现简单
- 用户立即知道结果

缺点：
- 同步处理，用户等待时间长
- 大文件可能超时
- 占用服务器资源

适用场景：
- 小批量（<1000条）
- 对实时性要求高

**方案二：异步导入+进度查询**

核心思想：
用户上传文件后立即返回，后台异步处理。

流程：
```text
1. 用户上传文件
2. 服务端：
   - 保存文件到OSS
   - 创建导入任务（状态：PENDING）
   - 返回任务ID
3. 后台Worker：
   - 异步处理导入任务
   - 更新任务进度
   - 完成后通知用户
4. 用户查询进度：
   GET /api/import-tasks/{taskId}
```

导入任务表：
```sql
import_task
├── task_id
├── tenant_id
├── file_url（OSS地址）
├── total_count（总数）
├── success_count（成功数）
├── fail_count（失败数）
├── status（PENDING/PROCESSING/SUCCESS/FAILED）
├── error_file_url（失败记录文件）
├── progress（进度百分比）
└── created_at

import_detail（导入明细，可选）
├── task_id
├── row_no（行号）
├── product_data（JSON）
├── status（SUCCESS/FAILED）
└── error_message
```

优点：
- 用户体验好（不用等待）
- 支持大批量
- 不占用Web线程

缺点：
- 实现复杂
- 需要进度查询接口

**方案三：流式导入+实时反馈**

核心思想：
使用WebSocket实时推送导入进度。

流程：
```text
1. 用户上传文件
2. 建立WebSocket连接
3. 服务端：
   - 边解析边处理
   - 每处理100条推送进度
   - 实时返回失败记录
4. 用户实时看到进度和错误
```

优点：
- 实时反馈
- 用户体验最好
- 可随时中断

缺点：
- 需要维护WebSocket连接
- 实现最复杂

**方案对比**：

| 维度 | 同步导入 | 异步导入 | 流式导入 |
|------|---------|---------|----------|
| 用户体验 | ★★☆☆☆ | ★★★★☆ | ★★★★★ |
| 支持规模 | ★★☆☆☆ | ★★★★★ | ★★★★☆ |
| 实施难度 | ★★★★★ | ★★★☆☆ | ★★☆☆☆ |
| 实时反馈 | ★★★★★ | ★★☆☆☆ | ★★★★★ |

**推荐方案**：
采用**异步导入+进度查询**。

实施要点：

1. **文件解析**：
   ```
   支持格式：
   - Excel（.xlsx）
   - CSV
   - JSON
   
   解析优化：
   - 流式解析（不一次加载全文件）
   - 分批处理（每100条一批）
   ```

2. **数据校验**：
   ```
   校验层级：
   L1：格式校验（必填字段、字段类型）
   L2：业务校验（价格合理性、类目有效性）
   L3：关联校验（品牌是否存在、图片URL是否有效）
   
   快速失败：
   - 格式错误直接返回，不处理后续数据
   ```

3. **事务处理**：
   ```
   方案A：全量事务
   - 全部成功才提交，任一失败全部回滚
   - 适合小批量、关联性强的数据
   
   方案B：分批事务（推荐）
   - 每100条一个事务
   - 部分失败不影响其他批次
   - 生成失败报告
   ```

4. **性能优化**：
   ```
   - 批量INSERT（100条一次）
   - 异步同步ES（不阻塞导入）
   - 限流（防止导入占用所有资源）
   - 分时段（凌晨处理大批量）
   ```

5. **失败处理**：
   ```
   失败记录：
   - 生成Excel文件，标注失败原因
   - 用户下载修改后重新导入
   
   部分成功：
   - 成功的商品已入库
   - 失败的记录在error_file中
   ```

**延伸思考**：
1. 如何支持导入任务的取消？
2. 导入过程中商品数据变更如何处理？
3. 如何设计商品导入的幂等性？

---

##### 💡 题目15：商品审核流程的设计

**问题描述**：
商家上传的商品需要经过审核才能上架（防止违规商品）。请设计商品审核系统，包括机审和人审。

**答案**：

**问题分析**：
商品审核的核心挑战：
1. 审核效率：大量商品等待审核
2. 审核准确性：机审误报，人审成本高
3. 审核优先级：重点类目优先审核
4. 申诉流程：商家对审核结果不满

**方案一：纯人工审核**

核心思想：
所有商品都由审核人员人工审核。

流程：
```text
1. 商家提交商品
2. 进入审核队列
3. 审核员登录审核后台
4. 逐个审核（通过/拒绝）
5. 通过的商品上架
```

优点：
- 准确性高
- 实现简单

缺点：
- 效率低
- 人力成本高
- 审核周期长

适用场景：
- 商品量少（每天<100个）
- 高风险类目（药品）

**方案二：机审+人审（推荐）**

核心思想：
机器审核过滤大部分，人工审核复杂case。

流程：
```text
商品提交 
→ 机器审核
  → 通过（80%）→ 直接上架
  → 不确定（15%）→ 人工审核
  → 拒绝（5%）→ 直接拒绝

机器审核规则：
1. 图片审核：
   - 调用内容安全API
   - 检测色情、暴恐、二维码
   - 置信度 > 0.9 → 拒绝
   - 置信度 0.7-0.9 → 转人审
   - 置信度 < 0.7 → 通过

2. 文本审核：
   - 标题敏感词检测
   - 虚假宣传检测（"最好"、"第一"）
   - 医疗广告检测

3. 价格审核：
   - 异常低价（低于市场价50%）
   - 异常高价（高于市场价200%）

4. 类目审核：
   - 类目与商品不匹配
   - 必填属性缺失
```

人工审核：
```text
审核任务分配：
- 按类目分配（服装审核员、3C审核员）
- 按优先级（大商家优先、付费商家优先）
- 负载均衡（平均分配）

审核操作：
- 通过：商品上架
- 拒绝：填写拒绝原因（类目错误、图片违规、价格虚高）
- 待定：标记问题，转高级审核员
```

优点：
- 效率高（机审处理80%）
- 成本可控
- 准确性较好

缺点：
- 需要维护审核规则
- 机审误报需要人工校正

**方案三：智能审核（AI审核）**

核心思想：
使用机器学习模型进行审核。

模型训练：
```text
训练数据：
- 正样本：审核通过的商品
- 负样本：审核拒绝的商品

特征工程：
- 文本特征：标题、描述的词频、TF-IDF
- 图片特征：图片分类、OCR文字
- 商家特征：店铺等级、历史通过率
- 类目特征：类目风险等级

模型：
- LR、GBDT、Deep Learning

输出：
- 通过概率：0.9 → 直接通过
- 拒绝概率：0.8 → 直接拒绝
- 中间态：0.5-0.8 → 人工审核
```

优点：
- 准确率高（持续学习）
- 自动化程度高
- 可处理复杂case

缺点：
- 需要算法团队
- 需要大量训练数据
- 模型维护成本高

**方案对比**：

| 维度 | 纯人审 | 机审+人审 | AI审核 |
|------|--------|-----------|--------|
| 审核效率 | ★★☆☆☆ | ★★★★☆ | ★★★★★ |
| 准确率 | ★★★★★ | ★★★★☆ | ★★★★★ |
| 成本 | ★★☆☆☆ | ★★★★☆ | ★★★☆☆ |
| 实施难度 | ★★★★★ | ★★★★☆ | ★★☆☆☆ |

**推荐方案**：
采用**机审+人审**，逐步引入AI审核。

实施要点：

1. **审核规则配置化**：
   ```
   审核规则表：
   review_rule
   ├── rule_id
   ├── rule_name
   ├── rule_type（IMAGE/TEXT/PRICE/CATEGORY）
   ├── rule_config（JSON）
   ├── severity（HIGH/MEDIUM/LOW）
   ├── action（REJECT/MANUAL_REVIEW/PASS）
   └── enabled
   
   示例规则：
   {
     "rule_name": "敏感词检测",
     "keywords": ["假货", "高仿", ...],
     "action": "REJECT"
   }
   ```

2. **审核任务队列**：
   ```
   优先级队列：
   P0：付费商家、大商家
   P1：普通商家
   P2：新商家
   
   分配策略：
   - P0优先分配
   - 同优先级按提交时间
   - 负载均衡（每个审核员任务量相当）
   ```

3. **审核SLA**：
   ```
   目标：
   - 机审：5秒内完成
   - 人审：2小时内完成（工作时间）
   
   超时告警：
   - 待审核任务积压 > 500
   - 人审超时 > 50个
   ```

4. **申诉流程**：
   ```
   商家不满审核结果：
   1. 点击"申诉"
   2. 填写申诉理由
   3. 转高级审核员复审
   4. 复审结果通知商家
   ```

**延伸思考**：
1. 如何设计审核人员的绩效考核？
2. 机审规则如何动态调整（根据审核质量）？
3. 如何防止商家恶意提交违规商品？

---

---

### 40.1.2 库存系统（17题）

##### 🔧 题目0扩展：库存是怎么创建出来的？

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

##### 🔧 题目0扩展B：库存如何和商品供给运营平台、商品生命周期联动？

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

##### 📊 题目1：设计防止库存超卖的方案

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

##### 🔧 题目2：如何设计分布式库存系统？

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

##### 💡 题目3：大促场景下的库存预热和削峰方案

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

##### 📊 题目4：库存预占与释放的设计

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

##### 🔧 题目5：如何设计库存的分级管理（前台可售vs仓库实际）？

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

##### 💡 题目6：库存扣减失败的补偿机制

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

##### 📊 题目7：设计库存盘点系统

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

##### 🔧 题目8：如何处理库存的并发更新？

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

##### 💡 题目9：虚拟库存vs实物库存的差异

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

##### 📊 题目10：多仓库场景下的库存分配策略

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

##### 🔧 题目11：如何设计库存安全水位和补货机制？

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

##### 💡 题目12：库存快照在订单中的应用

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

##### 📊 题目13：库存的实时性vs一致性权衡

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

##### 🔧 题目14：库存回滚机制的设计

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

##### 💡 题目15：跨境电商的库存管理（多国库存）

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

---

### 40.1.3 营销与计价系统（10题）

##### 📊 题目1：设计支持多种促销规则的价格计算引擎

**问题描述**：
电商平台有多种促销（满减、折扣、优惠券、满赠、阶梯价），用户下单时需要计算最终价格。如何设计灵活的价格计算引擎？

**答案**：

**问题分析**：
价格计算的核心挑战：
1. 规则类型多（满减、折扣、优惠券、积分抵扣）
2. 规则可组合（同时使用多种优惠）
3. 优先级和互斥（有些优惠不能同时用）
4. 实时计算性能

**方案一：硬编码规则**

核心思想：
在代码中直接编写每种促销规则的计算逻辑。

实现：
```java
public BigDecimal calculatePrice(Order order) {
  BigDecimal price = order.getOriginalPrice();
  
  // 应用满减
  if (order.getTotal() >= 200) {
    price = price.subtract(new BigDecimal("30"));
  }
  
  // 应用折扣
  if (order.hasDiscount()) {
    price = price.multiply(new BigDecimal("0.9"));
  }
  
  // 应用优惠券
  if (order.hasCoupon()) {
    price = price.subtract(order.getCouponAmount());
  }
  
  return price;
}
```

优点：
- 实现简单
- 性能好

缺点：
- 不灵活（新增规则需要改代码）
- 难维护
- 运营无法自主配置

适用场景：
- 规则简单且固定
- 小型电商

**方案二：规则引擎（推荐）**

核心思想：
将促销规则配置化，使用规则引擎动态执行。

设计：
```sql
promotion_rule（促销规则表）
├── rule_id
├── rule_name
├── rule_type（DISCOUNT/FULL_REDUCE/COUPON/GIFT/TIER_PRICE）
├── rule_config（JSON）
    {
      "type": "FULL_REDUCE",
      "threshold": 200,
      "reduce": 30,
      "priority": 10,
      "exclusive": false
    }
├── begin_time
├── end_time
├── priority（优先级，数字越小越优先）
├── exclusive（是否与其他规则互斥）
└── status

规则执行引擎：
public class PriceCalculator {
  public BigDecimal calculate(Order order) {
    // 1. 加载适用的规则
    List<Rule> rules = ruleEngine.getApplicableRules(order);
    
    // 2. 按优先级排序
    rules.sort(Comparator.comparing(Rule::getPriority));
    
    // 3. 依次应用规则
    BigDecimal finalPrice = order.getOriginalPrice();
    for (Rule rule : rules) {
      if (rule.isApplicable(order)) {
        finalPrice = rule.apply(finalPrice, order);
      }
    }
    
    return finalPrice;
  }
}
```

规则类型示例：
```text
满减规则：
{
  "type": "FULL_REDUCE",
  "threshold": 200,  // 满200
  "reduce": 30       // 减30
}

折扣规则：
{
  "type": "DISCOUNT",
  "rate": 0.85       // 8.5折
}

阶梯价：
{
  "type": "TIER_PRICE",
  "tiers": [
    {"quantity": 1, "price": 100},
    {"quantity": 10, "price": 90},
    {"quantity": 100, "price": 80}
  ]
}

满赠规则：
{
  "type": "GIFT",
  "threshold": 300,
  "giftSkuId": "gift_001"
}
```

优点：
- 灵活可配置
- 运营自主管理
- 易于扩展新规则
- 支持复杂组合

缺点：
- 实现复杂
- 性能略低于硬编码

**方案三：脚本引擎（Groovy/JavaScript）**

核心思想：
将规则写成脚本，动态加载执行。

设计：
```sql
promotion_script
├── script_id
├── script_name
├── script_content（Groovy脚本）
    """
    if (order.total >= 200) {
      return order.total - 30
    }
    return order.total
    """
├── priority
└── ...

执行：
public BigDecimal calculate(Order order) {
  for (Script script : scripts) {
    BigDecimal price = groovyEngine.eval(script, order);
    order.setPrice(price);
  }
  return order.getPrice();
}
```

优点：
- 极致灵活（可写任意逻辑）
- 无需发布代码

缺点：
- 安全风险（脚本注入）
- 调试困难
- 性能开销大

**方案对比**：

| 方案 | 灵活性 | 性能 | 运营友好 | 安全性 | 实施难度 |
|------|--------|------|---------|--------|---------|
| 硬编码 | ★★☆☆☆ | ★★★★★ | ★☆☆☆☆ | ★★★★★ | ★★★★★ |
| 规则引擎 | ★★★★☆ | ★★★★☆ | ★★★★★ | ★★★★☆ | ★★★☆☆ |
| 脚本引擎 | ★★★★★ | ★★★☆☆ | ★★★☆☆ | ★★☆☆☆ | ★★☆☆☆ |

**推荐方案**：
采用**规则引擎**。

实施要点：

1. **规则抽象**：
   ```java
   public interface PromotionRule {
     // 规则是否适用
     boolean isApplicable(Order order);
     
     // 应用规则，返回新价格
     BigDecimal apply(BigDecimal currentPrice, Order order);
     
     // 规则优先级
     int getPriority();
     
     // 是否与其他规则互斥
     boolean isExclusive();
   }
   
   // 满减规则实现
   public class FullReduceRule implements PromotionRule {
     private BigDecimal threshold;
     private BigDecimal reduceAmount;
     
     public boolean isApplicable(Order order) {
       return order.getTotal().compareTo(threshold) >= 0;
     }
     
     public BigDecimal apply(BigDecimal currentPrice, Order order) {
       return currentPrice.subtract(reduceAmount);
     }
   }
   ```

2. **规则组合策略**：
   ```
   互斥规则：
   - 满减和折扣互斥（选优惠力度大的）
   - 用户只能使用一张优惠券
   
   可叠加规则：
   - 满减 + 积分抵扣
   - 会员折扣 + 优惠券
   
   执行顺序：
   1. 商品级促销（商品折扣）
   2. 订单级促销（满减）
   3. 用户级促销（会员折扣）
   4. 优惠券
   5. 积分抵扣
   ```

3. **价格明细**：
   ```
   原价：¥500
   - 商品折扣：-¥50（9折）
   - 满减优惠：-¥30（满200减30）
   - 会员折扣：-¥42（额外9折）
   - 优惠券：-¥20
   = 实付：¥358
   
   用户可见每项优惠的金额
   ```

4. **性能优化**：
   ```
   规则缓存：
   - 缓存活跃的促销规则（Redis）
   - TTL 5分钟
   - 规则变更时主动刷新
   
   批量计算：
   - 购物车多商品批量计算
   - 减少数据库查询
   ```

5. **试算API**：
   ```
   POST /api/price/calculate
   {
     "items": [
       {"skuId": "123", "quantity": 2},
       {"skuId": "456", "quantity": 1}
     ],
     "couponCode": "SUMMER20",
     "usePoints": 100
   }
   
   响应：
   {
     "originalPrice": 500,
     "discounts": [
       {"type": "FULL_REDUCE", "amount": 30},
       {"type": "COUPON", "amount": 20}
     ],
     "finalPrice": 450
   }
   ```

**延伸思考**：
1. 如何设计促销规则的AB测试？
2. 多种促销组合时如何选择最优组合？
3. 促销规则变更如何保证已下单的订单价格不变？

---

##### 🔧 题目2：优惠券系统的设计

**问题描述**：
电商平台需要支持优惠券（满减券、折扣券、品类券）。如何设计优惠券系统，包括发放、使用、核销？

**答案**：

**问题分析**：
优惠券的核心要素：
1. 发放方式（批量发放、用户领取、定向发放）
2. 使用规则（满减、折扣、品类限制、商品限制）
3. 并发领取（秒杀券，1万人抢100张）
4. 防刷机制（防止用户重复领取）

**方案一：简单优惠券**

核心思想：
优惠券模板+用户优惠券实例。

设计：
```sql
coupon_template（优惠券模板）
├── template_id
├── name
├── coupon_type（FULL_REDUCE/DISCOUNT/CASH）
├── discount_amount（满减金额）
├── discount_rate（折扣率，如0.9）
├── threshold（使用门槛，如满200可用）
├── total_count（总发行量）
├── used_count（已使用数量）
├── begin_time
├── end_time
└── status

user_coupon（用户优惠券）
├── coupon_id
├── template_id
├── user_id
├── coupon_code（券码）
├── status（UNUSED/USED/EXPIRED）
├── used_order_id
├── received_at
├── used_at
└── expire_at
```

优点：
- 实现简单
- 易于理解

缺点：
- 功能单一
- 不支持复杂规则

**方案二：规则化优惠券（推荐）**

核心思想：
优惠券支持丰富的使用规则和发放规则。

设计：
```sql
coupon_template
├── template_id
├── name
├── coupon_type
├── discount_config（JSON）
    {
      "type": "FULL_REDUCE",
      "threshold": 200,
      "amount": 30
    }
├── usage_rule（JSON）
    {
      "validCategories": [1, 2, 3],  // 限定品类
      "validSkus": ["sku1", "sku2"],  // 限定商品
      "maxDiscountPerOrder": 50,      // 单笔最高优惠
      "excludeBrands": [10, 20]       // 排除品牌
    }
├── receive_rule（JSON）
    {
      "maxReceivePerUser": 1,         // 每人限领1张
      "newUserOnly": false,            // 是否新用户专享
      "memberLevelRequired": "VIP"    // 会员等级要求
    }
├── total_count
├── received_count
├── used_count
└── ...

user_coupon
├── coupon_id
├── template_id
├── user_id
├── status
├── lock_order_id（预占：锁定到某订单）
├── locked_at
└── ...
```

优点：
- 规则灵活
- 支持复杂场景
- 运营可配置

缺点：
- 实现复杂度高

**方案三：优惠券码模式**

核心思想：
预生成优惠券码，用户输入券码兑换。

设计：
```sql
coupon_code
├── code（券码，如SUMMER2024）
├── template_id
├── status（AVAILABLE/USED/EXPIRED）
├── user_id（已兑换用户）
├── used_at
└── expire_at

使用流程：
1. 运营批量生成券码
2. 用户输入券码兑换
3. 绑定到user_coupon
4. 下单时使用
```

优点：
- 支持券码分享
- 灵活发放（短信、广告）

缺点：
- 券码可能被盗用
- 需要生成大量券码

**方案对比**：

| 方案 | 灵活性 | 并发性能 | 防刷能力 | 实施难度 |
|------|--------|---------|---------|---------|
| 简单券 | ★★☆☆☆ | ★★★★☆ | ★★★☆☆ | ★★★★★ |
| 规则券 | ★★★★★ | ★★★★☆ | ★★★★☆ | ★★★☆☆ |
| 券码模式 | ★★★☆☆ | ★★★★★ | ★★☆☆☆ | ★★★☆☆ |

**推荐方案**：
采用**规则化优惠券**。

实施要点：

1. **领券流程**：
   ```
   用户点击"领取"：
   1. 检查用户是否已领取（防重复）
   2. 检查是否满足领取条件（新用户、会员等级）
   3. 检查库存（received_count < total_count）
   4. 扣减库存（乐观锁）
   5. 创建user_coupon记录
   
   并发控制：
   UPDATE coupon_template 
   SET received_count = received_count + 1
   WHERE template_id = ? 
     AND received_count < total_count
     AND version = ?;
   
   if (affected_rows == 0) {
     throw new CouponSoldOutException();
   }
   ```

2. **用券流程**：
   ```
   下单时使用优惠券：
   1. 检查优惠券是否属于当前用户
   2. 检查优惠券状态（UNUSED）
   3. 检查是否过期
   4. 检查订单是否满足使用条件（品类、金额）
   5. 锁定优惠券（防止重复使用）
   6. 计算优惠金额
   
   支付成功：
   - 核销优惠券（status=USED）
   
   订单取消：
   - 释放优惠券（status=UNUSED, lock_order_id=NULL）
   ```

3. **防刷策略**：
   ```
   策略1：用户限制
   - 每人限领1张
   - 同一手机号/设备ID限领
   
   策略2：行为检测
   - 短时间多次领取 → 拉黑
   - 领取后不使用 → 降低权重
   
   策略3：风控
   - 新注册用户限制
   - 异常IP拦截
   ```

4. **券叠加规则**：
   ```
   规则：
   - 单笔订单最多使用1张优惠券
   - 优惠券和满减活动可叠加
   - 优惠券和积分抵扣可叠加
   
   选券策略：
   - 自动选择优惠最大的券
   - 或用户手动选择
   ```

5. **券过期处理**：
   ```
   定时任务（每天凌晨）：
   1. 扫描即将过期的券（expire_at < NOW() + 3天）
   2. 发送提醒通知（App推送、短信）
   3. 扫描已过期的券
   4. 状态更新为EXPIRED
   ```

**延伸思考**：
1. 如何设计优惠券的转赠功能？
2. 优惠券如何支持多次使用（如月卡券）？
3. 如何设计优惠券的效果分析（发放ROI）？

---

##### 💡 题目3：阶梯价和批发价的设计

**问题描述**：
电商平台支持批发场景，购买数量越多价格越低（如买1件100元，买10件90元，买100件80元）。如何设计阶梯价系统？

**答案**：

**问题分析**：
阶梯价的核心要素：
1. 阶梯定义（数量区间和对应价格）
2. 混合SKU计算（多个商品如何累计数量）
3. 拆单问题（阶梯内和阶梯外商品分开发货）
4. 实时计算性能

**方案一：SKU级阶梯价**

核心思想：
每个SKU独立设置阶梯价，不同SKU不累计。

设计：
```sql
sku_tier_price
├── sku_id
├── tier_level（阶梯级别：1,2,3...）
├── min_quantity（最小数量）
├── max_quantity（最大数量，NULL表示无上限）
├── price
└── ...

示例数据：
SKU: iPhone15
tier_1: 1-9件, ¥7999
tier_2: 10-99件, ¥7500
tier_3: 100+件, ¥7000

价格计算：
if (quantity >= 100) {
  return 7000 * quantity;
} else if (quantity >= 10) {
  return 7500 * quantity;
} else {
  return 7999 * quantity;
}
```

优点：
- 简单直观
- 计算快速

缺点：
- 不支持跨SKU累计
- 批发商体验差（买不同商品无法享受折扣）

**方案二：品类级阶梯价**

核心思想：
同一品类的商品数量累计，达到阶梯享受折扣。

设计：
```sql
category_tier_price
├── category_id
├── tier_level
├── min_quantity
├── discount_rate（折扣率）
└── ...

示例：
手机品类阶梯折扣：
tier_1: 1-9件, 无折扣
tier_2: 10-99件, 95折
tier_3: 100+件, 90折

计算：
用户购买：
- iPhone15: 5件 × ¥7999
- 小米14: 6件 × ¥3999
- 总数量：11件（属于tier_2）
- 享受95折

最终价格：
(5 × 7999 + 6 × 3999) × 0.95
```

优点：
- 支持跨SKU累计
- 批发商友好

缺点：
- 品类定义需要清晰
- 计算复杂

**方案三：订单级阶梯价（推荐）**

核心思想：
按订单总金额或总件数，应用阶梯折扣。

设计：
```sql
order_tier_price
├── tier_id
├── tier_type（BY_QUANTITY/BY_AMOUNT）
├── min_value（最小值）
├── max_value
├── discount_type（RATE/AMOUNT）
├── discount_value
└── ...

示例1：按数量
tier_1: 1-9件, 无折扣
tier_2: 10-49件, 95折
tier_3: 50+件, 90折

示例2：按金额
tier_1: <¥1000, 无折扣
tier_2: ¥1000-¥5000, 减¥100
tier_3: >¥5000, 减¥500
```

优点：
- 灵活
- 适用多种场景
- 计算简单

缺点：
- 需要明确阶梯规则

**方案对比**：

| 方案 | 灵活性 | 批发友好 | 计算复杂度 | 适用场景 |
|------|--------|---------|-----------|----------|
| SKU级 | ★★☆☆☆ | ★★☆☆☆ | ★★★★★ | 零售 |
| 品类级 | ★★★★☆ | ★★★★☆ | ★★★☆☆ | 批发 |
| 订单级 | ★★★★★ | ★★★★★ | ★★★★☆ | 混合 |

**推荐方案**：
采用**订单级阶梯价**。

实施要点：

1. **阶梯计算引擎**：
   ```java
   public BigDecimal calculateTierPrice(Order order) {
     // 1. 计算订单总量/总额
     int totalQuantity = order.getTotalQuantity();
     BigDecimal totalAmount = order.getTotalAmount();
     
     // 2. 查找匹配的阶梯
     TierPrice tier = tierPriceService.findMatchingTier(
       totalQuantity, totalAmount
     );
     
     // 3. 应用折扣
     if (tier.getDiscountType() == RATE) {
       return totalAmount.multiply(tier.getDiscountRate());
     } else {
       return totalAmount.subtract(tier.getDiscountAmount());
     }
   }
   ```

2. **实时试算**：
   ```
   购物车实时显示：
   - 当前数量：8件
   - 当前价格：¥1000
   - 提示："再买2件，享受95折，可省¥50"
   
   动态提示：
   引导用户凑单，提高客单价
   ```

3. **拆单策略**：
   ```
   场景：用户购买120件商品
   - 100件享受阶梯价（¥80/件）
   - 20件普通价（¥100/件）
   
   方案A：不拆单
   - 所有商品按最高阶梯价
   - 用户体验好
   
   方案B：拆单
   - 100件一单，20件一单
   - 复杂，不推荐
   ```

4. **会员叠加**：
   ```
   规则：
   - 阶梯价和会员折扣可叠加
   - 先应用阶梯价，再应用会员折扣
   
   示例：
   原价：¥10000
   阶梯价（95折）：¥9500
   会员折扣（98折）：¥9310
   ```

5. **报表分析**：
   ```
   阶梯价效果分析：
   - 各阶梯成交订单数
   - 平均客单价提升
   - 转化率（凑单率）
   ```

**延伸思考**：
1. 阶梯价如何与优惠券组合？
2. 用户退货部分商品如何重新计算价格？
3. 大促期间阶梯价如何调整？

---

##### 📊 题目4：会员等级和积分体系的设计

**问题描述**：
电商平台有会员体系（普通、银卡、金卡、钻石），不同等级享受不同权益（折扣、包邮、专属客服）。如何设计会员和积分系统？

**答案**：

**问题分析**：
会员体系的核心要素：
1. 等级划分（如何升降级）
2. 权益设计（不同等级的差异化权益）
3. 积分规则（获取、消耗、过期）
4. 成长值体系（区分消费积分和成长值）

**方案一：简单会员（单一积分）**

核心思想：
只有积分，达到一定积分自动升级。

设计：
```sql
member
├── user_id
├── member_level（NORMAL/SILVER/GOLD/DIAMOND）
├── points（积分）
├── total_points（累计积分，用于升级）
└── ...

升级规则：
- 累计积分 >= 10000 → 钻石
- 累计积分 >= 5000 → 金卡
- 累计积分 >= 1000 → 银卡
```

优点：
- 实现简单
- 易于理解

缺点：
- 权益单一
- 无法区分消费和成长

**方案二：双轨制（积分+成长值，推荐）**

核心思想：
积分用于消费抵扣，成长值用于等级提升。

设计：
```sql
member
├── user_id
├── member_level
├── points（可消费积分）
├── growth_value（成长值，只增不减）
├── upgrade_time（升级时间）
├── downgrade_time（预计降级时间）
└── ...

member_level_config
├── level
├── min_growth_value（最低成长值）
├── benefits（JSON，权益配置）
    {
      "discount": 0.95,           // 95折
      "freeShipping": true,       // 包邮
      "pointsRate": 1.2,          // 积分倍率
      "birthdayCoupon": 50,       // 生日券
      "exclusiveService": true    // 专属客服
    }
└── ...

积分规则：
point_rule
├── rule_id
├── action（ORDER/CHECKIN/SHARE/REVIEW）
├── points_reward
├── growth_reward
└── ...

示例：
- 购物：每消费1元获得1积分 + 1成长值
- 签到：每天签到获得5积分 + 0成长值
- 分享：每次分享获得10积分 + 0成长值
- 评价：每次评价获得20积分 + 5成长值
```

优点：
- 积分和等级分离，科学
- 防止用户消费积分后降级
- 权益丰富

缺点：
- 复杂度高

**方案三：付费会员（Prime模式）**

核心思想：
用户付费购买会员资格，享受权益。

设计：
```sql
member_subscription
├── user_id
├── plan_type（MONTH/YEAR）
├── status（ACTIVE/EXPIRED/CANCELLED）
├── begin_time
├── end_time
├── auto_renew（是否自动续费）
└── ...

会员权益：
- 全场95折
- 全年包邮
- 专属客服
- 优先发货
- 会员专享价
```

优点：
- 现金流稳定
- 用户粘性高
- 权益明确

缺点：
- 需要足够吸引力的权益
- 续费率是关键

**方案对比**：

| 方案 | 用户粘性 | 权益丰富度 | 实施难度 | 盈利能力 |
|------|---------|-----------|---------|----------|
| 单一积分 | ★★★☆☆ | ★★☆☆☆ | ★★★★★ | ★★☆☆☆ |
| 双轨制 | ★★★★☆ | ★★★★★ | ★★★☆☆ | ★★★☆☆ |
| 付费会员 | ★★★★★ | ★★★★☆ | ★★★★☆ | ★★★★★ |

**推荐方案**：
采用**双轨制（积分+成长值）**。

实施要点：

1. **积分获取规则**：
   ```
   消费积分：
   - 订单完成后发放
   - 1元 = 1积分
   - 会员等级倍率（金卡1.5倍）
   
   行为积分：
   - 签到：5积分/天
   - 分享：10积分/次
   - 评价：20积分/次（带图50积分）
   - 首次购买：100积分
   ```

2. **积分消费**：
   ```
   抵扣规则：
   - 100积分 = 1元
   - 单笔订单最多抵扣订单金额的50%
   - 部分品类不支持积分抵扣（如iPhone）
   
   兑换商品：
   - 积分商城
   - 固定积分兑换商品
   ```

3. **等级维护**：
   ```
   升级：
   - 成长值达到阈值立即升级
   - 发送升级通知
   
   降级：
   - 每年12月31日统计年度成长值
   - 未达标的会员降级
   - 降级前1个月提醒
   - 保级活动（充值、消费保级）
   ```

4. **积分过期**：
   ```
   策略：
   - 积分有效期1年
   - 每年12月31日清零即将过期积分
   - 提前3个月、1个月、1周提醒
   ```

5. **防刷策略**：
   ```
   - 签到积分：每天限1次
   - 分享积分：每天限3次
   - 评价积分：每订单限1次
   - 异常行为检测（短时间大量操作）
   ```

**延伸思考**：
1. 如何设计会员等级的有效期（年度会员）？
2. 积分如何支持转赠功能？
3. 会员权益如何动态调整（AB测试）？

---

##### 🔧 题目5：秒杀活动的价格和库存设计

**问题描述**：
秒杀活动商品价格远低于平时，流量集中，如何设计秒杀的价格和库存系统，保证不超卖且性能可控？

**答案**：

**问题分析**：
秒杀的核心挑战：
1. 瞬时高并发（10万+ QPS）
2. 库存精准控制（100件商品，10万人抢）
3. 价格隔离（秒杀价和正常价不能混淆）
4. 防黄牛（防止脚本抢购）

**方案一：独立秒杀表**

核心思想：
秒杀商品和库存独立存储，与正常商品隔离。

设计：
```sql
seckill_activity（秒杀活动）
├── activity_id
├── name
├── start_time
├── end_time
└── status

seckill_product（秒杀商品）
├── seckill_id
├── activity_id
├── sku_id
├── seckill_price（秒杀价）
├── normal_price（原价）
├── total_stock（秒杀库存）
├── remaining_stock（剩余库存）
├── limit_per_user（每人限购）
└── ...

用户下单：
1. 检查活动时间
2. 检查用户是否已购买（限购）
3. 扣减秒杀库存（Redis）
4. 创建订单（秒杀价）
```

优点：
- 隔离性好
- 不影响正常业务
- 数据清晰

缺点：
- 数据冗余

**方案二：共享商品表+秒杀标记**

核心思想：
秒杀商品复用商品表，通过标记区分。

设计：
```sql
product
├── sku_id
├── normal_price
├── is_seckill（是否秒杀商品）
├── seckill_price
├── seckill_stock
└── ...

价格查询：
if (product.is_seckill && isInSeckillTime()) {
  return product.seckill_price;
} else {
  return product.normal_price;
}
```

优点：
- 无冗余
- 实现简单

缺点：
- 秒杀和正常业务混在一起
- 容易出错（价格混淆）

**方案三：分层架构（推荐）**

核心思想：
前台秒杀系统 + 后台正常系统，数据隔离。

架构：
```text
秒杀系统：
- 秒杀商品（独立表）
- 秒杀库存（Redis）
- 秒杀订单（独立表）
- 秒杀队列（削峰）

正常系统：
- 商品表
- 库存表
- 订单表

数据同步：
- 秒杀结束后同步到正常订单表
- 库存变更同步
```

优点：
- 完全隔离
- 互不影响
- 可针对性优化

缺点：
- 架构复杂
- 数据同步成本

**方案对比**：

| 方案 | 隔离性 | 性能 | 实施难度 | 适用场景 |
|------|--------|------|---------|----------|
| 独立秒杀表 | ★★★★☆ | ★★★★☆ | ★★★☆☆ | 中小型 |
| 共享表 | ★★☆☆☆ | ★★★☆☆ | ★★★★★ | 小型 |
| 分层架构 | ★★★★★ | ★★★★★ | ★★☆☆☆ | 大型 |

**推荐方案**：
采用**独立秒杀表**。

实施要点：

1. **秒杀库存**：
   ```
   Redis存储：
   key: seckill:stock:{seckill_id}
   value: 剩余库存数量
   
   扣减（Lua脚本）：
   local stock = redis.call('GET', KEYS[1])
   if tonumber(stock) > 0 then
     redis.call('DECR', KEYS[1])
     return 1
   else
     return 0
   end
   ```

2. **限购控制**：
   ```
   Redis Set记录已购买用户：
   key: seckill:bought:{seckill_id}
   value: Set<user_id>
   
   检查：
   if (redis.sismember(key, user_id)) {
     return "已购买，不能重复购买";
   }
   
   记录：
   redis.sadd(key, user_id);
   ```

3. **排队机制**：
   ```
   流程：
   1. 用户点击"立即抢购"
   2. 请求进入队列（Kafka）
   3. 显示排队位置
   4. Worker消费队列，限速扣减库存
   5. 扣减成功，通知用户
   6. 扣减失败，提示已售罄
   
   优点：
   - 削峰
   - 用户体验可控
   - 系统稳定
   ```

4. **防黄牛**：
   ```
   策略1：验证码
   - 点击抢购后弹出验证码
   - 通过验证才能提交订单
   
   策略2：实人认证
   - 首次参与秒杀需要实人认证
   - 人脸识别
   
   策略3：行为分析
   - 检测异常高频请求
   - IP黑名单
   - 设备指纹
   ```

5. **价格展示**：
   ```
   商品详情页：
   - 正常价：¥999（划线价）
   - 秒杀价：¥199（红色突出显示）
   - 倒计时：距开始还剩 01:23:45
   - 提醒：每人限购1件
   ```

**延伸思考**：
1. 秒杀订单未支付如何处理（是否释放库存）？
2. 秒杀活动如何预热（提前加载数据）？
3. 秒杀流量如何监控和应急处理？

---

##### 💡 题目6：动态定价系统的设计

**问题描述**：
电商平台希望实现动态定价（如机票、酒店根据供需实时调价）。如何设计动态定价系统？

**答案**：

**问题分析**：
动态定价的核心要素：
1. 定价因子（库存、时间、竞争对手、需求）
2. 定价策略（规则还是算法）
3. 价格变动频率
4. 用户体验（频繁变价影响用户信任）

**方案一：规则引擎定价**

核心思想：
根据预设规则调整价格。

规则示例：
```text
规则1：库存定价
- 库存 > 80% → 原价
- 库存 50%-80% → 原价 × 1.1
- 库存 20%-50% → 原价 × 1.2
- 库存 < 20% → 原价 × 1.3

规则2：时间定价
- 旺季（11-12月）→ 原价 × 1.2
- 淡季（3-4月）→ 原价 × 0.8

规则3：竞争对手定价
- 获取竞争对手价格
- 自己价格 = 竞对价格 × 0.95（低5%）

规则4：用户画像定价
- 高价值用户 → 原价
- 价格敏感用户 → 原价 × 0.9
```

优点：
- 可控
- 易于理解
- 运营可配置

缺点：
- 规则固定，不够灵活
- 无法自适应市场变化

**方案二：算法定价（推荐）**

核心思想：
使用机器学习预测最优价格。

设计：
```text
输入特征：
- 商品属性（品牌、类目、成本）
- 库存水位
- 历史销量
- 竞争对手价格
- 用户画像（购买力、价格敏感度）
- 时间特征（星期几、节假日）
- 外部因素（天气、事件）

模型：
- 回归模型：预测最优价格
- 强化学习：实时调整价格，最大化收益

输出：
- 推荐价格
- 置信度
```

优点：
- 智能化
- 自适应
- 收益最大化

缺点：
- 需要算法团队
- 冷启动问题
- 黑盒，不透明

**方案三：AB测试定价**

核心思想：
多个价格同时测试，选择效果最好的。

流程：
```text
1. 设定价格组：
   A: ¥99
   B: ¥109
   C: ¥119

2. 随机分流用户

3. 统计各价格组的转化率和收益

4. 选择最优价格作为主价格

5. 持续迭代测试
```

优点：
- 基于实际数据
- 科学决策

缺点：
- 测试周期长
- 需要流量支持

**方案对比**：

| 方案 | 灵活性 | 效果 | 实施难度 | 适用场景 |
|------|--------|------|---------|----------|
| 规则引擎 | ★★★☆☆ | ★★★☆☆ | ★★★★☆ | 标品 |
| 算法定价 | ★★★★★ | ★★★★★ | ★★☆☆☆ | 大平台 |
| AB测试 | ★★★★☆ | ★★★★☆ | ★★★★☆ | 新品 |

**推荐方案**：
采用**规则引擎+算法定价**的混合方案。

实施要点：

1. **定价数据收集**：
   ```sql
   price_history（价格历史）
   ├── sku_id
   ├── price
   ├── stock
   ├── sales_quantity（该价格下的销量）
   ├── conversion_rate（转化率）
   ├── start_time
   └── end_time
   
   competitor_price（竞对价格）
   ├── sku_id
   ├── competitor_name
   ├── price
   ├── crawled_at
   └── ...
   ```

2. **定价决策流程**：
   ```
   定时任务（每小时）：
   1. 收集数据（库存、销量、竞对价格）
   2. 输入定价模型
   3. 模型输出推荐价格
   4. 人工审核（可选）
   5. 更新商品价格
   6. 记录价格变更日志
   ```

3. **价格锁定**：
   ```
   用户加购物车：
   - 锁定当前价格15分钟
   - 15分钟内下单按锁定价
   - 超时按最新价
   
   或：
   - 不锁定价格
   - 下单时实时计算（用户体验差）
   ```

4. **价格展示**：
   ```
   对用户：
   - 显示当前价
   - 历史最低价（增加紧迫感）
   - 降价通知（用户订阅）
   
   对运营：
   - 价格趋势图
   - 竞对价格对比
   - 销量-价格关系
   ```

5. **价格保护**：
   ```
   规则：
   - 单次调价幅度 <= 20%
   - 每天最多调价3次
   - 价格不低于成本价 × 1.1（保证毛利）
   - 价格不高于市场价 × 1.5（防止离谱）
   ```

**延伸思考**：
1. 如何处理用户对频繁变价的不满？
2. 价格歧视（同一商品不同用户不同价）的法律风险？
3. 如何设计价格保护机制（买贵退差价）？

---

##### 📊 题目7：跨境电商的汇率和税费计算

**问题描述**：
跨境电商需要处理多币种和不同国家的税费。如何设计汇率转换和税费计算系统？

**答案**：

**问题分析**：
跨境价格的核心要素：
1. 汇率实时变动
2. 不同国家税率不同（关税、增值税）
3. 币种展示（用户看到本地币种）
4. 结算币种（实际收款币种）

**方案一：实时汇率**

核心思想：
每次计算价格时查询实时汇率。

设计：
```text
价格计算：
1. 商品基础价格（USD $100）
2. 查询实时汇率（USD/CNY = 7.2）
3. 转换为人民币（¥720）
4. 加税费（关税10%，增值税13%）
5. 最终价格（¥720 × 1.1 × 1.13 = ¥894）

汇率来源：
- 调用汇率API（如XE, OANDA）
- 每分钟更新一次
```

优点：
- 汇率准确
- 实时性好

缺点：
- 价格频繁变化
- 用户体验差
- API成本高

**方案二：固定汇率（推荐）**

核心思想：
每天固定汇率，当天内价格不变。

设计：
```sql
exchange_rate（汇率表）
├── from_currency
├── to_currency
├── rate
├── effective_date（生效日期）
└── created_at

价格计算：
1. 查询今日汇率（缓存）
2. 转换币种
3. 加税费
4. 展示价格

汇率更新：
- 每天凌晨0点更新汇率
- 或管理员手动更新
```

优点：
- 价格稳定
- 用户体验好
- 缓存友好

缺点：
- 汇率不是实时
- 可能有汇兑损失

**方案三：汇率浮动区间**

核心思想：
设置汇率波动阈值，超过阈值才更新。

设计：
```text
固定汇率：7.2（基准）
浮动区间：±2%（7.056 - 7.344）

实时汇率：7.25
→ 在区间内，使用固定汇率7.2

实时汇率：7.40
→ 超出区间，更新固定汇率为7.4
```

优点：
- 平衡稳定性和准确性
- 减少价格变化频率

缺点：
- 实现复杂度高

**方案对比**：

| 方案 | 准确性 | 稳定性 | 用户体验 | 实施难度 |
|------|--------|--------|---------|---------|
| 实时汇率 | ★★★★★ | ★★☆☆☆ | ★★☆☆☆ | ★★★☆☆ |
| 固定汇率 | ★★★☆☆ | ★★★★★ | ★★★★★ | ★★★★☆ |
| 浮动区间 | ★★★★☆ | ★★★★☆ | ★★★★☆ | ★★☆☆☆ |

**推荐方案**：
采用**固定汇率（每日更新）**。

实施要点：

1. **汇率管理**：
   ```java
   public class ExchangeRateService {
     @Scheduled(cron = "0 0 0 * * ?")  // 每天0点
     public void updateExchangeRate() {
       // 1. 调用汇率API获取最新汇率
       Map<String, BigDecimal> rates = fetchRatesFromAPI();
       
       // 2. 保存到数据库
       for (String pair : rates.keySet()) {
         ExchangeRate rate = new ExchangeRate();
         rate.setPair(pair);
         rate.setRate(rates.get(pair));
         rate.setEffectiveDate(LocalDate.now());
         repository.save(rate);
       }
       
       // 3. 刷新缓存
       cacheService.refreshRates(rates);
     }
   }
   ```

2. **税费计算**：
   ```sql
   tax_rule（税费规则）
   ├── country_code
   ├── category_id
   ├── import_duty_rate（关税率）
   ├── vat_rate（增值税率）
   ├── min_tax_free_amount（免税额）
   └── ...
   
   示例：
   中国：
   - 关税：10%
   - 增值税：13%
   - 免税额：¥5000以下免税
   
   美国：
   - 关税：0%
   - 州税：0-10%（各州不同）
   ```

3. **价格展示**：
   ```
   商品页展示：
   - 商品价格：$100
   - 运费：$20
   - 关税：$10（预估）
   - 总计：$130（约¥936）
   
   结算页：
   - 确认最终价格（包含税费）
   - 币种选择（CNY/USD）
   ```

4. **结算币种**：
   ```
   策略1：统一结算币种
   - 平台统一收USD
   - 用户支付CNY → 银行自动换汇
   
   策略2：多币种账户
   - 平台有USD、CNY、EUR账户
   - 用户付CNY → 直接入CNY账户
   - 减少汇兑成本
   ```

5. **汇率风险对冲**：
   ```
   风险：
   - 用户下单时汇率7.2
   - 商家收款时汇率7.0
   - 平台损失2%
   
   对冲策略：
   - 购买外汇期货
   - 设置汇率浮动保护（±1%）
   - 及时结汇
   ```

**延伸思考**：
1. 如何设计多币种支付（用户用USD支付CNY订单）？
2. 汇率变化导致退款金额不一致如何处理？
3. 跨境税费如何合规申报？

---

##### 🔧 题目8：组合促销的价格计算（满减+折扣+券）

**问题描述**：
用户下单时同时享受满减（满200减30）、商品折扣（9折）、优惠券（20元）。如何设计组合促销的价格计算逻辑？

**答案**：

**问题分析**：
组合促销的核心挑战：
1. 计算顺序（先满减还是先折扣影响最终价）
2. 规则冲突（有些促销不能同时用）
3. 最优组合（如何选择让用户优惠最大）
4. 性能（实时计算）

**方案一：固定计算顺序**

核心思想：
规定促销的固定计算顺序。

计算顺序：
```text
原价：¥500

顺序1：折扣 → 满减 → 优惠券
1. 商品折扣（9折）：¥500 × 0.9 = ¥450
2. 满减（满200减30）：¥450 - ¥30 = ¥420
3. 优惠券（20元）：¥420 - ¥20 = ¥400

顺序2：满减 → 折扣 → 优惠券
1. 满减：¥500 - ¥30 = ¥470
2. 折扣：¥470 × 0.9 = ¥423
3. 优惠券：¥423 - ¥20 = ¥403

结果不同！
```

推荐顺序：
```text
1. 商品级促销（商品折扣、第二件半价）
2. 订单级促销（满减、满赠）
3. 平台级促销（优惠券、积分抵扣）
4. 会员折扣

原则：
- 商品自身属性优先
- 门槛促销次之
- 通用促销最后
```

优点：
- 简单清晰
- 易于实现

缺点：
- 不够灵活
- 可能不是最优惠

**方案二：最优组合（推荐）**

核心思想：
尝试所有可能的组合，选择最优惠的。

算法：
```java
public BigDecimal calculateBestPrice(Order order) {
  // 1. 获取所有适用的促销
  List<Promotion> promotions = getApplicablePromotions(order);
  
  // 2. 生成所有可能的组合（考虑互斥规则）
  List<List<Promotion>> combinations = generateCombinations(promotions);
  
  // 3. 计算每种组合的最终价
  BigDecimal minPrice = order.getOriginalPrice();
  List<Promotion> bestCombination = null;
  
  for (List<Promotion> combo : combinations) {
    BigDecimal price = calculatePrice(order, combo);
    if (price.compareTo(minPrice) < 0) {
      minPrice = price;
      bestCombination = combo;
    }
  }
  
  // 4. 应用最优组合
  return applyPromotions(order, bestCombination);
}

生成组合时考虑互斥：
- 满减A和满减B互斥（只能选一个）
- 优惠券互斥（只能用一张）
- 其他可叠加
```

优点：
- 保证最优惠
- 用户体验最好

缺点：
- 计算量大（组合爆炸）
- 性能压力

优化：
```text
- 限制促销数量（最多5个）
- 剪枝（提前排除明显不优的组合）
- 缓存（相同商品+促销组合缓存结果）
```

**方案三：优先级+互斥**

核心思想：
促销有优先级，互斥的选优先级高的。

设计：
```text
促销列表（按优先级排序）：
1. 优惠券20元（优先级10，互斥组A）
2. 满减30元（优先级20，互斥组A）
3. 商品折扣9折（优先级30，可叠加）
4. 会员折扣95折（优先级40，可叠加）

计算逻辑：
1. 在互斥组A中选择优惠力度最大的（满减30元）
2. 应用可叠加的促销（商品折扣、会员折扣）

最终：
¥500 - ¥30（满减）× 0.9（商品折扣）× 0.95（会员折扣）= ¥401.5
```

优点：
- 平衡灵活性和性能
- 运营可配置

缺点：
- 可能不是全局最优

**方案对比**：

| 方案 | 最优性 | 性能 | 灵活性 | 实施难度 |
|------|--------|------|--------|---------|
| 固定顺序 | ★★☆☆☆ | ★★★★★ | ★★☆☆☆ | ★★★★★ |
| 最优组合 | ★★★★★ | ★★★☆☆ | ★★★★★ | ★★☆☆☆ |
| 优先级+互斥 | ★★★★☆ | ★★★★☆ | ★★★★☆ | ★★★☆☆ |

**推荐方案**：
采用**优先级+互斥**，必要时计算最优组合。

实施要点：

1. **促销配置**：
   ```sql
   promotion
   ├── promotion_id
   ├── name
   ├── type（DISCOUNT/FULL_REDUCE/COUPON）
   ├── priority（优先级）
   ├── exclusive_group（互斥组，NULL表示可叠加）
   ├── stackable（是否可叠加）
   └── ...
   ```

2. **价格明细**：
   ```
   订单价格明细：
   {
     "originalPrice": 500,
     "appliedPromotions": [
       {
         "name": "商品9折",
         "discountAmount": 50,
         "afterPrice": 450
       },
       {
         "name": "满200减30",
         "discountAmount": 30,
         "afterPrice": 420
       },
       {
         "name": "优惠券",
         "discountAmount": 20,
         "afterPrice": 400
       }
     ],
     "finalPrice": 400
   }
   
   用户可见每一步的优惠
   ```

3. **试算接口**：
   ```
   POST /api/price/preview
   {
     "items": [...],
     "promotions": [...],
     "coupon": "SUMMER20"
   }
   
   响应：
   {
     "scenarios": [
       {
         "name": "推荐方案",
         "finalPrice": 400,
         "savings": 100,
         "appliedPromotions": [...]
       },
       {
         "name": "仅用优惠券",
         "finalPrice": 480,
         "savings": 20,
         "appliedPromotions": [...]
       }
     ]
   }
   
   让用户选择方案
   ```

4. **性能优化**：
   ```
   缓存：
   key: price:calculate:{商品ID}:{促销IDs哈希}
   value: 计算结果
   TTL: 5分钟
   
   避免重复计算
   ```

**延伸思考**：
1. 如何向用户推荐最优促销组合？
2. 促销规则变更如何保证已下单的订单价格不变？
3. 如何设计促销的AB测试？

---

##### 💡 题目9：预售和定金膨胀的设计

**问题描述**：
预售活动中，用户支付定金（如50元），尾款时定金可抵100元。如何设计预售和定金膨胀系统？

**答案**：

**问题分析**：
预售定金的核心要素：
1. 定金不可退（锁定用户）
2. 定金膨胀（50元抵100元）
3. 尾款支付期限（超时定金不退）
4. 库存预占

**方案一：双订单模式**

核心思想：
定金订单和尾款订单分开。

设计：
```sql
presale_activity（预售活动）
├── activity_id
├── sku_id
├── deposit_amount（定金）
├── deposit_expand_amount（定金膨胀金额）
├── final_price（商品总价）
├── deposit_start_time
├── deposit_end_time
├── balance_start_time（尾款开始时间）
├── balance_end_time
└── ...

deposit_order（定金订单）
├── order_id
├── activity_id
├── user_id
├── deposit_amount
├── status（PAID/UNPAID）
└── ...

balance_order（尾款订单）
├── order_id
├── deposit_order_id（关联定金订单）
├── balance_amount（尾款金额 = 总价 - 定金膨胀）
├── status
└── ...

流程：
1. 预售期：用户支付定金 → 创建deposit_order
2. 尾款期：系统自动创建balance_order
3. 用户支付尾款
4. 发货
```

优点：
- 清晰分离
- 易于管理

缺点：
- 两个订单，用户理解成本高

**方案二：单订单分阶段（推荐）**

核心思想：
一个订单，分阶段支付。

设计：
```sql
order
├── order_id
├── order_type（PRESALE）
├── presale_activity_id
├── total_amount（商品总价）
├── deposit_amount（已付定金）
├── balance_amount（待付尾款）
├── current_stage（DEPOSIT/BALANCE/COMPLETED）
├── deposit_paid_at
├── balance_deadline
└── ...

order_payment（支付记录）
├── payment_id
├── order_id
├── payment_type（DEPOSIT/BALANCE）
├── amount
├── paid_at
└── ...

流程：
1. 预售期：用户下单，支付定金
   order.current_stage = DEPOSIT
   order.deposit_amount = 50
   order.balance_amount = 总价 - 定金膨胀金额
   
2. 尾款期：订单进入尾款阶段
   order.current_stage = BALANCE
   发送尾款提醒
   
3. 用户支付尾款
   order.current_stage = COMPLETED
   
4. 发货
```

优点：
- 订单统一
- 用户理解成本低
- 易于追踪

缺点：
- 订单状态复杂

**方案三：虚拟商品模式**

核心思想：
定金作为虚拟商品，尾款时抵扣。

设计：
```text
1. 用户购买"定金商品"（¥50）
2. 定金支付成功后，发放"抵扣券"（可抵¥100）
3. 尾款期，用户购买商品，使用抵扣券
4. 实付 = 商品价格 - 抵扣券金额
```

优点：
- 复用现有优惠券系统
- 灵活

缺点：
- 定金和尾款割裂
- 用户可能不理解

**方案对比**：

| 方案 | 清晰度 | 实施难度 | 用户体验 | 适用场景 |
|------|--------|---------|---------|----------|
| 双订单 | ★★★☆☆ | ★★★☆☆ | ★★★☆☆ | 复杂预售 |
| 单订单分阶段 | ★★★★★ | ★★★★☆ | ★★★★★ | 通用 |
| 虚拟商品 | ★★☆☆☆ | ★★★★☆ | ★★☆☆☆ | 简单预售 |

**推荐方案**：
采用**单订单分阶段**。

实施要点：

1. **定金膨胀计算**：
   ```
   商品总价：¥999
   定金：¥50
   定金膨胀：¥100（2倍膨胀）
   尾款：¥999 - ¥100 = ¥899
   
   用户总共支付：¥50 + ¥899 = ¥949（省¥50）
   ```

2. **库存管理**：
   ```
   定金支付成功：
   - 预占库存（reserved_stock +1）
   - 锁定到该订单
   
   尾款支付成功：
   - 确认库存（sold_stock +1, reserved_stock -1）
   
   超时未付尾款：
   - 释放库存（reserved_stock -1）
   - 定金不退
   ```

3. **尾款提醒**：
   ```
   提醒策略：
   - 尾款开始：立即推送
   - 尾款截止前3天：提醒
   - 尾款截止前1天：紧急提醒
   - 尾款截止前1小时：最后提醒
   
   提醒渠道：
   - App推送
   - 短信
   - 站内信
   ```

4. **超时处理**：
   ```
   定时任务（每小时）：
   1. 扫描超时未付尾款的订单
   2. 订单状态 → CLOSED
   3. 释放库存
   4. 定金记为平台收入（不退）
   5. 通知用户
   ```

5. **退款规则**：
   ```
   规则：
   - 支付定金后，不可取消订单
   - 定金不退
   - 尾款支付后，可申请退款
   - 退款金额 = 定金 + 尾款
   ```

**延伸思考**：
1. 如何防止用户恶意付定金占用库存？
2. 预售商品如何设置发货时间？
3. 定金膨胀活动如何设计ROI分析？

---

##### 📊 题目10：价格歧视与个性化定价的设计

**问题描述**：
电商平台希望根据用户画像（新老用户、购买力、价格敏感度）实现个性化定价。如何设计价格歧视系统？同时如何规避法律风险？

**答案**：

**问题分析**：
个性化定价的核心要素：
1. 用户分层（高价值、普通、价格敏感）
2. 定价策略（不同用户看到不同价格）
3. 法律风险（价格歧视在某些国家违法）
4. 用户信任（发现差价后的负面影响）

**方案一：明面价格歧视（不推荐）**

核心思想：
不同用户直接看到不同价格。

示例：
```text
用户A（新用户）：¥99
用户B（老用户）：¥129
用户C（高价值用户）：¥149

价格查询：
price = getPriceByUser(skuId, userId);
```

优点：
- 简单直接
- 收益最大化

缺点：
- 法律风险大（违反价格法）
- 用户信任崩塌（发现后口碑崩盘）
- 媒体曝光风险

**方案二：差异化优惠（推荐）**

核心思想：
价格统一，但不同用户获得不同优惠。

设计：
```text
基础价格：统一¥129

新用户：
- 新人专享券：¥30
- 实付：¥99

普通用户：
- 无优惠
- 实付：¥129

高价值用户：
- 会员折扣：9折
- 实付：¥116

关键：
- 价格统一展示
- 优惠透明（标注"新人专享"、"会员专享"）
```

优点：
- 合法合规
- 用户可接受
- 价格透明

缺点：
- 收益优化程度不如价格歧视

**方案三：隐性定价（灰色地带）**

核心思想：
通过算法展示不同的商品推荐和排序。

策略：
```text
高价值用户：
- 推荐高价商品
- 搜索结果优先展示高价商品

价格敏感用户：
- 推荐促销商品
- 搜索结果优先展示低价商品

不直接改价格，但影响用户选择
```

优点：
- 间接影响购买
- 法律风险小

缺点：
- 效果不如直接定价
- 算法复杂

**方案对比**：

| 方案 | 收益 | 合规性 | 用户信任 | 风险 |
|------|------|--------|---------|------|
| 明面歧视 | ★★★★★ | ★☆☆☆☆ | ★☆☆☆☆ | ★★★★★ |
| 差异化优惠 | ★★★★☆ | ★★★★★ | ★★★★☆ | ★★☆☆☆ |
| 隐性定价 | ★★★☆☆ | ★★★★☆ | ★★★★☆ | ★★★☆☆ |

**推荐方案**：
采用**差异化优惠**。

实施要点：

1. **用户分层**：
   ```
   基于RFM模型：
   - R（最近一次购买）
   - F（购买频次）
   - M（购买金额）
   
   用户分层：
   - 高价值用户（VIP）：R<30天, F>10次, M>1万
   - 活跃用户：R<90天, F>3次
   - 沉睡用户：R>90天
   - 新用户：注册<30天，F=0
   - 价格敏感用户：经常搜索低价、使用优惠券
   ```

2. **差异化优惠策略**：
   ```
   新用户：
   - 新人专享券（大额）
   - 首单免运费
   - 新人专区（低价引流商品）
   
   沉睡用户：
   - 唤醒券（定向发放）
   - "好久不见，给你优惠"
   
   高价值用户：
   - 会员折扣
   - 生日礼包
   - 专属客服
   
   价格敏感用户：
   - 推荐促销商品
   - 凑单优惠
   ```

3. **透明化展示**：
   ```
   商品页：
   - 价格：¥129（统一价格）
   - 您的优惠：
     ✓ 新人券：-¥30
     ✓ 首单免运费
   - 实付：¥99
   
   标注优惠来源，避免误解
   ```

4. **法律合规**：
   ```
   避免：
   - 同一商品同一时间不同价格（价格歧视）
   - 隐藏真实价格
   - 大数据杀熟
   
   合法：
   - 不同用户不同优惠（促销活动）
   - 会员专享价（明确标注）
   - 新人优惠（限定条件）
   ```

5. **监控与风控**：
   ```
   监控指标：
   - 用户投诉率（价格差异投诉）
   - 媒体舆情
   - 价格离散度（同商品价格差异）
   
   风控：
   - 价格差异 < 30%
   - 优惠透明化
   - 避免同一用户看到不同价格
   ```

**延伸思考**：
1. 如何平衡个性化定价和用户信任？
2. 用户发现价格差异后如何应对？
3. 如何设计价格歧视的AB测试（避免法律风险）？

---

---

---
