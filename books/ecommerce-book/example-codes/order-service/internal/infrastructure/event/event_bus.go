package event

import (
	"context"
	"sync"

	"order-service/internal/infrastructure/logger"
	"order-service/internal/model"
)

type EventBus struct {
	mu     sync.Mutex
	logger *logger.Logger
	events []model.Event
}

func NewEventBus(logger *logger.Logger) *EventBus {
	return &EventBus{logger: logger}
}

func (b *EventBus) Publish(ctx context.Context, event model.Event) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.events = append(b.events, event)
	b.logger.Info("infra.eventbus", "published event=%s payload=%+v", event.EventName(), event)
	return nil
}

func (b *EventBus) PublishedEvents() []model.Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	events := make([]model.Event, len(b.events))
	copy(events, b.events)
	return events
}
