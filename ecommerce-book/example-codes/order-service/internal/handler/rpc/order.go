package rpc

import (
	"context"

	"order-service/internal/infra"
	"order-service/internal/model"
	"order-service/internal/service"
)

type OrderRPCHandler struct {
	svc    *service.OrderService
	logger *infra.Logger
}

func NewOrderRPCHandler(svc *service.OrderService, logger *infra.Logger) *OrderRPCHandler {
	return &OrderRPCHandler{svc: svc, logger: logger}
}

func (h *OrderRPCHandler) CreateOrder(ctx context.Context, req model.CreateOrderRequest) (model.CreateOrderResponse, error) {
	h.logger.Info("handler.rpc", "OrderService.CreateOrder customer_id=%s", req.CustomerID)
	return h.svc.CreateOrder(ctx, req)
}
