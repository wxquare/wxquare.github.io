package persistence

import (
	"context"
	"fmt"
	"sync"

	"product-service/internal/domain"
)

type InventoryRepository struct {
	mu      sync.RWMutex
	records map[string]*domain.InventoryRecord
	ledgers map[string][]domain.InventoryLedger
}

func NewInventoryRepository() *InventoryRepository {
	return &InventoryRepository{
		records: make(map[string]*domain.InventoryRecord),
		ledgers: make(map[string][]domain.InventoryLedger),
	}
}

func (r *InventoryRepository) SaveInventory(ctx context.Context, config *domain.InventoryConfig, balance *domain.InventoryBalance, ledger *domain.InventoryLedger) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.records[config.InventoryKey]; ok {
		fmt.Printf("♻️  [DB Idempotent] InventoryKey=%s already exists, Available=%d\n",
			config.InventoryKey, existing.Balance.AvailableStock)
		return nil
	}
	r.records[config.InventoryKey] = &domain.InventoryRecord{
		Config:       cloneInventoryConfig(config),
		Balance:      cloneInventoryBalance(balance),
		Reservations: make(map[string]*domain.InventoryReservation),
	}
	if ledger != nil {
		r.ledgers[config.InventoryKey] = append(r.ledgers[config.InventoryKey], *cloneInventoryLedger(ledger))
	}
	fmt.Printf("💾 [DB Save] InventoryKey=%s, Total=%d, Available=%d\n",
		config.InventoryKey, balance.TotalStock, balance.AvailableStock)
	return nil
}

func (r *InventoryRepository) GetConfig(ctx context.Context, inventoryKey string) (*domain.InventoryConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	record, ok := r.records[inventoryKey]
	if !ok {
		return nil, fmt.Errorf("inventory config not found: %s", inventoryKey)
	}
	return cloneInventoryConfig(record.Config), nil
}

func (r *InventoryRepository) GetBalance(ctx context.Context, inventoryKey string) (*domain.InventoryBalance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	record, ok := r.records[inventoryKey]
	if !ok {
		return nil, fmt.Errorf("inventory balance not found: %s", inventoryKey)
	}
	return cloneInventoryBalance(record.Balance), nil
}

func (r *InventoryRepository) MutateInventory(ctx context.Context, inventoryKey string, mutate func(record *domain.InventoryRecord) (*domain.InventoryLedger, error)) (*domain.InventoryRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.records[inventoryKey]
	if !ok {
		return nil, fmt.Errorf("inventory record not found: %s", inventoryKey)
	}
	ledger, err := mutate(record)
	if err != nil {
		return nil, err
	}
	if ledger != nil {
		r.ledgers[inventoryKey] = append(r.ledgers[inventoryKey], *cloneInventoryLedger(ledger))
	}
	return cloneInventoryRecord(record), nil
}

func (r *InventoryRepository) ListLedger(ctx context.Context, inventoryKey string) ([]domain.InventoryLedger, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ledgers := r.ledgers[inventoryKey]
	result := make([]domain.InventoryLedger, 0, len(ledgers))
	for _, ledger := range ledgers {
		ledgerCopy := ledger
		result = append(result, ledgerCopy)
	}
	return result, nil
}

func cloneInventoryRecord(record *domain.InventoryRecord) *domain.InventoryRecord {
	if record == nil {
		return nil
	}
	cp := &domain.InventoryRecord{
		Config:       cloneInventoryConfig(record.Config),
		Balance:      cloneInventoryBalance(record.Balance),
		Reservations: make(map[string]*domain.InventoryReservation, len(record.Reservations)),
	}
	for k, v := range record.Reservations {
		reservationCopy := *v
		cp.Reservations[k] = &reservationCopy
	}
	return cp
}

func cloneInventoryConfig(config *domain.InventoryConfig) *domain.InventoryConfig {
	if config == nil {
		return nil
	}
	cp := *config
	return &cp
}

func cloneInventoryBalance(balance *domain.InventoryBalance) *domain.InventoryBalance {
	if balance == nil {
		return nil
	}
	cp := *balance
	return &cp
}

func cloneInventoryLedger(ledger *domain.InventoryLedger) *domain.InventoryLedger {
	if ledger == nil {
		return nil
	}
	cp := *ledger
	return &cp
}
