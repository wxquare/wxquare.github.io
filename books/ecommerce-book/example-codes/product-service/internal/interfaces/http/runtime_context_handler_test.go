package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"product-service/internal/application/service"
	"product-service/internal/domain"
	"product-service/internal/domain/strategy"
	"product-service/internal/infrastructure/persistence"
)

func TestRuntimeContextEndpointReturnsHotelContext(t *testing.T) {
	repo := persistence.NewRuntimeContextRepository()
	runtimeService := service.NewRuntimeContextService(repo, []domain.CategoryStrategy{
		strategy.NewTopupStrategy(),
		strategy.NewGiftCardStrategy(),
		strategy.NewFlightStrategy(),
		strategy.NewHotelStrategy(),
	})
	handler := NewProductHandler(nil, runtimeService)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/runtime-context?sku_id=40002&category_id=40104&scene=checkout", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		CategoryID int64 `json:"category_id"`
		Offer      struct {
			OfferType  string `json:"offer_type"`
			RatePlanID string `json:"rate_plan_id"`
		} `json:"offer"`
		Booking struct {
			Mode string `json:"mode"`
		} `json:"booking"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if body.CategoryID != 40104 {
		t.Fatalf("expected hotel category, got %d", body.CategoryID)
	}
	if body.Offer.OfferType != string(domain.OfferRatePlan) || body.Offer.RatePlanID == "" {
		t.Fatalf("expected hotel rate plan offer, got %+v", body.Offer)
	}
	if body.Booking.Mode != string(domain.BookingConfirmAfterPay) {
		t.Fatalf("expected confirm-after-pay booking, got %+v", body.Booking)
	}
}
