package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type JourneyMilestoneRemoved struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	MilestoneID string    `json:"milestoneId"`
	RemovedBy   string    `json:"removedBy"`
	OccurredOn  time.Time `json:"occurredOn"`
}

type JourneyMilestoneRemovedFields struct {
	ID          string
	MilestoneID string
	RemovedBy   string
}

func NewJourneyMilestoneRemoved(f JourneyMilestoneRemovedFields) JourneyMilestoneRemoved {
	return JourneyMilestoneRemoved{
		BaseEvent:   domain.NewBaseEvent(f.ID),
		ID:          f.ID,
		MilestoneID: f.MilestoneID,
		RemovedBy:   f.RemovedBy,
		OccurredOn:  time.Now().UTC(),
	}
}

func (e JourneyMilestoneRemoved) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e JourneyMilestoneRemoved) EventType() string { return pl.JourneyMilestoneRemoved }

func (e JourneyMilestoneRemoved) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"milestoneId": e.MilestoneID,
		"removedBy":   e.RemovedBy,
		"occurredOn":  e.OccurredOn,
	}
}
