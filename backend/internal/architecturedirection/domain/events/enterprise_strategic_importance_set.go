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
	ID                     string    `json:"id"`
	EnterpriseCapabilityID string    `json:"enterpriseCapabilityId"`
	PillarID               string    `json:"pillarId"`
	PillarName             string    `json:"pillarName"`
	Importance             int       `json:"importance"`
	Rationale              string    `json:"rationale"`
	SetAt                  time.Time `json:"setAt"`
}

func (e EnterpriseStrategicImportanceSet) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
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
