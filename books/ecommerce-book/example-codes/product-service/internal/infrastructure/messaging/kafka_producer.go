package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"product-service/internal/domain"
)

// EventPublisher 事件发布器接口
type EventPublisher interface {
	Publish(ctx context.Context, event domain.DomainEvent) error
	PublishBatch(ctx context.Context, events []domain.DomainEvent) error
}

// KafkaProducer Kafka生产者（基础设施层）
// 职责：将领域事件发布到Kafka Topic
type KafkaProducer struct {
	// 实际项目：
	// producer *kafka.Producer
}

func NewKafkaProducer() *KafkaProducer {
	return &KafkaProducer{}
}

// Publish 发布单个事件
func (p *KafkaProducer) Publish(ctx context.Context, event domain.DomainEvent) error {
	topic := p.getTopicByEventType(event.EventType())

	fmt.Printf("📨 [Infrastructure - Kafka Producer] Publishing to topic: %s\n", topic)
	fmt.Printf("   - EventType: %s\n", event.EventType())
	fmt.Printf("   - OccurredAt: %v\n", event.OccurredAt())

	// 序列化事件
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("序列化事件失败: %w", err)
	}

	// 实际项目：发送到Kafka
	// p.producer.Send(ctx, &kafka.Message{
	//     Topic: topic,
	//     Key:   []byte(event.EventType()),
	//     Value: data,
	//     Headers: []kafka.Header{
	//         {Key: "event_type", Value: []byte(event.EventType())},
	//         {Key: "timestamp", Value: []byte(fmt.Sprint(event.OccurredAt().Unix()))},
	//     },
	// })

	// Demo：打印消息
	fmt.Printf("✅ [Infrastructure - Kafka Producer] Message sent: %d bytes\n", len(data))

	return nil
}

// PublishBatch 批量发布事件
func (p *KafkaProducer) PublishBatch(ctx context.Context, events []domain.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	fmt.Printf("\n📨 [Infrastructure - Kafka Producer] Publishing %d events...\n", len(events))

	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return err
		}
	}

	fmt.Printf("✅ [Infrastructure - Kafka Producer] All events published\n")
	return nil
}

// 私有方法

func (p *KafkaProducer) getTopicByEventType(eventType string) string {
	// 根据事件类型映射到Kafka Topic
	topicMapping := map[string]string{
		"product.created":       "product-domain-events",
		"product.on_shelf":      "product-domain-events",
		"product.off_shelf":     "product-domain-events",
		"product.price_changed": "product-domain-events",
	}

	if topic, ok := topicMapping[eventType]; ok {
		return topic
	}
	return "product-domain-events" // 默认topic
}
