package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
)

var (
	ErrSectionOrderOutOfRange = errors.New("section order must be between 1 and 4")
)

type SectionOrder struct {
	value int
}

func NewSectionOrder(value int) (SectionOrder, error) {
	if value < 1 || value > 4 {
		return SectionOrder{}, ErrSectionOrderOutOfRange
	}
	return SectionOrder{value: value}, nil
}

func (s SectionOrder) Value() int {
	return s.value
}

func (s SectionOrder) Equals(other domain.ValueObject) bool {
	if otherOrder, ok := other.(SectionOrder); ok {
		return s.value == otherOrder.value
	}
	return false
}
