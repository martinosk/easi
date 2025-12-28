package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
)

type StrategyPillar struct {
	id          StrategyPillarID
	name        PillarName
	description PillarDescription
	active      bool
}

func NewStrategyPillar(id StrategyPillarID, name PillarName, description PillarDescription) (StrategyPillar, error) {
	return StrategyPillar{
		id:          id,
		name:        name,
		description: description,
		active:      true,
	}, nil
}

func NewInactiveStrategyPillar(id StrategyPillarID, name PillarName, description PillarDescription) (StrategyPillar, error) {
	return StrategyPillar{
		id:          id,
		name:        name,
		description: description,
		active:      false,
	}, nil
}

func (s StrategyPillar) ID() StrategyPillarID {
	return s.id
}

func (s StrategyPillar) Name() PillarName {
	return s.name
}

func (s StrategyPillar) Description() PillarDescription {
	return s.description
}

func (s StrategyPillar) IsActive() bool {
	return s.active
}

func (s StrategyPillar) WithUpdatedDetails(name PillarName, description PillarDescription) (StrategyPillar, error) {
	return StrategyPillar{
		id:          s.id,
		name:        name,
		description: description,
		active:      s.active,
	}, nil
}

func (s StrategyPillar) Deactivate() StrategyPillar {
	return StrategyPillar{
		id:          s.id,
		name:        s.name,
		description: s.description,
		active:      false,
	}
}

func (s StrategyPillar) Equals(other domain.ValueObject) bool {
	if otherPillar, ok := other.(StrategyPillar); ok {
		return s.id.Equals(otherPillar.id) &&
			s.name.Equals(otherPillar.name) &&
			s.description.Equals(otherPillar.description) &&
			s.active == otherPillar.active
	}
	return false
}
