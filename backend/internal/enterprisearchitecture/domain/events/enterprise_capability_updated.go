package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapabilityUpdated struct {
	domain.BaseEvent
	ID          string
	Name        string
	Description string
	Category    string
	UpdatedAt   time.Time
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
