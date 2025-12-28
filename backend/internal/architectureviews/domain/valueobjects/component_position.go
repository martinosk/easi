package valueobjects

import domain "easi/backend/internal/shared/eventsourcing"

// ComponentPosition represents the X, Y coordinates of a component on a view
type ComponentPosition struct {
	x float64
	y float64
}

// NewComponentPosition creates a new component position
func NewComponentPosition(x, y float64) ComponentPosition {
	return ComponentPosition{x: x, y: y}
}

// X returns the X coordinate
func (p ComponentPosition) X() float64 {
	return p.x
}

// Y returns the Y coordinate
func (p ComponentPosition) Y() float64 {
	return p.y
}

// Equals checks if two positions are equal
func (p ComponentPosition) Equals(other domain.ValueObject) bool {
	if otherPos, ok := other.(ComponentPosition); ok {
		return p.x == otherPos.x && p.y == otherPos.y
	}
	return false
}
