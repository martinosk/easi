package events

import (
	"time"

	"easi/backend/internal/shared/domain"
)

// ComponentRelationUpdated is raised when a component relation is updated
type ComponentRelationUpdated struct {
	domain.BaseEvent
	ID          string
	Name        string
	Description string
	UpdatedAt   time.Time
}

// NewComponentRelationUpdated creates a new ComponentRelationUpdated event
func NewComponentRelationUpdated(id, name, description string) ComponentRelationUpdated {
	return ComponentRelationUpdated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		UpdatedAt:   time.Now().UTC(),
	}
}

// EventType returns the event type name
func (e ComponentRelationUpdated) EventType() string {
	return "ComponentRelationUpdated"
}

// EventData returns the event data as a map for serialization
func (e ComponentRelationUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"updatedAt":   e.UpdatedAt,
	}
}
