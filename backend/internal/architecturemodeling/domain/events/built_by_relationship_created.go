package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type BuiltByRelationshipCreated struct {
	domain.BaseEvent
	ID             string
	InternalTeamID string
	ComponentID    string
	Notes          string
	CreatedAt      time.Time
}

func NewBuiltByRelationshipCreated(id, internalTeamID, componentID, notes string) BuiltByRelationshipCreated {
	return BuiltByRelationshipCreated{
		BaseEvent:      domain.NewBaseEvent(id),
		ID:             id,
		InternalTeamID: internalTeamID,
		ComponentID:    componentID,
		Notes:          notes,
		CreatedAt:      time.Now().UTC(),
	}
}

func (e BuiltByRelationshipCreated) EventType() string {
	return "BuiltByRelationshipCreated"
}

func (e BuiltByRelationshipCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":             e.ID,
		"internalTeamId": e.InternalTeamID,
		"componentId":    e.ComponentID,
		"notes":          e.Notes,
		"createdAt":      e.CreatedAt,
	}
}
