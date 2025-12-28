package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type SystemRealizationDeleted struct {
	domain.BaseEvent
	ID        string
	DeletedAt time.Time
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
