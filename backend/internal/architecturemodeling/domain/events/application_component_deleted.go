package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationComponentDeleted struct {
	domain.BaseEvent
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	DeletedAt time.Time `json:"deletedAt"`
}

func (e ApplicationComponentDeleted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewApplicationComponentDeleted(id, name string) ApplicationComponentDeleted {
	return ApplicationComponentDeleted{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		Name:      name,
		DeletedAt: time.Now().UTC(),
	}
}

func (e ApplicationComponentDeleted) EventType() string {
	return "ApplicationComponentDeleted"
}

func (e ApplicationComponentDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"name":      e.Name,
		"deletedAt": e.DeletedAt,
	}
}
