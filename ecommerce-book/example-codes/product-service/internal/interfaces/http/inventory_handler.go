package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"product-service/internal/application/dto"
)

func (h *ProductHandler) CreateInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.inventoryService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Inventory service not configured")
		return
	}
	fmt.Printf("\n🌐 [Interface Layer - HTTP] POST /api/v1/inventory/create\n")
	var req dto.CreateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	resp, err := h.inventoryService.CreateInventory(r.Context(), req)
	if err != nil {
		h.responseError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) CheckStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.inventoryService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Inventory service not configured")
		return
	}
	qty, _ := strconv.Atoi(r.URL.Query().Get("qty"))
	req := dto.CheckStockRequest{
		InventoryKey: r.URL.Query().Get("inventory_key"),
		Qty:          qty,
	}
	resp, err := h.inventoryService.CheckStock(r.Context(), req)
	if err != nil {
		h.responseError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) ReserveStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.inventoryService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Inventory service not configured")
		return
	}
	var req dto.ReserveStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	resp, err := h.inventoryService.ReserveStock(r.Context(), req)
	if err != nil {
		h.responseError(w, http.StatusConflict, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) ConfirmStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.inventoryService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Inventory service not configured")
		return
	}
	var req dto.ConfirmStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	resp, err := h.inventoryService.ConfirmStock(r.Context(), req)
	if err != nil {
		h.responseError(w, http.StatusConflict, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) ReleaseStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.inventoryService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Inventory service not configured")
		return
	}
	var req dto.ReleaseStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	resp, err := h.inventoryService.ReleaseStock(r.Context(), req)
	if err != nil {
		h.responseError(w, http.StatusConflict, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) AdjustInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.inventoryService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Inventory service not configured")
		return
	}
	var req dto.AdjustInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	resp, err := h.inventoryService.AdjustInventory(r.Context(), req)
	if err != nil {
		h.responseError(w, http.StatusConflict, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) GetInventoryLedger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.inventoryService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Inventory service not configured")
		return
	}
	resp, err := h.inventoryService.GetLedger(r.Context(), r.URL.Query().Get("inventory_key"))
	if err != nil {
		h.responseError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}
