package events

import (
	"time"

	"easi/backend/internal/shared/domain"
)

type CapabilityExpertAdded struct {
	domain.BaseEvent
	CapabilityID string
	ExpertName   string
	ExpertRole   string
	ContactInfo  string
	AddedAt      time.Time
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
