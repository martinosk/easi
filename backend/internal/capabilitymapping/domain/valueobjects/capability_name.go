package valueobjects

import (
	"errors"
	"strings"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrCapabilityNameEmpty = errors.New("capability name cannot be empty or whitespace only")
)

type CapabilityName struct {
	value string
}

func NewCapabilityName(value string) (CapabilityName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return CapabilityName{}, ErrCapabilityNameEmpty
	}

	return CapabilityName{value: trimmed}, nil
}

func (c CapabilityName) Value() string {
	return c.value
}

func (c CapabilityName) Equals(other domain.ValueObject) bool {
	if otherName, ok := other.(CapabilityName); ok {
		return c.value == otherName.value
	}
	return false
}

func (c CapabilityName) String() string {
	return c.value
}
