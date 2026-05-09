package strategy

import (
	"context"

	"product-service/internal/domain"
)

type baseStrategy struct {
	categoryID int64
}

func (s baseStrategy) CategoryID() int64 {
	return s.categoryID
}

func (s baseStrategy) BuildProductDefinition(ctx context.Context, input domain.BuildContextInput) (*domain.ProductDefinition, error) {
	return domain.BuildDefaultProductDefinition(input), nil
}

func moneyOrDefault(input domain.BuildContextInput) domain.Money {
	if input.Product.BasePrice.Currency == "" {
		return domain.Money{Amount: input.Product.BasePrice.Amount, Currency: "CNY"}
	}
	return input.Product.BasePrice
}
