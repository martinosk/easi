package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type AcquiredViaRelationshipCreated struct {
	domain.BaseEvent
	ID               string
	AcquiredEntityID string
	ComponentID      string
	Notes            string
	CreatedAt        time.Time
}

func NewAcquiredViaRelationshipCreated(id, acquiredEntityID, componentID, notes string) AcquiredViaRelationshipCreated {
	return AcquiredViaRelationshipCreated{
		BaseEvent:        domain.NewBaseEvent(id),
		ID:               id,
		AcquiredEntityID: acquiredEntityID,
		ComponentID:      componentID,
		Notes:            notes,
		CreatedAt:        time.Now().UTC(),
	}
}

func (e AcquiredViaRelationshipCreated) EventType() string {
	return "AcquiredViaRelationshipCreated"
}

func (e AcquiredViaRelationshipCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":               e.ID,
		"acquiredEntityId": e.AcquiredEntityID,
		"componentId":      e.ComponentID,
		"notes":            e.Notes,
		"createdAt":        e.CreatedAt,
	}
}
