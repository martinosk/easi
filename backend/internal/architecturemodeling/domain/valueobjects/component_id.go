package valueobjects

import (
	"easi/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// ComponentID represents a unique identifier for an application component
type ComponentID struct {
	value string
}

// NewComponentID creates a new component ID
func NewComponentID() ComponentID {
	return ComponentID{value: uuid.New().String()}
}

// NewComponentIDFromString creates a component ID from a string
func NewComponentIDFromString(value string) (ComponentID, error) {
	if value == "" {
		return ComponentID{}, domain.ErrEmptyValue
	}

	// Validate UUID format
	if _, err := uuid.Parse(value); err != nil {
		return ComponentID{}, domain.ErrInvalidValue
	}

	return ComponentID{value: value}, nil
}

// Value returns the string value of the ID
func (c ComponentID) Value() string {
	return c.value
}

// Equals checks if two component IDs are equal
func (c ComponentID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(ComponentID); ok {
		return c.value == otherID.value
	}
	return false
}

// String implements the Stringer interface
func (c ComponentID) String() string {
	return c.value
}
