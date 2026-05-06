package main

import (
	"context"
	"fmt"

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

	httpResp, err := app.HTTPHandler.CreateOrder(ctx, model.CreateOrderRequest{
		CustomerID: "CUST-001",
		Items: []model.OrderItem{
			{SKUID: "SKU-1001", Quantity: 2, UnitPriceCents: 3999},
			{SKUID: "SKU-2001", Quantity: 1, UnitPriceCents: 1299},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("HTTP response: %+v\n", httpResp)

	rpcResp, err := app.RPCHandler.CreateOrder(ctx, model.CreateOrderRequest{
		CustomerID: "CUST-002",
		Items: []model.OrderItem{
			{SKUID: "SKU-3001", Quantity: 1, UnitPriceCents: 9999},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("RPC response: %+v\n", rpcResp)
}
