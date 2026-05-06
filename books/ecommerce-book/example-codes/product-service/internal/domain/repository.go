package domain

import "context"

type ProductRepository interface {
	FindBySKUID(ctx context.Context, skuID SKU_ID) (*Product, error)
	BatchFindBySKUIDs(ctx context.Context, skuIDs []SKU_ID) ([]*Product, error)
	Save(ctx context.Context, product *Product) error
	Delete(ctx context.Context, skuID SKU_ID) error
}
