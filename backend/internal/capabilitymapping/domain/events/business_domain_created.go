package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type BusinessDomainCreated struct {
	domain.BaseEvent
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	DomainArchitectID string    `json:"domainArchitectId"`
	CreatedAt         time.Time `json:"createdAt"`
}

func NewBusinessDomainCreated(id, name, description, domainArchitectID string) BusinessDomainCreated {
	return BusinessDomainCreated{
		BaseEvent:         domain.NewBaseEvent(id),
		ID:                id,
		Name:              name,
		Description:       description,
		DomainArchitectID: domainArchitectID,
		CreatedAt:         time.Now().UTC(),
	}
}

func (e BusinessDomainCreated) EventType() string {
	return "BusinessDomainCreated"
}

func (e BusinessDomainCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                e.ID,
		"name":              e.Name,
		"description":       e.Description,
		"domainArchitectId": e.DomainArchitectID,
		"createdAt":         e.CreatedAt,
	}
}

func (e BusinessDomainCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
