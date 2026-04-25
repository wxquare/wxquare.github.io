package domain

import "context"

type ProductRuntimeSource struct {
	SKUID      int64
	SPUID      int64
	SKUCode    string
	Title      string
	CategoryID int64
	BasePrice  Money
	Attributes map[string]string
}

type BuildContextInput struct {
	Product    ProductRuntimeSource
	Capability CategoryCapability
	Scene      Scene
	UserInputs map[string]string
}

// CategoryStrategy hides category-specific product, availability, booking, fulfillment, and refund differences.
type CategoryStrategy interface {
	CategoryID() int64
	BuildProductDefinition(ctx context.Context, input BuildContextInput) (*ProductDefinition, error)
	BuildResource(ctx context.Context, input BuildContextInput) (*ResourceContext, error)
	BuildOffer(ctx context.Context, input BuildContextInput) (*OfferContext, error)
	CheckAvailability(ctx context.Context, input BuildContextInput) (*AvailabilityContext, error)
	BuildInputSchema(ctx context.Context, input BuildContextInput) (*InputSchema, error)
	BuildBooking(ctx context.Context, input BuildContextInput) (*BookingRequirement, error)
	BuildFulfillment(ctx context.Context, input BuildContextInput) (*FulfillmentContract, error)
	BuildRefundRule(ctx context.Context, input BuildContextInput) (*RefundRule, error)
}

func BuildDefaultProductDefinition(input BuildContextInput) *ProductDefinition {
	return &ProductDefinition{
		SKUID:      input.Product.SKUID,
		SPUID:      input.Product.SPUID,
		SKUCode:    input.Product.SKUCode,
		Title:      input.Product.Title,
		CategoryID: input.Product.CategoryID,
		BasePrice:  input.Product.BasePrice,
		Attributes: input.Product.Attributes,
	}
}
