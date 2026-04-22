package job

import (
	"context"

	"order-service/internal/infra"
)

type RetryPublishJob struct {
	logger *infra.Logger
}

func NewRetryPublishJob(logger *infra.Logger) *RetryPublishJob {
	return &RetryPublishJob{logger: logger}
}

func (j *RetryPublishJob) Run(ctx context.Context) error {
	j.logger.Info("handler.job", "retry unpublished outbox events")
	return nil
}
