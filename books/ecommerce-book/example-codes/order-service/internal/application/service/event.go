package service

import (
	"context"
	"time"

	"order-service/internal/model"
)

func (s *OrderService) publishOrderCreated(ctx context.Context, order model.Order) error {
	return s.eventBus.Publish(ctx, model.OrderCreatedEvent{
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		TotalCents: order.TotalCents,
		OccurredAt: time.Now(),
	})
}

func (s *OrderService) publishOrderPaid(ctx context.Context, order model.Order) error {
	paidAt := time.Now()
	if order.PaidAt != nil {
		paidAt = *order.PaidAt
	}

	return s.eventBus.Publish(ctx, model.OrderPaidEvent{
		OrderID:    order.ID,
		PaidAt:     paidAt,
		OccurredAt: time.Now(),
	})
}

func (s *OrderService) publishOrderCancelled(ctx context.Context, order model.Order, reason string) error {
	return s.eventBus.Publish(ctx, model.OrderCancelledEvent{
		OrderID:    order.ID,
		Reason:     reason,
		OccurredAt: time.Now(),
	})
}
