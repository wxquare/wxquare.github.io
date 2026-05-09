package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"product-service/internal/application/dto"
)

// ValidateTopupAccount validates user input for topup before checkout.
// HTTP POST /api/v1/topup/validate-account
func (h *ProductHandler) ValidateTopupAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.categoryActionService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Category action service not configured")
		return
	}

	var req dto.TopupValidateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.SKUID <= 0 || req.MobileNumber == "" {
		h.responseError(w, http.StatusBadRequest, "sku_id and mobile_number are required")
		return
	}

	resp, err := h.categoryActionService.ValidateTopupAccount(r.Context(), req)
	if err != nil {
		h.responseError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

// SearchFlights searches realtime flight offers.
// HTTP GET /api/v1/travel/flights/search?from=SHA&to=BJS&date=2026-05-01&adult=1
func (h *ProductHandler) SearchFlights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.categoryActionService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Category action service not configured")
		return
	}

	query := r.URL.Query()
	adult, err := strconv.Atoi(query.Get("adult"))
	if err != nil || adult <= 0 {
		adult = 1
	}

	resp, err := h.categoryActionService.SearchFlights(r.Context(), dto.FlightSearchRequest{
		From:  query.Get("from"),
		To:    query.Get("to"),
		Date:  query.Get("date"),
		Adult: adult,
	})
	if err != nil {
		h.responseError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}
