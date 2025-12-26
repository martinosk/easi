package valueobjects

import (
	"errors"
	"fmt"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrMaturityValueOutOfRange = errors.New("maturity value must be between 0 and 99")
)

type MaturityValue struct {
	value int
}

func NewMaturityValue(value int) (MaturityValue, error) {
	if value < 0 || value > 99 {
		return MaturityValue{}, ErrMaturityValueOutOfRange
	}
	return MaturityValue{value: value}, nil
}

func (m MaturityValue) Value() int {
	return m.value
}

func (m MaturityValue) Equals(other domain.ValueObject) bool {
	if otherVal, ok := other.(MaturityValue); ok {
		return m.value == otherVal.value
	}
	return false
}

func (m MaturityValue) String() string {
	return fmt.Sprintf("%d", m.value)
}

func (m MaturityValue) LessThan(other MaturityValue) bool {
	return m.value < other.value
}

func (m MaturityValue) LessThanOrEqual(other MaturityValue) bool {
	return m.value <= other.value
}
