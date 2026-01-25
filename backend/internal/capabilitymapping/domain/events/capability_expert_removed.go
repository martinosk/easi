package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type CapabilityExpertRemoved struct {
	domain.BaseEvent
	CapabilityID string    `json:"capabilityId"`
	ExpertName   string    `json:"expertName"`
	ExpertRole   string    `json:"expertRole"`
	ContactInfo  string    `json:"contactInfo"`
	RemovedAt    time.Time `json:"removedAt"`
}

func NewCapabilityExpertRemoved(capabilityID, expertName, expertRole, contactInfo string) CapabilityExpertRemoved {
	return CapabilityExpertRemoved{
		BaseEvent:    domain.NewBaseEvent(capabilityID),
		CapabilityID: capabilityID,
		ExpertName:   expertName,
		ExpertRole:   expertRole,
		ContactInfo:  contactInfo,
		RemovedAt:    time.Now().UTC(),
	}
}

func (e CapabilityExpertRemoved) EventType() string {
	return "CapabilityExpertRemoved"
}

func (e CapabilityExpertRemoved) EventData() map[string]interface{} {
	return map[string]interface{}{
		"capabilityId": e.CapabilityID,
		"expertName":   e.ExpertName,
		"expertRole":   e.ExpertRole,
		"contactInfo":  e.ContactInfo,
		"removedAt":    e.RemovedAt,
	}
}

func (e CapabilityExpertRemoved) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.CapabilityID
}
