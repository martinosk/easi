package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
	"strings"
)

const (
	// MinViewNameLength is the minimum allowed length for a view name
	MinViewNameLength = 1
	// MaxViewNameLength is the maximum allowed length for a view name
	MaxViewNameLength = 100
)

var (
	// ErrViewNameEmpty is returned when view name is empty or whitespace
	ErrViewNameEmpty = errors.New("view name cannot be empty or whitespace only")
	// ErrViewNameTooLong is returned when view name exceeds maximum length
	ErrViewNameTooLong = errors.New("view name cannot exceed 100 characters")
)

// ViewName represents the name of an architecture view
type ViewName struct {
	value string
}

// NewViewName creates a new view name with validation
func NewViewName(value string) (ViewName, error) {
	// Validate not empty or whitespace only
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ViewName{}, ErrViewNameEmpty
	}

	// Validate length (1-100 characters)
	if len(trimmed) > MaxViewNameLength {
		return ViewName{}, ErrViewNameTooLong
	}

	return ViewName{value: trimmed}, nil
}

// Value returns the string value of the name
func (v ViewName) Value() string {
	return v.value
}

// Equals checks if two view names are equal
func (v ViewName) Equals(other domain.ValueObject) bool {
	if otherName, ok := other.(ViewName); ok {
		return v.value == otherName.value
	}
	return false
}

// String implements the Stringer interface
func (v ViewName) String() string {
	return v.value
}
