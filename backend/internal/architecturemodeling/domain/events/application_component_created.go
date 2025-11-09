package events

import (
	"time"

	"easi/backend/internal/shared/domain"
)

// ApplicationComponentCreated is raised when a new application component is created
type ApplicationComponentCreated struct {
	domain.BaseEvent
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}

// NewApplicationComponentCreated creates a new ApplicationComponentCreated event
func NewApplicationComponentCreated(id, name, description string) ApplicationComponentCreated {
	return ApplicationComponentCreated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}
}

// EventType returns the event type name
func (e ApplicationComponentCreated) EventType() string {
	return "ApplicationComponentCreated"
}

// EventData returns the event data as a map for serialization
func (e ApplicationComponentCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"createdAt":   e.CreatedAt,
	}
}
