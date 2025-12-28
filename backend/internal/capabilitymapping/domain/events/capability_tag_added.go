package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type CapabilityTagAdded struct {
	domain.BaseEvent
	CapabilityID string
	Tag          string
	AddedAt      time.Time
}

func NewCapabilityTagAdded(capabilityID, tag string) CapabilityTagAdded {
	return CapabilityTagAdded{
		BaseEvent:    domain.NewBaseEvent(capabilityID),
		CapabilityID: capabilityID,
		Tag:          tag,
		AddedAt:      time.Now().UTC(),
	}
}

func (e CapabilityTagAdded) EventType() string {
	return "CapabilityTagAdded"
}

func (e CapabilityTagAdded) EventData() map[string]interface{} {
	return map[string]interface{}{
		"capabilityId": e.CapabilityID,
		"tag":          e.Tag,
		"addedAt":      e.AddedAt,
	}
}
