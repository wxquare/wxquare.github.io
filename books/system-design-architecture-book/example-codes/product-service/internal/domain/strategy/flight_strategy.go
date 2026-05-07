package strategy

import (
	"context"

	"product-service/internal/domain"
)

type FlightStrategy struct {
	baseStrategy
}

func NewFlightStrategy() *FlightStrategy {
	return &FlightStrategy{baseStrategy: baseStrategy{categoryID: 40102}}
}

func (s *FlightStrategy) BuildResource(ctx context.Context, input domain.BuildContextInput) (*domain.ResourceContext, error) {
	return &domain.ResourceContext{
		ResourceType: domain.ResourceFlight,
		ResourceID:   input.Product.Attributes["route_code"],
		Name:         input.Product.Attributes["route_name"],
		Attributes: map[string]string{
			"carrier":        input.Product.Attributes["carrier"],
			"departure_city": input.Product.Attributes["departure_city"],
			"arrival_city":   input.Product.Attributes["arrival_city"],
		},
	}, nil
}

func (s *FlightStrategy) BuildOffer(ctx context.Context, input domain.BuildContextInput) (*domain.OfferContext, error) {
	return &domain.OfferContext{
		OfferType:   domain.OfferRealtimeQuote,
		OfferID:     "supplier_quote_required",
		Price:       moneyOrDefault(input),
		NeedRefresh: true,
		Attributes: map[string]string{
			"quote_source": "supplier_realtime_api",
			"fare_class":   input.Product.Attributes["fare_class"],
		},
	}, nil
}

func (s *FlightStrategy) CheckAvailability(ctx context.Context, input domain.BuildContextInput) (*domain.AvailabilityContext, error) {
	return &domain.AvailabilityContext{
		AvailabilityType: domain.AvailabilityRealtimeSupplier,
		Sellable:         true,
		Quantity:         9,
		Message:          "航班库存以创单前供应商实时确认和占座结果为准",
		Attributes: map[string]string{
			"cache_ttl_seconds": "30",
		},
	}, nil
}

func (s *FlightStrategy) BuildInputSchema(ctx context.Context, input domain.BuildContextInput) (*domain.InputSchema, error) {
	return &domain.InputSchema{
		SchemaID: "flight_passenger_input",
		Fields: []domain.InputField{
			{Name: "passenger_name", Label: "乘机人姓名", Type: "string", Required: true},
			{Name: "document_no", Label: "证件号", Type: "string", Required: true},
			{Name: "birth_date", Label: "出生日期", Type: "date", Required: true},
		},
	}, nil
}

func (s *FlightStrategy) BuildBooking(ctx context.Context, input domain.BuildContextInput) (*domain.BookingRequirement, error) {
	return &domain.BookingRequirement{
		Mode:         domain.BookingPreLock,
		Required:     true,
		TTLSeconds:   900,
		ProviderSide: true,
		Attributes: map[string]string{
			"lock_target": "seat_or_fare_inventory",
		},
	}, nil
}

func (s *FlightStrategy) BuildFulfillment(ctx context.Context, input domain.BuildContextInput) (*domain.FulfillmentContract, error) {
	return &domain.FulfillmentContract{
		Type:       domain.FulfillmentTicket,
		Mode:       "ASYNC_TICKETING",
		TimeoutSec: 30,
	}, nil
}

func (s *FlightStrategy) BuildRefundRule(ctx context.Context, input domain.BuildContextInput) (*domain.RefundRule, error) {
	return &domain.RefundRule{
		RuleID:      "flight_supplier_refund_policy",
		Refundable:  true,
		NeedReview:  true,
		Description: "机票退改签以航司和供应商规则为准，通常需要计算退改手续费",
	}, nil
}
