package main

import (
	"context"

	"order-service/internal/application/service"
	"order-service/internal/bootstrap"
	"order-service/internal/model"
)

func main() {
	ctx := context.Background()
	app, err := bootstrap.NewApp()
	if err != nil {
		panic(err)
	}
	defer app.DB.Close()

	created, err := app.HTTPHandler.CreateOrder(ctx, model.CreateOrderRequest{
		CustomerID: "CUST-PAID",
		Items: []model.OrderItem{
			{SKUID: "SKU-1001", Quantity: 1, UnitPriceCents: 4999},
		},
	})
	if err != nil {
		panic(err)
	}

	if err := app.StockConsumer.HandleStockReserved(ctx, model.StockReservedEvent{OrderID: created.OrderID}); err != nil {
		panic(err)
	}
	if err := app.PaymentConsumer.HandlePaymentPaid(ctx, service.NewPaymentPaidEvent(created.OrderID)); err != nil {
		panic(err)
	}
}
