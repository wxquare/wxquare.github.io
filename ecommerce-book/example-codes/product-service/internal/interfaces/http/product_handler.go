package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"product-service/internal/application/dto"
	"product-service/internal/application/service"
)

// ProductHandler HTTP处理器（接口层）
type ProductHandler struct {
	productService        *service.ProductService
	runtimeContextService *service.RuntimeContextService
	categoryActionService *service.CategoryActionService
	productCenterService  *service.ProductCenterService
	inventoryService      *service.InventoryService
	supplyOpsService      *service.SupplyOpsService
}

// NewProductHandler constructs the HTTP adapter for sync APIs.
func NewProductHandler(productService *service.ProductService, optionalServices ...interface{}) *ProductHandler {
	var runtimeContextService *service.RuntimeContextService
	var categoryActionService *service.CategoryActionService
	handler := &ProductHandler{
		productService:        productService,
		runtimeContextService: runtimeContextService,
		categoryActionService: categoryActionService,
	}
	for _, optionalService := range optionalServices {
		switch svc := optionalService.(type) {
		case *service.RuntimeContextService:
			handler.runtimeContextService = svc
		case *service.CategoryActionService:
			handler.categoryActionService = svc
		case *service.ProductCenterService:
			handler.productCenterService = svc
		case *service.InventoryService:
			handler.inventoryService = svc
		case *service.SupplyOpsService:
			handler.supplyOpsService = svc
		}
	}
	return handler
}

// RegisterRoutes 注册路由
func (h *ProductHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/product-center/publish", h.PublishProductVersion)
	mux.HandleFunc("/api/v1/product-center/outbox", h.ListProductOutbox)
	mux.HandleFunc("/api/v1/product-center/products/", h.GetProductSnapshot)
	mux.HandleFunc("/api/v1/inventory/create", h.CreateInventory)
	mux.HandleFunc("/api/v1/inventory/check", h.CheckStock)
	mux.HandleFunc("/api/v1/inventory/reserve", h.ReserveStock)
	mux.HandleFunc("/api/v1/inventory/confirm", h.ConfirmStock)
	mux.HandleFunc("/api/v1/inventory/release", h.ReleaseStock)
	mux.HandleFunc("/api/v1/inventory/adjust", h.AdjustInventory)
	mux.HandleFunc("/api/v1/inventory/ledger", h.GetInventoryLedger)
	mux.HandleFunc("/api/v1/supply/drafts", h.CreateSupplyDraft)
	mux.HandleFunc("/api/v1/supply/drafts/", h.HandleSupplyDraftAction)
	mux.HandleFunc("/api/v1/supply/qc/", h.HandleSupplyQCAction)
	mux.HandleFunc("/api/v1/supply/staging/", h.HandleSupplyStagingAction)
	mux.HandleFunc("/api/v1/supply/logs", h.ListSupplyLogs)
	mux.HandleFunc("/api/v1/products/runtime-context", h.GetRuntimeContext)
	mux.HandleFunc("/api/v1/topup/validate-account", h.ValidateTopupAccount)
	mux.HandleFunc("/api/v1/travel/flights/search", h.SearchFlights)
	mux.HandleFunc("/api/v1/products/", h.GetProduct)
	mux.HandleFunc("/api/v1/products/on-shelf", h.OnShelf)
	mux.HandleFunc("/api/v1/products/update-price", h.UpdatePrice)
}

// GetProduct 查询单个商品
// HTTP GET /api/v1/products/:sku_id
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n🌐 [Interface Layer - HTTP] GET %s\n", r.URL.Path)

	// 解析路径参数
	skuIDStr := r.URL.Path[len("/api/v1/products/"):]
	skuID, err := strconv.ParseInt(skuIDStr, 10, 64)
	if err != nil {
		h.responseError(w, http.StatusBadRequest, "Invalid SKU ID")
		return
	}

	// 调用应用服务
	req := &dto.GetProductRequest{SKUID: skuID}
	resp, err := h.productService.GetProduct(r.Context(), req)
	if err != nil {
		h.responseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 返回响应
	h.responseJSON(w, http.StatusOK, resp)
	fmt.Printf("✅ [Interface Layer - HTTP] Response sent\n")
}

// OnShelf 商品上架
// HTTP POST /api/v1/products/on-shelf
func (h *ProductHandler) OnShelf(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	fmt.Printf("\n🌐 [Interface Layer - HTTP] POST /api/v1/products/on-shelf\n")

	// 解析请求体
	var req dto.OnShelfRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 调用应用服务
	resp, err := h.productService.OnShelf(r.Context(), &req)
	if err != nil {
		h.responseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 返回响应
	h.responseJSON(w, http.StatusOK, resp)
	fmt.Printf("✅ [Interface Layer - HTTP] Response sent\n")
}

// UpdatePrice 更新价格
// HTTP POST /api/v1/products/update-price
func (h *ProductHandler) UpdatePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	fmt.Printf("\n🌐 [Interface Layer - HTTP] POST /api/v1/products/update-price\n")

	// 解析请求体
	var req dto.UpdatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 调用应用服务
	resp, err := h.productService.UpdateBasePrice(r.Context(), &req)
	if err != nil {
		h.responseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 返回响应
	h.responseJSON(w, http.StatusOK, resp)
	fmt.Printf("✅ [Interface Layer - HTTP] Response sent\n")
}

// 私有方法

func (h *ProductHandler) responseJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *ProductHandler) responseError(w http.ResponseWriter, statusCode int, message string) {
	h.responseJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
