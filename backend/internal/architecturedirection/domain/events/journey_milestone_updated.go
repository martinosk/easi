package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type JourneyMilestoneUpdated struct {
	domain.BaseEvent
	ID           string            `json:"id"`
	MilestoneID  string            `json:"milestoneId"`
	Label        string            `json:"label"`
	TargetPeriod *TargetPeriodData `json:"targetPeriod,omitempty"`
	Status       string            `json:"status"`
	UpdatedBy    string            `json:"updatedBy"`
	OccurredOn   time.Time         `json:"occurredOn"`
}

func NewJourneyMilestoneUpdated(f JourneyMilestoneFields) JourneyMilestoneUpdated {
	return JourneyMilestoneUpdated{
		BaseEvent:    domain.NewBaseEvent(f.ID),
		ID:           f.ID,
		MilestoneID:  f.MilestoneID,
		Label:        f.Label,
		TargetPeriod: f.TargetPeriod,
		Status:       f.Status,
		UpdatedBy:    f.Actor,
		OccurredOn:   time.Now().UTC(),
	}
}

func (e JourneyMilestoneUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e JourneyMilestoneUpdated) EventType() string { return pl.JourneyMilestoneUpdated }

func (e JourneyMilestoneUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":           e.ID,
		"milestoneId":  e.MilestoneID,
		"label":        e.Label,
		"targetPeriod": targetPeriodEventData(e.TargetPeriod),
		"status":       e.Status,
		"updatedBy":    e.UpdatedBy,
		"occurredOn":   e.OccurredOn,
	}
}
