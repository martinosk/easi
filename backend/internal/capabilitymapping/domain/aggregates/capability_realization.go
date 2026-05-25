package aggregates

import (
	"fmt"
	"time"

	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
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
	componentName string,
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
		componentName,
		realizationLevel.Value(),
		notes.Value(),
	)

	aggregate.raise(event)

	return aggregate, nil
}

func LoadCapabilityRealizationFromHistory(events []domain.DomainEvent) (*CapabilityRealization, error) {
	aggregate := &CapabilityRealization{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	var applyErr error
	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
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

func (cr *CapabilityRealization) Update(
	realizationLevel valueobjects.RealizationLevel,
	notes valueobjects.Description,
) error {
	event := events.NewSystemRealizationUpdated(
		cr.ID(),
		realizationLevel.Value(),
		notes.Value(),
	)

	cr.raise(event)

	return nil
}

func (cr *CapabilityRealization) Delete() error {
	event := events.NewSystemRealizationDeleted(cr.ID())

	cr.raise(event)

	return nil
}

func (cr *CapabilityRealization) raise(event domain.DomainEvent) {
	if err := cr.apply(event); err != nil {
		panic(fmt.Sprintf("capabilitymapping: in-process apply failed: %v", err))
	}
	cr.RaiseEvent(event)
}

func (cr *CapabilityRealization) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.SystemLinkedToCapability:
		cr.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		capabilityID, err := valueobjects.NewCapabilityIDFromString(e.CapabilityID)
		if err != nil {
			return fmt.Errorf("%w: capability ID %q: %v", domain.ErrCorruptedEvent, e.CapabilityID, err)
		}
		cr.capabilityID = capabilityID
		componentID, err := valueobjects.NewComponentIDFromString(e.ComponentID)
		if err != nil {
			return fmt.Errorf("%w: component ID %q: %v", domain.ErrCorruptedEvent, e.ComponentID, err)
		}
		cr.componentID = componentID
		realizationLevel, err := valueobjects.NewRealizationLevel(e.RealizationLevel)
		if err != nil {
			return fmt.Errorf("%w: realization level %q: %v", domain.ErrCorruptedEvent, e.RealizationLevel, err)
		}
		cr.realizationLevel = realizationLevel
		cr.notes = valueobjects.MustNewDescription(e.Notes)
		cr.linkedAt = e.LinkedAt
	case events.SystemRealizationUpdated:
		realizationLevel, err := valueobjects.NewRealizationLevel(e.RealizationLevel)
		if err != nil {
			return fmt.Errorf("%w: realization level %q: %v", domain.ErrCorruptedEvent, e.RealizationLevel, err)
		}
		cr.realizationLevel = realizationLevel
		cr.notes = valueobjects.MustNewDescription(e.Notes)
	case events.SystemRealizationDeleted:
	}
	return nil
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
