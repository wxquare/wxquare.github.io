package messaging

import (
	"context"
	"fmt"

	eventHandler "product-service/internal/interfaces/event"
	// "github.com/confluentinc/confluent-kafka-go/kafka"
)

// KafkaConsumer Kafka消费者（基础设施层）
// 职责：管理Kafka连接、订阅Topic、接收消息、路由到Interface Layer
type KafkaConsumer struct {
	// 实际项目：
	// consumer *kafka.Consumer
	eventHandler *eventHandler.ProductEventHandler
	topics       []string
}

func NewKafkaConsumer(eventHandler *eventHandler.ProductEventHandler) *KafkaConsumer {
	return &KafkaConsumer{
		eventHandler: eventHandler,
		topics: []string{
			"supplier-product-events", // 供应商事件
			"pricing-events",          // 定价事件
		},
	}
}

// Start 启动消费者（阻塞）
func (c *KafkaConsumer) Start(ctx context.Context) error {
	fmt.Printf("\n📡 [Infrastructure - Kafka Consumer] Starting...\n")
	fmt.Printf("📡 [Infrastructure - Kafka Consumer] Subscribed topics: %v\n", c.topics)
	fmt.Printf("📡 [Infrastructure - Kafka Consumer] Consumer Group: product-service-group\n")

	// 实际项目：
	// c.consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
	//     "bootstrap.servers": "localhost:9092",
	//     "group.id":          "product-service-group",
	//     "auto.offset.reset": "earliest",
	//     "enable.auto.commit": false,  // 手动提交offset
	// })
	// if err != nil {
	//     return fmt.Errorf("创建Kafka Consumer失败: %w", err)
	// }
	//
	// c.consumer.SubscribeTopics(c.topics, nil)
	//
	// // 消费循环
	// for {
	//     select {
	//     case <-ctx.Done():
	//         fmt.Println("📡 [Infrastructure - Kafka Consumer] Stopping...")
	//         return ctx.Err()
	//     default:
	//         msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
	//         if err != nil {
	//             continue
	//         }
	//
	//         // ⭐️ 路由消息到Interface Layer的Handler
	//         messageType := string(msg.Key)
	//         if err := c.routeMessage(ctx, messageType, msg.Value); err != nil {
	//             fmt.Printf("❌ [Infrastructure - Kafka Consumer] Handle error: %v\n", err)
	//             // 错误处理：重试或发送到DLQ (Dead Letter Queue)
	//         } else {
	//             // ⭐️ 手动提交offset（保证at-least-once）
	//             c.consumer.CommitMessage(msg)
	//         }
	//     }
	// }

	fmt.Printf("✅ [Infrastructure - Kafka Consumer] Ready to consume\n")
	return nil
}

// routeMessage 路由消息到对应的Handler（Interface Layer）
func (c *KafkaConsumer) routeMessage(ctx context.Context, messageType string, data []byte) error {
	fmt.Printf("📬 [Infrastructure - Kafka Consumer] Routing message: %s\n", messageType)

	// ⭐️ 调用Interface Layer的Handler
	if err := c.eventHandler.HandleMessage(ctx, messageType, data); err != nil {
		return fmt.Errorf("处理消息失败: %w", err)
	}

	return nil
}

// Stop 停止消费者
func (c *KafkaConsumer) Stop() error {
	fmt.Println("📡 [Infrastructure - Kafka Consumer] Stopping...")
	// c.consumer.Close()
	return nil
}

// SimulateReceiveMessage 模拟接收消息（用于Demo）
func (c *KafkaConsumer) SimulateReceiveMessage(ctx context.Context, messageType string, data []byte) error {
	fmt.Printf("\n📬 [Infrastructure - Kafka Consumer] Simulating message: %s\n", messageType)
	return c.routeMessage(ctx, messageType, data)
}
