package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"order-service/internal/infra"
	"order-service/internal/model"
	"order-service/internal/repository"
)

type OrderService struct {
	repo     *repository.OrderRepository
	tx       *repository.TransactionManager
	eventBus *infra.EventBus
	logger   *infra.Logger
}

func NewOrderService(repo *repository.OrderRepository, tx *repository.TransactionManager, eventBus *infra.EventBus, logger *infra.Logger) *OrderService {
	return &OrderService{
		repo:     repo,
		tx:       tx,
		eventBus: eventBus,
		logger:   logger,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req model.CreateOrderRequest) (model.CreateOrderResponse, error) {
	s.logger.Info("service", "create order requested customer_id=%s", req.CustomerID)

	if req.CustomerID == "" {
		return model.CreateOrderResponse{}, errors.New("customer_id is required")
	}
	if len(req.Items) == 0 {
		return model.CreateOrderResponse{}, errors.New("order items are required")
	}
	for _, item := range req.Items {
		if item.SKUID == "" || item.Quantity <= 0 || item.UnitPriceCents <= 0 {
			return model.CreateOrderResponse{}, fmt.Errorf("invalid order item: %+v", item)
		}
	}

	var order model.Order
	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		orderID, err := s.repo.NextID(txCtx)
		if err != nil {
			return err
		}
		order = model.NewOrder(orderID, req, time.Now())
		return s.repo.Save(txCtx, order)
	}); err != nil {
		return model.CreateOrderResponse{}, err
	}

	if err := s.publishOrderCreated(ctx, order); err != nil {
		return model.CreateOrderResponse{}, err
	}

	return model.CreateOrderResponse{
		OrderID:    order.ID,
		TotalCents: order.TotalCents,
		Status:     order.Status,
	}, nil
}

func (s *OrderService) MarkOrderPaid(ctx context.Context, req model.MarkOrderPaidRequest) error {
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		order, err := s.repo.FindByID(txCtx, req.OrderID)
		if err != nil {
			return err
		}
		if order.Status != model.OrderStatusPendingPayment {
			return fmt.Errorf("order %s cannot be paid from status %s", order.ID, order.Status)
		}

		paidAt := req.PaidAt
		order.Status = model.OrderStatusPaid
		order.PaidAt = &paidAt

		if err := s.repo.Save(txCtx, order); err != nil {
			return err
		}
		return s.publishOrderPaid(txCtx, order)
	})
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID string, reason string) error {
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		order, err := s.repo.FindByID(txCtx, orderID)
		if err != nil {
			return err
		}
		if order.Status != model.OrderStatusPendingPayment {
			return fmt.Errorf("order %s cannot be cancelled from status %s", order.ID, order.Status)
		}

		now := time.Now()
		order.Status = model.OrderStatusCancelled
		order.CancelledAt = &now

		if err := s.repo.Save(txCtx, order); err != nil {
			return err
		}
		return s.publishOrderCancelled(txCtx, order, reason)
	})
}
