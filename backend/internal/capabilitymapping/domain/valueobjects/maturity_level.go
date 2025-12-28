package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
)

var (
	ErrInvalidMaturityLevel    = errors.New("invalid maturity level: must be Genesis, Custom Build, Product, or Commodity")
	ErrMaturityValueOutOfRange = errors.New("maturity value must be between 0 and 99")
)

const (
	DefaultMaturityValue = 12
)

type MaturityLevel struct {
	value int
}

var (
	MaturityGenesis     = MaturityLevel{value: 12}
	MaturityCustomBuild = MaturityLevel{value: 37}
	MaturityProduct     = MaturityLevel{value: 62}
	MaturityCommodity   = MaturityLevel{value: 87}
)

func NewMaturityLevel(value string) (MaturityLevel, error) {
	if value == "" {
		return MaturityGenesis, nil
	}

	midpoint, ok := legacyStringToValue(value)
	if !ok {
		return MaturityLevel{}, ErrInvalidMaturityLevel
	}
	return MaturityLevel{value: midpoint}, nil
}

func NewMaturityLevelFromValue(value int) (MaturityLevel, error) {
	if value < 0 || value > 99 {
		return MaturityLevel{}, ErrMaturityValueOutOfRange
	}
	return MaturityLevel{value: value}, nil
}

func legacyStringToValue(name string) (int, bool) {
	switch name {
	case "Genesis":
		return 12, true
	case "Custom Build":
		return 37, true
	case "Product":
		return 62, true
	case "Commodity":
		return 87, true
	default:
		return 0, false
	}
}

type maturitySection struct {
	name  string
	order int
	min   int
	max   int
}

var maturitySections = []maturitySection{
	{"Genesis", 1, 0, 24},
	{"Custom Build", 2, 25, 49},
	{"Product", 3, 50, 74},
	{"Commodity", 4, 75, 99},
}

func (m MaturityLevel) section() maturitySection {
	for _, s := range maturitySections {
		if m.value <= s.max {
			return s
		}
	}
	return maturitySections[3]
}

func (m MaturityLevel) Value() int {
	return m.value
}

func (m MaturityLevel) StringValue() string {
	return m.SectionName()
}

func (m MaturityLevel) SectionName() string {
	return m.section().name
}

func (m MaturityLevel) SectionOrder() int {
	return m.section().order
}

func (m MaturityLevel) SectionRange() (int, int) {
	s := m.section()
	return s.min, s.max
}

func (m MaturityLevel) Equals(other domain.ValueObject) bool {
	if otherLevel, ok := other.(MaturityLevel); ok {
		return m.value == otherLevel.value
	}
	return false
}

func (m MaturityLevel) NumericValue() int {
	return m.SectionOrder()
}
