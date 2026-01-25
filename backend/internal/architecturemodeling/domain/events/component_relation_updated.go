package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ComponentRelationUpdated struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (e ComponentRelationUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewComponentRelationUpdated(id, name, description string) ComponentRelationUpdated {
	return ComponentRelationUpdated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		UpdatedAt:   time.Now().UTC(),
	}
}

func (e ComponentRelationUpdated) EventType() string {
	return "ComponentRelationUpdated"
}

func (e ComponentRelationUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"updatedAt":   e.UpdatedAt,
	}
}
