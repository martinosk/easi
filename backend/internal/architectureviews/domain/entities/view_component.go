package entities

import (
	"github.com/easi/backend/internal/architectureviews/domain/valueobjects"
)

// ViewComponent represents a component placed on a view with its position
type ViewComponent struct {
	componentID string
	position    valueobjects.ComponentPosition
}

// NewViewComponent creates a new view component
func NewViewComponent(componentID string, position valueobjects.ComponentPosition) ViewComponent {
	return ViewComponent{
		componentID: componentID,
		position:    position,
	}
}

// ComponentID returns the component ID
func (vc ViewComponent) ComponentID() string {
	return vc.componentID
}

// Position returns the component position
func (vc ViewComponent) Position() valueobjects.ComponentPosition {
	return vc.position
}

// UpdatePosition returns a new ViewComponent with updated position (immutable)
func (vc ViewComponent) UpdatePosition(newPosition valueobjects.ComponentPosition) ViewComponent {
	return ViewComponent{
		componentID: vc.componentID,
		position:    newPosition,
	}
}
