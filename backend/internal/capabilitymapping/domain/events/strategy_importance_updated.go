package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type StrategyImportanceUpdated struct {
	domain.BaseEvent
	ID            string
	Importance    int
	Rationale     string
	OldImportance int
	OldRationale  string
	UpdatedAt     time.Time
}

type StrategyImportanceUpdatedParams struct {
	ID            string
	Importance    int
	Rationale     string
	OldImportance int
	OldRationale  string
}

func NewStrategyImportanceUpdated(params StrategyImportanceUpdatedParams) StrategyImportanceUpdated {
	return StrategyImportanceUpdated{
		BaseEvent:     domain.NewBaseEvent(params.ID),
		ID:            params.ID,
		Importance:    params.Importance,
		Rationale:     params.Rationale,
		OldImportance: params.OldImportance,
		OldRationale:  params.OldRationale,
		UpdatedAt:     time.Now().UTC(),
	}
}

func (e StrategyImportanceUpdated) EventType() string {
	return "StrategyImportanceUpdated"
}

func (e StrategyImportanceUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":            e.ID,
		"importance":    e.Importance,
		"rationale":     e.Rationale,
		"oldImportance": e.OldImportance,
		"oldRationale":  e.OldRationale,
		"updatedAt":     e.UpdatedAt,
	}
}
