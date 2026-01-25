package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type StrategyImportanceSet struct {
	domain.BaseEvent
	ID               string    `json:"id"`
	BusinessDomainID string    `json:"businessDomainId"`
	CapabilityID     string    `json:"capabilityId"`
	PillarID         string    `json:"pillarId"`
	PillarName       string    `json:"pillarName"`
	Importance       int       `json:"importance"`
	Rationale        string    `json:"rationale"`
	SetAt            time.Time `json:"setAt"`
}

type StrategyImportanceSetParams struct {
	ID               string
	BusinessDomainID string
	CapabilityID     string
	PillarID         string
	PillarName       string
	Importance       int
	Rationale        string
}

func NewStrategyImportanceSet(params StrategyImportanceSetParams) StrategyImportanceSet {
	return StrategyImportanceSet{
		BaseEvent:        domain.NewBaseEvent(params.ID),
		ID:               params.ID,
		BusinessDomainID: params.BusinessDomainID,
		CapabilityID:     params.CapabilityID,
		PillarID:         params.PillarID,
		PillarName:       params.PillarName,
		Importance:       params.Importance,
		Rationale:        params.Rationale,
		SetAt:            time.Now().UTC(),
	}
}

func (e StrategyImportanceSet) EventType() string {
	return "StrategyImportanceSet"
}

func (e StrategyImportanceSet) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":               e.ID,
		"businessDomainId": e.BusinessDomainID,
		"capabilityId":     e.CapabilityID,
		"pillarId":         e.PillarID,
		"pillarName":       e.PillarName,
		"importance":       e.Importance,
		"rationale":        e.Rationale,
		"setAt":            e.SetAt,
	}
}

func (e StrategyImportanceSet) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
