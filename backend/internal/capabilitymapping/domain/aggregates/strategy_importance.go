package aggregates

import (
	"fmt"
	"time"

	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type StrategyImportance struct {
	domain.AggregateRoot
	businessDomainID valueobjects.BusinessDomainID
	capabilityID     valueobjects.CapabilityID
	pillarID         valueobjects.PillarID
	importance       valueobjects.Importance
	rationale        valueobjects.Rationale
	setAt            time.Time
}

type NewImportanceParams struct {
	BusinessDomainID valueobjects.BusinessDomainID
	CapabilityID     valueobjects.CapabilityID
	PillarID         valueobjects.PillarID
	PillarName       string
	Importance       valueobjects.Importance
	Rationale        valueobjects.Rationale
}

func SetStrategyImportance(params NewImportanceParams) (*StrategyImportance, error) {
	aggregate := &StrategyImportance{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewStrategyImportanceSet(events.StrategyImportanceSetParams{
		ID:               aggregate.ID(),
		BusinessDomainID: params.BusinessDomainID.Value(),
		CapabilityID:     params.CapabilityID.Value(),
		PillarID:         params.PillarID.Value(),
		PillarName:       params.PillarName,
		Importance:       params.Importance.Value(),
		Rationale:        params.Rationale.Value(),
	})

	aggregate.raise(event)

	return aggregate, nil
}

func LoadStrategyImportanceFromHistory(eventHistory []domain.DomainEvent) (*StrategyImportance, error) {
	aggregate := &StrategyImportance{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	var applyErr error
	aggregate.LoadFromHistory(eventHistory, func(event domain.DomainEvent) {
		if applyErr != nil {
			return
		}
		applyErr = aggregate.apply(event)
	})
	if applyErr != nil {
		return nil, applyErr
	}

	return aggregate, nil
}

func (s *StrategyImportance) Update(importance valueobjects.Importance, rationale valueobjects.Rationale) error {
	event := events.NewStrategyImportanceUpdated(events.StrategyImportanceUpdatedParams{
		ID:            s.ID(),
		Importance:    importance.Value(),
		Rationale:     rationale.Value(),
		OldImportance: s.importance.Value(),
		OldRationale:  s.rationale.Value(),
	})

	s.raise(event)

	return nil
}

func (s *StrategyImportance) Remove() error {
	event := events.NewStrategyImportanceRemoved(
		s.ID(),
		s.businessDomainID.Value(),
		s.capabilityID.Value(),
		s.pillarID.Value(),
	)

	s.raise(event)

	return nil
}

func (s *StrategyImportance) raise(event domain.DomainEvent) {
	if err := s.apply(event); err != nil {
		panic(fmt.Sprintf("capabilitymapping: in-process apply failed: %v", err))
	}
	s.RaiseEvent(event)
}

func (s *StrategyImportance) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.StrategyImportanceSet:
		return s.applySet(e)
	case events.StrategyImportanceUpdated:
		return s.applyUpdated(e)
	case events.StrategyImportanceRemoved:
	}
	return nil
}

func (s *StrategyImportance) applySet(e events.StrategyImportanceSet) error {
	s.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	businessDomainID, err := valueobjects.NewBusinessDomainIDFromString(e.BusinessDomainID)
	if err != nil {
		return fmt.Errorf("%w: business domain ID %q: %v", domain.ErrCorruptedEvent, e.BusinessDomainID, err)
	}
	s.businessDomainID = businessDomainID
	capabilityID, err := valueobjects.NewCapabilityIDFromString(e.CapabilityID)
	if err != nil {
		return fmt.Errorf("%w: capability ID %q: %v", domain.ErrCorruptedEvent, e.CapabilityID, err)
	}
	s.capabilityID = capabilityID
	pillarID, err := valueobjects.NewPillarIDFromString(e.PillarID)
	if err != nil {
		return fmt.Errorf("%w: pillar ID %q: %v", domain.ErrCorruptedEvent, e.PillarID, err)
	}
	s.pillarID = pillarID
	importance, err := valueobjects.NewImportance(e.Importance)
	if err != nil {
		return fmt.Errorf("%w: importance %d: %v", domain.ErrCorruptedEvent, e.Importance, err)
	}
	s.importance = importance
	rationale, err := valueobjects.NewRationale(e.Rationale)
	if err != nil {
		return fmt.Errorf("%w: rationale: %v", domain.ErrCorruptedEvent, err)
	}
	s.rationale = rationale
	s.setAt = e.SetAt
	return nil
}

func (s *StrategyImportance) applyUpdated(e events.StrategyImportanceUpdated) error {
	importance, err := valueobjects.NewImportance(e.Importance)
	if err != nil {
		return fmt.Errorf("%w: importance %d: %v", domain.ErrCorruptedEvent, e.Importance, err)
	}
	s.importance = importance
	rationale, err := valueobjects.NewRationale(e.Rationale)
	if err != nil {
		return fmt.Errorf("%w: rationale: %v", domain.ErrCorruptedEvent, err)
	}
	s.rationale = rationale
	return nil
}

func (s *StrategyImportance) BusinessDomainID() valueobjects.BusinessDomainID {
	return s.businessDomainID
}

func (s *StrategyImportance) CapabilityID() valueobjects.CapabilityID {
	return s.capabilityID
}

func (s *StrategyImportance) PillarID() valueobjects.PillarID {
	return s.pillarID
}

func (s *StrategyImportance) Importance() valueobjects.Importance {
	return s.importance
}

func (s *StrategyImportance) Rationale() valueobjects.Rationale {
	return s.rationale
}

func (s *StrategyImportance) SetAt() time.Time {
	return s.setAt
}
