package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type StrategyPillarAdded struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	Version     int       `json:"version"`
	PillarID    string    `json:"pillarId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	ModifiedBy  string    `json:"modifiedBy"`
}

func (e StrategyPillarAdded) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewStrategyPillarAdded(params AddPillarParams) StrategyPillarAdded {
	return StrategyPillarAdded{
		BaseEvent:   domain.NewBaseEvent(params.ConfigID),
		ID:          params.ConfigID,
		TenantID:    params.TenantID,
		Version:     params.Version,
		PillarID:    params.PillarID,
		Name:        params.Name,
		Description: params.Description,
		ModifiedAt:  time.Now().UTC(),
		ModifiedBy:  params.ModifiedBy,
	}
}

func (e StrategyPillarAdded) EventType() string {
	return "StrategyPillarAdded"
}

func (e StrategyPillarAdded) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"tenantId":    e.TenantID,
		"version":     e.Version,
		"pillarId":    e.PillarID,
		"name":        e.Name,
		"description": e.Description,
		"modifiedAt":  e.ModifiedAt,
		"modifiedBy":  e.ModifiedBy,
	}
}
