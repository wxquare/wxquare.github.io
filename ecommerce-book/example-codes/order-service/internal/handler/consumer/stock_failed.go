package consumer

import (
	"context"

	"order-service/internal/model"
)

func (c *StockConsumer) HandleStockReserveFailed(ctx context.Context, event model.StockReserveFailedEvent) error {
	c.logger.Info("handler.consumer", "consume stock.reserve_failed order_id=%s", event.OrderID)
	return c.svc.HandleStockReserveFailed(ctx, event)
}
