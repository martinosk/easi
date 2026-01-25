package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type BusinessDomainDeleted struct {
	domain.BaseEvent
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
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

func (e BusinessDomainDeleted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
