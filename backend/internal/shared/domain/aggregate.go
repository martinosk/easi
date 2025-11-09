package domain

import (
	"github.com/google/uuid"
)

// AggregateRoot is the base for all aggregates in the system
type AggregateRoot struct {
	id      string
	version int
	changes []DomainEvent
}

// NewAggregateRoot creates a new aggregate root with a generated ID
func NewAggregateRoot() AggregateRoot {
	return AggregateRoot{
		id:      uuid.New().String(),
		version: 0,
		changes: make([]DomainEvent, 0),
	}
}

// NewAggregateRootWithID creates a new aggregate root with a specific ID
func NewAggregateRootWithID(id string) AggregateRoot {
	return AggregateRoot{
		id:      id,
		version: 0,
		changes: make([]DomainEvent, 0),
	}
}

// ID returns the aggregate's ID
func (a *AggregateRoot) ID() string {
	return a.id
}

// Version returns the current version of the aggregate
func (a *AggregateRoot) Version() int {
	return a.version
}

// GetUncommittedChanges returns all uncommitted domain events
func (a *AggregateRoot) GetUncommittedChanges() []DomainEvent {
	return a.changes
}

// MarkChangesAsCommitted clears the uncommitted changes
func (a *AggregateRoot) MarkChangesAsCommitted() {
	a.changes = make([]DomainEvent, 0)
}

// LoadFromHistory reconstructs the aggregate from historical events
func (a *AggregateRoot) LoadFromHistory(events []DomainEvent, applyFunc func(DomainEvent)) {
	for _, event := range events {
		applyFunc(event)
		a.version++
	}
}

// RaiseEvent records a new domain event
func (a *AggregateRoot) RaiseEvent(event DomainEvent) {
	a.changes = append(a.changes, event)
}

// IncrementVersion increments the aggregate version
func (a *AggregateRoot) IncrementVersion() {
	a.version++
}
