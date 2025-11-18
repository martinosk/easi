package valueobjects

import (
	"errors"
	"strings"

	"easi/backend/internal/shared/domain"
)

var (
	ErrInvalidStrategyPillar = errors.New("invalid strategy pillar: must be AlwaysOn, Grow, or Transform")
)

type StrategyPillar string

const (
	PillarAlwaysOn  StrategyPillar = "AlwaysOn"
	PillarGrow      StrategyPillar = "Grow"
	PillarTransform StrategyPillar = "Transform"
)

func NewStrategyPillar(value string) (StrategyPillar, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", nil
	}

	switch StrategyPillar(normalized) {
	case PillarAlwaysOn, PillarGrow, PillarTransform:
		return StrategyPillar(normalized), nil
	default:
		return "", ErrInvalidStrategyPillar
	}
}

func (s StrategyPillar) Value() string {
	return string(s)
}

func (s StrategyPillar) Equals(other domain.ValueObject) bool {
	if otherPillar, ok := other.(StrategyPillar); ok {
		return s == otherPillar
	}
	return false
}

func (s StrategyPillar) String() string {
	return string(s)
}

func (s StrategyPillar) IsEmpty() bool {
	return s == ""
}
