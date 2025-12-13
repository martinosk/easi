package valueobjects

import (
	"easi/backend/internal/shared/domain"

	"github.com/google/uuid"
)

type ComponentID struct {
	value string
}

func NewComponentIDFromString(value string) (ComponentID, error) {
	if value == "" {
		return ComponentID{}, domain.ErrEmptyValue
	}

	if _, err := uuid.Parse(value); err != nil {
		return ComponentID{}, domain.ErrInvalidValue
	}

	return ComponentID{value: value}, nil
}

func (c ComponentID) Value() string {
	return c.value
}

func (c ComponentID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(ComponentID); ok {
		return c.value == otherID.value
	}
	return false
}

func (c ComponentID) String() string {
	return c.value
}
