package events

import (
	"time"

	"easi/backend/internal/shared/domain"
)

type CapabilityAssignedToDomain struct {
	domain.BaseEvent
	ID               string
	BusinessDomainID string
	CapabilityID     string
	AssignedAt       time.Time
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
