package valueobjects

import (
	"errors"
	"strings"

	"easi/backend/internal/shared/domain"
)

var (
	ErrInvalidMaturityLevel = errors.New("invalid maturity level: must be Genesis, Custom Build, Product, or Commodity")
)

type MaturityLevel string

const (
	MaturityGenesis     MaturityLevel = "Genesis"
	MaturityCustomBuild MaturityLevel = "Custom Build"
	MaturityProduct     MaturityLevel = "Product"
	MaturityCommodity   MaturityLevel = "Commodity"
)

func NewMaturityLevel(value string) (MaturityLevel, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return MaturityGenesis, nil
	}

	switch MaturityLevel(normalized) {
	case MaturityGenesis, MaturityCustomBuild, MaturityProduct, MaturityCommodity:
		return MaturityLevel(normalized), nil
	default:
		return "", ErrInvalidMaturityLevel
	}
}

func (m MaturityLevel) Value() string {
	return string(m)
}

func (m MaturityLevel) Equals(other domain.ValueObject) bool {
	if otherLevel, ok := other.(MaturityLevel); ok {
		return m == otherLevel
	}
	return false
}

func (m MaturityLevel) String() string {
	return string(m)
}

func (m MaturityLevel) NumericValue() int {
	switch m {
	case MaturityGenesis:
		return 1
	case MaturityCustomBuild:
		return 2
	case MaturityProduct:
		return 3
	case MaturityCommodity:
		return 4
	default:
		return 0
	}
}
