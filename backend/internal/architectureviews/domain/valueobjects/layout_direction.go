package valueobjects

import (
	"errors"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidLayoutDirection = errors.New("invalid layout direction: must be 'TB', 'LR', 'BT', or 'RL'")
)

type LayoutDirection struct {
	value string
}

func NewLayoutDirection(value string) (LayoutDirection, error) {
	switch value {
	case "TB", "LR", "BT", "RL":
		return LayoutDirection{value: value}, nil
	default:
		return LayoutDirection{}, ErrInvalidLayoutDirection
	}
}

func DefaultLayoutDirection() LayoutDirection {
	return LayoutDirection{value: "TB"}
}

func (l LayoutDirection) Value() string {
	return l.value
}

func (l LayoutDirection) Equals(other domain.ValueObject) bool {
	if otherDirection, ok := other.(LayoutDirection); ok {
		return l.value == otherDirection.value
	}
	return false
}

func (l LayoutDirection) String() string {
	return l.value
}
