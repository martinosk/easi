package events

import (
	"time"

	"github.com/easi/backend/internal/shared/domain"
)

// ComponentRelationCreated is raised when a new component relation is created
type ComponentRelationCreated struct {
	domain.BaseEvent
	ID                string
	SourceComponentID string
	TargetComponentID string
	RelationType      string
	Name              string
	Description       string
	CreatedAt         time.Time
}

// NewComponentRelationCreated creates a new ComponentRelationCreated event
func NewComponentRelationCreated(id, sourceID, targetID, relationType, name, description string) ComponentRelationCreated {
	return ComponentRelationCreated{
		BaseEvent:         domain.NewBaseEvent(id),
		ID:                id,
		SourceComponentID: sourceID,
		TargetComponentID: targetID,
		RelationType:      relationType,
		Name:              name,
		Description:       description,
		CreatedAt:         time.Now().UTC(),
	}
}

// EventType returns the event type name
func (e ComponentRelationCreated) EventType() string {
	return "ComponentRelationCreated"
}

// EventData returns the event data as a map for serialization
func (e ComponentRelationCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                e.ID,
		"sourceComponentId": e.SourceComponentID,
		"targetComponentId": e.TargetComponentID,
		"relationType":      e.RelationType,
		"name":              e.Name,
		"description":       e.Description,
		"createdAt":         e.CreatedAt,
	}
}
