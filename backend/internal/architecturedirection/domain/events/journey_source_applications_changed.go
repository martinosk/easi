package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type JourneySourceApplicationsChanged struct {
	domain.BaseEvent
	ID               string    `json:"id"`
	FromComponentIDs []string  `json:"fromComponentIds"`
	ChangedBy        string    `json:"changedBy"`
	OccurredOn       time.Time `json:"occurredOn"`
}

type JourneySourceApplicationsChangedFields struct {
	ID               string
	FromComponentIDs []string
	ChangedBy        string
}

func NewJourneySourceApplicationsChanged(f JourneySourceApplicationsChangedFields) JourneySourceApplicationsChanged {
	return JourneySourceApplicationsChanged{
		BaseEvent:        domain.NewBaseEvent(f.ID),
		ID:               f.ID,
		FromComponentIDs: f.FromComponentIDs,
		ChangedBy:        f.ChangedBy,
		OccurredOn:       time.Now().UTC(),
	}
}

func (e JourneySourceApplicationsChanged) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e JourneySourceApplicationsChanged) EventType() string {
	return pl.JourneySourceApplicationsChanged
}

func (e JourneySourceApplicationsChanged) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":               e.ID,
		"fromComponentIds": e.FromComponentIDs,
		"changedBy":        e.ChangedBy,
		"occurredOn":       e.OccurredOn,
	}
}
