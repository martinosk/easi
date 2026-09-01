package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationHostingClassified struct {
	domain.BaseEvent
	ComponentID  string    `json:"componentId"`
	Hosting      string    `json:"hosting"`
	ClassifiedAt time.Time `json:"classifiedAt"`
}

func NewApplicationHostingClassified(componentID, hosting string) ApplicationHostingClassified {
	return ApplicationHostingClassified{
		BaseEvent:    domain.NewBaseEvent(componentID),
		ComponentID:  componentID,
		Hosting:      hosting,
		ClassifiedAt: time.Now().UTC(),
	}
}

func (e ApplicationHostingClassified) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func (e ApplicationHostingClassified) EventType() string {
	return "ApplicationHostingClassified"
}

func (e ApplicationHostingClassified) EventData() map[string]any {
	return map[string]any{
		"componentId":  e.ComponentID,
		"hosting":      e.Hosting,
		"classifiedAt": e.ClassifiedAt,
	}
}
