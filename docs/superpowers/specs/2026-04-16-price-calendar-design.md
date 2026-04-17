# 价格日历系统设计方案

**设计日期：** 2026-04-16  
**版本：** v1.0  
**作者：** wxquare  
**状态：** 设计完成，待实施

---

## 一、需求背景

### 1.1 业务目标

为B2B2C电商平台（主营酒店、机票等虚拟商品）设计并实现价格日历功能，帮助用户快速找到最优惠的预订日期，同时为运营团队提供价格管理工具。

### 1.2 核心需求

**用户端需求：**
- 用户选定某个酒店或航线后，查看未来30-60天的价格趋势
- 每个日期显示该日期的最低价格
- 点击日期后跳转到该日期的搜索结果页

**运营端需求：**
- 运营团队可以查看价格同步状态
- 支持手动刷新缓存
- 查看同步任务的成功率和失败原因

**非功能需求：**
- 查询响应时间P99 < 200ms
- 支持10000 QPS（单实例）
- 缓存命中率 > 80%
- 系统可用性 99.9%

### 1.3 业务约束

- **品类范围：** 初期支持酒店和机票，未来扩展到其他品类
- **时间范围：** 展示未来30-60天的价格数据
- **数据来源：** 定时同步供应商价格（非实时查询）
- **数据规模：** 100万酒店 + 10万航线，60天数据约6000万条记录

---

## 二、技术方案选择

### 2.1 备选方案对比

我们评估了三个技术方案：

| 方案 | 核心技术栈 | 优点 | 缺点 | 适用阶段 |
|-----|-----------|------|------|---------|
| **方案1** | MySQL + Redis | 技术栈成熟、快速上线、运维成本低 | 数据量增长需分库分表 | 初期（0-1） |
| **方案2** | MySQL主从 + Redis + MQ | 读写分离、高并发、削峰填谷 | 架构复杂、主从延迟 | 成熟期（1-10） |
| **方案3** | TimescaleDB + Redis | 时序优化、高压缩率、长期存储 | 学习成本高、运维复杂 | 长期（10+） |

### 2.2 最终选择：方案1（MySQL + Redis）

**理由：**
1. **快速上线：** 团队熟悉MySQL和Redis，2-3周即可完成开发
2. **风险可控：** 利用现有基础设施，无需引入新组件
3. **渐进式演进：** 后续可平滑升级到方案2或方案3

**演进路径：**
```
阶段1（现在）：MySQL + Redis，支撑10万QPS
         ↓
阶段2（6个月）：加入主从分离 + 消息队列
         ↓
阶段3（1年）：迁移到TimescaleDB（如需长期历史数据）
```

---

## 三、系统架构设计

### 3.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        客户端层                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Web前端     │  │  移动端App    │  │ 运营管理后台  │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└───────────────────────────┬─────────────────────────────────────┘
                            │ HTTPS
┌───────────────────────────┼─────────────────────────────────────┐
│                    API网关层                                      │
│         ┌──────────────────────────────────┐                     │
│         │   API Gateway (Kong/Nginx)       │                     │
│         │  - 限流                           │                     │
│         │  - 鉴权                           │                     │
│         │  - 路由                           │                     │
│         └──────────────────────────────────┘                     │
└───────────────────────────┼─────────────────────────────────────┘
                            │
┌───────────────────────────┼─────────────────────────────────────┐
│                     应用服务层                                     │
│  ┌─────────────────────────────────────────────────┐            │
│  │    Price Calendar Service (Go)                  │            │
│  │  ┌──────────────┐  ┌──────────────────────┐    │            │
│  │  │ Query API    │  │  Admin API           │    │            │
│  │  │ (客户查询)    │  │  (运营管理)           │    │            │
│  │  └──────────────┘  └──────────────────────┘    │            │
│  └─────────────────────────────────────────────────┘            │
│                            │                                      │
│  ┌─────────────────────────────────────────────────┐            │
│  │    Price Sync Service (Go)                      │            │
│  │  - 定时拉取供应商价格                             │            │
│  │  - 批量写入MySQL                                 │            │
│  │  - 刷新Redis缓存                                 │            │
│  └─────────────────────────────────────────────────┘            │
└───────────────────────────┼─────────────────────────────────────┘
                            │
┌───────────────────────────┼─────────────────────────────────────┐
│                      存储层                                        │
│  ┌──────────────────┐          ┌──────────────────┐             │
│  │  Redis Cluster   │          │  MySQL (InnoDB)  │             │
│  │  - 热点缓存      │          │  - 全量价格数据   │             │
│  │  - TTL 7天       │          │  - 保留60天      │             │
│  └──────────────────┘          └──────────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                            │
┌───────────────────────────┼─────────────────────────────────────┐
│                    外部依赖层                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ 酒店供应商API │  │ 机票供应商API │  │  监控系统     │          │
│  │ (Expedia等)  │  │ (Amadeus等)  │  │ (Prometheus) │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 核心模块职责

**Price Calendar Service（价格日历服务）**
- 提供C端用户查询价格日历的API
- 提供运营后台的价格管理API
- 负责查询逻辑：Redis → MySQL降级
- 水平扩展支持（无状态服务）

**Price Sync Service（价格同步服务）**
- 定时任务拉取供应商价格数据
- 批量写入MySQL
- 异步刷新Redis热点数据
- 支持全量同步和增量更新

**Redis层**
- 存储热门SKU的近7天价格
- Key结构：`price:{category}:{sku_id}` (Hash类型)
- 自动过期（TTL 7天）

**MySQL层**
- 主库：价格数据持久化
- 保留60天数据，自动清理过期数据
- 索引优化：(sku_id, date)复合索引

---

## 四、数据模型设计

### 4.1 MySQL表结构

```sql
-- 价格日历主表
CREATE TABLE price_calendar (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  sku_id VARCHAR(64) NOT NULL COMMENT '商品SKU ID（酒店ID/航线ID）',
  category ENUM('hotel', 'flight') NOT NULL COMMENT '品类',
  date DATE NOT NULL COMMENT '日期（酒店=入住日期，机票=出发日期）',
  min_price DECIMAL(10,2) NOT NULL COMMENT '最低价格',
  currency VARCHAR(3) DEFAULT 'CNY' COMMENT '货币单位',
  supplier_id VARCHAR(32) COMMENT '供应商ID',
  supplier_name VARCHAR(128) COMMENT '供应商名称',
  available_count INT DEFAULT 0 COMMENT '可售数量（酒店=房间数，机票=座位数）',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  -- 索引
  UNIQUE KEY uk_sku_date (sku_id, date),
  INDEX idx_date (date),
  INDEX idx_category_date (category, date),
  INDEX idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='价格日历表';

-- 分区策略（数据量大时启用）
ALTER TABLE price_calendar PARTITION BY RANGE (TO_DAYS(date)) (
  PARTITION p202604 VALUES LESS THAN (TO_DAYS('2026-05-01')),
  PARTITION p202605 VALUES LESS THAN (TO_DAYS('2026-06-01')),
  PARTITION p202606 VALUES LESS THAN (TO_DAYS('2026-07-01')),
  PARTITION pmax VALUES LESS THAN MAXVALUE
);
```

**字段说明：**
- `sku_id`：酒店ID（如"hotel_123456"）或航线ID（如"flight_PEK_SHA"）
- `date`：关键字段，酒店表示入住日期，机票表示出发日期
- `min_price`：该日期的最低价格（跨多个供应商的最低值）
- `available_count`：可售数量，用于判断是否售罄

### 4.2 Redis数据结构

**热点SKU价格缓存（Hash结构）**

```
Key: price:hotel:{sku_id}
Type: Hash
Fields:
  2026-04-20 → "299.00|sp001|10"  (价格|供应商ID|库存)
  2026-04-21 → "350.00|sp002|5"
  2026-04-22 → "280.00|sp001|20"
  ...
TTL: 7天

Key: price:flight:{sku_id}
Type: Hash
Fields:
  2026-04-20 → "450.00|sp003|15"
  2026-04-21 → "380.00|sp004|8"
  ...
TTL: 7天
```

**热点SKU列表（用于判断哪些SKU需要缓存）**

```
Key: hotkeys:hotel
Type: ZSet (Sorted Set)
Members:
  hotel_123456 score:10000 (访问次数)
  hotel_789012 score:8500
  ...
TTL: 1小时

Key: hotkeys:flight
Type: ZSet
Members:
  flight_PEK_SHA score:12000
  flight_SHA_SZX score:9000
  ...
TTL: 1小时
```

### 4.3 数据容量估算

**数据量：**
- 酒店：100万个 × 60天 = 6000万条记录
- 机票：10万条航线 × 60天 = 600万条记录
- 合计：约6600万条记录

**存储空间：**
- 单条记录大小：约100 bytes（不含索引）
- 数据大小：6600万 × 100 bytes ≈ 6.6 GB
- 索引大小：约2倍数据大小 ≈ 13 GB
- MySQL总空间：约20 GB（含冗余）

**Redis内存：**
- 热点SKU数量：100万（Top 10%）
- 每个SKU 7天数据：7 × 100 bytes = 700 bytes
- Hash结构开销：约20%
- 总内存：100万 × 700 × 1.2 ≈ 840 MB
- 推荐配置：4 GB（预留3倍空间）

---

## 五、API设计

### 5.1 客户端查询API

**查询价格日历**

```http
GET /api/v1/price-calendar

请求参数：
{
  "category": "hotel",           // 必填，品类：hotel/flight
  "sku_id": "hotel_123456",      // 必填，商品ID
  "start_date": "2026-04-20",    // 必填，开始日期
  "end_date": "2026-05-20",      // 必填，结束日期（最多查询60天）
  "currency": "CNY"              // 可选，货币单位，默认CNY
}

响应示例：
{
  "code": 0,
  "message": "success",
  "data": {
    "sku_id": "hotel_123456",
    "category": "hotel",
    "prices": [
      {
        "date": "2026-04-20",
        "min_price": 299.00,
        "currency": "CNY",
        "supplier_id": "sp001",
        "supplier_name": "Expedia",
        "available": true,
        "available_count": 10
      },
      {
        "date": "2026-04-21",
        "min_price": 350.00,
        "currency": "CNY",
        "supplier_id": "sp002",
        "supplier_name": "Booking",
        "available": true,
        "available_count": 5
      }
    ],
    "cache_hit": true,
    "total_days": 30
  }
}

错误响应：
{
  "code": 40001,
  "message": "date range too large, maximum 60 days",
  "data": null
}
```

**批量查询多个SKU的价格日历**

```http
POST /api/v1/price-calendar/batch

请求参数：
{
  "category": "hotel",
  "sku_ids": ["hotel_123456", "hotel_789012", "hotel_345678"],
  "start_date": "2026-04-20",
  "end_date": "2026-04-27"
}

响应：略（结构类似单个查询，data字段按sku_id分组）
```

### 5.2 运营管理API

**手动刷新价格缓存**

```http
POST /api/v1/admin/price-calendar/refresh

请求参数：
{
  "category": "hotel",
  "sku_ids": ["hotel_123456"],
  "date_range": {
    "start_date": "2026-04-20",
    "end_date": "2026-04-27"
  }
}

响应：
{
  "code": 0,
  "message": "refresh triggered",
  "data": {
    "task_id": "refresh_20260420_001",
    "sku_count": 1,
    "estimated_time": "30s"
  }
}
```

**查询同步任务状态**

```http
GET /api/v1/admin/sync-tasks

响应示例：
{
  "code": 0,
  "data": {
    "tasks": [
      {
        "task_id": "sync_hotel_20260420_0800",
        "category": "hotel",
        "status": "running",
        "started_at": "2026-04-20T08:00:00Z",
        "duration": "120s",
        "total_sku": 100000,
        "processed_sku": 85000,
        "failed_sku": 120,
        "error_message": null
      }
    ]
  }
}
```

### 5.3 API限流策略

```
客户端查询API：
- 单用户：100 QPS
- 单IP：500 QPS
- 全局：50000 QPS

运营管理API：
- IP白名单
- 刷新操作：10次/小时
```

---

## 六、核心业务流程

### 6.1 价格同步流程

```
定时任务触发（Cron: 每小时）
   ↓
步骤1: 获取需要同步的SKU列表
   ↓
步骤2: 批量调用供应商API（并发100）
   ↓
步骤3: 数据标准化和聚合（取最低价）
   ↓
步骤4: 批量写入MySQL（1000条/批）
   ↓
步骤5: 异步刷新Redis缓存（热点SKU）
   ↓
步骤6: 记录日志和监控指标
```

**伪代码：**

```go
func SyncPriceCalendar(category string) error {
    // 1. 获取SKU列表
    skus := getSKUList(category)
    
    // 2. 并发调用供应商API
    results := make(chan *PriceResult, len(skus))
    sem := make(chan struct{}, 100)
    
    for _, sku := range skus {
        sem <- struct{}{}
        go func(s string) {
            defer func() { <-sem }()
            priceData := callSupplierAPIWithRetry(s, 60)
            results <- priceData
        }(sku)
    }
    
    // 3. 收集结果并聚合
    var records []PriceRecord
    for i := 0; i < len(skus); i++ {
        result := <-results
        if result.Error != nil {
            continue
        }
        aggregated := aggregatePrices(result.Prices)
        records = append(records, aggregated...)
    }
    
    // 4. 批量写入MySQL
    batchInsertMySQL(records, 1000)
    
    // 5. 刷新热点SKU缓存
    hotSKUs := getHotSKUs(category, 1000)
    for _, sku := range hotSKUs {
        updateRedisCache(sku, records)
    }
    
    // 6. 上报监控指标
    metrics.RecordSyncSuccess(len(records))
    return nil
}
```

### 6.2 价格查询流程

```
用户请求 → API Gateway限流 → Price Calendar Service
   ↓
参数校验（日期范围<=60天）
   ↓
L1: 查询Redis（Hash HGETALL）
   ↓ 命中？
  Yes → 返回
   ↓ No
L2: 查询MySQL（WHERE sku_id AND date BETWEEN）
   ↓
异步写入Redis（热点SKU）
   ↓
返回结果
```

**伪代码：**

```go
func QueryPriceCalendar(req *QueryRequest) (*QueryResponse, error) {
    // 参数校验
    if err := validateRequest(req); err != nil {
        return nil, err
    }
    
    // L1: Redis查询
    cacheKey := fmt.Sprintf("price:%s:%s", req.Category, req.SKUID)
    cachedData, err := redis.HGetAll(cacheKey).Result()
    
    if err == nil && isCoverDateRange(cachedData, req.StartDate, req.EndDate) {
        metrics.RecordCacheHit("redis")
        return formatResponse(cachedData, req), nil
    }
    
    // L2: MySQL查询
    query := `SELECT * FROM price_calendar 
              WHERE sku_id = ? AND category = ? 
              AND date BETWEEN ? AND ? ORDER BY date`
    
    var records []PriceRecord
    db.Select(&records, query, req.SKUID, req.Category, req.StartDate, req.EndDate)
    
    // 异步更新Redis
    if isHotSKU(req.SKUID) {
        go updateRedisCache(cacheKey, records, 7*24*time.Hour)
    }
    
    metrics.RecordCacheMiss("redis")
    return formatResponse(records, req), nil
}
```

### 6.3 热点SKU识别流程

```
用户每次查询
   ↓
Redis记录访问计数：ZINCRBY hotkeys:hotel hotel_123 1
   ↓
定时任务（每小时）
   ↓
获取Top 1000热点SKU：ZREVRANGE hotkeys:hotel 0 999
   ↓
预加载价格到Redis
   ↓
重置计数器（每天凌晨）
```

---

## 七、数据生命周期管理

### 7.1 过期数据定义

**过期数据：** 日期早于当前日期的价格记录（date < CURDATE()）

### 7.2 清理策略

**MySQL清理：**

```sql
-- 定时删除（每天凌晨2点）
DELETE FROM price_calendar 
WHERE date < DATE_SUB(CURDATE(), INTERVAL 3 DAY)
LIMIT 10000;

-- 分区删除（每月1号删除3个月前的分区）
ALTER TABLE price_calendar DROP PARTITION p202601;
```

**Redis清理：**
- 自动过期：所有Key设置TTL=7天，Redis自动清理
- 主动清理：每天凌晨3点清理Hash中的过期日期字段

**监控指标：**
```
- price_calendar_oldest_date: 数据库中最早的日期（应该>=当前日期-3天）
- price_calendar_expired_rows_deleted: 每天删除的过期记录数
```

**告警规则：**
```
如果oldest_date < CURDATE() - 7天，触发告警（清理任务失败）
```

---

## 八、性能优化策略

### 8.1 数据库优化

**索引优化：**
- 核心索引：`uk_sku_date (sku_id, date)` 覆盖90%查询
- 辅助索引：`idx_date (date)` 用于清理过期数据

**批量查询优化：**
```sql
-- 拆分为多个单SKU查询并并发执行
SELECT * FROM price_calendar WHERE sku_id=? AND date BETWEEN ? AND ?;
```

**分库分表策略（数据量>5000万时）：**
- 分表键：sku_id（Hash取模）
- 分表数量：16个表
- 路由逻辑：`table_index = crc32(sku_id) % 16`

### 8.2 Redis优化

**内存优化：**
- 推荐配置：4 GB内存
- 最大内存策略：allkeys-lru（自动淘汰最少使用的Key）

**连接池配置：**
```go
MaxIdle:     100
MaxActive:   500
IdleTimeout: 300s
```

**缓存预热：**
- 系统启动时预热Top 1000热点SKU
- 从MySQL加载近7天数据写入Redis

### 8.3 供应商API调用优化

**并发控制：**
- 信号量控制并发数：最多100个并发请求
- 超时控制：单个请求3秒超时
- 熔断器：错误率>50%触发熔断，熔断时间5秒

**重试策略：**
- 最多重试2次
- 指数退避：100ms, 200ms, 400ms

### 8.4 应用层优化

**连接复用：**
```go
// HTTP客户端连接池
MaxIdleConns:        200
MaxIdleConnsPerHost: 100

// MySQL连接池
MaxOpenConns: 200
MaxIdleConns: 50
```

**批量操作：**
- MySQL批量插入：1000条/批
- 使用`INSERT ... ON DUPLICATE KEY UPDATE`

### 8.5 性能指标目标

```
查询接口：
  - P50延迟：< 50ms
  - P99延迟：< 200ms
  - 吞吐量：10000 QPS/实例
  - 缓存命中率：> 80%

同步任务：
  - 100万SKU同步时间：< 30分钟
  - 供应商API成功率：> 95%
  - 数据库写入速度：> 5000条/秒
```

---

## 九、错误处理与容灾

### 9.1 错误分类

**客户端错误（4xx）：**
- 400 Bad Request：参数错误（日期格式、范围超限）
- 404 Not Found：SKU不存在、无价格数据
- 429 Too Many Requests：限流

**服务端错误（5xx）：**
- 500 Internal Server Error：数据库/Redis连接失败
- 503 Service Unavailable：供应商API全部失败
- 504 Gateway Timeout：查询超时（>3秒）

### 9.2 降级策略

**多级降级：**
```
L1: Redis故障 → 降级到MySQL查询
L2: MySQL从库故障 → 降级到主库查询
L3: 全部数据库故障 → 返回静态默认数据
```

**降级开关：**
- 通过配置中心管理降级开关
- 支持动态开启/关闭Redis、MySQL

### 9.3 重试策略

**供应商API重试：**
- 最多重试2次
- 指数退避：100ms, 200ms, 400ms
- 只重试5xx错误，4xx不重试

### 9.4 数据一致性保证

**缓存一致性：**
```
更新流程：
1. 更新MySQL
2. 删除Redis缓存（Cache-Aside模式）
3. 下次查询时重新加载

防止缓存击穿：
- 使用分布式锁
- Double Check模式
```

### 9.5 监控与告警

**Prometheus指标：**
```yaml
- price_calendar_query_duration_seconds (histogram)
- price_calendar_query_total (counter)
- price_calendar_cache_hit_total (counter)
- price_sync_duration_seconds (histogram)
- price_sync_sku_total (counter)
- mysql_connection_pool_active (gauge)
- redis_memory_used_bytes (gauge)
```

**告警规则：**
```yaml
- P99延迟 > 500ms，持续5分钟 → warning
- 缓存命中率 < 60%，持续10分钟 → warning
- 同步失败率 > 100个SKU/小时 → critical
- MySQL连接数 > 180 → warning
- 数据堆积 > 1亿条 → critical
```

---

## 十、扩展性设计

### 10.1 水平扩展

**应用层扩展：**
```
负载均衡 → Price Calendar Service集群（3-20实例）
基于CPU使用率自动扩缩容（K8s HPA）
目标CPU：60%
```

**Redis扩展：**
```
单机(4GB) → 主从(读写分离) → Cluster(分片16节点)
```

**MySQL扩展：**
```
单库单表 → 主从分离 → 分库分表(16表)
```

### 10.2 跨品类扩展

**当前支持：** 酒店 + 机票  
**未来扩展：** 充值、电影票、火车票

**扩展方式：**
- 抽象接口：`PriceCalendarService`
- 品类特定实现：`HotelPriceService`, `FlightPriceService`
- 工厂模式创建实例

### 10.3 多区域部署

**单区域架构（初期）：**
```
Region: 中国
  ├─ AZ-1: 主MySQL + Redis主 + 应用实例
  ├─ AZ-2: 从MySQL + Redis从 + 应用实例
  └─ AZ-3: 从MySQL + Redis从 + 应用实例
```

**多区域架构（未来）：**
```
Region: 中国（主）+ Region: 东南亚（从）
MySQL跨region主从同步（异步）
Redis本地缓存（各region独立）
```

### 10.4 业界对比

| 维度 | Google Flights | Booking | Airbnb | 我们的方案 |
|-----|---------------|---------|--------|----------|
| 价格展示 | 趋势图+最低价 | 最低价 | 每晚价 | 最低价 |
| 缓存策略 | BigQuery | Redis | Memcached | Redis+MySQL |
| 价格预测 | ✓ AI预测 | ✗ | ✗ | ✗(后续可加) |
| 实时性 | 准实时 | 准实时 | 准实时 | 定时同步 |

**我们的优势：**
- 架构简单，快速上线
- 成本可控，适合初期规模
- 扩展性好，可平滑演进

---

## 十一、实施计划

### 11.1 开发排期

```
Phase 1（2周）：基础功能
  - 数据模型设计
  - MySQL表结构
  - 基础查询API
  
Phase 2（2周）：缓存优化
  - Redis缓存层
  - 热点识别
  - 缓存预热
  
Phase 3（2周）：同步服务
  - 供应商API调用
  - 批量写入优化
  - 定时任务调度
  
Phase 4（1周）：监控告警
  - Prometheus指标
  - Grafana看板
  - 告警规则
  
Phase 5（1周）：压测优化
  - 性能压测
  - 瓶颈分析
  - 优化调优

总计：8周
```

### 11.2 技术栈

```
语言：Go 1.21+
数据库：MySQL 8.0
缓存：Redis 7.0
监控：Prometheus + Grafana
日志：ELK Stack
部署：Kubernetes + Docker
```

### 11.3 团队配置

```
- 后端开发：2人（Go）
- 前端开发：1人（React/Vue）
- 测试：1人
- 运维：1人（兼职）
```

---

## 十二、风险评估

### 12.1 技术风险

| 风险 | 影响 | 概率 | 缓解措施 |
|-----|------|------|---------|
| 供应商API不稳定 | 高 | 中 | 重试+熔断+降级 |
| MySQL性能瓶颈 | 中 | 低 | 分库分表+缓存 |
| Redis内存不足 | 中 | 低 | 内存监控+LRU淘汰 |
| 数据一致性问题 | 高 | 低 | Cache-Aside+分布式锁 |

### 12.2 业务风险

| 风险 | 影响 | 概率 | 缓解措施 |
|-----|------|------|---------|
| 价格数据延迟 | 中 | 中 | 定时同步频率可调整 |
| 用户查询无结果 | 低 | 低 | 提示"暂无价格数据" |
| 流量突增 | 高 | 中 | 自动扩容+限流 |

---

## 十三、总结

### 13.1 核心亮点

1. **渐进式架构**：MySQL + Redis起步，可平滑演进
2. **性能优化**：二级缓存、热点识别、批量操作
3. **高可用**：降级策略、重试机制、多级容灾
4. **可观测性**：完整的监控指标、告警规则
5. **可扩展性**：支持品类扩展、水平扩展、多区域部署

### 13.2 待优化项（后续版本）

- [ ] 价格趋势图（折线图、柱状图）
- [ ] AI价格预测（需要历史数据积累）
- [ ] 实时价格监控（WebSocket推送）
- [ ] 价格波动告警（用户订阅）
- [ ] 跨品类价格对比（酒店vs民宿）

### 13.3 参考资料

- Google Flights价格日历设计分析
- Booking.com技术博客
- Airbnb Engineering Blog
- Redis官方文档：Hash数据结构最佳实践
- MySQL官方文档：分区表性能优化

---

**设计完成日期：** 2026-04-16  
**下一步：** 创建详细实施计划（Implementation Plan）
