package domain

import (
	"time"
)

// DomainEvent represents an event that has occurred in the domain
type DomainEvent interface {
	// AggregateID returns the ID of the aggregate that raised the event
	AggregateID() string

	// EventType returns the type name of the event
	EventType() string

	// OccurredAt returns when the event occurred
	OccurredAt() time.Time

	// EventData returns the event payload as a map for serialization
	EventData() map[string]interface{}
}

// BaseEvent provides common functionality for all events
type BaseEvent struct {
	aggregateID string
	occurredAt  time.Time
}

// NewBaseEvent creates a new base event
func NewBaseEvent(aggregateID string) BaseEvent {
	return BaseEvent{
		aggregateID: aggregateID,
		occurredAt:  time.Now().UTC(),
	}
}

// AggregateID returns the aggregate ID
func (e BaseEvent) AggregateID() string {
	return e.aggregateID
}

// OccurredAt returns when the event occurred
func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}
