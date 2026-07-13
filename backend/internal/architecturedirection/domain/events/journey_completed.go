package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type JourneyCompleted struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	CompletedBy string    `json:"completedBy"`
	OccurredOn  time.Time `json:"occurredOn"`
}

type JourneyCompletedFields struct {
	ID          string
	CompletedBy string
}

func NewJourneyCompleted(f JourneyCompletedFields) JourneyCompleted {
	return JourneyCompleted{
		BaseEvent:   domain.NewBaseEvent(f.ID),
		ID:          f.ID,
		CompletedBy: f.CompletedBy,
		OccurredOn:  time.Now().UTC(),
	}
}

func (e JourneyCompleted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e JourneyCompleted) EventType() string { return pl.JourneyCompleted }

func (e JourneyCompleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"completedBy": e.CompletedBy,
		"occurredOn":  e.OccurredOn,
	}
}
