package dto

// CreateProductRequest 创建商品请求
type CreateProductRequest struct {
	SupplierID  int64  `json:"supplier_id"`
	SupplierSKU string `json:"supplier_sku"`
	Title       string `json:"title"`
	BasePrice   int64  `json:"base_price"`
	CategoryID  int64  `json:"category_id"`
	Color       string `json:"color"`
	Size        string `json:"size"`
}

// CreateProductResponse 创建商品响应
type CreateProductResponse struct {
	SKUID   int64  `json:"sku_id"`
	Message string `json:"message"`
}

// GetProductRequest 查询商品请求
type GetProductRequest struct {
	SKUID int64 `json:"sku_id" binding:"required"`
}

// GetProductResponse 查询商品响应
type GetProductResponse struct {
	SKUID     int64             `json:"sku_id"`
	SPUID     int64             `json:"spu_id"`
	SKUCode   string            `json:"sku_code"`
	SKUName   string            `json:"sku_name"`
	BasePrice int64             `json:"base_price"` // 分
	Specs     map[string]string `json:"specs"`
	Status    string            `json:"status"`
	Images    []string          `json:"images"`
	CreatedAt int64             `json:"created_at"` // Unix timestamp
	UpdatedAt int64             `json:"updated_at"`
}

// OnShelfRequest 上架请求
type OnShelfRequest struct {
	SKUID int64 `json:"sku_id" binding:"required"`
}

// OnShelfResponse 上架响应
type OnShelfResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UpdatePriceRequest 更新价格请求
type UpdatePriceRequest struct {
	SKUID    int64 `json:"sku_id" binding:"required"`
	NewPrice int64 `json:"new_price" binding:"required,gt=0"`
}

// UpdatePriceResponse 更新价格响应
type UpdatePriceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
