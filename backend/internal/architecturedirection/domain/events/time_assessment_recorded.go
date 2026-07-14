package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type TimeAssessmentRecorded struct {
	domain.BaseEvent
	ID            string    `json:"id"`
	CapabilityID  string    `json:"capabilityId"`
	ComponentID   string    `json:"componentId"`
	RealizationID string    `json:"realizationId"`
	Grade         string    `json:"grade"`
	PreviousGrade string    `json:"previousGrade,omitempty"`
	Rationale     string    `json:"rationale"`
	AssessedBy    string    `json:"assessedBy"`
	OccurredOn    time.Time `json:"occurredOn"`
}

type TimeAssessmentRecordedFields struct {
	ID            string
	CapabilityID  string
	ComponentID   string
	RealizationID string
	Grade         string
	PreviousGrade string
	Rationale     string
	AssessedBy    string
}

func NewTimeAssessmentRecorded(f TimeAssessmentRecordedFields) TimeAssessmentRecorded {
	return TimeAssessmentRecorded{
		BaseEvent:     domain.NewBaseEvent(f.ID),
		ID:            f.ID,
		CapabilityID:  f.CapabilityID,
		ComponentID:   f.ComponentID,
		RealizationID: f.RealizationID,
		Grade:         f.Grade,
		PreviousGrade: f.PreviousGrade,
		Rationale:     f.Rationale,
		AssessedBy:    f.AssessedBy,
		OccurredOn:    time.Now().UTC(),
	}
}

func (e TimeAssessmentRecorded) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e TimeAssessmentRecorded) EventType() string { return pl.TimeAssessmentRecorded }

func (e TimeAssessmentRecorded) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":            e.ID,
		"capabilityId":  e.CapabilityID,
		"componentId":   e.ComponentID,
		"realizationId": e.RealizationID,
		"grade":         e.Grade,
		"previousGrade": e.PreviousGrade,
		"rationale":     e.Rationale,
		"assessedBy":    e.AssessedBy,
		"occurredOn":    e.OccurredOn,
	}
}
