package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type AcquiredViaRelationshipDeleted struct {
	domain.BaseEvent
	ID               string
	AcquiredEntityID string
	ComponentID      string
	DeletedAt        time.Time
}

func NewAcquiredViaRelationshipDeleted(id, acquiredEntityID, componentID string) AcquiredViaRelationshipDeleted {
	return AcquiredViaRelationshipDeleted{
		BaseEvent:        domain.NewBaseEvent(id),
		ID:               id,
		AcquiredEntityID: acquiredEntityID,
		ComponentID:      componentID,
		DeletedAt:        time.Now().UTC(),
	}
}

func (e AcquiredViaRelationshipDeleted) EventType() string {
	return "AcquiredViaRelationshipDeleted"
}

func (e AcquiredViaRelationshipDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":               e.ID,
		"acquiredEntityId": e.AcquiredEntityID,
		"componentId":      e.ComponentID,
		"deletedAt":        e.DeletedAt,
	}
}
