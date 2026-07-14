package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type JourneyAbandoned struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	AbandonedBy string    `json:"abandonedBy"`
	OccurredOn  time.Time `json:"occurredOn"`
}

type JourneyAbandonedFields struct {
	ID          string
	AbandonedBy string
}

func NewJourneyAbandoned(f JourneyAbandonedFields) JourneyAbandoned {
	return JourneyAbandoned{
		BaseEvent:   domain.NewBaseEvent(f.ID),
		ID:          f.ID,
		AbandonedBy: f.AbandonedBy,
		OccurredOn:  time.Now().UTC(),
	}
}

func (e JourneyAbandoned) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e JourneyAbandoned) EventType() string { return pl.JourneyAbandoned }

func (e JourneyAbandoned) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"abandonedBy": e.AbandonedBy,
		"occurredOn":  e.OccurredOn,
	}
}
