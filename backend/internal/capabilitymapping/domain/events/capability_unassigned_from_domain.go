package events

import (
	"time"

	"easi/backend/internal/shared/domain"
)

type CapabilityUnassignedFromDomain struct {
	domain.BaseEvent
	ID               string
	BusinessDomainID string
	CapabilityID     string
	UnassignedAt     time.Time
}

func NewCapabilityUnassignedFromDomain(id, businessDomainID, capabilityID string) CapabilityUnassignedFromDomain {
	return CapabilityUnassignedFromDomain{
		BaseEvent:        domain.NewBaseEvent(id),
		ID:               id,
		BusinessDomainID: businessDomainID,
		CapabilityID:     capabilityID,
		UnassignedAt:     time.Now().UTC(),
	}
}

func (e CapabilityUnassignedFromDomain) EventType() string {
	return "CapabilityUnassignedFromDomain"
}

func (e CapabilityUnassignedFromDomain) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":               e.ID,
		"businessDomainId": e.BusinessDomainID,
		"capabilityId":     e.CapabilityID,
		"unassignedAt":     e.UnassignedAt,
	}
}
