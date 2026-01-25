package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapabilityDeleted struct {
	domain.BaseEvent
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

func (e EnterpriseCapabilityDeleted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewEnterpriseCapabilityDeleted(id string) EnterpriseCapabilityDeleted {
	return EnterpriseCapabilityDeleted{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		DeletedAt: time.Now().UTC(),
	}
}

func (e EnterpriseCapabilityDeleted) EventType() string {
	return "EnterpriseCapabilityDeleted"
}

func (e EnterpriseCapabilityDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"deletedAt": e.DeletedAt,
	}
}
