package repository

import (
	"context"
	"fmt"
	"time"

	"order-service/internal/infra"
	"order-service/internal/model"
)

type OrderRepository struct {
	db     *infra.MySQLDB
	logger *infra.Logger
}

func NewOrderRepository(db *infra.MySQLDB, logger *infra.Logger) *OrderRepository {
	return &OrderRepository{db: db, logger: logger}
}

func (r *OrderRepository) NextID(ctx context.Context) (string, error) {
	id, err := r.db.NextOrderID(ctx)
	if err != nil {
		return "", err
	}
	r.logger.Info("repository", "generated order_id=%s", id)
	return id, nil
}

func (r *OrderRepository) Save(ctx context.Context, order model.Order) error {
	if err := r.db.SaveOrder(ctx, order); err != nil {
		return err
	}
	r.logger.Info("repository", "saved order_id=%s status=%s", order.ID, order.Status)
	return nil
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (model.Order, error) {
	order, ok, err := r.db.GetOrder(ctx, id)
	if err != nil {
		return model.Order{}, err
	}
	if !ok {
		return model.Order{}, fmt.Errorf("order %s not found", id)
	}

	r.logger.Info("repository", "loaded order_id=%s status=%s", order.ID, order.Status)
	return order, nil
}

func (r *OrderRepository) ListPendingPaymentBefore(ctx context.Context, before time.Time, limit int) ([]model.Order, error) {
	if limit <= 0 {
		limit = 100
	}

	result, err := r.db.ListPendingPaymentBefore(ctx, before, limit)
	if err != nil {
		return nil, err
	}

	r.logger.Info("repository", "found %d timeout orders before=%s", len(result), before.Format(time.RFC3339))
	return result, nil
}
