package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type StrategyPillarRemoved struct {
	domain.BaseEvent
	ID         string    `json:"id"`
	TenantID   string    `json:"tenantId"`
	Version    int       `json:"version"`
	PillarID   string    `json:"pillarId"`
	ModifiedAt time.Time `json:"modifiedAt"`
	ModifiedBy string    `json:"modifiedBy"`
}

func (e StrategyPillarRemoved) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewStrategyPillarRemoved(params PillarEventParams) StrategyPillarRemoved {
	return StrategyPillarRemoved{
		BaseEvent:  domain.NewBaseEvent(params.ConfigID),
		ID:         params.ConfigID,
		TenantID:   params.TenantID,
		Version:    params.Version,
		PillarID:   params.PillarID,
		ModifiedAt: time.Now().UTC(),
		ModifiedBy: params.ModifiedBy,
	}
}

func (e StrategyPillarRemoved) EventType() string {
	return "StrategyPillarRemoved"
}

func (e StrategyPillarRemoved) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":         e.ID,
		"tenantId":   e.TenantID,
		"version":    e.Version,
		"pillarId":   e.PillarID,
		"modifiedAt": e.ModifiedAt,
		"modifiedBy": e.ModifiedBy,
	}
}
