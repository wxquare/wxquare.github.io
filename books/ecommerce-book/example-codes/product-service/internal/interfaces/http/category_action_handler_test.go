package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"product-service/internal/application/service"
	"product-service/internal/domain"
	"product-service/internal/domain/strategy"
	"product-service/internal/infrastructure/persistence"
)

func TestTopupValidateAccountEndpoint(t *testing.T) {
	handler := newTestProductHandler()
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body := bytes.NewBufferString(`{"sku_id":10001,"mobile_number":"13800138000"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/topup/validate-account", body)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Valid    bool   `json:"valid"`
		Operator string `json:"operator"`
		Message  string `json:"message"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if !resp.Valid {
		t.Fatalf("expected valid topup account, got %+v", resp)
	}
	if resp.Operator == "" {
		t.Fatalf("expected operator to be filled")
	}
}

func TestFlightSearchEndpointReturnsRealtimeOffers(t *testing.T) {
	handler := newTestProductHandler()
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/travel/flights/search?from=SHA&to=BJS&date=2026-05-01&adult=1", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		RouteCode string `json:"route_code"`
		Offers    []struct {
			OfferToken       string `json:"offer_token"`
			OfferType        string `json:"offer_type"`
			AvailabilityType string `json:"availability_type"`
			BookingMode      string `json:"booking_mode"`
		} `json:"offers"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if resp.RouteCode != "SHA-BJS" {
		t.Fatalf("expected SHA-BJS route, got %s", resp.RouteCode)
	}
	if len(resp.Offers) == 0 {
		t.Fatal("expected at least one flight offer")
	}
	if resp.Offers[0].OfferType != string(domain.OfferRealtimeQuote) {
		t.Fatalf("expected realtime quote offer, got %+v", resp.Offers[0])
	}
	if resp.Offers[0].BookingMode != string(domain.BookingPreLock) {
		t.Fatalf("expected pre-lock booking, got %+v", resp.Offers[0])
	}
}

func newTestProductHandler() *ProductHandler {
	repo := persistence.NewRuntimeContextRepository()
	runtimeService := service.NewRuntimeContextService(repo, []domain.CategoryStrategy{
		strategy.NewTopupStrategy(),
		strategy.NewGiftCardStrategy(),
		strategy.NewFlightStrategy(),
		strategy.NewHotelStrategy(),
	})
	actionService := service.NewCategoryActionService(repo)
	return NewProductHandler(nil, runtimeService, actionService)
}
