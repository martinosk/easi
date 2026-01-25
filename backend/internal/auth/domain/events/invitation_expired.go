package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type InvitationExpired struct {
	domain.BaseEvent
	ID        string    `json:"id"`
	ExpiredAt time.Time `json:"expiredAt"`
}

func (e InvitationExpired) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewInvitationExpired(id string) InvitationExpired {
	now := time.Now().UTC()
	return InvitationExpired{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		ExpiredAt: now,
	}
}

func (e InvitationExpired) EventType() string {
	return "InvitationExpired"
}

func (e InvitationExpired) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"expiredAt": e.ExpiredAt.Format(time.RFC3339),
	}
}
