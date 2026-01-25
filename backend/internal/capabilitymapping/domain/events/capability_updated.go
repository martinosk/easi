package events

import domain "easi/backend/internal/shared/eventsourcing"

type CapabilityUpdated struct {
	domain.BaseEvent
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
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

func (e CapabilityUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
