package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type CapabilityUnassignedFromDomain struct {
	domain.BaseEvent
	ID               string    `json:"id"`
	BusinessDomainID string    `json:"businessDomainId"`
	CapabilityID     string    `json:"capabilityId"`
	UnassignedAt     time.Time `json:"unassignedAt"`
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

func (e CapabilityUnassignedFromDomain) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
