package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"product-service/internal/application/dto"
)

func (h *ProductHandler) PublishProductVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.productCenterService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Product center service not configured")
		return
	}
	fmt.Printf("\n🌐 [Interface Layer - HTTP] POST /api/v1/product-center/publish\n")
	var req dto.PublishProductVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	resp, err := h.productCenterService.PublishProductVersion(r.Context(), req)
	if err != nil {
		h.responseError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) GetProductSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.productCenterService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Product center service not configured")
		return
	}
	prefix := "/api/v1/product-center/products/"
	rest := strings.TrimPrefix(r.URL.Path, prefix)
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) != 2 || parts[1] != "snapshot" {
		h.responseError(w, http.StatusBadRequest, "Expected /api/v1/product-center/products/{item_id}/snapshot")
		return
	}
	itemID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || itemID <= 0 {
		h.responseError(w, http.StatusBadRequest, "Invalid item_id")
		return
	}
	version, err := strconv.ParseInt(r.URL.Query().Get("version"), 10, 64)
	if err != nil || version <= 0 {
		h.responseError(w, http.StatusBadRequest, "Invalid version")
		return
	}
	resp, err := h.productCenterService.GetProductSnapshot(r.Context(), itemID, version)
	if err != nil {
		h.responseError(w, http.StatusNotFound, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) ListProductOutbox(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.productCenterService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Product center service not configured")
		return
	}
	resp, err := h.productCenterService.ListOutbox(r.Context(), r.URL.Query().Get("status"))
	if err != nil {
		h.responseError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}
