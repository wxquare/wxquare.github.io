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

#### 后序扩展：
- images 字段为 JSON，建议单独建表管理图片，便于后续扩展（如视频、3D模型等）。
- 价格、库存等字段仅在 sku/item_state，若需支持价格历史、促销活动等，需额外设计。增加单独的库存表stock_tab和价格表 price_tab.
- 多语言支持
- 多币种支持

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


### 商品的价格和库存
#### 方案1. 价格和库存直接放在sku表中
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


#### 方案2. 价格和库存单独管理

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


### 商品的生命周期
- 商品的生命周期
- 如何处理商品的上下架
- 如何设计商品的审核流程？


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

### 商品快照 item_snapshots
1. 商品编辑时生成快照:
- 每次商品信息（如价格、描述、属性等）发生编辑时，生成一个新的商品快照。
- 将快照信息存储在 item_snapshots 表中，并生成一个唯一的 snapshot_id。

2. 订单创建时使用快照:
在用户下单时，查找当前商品的最新 snapshot_id。
在 order_items 表中记录该 snapshot_id，以确保订单项反映下单时的商品状态

(如何设计商品的版本控制？)
(商品的历史记录如何管理？)

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

### 用户操作日志
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
### 扩展，多语言多币种
### 扩展，物流


## 订单管理 Order Center

### 订单模型
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


### 订单状态机
<p align="center">
  <img src="/images/order_state_machine.png" width=800 height=800>
</p>

### 订单ID 生成策略。
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

### 创单
- 接口定义：通过OrderCreationStep接口定义了每个步骤必须实现的方法
- 上下文共享：使用OrderCreationContext在步骤间共享数据
- 步骤独立：每个步骤都是独立的，便于维护和测试
- 回滚机制：每个步骤都实现了回滚方法
- 流程管理：通过OrderCreationManager统一管理步骤的执行和回滚
- 错误处理：统一的错误处理和回滚机制
- 可扩展性：易于添加新的步骤或修改现有步骤
- 如何解决不同category 创单差异较大的问题。

```
// 定义错误码
type ErrorCode int

const (
    Success ErrorCode = iota
    ErrUserNotFound
    ErrProductNotFound
    ErrInsufficientStock
    ErrPriceChanged
    ErrPromotionExpired
    ErrSystemError
    // ... 其他错误码
)

// 错误处理结构
type OrderError struct {
    Code    ErrorCode
    Message string
    Step    string
    Details map[string]interface{}
}

// 错误处理方法
func handleError(err *OrderError) {
    // 1. 记录错误日志
    logError(err)
    
    // 2. 发送告警
    if isCriticalError(err.Code) {
        sendAlert(err)
    }
    
    // 3. 记录错误统计
    recordErrorMetrics(err)
}


// 回滚管理器
type RollbackManager struct {
    steps []RollbackStep
}

// 回滚步骤接口
type RollbackStep interface {
    Execute() error
    Rollback() error
}

// 订单创建回滚
func (rm *RollbackManager) Rollback() error {
    // 从后往前执行回滚
    for i := len(rm.steps) - 1; i >= 0; i-- {
        if err := rm.steps[i].Rollback(); err != nil {
            log.Printf("Rollback step %d failed: %v", i, err)
            // 继续执行其他回滚步骤
        }
    }
    return nil
}

func validateUser(userID string) (*UserValidationStep, error) {
    step := &UserValidationStep{
        userID: userID,
    }
    
    // 执行校验
    if err := step.Execute(); err != nil {
        return nil, &OrderError{
            Code:    ErrUserNotFound,
            Message: "User validation failed",
            Step:    "UserValidation",
            Details: map[string]interface{}{
                "userID": userID,
                "error":  err.Error(),
            },
        }
    }
    
    return step, nil
}

func validateAndDeductStock(productID string, quantity int) (*StockDeductionStep, error) {
    step := &StockDeductionStep{
        productID: productID,
        quantity:  quantity,
    }
    
    // 使用分布式锁
    lock := distributedLock.NewLock("stock:" + productID)
    if !lock.TryLock() {
        return nil, &OrderError{
            Code:    ErrLockFailed,
            Message: "Failed to acquire stock lock",
            Step:    "StockDeduction",
        }
    }
    defer lock.Unlock()
    
    // 执行库存扣减
    if err := step.Execute(); err != nil {
        return nil, &OrderError{
            Code:    ErrInsufficientStock,
            Message: "Stock deduction failed",
            Step:    "StockDeduction",
        }
    }
    
    return step, nil
}

func validateAndDeductPromotion(promoCode string) (*PromotionDeductionStep, error) {
    step := &PromotionDeductionStep{
        promoCode: promoCode,
    }
    
    // 执行促销活动处理
    if err := step.Execute(); err != nil {
        return nil, &OrderError{
            Code:    ErrPromotionExpired,
            Message: "Promotion validation failed",
            Step:    "PromotionDeduction",
        }
    }
    
    return step, nil
}

func CreateOrder(request OrderRequest) (*OrderResponse, error) {
    // 创建回滚管理器
    rollbackManager := &RollbackManager{}
    
    // 1. 用户校验
    userStep, err := validateUser(request.UserID)
    if err != nil {
        return nil, err
    }
    rollbackManager.steps = append(rollbackManager.steps, userStep)
    
    // 2. 商品校验
    productStep, err := validateProduct(request.ProductID)
    if err != nil {
        rollbackManager.Rollback()
        return nil, err
    }
    rollbackManager.steps = append(rollbackManager.steps, productStep)
    
    // 3. 库存校验和扣减
    stockStep, err := validateAndDeductStock(request.ProductID, request.Quantity)
    if err != nil {
        rollbackManager.Rollback()
        return nil, err
    }
    rollbackManager.steps = append(rollbackManager.steps, stockStep)
    
    // 4. 营销活动处理
    if request.PromoCode != "" {
        promoStep, err := validateAndDeductPromotion(request.PromoCode)
        if err != nil {
            rollbackManager.Rollback()
            return nil, err
        }
        rollbackManager.steps = append(rollbackManager.steps, promoStep)
    }
    
    // 5. 创建订单
    orderStep, err := createOrder(request)
    if err != nil {
        rollbackManager.Rollback()
        return nil, err
    }
    rollbackManager.steps = append(rollbackManager.steps, orderStep)
    
    return &OrderResponse{
        OrderID: orderStep.OrderID,
        Status:  "SUCCESS",
    }, nil
}

// 补偿任务
type CompensationTask struct {
    OrderID    string
    Step       string
    RetryCount int
    MaxRetries int
}

// 补偿执行器
func (ct *CompensationTask) Execute() error {
    if ct.RetryCount >= ct.MaxRetries {
        return errors.New("max retries exceeded")
    }
    
    // 根据步骤执行不同的补偿逻辑
    switch ct.Step {
    case "StockDeduction":
        return compensateStockDeduction(ct.OrderID)
    case "PromotionDeduction":
        return compensatePromotionDeduction(ct.OrderID)
    // ... 其他补偿逻辑
    }
    
    return nil
}


// 订单创建步骤接口
type OrderCreationStep interface {
    // 执行步骤
    Execute(ctx context.Context) error
    // 回滚步骤
    Rollback(ctx context.Context) error
    // 获取步骤名称
    GetStepName() string
}

// 订单创建上下文
type OrderCreationContext struct {
    OrderID      string
    UserID       string
    ProductID    string
    Quantity     int
    PromoCode    string
    PaymentInfo  PaymentInfo
    DeliveryInfo DeliveryInfo
    // 存储中间结果
    StepResults  map[string]interface{}
}

// 基础步骤实现
type BaseStep struct {
    ctx *OrderCreationContext
}

// 1. 用户校验步骤
type UserValidationStep struct {
    BaseStep
    userService UserService
}

func (s *UserValidationStep) Execute(ctx context.Context) error {
    // 1. 检查用户是否存在
    user, err := s.userService.GetUser(s.ctx.UserID)
    if err != nil {
        return fmt.Errorf("user not found: %v", err)
    }
    
    // 2. 检查用户状态
    if user.Status != UserStatusNormal {
        return fmt.Errorf("user status abnormal: %s", user.Status)
    }
    
    // 3. 存储校验结果
    s.ctx.StepResults["user"] = user
    return nil
}

func (s *UserValidationStep) Rollback(ctx context.Context) error {
    // 用户校验步骤通常不需要回滚
    return nil
}

func (s *UserValidationStep) GetStepName() string {
    return "UserValidation"
}

// 2. 商品校验步骤
type ProductValidationStep struct {
    BaseStep
    productService ProductService
}

func (s *ProductValidationStep) Execute(ctx context.Context) error {
    // 1. 获取商品信息
    product, err := s.productService.GetProduct(s.ctx.ProductID)
    if err != nil {
        return fmt.Errorf("product not found: %v", err)
    }
    
    // 2. 检查商品状态
    if product.Status != ProductStatusOnSale {
        return fmt.Errorf("product not on sale: %s", product.Status)
    }
    
    // 3. 存储商品信息
    s.ctx.StepResults["product"] = product
    return nil
}

func (s *ProductValidationStep) Rollback(ctx context.Context) error {
    // 商品校验步骤通常不需要回滚
    return nil
}

func (s *ProductValidationStep) GetStepName() string {
    return "ProductValidation"
}

// 3. 库存校验和扣减步骤
type StockDeductionStep struct {
    BaseStep
    inventoryService InventoryService
    lockService      LockService
}

func (s *StockDeductionStep) Execute(ctx context.Context) error {
    // 1. 获取分布式锁
    lock := s.lockService.NewLock("stock:" + s.ctx.ProductID)
    if !lock.TryLock() {
        return fmt.Errorf("failed to acquire stock lock")
    }
    defer lock.Unlock()
    
    // 2. 检查库存
    stock, err := s.inventoryService.GetStock(s.ctx.ProductID)
    if err != nil {
        return fmt.Errorf("failed to get stock: %v", err)
    }
    
    if stock < s.ctx.Quantity {
        return fmt.Errorf("insufficient stock")
    }
    
    // 3. 扣减库存
    if err := s.inventoryService.DeductStock(s.ctx.ProductID, s.ctx.Quantity); err != nil {
        return fmt.Errorf("failed to deduct stock: %v", err)
    }
    
    // 4. 记录扣减结果
    s.ctx.StepResults["stock_deduction"] = true
    return nil
}

func (s *StockDeductionStep) Rollback(ctx context.Context) error {
    // 回滚库存扣减
    if s.ctx.StepResults["stock_deduction"] == true {
        return s.inventoryService.ReturnStock(s.ctx.ProductID, s.ctx.Quantity)
    }
    return nil
}

func (s *StockDeductionStep) GetStepName() string {
    return "StockDeduction"
}

// 4. 营销活动处理步骤
type PromotionDeductionStep struct {
    BaseStep
    promotionService PromotionService
}

func (s *PromotionDeductionStep) Execute(ctx context.Context) error {
    if s.ctx.PromoCode == "" {
        return nil
    }
    
    // 1. 校验促销活动
    promotion, err := s.promotionService.ValidatePromotion(s.ctx.PromoCode)
    if err != nil {
        return fmt.Errorf("promotion validation failed: %v", err)
    }
    
    // 2. 扣减促销活动库存
    if err := s.promotionService.DeductPromotion(s.ctx.PromoCode); err != nil {
        return fmt.Errorf("failed to deduct promotion: %v", err)
    }
    
    // 3. 记录促销结果
    s.ctx.StepResults["promotion"] = promotion
    return nil
}

func (s *PromotionDeductionStep) Rollback(ctx context.Context) error {
    if s.ctx.PromoCode != "" && s.ctx.StepResults["promotion"] != nil {
        return s.promotionService.ReturnPromotion(s.ctx.PromoCode)
    }
    return nil
}

func (s *PromotionDeductionStep) GetStepName() string {
    return "PromotionDeduction"
}

// 5. 订单创建步骤
type OrderCreationStep struct {
    BaseStep
    orderService OrderService
}

func (s *OrderCreationStep) Execute(ctx context.Context) error {
    // 1. 构建订单信息
    order := &Order{
        OrderID:      s.ctx.OrderID,
        UserID:       s.ctx.UserID,
        ProductID:    s.ctx.ProductID,
        Quantity:     s.ctx.Quantity,
        PromoCode:    s.ctx.PromoCode,
        Status:       OrderStatusCreated,
        CreatedAt:    time.Now(),
    }
    
    // 2. 创建订单
    if err := s.orderService.CreateOrder(order); err != nil {
        return fmt.Errorf("failed to create order: %v", err)
    }
    
    // 3. 记录订单信息
    s.ctx.StepResults["order"] = order
    return nil
}

func (s *OrderCreationStep) Rollback(ctx context.Context) error {
    if order, ok := s.ctx.StepResults["order"].(*Order); ok {
        return s.orderService.CancelOrder(order.OrderID)
    }
    return nil
}

func (s *OrderCreationStep) GetStepName() string {
    return "OrderCreation"
}

// 订单创建流程管理器
type OrderCreationManager struct {
    steps []OrderCreationStep
    ctx   *OrderCreationContext
}

func NewOrderCreationManager(ctx *OrderCreationContext) *OrderCreationManager {
    return &OrderCreationManager{
        ctx: ctx,
        steps: []OrderCreationStep{
            &UserValidationStep{BaseStep{ctx: ctx}},
            &ProductValidationStep{BaseStep{ctx: ctx}},
            &StockDeductionStep{BaseStep{ctx: ctx}},
            &PromotionDeductionStep{BaseStep{ctx: ctx}},
            &OrderCreationStep{BaseStep{ctx: ctx}},
        },
    }
}

func (m *OrderCreationManager) Execute(ctx context.Context) error {
    // 记录已执行的步骤
    executedSteps := make([]OrderCreationStep, 0)
    
    // 按顺序执行步骤
    for _, step := range m.steps {
        if err := step.Execute(ctx); err != nil {
            // 发生错误时，从后往前回滚已执行的步骤
            for i := len(executedSteps) - 1; i >= 0; i-- {
                if rollbackErr := executedSteps[i].Rollback(ctx); rollbackErr != nil {
                    log.Printf("Failed to rollback step %s: %v", 
                        executedSteps[i].GetStepName(), rollbackErr)
                }
            }
            return fmt.Errorf("step %s failed: %v", step.GetStepName(), err)
        }
        executedSteps = append(executedSteps, step)
    }
    
    return nil
}


func CreateOrder(request OrderRequest) (*OrderResponse, error) {
    // 创建上下文
    ctx := &OrderCreationContext{
        OrderID:      generateOrderID(),
        UserID:       request.UserID,
        ProductID:    request.ProductID,
        Quantity:     request.Quantity,
        PromoCode:    request.PromoCode,
        StepResults:  make(map[string]interface{}),
    }
    
    // 创建订单创建管理器
    manager := NewOrderCreationManager(ctx)
    
    // 执行订单创建流程
    if err := manager.Execute(context.Background()); err != nil {
        return nil, err
    }
    
    // 获取创建的订单
    order := ctx.StepResults["order"].(*Order)
    
    return &OrderResponse{
        OrderID: order.OrderID,
        Status:  "SUCCESS",
    }, nil
}


// 1. 定义商品类目类型
type CategoryType string

const (
    CategoryTypePhysical    CategoryType = "PHYSICAL"    // 实物商品
    CategoryTypeVirtual     CategoryType = "VIRTUAL"     // 虚拟商品
    CategoryTypeSubscription CategoryType = "SUBSCRIPTION" // 订阅商品
    CategoryTypeService     CategoryType = "SERVICE"     // 服务类商品
)

// 2. 定义商品类目特定的创单步骤接口
type CategorySpecificStep interface {
    OrderCreationStep
    // 获取支持的类目类型
    GetSupportedCategories() []CategoryType
}

// 3. 定义不同类目的创单策略
type OrderCreationStrategy interface {
    // 获取类目特定的校验步骤
    GetValidationSteps() []CategorySpecificStep
    // 获取类目特定的库存步骤
    GetInventorySteps() []CategorySpecificStep
    // 获取类目特定的支付步骤
    GetPaymentSteps() []CategorySpecificStep
    // 获取类目特定的履约步骤
    GetFulfillmentSteps() []CategorySpecificStep
}

// 4. 实现不同类目的创单策略
// 4.1 实物商品策略
type PhysicalOrderStrategy struct {
    BaseStep
}

func (s *PhysicalOrderStrategy) GetValidationSteps() []CategorySpecificStep {
    return []CategorySpecificStep{
        &PhysicalProductValidationStep{BaseStep: s.BaseStep},
        &PhysicalInventoryValidationStep{BaseStep: s.BaseStep},
        &PhysicalDeliveryValidationStep{BaseStep: s.BaseStep},
    }
}

func (s *PhysicalOrderStrategy) GetInventorySteps() []CategorySpecificStep {
    return []CategorySpecificStep{
        &PhysicalStockDeductionStep{BaseStep: s.BaseStep},
        &PhysicalWarehouseSelectionStep{BaseStep: s.BaseStep},
    }
}

func (s *PhysicalOrderStrategy) GetPaymentSteps() []CategorySpecificStep {
    return []CategorySpecificStep{
        &PhysicalPaymentValidationStep{BaseStep: s.BaseStep},
    }
}

func (s *PhysicalOrderStrategy) GetFulfillmentSteps() []CategorySpecificStep {
    return []CategorySpecificStep{
        &PhysicalFulfillmentStep{BaseStep: s.BaseStep},
    }
}

// 4.2 虚拟商品策略
type VirtualOrderStrategy struct {
    BaseStep
}

func (s *VirtualOrderStrategy) GetValidationSteps() []CategorySpecificStep {
    return []CategorySpecificStep{
        &VirtualProductValidationStep{BaseStep: s.BaseStep},
        &VirtualInventoryValidationStep{BaseStep: s.BaseStep},
    }
}

func (s *VirtualOrderStrategy) GetInventorySteps() []CategorySpecificStep {
    return []CategorySpecificStep{
        &VirtualStockDeductionStep{BaseStep: s.BaseStep},
    }
}

func (s *VirtualOrderStrategy) GetPaymentSteps() []CategorySpecificStep {
    return []CategorySpecificStep{
        &VirtualPaymentValidationStep{BaseStep: s.BaseStep},
    }
}

func (s *VirtualOrderStrategy) GetFulfillmentSteps() []CategorySpecificStep {
    return []CategorySpecificStep{
        &VirtualFulfillmentStep{BaseStep: s.BaseStep},
    }
}

// 5. 策略工厂
type OrderStrategyFactory struct {
    strategies map[CategoryType]OrderCreationStrategy
}

func NewOrderStrategyFactory() *OrderStrategyFactory {
    return &OrderStrategyFactory{
        strategies: map[CategoryType]OrderCreationStrategy{
            CategoryTypePhysical:    &PhysicalOrderStrategy{},
            CategoryTypeVirtual:     &VirtualOrderStrategy{},
            CategoryTypeSubscription: &SubscriptionOrderStrategy{},
            CategoryTypeService:     &ServiceOrderStrategy{},
        },
    }
}

func (f *OrderStrategyFactory) GetStrategy(categoryType CategoryType) (OrderCreationStrategy, error) {
    if strategy, exists := f.strategies[categoryType]; exists {
        return strategy, nil
    }
    return nil, fmt.Errorf("unsupported category type: %s", categoryType)
}

// 6. 扩展订单创建上下文
type OrderCreationContext struct {
    // ... 原有字段 ...
    CategoryType CategoryType
    CategorySpecificData map[string]interface{}
}

// 7. 扩展订单创建管理器
type OrderCreationManager struct {
    steps []OrderCreationStep
    ctx   *OrderCreationContext
    strategy OrderCreationStrategy
}

func NewOrderCreationManager(ctx *OrderCreationContext) (*OrderCreationManager, error) {
    factory := NewOrderStrategyFactory()
    strategy, err := factory.GetStrategy(ctx.CategoryType)
    if err != nil {
        return nil, err
    }

    // 获取基础步骤
    baseSteps := []OrderCreationStep{
        &UserValidationStep{BaseStep{ctx: ctx}},
        &CommonProductValidationStep{BaseStep{ctx: ctx}},
    }

    // 获取类目特定步骤
    categorySteps := strategy.GetValidationSteps()
    inventorySteps := strategy.GetInventorySteps()
    paymentSteps := strategy.GetPaymentSteps()
    fulfillmentSteps := strategy.GetFulfillmentSteps()

    // 合并所有步骤
    allSteps := append(baseSteps, categorySteps...)
    allSteps = append(allSteps, inventorySteps...)
    allSteps = append(allSteps, paymentSteps...)
    allSteps = append(allSteps, fulfillmentSteps...)

    return &OrderCreationManager{
        steps: allSteps,
        ctx:   ctx,
        strategy: strategy,
    }, nil
}

// 8. 实现具体的类目特定步骤
// 8.1 实物商品特定步骤
type PhysicalProductValidationStep struct {
    BaseStep
}

func (s *PhysicalProductValidationStep) Execute(ctx context.Context) error {
    // 实物商品特定的校验逻辑
    // 1. 检查商品重量
    // 2. 检查商品尺寸
    // 3. 检查是否支持配送
    return nil
}

func (s *PhysicalProductValidationStep) GetSupportedCategories() []CategoryType {
    return []CategoryType{CategoryTypePhysical}
}

// 8.2 虚拟商品特定步骤
type VirtualProductValidationStep struct {
    BaseStep
}

func (s *VirtualProductValidationStep) Execute(ctx context.Context) error {
    // 虚拟商品特定的校验逻辑
    // 1. 检查激活码库存
    // 2. 检查有效期
    // 3. 检查使用限制
    return nil
}

func (s *VirtualProductValidationStep) GetSupportedCategories() []CategoryType {
    return []CategoryType{CategoryTypeVirtual}
}

func CreateOrder(request OrderRequest) (*OrderResponse, error) {
    // 创建上下文
    ctx := &OrderCreationContext{
        OrderID:      generateOrderID(),
        UserID:       request.UserID,
        ProductID:    request.ProductID,
        Quantity:     request.Quantity,
        PromoCode:    request.PromoCode,
        CategoryType: request.CategoryType,
        StepResults:  make(map[string]interface{}),
        CategorySpecificData: make(map[string]interface{}),
    }
    
    // 创建订单创建管理器
    manager, err := NewOrderCreationManager(ctx)
    if err != nil {
        return nil, err
    }
    
    // 执行订单创建流程
    if err := manager.Execute(context.Background()); err != nil {
        return nil, err
    }
    
    // 获取创建的订单
    order := ctx.StepResults["order"].(*Order)
    
    return &OrderResponse{
        OrderID: order.OrderID,
        Status:  "SUCCESS",
    }, nil
}
```




### 支付/支付回调

### 履约/履约回调



## 电商用户管理

### 核心功能
#### 用户注册
- 记录用户的基本信息
- 支持密码加密存储
- 发送验证码或邮件验证
- 通过邮箱、手机号或第三方账号注册

* 使用 bcrypt 对密码进行哈希存储、存储 salt，防止彩虹表攻击、防止弱密码（如 12345678）*

#### 用户登录
- 记住登录状态（JWT/Session）
- 通过邮箱、手机号+密码登录
- 支持 OAuth 登录（如 Google、微信、支付宝等）
- 账户锁定策略（防止暴力破解）

#### 其他功能（可扩展）
- 账户找回（忘记密码、重置密码）
- 用户权限管理（普通用户、VIP 用户、管理员）
- 多设备登录检测、账号安全管理（修改密码、绑定手机号、解绑社交账号）
- **社交登录（微信、Google、Apple ID）**  
- **用户等级 & 会员系统**  
- **黑名单风控（限制恶意 IP）**  
- **短信登录 & OAuth 认证**  


#### 注意事项
- 使用 **bcrypt** 对密码进行哈希存储
- 存储 `salt`，防止彩虹表攻击
- 防止**弱密码**（如 `12345678`）
- **JWT（JSON Web Token）**
  - 生成用户 Token 并存储在 `Authorization: Bearer <token>` 头部
  - 过期时间如 `7 天`
- **Session 认证**
  - 在 Redis 或数据库存储用户 Session
- **多次登录失败锁定账户**（5 次错误后，10 分钟内禁止登录）
- **短信/邮件验证码**（可选）
- **双因子认证（2FA）**（高级功能）

---

### 模型和数据库设计

#### 用户表（users）
```
CREATE TABLE users (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    username        VARCHAR(50) UNIQUE NOT NULL,
    email           VARCHAR(100) UNIQUE,
    phone           VARCHAR(20) UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    salt            VARCHAR(32) NOT NULL COMMENT '增强密码安全性，防止彩虹攻击',
    avatar          VARCHAR(255),
    status          TINYINT DEFAULT 1 COMMENT '1-正常, 0-禁用',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

#### 用户登录日志（user_login_logs）
```
CREATE TABLE user_login_logs (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id     BIGINT NOT NULL,
    login_ip    VARCHAR(50) NOT NULL,
    login_time  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    device      VARCHAR(100),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

---

###  核心流程以及API 设计

#### 用户注册 API
```
POST /api/register
Content-Type: application/json

{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "password123",
    "phone": "13812345678"
}

{
    "code": 200,
    "message": "注册成功"
}
```

#### 用户登录 API
```
POST /api/login
Content-Type: application/json

{
    "username": "john_doe",
    "password": "password123"
}

返回：
{
    "code": 200,
    "message": "登录成功",
    "token": "jwt_token_here"
}
```

## 5. 代码实现（Go + Gin + GORM 示例）

### 用户注册
```
go
package controllers

import (
	"ecommerce/models"
	"ecommerce/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 用户注册
func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}

	// 哈希加密密码
	hashedPassword, salt := utils.HashPassword(user.Password)
	user.Password = hashedPassword
	user.Salt = salt

	// 存入数据库
	if err := models.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "注册失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}
```

### 用户登录
```
go
func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	var user models.User
	if err := models.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户不存在"})
		return
	}

	// 校验密码
	if !utils.CheckPassword(req.Password, user.Password, user.Salt) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "密码错误"})
		return
	}

	// 生成 JWT Token
	token, _ := utils.GenerateToken(user.ID)
	c.JSON(http.StatusOK, gin.H{"message": "登录成功", "token": token})
}
```


### 核心业务流
### B 端
#### 首页运营和维护
#### 批量商品上传、编辑商品信息、价格、库存、状态
- mass/single upload
- mass/single edit
- verify，upload
- item sync fetch pull
- openapi

### APP端
#### 首页获取
#### 商品搜索
#### 
要求：
- 海量的数据，亿级别的商品量；
- 高并发查询，日 PV 过亿；
- 请求需要快速响应
特点：
- 商品数据已经结构化，但散布在商品、库存、价格、促销、仓储等多个系统
- 召回率要求高，保证每一个正常的商品均能够被搜索到
- 为保证用户体验，商品信息变更（比如价格、库存的变化）实时性要求高，导致更新量大，每天的更新量为千万级别
- 较强的个性化需求，由于是一个相对垂直的搜索领域，需要满足用户的个性化搜索意图，比如用户搜索“小说”有的用户希望找言情小说有的人需要找武侠小说有的人希望找到励志小说

#### 商品（商品详情）

#### 创单核心逻辑
- 用户校验
- 商品信息获取和校验
- 价格校验
- 营销活动校验
- antifraud
- 库存校验
- 生成payorderid和orderid
- 库存扣减和返还
- 营销活动扣减和返还
- 构建订单信息，插入DB
- 不同类型的创单逻辑会不同，这里通过接口定义基础的创单逻辑，后续不同类型的定义机遇这个逻辑扩展
```Go
package orderserver

// OrderRequest 包含创建订单所需的参数
type OrderRequest struct {
	UserID    string
	ProductID string
	Quantity  int
}

// OrderResponse 表示创建订单的响应
type OrderResponse struct {
	OrderID string
	Message string // 返回的信息，例如错误信息
}

// OrderServer 接口定义了创建订单的功能
type OrderServer interface {
	// validate
	ValidateUser(userID string) errors.ErrorCode
	GetProductInfo(productID string) (ProductInfo, errors.ErrorCode)
	ValidateProduct(productID string) errors.ErrorCode
	ValidatePrice(productID string) errors.ErrorCode
	ValidateInventory(productID string, quantity int) errors.ErrorCode
	ValidatePromotionCode(promoCode string) errors.ErrorCode
	CheckFraud() errors.ErrorCode

	GeneratePayOrderID() (string, errors.ErrorCode)
	GenerateOrderID() (string, errors.ErrorCode)

	DeductInventory(productID string, quantity int) errors.ErrorCode
	ReturnInventory(productID string, quantity int) errors.ErrorCode

	DeductPromotion(promoCode string) errors.ErrorCode
	ReturnPromotion(promoCode string) errors.ErrorCode

  BuildDBModels() errors.ErrorCode
	InsertOrder(order OrderRequest) (OrderResponse, error)
	LogOperation(orderID string, userID string) error

  PushOrderCreateEvent() errors.ErrorCode
}

// BaseOrderService 实现 OrderServer 接口
type BaseOrderService struct {
	// 可以添加数据库连接或其他依赖项
	req  OrderRequest
	resp OrderResponse

	OrderModel      *order.Model
  PayOrderModel   *order.PayModel
	OrderItemModels []*item.OrderItemModel
}

func (bos *BaseOrderService) ValidateUser(userID string) ErrorCode {
	// 简单的用户验证逻辑（示例）
	return Success
}
....
```
#### 订单支付和支付结果回调
<p align="center">
  <img src="/images/order_pay.png" width=500 height=1000>
</p>

```Go
  type OrderPayRequest struct {
      UserID      string
      OrderID     string
      PaymentMethod string // 支付方式，例如信用卡、支付宝等
      EVoucherCode string // 可选的电子券代码
      Coins        int     // 可用积分
  }

  // OrderPayResponse 表示支付请求的响应
  type OrderPayResponse struct {
      Success     bool
      Message     string
      PaymentID   string // 支付订单ID
  }

  // OrderPayService 接口定义了支付相关的功能
  type OrderPayService interface {
      ValidateUser(userID string) ErrorCode
      ValidateOrderStatus(orderID string) ErrorCode
      ValidatePrice(orderID string) ErrorCode
      ValidatePromotionCode(promoCode string) ErrorCode
      ValidateEVoucher(evoucherCode string) ErrorCode
      ValidateCoins(coins int) ErrorCode
      ValidatePaymentMethod(paymentMethod string) ErrorCode
      ValidatePaymentFees(orderID string, paymentMethod string) ErrorCode

      RedeemEVoucher(evoucherCode string) ErrorCode
      DeductCoins(coins int) ErrorCode

      InitializePayment(orderID string, paymentMethod string) (OrderPayResponse, ErrorCode)
      ConstructPaymentRequest(orderID string) (OrderPayResponse, ErrorCode)
      UpdateOrderStatus(orderID string, status string) ErrorCode
      HandleError(orderID string, err error) ErrorCode
  }
```

#### 订单履约和履约结果回调
<p align="center">
  <img src="/images/order_fulfillment.png" width=500 height=1000>
</p>

```Go
package fulfillmentserver

// OrderFulfillmentRequest 包含履行请求所需的参数
type OrderFulfillmentRequest struct {
    OrderID  string
    UserID   string
    Quantity int
}

// OrderFulfillmentResponse 表示履行请求的响应
type OrderFulfillmentResponse struct {
    Success   bool
    Message   string
    TrackingID string // 物流跟踪ID
}

// FulfillmentService 接口定义了订单履行相关的功能
type FulfillmentService interface {
    ValidateStock(orderID string, quantity int) ErrorCode
    ProcessOrder(request OrderFulfillmentRequest) (OrderFulfillmentResponse, ErrorCode)
    UpdateOrderStatus(orderID string, status string) ErrorCode
    HandleDelivery(orderID string) ErrorCode
    HandleError(orderID string, err error) ErrorCode
}
```
#### return & refund
<p align="center">
  <img src="/images/return&refund.png" width=500 height=1000>
</p>

- UserRefundOrderService、AdminRefundOrderService、FailedFulfillmentRefundOrderService
- Return
- refund

##### RefundPlaceOrder
```Go
package refundservice

// RefundOrderRequest 包含退款请求所需的参数
type RefundOrderRequest struct {
    OrderID  string
    UserID   string
    Amount   float64
}

// RefundOrderResponse 表示退款请求的响应
type RefundOrderResponse struct {
    Success   bool
    Message   string
    RefundID  string // 退款ID
}

// RefundOrderService 接口定义了退款相关的功能
type RefundOrderService interface {
    ValidateRefund(request RefundOrderRequest) ErrorCode
    PlaceOrder(request RefundOrderRequest) (RefundOrderResponse, ErrorCode)
}

// BaseRefundOrderService 实现 RefundOrderService 接口
type BaseRefundOrderService struct{}

func (bros *BaseRefundOrderService) ValidateRefund(request RefundOrderRequest) ErrorCode {
    if request.Amount <= 0 {
        return ErrInvalidAmount
    }
    if !orderExists(request.OrderID) {
        return ErrOrderNotFound
    }
    return Success
}
// 假设的辅助函数
func orderExists(orderID string) bool {
    // 检查订单是否存在的逻辑
    return true // 示例返回
}

func initiateRefund(orderID string, amount float64) string {
    // 处理退款并返回退款ID的逻辑
    return "refund123" // 示例返回
}


// UserRefundOrderService 实现用户退款订单的逻辑
type UserRefundOrderService struct {
    baseService *BaseRefundOrderService
}

func (uros *UserRefundOrderService) ValidateRefund(request RefundOrderRequest) ErrorCode {
    return uros.baseService.ValidateRefund(request)
}

func (uros *UserRefundOrderService) PlaceOrder(request RefundOrderRequest) (RefundOrderResponse, ErrorCode) {
    if errCode := uros.ValidateRefund(request); errCode != Success {
        return RefundOrderResponse{}, errCode
    }

    // 处理用户创建退款订单的逻辑
    refundID := initiateRefund(request.OrderID, request.Amount)

    return RefundOrderResponse{
        Success:  true,
        Message:  "User refund order created successfully.",
        RefundID: refundID,
    }, Success
}

// AdminRefundOrderService 实现管理员退款订单的逻辑
type AdminRefundOrderService struct {
    baseService *BaseRefundOrderService
}

func (aros *AdminRefundOrderService) ValidateRefund(request RefundOrderRequest) ErrorCode {
    return aros.baseService.ValidateRefund(request)
}

func (aros *AdminRefundOrderService) PlaceOrder(request RefundOrderRequest) (RefundOrderResponse, ErrorCode) {
    if errCode := aros.ValidateRefund(request); errCode != Success {
        return RefundOrderResponse{}, errCode
    }

    // 处理管理员创建退款订单的逻辑
    refundID := initiateRefund(request.OrderID, request.Amount)

    return RefundOrderResponse{
        Success:  true,
        Message:  "Admin refund order created successfully.",
        RefundID: refundID,
    }, Success
}

// FailedDeliveryRefundOrderService 实现发货失败退款订单的逻辑
type FailedFulfillmentRefundOrderService struct {
    baseService *BaseRefundOrderService
}

func (fdros *FailedFulfillmentRefundOrderService) HandleFailedDelivery(orderID string) (RefundOrderResponse, ErrorCode) {
    // 假设处理发货失败的逻辑
    refundRequest := RefundOrderRequest{
        OrderID: orderID,
        UserID:  "system", // 系统自动处理
        Amount:  0.0,      // 假设金额为0，具体金额需要根据业务逻辑设置
    }

    // 处理退款
    refundID := initiateRefund(refundRequest.OrderID, refundRequest.Amount)

    return RefundOrderResponse{
        Success:  true,
        Message:  "Refund order created due to failed delivery.",
        RefundID: refundID,
    }, Success
}
```

##### RefundApproveService
```Go
// RefundApproveRequest 包含退款审批请求所需的参数
type RefundApproveRequest struct {
    RefundID string
    Approve  bool
}

// RefundApproveResponse 表示退款审批的响应
type RefundApproveResponse struct {
    Success bool
    Message string
}

// RefundApproveService 接口定义了退款审批相关的功能
type RefundApproveService interface {
    ApproveRefund(request RefundApproveRequest) (RefundApproveResponse, ErrorCode)
}
```

##### ReturnPurchaseService
```Go
// ReturnPurchaseRequest 包含退货请求所需的参数
type ReturnPurchaseRequest struct {
    OrderID  string
    UserID   string
    Reason   string
    Amount   float64
}

// ReturnPurchaseResponse 表示退货请求的响应
type ReturnPurchaseResponse struct {
    Success   bool
    Message   string
    ReturnID  string // 退货ID
}
// ReturnPurchaseService 接口定义了退货相关的功能
type ReturnPurchaseService interface {
    ValidateReturn(request ReturnPurchaseRequest) ErrorCode
    ProcessReturn(request ReturnPurchaseRequest) (ReturnPurchaseResponse, ErrorCode)
}
```

##### RefundService
```Go
package refundservice

// RefundRequest 包含退款请求所需的参数
type RefundRequest struct {
    OrderID string
    UserID  string
    Amount  float64
}

// RefundResponse 表示退款请求的响应
type RefundResponse struct {
    Success  bool
    Message  string
    RefundID string // 退款ID
}

// RefundService 接口定义了退款相关的功能
type RefundService interface {
    ValidateRefund(request RefundRequest) ErrorCode
    ProcessRefund(request RefundRequest) (RefundResponse, ErrorCode)
}
```
#### 订单详情查询
## 系统挑战
### 如何维护订单状态的最终一致性？
<p align="center">
  <img src="/images/order_final_consistency_activity.png" width=600 height=600>
</p>

- 状态机一定要设计好，只有特定的原始状态 + 特定的事件才可以推进到指定的状态。
- 并发更新数据库前，要用乐观锁或者悲观锁，先使用select for update进行锁行记录，同时在更新时判断版本号是否是之前取出来的版本号，更新成功就结束，更新失败就组成消息发到消息队列，后面再消费。
- 通过补偿机制兜底，比如查询补单。
- 通过上述三个步骤，正常情况下，最终的数据状态一定是正确的。除非是某个系统有异常，比如外部渠道开始返回支付成功，然后又返回支付失败，说明依赖的外部系统已经异常，这样只能进人工差错处理流程。


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

3. 利用数据库自身特性 “主键唯一约束”，在插入订单记录时，带上主键值，如果订单重复，记录插入会失败。
操作过程如下：
引入一个服务，用于生成一个“全局唯一的订单号”；
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


### 电商系统稳定性建设

## 前言：怎样的系统算是稳定高可用的

首先回答另一个问题，怎样的系统算是稳定的？

Google SRE中(SRE三部曲[1])有一个层级模型来描述系统可靠性基础和高层次需求(Dickerson's Hierarchy of Service Reliability)，如下图：

<p align="center">
  <img src="/images/service-reliability-hierarchy.png" width=600 height=500>
  <br/>
</p>


该模型由Google SRE工程师Mikey Dickerson在2013年提出，将系统稳定性需求按照基础程度进行了不同层次的体系化区分，形成稳定性标准金字塔模型:
- 金字塔的底座是监控(Monitoring)，这是一个系统对于稳定性最基础的要求，缺少监控的系统，如同蒙上眼睛狂奔的野马，无从谈及可控性，更遑论稳定性。
- 更上层是应急响应(Incident Response)，从一个问题被监控发现到最终解决，这期间的耗时直接取决于应急响应机制的成熟度。合理的应急策略能保证当故障发生时，所有问题能得到有序且妥善的处理，而不是慌乱成一锅粥。
- 事后总结以及根因分析(Postmortem&Root Caue Analysis)，即我们平时谈到的“复盘”，虽然很多人都不太喜欢这项活动，但是不得不承认这是避免我们下次犯同样错误的最有效手段，只有当摸清故障的根因以及对应的缺陷，我们才能对症下药，合理进行规避。
- 测试和发布管控(Testing&Release procedures),大大小小的应用都离不开不断的变更与发布,有效的测试与发布策略能保障系统所有新增变量都处于可控稳定区间内，从而达到整体服务终态稳定
- 容量规划(Capacity Planning)则是针对于这方面变化进行的保障策略。现有系统体量是否足够支撑新的流量需求，整体链路上是否存在不对等的薄弱节点，都是容量规划需要考虑的问题。
- 位于金字塔模型最顶端的是产品设计(Product)与软件研发(Development)，即通过优秀的产品设计与软件设计使系统具备更高的可靠性，构建高可用产品架构体系，从而提升用户体验


## 系统稳定性建设概述
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


## 高可用的架构设计

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