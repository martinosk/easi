package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type StrategyImportanceRemoved struct {
	domain.BaseEvent
	ID               string
	BusinessDomainID string
	CapabilityID     string
	PillarID         string
	RemovedAt        time.Time
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
