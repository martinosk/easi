package events

import (
	"time"

	"easi/backend/internal/shared/domain"
)

type CapabilityDependencyDeleted struct {
	domain.BaseEvent
	ID        string
	DeletedAt time.Time
}

func NewCapabilityDependencyDeleted(id string) CapabilityDependencyDeleted {
	return CapabilityDependencyDeleted{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		DeletedAt: time.Now().UTC(),
	}
}

func (e CapabilityDependencyDeleted) EventType() string {
	return "CapabilityDependencyDeleted"
}

func (e CapabilityDependencyDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"deletedAt": e.DeletedAt,
	}
}
