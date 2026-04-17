---
title: 电商系统设计（十四）：价格日历系统设计（Hotel & Flight）
date: '2026-04-16'
categories:
  - 电商系统设计
tags:
  - price-calendar
  - hotel
  - flight
  - redis
  - mysql
  - cache
  - performance
  - system-design
---

> **电商系统设计（十四）**（价格优化专题；总索引见[（一）全景概览与领域划分](/system-design/20-ecommerce-overview/)）
> 
> 本文设计一个价格日历系统，参考 Google Flights、Booking.com、Airbnb 等业界实践，帮助用户快速找到最优惠的预订日期。适用于酒店、机票等按日期定价的商品品类。

---

## 一、需求背景

### 1.1 业务场景

在 B2B2C 电商平台（主营酒店、机票等虚拟商品）中，用户经常面临这样的困惑：

- 同一个酒店，不同日期价格差异巨大（周末 vs 工作日，旺季 vs 淡季）
- 用户需要逐日点击查询，效率低下
- 无法快速找到性价比最高的预订日期

**价格日历**就是为了解决这个问题：用户选定某个酒店或航线后，**一次性展示未来 30-60 天的价格趋势**，帮助用户做出最优决策。

### 1.2 业界参考

**Google Flights 价格日历：**
- 显示价格趋势图（折线图）
- 标注"低于平均价"的日期（绿色高亮）
- 支持灵活日期搜索（+/-3天）
- 价格预测功能（AI 预测未来价格走势）

**Booking.com 价格日历：**
- Hover 显示"从¥299起，有23家酒店"
- 售罄日期显示灰色
- 特价日期高亮显示（红色标签）
- 支持连住优惠提示

**Airbnb 价格日历：**
- 日历显示每晚价格
- 周末价格通常更高（不同颜色）
- 最少入住天数限制提示
- 清洁费分摊显示

### 1.3 核心需求

**用户端需求：**
- 用户选定某个酒店或航线后，查看未来 30-60 天的价格趋势
- 每个日期显示该日期的最低价格
- 点击日期后跳转到该日期的搜索结果页

**运营端需求：**
- 运营团队可以查看价格同步状态
- 支持手动刷新缓存
- 查看同步任务的成功率和失败原因

**非功能需求：**
- 查询响应时间 P99 < 200ms
- 支持 10000 QPS（单实例）
- 缓存命中率 > 80%
- 系统可用性 99.9%

### 1.4 业务约束

- **品类范围：** 初期支持酒店和机票，未来扩展到其他品类
- **时间范围：** 展示未来 30-60 天的价格数据
- **数据来源：** 定时同步供应商价格（非实时查询，避免百万级 SKU 的实时调用）
- **数据规模：** 100 万酒店 + 10 万航线，60 天数据约 6000 万条记录

---

## 二、技术方案选择

### 2.1 备选方案对比

我们评估了三个技术方案：

| 方案 | 核心技术栈 | 优点 | 缺点 | 适用阶段 |
|-----|-----------|------|------|---------|
| **方案1** | MySQL + Redis | 技术栈成熟、快速上线、运维成本低 | 数据量增长需分库分表 | 初期（0-1） |
| **方案2** | MySQL 主从 + Redis + MQ | 读写分离、高并发、削峰填谷 | 架构复杂、主从延迟 | 成熟期（1-10） |
| **方案3** | TimescaleDB + Redis | 时序优化、高压缩率、长期存储 | 学习成本高、运维复杂 | 长期（10+） |

### 2.2 最终选择：方案1（MySQL + Redis）⭐

**理由：**
1. **快速上线：** 团队熟悉 MySQL 和 Redis，2-3 周即可完成开发
2. **风险可控：** 利用现有基础设施，无需引入新组件
3. **渐进式演进：** 后续可平滑升级到方案2或方案3

**演进路径：**
```
阶段1（现在）：MySQL + Redis，支撑 10 万 QPS
         ↓
阶段2（6个月）：加入主从分离 + 消息队列
         ↓
阶段3（1年）：迁移到 TimescaleDB（如需长期历史数据）
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
- 提供 C 端用户查询价格日历的 API
- 提供运营后台的价格管理 API
- 负责查询逻辑：Redis → MySQL 降级
- 水平扩展支持（无状态服务）

**Price Sync Service（价格同步服务）**
- 定时任务拉取供应商价格数据
- 批量写入 MySQL
- 异步刷新 Redis 热点数据
- 支持全量同步和增量更新

**Redis 层**
- 存储热门 SKU 的近 7 天价格
- Key 结构：`price:{category}:{sku_id}` (Hash 类型)
- 自动过期（TTL 7天）

**MySQL 层**
- 主库：价格数据持久化
- 保留 60 天数据，自动清理过期数据
- 索引优化：(sku_id, date) 复合索引

---

## 四、数据模型设计

### 4.1 MySQL 表结构

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
- `sku_id`：酒店 ID（如"hotel_123456"）或航线 ID（如"flight_PEK_SHA"）
- `date`：关键字段，酒店表示入住日期，机票表示出发日期
- `min_price`：该日期的最低价格（跨多个供应商的最低值）
- `available_count`：可售数量，用于判断是否售罄

### 4.2 Redis 数据结构

**热点 SKU 价格缓存（Hash 结构）**

```
Key: price:hotel:{sku_id}
Type: Hash
Fields:
  2026-04-20 → "299.00|sp001|10"  (价格|供应商ID|库存)
  2026-04-21 → "350.00|sp002|5"
  2026-04-22 → "280.00|sp001|20"
  ...
TTL: 7天
```

**热点 SKU 列表（用于判断哪些 SKU 需要缓存）**

```
Key: hotkeys:hotel
Type: ZSet (Sorted Set)
Members:
  hotel_123456 score:10000 (访问次数)
  hotel_789012 score:8500
  ...
TTL: 1小时
```

### 4.3 数据容量估算

**数据量：**
- 酒店：100 万个 × 60 天 = 6000 万条记录
- 机票：10 万条航线 × 60 天 = 600 万条记录
- 合计：约 6600 万条记录

**存储空间：**
- 单条记录大小：约 100 bytes（不含索引）
- 数据大小：6600 万 × 100 bytes ≈ 6.6 GB
- 索引大小：约 2 倍数据大小 ≈ 13 GB
- MySQL 总空间：约 20 GB（含冗余）

**Redis 内存：**
- 热点 SKU 数量：100 万（Top 10%）
- 每个 SKU 7 天数据：7 × 100 bytes = 700 bytes
- Hash 结构开销：约 20%
- 总内存：100 万 × 700 × 1.2 ≈ 840 MB
- 推荐配置：4 GB（预留 3 倍空间）

---

## 五、API 设计

### 5.1 客户端查询 API

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
```

### 5.2 限流策略

```
客户端查询 API：
- 单用户：100 QPS
- 单 IP：500 QPS
- 全局：50000 QPS

运营管理 API：
- IP 白名单
- 刷新操作：10 次/小时
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
    sem := make(chan struct{}, 100) // 并发控制
    
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

### 6.3 热点 SKU 识别流程

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

**过期数据：** 日期早于当前日期的价格记录（`date < CURDATE()`）

用户不会查询过去的价格，这些数据对 C 端无价值，但占用大量存储空间。

### 7.2 清理策略

**MySQL 清理：**

```sql
-- 定时删除（每天凌晨2点）
DELETE FROM price_calendar 
WHERE date < DATE_SUB(CURDATE(), INTERVAL 3 DAY)
LIMIT 10000;  -- 分批删除，避免锁表

-- 分区删除（每月1号删除3个月前的分区）
ALTER TABLE price_calendar DROP PARTITION p202601;
```

**Redis 清理：**
- 自动过期：所有 Key 设置 TTL=7天，Redis 自动清理
- 主动清理：每天凌晨 3 点清理 Hash 中的过期日期字段

**监控指标：**
```
- price_calendar_oldest_date: 数据库中最早的日期（应该>=当前日期-3天）
- price_calendar_expired_rows_deleted: 每天删除的过期记录数
```

**告警规则：**
```
如果 oldest_date < CURDATE() - 7天，触发告警（清理任务失败）
```

---

## 八、性能优化策略

### 8.1 数据库优化

**索引优化：**
- 核心索引：`uk_sku_date (sku_id, date)` 覆盖 90% 查询
- 辅助索引：`idx_date (date)` 用于清理过期数据

**批量查询优化：**
```sql
-- 拆分为多个单SKU查询并并发执行
SELECT * FROM price_calendar WHERE sku_id=? AND date BETWEEN ? AND ?;
```

**分库分表策略（数据量>5000万时）：**
- 分表键：sku_id（Hash 取模）
- 分表数量：16 个表
- 路由逻辑：`table_index = crc32(sku_id) % 16`

### 8.2 Redis 优化

**内存优化：**
- 推荐配置：4 GB 内存
- 最大内存策略：allkeys-lru（自动淘汰最少使用的 Key）

**连接池配置：**
```go
MaxIdle:     100
MaxActive:   500
IdleTimeout: 300s
```

**缓存预热：**
- 系统启动时预热 Top 1000 热点 SKU
- 从 MySQL 加载近 7 天数据写入 Redis

### 8.3 供应商 API 调用优化

**并发控制：**
- 信号量控制并发数：最多 100 个并发请求
- 超时控制：单个请求 3 秒超时
- 熔断器：错误率>50% 触发熔断，熔断时间 5 秒

**重试策略：**
- 最多重试 2 次
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
- MySQL 批量插入：1000 条/批
- 使用 `INSERT ... ON DUPLICATE KEY UPDATE`

### 8.5 性能指标目标

```
查询接口：
  - P50 延迟：< 50ms
  - P99 延迟：< 200ms
  - 吞吐量：10000 QPS/实例
  - 缓存命中率：> 80%

同步任务：
  - 100万SKU同步时间：< 30分钟
  - 供应商API成功率：> 95%
  - 数据库写入速度：> 5000条/秒
```

---

## 九、错误处理与容灾

### 9.1 降级策略

**多级降级：**
```
L1: Redis故障 → 降级到MySQL查询
L2: MySQL从库故障 → 降级到主库查询
L3: 全部数据库故障 → 返回静态默认数据
```

**降级开关：**
- 通过配置中心管理降级开关
- 支持动态开启/关闭 Redis、MySQL

### 9.2 数据一致性保证

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

### 9.3 监控与告警

**Prometheus 指标：**
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
- P99 延迟 > 500ms，持续 5 分钟 → warning
- 缓存命中率 < 60%，持续 10 分钟 → warning
- 同步失败率 > 100 个SKU/小时 → critical
- MySQL 连接数 > 180 → warning
- 数据堆积 > 1 亿条 → critical
```

---

## 十、扩展性设计

### 10.1 水平扩展

**应用层扩展：**
```
负载均衡 → Price Calendar Service 集群（3-20实例）
基于CPU使用率自动扩缩容（K8s HPA）
目标CPU：60%
```

**Redis 扩展：**
```
单机(4GB) → 主从(读写分离) → Cluster(分片16节点)
```

**MySQL 扩展：**
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

### 10.3 业界对比

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

## 十一、总结

### 11.1 核心亮点

1. **渐进式架构**：MySQL + Redis 起步，可平滑演进到分库分表、时序数据库
2. **性能优化**：二级缓存（Redis → MySQL）、热点识别、批量操作
3. **高可用**：降级策略、重试机制、多级容灾
4. **可观测性**：完整的监控指标、告警规则、日志追踪
5. **可扩展性**：支持品类扩展、水平扩展、多区域部署

### 11.2 技术栈

```
语言：Go 1.21+
数据库：MySQL 8.0
缓存：Redis 7.0
监控：Prometheus + Grafana
日志：ELK Stack
部署：Kubernetes + Docker
```

### 11.3 实施计划

```
Phase 1（2周）：基础功能
  - 数据模型设计
  - MySQL 表结构
  - 基础查询 API
  
Phase 2（2周）：缓存优化
  - Redis 缓存层
  - 热点识别
  - 缓存预热
  
Phase 3（2周）：同步服务
  - 供应商 API 调用
  - 批量写入优化
  - 定时任务调度
  
Phase 4（1周）：监控告警
  - Prometheus 指标
  - Grafana 看板
  - 告警规则
  
Phase 5（1周）：压测优化
  - 性能压测
  - 瓶颈分析
  - 优化调优

总计：8 周
```

### 11.4 待优化项（后续版本）

- 价格趋势图（折线图、柱状图）
- AI 价格预测（需要历史数据积累）
- 实时价格监控（WebSocket 推送）
- 价格波动告警（用户订阅）
- 跨品类价格对比（酒店 vs 民宿）

---

## 参考资料

- Google Flights 价格日历设计分析
- Booking.com 技术博客
- Airbnb Engineering Blog
- Redis 官方文档：Hash 数据结构最佳实践
- MySQL 官方文档：分区表性能优化

---

> **相关文章：**
> - [（一）全景概览与领域划分](/system-design/20-ecommerce-overview/)
> - [（三）库存系统](/system-design/22-ecommerce-inventory/)
> - [（五）定价引擎（Pricing Engine）](/system-design/24-ecommerce-pricing-engine/)
> - [（六）定价域的 DDD 建模](/system-design/25-ecommerce-pricing-ddd/)
