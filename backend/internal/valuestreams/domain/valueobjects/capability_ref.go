package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrCapabilityRefEmpty = errors.New("capability reference cannot be empty")

type CapabilityRef struct {
	value string
}

func NewCapabilityRef(value string) (CapabilityRef, error) {
	if value == "" {
		return CapabilityRef{}, ErrCapabilityRefEmpty
	}
	return CapabilityRef{value: value}, nil
}

func (c CapabilityRef) Value() string {
	return c.value
}

func (c CapabilityRef) Equals(other domain.ValueObject) bool {
	if otherRef, ok := other.(CapabilityRef); ok {
		return c.value == otherRef.value
	}
	return false
}

func (c CapabilityRef) String() string {
	return c.value
}
