package service

import (
	"context"
	"testing"
	"time"

	"product-service/internal/domain"
	"product-service/internal/infrastructure/persistence"
)

func TestProductCenterPublishesSnapshotAndOutbox(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewProductCenterRepository()
	svc := NewProductCenterService(repo)

	result, err := svc.PublishCommand(ctx, domain.PublishProductVersionCommand{
		OperationID:        "op-test-publish",
		SourceType:         "LOCAL_OPS",
		BasePublishVersion: 0,
		Payload:            testProductPayload(91001, "inv:sku:91001:global", 10),
		RequestedAt:        time.Now(),
	})
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	if result.PublishVersion != 1 {
		t.Fatalf("expected publish version 1, got %d", result.PublishVersion)
	}
	snapshot, err := repo.GetSnapshot(ctx, result.ItemID, result.PublishVersion)
	if err != nil {
		t.Fatalf("snapshot not found: %v", err)
	}
	if snapshot.Payload.StockConfig.InventoryKey != "inv:sku:91001:global" {
		t.Fatalf("unexpected inventory key: %s", snapshot.Payload.StockConfig.InventoryKey)
	}
	events, err := repo.ListOutbox(ctx, domain.OutboxPending)
	if err != nil {
		t.Fatalf("list outbox failed: %v", err)
	}
	if len(events) != 1 || events[0].EventType != "ProductPublished" {
		t.Fatalf("expected one ProductPublished outbox event, got %+v", events)
	}
}

func TestProductCenterRejectsStalePublishVersion(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewProductCenterRepository()
	svc := NewProductCenterService(repo)
	payload := testProductPayload(91002, "inv:sku:91002:global", 10)

	first, err := svc.PublishCommand(ctx, domain.PublishProductVersionCommand{
		OperationID:        "op-test-publish-v1",
		SourceType:         "LOCAL_OPS",
		BasePublishVersion: 0,
		Payload:            payload,
		RequestedAt:        time.Now(),
	})
	if err != nil {
		t.Fatalf("first publish failed: %v", err)
	}
	payload.ItemID = first.ItemID
	payload.Title = "Updated Gift Card"

	_, err = svc.PublishCommand(ctx, domain.PublishProductVersionCommand{
		OperationID:        "op-test-publish-stale",
		SourceType:         "LOCAL_OPS",
		BasePublishVersion: 0,
		Payload:            payload,
		RequestedAt:        time.Now(),
	})
	if err == nil {
		t.Fatal("expected stale base publish version to fail")
	}
}

func testProductPayload(skuID int64, inventoryKey string, initialStock int) domain.ProductPublishPayload {
	return domain.ProductPublishPayload{
		SKUID:      skuID,
		SPUID:      skuID / 10,
		SKUCode:    "TEST-SKU",
		Title:      "Test Gift Card",
		CategoryID: 30105,
		BasePrice:  domain.Money{Amount: 10000, Currency: "CNY"},
		Resource: domain.ProductCenterResource{
			ResourceType: "GIFT_CARD",
			ResourceID:   "test_brand",
			Name:         "Test Brand",
		},
		Offer: domain.ProductCenterOffer{
			OfferID:   "offer_test",
			OfferType: domain.OfferFixedPrice,
			Price:     domain.Money{Amount: 10000, Currency: "CNY"},
		},
		StockConfig: domain.ProductStockConfig{
			InventoryKey:      inventoryKey,
			ManagementType:    domain.InventorySelfManaged,
			UnitType:          domain.InventoryUnitQuantity,
			DeductTiming:      domain.DeductOnOrder,
			InitialStock:      initialStock,
			LowStockThreshold: 1,
			Scope: domain.InventoryScope{
				ScopeType: "GLOBAL",
				ScopeID:   "0",
			},
		},
		InputSchema: domain.InputSchema{
			SchemaID: "test_input",
			Fields:   []domain.InputField{{Name: "email", Type: "string", Required: true}},
		},
		Fulfillment: domain.FulfillmentContract{
			Type: domain.FulfillmentIssueCode,
			Mode: "PAY_SUCCESS_THEN_ISSUE",
		},
		RefundRule: domain.RefundRule{
			RuleID:      "test_refund",
			Refundable:  true,
			Description: "before issue",
		},
	}
}
