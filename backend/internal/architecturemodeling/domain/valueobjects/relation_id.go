package valueobjects

import (
	"easi/backend/internal/shared/domain"

	"github.com/google/uuid"
)

// RelationID represents a unique identifier for a component relation
type RelationID struct {
	value string
}

// NewRelationID creates a new relation ID
func NewRelationID() RelationID {
	return RelationID{value: uuid.New().String()}
}

// NewRelationIDFromString creates a relation ID from a string
func NewRelationIDFromString(value string) (RelationID, error) {
	if value == "" {
		return RelationID{}, domain.ErrEmptyValue
	}

	// Validate UUID format
	if _, err := uuid.Parse(value); err != nil {
		return RelationID{}, domain.ErrInvalidValue
	}

	return RelationID{value: value}, nil
}

// Value returns the string value of the ID
func (r RelationID) Value() string {
	return r.value
}

// Equals checks if two relation IDs are equal
func (r RelationID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(RelationID); ok {
		return r.value == otherID.value
	}
	return false
}

// String implements the Stringer interface
func (r RelationID) String() string {
	return r.value
}
