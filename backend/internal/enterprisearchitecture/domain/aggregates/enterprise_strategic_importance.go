package aggregates

import (
	"easi/backend/internal/enterprisearchitecture/domain/events"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseStrategicImportance struct {
	domain.AggregateRoot
	enterpriseCapabilityID valueobjects.EnterpriseCapabilityID
	pillarID               valueobjects.PillarID
	importance             valueobjects.Importance
	rationale              valueobjects.Rationale
	setAt                  valueobjects.SetAt
}

type NewEnterpriseImportanceParams struct {
	EnterpriseCapabilityID valueobjects.EnterpriseCapabilityID
	PillarID               valueobjects.PillarID
	PillarName             string
	Importance             valueobjects.Importance
	Rationale              valueobjects.Rationale
}

func SetEnterpriseStrategicImportance(params NewEnterpriseImportanceParams) (*EnterpriseStrategicImportance, error) {
	id := valueobjects.NewEnterpriseStrategicImportanceIDFromComposite(
		params.EnterpriseCapabilityID,
		params.PillarID,
	)

	aggregate := &EnterpriseStrategicImportance{
		AggregateRoot: domain.NewAggregateRootWithID(id.Value()),
	}

	event := events.NewEnterpriseStrategicImportanceSet(events.EnterpriseStrategicImportanceSetParams{
		ID:                     aggregate.ID(),
		EnterpriseCapabilityID: params.EnterpriseCapabilityID.Value(),
		PillarID:               params.PillarID.Value(),
		PillarName:             params.PillarName,
		Importance:             params.Importance.Value(),
		Rationale:              params.Rationale.Value(),
	})

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadEnterpriseStrategicImportanceFromHistory(eventHistory []domain.DomainEvent) (*EnterpriseStrategicImportance, error) {
	aggregate := &EnterpriseStrategicImportance{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(eventHistory, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (s *EnterpriseStrategicImportance) Update(importance valueobjects.Importance, rationale valueobjects.Rationale) error {
	event := events.NewEnterpriseStrategicImportanceUpdated(events.EnterpriseStrategicImportanceUpdatedParams{
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

func (s *EnterpriseStrategicImportance) Remove() error {
	event := events.NewEnterpriseStrategicImportanceRemoved(
		s.ID(),
		s.enterpriseCapabilityID.Value(),
		s.pillarID.Value(),
	)

	s.apply(event)
	s.RaiseEvent(event)

	return nil
}

func (s *EnterpriseStrategicImportance) apply(event domain.DomainEvent) {
	switch evt := event.(type) {
	case events.EnterpriseStrategicImportanceSet:
		s.AggregateRoot = domain.NewAggregateRootWithID(evt.ID)
		s.enterpriseCapabilityID = mustNewEnterpriseCapabilityIDFromString(evt.EnterpriseCapabilityID)
		s.pillarID = mustNewPillarIDFromString(evt.PillarID)
		s.importance = mustNewImportance(evt.Importance)
		s.rationale = mustNewRationale(evt.Rationale)
		s.setAt = valueobjects.NewSetAtFromTime(evt.SetAt)
	case events.EnterpriseStrategicImportanceUpdated:
		s.importance = mustNewImportance(evt.Importance)
		s.rationale = mustNewRationale(evt.Rationale)
	case events.EnterpriseStrategicImportanceRemoved:
	}
}

func mustNewPillarIDFromString(value string) valueobjects.PillarID {
	id, err := valueobjects.NewPillarIDFromString(value)
	if err != nil {
		panic("corrupted event store: invalid pillar ID: " + value)
	}
	return id
}

func mustNewImportance(value int) valueobjects.Importance {
	imp, err := valueobjects.NewImportance(value)
	if err != nil {
		panic("corrupted event store: invalid importance value")
	}
	return imp
}

func mustNewRationale(value string) valueobjects.Rationale {
	rat, err := valueobjects.NewRationale(value)
	if err != nil {
		panic("corrupted event store: invalid rationale: " + value)
	}
	return rat
}

func (s *EnterpriseStrategicImportance) EnterpriseCapabilityID() valueobjects.EnterpriseCapabilityID {
	return s.enterpriseCapabilityID
}

func (s *EnterpriseStrategicImportance) PillarID() valueobjects.PillarID {
	return s.pillarID
}

func (s *EnterpriseStrategicImportance) Importance() valueobjects.Importance {
	return s.importance
}

func (s *EnterpriseStrategicImportance) Rationale() valueobjects.Rationale {
	return s.rationale
}

func (s *EnterpriseStrategicImportance) SetAt() valueobjects.SetAt {
	return s.setAt
}
