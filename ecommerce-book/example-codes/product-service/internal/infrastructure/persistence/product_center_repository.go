package persistence

import (
	"context"
	"fmt"
	"sync"

	"product-service/internal/domain"
)

type ProductCenterRepository struct {
	mu           sync.RWMutex
	nextItemID   int64
	products     map[int64]*domain.PublishedProduct
	productBySKU map[int64]int64
	snapshots    map[string]*domain.ProductSnapshot
	outbox       map[string]*domain.ProductOutboxEvent
}

func NewProductCenterRepository() *ProductCenterRepository {
	return &ProductCenterRepository{
		nextItemID:   800000,
		products:     make(map[int64]*domain.PublishedProduct),
		productBySKU: make(map[int64]int64),
		snapshots:    make(map[string]*domain.ProductSnapshot),
		outbox:       make(map[string]*domain.ProductOutboxEvent),
	}
}

func (r *ProductCenterRepository) NextItemID(ctx context.Context) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextItemID++
	return r.nextItemID
}

func (r *ProductCenterRepository) GetCurrentByItemID(ctx context.Context, itemID int64) (*domain.PublishedProduct, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	product, ok := r.products[itemID]
	if !ok {
		return nil, fmt.Errorf("published product not found: item_id=%d", itemID)
	}
	return clonePublishedProduct(product), nil
}

func (r *ProductCenterRepository) GetCurrentBySKUID(ctx context.Context, skuID int64) (*domain.PublishedProduct, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	itemID, ok := r.productBySKU[skuID]
	if !ok {
		return nil, fmt.Errorf("published product not found: sku_id=%d", skuID)
	}
	product, ok := r.products[itemID]
	if !ok {
		return nil, fmt.Errorf("published product not found: item_id=%d", itemID)
	}
	return clonePublishedProduct(product), nil
}

func (r *ProductCenterRepository) SavePublish(ctx context.Context, product *domain.PublishedProduct, snapshot *domain.ProductSnapshot, outbox *domain.ProductOutboxEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.products[product.ItemID] = clonePublishedProduct(product)
	r.productBySKU[product.SKUID] = product.ItemID
	r.snapshots[snapshotKey(snapshot.ItemID, snapshot.PublishVersion)] = cloneProductSnapshot(snapshot)
	r.outbox[outbox.EventID] = cloneProductOutbox(outbox)
	fmt.Printf("💾 [DB Save] ProductCenter item=%d version=%d snapshot=%s outbox=%s\n",
		product.ItemID, product.PublishVersion, snapshot.SnapshotID, outbox.EventID)
	return nil
}

func (r *ProductCenterRepository) GetSnapshot(ctx context.Context, itemID int64, publishVersion int64) (*domain.ProductSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	snapshot, ok := r.snapshots[snapshotKey(itemID, publishVersion)]
	if !ok {
		return nil, fmt.Errorf("product snapshot not found: item_id=%d version=%d", itemID, publishVersion)
	}
	return cloneProductSnapshot(snapshot), nil
}

func (r *ProductCenterRepository) ListOutbox(ctx context.Context, status domain.OutboxStatus) ([]domain.ProductOutboxEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	events := make([]domain.ProductOutboxEvent, 0)
	for _, event := range r.outbox {
		if status == "" || event.Status == status {
			events = append(events, *cloneProductOutbox(event))
		}
	}
	return events, nil
}

func snapshotKey(itemID int64, publishVersion int64) string {
	return fmt.Sprintf("%d:%d", itemID, publishVersion)
}

func clonePublishedProduct(p *domain.PublishedProduct) *domain.PublishedProduct {
	if p == nil {
		return nil
	}
	cp := *p
	return &cp
}

func cloneProductSnapshot(s *domain.ProductSnapshot) *domain.ProductSnapshot {
	if s == nil {
		return nil
	}
	cp := *s
	cp.Payload.Attributes = cloneStringMap(s.Payload.Attributes)
	cp.Payload.Images = append([]string(nil), s.Payload.Images...)
	cp.Payload.Resource.Attributes = cloneStringMap(s.Payload.Resource.Attributes)
	cp.Payload.Offer.Channels = append([]string(nil), s.Payload.Offer.Channels...)
	cp.Payload.Offer.Attributes = cloneStringMap(s.Payload.Offer.Attributes)
	cp.Payload.Fulfillment.Attributes = cloneStringMap(s.Payload.Fulfillment.Attributes)
	cp.Payload.RefundRule.Attributes = cloneStringMap(s.Payload.RefundRule.Attributes)
	return &cp
}

func cloneProductOutbox(e *domain.ProductOutboxEvent) *domain.ProductOutboxEvent {
	if e == nil {
		return nil
	}
	cp := *e
	cp.Payload = cloneInterfaceMap(e.Payload)
	return &cp
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	output := make(map[string]string, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}

func cloneInterfaceMap(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}
	output := make(map[string]interface{}, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}
