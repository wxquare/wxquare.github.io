package model

import "time"

type CreateOrderRequest struct {
	CustomerID string
	Items      []OrderItem
}

type CreateOrderResponse struct {
	OrderID    string
	TotalCents int64
	Status     OrderStatus
}

type MarkOrderPaidRequest struct {
	OrderID string
	PaidAt  time.Time
}

type CloseTimeoutOrdersRequest struct {
	Before time.Time
	Limit  int
}
