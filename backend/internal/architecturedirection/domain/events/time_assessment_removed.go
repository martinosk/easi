package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type TimeAssessmentRemoved struct {
	domain.BaseEvent
	ID           string    `json:"id"`
	CapabilityID string    `json:"capabilityId"`
	ComponentID  string    `json:"componentId"`
	RemovedBy    string    `json:"removedBy"`
	OccurredOn   time.Time `json:"occurredOn"`
}

type TimeAssessmentRemovedFields struct {
	ID           string
	CapabilityID string
	ComponentID  string
	RemovedBy    string
}

func NewTimeAssessmentRemoved(f TimeAssessmentRemovedFields) TimeAssessmentRemoved {
	return TimeAssessmentRemoved{
		BaseEvent:    domain.NewBaseEvent(f.ID),
		ID:           f.ID,
		CapabilityID: f.CapabilityID,
		ComponentID:  f.ComponentID,
		RemovedBy:    f.RemovedBy,
		OccurredOn:   time.Now().UTC(),
	}
}

func (e TimeAssessmentRemoved) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e TimeAssessmentRemoved) EventType() string { return pl.TimeAssessmentRemoved }

func (e TimeAssessmentRemoved) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":           e.ID,
		"capabilityId": e.CapabilityID,
		"componentId":  e.ComponentID,
		"removedBy":    e.RemovedBy,
		"occurredOn":   e.OccurredOn,
	}
}
