package job

import (
	"context"
	"time"

	"order-service/internal/infra"
	"order-service/internal/model"
	"order-service/internal/service"
)

type CloseTimeoutOrderJob struct {
	svc    *service.OrderService
	logger *infra.Logger
}

func NewCloseTimeoutOrderJob(svc *service.OrderService, logger *infra.Logger) *CloseTimeoutOrderJob {
	return &CloseTimeoutOrderJob{svc: svc, logger: logger}
}

func (j *CloseTimeoutOrderJob) Run(ctx context.Context, before time.Time) error {
	j.logger.Info("handler.job", "close timeout orders before=%s", before.Format(time.RFC3339))
	closed, err := j.svc.CloseTimeoutOrders(ctx, model.CloseTimeoutOrdersRequest{
		Before: before,
		Limit:  100,
	})
	if err != nil {
		return err
	}
	j.logger.Info("handler.job", "closed %d timeout orders", closed)
	return nil
}
