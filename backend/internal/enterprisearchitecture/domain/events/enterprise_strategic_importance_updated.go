package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseStrategicImportanceUpdatedParams struct {
	ID            string
	Importance    int
	Rationale     string
	OldImportance int
	OldRationale  string
}

type EnterpriseStrategicImportanceUpdated struct {
	domain.BaseEvent
	ID            string
	Importance    int
	Rationale     string
	OldImportance int
	OldRationale  string
	UpdatedAt     time.Time
}

func NewEnterpriseStrategicImportanceUpdated(params EnterpriseStrategicImportanceUpdatedParams) EnterpriseStrategicImportanceUpdated {
	return EnterpriseStrategicImportanceUpdated{
		BaseEvent:     domain.NewBaseEvent(params.ID),
		ID:            params.ID,
		Importance:    params.Importance,
		Rationale:     params.Rationale,
		OldImportance: params.OldImportance,
		OldRationale:  params.OldRationale,
		UpdatedAt:     time.Now().UTC(),
	}
}

func (e EnterpriseStrategicImportanceUpdated) EventType() string {
	return "EnterpriseStrategicImportanceUpdated"
}

func (e EnterpriseStrategicImportanceUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":            e.ID,
		"importance":    e.Importance,
		"rationale":     e.Rationale,
		"oldImportance": e.OldImportance,
		"oldRationale":  e.OldRationale,
		"updatedAt":     e.UpdatedAt,
	}
}
