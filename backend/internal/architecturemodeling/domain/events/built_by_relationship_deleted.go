package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type BuiltByRelationshipDeleted struct {
	domain.BaseEvent
	ID             string
	InternalTeamID string
	ComponentID    string
	DeletedAt      time.Time
}

func NewBuiltByRelationshipDeleted(id, internalTeamID, componentID string) BuiltByRelationshipDeleted {
	return BuiltByRelationshipDeleted{
		BaseEvent:      domain.NewBaseEvent(id),
		ID:             id,
		InternalTeamID: internalTeamID,
		ComponentID:    componentID,
		DeletedAt:      time.Now().UTC(),
	}
}

func (e BuiltByRelationshipDeleted) EventType() string {
	return "BuiltByRelationshipDeleted"
}

func (e BuiltByRelationshipDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":             e.ID,
		"internalTeamId": e.InternalTeamID,
		"componentId":    e.ComponentID,
		"deletedAt":      e.DeletedAt,
	}
}
