package aggregates

import (
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

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadStrategyImportanceFromHistory(eventHistory []domain.DomainEvent) (*StrategyImportance, error) {
	aggregate := &StrategyImportance{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(eventHistory, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

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

	s.apply(event)
	s.RaiseEvent(event)

	return nil
}

func (s *StrategyImportance) Remove() error {
	event := events.NewStrategyImportanceRemoved(
		s.ID(),
		s.businessDomainID.Value(),
		s.capabilityID.Value(),
		s.pillarID.Value(),
	)

	s.apply(event)
	s.RaiseEvent(event)

	return nil
}

func (s *StrategyImportance) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.StrategyImportanceSet:
		s.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		s.businessDomainID, _ = valueobjects.NewBusinessDomainIDFromString(e.BusinessDomainID)
		s.capabilityID, _ = valueobjects.NewCapabilityIDFromString(e.CapabilityID)
		s.pillarID, _ = valueobjects.NewPillarIDFromString(e.PillarID)
		s.importance, _ = valueobjects.NewImportance(e.Importance)
		s.rationale, _ = valueobjects.NewRationale(e.Rationale)
		s.setAt = e.SetAt
	case events.StrategyImportanceUpdated:
		s.importance, _ = valueobjects.NewImportance(e.Importance)
		s.rationale, _ = valueobjects.NewRationale(e.Rationale)
	case events.StrategyImportanceRemoved:
	}
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
