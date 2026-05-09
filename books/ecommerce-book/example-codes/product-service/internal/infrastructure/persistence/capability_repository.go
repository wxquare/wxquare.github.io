package persistence

import (
	"context"
	"fmt"

	"product-service/internal/domain"
)

type RuntimeContextRepository struct {
	capabilities map[int64]domain.CategoryCapability
	products     map[int64]domain.ProductRuntimeSource
}

func NewRuntimeContextRepository() *RuntimeContextRepository {
	return &RuntimeContextRepository{
		capabilities: map[int64]domain.CategoryCapability{
			10102: {
				CategoryID:              10102,
				CategoryName:            "Mobile Topup",
				ProductModelType:        domain.ProductModelAccountBased,
				ResourceType:            domain.ResourceNone,
				OfferType:               domain.OfferFixedPrice,
				AvailabilityType:        domain.AvailabilityUnlimited,
				InputSchemaID:           "mobile_topup_input",
				BookingMode:             domain.BookingNone,
				FulfillmentType:         domain.FulfillmentTopup,
				RefundRuleID:            "topup_no_refund_after_success",
				SupplierDependencyLevel: domain.SupplierDependencyMedium,
			},
			30105: {
				CategoryID:              30105,
				CategoryName:            "Gift Card",
				ProductModelType:        domain.ProductModelSingleSKU,
				ResourceType:            domain.ResourceGiftCard,
				OfferType:               domain.OfferFixedPrice,
				AvailabilityType:        domain.AvailabilityLocalPool,
				InputSchemaID:           "gift_card_input",
				BookingMode:             domain.BookingPayThenLock,
				FulfillmentType:         domain.FulfillmentIssueCode,
				RefundRuleID:            "gift_card_refund_before_code_assignment",
				SupplierDependencyLevel: domain.SupplierDependencyLow,
			},
			40102: {
				CategoryID:              40102,
				CategoryName:            "Flight",
				ProductModelType:        domain.ProductModelRealtimeOffer,
				ResourceType:            domain.ResourceFlight,
				OfferType:               domain.OfferRealtimeQuote,
				AvailabilityType:        domain.AvailabilityRealtimeSupplier,
				InputSchemaID:           "flight_passenger_input",
				BookingMode:             domain.BookingPreLock,
				FulfillmentType:         domain.FulfillmentTicket,
				RefundRuleID:            "flight_supplier_refund_policy",
				SupplierDependencyLevel: domain.SupplierDependencyHigh,
			},
			40104: {
				CategoryID:              40104,
				CategoryName:            "Hotel",
				ProductModelType:        domain.ProductModelResourceBased,
				ResourceType:            domain.ResourceHotel,
				OfferType:               domain.OfferRatePlan,
				AvailabilityType:        domain.AvailabilityRealtimeSupplier,
				InputSchemaID:           "hotel_guest_input",
				BookingMode:             domain.BookingConfirmAfterPay,
				FulfillmentType:         domain.FulfillmentBookingConfirm,
				RefundRuleID:            "hotel_rate_plan_refund_policy",
				SupplierDependencyLevel: domain.SupplierDependencyHigh,
			},
		},
		products: map[int64]domain.ProductRuntimeSource{
			10001: {
				SKUID:      10001,
				SPUID:      1001,
				SKUCode:    "TOPUP-100",
				Title:      "Mobile Topup 100",
				CategoryID: 10102,
				BasePrice:  domain.Money{Amount: 10000, Currency: "CNY"},
				Attributes: map[string]string{
					"denomination": "100",
					"operator":     "Generic Mobile Operator",
				},
			},
			10002: {
				SKUID:      10002,
				SPUID:      1002,
				SKUCode:    "GIFT-100",
				Title:      "Gift Card 100",
				CategoryID: 30105,
				BasePrice:  domain.Money{Amount: 10000, Currency: "CNY"},
				Attributes: map[string]string{
					"brand_code":    "GAME",
					"brand_name":    "Game Store",
					"denomination":  "100",
					"validity_days": "365",
				},
			},
			40001: {
				SKUID:      40001,
				SPUID:      4001,
				SKUCode:    "FLT-SHA-BJS",
				Title:      "Flight SHA to BJS",
				CategoryID: 40102,
				BasePrice:  domain.Money{Amount: 88000, Currency: "CNY"},
				Attributes: map[string]string{
					"route_code":     "SHA-BJS",
					"route_name":     "Shanghai to Beijing",
					"carrier":        "Generic Air",
					"departure_city": "Shanghai",
					"arrival_city":   "Beijing",
					"fare_class":     "Economy",
				},
			},
			40002: {
				SKUID:      40002,
				SPUID:      4002,
				SKUCode:    "HTL-DELUXE-STD",
				Title:      "Hotel Deluxe Room",
				CategoryID: 40104,
				BasePrice:  domain.Money{Amount: 56000, Currency: "CNY"},
				Attributes: map[string]string{
					"hotel_id":     "H10086",
					"hotel_name":   "City Garden Hotel",
					"room_type":    "Deluxe Room",
					"city":         "Shanghai",
					"star":         "5",
					"rate_plan_id": "STD-BREAKFAST-FREE-CANCEL",
					"meal":         "Breakfast",
					"cancellation": "Free cancel before deadline",
				},
			},
		},
	}
}

func (r *RuntimeContextRepository) GetCategoryCapability(ctx context.Context, categoryID int64) (domain.CategoryCapability, error) {
	capability, ok := r.capabilities[categoryID]
	if !ok {
		return domain.CategoryCapability{}, fmt.Errorf("category capability not found: %d", categoryID)
	}
	return capability, nil
}

func (r *RuntimeContextRepository) GetProductRuntimeSource(ctx context.Context, skuID int64) (domain.ProductRuntimeSource, error) {
	product, ok := r.products[skuID]
	if !ok {
		return domain.ProductRuntimeSource{}, fmt.Errorf("runtime product not found: %d", skuID)
	}
	return product, nil
}
