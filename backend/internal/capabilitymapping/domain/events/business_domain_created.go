package events

import (
	"time"

	"easi/backend/internal/shared/eventsourcing"
)

type BusinessDomainCreated struct {
	domain.BaseEvent
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}

func NewBusinessDomainCreated(id, name, description string) BusinessDomainCreated {
	return BusinessDomainCreated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}
}

func (e BusinessDomainCreated) EventType() string {
	return "BusinessDomainCreated"
}

func (e BusinessDomainCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"createdAt":   e.CreatedAt,
	}
}
