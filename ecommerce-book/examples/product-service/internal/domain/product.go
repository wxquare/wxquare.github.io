package domain

import (
	"errors"
	"time"
)

// Product 聚合根（SKU维度）
type Product struct {
	skuID        SKU_ID
	spu          *SPU
	skuCode      string
	basePrice    Price
	specs        Specifications
	status       ProductStatus
	images       []string
	createdAt    time.Time
	updatedAt    time.Time
	domainEvents []DomainEvent
}

// NewProduct 创建新商品
func NewProduct(skuID SKU_ID, spu *SPU, skuCode string, basePrice Price, specs Specifications) *Product {
	product := &Product{
		skuID:        skuID,
		spu:          spu,
		skuCode:      skuCode,
		basePrice:    basePrice,
		specs:        specs,
		status:       ProductStatusDraft,
		images:       make([]string, 0),
		createdAt:    time.Now(),
		updatedAt:    time.Now(),
		domainEvents: make([]DomainEvent, 0),
	}

	product.addDomainEvent(ProductCreatedEvent{
		SKUID:     skuID.Value(),
		SPUID:     spu.ID(),
		BasePrice: basePrice.Amount(),
		CreatedAt: product.createdAt,
	})

	return product
}

// ReconstructProduct 重建聚合根（从数据库加载）
func ReconstructProduct(skuID SKU_ID, spu *SPU, skuCode string, basePrice Price,
	specs Specifications, status ProductStatus, images []string,
	createdAt, updatedAt time.Time) *Product {
	return &Product{
		skuID:        skuID,
		spu:          spu,
		skuCode:      skuCode,
		basePrice:    basePrice,
		specs:        specs,
		status:       status,
		images:       images,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		domainEvents: make([]DomainEvent, 0),
	}
}

// Getters
func (p *Product) SKUID() SKU_ID       { return p.skuID }
func (p *Product) SPU() *SPU           { return p.spu }
func (p *Product) SKUCode() string     { return p.skuCode }
func (p *Product) BasePrice() Price    { return p.basePrice }
func (p *Product) Specs() Specifications { return p.specs }
func (p *Product) Status() ProductStatus { return p.status }
func (p *Product) Images() []string    { return p.images }
func (p *Product) CreatedAt() time.Time { return p.createdAt }
func (p *Product) UpdatedAt() time.Time { return p.updatedAt }
func (p *Product) DomainEvents() []DomainEvent { return p.domainEvents }

func (p *Product) ClearDomainEvents() {
	p.domainEvents = make([]DomainEvent, 0)
}

// OnShelf 上架
func (p *Product) OnShelf() error {
	if p.status == ProductStatusOnShelf {
		return errors.New("商品已上架")
	}
	if p.basePrice.Amount() <= 0 {
		return errors.New("商品价格必须大于0")
	}
	if len(p.images) == 0 {
		return errors.New("商品必须有至少一张图片")
	}

	p.status = ProductStatusOnShelf
	p.updatedAt = time.Now()

	p.addDomainEvent(ProductOnShelfEvent{
		SKUID:     p.skuID.Value(),
		OnShelfAt: p.updatedAt,
	})

	return nil
}

// OffShelf 下架
func (p *Product) OffShelf(reason string) error {
	if p.status == ProductStatusOffShelf {
		return errors.New("商品已下架")
	}

	p.status = ProductStatusOffShelf
	p.updatedAt = time.Now()

	p.addDomainEvent(ProductOffShelfEvent{
		SKUID:      p.skuID.Value(),
		Reason:     reason,
		OffShelfAt: p.updatedAt,
	})

	return nil
}

// UpdateBasePrice 更新基础价格
func (p *Product) UpdateBasePrice(newPrice Price) error {
	if newPrice.Amount() <= 0 {
		return errors.New("价格必须大于0")
	}

	oldPrice := p.basePrice
	p.basePrice = newPrice
	p.updatedAt = time.Now()

	p.addDomainEvent(PriceChangedEvent{
		SKUID:     p.skuID.Value(),
		OldPrice:  oldPrice.Amount(),
		NewPrice:  newPrice.Amount(),
		ChangedAt: p.updatedAt,
	})

	return nil
}

// AddImage 添加图片
func (p *Product) AddImage(imageURL string) {
	p.images = append(p.images, imageURL)
	p.updatedAt = time.Now()
}

func (p *Product) addDomainEvent(event DomainEvent) {
	p.domainEvents = append(p.domainEvents, event)
}
