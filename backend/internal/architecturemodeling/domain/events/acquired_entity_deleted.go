package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type AcquiredEntityDeleted struct {
	domain.BaseEvent
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	DeletedAt time.Time `json:"deletedAt"`
}

func (e AcquiredEntityDeleted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewAcquiredEntityDeleted(id, name string) AcquiredEntityDeleted {
	return AcquiredEntityDeleted{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		Name:      name,
		DeletedAt: time.Now().UTC(),
	}
}

func (e AcquiredEntityDeleted) EventType() string {
	return "AcquiredEntityDeleted"
}

func (e AcquiredEntityDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"name":      e.Name,
		"deletedAt": e.DeletedAt,
	}
}
