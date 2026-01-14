package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationComponentExpertRemoved struct {
	domain.BaseEvent
	ComponentID string
	ExpertName  string
	ExpertRole  string
	ContactInfo string
	RemovedAt   time.Time
}

func NewApplicationComponentExpertRemoved(componentID, expertName, expertRole, contactInfo string) ApplicationComponentExpertRemoved {
	return ApplicationComponentExpertRemoved{
		BaseEvent:   domain.NewBaseEvent(componentID),
		ComponentID: componentID,
		ExpertName:  expertName,
		ExpertRole:  expertRole,
		ContactInfo: contactInfo,
		RemovedAt:   time.Now().UTC(),
	}
}

func (e ApplicationComponentExpertRemoved) EventType() string {
	return "ApplicationComponentExpertRemoved"
}

func (e ApplicationComponentExpertRemoved) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"expertName":  e.ExpertName,
		"expertRole":  e.ExpertRole,
		"contactInfo": e.ContactInfo,
		"removedAt":   e.RemovedAt,
	}
}
