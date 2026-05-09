package strategy

import (
	"context"
	"testing"

	"product-service/internal/domain"
)

func TestTopupStrategyBuildsUnlimitedPayThenFulfillContext(t *testing.T) {
	s := NewTopupStrategy()
	input := sampleInput(10001, 10102, "Mobile Topup 100", domain.AvailabilityUnlimited, domain.OfferFixedPrice)

	availability, err := s.CheckAvailability(context.Background(), input)
	if err != nil {
		t.Fatalf("check availability failed: %v", err)
	}
	booking, err := s.BuildBooking(context.Background(), input)
	if err != nil {
		t.Fatalf("build booking failed: %v", err)
	}
	fulfillment, err := s.BuildFulfillment(context.Background(), input)
	if err != nil {
		t.Fatalf("build fulfillment failed: %v", err)
	}

	if availability.AvailabilityType != domain.AvailabilityUnlimited || !availability.Sellable {
		t.Fatalf("expected unlimited sellable topup availability, got %+v", availability)
	}
	if booking.Mode != domain.BookingNone || booking.Required {
		t.Fatalf("expected no booking for topup, got %+v", booking)
	}
	if fulfillment.Type != domain.FulfillmentTopup || fulfillment.Mode != "SYNC_AFTER_PAY" {
		t.Fatalf("expected sync topup fulfillment, got %+v", fulfillment)
	}
}

func TestGiftCardStrategyBuildsLocalPoolAndIssueCodeContext(t *testing.T) {
	s := NewGiftCardStrategy()
	input := sampleInput(10002, 30105, "Gift Card 100", domain.AvailabilityLocalPool, domain.OfferFixedPrice)

	availability, err := s.CheckAvailability(context.Background(), input)
	if err != nil {
		t.Fatalf("check availability failed: %v", err)
	}
	fulfillment, err := s.BuildFulfillment(context.Background(), input)
	if err != nil {
		t.Fatalf("build fulfillment failed: %v", err)
	}
	refund, err := s.BuildRefundRule(context.Background(), input)
	if err != nil {
		t.Fatalf("build refund rule failed: %v", err)
	}

	if availability.AvailabilityType != domain.AvailabilityLocalPool || availability.Quantity != 128 {
		t.Fatalf("expected local pool quantity 128, got %+v", availability)
	}
	if fulfillment.Type != domain.FulfillmentIssueCode {
		t.Fatalf("expected issue-code fulfillment, got %+v", fulfillment)
	}
	if !refund.Refundable {
		t.Fatalf("expected gift card to be refundable before code assignment")
	}
}

func TestFlightStrategyBuildsRealtimeQuoteAndPreLockContext(t *testing.T) {
	s := NewFlightStrategy()
	input := sampleInput(40001, 40102, "Flight SHA to BJS", domain.AvailabilityRealtimeSupplier, domain.OfferRealtimeQuote)

	offer, err := s.BuildOffer(context.Background(), input)
	if err != nil {
		t.Fatalf("build offer failed: %v", err)
	}
	booking, err := s.BuildBooking(context.Background(), input)
	if err != nil {
		t.Fatalf("build booking failed: %v", err)
	}

	if offer.OfferType != domain.OfferRealtimeQuote || !offer.NeedRefresh {
		t.Fatalf("expected realtime quote needing refresh, got %+v", offer)
	}
	if booking.Mode != domain.BookingPreLock || !booking.ProviderSide {
		t.Fatalf("expected supplier-side pre-lock, got %+v", booking)
	}
}

func TestHotelStrategyBuildsRatePlanAndConfirmAfterPayContext(t *testing.T) {
	s := NewHotelStrategy()
	input := sampleInput(40002, 40104, "Hotel Deluxe Room", domain.AvailabilityRealtimeSupplier, domain.OfferRatePlan)

	offer, err := s.BuildOffer(context.Background(), input)
	if err != nil {
		t.Fatalf("build offer failed: %v", err)
	}
	booking, err := s.BuildBooking(context.Background(), input)
	if err != nil {
		t.Fatalf("build booking failed: %v", err)
	}

	if offer.OfferType != domain.OfferRatePlan || offer.RatePlanID == "" {
		t.Fatalf("expected hotel rate plan, got %+v", offer)
	}
	if booking.Mode != domain.BookingConfirmAfterPay {
		t.Fatalf("expected confirm-after-pay booking, got %+v", booking)
	}
}

func sampleInput(skuID int64, categoryID int64, title string, availability domain.AvailabilityType, offer domain.OfferType) domain.BuildContextInput {
	return domain.BuildContextInput{
		Product: domain.ProductRuntimeSource{
			SKUID:      skuID,
			SPUID:      skuID / 10,
			SKUCode:    title,
			Title:      title,
			CategoryID: categoryID,
			BasePrice:  domain.Money{Amount: 10000, Currency: "CNY"},
			Attributes: map[string]string{"sample": "true"},
		},
		Capability: domain.CategoryCapability{
			CategoryID:       categoryID,
			CategoryName:     title,
			AvailabilityType: availability,
			OfferType:        offer,
		},
		Scene: domain.SceneDetail,
	}
}
