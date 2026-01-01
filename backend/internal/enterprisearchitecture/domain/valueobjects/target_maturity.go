package valueobjects

import (
	"errors"
	"fmt"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrTargetMaturityOutOfRange = errors.New("target maturity must be between 0 and 99")

type TargetMaturity struct {
	value int
}

func NewTargetMaturity(value int) (TargetMaturity, error) {
	if value < 0 || value > 99 {
		return TargetMaturity{}, ErrTargetMaturityOutOfRange
	}
	return TargetMaturity{value: value}, nil
}

func (m TargetMaturity) Value() int {
	return m.value
}

func (m TargetMaturity) SectionName() string {
	switch {
	case m.value <= 24:
		return "Genesis"
	case m.value <= 49:
		return "Custom Build"
	case m.value <= 74:
		return "Product"
	default:
		return "Commodity"
	}
}

func (m TargetMaturity) SectionOrder() int {
	switch {
	case m.value <= 24:
		return 1
	case m.value <= 49:
		return 2
	case m.value <= 74:
		return 3
	default:
		return 4
	}
}

func (m TargetMaturity) String() string {
	return fmt.Sprintf("%d (%s)", m.value, m.SectionName())
}

func (m TargetMaturity) Equals(other domain.ValueObject) bool {
	if otherMaturity, ok := other.(TargetMaturity); ok {
		return m.value == otherMaturity.value
	}
	return false
}
