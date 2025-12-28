package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type BusinessDomainUpdated struct {
	domain.BaseEvent
	ID          string
	Name        string
	Description string
	UpdatedAt   time.Time
}

func NewBusinessDomainUpdated(id, name, description string) BusinessDomainUpdated {
	return BusinessDomainUpdated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		UpdatedAt:   time.Now().UTC(),
	}
}

func (e BusinessDomainUpdated) EventType() string {
	return "BusinessDomainUpdated"
}

func (e BusinessDomainUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"updatedAt":   e.UpdatedAt,
	}
}
