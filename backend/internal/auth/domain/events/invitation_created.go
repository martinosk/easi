package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type InvitationCreated struct {
	domain.BaseEvent
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	InviterID    string    `json:"inviterID"`
	InviterEmail string    `json:"inviterEmail"`
	CreatedAt    time.Time `json:"createdAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

func (e InvitationCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
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
