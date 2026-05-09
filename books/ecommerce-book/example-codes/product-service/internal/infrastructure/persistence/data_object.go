package persistence

import "time"

// ProductDO 数据对象（与数据库表对应）
type ProductDO struct {
	ID        int64     `json:"id"`
	SKUID     int64     `json:"sku_id"`
	SPUID     int64     `json:"spu_id"`
	SKUCode   string    `json:"sku_code"`
	SKUName   string    `json:"sku_name"`
	BasePrice int64     `json:"base_price"` // 分
	SpecColor string    `json:"spec_color"`
	SpecSize  string    `json:"spec_size"`
	Status    string    `json:"status"`
	Images    []string  `json:"images"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SPUDO SPU数据对象
type SPUDO struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	CategoryID  int64     `json:"category_id"`
	BrandID     int64     `json:"brand_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
