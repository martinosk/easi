package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type CapabilityParentChanged struct {
	domain.BaseEvent
	CapabilityID string
	OldParentID  string
	NewParentID  string
	OldLevel     string
	NewLevel     string
	Timestamp    time.Time
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
