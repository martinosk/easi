package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type CapabilityExpertAdded struct {
	domain.BaseEvent
	CapabilityID string    `json:"capabilityId"`
	ExpertName   string    `json:"expertName"`
	ExpertRole   string    `json:"expertRole"`
	ContactInfo  string    `json:"contactInfo"`
	AddedAt      time.Time `json:"addedAt"`
}

func NewCapabilityExpertAdded(capabilityID, expertName, expertRole, contactInfo string) CapabilityExpertAdded {
	return CapabilityExpertAdded{
		BaseEvent:    domain.NewBaseEvent(capabilityID),
		CapabilityID: capabilityID,
		ExpertName:   expertName,
		ExpertRole:   expertRole,
		ContactInfo:  contactInfo,
		AddedAt:      time.Now().UTC(),
	}
}

func (e CapabilityExpertAdded) EventType() string {
	return "CapabilityExpertAdded"
}

func (e CapabilityExpertAdded) EventData() map[string]interface{} {
	return map[string]interface{}{
		"capabilityId": e.CapabilityID,
		"expertName":   e.ExpertName,
		"expertRole":   e.ExpertRole,
		"contactInfo":  e.ContactInfo,
		"addedAt":      e.AddedAt,
	}
}

func (e CapabilityExpertAdded) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.CapabilityID
}
