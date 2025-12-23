package domain

import (
	"encoding/json"
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

// GenericDomainEvent is a generic event implementation for deserialization
type GenericDomainEvent struct {
	aggregateID string
	eventType   string
	eventData   map[string]interface{}
	occurredAt  time.Time
}

// NewGenericDomainEvent creates a new generic domain event from stored data
func NewGenericDomainEvent(aggregateID, eventType string, jsonData []byte, occurredAt time.Time) DomainEvent {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		// If unmarshal fails, return empty data
		data = make(map[string]interface{})
	}

	return &GenericDomainEvent{
		aggregateID: aggregateID,
		eventType:   eventType,
		eventData:   data,
		occurredAt:  occurredAt,
	}
}

// AggregateID returns the aggregate ID
func (e *GenericDomainEvent) AggregateID() string {
	return e.aggregateID
}

// EventType returns the event type
func (e *GenericDomainEvent) EventType() string {
	return e.eventType
}

// OccurredAt returns when the event occurred
func (e *GenericDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// EventData returns the event payload
func (e *GenericDomainEvent) EventData() map[string]interface{} {
	return e.eventData
}
