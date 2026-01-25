package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationComponentExpertRemoved struct {
	domain.BaseEvent
	ComponentID string    `json:"componentId"`
	ExpertName  string    `json:"expertName"`
	ExpertRole  string    `json:"expertRole"`
	ContactInfo string    `json:"contactInfo"`
	RemovedAt   time.Time `json:"removedAt"`
}

func (e ApplicationComponentExpertRemoved) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
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
