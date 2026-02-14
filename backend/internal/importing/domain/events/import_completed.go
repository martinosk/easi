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
	ValueStreamsCreated int                      `json:"valueStreamsCreated"`
	RealizationsCreated int                      `json:"realizationsCreated"`
	CapabilityMappings  int                      `json:"capabilityMappings"`
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

func NewImportCompleted(id string, capabilitiesCreated, componentsCreated, valueStreamsCreated, realizationsCreated, capabilityMappings, domainAssignments int, errors []map[string]interface{}) ImportCompleted {
	return ImportCompleted{
		BaseEvent:           domain.NewBaseEvent(id),
		ID:                  id,
		CapabilitiesCreated: capabilitiesCreated,
		ComponentsCreated:   componentsCreated,
		ValueStreamsCreated: valueStreamsCreated,
		RealizationsCreated: realizationsCreated,
		CapabilityMappings:  capabilityMappings,
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
		"valueStreamsCreated": e.ValueStreamsCreated,
		"realizationsCreated": e.RealizationsCreated,
		"capabilityMappings":  e.CapabilityMappings,
		"domainAssignments":   e.DomainAssignments,
		"errors":              e.Errors,
		"completedAt":         e.CompletedAt,
	}
}
