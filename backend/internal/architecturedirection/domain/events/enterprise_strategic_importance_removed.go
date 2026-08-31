package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseStrategicImportanceRemoved struct {
	domain.BaseEvent
	ID                     string    `json:"id"`
	EnterpriseCapabilityID string    `json:"enterpriseCapabilityId"`
	PillarID               string    `json:"pillarId"`
	RemovedAt              time.Time `json:"removedAt"`
}

func (e EnterpriseStrategicImportanceRemoved) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewEnterpriseStrategicImportanceRemoved(id, enterpriseCapabilityID, pillarID string) EnterpriseStrategicImportanceRemoved {
	return EnterpriseStrategicImportanceRemoved{
		BaseEvent:              domain.NewBaseEvent(id),
		ID:                     id,
		EnterpriseCapabilityID: enterpriseCapabilityID,
		PillarID:               pillarID,
		RemovedAt:              time.Now().UTC(),
	}
}

func (e EnterpriseStrategicImportanceRemoved) EventType() string {
	return "EnterpriseStrategicImportanceRemoved"
}

func (e EnterpriseStrategicImportanceRemoved) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                     e.ID,
		"enterpriseCapabilityId": e.EnterpriseCapabilityID,
		"pillarId":               e.PillarID,
		"removedAt":              e.RemovedAt,
	}
}
