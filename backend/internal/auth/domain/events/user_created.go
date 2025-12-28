package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type UserCreated struct {
	domain.BaseEvent
	ID           string
	Email        string
	Name         string
	Role         string
	Status       string
	ExternalID   string
	InvitationID string
	CreatedAt    time.Time
}

func NewUserCreated(
	id string,
	email string,
	name string,
	role string,
	externalID string,
	invitationID string,
) UserCreated {
	return UserCreated{
		BaseEvent:    domain.NewBaseEvent(id),
		ID:           id,
		Email:        email,
		Name:         name,
		Role:         role,
		Status:       "active",
		ExternalID:   externalID,
		InvitationID: invitationID,
		CreatedAt:    time.Now().UTC(),
	}
}

func (e UserCreated) EventType() string {
	return "UserCreated"
}

func (e UserCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":           e.ID,
		"email":        e.Email,
		"name":         e.Name,
		"role":         e.Role,
		"status":       e.Status,
		"externalId":   e.ExternalID,
		"invitationId": e.InvitationID,
		"createdAt":    e.CreatedAt.Format(time.RFC3339),
	}
}
