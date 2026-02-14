package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type CapabilityLevelChanged struct {
	domain.BaseEvent
	CapabilityID string    `json:"capabilityId"`
	OldLevel     string    `json:"oldLevel"`
	NewLevel     string    `json:"newLevel"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewCapabilityLevelChanged(capabilityID, oldLevel, newLevel string) CapabilityLevelChanged {
	return CapabilityLevelChanged{
		BaseEvent:    domain.NewBaseEvent(capabilityID),
		CapabilityID: capabilityID,
		OldLevel:     oldLevel,
		NewLevel:     newLevel,
		Timestamp:    time.Now().UTC(),
	}
}

func (e CapabilityLevelChanged) EventType() string {
	return "CapabilityLevelChanged"
}

func (e CapabilityLevelChanged) EventData() map[string]interface{} {
	return map[string]interface{}{
		"capabilityId": e.CapabilityID,
		"oldLevel":     e.OldLevel,
		"newLevel":     e.NewLevel,
		"timestamp":    e.Timestamp,
	}
}

func (e CapabilityLevelChanged) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.CapabilityID
}
