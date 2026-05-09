package httpapi

import (
	"encoding/json"
	"net/http"

	"common-services/internal/idgen"
	"common-services/internal/infrastructure/metrics"
)

type Handler struct {
	service  idgen.Service
	registry idgen.Registry
	metrics  *metrics.Recorder
	ready    func() bool
	mux      *http.ServeMux
}

type issueRequest struct {
	Namespace string `json:"namespace"`
	Caller    string `json:"caller"`
	RequestID string `json:"request_id"`
	Count     int    `json:"count"`
}

func NewHandler(service idgen.Service, registry idgen.Registry, recorder *metrics.Recorder, ready func() bool) *Handler {
	if ready == nil {
		ready = func() bool { return true }
	}
	h := &Handler{service: service, registry: registry, metrics: recorder, ready: ready, mux: http.NewServeMux()}
	h.mux.HandleFunc("/api/v1/ids/next", h.next)
	h.mux.HandleFunc("/api/v1/ids/batch", h.batch)
	h.mux.HandleFunc("/api/v1/namespaces", h.namespaces)
	h.mux.HandleFunc("/healthz", h.healthz)
	h.mux.HandleFunc("/readyz", h.readyz)
	h.mux.HandleFunc("/metrics", h.metricsText)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) next(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, idgen.NewError(idgen.ErrInvalidRequest, "", "method not allowed", false), http.StatusMethodNotAllowed)
		return
	}
	var req issueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, idgen.NewError(idgen.ErrInvalidRequest, "", "invalid json", false), http.StatusBadRequest)
		return
	}
	result, err := h.service.Next(r.Context(), idgen.IssueRequest{Namespace: req.Namespace, Caller: req.Caller, RequestID: req.RequestID})
	if err != nil {
		h.metrics.Inc("idgen_requests_total", map[string]string{"namespace": req.Namespace, "result": "error"})
		writeServiceError(w, err)
		return
	}
	h.metrics.Inc("idgen_requests_total", map[string]string{"namespace": req.Namespace, "generator": string(result.Generator), "result": "success"})
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) batch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, idgen.NewError(idgen.ErrInvalidRequest, "", "method not allowed", false), http.StatusMethodNotAllowed)
		return
	}
	var req issueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, idgen.NewError(idgen.ErrInvalidRequest, "", "invalid json", false), http.StatusBadRequest)
		return
	}
	result, err := h.service.Batch(r.Context(), idgen.IssueRequest{Namespace: req.Namespace, Caller: req.Caller, RequestID: req.RequestID, Count: req.Count})
	if err != nil {
		h.metrics.Inc("idgen_batch_requests_total", map[string]string{"namespace": req.Namespace, "result": "error"})
		writeServiceError(w, err)
		return
	}
	h.metrics.Inc("idgen_batch_requests_total", map[string]string{"namespace": req.Namespace, "result": "success"})
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) namespaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, idgen.NewError(idgen.ErrInvalidRequest, "", "method not allowed", false), http.StatusMethodNotAllowed)
		return
	}
	configs, err := h.registry.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"namespaces": configs})
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) readyz(w http.ResponseWriter, r *http.Request) {
	if !h.ready() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *Handler) metricsText(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte(h.metrics.Text()))
}

func writeServiceError(w http.ResponseWriter, err error) {
	if svcErr, ok := idgen.AsServiceError(err); ok {
		status := http.StatusBadRequest
		if svcErr.Retryable {
			status = http.StatusServiceUnavailable
		}
		writeError(w, svcErr, status)
		return
	}
	writeError(w, idgen.NewError(idgen.ErrInvalidRequest, "", err.Error(), false), http.StatusInternalServerError)
}

func writeError(w http.ResponseWriter, err *idgen.ServiceError, status int) {
	writeJSON(w, status, map[string]any{"error": err})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
