package valueobjects

import (
	"errors"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidPillarWeight = errors.New("pillar weight must be between 0 and 100")
)

type PillarWeight struct {
	value int
}

func NewPillarWeight(value int) (PillarWeight, error) {
	if value < 0 || value > 100 {
		return PillarWeight{}, ErrInvalidPillarWeight
	}

	return PillarWeight{value: value}, nil
}

func (p PillarWeight) Value() int {
	return p.value
}

func (p PillarWeight) Equals(other domain.ValueObject) bool {
	if otherWeight, ok := other.(PillarWeight); ok {
		return p.value == otherWeight.value
	}
	return false
}
