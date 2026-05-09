package model

import "time"

type Event interface {
	EventName() string
}

type OrderCreatedEvent struct {
	OrderID    string
	CustomerID string
	TotalCents int64
	OccurredAt time.Time
}

func (e OrderCreatedEvent) EventName() string {
	return "order.created"
}

type OrderPaidEvent struct {
	OrderID    string
	PaidAt     time.Time
	OccurredAt time.Time
}

func (e OrderPaidEvent) EventName() string {
	return "order.paid"
}

type OrderCancelledEvent struct {
	OrderID    string
	Reason     string
	OccurredAt time.Time
}

func (e OrderCancelledEvent) EventName() string {
	return "order.cancelled"
}

type PaymentPaidEvent struct {
	OrderID string
	PaidAt  time.Time
}

func (e PaymentPaidEvent) EventName() string {
	return "payment.paid"
}

type StockReservedEvent struct {
	OrderID string
}

func (e StockReservedEvent) EventName() string {
	return "stock.reserved"
}

type StockReserveFailedEvent struct {
	OrderID string
	Reason  string
}

func (e StockReserveFailedEvent) EventName() string {
	return "stock.reserve_failed"
}
