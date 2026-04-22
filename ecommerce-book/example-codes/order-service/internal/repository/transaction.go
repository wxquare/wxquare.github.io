package repository

import (
	"context"

	"order-service/internal/infra"
)

type TransactionManager struct {
	logger *infra.Logger
}

func NewTransactionManager(logger *infra.Logger) *TransactionManager {
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
