package valueobjects

import (
	"errors"
	"strings"

	"easi/backend/internal/shared/domain"
)

var (
	ErrInvalidMaturityLevel = errors.New("invalid maturity level: must be Initial, Developing, Defined, Managed, or Optimizing")
)

type MaturityLevel string

const (
	MaturityInitial    MaturityLevel = "Initial"
	MaturityDeveloping MaturityLevel = "Developing"
	MaturityDefined    MaturityLevel = "Defined"
	MaturityManaged    MaturityLevel = "Managed"
	MaturityOptimizing MaturityLevel = "Optimizing"
)

func NewMaturityLevel(value string) (MaturityLevel, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return MaturityInitial, nil
	}

	switch MaturityLevel(normalized) {
	case MaturityInitial, MaturityDeveloping, MaturityDefined, MaturityManaged, MaturityOptimizing:
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
	case MaturityInitial:
		return 1
	case MaturityDeveloping:
		return 2
	case MaturityDefined:
		return 3
	case MaturityManaged:
		return 4
	case MaturityOptimizing:
		return 5
	default:
		return 0
	}
}
