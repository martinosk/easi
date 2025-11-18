package events

import (
	"easi/backend/internal/shared/domain"
)

type CapabilityUpdated struct {
	domain.BaseEvent
	ID          string
	Name        string
	Description string
}

func NewCapabilityUpdated(id, name, description string) CapabilityUpdated {
	return CapabilityUpdated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
	}
}

func (e CapabilityUpdated) EventType() string {
	return "CapabilityUpdated"
}

func (e CapabilityUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
	}
}
