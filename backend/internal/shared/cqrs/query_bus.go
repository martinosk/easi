package cqrs

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryQueryBus is a simple in-memory implementation of QueryBus
type InMemoryQueryBus struct {
	handlers map[string]QueryHandler
	mu       sync.RWMutex
}

// NewInMemoryQueryBus creates a new in-memory query bus
func NewInMemoryQueryBus() *InMemoryQueryBus {
	return &InMemoryQueryBus{
		handlers: make(map[string]QueryHandler),
	}
}

// Register registers a handler for a query type
func (b *InMemoryQueryBus) Register(queryName string, handler QueryHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[queryName] = handler
}

// Dispatch dispatches a query to its handler
func (b *InMemoryQueryBus) Dispatch(ctx context.Context, query Query) (interface{}, error) {
	b.mu.RLock()
	handler, exists := b.handlers[query.QueryName()]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no handler registered for query: %s", query.QueryName())
	}

	return handler.Handle(ctx, query)
}
