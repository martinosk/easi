package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

// ComponentAddedToView is raised when a component is added to a view
type ComponentAddedToView struct {
	domain.BaseEvent
	ViewID      string    `json:"viewId"`
	ComponentID string    `json:"componentId"`
	X           float64   `json:"x"`
	Y           float64   `json:"y"`
	AddedAt     time.Time `json:"addedAt"`
}

func (e ComponentAddedToView) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ViewID
}

// NewComponentAddedToView creates a new ComponentAddedToView event
func NewComponentAddedToView(viewID, componentID string, x, y float64) ComponentAddedToView {
	return ComponentAddedToView{
		BaseEvent:   domain.NewBaseEvent(viewID),
		ViewID:      viewID,
		ComponentID: componentID,
		X:           x,
		Y:           y,
		AddedAt:     time.Now().UTC(),
	}
}

// EventType returns the event type name
func (e ComponentAddedToView) EventType() string {
	return "ComponentAddedToView"
}

// EventData returns the event data as a map for serialization
func (e ComponentAddedToView) EventData() map[string]interface{} {
	return map[string]interface{}{
		"viewId":      e.ViewID,
		"componentId": e.ComponentID,
		"x":           e.X,
		"y":           e.Y,
		"addedAt":     e.AddedAt,
	}
}
