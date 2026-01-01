package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationFitScoreUpdated struct {
	domain.BaseEvent
	ID           string
	Score        int
	Rationale    string
	OldScore     int
	OldRationale string
	UpdatedAt    time.Time
	UpdatedBy    string
}

type ApplicationFitScoreUpdatedParams struct {
	ID           string
	Score        int
	Rationale    string
	OldScore     int
	OldRationale string
	UpdatedBy    string
}

func NewApplicationFitScoreUpdated(params ApplicationFitScoreUpdatedParams) ApplicationFitScoreUpdated {
	return ApplicationFitScoreUpdated{
		BaseEvent:    domain.NewBaseEvent(params.ID),
		ID:           params.ID,
		Score:        params.Score,
		Rationale:    params.Rationale,
		OldScore:     params.OldScore,
		OldRationale: params.OldRationale,
		UpdatedAt:    time.Now().UTC(),
		UpdatedBy:    params.UpdatedBy,
	}
}

func (e ApplicationFitScoreUpdated) EventType() string {
	return "ApplicationFitScoreUpdated"
}

func (e ApplicationFitScoreUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":           e.ID,
		"score":        e.Score,
		"rationale":    e.Rationale,
		"oldScore":     e.OldScore,
		"oldRationale": e.OldRationale,
		"updatedAt":    e.UpdatedAt,
		"updatedBy":    e.UpdatedBy,
	}
}
