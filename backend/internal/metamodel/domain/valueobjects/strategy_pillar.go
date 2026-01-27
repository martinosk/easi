package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
)

type StrategyPillar struct {
	id                StrategyPillarID
	name              PillarName
	description       PillarDescription
	active            bool
	fitScoringEnabled bool
	fitCriteria       FitCriteria
	fitType           FitType
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

func (s StrategyPillar) FitScoringEnabled() bool {
	return s.fitScoringEnabled
}

func (s StrategyPillar) FitCriteria() FitCriteria {
	return s.fitCriteria
}

func (s StrategyPillar) FitType() FitType {
	return s.fitType
}

func (s StrategyPillar) clone() StrategyPillar {
	return StrategyPillar{
		id: s.id, name: s.name, description: s.description, active: s.active,
		fitScoringEnabled: s.fitScoringEnabled, fitCriteria: s.fitCriteria, fitType: s.fitType,
	}
}

func (s StrategyPillar) WithUpdatedDetails(name PillarName, description PillarDescription) (StrategyPillar, error) {
	c := s.clone()
	c.name, c.description = name, description
	return c, nil
}

func (s StrategyPillar) WithFitConfiguration(enabled bool, criteria FitCriteria, fitType FitType) StrategyPillar {
	c := s.clone()
	c.fitScoringEnabled, c.fitCriteria, c.fitType = enabled, criteria, fitType
	return c
}

func (s StrategyPillar) WithFitType(fitType FitType) StrategyPillar {
	c := s.clone()
	c.fitType = fitType
	return c
}

func (s StrategyPillar) Deactivate() StrategyPillar {
	c := s.clone()
	c.active = false
	return c
}

func (s StrategyPillar) Equals(other domain.ValueObject) bool {
	if otherPillar, ok := other.(StrategyPillar); ok {
		return s.id.Equals(otherPillar.id) &&
			s.name.Equals(otherPillar.name) &&
			s.description.Equals(otherPillar.description) &&
			s.active == otherPillar.active &&
			s.fitScoringEnabled == otherPillar.fitScoringEnabled &&
			s.fitCriteria.Equals(otherPillar.fitCriteria) &&
			s.fitType.Equals(otherPillar.fitType)
	}
	return false
}
