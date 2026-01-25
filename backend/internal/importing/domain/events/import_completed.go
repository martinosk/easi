package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ImportCompleted struct {
	domain.BaseEvent
	ID                  string                   `json:"id"`
	CapabilitiesCreated int                      `json:"capabilitiesCreated"`
	ComponentsCreated   int                      `json:"componentsCreated"`
	RealizationsCreated int                      `json:"realizationsCreated"`
	DomainAssignments   int                      `json:"domainAssignments"`
	Errors              []map[string]interface{} `json:"errors"`
	CompletedAt         time.Time                `json:"completedAt"`
}

func (e ImportCompleted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewImportCompleted(id string, capabilitiesCreated, componentsCreated, realizationsCreated, domainAssignments int, errors []map[string]interface{}) ImportCompleted {
	return ImportCompleted{
		BaseEvent:           domain.NewBaseEvent(id),
		ID:                  id,
		CapabilitiesCreated: capabilitiesCreated,
		ComponentsCreated:   componentsCreated,
		RealizationsCreated: realizationsCreated,
		DomainAssignments:   domainAssignments,
		Errors:              errors,
		CompletedAt:         time.Now().UTC(),
	}
}

func (e ImportCompleted) EventType() string {
	return "ImportCompleted"
}

func (e ImportCompleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                  e.ID,
		"capabilitiesCreated": e.CapabilitiesCreated,
		"componentsCreated":   e.ComponentsCreated,
		"realizationsCreated": e.RealizationsCreated,
		"domainAssignments":   e.DomainAssignments,
		"errors":              e.Errors,
		"completedAt":         e.CompletedAt,
	}
}
