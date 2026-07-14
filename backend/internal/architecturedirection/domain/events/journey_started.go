package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type JourneyStarted struct {
	domain.BaseEvent
	ID         string    `json:"id"`
	StartedBy  string    `json:"startedBy"`
	OccurredOn time.Time `json:"occurredOn"`
}

type JourneyStartedFields struct {
	ID        string
	StartedBy string
}

func NewJourneyStarted(f JourneyStartedFields) JourneyStarted {
	return JourneyStarted{
		BaseEvent:  domain.NewBaseEvent(f.ID),
		ID:         f.ID,
		StartedBy:  f.StartedBy,
		OccurredOn: time.Now().UTC(),
	}
}

func (e JourneyStarted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e JourneyStarted) EventType() string { return pl.JourneyStarted }

func (e JourneyStarted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":         e.ID,
		"startedBy":  e.StartedBy,
		"occurredOn": e.OccurredOn,
	}
}
