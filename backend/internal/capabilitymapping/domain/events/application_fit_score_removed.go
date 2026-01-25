package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationFitScoreRemoved struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	ComponentID string    `json:"componentId"`
	PillarID    string    `json:"pillarId"`
	RemovedAt   time.Time `json:"removedAt"`
	RemovedBy   string    `json:"removedBy"`
}

func NewApplicationFitScoreRemoved(id, componentID, pillarID, removedBy string) ApplicationFitScoreRemoved {
	return ApplicationFitScoreRemoved{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		ComponentID: componentID,
		PillarID:    pillarID,
		RemovedAt:   time.Now().UTC(),
		RemovedBy:   removedBy,
	}
}

func (e ApplicationFitScoreRemoved) EventType() string {
	return "ApplicationFitScoreRemoved"
}

func (e ApplicationFitScoreRemoved) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"componentId": e.ComponentID,
		"pillarId":    e.PillarID,
		"removedAt":   e.RemovedAt,
		"removedBy":   e.RemovedBy,
	}
}

func (e ApplicationFitScoreRemoved) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
