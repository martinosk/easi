package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type StrategyImportanceRemoved struct {
	domain.BaseEvent
	ID               string    `json:"id"`
	BusinessDomainID string    `json:"businessDomainId"`
	CapabilityID     string    `json:"capabilityId"`
	PillarID         string    `json:"pillarId"`
	RemovedAt        time.Time `json:"removedAt"`
}

func NewStrategyImportanceRemoved(id, businessDomainID, capabilityID, pillarID string) StrategyImportanceRemoved {
	return StrategyImportanceRemoved{
		BaseEvent:        domain.NewBaseEvent(id),
		ID:               id,
		BusinessDomainID: businessDomainID,
		CapabilityID:     capabilityID,
		PillarID:         pillarID,
		RemovedAt:        time.Now().UTC(),
	}
}

func (e StrategyImportanceRemoved) EventType() string {
	return "StrategyImportanceRemoved"
}

func (e StrategyImportanceRemoved) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":               e.ID,
		"businessDomainId": e.BusinessDomainID,
		"capabilityId":     e.CapabilityID,
		"pillarId":         e.PillarID,
		"removedAt":        e.RemovedAt,
	}
}

func (e StrategyImportanceRemoved) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
