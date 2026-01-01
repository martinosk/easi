package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationFitScoreSet struct {
	domain.BaseEvent
	ID          string
	ComponentID string
	PillarID    string
	PillarName  string
	Score       int
	Rationale   string
	ScoredAt    time.Time
	ScoredBy    string
}

type ApplicationFitScoreSetParams struct {
	ID          string
	ComponentID string
	PillarID    string
	PillarName  string
	Score       int
	Rationale   string
	ScoredBy    string
}

func NewApplicationFitScoreSet(params ApplicationFitScoreSetParams) ApplicationFitScoreSet {
	return ApplicationFitScoreSet{
		BaseEvent:   domain.NewBaseEvent(params.ID),
		ID:          params.ID,
		ComponentID: params.ComponentID,
		PillarID:    params.PillarID,
		PillarName:  params.PillarName,
		Score:       params.Score,
		Rationale:   params.Rationale,
		ScoredAt:    time.Now().UTC(),
		ScoredBy:    params.ScoredBy,
	}
}

func (e ApplicationFitScoreSet) EventType() string {
	return "ApplicationFitScoreSet"
}

func (e ApplicationFitScoreSet) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"componentId": e.ComponentID,
		"pillarId":    e.PillarID,
		"pillarName":  e.PillarName,
		"score":       e.Score,
		"rationale":   e.Rationale,
		"scoredAt":    e.ScoredAt,
		"scoredBy":    e.ScoredBy,
	}
}
