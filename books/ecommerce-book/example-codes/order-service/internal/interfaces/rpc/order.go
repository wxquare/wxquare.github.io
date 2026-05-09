package rpc

import (
	"context"

	"order-service/internal/application/service"
	"order-service/internal/infrastructure/logger"
	"order-service/internal/model"
)

type OrderRPCHandler struct {
	svc    *service.OrderService
	logger *logger.Logger
}

func NewOrderRPCHandler(svc *service.OrderService, logger *logger.Logger) *OrderRPCHandler {
	return &OrderRPCHandler{svc: svc, logger: logger}
}

func (h *OrderRPCHandler) CreateOrder(ctx context.Context, req model.CreateOrderRequest) (model.CreateOrderResponse, error) {
	h.logger.Info("handler.rpc", "OrderService.CreateOrder customer_id=%s", req.CustomerID)
	return h.svc.CreateOrder(ctx, req)
}
