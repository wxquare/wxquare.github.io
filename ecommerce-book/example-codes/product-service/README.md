# Product Service - DDD 分层架构示例

## 📖 项目介绍

这是一个基于 **DDD（Domain-Driven Design）四层架构** 的商品服务示例，展示了从HTTP/gRPC/Event接口到数据持久化的完整数据流转。

**特别关注：** ⭐️ **事件订阅者的分层设计**（Interface Layer作为异步接口）

## 🏗️ 架构设计

### 四层架构

```
┌─────────────────────────────────────────────────────────────┐
│ Interface Layer (接口层)                                      │
│  - HTTP Handler        (同步接口)                             │
│  - gRPC Handler        (同步接口)                             │
│  - Event Subscriber ⭐️ (异步接口) - 本项目重点                │
├─────────────────────────────────────────────────────────────┤
│ Application Layer (应用层)                                    │
│  - Product Service     (业务编排)                             │
│  - DTO                 (数据传输对象)                          │
├─────────────────────────────────────────────────────────────┤
│ Domain Layer (领域层)                                         │
│  - Product             (聚合根)                               │
│  - SPU                 (实体)                                 │
│  - Value Objects       (值对象)                               │
│  - Domain Events       (领域事件)                             │
│  - Repository          (接口定义)                             │
├─────────────────────────────────────────────────────────────┤
│ Infrastructure Layer (基础设施层)                             │
│  - Product Repository  (数据持久化)                           │
│  - Cache               (三级缓存)                             │
│  - Kafka Producer   ⭐️ (事件发布 - 技术实现)                  │
│  - Kafka Consumer   ⭐️ (事件消费 - 技术实现)                  │
└─────────────────────────────────────────────────────────────┘
```

### ⭐️ 核心亮点：事件订阅者的分层

本项目展示了DDD中事件订阅者的推荐分层方式：

**设计原则**：事件订阅者是"异步接口"，与HTTP/gRPC同级

```
Kafka Topic
  ↓ 异步消息
Infrastructure Layer (KafkaConsumer) ← 技术实现（Kafka连接、消息接收）
  ↓ 消息路由
Interface Layer (ProductEventHandler) ← 协议适配（Kafka消息 → DTO）
  ↓ DTO
Application Layer (ProductService) ← 业务编排（复用HTTP/gRPC的同一个Service）
  ↓ Domain Model
Domain Layer (Product) ← 业务规则
  ↓ Repository
Infrastructure Layer (ProductRepository) ← 数据持久化
```

**详细说明**：详见 [EVENT_SUBSCRIBER_LAYER.md](EVENT_SUBSCRIBER_LAYER.md)

## 📁 目录结构

```
product-service/
├── cmd/
│   └── main.go                          # 程序入口（依赖注入容器）
├── internal/
│   ├── interfaces/                      # 接口层
│   │   ├── http/
│   │   │   └── product_handler.go       # HTTP Handler（同步）
│   │   ├── grpc/
│   │   │   ├── proto/product.proto      # Protobuf定义
│   │   │   └── product_handler.go       # gRPC Handler（同步）
│   │   └── event/ ⭐️
│   │       └── product_event_handler.go # Event Handler（异步接口）
│   │
│   ├── application/                     # 应用层
│   │   ├── service/
│   │   │   └── product_service.go       # 应用服务（业务编排）
│   │   └── dto/
│   │       └── product_dto.go           # DTO（请求/响应）
│   │
│   ├── domain/                          # 领域层
│   │   ├── product.go                   # Product聚合根
│   │   ├── spu.go                       # SPU实体
│   │   ├── value_objects.go             # 值对象（Price, SKU_ID等）
│   │   ├── events.go                    # 领域事件
│   │   └── repository.go                # Repository接口
│   │
│   └── infrastructure/                  # 基础设施层
│       ├── persistence/
│       │   ├── data_object.go           # DO（数据对象）
│       │   └── product_repository.go    # Repository实现
│       ├── cache/
│       │   └── cache.go                 # 三级缓存实现
│       └── messaging/ ⭐️
│           ├── kafka_producer.go        # Kafka生产者（事件发布）
│           └── kafka_consumer.go        # Kafka消费者（事件消费）
└── go.mod
```

## 🔑 核心特性

### 1. 完整的DDD战术设计

- ✅ **聚合根（Product）**：封装业务不变量，管理生命周期
- ✅ **值对象（Price, SKU_ID, Specifications）**：不可变、无身份
- ✅ **实体（SPU）**：有身份标识
- ✅ **领域事件（ProductOnShelfEvent, PriceChangedEvent）**：记录业务事实
- ✅ **Repository模式**：隔离基础设施细节

### 2. 三级缓存架构

```
L1 (Local Cache - sync.Map) → 进程内缓存，最快
  ↓ Miss
L2 (Redis) → 分布式缓存，中速
  ↓ Miss
L3 (MySQL) → 数据库，最慢
```

**Cache-Aside Pattern**：写操作删除缓存，由下次读操作回填

### 3. 事件驱动架构

#### 事件发布（Domain Events）

```go
// Domain Layer: 记录事件
product.OnShelf() // 产生 ProductOnShelfEvent

// Application Layer: 发布事件
service.publishDomainEvents(ctx, product)

// Infrastructure Layer: 发送到Kafka
kafkaProducer.Publish(ctx, event)
```

#### ⭐️ 事件订阅（推荐分层）

```go
// Infrastructure Layer: Kafka Consumer（技术实现）
kafkaConsumer.Start(ctx) // 订阅Topic、接收消息
  ↓
kafkaConsumer.routeMessage(ctx, messageType, data)
  ↓

// Interface Layer: Event Handler（协议适配）
eventHandler.HandleMessage(ctx, messageType, data)
  ↓
// 反序列化Kafka消息
json.Unmarshal(data, &kafkaEvent)
  ↓
// Kafka消息 → DTO
req := &dto.CreateProductRequest{...}
  ↓

// Application Layer: Product Service（业务编排）
productService.CreateProduct(ctx, req) // 复用同一个方法！
```

**关键点**：
- 📍 Event Handler 放在 **Interface Layer**（异步接口）
- 📍 Kafka Consumer 放在 **Infrastructure Layer**（技术实现）
- 📍 Application Service 被所有接口复用（HTTP、gRPC、Event）

### 4. 代码分层严格

- ✅ Domain Layer 无任何外部依赖（纯Go）
- ✅ Application Layer 依赖Domain接口，不依赖Infrastructure
- ✅ Infrastructure Layer 实现Domain定义的接口
- ✅ Interface Layer 调用Application Service（不直接访问Domain）

## 🚀 快速开始

### 运行Demo

```bash
cd /Users/wxquare/go/src/github.com/wxquare.github.io/ecommerce-book/examples/product-service
go run cmd/main.go
```

### Demo包含三个场景

#### Demo 1: 查询商品（三级缓存）

展示 L1 Miss → L2 Miss → L3 Hit 的完整缓存流程

#### Demo 2: 事件发布（领域事件）

展示商品上架/价格变更如何产生领域事件并发布到Kafka

#### Demo 3: 事件订阅（接收外部事件）⭐️

展示如何接收供应商服务/定价服务发送的事件，并调用Application Service处理

## 📚 详细文档

- **[QUICKSTART.md](QUICKSTART.md)** - 10分钟快速上手指南
- **[EVENT_PATTERN.md](EVENT_PATTERN.md)** - 事件发布和订阅的详细说明
- **[EVENT_SUBSCRIBER_LAYER.md](EVENT_SUBSCRIBER_LAYER.md)** ⭐️ - **事件订阅者的分层设计**（推荐阅读）

## 🎯 学习要点

### 1. 事件订阅者应该放在哪一层？

**推荐答案**：Interface Layer（异步接口）

**原因**：
- ✅ 与HTTP/gRPC同级，都是外部触发源
- ✅ 复用Application Service，业务逻辑只写一次
- ✅ 职责清晰：Interface负责协议适配，Infrastructure负责技术实现
- ✅ 易于测试和替换（换成RabbitMQ只需修改Infrastructure）

详见 [EVENT_SUBSCRIBER_LAYER.md](EVENT_SUBSCRIBER_LAYER.md)

### 2. 数据流转路径

#### HTTP同步调用

```
Client → HTTP Handler → Product Service → Product (Domain) → Repository → Cache/DB
```

#### Event异步调用

```
Kafka → Kafka Consumer (Infra) → Event Handler (Interface) → Product Service → Product (Domain) → Repository → Cache/DB
```

**关键差异**：Event多一层Infrastructure（Kafka技术实现）

### 3. DDD分层依赖方向

```
Interface → Application → Domain ← Infrastructure
                            ↑
                         (依赖倒置)
```

- Infrastructure **实现** Domain定义的接口
- 所有依赖指向Domain（依赖倒置原则）

## 🔧 实际项目建议

### 1. 服务分离部署

```
product-service-api (处理HTTP/gRPC)
├── 10副本
└── 职责：对外接口

product-service-consumer (处理Kafka事件)
├── 3副本
├── Consumer Group: product-service-group
└── 职责：消费事件

共享：
├── Application Service
├── Domain Model
└── Repository
```

### 2. Outbox Pattern（保证一致性）

```sql
CREATE TABLE domain_event_outbox (
    id BIGINT PRIMARY KEY,
    aggregate_type VARCHAR(50),
    aggregate_id BIGINT,
    event_type VARCHAR(100),
    event_data JSON,
    created_at TIMESTAMP,
    published_at TIMESTAMP,
    status ENUM('PENDING', 'PUBLISHED')
);
```

### 3. 事件幂等性

```go
// 使用event_id做幂等性检查
func (h *ProductEventHandler) handlePriceChanged(ctx, data) error {
    var event PriceChangedEvent
    json.Unmarshal(data, &event)
    
    // 检查是否已处理
    if h.isEventProcessed(ctx, event.EventID) {
        return nil // 幂等性：已处理，直接返回
    }
    
    // 处理业务逻辑...
    
    // 记录已处理
    h.markEventProcessed(ctx, event.EventID)
    
    return nil
}
```

## 🤝 贡献

欢迎提Issue或PR改进代码示例！

## 📄 许可证

MIT
