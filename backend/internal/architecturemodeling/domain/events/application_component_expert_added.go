package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationComponentExpertAdded struct {
	domain.BaseEvent
	ComponentID string
	ExpertName  string
	ExpertRole  string
	ContactInfo string
	AddedAt     time.Time
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
