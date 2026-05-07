package event

import (
	"context"
	"encoding/json"
	"fmt"

	"product-service/internal/application/dto"
	"product-service/internal/application/service"
	"product-service/internal/domain"
)

// ProductEventHandler 商品事件处理器（接口层）
// 职责：适配外部事件消息 → 调用Application Service
// 与HTTP/gRPC Handler同级，是"异步接口"
type ProductEventHandler struct {
	productService *service.ProductService
}

func NewProductEventHandler(productService *service.ProductService) *ProductEventHandler {
	return &ProductEventHandler{
		productService: productService,
	}
}

// HandleMessage 统一的消息处理入口
// 由Infrastructure Layer的Kafka Consumer调用
func (h *ProductEventHandler) HandleMessage(ctx context.Context, messageType string, data []byte) error {
	fmt.Printf("\n🔔 [Interface Layer - Event] Received message: %s\n", messageType)

	switch messageType {
	case "supplier.product.created":
		return h.handleSupplierProductCreated(ctx, data)

	case "pricing.price_changed":
		return h.handlePriceChanged(ctx, data)

	default:
		fmt.Printf("⚠️  [Interface Layer - Event] Unknown message type: %s\n", messageType)
		return nil
	}
}

// handleSupplierProductCreated 处理供应商商品创建事件
// 场景：供应商服务创建新商品后，通过Kafka通知商品服务同步
func (h *ProductEventHandler) handleSupplierProductCreated(ctx context.Context, data []byte) error {
	// Step 1: 反序列化Kafka消息
	var kafkaEvent struct {
		SupplierID  int64  `json:"supplier_id"`
		SupplierSKU string `json:"supplier_sku"`
		Title       string `json:"title"`
		BasePrice   int64  `json:"base_price"`
		CategoryID  int64  `json:"category_id"`
		Color       string `json:"color"`
		Size        string `json:"size"`
	}
	if err := json.Unmarshal(data, &kafkaEvent); err != nil {
		return fmt.Errorf("反序列化失败: %w", err)
	}

	// Step 2: Kafka消息 → DTO（协议适配）
	req := &dto.CreateProductRequest{
		SupplierID:  kafkaEvent.SupplierID,
		SupplierSKU: kafkaEvent.SupplierSKU,
		Title:       kafkaEvent.Title,
		BasePrice:   kafkaEvent.BasePrice,
		CategoryID:  kafkaEvent.CategoryID,
		Color:       kafkaEvent.Color,
		Size:        kafkaEvent.Size,
	}

	// Step 3: 调用应用服务（与HTTP/gRPC调用同一个方法）
	resp, err := h.productService.CreateProduct(ctx, req)
	if err != nil {
		return fmt.Errorf("创建商品失败: %w", err)
	}

	fmt.Printf("✅ [Interface Layer - Event] Product created from supplier, SKUID=%d\n", resp.SKUID)
	return nil
}

// handlePriceChanged 处理价格变更事件
// 场景：定价服务计算出新价格后，通知商品服务更新基础价格
func (h *ProductEventHandler) handlePriceChanged(ctx context.Context, data []byte) error {
	// Step 1: 反序列化Kafka消息
	var kafkaEvent struct {
		SKUID    int64 `json:"sku_id"`
		NewPrice int64 `json:"new_price"`
	}
	if err := json.Unmarshal(data, &kafkaEvent); err != nil {
		return fmt.Errorf("反序列化失败: %w", err)
	}

	// Step 2: Kafka消息 → DTO
	req := &dto.UpdatePriceRequest{
		SKUID:    kafkaEvent.SKUID,
		NewPrice: kafkaEvent.NewPrice,
	}

	// Step 3: 调用应用服务
	_, err := h.productService.UpdateBasePrice(ctx, req)
	if err != nil {
		return fmt.Errorf("更新价格失败: %w", err)
	}

	fmt.Printf("✅ [Interface Layer - Event] Price updated from pricing service, SKUID=%d\n", 
		kafkaEvent.SKUID)
	return nil
}

// HandleDomainEvent 处理本服务发布的领域事件（供其他服务消费）
// 注意：这个方法是给外部服务使用的示例代码
// 实际运行在SearchService/InventoryService/NotificationService中
func (h *ProductEventHandler) HandleDomainEvent(ctx context.Context, messageType string, data []byte) error {
	fmt.Printf("\n🔔 [External Service] Received domain event: %s\n", messageType)

	switch messageType {
	case "product.on_shelf":
		return h.handleProductOnShelfForExternalService(ctx, data)

	case "product.price_changed":
		return h.handlePriceChangedForExternalService(ctx, data)

	default:
		return nil
	}
}

// handleProductOnShelfForExternalService 外部服务处理商品上架事件
func (h *ProductEventHandler) handleProductOnShelfForExternalService(ctx context.Context, data []byte) error {
	var event domain.ProductOnShelfEvent
	json.Unmarshal(data, &event)

	fmt.Printf("   [SearchService] Updating ES index for SKUID=%d\n", event.SKUID)
	fmt.Printf("   [InventoryService] Checking stock for SKUID=%d\n", event.SKUID)
	fmt.Printf("   [NotificationService] Sending notification for SKUID=%d\n", event.SKUID)

	return nil
}

// handlePriceChangedForExternalService 外部服务处理价格变更事件
func (h *ProductEventHandler) handlePriceChangedForExternalService(ctx context.Context, data []byte) error {
	var event domain.PriceChangedEvent
	json.Unmarshal(data, &event)

	fmt.Printf("   [SearchService] Updating price in ES, SKUID=%d, NewPrice=¥%.2f\n",
		event.SKUID, float64(event.NewPrice)/100)

	return nil
}
