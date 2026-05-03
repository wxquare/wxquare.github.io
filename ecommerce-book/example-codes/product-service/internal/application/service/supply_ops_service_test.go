package service

import (
	"context"
	"testing"

	"product-service/internal/application/dto"
	"product-service/internal/domain"
	"product-service/internal/infrastructure/persistence"
)

func TestSupplyOpsLocalOpsCanPublishAndCreateInventory(t *testing.T) {
	ctx := context.Background()
	productRepo := persistence.NewProductCenterRepository()
	inventoryRepo := persistence.NewInventoryRepository()
	productCenter := NewProductCenterService(productRepo)
	inventory := NewInventoryService(inventoryRepo)
	supply := NewSupplyOpsService(persistence.NewSupplyOpsRepository(), productCenter, inventory)

	payload := testProductPayload(93001, "inv:sku:93001:global", 6)
	draft, err := supply.CreateDraft(ctx, dto.CreateSupplyDraftRequest{
		OperationID: "op-supply-local",
		SourceType:  string(domain.SupplySourceLocalOps),
		OperatorID:  100,
		Payload:     toProductPublishPayloadDTO(payload),
	})
	if err != nil {
		t.Fatalf("create draft failed: %v", err)
	}
	submitted, err := supply.SubmitDraft(ctx, draft.DraftID)
	if err != nil {
		t.Fatalf("submit draft failed: %v", err)
	}
	if submitted.QCPolicy != string(domain.QCPolicyAutoApprove) {
		t.Fatalf("local ops should auto approve, got %s", submitted.QCPolicy)
	}
	published, err := supply.PublishStaging(ctx, submitted.StagingID)
	if err != nil {
		t.Fatalf("publish staging failed: %v", err)
	}
	if !published.InventoryReady {
		t.Fatalf("expected inventory ready, got %+v", published)
	}
	check, err := inventory.CheckStockCommand(ctx, domain.CheckStockRequest{InventoryKey: payload.StockConfig.InventoryKey, Qty: 6})
	if err != nil {
		t.Fatalf("check inventory failed: %v", err)
	}
	if !check.Sellable || check.Available != 6 {
		t.Fatalf("unexpected inventory state: %+v", check)
	}
}

func TestSupplyOpsMerchantRequiresQCBeforePublish(t *testing.T) {
	ctx := context.Background()
	productCenter := NewProductCenterService(persistence.NewProductCenterRepository())
	inventory := NewInventoryService(persistence.NewInventoryRepository())
	supply := NewSupplyOpsService(persistence.NewSupplyOpsRepository(), productCenter, inventory)

	draft, err := supply.CreateDraft(ctx, dto.CreateSupplyDraftRequest{
		OperationID: "op-supply-merchant",
		SourceType:  string(domain.SupplySourceMerchant),
		OperatorID:  200,
		Payload:     toProductPublishPayloadDTO(testProductPayload(93002, "inv:sku:93002:global", 3)),
	})
	if err != nil {
		t.Fatalf("create draft failed: %v", err)
	}
	submitted, err := supply.SubmitDraft(ctx, draft.DraftID)
	if err != nil {
		t.Fatalf("submit draft failed: %v", err)
	}
	if submitted.QCReviewID == "" {
		t.Fatalf("merchant submission should create QC review: %+v", submitted)
	}
	if _, err := supply.PublishStaging(ctx, submitted.StagingID); err == nil {
		t.Fatal("expected publish before QC approval to fail")
	}
	if _, err := supply.ApproveQC(ctx, submitted.QCReviewID, dto.ApproveQCRequest{ReviewerID: 300}); err != nil {
		t.Fatalf("approve qc failed: %v", err)
	}
	if _, err := supply.PublishStaging(ctx, submitted.StagingID); err != nil {
		t.Fatalf("publish after QC approval failed: %v", err)
	}
}
