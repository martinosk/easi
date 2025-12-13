package events

import (
	"time"

	"easi/backend/internal/shared/domain"
)

type ImportCompleted struct {
	domain.BaseEvent
	ID                  string
	CapabilitiesCreated int
	ComponentsCreated   int
	RealizationsCreated int
	DomainAssignments   int
	Errors              []map[string]interface{}
	CompletedAt         time.Time
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
