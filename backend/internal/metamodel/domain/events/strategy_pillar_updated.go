package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type StrategyPillarUpdated struct {
	domain.BaseEvent
	ID             string    `json:"id"`
	TenantID       string    `json:"tenantId"`
	Version        int       `json:"version"`
	PillarID       string    `json:"pillarId"`
	NewName        string    `json:"newName"`
	NewDescription string    `json:"newDescription"`
	ModifiedAt     time.Time `json:"modifiedAt"`
	ModifiedBy     string    `json:"modifiedBy"`
}

func NewStrategyPillarUpdated(params UpdatePillarParams) StrategyPillarUpdated {
	return StrategyPillarUpdated{
		BaseEvent:      domain.NewBaseEvent(params.ConfigID),
		ID:             params.ConfigID,
		TenantID:       params.TenantID,
		Version:        params.Version,
		PillarID:       params.PillarID,
		NewName:        params.NewName,
		NewDescription: params.NewDescription,
		ModifiedAt:     time.Now().UTC(),
		ModifiedBy:     params.ModifiedBy,
	}
}

func (e StrategyPillarUpdated) EventType() string {
	return "StrategyPillarUpdated"
}

func (e StrategyPillarUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":             e.ID,
		"tenantId":       e.TenantID,
		"version":        e.Version,
		"pillarId":       e.PillarID,
		"newName":        e.NewName,
		"newDescription": e.NewDescription,
		"modifiedAt":     e.ModifiedAt,
		"modifiedBy":     e.ModifiedBy,
	}
}
