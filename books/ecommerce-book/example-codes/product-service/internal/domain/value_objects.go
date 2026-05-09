package domain

import (
	"errors"
	"fmt"
)

// SKU_ID 值对象
type SKU_ID struct {
	value int64
}

func NewSKU_ID(id int64) SKU_ID {
	return SKU_ID{value: id}
}

func (s SKU_ID) Value() int64 {
	return s.value
}

func (s SKU_ID) String() string {
	return fmt.Sprintf("SKU-%d", s.value)
}

// Price 值对象（使用分为单位，避免浮点精度问题）
type Price struct {
	amount   int64  // 金额（分）
	currency string // 货币（CNY）
}

func NewPrice(amount int64, currency string) (Price, error) {
	if amount < 0 {
		return Price{}, errors.New("价格不能为负数")
	}
	if currency == "" {
		currency = "CNY"
	}
	return Price{amount: amount, currency: currency}, nil
}

func (p Price) Amount() int64 {
	return p.amount
}

func (p Price) Currency() string {
	return p.currency
}

func (p Price) String() string {
	return fmt.Sprintf("¥%.2f", float64(p.amount)/100)
}

// Specifications 规格值对象
type Specifications struct {
	color string
	size  string
	attrs map[string]string
}

func NewSpecifications(color, size string, attrs map[string]string) Specifications {
	if attrs == nil {
		attrs = make(map[string]string)
	}
	return Specifications{
		color: color,
		size:  size,
		attrs: attrs,
	}
}

func (s Specifications) Color() string {
	return s.color
}

func (s Specifications) Size() string {
	return s.size
}

func (s Specifications) Attrs() map[string]string {
	return s.attrs
}

// ProductStatus 商品状态枚举
type ProductStatus int

const (
	ProductStatusDraft    ProductStatus = iota // 草稿
	ProductStatusOnShelf                       // 上架
	ProductStatusOffShelf                      // 下架
)

func (s ProductStatus) String() string {
	switch s {
	case ProductStatusDraft:
		return "DRAFT"
	case ProductStatusOnShelf:
		return "ON_SHELF"
	case ProductStatusOffShelf:
		return "OFF_SHELF"
	default:
		return "UNKNOWN"
	}
}

func ParseProductStatus(str string) ProductStatus {
	switch str {
	case "DRAFT":
		return ProductStatusDraft
	case "ON_SHELF":
		return ProductStatusOnShelf
	case "OFF_SHELF":
		return ProductStatusOffShelf
	default:
		return ProductStatusDraft
	}
}
