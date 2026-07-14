package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type JourneyMilestoneAdded struct {
	domain.BaseEvent
	ID           string            `json:"id"`
	MilestoneID  string            `json:"milestoneId"`
	Label        string            `json:"label"`
	TargetPeriod *TargetPeriodData `json:"targetPeriod,omitempty"`
	Status       string            `json:"status"`
	AddedBy      string            `json:"addedBy"`
	OccurredOn   time.Time         `json:"occurredOn"`
}

type JourneyMilestoneFields struct {
	ID           string
	MilestoneID  string
	Label        string
	TargetPeriod *TargetPeriodData
	Status       string
	Actor        string
}

func NewJourneyMilestoneAdded(f JourneyMilestoneFields) JourneyMilestoneAdded {
	return JourneyMilestoneAdded{
		BaseEvent:    domain.NewBaseEvent(f.ID),
		ID:           f.ID,
		MilestoneID:  f.MilestoneID,
		Label:        f.Label,
		TargetPeriod: f.TargetPeriod,
		Status:       f.Status,
		AddedBy:      f.Actor,
		OccurredOn:   time.Now().UTC(),
	}
}

func (e JourneyMilestoneAdded) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e JourneyMilestoneAdded) EventType() string { return pl.JourneyMilestoneAdded }

func (e JourneyMilestoneAdded) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":           e.ID,
		"milestoneId":  e.MilestoneID,
		"label":        e.Label,
		"targetPeriod": targetPeriodEventData(e.TargetPeriod),
		"status":       e.Status,
		"addedBy":      e.AddedBy,
		"occurredOn":   e.OccurredOn,
	}
}
