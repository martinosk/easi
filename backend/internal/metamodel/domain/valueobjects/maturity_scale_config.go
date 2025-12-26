package valueobjects

import (
	"errors"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrScaleFirstSectionMustStartAtZero = errors.New("first section must start at 0")
	ErrScaleLastSectionMustEndAt99      = errors.New("last section must end at 99")
	ErrScaleSectionsMustBeContiguous    = errors.New("sections must be contiguous with no gaps or overlaps")
	ErrScaleSectionsNotInOrder          = errors.New("sections must be ordered 1 through 4")
)

type MaturityScaleConfig struct {
	sections [4]MaturitySection
}

func NewMaturityScaleConfig(sections [4]MaturitySection) (MaturityScaleConfig, error) {
	if err := validateSectionOrder(sections); err != nil {
		return MaturityScaleConfig{}, err
	}
	if err := validateBoundaries(sections); err != nil {
		return MaturityScaleConfig{}, err
	}
	if err := validateContiguity(sections); err != nil {
		return MaturityScaleConfig{}, err
	}
	return MaturityScaleConfig{sections: sections}, nil
}

func validateSectionOrder(sections [4]MaturitySection) error {
	for i, section := range sections {
		if section.Order().Value() != i+1 {
			return ErrScaleSectionsNotInOrder
		}
	}
	return nil
}

func validateBoundaries(sections [4]MaturitySection) error {
	if sections[0].MinValue().Value() != 0 {
		return ErrScaleFirstSectionMustStartAtZero
	}
	if sections[3].MaxValue().Value() != 99 {
		return ErrScaleLastSectionMustEndAt99
	}
	return nil
}

func validateContiguity(sections [4]MaturitySection) error {
	for i := 0; i < 3; i++ {
		if !sections[i].IsAdjacentTo(sections[i+1]) {
			return ErrScaleSectionsMustBeContiguous
		}
	}
	return nil
}

func DefaultMaturityScaleConfig() MaturityScaleConfig {
	order1, _ := NewSectionOrder(1)
	name1, _ := NewSectionName("Genesis")
	min1, _ := NewMaturityValue(0)
	max1, _ := NewMaturityValue(24)
	section1, _ := NewMaturitySection(order1, name1, min1, max1)

	order2, _ := NewSectionOrder(2)
	name2, _ := NewSectionName("Custom Built")
	min2, _ := NewMaturityValue(25)
	max2, _ := NewMaturityValue(49)
	section2, _ := NewMaturitySection(order2, name2, min2, max2)

	order3, _ := NewSectionOrder(3)
	name3, _ := NewSectionName("Product")
	min3, _ := NewMaturityValue(50)
	max3, _ := NewMaturityValue(74)
	section3, _ := NewMaturitySection(order3, name3, min3, max3)

	order4, _ := NewSectionOrder(4)
	name4, _ := NewSectionName("Commodity")
	min4, _ := NewMaturityValue(75)
	max4, _ := NewMaturityValue(99)
	section4, _ := NewMaturitySection(order4, name4, min4, max4)

	return MaturityScaleConfig{
		sections: [4]MaturitySection{section1, section2, section3, section4},
	}
}

func (m MaturityScaleConfig) Sections() [4]MaturitySection {
	return m.sections
}

func (m MaturityScaleConfig) Equals(other domain.ValueObject) bool {
	if otherConfig, ok := other.(MaturityScaleConfig); ok {
		for i := 0; i < 4; i++ {
			if !m.sections[i].Equals(otherConfig.sections[i]) {
				return false
			}
		}
		return true
	}
	return false
}
