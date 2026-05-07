package bootstrap

import (
	"order-service/internal/application/service"
	eventbus "order-service/internal/infrastructure/event"
	"order-service/internal/infrastructure/logger"
	"order-service/internal/infrastructure/mysql"
	"order-service/internal/infrastructure/persistence"
	interfaceevent "order-service/internal/interfaces/event"
	httphandler "order-service/internal/interfaces/http"
	"order-service/internal/interfaces/job"
	"order-service/internal/interfaces/rpc"
)

type App struct {
	Logger          *logger.Logger
	DB              *mysql.MySQLDB
	EventBus        *eventbus.EventBus
	OrderService    *service.OrderService
	HTTPHandler     *httphandler.OrderHandler
	RPCHandler      *rpc.OrderRPCHandler
	CloseOrderJob   *job.CloseTimeoutOrderJob
	RetryPublishJob *job.RetryPublishJob
	PaymentConsumer *interfaceevent.PaymentConsumer
	StockConsumer   *interfaceevent.StockConsumer
}

func NewApp() (*App, error) {
	log := logger.NewLogger()
	db, err := mysql.NewMySQLDBFromEnv(log)
	if err != nil {
		return nil, err
	}
	eventBus := eventbus.NewEventBus(log)
	orderRepo := persistence.NewOrderRepository(db, log)
	tx := persistence.NewTransactionManager(log)
	orderSvc := service.NewOrderService(orderRepo, tx, eventBus, log)

	return &App{
		Logger:          log,
		DB:              db,
		EventBus:        eventBus,
		OrderService:    orderSvc,
		HTTPHandler:     httphandler.NewOrderHandler(orderSvc, log),
		RPCHandler:      rpc.NewOrderRPCHandler(orderSvc, log),
		CloseOrderJob:   job.NewCloseTimeoutOrderJob(orderSvc, log),
		RetryPublishJob: job.NewRetryPublishJob(log),
		PaymentConsumer: interfaceevent.NewPaymentConsumer(orderSvc, log),
		StockConsumer:   interfaceevent.NewStockConsumer(orderSvc, log),
	}, nil
}
