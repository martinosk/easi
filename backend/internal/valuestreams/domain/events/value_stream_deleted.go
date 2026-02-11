package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type ValueStreamDeleted struct {
	domain.BaseEvent
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

func NewValueStreamDeleted(id string) ValueStreamDeleted {
	return ValueStreamDeleted{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		DeletedAt: time.Now().UTC(),
	}
}

func (e ValueStreamDeleted) EventType() string {
	return "ValueStreamDeleted"
}

func (e ValueStreamDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"deletedAt": e.DeletedAt,
	}
}

func (e ValueStreamDeleted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
