---
title: 多品类统一商品·库存·价格管理系统设计：电商·虚拟商品·本地生活
date: 2026-02-26
categories:
- 系统设计
tags:
- 商品管理
- 库存系统
- 价格系统
- 电商
- 系统设计
- 高并发
toc: true
---

<!-- toc -->

## 一、背景与挑战

### 1.1 业务背景

在数字电商/本地生活平台中，**商品管理、库存管理、价格管理**是三大核心支柱系统，它们相互依赖、紧密协作，共同支撑着平台的商品从上架到售卖的完整生命周期。

```
商品上架 → 库存同步 → 价格配置 → 用户浏览列表 → 查看详情 → 加入购物车 → 用户下单 → 库存扣减 → 订单履约
   ↓         ↓         ↓            ↓            ↓         ↓          ↓         ↓         ↓
商品信息   库存状态   基础定价   批量价格计算   实时价格   价格快照   订单创建   预订/售出   发货/核销
                                (缓存优化)     (促销匹配)  (30分钟)
```

### 1.2 多品类差异与挑战

不同品类在商品属性、库存特性、定价逻辑上差异极大：

| 品类 | 商品特点 | 库存特性 | 价格特点 | 典型示例 |
|------|----------|----------|----------|----------|
| **电子券 (Deal)** | 券码制，每券唯一 | 券码池，预订扣减 | 面值 vs 售价 | 咖啡店电子券 |
| **虚拟服务券 (OPV)** | 数量制，分平台统计 | 数量制，预订扣减 | 固定价 + 促销 | 美甲服务券 |
| **酒店 (Hotel)** | 房型 × 日期 | 时间维度库存 | 日历价 + 动态定价 | 在线酒店预订 |
| **电影票 (Movie)** | 场次 × 座位 × 票种 | 座位制库存 | 场次定价 + Fee | IMAX 电影票 |
| **机票/票务** | 航班 × 舱位 | 座位/场次制 | 动态定价 | 航班经济舱 |
| **礼品卡 (Giftcard)** | 实时生成或预采购 | 券码制 / 无限 | 面值定价 | 应用商店充值卡 |
| **话费充值 (TopUp)** | 面额制 | 无限库存 | 面额 + 折扣 | 手机话费充值 |
| **本地生活套餐** | 组合型，多子项 | 组合库存联动 | 套餐价 + 子项加总 | 火锅双人套餐 |

### 1.3 核心痛点

#### 1.3.1 商品管理痛点

1. **流程不统一**：每个品类上架流程各异，代码无法复用
2. **状态管理混乱**：草稿、审核、上线、下线等状态散落在不同表中
3. **供应商对接不统一**：推送/拉取/API 各自实现，缺乏标准化
4. **审核策略不灵活**：无法根据数据来源（供应商/运营/商家）动态调整审核策略

#### 1.3.2 库存管理痛点

1. **模型割裂**：每个品类独立设计库存逻辑，无法复用
2. **数据不一致**：Redis 与 MySQL 之间、预订数量与实际状态脱节
3. **供应商策略不统一**：实时查询、定时同步、推送等策略混乱
4. **缺乏统一服务**：业务方直接操作 DB/Redis，维护成本高
5. **监控缺失**：超卖、库存差异、供应商同步延迟难以发现

#### 1.3.3 价格管理痛点

1. **价格散落多表**：基础价、营销价、费用、优惠券分散在不同模块
2. **计算逻辑分散**：各品类各自实现价格计算，重复代码多
3. **营销活动隔离**：促销规则硬编码在业务逻辑中，扩展性差
4. **Fee 管理混乱**：平台手续费、商户服务费、合作方费用等缺乏统一配置
5. **优惠券叠加复杂**：多种优惠方式叠加规则不清晰
6. **审计困难**：价格变更历史难以追溯，无法准确还原计算过程

### 1.4 设计目标

| 目标 | 说明 | 优先级 |
|------|------|--------|
| **统一模型** | 商品、库存、价格共用一套统一模型，多品类复用 | P0 |
| **高性能** | 支持万级 QPS 秒杀场景，P99 < 100ms | P0 |
| **灵活扩展** | 新品类接入无需修改核心代码 | P0 |
| **最终一致** | Redis 与 MySQL 数据最终一致 | P0 |
| **异步化** | 上传、审核、发布、价格快照异步化 | P0 |
| **状态可追溯** | 完整的状态变更历史记录 | P0 |
| **供应商集成** | 支持实时/定时/推送多种同步策略 | P1 |
| **多级降级** | 促销/优惠券服务不可用时，仍能返回基础价格 | P1 |

---

## 二、整体架构

### 2.1 三大系统总览

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         统一商品·库存·价格管理平台                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                      商品上架管理系统 (Listing)                          │ │
│  │  数据来源: 运营/商家/供应商Push/供应商Pull/API                            │ │
│  │  核心流程: DRAFT → Pending Audit → Approved → Online                    │ │
│  │  策略: 审核策略路由（免审/自动审核/人工审核/快速通道）                     │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                      ↓                                        │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                      统一库存管理系统 (Inventory)                         │ │
│  │  管理类型: 自管理/供应商管理/无限库存                                      │ │
│  │  单元类型: 券码制/数量制/时间维度/组合型                                   │ │
│  │  核心操作: BookStock / UnbookStock / SellStock / RefundStock            │ │
│  │  存储: Redis(热) + MySQL(冷) + Kafka(事件)                               │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                      ↓                                        │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                      统一价格管理系统 (Pricing)                           │ │
│  │  四层架构: Base Price → Promotion → Fee → Voucher                       │ │
│  │  核心能力: 价格计算引擎 / 营销匹配器 / 费用计算器 / 优惠券应用器          │ │
│  │  降级策略: 5级降级（促销/费用/优惠券可降级）                               │ │
│  │  审计: 价格快照 + 变更日志 + 人类可读公式                                 │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                      ↓                                        │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                        订单服务 (Order Service)                          │ │
│  │  下单 → 价格锁定(快照) → 库存预订 → 支付 → 库存售出 → 履约                │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                                                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 分层服务架构

```
┌───────────────────────────────────────────────────────────────────────────┐
│                            API Gateway / BFF                               │
└───────────────────────────────────────────────────────────────────────────┘
                                      │
        ┌─────────────────────────────┼─────────────────────────────┐
        │                             │                             │
        ▼                             ▼                             ▼
┌───────────────────┐     ┌───────────────────┐       ┌───────────────────┐
│ Listing Service   │     │ Inventory Service │       │ Pricing Service   │
│ ─────────────     │     │ ───────────────   │       │ ─────────────     │
│ • 上架API         │     │ • 库存查询        │       │ • 价格计算API      │
│ • 审核API         │     │ • 库存预订        │       │ • 快照API         │
│ • 发布API         │     │ • 库存售出        │       │ • 审计API         │
│ • 状态机引擎      │     │ • 库存退还        │       │ • 价格公式        │
│                   │     │                   │       │                   │
│ Workers:          │     │ Strategies:       │       │ Calculators:      │
│ • ExcelParser     │     │ • SelfManaged     │       │ • BasePriceCalc   │
│ • AuditWorker     │     │ • SupplierManaged │       │ • PromotionMatch  │
│ • PublishWorker   │     │ • Unlimited       │       │ • FeeCalculator   │
│ • Watchdog        │     │ • Estimated       │       │ • VoucherApplier  │
└───────────────────┘     └───────────────────┘       └───────────────────┘
        │                             │                             │
        └─────────────────────────────┼─────────────────────────────┘
                                      ▼
┌───────────────────────────────────────────────────────────────────────────┐
│                    Infrastructure & Data Layer                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │  MySQL   │  │  Redis   │  │  Kafka   │  │    ES    │  │   OSS    │   │
│  │  (分库表) │  │  Cluster │  │  Events  │  │  Search  │  │  Files   │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└───────────────────────────────────────────────────────────────────────────┘
```

### 2.3 核心设计思想

1. **统一模型 + 策略模式**：
   - 商品管理：统一状态机 + 审核策略路由
   - 库存管理：(ManagementType, UnitType) 二维分类 + 策略接口
   - 价格管理：四层计算架构 + 可插拔规则引擎

2. **异步化 + 事件驱动**：
   - 所有耗时操作（文件解析、审核、发布、快照）通过 Kafka + Worker 异步处理
   - 每个状态变更都发送 Kafka 事件，下游消费者解耦处理

3. **多级缓存 + 降级保障**：
   - L1 本地缓存 + L2 Redis + L3 MySQL，保证高性能
   - 5级降级策略，保证核心链路不中断

4. **数据一致性保障**：
   - Redis 是热路径，MySQL 是权威数据源
   - Kafka 异步持久化，定时对账修复

5. **审计与追溯**：
   - 价格快照保留完整计算明细
   - 库存操作日志留痕
   - 状态变更历史完整记录

### 2.4 核心业务流

平台业务可以划分为三大核心流：

```
┌────────────────────────────────────────────────────────────────────────┐
│                          三大核心业务流                                  │
├────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  流程一：商品管理 B 端流程 (Listing Management - B2B Operations)        │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  【商品供给侧】                                                │     │
│  │  供应商/运营/商家 → 批量上传/API推送 → 审核 → 发布 → 商品上线  │     │
│  │  • Excel 批量导入（单次最多 10000 SKU）                       │     │
│  │  • 供应商 Push/Pull（实时同步/定时拉取）                       │     │
│  │  • 运营后台表单（单品/批量编辑）                               │     │
│  │                                                                │     │
│  │  【运营管理侧】                                                │     │
│  │  商品编辑 → 价格调整 → 库存管理 → 类目维护 → 首页配置         │     │
│  │  • 价格批量调整（促销价、成本价）                              │     │
│  │  • 库存批量设置（导入券码、设置库存数）                         │     │
│  │  • Entrance/Group 首页入口配置                                │     │
│  │  • Tag 标签管理（推荐、热门、新品）                            │     │
│  └──────────────────────────────────────────────────────────────┘     │
│                                                                         │
│  流程二：用户交易流 (User Journey - C2C Customer Facing)                │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  首页浏览 → 搜索/筛选 → 查看详情 → 加购 → 下单 → 支付 → 查看订单│     │
│  │  • 列表页：批量价格计算 + 库存展示                             │     │
│  │  • 详情页：实时价格 + 促销匹配 + 库存校验                       │     │
│  │  • 购物车：价格快照锁定（30分钟）                              │     │
│  │  • 下单：价格验证 + 库存预订 + 订单创建                         │     │
│  │  • 支付：支付中台 + 优惠券核销 + 积分抵扣                      │     │
│  └──────────────────────────────────────────────────────────────┘     │
│                                                                         │
│  流程三：系统履约流 (System Fulfillment - Backend Processing)          │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  支付回调 → 库存确认 → 供应商履约 → 券码发放 → 订单完成         │     │
│  │  • 库存售出：booking → sold，Kafka 异步落库                    │     │
│  │  • 供应商履约：调供应商平台 API 创建订单/出票                   │     │
│  │  • 券码发放：电子券/礼品卡卡密展示                              │     │
│  │  • 退款处理：库存回退 + 优惠券/配额归还                         │     │
│  └──────────────────────────────────────────────────────────────┘     │
│                                                                         │
└────────────────────────────────────────────────────────────────────────┘
```

#### 2.4.1 流程职责划分

| 业务流 | 核心系统 | 主要用户 | 职责范围 | 关键指标 |
|--------|---------|---------|---------|---------|
| **商品管理 B 端流程** | Listing + Inventory + Pricing | 供应商、运营、商家 | 商品供给（上架/审核/发布）<br>运营管理（批量编辑/价格调整/库存管理/配置发布） | 上架成功率、审核通过率<br>供应商同步延迟、操作效率 |
| **用户交易 C 端流程** | Pricing + Inventory + Order | 终端用户（消费者） | 商品浏览、价格展示、库存查询<br>下单、支付、订单查询 | 转化率、下单成功率<br>支付成功率 |
| **系统履约流** | Order + Inventory + Supplier Platform | 系统自动化 | 支付回调处理、库存确认<br>供应商履约、券码发放、退款处理 | 履约成功率、履约时长<br>退款处理时长 |

**三大流程的关系**：
- **B 端流程**：负责"供给"，确保平台有丰富的商品可售
- **C 端流程**：负责"销售"，为用户提供流畅的购买体验  
- **履约流程**：负责"交付"，确保订单正确履约和售后处理

---

## 三、商品管理 B 端系统（Listing Management）

> **本章涵盖**：本章描述面向运营人员、商家、供应商的 B 端商品管理系统，包括**商品供给侧**（上架、审核、发布）和**运营管理侧**（批量编辑、价格调整、库存管理、配置管理）两大核心功能。

```
┌──────────────────────────────────────────────────────────────────────┐
│              商品管理 B 端系统全景 (Listing Management)                │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │  商品供给侧 (Supply Side - 3.1~3.4)                          │    │
│  │  ───────────────────────────────────────────────────────     │    │
│  │  目标：将商品快速、准确地上架到平台                           │    │
│  │                                                               │    │
│  │  数据来源 → 审核策略 → 状态流转 → 商品发布                   │    │
│  │  • 运营表单   免审核     DRAFT → Online                       │    │
│  │  • 商家上传   人工审核   DRAFT → Pending → Approved → Online  │    │
│  │  • 供应商推送 快速通道   DRAFT → Approved → Online            │    │
│  │  • 供应商拉取 自动审核   批量处理                              │    │
│  │  • Excel 批量 异步处理   Worker 队列                          │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                              ↓                                        │
│                        商品已上线                                     │
│                              ↓                                        │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │  运营管理侧 (Operation Side - 3.5)                            │    │
│  │  ──────────────────────────────────────────────────────       │    │
│  │  目标：高效管理已上线商品，批量调整价格、库存、配置            │    │
│  │                                                               │    │
│  │  商品管理    价格管理    库存管理    类目管理    首页配置     │    │
│  │  • 批量编辑  • 批量调价  • 批量设库  • 属性维护  • Entrance  │    │
│  │  • 搜索筛选  • 促销配置  • 券码导入  • 规则配置  • Tag 管理  │    │
│  │  • 上下线    • Fee 配置  • 对账修复  • 类目树    • 灰度发布  │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                       │
│  共同特点：                                                            │
│  • 用户：运营人员、商家、供应商（B 端）                                │
│  • 场景：批量操作、高效率、低延迟要求                                  │
│  • 核心诉求：效率、准确性、可追溯                                      │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

### 3.1 商品供给侧：上架与审核

#### 3.1.1 统一状态机

所有品类共享同一套状态流转：

```
┌──────────┐
│  DRAFT   │  草稿（0）
│          │  • 运营创建/编辑商品
└─────┬────┘
      │ submit()
      ▼
┌──────────────┐
│Pending Audit │  待审核（10）
│              │  • 提交后不可编辑
└──────┬───────┘
 ┌─────┴─────┐
 │           │
 │ approve() │ reject()
 ▼           ▼
┌────────┐ ┌────────┐
│Approved│ │Rejected│  审核拒绝（12）→ 可重新提交
│  (11)  │ │  (12)  │
└───┬────┘ └────────┘
    │ publish()
    ▼
┌────────┐
│ Online │  已上线（20）→ 商品可售
│  (20)  │
└───┬────┘
    │
    ├── offline()      → Offline (21)    下线
    ├── maintain()     → Maintain (22)   维护中
    └── outOfStock()   → OutOfStock (23) 缺货
```

#### 3.1.2 审核策略路由

根据数据来源自动选择审核策略：

| 数据来源 | 审核策略 | 说明 |
|---------|---------|------|
| 供应商 Push/Pull | 快速通道（自动审核） | 仅校验必填项和格式，秒级完成 |
| 运营上传 | 免审核 | 跳过审核环节，直接发布 |
| 商家上传 | 人工审核 | 完整校验规则，推送审核队列 |
| API 接口 | 按配置决策 | 根据调用方配置 |

### 3.2 核心数据模型

#### 3.2.1 上架任务表（listing_task_tab）

```sql
CREATE TABLE listing_task_tab (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  task_code       VARCHAR(64) NOT NULL COMMENT '任务编码(雪花算法)',
  task_type       VARCHAR(50) NOT NULL COMMENT 'single_create/batch_import/supplier_sync/api_import',
  category_id     BIGINT NOT NULL COMMENT '类目ID',
  item_id         BIGINT COMMENT '商品ID(创建成功后关联)',
  
  -- 状态
  status          TINYINT NOT NULL DEFAULT 0 COMMENT '主状态(状态机)',
  sub_status      VARCHAR(50) COMMENT '子状态: processing/waiting_retry/failed',
  
  -- 任务数据
  source_type     VARCHAR(50) NOT NULL COMMENT 'operator_form/merchant_portal/excel_batch/supplier_push/supplier_pull/api',
  source_user_type VARCHAR(50) COMMENT '来源用户类型: operator/merchant/system',
  item_data       JSON NOT NULL COMMENT '商品数据(待处理)',
  validation_result JSON COMMENT '校验结果',
  error_message   TEXT COMMENT '错误信息',
  
  -- 审核信息
  audit_type      VARCHAR(50) DEFAULT 'auto' COMMENT 'auto/manual',
  auditor_id      BIGINT,
  audit_time      TIMESTAMP NULL,
  audit_comment   TEXT,
  
  -- 乐观锁
  version         INT NOT NULL DEFAULT 0,
  
  created_by      BIGINT NOT NULL,
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  UNIQUE KEY uk_task_code (task_code),
  KEY idx_category_status (category_id, status)
);
```

#### 3.2.2 审核策略配置表（listing_audit_config_tab）

```sql
CREATE TABLE listing_audit_config_tab (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  category_id     BIGINT NOT NULL,
  source_type     VARCHAR(50) NOT NULL COMMENT '数据来源类型',
  source_user_type VARCHAR(50) COMMENT '用户类型: operator/merchant/system',
  
  -- 审核策略
  audit_strategy  VARCHAR(50) NOT NULL COMMENT 'skip/auto/manual/fast_track',
  skip_audit      BOOLEAN DEFAULT FALSE,
  fast_track      BOOLEAN DEFAULT FALSE,
  require_manual  BOOLEAN DEFAULT FALSE,
  
  -- 审核规则
  validation_rules JSON COMMENT '校验规则配置',
  auto_approve_conditions JSON COMMENT '自动通过条件',
  
  UNIQUE KEY uk_category_source (category_id, source_type, source_user_type)
);

-- 示例配置
INSERT INTO listing_audit_config_tab VALUES
  (1, 'supplier_push', 'system', 'fast_track', FALSE, TRUE),  -- 供应商推送：快速通道
  (1, 'operator_form', 'operator', 'skip', TRUE, FALSE),      -- 运营上传：免审核
  (1, 'merchant_portal', 'merchant', 'manual', FALSE, FALSE); -- 商家上传：人工审核
```

### 3.3 核心流程

#### 3.3.1 单品上架流程

```
用户提交表单
  │
  ▼
1. ListingUploadService.createSingle()
   • 数据校验（必填项、格式、范围）
   • 创建 listing_task (status=DRAFT)
   • 返回 task_code
  │
  ▼
2. submit() → 状态: DRAFT → Pending (10)
   • 发送 Kafka: listing.audit.pending
  │
  ▼
3. AuditWorker 消费处理
   • 执行审核规则引擎
   • 状态: Pending → Approved (11)
   • 发送 Kafka: listing.publish.ready
  │
  ▼
4. PublishWorker 消费处理（Saga 事务）
   • 创建 item_tab / sku_tab 记录
   • 状态: Approved → Online (20)
   • 同步 ES + 清除缓存
  │
  ▼
5. 商品上线成功
```

#### 3.3.2 供应商推送同步流程

```
供应商发送变更消息 (MQ)
  │
  ▼
1. SupplierPushConsumer 消费
   • 数据映射转换
   • 创建 listing_task (source_type=supplier_push)
  │
  ▼
2. 自动审核（快速通道）
   • 仅校验必填项
   • 状态: DRAFT → Approved
  │
  ▼
3. PublishWorker
   • 创建 item/sku
   • 同步缓存和 ES
  │
  ▼
4. 商品自动上线
```

### 3.4 分布式事务（Saga）

商品发布涉及多表写入，使用 Saga 编排模式保证一致性：

```go
type PublishSaga struct {
    steps []SagaStep
}

// 发布步骤
steps := []SagaStep{
    &CreateItemStep{},      // 创建商品主体
    &CreateSKUStep{},       // 创建SKU
    &CreateAttributesStep{},// 创建属性
    &UpdateStatusStep{},    // 更新状态（最后提交）
    &PublishEventStep{},    // 发送事件（本地消息表）
    &UpdateCacheStep{},     // 更新缓存
    &SyncESStep{},          // 同步ES
}

// 执行失败 → 逆序回滚已完成的步骤
func (s *PublishSaga) compensate(ctx context.Context) {
    for i := len(s.completed) - 1; i >= 0; i-- {
        step := s.completed[i]
        step.Compensate(ctx)
    }
}
```

### 3.5 运营管理侧：批量操作与配置管理

> **职责说明**：本节描述运营人员日常使用的管理工具，包括商品编辑、价格调整、库存管理、类目维护、首页配置等批量操作功能。

#### 3.5.1 运营管理全景

```
┌────────────────────────────────────────────────────────────────┐
│                     运营管理后台 (Admin Portal)                  │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  商品管理 (Item Management)                              │   │
│  │  • 商品查询 & 筛选（按类目/状态/创建时间/供应商）        │   │
│  │  • 单品编辑（标题/描述/图片/属性）                       │   │
│  │  • 批量编辑（Excel 导入/导出）                           │   │
│  │  • 商品上下线操作                                        │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  价格管理 (Price Management)                             │   │
│  │  • 基础价格批量调整                                       │   │
│  │  • 促销活动创建 & 配置（折扣/满减/秒杀）                  │   │
│  │  • 费用规则配置（平台手续费/商户服务费/税费）              │   │
│  │  • 价格变更日志查询                                       │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  库存管理 (Inventory Management)                         │   │
│  │  • 库存批量设置（Self 类型）                              │   │
│  │  • 券码池导入（Excel/CSV）                               │   │
│  │  • 供应商同步监控 & 手动触发                             │   │
│  │  • 库存对账报告 & 差异处理                               │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  类目管理 (Category Management)                          │   │
│  │  • 类目树维护（一级/二级/三级）                           │   │
│  │  • 类目属性配置（必填项/可选项）                          │   │
│  │  • 类目关联校验规则                                       │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  首页配置 (Entrance Management)                          │   │
│  │  • FE Group 配置 & 排序                                  │   │
│  │  • Category 关联 Entrance                                │   │
│  │  • 合作方/品牌白名单配置                                  │   │
│  │  • 配置发布 & 灰度（Redis + CDN）                         │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Tag 管理 (Tag Management)                               │   │
│  │  • 标签创建（推荐/热门/新品/限时特惠）                    │   │
│  │  • 商品批量打标                                          │   │
│  │  • 标签权重配置（影响排序）                              │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

#### 3.5.2 核心运营功能

**功能 1：批量价格调整**

```go
// 批量调整价格（支持百分比/固定金额调整）
func (s *PriceOperationService) BatchAdjustPrice(req *BatchPriceAdjustRequest) error {
    // 1. 查询目标商品
    items := s.itemRepo.QueryByCategory(req.CategoryID, req.Filters)
    
    // 2. 计算新价格
    updates := make([]*PriceUpdate, 0)
    for _, item := range items {
        oldPrice := item.Price
        var newPrice decimal.Decimal
        
        switch req.AdjustType {
        case "percentage":
            newPrice = oldPrice.Mul(decimal.NewFromFloat(1 + req.AdjustValue/100))
        case "fixed_amount":
            newPrice = oldPrice.Add(decimal.NewFromFloat(req.AdjustValue))
        }
        
        updates = append(updates, &PriceUpdate{
            SKUID: item.SKUID, OldPrice: oldPrice, NewPrice: newPrice,
        })
    }
    
    // 3. 批量更新数据库
    if err := s.priceRepo.BatchUpdate(updates); err != nil {
        return err
    }
    
    // 4. 发送价格变更事件 → 失效缓存
    for _, update := range updates {
        s.eventPublisher.Publish(&PriceChangeEvent{
            SKUID: update.SKUID, OldPrice: update.OldPrice, NewPrice: update.NewPrice,
        })
    }
    
    return nil
}
```

**功能 2：库存批量设置**

```go
// 批量设置库存（Excel 导入）
func (s *InventoryOperationService) BatchSetStock(file *multipart.FileHeader) (*BatchResult, error) {
    // 1. 解析 Excel
    rows, err := parseExcel(file)  // [{sku_id, total_stock, available_stock}, ...]
    
    // 2. 数据校验
    validRows := make([]*InventoryRow, 0)
    for _, row := range rows {
        if err := s.validateInventoryRow(row); err != nil {
            recordError(row.RowNumber, err)
            continue
        }
        validRows = append(validRows, row)
    }
    
    // 3. 批量更新库存表
    for _, row := range validRows {
        s.inventoryRepo.UpdateStock(&Inventory{
            SKUID: row.SKUID,
            TotalStock: row.TotalStock,
            AvailableStock: row.AvailableStock,
        })
        
        // 4. 同步到 Redis
        s.syncStockToRedis(row.SKUID, row.AvailableStock)
    }
    
    // 5. 生成结果文件
    return &BatchResult{
        TotalCount: len(rows),
        SuccessCount: len(validRows),
        FailedCount: len(rows) - len(validRows),
        ResultFile: generateResultFile(rows),
    }, nil
}
```

**功能 3：首页 Entrance 配置**

```go
// Entrance/Group 配置发布（避免热 Key）
func (s *EntranceService) PublishEntranceConfig(req *PublishEntranceRequest) error {
    config := &EntranceConfig{
        GroupID: req.GroupID,
        Region: req.Region,
        Categories: req.Categories,
        Carriers: req.Carriers,
    }
    
    // 1. 生成配置 JSON
    configJSON, _ := json.Marshal(config)
    
    // 2. 上传到 CDN（静态资源）
    cdnURL := s.uploadToCDN(configJSON, req.Region, req.Version)
    
    // 3. 写入 Redis（分散热 Key：按用户 ID 哈希到不同 Key）
    // 避免单一热 Key，拆分为 100 个 Key
    for i := 0; i < 100; i++ {
        key := fmt.Sprintf("dp:entrance_snapshot_%d_%d:%s:%s", 
            req.GroupID, i, req.Env, req.Region)
        s.redis.Set(ctx, key, configJSON, 10*time.Minute)
    }
    
    // 4. 用户访问时根据 user_id % 100 路由到对应 Key
    // 分散流量，避免热 Key 问题
    
    return nil
}
```

#### 3.5.3 B 端效率优化成果

> **优化目标**：提升运营人员操作效率，降低人力成本，支持海量商品批量管理。

| 优化点 | 优化前 | 优化后 | 提升 |
|--------|--------|--------|------|
| **批量价格调整** | 手动逐个修改 | Excel 批量导入 + 批量 SQL | 效率提升 100 倍 |
| **券码导入** | API 逐条插入 | Excel 流式解析 + 批量写入 | 10000 条从 30 分钟降至 2 分钟 |
| **首页配置发布** | 单一 Redis Key | 100 个 Key 分散 + CDN | 热 Key QPS 从 6 万降至 600 |
| **商品搜索** | MySQL LIKE 查询 | ES 索引 + 缓存 | 查询耗时从 2s 降至 50ms |

**典型操作时间对比**：
- 单品上架：运营表单 < 3 秒（免审核），商家上传 < 5 分钟（人工审核）
- 批量上传：10000 SKU < 10 分钟（Excel 导入 + 自动审核）
- 价格调整：单品实时生效，批量 1000 SKU < 30 秒
- 库存设置：单品实时生效，批量 10000 SKU < 5 分钟

---

## 四、统一库存管理系统

### 4.1 核心概念

#### 4.1.1 二维分类模型

所有品类抽象为两个正交维度：

**维度一：谁管库存？（Management Type）**

```go
const (
    SelfManaged     = 1 // 自管理：平台维护库存数据
    SupplierManaged = 2 // 供应商管理：第三方维护，平台定期同步
    Unlimited       = 3 // 无限库存：无需库存管理
)
```

**维度二：库存长什么样？（Unit Type）**

```go
const (
    CodeBased     = 1 // 券码制：每个库存是唯一券码（Deal、Giftcard）
    QuantityBased = 2 // 数量制：库存是一个数字（OPV、本地服务）
    TimeBased     = 3 // 时间维度：按日期/时段管理（酒店、票务）
    BundleBased   = 4 // 组合型：多子项联动扣减（套餐）
)
```

#### 4.1.2 品类分类矩阵

| 品类 | 管理类型 | 单元类型 | 扣减时机 |
|------|----------|----------|----------|
| 电子券 (Deal) | Self | Code | 下单 |
| 虚拟服务券 (OPV) | Self | Quantity | 下单 |
| 酒店 | Supplier | Time | 支付 |
| 机票 | Supplier | Quantity | 支付 |
| 话费充值 | Unlimited | - | 无 |
| 礼品卡(预采购) | Self | Code | 下单 |
| 套餐组合 | Self | Bundle | 下单 |

### 4.2 核心数据模型

#### 4.2.1 库存配置表（inventory_config）

```sql
CREATE TABLE inventory_config (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  item_id         BIGINT NOT NULL,
  sku_id          BIGINT NOT NULL DEFAULT 0,
  
  -- 库存分类（核心）
  management_type INT NOT NULL COMMENT '1=自管理,2=供应商,3=无限',
  unit_type       INT NOT NULL COMMENT '1=券码,2=数量,3=时间,4=组合',
  deduct_timing   INT NOT NULL DEFAULT 1 COMMENT '1=下单,2=支付,3=发货',
  
  -- 供应商配置
  supplier_id     BIGINT NOT NULL DEFAULT 0,
  sync_strategy   INT NOT NULL DEFAULT 0 COMMENT '1=定时,2=实时,3=推送',
  sync_interval   INT NOT NULL DEFAULT 300 COMMENT '同步间隔(秒)',
  
  UNIQUE KEY uk_item_sku (item_id, sku_id)
);
```

#### 4.2.2 核心库存表（inventory）

```sql
CREATE TABLE inventory (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  item_id         BIGINT NOT NULL,
  sku_id          BIGINT NOT NULL,
  batch_id        BIGINT NOT NULL DEFAULT 0 COMMENT '批次(券码制)',
  calendar_date   DATE DEFAULT NULL COMMENT '日期(时间维度)',
  
  -- 核心库存字段
  total_stock     INT NOT NULL DEFAULT 0 COMMENT '总库存',
  available_stock INT NOT NULL DEFAULT 0 COMMENT '可售库存',
  booking_stock   INT NOT NULL DEFAULT 0 COMMENT '预订(已下单未支付)',
  locked_stock    INT NOT NULL DEFAULT 0 COMMENT '锁定(营销活动)',
  sold_stock      INT NOT NULL DEFAULT 0 COMMENT '已售',
  
  -- 供应商同步
  supplier_stock     INT NOT NULL DEFAULT 0,
  supplier_sync_time BIGINT NOT NULL DEFAULT 0,
  
  UNIQUE KEY uk_sku_batch_date (sku_id, batch_id, calendar_date)
);
```

**库存恒等式**：

```
total_stock = available_stock + booking_stock + locked_stock + sold_stock
```

#### 4.2.3 券码池表（inventory_code_pool，分100张表）

```sql
CREATE TABLE inventory_code_pool_00 (
  id           BIGINT PRIMARY KEY COMMENT '雪花算法',
  item_id      BIGINT NOT NULL,
  sku_id       BIGINT NOT NULL,
  batch_id     BIGINT NOT NULL,
  
  code         VARCHAR(255) NOT NULL COMMENT '券码(唯一)',
  serial_number VARCHAR(255) DEFAULT '' COMMENT '序列号/PIN',
  
  status       INT NOT NULL DEFAULT 1 COMMENT '1=可用,2=预订,3=已售,4=已核销,5=退款',
  order_id     BIGINT NOT NULL DEFAULT 0,
  
  booking_time  BIGINT DEFAULT 0,
  purchase_time BIGINT DEFAULT 0,
  expire_time   BIGINT DEFAULT 0,
  
  UNIQUE KEY uk_code (code),
  KEY idx_status (status)
);
-- 分表规则：item_id % 100
```

### 4.3 策略模式架构

```go
// InventoryStrategy 库存管理策略接口
type InventoryStrategy interface {
    CheckStock(ctx context.Context, req *CheckStockReq) (*CheckStockResp, error)
    BookStock(ctx context.Context, req *BookStockReq) (*BookStockResp, error)
    UnbookStock(ctx context.Context, req *UnbookStockReq) error
    SellStock(ctx context.Context, req *SellStockReq) error
    RefundStock(ctx context.Context, req *RefundStockReq) error
}

// StrategyFactory 策略工厂
func GetStrategy(mgmtType int) InventoryStrategy {
    switch mgmtType {
    case SelfManaged:
        return &SelfManagedStrategy{}
    case SupplierManaged:
        return &SupplierManagedStrategy{}
    case Unlimited:
        return &UnlimitedStrategy{}
    }
}
```

### 4.4 核心流程

#### 4.4.1 券码制出货 + 补货

```
用户下单
  │
  ▼
1. 检查库存空标志 ──── 命中 → 返回缺货
  │ 未命中
  ▼
2. 从 Redis LIST 原子出货 (Lua: LRANGE + LTRIM)
  │
  ├── 出货成功 → 步骤 4
  │
  └── 库存不足 → 3. 补货（从 MySQL 查可用券码 → RPUSH 到 Redis）
                       │
                       ├── 补货成功 → 再次出货
                       └── DB 也无库存 → 设置空标志(1h)
  │
  ▼
4. 更新 MySQL 券码状态: AVAILABLE → BOOKING
  │
  ▼
5. 发送 Kafka 事件 (异步)
```

**出货 Lua 脚本**：

```lua
-- 原子取出 N 个券码
local result = redis.call('LRANGE', KEYS[1], 0, ARGV[1] - 1)
redis.call('LTRIM', KEYS[1], ARGV[1], -1)
return result
```

#### 4.4.2 数量制预订

Redis 存储结构：

```
Key:   inventory:qty:stock:{itemID}:{skuID}
Type:  HASH
Fields:
  "available"   : 10000       # 可售库存
  "booking"     : 50          # 预订中
  "issued"      : 5000        # 已售
  "locked"      : 500         # 营销锁定
  "{promotionID}": 200        # 营销活动独立库存（动态字段）
```

**预订 Lua 脚本**：

```lua
local key = KEYS[1]
local book_num = tonumber(ARGV[1])
local promotion_id = ARGV[2]

-- 1. 获取可用库存
local available = tonumber(redis.call('HGET', key, 'available') or 0)

-- 2. 如果有营销活动，合并计算
local promo_stock = 0
if promotion_id ~= '' then
    promo_stock = tonumber(redis.call('HGET', key, promotion_id) or 0)
end
local total_available = available + promo_stock

-- 3. 检查库存
if book_num > total_available then
    return -1  -- 库存不足
end

-- 4. 优先扣营销库存，不足时扣普通库存
if promo_stock > 0 then
    if book_num <= promo_stock then
        redis.call('HINCRBY', key, promotion_id, -book_num)
    else
        redis.call('HSET', key, promotion_id, 0)
        redis.call('HINCRBY', key, 'available', -(book_num - promo_stock))
    end
else
    redis.call('HINCRBY', key, 'available', -book_num)
end

-- 5. 增加预订数
redis.call('HINCRBY', key, 'booking', book_num)

return total_available - book_num
```

#### 4.4.3 供应商异步预订（booking 轮询）

**场景**：部分供应商系统不稳定，预订流程为：
1. 创建 booking 单 → 立即返回 `booking_id`（状态 `PENDING`）
2. 轮询查询 booking 状态 → 最终返回 `CONFIRMED` / `FAILED`

```go
// 1. 用户下单时：创建 booking 单，立即返回"处理中"
func (s *SupplierManagedStrategy) BookStock(ctx context.Context, req *BookStockReq) (*BookStockResp, error) {
    // 调供应商创建 booking
    resp, err := supplierClient.CreateBooking(ctx, req.SupplierID, req.ProductID)
    
    // 保存 booking 记录（状态 PENDING）
    saveSupplierBooking(&SupplierBooking{
        OrderID:       req.OrderID,
        BookingID:     resp.BookingID,
        BookingStatus: "PENDING",
    })
    
    // 发送到 MQ 异步轮询
    publishToMQ(&BookingPollTask{OrderID: req.OrderID, BookingID: resp.BookingID})
    
    return &BookStockResp{Status: "PROCESSING"}, nil
}

// 2. 异步 Consumer：轮询 booking 状态
func PollBookingStatus(task *BookingPollTask) {
    ticker := time.NewTicker(2 * time.Second)
    timeout := time.After(30 * time.Second)
    
    for {
        select {
        case <-ticker.C:
            status, err := supplierClient.QueryBookingStatus(task.SupplierID, task.BookingID)
            
            switch status {
            case "CONFIRMED":
                handleBookingSuccess(task.OrderID)
                return
            case "FAILED":
                handleBookingFailed(task.OrderID)
                return
            }
        case <-timeout:
            handleBookingTimeout(task.OrderID)
            return
        }
    }
}
```

### 4.5 数据一致性保障

#### 4.5.1 Redis 与 MySQL 双写策略

| 操作 | Redis | MySQL | 一致性保障 |
|------|-------|-------|-----------|
| **预订 (Book)** | 同步扣减（Lua 原子） | Kafka 异步更新 | 最终一致 |
| **支付 (Sell)** | 同步更新 | Kafka 异步更新 | 最终一致 |
| **营销锁定 (Lock)** | 同步 | 同步（DB 事务） | 强一致 |
| **补货 (Replenish)** | 同步写入 | 不变 | - |

#### 4.5.2 定时对账（每小时）

```go
func Reconcile() {
    for _, cfg := range queryAllSelfManagedConfigs() {
        redisStock := getRedisAvailable(cfg.ItemID, cfg.SKUID)
        mysqlStock := getMySQLAvailable(cfg.ItemID, cfg.SKUID)
        diff := redisStock - mysqlStock

        // Redis vs MySQL 差异
        if abs(diff) > 100 || abs(diff) > mysqlStock/10 {
            alert("库存差异过大: item=%d, redis=%d, mysql=%d", 
                cfg.ItemID, redisStock, mysqlStock)
        }

        // 自动修复（可选，以 MySQL 为准）
        if cfg.AutoReconcile {
            syncRedisFromMySQL(cfg.ItemID, cfg.SKUID)
        }
    }
}
```

---

## 五、用户交易 C 端流程（User Journey）

> **本章涵盖**：本章描述面向终端用户（消费者）的 C 端交易流程，从首页浏览到支付完成的完整用户旅程，包括商品展示、价格计算、库存校验、下单支付等核心环节。

### 5.1 完整用户旅程

```
┌────────────────────────────────────────────────────────────────────────┐
│                      用户交易完整旅程 (User Journey)                     │
├────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Phase 1: 首页浏览 (Homepage Browsing)                                  │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  用户打开 APP → 加载首页 Entrance/Group 配置                  │     │
│  │  • 拉取首页配置（CDN + Redis，分散热 Key）                    │     │
│  │  • 展示品类卡片（电影/酒店/美食/娱乐/充值）                   │     │
│  │  • 展示营销 Banner（秒杀/新人专享/限时特惠）                  │     │
│  └──────────────────────────────────────────────────────────────┘     │
│         │                                                               │
│         ▼                                                               │
│  Phase 2: 商品列表 (Product List)                                       │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  用户点击品类 → 进入商品列表                                  │     │
│  │  • ES 搜索（按类目/Tag/筛选条件）                             │     │
│  │  • 批量价格计算（BatchCalculate，20-50 商品/页）             │     │
│  │  • 库存状态展示（有货/缺货/少量库存）                         │     │
│  │  • 促销标签展示（限时特惠/新人专享/买一送一）                 │     │
│  └──────────────────────────────────────────────────────────────┘     │
│         │                                                               │
│         ▼                                                               │
│  Phase 3: 商品详情 (Item Detail)                                        │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  用户点击商品 → 进入详情页                                    │     │
│  │  • 查询商品详情（L1 本地缓存 → L2 Redis → L3 MySQL）         │     │
│  │  • 实时价格计算（Calculate API）                              │     │
│  │    - Base Price: 450฿                                         │     │
│  │    - Promotion: -50฿ (新人立减)                               │     │
│  │    - Fee: +15฿ (平台手续费)                                   │     │
│  │    - Final: 415฿                                              │     │
│  │  • 库存实时查询（Redis CheckStock）                           │     │
│  │  • SKU 切换（规格选择）                                       │     │
│  │  • 推荐商品（"你可能还喜欢"）                                 │     │
│  └──────────────────────────────────────────────────────────────┘     │
│         │                                                               │
│         ▼                                                               │
│  Phase 4: 加入购物车 (Add to Cart)                                      │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  用户点击"加入购物车"                                          │     │
│  │  • 创建购物车项（cart_item_tab）                              │     │
│  │  • 生成价格快照（30 分钟有效期）                              │     │
│  │  • 锁定价格公式和优惠明细                                     │     │
│  │  • 展示"已加入购物车，共 N 件商品"                            │     │
│  └──────────────────────────────────────────────────────────────┘     │
│         │                                                               │
│         ▼                                                               │
│  Phase 5: 购物车结算 (Cart Checkout)                                    │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  用户进入购物车 → 点击"去结算"                                │     │
│  │  • 批量验证价格快照（是否过期？）                             │     │
│  │    - 未过期：使用快照价格                                     │     │
│  │    - 已过期：重新计算 → 价格变动提示用户                      │     │
│  │  • 优惠券选择（展示可用券列表）                               │     │
│  │  • 实时校验库存（批量 CheckStock）                            │     │
│  │  • 计算订单总价（Subtotal + Fee - Voucher）                  │     │
│  └──────────────────────────────────────────────────────────────┘     │
│         │                                                               │
│         ▼                                                               │
│  Phase 6: 创建订单 (Create Order)                                       │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  用户点击"提交订单"                                            │     │
│  │  • 验证价格快照（最后一次检查）                               │     │
│  │  • 库存预订（BookStock，Redis 原子扣减）                      │     │
│  │    - 成功：booking_stock += quantity                          │     │
│  │    - 失败：返回"库存不足，请选择其他商品"                      │     │
│  │  • 营销配额扣减（促销活动/优惠券配额）                         │     │
│  │  • 创建订单（order_tab，status=PENDING_PAYMENT）              │     │
│  │  • 返回订单号 + 支付二维码                                    │     │
│  └──────────────────────────────────────────────────────────────┘     │
│         │                                                               │
│         ▼                                                               │
│  Phase 7: 支付 (Payment)                                                │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  用户扫码支付 / 使用电子钱包 / 积分抵扣                       │     │
│  │  • 跳转支付中台（统一收银台）                                 │     │
│  │  • 选择支付方式（钱包余额/信用卡/借记卡）                      │     │
│  │  • 积分部分抵扣（100 积分 = 1฿）                              │     │
│  │  • 优惠券最终核销（锁定优惠券，扣减配额）                      │     │
│  │  • 支付成功 → Webhook 回调订单服务                            │     │
│  └──────────────────────────────────────────────────────────────┘     │
│         │                                                               │
│         ▼                                                               │
│  Phase 8: 查看订单 (View Order)                                         │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  支付成功 → 跳转订单详情页                                    │     │
│  │  • 订单状态：PAID → PROCESSING → COMPLETED                    │     │
│  │  • 展示价格明细（Base/Promotion/Fee/Voucher/Final）          │     │
│  │  • 券码展示（电子券直接可用）                                 │     │
│  │  • 履约进度（订单创建 → 供应商确认 → 券码发放）               │     │
│  │  • 订单操作：申请退款 / 联系客服 / 查看详情                   │     │
│  └──────────────────────────────────────────────────────────────┘     │
│                                                                         │
└────────────────────────────────────────────────────────────────────────┘
```

### 5.2 关键节点详细流程

#### 5.2.1 列表页批量价格计算

**场景**：用户浏览商品列表，单页展示 20-50 件商品，需要批量计算价格并展示促销信息。

```go
// 列表页批量价格计算（优化版）
func (s *ListPageService) LoadProductList(ctx context.Context, req *ListPageRequest) (*ListPageResponse, error) {
    // 1. ES 搜索商品（按类目/Tag/筛选条件）
    esResp, _ := s.esClient.Search(ctx, &ESSearchRequest{
        CategoryID: req.CategoryID,
        Tags: req.Tags,
        Filters: req.Filters,
        Page: req.Page,
        PageSize: 20,
    })
    
    // 2. 提取 SKU ID 列表
    skuIDs := make([]int64, 0)
    for _, item := range esResp.Items {
        skuIDs = append(skuIDs, item.DefaultSKUID)
    }
    
    // 3. 批量价格计算（单次调用）
    priceReqs := make([]*PriceRequest, len(skuIDs))
    for i, skuID := range skuIDs {
        priceReqs[i] = &PriceRequest{
            SKUID: skuID,
            Quantity: 1,
            UserID: req.UserID,
        }
    }
    priceResults, _ := s.pricingEngine.BatchCalculate(ctx, priceReqs)
    
    // 4. 批量库存查询（Redis Pipeline）
    stockResults, _ := s.inventoryService.BatchCheckStock(ctx, skuIDs)
    
    // 5. 组装返回结果
    products := make([]*ProductCard, 0)
    for i, item := range esResp.Items {
        products = append(products, &ProductCard{
            ItemID: item.ItemID,
            Title: item.Title,
            ImageURL: item.ImageURL,
            BasePrice: priceResults[i].BasePrice,
            FinalPrice: priceResults[i].FinalPrice,
            PromotionTag: priceResults[i].PromotionTag,  // "限时特惠", "新人专享"
            StockStatus: stockResults[i].Status,         // "有货", "缺货", "少量库存"
        })
    }
    
    return &ListPageResponse{Products: products}, nil
}
```

**性能优化**：
- ES 搜索 + 批量价格计算：P99 < 150ms
- 本地缓存命中：80%+，P99 < 50ms
- Redis Pipeline 批量查库存：10ms/次

#### 5.2.2 详情页实时价格计算

**场景**：用户进入详情页，需要实时计算价格并展示促销活动、优惠券、费用明细。

```go
// 详情页实时价格计算
func (s *DetailPageService) LoadItemDetail(ctx context.Context, req *DetailPageRequest) (*DetailPageResponse, error) {
    // 1. 查询商品详情（多级缓存）
    item, _ := s.itemService.GetByID(ctx, req.ItemID)
    
    // 2. 实时价格计算
    priceResp, _ := s.pricingEngine.Calculate(ctx, &PriceRequest{
        ItemID: req.ItemID,
        SKUID: req.SKUID,
        Quantity: req.Quantity,
        UserID: req.UserID,
        VoucherCodes: req.VoucherCodes,  // 用户选择的优惠券
    })
    
    // 3. 库存实时查询
    stockResp, _ := s.inventoryService.CheckStock(ctx, req.SKUID)
    
    // 4. 查询可用优惠券
    vouchers, _ := s.voucherService.GetAvailableVouchers(ctx, req.UserID, req.ItemID)
    
    // 5. 推荐商品（协同过滤）
    recommendations, _ := s.recommendService.GetRecommendations(ctx, req.UserID, req.ItemID)
    
    return &DetailPageResponse{
        Item: item,
        Price: &PriceDetail{
            BasePrice: priceResp.BasePrice,
            PromotionDiscount: priceResp.PromotionDiscount,
            PromotionDetails: priceResp.PromotionDetails,  // 具体促销活动列表
            TotalFee: priceResp.TotalFee,
            FeeDetails: priceResp.FeeDetails,              // 平台手续费, 商户服务费, 其他费用
            VoucherDiscount: priceResp.VoucherDiscount,
            FinalPrice: priceResp.FinalPrice,
            PriceFormula: priceResp.PriceFormula,          // "450฿ - 50฿ + 15฿ = 415฿"
        },
        Stock: &StockInfo{
            Available: stockResp.AvailableStock,
            Status: stockResp.Status,  // "有货" / "仅剩 5 件" / "缺货"
        },
        Vouchers: vouchers,
        Recommendations: recommendations,
    }, nil
}
```

#### 5.2.3 购物车价格快照

**场景**：用户加入购物车，需要锁定价格 30 分钟，避免结算时价格变动。

```go
// 加入购物车 + 价格快照
func (s *CartService) AddToCart(ctx context.Context, req *AddToCartRequest) (*AddToCartResponse, error) {
    // 1. 实时计算价格
    priceResp, _ := s.pricingEngine.Calculate(ctx, &PriceRequest{
        ItemID: req.ItemID,
        SKUID: req.SKUID,
        Quantity: req.Quantity,
        UserID: req.UserID,
    })
    
    // 2. 生成价格快照（30 分钟有效）
    snapshot := &PriceSnapshot{
        SnapshotCode: generateSnapshotCode(),
        UserID: req.UserID,
        SKUID: req.SKUID,
        Quantity: req.Quantity,
        BasePrice: priceResp.BasePrice,
        PromotionDiscount: priceResp.PromotionDiscount,
        PromotionDetails: priceResp.PromotionDetails,
        TotalFee: priceResp.TotalFee,
        FeeDetails: priceResp.FeeDetails,
        FinalPrice: priceResp.FinalPrice,
        PriceFormula: priceResp.PriceFormula,
        ExpiredAt: time.Now().Add(30 * time.Minute),
    }
    s.snapshotRepo.Create(ctx, snapshot)
    
    // 3. 创建购物车项
    cartItem := &CartItem{
        UserID: req.UserID,
        ItemID: req.ItemID,
        SKUID: req.SKUID,
        Quantity: req.Quantity,
        SnapshotCode: snapshot.SnapshotCode,  // 关联价格快照
    }
    s.cartRepo.Create(ctx, cartItem)
    
    return &AddToCartResponse{
        Success: true,
        SnapshotCode: snapshot.SnapshotCode,
        FinalPrice: priceResp.FinalPrice,
    }, nil
}
```

#### 5.2.4 下单流程（核心）

```go
// 创建订单（含库存预订 + 配额扣减）
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
    // Step 1: 验证价格快照
    snapshot, err := s.snapshotRepo.GetByCode(ctx, req.SnapshotCode)
    if err != nil {
        return nil, errors.New("价格快照不存在")
    }
    if snapshot.ExpiredAt.Before(time.Now()) {
        // 快照已过期，重新计算价格
        newPrice, _ := s.pricingEngine.Calculate(ctx, &PriceRequest{
            ItemID: snapshot.ItemID, SKUID: snapshot.SKUID, Quantity: snapshot.Quantity,
        })
        if !newPrice.FinalPrice.Equal(snapshot.FinalPrice) {
            return nil, errors.New("价格已变动，请重新确认")
        }
    }
    
    // Step 2: 库存预订（原子操作）
    bookResp, err := s.inventoryService.BookStock(ctx, &BookStockReq{
        ItemID: snapshot.ItemID,
        SKUID: snapshot.SKUID,
        Quantity: snapshot.Quantity,
        OrderID: generateOrderID(),
    })
    if err != nil || !bookResp.Success {
        return nil, errors.New("库存不足")
    }
    
    // Step 3: 营销配额扣减（Redis Lua 原子操作）
    if len(snapshot.PromotionDetails) > 0 {
        for _, promo := range snapshot.PromotionDetails {
            if err := s.promotionService.ConsumeQuota(ctx, promo.ActivityID, 1); err != nil {
                // 配额不足，回滚库存
                s.inventoryService.UnbookStock(ctx, bookResp.BookingID)
                return nil, errors.New("活动配额已用完")
            }
        }
    }
    
    // Step 4: 优惠券锁定（待支付后核销）
    if req.VoucherCode != "" {
        if err := s.voucherService.LockVoucher(ctx, req.VoucherCode, req.UserID); err != nil {
            // 优惠券锁定失败，回滚库存和配额
            s.inventoryService.UnbookStock(ctx, bookResp.BookingID)
            return nil, errors.New("优惠券不可用")
        }
    }
    
    // Step 5: 创建订单
    order := &Order{
        OrderID: bookResp.OrderID,
        UserID: req.UserID,
        ItemID: snapshot.ItemID,
        SKUID: snapshot.SKUID,
        Quantity: snapshot.Quantity,
        SnapshotCode: snapshot.SnapshotCode,
        Status: OrderStatusPendingPayment,
        TotalAmount: snapshot.FinalPrice,
    }
    s.orderRepo.Create(ctx, order)
    
    // Step 6: 发送 Kafka 事件
    s.eventPublisher.Publish(&OrderCreatedEvent{OrderID: order.OrderID})
    
    return &CreateOrderResponse{
        OrderID: order.OrderID,
        PaymentURL: s.generatePaymentURL(order.OrderID),
    }, nil
}
```

### 5.3 用户体验优化

#### 5.3.1 价格变动提示

```go
// 购物车结算时检查价格变动
func (s *CartService) CheckPriceChange(ctx context.Context, snapshotCode string) (*PriceChangeAlert, error) {
    snapshot, _ := s.snapshotRepo.GetByCode(ctx, snapshotCode)
    
    // 重新计算当前价格
    currentPrice, _ := s.pricingEngine.Calculate(ctx, &PriceRequest{
        ItemID: snapshot.ItemID, SKUID: snapshot.SKUID, Quantity: snapshot.Quantity,
    })
    
    priceDiff := currentPrice.FinalPrice.Sub(snapshot.FinalPrice)
    
    if priceDiff.Abs().GreaterThan(decimal.NewFromFloat(0.01)) {
        return &PriceChangeAlert{
            HasChange: true,
            OldPrice: snapshot.FinalPrice,
            NewPrice: currentPrice.FinalPrice,
            ChangeReason: "促销活动已结束",  // 或 "基础价格调整"
            Message: fmt.Sprintf("价格已从 %s฿ 变为 %s฿", 
                snapshot.FinalPrice.String(), currentPrice.FinalPrice.String()),
        }, nil
    }
    
    return &PriceChangeAlert{HasChange: false}, nil
}
```

#### 5.3.2 缺货降级提示

```
当前商品暂时缺货，您可以：
1. 查看相似商品（推荐 3 款同类商品）
2. 到货通知（输入手机号，库存补充时短信提醒）
3. 联系客服咨询（跳转客服对话）
```

---

## 六、统一价格管理系统

### 6.1 核心概念

#### 6.1.1 四层价格架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                     统一价格计算架构                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Layer 1: 基础价格层 (Base Price Layer)                              │
│  ┌──────────────────────────────────────────────────┐               │
│  │  • sku_tab.price             (SKU 基础价)         │               │
│  │  • hotel_price_calendar_tab  (酒店价格日历)       │               │
│  │  • dynamic_pricing_rule_tab  (动态定价规则)       │               │
│  └──────────────────────────────────────────────────┘               │
│                          ↓                                           │
│  Layer 2: 营销价格层 (Promotion Price Layer)                         │
│  ┌──────────────────────────────────────────────────┐               │
│  │  • promotion_activity_tab    (营销活动主表)       │               │
│  │  • promotion_rule_tab        (规则配置)           │               │
│  │  • promotion_priority_tab    (优先级 & 互斥)      │               │
│  └──────────────────────────────────────────────────┘               │
│                          ↓                                           │
│  Layer 3: 费用层 (Fee Layer)                                         │
│  ┌──────────────────────────────────────────────────┐               │
│  │  • fee_config_tab            (费用配置表)         │               │
│  │  • fee_type: dp_fee, hub_fee, service_fee, tax   │               │
│  └──────────────────────────────────────────────────┘               │
│                          ↓                                           │
│  Layer 4: 优惠券层 (Voucher Layer)                                   │
│  ┌──────────────────────────────────────────────────┐               │
│  │  • voucher_tab               (优惠券主表)         │               │
│  │  • user_voucher_tab          (用户持有)           │               │
│  └──────────────────────────────────────────────────┘               │
│                          ↓                                           │
│  ═══════════════════════════════════════════════════                 │
│  最终价格 = Base - Promotion + Fee - Voucher                         │
│  ═══════════════════════════════════════════════════                 │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

#### 6.1.2 价格组件定义

| 组件 | 说明 | 计算方式 | 示例 |
|------|------|----------|------|
| **Base Price** | SKU 基础售价 | 固定 / 时间动态 | 480฿（电影票） |
| **Promotion Discount** | 营销活动折扣 | 百分比 / 固定金额 / 满减 | -50฿（新用户立减） |
| **Platform Fee** | 平台手续费 | 固定 / 百分比 | +10฿ |
| **Merchant Service Fee** | 商户服务费 | 固定 / 百分比 / 阶梯 | +5฿ |
| **Voucher Discount** | 优惠券抵扣 | 固定 / 百分比 / 封顶 | -30฿ |
| **Final Price** | 最终支付价格 | Base - Promo + Fee - Voucher | 435฿ |

### 6.2 核心数据模型

#### 6.2.1 营销活动表（promotion_activity_tab）

```sql
CREATE TABLE promotion_activity_tab (
    id              BIGINT NOT NULL AUTO_INCREMENT,
    activity_code   VARCHAR(64) NOT NULL COMMENT '活动编码',
    activity_name   VARCHAR(255) NOT NULL COMMENT '活动名称',
    activity_type   VARCHAR(50) NOT NULL COMMENT '活动类型: discount, full_reduction, bundle, flash_sale',
    category_ids    JSON COMMENT '适用类目',
    item_ids        JSON COMMENT '指定商品ID',
    user_type       VARCHAR(50) DEFAULT 'all' COMMENT '用户类型: all, new, vip',
    discount_type   VARCHAR(50) NOT NULL COMMENT '折扣类型: percentage, fixed_amount, full_reduction',
    discount_value  JSON NOT NULL COMMENT '折扣配置',
    priority        INT DEFAULT 0 COMMENT '优先级（数字越大越优先）',
    exclusivity     TINYINT DEFAULT 0 COMMENT '互斥性: 0=可叠加, 1=与其他活动互斥',
    voucher_compatible TINYINT DEFAULT 1 COMMENT '优惠券兼容: 0=互斥, 1=可叠加',
    status          TINYINT DEFAULT 1,
    start_time      DATETIME NOT NULL,
    end_time        DATETIME NOT NULL,
    total_quota     INT COMMENT '总配额',
    used_quota      INT DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uk_activity_code (activity_code)
);
```

**discount_value** JSON 配置：

| 折扣类型 | JSON 配置示例 |
|---------|--------------|
| percentage | `{"percentage": 20}` |
| fixed_amount | `{"amount": 50}` |
| full_reduction | `{"threshold": 3000, "discount": 200}` |

#### 6.2.2 费用配置表（fee_config_tab）

```sql
CREATE TABLE fee_config_tab (
    id              BIGINT NOT NULL AUTO_INCREMENT,
    fee_code        VARCHAR(64) NOT NULL,
    fee_name        VARCHAR(255) NOT NULL,
    fee_type        VARCHAR(50) NOT NULL COMMENT 'platform_fee, merchant_service_fee, service_fee, tax',
    category_id     BIGINT COMMENT '类目ID',
    calculation_type VARCHAR(50) NOT NULL COMMENT 'fixed, percentage, tiered',
    calculation_config JSON NOT NULL,
    can_be_discounted TINYINT DEFAULT 0 COMMENT '是否可被优惠券抵扣',
    PRIMARY KEY (id),
    UNIQUE KEY uk_fee_code (fee_code)
);
```

#### 6.2.3 价格快照表（price_snapshot_tab）

```sql
CREATE TABLE price_snapshot_tab (
    id              BIGINT NOT NULL AUTO_INCREMENT,
    snapshot_code   VARCHAR(64) NOT NULL COMMENT '快照编码',
    order_id        BIGINT COMMENT '订单ID',
    user_id         BIGINT NOT NULL,
    sku_id          BIGINT NOT NULL,
    quantity        INT NOT NULL DEFAULT 1,
    currency        VARCHAR(3) NOT NULL,
    
    -- 基础价格
    base_price      DECIMAL(20,2) NOT NULL,
    subtotal        DECIMAL(20,2) NOT NULL,
    
    -- 营销折扣
    promotion_discount DECIMAL(20,2) DEFAULT 0,
    promotion_details JSON,
    
    -- 费用
    total_fee       DECIMAL(20,2) DEFAULT 0,
    fee_details     JSON,
    
    -- 优惠券
    voucher_discount DECIMAL(20,2) DEFAULT 0,
    voucher_details JSON,
    
    -- 最终价格
    final_price     DECIMAL(20,2) NOT NULL,
    price_formula   TEXT COMMENT '价格计算公式（人类可读）',
    
    expired_at      DATETIME COMMENT '快照过期时间（30分钟）',
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_snapshot_code (snapshot_code),
    KEY idx_order (order_id)
);
```

### 6.3 价格计算引擎

```go
// Calculate 计算价格（核心入口）
func (e *PricingEngine) Calculate(ctx context.Context, req *PriceRequest) (*PriceResponse, error) {
    resp := &PriceResponse{Currency: req.Currency}

    // Step 1: 计算基础价格
    basePrice, err := e.basePriceCalculator.Calculate(ctx, req)
    resp.BasePrice = basePrice.Price
    resp.Subtotal = basePrice.Price.Mul(decimal.NewFromInt(int64(req.Quantity)))

    // Step 2: 匹配并应用营销活动（支持降级）
    promotions, err := e.matchPromotionsWithDegrade(ctx, req, basePrice.Price)
    totalDiscount := decimal.Zero
    for _, promo := range promotions {
        discount := e.calculatePromotionDiscount(promo, resp.Subtotal)
        totalDiscount = totalDiscount.Add(discount)
        resp.PromotionDetails = append(resp.PromotionDetails, PromotionDetail{
            ActivityID: promo.ActivityID, Discount: discount,
        })
    }
    resp.PromotionDiscount = totalDiscount

    // Step 3: 计算费用（支持降级）
    fees, err := e.calculateFeesWithDegrade(ctx, req, basePrice.Price)
    totalFee := decimal.Zero
    for _, fee := range fees {
        totalFee = totalFee.Add(fee.Amount)
        resp.FeeDetails = append(resp.FeeDetails, FeeDetail{
            FeeType: fee.FeeType, Amount: fee.Amount,
        })
    }
    resp.TotalFee = totalFee

    // Step 4: 应用优惠券（支持降级）
    if len(req.VoucherCodes) > 0 {
        voucherResult, _ := e.applyVouchersWithDegrade(ctx, req, resp)
        resp.VoucherDiscount = voucherResult.TotalDiscount
        resp.VoucherDetails = voucherResult.Details
    }

    // Step 5: 计算最终价格
    resp.FinalPrice = resp.Subtotal.
        Sub(resp.PromotionDiscount).
        Add(resp.TotalFee).
        Sub(resp.VoucherDiscount)

    // 确保价格不为负
    if resp.FinalPrice.LessThan(decimal.Zero) {
        resp.FinalPrice = decimal.Zero
    }

    // Step 6: 币种精度对齐
    resp.FinalPrice = e.roundByCurrency(resp.FinalPrice, req.Currency)

    // Step 7: 生成价格公式（人类可读）
    resp.PriceFormula = e.buildPriceFormula(resp)

    // Step 8: 生成价格快照（异步写入）
    go e.snapshotGenerator.Generate(context.Background(), req, resp)

    return resp, nil
}
```

### 6.4 币种精度处理

| 币种 | 代码 | 小数位 | 示例 |
|------|------|--------|------|
| 泰铢 | THB | 2 | 480.50฿ |
| 越南盾 | VND | 0 | 120000₫ |
| 印尼盾 | IDR | 0 | 85000Rp |

```go
func (e *PricingEngine) roundByCurrency(amount decimal.Decimal, currency string) decimal.Decimal {
    switch currency {
    case "VND", "IDR":
        return amount.Ceil()  // 向上取整到最小货币单位
    case "THB", "MYR", "SGD", "PHP":
        return amount.Round(2)  // 2位小数，银行家舍入
    }
}
```

### 6.5 降级策略（5级降级）

```
┌─────────────────────────────────────────────────────────────┐
│ Level 0  │ 全部正常          │ 完整计算（Base+Promo+Fee+Voucher）│
│ Level 1  │ 优惠券服务不可用   │ 跳过优惠券，提示用户稍后重试    │
│ Level 2  │ 促销服务不可用     │ 仅 Base + Fee，无促销折扣      │
│ Level 3  │ 费用服务不可用     │ 仅 Base，无额外费用            │
│ Level 4  │ 缓存全部失效       │ 直接查询 MySQL 基础价格        │
│ Level 5  │ MySQL 也不可用     │ 返回上一次缓存的价格快照       │
└─────────────────────────────────────────────────────────────┘
```

---

## 七、三大系统协作流程

> **本章涵盖**：本章描述 Listing、Inventory、Pricing 三大核心系统的协作流程，包括完整下单流程、秒杀场景、供应商同步场景、系统履约流程等关键业务场景的系统协作细节。

### 7.1 完整下单流程

```
┌──────────────────────────────────────────────────────────────────────┐
│                      商品从上架到售卖的完整链路                        │
└──────────────────────────────────────────────────────────────────────┘

Phase 1: 商品上架
  │
  ├─ 供应商推送/运营上传 → ListingService
  ├─ 创建 listing_task → 审核策略路由
  ├─ PublishWorker (Saga) → 创建 item/sku
  └─ 商品状态: DRAFT → Approved → Online

Phase 2: 库存初始化
  │
  ├─ 读取 inventory_config → 确定管理类型和单元类型
  ├─ Self+Code: 券码批量导入 → 写入 code_pool → 预热到 Redis LIST
  ├─ Self+Quantity: 设置 total_stock → 同步到 Redis HASH
  ├─ Supplier: 配置 sync_strategy → 定时拉取/实时推送
  └─ 库存状态: 可售

Phase 3: 价格配置
  │
  ├─ 配置 SKU 基础价 → sku_tab.price
  ├─ 创建营销活动 → promotion_activity_tab
  ├─ 配置费用规则 → fee_config_tab
  └─ 价格生效 → 缓存预热

Phase 4: 用户浏览
  │
  ├─ 列表页: BatchCalculate() → 批量价格计算（缓存命中）
  ├─ 详情页: Calculate() → 实时价格计算
  └─ 返回: final_price, promotion_details, fee_details

Phase 5: 加入购物车
  │
  ├─ CartService.AddToCart()
  ├─ 实时计算价格 → 生成价格快照（30分钟有效）
  ├─ 检查库存可用性 → CheckStock()
  └─ 购物车关联: cart_item.snapshot_code

Phase 6: 用户下单
  │
  ├─ OrderService.CreateOrder()
  ├─ 验证价格快照 → 检查是否过期
  │   ├─ 未过期 → 使用快照价格
  │   └─ 已过期 → 重新计算 → 价格变动提示用户确认
  │
  ├─ 库存预订 → InventoryManager.BookStock()
  │   ├─ Self+Code: Redis LIST LPOP 券码 → MySQL 状态 AVAILABLE→BOOKING
  │   ├─ Self+Quantity: Redis HASH Lua 原子扣减 available
  │   ├─ Supplier: 调供应商 booking API / 异步轮询
  │   └─ 预订成功 → booking_stock += quantity
  │
  └─ 订单创建: status=PENDING_PAYMENT

Phase 7: 用户支付
  │
  ├─ PaymentService.ProcessPayment()
  ├─ 支付成功回调 → OrderService.OnPaymentSuccess()
  │
  ├─ 库存确认 → InventoryManager.SellStock()
  │   ├─ Self+Code: MySQL 券码状态 BOOKING→SOLD
  │   ├─ Self+Quantity: Redis booking--, sold++
  │   ├─ Supplier: 调供应商确认接口
  │   ├─ Giftcard(实时生成): 调供应商 API 生成卡密
  │   └─ MySQL: booking_stock--, sold_stock++
  │
  ├─ 优惠券核销 → VoucherService.Consume()
  ├─ 配额扣减 → PromotionService.ConsumeQuota()
  └─ 订单状态: PAID → PROCESSING

Phase 8: 订单履约
  │
  ├─ 电子券: 券码发放 → 用户可查看
  ├─ 酒店/机票: 确认单发送
  ├─ 礼品卡: 卡密展示（脱敏）
  └─ 订单状态: COMPLETED

Phase 9: 退款场景
  │
  ├─ RefundService.ProcessRefund()
  │
  ├─ 库存回退 → InventoryManager.RefundStock()
  │   ├─ Self+Code: MySQL 券码状态 SOLD→REFUND, RPUSH 回 Redis
  │   ├─ Self+Quantity: Redis sold--, available++
  │   ├─ Supplier: 调供应商取消接口
  │   └─ MySQL: sold_stock--, available_stock++
  │
  ├─ 优惠券回退 → VoucherService.Rollback()
  ├─ 配额回退 → PromotionService.RollbackQuota()
  └─ 订单状态: REFUNDED
```

### 7.2 秒杀场景高并发链路

```
秒杀开始（20k QPS）
  │
  ▼
1. 价格计算（并行批量）
   • BatchCalculate() 批量查基础价
   • 本地缓存拦截 80% 读请求
   • P99 < 30ms
  │
  ▼
2. 库存预订（Redis 原子操作）
   • Self+Quantity: Redis Lua 原子扣减
   • 配额管理: Redis Lua 原子扣减
   • 无 DB 查询，纯内存操作
   • P99 < 10ms
  │
  ▼
3. 订单创建（异步落库）
   • 订单写入 MySQL（批量写入优化）
   • 库存变更通过 Kafka 异步持久化
   • 不阻塞下单主流程
  │
  ▼
4. 支付成功（异步确认）
   • 库存售出 → Kafka 异步更新 MySQL
   • 券码状态变更 → 批量更新
   • 对账机制保证最终一致
```

### 7.3 供应商同步场景

```
定时任务触发（每 5 分钟）
  │
  ▼
1. SupplierPullScheduler
   • 读取 last_sync_time
   • 调供应商 API: GET /api/hotels/changes?since=xxx
   • 获取增量数据（酒店+房型+价格日历）
  │
  ▼
2. 数据转换 & 批量上架
   • 供应商数据 → 平台数据映射
   • 批量创建 listing_task (source_type=supplier_pull)
   • 创建 listing_batch_task
  │
  ▼
3. 批量审核（快速通道）
   • BatchAutoAuditWorker
   • 校验价格日历合法性
  │
  ▼
4. 批量发布
   • BatchPublishWorker
   • 批量创建 item/sku
   • 批量创建价格日历记录
  │
  ▼
5. 库存同步
   • InventoryService.SyncFromSupplier()
   • 更新 inventory.supplier_stock
   • 同步到 Redis 缓存
  │
  ▼
6. 价格同步
   • 批量更新 sku_tab.price
   • 发送 Kafka: price_changed
   • 缓存失效 + 预热
  │
  ▼
7. 更新 last_sync_time
```

### 7.4 系统履约核心流（Fulfillment）

#### 7.4.1 完整履约流程

```
┌────────────────────────────────────────────────────────────────────────┐
│                     系统履约流 (Fulfillment Pipeline)                    │
├────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Phase 1: 支付成功回调 (Payment Callback)                               │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  支付中台 → Webhook 回调 → 订单服务                           │     │
│  │  • 验证签名（防止伪造回调）                                    │     │
│  │  • 幂等性校验（防止重复回调）                                  │     │
│  │  • 订单状态校验：PENDING_PAYMENT → PAID                       │     │
│  │  • 发送 Kafka: order.paid                                     │     │
│  └──────────────────────────────────────────────────────────────┘     │
│         │                                                               │
│         ▼                                                               │
│  Phase 2: 库存确认 (Stock Confirmation)                                 │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  FulfillmentWorker 消费 order.paid 事件                       │     │
│  │  • 调用 InventoryService.SellStock()                          │     │
│  │  • 券码制：MySQL 券码状态 BOOKING → SOLD                      │     │
│  │  • 数量制：Redis booking--, sold++, MySQL 异步落库            │     │
│  │  • 供应商管理：调供应商确认接口 (confirm booking)              │     │
│  │  • 失败自动重试（最多 3 次，指数退避）                         │     │
│  └──────────────────────────────────────────────────────────────┘     │
│         │                                                               │
│         ▼                                                               │
│  Phase 3: 优惠券核销 (Voucher Consumption)                              │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  VoucherService.ConsumeVoucher()                              │     │
│  │  • 用户优惠券状态：LOCKED → USED                               │     │
│  │  • 优惠券配额扣减（Redis 原子操作）                            │     │
│  │  • 记录优惠券使用日志                                          │     │
│  └──────────────────────────────────────────────────────────────┘     │
│         │                                                               │
│         ▼                                                               │
│  Phase 4: 供应商履约 (Supplier Fulfillment)                             │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  调用供应商平台 API 创建供应商订单                             │     │
│  │  • 电子券 (Deal): 无需供应商履约，券码已预分配                 │     │
│  │  • 虚拟服务 (OPV): 调供应商平台 API 创建服务单                 │     │
│  │  • 酒店: 调供应商 API 创建预订单 (booking reference)           │     │
│  │  • 机票: 调航司 API 出票 (PNR + ticket number)                │     │
│  │  • 礼品卡 (实时生成): 调供应商 API 生成卡密                    │     │
│  │  • 失败处理：自动重试 → 人工介入 → 补发/退款                   │     │
│  └──────────────────────────────────────────────────────────────┘     │
│         │                                                               │
│         ▼                                                               │
│  Phase 5: 券码发放/信息展示 (Code Distribution)                         │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  根据品类类型处理：                                            │     │
│  │  • 电子券: 直接展示券码 + 二维码 + 使用说明                    │     │
│  │  • 礼品卡: 展示卡号 + 卡密（脱敏：****1234）                  │     │
│  │  • 酒店: 展示确认号 + 预订详情                                │     │
│  │  • 机票: 展示 PNR + 电子票号 + 行程单                         │     │
│  │  • 虚拟服务: 展示核销码 + 商家地址 + 联系方式                  │     │
│  └──────────────────────────────────────────────────────────────┘     │
│         │                                                               │
│         ▼                                                               │
│  Phase 6: 订单完成 (Order Completion)                                   │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  订单状态：PAID → PROCESSING → COMPLETED                      │     │
│  │  • 发送确认短信/邮件（券码、确认号）                           │     │
│  │  • 发送 Push 通知（"您的订单已完成，可立即使用"）              │     │
│  │  • 更新用户订单列表                                            │     │
│  │  • 触发积分奖励（如有）                                        │     │
│  └──────────────────────────────────────────────────────────────┘     │
│                                                                         │
└────────────────────────────────────────────────────────────────────────┘
```

#### 7.4.2 履约异常处理

```go
// 履约失败自动重试 + 降级处理
func (w *FulfillmentWorker) processFulfillment(ctx context.Context, order *Order) error {
    maxRetries := 3
    backoff := time.Second
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        // Step 1: 库存确认
        if err := w.inventoryService.SellStock(ctx, &SellStockReq{
            OrderID: order.OrderID,
            ItemID: order.ItemID,
            SKUID: order.SKUID,
            Quantity: order.Quantity,
        }); err != nil {
            log.Errorf("库存确认失败 attempt=%d, error=%v", attempt, err)
            time.Sleep(backoff)
            backoff *= 2  // 指数退避
            continue
        }
        
        // Step 2: 供应商履约
        if order.RequireSupplierFulfillment() {
            supplierResp, err := w.supplierClient.CreateSupplierOrder(ctx, &SupplierOrderRequest{
                OrderID: order.OrderID,
                SupplierID: order.SupplierID,
                ProductCode: order.SupplierProductCode,
                Quantity: order.Quantity,
            })
            if err != nil {
                log.Errorf("供应商履约失败 attempt=%d, error=%v", attempt, err)
                
                if isRetryable(err) {
                    time.Sleep(backoff)
                    backoff *= 2
                    continue
                } else {
                    // 不可重试错误（如供应商缺货）→ 直接失败
                    return w.handleFulfillmentFailed(ctx, order, err)
                }
            }
            
            // 更新供应商订单号
            order.SupplierOrderID = supplierResp.SupplierOrderID
            w.orderRepo.Update(ctx, order)
        }
        
        // Step 3: 券码发放
        if err := w.distributeCode(ctx, order); err != nil {
            log.Errorf("券码发放失败 attempt=%d, error=%v", attempt, err)
            time.Sleep(backoff)
            backoff *= 2
            continue
        }
        
        // Step 4: 更新订单状态
        order.Status = OrderStatusCompleted
        order.CompletedAt = time.Now()
        w.orderRepo.Update(ctx, order)
        
        // Step 5: 发送通知
        w.notificationService.SendOrderCompletedNotification(ctx, order)
        
        return nil
    }
    
    // 重试 3 次全部失败 → 人工介入队列
    return w.handleFulfillmentFailed(ctx, order, errors.New("max retries exceeded"))
}

// 履约失败处理
func (w *FulfillmentWorker) handleFulfillmentFailed(ctx context.Context, order *Order, err error) error {
    // 1. 订单状态标记为履约失败
    order.Status = OrderStatusFulfillmentFailed
    order.FailureReason = err.Error()
    w.orderRepo.Update(ctx, order)
    
    // 2. 推送到人工处理队列
    w.manualInterventionQueue.Push(&ManualTask{
        TaskType: "fulfillment_failed",
        OrderID: order.OrderID,
        Reason: err.Error(),
        CreatedAt: time.Now(),
    })
    
    // 3. 通知用户（客服将联系处理）
    w.notificationService.SendFulfillmentDelayNotification(ctx, order)
    
    // 4. 告警运维
    w.alertService.Alert("履约失败", map[string]interface{}{
        "order_id": order.OrderID,
        "reason": err.Error(),
    })
    
    return nil
}
```

#### 7.4.3 退款履约流程

```go
// 退款履约（库存回退 + 优惠券归还）
func (s *RefundService) ProcessRefund(ctx context.Context, req *RefundRequest) error {
    order, _ := s.orderRepo.GetByID(ctx, req.OrderID)
    
    // Step 1: 创建退款单
    refund := &Refund{
        RefundID: generateRefundID(),
        OrderID: req.OrderID,
        Amount: order.TotalAmount,
        Reason: req.Reason,
        Status: RefundStatusPending,
    }
    s.refundRepo.Create(ctx, refund)
    
    // Step 2: 库存回退
    if err := s.inventoryService.RefundStock(ctx, &RefundStockReq{
        OrderID: order.OrderID,
        ItemID: order.ItemID,
        SKUID: order.SKUID,
        Quantity: order.Quantity,
    }); err != nil {
        return errors.Wrap(err, "库存回退失败")
    }
    // 券码制：SOLD → REFUND, RPUSH 回 Redis
    // 数量制：sold--, available++
    
    // Step 3: 供应商取消
    if order.SupplierOrderID != "" {
        if err := s.supplierClient.CancelSupplierOrder(ctx, order.SupplierOrderID); err != nil {
            log.Errorf("供应商取消失败: %v", err)
            // 失败不阻塞退款，后续人工处理
        }
    }
    
    // Step 4: 优惠券回退
    if order.VoucherCode != "" {
        s.voucherService.RollbackVoucher(ctx, order.VoucherCode, order.UserID)
        // 用户优惠券状态：USED → UNUSED
    }
    
    // Step 5: 营销配额回退
    if order.PromotionActivityID > 0 {
        s.promotionService.RollbackQuota(ctx, order.PromotionActivityID, order.Quantity)
    }
    
    // Step 6: 发起退款（调支付中台）
    paymentResp, err := s.paymentClient.Refund(ctx, &PaymentRefundRequest{
        OrderID: order.OrderID,
        Amount: order.TotalAmount,
    })
    if err != nil {
        return errors.Wrap(err, "退款失败")
    }
    
    // Step 7: 更新订单和退款单状态
    order.Status = OrderStatusRefunded
    refund.Status = RefundStatusCompleted
    refund.CompletedAt = time.Now()
    s.orderRepo.Update(ctx, order)
    s.refundRepo.Update(ctx, refund)
    
    // Step 8: 通知用户
    s.notificationService.SendRefundCompletedNotification(ctx, order, refund)
    
    return nil
}
```

---

## 八、数据一致性保障

### 8.1 分布式事务

#### 8.1.1 Saga 模式（商品发布）

```go
type PublishSaga struct {
    steps []SagaStep
}

steps := []SagaStep{
    &CreateItemStep{},      // 创建商品主体
    &CreateSKUStep{},       // 创建SKU
    &CreatePriceStep{},     // 创建价格
    &InitInventoryStep{},   // 初始化库存
    &UpdateStatusStep{},    // 更新状态（最后提交）
    &PublishEventStep{},    // 发送事件（本地消息表）
}

// 执行失败 → 逆序回滚
func (s *PublishSaga) compensate(ctx context.Context) {
    for i := len(s.completed) - 1; i >= 0; i-- {
        s.completed[i].Compensate(ctx)
    }
}
```

#### 8.1.2 本地消息表（可靠事件发布）

```sql
CREATE TABLE listing_outbox_tab (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  task_id         BIGINT NOT NULL,
  event_type      VARCHAR(50) NOT NULL,
  event_payload   JSON NOT NULL,
  status          VARCHAR(50) DEFAULT 'pending',
  retry_count     INT DEFAULT 0,
  published_at    TIMESTAMP NULL,
  KEY idx_status_retry (status, published_at)
);
```

```go
// 1. 本地事务写入消息表
func (s *PublishEventStep) Execute(ctx context.Context) error {
    outbox := &OutboxMessage{
        TaskID:       s.taskID,
        EventType:    "listing.published",
        EventPayload: payload,
        Status:       "pending",
    }
    db.Create(outbox)  // 与业务数据在同一事务
    return nil
}

// 2. 独立 Worker 轮询发送
func (p *OutboxPublisher) publishPendingMessages() {
    messages := db.Where("status = 'pending'").Limit(100).Find()
    for _, msg := range messages {
        err := kafka.Publish("listing.events", msg.EventPayload)
        if err == nil {
            db.Update(&msg, "status", "published")
        } else {
            // 指数退避重试
        }
    }
}
```

### 8.2 缓存一致性

#### 8.2.1 多级缓存架构

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  L1: 本地缓存  │───▶│  L2: Redis    │───▶│  L3: MySQL    │
│  (sync.Map)   │    │  (Cluster)    │    │  (主从)       │
│  TTL: 10s     │    │  TTL: 5min    │    │  持久存储      │
└──────────────┘    └──────────────┘    └──────────────┘
```

#### 8.2.2 缓存失效策略

```go
// 价格变更时主动失效缓存（Kafka 驱动）
func (h *PriceCacheInvalidHandler) HandlePriceChange(ctx context.Context, event *PriceChangeEvent) error {
    keys := []string{
        fmt.Sprintf("price:base:%d", event.SKUID),
        fmt.Sprintf("price:calc:%d:*", event.SKUID),
    }

    for _, key := range keys {
        if strings.Contains(key, "*") {
            redis.ScanAndDelete(ctx, key)  // 模式匹配删除
        } else {
            redis.Del(ctx, key)
        }
    }

    localCache.Delete(fmt.Sprintf("price:base:%d", event.SKUID))
    return nil
}
```

### 8.3 幂等性保障

#### 8.3.1 唯一索引保证幂等

```go
// task_code 唯一索引保证同一上架操作不会重复创建
func CreateTask(req *CreateTaskRequest) (*ListingTask, error) {
    taskCode := generateTaskCode(req.CategoryID, req.CreatedBy, time.Now())

    err := db.Create(&ListingTask{TaskCode: taskCode, ...})
    if isDuplicateKeyError(err) {
        return db.GetByTaskCode(taskCode)  // 幂等返回已存在任务
    }
    return task, err
}
```

#### 8.3.2 Kafka 消费幂等

```go
func (h *Handler) Handle(ctx context.Context, msg *kafka.Message) error {
    eventID := string(msg.Key)

    // Redis 幂等键（24小时过期）
    idempotentKey := fmt.Sprintf("event:idempotent:%s", eventID)
    set, err := redis.SetNX(ctx, idempotentKey, "1", 24*time.Hour).Result()
    if !set {
        return nil  // 重复消息，跳过
    }

    // 处理消息...
}
```

---

## 九、监控与运维

### 9.1 核心监控指标

#### 9.1.1 商品上架监控

```
# 上架成功率
listing_task_total{type="single|batch|supplier", status="success|fail"}

# 上架时长
listing_task_duration_seconds{stage="audit|publish"}

# Worker 队列积压
listing_worker_queue_size{worker="audit|publish|parse"}

# 供应商同步延迟
listing_supplier_sync_lag_seconds{category, supplier_id}
```

#### 9.1.2 库存监控

```
# 操作计数
inventory_operation_total{op="book|sell|refund", mgmt="self|supplier", status="ok|fail"}

# 操作延迟
inventory_operation_duration_seconds{op="book|sell"}

# 库存差异
inventory_reconcile_diff{item_id, sku_id}

# 缺货次数
inventory_out_of_stock_total{item_id}

# 超卖告警（严重）
inventory_oversell_total{item_id}  # 目标: = 0
```

#### 9.1.3 价格监控

```
# 价格计算 QPS 和延迟
pricing_calculate_total{scene="list|detail|cart|checkout", status="success|fail|degrade"}
pricing_calculate_duration_seconds{scene, quantile="0.5|0.9|0.99"}

# 营销活动命中
pricing_promotion_match_total{activity_type, result="hit|miss"}
pricing_promotion_quota_remaining{activity_id}

# 缓存命中率
pricing_cache_hit_rate{cache_level="l1|l2", data_type}

# 降级计数
pricing_degrade_total{level="promotion|fee|voucher"}

# 价格异常波动
pricing_price_variance{sku_id}
```

### 9.2 告警规则

| 告警名称 | 条件 | 级别 | 处理 |
|---------|------|------|------|
| **超卖告警** | inventory_oversell > 0 | P0 | 紧急介入，人工补发 |
| **库存严重差异** | reconcile_diff > 1000 | P0 | 检查 Kafka 消费 + Redis/MySQL 状态 |
| **价格计算失败率** | pricing_fail_rate > 1% 持续3分钟 | P0 | 检查依赖服务（促销/优惠券） |
| **上架失败率** | listing_fail_rate > 5% 持续5分钟 | P1 | 检查 Worker 状态、DB 连接 |
| **供应商同步延迟** | sync_lag > 10min | P1 | 检查供应商 API 可用性 |
| **价格降级率** | degrade_rate > 5% 持续3分钟 | P1 | 检查促销/优惠券服务 |
| **缓存命中率下降** | cache_hit_rate < 70% 持续5分钟 | P2 | 检查 Redis 集群状态 |

### 9.3 日志与审计

#### 9.3.1 价格审计工具

```go
// 根据订单还原价格计算（客服工具）
func (s *AuditService) ReconstructPriceCalculation(ctx context.Context, orderID int64) (*PriceAuditResult, error) {
    // 1. 获取价格快照
    snapshot, _ := snapshotRepo.GetByOrderID(ctx, orderID)

    // 2. 验证计算正确性
    calculatedFinalPrice := snapshot.Subtotal.
        Sub(snapshot.PromotionDiscount).
        Add(snapshot.TotalFee).
        Sub(snapshot.VoucherDiscount)

    isValid := calculatedFinalPrice.Equal(snapshot.FinalPrice)

    return &PriceAuditResult{
        SnapshotCode:     snapshot.SnapshotCode,
        BasePrice:        snapshot.BasePrice,
        PromotionDetails: parseJSON(snapshot.PromotionDetails),
        FeeDetails:       parseJSON(snapshot.FeeDetails),
        VoucherDetails:   parseJSON(snapshot.VoucherDetails),
        FinalPrice:       snapshot.FinalPrice,
        PriceFormula:     snapshot.PriceFormula,  // 人类可读公式
        IsValid:          isValid,
    }, nil
}
```

#### 9.3.2 关键日志记录

```
# 商品上架日志
listing.task.created     {task_code, category, source_type}
listing.task.audited     {task_code, audit_result, auditor}
listing.task.published   {task_code, item_id, duration_ms}

# 库存操作日志
inventory.book.success   {order_id, item_id, quantity, stock_after}
inventory.book.failed    {order_id, reason, stock_available}
inventory.reconcile      {item_id, redis_stock, mysql_stock, diff}

# 价格计算日志
pricing.calculate        {snapshot_code, base, promotion, fee, voucher, final}
pricing.degrade          {level, reason, fallback_price}
pricing.price_changed    {sku_id, old_price, new_price, change_reason}
```

### 9.4 运维工具

#### 9.4.1 数据对账工具

```go
// 库存对账工具（每日定时任务）
func InventoryReconcileJob() {
    for _, cfg := range queryAllSelfManagedConfigs() {
        redisStock := getRedisAvailable(cfg.ItemID, cfg.SKUID)
        mysqlStock := getMySQLAvailable(cfg.ItemID, cfg.SKUID)
        diff := abs(redisStock - mysqlStock)

        if diff > 100 || diff > mysqlStock*0.1 {
            alert("库存差异: item=%d, redis=%d, mysql=%d, diff=%d",
                cfg.ItemID, redisStock, mysqlStock, diff)
            
            // 自动修复（以 MySQL 为准）
            if cfg.AutoReconcile {
                syncRedisFromMySQL(cfg.ItemID, cfg.SKUID)
            }
        }
    }
}
```

#### 9.4.2 缓存预热工具

```go
// 秒杀活动前缓存预热
func PreheatForFlashSale(activityID int64) error {
    activity := getActivity(activityID)
    
    // 1. 预热商品基础信息
    items := getItemsByActivity(activityID)
    for _, item := range items {
        cacheItem(item)
    }
    
    // 2. 预热价格
    for _, item := range items {
        priceResp := pricingEngine.Calculate(ctx, &PriceRequest{
            ItemID: item.ID, SKUID: item.DefaultSKUID,
        })
        cachePrice(item.ID, priceResp)
    }
    
    // 3. 预热库存（券码制提前加载到 Redis）
    for _, item := range items {
        if isCodeBased(item) {
            replenishCodesForFlashSale(item.ID, 10000)  // 预热1万张券码
        }
    }
    
    return nil
}
```

---

## 十、性能优化

### 10.1 秒杀场景优化

#### 10.1.1 券码预热

```go
// 活动前 1 小时预热券码到 Redis
func PreheatCodes(itemID, skuID, batchID int64, count int) error {
    // 从 MySQL 查询可用券码
    codes := db.Query(`
        SELECT id FROM inventory_code_pool_xx
        WHERE item_id=? AND sku_id=? AND batch_id=? AND status=1
        ORDER BY id LIMIT ?
    `, itemID, skuID, batchID, count)

    // 批量写入 Redis LIST
    stockKey := fmt.Sprintf("inventory:code:pool:%d:%d:%d", itemID, skuID, batchID)
    redis.RPush(ctx, stockKey, codes...)

    log.Infof("Preheated %d codes for item %d", len(codes), itemID)
    return nil
}
```

#### 10.1.2 热点 Key 分散

```
问题: 单个爆款商品，所有请求打到同一个 Redis Key
解决方案:
1. 本地缓存 (Caffeine) 拦截 80% 读请求
2. Key 分散: stock:item_123:0 ~ stock:item_123:9
   • 读请求随机路由
   • 写请求同步更新所有副本
3. 限流前置: 网关层按 item_id 限流
```

#### 10.1.3 批量写入优化

```go
// Kafka 消费批量写入 MySQL
func BatchInsertInventoryLog(logs []*InventoryOperationLog) error {
    // 攒批 100 条，批量 INSERT
    sql := `INSERT INTO inventory_operation_log 
            (item_id, operation_type, quantity, ...)
            VALUES (?, ?, ?), (?, ?, ?), ...`  // 100 rows
    
    db.Exec(sql, flattenParams(logs)...)
    
    // TPS 从 5k 提升至 80k
}
```

### 10.2 价格计算优化

#### 10.2.1 并行计算

```go
func (e *PricingEngine) parallelCalculate(ctx context.Context, req *PriceRequest) {
    var wg sync.WaitGroup
    wg.Add(3)

    go func() {
        defer wg.Done()
        basePrice = e.basePriceCalculator.Calculate(ctx, req)
    }()

    go func() {
        defer wg.Done()
        promotions = e.promotionMatcher.Match(ctx, req)
    }()

    go func() {
        defer wg.Done()
        fees = e.feeCalculator.Calculate(ctx, req)
    }()

    wg.Wait()
    // 基础价格、促销、费用并行计算，延迟降低 40%
}
```

#### 10.2.2 批量价格计算（列表页）

```go
func (e *PricingEngine) BatchCalculate(ctx context.Context, reqs []*PriceRequest) ([]*PriceResponse, error) {
    // 1. 批量获取 SKU 基础价格（一次 DB 查询）
    skuIDs := extractSKUIDs(reqs)
    basePrices := e.basePriceCalculator.BatchGet(ctx, skuIDs)

    // 2. 批量获取品类活动（按 category 分组）
    categoryIDs := extractCategoryIDs(reqs)
    promotionsByCategory := e.promotionMatcher.BatchLoadActive(ctx, categoryIDs)

    // 3. 并行计算每个商品（控制并发数20）
    results := make([]*PriceResponse, len(reqs))
    sem := make(chan struct{}, 20)
    
    for i, req := range reqs {
        sem <- struct{}{}
        go func(idx int, r *PriceRequest) {
            defer func() { <-sem }()
            results[idx] = e.calculateSingle(ctx, r, basePrices, promotionsByCategory)
        }(i, req)
    }
    
    return results
}
```

---

## 十一、新品类接入指南

### 11.1 接入检查清单

| 检查项 | 商品管理 | 库存管理 | 价格管理 |
|--------|----------|----------|----------|
| **数据模型** | 确定品类属性和 SKU 结构 | 确定 (ManagementType, UnitType) | 确定基础价格来源 |
| **流程配置** | 配置审核策略（运营/商家/供应商） | 配置扣减时机（下单/支付） | 配置费用类型 |
| **供应商对接** | 配置 Push/Pull 策略 | 配置同步策略（实时/定时） | 配置动态定价规则 |
| **校验规则** | 注册品类校验规则 | - | - |
| **自定义逻辑** | 实现 ValidationRule 接口 | 实现 Strategy 接口（如需） | 实现 BasePriceCalculator（如需） |

### 11.2 四步接入示例：演唱会门票

#### Step 1: 商品管理配置

```sql
-- 1. 配置审核策略
INSERT INTO listing_audit_config_tab (category_id, source_type, audit_strategy)
VALUES 
  (50001, 'supplier_push', 'fast_track'),  -- 供应商推送：快速通道
  (50001, 'operator_form', 'skip'),        -- 运营上传：免审核
  (50001, 'merchant_portal', 'manual');    -- 商家上传：人工审核
```

```go
// 2. 注册校验规则
engine.RegisterRule("concert", &ConcertValidationRule{
    // 演唱会特殊校验：场次时间、座位区域、票价范围
})
```

#### Step 2: 库存管理配置

```sql
-- 1. 配置库存策略
INSERT INTO inventory_config (item_id, management_type, unit_type, deduct_timing, supplier_id, sync_strategy)
VALUES (900001, 2, 3, 2, 700001, 2);
-- 供应商管理 + 时间维度 + 支付扣减 + 实时同步
```

```go
// 2. 调用统一接口（无需修改核心代码）
inventoryManager.BookStock(ctx, &BookStockReq{
    ItemID:   900001,
    SKUID:    900001001,
    Quantity: 2,
    OrderID:  orderID,
    Context:  map[string]interface{}{"session_id": "202608150900"},
})
```

#### Step 3: 价格管理配置

```sql
-- 1. 配置基础价格（按场次 × 区域）
INSERT INTO sku_tab (item_id, sku_code, sku_name, price)
VALUES (900001, 'SKU_CONCERT_VIP', 'VIP区', 2800.00);

-- 2. 配置费用
INSERT INTO fee_config_tab (fee_code, fee_name, fee_type, category_id, calculation_type, calculation_config)
VALUES 
  ('FEE_PLATFORM_CONCERT', '平台手续费', 'platform_fee', 50001, 'percentage', '{"percentage": 3}'),
  ('FEE_TICKET_SERVICE', '票务服务费', 'service_fee', 50001, 'fixed', '{"amount": 15}');

-- 3. 关联营销活动（在活动 category_ids 中加入 50001）
UPDATE promotion_activity_tab 
SET category_ids = JSON_ARRAY_APPEND(category_ids, '$', 50001)
WHERE activity_code = 'PROMO_SUMMER_SALE';
```

#### Step 4: 验证接入

```go
// 完整流程测试
func TestConcertTicketFlow(t *testing.T) {
    // 1. 创建演唱会商品
    task := listingService.CreateTask(&CreateTaskRequest{
        CategoryID: 50001,
        SourceType: "operator_form",
        ItemData: map[string]interface{}{
            "title": "周杰伦2026演唱会",
            "venue": "鸟巢",
            "date": "2026-08-15",
            "session_time": "19:00",
        },
    })
    assert.Equal(t, StatusOnline, task.Status)

    // 2. 查询价格
    priceResp := pricingEngine.Calculate(ctx, &PriceRequest{
        ItemID: task.ItemID,
        SKUID: 900001001,
        Quantity: 2,
    })
    assert.Equal(t, decimal.NewFromFloat(5630.00), priceResp.FinalPrice)
    // 2800 × 2 + 3% 平台手续费 + 15 × 2 服务费 = 5630

    // 3. 预订库存
    bookResp := inventoryManager.BookStock(ctx, &BookStockReq{
        ItemID: task.ItemID,
        SKUID: 900001001,
        Quantity: 2,
    })
    assert.True(t, bookResp.Success)

    // 4. 支付 → 库存售出
    inventoryManager.SellStock(ctx, &SellStockReq{
        ItemID: task.ItemID,
        SKUID: 900001001,
        Quantity: 2,
        OrderID: 123456,
    })
}
```

---

## 十二、设计总结

### 12.1 核心设计决策

| 维度 | 决策 | 原因 |
|------|------|------|
| **统一 vs 独立** | 统一模型 + 策略模式 | 复用逻辑，新品类零代码接入 |
| **同步 vs 异步** | 热路径同步，冷路径异步 | 高性能 + 可靠性 |
| **缓存策略** | L1 本地 + L2 Redis + L3 MySQL | 多级缓存，高性能 + 可靠 |
| **一致性保障** | Saga + 本地消息表 + 定时对账 | 最终一致性 |
| **降级策略** | 5级降级，基础价不可降级 | 保证核心链路可用 |
| **精度处理** | decimal.Decimal + 币种精度表 | 避免浮点误差 |
| **并发控制** | 乐观锁 + Redis Lua 原子操作 | 轻量级，无分布式锁 |
| **审计追溯** | 价格快照 + 状态历史 + 变更日志 | 完整记录，客诉可还原 |

### 12.2 关键技术栈

| 组件 | 技术选型 | 说明 |
|------|----------|------|
| Listing Service | Go 微服务 + Kafka + Worker Pool | 商品上架异步化 |
| Inventory Service | Go + Redis Cluster + MySQL | 库存管理策略模式 |
| Pricing Service | Go + Redis + decimal.Decimal | 价格计算引擎 |
| 缓存层 | Redis Cluster + sync.Map | 多级缓存 |
| 消息队列 | Kafka | 事件驱动 + 异步持久化 |
| 数据库 | MySQL (分库分表) | 持久化存储 |
| 搜索 | Elasticsearch | 商品搜索 |
| 监控 | Prometheus + Grafana | 性能监控与告警 |
| 配置中心 | Apollo / Nacos | 降级开关、超时配置 |

### 12.3 业务规模与性能

| 指标 | 数值 | 说明 |
|------|------|------|
| **秒杀峰值 QPS** | 20,000 | 单个爆款商品 |
| **日均订单量** | 2,000,000 | 支付成功订单 |
| **价格计算 P99** | < 100ms | 含促销/费用/优惠券 |
| **库存预订 P99** | < 50ms | Redis 原子操作 |
| **上架成功率** | > 95% | 全品类平均 |
| **库存准确率** | > 99.9% | 对账后修复 |
| **缓存命中率** | > 85% | L1 + L2 |
| **超卖次数** | 0 | 严格监控 |

### 12.4 成本分析

| 资源 | 配置 | 数量 | 月成本 |
|------|------|------|--------|
| Redis Cluster | 32GB × 6 节点 | 1 套 | $800 |
| MySQL | 64GB 主库 + 32GB × 2 从库 | 1 套 | $1,200 |
| 应用服务 | 4C8G Pod | 30 台 | $1,800 |
| Kafka | 8C16G Broker | 3 台 | $900 |
| Elasticsearch | 8C32G 节点 | 3 台 | $1,200 |
| **总计** | - | - | **$5,900/月** |

**日均订单成本**：$5,900 / 2,000,000 = **$0.00295/单**（0.3 美分）

---

## 十三、未来改进路线

```
┌─────────────────────────────────────────────────────────────────┐
│                  改进路线图（2026-2027）                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Q2 2026: 基础完善（当前）                                        │
│  ✅ 统一商品·库存·价格管理中心上线                                 │
│  ✅ 多品类策略模式 + 异步化流程                                    │
│  ✅ 价格快照 + 多级缓存 + 5级降级                                  │
│                                                                  │
│  Q3 2026: 智能化初步                                              │
│  🔲 动态定价规则增强（需求预测 + 竞品监控）                         │
│  🔲 A/B 测试平台集成（价格/促销策略）                               │
│  🔲 用户分层定价（新用户/VIP/高净值）                               │
│  🔲 智能库存预警（缺货预测 + 自动补货）                             │
│                                                                  │
│  Q4 2026: ML 模型引入                                             │
│  🔲 价格弹性分析（ML 模型）                                        │
│  🔲 需求预测模型（LSTM）                                           │
│  🔲 促销效果评估（ROI 计算）                                        │
│  🔲 库存优化模型（安全库存 + 补货策略）                             │
│                                                                  │
│  Q1 2027: 实时智能                                                │
│  🔲 实时特征平台（Flink）                                          │
│  🔲 在线模型更新（自适应定价）                                      │
│  🔲 多目标优化（收入/转化/毛利平衡）                                │
│  🔲 实时库存调拨（跨区域优化）                                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 附录：相关文档

1. [多品类统一商品上架系统设计](./19-listing-upload-system-design.md)
2. [多品类统一库存系统设计](./18-inventory-system-design.md)
3. [多品类统一价格管理与计价系统设计](./21-pricing.md)
