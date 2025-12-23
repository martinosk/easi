package events

import (
	"time"

	"easi/backend/internal/shared/eventsourcing"
)

type InvitationRevoked struct {
	domain.BaseEvent
	ID        string
	RevokedAt time.Time
}

func NewInvitationRevoked(id string) InvitationRevoked {
	now := time.Now().UTC()
	return InvitationRevoked{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		RevokedAt: now,
	}
}

func (e InvitationRevoked) EventType() string {
	return "InvitationRevoked"
}

func (e InvitationRevoked) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"revokedAt": e.RevokedAt.Format(time.RFC3339),
	}
}
