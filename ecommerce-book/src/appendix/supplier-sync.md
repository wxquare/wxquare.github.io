# 附录F 供应商数据同步链路

## 1. 背景

数字商品平台需要从外部供应商同步供给数据。本方案讨论的是一条通用的供应商数据同步链路，并以**酒店供给全量同步**为例展开。酒店数据规模大、结构复杂、变化频率不一致：酒店名称、地址、设施、图片等静态信息变化较慢；房型、套餐、最低价、可售状态等半动态信息需要更高频刷新；下单前房态房价必须实时确认。

本设计聚焦一个典型任务：

```text
通过遍历所有城市，从供应商拉取酒店信息
酒店规模约 100 万
任务预计运行 10 小时
需要支持断点续跑、失败补偿、数据追溯和质量监控
```

这类任务不能只依赖进程内状态做一个长循环。第一阶段更推荐设计成 **Batch + Checkpoint + DLQ** 的可恢复流水线：任务可以按城市和分页顺序遍历，进度持久化在数据库里，失败后从 checkpoint 继续。任务分片和分布式 Worker 抢占可以作为后续优化项目，而不是一开始就进入主链路。

## 2. 设计目标

1. **可恢复**：任务中断后可以从 checkpoint 继续，不从头重跑。
2. **可追溯**：保存供应商原始数据 Raw Snapshot，支持问题排查和回放。
3. **可治理**：通过标准化、质量校验、Diff、版本控制，避免错误数据污染平台模型。
4. **可补偿**：失败数据进入 DLQ，支持自动重试、人工修复和重新投递。
5. **可观测**：实时查看任务进度、失败原因、供应商质量和业务影响指标。
6. **不影响交易安全**：列表页可缓存，详情页更接近实时，创单前必须实时确认。

## 3. 核心难点

| 难点 | 说明 | 设计策略 |
|------|------|----------|
| 任务时间长 | 100 万酒店跑 10 小时，中途失败概率高 | Batch + Page/Cursor Checkpoint |
| 数据量大 | 全量同步可能包含酒店、房型、图片、设施等大 payload | Raw Snapshot 存引用，主表保持轻量 |
| 供应商不稳定 | 超时、限流、5xx、分页游标失效 | 限流、熔断、指数退避、DLQ |
| 模型不一致 | 供应商酒店/房型/套餐与平台 Resource/SPU/SKU/Offer 不一致 | 标准化映射 + supplier mapping |
| 数据质量不稳定 | 字段缺失、城市映射失败、价格异常、坐标漂移 | 分层质量校验 + 部分成功 |
| 发布风险 | 同步成功不代表可以发布 | sync version、snapshot version、publish version 分离 |
| 下游一致性 | DB 更新成功但 ES、缓存、事件可能失败 | Outbox + 索引补偿 |

## 4. 总体架构

```text
Full Sync Task
  → Sync Batch
  → Page Fetch
  → Raw Snapshot
  → Normalize
  → Quality Check
  → Resource Mapping
  → Diff
  → Publish
  → Search / Cache / Downstream Event
  → Metrics / DLQ / Compensation
```

架构图见：

![供应商数据同步链路架构图](../../images/supplier-sync-architecture.png)

Data Flow Diagram 见：

![供应商数据同步 Data Flow Diagram](../../images/supplier-sync-data-flow.png)

图文件：

- `ecommerce-book/images/supplier-sync-architecture.png`
- `ecommerce-book/images/supplier-sync-architecture.svg`
- `ecommerce-book/images/supplier-sync-data-flow.png`
- `ecommerce-book/images/supplier-sync-data-flow.svg`

## 5. 任务模型

### 5.1 Task：同步任务定义

`supplier_sync_task` 描述“要同步什么、怎么同步、多久同步一次”。

```sql
CREATE TABLE supplier_sync_task (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_code VARCHAR(64) NOT NULL,
    supplier_id BIGINT NOT NULL,
    category_code VARCHAR(32) NOT NULL,
    sync_mode VARCHAR(32) NOT NULL COMMENT 'FULL/INCREMENTAL/PUSH/REFRESH',
    data_scope VARCHAR(64) NOT NULL COMMENT 'RESOURCE/PRODUCT/OFFER/STOCK_PRICE',
    schedule_type VARCHAR(32) NOT NULL COMMENT 'CRON/MANUAL/PUSH',
    cron_expr VARCHAR(64) DEFAULT NULL,
    status VARCHAR(32) NOT NULL COMMENT 'ENABLED/DISABLED',
    concurrency_policy VARCHAR(32) NOT NULL DEFAULT 'SKIP_IF_RUNNING'
        COMMENT 'SKIP_IF_RUNNING/CANCEL_PREVIOUS/ALLOW_PARALLEL',
    last_batch_id VARCHAR(64) DEFAULT NULL,
    owner_team VARCHAR(64) DEFAULT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_task_code (task_code),
    KEY idx_supplier_category (supplier_id, category_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商同步任务定义';
```

样例：

```text
task_code: hotel_supplier_full_resource
supplier_id: 1001
category_code: HOTEL
sync_mode: FULL
data_scope: RESOURCE
schedule_type: MANUAL
status: ENABLED
concurrency_policy: SKIP_IF_RUNNING
owner_team: product-sync
```

### 5.2 Batch：一次任务执行批次

`supplier_sync_batch` 记录一次任务执行的状态、水位、统计和版本。

```sql
CREATE TABLE supplier_sync_batch (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    batch_id VARCHAR(64) NOT NULL,
    task_code VARCHAR(64) NOT NULL,
    trigger_source VARCHAR(32) NOT NULL COMMENT 'CRON/MANUAL/COMPENSATION',
    trigger_id VARCHAR(64) DEFAULT NULL COMMENT '外部触发幂等 ID',
    supplier_id BIGINT NOT NULL,
    category_code VARCHAR(32) NOT NULL,
    sync_mode VARCHAR(32) NOT NULL,
    data_scope VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL COMMENT 'PENDING/RUNNING/SUCCESS/PARTIAL_FAILED/FAILED/CANCELLED',
    sync_batch_version BIGINT NOT NULL,
    start_checkpoint VARCHAR(512) DEFAULT NULL,
    end_checkpoint VARCHAR(512) DEFAULT NULL,
    total_count INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    failed_count INT NOT NULL DEFAULT 0,
    skipped_count INT NOT NULL DEFAULT 0,
    current_city_code VARCHAR(64) DEFAULT NULL,
    current_page INT DEFAULT NULL,
    progress_percent DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    worker_id VARCHAR(64) DEFAULT NULL,
    lease_token VARCHAR(64) DEFAULT NULL,
    lease_until DATETIME DEFAULT NULL,
    heartbeat_at DATETIME DEFAULT NULL,
    last_heartbeat_stage VARCHAR(64) DEFAULT NULL,
    last_heartbeat_message VARCHAR(512) DEFAULT NULL,
    last_checkpoint_at DATETIME DEFAULT NULL,
    created_at DATETIME NOT NULL,
    started_at DATETIME DEFAULT NULL,
    finished_at DATETIME DEFAULT NULL,
    updated_at DATETIME NOT NULL,
    error_message VARCHAR(1024) DEFAULT NULL,
    UNIQUE KEY uk_batch_id (batch_id),
    UNIQUE KEY uk_task_trigger (task_code, trigger_id),
    KEY idx_task_status (task_code, status),
    KEY idx_status_lease (status, lease_until),
    KEY idx_supplier_time (supplier_id, started_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商同步批次';
```

样例：

```text
batch_id: batch_20260427_hotel_full_001
task_code: hotel_supplier_full_resource
trigger_source: MANUAL
trigger_id: req_20260427_0001
supplier_id: 1001
category_code: HOTEL
sync_mode: FULL
data_scope: RESOURCE
status: RUNNING
sync_batch_version: 202604270001
total_count: 1000000
success_count: 688200
failed_count: 320
skipped_count: 12000
current_city_code: BKK
current_page: 120
progress_percent: 68.82
worker_id: hotel-sync-worker-pod-a1b2c3-12345-20260427T103000Z
lease_token: 7f2d4c77-5d5b-4f1f-aeb0-74f7f21c6e2a
lease_until: 2026-04-27 10:35:00
heartbeat_at: 2026-04-27 10:30:00
last_heartbeat_stage: FETCHING
last_heartbeat_message: fetching city=BKK page=120
last_checkpoint_at: 2026-04-27 10:29:50
```

## 6. 任务创建、互斥与执行恢复

### 6.1 任务创建流程

一次同步任务通常由定时调度、运营手动触发或系统补偿触发。无论来源是什么，都不应该直接启动一个进程开始跑，而是先创建 batch，再由执行器领取 batch。

```text
触发同步
  → 查询 supplier_sync_task
  → 检查任务是否 ENABLED
  → 检查 trigger_id 幂等
  → 检查互斥策略
  → 创建 supplier_sync_batch(status=PENDING)
  → 执行器抢占 batch
  → 执行同步
```

创建 batch 时要初始化：

| 字段 | 说明 |
|------|------|
| `batch_id` | 本次执行唯一 ID |
| `trigger_source` / `trigger_id` | 触发来源和外部请求幂等 ID |
| `sync_batch_version` | 本次同步批次版本 |
| `status` | 初始为 `PENDING` |
| `start_checkpoint` | 本次任务起点，通常为空或上次成功水位 |
| `end_checkpoint` | 当前进度，任务执行过程中不断推进 |
| `total_count` | 预计处理数量，可先为空或估算 |
| `worker_id` / `lease_token` | 执行器抢占后写入 |

任务创建也要做幂等。运营后台重复点击、调度器重试、网络超时后重发，都可能重复触发同一个任务。推荐由调用方传入 `trigger_id`，例如运营后台的 `manual_request_id` 或调度系统的 `fire_id`：

```text
同一个 task_code + trigger_id
  → 只允许创建一个 batch
  → 重复请求直接返回已存在 batch
```

如果是定时任务，可以用计划触发时间生成 `trigger_id`：

```text
trigger_id = hotel_supplier_full_resource:2026-04-27T02:00:00Z
```

### 6.2 上一次任务还没执行完怎么办

同一个供应商、同一个品类、同一个数据范围的全量任务，通常不应该同时跑多个，否则会造成供应商限流、重复写入、发布版本乱序和进度混乱。这里需要显式定义互斥策略。

| 策略 | 含义 | 适用场景 |
|------|------|----------|
| `SKIP_IF_RUNNING` | 如果已有运行中的 batch，新触发直接跳过 | 定时全量同步、普通刷新 |
| `CANCEL_PREVIOUS` | 取消旧 batch，启动新 batch | 人工修复后需要重新跑全量 |
| `ALLOW_PARALLEL` | 允许并行，但必须保证数据范围不重叠 | 不同城市、不同供应商、不同数据 scope |

默认建议使用 `SKIP_IF_RUNNING`。创建 batch 前先检查：

```sql
SELECT batch_id, status, heartbeat_at, lease_until
FROM supplier_sync_batch
WHERE task_code = ?
  AND status IN ('PENDING', 'RUNNING')
ORDER BY created_at DESC
LIMIT 1;
```

如果存在未完成 batch：

```text
concurrency_policy = SKIP_IF_RUNNING
  → 不创建新 batch，记录 SKIPPED 日志

concurrency_policy = CANCEL_PREVIOUS
  → 将旧 batch 标记 CANCELLED
  → 创建新 batch

concurrency_policy = ALLOW_PARALLEL
  → 检查数据范围是否重叠
  → 不重叠才允许创建
```

面试时可以强调：**重复下发不是靠“大家约定别点两次”解决，而要靠任务互斥策略和数据库状态控制解决**。

### 6.3 Batch 抢占

即使第一阶段不做任务分片，也建议 batch 由执行器通过 CAS 抢占，避免多个进程同时执行同一个 batch。抢占不是“查出来再更新”，而是用一条带条件的 `UPDATE` 完成。

```sql
UPDATE supplier_sync_batch
SET status = 'RUNNING',
    worker_id = ?,
    lease_token = ?,
    lease_until = DATE_ADD(NOW(), INTERVAL 5 MINUTE),
    heartbeat_at = NOW(),
    last_heartbeat_stage = 'CLAIMED',
    last_heartbeat_message = 'batch claimed',
    started_at = IFNULL(started_at, NOW()),
    updated_at = NOW()
WHERE batch_id = ?
  AND status = 'PENDING';
```

`rows_affected = 1` 表示抢占成功；`rows_affected = 0` 表示已经被其他执行器抢走，当前 worker 必须放弃执行。

对于机器重启、进程 OOM、发布中断后遗留的 `RUNNING` batch，可以允许抢占 lease 已经过期的 batch：

```sql
UPDATE supplier_sync_batch
SET worker_id = ?,
    lease_token = ?,
    lease_until = DATE_ADD(NOW(), INTERVAL 5 MINUTE),
    heartbeat_at = NOW(),
    last_heartbeat_stage = 'RECLAIMED',
    last_heartbeat_message = 'expired batch reclaimed',
    updated_at = NOW()
WHERE batch_id = ?
  AND status = 'RUNNING'
  AND lease_until < NOW();
```

注意，这里只抢占“租约过期”的任务，不抢占“心跳正常”的任务。否则一个慢请求、一次 GC 或一次网络抖动都可能导致双 worker 写同一个 batch。

### 6.4 `worker_id` 与 `lease_token`

`worker_id` 用来标识“哪个执行器实例在跑任务”，`lease_token` 用来标识“本次抢占的所有权”。两者要同时使用。

| 字段 | 作用 | 是否稳定 |
|------|------|----------|
| `worker_id` | 标识执行器实例，方便排查、日志关联和监控展示 | 进程生命周期内稳定 |
| `lease_token` | 标识一次抢占行为，防止旧 worker 恢复后覆盖新 worker | 每次抢占重新生成 |

`worker_id` 可以用“服务名 + 机器/容器名 + 进程号 + 启动时间”生成：

```go
func GenerateWorkerID(serviceName string) string {
    host := os.Getenv("POD_NAME")
    if host == "" {
        host = os.Getenv("HOSTNAME")
    }
    if host == "" {
        host, _ = os.Hostname()
    }

    pid := os.Getpid()
    startedAt := time.Now().UTC().Format("20060102T150405Z")
    return fmt.Sprintf("%s-%s-%d-%s", serviceName, host, pid, startedAt)
}
```

示例：

```text
worker_id   = hotel-sync-worker-pod-a1b2c3-12345-20260427T103000Z
lease_token = 7f2d4c77-5d5b-4f1f-aeb0-74f7f21c6e2a
```

为什么还需要 `lease_token`？因为容器名或机器名可能复用，旧进程在长 GC 后也可能恢复。只有 `worker_id` 不够严格；`lease_token` 能保证“只有当前这次抢占的持有者”才能续租、推进 checkpoint 和结束任务。

所有关键更新都必须带上三个条件：

```sql
WHERE batch_id = ?
  AND worker_id = ?
  AND lease_token = ?
```

如果更新影响行数为 0，要立即停止当前任务，并记录 `LEASE_LOST` 日志。

### 6.5 心跳与租约

长任务不能只依赖 `status=RUNNING` 判断是否还活着。机器重启、进程 OOM、发布重启都可能导致状态永远卡在 `RUNNING`。因此 batch 要同时有“租约”和“心跳”。

| 概念 | 解决的问题 | 典型字段 |
|------|------------|----------|
| 心跳 Heartbeat | worker 是否还活着 | `heartbeat_at`、`last_heartbeat_stage` |
| 租约 Lease | 当前谁拥有任务执行权 | `worker_id`、`lease_token`、`lease_until` |
| Checkpoint | 任务恢复时从哪里继续 | `end_checkpoint`、`last_checkpoint_at` |

执行器每 15 到 30 秒续租一次，租约建议设置为 2 到 5 分钟。心跳间隔要远小于租约时长，给短暂网络抖动留下余量。

```sql
UPDATE supplier_sync_batch
SET heartbeat_at = NOW(),
    lease_until = DATE_ADD(NOW(), INTERVAL 5 MINUTE),
    last_heartbeat_stage = ?,
    last_heartbeat_message = ?,
    updated_at = NOW()
WHERE batch_id = ?
  AND worker_id = ?
  AND lease_token = ?
  AND status = 'RUNNING';
```

心跳建议上报的不只是“我还活着”，还要包含当前阶段：

| 阶段 | 含义 | 示例 message |
|------|------|--------------|
| `FETCHING` | 正在请求供应商接口 | `fetching city=BKK page=120` |
| `SNAPSHOT_SAVING` | 正在保存 Raw Snapshot | `saving raw snapshot page=120` |
| `NORMALIZING` | 正在做字段标准化 | `normalizing 100 hotels` |
| `VALIDATING` | 正在做质量校验 | `validating schema and city mapping` |
| `PUBLISHING` | 正在发布平台模型 | `publishing resource changes` |
| `CHECKPOINTING` | 正在推进 checkpoint | `checkpoint to page=121` |

如果心跳更新失败：

```text
rows_affected = 0
  → 当前 worker 不再拥有任务
  → 停止拉取供应商
  → 停止写平台表
  → 打印 LEASE_LOST 日志
  → 退出执行
```

这一步非常关键。不能因为“当前进程还活着”就继续跑，因为数据库里的执行权可能已经被新 worker 抢走。

### 6.6 心跳正常但 Checkpoint 不动怎么办

心跳和 checkpoint 是两个维度。心跳正常只能说明 worker 还活着，不代表任务在前进。可能出现：

1. 供应商接口一直卡在慢请求。
2. 某个城市数据量异常大。
3. Raw Snapshot 存储变慢。
4. 发布阶段被数据库锁阻塞。
5. worker 进入了内部死循环，但心跳线程仍然正常。

因此需要同时监控：

```text
heartbeat_lag = now - heartbeat_at
checkpoint_lag = now - last_checkpoint_at
```

处理策略：

| 现象 | 判断 | 动作 |
|------|------|------|
| `heartbeat_lag` 超过租约 | worker 失联 | 允许新 worker 抢占 |
| `heartbeat_lag` 正常，`checkpoint_lag` 过大 | worker 活着但进度卡住 | 告警，不立即抢占 |
| `heartbeat_lag` 正常，阶段长期不变 | 某阶段阻塞 | 根据阶段定位供应商、存储或发布问题 |

不要在心跳正常时强行抢占。否则可能造成两个 worker 同时处理同一页，只是其中一个更慢。

### 6.7 机器重启后如何恢复

机器重启后，原 worker 不再续租。调度器或新 worker 会发现：

```sql
SELECT batch_id
FROM supplier_sync_batch
WHERE status = 'RUNNING'
  AND lease_until < NOW();
```

恢复流程：

```text
worker-01 执行 batch
  → 机器重启，心跳停止
  → lease_until 过期
  → worker-02 生成新的 worker_id 和 lease_token
  → worker-02 抢占过期 batch
  → 读取 end_checkpoint
  → 从 city/page/cursor 继续
```

这时可能重复处理上一页，所以处理逻辑必须幂等：

```text
supplier_id + supplier_resource_code + supplier_product_code
```

Checkpoint 负责减少重跑范围，幂等负责保证重复处理也不会写错。

### 6.8 进度上报

进度不要只写日志，要落到 batch 表，便于运营后台、告警系统和排查工具读取。

每处理完一页，更新 checkpoint、统计、进度和心跳：

```sql
UPDATE supplier_sync_batch
SET end_checkpoint = ?,
    current_city_code = ?,
    current_page = ?,
    success_count = success_count + ?,
    failed_count = failed_count + ?,
    skipped_count = skipped_count + ?,
    progress_percent = ?,
    heartbeat_at = NOW(),
    lease_until = DATE_ADD(NOW(), INTERVAL 5 MINUTE),
    last_heartbeat_stage = 'CHECKPOINTING',
    last_heartbeat_message = ?,
    last_checkpoint_at = NOW(),
    updated_at = NOW()
WHERE batch_id = ?
  AND worker_id = ?
  AND lease_token = ?
  AND status = 'RUNNING';
```

上报频率建议按“页”或“固定时间窗口”控制：

| 上报方式 | 优点 | 缺点 |
|----------|------|------|
| 每条酒店上报 | 精确 | DB 写入过多 |
| 每页上报 | 性能和准确性平衡 | 失败时最多重复一页 |
| 每 30 秒上报 | 写入少 | 进度略滞后 |

推荐：**每页处理完成后推进 checkpoint，同时每 15 到 30 秒续租心跳**。如果一页处理时间可能超过心跳间隔，则需要独立心跳协程，不能等整页处理完成才心跳。

### 6.9 边界场景处理

| 场景 | 风险 | 处理 |
|------|------|------|
| 定时任务重复触发 | 同一任务多个 batch 并发 | `concurrency_policy=SKIP_IF_RUNNING` |
| 人工重复点击执行 | 重复创建全量任务 | 用 `task_code + status` 互斥 |
| 机器重启 | batch 卡在 RUNNING | lease 过期后新 worker 抢占 |
| 旧 worker 恢复 | 覆盖新 worker checkpoint | 更新时校验 `worker_id + lease_token` |
| 心跳正常但 checkpoint 不动 | worker 活着但卡住 | 告警定位，不立即抢占 |
| checkpoint 更新失败 | 下次重复处理上一页 | 页内写入必须幂等 |
| checkpoint 先更新后处理失败 | 数据被跳过 | 必须先处理成功再推进 checkpoint |
| 供应商短暂失败 | 任务频繁失败 | 指数退避、限流、熔断 |
| 任务被取消 | 仍有 worker 在跑 | worker 每页检查 batch status |
| 发布新版本 | 进程退出 | checkpoint + lease 恢复 |

## 7. Checkpoint 与断点续跑

### 7.1 为什么需要 Checkpoint

100 万酒店、10 小时任务，如果只把进度放在内存里，会有三个问题：

1. 任务中断后恢复困难。
2. 机器重启后只能从头开始。
3. 进度不可观测，不知道当前卡在哪里。

因此，第一阶段主设计不引入任务分片，而是在 `supplier_sync_batch` 上保存 checkpoint。任务仍然可以按城市和分页遍历，但每处理完一页就推进一次 checkpoint。

```text
batch_001
  → city = BKK, page = 1
  → city = BKK, page = 2
  → ...
  → city = JKT, page = 1
  → ...
```

### 7.2 Checkpoint 存储

Checkpoint 可以先复用 `supplier_sync_batch.start_checkpoint` 和 `supplier_sync_batch.end_checkpoint`，也可以在后续演进中拆出独立 checkpoint 表。

主链路里的 checkpoint 建议记录：

| 字段 | 含义 |
|------|------|
| `city_code` | 当前遍历到哪个城市 |
| `page` | 当前处理到第几页 |
| `cursor` | 供应商返回的下一页游标 |
| `last_supplier_hotel_id` | 上一次成功处理的供应商酒店 ID |
| `success_count` | 当前批次已成功处理数量 |
| `failed_count` | 当前批次失败数量 |
| `updated_at` | checkpoint 更新时间 |

### 7.3 Checkpoint 是什么

Checkpoint 是同步任务“跑到哪里了”的进度记录。它用于断点续跑。

示例：

```json
{
  "city_code": "BKK",
  "page": 120,
  "cursor": "abc123",
  "last_supplier_hotel_id": "H998877"
}
```

如果 Bangkok 第 120 页失败，下次可以从 page 120 或 cursor `abc123` 继续，而不是从第一页重跑。

### 7.4 Checkpoint 怎么使用

推荐顺序是：**先处理本页数据，再推进 checkpoint**。

```text
拉取 BKK page=120
  → 保存 Raw Snapshot
  → 标准化
  → 质量校验
  → 平台模型映射
  → Diff / Publish
  → 本页处理成功
  → checkpoint = BKK page=121
```

不要先推进 checkpoint 再处理数据，否则机器在中间宕机会跳过未处理页面。

机器重启时的恢复流程：

```text
机器重启 / 进程退出
  → 调度器重新启动 batch
  → 读取 batch.end_checkpoint
  → 从 city/page/cursor 继续拉取
  → 已处理过的一页允许重复处理
  → 通过 supplier_id + supplier_resource_code 幂等去重
```

Checkpoint 只能保证“不大范围重跑”，不能保证“绝不重复处理”。因此它必须和幂等设计配合使用。

## 8. 拉取与限流

同步任务按城市和分页拉取：

```text
city = BKK
page_size = 100
page = 1..N
```

容量估算：

```text
1000000 hotels / 10 hours = 27.8 hotels/s
```

如果每页 100 个酒店：

```text
1000000 / 100 = 10000 pages
10000 pages / 10 hours = 0.28 page/s
```

如果需要逐个拉酒店详情：

```text
1000000 detail calls / 10 hours = 27.8 QPS
```

拉取并发度要受供应商限流约束：

```text
fetch_concurrency = min(供应商限流 QPS / 单请求 QPS, 系统处理能力)
```

必须支持：

1. 每供应商限流。
2. 每城市请求节流。
3. 超时控制。
4. 失败指数退避。
5. 供应商异常时熔断。

## 9. Raw Snapshot 与标准化

### 9.1 Raw Snapshot

Raw Snapshot 是供应商原始响应数据的快照。它不是平台商品模型，也不是最终发布数据，而是证据和可回放数据。

作用：

1. 排查问题：线上价格或酒店信息异常时，可以还原供应商当时返回了什么。
2. 支持回放：修复映射规则后，可以用原始数据重新跑同步。
3. 支持 Diff：比较本次和上次数据变化。
4. 明确责任：区分供应商数据错误和平台清洗映射错误。

### 9.2 Snapshot 表

```sql
CREATE TABLE supplier_sync_snapshot (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    snapshot_id VARCHAR(64) NOT NULL,
    batch_id VARCHAR(64) NOT NULL,
    supplier_id BIGINT NOT NULL,
    category_code VARCHAR(32) NOT NULL,
    supplier_resource_code VARCHAR(128) DEFAULT NULL,
    supplier_product_code VARCHAR(128) DEFAULT NULL,
    snapshot_type VARCHAR(32) NOT NULL COMMENT 'RAW/NORMALIZED',
    snapshot_version BIGINT NOT NULL,
    payload_ref VARCHAR(512) DEFAULT NULL,
    payload_hash VARCHAR(64) NOT NULL,
    created_at DATETIME NOT NULL,
    UNIQUE KEY uk_snapshot_id (snapshot_id),
    KEY idx_batch (batch_id),
    KEY idx_supplier_object (supplier_id, supplier_resource_code, supplier_product_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商同步快照';
```

样例：

```text
snapshot_id: rs_20260427_000001
batch_id: batch_20260427_hotel_full_001
supplier_id: 1001
category_code: HOTEL
supplier_resource_code: hotel_8848
supplier_product_code: room_deluxe
snapshot_type: RAW
snapshot_version: 8
payload_ref: s3://hotel-sync/raw/2026/04/27/batch001/BKK/page120.json
payload_hash: 9a0f...e31c
```

### 9.3 标准化

供应商字段需要转换成平台标准模型：

| 供应商字段 | 平台字段 |
|-----------|----------|
| `supplier_hotel_id` | `supplier_resource_code` |
| `hotel_name` | `resource_name` |
| `city_code` | `platform_city_id` |
| `address` | `address` |
| `latitude` | `geo.lat` |
| `longitude` | `geo.lng` |
| `facilities` | `ext_info.facilities` |

标准化后生成 `NORMALIZED` snapshot。

## 10. 质量校验

质量校验分为五层：

| 校验层 | 校验内容 | 失败处理 |
|--------|----------|----------|
| Schema 校验 | 必填字段、类型、枚举、时间格式、货币单位 | 进入失败明细 |
| 主数据校验 | 城市、国家、商圈、品牌是否存在 | 进入人工映射 |
| 模型校验 | 是否能映射 Resource / SPU / SKU / Offer | 阻断发布 |
| 交易校验 | 价格异常、库存异常、可售状态矛盾 | 高风险拦截 |
| 业务规则校验 | 站点、渠道、品类是否允许售卖 | 审核或灰度 |

质量校验要支持部分成功。100 万酒店同步中，不能因为 100 条失败就整批失败。

```text
成功数据：继续发布
失败数据：写入 DLQ
高风险数据：进入审核或人工修复
```

## 11. 平台模型映射

酒店通常作为 `Resource` 沉淀：

```text
supplier_hotel_id
  → supplier_product_mapping
  → platform_resource_id
```

如果 mapping 存在：

```text
更新 resource / ext_info / room 信息
```

如果 mapping 不存在：

```text
创建 resource
创建 supplier mapping
必要时创建 SPU / SKU / Offer
```

酒店同步的核心落库模型：

| 平台模型 | 说明 |
|----------|------|
| `resource_tab` | 酒店资源 |
| `resource_ext_hotel_tab` | 酒店扩展信息，如地址、设施、坐标、评分 |
| `supplier_product_mapping_tab` | 供应商酒店 ID 与平台酒店 ID 的映射 |
| `product_spu_tab` | 需要平台售卖承接时创建 |
| `product_sku_tab` | 固定售卖单元，部分酒店业务可不沉淀完整 SKU |
| `product_offer_tab` | 套餐、房型、房价计划等销售配置 |

## 12. 版本与 Diff

版本分为三类：

| 版本 | 含义 | 用途 |
|------|------|------|
| `sync_batch_version` | 本次同步任务版本 | 排查哪次同步带来了变化 |
| `data_snapshot_version` | 原始/标准化数据快照版本 | 支持回放、diff、回滚 |
| `publish_version` | 平台正式发布版本 | 控制搜索、缓存、下游事件一致性 |

Diff 是标准化后的数据与当前线上发布版本之间的变化。

```text
Normalized Snapshot
  vs
Current Published Resource
```

Diff 类型：

| Diff 类型 | 示例 | 动作 |
|-----------|------|------|
| `NO_CHANGE` | 无变化 | 跳过 |
| `CONTENT_CHANGED` | 酒店名称、地址变化 | 更新详情缓存 |
| `IMAGE_CHANGED` | 图片变化 | 更新图片和缓存 |
| `GEO_CHANGED` | 城市、坐标变化 | 高风险，进入审核 |
| `ROOM_CHANGED` | 房型变化 | 更新房型或 Offer |
| `SELLABILITY_CHANGED` | 可售状态变化 | 刷新可售状态 |

Diff 表：

```sql
CREATE TABLE supplier_sync_diff_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    diff_id VARCHAR(64) NOT NULL,
    batch_id VARCHAR(64) NOT NULL,
    supplier_id BIGINT NOT NULL,
    category_code VARCHAR(32) NOT NULL,
    supplier_resource_code VARCHAR(128) DEFAULT NULL,
    supplier_product_code VARCHAR(128) DEFAULT NULL,
    platform_resource_id BIGINT DEFAULT NULL,
    spu_id BIGINT DEFAULT NULL,
    sku_id BIGINT DEFAULT NULL,
    offer_id BIGINT DEFAULT NULL,
    old_publish_version BIGINT DEFAULT NULL,
    new_snapshot_version BIGINT NOT NULL,
    diff_type VARCHAR(64) NOT NULL COMMENT 'NO_CHANGE/CONTENT_CHANGED/PRICE_CHANGED/STOCK_CHANGED/RULE_CHANGED',
    changed_fields JSON NOT NULL,
    risk_level VARCHAR(32) NOT NULL COMMENT 'LOW/MEDIUM/HIGH',
    action VARCHAR(64) NOT NULL COMMENT 'IGNORE/AUTO_PUBLISH/REVIEW/DLQ',
    created_at DATETIME NOT NULL,
    UNIQUE KEY uk_diff_id (diff_id),
    KEY idx_batch (batch_id),
    KEY idx_action (action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商同步差异日志';
```

样例：

```text
diff_id: diff_20260427_000001
batch_id: batch_20260427_hotel_full_001
supplier_id: 1001
category_code: HOTEL
supplier_resource_code: hotel_8848
platform_resource_id: 50001
old_publish_version: 22
new_snapshot_version: 8
diff_type: CONTENT_CHANGED
changed_fields:
[
  {"field": "address", "old": "Old Road", "new": "New Road"},
  {"field": "facilities", "old": ["wifi"], "new": ["wifi", "pool"]}
]
risk_level: LOW
action: AUTO_PUBLISH
```

## 13. 发布与下游刷新

发布时生成新的 `publish_version`：

```text
resource_id = 50001
old_publish_version = 21
new_publish_version = 22
```

发布后通过 Outbox 发事件：

```text
HotelResourceUpdated
HotelMappingCreated
HotelContentChanged
HotelSearchIndexRefreshRequired
```

下游动作：

1. 搜索索引刷新。
2. 详情缓存失效。
3. 商品质量报表更新。
4. 数据平台 CDC。
5. 营销、计价、订单读取新版本商品上下文。

## 14. DLQ 与补偿

### 14.1 为什么用 MySQL DLQ

酒店同步失败通常不是单纯消息失败，而是字段缺失、映射失败、价格异常、发布失败、索引失败等需要人工修复、状态流转和审计的问题。因此推荐：

```text
Kafka DLQ：短期失败消息缓冲，可选
MySQL DLQ：权威问题单和补偿状态
```

### 14.2 DLQ 表

```sql
CREATE TABLE supplier_sync_dead_letter (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    dead_letter_id VARCHAR(64) NOT NULL,
    batch_id VARCHAR(64) NOT NULL,
    task_code VARCHAR(64) NOT NULL,
    sync_mode VARCHAR(32) NOT NULL,
    category_code VARCHAR(32) NOT NULL,
    supplier_id BIGINT NOT NULL,
    supplier_resource_code VARCHAR(128) DEFAULT NULL,
    supplier_product_code VARCHAR(128) DEFAULT NULL,
    platform_resource_id BIGINT DEFAULT NULL,
    spu_id BIGINT DEFAULT NULL,
    sku_id BIGINT DEFAULT NULL,
    offer_id BIGINT DEFAULT NULL,
    error_stage VARCHAR(64) NOT NULL COMMENT 'ADAPTER/VALIDATION/MAPPING/PUBLISH/INDEX',
    error_type VARCHAR(64) NOT NULL COMMENT 'RETRYABLE/NON_RETRYABLE/MAPPING_REQUIRED/RISK_BLOCKED',
    error_code VARCHAR(128) NOT NULL,
    error_message VARCHAR(1024) NOT NULL,
    raw_payload_ref VARCHAR(512) DEFAULT NULL,
    raw_payload_hash VARCHAR(64) DEFAULT NULL,
    normalized_payload_ref VARCHAR(512) DEFAULT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING'
        COMMENT 'PENDING/RETRYING/MANUAL_FIX/RESOLVED/IGNORED/FAILED',
    retry_count INT NOT NULL DEFAULT 0,
    max_retry_count INT NOT NULL DEFAULT 5,
    next_retry_at DATETIME DEFAULT NULL,
    last_retry_at DATETIME DEFAULT NULL,
    owner_team VARCHAR(64) DEFAULT NULL,
    assignee VARCHAR(64) DEFAULT NULL,
    fix_note VARCHAR(1024) DEFAULT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    resolved_at DATETIME DEFAULT NULL,
    UNIQUE KEY uk_dead_letter_id (dead_letter_id),
    UNIQUE KEY uk_dedup (
        batch_id,
        supplier_id,
        supplier_resource_code,
        supplier_product_code,
        error_stage,
        raw_payload_hash
    ),
    KEY idx_status_next_retry (status, next_retry_at),
    KEY idx_supplier_status (supplier_id, status),
    KEY idx_category_status (category_code, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商同步死信队列';
```

样例：

```text
dead_letter_id: dlq_20260427_000001
batch_id: batch_20260427_hotel_full_001
task_code: hotel_supplier_full_resource
sync_mode: FULL
category_code: HOTEL
supplier_id: 1001
supplier_resource_code: hotel_8848
error_stage: MAPPING
error_type: MAPPING_REQUIRED
error_code: CITY_NOT_FOUND
error_message: supplier city code BKK-OLD cannot map to platform city
raw_payload_ref: s3://hotel-sync/raw/2026/04/27/batch001/BKK/page120.json
status: MANUAL_FIX
owner_team: product-sync
assignee: ops_user_01
```

### 14.3 状态机

```text
PENDING
  → RETRYING
  → RESOLVED

PENDING
  → MANUAL_FIX
  → RETRYING
  → RESOLVED

PENDING
  → IGNORED

RETRYING
  → FAILED
```

### 14.4 补偿 Job

```sql
SELECT *
FROM supplier_sync_dead_letter
WHERE status IN ('PENDING', 'FAILED')
  AND next_retry_at <= NOW()
  AND retry_count < max_retry_count
ORDER BY next_retry_at ASC
LIMIT 100;
```

重试时间使用指数退避：

```text
next_retry_at = now + min(2^retry_count minutes, 1 hour)
```

## 15. 监控指标

| 指标类型 | 指标 |
|----------|------|
| 任务进度 | 总城市数、已完成城市数、当前城市、当前 page/cursor |
| 处理统计 | 酒店总数、成功数、失败数、跳过数 |
| 性能指标 | 任务耗时、供应商 QPS、平均耗时、P99 耗时 |
| 质量指标 | 字段缺失率、映射失败率、重复数据率、异常价格率 |
| 新鲜度指标 | 数据延迟、过期数据比例、热门酒店刷新延迟 |
| 补偿指标 | DLQ 数量、重试成功率、人工修复数量 |
| 下游指标 | ES 刷新失败数、缓存刷新失败数、事件发布失败数 |

核心指标公式：

```text
同步成功率 = 成功处理酒店数 / 总酒店数
映射失败率 = 映射失败酒店数 / 总酒店数
字段缺失率 = 缺失关键字段酒店数 / 总酒店数
数据新鲜度延迟 = now - last_success_sync_time
DLQ 修复率 = resolved_dlq_count / total_dlq_count
```

## 16. 异常场景

| 异常 | 处理 |
|------|------|
| 某城市同步失败 | 从该城市对应 checkpoint 继续 |
| 某页接口超时 | 从 page checkpoint 重试 |
| 单个酒店字段缺失 | 写入 DLQ，不阻塞整批 |
| 供应商限流 | 降低 worker 数，指数退避 |
| 城市映射失败 | 进入人工映射，修复后重新投递 |
| ES 刷新失败 | Outbox 补偿重试 |
| 发布版本异常 | 保留旧版本，新版本不生效 |

## 17. 面试总结

> 100 万酒店、10 小时的供应商全量同步任务，我不会设计成依赖进程内状态的单进程长循环，而会先设计成 Batch + Page/Cursor Checkpoint 的可恢复流水线。供应商原始响应先保存 Raw Snapshot，然后做标准化、质量校验、平台模型映射、Diff 和发布。同步版本、快照版本和发布版本要分离，保证可追溯、可回放和可审计。失败数据进入 MySQL DLQ，支持自动重试、人工修复和重新投递。监控上不仅看任务成功率，还要看字段缺失率、映射失败率、新鲜度延迟、DLQ 修复率和下游刷新失败率。任务分片和分布式 Worker 抢占可以作为后续优化，在第一阶段不强行复杂化主链路。

## 18. 面试题目

### 18.1 基础理解

**问题 1：为什么不能把 100 万酒店同步设计成一个单进程长循环？**

参考回答：因为任务运行时间长，中途机器重启、供应商超时、进程发布、网络抖动的概率都很高。单进程长循环的进度通常在内存里，失败后只能从头跑，排查也困难。更合理的设计是 `Task + Batch + Checkpoint + DLQ`：任务先落库，执行过程持续推进 checkpoint，失败数据进入 DLQ，机器重启后从 checkpoint 恢复。

**问题 2：Task 和 Batch 有什么区别？**

参考回答：Task 是任务定义，描述“同步什么、怎么同步、什么时候同步”，比如某供应商酒店全量同步；Batch 是一次具体执行，描述“这一次跑到了哪里、成功多少、失败多少、当前 worker 是谁、租约什么时候过期”。一个 Task 会产生多次 Batch。

**问题 3：Checkpoint 是什么，什么时候更新？**

参考回答：Checkpoint 是任务进度水位，例如当前城市、页码、供应商 cursor、最后处理成功的供应商酒店 ID。它用于断点续跑。推荐先处理本页数据，再推进 checkpoint。这样即使机器在中间宕机，最多重复处理上一页，不会跳过未处理数据。

### 18.2 分布式执行与恢复

**问题 4：worker 如何抢占任务？**

参考回答：通过数据库 CAS 抢占。worker 执行一条带条件的 `UPDATE`，只有 `status=PENDING` 或 `status=RUNNING AND lease_until < NOW()` 的 batch 才能被抢占。`rows_affected=1` 才说明抢占成功，其他 worker 必须退出。

**问题 5：`worker_id` 和 `lease_token` 的区别是什么？**

参考回答：`worker_id` 标识执行器实例，通常由服务名、机器名或容器名、进程号、启动时间组成，方便排查和监控。`lease_token` 标识一次抢占行为，每次抢占都重新生成。关键写操作必须同时校验 `batch_id + worker_id + lease_token`，防止旧 worker 恢复后覆盖新 worker 的进度。

**问题 6：心跳、租约、checkpoint 分别解决什么问题？**

参考回答：心跳说明 worker 是否还活着；租约说明当前任务执行权属于谁；checkpoint 说明任务恢复时从哪里继续。心跳正常不代表任务在前进，所以还要看 `last_checkpoint_at`。租约过期才允许新 worker 抢占，checkpoint 用于恢复位置。

**问题 7：机器重启后怎么恢复？**

参考回答：机器重启后原 worker 不再续租，`lease_until` 到期。新 worker 通过 CAS 抢占过期 batch，读取 `end_checkpoint`，从对应城市、页码或 cursor 继续。由于可能重复处理上一页，所以落库必须基于 `supplier_id + supplier_resource_code + supplier_product_code` 做幂等。

**问题 8：旧 worker 在长 GC 或网络抖动后恢复了怎么办？**

参考回答：旧 worker 恢复后可能以为自己还拥有任务。所有续租、checkpoint、发布和结束任务的 SQL 都必须带 `worker_id + lease_token` 条件。如果更新影响行数为 0，说明租约已经丢失，旧 worker 必须停止执行，不能继续写平台表。

### 18.3 任务互斥与边界场景

**问题 9：上一次任务还没跑完，又下发了一次任务怎么办？**

参考回答：要用显式互斥策略。默认 `SKIP_IF_RUNNING`，如果已有同 `task_code` 的 `PENDING/RUNNING` batch，新任务直接跳过；人工强制重跑可使用 `CANCEL_PREVIOUS`；只有数据范围不重叠时才允许 `ALLOW_PARALLEL`。

**问题 10：心跳正常但 checkpoint 长时间不动，说明什么？**

参考回答：说明 worker 还活着，但任务可能卡在某个阶段，例如供应商慢请求、对象存储写入慢、数据库锁等待、发布阻塞。此时不应立即抢占，而应告警并结合 `last_heartbeat_stage` 定位卡点。只有租约过期才允许新 worker 抢占。

**问题 11：checkpoint 更新失败怎么办？**

参考回答：如果本页处理成功但 checkpoint 更新失败，下次恢复可能重复处理本页。因此页内落库和发布必须幂等。相反，不能先更新 checkpoint 再处理数据，否则宕机会跳过未处理页面。

**问题 12：任务被人工取消时 worker 还在跑，怎么停？**

参考回答：worker 不应该只在启动时读取状态，而要在每页处理前后检查 batch status。如果发现 `CANCELLED`，停止继续拉供应商，不再发布新数据，只做必要的清理和日志记录。

### 18.4 数据治理与补偿

**问题 13：Raw Snapshot 的价值是什么？**

参考回答：Raw Snapshot 是供应商原始响应的证据，不是平台商品模型。它用于排查线上问题、回放同步、验证标准化规则、做 diff，也能区分是供应商数据错误还是平台映射错误。

**问题 14：为什么需要 Diff？**

参考回答：同步成功不等于应该发布。Diff 用来比较标准化后的数据和当前线上发布版本，识别字段变化、图片变化、坐标变化、房型变化和可售变化。低风险变化可以自动发布，高风险变化进入审核或 DLQ。

**问题 15：DLQ 为什么建议用 MySQL，而不是只用消息队列？**

参考回答：供应商同步失败往往不是简单消息消费失败，而是字段缺失、城市映射失败、价格异常、发布失败等需要人工修复、状态流转和审计的问题。MySQL DLQ 可以作为权威问题单，支持查询、分派、重试、忽略、修复和报表。消息队列 DLQ 可以做短期缓冲，但不适合作为运营修复台账。

**问题 16：100 万酒店 10 小时如何估算吞吐？**

参考回答：100 万 / 10 小时约等于 27.8 个酒店/秒。如果每页 100 个酒店，大约需要 10000 页，10 小时内只需要 0.28 页/秒。但如果要逐个拉详情，就是 27.8 QPS，还要受供应商限流、超时、重试和数据处理速度约束。

### 18.5 Redis 抢占问题

**问题 17：worker 可以从 Redis 中抢占任务吗？**

参考回答：可以，但我会把 Redis 作为加速锁或短租约，不把它作为唯一权威状态。任务状态、checkpoint、统计、DLQ 和审计仍然落 MySQL。Redis 可以用 `SET lock_key value NX EX 300` 抢锁，用 Lua 保证续租和释放的原子性。但真正开始执行前仍要更新 MySQL batch 的 `worker_id + lease_token + lease_until`，避免 Redis 主从切换、锁丢失或网络分区导致状态不可追溯。

## 19. 后续优化项目

### 19.1 任务分片

当单批次同步时间继续变长，或者需要多个 Worker 并行提升吞吐时，可以把任务从“Batch + Checkpoint”演进为“Batch + Shard + Checkpoint”。

典型分片方式：

```text
batch_001
  ├─ city_shard_BKK
  ├─ city_shard_JKT
  ├─ city_shard_SIN
  └─ ...
```

Shard 表可以这样设计：

```sql
CREATE TABLE supplier_sync_shard (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    batch_id VARCHAR(64) NOT NULL,
    shard_type VARCHAR(32) NOT NULL COMMENT 'CITY',
    shard_key VARCHAR(128) NOT NULL COMMENT 'city_code or city_id',
    status VARCHAR(32) NOT NULL COMMENT 'PENDING/RUNNING/SUCCESS/FAILED',
    checkpoint VARCHAR(1024) DEFAULT NULL,
    total_count INT DEFAULT 0,
    success_count INT DEFAULT 0,
    failed_count INT DEFAULT 0,
    skipped_count INT DEFAULT 0,
    worker_id VARCHAR(64) DEFAULT NULL,
    lease_token VARCHAR(64) DEFAULT NULL,
    lease_until DATETIME DEFAULT NULL,
    heartbeat_at DATETIME DEFAULT NULL,
    started_at DATETIME DEFAULT NULL,
    finished_at DATETIME DEFAULT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_batch_shard (batch_id, shard_key),
    KEY idx_status (status),
    KEY idx_lease (status, lease_until),
    KEY idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商同步分片';
```

### 19.2 分布式 Worker 抢占

多个 Worker 可以通过数据库 CAS 抢占 `PENDING` shard：

```sql
UPDATE supplier_sync_shard
SET status = 'RUNNING',
    worker_id = 'worker-01',
    lease_token = 'token-abc',
    lease_until = DATE_ADD(NOW(), INTERVAL 5 MINUTE),
    heartbeat_at = NOW(),
    updated_at = NOW()
WHERE id = 123
  AND status = 'PENDING';
```

`rows_affected = 1` 表示抢占成功，`rows_affected = 0` 表示已经被其他 Worker 抢走。

执行过程中 Worker 定期续租：

```sql
UPDATE supplier_sync_shard
SET heartbeat_at = NOW(),
    lease_until = DATE_ADD(NOW(), INTERVAL 5 MINUTE)
WHERE id = ?
  AND worker_id = ?
  AND lease_token = ?
  AND status = 'RUNNING';
```

如果 Worker 宕机，租约过期后，调度器把 shard 释放回 `PENDING`，其他 Worker 读取 shard checkpoint 继续执行。

### 19.3 Redis 抢占与数据库权威状态

当 batch 或 shard 数量非常多，多个 worker 高频抢占数据库导致压力上升时，可以引入 Redis 作为抢占加速层。

基本做法：

```text
worker 抢 Redis 锁
  → SET lock:sync:batch:{batch_id} value NX EX 300
  → 抢到 Redis 锁后，再 CAS 更新 MySQL batch
  → MySQL 更新成功，才真正执行任务
  → 执行期间同时续 Redis 锁和 MySQL lease
```

Redis 抢锁示例：

```text
SET lock:sync:batch:batch_001 worker_id:lease_token NX EX 300
```

续租和释放必须用 Lua 校验 value，不能直接 `DEL`：

```text
if redis.call("GET", key) == value then
    return redis.call("EXPIRE", key, ttl)
else
    return 0
end
```

释放锁同理：

```text
if redis.call("GET", key) == value then
    return redis.call("DEL", key)
else
    return 0
end
```

Redis 抢占的关键原则：

1. Redis 只做短期锁，不做任务事实表。
2. MySQL 仍然是 batch 状态、checkpoint、统计和审计的权威存储。
3. worker 只有同时持有 Redis 锁和 MySQL lease，才允许继续执行。
4. 如果 Redis 锁续租失败，但 MySQL lease 还在，可以选择停止任务并释放 MySQL lease，避免双写风险。
5. 如果 MySQL lease 更新失败，即使 Redis 锁还在，也必须停止任务。

是否使用 Redis，要看瓶颈在哪里。对于“一个 10 小时酒店全量任务”的第一阶段，MySQL CAS 足够简单可靠；对于“上万个 shard、大量 worker 高频抢占”的阶段，Redis 才更有价值。

### 19.4 为什么放在后续优化

任务分片和分布式 Worker 会引入额外复杂度：

1. Shard 状态机。
2. Worker 租约和心跳。
3. 旧 Worker 恢复后的并发写保护。
4. 跨 shard 的批次统计聚合。
5. 热点城市和长尾城市的任务倾斜。

如果第一阶段的 10 小时任务可以接受，优先实现 Batch + Checkpoint + DLQ 的简单闭环。等同步窗口、供应商限流、数据规模或恢复时间成为瓶颈，再引入 shard 和分布式 Worker。
