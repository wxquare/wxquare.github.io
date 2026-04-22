package bootstrap

import (
	"order-service/internal/handler/consumer"
	httphandler "order-service/internal/handler/http"
	"order-service/internal/handler/job"
	"order-service/internal/handler/rpc"
	"order-service/internal/infra"
	"order-service/internal/repository"
	"order-service/internal/service"
)

type App struct {
	Logger          *infra.Logger
	DB              *infra.MySQLDB
	EventBus        *infra.EventBus
	OrderService    *service.OrderService
	HTTPHandler     *httphandler.OrderHandler
	RPCHandler      *rpc.OrderRPCHandler
	CloseOrderJob   *job.CloseTimeoutOrderJob
	RetryPublishJob *job.RetryPublishJob
	PaymentConsumer *consumer.PaymentConsumer
	StockConsumer   *consumer.StockConsumer
}

func NewApp() (*App, error) {
	logger := infra.NewLogger()
	db, err := infra.NewMySQLDBFromEnv(logger)
	if err != nil {
		return nil, err
	}
	eventBus := infra.NewEventBus(logger)
	orderRepo := repository.NewOrderRepository(db, logger)
	tx := repository.NewTransactionManager(logger)
	orderSvc := service.NewOrderService(orderRepo, tx, eventBus, logger)

	return &App{
		Logger:          logger,
		DB:              db,
		EventBus:        eventBus,
		OrderService:    orderSvc,
		HTTPHandler:     httphandler.NewOrderHandler(orderSvc, logger),
		RPCHandler:      rpc.NewOrderRPCHandler(orderSvc, logger),
		CloseOrderJob:   job.NewCloseTimeoutOrderJob(orderSvc, logger),
		RetryPublishJob: job.NewRetryPublishJob(logger),
		PaymentConsumer: consumer.NewPaymentConsumer(orderSvc, logger),
		StockConsumer:   consumer.NewStockConsumer(orderSvc, logger),
	}, nil
}
