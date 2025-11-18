package aggregates

import (
	"time"

	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/domain"
)

type CapabilityRealization struct {
	domain.AggregateRoot
	capabilityID     valueobjects.CapabilityID
	componentID      valueobjects.ComponentID
	realizationLevel valueobjects.RealizationLevel
	notes            valueobjects.Description
	linkedAt         time.Time
}

func NewCapabilityRealization(
	capabilityID valueobjects.CapabilityID,
	componentID valueobjects.ComponentID,
	realizationLevel valueobjects.RealizationLevel,
	notes valueobjects.Description,
) (*CapabilityRealization, error) {
	aggregate := &CapabilityRealization{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewSystemLinkedToCapability(
		aggregate.ID(),
		capabilityID.Value(),
		componentID.Value(),
		realizationLevel.Value(),
		notes.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadCapabilityRealizationFromHistory(events []domain.DomainEvent) (*CapabilityRealization, error) {
	aggregate := &CapabilityRealization{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (cr *CapabilityRealization) Update(
	realizationLevel valueobjects.RealizationLevel,
	notes valueobjects.Description,
) error {
	event := events.NewSystemRealizationUpdated(
		cr.ID(),
		realizationLevel.Value(),
		notes.Value(),
	)

	cr.apply(event)
	cr.RaiseEvent(event)

	return nil
}

func (cr *CapabilityRealization) Delete() error {
	event := events.NewSystemRealizationDeleted(cr.ID())

	cr.apply(event)
	cr.RaiseEvent(event)

	return nil
}

func (cr *CapabilityRealization) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.SystemLinkedToCapability:
		cr.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		cr.capabilityID, _ = valueobjects.NewCapabilityIDFromString(e.CapabilityID)
		cr.componentID, _ = valueobjects.NewComponentIDFromString(e.ComponentID)
		cr.realizationLevel, _ = valueobjects.NewRealizationLevel(e.RealizationLevel)
		cr.notes = valueobjects.NewDescription(e.Notes)
		cr.linkedAt = e.LinkedAt
	case events.SystemRealizationUpdated:
		cr.realizationLevel, _ = valueobjects.NewRealizationLevel(e.RealizationLevel)
		cr.notes = valueobjects.NewDescription(e.Notes)
	case events.SystemRealizationDeleted:
	}
}

func (cr *CapabilityRealization) CapabilityID() valueobjects.CapabilityID {
	return cr.capabilityID
}

func (cr *CapabilityRealization) ComponentID() valueobjects.ComponentID {
	return cr.componentID
}

func (cr *CapabilityRealization) RealizationLevel() valueobjects.RealizationLevel {
	return cr.realizationLevel
}

func (cr *CapabilityRealization) Notes() valueobjects.Description {
	return cr.notes
}

func (cr *CapabilityRealization) LinkedAt() time.Time {
	return cr.linkedAt
}
