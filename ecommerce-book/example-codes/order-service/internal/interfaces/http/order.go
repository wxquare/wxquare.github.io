package http

import (
	"context"

	"order-service/internal/application/service"
	"order-service/internal/infrastructure/logger"
	"order-service/internal/model"
)

type OrderHandler struct {
	svc    *service.OrderService
	logger *logger.Logger
}

func NewOrderHandler(svc *service.OrderService, logger *logger.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, logger: logger}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req model.CreateOrderRequest) (model.CreateOrderResponse, error) {
	h.logger.Info("handler.http", "POST /orders customer_id=%s", req.CustomerID)
	return h.svc.CreateOrder(ctx, req)
}
