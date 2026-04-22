package event

import (
	"context"

	"order-service/internal/application/service"
	"order-service/internal/infrastructure/logger"
	"order-service/internal/model"
)

type PaymentConsumer struct {
	svc    *service.OrderService
	logger *logger.Logger
}

func NewPaymentConsumer(svc *service.OrderService, logger *logger.Logger) *PaymentConsumer {
	return &PaymentConsumer{svc: svc, logger: logger}
}

func (c *PaymentConsumer) HandlePaymentPaid(ctx context.Context, event model.PaymentPaidEvent) error {
	c.logger.Info("handler.consumer", "consume payment.paid order_id=%s", event.OrderID)
	return c.svc.HandlePaymentPaid(ctx, event)
}
