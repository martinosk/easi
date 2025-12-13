package valueobjects

import (
	"easi/backend/internal/shared/domain"

	"github.com/google/uuid"
)

// ViewID represents a unique identifier for an architecture view
type ViewID struct {
	value string
}

// NewViewID creates a new view ID
func NewViewID() ViewID {
	return ViewID{value: uuid.New().String()}
}

// NewViewIDFromString creates a view ID from a string
func NewViewIDFromString(value string) (ViewID, error) {
	if value == "" {
		return ViewID{}, domain.ErrEmptyValue
	}

	// Validate UUID format
	if _, err := uuid.Parse(value); err != nil {
		return ViewID{}, domain.ErrInvalidValue
	}

	return ViewID{value: value}, nil
}

// Value returns the string value of the ID
func (v ViewID) Value() string {
	return v.value
}

// Equals checks if two view IDs are equal
func (v ViewID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(ViewID); ok {
		return v.value == otherID.value
	}
	return false
}

// String implements the Stringer interface
func (v ViewID) String() string {
	return v.value
}
