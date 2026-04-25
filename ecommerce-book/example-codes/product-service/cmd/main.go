package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"product-service/internal/application/dto"
	"product-service/internal/application/service"
	"product-service/internal/domain"
	"product-service/internal/domain/strategy"
	"product-service/internal/infrastructure/cache"
	"product-service/internal/infrastructure/messaging"
	"product-service/internal/infrastructure/persistence"
	eventHandler "product-service/internal/interfaces/event"
	httpHandler "product-service/internal/interfaces/http"
	// grpcHandler "product-service/internal/interfaces/grpc"
	// "google.golang.org/grpc"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("🚀 Product Service - DDD 四层架构 Demo")
	fmt.Println("   事件订阅者分层设计：Interface Layer")
	fmt.Println("===========================================")
	fmt.Println("")

	// 初始化依赖（依赖注入）
	dependencies := initDependencies()

	// 准备测试数据
	dependencies.repo.InitTestData()
	fmt.Println("✅ Test data initialized")
	fmt.Println("")

	// 启动HTTP服务器（非阻塞）
	go startHTTPServer(dependencies.httpHandler)

	// ⭐️ 启动Kafka Consumer（非阻塞）
	go startKafkaConsumer(dependencies.kafkaConsumer)

	// 启动gRPC服务器（示例，实际需要protoc生成代码）
	// go startGRPCServer(dependencies.grpcHandler)

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 运行完整Demo
	runCompleteDemo()

	fmt.Println("\n===========================================")
	fmt.Println("✅ Demo Completed!")
	fmt.Println("===========================================")
}

// Dependencies 依赖容器
type Dependencies struct {
	localCache     *cache.LocalCache
	redisCache     *cache.RedisCache
	eventPublisher messaging.EventPublisher
	repo           *persistence.ProductRepositoryImpl
	productService *service.ProductService
	runtimeRepo    *persistence.RuntimeContextRepository
	runtimeService *service.RuntimeContextService
	actionService  *service.CategoryActionService
	httpHandler    *httpHandler.ProductHandler
	eventHandler   *eventHandler.ProductEventHandler // ⭐️ Interface Layer
	kafkaConsumer  *messaging.KafkaConsumer          // ⭐️ Infrastructure Layer
	// grpcHandler    *grpcHandler.ProductServiceServer
}

// initDependencies 初始化依赖（依赖注入容器）
func initDependencies() *Dependencies {
	fmt.Println("📦 Initializing dependencies...")

	// Infrastructure Layer - Cache
	localCache := cache.NewLocalCache()
	redisCache := cache.NewRedisCache()
	repo := persistence.NewProductRepository(localCache, redisCache)

	// Infrastructure Layer - Messaging
	eventPublisher := messaging.NewKafkaProducer()

	// Application Layer
	productService := service.NewProductService(repo, eventPublisher)
	runtimeRepo := persistence.NewRuntimeContextRepository()
	runtimeService := service.NewRuntimeContextService(runtimeRepo, []domain.CategoryStrategy{
		strategy.NewTopupStrategy(),
		strategy.NewGiftCardStrategy(),
		strategy.NewFlightStrategy(),
		strategy.NewHotelStrategy(),
	})
	actionService := service.NewCategoryActionService(runtimeRepo)

	// Interface Layer - HTTP
	handler := httpHandler.NewProductHandler(productService, runtimeService, actionService)

	// ⭐️ Interface Layer - Event (异步接口)
	evtHandler := eventHandler.NewProductEventHandler(productService)

	// ⭐️ Infrastructure Layer - Kafka Consumer (技术实现)
	kafkaConsumer := messaging.NewKafkaConsumer(evtHandler)

	// Interface Layer - gRPC (demo only)
	// grpcHandler := grpcHandler.NewProductServiceServer(productService)

	fmt.Println("✅ Dependencies initialized")
	fmt.Println("   📍 Event Handler: interfaces/event/ (协议适配)")
	fmt.Println("   📍 Kafka Consumer: infrastructure/messaging/ (技术实现)")
	fmt.Println("")

	return &Dependencies{
		localCache:     localCache,
		redisCache:     redisCache,
		eventPublisher: eventPublisher,
		repo:           repo,
		productService: productService,
		runtimeRepo:    runtimeRepo,
		runtimeService: runtimeService,
		actionService:  actionService,
		httpHandler:    handler,
		eventHandler:   evtHandler,
		kafkaConsumer:  kafkaConsumer,
		// grpcHandler:    grpcHandler,
	}
}

// startHTTPServer 启动HTTP服务器
func startHTTPServer(handler *httpHandler.ProductHandler) {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	fmt.Println("🌐 HTTP Server starting on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("❌ HTTP Server error: %v\n", err)
	}
}

// ⭐️ startKafkaConsumer 启动Kafka消费者
func startKafkaConsumer(consumer *messaging.KafkaConsumer) {
	ctx := context.Background()
	if err := consumer.Start(ctx); err != nil {
		fmt.Printf("❌ Kafka Consumer error: %v\n", err)
	}
}

// startGRPCServer 启动gRPC服务器（示例）
// func startGRPCServer(handler *grpcHandler.ProductServiceServer) {
// 	lis, err := net.Listen("tcp", ":9090")
// 	if err != nil {
// 		fmt.Printf("❌ gRPC Server error: %v\n", err)
// 		return
// 	}
//
// 	s := grpc.NewServer()
// 	proto.RegisterProductServiceServer(s, handler)
//
// 	fmt.Println("🌐 gRPC Server starting on :9090...")
// 	if err := s.Serve(lis); err != nil {
// 		fmt.Printf("❌ gRPC Server error: %v\n", err)
// 	}
// }

// runCompleteDemo 运行完整示例（展示数据流转）
func runCompleteDemo() {
	baseURL := "http://localhost:8080"

	fmt.Println("\n===========================================")
	fmt.Println("📋 Demo 1: 查询商品（展示完整数据流转）")
	fmt.Println("===========================================")
	fmt.Println("\n【数据流转路径】")
	fmt.Println("HTTP Request → Interface Layer → Application Layer → Domain Layer → Infrastructure Layer")
	fmt.Println("             ↓                  ↓                    ↓               ↓")
	fmt.Println("           解析请求          业务编排             业务规则        三级缓存查询")
	fmt.Println("             ↓                  ↓                    ↓               ↓")
	fmt.Println("           DTO转换          调用Repository      聚合根方法      L1→L2→L3")
	fmt.Println("")

	// 第一次查询：Cache Miss，从数据库加载
	fmt.Println("▶️  第一次查询 SKUID=10001 (预期：L1 Miss → L2 Miss → L3 Hit)")
	getProduct(baseURL, 10001)
	time.Sleep(200 * time.Millisecond)

	// 第二次查询：L1 Hit
	fmt.Println("\n▶️  第二次查询 SKUID=10001 (预期：L1 Hit)")
	getProduct(baseURL, 10001)
	time.Sleep(200 * time.Millisecond)

	// 第三次查询：不同商品
	fmt.Println("\n▶️  第三次查询 SKUID=10002 (预期：L1 Miss → L2 Miss → L3 Hit)")
	getProduct(baseURL, 10002)
	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n===========================================")
	fmt.Println("📋 Demo 2: 事件发布（领域事件）")
	fmt.Println("===========================================")
	fmt.Println("\n【事件流程】")
	fmt.Println("1. 商品上架 → Domain Layer产生事件 → Application Layer发布")
	fmt.Println("2. Event Publisher → Kafka Topic (product-domain-events)")
	fmt.Println("3. Event Subscribers消费事件 → SearchService/InventoryService/NotificationService")
	fmt.Println("")

	// 商品上架（触发事件）
	fmt.Println("▶️  商品上架 SKUID=10001")
	onShelf(baseURL, 10001)
	time.Sleep(500 * time.Millisecond) // 等待事件处理

	// 更新价格（触发事件）
	fmt.Println("\n▶️  更新价格 SKUID=10001, NewPrice=¥7999.00")
	updatePrice(baseURL, 10001, 799900)
	time.Sleep(500 * time.Millisecond) // 等待事件处理

	fmt.Println("\n===========================================")
	fmt.Println("📋 Demo 3: 事件订阅（接收外部服务事件）")
	fmt.Println("===========================================")
	fmt.Println("\n【事件订阅流程】")
	fmt.Println("Kafka Topic (supplier-product-events)")
	fmt.Println("  ↓ 异步消息")
	fmt.Println("Infrastructure Layer (Kafka Consumer) ← 接收消息、技术实现")
	fmt.Println("  ↓ 消息路由")
	fmt.Println("Interface Layer (Event Handler) ← 协议适配（Kafka → DTO）")
	fmt.Println("  ↓ DTO")
	fmt.Println("Application Layer (Product Service) ← 业务编排")
	fmt.Println("  ↓ Domain Model")
	fmt.Println("Domain Layer (Product Aggregate) ← 业务规则")
	fmt.Println("")

	// 模拟接收供应商商品创建事件
	fmt.Println("▶️  模拟接收供应商事件: supplier.product.created")
	simulateSupplierProductCreated()
	time.Sleep(500 * time.Millisecond)

	// 模拟接收定价变更事件
	fmt.Println("\n▶️  模拟接收定价事件: pricing.price_changed")
	simulatePricingPriceChanged()
	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n===========================================")
	fmt.Println("📋 Demo 4: 八层商品交易模型（ProductRuntimeContext）")
	fmt.Println("===========================================")
	fmt.Println("\n【八层模型】Product Definition → Resource → Offer → Availability → Input Schema → Booking → Fulfillment → Refund")
	fmt.Println("")
	getRuntimeContext(baseURL, 10001, 10102, "detail")
	getRuntimeContext(baseURL, 10002, 30105, "detail")
	getRuntimeContext(baseURL, 40001, 40102, "checkout")
	getRuntimeContext(baseURL, 40002, 40104, "checkout")

	fmt.Println("\n===========================================")
	fmt.Println("📋 Demo 5: 品类动作接口与垂直实时供给接口")
	fmt.Println("===========================================")
	validateTopupAccount(baseURL, 10001, "13800138000")
	searchFlights(baseURL, "SHA", "BJS", "2026-05-01", 1)
}

// HTTP客户端辅助函数

func getProduct(baseURL string, skuID int64) {
	url := fmt.Sprintf("%s/api/v1/products/%d", baseURL, skuID)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Request error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result dto.GetProductResponse
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("   Response: SKUID=%d, Name=%s, Price=%s, Status=%s\n",
		result.SKUID, result.SKUName, formatPrice(result.BasePrice), result.Status)
}

func onShelf(baseURL string, skuID int64) {
	url := fmt.Sprintf("%s/api/v1/products/on-shelf", baseURL)
	reqBody := dto.OnShelfRequest{SKUID: skuID}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		fmt.Printf("❌ Request error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result dto.OnShelfResponse
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("   Response: Success=%v, Message=%s\n", result.Success, result.Message)
}

func updatePrice(baseURL string, skuID int64, newPrice int64) {
	url := fmt.Sprintf("%s/api/v1/products/update-price", baseURL)
	reqBody := dto.UpdatePriceRequest{SKUID: skuID, NewPrice: newPrice}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		fmt.Printf("❌ Request error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result dto.UpdatePriceResponse
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("   Response: Success=%v, Message=%s\n", result.Success, result.Message)
}

func getRuntimeContext(baseURL string, skuID int64, categoryID int64, scene string) {
	url := fmt.Sprintf("%s/api/v1/products/runtime-context?sku_id=%d&category_id=%d&scene=%s", baseURL, skuID, categoryID, scene)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Request error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result dto.RuntimeContextResponse
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("   SKU=%d Category=%s Offer=%s Availability=%s Booking=%s Fulfillment=%s\n",
		result.SKUID,
		result.CategoryName,
		result.Offer.OfferType,
		result.Availability.AvailabilityType,
		result.Booking.Mode,
		result.Fulfillment.Type,
	)
}

func validateTopupAccount(baseURL string, skuID int64, mobileNumber string) {
	url := fmt.Sprintf("%s/api/v1/topup/validate-account", baseURL)
	reqBody := dto.TopupValidateAccountRequest{SKUID: skuID, MobileNumber: mobileNumber}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		fmt.Printf("❌ Request error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result dto.TopupValidateAccountResponse
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("   Topup Validate: SKU=%d Mobile=%s Valid=%v Operator=%s\n",
		result.SKUID, result.MobileNumber, result.Valid, result.Operator)
}

func searchFlights(baseURL string, from string, to string, date string, adult int) {
	url := fmt.Sprintf("%s/api/v1/travel/flights/search?from=%s&to=%s&date=%s&adult=%d", baseURL, from, to, date, adult)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Request error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result dto.FlightSearchResponse
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("   Flight Search: Route=%s Offers=%d\n", result.RouteCode, len(result.Offers))
	for _, offer := range result.Offers {
		fmt.Printf("     Offer=%s Flight=%s Price=%s Booking=%s\n",
			offer.OfferToken, offer.FlightNo, formatPrice(offer.Price.Amount), offer.BookingMode)
	}
}

func formatPrice(amount int64) string {
	return fmt.Sprintf("¥%.2f", float64(amount)/100)
}

// 事件订阅模拟函数

func simulateSupplierProductCreated() {
	kafkaMessage := map[string]interface{}{
		"supplier_id":  2001,
		"supplier_sku": "SUP-SKU-999",
		"title":        "模拟供应商商品 - iPhone 17 Pro Max",
		"base_price":   999900,
		"category_id":  1,
		"color":        "深空灰",
		"size":         "512GB",
	}
	data, _ := json.Marshal(kafkaMessage)

	// 模拟Kafka Consumer接收消息
	// 实际项目中这是在Kafka Consumer的循环中自动接收的
	fmt.Printf("   [Simulating] Kafka message received on topic: supplier-product-events (%d bytes)\n", len(data))

	// 注意：实际项目不需要手动调用，这里为了演示
	// dependencies.kafkaConsumer.SimulateReceiveMessage(context.Background(), "supplier.product.created", data)
	fmt.Println("   [Simulating] Message would be routed to Interface Layer → Application Layer")
}

func simulatePricingPriceChanged() {
	kafkaMessage := map[string]interface{}{
		"sku_id":    10001,
		"new_price": 599900, // ¥5999.00
	}
	data, _ := json.Marshal(kafkaMessage)

	fmt.Printf("   [Simulating] Kafka message received on topic: pricing-events (%d bytes)\n", len(data))
	fmt.Println("   [Simulating] Message would be routed to Interface Layer → Application Layer")
	// dependencies.kafkaConsumer.SimulateReceiveMessage(context.Background(), "pricing.price_changed", data)
}

// gRPC客户端示例（不要求编译通过）
// func getProductGRPC(skuID int64) {
// 	conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
// 	if err != nil {
// 		fmt.Printf("❌ gRPC connection error: %v\n", err)
// 		return
// 	}
// 	defer conn.Close()
//
// 	client := proto.NewProductServiceClient(conn)
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()
//
// 	req := &proto.GetProductRequest{SkuId: skuID}
// 	resp, err := client.GetProduct(ctx, req)
// 	if err != nil {
// 		fmt.Printf("❌ gRPC call error: %v\n", err)
// 		return
// 	}
//
// 	fmt.Printf("   gRPC Response: SKUID=%d, Name=%s\n",
// 		resp.Product.SkuId, resp.Product.SkuName)
// }
