package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type InvitationAccepted struct {
	domain.BaseEvent
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	AcceptedAt time.Time `json:"acceptedAt"`
}

func (e InvitationAccepted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewInvitationAccepted(id string, email string) InvitationAccepted {
	now := time.Now().UTC()
	return InvitationAccepted{
		BaseEvent:  domain.NewBaseEvent(id),
		ID:         id,
		Email:      email,
		AcceptedAt: now,
	}
}

func (e InvitationAccepted) EventType() string {
	return "InvitationAccepted"
}

func (e InvitationAccepted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":         e.ID,
		"email":      e.Email,
		"acceptedAt": e.AcceptedAt.Format(time.RFC3339),
	}
}
