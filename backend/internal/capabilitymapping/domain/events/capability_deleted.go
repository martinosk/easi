package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type CapabilityDeleted struct {
	domain.BaseEvent
	ID        string
	DeletedAt time.Time
}

func NewCapabilityDeleted(id string) CapabilityDeleted {
	return CapabilityDeleted{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		DeletedAt: time.Now().UTC(),
	}
}

func (e CapabilityDeleted) EventType() string {
	return "CapabilityDeleted"
}

func (e CapabilityDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"deletedAt": e.DeletedAt,
	}
}
