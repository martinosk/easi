package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapabilityUpdated struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (e EnterpriseCapabilityUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewEnterpriseCapabilityUpdated(id, name, description, category string) EnterpriseCapabilityUpdated {
	return EnterpriseCapabilityUpdated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		Category:    category,
		UpdatedAt:   time.Now().UTC(),
	}
}

func (e EnterpriseCapabilityUpdated) EventType() string {
	return "EnterpriseCapabilityUpdated"
}

func (e EnterpriseCapabilityUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"category":    e.Category,
		"updatedAt":   e.UpdatedAt,
	}
}
