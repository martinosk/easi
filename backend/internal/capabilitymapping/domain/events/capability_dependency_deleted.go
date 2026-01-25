package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type CapabilityDependencyDeleted struct {
	domain.BaseEvent
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
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

func (e CapabilityDependencyDeleted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
