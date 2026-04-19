# 事件发布和订阅机制详解 📨

## 完整流程图

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. 用户操作 (HTTP/gRPC)                                          │
│    POST /api/v1/products/on-shelf {sku_id: 10001}              │
└────────────────────────────┬────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│ 2. Interface Layer (接口层)                                      │
│    File: internal/interfaces/http/product_handler.go           │
│    - 解析HTTP请求                                                │
│    - 调用应用服务                                                │
└────────────────────────────┬────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│ 3. Application Layer (应用层)                                    │
│    File: internal/application/service/product_service.go       │
│                                                                  │
│    Step 1: 查询聚合根                                            │
│      product = repo.FindBySKUID(ctx, skuID)                    │
│                                                                  │
│    Step 2: 调用领域方法（产生领域事件）                          │
│      product.OnShelf() → 产生 ProductOnShelfEvent              │
│                                                                  │
│    Step 3: 保存聚合根                                            │
│      repo.Save(ctx, product)                                   │
│                                                                  │
│    Step 4: 发布领域事件 ⭐️                                       │
│      publishDomainEvents(ctx, product)                         │
└────────────────────────────┬────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│ 4. Domain Layer (领域层)                                         │
│    File: internal/domain/product.go                            │
│                                                                  │
│    聚合根内部记录事件：                                          │
│    func (p *Product) OnShelf() error {                         │
│        // 业务规则校验                                           │
│        if p.basePrice.Amount() <= 0 {                          │
│            return errors.New("价格必须大于0")                   │
│        }                                                        │
│                                                                  │
│        // 状态转换                                              │
│        p.status = ProductStatusOnShelf                         │
│                                                                  │
│        // ⭐️ 记录领域事件                                        │
│        p.addDomainEvent(ProductOnShelfEvent{                   │
│            SKUID: p.skuID.Value(),                             │
│            OnShelfAt: time.Now(),                              │
│        })                                                       │
│                                                                  │
│        return nil                                              │
│    }                                                            │
└────────────────────────────┬────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│ 5. Infrastructure Layer (基础设施层 - 事件发布)                  │
│    File: internal/infrastructure/event/event_publisher.go      │
│                                                                  │
│    KafkaEventPublisher.Publish():                              │
│      ├─ 序列化事件为JSON                                         │
│      ├─ 确定Kafka Topic ("product-events")                     │
│      ├─ 发送到Kafka (实际项目)                                  │
│      │  producer.Send(ctx, &kafka.Message{                     │
│      │      Topic: "product-events",                           │
│      │      Key: "product.on_shelf",                           │
│      │      Value: json.Marshal(event),                        │
│      │  })                                                      │
│      └─ 通知本地订阅者 (Demo简化实现)                            │
└────────────────────────────┬────────────────────────────────────┘
                             ↓
                    Kafka Topic: product-events
                             │
                ┌────────────┼────────────┐
                ↓            ↓            ↓
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│ 6. Event Subscribers (事件订阅者 - 其他微服务)                  │
│                                                                │
│ SearchService    │ │ InventoryService │ │ NotificationService│
│ ─────────────    │ │ ────────────────│ │ ──────────────────│
│ 更新ES索引       │ │ 检查库存状态     │ │ 发送通知          │
│                  │ │                  │ │                   │
│ ES.Update({      │ │ stock = get()   │ │ Email.Send({      │
│   status: "ON",  │ │ if stock < 10 { │ │   to: "ops@..",  │
│   ...            │ │   alert()       │ │   subject: "...", │
│ })               │ │ }               │ │ })                │
└──────────────────┘ └──────────────────┘ └──────────────────┘
```

---

## 关键设计要点

### 1. 聚合根内部记录事件

```go
// internal/domain/product.go

type Product struct {
    // ... 其他字段
    domainEvents []DomainEvent  // ⭐️ 事件列表
}

func (p *Product) OnShelf() error {
    // 业务规则校验 + 状态转换
    p.status = ProductStatusOnShelf
    
    // ⭐️ 记录事件（不是立即发布）
    p.addDomainEvent(ProductOnShelfEvent{
        SKUID:     p.skuID.Value(),
        OnShelfAt: time.Now(),
    })
    
    return nil
}
```

**为什么在聚合根内部记录？**
- ✅ 事件是业务事实的一部分，属于领域层
- ✅ 保证事件和状态变更的原子性
- ✅ 领域层不依赖任何技术实现（如Kafka）

---

### 2. 应用层统一发布事件

```go
// internal/application/service/product_service.go

func (s *ProductService) OnShelf(ctx context.Context, req *dto.OnShelfRequest) (*dto.OnShelfResponse, error) {
    // Step 1: 查询聚合根
    product, _ := s.repo.FindBySKUID(ctx, skuID)
    
    // Step 2: 调用领域方法（产生事件）
    product.OnShelf()
    
    // Step 3: 保存聚合根
    s.repo.Save(ctx, product)
    
    // Step 4: ⭐️ 统一发布事件
    s.publishDomainEvents(ctx, product)
    
    return &dto.OnShelfResponse{Success: true}, nil
}

func (s *ProductService) publishDomainEvents(ctx context.Context, product *domain.Product) error {
    events := product.DomainEvents()  // 获取事件列表
    
    // 批量发布
    s.eventPublisher.PublishBatch(ctx, events)
    
    // ⭐️ 清除已发布的事件
    product.ClearDomainEvents()
    
    return nil
}
```

**为什么在应用层发布？**
- ✅ 应用层负责事务管理和编排
- ✅ 可以在事务提交后再发布事件
- ✅ 统一的事件发布逻辑，便于监控和重试

---

### 3. 基础设施层实现发布器

```go
// internal/infrastructure/event/event_publisher.go

type EventPublisher interface {
    Publish(ctx context.Context, event domain.DomainEvent) error
    PublishBatch(ctx context.Context, events []domain.DomainEvent) error
}

type KafkaEventPublisher struct {
    // 实际项目：Kafka Producer
    // producer *kafka.Producer
}

func (p *KafkaEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
    // 1. 序列化事件
    data, _ := json.Marshal(event)
    
    // 2. 确定Topic
    topic := p.getTopicByEventType(event.EventType())
    
    // 3. 发送到Kafka
    // p.producer.Send(ctx, &kafka.Message{
    //     Topic: topic,
    //     Key:   []byte(event.EventType()),
    //     Value: data,
    // })
    
    return nil
}
```

**为什么在基础设施层？**
- ✅ Kafka是技术实现细节，不属于领域层
- ✅ 可以轻松替换为其他消息队列（RabbitMQ, Redis Stream等）
- ✅ 测试时可以使用内存实现

---

### 4. 事件订阅者（消费者）

```go
// internal/infrastructure/event/event_handlers.go

type SearchServiceEventHandler struct {
    serviceName string
}

func (h *SearchServiceEventHandler) Handle(ctx context.Context, data []byte) error {
    var event map[string]interface{}
    json.Unmarshal(data, &event)
    
    eventType, _ := event["EventType"].(string)
    
    switch eventType {
    case "product.on_shelf":
        // 更新ES索引
        h.updateESIndex(ctx, event)
    case "product.price_changed":
        // 更新ES中的价格
        h.updatePrice(ctx, event)
    }
    
    return nil
}
```

**实际项目中的部署：**
```
# 每个微服务独立订阅Kafka Topic

ProductService (生产者)
    ↓ 发布事件到 Kafka Topic: product-events
    
SearchService (消费者1)
    ↓ 订阅 product-events
    └─ Consumer Group: search-service-group
    
InventoryService (消费者2)
    ↓ 订阅 product-events
    └─ Consumer Group: inventory-service-group
    
NotificationService (消费者3)
    ↓ 订阅 product-events
    └─ Consumer Group: notification-service-group
```

---

## 事件类型和用途

| 事件名 | 触发时机 | 消费方 | 用途 |
|-------|---------|--------|------|
| `product.created` | 创建新商品 | SearchService | 创建ES索引 |
| `product.on_shelf` | 商品上架 | SearchService<br>InventoryService<br>NotificationService | 更新ES状态<br>检查库存<br>发送通知 |
| `product.off_shelf` | 商品下架 | SearchService | 标记ES不可见 |
| `product.price_changed` | 价格变更 | SearchService<br>NotificationService | 更新ES价格<br>发送价格变更通知 |

---

## Outbox Pattern（可选）

为保证事件**最终一定会发布**，可以使用Outbox Pattern：

```sql
CREATE TABLE event_outbox (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    aggregate_id BIGINT NOT NULL COMMENT '聚合根ID',
    event_type VARCHAR(100) NOT NULL,
    event_data JSON NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',  -- PENDING/SENT/FAILED
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP,
    INDEX idx_status (status, created_at)
);
```

**流程**：
1. 应用层：在**同一个数据库事务**中，保存聚合根 + 插入event_outbox
2. 定时任务：扫描event_outbox中的PENDING事件，发布到Kafka
3. 发布成功后：更新status为SENT

**优点**：
- ✅ 保证事件不丢失（与业务数据在同一事务）
- ✅ 支持重试（定时任务不断扫描）
- ✅ 最终一致性保证

---

## 运行Demo查看事件流程

```bash
cd /Users/wxquare/go/src/github.com/wxquare.github.io/ecommerce-book/examples/product-service
go run cmd/main.go
```

**输出示例**：
```
📋 Demo: 事件发布和订阅
===========================================

▶️  商品上架 SKUID=10001

🚀 [Application Layer] OnShelf called, SKUID=10001
✅ [L1 Hit] SKUID=10001 from Local Cache
💾 [DB Save] SKUID=10001 saved to MySQL
🗑️  [Cache Delete] SKUID=10001 cache invalidated

📨 [Application Layer] Publishing 2 domain events...
📨 [Event Publisher] Publishing event to topic: product-events
   - EventType: product.created
   - OccurredAt: 2026-04-19 12:00:00
✅ [Event Publisher] Event published successfully

📨 [Event Publisher] Publishing event to topic: product-events
   - EventType: product.on_shelf
   - OccurredAt: 2026-04-19 12:00:05
✅ [Event Publisher] Event published successfully

🔔 [SearchService] Received event
   ✅ [SearchService] Marking product as available in ES, SKUID=10001

🔔 [InventoryService] Received event
   ✅ [InventoryService] Checking inventory for SKUID=10001

🔔 [NotificationService] Received event
   ✅ [NotificationService] Sending notification: Product SKUID=10001 is now on shelf
```

---

## 总结

### 事件发布（生产者）

1. **Domain Layer**: 聚合根内部记录事件
2. **Application Layer**: 统一发布事件
3. **Infrastructure Layer**: 具体实现（Kafka/RabbitMQ）

### 事件订阅（消费者）

1. **独立的微服务**: SearchService, InventoryService, NotificationService
2. **各自订阅**: Kafka Consumer Group
3. **异步处理**: 不阻塞主流程

### 优势

✅ **解耦**: 生产者和消费者互不依赖  
✅ **可扩展**: 新增消费者无需修改生产者  
✅ **异步**: 不阻塞主流程，提升性能  
✅ **最终一致性**: 保证数据最终同步  

对应章节：**16.6.1.5.3 领域事件**
