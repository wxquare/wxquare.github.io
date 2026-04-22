package service

import (
	"context"
	"fmt"
	"time"

	"product-service/internal/application/dto"
	"product-service/internal/domain"
	"product-service/internal/infrastructure/messaging"
)

// ProductService 应用服务（业务编排层）
type ProductService struct {
	repo           domain.ProductRepository
	eventPublisher messaging.EventPublisher
}

func NewProductService(repo domain.ProductRepository, eventPublisher messaging.EventPublisher) *ProductService {
	return &ProductService{
		repo:           repo,
		eventPublisher: eventPublisher,
	}
}

// CreateProduct 创建商品（命令方法）
// 场景：接收来自HTTP、gRPC或Kafka Event的创建请求
func (s *ProductService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.CreateProductResponse, error) {
	fmt.Printf("\n🚀 [Application Layer] CreateProduct called, Title=%s\n", req.Title)

	// Step 1: 创建领域对象
	spu := domain.NewSPU(req.CategoryID, req.Title, req.CategoryID, 0)
	price, _ := domain.NewPrice(req.BasePrice, "CNY")
	specs := domain.NewSpecifications(req.Color, req.Size, nil)
	
	// 生成SKU ID（实际项目应该用分布式ID生成器）
	skuID := domain.NewSKU_ID(time.Now().UnixNano() % 100000 + 10000)
	
	product := domain.NewProduct(
		skuID,
		spu,
		req.SupplierSKU,
		price,
		specs,
	)

	// Step 2: 保存聚合根
	if err := s.repo.Save(ctx, product); err != nil {
		return nil, fmt.Errorf("保存商品失败: %w", err)
	}

	// Step 3: 发布领域事件（ProductCreatedEvent）
	if err := s.publishDomainEvents(ctx, product); err != nil {
		fmt.Printf("⚠️  [Application Layer] Failed to publish events: %v\n", err)
	}

	fmt.Printf("✅ [Application Layer] CreateProduct completed, SKUID=%d\n", skuID.Value())
	return &dto.CreateProductResponse{
		SKUID:   skuID.Value(),
		Message: "商品创建成功",
	}, nil
}

// GetProduct 查询单个商品（查询方法）
// 数据流转：DTO → Domain Model → Infrastructure → Domain Model → DTO
func (s *ProductService) GetProduct(ctx context.Context, req *dto.GetProductRequest) (*dto.GetProductResponse, error) {
	fmt.Printf("\n🚀 [Application Layer] GetProduct called, SKUID=%d\n", req.SKUID)

	// Step 1: 调用Repository查询（会触发三级缓存）
	product, err := s.repo.FindBySKUID(ctx, domain.NewSKU_ID(req.SKUID))
	if err != nil {
		return nil, fmt.Errorf("查询商品失败: %w", err)
	}

	// Step 2: Domain Model → DTO
	response := s.toDTO(product)

	fmt.Printf("✅ [Application Layer] GetProduct completed\n")
	return response, nil
}

// OnShelf 商品上架（命令方法 - 展示事件发布）
func (s *ProductService) OnShelf(ctx context.Context, req *dto.OnShelfRequest) (*dto.OnShelfResponse, error) {
	fmt.Printf("\n🚀 [Application Layer] OnShelf called, SKUID=%d\n", req.SKUID)

	// Step 1: 查询商品
	product, err := s.repo.FindBySKUID(ctx, domain.NewSKU_ID(req.SKUID))
	if err != nil {
		return &dto.OnShelfResponse{
			Success: false,
			Message: fmt.Sprintf("商品不存在: %v", err),
		}, nil
	}

	// Step 2: 调用领域对象的业务方法（会产生领域事件）
	if err := product.OnShelf(); err != nil {
		return &dto.OnShelfResponse{
			Success: false,
			Message: fmt.Sprintf("上架失败: %v", err),
		}, nil
	}

	// Step 3: 保存聚合根
	if err := s.repo.Save(ctx, product); err != nil {
		return &dto.OnShelfResponse{
			Success: false,
			Message: fmt.Sprintf("保存失败: %v", err),
		}, nil
	}

	// Step 4: 发布领域事件（关键步骤！）
	if err := s.publishDomainEvents(ctx, product); err != nil {
		fmt.Printf("⚠️  [Application Layer] Failed to publish events: %v\n", err)
		// 注意：事件发布失败不应该回滚事务
		// 可以考虑使用Outbox Pattern保证最终一致性
	}

	fmt.Printf("✅ [Application Layer] OnShelf completed\n")
	return &dto.OnShelfResponse{
		Success: true,
		Message: "上架成功",
	}, nil
}

// UpdateBasePrice 更新基础价格（命令方法 - 展示事件发布）
func (s *ProductService) UpdateBasePrice(ctx context.Context, req *dto.UpdatePriceRequest) (*dto.UpdatePriceResponse, error) {
	fmt.Printf("\n🚀 [Application Layer] UpdateBasePrice called, SKUID=%d, NewPrice=%d\n", 
		req.SKUID, req.NewPrice)

	// Step 1: 查询商品
	product, err := s.repo.FindBySKUID(ctx, domain.NewSKU_ID(req.SKUID))
	if err != nil {
		return &dto.UpdatePriceResponse{
			Success: false,
			Message: fmt.Sprintf("商品不存在: %v", err),
		}, nil
	}

	// Step 2: 调用领域对象的业务方法（会产生领域事件）
	newPrice, _ := domain.NewPrice(req.NewPrice, "CNY")
	if err := product.UpdateBasePrice(newPrice); err != nil {
		return &dto.UpdatePriceResponse{
			Success: false,
			Message: fmt.Sprintf("更新价格失败: %v", err),
		}, nil
	}

	// Step 3: 保存聚合根
	if err := s.repo.Save(ctx, product); err != nil {
		return &dto.UpdatePriceResponse{
			Success: false,
			Message: fmt.Sprintf("保存失败: %v", err),
		}, nil
	}

	// Step 4: 发布领域事件
	if err := s.publishDomainEvents(ctx, product); err != nil {
		fmt.Printf("⚠️  [Application Layer] Failed to publish events: %v\n", err)
	}

	fmt.Printf("✅ [Application Layer] UpdateBasePrice completed\n")
	return &dto.UpdatePriceResponse{
		Success: true,
		Message: "价格更新成功",
	}, nil
}

// 私有方法

// publishDomainEvents 发布领域事件（统一处理）
func (s *ProductService) publishDomainEvents(ctx context.Context, product *domain.Product) error {
	events := product.DomainEvents()
	if len(events) == 0 {
		return nil
	}

	fmt.Printf("\n📨 [Application Layer] Publishing %d domain events...\n", len(events))

	// 批量发布事件
	if err := s.eventPublisher.PublishBatch(ctx, events); err != nil {
		return err
	}

	// 清除已发布的事件
	product.ClearDomainEvents()

	fmt.Printf("✅ [Application Layer] All events published\n")
	return nil
}

func (s *ProductService) toDTO(product *domain.Product) *dto.GetProductResponse {
	specs := map[string]string{
		"color": product.Specs().Color(),
		"size":  product.Specs().Size(),
	}

	return &dto.GetProductResponse{
		SKUID:     product.SKUID().Value(),
		SPUID:     product.SPU().ID(),
		SKUCode:   product.SKUCode(),
		SKUName:   product.SPU().Title(),
		BasePrice: product.BasePrice().Amount(),
		Specs:     specs,
		Status:    product.Status().String(),
		Images:    product.Images(),
		CreatedAt: product.CreatedAt().Unix(),
		UpdatedAt: product.UpdatedAt().Unix(),
	}
}
