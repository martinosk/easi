package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type JourneyDetailsUpdated struct {
	domain.BaseEvent
	ID            string            `json:"id"`
	Note          string            `json:"note"`
	TargetPeriod  *TargetPeriodData `json:"targetPeriod,omitempty"`
	ResultingName string            `json:"resultingName,omitempty"`
	UpdatedBy     string            `json:"updatedBy"`
	OccurredOn    time.Time         `json:"occurredOn"`
}

type JourneyDetailsUpdatedFields struct {
	ID            string
	Note          string
	TargetPeriod  *TargetPeriodData
	ResultingName string
	UpdatedBy     string
}

func NewJourneyDetailsUpdated(f JourneyDetailsUpdatedFields) JourneyDetailsUpdated {
	return JourneyDetailsUpdated{
		BaseEvent:     domain.NewBaseEvent(f.ID),
		ID:            f.ID,
		Note:          f.Note,
		TargetPeriod:  f.TargetPeriod,
		ResultingName: f.ResultingName,
		UpdatedBy:     f.UpdatedBy,
		OccurredOn:    time.Now().UTC(),
	}
}

func (e JourneyDetailsUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e JourneyDetailsUpdated) EventType() string { return pl.JourneyDetailsUpdated }

func (e JourneyDetailsUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":            e.ID,
		"note":          e.Note,
		"targetPeriod":  targetPeriodEventData(e.TargetPeriod),
		"resultingName": e.ResultingName,
		"updatedBy":     e.UpdatedBy,
		"occurredOn":    e.OccurredOn,
	}
}
