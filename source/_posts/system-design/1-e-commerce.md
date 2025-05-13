---
title: 互联网业务系统 - 电商系统设计
categories:
- 系统设计
---


## 电商系统整体架构设计
### 业务流 (business process)
<p align="center">
  <img src="/images/E-commerce-whole-business-process.webp" width=800 height=500>
  <br/>
  <strong><a href="https://axureboutique.com/blogs/product-design/understanding-the-structure-of-e-commerce-products">E-commerce process</a></strong>
</p>

### 系统流 (system process)
<p align="center">
  <img src="/images/E-commerce-whole-system-process.webp" width=800 height=500>
  <br/>
  <strong><a href="https://axureboutique.com/blogs/product-design/understanding-the-structure-of-e-commerce-products">E-commerce whole process of system</a></strong>
</p>

### 系统和产品架构 (Product Structure)
<p align="center">
  <img src="/images/E-commerce-product-structure.webp" width=800 height=600>
  <br/>
  <strong><a href="https://axureboutique.com/blogs/product-design/understanding-the-structure-of-e-commerce-products">E-commerce product structure</a></strong>
</p>

<p align="center">
  <img src="/images/e-commerce-system.png" width=800 height=1300>
</p>

### 应用架构

### 技术架构

### 数据架构

## 商品管理 Product Center

### 商品信息包括哪些内容
<p align="center">
  <img src="/images/item-info.png" width=700 height=700>
</p>

### 商品系统的演进

| 阶段         | 主要特征/能力                                                         | 技术架构/数据模型           | 适用场景/目标                         | 实现方式简单说明 |
|--------------|---------------------------------------------------------------------|-----------------------------|--------------------------------------|----------------|
| 初始阶段     | - 商品信息简单，字段少<br>- SKU/SPU未严格区分<br>- 价格库存直接在商品表<br>- 仅支持基本的增删改查 | 单表/简单表结构              | 小型电商、业务初期，SKU数量少         | 单体应用，单表存储，简单业务逻辑，直接数据库操作 |
| 成长阶段     | - 引入SPU/SKU模型<br>- 属性、类目、品牌等实体独立<br>- 支持多规格商品<br>- 价格库存可拆分为独立表 | 关系型数据库，ER模型优化      | SKU多样化，品类扩展，业务快速增长     | 关系型数据库，ER模型优化，多表存储，业务逻辑复杂 |
| 成熟阶段     | - 商品中台化，支持多业务线/多渠道<br>- 属性体系灵活可扩展<br>- 多级类目、标签、图片、描述等丰富<br>- 商品快照、操作日志、版本控制 | 中台架构，微服务/多表/NoSQL   | 大型平台，业务复杂，需支撑多业务场景   | 分布式服务，插件化/配置化流程，状态机驱动，异步消息，灵活数据模型 |
| 未来演进     | - 多语言多币种支持<br>- 商品内容多媒体化（视频、3D等）<br>- AI智能标签/推荐<br>- 商品数据实时分析与洞察 | 分布式/云原生/大数据平台      | 国际化、智能化、数据驱动的电商生态    | 云原生架构，AI/大数据分析，自动化运维，弹性伸缩，智能路由与风控 |

### 什么是SPU、SKU
方案一：同时创建多个SKU，并同步生成关联的SPU。整体方案是直接创建SKU，并维护多个不同的属性；该方案适用于大多数C2C综合电商平台（例如，阿里巴巴就是采用这种方式创建商品）。
方案二：先创建SPU，再根据SPU创建SKU。整体方案是由平台的主数据团队负责维护SPU，商家（包括自营和POP）根据SPU维护SKU。在创建SKU时，首先选择SPU（SPU中的基本属性由数据团队维护），然后基于SPU维护销售属性和物流属性，最后生成SKU；该方案适用于高度专业化的垂直B2B行业，如汽车、医药等。
这两种方案的原因是：垂直B2B平台上的业务（传统行业、年长的商家）操作能力有限，维护产品属性的错误率远高于C2C平台，同时平台对产品结构控制的要求较高。为了避免同一产品被不同商家维护成多个不同的属性（例如，汽车轮胎的胎面宽度、尺寸等属性），平台通常选择专门的数据团队来维护产品的基本属性，即维护SPU。
此外，B2B垂直电商的品类较少，SKU数量相对较小，品类标准化程度高，平台统一维护的可行性较高。
对于拥有成千上万品类的综合电商平台，依靠平台数据团队的统一维护是不现实的，或者像服装这样非标准化的品类对商品结构化管理的要求较低。因此，综合平台（阿里巴巴和亚马逊）的设计方向与垂直平台有所不同。
实际上，即使对于综合平台，不同的品类也会有不同的设计方法。一些品类具有垂直深度，因此也采用平台维护SPU和商家创建SKU的方式

### 数据库模型
- 类目category
- 品牌brand
- 属性attribute
- 标签tag
- 商品主表/spu表/item表、item_stat 统计表、item属性值表
- 商品变体表/variant表/sku表、sku attribute表
- 其它实体表、其它实体和商品表的关联表

<p align="center">
  <img src="/images/product_ER.png" width=700 height=700>
  <br/>
  <strong><a href="https://axureboutique.com/blogs/product-design/build-an-e-commerce-product-center-from-scratch">E-commerce product center</a></strong>
</p>

#### 模型说明：
- 商品（item/SPU）与商品变体（sku）分离，便于管理不同规格、价格、库存的商品。
- 属性（attribute）、类目（category）、品牌（brand）等实体独立，便于扩展和维护
- 商品分类体系如何设计？采用多级分类？分类的动态扩展只需插入新分类，指定其 parent_id，即可动态扩展任意层级
- 灵活的属性体系。通过 category_attribute 和 spu_attr_value 支持不同类目下的不同属性，适应多样化商品需求。属性值与商品解耦，支持动态扩展
- item_stat 单独存储统计信息，便于高并发下的读写优化。
- 可以方便地增加标签（tag）、图片、描述、规格等字段，适应业务变化

#### 商品信息录入JSON示例

##### 实体商品

<details>
<summary>1、实体商品男士T恤 JSON 数据</summary>
<pre><code class="json">
```json
{
  "categoryId": 1003001,
  "spu": {
    "name": "经典圆领男士T恤",
    "brandId": 2001,
    "description": "柔软舒适，100%纯棉"
  },
  "basicAttributes": [
    {
      "attributeId": 101,         // 品牌
      "attributeName": "品牌",
      "value": "NIKE"
    },
    {
      "attributeId": 102,         // 材质
      "attributeName": "材质",
      "value": "棉"
    },
    {
      "attributeId": 103,         // 产地
      "attributeName": "产地",
      "value": "中国"
    },
    {
      "attributeId": 104,         // 袖型
      "attributeName": "袖型",
      "value": "短袖"
    }
  ],
  "skus": [
    {
      "skuName": "黑色 L",
      "price": 79.00,
      "stock": 100,
      "salesAttributes": [
        {
          "attributeId": 201,
          "attributeName": "颜色",
          "value": "黑色"
        },
        {
          "attributeId": 202,
          "attributeName": "尺码",
          "value": "L"
        }
      ]
    },
    {
      "skuName": "白色 M",
      "price": 79.00,
      "stock": 150,
      "salesAttributes": [
        {
          "attributeId": 201,
          "attributeName": "颜色",
          "value": "白色"
        },
        {
          "attributeId": 202,
          "attributeName": "尺码",
          "value": "M"
        }
      ]
    }
  ]
}
```
</code></pre>
</details> 

##### 虚拟商品

<details>
<summary>2、虚拟商品流量充值 JSON 数据</summary>
<pre><code class="json">
```json
{
  "categoryId": 1005002,
  "spu": {
    "name": "中国移动流量包充值",
    "brandId": 3001,
    "description": "全国通用流量包充值，按需选择，自动到账"
  },
  "basicAttributes": [
    {
      "attributeId": 301,
      "attributeName": "运营商",
      "value": "中国移动"
    },
    {
      "attributeId": 302,
      "attributeName": "适用网络",
      "value": "4G/5G"
    }
  ],
  "skus": [
    {
      "skuName": "中国移动1GB全国流量包（7天）",
      "price": 5.00,
      "stock": 9999,
      "salesAttributes": [
        {
          "attributeId": 401,
          "attributeName": "流量容量",
          "value": "1GB"
        },
        {
          "attributeId": 402,
          "attributeName": "有效期",
          "value": "7天"
        },
        {
          "attributeId": 403,
          "attributeName": "流量类型",
          "value": "全国通用"
        }
      ]
    },
    {
      "skuName": "中国移动5GB全国流量包（30天）",
      "price": 20.00,
      "stock": 9999,
      "salesAttributes": [
        {
          "attributeId": 401,
          "attributeName": "流量容量",
          "value": "5GB"
        },
        {
          "attributeId": 402,
          "attributeName": "有效期",
          "value": "30天"
        },
        {
          "attributeId": 403,
          "attributeName": "流量类型",
          "value": "全国通用"
        }
      ]
    },
    {
      "skuName": "中国移动10GB全国流量包（90天）",
      "price": 38.00,
      "stock": 9999,
      "salesAttributes": [
        {
          "attributeId": 401,
          "attributeName": "流量容量",
          "value": "10GB"
        },
        {
          "attributeId": 402,
          "attributeName": "有效期",
          "value": "90天"
        },
        {
          "attributeId": 403,
          "attributeName": "流量类型",
          "value": "全国通用"
        }
      ]
    }
  ]
}

```
</code></pre>
</details> 

#### 商品的价格和库存

##### 方案1. 价格和库存直接放在sku表中 （变化小）
在这种方案中，SKU（Stock Keeping Unit） 表包含商品的所有信息，包括价格和库存数量。每个 SKU 记录一个独立的商品实例，它有唯一的标识符，直接关联价格和库存。
```sql
CREATE TABLE sku_tab (
    sku_id INT PRIMARY KEY,             -- SKU ID
    product_id INT,                     -- 商品ID (外键，指向商品表)
    sku_name VARCHAR(255),              -- SKU 名称
    original_price DECIMAL(10, 2),      -- 原始价格
    price DECIMAL(10, 2),               -- 销售价格
    discount_price DECIMAL(10, 2),      -- 折扣价格（如果有）
    stock_quantity INT,                 -- 库存数量
    warehouse_id INT,                   -- 仓库ID（如果有多个仓库）
    created_at TIMESTAMP,               -- 创建时间
    updated_at TIMESTAMP                -- 更新时间
);
```

优点：
- 简单：所有信息都集中在一个表中，查询和管理都很方便。
- 查询效率：查询某个商品的价格和库存不需要多表联接，减少了数据库查询的复杂度。
- 维护方便：商品的所有信息（包括价格和库存）都在一个地方，减少了冗余数据和数据不一致的可能性。

缺点：
- 灵活性差：如果价格和库存的管理策略较复杂（如促销、库存管理、动态定价等），这种方式可能不太适用。修改价格或库存时需要直接更新 SKU 表。
- 扩展性差：对于一些复杂的定价和库存管理需求（如多层次的定价结构、分仓库管理等），直接放在 SKU 表中可能不够灵活。

适用场景：
- 商品种类较少，SKU 数量相对固定且不复杂的场景。
- 价格和库存变动较少，不涉及复杂的促销或动态定价的场景


##### 方案2. 价格和库存单独管理（变化大）

```sql

CREATE TABLE price_tab (
    price_id INT PRIMARY KEY,         -- 价格ID
    sku_id INT,                       -- SKU ID (外键)
    price DECIMAL(10, 2),             -- 商品价格
    discount_price DECIMAL(10, 2),    -- 折扣价格
    effective_date TIMESTAMP,         -- 价格生效时间
    expiry_date TIMESTAMP,            -- 价格失效时间
    price_type VARCHAR(50),           -- 价格类型（如标准价、促销价等）
    FOREIGN KEY (sku_id) REFERENCES ProductSKUs(sku_id)
);

CREATE TABLE inventory_tab (
    inventory_id INT PRIMARY KEY,     -- 库存ID
    sku_id INT,                        -- SKU ID (外键)
    quantity INT,                      -- 库存数量
    warehouse_id INT,                  -- 仓库ID（如果有多个仓库）
    updated_at TIMESTAMP,              -- 库存更新时间
    FOREIGN KEY (sku_id) REFERENCES ProductSKUs(sku_id)
);
```
优点：
- 灵活性高：价格和库存信息可以独立管理，更容易支持多样化的定价策略、促销活动、库存管理等。
- 可扩展性强：对于需要频繁更新价格、库存、促销等信息的商品，这种方案更容易扩展和适应变化。例如，可以灵活地增加新的价格策略或库存仓库。
- 数据结构清晰：避免了价格和库存在 SKU 表中的冗余存储，使得数据结构更清晰。

缺点：
- 查询复杂：获取某个商品的价格和库存信息时，需要联接多个表，查询效率可能会降低，尤其是在数据量大时。
- 管理复杂：需要更多的表和关系，增加了维护成本和系统复杂度。

适用场景：
- 商品种类繁多，SKU 数量较大，且需要支持动态定价、促销、库存管理等复杂需求的场景。
- 需要频繁变动价格或库存的商品，且这些信息与 SKU 无法紧密绑定的场景


#### 商品快照 item_snapshots
1. 商品编辑时生成快照:
- 每次商品信息（如价格、描述、属性等）发生编辑时，生成一个新的商品快照。
- 将快照信息存储在 item_snapshots 表中，并生成一个唯一的 snapshot_id。

2. 订单创建时使用快照:
在用户下单时，查找当前商品的最新 snapshot_id。
在 order_items 表中记录该 snapshot_id，以确保订单项反映下单时的商品状态
```sql
  CREATE TABLE `snapshot_tab` (
    `snapshot_id` int(11) NOT NULL AUTO_INCREMENT,
    `snapshot_type` int(11) NOT NULL, 
    `create_time` int(11) NOT NULL DEFAULT '0',
    `data` text NOT NULL,
    `entity_id` int(11) DEFAULT NULL,
    PRIMARY KEY (`snapshot_id`),
    KEY `idx_entity_id` (`entity_id`)
  ) 
```

#### 用户操作日志
```sql
CREATE TABLE user_operation_logs (
  log_id INT PRIMARY KEY AUTO_INCREMENT,  -- Unique identifier for each log entry
  user_id INT NOT NULL,                   -- ID of the user who made the edit
  entity_id INT NOT NULL,                 -- ID of the entity being edited
  entity_type VARCHAR(50) NOT NULL,       -- Type of entity (e.g., SPU, SKU, Price, Stock)
  operation_type VARCHAR(50) NOT NULL,    -- Type of operation (e.g., CREATE, UPDATE, DELETE)
  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- Time of the operation
  details TEXT,                           -- Additional details about the operation
  FOREIGN KEY (user_id) REFERENCES users(id)  -- Assuming a users table exists
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```
#### 商品的统计信息



### 缓存的使用
### 核心流程
#### B端：商品创建和发布的流程
- 批量上传、批量编辑
- 单个上传、编辑
- 审核、发布
- OpenAPI，支持外部同步API push 商品
- auto-sync，自动同步外部商品
#### C端：商品搜索、商品详情
- 商品搜索
  - elastic search 索引构建。获取商品列表(首页索引)
  - 如何处理商品的SEO优化？
<details>
<summary>1、item index</summary>
<pre><code class="json">
```json
POST /products/_doc/1
{
  "product_id": "123456",
  "name": "Wireless Bluetooth Headphones",
  "description": "High-quality wireless headphones with noise-cancellation.",
  "price": 99.99,
  "stock": 50,
  "category": "Electronics",
  "brand": "SoundMax",
  "sku": "SM-123",
  "spu": "SPU-456",
  "image_urls": [
    "http://example.com/images/headphones1.jpg",
    "http://example.com/images/headphones2.jpg"
  ],
  "ratings": 4.5,
  "seller_info": {
    "seller_id": "78910",
    "seller_name": "BestSeller"
  },
  "attributes": {
    "color": "Black",
    "size": "Standard",
    "material": "Plastic"
  },
  "release_date": "2023-01-15",
  "location": {
    "lat": 40.7128,
    "lon": -74.0060
  },
  "tags": ["headphones", "bluetooth", "wireless"],
  "promotional_info": "20% off for a limited time"
}
```
</code></pre>
</details> 
- 商品推荐
  - 商品的A/B测试如何设计？
  - 如何设计商品的推荐算法？
  - 商品的个性化定制如何实现？
- 获取商品详情


## 订单管理 Order Center
[订单系统，平台的"生命中轴线"](https://www.woshipm.com/pd/753646.html)
### 订单中需要包含哪些信息
<p align="center">
  <img src="/images/order_content.webp" width=800 height=500>
</p>

### 常见的订单类型
1. 实物订单
典型场景：电商平台购物（如买衣服、家电）
核心特征：
需要物流配送，涉及收货地址、运费、物流跟踪
需要库存校验与扣减
售后流程（退货、换货、退款）复杂
订单状态多（待发货、已发货、已收货等）

2. 虚拟订单
典型场景：会员卡、电子券、游戏点卡、电影票等
核心特征：
无物流配送，不需要收货地址和运费
通常无需库存（或库存为虚拟库存）
订单完成后直接发放虚拟物品或凭证
售后流程简单或无售后

3. 预售订单
典型场景：新品预售、定金膨胀、众筹等
核心特征：
订单分为定金和尾款两阶段
需校验定金支付、尾款支付的时效
可能涉及定金不退、尾款未付订单自动关闭等规则
发货时间通常在尾款支付后

4. O2O订单，外卖订单
典型场景：酒店预订
核心特征：
需选择入住/离店日期、房型、入住人信息
需对接第三方酒店系统实时查房、锁房
取消、变更政策复杂，可能涉及违约金
无物流，但有电子凭证或入住确认

### 订单系统的演进

| 阶段         | 主要特征/能力                                                                 | 技术架构/数据模型           | 适用场景/目标                         | 实现方式简单说明 |
|--------------|----------------------------------------------------------------------------|-----------------------------|--------------------------------------|----------------|
| 初始阶段     | - 实现订单基本流转（下单、支付、发货、收货、取消）<br>- 单一订单类型（实物订单）<br>- 订单与商品、用户简单关联 | 单体应用/单表或少量表结构    | 业务初期，订单量小，流程简单，SKU/商家数量有限 | 单体应用，单表存储，简单业务逻辑，直接数据库操作 |
| 成长阶段（订单中心） | - 支持订单拆单、合单（如多仓发货、合并支付）<br>- 支持多品类订单（如实物+虚拟）<br>- 订单中心化，订单与支付、配送、售后等子系统解耦<br>- 订单与商品快照、操作日志关联 | 微服务/多表/订单中心架构         | 平台型电商，业务扩展，需支持多商家、多类型订单，订单量大幅增长 | 订单中心服务，微服务拆分，多表关联，服务间接口调用，快照与日志表设计 |
| 成熟期（平台化）   | - 支持多样化订单类型（预售、虚拟、O2O、定制、JIT等）<br>- 订单流程可配置/插件化/工作流引擎/状态机框架/规则引擎等<br>- 订单状态机、履约、支付、退款等子流程解耦<br>- 支持复杂的促销、分账、履约模式 | 分布式/服务化/灵活数据模型    | 大型/综合电商，业务复杂，需快速适应新业务模式和高并发场景 | 分布式服务，插件化/配置化流程，状态机驱动，异步消息，灵活数据模型 |
| 未来智能化   | - 订单智能路由与分配（如智能分仓、智能客服）<br>- 实时风控与反欺诈<br>- 订单数据实时分析与洞察<br>- 高可用、弹性伸缩、自动化运维 | 云原生/大数据/AI驱动架构      | 超大规模平台，国际化、智能化、数据驱动，需极致稳定与创新能力 | 云原生架构，AI/大数据分析，自动化运维，弹性伸缩，智能路由与风控 |

### 常见的订单模型设计
<p align="center">
  <img src="/images/order_er.png" width=800 height=600>
</p>

#### 订单表（order_tab）：记录用户的购买订单信息。主键为 order_id。
  - pay_order_id：支付订单ID，作为外键关联支付订单。
  - user_id：用户ID，标识购买订单的用户。
  - total_amount：订单的总金额。
  - order_status：订单状态，如已完成、已取消等。
  - payment_status：支付状态，与支付订单相关。
  - fulfillment_status：履约状态，表示订单的配送或服务状态。
  - refund_status：退款状态，用于标识订单是否有退款

#### 订单商品表（order_item_tab：记录订单中具体商品的信息。主键为 order_item_id。
  - order_id：订单ID，作为外键关联订单。
  - item_id：商品ID，表示订单中的商品。
  - item_snapshot_id：商品快照ID，记录当时购买时的商品信息快照。
  - item_status：商品状态，如已发货、退货等。
  - quantity：购买数量。
  - price：商品单价。
  - discount：商品折扣金额

#### 订单支付表（pay_order_tab）：主要用于记录用户的支付信息。主键为 pay_order_id，标识唯一的支付订单。
  - user_id：用户ID，标识支付的用户。
  - payment_method：支付方式，如信用卡、支付宝等。
  - payment_status：支付状态，如已支付、未支付等。
  - pay_amount、cash_amount、coin_amount、voucher_amount：支付金额、现金支付金额、代币支付金额、优惠券使用金额。
  - 时间戳字段包括创建时间、初始化时间和更新时间

#### 退款表（refund_tab）：记录订单或订单项的退款信息。主键为 refund_id。
  - order_id：订单ID，作为外键关联订单。
  - order_item_id：订单项ID，标识具体商品的退款。
  - refund_amount：退款金额。
  - reason：退款原因。
  - quantity：退款的商品数量。
  - refund_status：退款状态。
  - refund_time：退款操作时间。

#### 实体间关系：
- 支付订单与订单：
- 一个支付订单可能关联多个购买订单，形成 一对多 关系。
例如，用户可以通过一次支付购买多个不同的订单。
- 订单与订单商品：
一个订单可以包含多个订单项，形成 一对多 关系。
订单项代表订单中所购买的每个商品的详细信息。
- 订单与退款：
  - 一个订单可能包含多个退款，形成 一对多 关系。
  - 退款可以是针对订单整体，也可以针对订单中的某个商品


### 订单状态机设计


#### Order 主状态机
<p align="center">
  <img src="/images/order_status.png" width=500 height=450>
</p>


#### 支付状态机
<p align="center">
  <img src="/images/pay_status.png" width=700 height=400>
</p>


#### 履约状态机
<p align="center">
  <img src="/images/fulfillment_status.png" width=500 height=400>
</p>

#### 退货退款状体机
<p align="center">
  <img src="/images/refund_status.png" width=700 height=1100>
</p>


#### 异常单人工介入
- 用户发起退款单拒绝
- 退货失败，订单状态无法流转
- 退款失败
- 退营销失败


### 订单ID 生成策略
``` python 
  # 时间戳 + 机器id + uid % 1000 + 自增序号

import time
import threading
from typing import Union

class OrderNoGenerator:
    def __init__(self, machine_id: int):
        """
        初始化订单号生成器
        :param machine_id: 机器ID (0-999)
        """
        if not 0 <= machine_id <= 999:
            raise ValueError("机器ID必须在0-999之间")
        
        self.machine_id = machine_id
        self.sequence = 0
        self.last_timestamp = -1
        self.lock = threading.Lock()  # 线程锁，保证线程安全

    def _wait_next_second(self, last_timestamp: int) -> int:
        """
        等待下一秒
        :param last_timestamp: 上次时间戳
        :return: 新的时间戳
        """
        timestamp = int(time.time())
        while timestamp <= last_timestamp:
            timestamp = int(time.time())
        return timestamp

    def generate_order_no(self, user_id: int) -> Union[int, str]:
        """
        生成订单号
        :param user_id: 用户ID
        :return: 订单号（整数或字符串形式）
        """
        with self.lock:  # 使用线程锁保证线程安全
            # 获取当前时间戳（秒级）
            timestamp = int(time.time())
            
            # 处理时间回拨
            if timestamp < self.last_timestamp:
                raise RuntimeError("系统时间回拨，拒绝生成订单号")
            
            # 如果是同一秒，序列号自增
            if timestamp == self.last_timestamp:
                self.sequence = (self.sequence + 1) % 1000
                # 如果序列号用完了，等待下一秒
                if self.sequence == 0:
                    timestamp = self._wait_next_second(self.last_timestamp)
            else:
                # 不同秒，序列号重置
                self.sequence = 0
            
            self.last_timestamp = timestamp
            
            # 获取用户ID的后3位
            user_id_suffix = user_id % 1000
            
            # 组装订单号
            order_no = (timestamp * 1000000000 +  # 时间戳左移9位
                       self.machine_id * 1000000 +  # 机器ID左移6位
                       user_id_suffix * 1000 +      # 用户ID左移3位
                       self.sequence)               # 序列号
            
            return order_no

    def generate_order_no_str(self, user_id: int) -> str:
        """
        生成字符串形式的订单号
        :param user_id: 用户ID
        :return: 字符串形式的订单号
        """
        order_no = self.generate_order_no(user_id)
        return f"{order_no:019d}"  # 补零到19位

# 使用示例
def main():
    # 创建订单号生成器实例
    generator = OrderNoGenerator(machine_id=1)
    
    # 生成订单号
    user_id = 12345
    order_no = generator.generate_order_no(user_id)
    order_no_str = generator.generate_order_no_str(user_id)
    
    print(f"整数形式订单号: {order_no}")
    print(f"字符串形式订单号: {order_no_str}")
    
    # 测试并发
    def test_concurrent():
        for _ in range(5):
            order_no = generator.generate_order_no(user_id)
            print(f"并发生成的订单号: {order_no}")
    
    # 创建多个线程测试并发
    threads = []
    for _ in range(3):
        t = threading.Thread(target=test_concurrent)
        threads.append(t)
        t.start()
    
    # 等待所有线程完成
    for t in threads:
        t.join()

if __name__ == "__main__":
    main()
```

### 订单商品快照

#### 方案1. 直接使用商品系统的item snapshot。（由商品系统维护快照）
- 商品系统负责维护商品快照
- 订单系统通过引用商品快照ID来关联商品信息
- 商品信息变更时，商品系统生成新的快照版本
```sql
-- 商品系统维护的快照表
CREATE TABLE item_snapshot_tab (
    snapshot_id BIGINT PRIMARY KEY,
    item_id BIGINT NOT NULL,
    version INT NOT NULL,
    data JSON NOT NULL,  -- 存储商品完整信息
    created_at TIMESTAMP NOT NULL,
    INDEX idx_item_version (item_id, version)
);

-- 订单系统引用快照
CREATE TABLE order_item_tab (
    order_id BIGINT,
    item_id BIGINT,
    snapshot_id BIGINT,  -- 引用商品快照
    quantity INT,
    price DECIMAL(10,2),
    FOREIGN KEY (snapshot_id) REFERENCES item_snapshot(snapshot_id)
);
```
优点
- 数据一致性高：商品系统统一管理快照，避免数据不一致
- 存储效率高：多个订单可以共享同一个快照版本
- 维护成本低：订单系统不需要关心快照的生成和管理
- 查询性能好：可以直接通过快照ID获取完整商品信息

缺点
- 系统耦合度高：订单系统强依赖商品系统的快照服务
- 扩展性受限：商品系统需要支持所有订单系统可能需要的商品信息
- 版本管理复杂：需要处理快照的版本控制和清理
- 跨系统调用：订单系统需要调用商品系统获取快照信息

#### 方案2. 创单时提供商品详情信息。（由订单维护商品快照)
```sql
CREATE TABLE order_item (
    order_id BIGINT,
    item_id BIGINT,
    quantity INT,
    price DECIMAL(10,2),
    snapshot_data JSON NOT NULL,  -- 存储下单时的商品信息
    FOREIGN KEY (order_id, item_id) REFERENCES order_item_snapshot(order_id, item_id)
);
```

#### 方案3. 创单时提供商品详情信息。（由订单维护商品快照）+ 快照复用
设计思路：
- 订单系统维护自己的快照表，但增加快照复用机制
- 使用商品信息的摘要(摘要算法如MD5)作为快照的唯一标识
- 相同摘要的商品信息共享同一个快照记录
- 创单时先检查摘要是否存在，存在则复用，不存在则创建新快照

```sql
-- 订单系统维护的快照表
CREATE TABLE order_item_snapshot (
    snapshot_id BIGINT PRIMARY KEY,
    item_id BIGINT NOT NULL,
    item_hash VARCHAR(32) NOT NULL COMMENT '商品信息摘要',
    snapshot_data JSON NOT NULL COMMENT '存储下单时的商品信息',
    created_at TIMESTAMP NOT NULL,
    INDEX idx_item_hash (item_hash),
    INDEX idx_item_id (item_id)
);
-- 订单商品表
CREATE TABLE order_item (
    order_id BIGINT,
    item_id BIGINT,
    snapshot_id BIGINT,
    quantity INT,
    price DECIMAL(10,2),
    FOREIGN KEY (snapshot_id) REFERENCES order_item_snapshot(snapshot_id)
);
```
适用场景：
- 商品模型比较固定，项目初期，团队比较小，能接受系统之间的耦合，可以考虑用1
- 不同商品差异比较大，商品信息结构复杂，考虑用2
- 订单量太大，考虑复用快照
### 核心流程
#### 正常流程和逆向流程
#### 创单
##### 核心步骤
1. 参数校验。用户校验，是否异常用户。
2. 商品与价格校验。校验商品是否存在、是否上架、价格是否有效
3. 库存校验与预占。检查库存是否充足，部分场景下进行库存预占（锁库存）。
4. 营销信息校验。校验优惠券、积分等是否可用，计算优惠金额。
6. 订单金额计算。计算订单总金额、应付金额、各项明细。
7. 生成订单号。生成全局唯一订单号，保证幂等性。
8. 订单数据落库。写入订单主表、订单明细表、扩展表等。
9. 扣减库存、扣减实际库存（有的系统在支付后扣减）。
10. 发送消息/异步处理。发送订单创建成功消息，通知库存、物流、营销等系统。
11. 返回下单结果。返回订单号、支付信息等给前端。

##### 实现思路
- 接口定义：通过OrderCreationStep接口定义了每个步骤必须实现的方法
- 上下文共享：使用OrderCreationContext在步骤间共享数据
- 步骤独立：每个步骤都是独立的，便于维护和测试
- 回滚机制：每个步骤都实现了回滚方法
- 流程管理：通过OrderCreationManager统一管理步骤的执行和回滚
- 错误处理：统一的错误处理和回滚机制
- 可扩展性：易于添加新的步骤或修改现有步骤
- 如何解决不同category 创单差异较大的问题？
  - 插件化/策略模式。将订单处理流程拆分为多个步骤（如校验、支付、通知等）。不同订单类型实现各自的处理逻辑，通过策略模式动态选择。
  2. 订单类型标识。在订单主表中增加订单类型字段，根据类型选择不同的处理流程。
  3. 扩展字段。使用JSON或扩展表存储特定订单类型的特殊字段（如酒店的入住日期、机票的航班信息）。
  4. 流程引擎。使用流程引擎（如BPMN）定义和管理复杂的订单处理流程，支持动态调整。

<details>
<summary>点击查看创单核心逻辑代码实现</summary>
<pre><code class="go">

``` Go
package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
)

// OrderType 订单类型
type OrderType string

const (
	OrderTypePhysical OrderType = "physical" // 实物订单
	OrderTypeVirtual  OrderType = "virtual"  // 虚拟订单
	OrderTypePresale  OrderType = "presale"  // 预售订单
	OrderTypeHotel    OrderType = "hotel"    // 酒店订单
	OrderTypeTopUp    OrderType = "topup"    // 充值订单
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusInit     OrderStatus = "init"     // 初始化
	OrderStatusPending  OrderStatus = "pending"  // 待支付
	OrderStatusPaid     OrderStatus = "paid"     // 已支付
	OrderStatusShipping OrderStatus = "shipping" // 发货中
	OrderStatusSuccess  OrderStatus = "success"  // 成功
	OrderStatusFailed   OrderStatus = "failed"   // 失败
	OrderStatusCanceled OrderStatus = "canceled" // 已取消
)

// Order 订单基础信息
type Order struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Type      OrderType       `json:"type"`
	Status    OrderStatus     `json:"status"`
	Amount    float64         `json:"amount"`
	Detail    json.RawMessage `json:"detail"` // 不同类型订单的特殊字段
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// OrderCreationContext 创单上下文
type OrderCreationContext struct {
	Ctx                 context.Context
	Order               *Order
	Params              map[string]interface{} // 创单参数
	Cache               map[string]interface{} // 步骤间共享数据
	Errors              []error                // 错误记录
	StepResults         map[string]StepResult  // 每个步骤的执行结果
	RollbackFailedSteps []string               // 记录回滚失败的步骤
}

// StepResult 步骤执行结果
type StepResult struct {
	Success        bool
	Error          error
	Data           interface{}
	CompensateData interface{} // 用于补偿的数据
}

// OrderCreationStep 创单步骤接口
type OrderCreationStep interface {
	Execute(ctx *OrderCreationContext) error
	Rollback(ctx *OrderCreationContext) error
	Compensate(ctx *OrderCreationContext) error // 异步补偿
	Name() string
}

// 错误定义
var (
	ErrInvalidParams     = errors.New("invalid parameters")
	ErrProductNotFound   = errors.New("product not found")
	ErrProductOffline    = errors.New("product is offline")
	ErrStockInsufficient = errors.New("stock insufficient")
	ErrUserBlocked       = errors.New("user is blocked")
	ErrSystemBusy        = errors.New("system is busy")
)

// OrderError 订单错误
type OrderError struct {
	Step    string
	Message string
	Err     error
}

func (e *OrderError) Error() string {
	return fmt.Sprintf("step: %s, message: %s, error: %v", e.Step, e.Message, e.Err)
}

// 参数校验步骤
type ParamValidationStep struct{}

func (s *ParamValidationStep) Execute(ctx *OrderCreationContext) error {
	// 通用参数校验
	if ctx.Order.UserID == "" || ctx.Order.Type == "" {
		return &OrderError{Step: s.Name(), Message: "missing required fields", Err: ErrInvalidParams}
	}

	// 订单类型特殊参数校验
	switch ctx.Order.Type {
	case OrderTypePhysical:
		if addr, ok := ctx.Params["address"].(string); !ok || addr == "" {
			return &OrderError{Step: s.Name(), Message: "missing address for physical order", Err: ErrInvalidParams}
		}
	case OrderTypeHotel:
		if _, ok := ctx.Params["check_in_date"].(time.Time); !ok {
			return &OrderError{Step: s.Name(), Message: "missing check-in date for hotel order", Err: ErrInvalidParams}
		}
	}
	return nil
}

func (s *ParamValidationStep) Rollback(ctx *OrderCreationContext) error {
	// 参数校验步骤无需回滚
	return nil
}

func (s *ParamValidationStep) Compensate(ctx *OrderCreationContext) error {
	// 参数校验步骤无需补偿
	return nil
}

func (s *ParamValidationStep) Name() string {
	return "param_validation"
}

// Product 商品信息
type Product struct {
	ID       string
	Name     string
	Price    float64
	IsOnSale bool
}

// ProductService 商品服务接口
type ProductService interface {
	GetProduct(ctx context.Context, productID string) (*Product, error)
}

// StockService 库存服务接口
type StockService interface {
	LockStock(ctx context.Context, productID string, quantity int) (string, error)
	UnlockStock(ctx context.Context, lockID string) error
	DeductStock(ctx context.Context, productID string, quantity int) error
	RevertDeductStock(ctx context.Context, productID string, quantity int) error
}

// PromotionService 营销服务接口
type PromotionService interface {
	ValidateCoupon(ctx context.Context, couponCode string, userID string, orderAmount float64) (*Coupon, error)
	UseCoupon(ctx context.Context, couponCode string, userID string, orderID string) error
	RevertCouponUsage(ctx context.Context, couponCode string, userID string, orderID string) error
	DeductPoints(ctx context.Context, userID string, points int) error
	RevertPointsDeduction(ctx context.Context, userID string, points int) error
}

// Coupon 优惠券信息
type Coupon struct {
	Code       string
	Type       string
	Amount     float64
	Threshold  float64
	ExpireTime time.Time
}

// 商品校验步骤
type ProductValidationStep struct {
	productService ProductService
}

func (s *ProductValidationStep) Execute(ctx *OrderCreationContext) error {
	productID := ctx.Params["product_id"].(string)
	product, err := s.productService.GetProduct(ctx.Ctx, productID)
	if err != nil {
		return &OrderError{Step: s.Name(), Message: "failed to get product", Err: err}
	}

	if !product.IsOnSale {
		return &OrderError{Step: s.Name(), Message: "product is offline", Err: ErrProductOffline}
	}

	ctx.Cache["product"] = product
	return nil
}

func (s *ProductValidationStep) Rollback(ctx *OrderCreationContext) error {
	return nil
}

func (s *ProductValidationStep) Compensate(ctx *OrderCreationContext) error {
	return nil
}

func (s *ProductValidationStep) Name() string {
	return "product_validation"
}

// 库存校验步骤
type StockValidationStep struct {
	stockService StockService
}

func (s *StockValidationStep) Execute(ctx *OrderCreationContext) error {
	if ctx.Order.Type == OrderTypeVirtual || ctx.Order.Type == OrderTypeTopUp {
		return nil
	}

	productID := ctx.Params["product_id"].(string)
	quantity := ctx.Params["quantity"].(int)

	lockID, err := s.stockService.LockStock(ctx.Ctx, productID, quantity)
	if err != nil {
		return &OrderError{Step: s.Name(), Message: "failed to lock stock", Err: err}
	}

	ctx.Cache["stock_lock_id"] = lockID
	return nil
}

func (s *StockValidationStep) Rollback(ctx *OrderCreationContext) error {
	if lockID, ok := ctx.Cache["stock_lock_id"].(string); ok {
		return s.stockService.UnlockStock(ctx.Ctx, lockID)
	}
	return nil
}

func (s *StockValidationStep) Compensate(ctx *OrderCreationContext) error {
	return nil
}

func (s *StockValidationStep) Name() string {
	return "stock_validation"
}

// 库存扣减步骤
type StockDeductionStep struct {
	stockService StockService
}

func (s *StockDeductionStep) Execute(ctx *OrderCreationContext) error {
	// 虚拟商品和充值订单跳过库存扣减
	if ctx.Order.Type == OrderTypeVirtual || ctx.Order.Type == OrderTypeTopUp {
		return nil
	}

	productID := ctx.Params["product_id"].(string)
	quantity := ctx.Params["quantity"].(int)

	// 执行库存扣减
	if err := s.stockService.DeductStock(ctx.Ctx, productID, quantity); err != nil {
		return &OrderError{
			Step:    s.Name(),
			Message: "failed to deduct stock",
			Err:     err,
		}
	}

	// 记录扣减信息，用于回滚
	ctx.Cache["stock_deducted"] = map[string]interface{}{
		"product_id": productID,
		"quantity":   quantity,
	}

	return nil
}

func (s *StockDeductionStep) Rollback(ctx *OrderCreationContext) error {
	deducted, ok := ctx.Cache["stock_deducted"].(map[string]interface{})
	if !ok {
		return nil
	}

	productID := deducted["product_id"].(string)
	quantity := deducted["quantity"].(int)

	return s.stockService.RevertDeductStock(ctx.Ctx, productID, quantity)
}

func (s *StockDeductionStep) Compensate(ctx *OrderCreationContext) error {
	deducted, ok := ctx.Cache["stock_deducted"].(map[string]interface{})
	if !ok {
		return nil
	}

	productID := deducted["product_id"].(string)
	quantity := deducted["quantity"].(int)

	// 创建补偿消息
	compensationMsg := StockCompensationMessage{
		OrderID:   ctx.Order.ID,
		ProductID: productID,
		Quantity:  quantity,
		Timestamp: time.Now(),
	}

	// TODO: 实现发送到补偿队列的逻辑
	// return sendToCompensationQueue("stock_compensation", compensationMsg)
	return nil
}

func (s *StockDeductionStep) Name() string {
	return "stock_deduction"
}

// 营销活动扣减步骤
type PromotionDeductionStep struct {
	promotionService PromotionService
}

func (s *PromotionDeductionStep) Execute(ctx *OrderCreationContext) error {
	// 处理优惠券
	if couponCode, ok := ctx.Params["coupon_code"].(string); ok {
		// 验证优惠券
		coupon, err := s.promotionService.ValidateCoupon(
			ctx.Ctx,
			couponCode,
			ctx.Order.UserID,
			ctx.Order.Amount,
		)
		if err != nil {
			return &OrderError{
				Step:    s.Name(),
				Message: "invalid coupon",
				Err:     err,
			}
		}

		// 使用优惠券
		if err := s.promotionService.UseCoupon(ctx.Ctx, couponCode, ctx.Order.UserID, ctx.Order.ID); err != nil {
			return &OrderError{
				Step:    s.Name(),
				Message: "failed to use coupon",
				Err:     err,
			}
		}

		// 记录优惠券使用信息，用于回滚
		ctx.Cache["used_coupon"] = couponCode

		// 更新订单金额
		ctx.Order.Amount -= coupon.Amount
	}

	// 处理积分抵扣
	if points, ok := ctx.Params["use_points"].(int); ok && points > 0 {
		// 扣减积分
		if err := s.promotionService.DeductPoints(ctx.Ctx, ctx.Order.UserID, points); err != nil {
			return &OrderError{
				Step:    s.Name(),
				Message: "failed to deduct points",
				Err:     err,
			}
		}

		// 记录积分扣减信息，用于回滚
		ctx.Cache["deducted_points"] = points

		// 更新订单金额（假设1积分=0.01元）
		ctx.Order.Amount -= float64(points) * 0.01
	}

	return nil
}

func (s *PromotionDeductionStep) Rollback(ctx *OrderCreationContext) error {
	// 回滚优惠券使用
	if couponCode, ok := ctx.Cache["used_coupon"].(string); ok {
		if err := s.promotionService.RevertCouponUsage(ctx.Ctx, couponCode, ctx.Order.UserID, ctx.Order.ID); err != nil {
			return err
		}
	}

	// 回滚积分扣减
	if points, ok := ctx.Cache["deducted_points"].(int); ok {
		if err := s.promotionService.RevertPointsDeduction(ctx.Ctx, ctx.Order.UserID, points); err != nil {
			return err
		}
	}

	return nil
}

func (s *PromotionDeductionStep) Compensate(ctx *OrderCreationContext) error {
	// 优惠券补偿
	if couponCode, ok := ctx.Cache["used_coupon"].(string); ok {
		// TODO: 实现优惠券补偿逻辑
		// 1. 发送到补偿队列
		// 2. 记录补偿日志
		// 3. 通知运营人员
	}

	// 积分补偿
	if points, ok := ctx.Cache["deducted_points"].(int); ok {
		// TODO: 实现积分补偿逻辑
		// 1. 发送到补偿队列
		// 2. 记录补偿日志
		// 3. 通知运营人员
	}

	return nil
}

func (s *PromotionDeductionStep) Name() string {
	return "promotion_deduction"
}

// OrderFactory 订单工厂
type OrderFactory struct {
	commonSteps []OrderCreationStep
	typeSteps   map[OrderType][]OrderCreationStep
}

func NewOrderFactory() *OrderFactory {
	f := &OrderFactory{
		commonSteps: []OrderCreationStep{
			&ParamValidationStep{},
			&ProductValidationStep{},
			&PromotionDeductionStep{},
		},
		typeSteps: make(map[OrderType][]OrderCreationStep),
	}

	// 实物订单特有步骤
	f.typeSteps[OrderTypePhysical] = []OrderCreationStep{
		&StockValidationStep{},
		&StockDeductionStep{},
	}

	// 虚拟订单特有步骤
	f.typeSteps[OrderTypeVirtual] = []OrderCreationStep{}

	// 预售订单特有步骤
	f.typeSteps[OrderTypePresale] = []OrderCreationStep{
		&StockValidationStep{},
	}

	// 酒店订单特有步骤
	f.typeSteps[OrderTypeHotel] = []OrderCreationStep{}

	return f
}

func (f *OrderFactory) GetSteps(orderType OrderType) []OrderCreationStep {
	steps := make([]OrderCreationStep, 0)
	steps = append(steps, f.commonSteps...)
	if typeSteps, ok := f.typeSteps[orderType]; ok {
		steps = append(steps, typeSteps...)
	}
	return steps
}

// Logger 日志接口
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// OrderCreationManager 订单创建管理器
type OrderCreationManager struct {
	factory *OrderFactory
	logger  Logger
}

func (m *OrderCreationManager) CreateOrder(ctx context.Context, params map[string]interface{}) (*Order, error) {
	orderCtx := &OrderCreationContext{
		Ctx:                 ctx,
		Params:              params,
		Cache:               make(map[string]interface{}),
		StepResults:         make(map[string]StepResult),
		RollbackFailedSteps: make([]string, 0),
	}

	// 初始化订单
	order := &Order{
		ID:        generateOrderID(),
		UserID:    params["user_id"].(string),
		Type:      OrderType(params["type"].(string)),
		Status:    OrderStatusInit,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	orderCtx.Order = order

	// 获取订单类型对应的处理步骤
	steps := m.factory.GetSteps(order.Type)

	// 执行步骤
	executedSteps := make([]OrderCreationStep, 0)
	for _, step := range steps {
		stepName := step.Name()
		m.logger.Info("executing step", "step", stepName)

		err := step.Execute(orderCtx)
		if err != nil {
			m.logger.Error("step execution failed", "step", stepName, "error", err)

			orderCtx.Errors = append(orderCtx.Errors, err)

			// 执行回滚，并记录回滚失败的步骤
			m.rollbackSteps(orderCtx, executedSteps)

			// 只对回滚失败的步骤进行补偿
			if len(orderCtx.RollbackFailedSteps) > 0 {
				go m.compensateFailedRollbacks(orderCtx)
			}

			return nil, err
		}

		executedSteps = append(executedSteps, step)
		m.logger.Info("step executed successfully", "step", stepName)
	}

	return order, nil
}

// 修改回滚逻辑，记录回滚失败的步骤
func (m *OrderCreationManager) rollbackSteps(ctx *OrderCreationContext, steps []OrderCreationStep) {
	for i := len(steps) - 1; i >= 0; i-- {
		step := steps[i]
		stepName := step.Name()

		if err := step.Rollback(ctx); err != nil {
			m.logger.Error("step rollback failed", "step", stepName, "error", err)
			// 记录回滚失败的步骤
			ctx.RollbackFailedSteps = append(ctx.RollbackFailedSteps, stepName)
		}
	}
}

// 新的补偿方法，只处理回滚失败的步骤
func (m *OrderCreationManager) compensateFailedRollbacks(ctx *OrderCreationContext) {
	m.logger.Info("starting compensation for failed rollbacks",
		"failed_steps", ctx.RollbackFailedSteps)

	// 获取所有步骤的映射
	allSteps := make(map[string]OrderCreationStep)
	for _, step := range m.factory.GetSteps(ctx.Order.Type) {
		allSteps[step.Name()] = step
	}

	// 只对回滚失败的步骤进行补偿
	for _, failedStepName := range ctx.RollbackFailedSteps {
		if step, ok := allSteps[failedStepName]; ok {
			if err := step.Compensate(ctx); err != nil {
				m.logger.Error("step compensation failed",
					"step", failedStepName,
					"error", err)

				// 补偿失败处理
				m.handleCompensationFailure(ctx, failedStepName, err)
			}
		}
	}
}

// 处理补偿失败的情况
func (m *OrderCreationManager) handleCompensationFailure(ctx *OrderCreationContext, stepName string, err error) {
	// 创建补偿任务
	compensationTask := CompensationTask{
		OrderID:    ctx.Order.ID,
		StepName:   stepName,
		Params:     ctx.Params,
		Cache:      ctx.Cache,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	// 记录错误日志
	m.logger.Error("compensation task created for failed step",
		"order_id", compensationTask.OrderID,
		"step", compensationTask.StepName,
		"error", err)

	// TODO: 实现具体的补偿任务处理逻辑
	// 1. 将任务保存到数据库
	// 2. 发送到消息队列
	// 3. 触发告警
}

// DefaultLogger 默认日志实现
type DefaultLogger struct{}

func NewDefaultLogger() Logger {
	return &DefaultLogger{}
}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	log.Printf("INFO: "+msg, args...)
}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	log.Printf("ERROR: "+msg, args...)
}

// 辅助函数
func generateOrderID() string {
	return fmt.Sprintf("ORDER_%d", time.Now().UnixNano())
}

// CompensationTask 补偿任务结构
type CompensationTask struct {
	OrderID    string
	StepName   string
	Params     map[string]interface{}
	Cache      map[string]interface{}
	RetryCount int
	MaxRetries int
	CreatedAt  time.Time
}

// StockCompensationMessage 库存补偿消息
type StockCompensationMessage struct {
	OrderID   string
	ProductID string
	Quantity  int
	Timestamp time.Time
}
```

</code></pre>
</details>


#### 支付

##### 支付流程
<p align="center">
  <img src="/images/order_pay.png" width=600 height=1200>
</p>
1. 支付校验。用户校验，订单状态校验等
2. 营销活动扣减deduction、回滚rollback、补偿compensation.
3. 支付初始化
4. 支付回调
5. 补偿队列
6. OrderBus 订单事件

##### 支付状态的设计
```
P0: PAYMENT_NOT_STARTED - 未开始
P1: PAYMENT_PENDING - 支付中,用户点击了pay按钮，等待支付）

P2: MARKETING_Init - 营销初始化
P3: MARKETING_FAILED - 营销扣减失败
P4: MARKETING_SUCCESS - 营销扣减成功

P5: PAYMENT_INITIALIZED - 支付初始化
P6: PAYMENT_INITIALIZED_FAILED - 支付初始化失败
P7: PAYMENT_PROCESSING - 支付处理中。（支付系统正在处理支付请求）
P8: PAYMENT_SUCCESS - 支付成功
P9: PAYMENT_FAILED - 支付失败
P10: PAYMENT_CANCELLED - 支付取消
P11: PAYMENT_TIMEOUT - 支付超时
```
##### 异常和补偿设计
常见的异常：
营销部分：
1. 营销扣减补偿操作重复。（营销接口幂等设计）
2. 营销已经扣减了，但是后续步骤失败，需要回滚扣减的操作。（业务代码中需要有rollback操作）
3. 营销已经扣减了，回滚扣减失败。延时队列任务补偿。（回滚失败发送延时队列，任务补偿）
4. 营销已经扣减了，写延时队列失败，任务没有补偿成功。（补偿任务通过扫描异常单进行补偿）
5. 营销已经扣减了，延时队列消息重复，重复回滚。（依赖营销系统的幂等操作）
6. 营销已经扣减了，请求已经发给了营销服务，营销服务已经扣减了，但是回包失败。（请求营销接口之前更新订单状态为P2,针对P2的订单进行补偿）

支付部分：
1. 重复支付。（支付接口幂等设计）
2. 支付初始化请求支付成功，但是回包失败（重续针对P5的订单进行补偿，查询支付系统是否收单，已经支付结果查询）
3. 支付回调包重复，更新回调结果幂等。
4. 支付回调包丢失，对于P7支付单需要补偿。

#### 履约
##### 履约核心流程
<p align="center">
  <img src="/images/order_fulfillment.png" width=500 height=1000>
</p>


##### 履约状态机的设计
```
  F0: FULFILLMENT_NOT_STARTED - 未开始
  F1: FULFILLMENT_PENDING - 履约开始
  F2: FULFILLMENT_PROCESSING - 履约处理中
  F3: FULFILLMENT_FAILED - 履约失败
  F4: FULFILLMENT_SUCCESS - 履约成功
  F5: FULFILLMENT_CANCELLED - 履约取消
```

##### 异常和补偿的设计
1. 订阅支付完成的事件O2
2. 在请求fulfillment/init履约初始化之前，更新订单状态为F1
3. fulfillment/init 接口的回包丢了。（针对F1订单进行补偿）
4. fulfillment/init 重复请求（幂等设计）
5. F2订单补偿。（fulfillment/callback 丢包，处理失败等）


#### return & refund
<p align="center">
  <img src="/images/return-refund.png" width=500 height=1000>
</p>

##### 主要流程
1. 订单服务作为协调者。与履约服务、营销服务、支付服务解耦
2. 用 OrderBus 进行事件传递
3. 状体机设计
4. 异常处理

##### 异常和补偿机制
1. 退货环节异常
- 退货初始化失败：直接发送退款失败事件
- 退货回调失败：更新状态为 R7，发送失败事件
2. 营销退款异常
- 营销处理失败：更新状态为 R14，发送失败事件
- 营销处理成功：更新状态为 R13，继续后续流程
3. 支付退款异常
- 支付退款失败：更新状态为 R11，发送失败事件
- 支付退款成功：更新状态为 R10，发送成功事件

##### 订单详情查询


## 系统挑战和解决方案
### 如何维护订单状态的最终一致性？
<p align="center">
  <img src="/images/order_final_consistency_activity.png" width=600 height=600>
</p>

#### 不一致的原因
- 重复请求
- 丢包。例如，请求发货，对方收单，回包失败。
- 资源回滚：营销、库存
- 并发问题

#### 状态机
  - 设计层面，严格的状态转换规则 + 状态转换的触发事件
  - 状态转换的原子性。（事务性）

#### 并发更新数据库前，要用乐观锁或者悲观锁，
  - 乐观锁：同时在更新时判断版本号是否是之前取出来的版本号，更新成功就结束
  - 悲观锁：先使用select for update进行锁行记录，然后更新

```sql
 UPDATE orders 
   SET status = 'NEW_STATUS', 
       version = version + 1 
   WHERE id = ? AND version = ?

   BEGIN;
   SELECT * FROM orders WHERE id = ? FOR UPDATE;
   UPDATE orders SET status = 'NEW_STATUS' WHERE id = ?;
   COMMIT;
```


#### 幂等设计。比如重复支付、重复扣减营销、重复履约等
  - 支付重复支付，支付回调幂等设计。
  - 重复营销扣减，回滚，
  - 重复履约
  - 重复回调

``` java
   // 使用支付单号作为幂等键
   @Transactional
   public void handlePaymentCallback(String paymentId, String status) {
       // 检查是否已处理
       if (isProcessed(paymentId)) {
           return;
       }
       // 处理支付回调
       processPaymentCallback(paymentId, status);
       // 记录处理状态
       markAsProcessed(paymentId);
   }


      // 使用订单号+营销资源ID作为幂等键
   @Transactional
   public void deductMarketingResource(String orderId, String resourceId) {
       if (isDeducted(orderId, resourceId)) {
           return;
       }
       // 扣减营销资源
       deductResource(orderId, resourceId);
       // 记录扣减状态
       markAsDeducted(orderId, resourceId);
   }

```

#### 补偿机制兜底
  - 异常回滚。营销扣减回滚
  - 消息队列补偿：补偿队列，重试。（可能丢消息）
  - 定时任务补偿：扫表补偿
  - 依赖方支付查询和幂等设计

#### 分布式事务
 - 营销扣减
 - 库存扣减
 - 支付等业务
 - 实现状态转换和业务操作在同一个事务中完成
 ```java
  @Transactional
   public void processOrderWithDistributedTransaction(Order order) {
       try {
           // 1. 更新订单状态
           updateOrderStatus(order);
           // 2. 扣减库存
           deductInventory(order);
           // 3. 创建物流单
           createLogistics(order);
       } catch (Exception e) {
           // 触发补偿机制
           triggerCompensation(order);
       }
   }
 ```

#### 异常单人工介入
#### 对账机制

### 商品信息缓存和数据一致性
<p align="center">
  <img src="/images/item-info-cache.png" width=600 height=500>
</p>

### 主从架构中如何获取最新的数据，避免因为主从延时导致获得脏数据
<p align="center">
  <img src="/images/master-slave-get-latest-data.png" width=500 height=400>
</p>

| **策略**     | **优点**                                                                         | **缺点**                                                                        |
|-------------|----------------------------------------------------------------------------------|--------------------------------------------------------------------------------|
| **1. 直接读取主库** | - **一致性:** 始终获取最新的数据。                | - **性能:** 增加主库的负载，可能导致性能瓶颈。                                                                     |
|                   | - **简单性:** 实现简单直接，因为它直接查询可信的源。 | - **可扩展性:** 主库可能成为瓶颈，限制系统在高读流量下有效扩展的能力。                                                 |
|                   |                                               |                                                                                                             |
| **2. 使用VersionCache与从库** | - **性能:** 分散读取负载到从库，减少主库的压力。  | - **复杂性:** 实现更加复杂，需要进行缓存管理并处理潜在的不一致性问题。                                   |
|                             | - **可扩展性:** 通过将大部分读取操作卸载到从库，实现更好的扩展性。      | - **缓存管理:** 需要进行适当的缓存失效处理和同步，以确保数据的一致性。                 |
|                             | - **一致性:** 通过比较版本并在必要时回退到主库，提供确保最新数据的机制。| - **潜在延迟:** 从库的数据可能仍然存在不同步的可能性，导致数据更新前有轻微延迟。         |


### 常见问题1: 重复下单、支付、履约问题（重复和幂等问题)
场景：
1. 下单、去重、DB唯一键兜底。去重逻辑是约定的
2. 支付、checkoutid，唯一键
3. 履约、先获取reference id，再履约

解决方案：
1. 前端方案
前端通过js脚本控制，无法解决用户刷新提交的请求。另外也无法解决恶意提交。
不建议采用该方案，如果想用，也只是作为一个补充方案。


2. 中间环节去重。根据请求参数中间去重
当用户点击购买按钮时，渲染下单页面，展示商品、收货地址、运费、价格等信息，同时页面会埋上 Token 信息，用户提交订单时，后端业务逻辑会校验token，有且匹配才认为是合理请求。

3. 利用数据库自身特性 "主键唯一约束"，在插入订单记录时，带上主键值，如果订单重复，记录插入会失败。
操作过程如下：
引入一个服务，用于生成一个"全局唯一的订单号"；
进入创建订单页面时，前端请求该服务，预生成订单ID；
提交订单时，请求参数除了业务参数外，还要带上这个预生成订单ID

### 快照和操作日志
为了保证数据的 完整性、可追溯性，写操作需要关注的问题
场景：
商品信息是可以修改的，当用户下单后，为了更好解决后面可能存在的买卖纠纷，创建订单时会同步保存一份商品详情信息，称之为订单快照

解决方案：
同一件商品，会有很多用户会购买，如果热销商品，短时间就会有上万的订单。如果每个订单都创建一份快照，存储成本太高。另外商品信息虽然支持修改，但毕竟是一个低频动作。我们可以理解成，大部分订单的商品快照信息都是一样的，除非下单时用户修改过。
如何实时识别修改动作是解决快照成本的关键所在。我们采用摘要比对的方法‍。创建订单时，先检查商品信息摘要是否已经存在，如果不存在，会创建快照记录。订单明细会关联商品的快照主键。

账户余额更新，保证事务
用户支付，我们要从买家账户减掉一定金额，再往卖家增加一定金额，为了保证数据的 完整性、可追溯性， 变更余额时，我们通常会同时插入一条 记录流水。

账户流水核心字段： 流水ID、金额、交易双方账户、交易时间戳、订单号。
账户流水只能新增，不能修改和删除。流水号必须是自增的。
后续，系统对账时，我们只需要对交易流水明细数据做累计即可，如果出现和余额不一致情况，一般以交易流水为准来修复余额数据。
更新余额、记录流水 虽属于两个操作，但是要保证要么都成功，要么都失败。要做到事务。
当然，如果涉及多个微服务调用，会用到 分布式事务。
分布式事务，细想下也很容易理解，就是 将一个大事务拆分为多个本地事务， 本地事务依然借助于数据库自身事务来解决，难点在于解决这个分布式一致性问题，借助重试机制，保证最终一致是我们常用的方案。

### 常见问题3: 并发更新的ABA问题 （订单表的version)
场景：
商家发货，填写运单号，开始填了 123，后来发现填错了，然后又修改为 456。此时，如果就为某种特殊场景埋下错误伏笔，具体我们来看下，过程如下：
开始「请求A」发货，调订单服务接口，更新运单号 123，但是响应有点慢，超时了；
此时，商家发现运单号填错了，发起了「请求B」，更新运单号为 456 ，订单服务也响应成功了；
这时，「请求A」触发了重试，再次调用订单服务，更新运单号 123，订单服务也响应成功了；订单服务最后保存的 运单号 是 123。

是不是犯错了！！！！，那么有什么好的解决方案吗？
数据库表引入一个额外字段 version ，每次更新时，判断表中的版本号与请求参数携带的版本号是否一致。这个版本字段可以是时间戳
复制
update order
set logistics_num = #{logistics_num} , version = #{version} + 1
where order_id= 1111 and version = #{version}

### 秒杀系统中的库存管理和订单蓄洪
常见的库存扣减方式有：
下单减库存： 即当买家下单后，在商品的总库存中减去买家购买数量。下单减库存是最简单的减库存方式，也是控制最精确的一种，但是有些人下完单可能并不会付款。
付款减库存： 即买家下单后，并不立即减库存，而是等到有用户付款后才真正减库存，否则库存一直保留给其他买家。但因为付款时才减库存，如果并发比较高，有可能出现买家下单后付不了款的情况，因为可能商品已经被其他人买走了。
预扣库存： 这种方式相对复杂一些，买家下单后，库存为其保留一定的时间（如 30 分钟），超过这个时间，库存将会自动释放，释放后其他买家就可以继续购买。在买家付款前，系统会校验该订单的库存是否还有保留：如果没有保留，则再次尝试预扣；
方案一：数据库乐观锁扣减库存
通常在扣减库存的场景下使用行级锁，通过数据库引擎本身对记录加锁的控制，保证数据库的更新的安全性，并且通过where语句的条件，保证库存不会被减到 0 以下，也就是能够有效的控制超卖的场景。
先查库存
然后乐观锁更新：update ... set amount = amount - 1 where id = $id and amount = x
设置数据库的字段数据为无符号整数，这样减后库存字段值小于零时 SQL 语句会报错
方案二：redis 扣减库存，异步同步到DB
redis 原子操作扣减库存
异步通过MQ消息同步到DB


### 购物车模块的实现和优化
技术设计并不是特别复杂，存储的信息也相对有限（用户id、商品id、sku_id、数量、添加时间）。这里特别拿出来单讲主要是用户体验层面要注意几个问题：
添加购物车时，后端校验用户未登录，常规思路，引导用户跳转登录页，待登录成功后，再添加购物车。多了一步操作，给用户一种强迫的感觉，体验会比较差。有没有更好的方式？
如果细心体验京东、淘宝等大平台，你会发现即使未登录态也可以添加购物车，这到底是怎么实现的？
细细琢磨其实原理并不复杂，服务端这边在用户登录态校验时，做了分支路由，当用户未登录时，会创建一个临时Token，作为用户的唯一标识，购物车数据挂载在该Token下，为了避免购物车数据相互影响以及设计的复杂度，这里会有一个临时购物车表。
当然，临时购物车表的数据量并不会太大，why？用户不会一直闲着添加购物车玩，当用户登录后，查看自己的购物车，服务端会从请求的cookie里查找购物车Token标识，并查询临时购物车表是否有数据，然后合并到正式购物车表里。
临时购物车是不是一定要在服务端存储？未必。
有架构师倾向前置存储，将数据存储在浏览器或者 APP LocalStorage， 这部分数据毕竟不是共享的，但是不太好的增加了设计的复杂度。

客户端需要借助本地数据索引，远程请求查完整信息；
如果是登录态，还要增加数据合并逻辑；
考虑到这两部分数据只是用户标识的差异性，所以作者还是建议统一存到服务端，日后即使业务逻辑变更，只需要改一处就可以了，毕竟自运营系统，良好的可维护性也需要我们非常关注的。

购物车是电商系统的标配功能，暂存用户想要购买的商品。
- 分为添加商品、列表查看、结算下单三个动作。
- 用户未登录时，将数据存储在浏览器或者 APP LocalStorage。登录后写入后端
- 后端使用DB，为了性能考虑可以结合redis和DB联合存储
- 存redis定期同步到DB
- 前后端联合存储


### 系统中的分布式ID是怎么生成的
item ID 自增。（100w级别）
order id. 时间戳 + 机器ID + uid % 100 + sequence
DP唯一ID生成调研说明
request 生成方法：时间戳 + 机器mac地址 + sequence


## 系统稳定性建设

### 前言：怎样的系统算是稳定高可用的

首先回答另一个问题，怎样的系统算是稳定的？

Google SRE中(SRE三部曲[1])有一个层级模型来描述系统可靠性基础和高层次需求(Dickerson's Hierarchy of Service Reliability)，如下图：

<p align="center">
  <img src="/images/service-reliability-hierarchy.png" width=600 height=500>
  <br/>
</p>


该模型由Google SRE工程师Mikey Dickerson在2013年提出，将系统稳定性需求按照基础程度进行了不同层次的体系化区分，形成稳定性标准金字塔模型:
- 金字塔的底座是监控(Monitoring)，这是一个系统对于稳定性最基础的要求，缺少监控的系统，如同蒙上眼睛狂奔的野马，无从谈及可控性，更遑论稳定性。
- 更上层是应急响应(Incident Response)，从一个问题被监控发现到最终解决，这期间的耗时直接取决于应急响应机制的成熟度。合理的应急策略能保证当故障发生时，所有问题能得到有序且妥善的处理，而不是慌乱成一锅粥。
- 事后总结以及根因分析(Postmortem&Root Caue Analysis)，即我们平时谈到的"复盘"，虽然很多人都不太喜欢这项活动，但是不得不承认这是避免我们下次犯同样错误的最有效手段，只有当摸清故障的根因以及对应的缺陷，我们才能对症下药，合理进行规避。
- 测试和发布管控(Testing&Release procedures),大大小小的应用都离不开不断的变更与发布,有效的测试与发布策略能保障系统所有新增变量都处于可控稳定区间内，从而达到整体服务终态稳定
- 容量规划(Capacity Planning)则是针对于这方面变化进行的保障策略。现有系统体量是否足够支撑新的流量需求，整体链路上是否存在不对等的薄弱节点，都是容量规划需要考虑的问题。
- 位于金字塔模型最顶端的是产品设计(Product)与软件研发(Development)，即通过优秀的产品设计与软件设计使系统具备更高的可靠性，构建高可用产品架构体系，从而提升用户体验


### 系统稳定性建设概述
<p align="center">
  <img src="/images/system-stability.png" width=800 height=800>
  <br/>
</p>

从金字塔模型我们可以看到构建维护一个高可用服务所需要做到的几方面工作：
- 产品、技术、架构的设计，高可用的架构体系
- 系统链路&业务策略梳理和维护（System & Biz Profiling）
- 容量规划（Capacity Planning）
- 应急响应（Incident Response）
- 测试
- 事后总结（Testing & Postmortem）
- 监控（Monitoring）
- 资损体系
- 风控体系
- 大促保障
- 性能优化

<p align="center">
  <img src="/images/6-reliability-steps.png" width=600 height=500>
  <br/>
</p>


### 高可用的架构设计

## 系统链路梳理和维护 System & Biz Profiling

系统链路梳理是所有保障工作的基础，如同对整体应用系统进行一次全面体检，从流量入口开始，按照链路轨迹，逐级分层节点，得到系统全局画像与核心保障点。

### 入口梳理盘点
一个系统往往存在十几个甚至更多流量入口，包含HTTP、RPC、消息等都多种来源。如果无法覆盖所有所有链路，可以从以下三类入口开始进行梳理：
- 核心重保流量入口
  - 用户承诺服务SLI较高，对数据准确性、服务响应时间、可靠度具有明确要求。
  - 业务核心链路，浏览、下单、支付、履约
  - 面向企业级用户
- 资损事件对应入口
  - 关联到公司资金收入或者客户资金收入收费服务
- 大流量入口
  - 系统TPS&QPS TOP5~10
  - 该类入口虽然不涉及较高SLI与资损要求，但是流量较高，对整体系统负载有较大影响


### 节点分层判断
对于复杂场景可以做节点分层判断

流量入口就如同线团中的线头，挑出线头后就可按照流量轨迹对链路上的节点(HSF\DB\Tair\HBase等一切外部依赖)按照依赖程度、可用性、可靠性进行初级分层区分。

1. 强弱依赖节点判断
  - 若节点不可用，链路业务逻辑被中断 or 高级别有损(存在一定耐受阈值)，则为业务强依赖；反之为弱依赖。
  - 若节点不可用，链路执行逻辑被中断(return error)，则为系统强依赖；反之为弱依赖。
  - 若节点不可用，系统性能受影响，则为系统强依赖；反之为弱依赖。
  - 按照快速失败设计逻辑，该类节点不应存在，但是在不变更应用代码前提下，如果出现该类节点，应作为强依赖看待。
  - 若节点无感可降级 or 存在业务轻微损伤替换方案，则为弱依赖。
2. 低可用依赖节点判断
  - 节点服务日常超时严重
  - 节点对应系统资源不足

3. 高风险节点判断
  - 上次大促后，节点存在大版本系统改造
  - 新上线未经历过大促的节点
  - 节点对应系统是否曾经出现高级别故障
  - 节点故障后存在资损风险


### 应产出数据
- 识别核心接口（流程）调用拓扑图或者时序图（借用分布式链路追踪系统获得调用拓扑图）
- 调用比
- 识别资损风险
- 识别内外部依赖

完成该项梳理工作后，我们应该产出以下数据：对应业务域所有核心链路分析，技术&业务强依赖、核心上游、下游系统、资损风险应明确标注。

## 监控&告警梳理 -- Monitoring
站在监控的角度看，我们的系统从上到下一般可以分为三层：业务（Biz）、应用（Application）、系统（System）。系统层为最下层基础，表示操作系统相关状态；应用层为JVM层，涵盖主应用进程与中间件运行状态；业务层为最上层，为业务视角下服务对外运行状态。因此进行大促稳定性监控梳理时，可以先脱离现有监控，先从核心、资损链路开始，按照业务、应用（中间件、JVM、DB）、系统三个层次梳理需要哪些监控，再从根据这些索引找到对应的监控告警，如果不存在，则相应补上；如果存在则检查阈值、时间、告警人是否合理。

### 监控
监控系统一般有四项黄金指标：延时（Latency）, 错误（Error）,流量（Traffic）, 饱和度（Situation），各层的关键性监控同样也可以按照这四项指标来进行归类，具体如下：

<p align="center">
  <img src="/images/how-to-monitor.png" width=900 height=500>
  <br/>
</p>


### 告警
是不是每项监控都需要告警？答案当然是否定的。建议优先设置Biz层告警，因为Biz层我们对外服务最直观业务表现，最贴切用户感受。Application&System层指标主要用于监控，部分关键&高风险指标可设置告警，用于问题排查定位以及故障提前发现。对于一项告警，我们一般需要关注级别、阈值、通知人等几个点。

1. 级别
即当前告警被触发时，问题的严重程度，一般来说有几个衡量点：

- 是否关联NOC
- 是否产生严重业务影响
- 是否产生资损

2. 阈值
- 即一项告警的触发条件&时间，需根据具体场景合理制定。一般遵循以下原则：
- 不可过于迟钝。一个合理的监控体系中，任何异常发生后都应触发相关告警。
- 不可过于敏感。过于敏感的阈值会造成频繁告警，从而导致响应人员疲劳应对，无法筛选真实异常。若一个告警频繁出现，一般是两个原因：系统设计不合理 or 阈值设置不合理。
- 若单一指标无法反馈覆盖整体业务场景，可结合多项指标关联构建。
- 需符合业务波动曲线，不同时段可设置不同条件&通知策略。

3. 通知人&方式
- 若为业务指标异常(Biz层告警)，通知人应为问题处理人员(开发、运维同学)与业务关注人员(TL、业务同学)的集合，通知方式较为实时，比如电话通知。
- 若为应用 & 系统层告警，主要用于定位异常原因，通知人设置问题排查处理人员即可，通知方式可考虑钉钉、短信等低干扰方式。
- 除了关联层次，对于不同级别的告警，通知人范围也可适当扩大，尤其是关联GOC故障的告警指标，应适当放宽范围，通知方式也应更为实时直接

#### 应产出数据

完成该项梳理工作后，我们应该产出以下数据：
1. 系统监控模型，格式同表1
  - Biz、Application、System 分别存在哪些待监控点
  - 监控点是否已全部存在指标，仍有哪些待补充

2. 系统告警模型列表，需包含以下数据
  - 关联监控指标（链接）
  - 告警关键级别
  - 是否推送GOC
  - 是否产生资损
  - 是否关联故障
  - 是否关联预案
3. 业务指标大盘，包含Biz层重点监控指标数据
4. 系统&应用指标大盘，包含核心系统关键系统指标，可用于白盒监控定位问题。



## 业务策略&容量规划 Capacity Planning - 容量规划

### 业务策略
不同于高可用系统建设体系，大促稳定性保障体系与面向特定业务活动的针对性保障建设，因此，业务策略与数据是我们进行保障前不可或缺的数据。
一般大促业务数据可分为两类，全局业务形态评估以及应急策略&玩法。

#### 全局评估

该类数据从可以帮助我们进行精准流量评估、峰值预测、大促人力排班等等，一般包含下面几类：
- 业务量预估体量（日常X倍）
- 预估峰值日期
- 大促业务时长（XX日-XX日）
- 业务场景预估流量分配

#### 应急策略
- 该类数据指相较于往年大促活动，本次大促业务变量，可用于应急响应预案与高风险节点评估等，一般包含下面两类：
- 特殊业务玩法

容量规划的本质是追求计算风险最小化和计算成本最小化之间的平衡，只追求任意其一都不是合理的。为了达到这两者的最佳平衡点，需尽量精准计算系统峰值负载流量，再将流量根据单点资源负载上限换算成相应容量，得到最终容量规划模型。


### 流量模型评估

1. 入口流量

对于一次大促，系统峰值入口流量一般由常规业务流量与非常规增量（比如容灾预案&业务营销策略变化带来的流量模型配比变化）叠加拟合而成。

- 常规业务流量一般有两类计算方式：
  - 历史流量算法：该类算法假设当年大促增幅完全符合历史流量模型，根据当前&历年日常流量，计算整体业务体量同比增量模型；然后根据历年大促-日常对比，计算预估流量环比增量模型；最后二者拟合得到最终评估数据。
  - 由于计算时无需依赖任何业务信息输入，该类算法可用于保障工作初期业务尚未给出业务总量评估时使用，得到初估业务流量。
  - 业务量-流量转化算法(GMV\DAU\订单量)：该类算法一般以业务预估总量（GMV\DAU\订单量）为输入，根据历史大促&日常业务量-流量转化模型（比如经典漏洞模型）换算得到对应子域业务体量评估。- 该种方式强依赖业务总量预估，可在保障工作中后期使用，在初估业务流量基础上纳入业务评估因素考虑。
- 非常规增量一般指前台业务营销策略变更或系统应急预案执行后流量模型变化造成的增量流量。例如，NA61机房故障时，流量100%切换到NA62后，带来的增量变化.考虑到成本最小化，非常规增量P计算时一般无需与常规业务流量W一起，全量纳入叠加入口流量K，一般会将非常规策略发生概率λ作为权重

2. 节点流量
节点流量由入口流量根据流量分支模型，按比例转化而来。分支流量模型以系统链路为计算基础，遵循以下原则：
- 同一入口，不同链路占比流量独立计算。
- 针对同一链路上同一节点，若存在多次调用，需计算按倍数同比放大（比如DB\Tair等）。
- DB写流量重点关注，可能出现热点造成DB HANG死。


### 容量转化

节点容量是指一个节点在运行过程中，能够**同时处理的最大请求数**。它反映了系统的瞬时负载能力。


1）Little Law衍生法则
不同类型资源节点(应用容器、Tair、DB、HBASE等)流量-容量转化比各不相同，但都服从Little Law衍生法则，即：
  节点容量=节点吞吐率×平均响应时间

2）N + X 冗余原则

在满足目标流量所需要的最小容量基础上，冗余保留X单位冗余能力
X与目标成本与资源节点故障概率成正相关，不可用概率越高，X越高
对于一般应用容器集群，可考虑X = 0.2N


### 全链路压测(TODO)
- 上述法则只能用于容量初估(大促压测前&新依赖)，最终精准系统容量还是需要结合系统周期性压力测试得出。

### 应产出数据
- 基于模型评估的入口流量模型 & 集群自身容量转化结果（若为非入口应用，则为限流点梳理）。
- 基于链路梳理的分支流量模型 & 外部依赖容量转化结果。


## 大促保障
### Incident Response - 紧急&前置预案梳理
要想在大促高并发流量场景下快速对线上紧急事故进行响应处理，仅仅依赖值班同学临场发挥是远远不够的。争分夺秒的情况下，无法给处理人员留有充足的策略思考空间，而错误的处理决策，往往会导致更为失控严重的业务&系统影响。因此，要想在大促现场快速而正确的响应问题，值班同学需要做的是选择题(Which)，而不是陈述题(What)。而选项的构成，便是我们的业务&系统预案。从执行时机与解决问题属性来划分，预案可分为技术应急预案、技术前置预案、业务应急预案、业务前置预案等四大类。结合之前的链路梳理和业务评估结果，我们可以快速分析出链路中需要的预案，遵循以下原则：
- 技术应急预案：该类预案用于处理系统链路中，某层次节点不可用的情况，例如技术/业务强依赖、弱稳定性、高风险等节点不可用等异常场景。
- 技术前置预案：该类预案用于平衡整体系统风险与单节点服务可用性，通过熔断等策略保障全局服务可靠。例如弱稳定性&弱依赖服务提前降级、与峰值流量时间冲突的离线任务提前暂定等。
- 业务应急预案：该类预案用于应对业务变更等非系统性异常带来的需应急处理问题，例如业务数据错误（数据正确性敏感节点）、务策略调整（配合业务应急策略）等
- 业务前置预案：该类预案用于配和业务全局策略进行的前置服务调整（非系统性需求）

### 应产出数据
完成该项梳理工作后，我们应该产出以下数据：
- 执行&关闭时间（前置预案）
- 触发阈值（紧急预案，须关联相关告警）
- 关联影响（系统&业务）
- 决策&执行&验证人员
- 开启验证方式
- 关闭阈值（紧急预案）
- 关闭验证方式


阶段性产出-全链路作战地图

进行完上述几项保障工作，我们基本可得到全局链路作战地图，包含链路分支流量模型、强弱依赖节点、资损评估、对应预案&处理策略等信息。大促期间可凭借该地图快速从全局视角查看应急事件相关影响，同时也可根据地图反向评估预案、容量等梳理是否完善合理。

#### Incident Response - 作战手册梳理

作战手册是整个大促保障的行动依据，贯穿于整个大促生命周期，可从事前、事中、事后三个阶段展开考虑。整体梳理应本着精准化、精细化的原则，理想状态下，即便是对业务、系统不熟悉的轮班同学，凭借手册也能快速响应处理线上问题。
**事前**
1）前置检查事项清单
- 大促前必须执行事项checklist,通常包含以下事项：
- 集群机器重启 or 手动FGC
- 影子表数据清理
- 检查上下游机器权限
- 检查限流值
- 检查机器开关一致性
- 检查数据库配置
- 检查中间件容量、配置(DB\缓存\NoSQL等)
- 检查监控有效性（业务大盘、技术大盘、核心告警）
- 每个事项都需包含具体执行人、检查方案、检查结果三列数据
2）前置预案
- 域内所有业务&技术前置预案。

**事中**
1. 紧急技术&业务预案
需要包含的内容基本同前置预案，差异点如下：
- 执行条件&恢复条件：具体触发阈值，对应监控告警项。
- 通知决策人。
2. 应急工具&脚本
常见故障排查方式、核心告警止血方式(强弱依赖不可用等)，业务相关日志捞取脚本等。
3. 告警&大盘
- 应包含业务、系统集群及中间件告警监控梳理结果，核心业务以及系统大盘，对应日志数据源明细等数据：
- 日志数据源明细：数据源名称、文件位置、样例、切分格式。
- 业务、系统集群及中间件告警监控梳理结果：关联监控指标（链接）、告警关键级别、是否推送GOC、是否产生资损、是否关联故障、是否关联预案。
- 核心业务&系统大盘：大盘地址、包含指标明细(含义、是否关联告警、对应日志)。

4. 上下游机器分组
- 应包含核心系统、上下游系统，在不同机房、单元集群分组、应用名，可用于事前-机器权限检查、事中-应急问题排查黑屏处理。
5. 值班注意事项
- 包含每班轮班同学值班必做事项、应急变更流程、核心大盘链接等。
6. 核心播报指标
- 包含核心系统&服务指标(CPU\LOAD\RT)、业务关注指标等，每项指标应明确具体监控地址、采集方式。
7. 域内&关联域人员通讯录、值班
- 包含域内技术、TL、业务方对应排班情况、联系方式(电话)，相关上下游、基础组件(DB、中间件等)对应值班情况。
8. 值班问题记录
- 作战记录，记录工单、业务问题、预案(前置\紧急)（至少包含：时间、问题描述（截图）、影响分析、决策&解决过程等）。值班同学在值班结束前，进行记录。
**事后**
1. 系统恢复设置事项清单(限流、缩容)
一般与事前检查事项清单对应，包含限流阈值调整、集群缩容等大促后恢复操作。
2. 大促问题复盘记录
- 应包含大促遇到的核心事件总结梳理。



### 沙盘推演和演练 Incident Response

实战沙盘演练是应急响应方面的最后一项保障工作，以历史真实故障CASE作为应急场景输入，模拟大促期间紧急状况，旨在考验值班同学们对应急问题处理的响应情况。
一般来说，一个线上问题从发现到解决，中间需要经历定位&排查&诊断&修复等过程，总体遵循以下几点原则：
- 尽最大可能让系统先恢复服务，同时为根源调查保护现场（机器、日志、水位记录）。
- 避免盲目搜索，依据白盒监控针对性诊断定位。
- 有序分工，各司其职，避免一窝蜂失控乱象。
- 依据现场情况实时评估影响范围，实在无法通过技术手段挽救的情况(例如强依赖不可用)，转化为业务问题思考（影响范围、程度、是否有资损、如何协同业务方）。
- 沙盘演练旨在检验值班同学故障处理能力，着重关注止血策略、分工安排、问题定位等三个方面：
国际化中台双11买家域演练
根据故障类型，常见止血策略有以下解决思路：
- 入口限流：调低对应Provider服务来源限流值
- 应对突发流量过高导致自身系统、下游强依赖负载被打满。
- 下游降级：降级对应下游服务
- 下游弱依赖不可用。
- 下游业务强依赖经业务同意后降级（业务部分有损）。
- 单点失败移除：摘除不可用节点
- 单机水位飙高时，先下线不可用单机服务（无需下线机器，保留现场）。
- 应对集群单点不可用、性能差。
- 切换：单元切流或者切换备份

应对单库或某单元依赖因为自身原因（宿主机或网络），造成局部流量成功率下跌下跌。
Google SRE中，对于紧急事故管理有以下几点要素：
- 嵌套式职责分离，即分确的职能分工安排
- 控制中心\作战室
- 实时事故状态文档
- 明确公开的职责交接
- 其中嵌套式职责分离，即分确的职能分工安排，达到各司其职，有序处理的效果，一般可分为下列几个角色：
事故总控：负责协调分工以及未分配事务兜底工作，掌握全局概要信息，一般为PM/TL担任。
事务处理团队：事故真正处理人员，可根据具体业务场景&系统特性分为多个小团队。团队内部存在域内负责人，与事故总控人员进行沟通。
发言人：事故对外联络人员，负责对事故处理内部成员以及外部关注人员信息做周期性信息同步，同时需要实时维护更新事故文档。
规划负责人：负责外部持续性支持工作，比如当大型故障出现，多轮排班轮转时，负责组织职责交接记录

## 资损体系

### 定期review资损风险

### 事中及时发现

<p align="center">
  <img src="/images/realtime-verify.webp" width=800 height=600>
  <br/>
  <strong><a href="https://segmentfault.com/a/1190000040286146">【得物技术】浅谈资损防控</a></strong>
  <br/>
</p>

### 事后复盘和知识沉淀

### 参考学习
- [资损防控技术体系简介及实践](https://tech.dewu.com/article?id=73)
- [浅谈资损防控](https://segmentfault.com/a/1190000040286146)

## 风控体系

## 性能优化

<p align="center">
  <img src="/images/performance.png" width=800 height=800>
  <br/>
</p>

学习资料：
- https://landing.google.com/sre/books/
- https://sre.google/sre-book/table-of-contents/
- https://sre.google/workbook/table-of-contents/
- https://mp.weixin.qq.com/s/w2tOXR6rcTmUHGsJKJilzg?spm=a2c6h.12873639.article-detail.7.31fc2988tIxeaF


## 面试题

### 基础概念与架构设计
- 电商后台系统的核心架构设计原则有哪些？
- 电商后台系统与前端系统的交互方式有哪些？各自的特点是什么？
- 如何设计电商后台系统的用户权限管理模块？
- 电商后台系统中，微服务架构和单体架构的适用场景分别是什么？
- 简述电商后台系统的分层架构设计，各层的主要职责是什么？
- 如何实现电商后台系统的接口幂等性？
- 电商后台系统中，分布式 Session 管理有哪些常见方案？
- 设计电商后台系统时，如何考虑系统的可扩展性和可维护性？
- 电商后台系统的 API 设计规范应包含哪些内容？
- 如何设计电商后台系统的异常处理机制？



### 商品管理
- 什么是SPU和SKU？它们之间的关系是什么？
- 电商系统中的商品分类体系是如何设计的？ category 父子类目
- 什么是商品属性？如何区分规格属性（Sales Attributes））和非规格属性？属性，会影响商品SKU的属性直接关系到库存和价格，用户购买时需要选择的属性，例如：颜色、尺码、内存容量等。非规格属性（Basic Attributes）：用于描述商品特征，产地、材质、生产日期
- 商品的生命周期包含哪些状态？
```
  创建阶段
  DRAFT(0): 草稿状态
  PENDING_AUDIT(1): 待审核
  AUDIT_REJECTED(2): 审核拒绝
  AUDIT_APPROVED(3): 审核通过
  销售阶段
  ON_SHELF(10): 在售/上架
  OFF_SHELF(11): 下架
  SOLD_OUT(12): 售罄
  特殊状态
  FROZEN(20): 冻结（违规/投诉）
  DELETED(99): 删除
```
什么是商品快照？为什么需要商品快照？

- 商品的 SKU 和 SPU 概念在后台系统中如何体现？两者的关系是怎样的？
- 电商后台系统中商品的基础信息包括哪些？如何设计商品表的数据库模型？
- 商品的库存管理和价格管理在后台系统中是如何关联的？
- 如何处理商品的多规格（如颜色、尺寸、型号等）信息？数据库表结构如何设计？
- 商品详情页的信息（如描述、图片、参数）在后台系统中如何存储和管理？
- 商品上下架的逻辑在后台系统中是如何实现的？需要考虑哪些因素（如库存、审核状态等）？
- 商品的搜索和筛选功能在后台系统中是如何实现的？涉及哪些技术（如全文搜索、数据库索引等）？
- 新品发布和商品淘汰在后台系统中的处理流程是怎样的？
- 如何保证商品信息的唯一性和完整性，避免重复录入和数据错误？
1) 唯一性保证：
使用唯一索引
引入商品编码系统
查重机制
2) 完整性保证：
必填字段验证
数据格式校验
关联完整性检查
业务规则校验


### 订单管理
- 电商后台系统中订单的主要状态有哪些？状态流转的触发条件和处理逻辑是怎样的？
- 订单的创建流程在后台系统中是如何处理的？涉及哪些模块（如库存、价格、用户信息等）的交互？
- 如何实现订单的分单处理（如不同仓库发货、不同店铺订单拆分）？
- 订单的支付状态如何与支付系统进行同步？后台系统需要处理哪些异常情况？
- 订单的取消、修改（如收货地址、商品数量）在后台系统中有哪些限制和处理逻辑？
- 如何计算订单的总价（包括商品价格、运费、优惠活动等）？优惠分摊的逻辑是怎样的？
- 订单的物流信息在后台系统中如何获取和更新？与物流服务商的接口如何对接？
- 历史订单的存储和查询在后台系统中如何优化？涉及大量数据时如何提高查询效率？
- 如何设计订单的反欺诈机制，识别和防范恶意订单？
- 订单的售后服务（如退货、换货、退款）在后台系统中的处理流程是怎样的？与库存、财务等模块如何交互？


### 用户与账户管理
- 电商后台系统中用户信息通常包含哪些字段？如何设计用户表的数据库结构？
- 如何实现用户的注册、登录（包括第三方登录）功能在后台系统中的处理逻辑？
- 怎样处理用户密码的加密存储和找回功能？
- 电商后台如何管理用户的收货地址？地址数据的增删改查逻辑是怎样的？
- 用户账户余额和积分的管理在后台系统中有哪些注意事项？如何保证数据的一致性？
- 如何实现用户权限的分级管理（如普通用户、VIP 用户、管理员等）？
- 当用户账户出现异常登录时，后台系统应如何处理和记录？
- 电商后台如何统计用户的活跃度、留存率等指标？数据来源和计算逻辑是怎样的？
- 用户信息修改（如手机号、邮箱）时，后台系统需要进行哪些验证和处理？
- 如何设计用户操作日志的记录和查询功能，以满足审计和问题排查需求？

### 库存管理
- 电商后台系统中库存管理的主要目标是什么？常见的库存管理策略有哪些？
- 如何实现库存的实时更新？在高并发场景下如何保证库存数据的一致性？
- 库存预警机制如何设计？预警的条件（如安全库存、滞销库存等）和通知方式是怎样的？
- 多仓库库存管理在后台系统中如何实现？库存的分配和调拨逻辑是怎样的？
- 库存盘点功能在后台系统中的实现步骤是怎样的？如何处理盘点差异？
- 预售商品的库存管理与普通商品有何不同？后台系统需要特殊处理哪些方面？
- 如何防止超卖现象的发生？在库存不足时，订单的处理逻辑是怎样的？
- 库存数据与订单、采购、物流等模块的交互接口是如何设计的？
- 对于虚拟商品（如电子卡券），库存管理的方式与实物商品有何区别？
- 如何统计库存的周转率、缺货率等指标？数据来源和计算方法是怎样的？



### 支付与结算
- 电商后台系统支持哪些支付方式？每种支付方式的对接流程和注意事项是什么？
- 支付系统与电商后台系统的交互接口应包含哪些关键信息？如何保证支付数据的安全性？
- 支付过程中的异步通知机制是如何实现的？后台系统如何处理重复通知和通知失败的情况？
- 如何实现支付订单与业务订单的关联和对账功能？
- 结算周期和结算规则在后台系统中如何配置和管理？（如供应商结算、平台佣金结算等）
- 支付过程中的手续费计算和分摊逻辑是怎样的？如何在后台系统中实现？
- 对于跨境支付，后台系统需要处理哪些特殊问题（如汇率转换、支付合规性等）？
- 如何设计支付系统的异常处理和回滚机制？
- 支付成功后，后台系统如何触发后续的业务流程（如订单发货、积分发放等）？
- 财务对账在后台系统中的实现方式有哪些？如何保证财务数据与业务数据的一致性？

### 物流与供应链
- 电商后台系统如何与物流服务商（如快递、仓储）进行接口对接？需要获取哪些物流信息？
- 物流单号的生成和管理在后台系统中是如何实现的？如何避免重复和错误？
- 发货流程在后台系统中的处理逻辑是怎样的？涉及哪些部门或系统的协作（如仓库、库存、订单等）？
- 如何实现物流信息的实时追踪和更新？在后台系统中如何展示给用户和客服？
- 退换货的物流处理在后台系统中有哪些特殊流程？如何与原订单和库存进行关联？
- 供应链管理在电商后台系统中包括哪些主要功能？如何实现供应商管理、采购管理和库存管理的协同？
- 如何根据商品的特性和用户地址选择合适的物流方案（如快递类型、运费模板等）？
- 物流异常（如包裹丢失、破损）在后台系统中的处理流程是怎样的？如何与用户和物流服务商沟通协调？
- 如何统计物流成本和物流效率（如发货时效、配送成功率等）？数据来源和分析方法是怎样的？
- 对于海外仓和跨境物流，后台系统需要处理哪些额外的业务逻辑（如清关、关税计算等）？

### 营销与促销
电商后台系统中常见的促销策略有哪些（如满减、打折、优惠券、秒杀、拼团等）？如何设计支持多种促销策略的模块？
优惠券的生成、发放、使用和核销在后台系统中的处理流程是怎样的？
促销活动的时间管理和范围管理（如针对特定用户群体、特定商品、特定时间段）如何实现？
如何避免促销活动中的超卖和优惠叠加错误？后台系统的校验逻辑是怎样的？
秒杀活动在后台系统中如何应对高并发场景？需要进行哪些技术优化？
营销活动的效果评估指标（如转化率、客单价提升、销售额增长等）在后台系统中如何统计和分析？
如何设计推荐系统与后台营销模块的集成，实现个性化的促销推荐？
会员体系（如 VIP 等级、积分兑换）在后台系统中如何与营销活动结合？
促销活动的库存预留和释放逻辑是怎样的？如何与库存管理模块进行交互？
营销费用的预算管理和成本核算在后台系统中如何实现？

### 数据统计与分析
- 电商后台系统需要统计哪些核心业务指标（如 GMV、UV、PV、转化率、复购率等）？数据采集的方式和频率是怎样的？
- 如何设计数据报表功能，支持不同角色（如运营、管理层、客服）的个性化报表需求？
- 数据统计中的维度和指标如何定义和管理？如何实现多维度的交叉分析？
- 实时数据统计和离线数据统计在后台系统中的实现方式有何不同？各自的适用场景是什么？
- 如何保证数据统计的准确性和完整性？数据清洗和校验的流程是怎样的？
- 数据分析结果如何反馈到业务模块（如库存调整、促销策略优化等）？
- 数据可视化在后台系统中的实现方式有哪些（如图表、仪表盘、数据大屏等）？
- 对于海量数据的统计分析，后台系统需要进行哪些性能优化（如分布式计算、缓存、索引等）？
- 如何设计数据权限管理，确保不同用户只能查看和操作其权限范围内的数据？
- 数据统计分析模块与其他业务模块（如订单、商品、用户等）的数据接口是如何设计的？
### 其他综合问题
- 电商后台系统在应对大促（如双 11、618）时，需要进行哪些准备工作和技术优化？
- 如何保障电商后台系统的高可用性和容灾能力？常见的解决方案有哪些？
- 后台系统的代码维护和版本管理有哪些最佳实践？如何保证多人协作开发的效率和代码质量？
- 当电商业务拓展到新的领域或增加新的业务模块时，后台系统如何进行适应性改造？
- 如何处理不同国家和地区的电商业务在后台系统中的差异化需求（如语言、货币、法规等）？
- 电商后台系统中的日志管理有哪些重要性？如何设计日志的记录、存储和查询功能？
- 对于第三方服务（如短信验证码、邮件通知、数据分析工具等）的接入，后台系统需要注意哪些问题？
- 如何评估电商后台系统的性能瓶颈？常见的性能测试工具和方法有哪些？
- 简述你在以往项目中参与过的电商后台系统开发经验，遇到过哪些挑战，是如何解决的？
- 对于电商后台系统的未来发展趋势（如智能化、自动化、区块链应用等），你有哪些了解和思考？






## 参考:
- [订单状态机的设计和实现](https://www.cnblogs.com/wanglifeng717/p/16214122.html)
- [Understanding the Structure of E-Commerce Products](https://axureboutique.com/blogs/product-design/understanding-the-structure-of-e-commerce-products?srsltid=AfmBOorcMDfLRBbuCUYyKtkpkf5Vf8yQUjJSRKR0FzQSI2lvcvMmIK--)
- [Build an E-Commerce Product Center from Scratch](https://axureboutique.com/blogs/product-design/build-an-e-commerce-product-center-from-scratch)





 1000336300005941988
 2746699375374033603