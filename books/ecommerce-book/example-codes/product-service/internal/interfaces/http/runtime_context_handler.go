package http

import (
	"fmt"
	"net/http"
	"strconv"
)

// GetRuntimeContext builds the eight-layer product transaction context.
// HTTP GET /api/v1/products/runtime-context?sku_id=40002&category_id=40104&scene=checkout
func (h *ProductHandler) GetRuntimeContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.runtimeContextService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Runtime context service not configured")
		return
	}

	fmt.Printf("\n🌐 [Interface Layer - HTTP] GET %s\n", r.URL.String())

	query := r.URL.Query()
	skuID, err := strconv.ParseInt(query.Get("sku_id"), 10, 64)
	if err != nil || skuID <= 0 {
		h.responseError(w, http.StatusBadRequest, "Invalid sku_id")
		return
	}
	categoryID, err := strconv.ParseInt(query.Get("category_id"), 10, 64)
	if err != nil || categoryID <= 0 {
		h.responseError(w, http.StatusBadRequest, "Invalid category_id")
		return
	}
	scene := query.Get("scene")
	if scene == "" {
		scene = "detail"
	}

	resp, err := h.runtimeContextService.BuildRuntimeContext(r.Context(), skuID, categoryID, scene)
	if err != nil {
		h.responseError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.responseJSON(w, http.StatusOK, resp)
	fmt.Printf("✅ [Interface Layer - HTTP] Runtime context response sent\n")
}
