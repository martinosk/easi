package valueobjects

import (
	"errors"
	"strings"

	"github.com/easi/backend/internal/shared/domain"
)

var (
	// ErrComponentNameEmpty is returned when component name is empty or whitespace
	ErrComponentNameEmpty = errors.New("component name cannot be empty or whitespace only")
)

// ComponentName represents the name of an application component
type ComponentName struct {
	value string
}

// NewComponentName creates a new component name with validation
func NewComponentName(value string) (ComponentName, error) {
	// Validate not empty or whitespace only
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ComponentName{}, ErrComponentNameEmpty
	}

	return ComponentName{value: trimmed}, nil
}

// Value returns the string value of the name
func (c ComponentName) Value() string {
	return c.value
}

// Equals checks if two component names are equal
func (c ComponentName) Equals(other domain.ValueObject) bool {
	if otherName, ok := other.(ComponentName); ok {
		return c.value == otherName.value
	}
	return false
}

// String implements the Stringer interface
func (c ComponentName) String() string {
	return c.value
}
