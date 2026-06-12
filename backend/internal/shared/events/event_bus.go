package events

import (
	"context"
	"errors"
	"fmt"
	"sync"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EventHandler interface {
	Handle(ctx context.Context, event domain.DomainEvent) error
}

type EventHandlerFunc func(ctx context.Context, event domain.DomainEvent) error

func (f EventHandlerFunc) Handle(ctx context.Context, event domain.DomainEvent) error {
	return f(ctx, event)
}

type EventBus interface {
	Publish(ctx context.Context, events []domain.DomainEvent) error
	Subscribe(eventType string, handler EventHandler)
	SubscribeAll(handler EventHandler)
}

type InMemoryEventBus struct {
	mu             sync.RWMutex
	handlers       map[string][]EventHandler
	globalHandlers []EventHandler
}

func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers:       make(map[string][]EventHandler),
		globalHandlers: make([]EventHandler, 0),
	}
}

func (b *InMemoryEventBus) Publish(ctx context.Context, events []domain.DomainEvent) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var failures []error
	for _, event := range events {
		failures = append(failures, deliverToAll(ctx, b.globalHandlers, event, "global handler")...)
		failures = append(failures, deliverToAll(ctx, b.handlers[event.EventType()], event, "handler")...)
	}

	return errors.Join(failures...)
}

func deliverToAll(ctx context.Context, handlers []EventHandler, event domain.DomainEvent, label string) []error {
	var failures []error
	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			failures = append(failures, fmt.Errorf("%s failed for event %s: %w", label, event.EventType(), err))
		}
	}
	return failures
}

func (b *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.handlers[eventType]; !exists {
		b.handlers[eventType] = make([]EventHandler, 0)
	}
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *InMemoryEventBus) SubscribeAll(handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.globalHandlers = append(b.globalHandlers, handler)
}
