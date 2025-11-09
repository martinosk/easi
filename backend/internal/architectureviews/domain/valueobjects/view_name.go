package valueobjects

import (
	"errors"
	"strings"

	"github.com/easi/backend/internal/shared/domain"
)

var (
	// ErrViewNameEmpty is returned when view name is empty or whitespace
	ErrViewNameEmpty = errors.New("view name cannot be empty or whitespace only")
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
