package valueobjects

import (
	"strings"

	"easi/backend/internal/shared/domain"
)

// Description represents a textual description
type Description struct {
	value string
}

// NewDescription creates a new description (can be empty)
func NewDescription(value string) Description {
	return Description{value: strings.TrimSpace(value)}
}

// Value returns the string value of the description
func (d Description) Value() string {
	return d.value
}

// IsEmpty checks if the description is empty
func (d Description) IsEmpty() bool {
	return d.value == ""
}

// Equals checks if two descriptions are equal
func (d Description) Equals(other domain.ValueObject) bool {
	if otherDesc, ok := other.(Description); ok {
		return d.value == otherDesc.value
	}
	return false
}

// String implements the Stringer interface
func (d Description) String() string {
	return d.value
}
