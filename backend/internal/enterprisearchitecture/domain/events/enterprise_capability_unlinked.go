package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapabilityUnlinked struct {
	domain.BaseEvent
	ID                     string    `json:"id"`
	EnterpriseCapabilityID string    `json:"enterpriseCapabilityId"`
	DomainCapabilityID     string    `json:"domainCapabilityId"`
	UnlinkedAt             time.Time `json:"unlinkedAt"`
}

func (e EnterpriseCapabilityUnlinked) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewEnterpriseCapabilityUnlinked(id, enterpriseCapabilityID, domainCapabilityID string) EnterpriseCapabilityUnlinked {
	return EnterpriseCapabilityUnlinked{
		BaseEvent:              domain.NewBaseEvent(id),
		ID:                     id,
		EnterpriseCapabilityID: enterpriseCapabilityID,
		DomainCapabilityID:     domainCapabilityID,
		UnlinkedAt:             time.Now().UTC(),
	}
}

func (e EnterpriseCapabilityUnlinked) EventType() string {
	return "EnterpriseCapabilityUnlinked"
}

func (e EnterpriseCapabilityUnlinked) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                     e.ID,
		"enterpriseCapabilityId": e.EnterpriseCapabilityID,
		"domainCapabilityId":     e.DomainCapabilityID,
		"unlinkedAt":             e.UnlinkedAt,
	}
}
