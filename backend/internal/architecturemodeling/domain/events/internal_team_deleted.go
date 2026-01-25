package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type InternalTeamDeleted struct {
	domain.BaseEvent
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	DeletedAt time.Time `json:"deletedAt"`
}

func (e InternalTeamDeleted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewInternalTeamDeleted(id, name string) InternalTeamDeleted {
	return InternalTeamDeleted{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		Name:      name,
		DeletedAt: time.Now().UTC(),
	}
}

func (e InternalTeamDeleted) EventType() string {
	return "InternalTeamDeleted"
}

func (e InternalTeamDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"name":      e.Name,
		"deletedAt": e.DeletedAt,
	}
}
