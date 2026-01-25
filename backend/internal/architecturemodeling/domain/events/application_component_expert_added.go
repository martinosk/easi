package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationComponentExpertAdded struct {
	domain.BaseEvent
	ComponentID string    `json:"componentId"`
	ExpertName  string    `json:"expertName"`
	ExpertRole  string    `json:"expertRole"`
	ContactInfo string    `json:"contactInfo"`
	AddedAt     time.Time `json:"addedAt"`
}

func (e ApplicationComponentExpertAdded) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewApplicationComponentExpertAdded(componentID, expertName, expertRole, contactInfo string) ApplicationComponentExpertAdded {
	return ApplicationComponentExpertAdded{
		BaseEvent:   domain.NewBaseEvent(componentID),
		ComponentID: componentID,
		ExpertName:  expertName,
		ExpertRole:  expertRole,
		ContactInfo: contactInfo,
		AddedAt:     time.Now().UTC(),
	}
}

func (e ApplicationComponentExpertAdded) EventType() string {
	return "ApplicationComponentExpertAdded"
}

func (e ApplicationComponentExpertAdded) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"expertName":  e.ExpertName,
		"expertRole":  e.ExpertRole,
		"contactInfo": e.ContactInfo,
		"addedAt":     e.AddedAt,
	}
}
