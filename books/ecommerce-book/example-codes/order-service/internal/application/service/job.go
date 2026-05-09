package service

import (
	"context"

	"order-service/internal/model"
)

func (s *OrderService) CloseTimeoutOrders(ctx context.Context, req model.CloseTimeoutOrdersRequest) (int, error) {
	orders, err := s.repo.ListPendingPaymentBefore(ctx, req.Before, req.Limit)
	if err != nil {
		return 0, err
	}

	closed := 0
	for _, order := range orders {
		if err := s.CancelOrder(ctx, order.ID, "payment timeout"); err != nil {
			return closed, err
		}
		closed++
	}
	return closed, nil
}
