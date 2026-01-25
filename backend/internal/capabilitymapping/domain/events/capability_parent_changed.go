package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type CapabilityParentChanged struct {
	domain.BaseEvent
	CapabilityID string    `json:"capabilityId"`
	OldParentID  string    `json:"oldParentId"`
	NewParentID  string    `json:"newParentId"`
	OldLevel     string    `json:"oldLevel"`
	NewLevel     string    `json:"newLevel"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewCapabilityParentChanged(capabilityID, oldParentID, newParentID, oldLevel, newLevel string) CapabilityParentChanged {
	return CapabilityParentChanged{
		BaseEvent:    domain.NewBaseEvent(capabilityID),
		CapabilityID: capabilityID,
		OldParentID:  oldParentID,
		NewParentID:  newParentID,
		OldLevel:     oldLevel,
		NewLevel:     newLevel,
		Timestamp:    time.Now().UTC(),
	}
}

func (e CapabilityParentChanged) EventType() string {
	return "CapabilityParentChanged"
}

func (e CapabilityParentChanged) EventData() map[string]interface{} {
	return map[string]interface{}{
		"capabilityId": e.CapabilityID,
		"oldParentId":  e.OldParentID,
		"newParentId":  e.NewParentID,
		"oldLevel":     e.OldLevel,
		"newLevel":     e.NewLevel,
		"timestamp":    e.Timestamp,
	}
}

func (e CapabilityParentChanged) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.CapabilityID
}
