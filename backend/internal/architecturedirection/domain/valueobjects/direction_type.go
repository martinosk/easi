package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidDirectionType = errors.New("direction type must be one of consolidate, decompose, stay")

const (
	DirectionTypeConsolidate = "consolidate"
	DirectionTypeDecompose   = "decompose"
	DirectionTypeStay        = "stay"
)

type DirectionType struct {
	value string
}

func NewDirectionType(value string) (DirectionType, error) {
	switch value {
	case DirectionTypeConsolidate, DirectionTypeDecompose, DirectionTypeStay:
		return DirectionType{value: value}, nil
	default:
		return DirectionType{}, ErrInvalidDirectionType
	}
}

func (d DirectionType) Value() string { return d.value }

func (d DirectionType) IsConsolidate() bool { return d.value == DirectionTypeConsolidate }
func (d DirectionType) IsDecompose() bool   { return d.value == DirectionTypeDecompose }
func (d DirectionType) IsStay() bool        { return d.value == DirectionTypeStay }

func (d DirectionType) RequiresExactlyOneSource() bool {
	return d.IsDecompose() || d.IsStay()
}

func (d DirectionType) ExactSourceCount() int {
	if d.RequiresExactlyOneSource() {
		return 1
	}
	return 0
}

func (d DirectionType) MinSourceCount() int {
	if d.IsConsolidate() {
		return 2
	}
	return 1
}

func (d DirectionType) IsValidPlacementCount(count int) bool {
	switch d.value {
	case DirectionTypeConsolidate:
		return count == 1
	case DirectionTypeDecompose:
		return count >= 1
	case DirectionTypeStay:
		return count == 0
	}
	return false
}

func (d DirectionType) Equals(other domain.ValueObject) bool {
	if otherType, ok := other.(DirectionType); ok {
		return d.value == otherType.value
	}
	return false
}
