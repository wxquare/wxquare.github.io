package strategy

import (
	"context"

	"product-service/internal/domain"
)

type TopupStrategy struct {
	baseStrategy
}

func NewTopupStrategy() *TopupStrategy {
	return &TopupStrategy{baseStrategy: baseStrategy{categoryID: 10102}}
}

func (s *TopupStrategy) BuildResource(ctx context.Context, input domain.BuildContextInput) (*domain.ResourceContext, error) {
	return &domain.ResourceContext{
		ResourceType: domain.ResourceNone,
		Attributes: map[string]string{
			"resource_owner": "operator",
		},
	}, nil
}

func (s *TopupStrategy) BuildOffer(ctx context.Context, input domain.BuildContextInput) (*domain.OfferContext, error) {
	return &domain.OfferContext{
		OfferType: domain.OfferFixedPrice,
		OfferID:   input.Product.SKUCode,
		Price:     moneyOrDefault(input),
		Attributes: map[string]string{
			"denomination": input.Product.Attributes["denomination"],
		},
	}, nil
}

func (s *TopupStrategy) CheckAvailability(ctx context.Context, input domain.BuildContextInput) (*domain.AvailabilityContext, error) {
	return &domain.AvailabilityContext{
		AvailabilityType: domain.AvailabilityUnlimited,
		Sellable:         true,
		Quantity:         999999,
		Message:          "充值类商品使用近似无限库存，最终以充值接口结果为准",
	}, nil
}

func (s *TopupStrategy) BuildInputSchema(ctx context.Context, input domain.BuildContextInput) (*domain.InputSchema, error) {
	return &domain.InputSchema{
		SchemaID: "mobile_topup_input",
		Fields: []domain.InputField{
			{Name: "mobile_number", Label: "手机号", Type: "string", Required: true, Pattern: `^\d{8,15}$`},
		},
	}, nil
}

func (s *TopupStrategy) BuildBooking(ctx context.Context, input domain.BuildContextInput) (*domain.BookingRequirement, error) {
	return &domain.BookingRequirement{
		Mode:     domain.BookingNone,
		Required: false,
	}, nil
}

func (s *TopupStrategy) BuildFulfillment(ctx context.Context, input domain.BuildContextInput) (*domain.FulfillmentContract, error) {
	return &domain.FulfillmentContract{
		Type:       domain.FulfillmentTopup,
		Mode:       "SYNC_AFTER_PAY",
		TimeoutSec: 3,
	}, nil
}

func (s *TopupStrategy) BuildRefundRule(ctx context.Context, input domain.BuildContextInput) (*domain.RefundRule, error) {
	return &domain.RefundRule{
		RuleID:      "topup_no_refund_after_success",
		Refundable:  false,
		NeedReview:  false,
		Description: "充值成功后不可退款；供应商失败时由平台自动补偿或退款",
	}, nil
}
