package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EditGrantExpired struct {
	domain.BaseEvent
	ID        string    `json:"id"`
	ExpiredAt time.Time `json:"expiredAt"`
}

func (e EditGrantExpired) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewEditGrantExpired(id string) EditGrantExpired {
	now := time.Now().UTC()
	return EditGrantExpired{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		ExpiredAt: now,
	}
}

func (e EditGrantExpired) EventType() string {
	return "EditGrantExpired"
}

func (e EditGrantExpired) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"expiredAt": e.ExpiredAt.Format(time.RFC3339),
	}
}
