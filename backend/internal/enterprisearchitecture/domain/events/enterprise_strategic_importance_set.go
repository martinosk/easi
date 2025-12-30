package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseStrategicImportanceSetParams struct {
	ID                     string
	EnterpriseCapabilityID string
	PillarID               string
	PillarName             string
	Importance             int
	Rationale              string
}

type EnterpriseStrategicImportanceSet struct {
	domain.BaseEvent
	ID                     string
	EnterpriseCapabilityID string
	PillarID               string
	PillarName             string
	Importance             int
	Rationale              string
	SetAt                  time.Time
}

func NewEnterpriseStrategicImportanceSet(params EnterpriseStrategicImportanceSetParams) EnterpriseStrategicImportanceSet {
	return EnterpriseStrategicImportanceSet{
		BaseEvent:              domain.NewBaseEvent(params.ID),
		ID:                     params.ID,
		EnterpriseCapabilityID: params.EnterpriseCapabilityID,
		PillarID:               params.PillarID,
		PillarName:             params.PillarName,
		Importance:             params.Importance,
		Rationale:              params.Rationale,
		SetAt:                  time.Now().UTC(),
	}
}

func (e EnterpriseStrategicImportanceSet) EventType() string {
	return "EnterpriseStrategicImportanceSet"
}

func (e EnterpriseStrategicImportanceSet) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                     e.ID,
		"enterpriseCapabilityId": e.EnterpriseCapabilityID,
		"pillarId":               e.PillarID,
		"pillarName":             e.PillarName,
		"importance":             e.Importance,
		"rationale":              e.Rationale,
		"setAt":                  e.SetAt,
	}
}
