package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type InheritedRealization struct {
	CapabilityID         string    `json:"capabilityId"`
	ComponentID          string    `json:"componentId"`
	ComponentName        string    `json:"componentName"`
	RealizationLevel     string    `json:"realizationLevel"`
	Notes                string    `json:"notes"`
	Origin               string    `json:"origin"`
	SourceRealizationID  string    `json:"sourceRealizationId"`
	SourceCapabilityID   string    `json:"sourceCapabilityId"`
	SourceCapabilityName string    `json:"sourceCapabilityName"`
	LinkedAt             time.Time `json:"linkedAt"`
}

type CapabilityRealizationsInherited struct {
	domain.BaseEvent
	CapabilityID           string               `json:"capabilityId"`
	InheritedRealizations  []InheritedRealization `json:"inheritedRealizations"`
	Timestamp              time.Time            `json:"timestamp"`
}

func NewCapabilityRealizationsInherited(capabilityID string, realizations []InheritedRealization) CapabilityRealizationsInherited {
	return CapabilityRealizationsInherited{
		BaseEvent:           domain.NewBaseEvent(capabilityID),
		CapabilityID:        capabilityID,
		InheritedRealizations: realizations,
		Timestamp:           time.Now().UTC(),
	}
}

func (e CapabilityRealizationsInherited) EventType() string {
	return "CapabilityRealizationsInherited"
}

func (e CapabilityRealizationsInherited) EventData() map[string]interface{} {
	return map[string]interface{}{
		"capabilityId":          e.CapabilityID,
		"inheritedRealizations": e.InheritedRealizations,
		"timestamp":             e.Timestamp,
	}
}

func (e CapabilityRealizationsInherited) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.CapabilityID
}
