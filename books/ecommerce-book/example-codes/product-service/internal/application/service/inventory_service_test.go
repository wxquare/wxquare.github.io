package service

import (
	"context"
	"testing"
	"time"

	"product-service/internal/domain"
	"product-service/internal/infrastructure/persistence"
)

func TestInventoryReserveIsAtomicAndIdempotent(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewInventoryRepository()
	svc := NewInventoryService(repo)
	key := "inv:sku:92001:global"

	_, err := svc.CreateInventoryCommand(ctx, domain.CreateInventoryCommand{
		OperationID:  "op-inventory-init",
		SourceType:   "PRODUCT_PUBLISH",
		SourceID:     "staging-test",
		Reason:       "test init",
		InitialStock: 5,
		Config: domain.InventoryConfig{
			InventoryKey:   key,
			ItemID:         1,
			SKUID:          92001,
			ManagementType: domain.InventorySelfManaged,
			UnitType:       domain.InventoryUnitQuantity,
			DeductTiming:   domain.DeductOnOrder,
			Scope:          domain.InventoryScope{ScopeType: "GLOBAL", ScopeID: "0"},
		},
	})
	if err != nil {
		t.Fatalf("create inventory failed: %v", err)
	}

	reserved, err := svc.ReserveStockCommand(ctx, domain.ReserveStockRequest{
		InventoryKey:   key,
		OrderID:        "ORDER-1",
		Qty:            3,
		TTL:            15 * time.Minute,
		IdempotencyKey: "ORDER-1:" + key,
		OperatorType:   "ORDER",
	})
	if err != nil {
		t.Fatalf("reserve failed: %v", err)
	}
	if reserved.Remaining != 2 {
		t.Fatalf("expected remaining 2, got %d", reserved.Remaining)
	}

	again, err := svc.ReserveStockCommand(ctx, domain.ReserveStockRequest{
		InventoryKey:   key,
		OrderID:        "ORDER-1",
		Qty:            3,
		TTL:            15 * time.Minute,
		IdempotencyKey: "ORDER-1:" + key,
		OperatorType:   "ORDER",
	})
	if err != nil {
		t.Fatalf("repeat reserve failed: %v", err)
	}
	if !again.Idempotent || again.ReservationID != reserved.ReservationID {
		t.Fatalf("expected idempotent reserve, first=%+v again=%+v", reserved, again)
	}

	check, err := svc.CheckStockCommand(ctx, domain.CheckStockRequest{InventoryKey: key, Qty: 3})
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}
	if check.Sellable {
		t.Fatalf("expected qty=3 to be unsellable after reservation, got %+v", check)
	}
}

func TestInventoryConfirmAndReleaseStateMachine(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewInventoryRepository()
	svc := NewInventoryService(repo)
	key := "inv:sku:92002:global"

	_, err := svc.CreateInventoryCommand(ctx, domain.CreateInventoryCommand{
		OperationID:  "op-inventory-init-2",
		SourceType:   "PRODUCT_PUBLISH",
		Reason:       "test init",
		InitialStock: 4,
		Config: domain.InventoryConfig{
			InventoryKey:   key,
			ItemID:         2,
			SKUID:          92002,
			ManagementType: domain.InventorySelfManaged,
			UnitType:       domain.InventoryUnitQuantity,
			DeductTiming:   domain.DeductOnOrder,
			Scope:          domain.InventoryScope{ScopeType: "GLOBAL", ScopeID: "0"},
		},
	})
	if err != nil {
		t.Fatalf("create inventory failed: %v", err)
	}
	_, err = svc.ReserveStockCommand(ctx, domain.ReserveStockRequest{
		InventoryKey:   key,
		OrderID:        "ORDER-2",
		Qty:            2,
		TTL:            15 * time.Minute,
		IdempotencyKey: "ORDER-2:" + key,
		OperatorType:   "ORDER",
	})
	if err != nil {
		t.Fatalf("reserve failed: %v", err)
	}
	if err := svc.ConfirmStockCommand(ctx, domain.ConfirmStockRequest{InventoryKey: key, OrderID: "ORDER-2", EventID: "pay-1"}); err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	if err := svc.ConfirmStockCommand(ctx, domain.ConfirmStockRequest{InventoryKey: key, OrderID: "ORDER-2", EventID: "pay-1-replay"}); err != nil {
		t.Fatalf("repeat confirm should be idempotent: %v", err)
	}
	if err := svc.ReleaseStockCommand(ctx, domain.ReleaseStockRequest{InventoryKey: key, OrderID: "ORDER-2", EventID: "close-1"}); err == nil {
		t.Fatal("expected release after confirm to fail")
	}
	balance, err := repo.GetBalance(ctx, key)
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	if balance.AvailableStock != 2 || balance.SoldStock != 2 || !balance.InvariantHolds() {
		t.Fatalf("unexpected balance after confirm: %+v", balance)
	}
}
