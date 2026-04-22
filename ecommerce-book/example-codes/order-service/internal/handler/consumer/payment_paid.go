package consumer

import (
	"context"

	"order-service/internal/infra"
	"order-service/internal/model"
	"order-service/internal/service"
)

type PaymentConsumer struct {
	svc    *service.OrderService
	logger *infra.Logger
}

func NewPaymentConsumer(svc *service.OrderService, logger *infra.Logger) *PaymentConsumer {
	return &PaymentConsumer{svc: svc, logger: logger}
}

func (c *PaymentConsumer) HandlePaymentPaid(ctx context.Context, event model.PaymentPaidEvent) error {
	c.logger.Info("handler.consumer", "consume payment.paid order_id=%s", event.OrderID)
	return c.svc.HandlePaymentPaid(ctx, event)
}
