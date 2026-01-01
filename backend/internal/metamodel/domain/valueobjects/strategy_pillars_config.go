package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrTooManyPillars               = errors.New("cannot have more than 20 pillars")
	ErrPillarNameDuplicate          = errors.New("pillar name already exists")
	ErrPillarNotFound               = errors.New("pillar not found")
	ErrCannotRemoveLastActivePillar = errors.New("cannot remove the last active pillar")
	ErrPillarAlreadyInactive        = errors.New("pillar is already inactive")
)

const MaxPillars = 20

type StrategyPillarsConfig struct {
	pillars []StrategyPillar
}

func NewStrategyPillarsConfig(pillars []StrategyPillar) (StrategyPillarsConfig, error) {
	if len(pillars) > MaxPillars {
		return StrategyPillarsConfig{}, ErrTooManyPillars
	}
	if err := validateUniqueNames(pillars); err != nil {
		return StrategyPillarsConfig{}, err
	}
	pillarsCopy := make([]StrategyPillar, len(pillars))
	copy(pillarsCopy, pillars)
	return StrategyPillarsConfig{pillars: pillarsCopy}, nil
}

func validateUniqueNames(pillars []StrategyPillar) error {
	seen := make(map[string]bool)
	for _, p := range pillars {
		normalizedName := p.Name().Value()
		for seenName := range seen {
			existingName, _ := NewPillarName(seenName)
			if p.Name().EqualsIgnoreCase(existingName) {
				return ErrPillarNameDuplicate
			}
		}
		seen[normalizedName] = true
	}
	return nil
}

func DefaultStrategyPillarsConfig() StrategyPillarsConfig {
	alwaysOnName, _ := NewPillarName("Always On")
	alwaysOnDesc, _ := NewPillarDescription("Core capabilities that must always be operational")
	alwaysOn, _ := NewStrategyPillar(NewStrategyPillarID(), alwaysOnName, alwaysOnDesc)

	growName, _ := NewPillarName("Grow")
	growDesc, _ := NewPillarDescription("Capabilities driving business growth")
	grow, _ := NewStrategyPillar(NewStrategyPillarID(), growName, growDesc)

	transformName, _ := NewPillarName("Transform")
	transformDesc, _ := NewPillarDescription("Capabilities enabling digital transformation")
	transform, _ := NewStrategyPillar(NewStrategyPillarID(), transformName, transformDesc)

	config, _ := NewStrategyPillarsConfig([]StrategyPillar{alwaysOn, grow, transform})
	return config
}

func (s StrategyPillarsConfig) Pillars() []StrategyPillar {
	result := make([]StrategyPillar, len(s.pillars))
	copy(result, s.pillars)
	return result
}

func (s StrategyPillarsConfig) ActivePillars() []StrategyPillar {
	result := make([]StrategyPillar, 0)
	for _, p := range s.pillars {
		if p.IsActive() {
			result = append(result, p)
		}
	}
	return result
}

func (s StrategyPillarsConfig) CountActive() int {
	count := 0
	for _, p := range s.pillars {
		if p.IsActive() {
			count++
		}
	}
	return count
}

func (s StrategyPillarsConfig) FindByID(id StrategyPillarID) (StrategyPillar, bool) {
	for _, p := range s.pillars {
		if p.ID().Equals(id) {
			return p, true
		}
	}
	return StrategyPillar{}, false
}

func (s StrategyPillarsConfig) HasPillarWithName(name PillarName) bool {
	for _, p := range s.pillars {
		if p.Name().EqualsIgnoreCase(name) {
			return true
		}
	}
	return false
}

func (s StrategyPillarsConfig) hasPillarWithNameExcluding(name PillarName, excludeID StrategyPillarID) bool {
	for _, p := range s.pillars {
		if p.ID().Equals(excludeID) {
			continue
		}
		if p.Name().EqualsIgnoreCase(name) {
			return true
		}
	}
	return false
}

func (s StrategyPillarsConfig) WithAddedPillar(pillar StrategyPillar) (StrategyPillarsConfig, error) {
	if len(s.pillars) >= MaxPillars {
		return StrategyPillarsConfig{}, ErrTooManyPillars
	}
	if s.HasPillarWithName(pillar.Name()) {
		return StrategyPillarsConfig{}, ErrPillarNameDuplicate
	}
	newPillars := make([]StrategyPillar, len(s.pillars), len(s.pillars)+1)
	copy(newPillars, s.pillars)
	newPillars = append(newPillars, pillar)
	return StrategyPillarsConfig{pillars: newPillars}, nil
}

func (s StrategyPillarsConfig) WithUpdatedPillar(id StrategyPillarID, name PillarName, description PillarDescription) (StrategyPillarsConfig, error) {
	found := false
	var idx int
	for i, p := range s.pillars {
		if p.ID().Equals(id) {
			found = true
			idx = i
			break
		}
	}
	if !found {
		return StrategyPillarsConfig{}, ErrPillarNotFound
	}
	if s.hasPillarWithNameExcluding(name, id) {
		return StrategyPillarsConfig{}, ErrPillarNameDuplicate
	}
	newPillars := make([]StrategyPillar, len(s.pillars))
	copy(newPillars, s.pillars)
	updated, _ := newPillars[idx].WithUpdatedDetails(name, description)
	newPillars[idx] = updated
	return StrategyPillarsConfig{pillars: newPillars}, nil
}

func (s StrategyPillarsConfig) WithRemovedPillar(id StrategyPillarID) (StrategyPillarsConfig, error) {
	found := false
	var idx int
	var pillar StrategyPillar
	for i, p := range s.pillars {
		if p.ID().Equals(id) {
			found = true
			idx = i
			pillar = p
			break
		}
	}
	if !found {
		return StrategyPillarsConfig{}, ErrPillarNotFound
	}
	if !pillar.IsActive() {
		return StrategyPillarsConfig{}, ErrPillarAlreadyInactive
	}
	if s.CountActive() <= 1 {
		return StrategyPillarsConfig{}, ErrCannotRemoveLastActivePillar
	}
	newPillars := make([]StrategyPillar, len(s.pillars))
	copy(newPillars, s.pillars)
	newPillars[idx] = newPillars[idx].Deactivate()
	return StrategyPillarsConfig{pillars: newPillars}, nil
}

func (s StrategyPillarsConfig) WithUpdatedPillarFitConfiguration(id StrategyPillarID, enabled bool, criteria FitCriteria) (StrategyPillarsConfig, error) {
	found := false
	var idx int
	for i, p := range s.pillars {
		if p.ID().Equals(id) {
			found = true
			idx = i
			break
		}
	}
	if !found {
		return StrategyPillarsConfig{}, ErrPillarNotFound
	}
	newPillars := make([]StrategyPillar, len(s.pillars))
	copy(newPillars, s.pillars)
	newPillars[idx] = newPillars[idx].WithFitConfiguration(enabled, criteria)
	return StrategyPillarsConfig{pillars: newPillars}, nil
}

func (s StrategyPillarsConfig) Equals(other domain.ValueObject) bool {
	otherConfig, ok := other.(StrategyPillarsConfig)
	if !ok {
		return false
	}
	return s.equalsPillars(otherConfig)
}

func (s StrategyPillarsConfig) equalsPillars(other StrategyPillarsConfig) bool {
	if len(s.pillars) != len(other.pillars) {
		return false
	}
	for i := range s.pillars {
		if !s.pillars[i].Equals(other.pillars[i]) {
			return false
		}
	}
	return true
}
