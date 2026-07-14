package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type JourneyProgressUpdated struct {
	domain.BaseEvent
	ID         string    `json:"id"`
	Progress   int       `json:"progress"`
	UpdatedBy  string    `json:"updatedBy"`
	OccurredOn time.Time `json:"occurredOn"`
}

type JourneyProgressUpdatedFields struct {
	ID        string
	Progress  int
	UpdatedBy string
}

func NewJourneyProgressUpdated(f JourneyProgressUpdatedFields) JourneyProgressUpdated {
	return JourneyProgressUpdated{
		BaseEvent:  domain.NewBaseEvent(f.ID),
		ID:         f.ID,
		Progress:   f.Progress,
		UpdatedBy:  f.UpdatedBy,
		OccurredOn: time.Now().UTC(),
	}
}

func (e JourneyProgressUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e JourneyProgressUpdated) EventType() string { return pl.JourneyProgressUpdated }

func (e JourneyProgressUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":         e.ID,
		"progress":   e.Progress,
		"updatedBy":  e.UpdatedBy,
		"occurredOn": e.OccurredOn,
	}
}
