package event

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"product-service/internal/domain"
)

// EventPublisher 事件发布器接口
type EventPublisher interface {
	Publish(ctx context.Context, event domain.DomainEvent) error
	PublishBatch(ctx context.Context, events []domain.DomainEvent) error
}

// KafkaEventPublisher Kafka事件发布器（简化实现）
// 实际项目应该使用真实的Kafka客户端（如 confluent-kafka-go）
type KafkaEventPublisher struct {
	subscribers map[string][]EventHandler // topic -> handlers
	mu          sync.RWMutex
}

func NewKafkaEventPublisher() *KafkaEventPublisher {
	return &KafkaEventPublisher{
		subscribers: make(map[string][]EventHandler),
	}
}

// Publish 发布单个事件
func (p *KafkaEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	topic := p.getTopicByEventType(event.EventType())
	
	fmt.Printf("📨 [Event Publisher] Publishing event to topic: %s\n", topic)
	fmt.Printf("   - EventType: %s\n", event.EventType())
	fmt.Printf("   - OccurredAt: %v\n", event.OccurredAt())
	
	// 序列化事件
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("序列化事件失败: %w", err)
	}
	
	// 实际项目：发送到Kafka
	// producer.Send(ctx, &kafka.Message{
	//     Topic: topic,
	//     Key:   []byte(event.EventType()),
	//     Value: data,
	// })
	
	// Demo：同步调用订阅者（模拟异步消费）
	p.notifySubscribers(topic, data)
	
	fmt.Printf("✅ [Event Publisher] Event published successfully\n")
	return nil
}

// PublishBatch 批量发布事件
func (p *KafkaEventPublisher) PublishBatch(ctx context.Context, events []domain.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}
	
	fmt.Printf("📨 [Event Publisher] Publishing %d events...\n", len(events))
	
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return err
		}
	}
	
	return nil
}

// Subscribe 订阅事件（用于Demo演示）
func (p *KafkaEventPublisher) Subscribe(topic string, handler EventHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.subscribers[topic] = append(p.subscribers[topic], handler)
	fmt.Printf("📝 [Event Publisher] Handler subscribed to topic: %s\n", topic)
}

// 私有方法

func (p *KafkaEventPublisher) getTopicByEventType(eventType string) string {
	// 根据事件类型映射到Kafka Topic
	topicMapping := map[string]string{
		"product.created":       "product-events",
		"product.on_shelf":      "product-events",
		"product.off_shelf":     "product-events",
		"product.price_changed": "product-events",
	}
	
	if topic, ok := topicMapping[eventType]; ok {
		return topic
	}
	return "product-events" // 默认topic
}

func (p *KafkaEventPublisher) notifySubscribers(topic string, data []byte) {
	p.mu.RLock()
	handlers := p.subscribers[topic]
	p.mu.RUnlock()
	
	for _, handler := range handlers {
		// 模拟异步处理（实际应该在goroutine中）
		go func(h EventHandler) {
			if err := h.Handle(context.Background(), data); err != nil {
				fmt.Printf("❌ [Event Handler] Error: %v\n", err)
			}
		}(handler)
	}
}

// EventHandler 事件处理器接口
type EventHandler interface {
	Handle(ctx context.Context, data []byte) error
}
