package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationFitScoreUpdated struct {
	domain.BaseEvent
	ID           string    `json:"id"`
	Score        int       `json:"score"`
	Rationale    string    `json:"rationale"`
	OldScore     int       `json:"oldScore"`
	OldRationale string    `json:"oldRationale"`
	UpdatedAt    time.Time `json:"updatedAt"`
	UpdatedBy    string    `json:"updatedBy"`
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

func (e ApplicationFitScoreUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
