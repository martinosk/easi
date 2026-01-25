package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationComponentCreated struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (e ApplicationComponentCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewApplicationComponentCreated(id, name, description string) ApplicationComponentCreated {
	return ApplicationComponentCreated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}
}

func (e ApplicationComponentCreated) EventType() string {
	return "ApplicationComponentCreated"
}

func (e ApplicationComponentCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"createdAt":   e.CreatedAt,
	}
}
