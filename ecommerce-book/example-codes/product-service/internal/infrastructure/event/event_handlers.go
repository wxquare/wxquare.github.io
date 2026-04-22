package event

import (
	"context"
	"encoding/json"
	"fmt"

	"product-service/internal/domain"
)

// SearchServiceEventHandler 搜索服务事件处理器
// 用途：监听商品事件，更新ES索引
type SearchServiceEventHandler struct {
	serviceName string
}

func NewSearchServiceEventHandler() *SearchServiceEventHandler {
	return &SearchServiceEventHandler{
		serviceName: "SearchService",
	}
}

func (h *SearchServiceEventHandler) Handle(ctx context.Context, data []byte) error {
	var event map[string]interface{}
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	fmt.Printf("\n🔔 [%s] Received event\n", h.serviceName)
	
	// 解析事件类型
	eventType, _ := event["EventType"].(string)
	
	switch eventType {
	case "product.created":
		var e domain.ProductCreatedEvent
		json.Unmarshal(data, &e)
		return h.handleProductCreated(ctx, &e)
		
	case "product.on_shelf":
		var e domain.ProductOnShelfEvent
		json.Unmarshal(data, &e)
		return h.handleProductOnShelf(ctx, &e)
		
	case "product.off_shelf":
		var e domain.ProductOffShelfEvent
		json.Unmarshal(data, &e)
		return h.handleProductOffShelf(ctx, &e)
		
	case "product.price_changed":
		var e domain.PriceChangedEvent
		json.Unmarshal(data, &e)
		return h.handlePriceChanged(ctx, &e)
	}
	
	return nil
}

func (h *SearchServiceEventHandler) handleProductCreated(ctx context.Context, event *domain.ProductCreatedEvent) error {
	fmt.Printf("   ✅ [%s] Updating ES index for new product, SKUID=%d\n", h.serviceName, event.SKUID)
	// 实际操作：调用ES API创建文档
	// esClient.Index(ctx, "products", productDoc)
	return nil
}

func (h *SearchServiceEventHandler) handleProductOnShelf(ctx context.Context, event *domain.ProductOnShelfEvent) error {
	fmt.Printf("   ✅ [%s] Marking product as available in ES, SKUID=%d\n", h.serviceName, event.SKUID)
	// 实际操作：更新ES中的status字段
	// esClient.Update(ctx, "products", skuID, map[string]string{"status": "ON_SHELF"})
	return nil
}

func (h *SearchServiceEventHandler) handleProductOffShelf(ctx context.Context, event *domain.ProductOffShelfEvent) error {
	fmt.Printf("   ✅ [%s] Marking product as unavailable in ES, SKUID=%d, Reason=%s\n", 
		h.serviceName, event.SKUID, event.Reason)
	// 实际操作：更新ES中的status字段
	return nil
}

func (h *SearchServiceEventHandler) handlePriceChanged(ctx context.Context, event *domain.PriceChangedEvent) error {
	fmt.Printf("   ✅ [%s] Updating product price in ES, SKUID=%d, NewPrice=¥%.2f\n", 
		h.serviceName, event.SKUID, float64(event.NewPrice)/100)
	// 实际操作：更新ES中的价格字段
	return nil
}

// InventoryServiceEventHandler 库存服务事件处理器
// 用途：监听商品上架事件，检查库存状态
type InventoryServiceEventHandler struct {
	serviceName string
}

func NewInventoryServiceEventHandler() *InventoryServiceEventHandler {
	return &InventoryServiceEventHandler{
		serviceName: "InventoryService",
	}
}

func (h *InventoryServiceEventHandler) Handle(ctx context.Context, data []byte) error {
	var event map[string]interface{}
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	fmt.Printf("\n🔔 [%s] Received event\n", h.serviceName)
	
	// 只处理上架事件
	eventType, _ := event["EventType"].(string)
	if eventType == "product.on_shelf" {
		var e domain.ProductOnShelfEvent
		json.Unmarshal(data, &e)
		return h.handleProductOnShelf(ctx, &e)
	}
	
	return nil
}

func (h *InventoryServiceEventHandler) handleProductOnShelf(ctx context.Context, event *domain.ProductOnShelfEvent) error {
	fmt.Printf("   ✅ [%s] Checking inventory for SKUID=%d\n", h.serviceName, event.SKUID)
	// 实际操作：检查库存是否充足
	// stock := inventoryRepo.GetStock(ctx, skuID)
	// if stock < threshold {
	//     alert.Send("库存不足预警")
	// }
	return nil
}

// NotificationServiceEventHandler 通知服务事件处理器
// 用途：监听商品事件，发送通知给运营人员
type NotificationServiceEventHandler struct {
	serviceName string
}

func NewNotificationServiceEventHandler() *NotificationServiceEventHandler {
	return &NotificationServiceEventHandler{
		serviceName: "NotificationService",
	}
}

func (h *NotificationServiceEventHandler) Handle(ctx context.Context, data []byte) error {
	var event map[string]interface{}
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	fmt.Printf("\n🔔 [%s] Received event\n", h.serviceName)
	
	eventType, _ := event["EventType"].(string)
	
	switch eventType {
	case "product.on_shelf":
		var e domain.ProductOnShelfEvent
		json.Unmarshal(data, &e)
		return h.sendOnShelfNotification(ctx, &e)
		
	case "product.price_changed":
		var e domain.PriceChangedEvent
		json.Unmarshal(data, &e)
		return h.sendPriceChangeNotification(ctx, &e)
	}
	
	return nil
}

func (h *NotificationServiceEventHandler) sendOnShelfNotification(ctx context.Context, event *domain.ProductOnShelfEvent) error {
	fmt.Printf("   ✅ [%s] Sending notification: Product SKUID=%d is now on shelf\n", 
		h.serviceName, event.SKUID)
	// 实际操作：发送邮件/短信/推送通知
	// notificationClient.Send(ctx, Notification{
	//     Type: "email",
	//     To: "operations@example.com",
	//     Subject: "商品已上架",
	//     Body: fmt.Sprintf("商品 %d 已成功上架", event.SKUID),
	// })
	return nil
}

func (h *NotificationServiceEventHandler) sendPriceChangeNotification(ctx context.Context, event *domain.PriceChangedEvent) error {
	fmt.Printf("   ✅ [%s] Sending notification: Price changed for SKUID=%d (¥%.2f → ¥%.2f)\n", 
		h.serviceName, event.SKUID, 
		float64(event.OldPrice)/100, float64(event.NewPrice)/100)
	// 实际操作：发送价格变更通知
	return nil
}
