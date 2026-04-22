package service

import (
	"context"
	"time"

	"order-service/internal/model"
)

func (s *OrderService) HandlePaymentPaid(ctx context.Context, event model.PaymentPaidEvent) error {
	return s.MarkOrderPaid(ctx, model.MarkOrderPaidRequest{
		OrderID: event.OrderID,
		PaidAt:  event.PaidAt,
	})
}

func (s *OrderService) HandleStockReserved(ctx context.Context, event model.StockReservedEvent) error {
	s.logger.Info("service", "stock reserved for order_id=%s", event.OrderID)
	return nil
}

func (s *OrderService) HandleStockReserveFailed(ctx context.Context, event model.StockReserveFailedEvent) error {
	reason := event.Reason
	if reason == "" {
		reason = "stock reserve failed"
	}
	return s.CancelOrder(ctx, event.OrderID, reason)
}

func NewPaymentPaidEvent(orderID string) model.PaymentPaidEvent {
	return model.PaymentPaidEvent{
		OrderID: orderID,
		PaidAt:  time.Now(),
	}
}
