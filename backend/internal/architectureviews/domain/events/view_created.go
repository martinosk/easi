package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

// ViewCreated is raised when a new architecture view is created
type ViewCreated struct {
	domain.BaseEvent
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}

// NewViewCreated creates a new ViewCreated event
func NewViewCreated(id, name, description string) ViewCreated {
	return ViewCreated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}
}

// EventType returns the event type name
func (e ViewCreated) EventType() string {
	return "ViewCreated"
}

// EventData returns the event data as a map for serialization
func (e ViewCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"createdAt":   e.CreatedAt,
	}
}
