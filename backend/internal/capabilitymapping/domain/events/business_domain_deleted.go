package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type BusinessDomainDeleted struct {
	domain.BaseEvent
	ID        string
	DeletedAt time.Time
}

func NewBusinessDomainDeleted(id string) BusinessDomainDeleted {
	return BusinessDomainDeleted{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		DeletedAt: time.Now().UTC(),
	}
}

func (e BusinessDomainDeleted) EventType() string {
	return "BusinessDomainDeleted"
}

func (e BusinessDomainDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"deletedAt": e.DeletedAt,
	}
}
