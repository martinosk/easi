package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type BusinessDomainUpdated struct {
	domain.BaseEvent
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	DomainArchitectID string    `json:"domainArchitectId"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func NewBusinessDomainUpdated(id, name, description, domainArchitectID string) BusinessDomainUpdated {
	return BusinessDomainUpdated{
		BaseEvent:         domain.NewBaseEvent(id),
		ID:                id,
		Name:              name,
		Description:       description,
		DomainArchitectID: domainArchitectID,
		UpdatedAt:         time.Now().UTC(),
	}
}

func (e BusinessDomainUpdated) EventType() string {
	return "BusinessDomainUpdated"
}

func (e BusinessDomainUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                e.ID,
		"name":              e.Name,
		"description":       e.Description,
		"domainArchitectId": e.DomainArchitectID,
		"updatedAt":         e.UpdatedAt,
	}
}

func (e BusinessDomainUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
