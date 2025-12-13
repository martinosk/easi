package valueobjects

import (
	"easi/backend/internal/shared/domain"

	"github.com/google/uuid"
)

// ViewComponentID represents a unique identifier for a component within a view
type ViewComponentID struct {
	value string
}

// NewViewComponentID creates a new view component ID
func NewViewComponentID() ViewComponentID {
	return ViewComponentID{value: uuid.New().String()}
}

// NewViewComponentIDFromString creates a view component ID from a string
func NewViewComponentIDFromString(value string) (ViewComponentID, error) {
	if value == "" {
		return ViewComponentID{}, domain.ErrEmptyValue
	}

	// Validate UUID format
	if _, err := uuid.Parse(value); err != nil {
		return ViewComponentID{}, domain.ErrInvalidValue
	}

	return ViewComponentID{value: value}, nil
}

// Value returns the string value of the ID
func (v ViewComponentID) Value() string {
	return v.value
}

// Equals checks if two view component IDs are equal
func (v ViewComponentID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(ViewComponentID); ok {
		return v.value == otherID.value
	}
	return false
}

// String implements the Stringer interface
func (v ViewComponentID) String() string {
	return v.value
}
