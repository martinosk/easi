package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

// ViewCreated is raised when a new architecture view is created
type ViewCreated struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPrivate   bool      `json:"isPrivate"`
	OwnerUserID string    `json:"ownerUserId"`
	OwnerEmail  string    `json:"ownerEmail"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (e ViewCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

// NewViewCreated creates a new ViewCreated event
func NewViewCreated(id, name, description string, isPrivate bool, ownerUserID, ownerEmail string) ViewCreated {
	return ViewCreated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		IsPrivate:   isPrivate,
		OwnerUserID: ownerUserID,
		OwnerEmail:  ownerEmail,
		CreatedAt:   time.Now().UTC(),
	}
}

// EventType returns the event type name
func (e ViewCreated) EventType() string {
	return "ViewCreated"
}

// EventData returns the event data as a map for serialization
func (e ViewCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"isPrivate":   e.IsPrivate,
		"ownerUserId": e.OwnerUserID,
		"ownerEmail":  e.OwnerEmail,
		"createdAt":   e.CreatedAt,
	}
}
