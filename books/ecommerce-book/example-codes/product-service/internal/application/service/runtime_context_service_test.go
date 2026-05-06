package service

import (
	"context"
	"testing"

	"product-service/internal/domain"
	"product-service/internal/domain/strategy"
	"product-service/internal/infrastructure/persistence"
)

func TestRuntimeContextServiceBuildsFlightContext(t *testing.T) {
	repo := persistence.NewRuntimeContextRepository()
	svc := NewRuntimeContextService(repo, []domain.CategoryStrategy{
		strategy.NewTopupStrategy(),
		strategy.NewGiftCardStrategy(),
		strategy.NewFlightStrategy(),
		strategy.NewHotelStrategy(),
	})

	resp, err := svc.BuildRuntimeContext(context.Background(), 40001, 40102, "checkout")
	if err != nil {
		t.Fatalf("build runtime context failed: %v", err)
	}

	if resp.CategoryID != 40102 {
		t.Fatalf("expected flight category, got %d", resp.CategoryID)
	}
	if resp.Offer.OfferType != string(domain.OfferRealtimeQuote) {
		t.Fatalf("expected realtime quote, got %+v", resp.Offer)
	}
	if resp.Booking.Mode != string(domain.BookingPreLock) {
		t.Fatalf("expected pre-lock booking, got %+v", resp.Booking)
	}
	if !resp.Availability.Sellable {
		t.Fatalf("expected flight context to be sellable")
	}
}

func TestRuntimeContextServiceRejectsMismatchedCategory(t *testing.T) {
	repo := persistence.NewRuntimeContextRepository()
	svc := NewRuntimeContextService(repo, []domain.CategoryStrategy{
		strategy.NewTopupStrategy(),
		strategy.NewGiftCardStrategy(),
		strategy.NewFlightStrategy(),
		strategy.NewHotelStrategy(),
	})

	_, err := svc.BuildRuntimeContext(context.Background(), 40001, 10102, "detail")
	if err == nil {
		t.Fatal("expected mismatched category to fail")
	}
}
