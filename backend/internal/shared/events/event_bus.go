package events

import (
	"context"
	"errors"
	"fmt"
	"sync"

	domain "easi/backend/internal/shared/eventsourcing"
)

// EventHandler handles a domain event
type EventHandler interface {
	Handle(ctx context.Context, event domain.DomainEvent) error
}

// EventHandlerFunc is a function adapter for EventHandler
type EventHandlerFunc func(ctx context.Context, event domain.DomainEvent) error

// Handle implements EventHandler interface
func (f EventHandlerFunc) Handle(ctx context.Context, event domain.DomainEvent) error {
	return f(ctx, event)
}

// EventBus is responsible for publishing events to registered handlers
type EventBus interface {
	// Publish publishes events to all registered handlers
	Publish(ctx context.Context, events []domain.DomainEvent) error

	// Subscribe registers a handler for a specific event type
	Subscribe(eventType string, handler EventHandler)

	// SubscribeAll registers a handler for all event types
	SubscribeAll(handler EventHandler)
}

// InMemoryEventBus is an in-memory implementation of EventBus
type InMemoryEventBus struct {
	mu             sync.RWMutex
	handlers       map[string][]EventHandler // Handlers by event type
	globalHandlers []EventHandler            // Handlers that receive all events
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers:       make(map[string][]EventHandler),
		globalHandlers: make([]EventHandler, 0),
	}
}

// Publish publishes events to all registered handlers
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

// Subscribe registers a handler for a specific event type
func (b *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.handlers[eventType]; !exists {
		b.handlers[eventType] = make([]EventHandler, 0)
	}
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// SubscribeAll registers a handler for all event types
func (b *InMemoryEventBus) SubscribeAll(handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.globalHandlers = append(b.globalHandlers, handler)
}
