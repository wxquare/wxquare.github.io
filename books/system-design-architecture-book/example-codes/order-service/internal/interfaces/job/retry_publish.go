package job

import (
	"context"

	"order-service/internal/infrastructure/logger"
)

type RetryPublishJob struct {
	logger *logger.Logger
}

func NewRetryPublishJob(logger *logger.Logger) *RetryPublishJob {
	return &RetryPublishJob{logger: logger}
}

func (j *RetryPublishJob) Run(ctx context.Context) error {
	j.logger.Info("handler.job", "retry unpublished outbox events")
	return nil
}
