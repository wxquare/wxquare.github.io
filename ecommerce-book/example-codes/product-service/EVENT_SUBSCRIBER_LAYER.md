# DDD中事件订阅者的分层设计 🏗️

## 核心问题

**事件订阅者（Event Subscriber）应该放在哪一层？**

这个问题等同于：**异步事件驱动的入口应该放在哪里？**

---

## 两种主流方案对比

### 方案A：订阅者在Interface Layer（推荐）⭐️

**设计思想**：事件订阅者是"异步的接口层"，与HTTP/gRPC同级

```
┌─────────────────────────────────────────────────────────────┐
│              Interface Layer (接口层)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │ HTTP Handler │  │ gRPC Handler │  │Event Subscriber│     │
│  │              │  │              │  │ (异步接口)    │       │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘       │
│         │ 同步调用          │ 同步调用          │ 异步消费      │
└─────────┼──────────────────┼──────────────────┼─────────────┘
          ↓                  ↓                  ↓
┌─────────────────────────────────────────────────────────────┐
│              Application Layer (应用层)                        │
│  ┌──────────────────────────────────────────────────────────┐│
│  │         同一个 Application Service                         ││
│  │    productService.GetProduct(req)                         ││
│  │    productService.OnShelf(req)                            ││
│  └──────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

**优点**：
- ✅ **统一的入口抽象**：HTTP、gRPC、Event都是触发源
- ✅ **复用Application Service**：业务逻辑只写一次
- ✅ **职责清晰**：Interface Layer负责"适配外部协议"
- ✅ **易于测试**：直接测试Application Service

**缺点**：
- ⚠️ 需要理解"事件订阅也是接口"的概念

---

### 方案B：订阅者在Infrastructure Layer

**设计思想**：Kafka Consumer是技术实现，放在基础设施层

```
┌─────────────────────────────────────────────────────────────┐
│              Interface Layer (接口层)                          │
│  ┌──────────────┐  ┌──────────────┐                          │
│  │ HTTP Handler │  │ gRPC Handler │                          │
│  └──────┬───────┘  └──────┬───────┘                          │
└─────────┼──────────────────┼─────────────────────────────────┘
          ↓                  ↓
┌─────────────────────────────────────────────────────────────┐
│              Application Layer (应用层)                        │
│         productService.GetProduct(req)                        │
└─────────────────────────────────────────────────────────────┘
          ↑
          │ 反向依赖（不推荐）
          │
┌─────────┴───────────────────────────────────────────────────┐
│        Infrastructure Layer (基础设施层)                       │
│  ┌──────────────────────────────────────────────────────────┐│
│  │ KafkaConsumer.OnMessage() {                              ││
│  │     // 直接调用Application Service                        ││
│  │     productService.OnShelf(...)                          ││
│  │ }                                                         ││
│  └──────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

**优点**：
- ✅ Kafka Consumer确实是技术实现

**缺点**：
- ❌ **违反依赖倒置原则**：Infrastructure依赖Application
- ❌ 职责不清晰：Infrastructure既是实现层又是入口层
- ❌ 难以扩展：如果换成RabbitMQ需要大量修改

---

## 推荐方案A的完整实现

### 目录结构

```
internal/
├── interfaces/                        # 接口层（统一入口）
│   ├── http/
│   │   └── product_handler.go        # HTTP同步接口
│   ├── grpc/
│   │   └── product_handler.go        # gRPC同步接口
│   └── event/                         # ⭐️ 事件异步接口
│       └── product_event_handler.go   # Kafka消费者Handler
│
├── application/
│   └── service/
│       └── product_service.go         # 应用服务（被所有接口复用）
│
├── infrastructure/
│   ├── messaging/                     # ⭐️ 消息技术实现
│   │   ├── kafka_consumer.go         # Kafka消费者（技术细节）
│   │   └── kafka_producer.go         # Kafka生产者（技术细节）
│   ├── persistence/
│   │   └── product_repository.go
│   └── cache/
│       └── cache.go
│
└── domain/
    └── product.go
```

---

## 完整代码示例

### Step 1: Interface Layer - 事件订阅接口

```go
// internal/interfaces/event/product_event_handler.go

package event

import (
	"context"
	"encoding/json"
	"fmt"

	"product-service/internal/application/dto"
	"product-service/internal/application/service"
)

// ProductEventHandler 商品事件处理器（接口层）
// 职责：适配Kafka消息 → 调用Application Service
type ProductEventHandler struct {
	productService *service.ProductService
}

func NewProductEventHandler(productService *service.ProductService) *ProductEventHandler {
	return &ProductEventHandler{
		productService: productService,
	}
}

// HandleProductCreatedEvent 处理商品创建事件（来自其他服务）
// 例如：供应商服务发布"supplier.product.created"事件
func (h *ProductEventHandler) HandleProductCreatedEvent(ctx context.Context, data []byte) error {
	fmt.Printf("\n🔔 [Interface Layer - Event] Received: supplier.product.created\n")

	// 解析事件
	var event struct {
		SupplierID   int64  `json:"supplier_id"`
		SupplierSKU  string `json:"supplier_sku"`
		Title        string `json:"title"`
		BasePrice    int64  `json:"base_price"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("解析事件失败: %w", err)
	}

	// 转换事件数据 → DTO
	req := &dto.CreateProductRequest{
		SupplierID:  event.SupplierID,
		SupplierSKU: event.SupplierSKU,
		Title:       event.Title,
		BasePrice:   event.BasePrice,
	}

	// 调用应用服务（与HTTP/gRPC调用同一个方法）
	resp, err := h.productService.CreateProduct(ctx, req)
	if err != nil {
		return fmt.Errorf("创建商品失败: %w", err)
	}

	fmt.Printf("✅ [Interface Layer - Event] Product created, SKUID=%d\n", resp.SKUID)
	return nil
}

// HandlePriceChangedEvent 处理价格变更事件（来自定价服务）
func (h *ProductEventHandler) HandlePriceChangedEvent(ctx context.Context, data []byte) error {
	fmt.Printf("\n🔔 [Interface Layer - Event] Received: pricing.price_changed\n")

	var event struct {
		SKUID    int64 `json:"sku_id"`
		NewPrice int64 `json:"new_price"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	// 转换事件数据 → DTO
	req := &dto.UpdatePriceRequest{
		SKUID:    event.SKUID,
		NewPrice: event.NewPrice,
	}

	// 调用应用服务
	_, err := h.productService.UpdateBasePrice(ctx, req)
	return err
}
```

**关键点**：
- 📍 放在 `interfaces/event/` 目录
- 📍 职责：**协议适配**（Kafka消息 → DTO）
- 📍 调用Application Service（**复用业务逻辑**）

---

### Step 2: Infrastructure Layer - Kafka Consumer（技术实现）

```go
// internal/infrastructure/messaging/kafka_consumer.go

package messaging

import (
	"context"
	"fmt"

	eventHandler "product-service/internal/interfaces/event"
	// "github.com/confluentinc/confluent-kafka-go/kafka"
)

// KafkaConsumer Kafka消费者（技术实现）
// 职责：订阅Topic、接收消息、调用Interface Layer的Handler
type KafkaConsumer struct {
	// consumer *kafka.Consumer  // 实际项目使用
	handler *eventHandler.ProductEventHandler
}

func NewKafkaConsumer(handler *eventHandler.ProductEventHandler) *KafkaConsumer {
	return &KafkaConsumer{
		handler: handler,
	}
}

// Start 启动消费者
func (c *KafkaConsumer) Start(ctx context.Context) error {
	fmt.Println("📡 [Infrastructure - Kafka Consumer] Starting...")

	// 订阅Topic
	topics := []string{
		"supplier-product-events",  // 供应商商品事件
		"pricing-events",           // 定价事件
	}

	// 实际项目：
	// c.consumer.SubscribeTopics(topics, nil)
	// for {
	//     msg, err := c.consumer.ReadMessage(-1)
	//     if err != nil {
	//         continue
	//     }
	//     c.routeMessage(ctx, msg)
	// }

	fmt.Printf("✅ [Infrastructure - Kafka Consumer] Subscribed to topics: %v\n", topics)
	return nil
}

// routeMessage 路由消息到对应的Handler
func (c *KafkaConsumer) routeMessage(ctx context.Context, msg *Message) error {
	switch msg.Key {
	case "supplier.product.created":
		return c.handler.HandleProductCreatedEvent(ctx, msg.Value)
		
	case "pricing.price_changed":
		return c.handler.HandlePriceChangedEvent(ctx, msg.Value)
		
	default:
		fmt.Printf("⚠️  Unknown message type: %s\n", msg.Key)
		return nil
	}
}

// Message 简化的消息结构
type Message struct {
	Key   string
	Value []byte
}
```

**关键点**：
- 📍 放在 `infrastructure/messaging/` 目录
- 📍 职责：**技术实现**（Kafka连接、消息接收）
- 📍 调用Interface Layer的EventHandler（**不直接调用Application Service**）

---

### Step 3: Main程序 - 启动消费者

```go
// cmd/main.go

func main() {
	// 初始化依赖
	dependencies := initDependencies()
	
	// 启动HTTP服务器
	go startHTTPServer(dependencies.httpHandler)
	
	// ⭐️ 启动Kafka消费者
	go startKafkaConsumer(dependencies.kafkaConsumer)
	
	// 阻塞主goroutine
	select {}
}

func initDependencies() *Dependencies {
	// Infrastructure Layer
	localCache := cache.NewLocalCache()
	repo := persistence.NewProductRepository(localCache, ...)
	
	// Application Layer
	productService := service.NewProductService(repo, ...)
	
	// Interface Layer - HTTP
	httpHandler := httpHandler.NewProductHandler(productService)
	
	// Interface Layer - Event ⭐️
	eventHandler := eventHandler.NewProductEventHandler(productService)
	
	// Infrastructure Layer - Kafka Consumer
	kafkaConsumer := messaging.NewKafkaConsumer(eventHandler)
	
	return &Dependencies{
		httpHandler:   httpHandler,
		eventHandler:  eventHandler,
		kafkaConsumer: kafkaConsumer,
	}
}

func startKafkaConsumer(consumer *messaging.KafkaConsumer) {
	ctx := context.Background()
	if err := consumer.Start(ctx); err != nil {
		fmt.Printf("❌ Kafka Consumer error: %v\n", err)
	}
}
```

---

## 完整调用链路对比

### HTTP同步调用链路

```
Client
  ↓ HTTP Request
┌─────────────────────────────────┐
│ Interface Layer - HTTP Handler  │
│ product_handler.go              │
└──────────────┬──────────────────┘
               ↓ DTO
┌─────────────────────────────────┐
│ Application Layer               │
│ product_service.go              │
└──────────────┬──────────────────┘
               ↓ Domain Model
┌─────────────────────────────────┐
│ Domain Layer                    │
│ product.go                      │
└──────────────┬──────────────────┘
               ↓ Repository
┌─────────────────────────────────┐
│ Infrastructure Layer            │
│ product_repository.go           │
└─────────────────────────────────┘
```

### Kafka异步调用链路

```
Kafka Topic
  ↓ 异步消息
┌─────────────────────────────────┐
│ Infrastructure Layer            │
│ kafka_consumer.go               │  ← 技术实现（订阅、接收）
└──────────────┬──────────────────┘
               ↓ 消息路由
┌─────────────────────────────────┐
│ Interface Layer - Event Handler │  ← 协议适配（Kafka → DTO）
│ product_event_handler.go        │
└──────────────┬──────────────────┘
               ↓ DTO
┌─────────────────────────────────┐
│ Application Layer               │  ← 复用业务逻辑
│ product_service.go              │
└──────────────┬──────────────────┘
               ↓ Domain Model
┌─────────────────────────────────┐
│ Domain Layer                    │
│ product.go                      │
└──────────────┬──────────────────┘
               ↓ Repository
┌─────────────────────────────────┐
│ Infrastructure Layer            │
│ product_repository.go           │
└─────────────────────────────────┘
```

**关键差异**：
- HTTP: `Interface → Application`
- Event: `Infrastructure → Interface → Application`

**为什么多一层？**
- Infrastructure负责技术细节（Kafka连接、反序列化）
- Interface负责业务适配（事件 → DTO转换）

---

## 方案A的完整代码示例

### 目录结构

```
internal/
├── interfaces/                      # 接口层（统一入口）
│   ├── http/
│   │   └── product_handler.go       # HTTP接口
│   ├── grpc/
│   │   └── product_handler.go       # gRPC接口
│   └── event/                        # ⭐️ 事件接口
│       ├── product_event_handler.go  # 事件Handler（协议适配）
│       └── supplier_event_handler.go # 供应商事件Handler
│
├── application/
│   └── service/
│       └── product_service.go        # 应用服务（被所有接口复用）
│
├── infrastructure/
│   ├── messaging/                    # ⭐️ 消息中间件
│   │   ├── kafka_consumer.go        # Kafka消费者（技术实现）
│   │   └── kafka_producer.go        # Kafka生产者（技术实现）
│   ├── persistence/
│   └── cache/
│
└── domain/
```

### Interface Layer - Event Handler

```go
// internal/interfaces/event/product_event_handler.go

package event

import (
	"context"
	"encoding/json"
	"fmt"

	"product-service/internal/application/dto"
	"product-service/internal/application/service"
)

// ProductEventHandler 商品事件处理器（接口层）
// 职责：Kafka消息 → DTO → 调用Application Service
type ProductEventHandler struct {
	productService *service.ProductService
}

func NewProductEventHandler(productService *service.ProductService) *ProductEventHandler {
	return &ProductEventHandler{
		productService: productService,
	}
}

// HandleMessage 统一的消息处理入口
func (h *ProductEventHandler) HandleMessage(ctx context.Context, messageType string, data []byte) error {
	switch messageType {
	case "supplier.product.created":
		return h.handleSupplierProductCreated(ctx, data)
		
	case "pricing.price_changed":
		return h.handlePriceChanged(ctx, data)
		
	default:
		return fmt.Errorf("unknown message type: %s", messageType)
	}
}

// handleSupplierProductCreated 处理供应商商品创建事件
func (h *ProductEventHandler) handleSupplierProductCreated(ctx context.Context, data []byte) error {
	fmt.Printf("\n🔔 [Interface Layer - Event] supplier.product.created\n")

	// 1. 反序列化Kafka消息
	var kafkaEvent struct {
		SupplierID   int64  `json:"supplier_id"`
		SupplierSKU  string `json:"supplier_sku"`
		Title        string `json:"title"`
		BasePrice    int64  `json:"base_price"`
		CategoryID   int64  `json:"category_id"`
	}
	if err := json.Unmarshal(data, &kafkaEvent); err != nil {
		return fmt.Errorf("反序列化失败: %w", err)
	}

	// 2. Kafka消息 → DTO（协议适配）
	req := &dto.CreateProductRequest{
		SupplierID:  kafkaEvent.SupplierID,
		SupplierSKU: kafkaEvent.SupplierSKU,
		Title:       kafkaEvent.Title,
		BasePrice:   kafkaEvent.BasePrice,
		CategoryID:  kafkaEvent.CategoryID,
	}

	// 3. 调用应用服务（与HTTP调用同一个方法）
	resp, err := h.productService.CreateProduct(ctx, req)
	if err != nil {
		return fmt.Errorf("创建商品失败: %w", err)
	}

	fmt.Printf("✅ [Interface Layer - Event] Product created, SKUID=%d\n", resp.SKUID)
	return nil
}

// handlePriceChanged 处理价格变更事件（来自定价服务）
func (h *ProductEventHandler) handlePriceChanged(ctx context.Context, data []byte) error {
	fmt.Printf("\n🔔 [Interface Layer - Event] pricing.price_changed\n")

	var kafkaEvent struct {
		SKUID    int64 `json:"sku_id"`
		NewPrice int64 `json:"new_price"`
	}
	if err := json.Unmarshal(data, &kafkaEvent); err != nil {
		return err
	}

	// Kafka消息 → DTO
	req := &dto.UpdatePriceRequest{
		SKUID:    kafkaEvent.SKUID,
		NewPrice: kafkaEvent.NewPrice,
	}

	// 调用应用服务
	_, err := h.productService.UpdateBasePrice(ctx, req)
	return err
}
```

### Infrastructure Layer - Kafka Consumer

```go
// internal/infrastructure/messaging/kafka_consumer.go

package messaging

import (
	"context"
	"fmt"

	eventHandler "product-service/internal/interfaces/event"
	// "github.com/confluentinc/confluent-kafka-go/kafka"
)

// KafkaConsumer Kafka消费者（技术实现）
// 职责：管理Kafka连接、订阅Topic、接收消息
type KafkaConsumer struct {
	// consumer *kafka.Consumer  // 实际Kafka客户端
	eventHandler *eventHandler.ProductEventHandler
	topics       []string
}

func NewKafkaConsumer(eventHandler *eventHandler.ProductEventHandler) *KafkaConsumer {
	return &KafkaConsumer{
		eventHandler: eventHandler,
		topics: []string{
			"supplier-product-events",  // 供应商事件
			"pricing-events",           // 定价事件
		},
	}
}

// Start 启动消费者（阻塞）
func (c *KafkaConsumer) Start(ctx context.Context) error {
	fmt.Printf("📡 [Infrastructure - Kafka Consumer] Starting...\n")
	fmt.Printf("📡 [Infrastructure - Kafka Consumer] Topics: %v\n", c.topics)

	// 实际项目：
	// c.consumer, _ = kafka.NewConsumer(&kafka.ConfigMap{
	//     "bootstrap.servers": "localhost:9092",
	//     "group.id":          "product-service-group",
	//     "auto.offset.reset": "earliest",
	// })
	//
	// c.consumer.SubscribeTopics(c.topics, nil)
	//
	// for {
	//     select {
	//     case <-ctx.Done():
	//         return ctx.Err()
	//     default:
	//         msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
	//         if err != nil {
	//             continue
	//         }
	//         
	//         // ⭐️ 调用Interface Layer的Handler
	//         messageType := string(msg.Key)
	//         if err := c.eventHandler.HandleMessage(ctx, messageType, msg.Value); err != nil {
	//             fmt.Printf("❌ Handle error: %v\n", err)
	//         }
	//     }
	// }

	return nil
}

// Stop 停止消费者
func (c *KafkaConsumer) Stop() error {
	// c.consumer.Close()
	return nil
}
```

---

## 方案A vs 方案B 深度对比

### 依赖关系

**方案A（推荐）**：
```
Infrastructure (KafkaConsumer)
    ↓ 调用
Interface (ProductEventHandler)
    ↓ 调用
Application (ProductService)
    ↓ 调用
Domain (Product)
```

**方案B（不推荐）**：
```
Infrastructure (KafkaConsumer)
    ↓ 直接调用（违反依赖倒置）
Application (ProductService)
    ↓ 调用
Domain (Product)
```

### 代码复用性

| 场景 | 方案A | 方案B |
|-----|-------|-------|
| HTTP调用 | `httpHandler → productService.OnShelf()` | 同左 |
| gRPC调用 | `grpcHandler → productService.OnShelf()` | 同左 |
| Kafka调用 | `kafkaConsumer → eventHandler → productService.OnShelf()` | `kafkaConsumer → productService.OnShelf()` |
| **复用程度** | ✅ Application Service被所有接口复用 | ⚠️ 部分复用 |

### 职责分离

| 层级 | 方案A职责 | 方案B职责 |
|-----|----------|----------|
| Interface Layer | 协议适配（HTTP/gRPC/Event → DTO） | 协议适配（HTTP/gRPC → DTO） |
| Infrastructure Layer | 技术实现（Kafka/MySQL/Redis） | 技术实现 + 部分协议适配（混乱） |

---

## 实际项目中的完整示例

### 场景：供应商服务发布"新商品创建"事件

#### Step 1: 供应商服务发布事件

```go
// supplier-service/internal/application/service/supplier_service.go

func (s *SupplierService) CreateProduct(...) error {
	// 创建商品
	product := ...
	s.repo.Save(ctx, product)
	
	// ⭐️ 发布领域事件
	s.eventPublisher.Publish(ctx, SupplierProductCreatedEvent{
		SupplierID:  product.SupplierID(),
		SupplierSKU: product.SKU(),
		Title:       product.Title(),
		BasePrice:   product.BasePrice(),
	})
	
	return nil
}
```

#### Step 2: Kafka中转

```
Topic: supplier-product-events
Partition: 0
Message: {
  "supplier_id": 2001,
  "supplier_sku": "SUP-SKU-001",
  "title": "iPhone 17 Pro Max",
  "base_price": 999900
}
```

#### Step 3: 商品服务消费事件

```go
// product-service/internal/infrastructure/messaging/kafka_consumer.go

func (c *KafkaConsumer) Start(ctx) {
	for {
		msg := c.consumer.ReadMessage(-1)
		
		// ⭐️ 调用Interface Layer处理
		c.eventHandler.HandleMessage(ctx, string(msg.Key), msg.Value)
	}
}
```

#### Step 4: Interface Layer适配

```go
// product-service/internal/interfaces/event/product_event_handler.go

func (h *ProductEventHandler) handleSupplierProductCreated(ctx, data) error {
	// 1. 反序列化Kafka消息
	var kafkaEvent SupplierProductCreatedEvent
	json.Unmarshal(data, &kafkaEvent)
	
	// 2. Kafka消息 → DTO
	req := &dto.CreateProductRequest{
		SupplierID:  kafkaEvent.SupplierID,
		SupplierSKU: kafkaEvent.SupplierSKU,
		Title:       kafkaEvent.Title,
		BasePrice:   kafkaEvent.BasePrice,
	}
	
	// 3. 调用Application Service
	h.productService.CreateProduct(ctx, req)
	
	return nil
}
```

#### Step 5: Application Service执行业务逻辑

```go
// product-service/internal/application/service/product_service.go

func (s *ProductService) CreateProduct(ctx, req) (*dto.CreateProductResponse, error) {
	// 创建领域对象
	product := domain.NewProduct(...)
	
	// 保存
	s.repo.Save(ctx, product)
	
	// 发布自己的领域事件
	s.publishDomainEvents(ctx, product)
	
	return &dto.CreateProductResponse{SKUID: product.SKUID()}, nil
}
```

---

## 为什么选择方案A？

### 1. 符合依赖倒置原则

```
外层（Infrastructure）依赖内层（Interface）
内层（Interface）依赖更内层（Application）
✅ 依赖方向：外 → 内
```

### 2. 对称性（美学）

```
同步接口：Client → Interface → Application
异步接口：Kafka → Infrastructure → Interface → Application
          (多一层技术实现)
```

### 3. 易于测试

```go
// 测试Interface Layer的EventHandler
func TestProductEventHandler(t *testing.T) {
	mockService := &MockProductService{}
	handler := event.NewProductEventHandler(mockService)
	
	data := []byte(`{"sku_id": 10001, "new_price": 999900}`)
	err := handler.HandleMessage(ctx, "pricing.price_changed", data)
	
	assert.NoError(t, err)
	assert.True(t, mockService.UpdateBasePriceCalled)
}
```

### 4. 易于替换消息队列

```go
// 从Kafka切换到RabbitMQ，只需修改Infrastructure层
// Interface Layer完全不变

// kafka_consumer.go → rabbitmq_consumer.go
type RabbitMQConsumer struct {
	eventHandler *eventHandler.ProductEventHandler  // 复用同一个Handler
}
```

---

## 实际项目部署架构

### 单体架构（Demo）

```
product-service (单个进程)
├── HTTP Server (Port 8080)
├── gRPC Server (Port 9090)
└── Kafka Consumer (异步线程)
```

### 微服务架构（实际项目）

```
product-service-api (处理HTTP/gRPC)
├── Deployment: 10副本
└── 职责：对外接口

product-service-consumer (处理Kafka事件)
├── Deployment: 3副本
├── Consumer Group: product-service-group
└── 职责：消费事件

共享：
├── Application Service
├── Domain Model
└── Repository
```

**分离的好处**：
- ✅ API服务可以独立扩容（根据QPS）
- ✅ Consumer服务可以独立扩容（根据消息堆积）
- ✅ Consumer故障不影响API可用性

---

## 总结

### 推荐方案：方案A

**事件订阅者的分层**：
1. **Interface Layer**: `interfaces/event/` - 协议适配（Kafka → DTO）
2. **Infrastructure Layer**: `infrastructure/messaging/` - 技术实现（Kafka客户端）

**调用链路**：
```
Kafka Topic 
  → Infrastructure (接收消息)
  → Interface (协议适配) 
  → Application (业务编排)
  → Domain (业务规则)
  → Infrastructure (持久化)
```

**核心原则**：
- ✅ 事件订阅是"异步的接口"，放在Interface Layer
- ✅ Kafka/RabbitMQ是"技术实现"，放在Infrastructure Layer
- ✅ 复用Application Service，业务逻辑只写一次
- ✅ 保持依赖方向：外层依赖内层

对应章节：**16.6.1.5.3 领域事件**
