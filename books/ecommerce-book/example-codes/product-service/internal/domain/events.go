package domain

import "time"

type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}

type ProductCreatedEvent struct {
	SKUID     int64
	SPUID     int64
	BasePrice int64
	CreatedAt time.Time
}

func (e ProductCreatedEvent) EventType() string   { return "product.created" }
func (e ProductCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

type ProductOnShelfEvent struct {
	SKUID     int64
	OnShelfAt time.Time
}

func (e ProductOnShelfEvent) EventType() string   { return "product.on_shelf" }
func (e ProductOnShelfEvent) OccurredAt() time.Time { return e.OnShelfAt }

type ProductOffShelfEvent struct {
	SKUID      int64
	Reason     string
	OffShelfAt time.Time
}

func (e ProductOffShelfEvent) EventType() string   { return "product.off_shelf" }
func (e ProductOffShelfEvent) OccurredAt() time.Time { return e.OffShelfAt }

type PriceChangedEvent struct {
	SKUID     int64
	OldPrice  int64
	NewPrice  int64
	ChangedAt time.Time
}

func (e PriceChangedEvent) EventType() string   { return "product.price_changed" }
func (e PriceChangedEvent) OccurredAt() time.Time { return e.ChangedAt }
