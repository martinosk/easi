package valueobjects

import (
	"errors"
	"fmt"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidContainmentKind = errors.New("invalid containment kind: must be composition or aggregation")

const (
	ContainmentComposition = "composition"
	ContainmentAggregation = "aggregation"
)

type ContainmentKind struct {
	value string
}

func NewContainmentKind(value string) (ContainmentKind, error) {
	switch value {
	case ContainmentComposition, ContainmentAggregation:
		return ContainmentKind{value: value}, nil
	default:
		return ContainmentKind{}, fmt.Errorf("%w: %s", ErrInvalidContainmentKind, value)
	}
}

func (k ContainmentKind) String() string {
	return k.value
}

func (k ContainmentKind) IsComposition() bool {
	return k.value == ContainmentComposition
}

func (k ContainmentKind) Equals(other domain.ValueObject) bool {
	if otherKind, ok := other.(ContainmentKind); ok {
		return k.value == otherKind.value
	}
	return false
}
