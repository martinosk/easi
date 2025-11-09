package events

import (
	"time"

	"github.com/easi/backend/internal/shared/domain"
)

// ComponentPositionUpdated is raised when a component's position is updated on a view
type ComponentPositionUpdated struct {
	domain.BaseEvent
	ViewID      string
	ComponentID string
	X           float64
	Y           float64
	UpdatedAt   time.Time
}

// NewComponentPositionUpdated creates a new ComponentPositionUpdated event
func NewComponentPositionUpdated(viewID, componentID string, x, y float64) ComponentPositionUpdated {
	return ComponentPositionUpdated{
		BaseEvent:   domain.NewBaseEvent(viewID),
		ViewID:      viewID,
		ComponentID: componentID,
		X:           x,
		Y:           y,
		UpdatedAt:   time.Now().UTC(),
	}
}

// EventType returns the event type name
func (e ComponentPositionUpdated) EventType() string {
	return "ComponentPositionUpdated"
}

// EventData returns the event data as a map for serialization
func (e ComponentPositionUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"viewId":      e.ViewID,
		"componentId": e.ComponentID,
		"x":           e.X,
		"y":           e.Y,
		"updatedAt":   e.UpdatedAt,
	}
}
