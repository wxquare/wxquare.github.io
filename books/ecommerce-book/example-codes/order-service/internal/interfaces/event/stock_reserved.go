package event

import (
	"context"

	"order-service/internal/application/service"
	"order-service/internal/infrastructure/logger"
	"order-service/internal/model"
)

type StockConsumer struct {
	svc    *service.OrderService
	logger *logger.Logger
}

func NewStockConsumer(svc *service.OrderService, logger *logger.Logger) *StockConsumer {
	return &StockConsumer{svc: svc, logger: logger}
}

func (c *StockConsumer) HandleStockReserved(ctx context.Context, event model.StockReservedEvent) error {
	c.logger.Info("handler.consumer", "consume stock.reserved order_id=%s", event.OrderID)
	return c.svc.HandleStockReserved(ctx, event)
}
