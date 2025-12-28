package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
)

var (
	ErrMaturitySectionInvalidRange = errors.New("section minValue must be less than or equal to maxValue")
)

type MaturitySection struct {
	order    SectionOrder
	name     SectionName
	minValue MaturityValue
	maxValue MaturityValue
}

func NewMaturitySection(order SectionOrder, name SectionName, minValue MaturityValue, maxValue MaturityValue) (MaturitySection, error) {
	if !minValue.LessThanOrEqual(maxValue) {
		return MaturitySection{}, ErrMaturitySectionInvalidRange
	}
	return MaturitySection{
		order:    order,
		name:     name,
		minValue: minValue,
		maxValue: maxValue,
	}, nil
}

func (m MaturitySection) Order() SectionOrder {
	return m.order
}

func (m MaturitySection) Name() SectionName {
	return m.name
}

func (m MaturitySection) MinValue() MaturityValue {
	return m.minValue
}

func (m MaturitySection) MaxValue() MaturityValue {
	return m.maxValue
}

func (m MaturitySection) IsAdjacentTo(other MaturitySection) bool {
	return m.maxValue.Value()+1 == other.minValue.Value()
}

func (m MaturitySection) Equals(other domain.ValueObject) bool {
	if otherSection, ok := other.(MaturitySection); ok {
		return m.order.Equals(otherSection.order) &&
			m.name.Equals(otherSection.name) &&
			m.minValue.Equals(otherSection.minValue) &&
			m.maxValue.Equals(otherSection.maxValue)
	}
	return false
}
