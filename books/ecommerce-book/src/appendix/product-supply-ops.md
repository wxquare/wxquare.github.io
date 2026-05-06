# 附录G 商品供给与运营治理平台

## 1. 背景

商品供给与运营治理平台解决的是“商品如何进入平台、如何被审核发布、如何创建和修改库存、上线后如何持续维护”的问题。它不是运营后台的 CRUD，也不是供应商同步的一组定时任务，而是一条长期运行的供给治理流水线。

在数字商品平台中，商品供给来源通常有五类：

```text
人工创建/上传
  → 运营或商家从 0 到 1 创建商品

批量导入
  → 通过模板、Excel、CSV 或文件批量创建和修改商品

运营编辑
  → 对线上商品做标题、图片、类目、价格、库存、上下架、履约和退款规则变更

库存创建 / 修改
  → 初始化库存、补货、导入券码、系统生码、锁库存、门店和日期库存调整

供应商同步
  → 从外部供应商全量、增量、Push 或主动刷新供给数据
```

供应商同步属于商品供给链路，但它不是商品供给链路的全部。更合理的设计是：用一套统一的供给治理控制面承接五类入口，共享任务模型、暂存区、校验、审核、发布版本、Outbox、DLQ、补偿和质量监控；其中供应商同步因为有长任务、Checkpoint、Raw Snapshot、Worker 租约和数据新鲜度问题，单独作为专项链路展开。供应商同步的完整设计见[附录F：供应商数据同步链路](./supplier-sync.md)。

本附录聚焦统一供给治理平台，尤其补足人工上传、批量导入、运营编辑和库存运营四条控制面链路。

## 2. 设计目标

1. **入口统一**：人工创建、批量导入、运营编辑、库存创建 / 修改、供应商同步进入统一任务和发布框架。
2. **线上隔离**：所有未校验、未审核、未发布的数据只进入 Draft / Staging，不污染正式商品表。
3. **质量可控**：通过类目模板、主数据校验、交易契约校验、风险规则和审核流控制发布质量。
4. **发布一致**：商品主数据、资源映射、Offer、库存控制面、履约规则、退款规则、营销协同、搜索索引、缓存和计价上下文最终一致。
5. **失败可恢复**：任务、行级明细、错误文件、DLQ、Outbox 和补偿任务形成闭环。
6. **变更可追溯**：每次发布都有 Diff、审核记录、操作者、TraceID、发布版本和商品快照。
7. **运营可用**：运营能看到任务进度、失败原因、错误文件、审核状态、发布结果和质量报表。

## 3. 核心难点与解决方法

| 难点 | 典型表现 | 风险 | 解决方法 |
|------|----------|------|----------|
| 入口多且语义不同 | 人工创建、批量导入、运营编辑、库存创建 / 修改、供应商同步都在改变供给能力 | 流程混乱、重复逻辑、审计缺失 | 统一为 Supply Task，但按 `task_type` 路由不同策略 |
| 未发布数据污染线上 | 表单保存、导入半成品直接写商品正式表 | 前台展示脏数据，订单拿到半成品契约 | Draft / Staging 与正式表分离，只有发布事务写正式表 |
| 类目差异大 | 酒店、话费、账单、券码、电影票字段完全不同 | 表单和校验 if-else 爆炸 | 类目模板 + 能力矩阵 + Schema 驱动表单和校验 |
| 批量导入规模大 | 大促前一次导入 10 万行商品或价格 | 内存爆、长事务、失败难定位 | 流式解析、行级任务、分批处理、部分成功、错误文件 |
| 运营误操作 | 批量改价、类目迁移、退款规则变更 | 资损、投诉、履约失败 | Diff、风险评分、二次确认、人工审核、灰度发布 |
| 供应商与运营冲突 | 供应商同步覆盖运营修正字段 | 运营修复失效，线上数据反复抖动 | 字段主导权、保护期、版本锁、冲突日志 |
| 审核策略粗糙 | 所有变更都人工审核或全部自动通过 | 效率低或风险失控 | 风险分级：低风险自动，中风险规则校验，高风险强审 |
| 发布不一致 | DB 成功，ES / 缓存 / 计价上下文没刷新，或营销活动协同失败 | 搜不到、价格错、活动不可用、下单失败 | 发布事务 + Outbox + 营销命令 + 异步投影 + 补偿重试 |
| 历史订单受影响 | 商品改价、改退款规则后影响旧订单 | 售后争议、财务对不上 | 创单保存商品快照、报价快照、履约和退款规则快照 |
| 失败不可运营 | 只在日志里记录导入失败 | 运营不知道怎么修 | MySQL DLQ + 错误文件 + 修复建议 + 重新投递 |
| 质量缺陷长期存在 | 缺图、缺价、无库存、无履约规则 | 转化差、履约失败 | 商品质量巡检、质量分、自动下架或告警 |

核心判断已统一收录到[附录B](./interview.md)。

## 4. 总体架构

架构图如下：

![商品供给与运营治理平台总体架构](../../images/product-supply-ops-architecture.png)

图源文件：

- `books/ecommerce-book/images/product-supply-ops-architecture.png`
- `books/ecommerce-book/images/product-supply-ops-architecture.svg`
- `source/diagrams/Excalidraw/product-supply-ops-architecture.excalidraw`

```text
Supply Entry
  → Draft / Staging
  → Supply Task
  → Standardization
  → Quality Validation
  → Diff & Risk Scoring
  → Review / Auto Approval
  → Publish Transaction
  → Outbox Event
  → Search / Cache / Pricing Context / Data Platform
  → Marketing Command / Eligibility Event
  → DLQ / Compensation / Quality Inspection
```

分层职责如下：

| 层级 | 职责 | 关键产物 |
|------|------|----------|
| 供给入口层 | 接收表单、文件、API、供应商同步数据 | 原始输入、来源、操作者、TraceID |
| 暂存层 | 保存未发布数据 | Draft、Staging Snapshot、payload hash |
| 任务层 | 编排一次供给动作 | Task、Task Item、进度、错误文件 |
| 标准化层 | 转成平台统一模型 | Resource、SPU、SKU、Offer、Rule |
| 校验层 | 判断是否完整、合法、可售 | 校验结果、错误码、质量分 |
| 风险审核层 | 判断是否自动通过或人工审核 | Diff、风险等级、审核单 |
| 发布层 | 写正式表、生成版本 | publish version、product snapshot |
| 集成层 | 通过 Outbox 通知搜索、缓存、计价上下文和数据平台，通过营销命令协同活动配置 | Outbox、索引任务、缓存失效任务、营销协同任务 |
| 治理层 | 失败补偿、质量巡检、报表 | DLQ、补偿任务、质量日报 |

### 4.1 核心表分组

商品供给与运营链路的表设计要覆盖草稿、任务、行级处理、暂存、校验、变更审核、发布、补偿审计八类能力。

| 表组 | 典型表 | 作用 |
|------|--------|------|
| Draft 草稿表 | `product_supply_draft`、`product_supply_draft_version` | 保存单商品创建、单商品编辑过程中的草稿，草稿可反复保存，不进入发布 |
| Task 任务表 | `product_supply_task` | 记录一次供给动作：单商品创建、单商品编辑、批量导入、批量编辑、供应商同步后的商品变更 |
| Task Item 明细表 | `product_supply_task_item` | 记录任务中每一行、每个商品、每个 Offer 或每条规则的处理状态 |
| Staging 暂存表 | `product_supply_staging`、`product_supply_staging_snapshot` | 保存已经提交、已经标准化、但还没有发布到正式表的数据 |
| Validation 校验表 | `product_validation_result` | 保存字段、类目、主数据、商品模型、交易契约、风险规则的校验结果 |
| Change / Audit 表 | `product_change_request`、`product_audit_log` | 保存字段 Diff、风险等级、审核策略、审核人、审核结论和驳回原因 |
| Publish / Snapshot 表 | `product_publish_record`、`product_publish_snapshot`、`product_change_log` | 保存发布批次、商品完整快照和正式发布后的变更日志 |
| Outbox / DLQ / Compensation 表 | `product_outbox_event`、`product_supply_dead_letter`、`product_compensation_task`、`product_quality_issue` | 保证下游一致性，承接失败问题单、补偿任务和质量巡检 |

这些表不是为了把商品中心再复制一遍。供给平台负责流程治理和发布编排，正式商品数据仍然写入商品中心主数据表，例如：

```text
resource_tab
product_spu_tab
product_sku_tab
product_offer_tab
rate_plan_tab
stock_config_tab
sellable_rule_tab
fulfillment_rule_tab
refund_rule_tab
```

第一期建议保留最小闭环：

```text
product_supply_draft
product_supply_task
product_supply_task_item
product_supply_staging
product_validation_result
product_change_request
product_audit_log
product_publish_snapshot
product_change_log
product_outbox_event
product_supply_dead_letter
```

供应商同步执行层独立维护 `supplier_sync_task`、`supplier_sync_batch`、`supplier_sync_snapshot`、`supplier_sync_dead_letter`，但标准化后的商品变更要进入供给平台：

```text
supplier_sync_batch
  → Normalize
  → product_supply_task(task_type=SUPPLIER_SYNC_IMPORT)
  → product_supply_task_item
  → product_supply_staging
  → product_validation_result
  → product_change_request
  → Publish
```

## 5. 领域边界

商品供给与运营平台不应该替代商品中心、库存系统、计价系统、搜索系统、订单系统或营销系统。它的职责是“供给流程和发布治理”，不是所有商品数据的唯一存储，也不是到处同步写下游的超级后台。

| 系统 | 负责什么 | 不负责什么 |
|------|----------|------------|
| 供给与运营平台 | 入口、任务、暂存、校验、审核、发布编排、库存创建 / 修改运营入口、营销活动配置入口、补偿、审计 | C 端高 QPS 商品查询、库存扣减、库存账本事实、计价试算、搜索索引直写、订单状态维护、营销优惠计算 |
| 商品中心 | Resource、SPU、SKU、Offer、Rate Plan、类目、属性正式模型 | 运营任务进度和错误文件 |
| 库存系统 | 库存事实、库存创建命令执行、库存扣减、券码池、实时可售、库存账本 | 商品标题、图片、类目、运营审核流 |
| 计价系统 | 价格规则、试算、应付金额、优惠叠加 | 商品生命周期审核 |
| 营销系统 | 活动、券、补贴、预算、营销库存、圈品规则、优惠计算规则 | 商品供给流程、商品生命周期和库存账本 |
| 搜索系统 | 可检索字段、召回、排序、索引刷新 | 商品发布事务 |
| 订单系统 | 商品快照、报价快照、履约契约快照 | 商品最新主数据维护 |

设计原则：

1. 供给平台负责流程，商品中心负责正式模型。
2. 库存创建 / 修改的运营入口在供给平台，库存事实和扣减账本在库存系统。
3. 搜索、缓存、计价上下文和数据平台通过 Outbox 事件感知变更，不由运营后台直接写入。
4. 营销系统通过活动配置命令或营销资格事件协同，但活动规则、预算、券、补贴、营销库存和优惠计算仍归营销系统。
5. 订单只相信创单时保存的快照，不回读最新商品配置解释历史订单，也不由供给平台直接修改订单。

## 6. 任务模型

### 6.1 Task：一次供给动作

`product_supply_task` 记录一次人工创建、批量导入、运营编辑或供应商同步动作。

```sql
CREATE TABLE product_supply_task (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id VARCHAR(64) NOT NULL,
    task_type VARCHAR(32) NOT NULL
        COMMENT 'MANUAL_CREATE/BATCH_IMPORT/OPS_EDIT/SUPPLIER_SYNC',
    execution_mode VARCHAR(16) NOT NULL DEFAULT 'SYNC'
        COMMENT 'SYNC/ASYNC',
    source_type VARCHAR(32) NOT NULL COMMENT 'OPS/MERCHANT/SUPPLIER/SYSTEM',
    source_id VARCHAR(64) DEFAULT NULL,
    category_code VARCHAR(32) NOT NULL,
    operator_id VARCHAR(64) DEFAULT NULL,
    trigger_id VARCHAR(64) DEFAULT NULL COMMENT '外部幂等 ID',
    template_version VARCHAR(64) DEFAULT NULL,
    status VARCHAR(32) NOT NULL
        COMMENT 'DRAFT/PENDING/PARSING/RUNNING/VALIDATING/REVIEWING/APPROVED/PUBLISHING/PUBLISHED/PARTIAL_FAILED/REJECTED/FAILED/CANCELLED',
    total_count INT NOT NULL DEFAULT 0,
    parsed_count INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    failed_count INT NOT NULL DEFAULT 0,
    skipped_count INT NOT NULL DEFAULT 0,
    current_stage VARCHAR(64) DEFAULT NULL,
    input_file_ref VARCHAR(512) DEFAULT NULL,
    parse_checkpoint VARCHAR(1024) DEFAULT NULL,
    error_file_ref VARCHAR(512) DEFAULT NULL,
    publish_version BIGINT DEFAULT NULL,
    worker_id VARCHAR(64) DEFAULT NULL,
    lease_token VARCHAR(64) DEFAULT NULL,
    lease_until DATETIME DEFAULT NULL,
    heartbeat_at DATETIME DEFAULT NULL,
    created_at DATETIME NOT NULL,
    started_at DATETIME DEFAULT NULL,
    finished_at DATETIME DEFAULT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_task_id (task_id),
    UNIQUE KEY uk_task_trigger (task_type, trigger_id),
    KEY idx_status (status),
    KEY idx_category_status (category_code, status),
    KEY idx_operator_time (operator_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品供给任务';
```

### 6.2 Task Item：行级或对象级明细

批量导入和供应商同步必须支持部分成功，因此任务要拆到 item 维度。

```sql
CREATE TABLE product_supply_task_item (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id VARCHAR(64) NOT NULL,
    item_no VARCHAR(64) NOT NULL COMMENT '文件行号、表单对象序号或外部对象序号',
    item_type VARCHAR(32) NOT NULL COMMENT 'RESOURCE/SPU/SKU/OFFER/RATE_PLAN/STOCK/RULE',
    idempotency_key VARCHAR(128) NOT NULL,
    platform_resource_id BIGINT DEFAULT NULL,
    spu_id BIGINT DEFAULT NULL,
    sku_id BIGINT DEFAULT NULL,
    offer_id BIGINT DEFAULT NULL,
    status VARCHAR(32) NOT NULL
        COMMENT 'PENDING/NORMALIZING/VALIDATING/STAGING/DIFFING/REVIEWING/PUBLISHING/SUCCESS/FAILED/DLQ/SKIPPED',
    risk_level VARCHAR(32) DEFAULT NULL COMMENT 'LOW/MEDIUM/HIGH',
    error_code VARCHAR(128) DEFAULT NULL,
    error_message VARCHAR(1024) DEFAULT NULL,
    raw_row_ref VARCHAR(512) DEFAULT NULL,
    staging_id VARCHAR(64) DEFAULT NULL,
    change_id VARCHAR(64) DEFAULT NULL,
    normalized_ref VARCHAR(512) DEFAULT NULL,
    normalized_payload_hash VARCHAR(64) DEFAULT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    max_retry_count INT NOT NULL DEFAULT 5,
    next_retry_at DATETIME DEFAULT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_task_item (task_id, item_no),
    UNIQUE KEY uk_task_idempotency (task_id, idempotency_key),
    KEY idx_task_status (task_id, status),
    KEY idx_platform_object (platform_resource_id, spu_id, sku_id, offer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品供给任务明细';
```

### 6.3 状态机

```text
DRAFT
  → PENDING
  → PARSING
  → RUNNING
  → VALIDATING
  → REVIEWING
  → APPROVED
  → PUBLISHING
  → PUBLISHED

PARSING / RUNNING / VALIDATING / REVIEWING / PUBLISHING
  → PARTIAL_FAILED / FAILED / REJECTED

PENDING / PARSING / RUNNING / VALIDATING / REVIEWING
  → CANCELLED
```

状态说明：

| 状态 | 含义 |
|------|------|
| `DRAFT` | 表单草稿或导入任务草稿 |
| `PENDING` | 已提交，等待执行 |
| `PARSING` | 批量任务正在解析文件并生成 item |
| `RUNNING` | 批量任务正在分批处理 item |
| `VALIDATING` | 正在标准化和质量校验 |
| `REVIEWING` | 有高风险项进入审核 |
| `APPROVED` | 审核通过，等待发布 |
| `PUBLISHING` | 正在写正式表和 Outbox |
| `PUBLISHED` | 全部发布成功 |
| `PARTIAL_FAILED` | 部分 item 成功、部分失败 |
| `REJECTED` | 审核驳回 |
| `FAILED` | 整体失败 |
| `CANCELLED` | 人工取消 |

## 7. 暂存区与快照

所有入口都必须先写暂存区，不能直接写商品正式表。

```sql
CREATE TABLE product_supply_staging (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    staging_id VARCHAR(64) NOT NULL,
    task_id VARCHAR(64) NOT NULL,
    item_no VARCHAR(64) NOT NULL,
    object_type VARCHAR(32) NOT NULL
        COMMENT 'RESOURCE/SPU/SKU/OFFER/RATE_PLAN/STOCK_CONFIG/INPUT_SCHEMA/FULFILLMENT_RULE/REFUND_RULE',
    object_key VARCHAR(128) NOT NULL,
    source_type VARCHAR(32) NOT NULL,
    source_ref VARCHAR(512) DEFAULT NULL,
    raw_payload_ref VARCHAR(512) DEFAULT NULL,
    normalized_payload JSON NOT NULL,
    payload_hash VARCHAR(64) NOT NULL,
    base_publish_version BIGINT DEFAULT NULL,
    status VARCHAR(32) NOT NULL
        COMMENT 'DRAFT/VALIDATED/REVIEWING/APPROVED/PUBLISHED/REJECTED',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_staging_id (staging_id),
    UNIQUE KEY uk_task_object (task_id, object_type, object_key),
    KEY idx_status (status),
    KEY idx_object_key (object_type, object_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品供给暂存数据';
```

暂存区的作用：

1. 保护线上正式表，不让半成品商品被搜索或下单。
2. 支持审核员查看发布前快照。
3. 支持 Diff、风险评分、回放和问题排查。
4. 支持失败后修复并重新发布。

## 8. 人工创建链路

人工创建适合运营或商家少量创建商品，例如本地生活券、礼品卡、账单缴费入口、活动套餐。

```text
选择类目
  → 加载类目模板
  → 填写 Resource / SPU / SKU / Offer / Rule
  → 前端实时校验
  → 保存 Draft
  → 提交 Supply Task
  → 后端强校验
  → 生成 Staging Snapshot
  → 审核
  → 发布
```

关键难点：

| 难点 | 解决方法 |
|------|----------|
| 不同品类字段差异巨大 | 类目模板驱动表单，模板定义字段、类型、是否必填、校验规则 |
| 运营只填商品标题和价格，遗漏交易契约 | 提交时强校验 Offer、库存来源、履约规则、退款规则、Input Schema |
| 新商品审核缺少上下文 | 审核页展示标准化快照、类目模板、风险命中、历史相似商品 |
| 草稿反复修改 | Draft 与 Staging 分离，草稿不生成发布版本 |
| 创建成功但无法下单 | 发布后做可售校验：库存、价格、履约、退款、搜索索引状态 |

类目模板示例：

```json
{
  "category_code": "HOTEL",
  "required_objects": ["RESOURCE", "SPU", "OFFER", "RATE_PLAN", "REFUND_RULE"],
  "fields": [
    {"name": "hotel_name", "type": "string", "required": true},
    {"name": "city_code", "type": "string", "required": true},
    {"name": "geo.lat", "type": "decimal", "required": true},
    {"name": "geo.lng", "type": "decimal", "required": true}
  ]
}
```

## 9. 批量导入链路

批量导入适合大促、类目迁移、商家批量上新、套餐批量配置。

```text
下载模板
  → 上传文件
  → 文件格式预检
  → 创建 product_supply_task
  → 流式解析
  → 每行生成 product_supply_task_item
  → 分批标准化
  → 行级校验
  → 成功项发布或审核
  → 失败项生成错误文件
  → 汇总任务状态
```

核心难点与解决方法：

| 难点 | 解决方法 |
|------|----------|
| 文件过大 | 流式解析，不一次性读入内存 |
| 导入耗时长 | 分批提交，后台异步执行，前台轮询进度 |
| 局部失败 | 行级状态，成功项继续，失败项生成错误文件 |
| 重复上传 | `task_type + trigger_id` 幂等，行级 `idempotency_key` 去重 |
| 模板演进 | 文件记录 `template_version`，旧模板兼容或拒绝 |
| 批量事故 | 高风险字段批量变更进入抽样审核或二次确认 |
| 下游被打爆 | 发布和索引刷新限速，使用 Outbox 背压 |

### 9.1 异步执行总流程

批量导入和批量编辑不能只有一个后台线程从头跑到尾。更稳妥的方式是拆成解析、行级处理、审核发布和结果归档几个阶段。

```text
上传文件 / 批量提交
  → 创建 product_supply_task(status=PENDING, execution_mode=ASYNC)
  → Parser Worker 流式解析文件
  → 批量写入 product_supply_task_item
  → Item Worker 分批处理 item
  → 标准化 / 校验 / Staging / Diff
  → 低风险自动发布，高风险进入审核
  → Publish Worker 发布正式表并写 Outbox
  → 生成错误文件 / DLQ / 质量报告
```

这个拆法有三个好处：

1. 解析失败不会污染正式商品表。
2. 行级失败不会拖垮整批任务。
3. 发布和下游刷新可以限速、重试和补偿。

### 9.2 Parser Worker：只解析，不发布

Parser Worker 的职责边界要非常窄：只负责把文件拆成 `product_supply_task_item`，不做正式发布。

```text
1. CAS 抢占 product_supply_task
2. 校验 input_file_ref、文件 hash、模板版本和列结构
3. 流式读取文件，不能一次性加载到内存
4. 每 N 行批量插入 product_supply_task_item
5. 更新 parsed_count、parse_checkpoint、heartbeat_at
6. 解析完成后写 total_count
7. task.status 从 PARSING 推进到 RUNNING
```

`parse_checkpoint` 用来恢复解析进度：

```json
{
  "sheet": "Sheet1",
  "row_no": 12000,
  "byte_offset": 8842211
}
```

如果 Parser Worker 在第 12000 行宕机，下次恢复时允许重复解析上一小批。重复数据由两个唯一键兜住：

```text
UNIQUE(task_id, item_no)
UNIQUE(task_id, idempotency_key)
```

注意：Excel 这类格式不一定天然支持稳定的 `byte_offset` 恢复。工程上可以先把上传文件转换成规范化 CSV 或行级 JSONL，再按 offset 恢复；也可以按 `row_no` 从头快速跳过。核心原则是 checkpoint 控制重跑范围，幂等保证重复处理不写错。

### 9.3 Task Item：行级事实表

`product_supply_task_item` 是批量链路里最重要的表。Task 只说明“这次批量任务怎么样”，Item 才能回答“第几行、哪个商品、哪个 Offer 为什么失败”。

Item 状态机建议设计为：

```text
PENDING
  → NORMALIZING
  → VALIDATING
  → STAGING
  → DIFFING
  → REVIEWING
  → PUBLISHING
  → SUCCESS

失败分支：
NORMALIZING / VALIDATING / STAGING / DIFFING / PUBLISHING
  → FAILED / DLQ / SKIPPED
```

关键字段含义：

| 字段 | 作用 |
|------|------|
| `item_no` | 文件行号或批量对象序号 |
| `idempotency_key` | 行级业务幂等键，防止重复导入 |
| `raw_row_ref` | 原始行数据引用，方便生成错误文件和回放 |
| `normalized_ref` | 标准化后 payload 引用 |
| `staging_id` | 通过校验后的暂存数据 |
| `change_id` | Diff 后生成的变更单 |
| `retry_count` / `next_retry_at` | 自动重试控制 |

### 9.4 Item Worker：分批处理行级任务

Item Worker 不按“整个文件”处理，而是扫描一小批待处理 item。

```sql
SELECT *
FROM product_supply_task_item
WHERE task_id = ?
  AND status IN ('PENDING', 'FAILED')
  AND next_retry_at <= NOW()
ORDER BY item_no ASC
LIMIT 500;
```

每个 item 或小批次使用独立事务：

```text
读取 raw_row_ref
  → CAS 将 item 推进到 NORMALIZING
  → 按类目模板标准化成 Resource / SPU / SKU / Offer / Rule
  → 写 normalized_ref
  → 执行 Schema / 主数据 / 商品模型 / 交易契约校验
  → 校验通过后写 product_supply_staging
  → 与线上 publish_version 做 Diff
  → 生成 product_change_request
  → 根据 risk_level 自动发布或进入 REVIEWING
  → 更新 item.status
```

不要用一个大事务包住 500 行。正确做法是行级或小批次事务，否则一行失败会拖垮整批，也会造成长事务、锁等待和回滚成本过高。

### 9.5 Staging、Diff 与发布合流

Item Worker 校验通过后，只能写 `product_supply_staging`，不能直接写正式商品表。

```text
product_supply_task_item
  → product_supply_staging
  → product_change_request
  → product_publish_snapshot
  → product_outbox_event
```

`base_publish_version` 很重要。批量导入或批量编辑可能基于旧版本生成，如果发布时线上商品已经被别人改过，必须识别版本冲突，不能静默覆盖。

风险分流建议如下：

| 风险等级 | 处理 |
|----------|------|
| `LOW` | 自动准入，进入发布 |
| `MEDIUM` | 规则校验通过后发布，异常进入审核 |
| `HIGH` | 强制进入人工审核 |

Publish Worker 只处理已经 `APPROVED` 或 `AUTO_APPROVE` 的变更：

```text
读取 approved change
  → 开启发布事务
  → 写 Resource / SPU / SKU / Offer / Rule
  → 写 publish_snapshot
  → 写 product_change_log
  → 写 product_outbox_event
  → 提交事务
  → item.status = SUCCESS
```

ES、缓存和计价上下文不要放在发布事务里同步调用，统一由 Outbox 消费者异步刷新；营销活动配置走营销系统命令或营销资格事件，不在发布事务内同步写营销规则。

### 9.6 Task 状态汇总

Task 状态不要靠 Worker 主观判断，而要从 item 状态聚合。

| Item 汇总结果 | Task 状态 |
|---------------|-----------|
| 全部 `SUCCESS` | `PUBLISHED` |
| 部分 `SUCCESS`，部分 `FAILED/DLQ` | `PARTIAL_FAILED` |
| 全部失败 | `FAILED` |
| 存在 `REVIEWING` | `REVIEWING` |
| 存在 `PUBLISHING` | `PUBLISHING` |
| 任务被人工取消 | `CANCELLED` |

统计可以每批 item 处理完成后增量更新，也可以由定时聚合 Job 修正。运营后台看到的进度来自 task 计数，但失败定位必须下钻到 item。

### 9.7 失败处理

批量异步链路的失败要按阶段处理：

| 失败阶段 | 示例 | 处理 |
|----------|------|------|
| 文件级失败 | 文件损坏、模板版本不支持、列结构缺失 | task 直接 `FAILED`，不生成大量 item |
| 行级格式失败 | 价格非法、字段缺失、枚举非法 | item `FAILED`，写错误文件 |
| 主数据失败 | 城市、商户、品牌不存在 | item `DLQ` 或 `MANUAL_FIX` |
| 风险失败 | 改价过大、退款规则变化 | change_request `REVIEWING` |
| 发布失败 | 版本冲突、DB 冲突、唯一键冲突 | item 延迟重试，超过次数进 DLQ |
| 下游失败 | ES、缓存刷新失败 | Outbox 补偿，不回滚发布事务 |

错误文件应该从 `product_supply_task_item` 和 `product_validation_result` 生成，而不是从日志拼出来。

### 9.8 设计原则

1. **Parser Worker 只解析，不发布。**
2. **Item Worker 按行级状态推进，支持部分成功。**
3. **所有 item 处理必须幂等。**
4. **Staging 是正式表前的隔离层。**
5. **发布必须版本化，不能覆盖未知的新版本。**
6. **下游刷新走 Outbox，不阻塞发布事务。**
7. **Task 管整体进度，Item 才是真正的问题定位单元。**

错误文件要能指导运营修复，而不是只写“导入失败”：

```text
row_no, object_key, field, error_code, error_message, suggestion
12, SKU_001, price, PRICE_TOO_LOW, price lower than floor price, adjust price >= 100
25, OFFER_014, refund_rule, REFUND_RULE_MISSING, refund rule is required, choose a refund template
31, HOTEL_020, city_code, CITY_NOT_FOUND, city cannot map to platform city, add city mapping first
```

## 10. 运营编辑链路

运营编辑针对线上商品，需要解决“谁能改、改什么、是否覆盖供应商数据、什么时候生效、如何回滚”的问题。

```text
读取当前 publish_version
  → 创建编辑草稿
  → 修改字段
  → 生成 Diff
  → 字段主导权判断
  → 风险评分
  → 自动通过 / 人工审核 / 阻断
  → 发布新 publish_version
  → Outbox 通知读侧投影
  → 营销活动配置异步协同
```

### 10.1 字段主导权

| 字段 | 主导方 | 供应商同步能否覆盖 | 运营策略 |
|------|--------|-------------------|----------|
| 标题、卖点、活动标签 | 平台运营 | 否 | 运营编辑为准 |
| 酒店名称、地址、设施 | 供应商/平台治理 | 低风险可覆盖，高风险审核 | 可人工修正并设置保护期 |
| 展示图片 | 平台运营/供应商 | 取决于来源质量 | 图片变更需要质量校验 |
| 基础价、Rate Plan | 供应商/计价 | 取决于品类 | 超阈值审核 |
| 库存水位、可售状态 | 库存域/供应商 | 是 | 人工覆盖必须有有效期 |
| 退款规则、履约规则 | 平台/供应商契约 | 高风险覆盖 | 强制审核 |
| 类目、Resource 映射 | 平台治理 | 否 | 强制审核和数据巡检 |

### 10.2 冲突处理

常见冲突：

```text
运营改了酒店名称
  → 供应商增量同步又推回旧名称

运营批量下架一批商品
  → 供应商同步推送可售状态为可售

运营修复城市映射
  → 供应商全量同步发现城市字段不同
```

解决方法：

1. 对每个字段定义 `owner_type`：OPS、SUPPLIER、SYSTEM。
2. 运营覆盖供应商字段时记录 `override_until` 和 `override_reason`。
3. 供应商同步遇到运营保护字段时只记录 Diff，不自动覆盖。
4. 高风险冲突进入审核队列。
5. 保护期到期后由巡检任务决定是否恢复供应商主导。

## 11. 标准化与质量校验

质量校验要分层，不要只做字段必填。

| 校验层 | 校验内容 | 失败处理 |
|--------|----------|----------|
| Schema 校验 | 类型、必填、枚举、长度、格式 | 行级失败 |
| 类目模板校验 | 类目要求的对象和字段是否完整 | 阻断提交 |
| 主数据校验 | 城市、商户、品牌、Resource 是否存在 | 进入人工映射 |
| 商品模型校验 | SPU、SKU、Offer、Rate Plan 关系是否成立 | 阻断发布 |
| 交易契约校验 | 库存来源、Input Schema、履约规则、退款规则 | 阻断发布 |
| 可售校验 | 商品状态、库存、价格、渠道、站点是否允许售卖 | 阻断上线或告警 |
| 风险校验 | 价格、类目、履约、退款、映射是否高风险 | 进入审核 |

质量分可以作为运营看板：

```text
quality_score =
  content_score
  + model_score
  + sellability_score
  + fulfillment_score
  + risk_score
```

如果商品缺图、缺价、无库存、无履约规则，即使主表写入成功，也不能认为供给成功。

## 12. Diff 与风险审核

审核不是所有变更都走人工。系统应该根据 Diff 和风险规则决定处理方式。

```sql
CREATE TABLE product_change_request (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    change_id VARCHAR(64) NOT NULL,
    task_id VARCHAR(64) NOT NULL,
    object_type VARCHAR(32) NOT NULL,
    object_id BIGINT DEFAULT NULL,
    old_publish_version BIGINT DEFAULT NULL,
    new_staging_id VARCHAR(64) NOT NULL,
    changed_fields JSON NOT NULL,
    risk_level VARCHAR(32) NOT NULL COMMENT 'LOW/MEDIUM/HIGH',
    review_policy VARCHAR(32) NOT NULL COMMENT 'AUTO_APPROVE/MANUAL_REVIEW/BLOCK',
    status VARCHAR(32) NOT NULL COMMENT 'PENDING/APPROVED/REJECTED/PUBLISHED',
    reviewer_id VARCHAR(64) DEFAULT NULL,
    review_note VARCHAR(1024) DEFAULT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_change_id (change_id),
    KEY idx_task (task_id),
    KEY idx_status_risk (status, risk_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品供给变更单';
```

风险策略：

| 变更类型 | 风险等级 | 策略 |
|----------|----------|------|
| 标题、描述、小图修正 | 低 | 自动通过，记录日志 |
| 普通图片变更 | 低/中 | 图片质量校验通过后发布 |
| 库存水位调整 | 中 | 自动校验，通过后发布，异常告警 |
| 价格或 Offer 规则变更 | 中高 | 超阈值人工审核 |
| 类目变更 | 高 | 强制审核 |
| 履约类型或退款规则变更 | 高 | 强制审核 |
| Resource / Supplier Mapping 变更 | 高 | 强制审核并触发巡检 |

风险评分示例：

```text
risk_score =
  field_weight
  + change_ratio_weight
  + category_weight
  + product_heat_weight
  + operator_history_weight
  + source_trust_weight
```

## 13. 发布一致性设计

审核通过不等于商品可售。发布阶段要把商品主数据和交易前契约一次性落到可追溯版本上。

```text
开始发布事务
  → 校验 base_publish_version
  → 写 Resource / SPU / SKU / Offer / Rate Plan
  → 写 Stock Config / Sellable Rule
  → 写 Input Schema / Fulfillment Rule / Refund Rule
  → 写 Supplier Mapping 或 Merchant Mapping
  → 生成 publish_version
  → 生成 product_snapshot
  → 写 product_change_log
  → 写 outbox_event
提交事务
  → 异步刷新搜索、缓存、计价上下文、数据平台
  → 如涉及活动配置，异步调用营销系统命令
```

关键设计：

| 设计点 | 解决的问题 |
|--------|------------|
| `base_publish_version` 乐观锁 | 防止基于旧版本覆盖新版本 |
| `publish_version` | 支持回滚、审计、对账 |
| `product_snapshot` | 支持订单快照、问题排查 |
| `outbox_event` | 防止商品已变更但下游没收到事件 |
| 异步刷新 | 避免发布事务被 ES、缓存、计价上下文和营销协同拖慢 |
| 补偿任务 | 下游刷新失败后可重试 |

Outbox 事件：

```text
ProductPublished
ProductContentChanged
OfferChanged
RatePlanChanged
SellableRuleChanged
FulfillmentRuleChanged
RefundRuleChanged
SearchIndexRefreshRequired
ProductCacheInvalidationRequired
```

## 14. DLQ 与补偿

人工供给和运营编辑也需要 DLQ。它们的失败通常不是供应商接口失败，而是输入错误、映射错误、审核驳回、版本冲突、发布失败和下游刷新失败。

```sql
CREATE TABLE product_supply_dead_letter (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    dead_letter_id VARCHAR(64) NOT NULL,
    task_id VARCHAR(64) NOT NULL,
    task_type VARCHAR(32) NOT NULL,
    item_no VARCHAR(64) DEFAULT NULL,
    object_type VARCHAR(32) DEFAULT NULL,
    object_key VARCHAR(128) DEFAULT NULL,
    platform_resource_id BIGINT DEFAULT NULL,
    spu_id BIGINT DEFAULT NULL,
    sku_id BIGINT DEFAULT NULL,
    offer_id BIGINT DEFAULT NULL,
    error_stage VARCHAR(64) NOT NULL COMMENT 'PARSE/VALIDATION/MAPPING/REVIEW/PUBLISH/OUTBOX/INDEX/CACHE',
    error_type VARCHAR(64) NOT NULL COMMENT 'RETRYABLE/NON_RETRYABLE/MAPPING_REQUIRED/RISK_BLOCKED/VERSION_CONFLICT',
    error_code VARCHAR(128) NOT NULL,
    error_message VARCHAR(1024) NOT NULL,
    payload_ref VARCHAR(512) DEFAULT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING'
        COMMENT 'PENDING/RETRYING/MANUAL_FIX/RESOLVED/IGNORED/FAILED',
    retry_count INT NOT NULL DEFAULT 0,
    max_retry_count INT NOT NULL DEFAULT 5,
    next_retry_at DATETIME DEFAULT NULL,
    owner_team VARCHAR(64) DEFAULT NULL,
    assignee VARCHAR(64) DEFAULT NULL,
    fix_note VARCHAR(1024) DEFAULT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    resolved_at DATETIME DEFAULT NULL,
    UNIQUE KEY uk_dead_letter_id (dead_letter_id),
    KEY idx_status_next_retry (status, next_retry_at),
    KEY idx_task (task_id),
    KEY idx_error_code (error_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品供给死信队列';
```

补偿策略：

| 失败类型 | 示例 | 处理方式 |
|----------|------|----------|
| 可重试失败 | DB 短暂失败、Outbox 发送失败 | 指数退避重试 |
| 输入失败 | 文件字段非法、必填缺失 | 生成错误文件，运营修复后重新提交 |
| 映射失败 | 城市、商户、品牌找不到 | 人工补映射后重新投递 |
| 风险阻断 | 价格异常、退款规则风险 | 人工审核或驳回 |
| 版本冲突 | 基于旧版本编辑 | 重新拉取最新版本再编辑 |
| 下游失败 | ES、缓存刷新失败 | Outbox 补偿重试 |

## 15. 可观测性

运营后台和监控系统要能回答五个问题：

1. 任务跑到哪里了？
2. 为什么失败？
3. 谁需要处理？
4. 修复后如何重新投递？
5. 发布后前台是否真的可见、可买、可履约？

核心指标：

| 指标类型 | 指标 |
|----------|------|
| 任务进度 | 总数、成功数、失败数、跳过数、当前阶段 |
| 效率指标 | 任务完成耗时、P95 / P99、排队时间 |
| 质量指标 | 字段缺失率、映射失败率、缺图率、缺价率、无库存率 |
| 审核指标 | 自动审核占比、人工审核耗时、驳回率 |
| 发布指标 | 发布成功率、版本冲突数、回滚次数 |
| 下游指标 | ES 刷新失败数、缓存失效失败数、Outbox 堆积 |
| 运营指标 | 错误文件下载次数、人工修复耗时、DLQ 修复率 |

质量巡检任务：

1. 缺图商品巡检。
2. 缺价商品巡检。
3. 无库存商品巡检。
4. 无履约规则商品巡检。
5. 退款规则缺失巡检。
6. 搜索索引与发布版本一致性巡检。
7. 缓存版本与发布版本一致性巡检。
8. 运营覆盖字段到期巡检。

## 16. 典型异常场景

| 异常 | 风险 | 处理 |
|------|------|------|
| 运营重复点击提交 | 重复创建任务 | `task_type + trigger_id` 幂等 |
| 导入文件 10 万行中 500 行失败 | 整批回滚影响效率 | 部分成功，失败行生成错误文件 |
| 发布时发现版本冲突 | 覆盖别人刚发布的变更 | `base_publish_version` 乐观锁，要求重新编辑 |
| 审核通过但 ES 刷新失败 | 前台搜不到 | Outbox 补偿刷新 |
| 商品发布成功但库存未初始化 | 前台可见不可买 | 可售校验不通过，不进入 ONLINE |
| 运营改标题后被供应商覆盖 | 人工修复失效 | 字段主导权和保护期 |
| 大批量改价低于底价 | 资损 | 风险规则阻断，人工审核 |
| 退款规则变更影响历史订单 | 售后争议 | 订单保存退款规则快照 |
| 质量巡检发现缺履约规则 | 下单后无法履约 | 自动下架或阻断可售，进入 DLQ |

## 17. 与供应商同步的关系

统一供给治理平台和供应商同步专项链路的关系如下：

```text
统一供给治理平台
  → 统一任务模型
  → 统一暂存与发布模型
  → 统一校验、Diff、审核、Outbox、补偿

供应商同步专项链路
  → Raw Snapshot
  → Sync Batch
  → Checkpoint
  → Worker Lease
  → Supplier Mapping
  → 数据新鲜度
```

供应商同步产生的标准化数据最终也应该进入供给治理平台的校验、Diff、审核和发布机制。不同点在于，供应商同步多了拉取、分页、断点续跑、租约抢占、原始快照和供应商质量监控。

## 18. 答辩材料

本专题相关总结、常见问题和参考回答已统一收录到[附录B](./interview.md)。
