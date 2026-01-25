package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type InvitationRevoked struct {
	domain.BaseEvent
	ID        string    `json:"id"`
	RevokedAt time.Time `json:"revokedAt"`
}

func (e InvitationRevoked) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
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
