package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type RealizationInheritanceRemoval struct {
	SourceRealizationID string   `json:"sourceRealizationId"`
	CapabilityIDs       []string `json:"capabilityIds"`
}

type CapabilityRealizationsUninherited struct {
	domain.BaseEvent
	CapabilityID string                        `json:"capabilityId"`
	Removals     []RealizationInheritanceRemoval `json:"removals"`
	Timestamp    time.Time                     `json:"timestamp"`
}

func NewCapabilityRealizationsUninherited(capabilityID string, removals []RealizationInheritanceRemoval) CapabilityRealizationsUninherited {
	return CapabilityRealizationsUninherited{
		BaseEvent:    domain.NewBaseEvent(capabilityID),
		CapabilityID: capabilityID,
		Removals:     removals,
		Timestamp:    time.Now().UTC(),
	}
}

func (e CapabilityRealizationsUninherited) EventType() string {
	return "CapabilityRealizationsUninherited"
}

func (e CapabilityRealizationsUninherited) EventData() map[string]interface{} {
	return map[string]interface{}{
		"capabilityId": e.CapabilityID,
		"removals":     e.Removals,
		"timestamp":    e.Timestamp,
	}
}

func (e CapabilityRealizationsUninherited) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.CapabilityID
}
