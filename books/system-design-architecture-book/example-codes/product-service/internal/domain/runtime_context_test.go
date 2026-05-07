package domain

import "testing"

func TestCategoryCapabilityDescribesEightLayerRequirements(t *testing.T) {
	capability := CategoryCapability{
		CategoryID:              10102,
		CategoryName:            "Mobile Topup",
		ProductModelType:        ProductModelAccountBased,
		ResourceType:            ResourceNone,
		OfferType:               OfferFixedPrice,
		AvailabilityType:        AvailabilityUnlimited,
		InputSchemaID:           "mobile_topup_input",
		BookingMode:             BookingNone,
		FulfillmentType:         FulfillmentTopup,
		RefundRuleID:            "topup_no_refund_after_success",
		SupplierDependencyLevel: SupplierDependencyMedium,
	}

	if capability.ProductModelType != ProductModelAccountBased {
		t.Fatalf("expected account-based product model, got %s", capability.ProductModelType)
	}
	if capability.AvailabilityType != AvailabilityUnlimited {
		t.Fatalf("expected unlimited availability, got %s", capability.AvailabilityType)
	}
	if capability.FulfillmentType != FulfillmentTopup {
		t.Fatalf("expected topup fulfillment, got %s", capability.FulfillmentType)
	}
}

func TestProductRuntimeContextIsCompleteWhenAllEightLayersExist(t *testing.T) {
	ctx := &ProductRuntimeContext{
		SKUID:      10001,
		CategoryID: 10102,
		Scene:      SceneDetail,
		ProductDefinition: &ProductDefinition{
			SKUID:     10001,
			SKUCode:   "TOPUP-100",
			Title:     "Mobile Topup 100",
			BasePrice: Money{Amount: 10000, Currency: "CNY"},
		},
		Resource: &ResourceContext{
			ResourceType: ResourceNone,
		},
		Offer: &OfferContext{
			OfferType: OfferFixedPrice,
			Price:     Money{Amount: 10000, Currency: "CNY"},
		},
		Availability: &AvailabilityContext{
			AvailabilityType: AvailabilityUnlimited,
			Sellable:         true,
		},
		InputSchema: &InputSchema{
			SchemaID: "mobile_topup_input",
			Fields:   []InputField{{Name: "mobile_number", Required: true}},
		},
		Booking: &BookingRequirement{
			Mode: BookingNone,
		},
		Fulfillment: &FulfillmentContract{
			Type: FulfillmentTopup,
		},
		RefundRule: &RefundRule{
			RuleID:      "topup_no_refund_after_success",
			Refundable:  false,
			Description: "充值成功后不可退款",
		},
	}

	if !ctx.IsComplete() {
		t.Fatal("expected runtime context to be complete")
	}
}
