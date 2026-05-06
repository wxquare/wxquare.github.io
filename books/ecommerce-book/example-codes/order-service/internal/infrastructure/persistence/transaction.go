package persistence

import (
	"context"

	"order-service/internal/infrastructure/logger"
)

type TransactionManager struct {
	logger *logger.Logger
}

func NewTransactionManager(logger *logger.Logger) *TransactionManager {
	return &TransactionManager{logger: logger}
}

func (m *TransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	m.logger.Info("repository.tx", "begin transaction")
	if err := fn(ctx); err != nil {
		m.logger.Info("repository.tx", "rollback transaction: %v", err)
		return err
	}
	m.logger.Info("repository.tx", "commit transaction")
	return nil
}
