package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ValueStreamCreated struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

func NewValueStreamCreated(id, name, description string) ValueStreamCreated {
	return ValueStreamCreated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}
}

func (e ValueStreamCreated) EventType() string {
	return "ValueStreamCreated"
}

func (e ValueStreamCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"createdAt":   e.CreatedAt,
	}
}

func (e ValueStreamCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
