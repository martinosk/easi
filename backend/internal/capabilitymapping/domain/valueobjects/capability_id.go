package valueobjects

import (
	"easi/backend/internal/shared/domain"

	"github.com/google/uuid"
)

type CapabilityID struct {
	value string
}

func NewCapabilityID() CapabilityID {
	return CapabilityID{value: uuid.New().String()}
}

func NewCapabilityIDFromString(value string) (CapabilityID, error) {
	if value == "" {
		return CapabilityID{}, domain.ErrEmptyValue
	}

	if _, err := uuid.Parse(value); err != nil {
		return CapabilityID{}, domain.ErrInvalidValue
	}

	return CapabilityID{value: value}, nil
}

func (c CapabilityID) Value() string {
	return c.value
}

func (c CapabilityID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(CapabilityID); ok {
		return c.value == otherID.value
	}
	return false
}

func (c CapabilityID) String() string {
	return c.value
}
