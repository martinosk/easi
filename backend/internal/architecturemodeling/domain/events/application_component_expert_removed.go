package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationComponentExpertRemoved struct {
	domain.BaseEvent
	ComponentID string
	ExpertName  string
	RemovedAt   time.Time
}

func NewApplicationComponentExpertRemoved(componentID, expertName string) ApplicationComponentExpertRemoved {
	return ApplicationComponentExpertRemoved{
		BaseEvent:   domain.NewBaseEvent(componentID),
		ComponentID: componentID,
		ExpertName:  expertName,
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
		"removedAt":   e.RemovedAt,
	}
}
