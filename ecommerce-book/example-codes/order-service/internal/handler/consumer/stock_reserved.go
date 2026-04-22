package consumer

import (
	"context"

	"order-service/internal/infra"
	"order-service/internal/model"
	"order-service/internal/service"
)

type StockConsumer struct {
	svc    *service.OrderService
	logger *infra.Logger
}

func NewStockConsumer(svc *service.OrderService, logger *infra.Logger) *StockConsumer {
	return &StockConsumer{svc: svc, logger: logger}
}

func (c *StockConsumer) HandleStockReserved(ctx context.Context, event model.StockReservedEvent) error {
	c.logger.Info("handler.consumer", "consume stock.reserved order_id=%s", event.OrderID)
	return c.svc.HandleStockReserved(ctx, event)
}
