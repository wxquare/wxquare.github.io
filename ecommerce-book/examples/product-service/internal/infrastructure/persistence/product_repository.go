package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"product-service/internal/domain"
	"product-service/internal/infrastructure/cache"
)

// ProductRepositoryImpl Repository实现（三级缓存）
type ProductRepositoryImpl struct {
	localCache *cache.LocalCache
	redisCache *cache.RedisCache
	mockDB     map[int64]*ProductDO
	mockSPU    map[int64]*SPUDO
	mu         sync.RWMutex
}

func NewProductRepository(localCache *cache.LocalCache, redisCache *cache.RedisCache) *ProductRepositoryImpl {
	return &ProductRepositoryImpl{
		localCache: localCache,
		redisCache: redisCache,
		mockDB:     make(map[int64]*ProductDO),
		mockSPU:    make(map[int64]*SPUDO),
	}
}

// FindBySKUID 查询商品（三级缓存）
// 数据流转：L1本地缓存 → L2 Redis缓存 → L3 MySQL
func (r *ProductRepositoryImpl) FindBySKUID(ctx context.Context, skuID domain.SKU_ID) (*domain.Product, error) {
	cacheKey := fmt.Sprintf("product:%d", skuID.Value())

	// Step 1: 查询L1本地缓存
	if cached, ok := r.localCache.Get(cacheKey); ok {
		fmt.Printf("✅ [L1 Hit] SKUID=%d from Local Cache\n", skuID.Value())
		return cached.(*domain.Product), nil
	}
	fmt.Printf("❌ [L1 Miss] SKUID=%d, checking L2...\n", skuID.Value())

	// Step 2: 查询L2 Redis缓存
	if cached, err := r.redisCache.Get(ctx, cacheKey); err == nil {
		fmt.Printf("✅ [L2 Hit] SKUID=%d from Redis Cache\n", skuID.Value())
		product := r.unmarshalProduct(cached)

		// 回写L1缓存
		r.localCache.Set(cacheKey, product, 1*time.Minute)
		return product, nil
	}
	fmt.Printf("❌ [L2 Miss] SKUID=%d, checking L3...\n", skuID.Value())

	// Step 3: 查询MySQL数据库
	productDO, err := r.queryFromDB(ctx, skuID.Value())
	if err != nil {
		return nil, err
	}
	fmt.Printf("✅ [L3 Hit] SKUID=%d from MySQL\n", skuID.Value())

	// Step 4: 转换DO → Domain Model
	product := r.toDomain(productDO)

	// Step 5: 回写缓存（L2 Redis + L1 Local）
	r.writeCache(ctx, cacheKey, product)

	return product, nil
}

// BatchFindBySKUIDs 批量查询商品
func (r *ProductRepositoryImpl) BatchFindBySKUIDs(ctx context.Context, skuIDs []domain.SKU_ID) ([]*domain.Product, error) {
	products := make([]*domain.Product, 0, len(skuIDs))

	for _, skuID := range skuIDs {
		product, err := r.FindBySKUID(ctx, skuID)
		if err != nil {
			continue
		}
		products = append(products, product)
	}

	return products, nil
}

// Save 保存商品（新增或更新）
func (r *ProductRepositoryImpl) Save(ctx context.Context, product *domain.Product) error {
	// Step 1: 转换Domain Model → DO
	productDO := r.toDataObject(product)

	// Step 2: 保存到MySQL
	r.mu.Lock()
	r.mockDB[product.SKUID().Value()] = productDO
	r.mu.Unlock()
	fmt.Printf("💾 [DB Save] SKUID=%d saved to MySQL\n", product.SKUID().Value())

	// Step 3: 删除缓存（Cache Aside模式）
	cacheKey := fmt.Sprintf("product:%d", product.SKUID().Value())
	r.localCache.Delete(cacheKey)
	r.redisCache.Delete(ctx, cacheKey)
	fmt.Printf("🗑️  [Cache Delete] SKUID=%d cache invalidated\n", product.SKUID().Value())

	return nil
}

// Delete 删除商品
func (r *ProductRepositoryImpl) Delete(ctx context.Context, skuID domain.SKU_ID) error {
	r.mu.Lock()
	delete(r.mockDB, skuID.Value())
	r.mu.Unlock()

	cacheKey := fmt.Sprintf("product:%d", skuID.Value())
	r.localCache.Delete(cacheKey)
	r.redisCache.Delete(ctx, cacheKey)

	return nil
}

// InitTestData 初始化测试数据
func (r *ProductRepositoryImpl) InitTestData() {
	// 创建SPU
	spu := &SPUDO{
		ID:          1001,
		Title:       "Apple iPhone 17",
		CategoryID:  1,
		BrandID:     1,
		Description: "最新款iPhone",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	r.mockSPU[spu.ID] = spu

	// 创建商品1
	product1 := &ProductDO{
		ID:        1,
		SKUID:     10001,
		SPUID:     1001,
		SKUCode:   "SKU-10001",
		SKUName:   "iPhone 17 黑色 128GB",
		BasePrice: 759900, // ¥7599.00
		SpecColor: "黑色",
		SpecSize:  "128GB",
		Status:    "DRAFT",
		Images:    []string{"https://example.com/iphone17-black.jpg"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	r.mockDB[product1.SKUID] = product1

	// 创建商品2
	product2 := &ProductDO{
		ID:        2,
		SKUID:     10002,
		SPUID:     1001,
		SKUCode:   "SKU-10002",
		SKUName:   "iPhone 17 白色 256GB",
		BasePrice: 859900, // ¥8599.00
		SpecColor: "白色",
		SpecSize:  "256GB",
		Status:    "DRAFT",
		Images:    []string{"https://example.com/iphone17-white.jpg"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	r.mockDB[product2.SKUID] = product2
}

// 私有方法

func (r *ProductRepositoryImpl) queryFromDB(ctx context.Context, skuID int64) (*ProductDO, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	productDO, ok := r.mockDB[skuID]
	if !ok {
		return nil, fmt.Errorf("product not found: %d", skuID)
	}
	return productDO, nil
}

func (r *ProductRepositoryImpl) toDomain(do *ProductDO) *domain.Product {
	// 查询SPU
	r.mu.RLock()
	spuDO, _ := r.mockSPU[do.SPUID]
	r.mu.RUnlock()

	spu := domain.NewSPU(spuDO.ID, spuDO.Title, spuDO.CategoryID, spuDO.BrandID)
	spu.SetDescription(spuDO.Description)

	price, _ := domain.NewPrice(do.BasePrice, "CNY")
	specs := domain.NewSpecifications(do.SpecColor, do.SpecSize, nil)
	status := domain.ParseProductStatus(do.Status)

	return domain.ReconstructProduct(
		domain.NewSKU_ID(do.SKUID),
		spu,
		do.SKUCode,
		price,
		specs,
		status,
		do.Images,
		do.CreatedAt,
		do.UpdatedAt,
	)
}

func (r *ProductRepositoryImpl) toDataObject(product *domain.Product) *ProductDO {
	return &ProductDO{
		SKUID:     product.SKUID().Value(),
		SPUID:     product.SPU().ID(),
		SKUCode:   product.SKUCode(),
		SKUName:   product.SPU().Title(),
		BasePrice: product.BasePrice().Amount(),
		SpecColor: product.Specs().Color(),
		SpecSize:  product.Specs().Size(),
		Status:    product.Status().String(),
		Images:    product.Images(),
		CreatedAt: product.CreatedAt(),
		UpdatedAt: product.UpdatedAt(),
	}
}

func (r *ProductRepositoryImpl) unmarshalProduct(data []byte) *domain.Product {
	var productDO ProductDO
	json.Unmarshal(data, &productDO)
	return r.toDomain(&productDO)
}

func (r *ProductRepositoryImpl) writeCache(ctx context.Context, key string, product *domain.Product) {
	// 序列化为DO（更适合缓存）
	productDO := r.toDataObject(product)
	data, _ := json.Marshal(productDO)

	// 写L2 Redis（TTL 30分钟）
	r.redisCache.Set(ctx, key, data, 30*time.Minute)

	// 写L1本地缓存（TTL 1分钟）
	r.localCache.Set(key, product, 1*time.Minute)

	fmt.Printf("📝 [Cache Write] %s written to L1+L2\n", key)
}
