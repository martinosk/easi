package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type AcquiredEntityDeleted struct {
	domain.BaseEvent
	ID        string
	Name      string
	DeletedAt time.Time
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
