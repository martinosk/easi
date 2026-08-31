package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapabilityCreated struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (e EnterpriseCapabilityCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewEnterpriseCapabilityCreated(id, name, description, category string) EnterpriseCapabilityCreated {
	return EnterpriseCapabilityCreated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		Category:    category,
		Active:      true,
		CreatedAt:   time.Now().UTC(),
	}
}

func (e EnterpriseCapabilityCreated) EventType() string {
	return "EnterpriseCapabilityCreated"
}

func (e EnterpriseCapabilityCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"category":    e.Category,
		"active":      e.Active,
		"createdAt":   e.CreatedAt,
	}
}
