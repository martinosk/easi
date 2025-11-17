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

func NewAggregateRoot() AggregateRoot {
	return AggregateRoot{
		id:      uuid.New().String(),
		version: 0,
		changes: make([]DomainEvent, 0),
	}
}

func NewAggregateRootWithID(id string) AggregateRoot {
	return AggregateRoot{
		id:      id,
		version: 0,
		changes: make([]DomainEvent, 0),
	}
}

func (a *AggregateRoot) ID() string {
	return a.id
}

func (a *AggregateRoot) Version() int {
	return a.version
}

func (a *AggregateRoot) GetUncommittedChanges() []DomainEvent {
	return a.changes
}

func (a *AggregateRoot) MarkChangesAsCommitted() {
	a.changes = make([]DomainEvent, 0)
}

func (a *AggregateRoot) LoadFromHistory(events []DomainEvent, applyFunc func(DomainEvent)) {
	for _, event := range events {
		applyFunc(event)
		a.version++
	}
}

func (a *AggregateRoot) RaiseEvent(event DomainEvent) {
	a.changes = append(a.changes, event)
	a.version++
}

func (a *AggregateRoot) IncrementVersion() {
	a.version++
}
