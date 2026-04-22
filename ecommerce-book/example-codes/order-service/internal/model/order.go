package model

import "time"

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "pending_payment"
	OrderStatusPaid           OrderStatus = "paid"
	OrderStatusCancelled      OrderStatus = "cancelled"
)

type Order struct {
	ID          string
	CustomerID  string
	Items       []OrderItem
	TotalCents  int64
	Status      OrderStatus
	CreatedAt   time.Time
	PaidAt      *time.Time
	CancelledAt *time.Time
}

type OrderItem struct {
	SKUID          string
	Quantity       int
	UnitPriceCents int64
}

func NewOrder(id string, req CreateOrderRequest, now time.Time) Order {
	items := make([]OrderItem, len(req.Items))
	copy(items, req.Items)

	return Order{
		ID:         id,
		CustomerID: req.CustomerID,
		Items:      items,
		TotalCents: CalculateTotalCents(items),
		Status:     OrderStatusPendingPayment,
		CreatedAt:  now,
	}
}

func CalculateTotalCents(items []OrderItem) int64 {
	var total int64
	for _, item := range items {
		total += int64(item.Quantity) * item.UnitPriceCents
	}
	return total
}
