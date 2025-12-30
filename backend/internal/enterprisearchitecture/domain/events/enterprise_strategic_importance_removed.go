package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseStrategicImportanceRemoved struct {
	domain.BaseEvent
	ID                     string
	EnterpriseCapabilityID string
	PillarID               string
	RemovedAt              time.Time
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
