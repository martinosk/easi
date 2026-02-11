package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type ValueStreamUpdated struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func NewValueStreamUpdated(id, name, description string) ValueStreamUpdated {
	return ValueStreamUpdated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		UpdatedAt:   time.Now().UTC(),
	}
}

func (e ValueStreamUpdated) EventType() string {
	return "ValueStreamUpdated"
}

func (e ValueStreamUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"updatedAt":   e.UpdatedAt,
	}
}

func (e ValueStreamUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
