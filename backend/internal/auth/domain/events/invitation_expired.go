package events

import (
	"time"

	"easi/backend/internal/shared/eventsourcing"
)

type InvitationExpired struct {
	domain.BaseEvent
	ID        string
	ExpiredAt time.Time
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
