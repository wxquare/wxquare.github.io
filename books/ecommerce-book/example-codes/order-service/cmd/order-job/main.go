package main

import (
	"context"
	"time"

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

	_, err = app.HTTPHandler.CreateOrder(ctx, model.CreateOrderRequest{
		CustomerID: "CUST-TIMEOUT",
		Items: []model.OrderItem{
			{SKUID: "SKU-1001", Quantity: 1, UnitPriceCents: 1999},
		},
	})
	if err != nil {
		panic(err)
	}

	before := time.Now().Add(time.Second)
	if err := app.CloseOrderJob.Run(ctx, before); err != nil {
		panic(err)
	}
	if err := app.RetryPublishJob.Run(ctx); err != nil {
		panic(err)
	}
}
