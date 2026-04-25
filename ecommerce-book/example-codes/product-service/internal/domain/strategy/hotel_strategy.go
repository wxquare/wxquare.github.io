package strategy

import (
	"context"

	"product-service/internal/domain"
)

type HotelStrategy struct {
	baseStrategy
}

func NewHotelStrategy() *HotelStrategy {
	return &HotelStrategy{baseStrategy: baseStrategy{categoryID: 40104}}
}

func (s *HotelStrategy) BuildResource(ctx context.Context, input domain.BuildContextInput) (*domain.ResourceContext, error) {
	return &domain.ResourceContext{
		ResourceType: domain.ResourceHotel,
		ResourceID:   input.Product.Attributes["hotel_id"],
		Name:         input.Product.Attributes["hotel_name"],
		Attributes: map[string]string{
			"room_type": input.Product.Attributes["room_type"],
			"city":      input.Product.Attributes["city"],
			"star":      input.Product.Attributes["star"],
		},
	}, nil
}

func (s *HotelStrategy) BuildOffer(ctx context.Context, input domain.BuildContextInput) (*domain.OfferContext, error) {
	ratePlanID := input.Product.Attributes["rate_plan_id"]
	if ratePlanID == "" {
		ratePlanID = "STANDARD_RATE_PLAN"
	}

	return &domain.OfferContext{
		OfferType:   domain.OfferRatePlan,
		OfferID:     input.Product.SKUCode,
		RatePlanID:  ratePlanID,
		Price:       moneyOrDefault(input),
		NeedRefresh: true,
		Attributes: map[string]string{
			"meal":             input.Product.Attributes["meal"],
			"cancellation":     input.Product.Attributes["cancellation"],
			"price_source":     "calendar_or_supplier_quote",
			"settlement_model": "prepay",
		},
	}, nil
}

func (s *HotelStrategy) CheckAvailability(ctx context.Context, input domain.BuildContextInput) (*domain.AvailabilityContext, error) {
	return &domain.AvailabilityContext{
		AvailabilityType: domain.AvailabilityRealtimeSupplier,
		Sellable:         true,
		Quantity:         3,
		Message:          "房态房价按入住日期、间夜和入住人数动态确认",
		Attributes: map[string]string{
			"calendar_required": "true",
		},
	}, nil
}

func (s *HotelStrategy) BuildInputSchema(ctx context.Context, input domain.BuildContextInput) (*domain.InputSchema, error) {
	return &domain.InputSchema{
		SchemaID: "hotel_guest_input",
		Fields: []domain.InputField{
			{Name: "check_in", Label: "入住日期", Type: "date", Required: true},
			{Name: "check_out", Label: "离店日期", Type: "date", Required: true},
			{Name: "guest_name", Label: "入住人姓名", Type: "string", Required: true},
			{Name: "contact_phone", Label: "联系电话", Type: "string", Required: true},
		},
	}, nil
}

func (s *HotelStrategy) BuildBooking(ctx context.Context, input domain.BuildContextInput) (*domain.BookingRequirement, error) {
	return &domain.BookingRequirement{
		Mode:         domain.BookingConfirmAfterPay,
		Required:     true,
		TTLSeconds:   1200,
		ProviderSide: true,
		Attributes: map[string]string{
			"confirm_type": "supplier_confirmation",
		},
	}, nil
}

func (s *HotelStrategy) BuildFulfillment(ctx context.Context, input domain.BuildContextInput) (*domain.FulfillmentContract, error) {
	return &domain.FulfillmentContract{
		Type:       domain.FulfillmentBookingConfirm,
		Mode:       "ASYNC_CONFIRMATION",
		TimeoutSec: 60,
	}, nil
}

func (s *HotelStrategy) BuildRefundRule(ctx context.Context, input domain.BuildContextInput) (*domain.RefundRule, error) {
	return &domain.RefundRule{
		RuleID:      "hotel_rate_plan_refund_policy",
		Refundable:  true,
		NeedReview:  false,
		Description: "酒店退款取决于 Rate Plan，免费取消、限时取消和不可取消规则需要分别处理",
	}, nil
}
