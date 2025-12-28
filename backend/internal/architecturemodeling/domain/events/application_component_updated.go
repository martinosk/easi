package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

// ApplicationComponentUpdated is raised when an application component is updated
type ApplicationComponentUpdated struct {
	domain.BaseEvent
	ID          string
	Name        string
	Description string
	UpdatedAt   time.Time
}

// NewApplicationComponentUpdated creates a new ApplicationComponentUpdated event
func NewApplicationComponentUpdated(id, name, description string) ApplicationComponentUpdated {
	return ApplicationComponentUpdated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		UpdatedAt:   time.Now().UTC(),
	}
}

// EventType returns the event type name
func (e ApplicationComponentUpdated) EventType() string {
	return "ApplicationComponentUpdated"
}

// EventData returns the event data as a map for serialization
func (e ApplicationComponentUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"updatedAt":   e.UpdatedAt,
	}
}
