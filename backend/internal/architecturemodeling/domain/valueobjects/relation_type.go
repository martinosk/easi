package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
)

var (
	// ErrInvalidRelationType is returned when relation type is not valid
	ErrInvalidRelationType = errors.New("relation type must be either 'Triggers' or 'Serves'")
)

// RelationType represents the type of relation between components
type RelationType string

const (
	// RelationTypeTriggers indicates source initiates functionality in target
	RelationTypeTriggers RelationType = "Triggers"

	// RelationTypeServes indicates source provides services to target
	RelationTypeServes RelationType = "Serves"
)

// NewRelationType creates a new relation type with validation
func NewRelationType(value string) (RelationType, error) {
	rt := RelationType(value)

	switch rt {
	case RelationTypeTriggers, RelationTypeServes:
		return rt, nil
	default:
		return "", ErrInvalidRelationType
	}
}

// Value returns the string value of the relation type
func (r RelationType) Value() string {
	return string(r)
}

// Equals checks if two relation types are equal
func (r RelationType) Equals(other domain.ValueObject) bool {
	if otherType, ok := other.(RelationType); ok {
		return r == otherType
	}
	return false
}

// String implements the Stringer interface
func (r RelationType) String() string {
	return string(r)
}
