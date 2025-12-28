package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type ComponentRelationDeleted struct {
	domain.BaseEvent
	ID                string
	SourceComponentID string
	TargetComponentID string
	DeletedAt         time.Time
}

func NewComponentRelationDeleted(id, sourceComponentID, targetComponentID string) ComponentRelationDeleted {
	return ComponentRelationDeleted{
		BaseEvent:         domain.NewBaseEvent(id),
		ID:                id,
		SourceComponentID: sourceComponentID,
		TargetComponentID: targetComponentID,
		DeletedAt:         time.Now().UTC(),
	}
}

func (e ComponentRelationDeleted) EventType() string {
	return "ComponentRelationDeleted"
}

func (e ComponentRelationDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                e.ID,
		"sourceComponentID": e.SourceComponentID,
		"targetComponentID": e.TargetComponentID,
		"deletedAt":         e.DeletedAt,
	}
}
