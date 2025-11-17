package cqrs

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryCommandBus is a simple in-memory implementation of CommandBus
type InMemoryCommandBus struct {
	handlers map[string]CommandHandler
	mu       sync.RWMutex
}

func NewInMemoryCommandBus() *InMemoryCommandBus {
	return &InMemoryCommandBus{
		handlers: make(map[string]CommandHandler),
	}
}

func (b *InMemoryCommandBus) Register(commandName string, handler CommandHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[commandName] = handler
}

func (b *InMemoryCommandBus) Dispatch(ctx context.Context, cmd Command) error {
	b.mu.RLock()
	handler, exists := b.handlers[cmd.CommandName()]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handler registered for command: %s", cmd.CommandName())
	}

	return handler.Handle(ctx, cmd)
}
