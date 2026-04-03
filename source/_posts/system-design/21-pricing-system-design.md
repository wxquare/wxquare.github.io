---
title: 多品类统一价格管理与计价系统设计：电商·虚拟商品·本地生活
date: 2026-01-15
categories:
- 系统设计
tags:
- 价格系统
- 电商
- 系统设计
- 高并发
toc: true
---

<!-- toc -->

## 一、背景与挑战

### 1.1 多品类价格差异

在数字电商/本地生活平台中，不同品类的定价逻辑差异极大：

| 品类 | 价格特点 | 定价维度 | 费用组成 | 典型示例 |
|------|----------|----------|----------|----------|
| **酒店 (Hotel)** | 时间维度定价，每日价格独立 | 日期 × 房型 × 早餐 | 房价 + DP Fee + Hub Fee | 曼谷万豪豪华房 |
| **电影票 (Movie)** | 场次 × 座位 × 票种定价 | 场次 × 座位区 × 票种 | 票价 + DP Fee + 选座费 | 阿凡达3 IMAX 成人票 |
| **话费充值 (TopUp)** | 面额定价，无 SKU 变体 | 运营商 × 面额 | 面额 + 手续费 - 补贴 | AIS 100฿ 充值 |
| **电子券 (E-voucher)** | 面值 vs 售价差异 | 品牌 × 面值 | 面值 - 平台折扣 + DP Fee | 星巴克 500฿ 电子券 |
| **礼品卡 (Giftcard)** | 面值定价 + 平台折扣 | 品牌 × 面值 | 面值 - 折扣 + DP Fee | Google Play 充值卡 |
| **本地生活套餐** | 组合定价，子项加总 | 套餐 × 子项 × 份数 | 套餐价 + 服务费 | 海底捞双人套餐 |

### 1.2 核心痛点

1. **价格散落多表**：基础价、营销价、费用、优惠券分散在不同模块，缺乏统一视图。
2. **计算逻辑分散**：各品类各自实现价格计算，重复代码多，难以维护。
3. **营销活动隔离**：促销规则硬编码在业务逻辑中，扩展性差。
4. **Fee 管理混乱**：DP Fee、Hub Fee、Carrier Fee 等多种费用缺乏统一配置。
5. **优惠券叠加复杂**：Voucher、Promo Code、积分抵扣等多种优惠方式叠加规则不清晰。
6. **审计困难**：价格变更历史难以追溯，用户投诉时无法准确还原计算过程。
7. **精度与币种问题**：多币种场景下精度不一致，浮点运算导致分摊误差。

### 1.3 设计目标

| 目标 | 说明 | 优先级 |
|------|------|--------|
| **统一价格中心** | 所有价格计算通过统一的 Pricing Engine 进行 | P0 |
| **分层价格模型** | 基础价 → 营销价 → 费用 → 优惠券 → 最终价，层次清晰 | P0 |
| **规则引擎化** | 营销活动、费用规则可配置，支持动态调整 | P0 |
| **价格快照** | 每笔订单保留完整价格计算明细，支持审计和追溯 | P0 |
| **高性能** | P99 < 100ms，支持万级 QPS 并发价格计算 | P1 |
| **多级降级** | 促销/优惠券服务不可用时，仍能返回基础价格 | P1 |
| **可扩展** | 新增营销活动类型、费用类型、优惠券类型无需改动核心架构 | P1 |

---

## 二、整体架构

### 2.1 四层价格架构

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
│  │  • voucher_usage_log         (使用记录)           │               │
│  └──────────────────────────────────────────────────┘               │
│                          ↓                                           │
│  ═══════════════════════════════════════════════════                 │
│  最终价格 = Base - Promotion + Fee - Voucher                         │
│  ═══════════════════════════════════════════════════                 │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 分层服务架构

```
┌──────────────────────────────────────────────────────────────────┐
│                        API Gateway                                │
├──────────────────────────────────────────────────────────────────┤
│                                                                    │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                   Pricing Engine Service                   │    │
│  │  ┌────────────┐ ┌─────────────┐ ┌──────────────────┐    │    │
│  │  │ Price API   │ │ Snapshot API│ │  Audit API        │    │    │
│  │  └────────────┘ └─────────────┘ └──────────────────┘    │    │
│  │                                                            │    │
│  │  ┌────────────────────────────────────────────────────┐  │    │
│  │  │              Price Calculator Pipeline              │  │    │
│  │  │  Base → Promotion → Fee → Voucher → Final          │  │    │
│  │  └────────────────────────────────────────────────────┘  │    │
│  └──────────────────────────────────────────────────────────┘    │
│                                                                    │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐  │
│  │  Promotion  │ │    Fee     │ │  Voucher   │ │  Snapshot  │  │
│  │  Service    │ │  Service   │ │  Service   │ │  Service   │  │
│  └────────────┘ └────────────┘ └────────────┘ └────────────┘  │
│                                                                    │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │  Infrastructure: Redis | MySQL | Kafka | Prometheus       │    │
│  └──────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────┘
```

### 2.3 价格组件定义

| 组件 | 说明 | 计算方式 | 示例 |
|------|------|----------|------|
| **Base Price** | SKU 基础售价 | 固定 / 时间动态 | 480฿（电影票） |
| **Promotion Discount** | 营销活动折扣 | 百分比 / 固定金额 / 满减 | -50฿（新用户立减） |
| **DP Fee** | DP 平台手续费 | 固定 / 百分比 | +10฿ |
| **Hub Fee** | Hub 商户服务费 | 固定 / 百分比 / 阶梯 | +5฿ |
| **Service Fee** | 附加服务费 | 固定 | +20฿（选座费） |
| **Tax** | 税费 | 百分比 | +7%（VAT） |
| **Voucher Discount** | 优惠券抵扣 | 固定 / 百分比 / 封顶 | -30฿ |
| **Final Price** | 最终支付价格 | Base - Promo + Fee - Voucher | 435฿ |

---

## 三、数据模型

### 3.1 基础价格表（继承商品模型）

**sku_tab**（SKU 基础价格，已在商品模型中定义）：

```sql
-- 核心价格字段:
--   price          DECIMAL(20,2)  -- SKU 基础价格
--   original_price DECIMAL(20,2)  -- 原价（划线价）
--   cost_price     DECIMAL(20,2)  -- 成本价
--   currency       VARCHAR(3)     -- 币种 (THB, VND, IDR, MYR, SGD, PHP)
```

**dynamic_pricing_rule_tab**（动态定价规则）：

```sql
CREATE TABLE `dynamic_pricing_rule_tab` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '规则ID',
    `rule_code` VARCHAR(64) NOT NULL COMMENT '规则编码',
    `rule_name` VARCHAR(255) NOT NULL COMMENT '规则名称',
    `category_id` BIGINT COMMENT '类目ID (NULL表示全品类)',
    `rule_type` VARCHAR(50) NOT NULL COMMENT '规则类型: demand_based, inventory_based, time_based, competitor_based',
    `trigger_condition` JSON NOT NULL COMMENT '触发条件: {"inventory_threshold": 10, "time_window": "18:00-22:00"}',
    `adjustment_type` VARCHAR(20) NOT NULL COMMENT '调整类型: percentage, fixed_amount',
    `adjustment_value` DECIMAL(10,2) NOT NULL COMMENT '调整值: 10表示加价10%或加价10元',
    `min_price` DECIMAL(20,2) COMMENT '最低价格限制（价格地板）',
    `max_price` DECIMAL(20,2) COMMENT '最高价格限制（价格天花板）',
    `priority` INT DEFAULT 0 COMMENT '优先级（数字越大越优先）',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1=启用, 0=禁用',
    `effective_start` DATETIME COMMENT '生效开始时间',
    `effective_end` DATETIME COMMENT '生效结束时间',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_rule_code` (`rule_code`),
    KEY `idx_category` (`category_id`),
    KEY `idx_status_time` (`status`, `effective_start`, `effective_end`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='动态定价规则表';
```

### 3.2 营销活动表

**promotion_activity_tab**（营销活动主表）：

```sql
CREATE TABLE `promotion_activity_tab` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '活动ID',
    `activity_code` VARCHAR(64) NOT NULL COMMENT '活动编码',
    `activity_name` VARCHAR(255) NOT NULL COMMENT '活动名称',
    `activity_type` VARCHAR(50) NOT NULL COMMENT '活动类型: discount, full_reduction, bundle, flash_sale, first_order, new_user',
    `category_ids` JSON COMMENT '适用类目: [10001, 30001] (NULL表示全品类)',
    `item_ids` JSON COMMENT '指定商品ID: [100001, 200001] (NULL表示不限)',
    `sku_ids` JSON COMMENT '指定SKU ID: [1000001, 2000001]',
    `user_type` VARCHAR(50) DEFAULT 'all' COMMENT '用户类型: all, new, vip, specific',
    `user_ids` JSON COMMENT '指定用户ID (user_type=specific时)',
    `discount_type` VARCHAR(50) NOT NULL COMMENT '折扣类型: percentage, fixed_amount, full_reduction, buy_n_get_m, tiered_discount',
    `discount_value` JSON NOT NULL COMMENT '折扣配置（见下方说明）',
    `max_discount_amount` DECIMAL(20,2) COMMENT '最大折扣金额上限',
    `min_purchase_amount` DECIMAL(20,2) COMMENT '最低购买金额',
    `min_purchase_quantity` INT COMMENT '最低购买数量',
    `priority` INT DEFAULT 0 COMMENT '优先级（数字越大越优先）',
    `exclusivity` TINYINT DEFAULT 0 COMMENT '互斥性: 0=可叠加, 1=与其他活动互斥',
    `voucher_compatible` TINYINT DEFAULT 1 COMMENT '优惠券兼容: 0=与优惠券互斥, 1=可叠加',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1=启用, 0=禁用',
    `start_time` DATETIME NOT NULL COMMENT '开始时间',
    `end_time` DATETIME NOT NULL COMMENT '结束时间',
    `daily_quota` INT COMMENT '每日配额',
    `total_quota` INT COMMENT '总配额',
    `used_quota` INT DEFAULT 0 COMMENT '已使用配额',
    `per_user_limit` INT DEFAULT 0 COMMENT '每用户限享次数 (0=不限)',
    `description` TEXT COMMENT '活动描述',
    `created_by` BIGINT COMMENT '创建人',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_activity_code` (`activity_code`),
    KEY `idx_type_status` (`activity_type`, `status`),
    KEY `idx_time` (`start_time`, `end_time`),
    KEY `idx_priority` (`priority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='营销活动主表';
```

**discount_value** JSON 配置说明：

| 折扣类型 | JSON 配置示例 | 说明 |
|---------|--------------|------|
| percentage | `{"percentage": 20}` | 打8折（减20%） |
| fixed_amount | `{"amount": 50}` | 立减50฿ |
| full_reduction | `{"threshold": 3000, "discount": 200}` | 满3000减200 |
| buy_n_get_m | `{"buy": 3, "free": 1}` | 买3送1 |
| tiered_discount | `{"tiers": [{"threshold": 500, "percentage": 5}, {"threshold": 200, "percentage": 3}]}` | 阶梯折扣 |

**promotion_usage_log_tab**（活动使用记录）：

```sql
CREATE TABLE `promotion_usage_log_tab` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `activity_id` BIGINT NOT NULL COMMENT '活动ID',
    `user_id` BIGINT NOT NULL COMMENT '用户ID',
    `order_id` BIGINT NOT NULL COMMENT '订单ID',
    `item_id` BIGINT NOT NULL COMMENT '商品ID',
    `sku_id` BIGINT NOT NULL COMMENT 'SKU ID',
    `original_price` DECIMAL(20,2) NOT NULL COMMENT '原价',
    `discount_amount` DECIMAL(20,2) NOT NULL COMMENT '折扣金额',
    `final_price` DECIMAL(20,2) NOT NULL COMMENT '最终价格',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_activity` (`activity_id`),
    KEY `idx_user` (`user_id`),
    KEY `idx_order` (`order_id`),
    KEY `idx_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='营销活动使用记录表';
```

### 3.3 费用配置表

**fee_config_tab**（费用配置表）：

```sql
CREATE TABLE `fee_config_tab` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '费用配置ID',
    `fee_code` VARCHAR(64) NOT NULL COMMENT '费用编码',
    `fee_name` VARCHAR(255) NOT NULL COMMENT '费用名称',
    `fee_type` VARCHAR(50) NOT NULL COMMENT '费用类型: dp_fee, hub_fee, service_fee, carrier_fee, seat_fee, tax',
    `category_id` BIGINT COMMENT '类目ID (NULL表示全品类)',
    `item_id` BIGINT COMMENT '商品ID (NULL表示不限)',
    `sku_id` BIGINT COMMENT 'SKU ID (NULL表示不限)',
    `entity_type` VARCHAR(50) COMMENT '关联实体类型: carrier, cinema, merchant',
    `entity_id` BIGINT COMMENT '关联实体ID',
    `region` VARCHAR(10) COMMENT '国家/地区: TH, VN, ID, MY, SG, PH (NULL表示全地区)',
    `calculation_type` VARCHAR(50) NOT NULL COMMENT '计算方式: fixed, percentage, tiered',
    `calculation_config` JSON NOT NULL COMMENT '计算配置（见下方说明）',
    `min_fee` DECIMAL(20,2) COMMENT '最小费用',
    `max_fee` DECIMAL(20,2) COMMENT '最大费用',
    `display_type` VARCHAR(50) DEFAULT 'separate' COMMENT '展示方式: separate(单独展示), included(包含在总价)',
    `can_be_discounted` TINYINT DEFAULT 0 COMMENT '是否可被优惠券抵扣: 0=否, 1=是',
    `priority` INT DEFAULT 0 COMMENT '计算优先级',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1=启用, 0=禁用',
    `effective_start` DATETIME COMMENT '生效开始时间',
    `effective_end` DATETIME COMMENT '生效结束时间',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_fee_code` (`fee_code`),
    KEY `idx_type_category` (`fee_type`, `category_id`),
    KEY `idx_entity` (`entity_type`, `entity_id`),
    KEY `idx_region` (`region`),
    KEY `idx_status_time` (`status`, `effective_start`, `effective_end`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='费用配置表';
```

**calculation_config** JSON 配置说明：

| 计算方式 | JSON 配置示例 | 说明 |
|---------|--------------|------|
| fixed | `{"amount": 10}` | 固定10฿ |
| percentage | `{"percentage": 2.5}` | 按基础价的2.5% |
| tiered | `{"tiers": [{"threshold": 5000, "fee": 150}, {"threshold": 3000, "fee": 100}, {"threshold": 0, "fee": 50}]}` | 金额阶梯 |

### 3.4 优惠券表

**voucher_tab**（优惠券主表）：

```sql
CREATE TABLE `voucher_tab` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '优惠券ID',
    `voucher_code` VARCHAR(64) NOT NULL COMMENT '优惠券编码',
    `voucher_name` VARCHAR(255) NOT NULL COMMENT '优惠券名称',
    `voucher_type` VARCHAR(50) NOT NULL COMMENT '券类型: discount, cashback, gift, shipping_free',
    `discount_type` VARCHAR(50) NOT NULL COMMENT '折扣类型: percentage, fixed_amount, full_reduction',
    `discount_value` JSON NOT NULL COMMENT '折扣配置',
    `max_discount_amount` DECIMAL(20,2) COMMENT '最大折扣金额',
    `min_purchase_amount` DECIMAL(20,2) COMMENT '最低消费金额',
    `category_ids` JSON COMMENT '适用类目',
    `item_ids` JSON COMMENT '适用商品',
    `exclude_item_ids` JSON COMMENT '排除商品',
    `stackable_with_promotion` TINYINT DEFAULT 0 COMMENT '是否与促销叠加: 0=互斥, 1=可叠加',
    `stackable_with_voucher` TINYINT DEFAULT 0 COMMENT '是否与其他券叠加: 0=互斥, 1=可叠加',
    `total_quantity` INT NOT NULL COMMENT '发行总量',
    `claimed_quantity` INT DEFAULT 0 COMMENT '已领取数量',
    `used_quantity` INT DEFAULT 0 COMMENT '已使用数量',
    `per_user_limit` INT DEFAULT 1 COMMENT '每用户限领数量',
    `valid_days` INT COMMENT '领取后有效天数 (NULL表示固定有效期)',
    `valid_start` DATETIME COMMENT '固定有效期开始',
    `valid_end` DATETIME COMMENT '固定有效期结束',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1=可用, 0=禁用',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_voucher_code` (`voucher_code`),
    KEY `idx_type_status` (`voucher_type`, `status`),
    KEY `idx_valid_time` (`valid_start`, `valid_end`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券主表';
```

**user_voucher_tab**（用户优惠券表）：

```sql
CREATE TABLE `user_voucher_tab` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '用户券ID',
    `voucher_id` BIGINT NOT NULL COMMENT '优惠券ID',
    `user_id` BIGINT NOT NULL COMMENT '用户ID',
    `voucher_code` VARCHAR(64) NOT NULL COMMENT '券码',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1=未使用, 2=已使用, 3=已过期, 4=已退还',
    `claimed_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '领取时间',
    `valid_start` DATETIME NOT NULL COMMENT '有效期开始',
    `valid_end` DATETIME NOT NULL COMMENT '有效期结束',
    `used_at` TIMESTAMP NULL COMMENT '使用时间',
    `order_id` BIGINT COMMENT '关联订单ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_voucher_code` (`user_id`, `voucher_code`),
    KEY `idx_voucher` (`voucher_id`),
    KEY `idx_user_status` (`user_id`, `status`),
    KEY `idx_valid_time` (`valid_start`, `valid_end`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户优惠券表';
```

### 3.5 价格快照表

**price_snapshot_tab**（价格快照表）：

```sql
CREATE TABLE `price_snapshot_tab` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '快照ID',
    `snapshot_code` VARCHAR(64) NOT NULL COMMENT '快照编码',
    `order_id` BIGINT COMMENT '订单ID (下单后关联)',
    `user_id` BIGINT NOT NULL COMMENT '用户ID',
    `item_id` BIGINT NOT NULL COMMENT '商品ID',
    `sku_id` BIGINT NOT NULL COMMENT 'SKU ID',
    `quantity` INT NOT NULL DEFAULT 1 COMMENT '购买数量',
    `currency` VARCHAR(3) NOT NULL COMMENT '币种',
    
    -- 基础价格
    `base_price` DECIMAL(20,2) NOT NULL COMMENT '基础价格（单价）',
    `original_price` DECIMAL(20,2) COMMENT '原价（划线价）',
    `subtotal` DECIMAL(20,2) NOT NULL COMMENT '小计 (base_price * quantity)',
    
    -- 营销折扣
    `promotion_discount` DECIMAL(20,2) DEFAULT 0 COMMENT '营销折扣总额',
    `promotion_details` JSON COMMENT '促销明细: [{"activity_id": 123, "name": "...", "discount": 50}]',
    
    -- 费用
    `total_fee` DECIMAL(20,2) DEFAULT 0 COMMENT '费用总额',
    `fee_details` JSON COMMENT '费用明细: [{"fee_type": "dp_fee", "amount": 10}]',
    
    -- 优惠券
    `voucher_discount` DECIMAL(20,2) DEFAULT 0 COMMENT '优惠券抵扣',
    `voucher_details` JSON COMMENT '优惠券明细: [{"voucher_id": 456, "discount": 30}]',
    
    -- 最终价格
    `final_price` DECIMAL(20,2) NOT NULL COMMENT '最终价格',
    `price_formula` TEXT COMMENT '价格计算公式（人类可读）',
    
    -- 计算上下文
    `calculation_context` JSON COMMENT '计算上下文: {"engine_version": "v1.0", "region": "TH"}',
    `expired_at` DATETIME COMMENT '快照过期时间（加购场景30分钟有效）',
    
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_snapshot_code` (`snapshot_code`),
    KEY `idx_order` (`order_id`),
    KEY `idx_user_item` (`user_id`, `item_id`),
    KEY `idx_expired` (`expired_at`),
    KEY `idx_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='价格快照表';
```

### 3.6 价格变更日志

**price_change_log_tab**：

```sql
CREATE TABLE `price_change_log_tab` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `sku_id` BIGINT NOT NULL COMMENT 'SKU ID',
    `change_type` VARCHAR(50) NOT NULL COMMENT '变更类型: base_price, promotion, fee_config, dynamic_rule',
    `old_value` JSON COMMENT '旧值',
    `new_value` JSON NOT NULL COMMENT '新值',
    `change_reason` VARCHAR(255) COMMENT '变更原因',
    `changed_by` BIGINT COMMENT '变更人 (0=系统自动)',
    `changed_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_sku_time` (`sku_id`, `changed_at`),
    KEY `idx_type` (`change_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='价格变更日志';
```

---

## 四、价格计算引擎

### 4.1 核心数据结构

```go
// PricingEngine 价格计算引擎
type PricingEngine struct {
    basePriceCalculator BasePriceCalculator
    promotionMatcher    PromotionMatcher
    feeCalculator       FeeCalculator
    voucherApplier      VoucherApplier
    snapshotGenerator   SnapshotGenerator
    priceCache          cache.Cache
    degradeConfig       *DegradeConfig // 降级配置
}

// PriceRequest 价格计算请求
type PriceRequest struct {
    UserID       int64                  `json:"user_id"`
    ItemID       int64                  `json:"item_id"`
    SKUID        int64                  `json:"sku_id"`
    Quantity     int                    `json:"quantity"`
    CategoryID   int64                  `json:"category_id"`
    Region       string                 `json:"region"`       // 国家/地区: TH, VN, ID...
    Currency     string                 `json:"currency"`     // 币种: THB, VND, IDR...
    VoucherCodes []string               `json:"voucher_codes"`
    Context      map[string]interface{} `json:"context"`      // 额外上下文（日期、场次等）
    Scene        string                 `json:"scene"`        // 场景: list, detail, cart, checkout
}

// PriceResponse 价格计算响应
type PriceResponse struct {
    // 基础价格
    BasePrice     decimal.Decimal `json:"base_price"`
    OriginalPrice decimal.Decimal `json:"original_price"`
    Subtotal      decimal.Decimal `json:"subtotal"`

    // 营销折扣
    PromotionDiscount decimal.Decimal   `json:"promotion_discount"`
    PromotionDetails  []PromotionDetail `json:"promotion_details"`

    // 费用
    TotalFee   decimal.Decimal `json:"total_fee"`
    FeeDetails []FeeDetail     `json:"fee_details"`

    // 优惠券
    VoucherDiscount decimal.Decimal   `json:"voucher_discount"`
    VoucherDetails  []VoucherDetail   `json:"voucher_details"`

    // 最终价格
    FinalPrice   decimal.Decimal `json:"final_price"`
    Currency     string          `json:"currency"`
    PriceFormula string          `json:"price_formula"`

    // 快照
    SnapshotCode string `json:"snapshot_code"`

    // 降级标记
    Degraded     bool   `json:"degraded"`      // 是否降级
    DegradeLevel string `json:"degrade_level"` // 降级级别: none, promotion, voucher, fee
}

// PromotionDetail 促销明细
type PromotionDetail struct {
    ActivityID   int64           `json:"activity_id"`
    ActivityName string          `json:"activity_name"`
    ActivityType string          `json:"activity_type"`
    Discount     decimal.Decimal `json:"discount"`
}

// FeeDetail 费用明细
type FeeDetail struct {
    FeeType     string          `json:"fee_type"`
    FeeName     string          `json:"fee_name"`
    Amount      decimal.Decimal `json:"amount"`
    CanDiscount bool            `json:"can_discount"` // 是否可被优惠券抵扣
    DisplayType string          `json:"display_type"` // separate / included
}

// VoucherDetail 优惠券明细
type VoucherDetail struct {
    VoucherID   int64           `json:"voucher_id"`
    VoucherCode string          `json:"voucher_code"`
    VoucherName string          `json:"voucher_name"`
    Discount    decimal.Decimal `json:"discount"`
}
```

### 4.2 价格计算核心流程

```go
// Calculate 计算价格（核心入口）
func (e *PricingEngine) Calculate(ctx context.Context, req *PriceRequest) (*PriceResponse, error) {
    resp := &PriceResponse{Currency: req.Currency}

    // Step 1: 计算基础价格
    basePrice, err := e.basePriceCalculator.Calculate(ctx, req)
    if err != nil {
        return nil, errors.Wrap(err, "calculate base price failed")
    }
    resp.BasePrice = basePrice.Price
    resp.OriginalPrice = basePrice.OriginalPrice
    resp.Subtotal = basePrice.Price.Mul(decimal.NewFromInt(int64(req.Quantity)))

    // Step 2: 匹配并应用营销活动（支持降级）
    promotions, err := e.matchPromotionsWithDegrade(ctx, req, basePrice.Price)
    if err != nil {
        // 促销服务故障时降级，不影响核心流程
        log.Warnf("promotion match degraded: %v", err)
        resp.Degraded = true
        resp.DegradeLevel = "promotion"
    }

    totalDiscount := decimal.Zero
    for _, promo := range promotions {
        discount := e.calculatePromotionDiscount(promo, resp.Subtotal)
        totalDiscount = totalDiscount.Add(discount)
        resp.PromotionDetails = append(resp.PromotionDetails, PromotionDetail{
            ActivityID:   promo.ActivityID,
            ActivityName: promo.ActivityName,
            ActivityType: promo.ActivityType,
            Discount:     discount,
        })
    }
    resp.PromotionDiscount = totalDiscount

    // Step 3: 计算费用（支持降级）
    fees, err := e.calculateFeesWithDegrade(ctx, req, basePrice.Price)
    if err != nil {
        log.Warnf("fee calculation degraded: %v", err)
        resp.Degraded = true
        resp.DegradeLevel = "fee"
    }

    totalFee := decimal.Zero
    for _, fee := range fees {
        totalFee = totalFee.Add(fee.Amount)
        resp.FeeDetails = append(resp.FeeDetails, FeeDetail{
            FeeType:     fee.FeeType,
            FeeName:     fee.FeeName,
            Amount:      fee.Amount,
            CanDiscount: fee.CanDiscount,
            DisplayType: fee.DisplayType,
        })
    }
    resp.TotalFee = totalFee

    // Step 4: 应用优惠券（支持降级）
    if len(req.VoucherCodes) > 0 {
        voucherResult, err := e.applyVouchersWithDegrade(ctx, req, resp)
        if err != nil {
            log.Warnf("voucher apply degraded: %v", err)
            resp.Degraded = true
            resp.DegradeLevel = "voucher"
        } else {
            resp.VoucherDiscount = voucherResult.TotalDiscount
            resp.VoucherDetails = voucherResult.Details
        }
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

    // Step 6: 币种精度对齐（不同币种小数位不同）
    resp.FinalPrice = e.roundByCurrency(resp.FinalPrice, req.Currency)

    // Step 7: 生成价格公式（人类可读）
    resp.PriceFormula = e.buildPriceFormula(resp)

    // Step 8: 生成价格快照（异步写入，不阻塞主流程）
    go func() {
        snapshot, err := e.snapshotGenerator.Generate(context.Background(), req, resp)
        if err != nil {
            log.Errorf("generate snapshot failed: %v", err)
        } else {
            resp.SnapshotCode = snapshot.SnapshotCode
        }
    }()

    return resp, nil
}
```

### 4.3 币种精度处理

不同东南亚国家的币种精度差异很大，这是必须处理的核心问题：

| 币种 | 代码 | 小数位 | 最小单位 | 示例 |
|------|------|--------|----------|------|
| 泰铢 | THB | 2 | 0.01 | 480.50฿ |
| 越南盾 | VND | 0 | 1 | 120000₫ |
| 印尼盾 | IDR | 0 | 1 | 85000Rp |
| 马来西亚林吉特 | MYR | 2 | 0.01 | RM 25.90 |
| 新加坡元 | SGD | 2 | 0.01 | S$12.50 |
| 菲律宾比索 | PHP | 2 | 0.01 | ₱350.00 |

```go
// roundByCurrency 按币种精度四舍五入
func (e *PricingEngine) roundByCurrency(amount decimal.Decimal, currency string) decimal.Decimal {
    switch currency {
    case "VND", "IDR":
        // 越南盾、印尼盾无小数位，向上取整到最小货币单位
        return amount.Ceil()
    case "THB", "MYR", "SGD", "PHP":
        // 2位小数，Banker's Rounding（银行家舍入）
        return amount.Round(2)
    default:
        return amount.Round(2)
    }
}

// 多商品分摊时的精度处理（确保分摊总和 = 原始金额）
func spreadDiscount(totalDiscount decimal.Decimal, items []OrderItem) []decimal.Decimal {
    total := decimal.Zero
    for _, item := range items {
        total = total.Add(item.Subtotal)
    }

    results := make([]decimal.Decimal, len(items))
    allocated := decimal.Zero

    for i, item := range items {
        if i == len(items)-1 {
            // 最后一个商品承担所有剩余（消除分摊误差）
            results[i] = totalDiscount.Sub(allocated)
        } else {
            // 按比例分摊
            ratio := item.Subtotal.Div(total)
            results[i] = totalDiscount.Mul(ratio).Round(2)
            allocated = allocated.Add(results[i])
        }
    }
    return results
}
```

### 4.4 营销活动匹配器

```go
// PromotionMatcher 营销活动匹配器
type PromotionMatcher struct {
    promotionRepo repository.PromotionRepository
    userService   UserService   // 查询用户类型（新用户/VIP等）
    quotaService  QuotaService  // 配额管理（Redis 原子操作）
    cache         cache.Cache
}

// Match 匹配营销活动
func (m *PromotionMatcher) Match(ctx context.Context, req *PromotionMatchRequest) ([]*MatchedPromotion, error) {
    // 1. 获取所有生效的营销活动（优先从缓存）
    activities, err := m.loadActivePromotions(ctx, req.CategoryID)
    if err != nil {
        return nil, err
    }

    // 2. 按优先级排序（数字越大越优先）
    sort.Slice(activities, func(i, j int) bool {
        return activities[i].Priority > activities[j].Priority
    })

    var matched []*MatchedPromotion
    hasExclusive := false

    for _, activity := range activities {
        // 3. 检查生效条件
        if !m.checkConditions(ctx, activity, req) {
            continue
        }

        // 4. 互斥规则判断
        if activity.Exclusivity == 1 && len(matched) > 0 {
            continue // 当前活动互斥，且已有其他活动命中
        }
        if hasExclusive {
            continue // 已有互斥活动命中，跳过后续所有活动
        }

        // 5. 检查优惠券兼容性
        if req.HasVoucher && activity.VoucherCompatible == 0 {
            continue
        }

        // 6. 配额检查（Redis 原子操作）
        if activity.TotalQuota != nil {
            ok, err := m.quotaService.TryConsume(ctx, activity.ID, req.UserID)
            if err != nil || !ok {
                continue
            }
        }

        matched = append(matched, &MatchedPromotion{
            ActivityID:       activity.ID,
            ActivityName:     activity.ActivityName,
            ActivityType:     activity.ActivityType,
            DiscountType:     activity.DiscountType,
            DiscountValue:    activity.DiscountValue,
            MaxDiscountAmount: activity.MaxDiscountAmount,
        })

        // 7. 互斥活动命中后，停止遍历
        if activity.Exclusivity == 1 {
            hasExclusive = true
            break
        }
    }

    return matched, nil
}

// checkConditions 检查活动生效条件
func (m *PromotionMatcher) checkConditions(ctx context.Context, activity *model.PromotionActivity, req *PromotionMatchRequest) bool {
    now := time.Now()

    // 时间检查
    if now.Before(activity.StartTime) || now.After(activity.EndTime) {
        return false
    }

    // 类目检查
    if len(activity.CategoryIDs) > 0 && !contains(activity.CategoryIDs, req.CategoryID) {
        return false
    }

    // 商品/SKU 检查
    if len(activity.ItemIDs) > 0 && !contains(activity.ItemIDs, req.ItemID) {
        return false
    }
    if len(activity.SKUIDs) > 0 && !contains(activity.SKUIDs, req.SKUID) {
        return false
    }

    // 用户类型检查（调用 User Service）
    if activity.UserType != "all" {
        userType, _ := m.userService.GetUserType(ctx, req.UserID)
        if userType != activity.UserType {
            return false
        }
    }

    // 最低金额检查
    if activity.MinPurchaseAmount != nil {
        subtotal := req.BasePrice.Mul(decimal.NewFromInt(int64(req.Quantity)))
        if subtotal.LessThan(*activity.MinPurchaseAmount) {
            return false
        }
    }

    // 最低数量检查
    if activity.MinPurchaseQuantity != nil && req.Quantity < *activity.MinPurchaseQuantity {
        return false
    }

    return true
}
```

### 4.5 费用计算器

```go
// FeeCalculator 费用计算器
type FeeCalculator struct {
    feeRepo repository.FeeRepository
    cache   cache.Cache
}

// Calculate 计算费用
func (c *FeeCalculator) Calculate(ctx context.Context, req *FeeCalcRequest) ([]*CalculatedFee, error) {
    // 1. 获取适用的费用配置（按 category → item → sku → entity 多维度匹配）
    feeConfigs, err := c.loadApplicableFees(ctx, req)
    if err != nil {
        return nil, err
    }

    // 2. 按优先级排序
    sort.Slice(feeConfigs, func(i, j int) bool {
        return feeConfigs[i].Priority > feeConfigs[j].Priority
    })

    var result []*CalculatedFee

    // 3. 逐项计算，同类型费用仅取最高优先级
    feeTypeSeen := make(map[string]bool)
    for _, config := range feeConfigs {
        if feeTypeSeen[config.FeeType] {
            continue // 同类型费用只取优先级最高的一条
        }
        feeTypeSeen[config.FeeType] = true

        amount, err := c.calculateFeeAmount(config, req)
        if err != nil {
            log.Errorf("calculate fee %s failed: %v", config.FeeCode, err)
            continue
        }

        result = append(result, &CalculatedFee{
            FeeType:     config.FeeType,
            FeeName:     config.FeeName,
            Amount:      amount,
            CanDiscount: config.CanBeDiscounted == 1,
            DisplayType: config.DisplayType,
        })
    }

    return result, nil
}

// calculateFeeAmount 计算费用金额
func (c *FeeCalculator) calculateFeeAmount(config *model.FeeConfig, req *FeeCalcRequest) (decimal.Decimal, error) {
    baseAmount := req.BasePrice.Mul(decimal.NewFromInt(int64(req.Quantity)))

    var amount decimal.Decimal

    switch config.CalculationType {
    case "fixed":
        val := config.CalculationConfig["amount"].(float64)
        amount = decimal.NewFromFloat(val).Mul(decimal.NewFromInt(int64(req.Quantity)))

    case "percentage":
        pct := config.CalculationConfig["percentage"].(float64)
        amount = baseAmount.Mul(decimal.NewFromFloat(pct / 100))

    case "tiered":
        tiers := config.CalculationConfig["tiers"].([]interface{})
        for _, tier := range tiers {
            t := tier.(map[string]interface{})
            threshold := decimal.NewFromFloat(t["threshold"].(float64))
            if baseAmount.GreaterThanOrEqual(threshold) {
                amount = decimal.NewFromFloat(t["fee"].(float64))
                break
            }
        }

    default:
        return decimal.Zero, fmt.Errorf("unsupported calculation type: %s", config.CalculationType)
    }

    // 应用最小/最大限制
    if config.MinFee != nil && amount.LessThan(*config.MinFee) {
        amount = *config.MinFee
    }
    if config.MaxFee != nil && amount.GreaterThan(*config.MaxFee) {
        amount = *config.MaxFee
    }

    return amount, nil
}
```

### 4.6 价格公式生成

```go
// buildPriceFormula 构建人类可读的价格公式
func (e *PricingEngine) buildPriceFormula(resp *PriceResponse) string {
    formula := fmt.Sprintf("%.2f", resp.Subtotal)

    if resp.PromotionDiscount.GreaterThan(decimal.Zero) {
        formula += fmt.Sprintf(" - %.2f (促销)", resp.PromotionDiscount)
        for _, p := range resp.PromotionDetails {
            formula += fmt.Sprintf(" [%s: -%.2f]", p.ActivityName, p.Discount)
        }
    }

    if resp.TotalFee.GreaterThan(decimal.Zero) {
        formula += fmt.Sprintf(" + %.2f (费用)", resp.TotalFee)
        for _, f := range resp.FeeDetails {
            formula += fmt.Sprintf(" [%s: +%.2f]", f.FeeName, f.Amount)
        }
    }

    if resp.VoucherDiscount.GreaterThan(decimal.Zero) {
        formula += fmt.Sprintf(" - %.2f (优惠券)", resp.VoucherDiscount)
    }

    formula += fmt.Sprintf(" = %.2f %s", resp.FinalPrice, resp.Currency)
    return formula
}
```

---

## 五、核心流程设计

### 5.1 价格计算全流程

```
用户请求价格计算
        │
        ▼
┌─────────────────────────────────┐
│  1. 获取基础价格 (Base Price)    │
│     - SKU 基础价                │
│     - 时间维度价格（酒店/场次）  │
│     - 动态定价调整              │
└─────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────┐
│  2. 匹配营销活动 (Promotion)     │
│     - 按优先级遍历活动          │
│     - 检查生效条件              │
│     - 互斥/叠加规则判断         │
│     - 配额原子扣减              │
└─────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────┐
│  3. 计算费用 (Fee)              │
│     - DP Fee / Hub Fee          │
│     - Service Fee / Tax         │
│     - 同类型取最高优先级        │
└─────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────┐
│  4. 应用优惠券 (Voucher)         │
│     - 检查持有 & 有效期         │
│     - 验证使用条件              │
│     - 计算可抵扣金额            │
│     - 与促销互斥规则            │
└─────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────┐
│  5. 最终价格 & 精度处理          │
│     - FinalPrice = Subtotal     │
│       - Promotion + Fee         │
│       - Voucher                 │
│     - 币种精度对齐              │
│     - 确保价格 >= 0             │
└─────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────┐
│  6. 生成快照 (异步)             │
│     - 记录完整计算明细          │
│     - 关联订单（下单后）        │
│     - 价格公式人类可读          │
└─────────────────────────────────┘
        │
        ▼
返回最终价格（含明细 & 快照码）
```

### 5.2 价格一致性保障

**问题**：用户在商品列表、详情页、购物车、订单确认页看到的价格不一致。

```
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│ 列表页    │   │ 详情页    │   │ 购物车    │   │ 下单页    │
│ 实时计算  │   │ 实时计算  │   │ 生成快照  │   │ 验证快照  │
│ (可缓存)  │   │ (实时)    │   │ (30min)   │   │ (锁定)    │
└──────────┘   └──────────┘   └──────────┘   └──────────┘
      │              │              │               │
      └──────────────┴──────────────┴───────────────┘
                           │
                  统一 Pricing Engine API
```

**方案**：

```go
// 加入购物车 → 生成价格快照
func (s *CartService) AddToCart(ctx context.Context, req *AddToCartRequest) error {
    // 1. 实时计算价格
    priceResp, err := s.pricingEngine.Calculate(ctx, &PriceRequest{
        UserID: req.UserID, SKUID: req.SKUID, Quantity: req.Quantity,
        Scene: "cart",
    })
    if err != nil {
        return err
    }

    // 2. 生成价格快照（有效期30分钟）
    snapshot := &model.PriceSnapshot{
        SnapshotCode: generateSnapshotCode(),
        UserID:       req.UserID,
        SKUID:        req.SKUID,
        FinalPrice:   priceResp.FinalPrice,
        ExpiredAt:    time.Now().Add(30 * time.Minute),
    }
    if err := s.snapshotRepo.Create(ctx, snapshot); err != nil {
        return err
    }

    // 3. 购物车关联快照
    return s.cartRepo.Add(ctx, &model.CartItem{
        UserID: req.UserID, SKUID: req.SKUID,
        SnapshotCode: snapshot.SnapshotCode,
    })
}

// 下单 → 验证快照价格
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*model.Order, error) {
    // 1. 获取快照
    snapshot, err := s.snapshotRepo.GetByCode(ctx, req.SnapshotCode)
    if err != nil {
        return nil, err
    }

    // 2. 检查快照是否过期
    if snapshot.ExpiredAt.Before(time.Now()) {
        // 重新计算价格
        newPrice, _ := s.pricingEngine.Calculate(ctx, &PriceRequest{
            UserID: req.UserID, SKUID: req.SKUID, Scene: "checkout",
        })
        // 价格变动 → 要求用户重新确认
        if !newPrice.FinalPrice.Equal(snapshot.FinalPrice) {
            return nil, &PriceChangedError{
                OldPrice: snapshot.FinalPrice,
                NewPrice: newPrice.FinalPrice,
            }
        }
    }

    // 3. 使用快照价格创建订单
    order := &model.Order{
        FinalPrice:   snapshot.FinalPrice,
        SnapshotCode: snapshot.SnapshotCode,
    }
    return order, s.orderRepo.Create(ctx, order)
}
```

### 5.3 营销活动优先级与互斥

**规则矩阵**：

| 活动 | Priority | Exclusivity | Voucher Compatible | 说明 |
|------|----------|-------------|-------------------|------|
| 秒杀特价 | 15 | 1（互斥） | 0（与券互斥） | 独占，与券互斥 |
| 新用户立减50฿ | 10 | 0（可叠加） | 1（可与券叠加） | 可叠加，可与券叠加 |
| 满3000减200 | 5 | 0（可叠加） | 1（可与券叠加） | 可叠加，可与券叠加 |
| 周末特惠 | 3 | 0（可叠加） | 1（可与券叠加） | 可叠加，可与券叠加 |

**匹配流程**：

```
活动列表（按 Priority DESC 排序）
        │
        ▼
    秒杀特价 (P=15, 互斥)
    ├── 条件匹配 ✓ → 命中！exclusivity=1 → 停止遍历（独占）
    │   结果: 仅秒杀生效
    │
    └── 条件不匹配 ✗ → 继续
        │
        ▼
    新用户立减 (P=10, 可叠加)
    ├── 条件匹配 ✓ → 加入匹配列表
    │
    └── 条件不匹配 ✗ → 继续
        │
        ▼
    满3000减200 (P=5, 可叠加)
    ├── 条件匹配 ✓ → 加入匹配列表
    │
    └── 条件不匹配 ✗ → 继续
```

**计算示例**：

- **场景A**：新用户 + 满减 → 叠加生效（-50 + -200 = -250）
- **场景B**：秒杀 + 新用户 → 仅秒杀生效（秒杀优先级高且独占）
- **场景C**：新用户 + 优惠券 → 叠加生效
- **场景D**：秒杀 + 优惠券 → 仅秒杀生效（秒杀与券互斥）

### 5.4 Fee 可折扣性设计

**问题**：哪些费用可以被优惠券抵扣？

| 费用类型 | can_be_discounted | 原因 |
|---------|------------------|------|
| DP Fee | 0（不可抵扣） | 平台收入，不参与优惠 |
| Hub Fee | 按协议配置 | 商户协议决定 |
| Service Fee（选座费） | 1（可抵扣） | 增值服务，可参与优惠 |
| Tax | 0（不可抵扣） | 税费不可被优惠 |

**抵扣逻辑**：

```go
// calculateDiscountableAmount 计算优惠券可抵扣金额
func (a *VoucherApplier) calculateDiscountableAmount(req *VoucherApplyRequest) decimal.Decimal {
    // 基础金额 = 小计 - 促销折扣
    discountableAmount := req.Subtotal.Sub(req.PromotionDiscount)

    // 加上可被抵扣的费用
    for _, fee := range req.FeeDetails {
        if fee.CanDiscount {
            discountableAmount = discountableAmount.Add(fee.Amount)
        }
    }

    return discountableAmount
}
```

**示例**：

```
商品小计:      1000฿
促销折扣:      -100฿
DP Fee:        +10฿  (不可抵扣)
Hub Fee:       +20฿  (可抵扣)
Service Fee:   +5฿   (可抵扣)
──────────────────────────────
优惠券可抵扣金额 = (1000 - 100) + 20 + 5 = 925฿
优惠券面额 50฿ → 实际抵扣 50฿
最终价格 = 1000 - 100 + 10 + 20 + 5 - 50 = 885฿
```

---

## 六、配额并发安全

### 6.1 营销活动配额管理

秒杀/限量活动场景下，配额扣减必须保证原子性。使用 Redis Lua 脚本实现：

```lua
-- promotion_quota_consume.lua
-- 营销活动配额原子扣减
-- KEYS[1]: promotion:{activity_id}:quota
-- KEYS[2]: promotion:{activity_id}:user:{user_id}
-- ARGV[1]: total_quota (总配额)
-- ARGV[2]: per_user_limit (每用户限制, 0=不限)
-- ARGV[3]: activity_id
-- 返回: 1=成功, 0=配额不足, -1=用户已达上限

local total_key = KEYS[1]
local user_key = KEYS[2]
local total_quota = tonumber(ARGV[1])
local per_user_limit = tonumber(ARGV[2])

-- 1. 检查用户限制
if per_user_limit > 0 then
    local user_count = tonumber(redis.call('GET', user_key) or '0')
    if user_count >= per_user_limit then
        return -1  -- 用户已达上限
    end
end

-- 2. 原子扣减总配额
local used = tonumber(redis.call('GET', total_key) or '0')
if used >= total_quota then
    return 0  -- 配额不足
end

redis.call('INCR', total_key)

-- 3. 增加用户使用次数
if per_user_limit > 0 then
    redis.call('INCR', user_key)
    redis.call('EXPIRE', user_key, 86400 * 30)  -- 30天过期
end

return 1  -- 成功
```

### 6.2 优惠券核销并发安全

```go
// 优惠券核销：Redis + MySQL 双写
func (a *VoucherApplier) consumeVoucher(ctx context.Context, userID int64, voucherCode string, orderID int64) error {
    lockKey := fmt.Sprintf("voucher:lock:%d:%s", userID, voucherCode)

    // 1. Redis 分布式锁（防止并发核销）
    acquired, err := a.redis.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
    if err != nil || !acquired {
        return errors.New("voucher is being used by another request")
    }
    defer a.redis.Del(ctx, lockKey)

    // 2. 检查优惠券状态
    userVoucher, err := a.userVoucherRepo.GetByUserAndCode(ctx, userID, voucherCode)
    if err != nil {
        return err
    }
    if userVoucher.Status != StatusUnused {
        return errors.New("voucher already used or expired")
    }

    // 3. 更新状态（乐观锁）
    affected, err := a.userVoucherRepo.UpdateStatus(ctx, userVoucher.ID, StatusUnused, StatusUsed, orderID)
    if err != nil {
        return err
    }
    if affected == 0 {
        return errors.New("voucher status changed, retry")
    }

    // 4. 更新优惠券使用数量
    a.voucherRepo.IncrUsedQuantity(ctx, userVoucher.VoucherID)

    return nil
}
```

### 6.3 退款时优惠券/配额回退

```go
// 订单取消/退款 → 回退优惠券和配额
func (s *RefundService) rollbackPriceComponents(ctx context.Context, order *model.Order) error {
    snapshot, err := s.snapshotRepo.GetByOrderID(ctx, order.ID)
    if err != nil {
        return err
    }

    // 1. 回退优惠券
    for _, v := range snapshot.VoucherDetails {
        if err := s.voucherApplier.RollbackVoucher(ctx, order.UserID, v.VoucherCode); err != nil {
            log.Errorf("rollback voucher %s failed: %v", v.VoucherCode, err)
            // 不中断，继续回退其他组件
        }
    }

    // 2. 回退营销活动配额
    for _, p := range snapshot.PromotionDetails {
        if err := s.quotaService.Rollback(ctx, p.ActivityID, order.UserID); err != nil {
            log.Errorf("rollback promotion %d quota failed: %v", p.ActivityID, err)
        }
    }

    // 3. 发送退款事件（Kafka）
    s.eventPublisher.Publish(ctx, &PriceRefundEvent{
        OrderID:    order.ID,
        SnapshotCode: snapshot.SnapshotCode,
        RefundAmount: snapshot.FinalPrice,
    })

    return nil
}
```

---

## 七、样例数据与计算场景

### 7.1 场景一：电影票购买（促销 + Fee + Voucher）

**基础数据**：

```sql
-- SKU: 阿凡达3 IMAX 成人票
INSERT INTO `sku_tab` (id, sku_code, item_id, sku_name, price, original_price, cost_price)
VALUES (2000001, 'SKU_MOVIE_AVATAR3_ADULT', 200001, 'IMAX 3D 成人票', 480.00, 550.00, 380.00);

-- 营销活动: 新用户立减50฿
INSERT INTO `promotion_activity_tab` (id, activity_code, activity_name, activity_type, category_ids, user_type,
    discount_type, discount_value, priority, exclusivity, voucher_compatible, status, start_time, end_time)
VALUES (1001, 'PROMO_NEW_USER_50', '新用户立减50฿', 'new_user', '[30001]', 'new',
    'fixed_amount', '{"amount": 50}', 10, 0, 1, 1, '2026-01-01', '2026-12-31');

-- 费用: DP Fee 10฿
INSERT INTO `fee_config_tab` (id, fee_code, fee_name, fee_type, category_id, calculation_type, calculation_config, can_be_discounted)
VALUES (101, 'FEE_DP_MOVIE', 'DP 平台服务费', 'dp_fee', 30001, 'fixed', '{"amount": 10}', 0);

-- 费用: 选座费 5฿
INSERT INTO `fee_config_tab` (id, fee_code, fee_name, fee_type, category_id, calculation_type, calculation_config, can_be_discounted)
VALUES (102, 'FEE_SEAT_SELECT', '选座服务费', 'service_fee', 30001, 'fixed', '{"amount": 5}', 0);

-- 优惠券: 电影通用券30฿
INSERT INTO `voucher_tab` (id, voucher_code, voucher_name, voucher_type, discount_type, discount_value, min_purchase_amount, category_ids)
VALUES (5001, 'VOUCHER_MOVIE_30', '电影通用券30฿', 'discount', 'fixed_amount', '{"amount": 30}', 10.00, '[30001]');
```

**计算过程**：

```
请求: user_id=100001, sku_id=2000001, quantity=2, voucher=VOUCHER_MOVIE_30

Step 1: 基础价格
  单价: 480.00฿ × 2张 = 960.00฿

Step 2: 营销活动
  命中: 新用户立减50฿ (per unit)
  折扣: 50.00 × 2 = 100.00฿
  促销后: 860.00฿

Step 3: 费用
  DP Fee:  10.00 × 2 = 20.00฿  (不可抵扣)
  选座费:   5.00 × 2 = 10.00฿  (不可抵扣)
  费用合计: 30.00฿
  加费后: 890.00฿

Step 4: 优惠券
  电影通用券: -30.00฿
  券后: 860.00฿

Step 5: 最终价格
  公式: 960.00 - 100.00 (促销) + 30.00 (费用) - 30.00 (优惠券) = 860.00 THB
```

### 7.2 场景二：酒店预订（动态定价 + 满减 + 阶梯 Fee）

**基础数据**：

```sql
-- 动态定价: 库存紧张加价15%
INSERT INTO `dynamic_pricing_rule_tab` (id, rule_code, rule_name, category_id, rule_type,
    trigger_condition, adjustment_type, adjustment_value, priority, status, effective_start, effective_end)
VALUES (201, 'RULE_HOTEL_INVENTORY', '库存紧张加价', 10001, 'inventory_based',
    '{"inventory_threshold": 5}', 'percentage', 15.00, 1, 1, '2026-01-01', '2026-12-31');

-- 营销活动: 满3000减200
INSERT INTO `promotion_activity_tab` (id, activity_code, activity_name, activity_type, category_ids,
    discount_type, discount_value, priority, status, start_time, end_time)
VALUES (1002, 'PROMO_HOTEL_FULL_RED', '满3000减200', 'full_reduction', '[10001]',
    'full_reduction', '{"threshold": 3000, "discount": 200}', 5, 1, '2026-01-01', '2026-03-31');

-- 费用: 阶梯式 Hub Fee
INSERT INTO `fee_config_tab` (id, fee_code, fee_name, fee_type, category_id, calculation_type, calculation_config, min_fee, max_fee)
VALUES (201, 'FEE_HUB_HOTEL_TIERED', 'Hub 服务费（阶梯）', 'hub_fee', 10001, 'tiered',
    '{"tiers": [{"threshold": 5000, "fee": 150}, {"threshold": 3000, "fee": 100}, {"threshold": 0, "fee": 50}]}', 50.00, 150.00);
```

**计算过程**：

```
请求: user_id=100002, sku_id=1000002(豪华房), quantity=1, context.nights=2, context.available_rooms=3

Step 1: 基础价格
  基础房价: 4200.00฿/晚 × 2晚 = 8400.00฿
  动态定价: 库存=3间 <= 阈值5间 → 加价15%
  加价金额: 8400.00 × 15% = 1260.00฿
  小计: 9660.00฿

Step 2: 营销活动
  命中: 满3000减200 (9660 >= 3000 ✓)
  折扣: 200.00฿
  促销后: 9460.00฿

Step 3: 费用
  Hub Fee: 9460.00 >= 5000 → 阶梯费 150.00฿
  加费后: 9610.00฿

Step 4: 优惠券
  未使用

Step 5: 最终价格
  公式: 9660.00 - 200.00 (促销) + 150.00 (费用) = 9610.00 THB
```

### 7.3 场景三：TopUp 充值（阶梯优惠 + 平台补贴）

```
请求: user_id=100003, item=AIS 500฿充值, sku_id=3000001, quantity=1

Step 1: 基础价格
  面额: 500.00฿

Step 2: 营销活动
  命中: 充值阶梯优惠 (500 >= 阈值500 → 5%折扣)
  折扣: 500.00 × 5% = 25.00฿
  封顶: min(25.00, 50.00) = 25.00฿
  促销后: 475.00฿

Step 3: 费用
  无额外费用

Step 4: 优惠券
  未使用

Step 5: 最终价格
  公式: 500.00 - 25.00 (促销) = 475.00 THB
  用户支付 475฿ 获得 500฿ 话费
```

---

## 八、降级策略

### 8.1 多级降级方案

价格服务是核心链路，任何组件故障不能导致整个计算失败。

```
┌─────────────────────────────────────────────────────────────┐
│                    降级策略分级                               │
├──────────┬──────────────────┬───────────────────────────────┤
│ 降级级别  │ 触发条件          │ 降级行为                      │
├──────────┼──────────────────┼───────────────────────────────┤
│ Level 0  │ 全部正常          │ 完整计算（Base+Promo+Fee+Voucher）│
│ Level 1  │ 优惠券服务不可用   │ 跳过优惠券，提示用户稍后重试    │
│ Level 2  │ 促销服务不可用     │ 仅 Base + Fee，无促销折扣      │
│ Level 3  │ 费用服务不可用     │ 仅 Base，无额外费用            │
│ Level 4  │ 缓存全部失效       │ 直接查询 MySQL 基础价格        │
│ Level 5  │ MySQL 也不可用     │ 返回上一次缓存的价格快照       │
└──────────┴──────────────────┴───────────────────────────────┘
```

### 8.2 降级实现

```go
// DegradeConfig 降级配置
type DegradeConfig struct {
    PromotionTimeout   time.Duration `json:"promotion_timeout"`   // 促销匹配超时: 50ms
    FeeTimeout         time.Duration `json:"fee_timeout"`         // 费用计算超时: 30ms
    VoucherTimeout     time.Duration `json:"voucher_timeout"`     // 优惠券应用超时: 50ms
    PromotionDegrade   bool          `json:"promotion_degrade"`   // 促销服务强制降级开关
    VoucherDegrade     bool          `json:"voucher_degrade"`     // 优惠券服务强制降级开关
}

// matchPromotionsWithDegrade 带降级的促销匹配
func (e *PricingEngine) matchPromotionsWithDegrade(ctx context.Context, req *PriceRequest, basePrice decimal.Decimal) ([]*MatchedPromotion, error) {
    // 强制降级开关
    if e.degradeConfig.PromotionDegrade {
        return nil, nil
    }

    // 超时控制
    ctx, cancel := context.WithTimeout(ctx, e.degradeConfig.PromotionTimeout)
    defer cancel()

    promotions, err := e.promotionMatcher.Match(ctx, &PromotionMatchRequest{
        UserID: req.UserID, ItemID: req.ItemID, SKUID: req.SKUID,
        CategoryID: req.CategoryID, Quantity: req.Quantity, BasePrice: basePrice,
    })
    if err != nil {
        // 促销匹配失败 → 降级，不影响核心流程
        metrics.PromotionDegradeCounter.Inc()
        return nil, err
    }

    return promotions, nil
}
```

### 8.3 降级开关配置

通过配置中心（Apollo / Nacos）动态控制降级开关，无需重启服务：

```json
{
  "pricing_degrade": {
    "promotion_degrade": false,
    "voucher_degrade": false,
    "fee_degrade": false,
    "promotion_timeout_ms": 50,
    "fee_timeout_ms": 30,
    "voucher_timeout_ms": 50
  }
}
```

---

## 九、缓存策略

### 9.1 多级缓存架构

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  L1: 本地缓存  │───▶│  L2: Redis    │───▶│  L3: MySQL    │
│  (sync.Map)   │    │  (Cluster)    │    │  (主从)       │
│  TTL: 10s     │    │  TTL: 5min    │    │  持久存储      │
│  大小: 10000   │    │  大小: 不限     │    │               │
└──────────────┘    └──────────────┘    └──────────────┘
```

### 9.2 缓存 Key 设计

| 数据类型 | Redis Key | TTL | 说明 |
|---------|-----------|-----|------|
| SKU 基础价格 | `price:base:{sku_id}` | 5min | 基础价格变化不频繁 |
| 品类活动列表 | `promo:active:cat:{category_id}` | 5min | 缓存生效活动 |
| 费用配置 | `fee:config:cat:{category_id}` | 10min | 费用配置变化少 |
| 优惠券信息 | `voucher:info:{voucher_code}` | 5min | 券信息 |
| 活动配额 | `promo:quota:{activity_id}` | 无过期 | 配额计数器 |
| 价格计算结果 | `price:calc:{sku_id}:{user_hash}` | 60s | 短缓存，防穿透 |

### 9.3 缓存更新策略

```go
// 价格变更时主动失效缓存（通过 Kafka 消息驱动）
func (h *PriceCacheInvalidHandler) HandlePriceChange(ctx context.Context, event *PriceChangeEvent) error {
    keys := []string{
        fmt.Sprintf("price:base:%d", event.SKUID),
        fmt.Sprintf("price:calc:%d:*", event.SKUID), // 模式匹配删除
    }

    for _, key := range keys {
        if strings.Contains(key, "*") {
            // Scan + Delete 模式匹配
            h.redis.ScanAndDelete(ctx, key)
        } else {
            h.redis.Del(ctx, key)
        }
    }

    // 同时清除本地缓存
    h.localCache.Delete(fmt.Sprintf("price:base:%d", event.SKUID))

    return nil
}
```

---

## 十、Kafka 事件设计

### 10.1 事件类型

| 事件 | Topic | 生产者 | 消费者 | 说明 |
|------|-------|--------|--------|------|
| 价格变更 | `dp.pricing.price_changed` | 价格管理后台 | Pricing Engine、搜索服务 | SKU 基础价格变更 |
| 活动创建/更新 | `dp.pricing.promotion_changed` | 营销管理后台 | Pricing Engine | 营销活动变更 |
| 费用配置变更 | `dp.pricing.fee_changed` | 费用管理后台 | Pricing Engine | 费用配置变更 |
| 价格快照生成 | `dp.pricing.snapshot_created` | Pricing Engine | 数据仓库 | 快照异步落库 |
| 优惠券核销 | `dp.pricing.voucher_consumed` | Pricing Engine | 优惠券服务 | 优惠券使用 |
| 退款回退 | `dp.pricing.refund_rollback` | 订单服务 | Pricing Engine | 退款时回退配额/券 |

### 10.2 事件 Schema

```protobuf
// PriceChangeEvent 价格变更事件
message PriceChangeEvent {
    int64  sku_id       = 1;
    string change_type  = 2;  // base_price, promotion, fee, dynamic_rule
    string old_value    = 3;  // JSON
    string new_value    = 4;  // JSON
    string reason       = 5;
    int64  operator_id  = 6;
    int64  timestamp_ms = 7;
}

// SnapshotCreatedEvent 快照生成事件
message SnapshotCreatedEvent {
    string snapshot_code     = 1;
    int64  user_id           = 2;
    int64  sku_id            = 3;
    string final_price       = 4;
    string price_formula     = 5;
    string promotion_details = 6;  // JSON
    string fee_details       = 7;  // JSON
    string voucher_details   = 8;  // JSON
    int64  timestamp_ms      = 9;
}
```

### 10.3 幂等消费

```go
// 幂等处理：避免重复消费导致配额多次扣减
func (h *PromotionChangeHandler) Handle(ctx context.Context, msg *kafka.Message) error {
    eventID := string(msg.Key)

    // Redis 幂等键（24小时过期）
    idempotentKey := fmt.Sprintf("pricing:event:idempotent:%s", eventID)
    set, err := h.redis.SetNX(ctx, idempotentKey, "1", 24*time.Hour).Result()
    if err != nil || !set {
        log.Infof("duplicate event %s, skip", eventID)
        return nil // 重复消息，跳过
    }

    // 处理消息
    event := &PriceChangeEvent{}
    if err := proto.Unmarshal(msg.Value, event); err != nil {
        return err
    }

    // 失效缓存
    return h.cacheInvalidator.Invalidate(ctx, event)
}
```

---

## 十一、性能优化

### 11.1 并行计算

基础价格、促销匹配、费用计算三者互相独立，可并行执行：

```go
func (e *PricingEngine) parallelCalculate(ctx context.Context, req *PriceRequest) (
    *BasePriceResult, []*MatchedPromotion, []*CalculatedFee, error) {

    var (
        basePrice  *BasePriceResult
        promotions []*MatchedPromotion
        fees       []*CalculatedFee
        baseErr, promoErr, feeErr error
        wg sync.WaitGroup
    )

    wg.Add(3)

    go func() {
        defer wg.Done()
        basePrice, baseErr = e.basePriceCalculator.Calculate(ctx, req)
    }()

    go func() {
        defer wg.Done()
        promotions, promoErr = e.promotionMatcher.Match(ctx, &PromotionMatchRequest{...})
    }()

    go func() {
        defer wg.Done()
        fees, feeErr = e.feeCalculator.Calculate(ctx, &FeeCalcRequest{...})
    }()

    wg.Wait()

    // 基础价格是必须的，其他可降级
    if baseErr != nil {
        return nil, nil, nil, baseErr
    }

    return basePrice, promotions, fees, nil
}
```

### 11.2 批量价格计算

列表页需要同时计算多个商品的价格，使用批量接口减少 RPC 和 DB 调用：

```go
// BatchCalculate 批量价格计算（列表页场景）
func (e *PricingEngine) BatchCalculate(ctx context.Context, reqs []*PriceRequest) ([]*PriceResponse, error) {
    // 1. 批量获取 SKU 基础价格（一次 DB 查询）
    skuIDs := extractSKUIDs(reqs)
    basePrices, err := e.basePriceCalculator.BatchGet(ctx, skuIDs)
    if err != nil {
        return nil, err
    }

    // 2. 批量获取品类活动（按 category 分组，减少缓存查询）
    categoryIDs := extractCategoryIDs(reqs)
    promotionsByCategory, err := e.promotionMatcher.BatchLoadActive(ctx, categoryIDs)
    if err != nil {
        log.Warnf("batch load promotions failed, degrade: %v", err)
    }

    // 3. 并行计算每个商品的价格
    results := make([]*PriceResponse, len(reqs))
    var wg sync.WaitGroup

    // 控制并发数
    sem := make(chan struct{}, 20)

    for i, req := range reqs {
        wg.Add(1)
        sem <- struct{}{}
        go func(idx int, r *PriceRequest) {
            defer wg.Done()
            defer func() { <-sem }()
            results[idx], _ = e.calculateSingle(ctx, r, basePrices, promotionsByCategory)
        }(i, req)
    }

    wg.Wait()
    return results, nil
}
```

### 11.3 性能指标

| 指标 | 目标值 | 监控方式 |
|------|--------|----------|
| **单价计算 P50** | < 20ms | Prometheus |
| **单价计算 P99** | < 100ms | Prometheus |
| **批量计算 P99**（20个） | < 200ms | Prometheus |
| **缓存命中率** | > 85% | Redis Monitor |
| **QPS** | > 10,000 | Load Test |
| **CPU 使用率** | < 60% | Grafana |

---

## 十二、监控与告警

### 12.1 核心监控指标

```
# 价格计算 QPS 和延迟
pricing_calculate_total{scene="list|detail|cart|checkout", status="success|fail|degrade"}
pricing_calculate_duration_seconds{scene, quantile="0.5|0.9|0.99"}

# 营销活动命中
pricing_promotion_match_total{activity_type, result="hit|miss"}
pricing_promotion_quota_remaining{activity_id}

# 费用计算
pricing_fee_calculate_total{fee_type, status="success|fail"}
pricing_fee_amount_sum{fee_type}

# 优惠券
pricing_voucher_apply_total{voucher_type, status="success|fail|expired|conflict"}

# 缓存
pricing_cache_hit_total{cache_level="l1|l2", data_type="base_price|promotion|fee"}
pricing_cache_miss_total{cache_level, data_type}

# 降级
pricing_degrade_total{level="promotion|fee|voucher|full"}

# 价格快照
pricing_snapshot_create_total{status="success|fail"}
pricing_snapshot_expired_total
```

### 12.2 告警规则

| 告警名称 | 条件 | 级别 | 处理 |
|---------|------|------|------|
| **价格计算超时** | P99 > 200ms 持续5分钟 | P1 | 检查缓存、DB、下游服务 |
| **降级率过高** | 降级率 > 5% 持续3分钟 | P1 | 检查促销/优惠券服务可用性 |
| **价格计算失败率** | 失败率 > 1% 持续3分钟 | P0 | 紧急排查，可能影响下单 |
| **缓存命中率下降** | 命中率 < 70% 持续5分钟 | P2 | 检查 Redis 集群状态 |
| **配额即将耗尽** | 剩余配额 < 10% | P3 | 通知运营补充配额 |
| **价格异常波动** | 同 SKU 价格波动 > 30% | P2 | 检查动态定价规则 |

### 12.3 价格审计

```go
// ReconstructPriceCalculation 根据订单还原价格计算（客服工具）
func (s *AuditService) ReconstructPriceCalculation(ctx context.Context, orderID int64) (*PriceAuditResult, error) {
    // 1. 获取价格快照
    snapshot, err := s.snapshotRepo.GetByOrderID(ctx, orderID)
    if err != nil {
        return nil, err
    }

    // 2. 解析计算明细
    promotionDetails := parseJSON[[]PromotionDetail](snapshot.PromotionDetails)
    feeDetails := parseJSON[[]FeeDetail](snapshot.FeeDetails)
    voucherDetails := parseJSON[[]VoucherDetail](snapshot.VoucherDetails)

    // 3. 验证计算正确性
    calculatedFinalPrice := snapshot.Subtotal.
        Sub(snapshot.PromotionDiscount).
        Add(snapshot.TotalFee).
        Sub(snapshot.VoucherDiscount)

    isValid := calculatedFinalPrice.Equal(snapshot.FinalPrice)
    if !isValid {
        log.Errorf("price snapshot %s inconsistency: calculated=%s, recorded=%s",
            snapshot.SnapshotCode, calculatedFinalPrice, snapshot.FinalPrice)
    }

    // 4. 返回审计结果
    return &PriceAuditResult{
        SnapshotCode:     snapshot.SnapshotCode,
        BasePrice:        snapshot.BasePrice,
        Subtotal:         snapshot.Subtotal,
        PromotionDetails: promotionDetails,
        FeeDetails:       feeDetails,
        VoucherDetails:   voucherDetails,
        FinalPrice:       snapshot.FinalPrice,
        PriceFormula:     snapshot.PriceFormula,
        CalculationTime:  snapshot.CreatedAt,
        IsValid:          isValid,
    }, nil
}
```

---

## 十三、新品类接入指南

**四步接入**：

1. **配置基础价格策略**：确定价格来源（SKU 固定价 / 日历价 / 动态定价规则）。
2. **配置费用规则**：确定该品类涉及哪些费用类型，INSERT `fee_config_tab`。
3. **关联营销活动**：在 `promotion_activity_tab` 中 `category_ids` 加入新品类 ID。
4. **注册品类策略**（可选）：如有特殊计价逻辑，实现 `BasePriceCalculator` 接口。

```go
// 示例：接入新品类 "演唱会门票"

// 1. 注册品类基础价格计算器（如有特殊逻辑）
engine.RegisterBasePriceCalculator("concert", &ConcertBasePriceCalculator{
    // 按场次 × 区域 × 票种计算基础价
})

// 2. 配置费用（SQL）
// INSERT INTO fee_config_tab (fee_code, fee_name, fee_type, category_id, calculation_type, calculation_config)
// VALUES ('FEE_DP_CONCERT', 'DP平台费', 'dp_fee', 50001, 'percentage', '{"percentage": 3}');
// VALUES ('FEE_TICKET_SERVICE', '票务服务费', 'service_fee', 50001, 'fixed', '{"amount": 15}');

// 3. 营销活动自动关联（在活动 category_ids 中加入 50001 即可）

// 4. 无需修改 Pricing Engine 核心代码
```

**品类接入检查清单**：

| 检查项 | 说明 | 必须 |
|--------|------|------|
| 基础价格来源 | SKU 固定价 / 日历价 / 动态定价 | ✅ |
| 费用配置 | 确定涉及哪些费用类型 | ✅ |
| 币种精度 | 确认该品类对应的币种和精度 | ✅ |
| 营销活动兼容 | 验证现有活动对新品类是否生效 | ✅ |
| 价格快照字段 | 确认 context 中是否需要额外字段 | ⬜ |
| 自定义计算器 | 是否需要特殊基础价格计算逻辑 | ⬜ |

---

## 十四、业界价格模型演进与对比

### 14.1 价格模型演进史

```
┌─────────────────────────────────────────────────────────────────┐
│                  价格引擎技术演进路径                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Phase 1: 单一价格时代（2000年前）                               │
│  ┌─────────────────────────────┐                               │
│  │ 商品一口价，总价 = 单价 × 数量 │                               │
│  │ 代表: 早期零售 ERP            │                               │
│  └─────────────────────────────┘                               │
│               ↓                                                  │
│  Phase 2: 促销价格分离（2000-2010）                              │
│  ┌─────────────────────────────┐                               │
│  │ original_price + sale_price │                               │
│  │ 简单时间段促销              │                               │
│  │ 代表: 早期淘宝、eBay        │                               │
│  └─────────────────────────────┘                               │
│               ↓                                                  │
│  Phase 3: 规则引擎化（2010-2015）                                │
│  ┌─────────────────────────────┐                               │
│  │ 促销规则独立成表            │                               │
│  │ 优先级 & 互斥机制          │                               │
│  │ 独立优惠券系统              │                               │
│  │ 代表: 淘宝天猫、京东        │                               │
│  └─────────────────────────────┘                               │
│               ↓                                                  │
│  Phase 4: 智能定价（2015-2020）                                  │
│  ┌─────────────────────────────┐                               │
│  │ 动态定价（供需、库存、竞品）│                               │
│  │ 个性化定价（用户画像）      │                               │
│  │ 价格快照 & 审计             │                               │
│  │ 代表: Uber、Airbnb、Amazon  │                               │
│  └─────────────────────────────┘                               │
│               ↓                                                  │
│  Phase 5: AI 赋能（2020至今）                                    │
│  ┌─────────────────────────────┐                               │
│  │ ML 定价模型（需求预测）     │                               │
│  │ 实时竞品监控                │                               │
│  │ A/B 测试自动化              │                               │
│  │ 代表: Amazon ML、阿里智能   │                               │
│  └─────────────────────────────┘                               │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 14.2 各阶段对比

| 维度 | Phase 1 | Phase 2 | Phase 3 | Phase 4 | Phase 5 |
|------|---------|---------|---------|---------|---------|
| **计算方式** | 硬编码 | 配置化 | 规则引擎 | ML 模型 | 在线学习 |
| **响应时间** | <10ms | <50ms | <100ms | <200ms | <100ms |
| **扩展性** | 差 | 中 | 好 | 好 | 优秀 |
| **维护成本** | 高 | 中 | 中 | 中 | 低 |
| **智能程度** | 无 | 低 | 中 | 高 | 极高 |
| **代表企业** | 早期电商 | 淘宝/京东 | 天猫/Amazon | Uber/Airbnb | Amazon/阿里 |

### 14.3 主流电商平台对比

#### 淘宝/天猫

```
价格计算链路: 商品价格 → 营销活动 → 优惠券 → 积分 → 红包
技术栈: Drools 规则引擎 + Tair 分布式缓存 + HSF/Dubbo RPC
性能: P99 < 50ms
特色: 双11 大促复杂叠加规则（店铺券 + 品类券 + 跨店券 + 红包 + 津贴）
```

#### 京东

```
价格计算链路: 基础价 → 会员价(Plus) → 活动 → 优惠券 → 京豆 → 运费
技术栈: 自研规则引擎 + Redis 集群
性能: P99 < 80ms
特色: Plus 会员价体系 + 京豆积分深度融合
```

#### Amazon

```
价格计算链路: 基础价 → ML 动态定价 → Prime 折扣 → Lightning Deal → Subscribe & Save
技术栈: 自研 ML Pipeline + ElastiCache
性能: P99 < 100ms
特色: ML 驱动定价（需求弹性 × 竞品价格 × 库存水平）
```

#### Uber (动态定价)

```
定价逻辑: base_price = distance × rate + duration × rate → surge_multiplier
技术栈: 实时供需计算 + ML 预测 + 平滑处理
特色: 供需比驱动 Surge Pricing，平滑处理避免价格突变
```

### 14.4 我们的定位与对比

| 特性 | 我们的设计 | 淘宝天猫 | 京东 | Amazon |
|------|-----------|---------|------|--------|
| **分层架构** | 4层（Base/Promo/Fee/Voucher） | 5层（含积分/红包） | 5层（含京豆） | 3层（简化） |
| **规则引擎** | DB 配置 + JSON | Drools + 自研 | 自研 | 自研 |
| **动态定价** | 规则驱动 | ML 驱动 | 规则驱动 | ML 驱动 |
| **价格快照** | ✅ 完整支持 | ✅ 支持 | ✅ 支持 | ✅ 支持 |
| **费用拆分** | ✅ 详细（DP/Hub/Service/Tax） | ✅ 详细 | ✅ 详细 | ✅ 详细 |
| **多币种** | ✅ 6种东南亚货币 | 单一（CNY） | 单一（CNY） | 全球多币种 |
| **降级策略** | ✅ 5级降级 | ✅ 完善 | ✅ 完善 | ✅ 完善 |
| **ML 定价** | ❌ 待规划 | ✅ 深度应用 | ⚠️ 部分 | ✅ 核心能力 |

**我们的差异化优势**：

1. **专注虚拟商品**：无物流成本，关注时间维度定价（酒店日历、电影场次）。
2. **费用透明化**：DP Fee、Hub Fee 独立管理，可配置是否可被优惠券抵扣。
3. **多币种原生支持**：东南亚6国币种精度处理内置。
4. **审计合规**：完整价格快照 + 人类可读公式，满足监管要求。

---

## 十五、设计总结

### 15.1 核心设计决策

| 决策 | 选择 | 原因 |
|------|------|------|
| **统一 vs 独立** | 统一 Pricing Engine + 策略模式 | 复用计算逻辑，新品类零代码接入 |
| **同步 vs 异步** | 价格计算同步，快照异步写入 | 热路径极速，冷路径可靠 |
| **缓存策略** | L1 本地 + L2 Redis + L3 MySQL | 多级缓存，高性能 + 可靠 |
| **精度处理** | decimal.Decimal + 币种精度表 | 避免浮点误差 |
| **降级策略** | 5级降级，促销/券可降级，基础价不可降级 | 保证核心链路可用 |
| **互斥规则** | Priority + Exclusivity 字段 | 灵活配置，无需改代码 |
| **配额管理** | Redis Lua 原子操作 | 高并发场景不超卖 |
| **审计追溯** | 价格快照 + 变更日志 | 完整记录，客诉可还原 |

### 15.2 关键技术栈

| 组件 | 技术选型 | 说明 |
|------|----------|------|
| Pricing Engine | Go 微服务 | 价格计算核心引擎 |
| 缓存层 | Redis Cluster + sync.Map | 多级缓存 |
| 规则引擎 | MySQL + JSON 配置 | 灵活配置营销规则 |
| 价格快照 | MySQL | 持久化价格明细 |
| 消息队列 | Kafka | 价格变动事件、快照异步写入 |
| 配额管理 | Redis Lua | 原子扣减 |
| 监控 | Prometheus + Grafana | 性能监控与告警 |
| 配置中心 | Apollo / Nacos | 降级开关、超时配置 |

### 15.3 实施路线

| 阶段 | 时间 | 内容 | 交付物 |
|------|------|------|--------|
| **P1: 基础架构** | 2周 | 数据库表、Pricing Engine 框架、基础价格层 | 可运行的计算引擎 |
| **P2: 营销活动** | 3周 | 促销规则引擎、匹配器、配额管理、管理后台 | 支持折扣/满减/秒杀 |
| **P3: 费用与优惠券** | 2周 | Fee 计算器、Voucher 系统、叠加规则 | 完整4层计算 |
| **P4: 性能与审计** | 2周 | 多级缓存、并行计算、价格快照、审计工具 | P99 < 100ms |
| **P5: 全量上线** | 1周 | 全品类覆盖、监控告警、文档培训 | 生产环境稳定运行 |

### 15.4 改进路线图

```
┌─────────────────────────────────────────────────────────────────┐
│                  价格引擎改进路线图                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Q1 2026: 基础完善                                               │
│  ✅ 统一价格中心上线                                             │
│  ✅ 价格快照机制                                                 │
│  ✅ 多级缓存策略                                                 │
│  ✅ 多级降级方案                                                 │
│                                                                  │
│  Q2 2026: 智能化初步                                             │
│  🔲 动态定价规则增强                                             │
│  🔲 A/B 测试平台集成                                             │
│  🔲 用户分层定价（新用户/VIP/高净值）                             │
│                                                                  │
│  Q3 2026: ML 模型引入                                            │
│  🔲 需求预测模型（LSTM）                                         │
│  🔲 价格弹性分析                                                 │
│  🔲 竞品价格监控                                                 │
│                                                                  │
│  Q4 2026: 实时智能                                               │
│  🔲 实时特征平台（Flink）                                        │
│  🔲 在线模型更新                                                 │
│  🔲 多目标优化（收入/转化/毛利）                                  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 15.5 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| **规则配置错误** | 价格异常 | 配置审批流程 + 灰度发布 + 回滚机制 |
| **性能瓶颈** | 响应超时 | 缓存预热 + 限流降级 + 水平扩展 |
| **价格不一致** | 用户投诉 | 统一 API + 快照机制 + 实时校验 |
| **促销超卖** | 成本损失 | Redis Lua 原子扣减 + 异步对账 |
| **多币种精度** | 金额误差 | decimal.Decimal + 币种精度表 + 分摊尾差处理 |
| **缓存穿透** | DB 压力 | 布隆过滤器 + 空值缓存 + 限流 |
