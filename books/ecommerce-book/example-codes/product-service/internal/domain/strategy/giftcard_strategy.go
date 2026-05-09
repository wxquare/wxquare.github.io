package strategy

import (
	"context"

	"product-service/internal/domain"
)

type GiftCardStrategy struct {
	baseStrategy
}

func NewGiftCardStrategy() *GiftCardStrategy {
	return &GiftCardStrategy{baseStrategy: baseStrategy{categoryID: 30105}}
}

func (s *GiftCardStrategy) BuildResource(ctx context.Context, input domain.BuildContextInput) (*domain.ResourceContext, error) {
	return &domain.ResourceContext{
		ResourceType: domain.ResourceGiftCard,
		ResourceID:   input.Product.Attributes["brand_code"],
		Name:         input.Product.Attributes["brand_name"],
		Attributes: map[string]string{
			"validity_days": input.Product.Attributes["validity_days"],
		},
	}, nil
}

func (s *GiftCardStrategy) BuildOffer(ctx context.Context, input domain.BuildContextInput) (*domain.OfferContext, error) {
	return &domain.OfferContext{
		OfferType: domain.OfferFixedPrice,
		OfferID:   input.Product.SKUCode,
		Price:     moneyOrDefault(input),
		Attributes: map[string]string{
			"denomination": input.Product.Attributes["denomination"],
		},
	}, nil
}

func (s *GiftCardStrategy) CheckAvailability(ctx context.Context, input domain.BuildContextInput) (*domain.AvailabilityContext, error) {
	return &domain.AvailabilityContext{
		AvailabilityType: domain.AvailabilityLocalPool,
		Sellable:         true,
		Quantity:         128,
		Message:          "库存来自平台券码池，发码时原子分配券码",
	}, nil
}

func (s *GiftCardStrategy) BuildInputSchema(ctx context.Context, input domain.BuildContextInput) (*domain.InputSchema, error) {
	return &domain.InputSchema{
		SchemaID: "gift_card_input",
		Fields: []domain.InputField{
			{Name: "recipient_email", Label: "收件邮箱", Type: "email", Required: false},
		},
	}, nil
}

func (s *GiftCardStrategy) BuildBooking(ctx context.Context, input domain.BuildContextInput) (*domain.BookingRequirement, error) {
	return &domain.BookingRequirement{
		Mode:       domain.BookingPayThenLock,
		Required:   true,
		TTLSeconds: 300,
		Attributes: map[string]string{
			"lock_target": "voucher_code_pool",
		},
	}, nil
}

func (s *GiftCardStrategy) BuildFulfillment(ctx context.Context, input domain.BuildContextInput) (*domain.FulfillmentContract, error) {
	return &domain.FulfillmentContract{
		Type:       domain.FulfillmentIssueCode,
		Mode:       "LOCAL_CODE_ASSIGNMENT",
		TimeoutSec: 1,
	}, nil
}

func (s *GiftCardStrategy) BuildRefundRule(ctx context.Context, input domain.BuildContextInput) (*domain.RefundRule, error) {
	return &domain.RefundRule{
		RuleID:      "gift_card_refund_before_code_assignment",
		Refundable:  true,
		NeedReview:  false,
		Description: "未分配券码前可退款；券码已发放后通常不可退款",
	}, nil
}
