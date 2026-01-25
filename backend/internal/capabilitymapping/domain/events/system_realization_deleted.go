package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type SystemRealizationDeleted struct {
	domain.BaseEvent
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

func NewSystemRealizationDeleted(id string) SystemRealizationDeleted {
	return SystemRealizationDeleted{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		DeletedAt: time.Now().UTC(),
	}
}

func (e SystemRealizationDeleted) EventType() string {
	return "SystemRealizationDeleted"
}

func (e SystemRealizationDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"deletedAt": e.DeletedAt,
	}
}

func (e SystemRealizationDeleted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
