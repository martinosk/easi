package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type InvitationCreated struct {
	domain.BaseEvent
	ID           string
	Email        string
	Role         string
	InviterID    string
	InviterEmail string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

func NewInvitationCreated(
	id string,
	email string,
	role string,
	inviterID string,
	inviterEmail string,
	expiresAt time.Time,
) InvitationCreated {
	return InvitationCreated{
		BaseEvent:    domain.NewBaseEvent(id),
		ID:           id,
		Email:        email,
		Role:         role,
		InviterID:    inviterID,
		InviterEmail: inviterEmail,
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    expiresAt,
	}
}

func (e InvitationCreated) EventType() string {
	return "InvitationCreated"
}

func (e InvitationCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":           e.ID,
		"email":        e.Email,
		"role":         e.Role,
		"inviterID":    e.InviterID,
		"inviterEmail": e.InviterEmail,
		"createdAt":    e.CreatedAt.Format(time.RFC3339),
		"expiresAt":    e.ExpiresAt.Format(time.RFC3339),
	}
}
