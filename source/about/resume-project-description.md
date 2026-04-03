# B 端多品类统一商品运营管理平台

**项目角色**：主架构师 / 核心开发  
**项目周期**：2024.01 - 2025.12  
**团队规模**：8 人（2 后端、2 前端、1 测试、1 DBA、1 运维、1 产品）

---

## 项目背景

在数字电商/本地生活平台中，针对**虚拟商品多品类**（话费充值、电影票、酒店、本地生活 Deal、礼品卡、账单代缴等 10+ 品类）导致的**操作流混乱、代码冗余、供给效率低下**的痛点，主导设计并实现了高度可扩展的统一商品供给与运营管理体系。

**核心挑战**：
- **数据源复杂**：供应商 Push/Pull、运营批量导入、商家上传、API 接口等 6 类数据源，审核策略各异
- **品类差异大**：每个品类的商品属性、库存模型（券码制/数量制/时间维度）、定价逻辑（面值/日历价/动态定价）、审核策略完全不同
- **架构不统一，流程碎片化**：每个品类有独立系统，代码逻辑分散，运营需切换多个后台，代码复用率 < 10%
- **缺少监控与追溯**：缺少跟踪、进度、审计等信息，批量操作无法溯源，问题排查困难
- **批量操作瓶颈**：同步批量操作限制在 100 个，万级 SKU 调价需数小时，10 万券码导入 30 分钟，运营效率低

---

## 核心贡献

### 1. 多源异步上架引擎（商品供给侧）

**技术方案**：
- **统一入口 + 多源适配**：针对商家 Portal、运营单品/批量导入、供应商 Push/Pull、API 接口等 6 类数据源，设计了统一的数据接入层，通过 `source_type + source_user_type` 识别来源并动态路由审核策略
- **Kafka + Worker 异步处理**：解耦上传与处理，创建任务后立即返回 `task_code`，后续通过 5 类 Worker（ExcelParse、Audit、Publish、BatchAudit、BatchPublish）异步处理，支持万级并发
- **本地消息表 + Outbox 模式**：引入 `outbox_message_tab` 保证消息可靠发布，配合定时扫描兜底机制，确保高并发下数据最终一致性
- **乐观锁 + 唯一索引**：所有状态变更使用 `version` 字段乐观锁，防止并发冲突，成功率 > 99.9%

**技术细节**：
```go
// 核心流程：创建任务 → Kafka 异步 → Worker 处理
1. API 层：CreateTask() → 写入 listing_task_tab + outbox_message_tab
2. OutboxPublisher：定时扫描 outbox → 发送 Kafka（listing.audit.pending）
3. AuditWorker：消费 Kafka → 规则校验 → 更新状态 → 发送下一阶段事件
4. PublishWorker：Saga 事务发布 → 创建 item/sku/属性表 → 缓存/ES 同步
```

**业务价值**：
- 支持 6 类数据源统一接入，审核策略配置化（供应商快速通道、运营免审、商家人工审核）
- 上架成功率从 85% 提升至 **99%+**，失败任务自动重试 + 看门狗兜底
- 批量上传（10000 SKU）从无法支持（OOM）到 **< 10 分钟**完成

---

### 2. 统一模型与状态机（多品类复用）

**技术方案**：
- **统一有限状态机（FSM）**：所有品类共享 `DRAFT → Pending Audit → Approved/Rejected → Online → Offline` 状态流转，状态变更通过乐观锁保证并发安全
- **策略模式（Strategy Pattern）**：解耦品类差异业务逻辑
  - `ValidationRule` 接口：HotelPriceValidationRule（价格日历校验）、MovieSessionValidationRule（场次时间校验）、TopUpDenominationRule（面额范围校验）
  - `PublishStrategy` 接口：HotelPublishStrategy（创建价格日历表）、MoviePublishStrategy（创建场次座位表）、DealPublishStrategy（关联券码池）
- **规则配置驱动**：审核策略、库存策略、定价策略通过数据库配置表（`listing_audit_config_tab`、`inventory_config`、`fee_config_tab`）动态加载，无需改代码
- **分层存储设计**：
  - 简单品类（话费、礼品卡）：核心表 + JSON 字段扩展
  - 复杂品类（酒店、电影）：核心表 + 专属扩展表（`hotel_price_calendar_tab`、`movie_session_tab`）

**技术细节**：
```go
// 策略模式核心实现
type ValidationRule interface {
    Validate(ctx context.Context, data map[string]interface{}) error
}

type PublishStrategy interface {
    Publish(ctx context.Context, task *ListingTask) error
    Rollback(ctx context.Context, task *ListingTask) error
}

// 运行时动态选择策略
strategy := publishRegistry.Get(task.CategoryID)
saga := NewPublishSaga(strategy.GetSteps(task))
saga.Execute()  // 支持自动回滚
```

**业务价值**：
- 业务代码复用率从 < 10% 提升至 **90%+**
- 新品类接入从 2 周缩短至 **2 天**（只需配置 + 注册策略）
- 支持 **10+ 品类**统一管理，运营学习成本降低 **70%**

---

### 3. 状态可追溯与审计体系（合规 & 问题定位）

**技术方案**：
- **三层审计设计**：
  1. **状态快照表**（`listing_state_history_tab`）：记录每次状态流转（DRAFT→Pending→Approved），包含操作人、时间、原因
  2. **审计日志表**（`listing_audit_log_tab`）：记录审核详情（审核策略、校验规则结果、审核意见）
  3. **业务变更表**（`sku_price_change_log_tab`、`inventory_change_log_tab`）：记录价格/库存的每次变更（旧值→新值、操作类型）
- **完整的变更链路**：通过 `task_id` 关联上架任务、状态历史、审计日志、业务变更，支持从商品 ID 反查全生命周期操作
- **统一批量操作审计**（创新点）：
  - `operation_batch_task_tab`：记录批次级元数据（批次号、操作类型、总数、成功/失败数、结果文件）
  - `operation_batch_item_tab`：记录每条明细（target_id、before_value、after_value、status、error_message）
  - 支持任意维度查询：按批次号、操作人、时间范围、SKU ID 等多维度追溯
  - **审计覆盖率从 33% 提升到 100%**（从仅上架有审计，到所有批量操作全覆盖）

**技术细节**：
```sql
-- 状态历史表
CREATE TABLE listing_state_history_tab (
  task_id         BIGINT NOT NULL,
  from_status     TINYINT NOT NULL,
  to_status       TINYINT NOT NULL,
  action          VARCHAR(50),     -- submit/approve/reject/publish/offline
  operator_id     BIGINT,
  operator_type   VARCHAR(50),     -- system/operator/merchant
  reason          TEXT,
  created_at      TIMESTAMP
);

-- 价格变更日志
CREATE TABLE sku_price_change_log_tab (
  sku_id          BIGINT NOT NULL,
  old_price       DECIMAL(20,2),
  new_price       DECIMAL(20,2),
  change_type     VARCHAR(50),     -- batch/single/promotion/dynamic
  operator_id     BIGINT,
  reason          VARCHAR(500),
  created_at      TIMESTAMP
);
```

**业务价值**：
- 满足合规要求（金融、电商行业审计要求）
- 异常问题定位时间从 **数小时降至 5 分钟**（通过 task_id 快速追溯）
- 支持运营回溯分析（如"为什么这批商品被拒绝？"、"谁批量修改了价格？"）

---

### 4. 统一批量操作框架（架构创新 & 效率提升）

**核心痛点**：
- **架构不统一**：商品上架有完整批次框架（task+item+进度+结果），但价格调整、库存设置却是同步接口，缺少跟踪
- **代码重复**：每种批量操作都要实现一遍（任务创建、进度更新、结果生成、错误处理），复用率 < 15%
- **可追溯性差**：批量调价/设库无审计记录，出问题无法回溯"谁改了什么"
- **新功能成本高**：新增一种批量操作需要 550 行代码、2 周开发时间

**技术方案**：
- **统一数据模型**：设计 `operation_batch_task_tab` 和 `operation_batch_item_tab` 两张泛化表，通过 `operation_type` 字段区分不同批量操作（price_adjust、inventory_update、voucher_code_import、item_edit 等），所有批量操作复用同一套表结构
- **统一事件驱动**：单一 Kafka Topic `operation.batch.created`，Worker 通过 `operation_type` 字段动态路由到不同处理器（PriceUpdateWorker、InventoryUpdateWorker 等），避免重复创建 Topic
- **模板化Worker设计**：抽象 `BaseBatchOperationWorker` 提供通用处理模板（流式解析 → 创建明细 → Worker Pool并发 → 结果文件 → 状态更新），新增批量操作仅需实现 50 行业务逻辑
- **完整审计链路**：每条 `operation_batch_item` 记录 `before_value` 和 `after_value` JSON，支持变更前后对比、完整追溯、合规审计
- **流式解析 + 分批处理**：使用 `excelize.NewReader()` 逐行读取 Excel（内存恒定 < 200MB），分批（100条/批）查询+更新，避免 OOM
- **Worker Pool 并发**：20 goroutine 并发处理，配合 Channel 任务分发，充分利用多核 CPU，并发度提升 **30 倍**
- **统一监控告警**：Prometheus 统一指标 `operation_batch_task_total{operation_type, status}`，一套告警规则覆盖所有批量操作

**技术细节**：
```go
// 统一批量操作架构
1. API层统一入口：
   BatchService.Create() → operation_batch_task_tab + operation_batch_item_tab
   → 发送 operation.batch.created 事件

2. Worker动态路由：
   KafkaConsumer → event.operation_type → 路由到具体Worker
   WorkerManager.Register(PriceUpdateWorker)   // 批量调价
   WorkerManager.Register(InventoryUpdateWorker) // 批量设库存
   WorkerManager.Register(VoucherCodeImportWorker) // 券码导入
   WorkerManager.Register(ItemBatchEditWorker)  // 批量编辑

3. Worker通用处理模板（BaseBatchOperationWorker）：
   流式解析Excel → 分批查询items（100条/批） → Worker Pool(20并发)
   → 记录before/after → 更新进度 → 生成结果文件

4. 新增批量操作接入（仅需3步，50行代码，2天）：
   - 实现 BatchOperationWorker 接口（Process方法）
   - 注册到 WorkerManager
   - API层调用统一 BatchService.Create()
```

**性能对比**：

| 操作 | 优化前 | 优化后 | 提升 |
|------|--------|--------|------|
| **批量上传（10000 SKU）** | 无法完成（OOM） | < 10 分钟 | ∞ |
| **批量调价（1000 SKU）** | 手动逐个（数小时） | < 30 秒 | **100 倍+** |
| **券码导入（10 万条）** | 30 分钟 | < 2 分钟 | **15 倍** |
| **批量设库存（10000 SKU）** | 不支持（同步OOM） | < 5 分钟 | 新增能力 |
| **跨品类批量编辑（500商品）** | 不支持 | < 2 分钟 | 新增能力 |

**架构优化成果**：

| 维度 | 优化前 | 优化后 | 提升 |
|------|--------|--------|------|
| **代码复用率** | 15%（各自实现） | **80%**（统一模板） | **5.3 倍** |
| **新增批量操作成本** | 550 行代码，2 周 | 50 行代码，2 天 | **开发效率提升 7 倍** |
| **审计覆盖率** | 33%（仅上架有） | **100%**（全覆盖） | **3 倍** |
| **资源节约** | 12 核 24GB（3种×2副本） | 11 核 22GB（统一框架） | 节约 1 核 2GB |

**业务价值**：
- 运营人力成本降低 **60%**（批量操作从数小时 → 分钟级）
- 大促准备时间从 **3 天 → 半天**（批量调价、设库存、券码导入一站式完成）
- 错误率从 5%（人工操作）降至 **< 0.1%**（自动化 + 完整审计）
- **架构优势**：新增批量操作从 2 周降至 2 天，开发效率提升 **7 倍**

---

### 5. 运营管理工具集（后台统一化 & 用户体验）

**功能模块**：
- **批量价格管理**：支持 Excel 批量导入、百分比/固定金额调价、促销配置、Fee 配置
- **批量库存管理**：支持批量设库存、券码批量导入（10 万+）、库存对账修复
- **批量操作历史**：统一查询所有批量操作（调价、设库、券码导入、商品编辑），支持按类型、状态、时间筛选
- **配置管理**：首页 Entrance 配置、Tag 标签管理、类目属性维护
- **跨品类搜索**：基于 Elasticsearch，支持全文检索、多维度筛选（品类、状态、价格区间）

**用户体验优化**（基于统一批量框架）：
- **实时进度反馈**：WebSocket 每 2 秒推送进度（0-100%）、成功数、失败数、预估剩余时间
- **结果文件自动生成**：所有批量操作自动生成 Excel 结果文件，包含 before/after 对比、成功/失败明细、错误原因
- **统一操作历史**：一个界面查看所有批量操作历史（调价、设库、券码导入、编辑），支持下载历史结果文件
- **数据对账工具**：自动检测 Redis/MySQL 差异，一键修复

**技术亮点**：
- 统一后台支持所有品类，运营只需一个系统
- 批量操作用户体验一致性（进度、结果、历史）
- 运营人员满意度显著提升（从无反馈 → 实时可视化）

---

### 6. 供应商对接标准化（数据实时性保障）

**技术方案**：
- **Pull 模式**（定时拉取）：Cron 定时任务调用供应商 API，增量同步酒店价格日历、房型库存，支持断点续传
- **Push 模式**（实时推送）：供应商通过 Kafka 实时推送电影场次变更、库存更新，秒级上线
- **混合模式**：酒店采用定时 Pull（30 分钟）+ 实时 Push（紧急变更）
- **同步监控**：监控供应商同步延迟，超过 15 分钟告警，自动重试 + 指数退避

**技术细节**：
```go
// 供应商同步架构
1. Pull 模式（Hotel）：
   Cron(30分钟) → SupplierPullWorker → 调用供应商 API → 批量创建任务
   
2. Push 模式（Movie）：
   供应商 → Kafka(supplier.movie.updates) → SupplierPushConsumer → 创建任务
   
3. 同步监控：
   Cron(5分钟) → SupplierSyncMonitorWorker → 检查同步延迟 → 告警/重试
```

**业务价值**：
- 酒店价格同步延迟从 **30 分钟 → 实时**（Push 模式）
- 电影票上线速度 **< 500ms**（实时推送）
- 供应商对接效率提升 **50%**（标准化接口，新供应商 2 天完成对接）

---

## 技术架构

### 核心技术栈

- **开发语言**：Go（主服务）、Java（部分老系统对接）
- **消息队列**：Kafka（事件驱动、异步解耦、Topic 分区并行消费）
- **缓存 & 锁**：Redis（分布式锁、两级缓存 L1+L2、库存原子操作）
- **数据库**：MySQL（乐观锁、分库分表 16 张、本地消息表 Outbox）
- **搜索引擎**：Elasticsearch（全文检索、实时统计、聚合分析）
- **对象存储**：OSS（Excel 文件存储、结果文件下载）

### 架构模式

- **事件驱动架构**：Kafka + 27 类 Worker（解耦上传/审核/发布/同步/监控/批量操作）
- **策略模式**：ValidationRule、PublishStrategy、InventoryStrategy 接口解耦品类差异
- **模板方法模式**：BaseBatchOperationWorker 提供统一批量处理模板，子类只需实现业务逻辑
- **Saga 分布式事务**：商品发布跨多表操作，支持自动补偿回滚
- **两级缓存架构**：本地缓存（5 分钟）+ Redis 缓存（1 小时），缓存命中率 **85%+**
- **分库分表**：`listing_task_tab` 按 `item_id` 取模分 16 张表，单表 < 5000 万
- **数据归档**：定时任务归档 1 年前数据，保持查询性能
- **统一批量操作框架**：单一事件 Topic + 动态 Worker 路由 + 模板化处理，代码复用率 **80%**

---

## 项目成果

### 核心指标

| 指标 | 优化前 | 优化后 | 提升 |
|------|--------|--------|------|
| **品类支持数** | 3 个（独立系统） | **10+ 个**（统一系统） | - |
| **代码复用率** | < 10%（品类间）<br/>15%（批量操作间） | **90%+**（品类间）<br/>**80%**（批量操作间） | **6-9 倍** |
| **新品类接入时间** | 2 周 | **2 天** | **7 倍** |
| **新批量操作接入** | 550 行，2 周 | **50 行，2 天** | **开发效率提升 7 倍** |
| **批量上传（10000 SKU）** | 无法支持（OOM） | **< 10 分钟** | ∞ |
| **批量调价（1000 SKU）** | 数小时（手动） | **< 30 秒** | **100 倍+** |
| **券码导入（10 万条）** | 30 分钟 | **< 2 分钟** | **15 倍** |
| **批量设库存（10000 SKU）** | 不支持（同步OOM） | **< 5 分钟** | 新增能力 |
| **跨品类批量编辑（500商品）** | 不支持 | **< 2 分钟** | 新增能力 |
| **上架成功率** | 85% | **99%+** | **提升 14%** |
| **批量操作审计覆盖率** | 33%（仅上架） | **100%**（全操作） | **3 倍** |
| **批量处理成功率** | - | **99.99%** | - |
| **供应商同步延迟** | 30 分钟 | **< 30 秒**（秒级） | **60 倍** |
| **运营效率** | - | **人力成本降低 60%** | - |
| **系统故障率** | 5% | **< 0.1%** | **50 倍** |

### 性能指标

| 维度 | 指标 | 说明 |
|------|------|------|
| **并发能力** | 支持 1000 QPS | 批量上传高峰期 |
| **响应时间** | API P99 < 200ms | 单品上传 < 3 秒 |
| **数据规模** | 单表 < 5000 万 | 分表 + 归档策略 |
| **缓存命中率** | 85%+ | L1 本地 + L2 Redis |
| **Kafka 吞吐** | 10000 条/秒 | 批量处理峰值 |
| **Worker 并发** | 20 goroutines | 充分利用多核 |

### 业务价值

- **运营效率提升**：批量操作从数小时降至分钟级，人力成本降低 60%
- **品类扩展能力**：新品类接入从 2 周降至 2 天，接入效率提升 7 倍
- **数据实时性**：供应商同步延迟从 30 分钟降至秒级，用户体验提升显著
- **系统稳定性**：故障率从 5% 降至 < 0.1%，上架成功率 99%+
- **合规性保障**：完整审计链路，满足金融/电商行业监管要求

---

## 技术亮点总结

1. **多源异步上架引擎**：Kafka + Worker + 本地消息表保证最终一致性，支持 6 类数据源统一接入
2. **统一状态机 + 策略模式**：90% 代码复用，新品类 2 天接入
3. **统一批量操作框架**（架构创新）：
   - 单一泛化表 + 单一事件 Topic + 动态 Worker 路由，代码复用率从 15% → **80%**
   - 模板方法模式（BaseBatchOperationWorker），新增批量操作从 550 行/2 周 → **50 行/2 天**
   - 统一审计链路（before/after value），审计覆盖率从 33% → **100%**
   - 流式解析 + Worker Pool，万级数据分钟级完成，内存占用恒定 < 200MB
4. **三层审计体系**：状态快照、审计日志、业务变更，完整追溯链路
5. **供应商对接标准化**：Pull/Push 混合模式，实时性从 30 分钟 → 秒级
6. **高可用架构**：乐观锁 + Saga 事务 + 看门狗机制，成功率 99.99%

---

## 个人职责

- 主导系统架构设计，包括状态机、策略模式、异步处理链路、审计体系、**统一批量操作框架**等核心模块
- **架构创新**：设计并实现统一批量操作框架（单一表模型 + 事件驱动 + 模板方法模式），代码复用率提升 **5 倍**，开发效率提升 **7 倍**
- 负责核心代码开发（70% 代码量），包括 Kafka + Worker 框架、批量处理引擎（流式解析 + Worker Pool）、Saga 事务引擎、BaseBatchOperationWorker 模板
- 制定技术规范（数据库表设计、接口规范、代码规范、部署规范、批量操作接入规范）
- 指导团队成员完成各品类策略开发、供应商对接、运营工具开发、新批量操作接入（3 步接入法）
- 性能优化和故障排查（批量操作 OOM 问题、Kafka 消费延迟、Redis 热 Key 问题、Worker Pool 并发调优）
- 编写技术文档（系统设计文档、接口文档、运维手册、批量操作开发指南）
