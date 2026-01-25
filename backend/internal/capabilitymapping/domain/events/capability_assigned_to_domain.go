package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type CapabilityAssignedToDomain struct {
	domain.BaseEvent
	ID               string    `json:"id"`
	BusinessDomainID string    `json:"businessDomainId"`
	CapabilityID     string    `json:"capabilityId"`
	AssignedAt       time.Time `json:"assignedAt"`
}

func NewCapabilityAssignedToDomain(id, businessDomainID, capabilityID string) CapabilityAssignedToDomain {
	return CapabilityAssignedToDomain{
		BaseEvent:        domain.NewBaseEvent(id),
		ID:               id,
		BusinessDomainID: businessDomainID,
		CapabilityID:     capabilityID,
		AssignedAt:       time.Now().UTC(),
	}
}

func (e CapabilityAssignedToDomain) EventType() string {
	return "CapabilityAssignedToDomain"
}

func (e CapabilityAssignedToDomain) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":               e.ID,
		"businessDomainId": e.BusinessDomainID,
		"capabilityId":     e.CapabilityID,
		"assignedAt":       e.AssignedAt,
	}
}

func (e CapabilityAssignedToDomain) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
