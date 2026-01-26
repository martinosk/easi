package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type PillarFitConfigurationUpdated struct {
	domain.BaseEvent
	ID                string    `json:"id"`
	TenantID          string    `json:"tenantId"`
	Version           int       `json:"version"`
	PillarID          string    `json:"pillarId"`
	FitScoringEnabled bool      `json:"fitScoringEnabled"`
	FitCriteria       string    `json:"fitCriteria"`
	FitType           string    `json:"fitType"`
	ModifiedAt        time.Time `json:"modifiedAt"`
	ModifiedBy        string    `json:"modifiedBy"`
}

func (e PillarFitConfigurationUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewPillarFitConfigurationUpdated(params UpdatePillarFitConfigParams) PillarFitConfigurationUpdated {
	return PillarFitConfigurationUpdated{
		BaseEvent:         domain.NewBaseEvent(params.ConfigID),
		ID:                params.ConfigID,
		TenantID:          params.TenantID,
		Version:           params.Version,
		PillarID:          params.PillarID,
		FitScoringEnabled: params.FitScoringEnabled,
		FitCriteria:       params.FitCriteria,
		FitType:           params.FitType,
		ModifiedAt:        time.Now().UTC(),
		ModifiedBy:        params.ModifiedBy,
	}
}

func (e PillarFitConfigurationUpdated) EventType() string {
	return "PillarFitConfigurationUpdated"
}

func (e PillarFitConfigurationUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                e.ID,
		"tenantId":          e.TenantID,
		"version":           e.Version,
		"pillarId":          e.PillarID,
		"fitScoringEnabled": e.FitScoringEnabled,
		"fitCriteria":       e.FitCriteria,
		"fitType":           e.FitType,
		"modifiedAt":        e.ModifiedAt,
		"modifiedBy":        e.ModifiedBy,
	}
}
