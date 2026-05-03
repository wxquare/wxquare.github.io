package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"product-service/internal/application/dto"
)

func (h *ProductHandler) CreateSupplyDraft(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.supplyOpsService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Supply ops service not configured")
		return
	}
	var req dto.CreateSupplyDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	resp, err := h.supplyOpsService.CreateDraft(r.Context(), req)
	if err != nil {
		h.responseError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) HandleSupplyDraftAction(w http.ResponseWriter, r *http.Request) {
	if h.supplyOpsService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Supply ops service not configured")
		return
	}
	parts := splitPathAfter(r.URL.Path, "/api/v1/supply/drafts/")
	if len(parts) != 2 || parts[1] != "submit" {
		h.responseError(w, http.StatusBadRequest, "Expected /api/v1/supply/drafts/{draft_id}/submit")
		return
	}
	if r.Method != http.MethodPost {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	resp, err := h.supplyOpsService.SubmitDraft(r.Context(), parts[0])
	if err != nil {
		h.responseError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) HandleSupplyQCAction(w http.ResponseWriter, r *http.Request) {
	if h.supplyOpsService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Supply ops service not configured")
		return
	}
	parts := splitPathAfter(r.URL.Path, "/api/v1/supply/qc/")
	if len(parts) != 2 || parts[1] != "approve" {
		h.responseError(w, http.StatusBadRequest, "Expected /api/v1/supply/qc/{review_id}/approve")
		return
	}
	if r.Method != http.MethodPost {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var req dto.ApproveQCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	resp, err := h.supplyOpsService.ApproveQC(r.Context(), parts[0], req)
	if err != nil {
		h.responseError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) HandleSupplyStagingAction(w http.ResponseWriter, r *http.Request) {
	if h.supplyOpsService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Supply ops service not configured")
		return
	}
	parts := splitPathAfter(r.URL.Path, "/api/v1/supply/staging/")
	if len(parts) != 2 || parts[1] != "publish" {
		h.responseError(w, http.StatusBadRequest, "Expected /api/v1/supply/staging/{staging_id}/publish")
		return
	}
	if r.Method != http.MethodPost {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	resp, err := h.supplyOpsService.PublishStaging(r.Context(), parts[0])
	if err != nil {
		h.responseError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) ListSupplyLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.supplyOpsService == nil {
		h.responseError(w, http.StatusServiceUnavailable, "Supply ops service not configured")
		return
	}
	resp, err := h.supplyOpsService.ListLogs(r.Context(), r.URL.Query().Get("trace_id"))
	if err != nil {
		h.responseError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.responseJSON(w, http.StatusOK, resp)
}

func splitPathAfter(path string, prefix string) []string {
	rest := strings.TrimPrefix(path, prefix)
	if rest == path {
		return nil
	}
	return strings.Split(strings.Trim(rest, "/"), "/")
}
