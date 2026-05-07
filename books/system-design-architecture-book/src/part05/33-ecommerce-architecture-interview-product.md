# 35.2.1 商品中心系统题库

## 35.2.1 商品中心系统（16题）

#### 📊 题目1：设计支持多品类的SPU/SKU数据模型

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

#### 🔧 题目2：商品详情页的缓存架构设计

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

#### 💡 题目3：如何解决商品信息变更后搜索不一致问题？

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

#### 🔧 题目3 扩展：直接订阅 Binlog 同步 ES 的弊端是什么？如果不同变更之间存在依赖关系，应该怎么处理？

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

#### 📊 题目4：设计商品类目体系和属性管理

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

#### 🔧 题目5：商品图片的存储和CDN方案

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

#### 💡 题目6：虚拟商品vs实物商品的设计差异

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

#### 📊 题目7：商品上架流程的工作流设计

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

#### 🔧 题目8：如何支持商品的多规格选择（颜色、尺码等）？

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

#### 💡 题目9：商品快照在订单中的应用

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

#### 📊 题目10：设计商品推荐系统的架构

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

#### 🔧 题目11：商品搜索的倒排索引设计

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

#### 💡 题目12：如何处理商品数据的历史版本？

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

#### 📊 题目13：多租户场景下的商品数据隔离

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

#### 🔧 题目14：商品导入的批量处理优化

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

#### 💡 题目15：商品审核流程的设计

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
