package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapabilityDeleted struct {
	domain.BaseEvent
	ID        string
	DeletedAt time.Time
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
