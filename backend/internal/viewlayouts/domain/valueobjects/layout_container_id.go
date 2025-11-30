package valueobjects

import (
	"easi/backend/internal/shared/domain"

	"github.com/google/uuid"
)

type LayoutContainerID struct {
	value string
}

func NewLayoutContainerID() LayoutContainerID {
	return LayoutContainerID{value: uuid.New().String()}
}

func NewLayoutContainerIDFromString(value string) (LayoutContainerID, error) {
	if value == "" {
		return LayoutContainerID{}, domain.ErrEmptyValue
	}

	if _, err := uuid.Parse(value); err != nil {
		return LayoutContainerID{}, domain.ErrInvalidValue
	}

	return LayoutContainerID{value: value}, nil
}

func (l LayoutContainerID) Value() string {
	return l.value
}

func (l LayoutContainerID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(LayoutContainerID); ok {
		return l.value == otherID.value
	}
	return false
}

func (l LayoutContainerID) String() string {
	return l.value
}
