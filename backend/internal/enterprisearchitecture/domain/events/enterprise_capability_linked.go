package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapabilityLinked struct {
	domain.BaseEvent
	ID                     string
	EnterpriseCapabilityID string
	DomainCapabilityID     string
	LinkedBy               string
	LinkedAt               time.Time
}

func NewEnterpriseCapabilityLinked(id, enterpriseCapabilityID, domainCapabilityID, linkedBy string) EnterpriseCapabilityLinked {
	return EnterpriseCapabilityLinked{
		BaseEvent:              domain.NewBaseEvent(id),
		ID:                     id,
		EnterpriseCapabilityID: enterpriseCapabilityID,
		DomainCapabilityID:     domainCapabilityID,
		LinkedBy:               linkedBy,
		LinkedAt:               time.Now().UTC(),
	}
}

func (e EnterpriseCapabilityLinked) EventType() string {
	return "EnterpriseCapabilityLinked"
}

func (e EnterpriseCapabilityLinked) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                     e.ID,
		"enterpriseCapabilityId": e.EnterpriseCapabilityID,
		"domainCapabilityId":     e.DomainCapabilityID,
		"linkedBy":               e.LinkedBy,
		"linkedAt":               e.LinkedAt,
	}
}
