package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type JourneyMilestonesReordered struct {
	domain.BaseEvent
	ID           string    `json:"id"`
	MilestoneIDs []string  `json:"milestoneIds"`
	ReorderedBy  string    `json:"reorderedBy"`
	OccurredOn   time.Time `json:"occurredOn"`
}

type JourneyMilestonesReorderedFields struct {
	ID           string
	MilestoneIDs []string
	ReorderedBy  string
}

func NewJourneyMilestonesReordered(f JourneyMilestonesReorderedFields) JourneyMilestonesReordered {
	return JourneyMilestonesReordered{
		BaseEvent:    domain.NewBaseEvent(f.ID),
		ID:           f.ID,
		MilestoneIDs: f.MilestoneIDs,
		ReorderedBy:  f.ReorderedBy,
		OccurredOn:   time.Now().UTC(),
	}
}

func (e JourneyMilestonesReordered) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e JourneyMilestonesReordered) EventType() string { return pl.JourneyMilestonesReordered }

func (e JourneyMilestonesReordered) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":           e.ID,
		"milestoneIds": e.MilestoneIDs,
		"reorderedBy":  e.ReorderedBy,
		"occurredOn":   e.OccurredOn,
	}
}
