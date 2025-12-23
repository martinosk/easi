package events

import (
	"time"

	"easi/backend/internal/shared/eventsourcing"
)

// ComponentRemovedFromView is raised when a component is removed from a view
type ComponentRemovedFromView struct {
	domain.BaseEvent
	ViewID      string
	ComponentID string
	RemovedAt   time.Time
}

// NewComponentRemovedFromView creates a new ComponentRemovedFromView event
func NewComponentRemovedFromView(viewID, componentID string) ComponentRemovedFromView {
	return ComponentRemovedFromView{
		BaseEvent:   domain.NewBaseEvent(viewID),
		ViewID:      viewID,
		ComponentID: componentID,
		RemovedAt:   time.Now().UTC(),
	}
}

// EventType returns the event type name
func (e ComponentRemovedFromView) EventType() string {
	return "ComponentRemovedFromView"
}

// EventData returns the event data as a map for serialization
func (e ComponentRemovedFromView) EventData() map[string]interface{} {
	return map[string]interface{}{
		"viewId":      e.ViewID,
		"componentId": e.ComponentID,
		"removedAt":   e.RemovedAt,
	}
}
