package aggregates

import (
	"fmt"
	"time"

	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type BusinessDomainAssignment struct {
	domain.AggregateRoot
	businessDomainID valueobjects.BusinessDomainID
	capabilityID     valueobjects.CapabilityID
	assignedAt       time.Time
}

func AssignCapabilityToDomain(
	businessDomainID valueobjects.BusinessDomainID,
	capabilityID valueobjects.CapabilityID,
) (*BusinessDomainAssignment, error) {
	aggregate := &BusinessDomainAssignment{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewCapabilityAssignedToDomain(
		aggregate.ID(),
		businessDomainID.Value(),
		capabilityID.Value(),
	)

	aggregate.raise(event)

	return aggregate, nil
}

func LoadBusinessDomainAssignmentFromHistory(events []domain.DomainEvent) (*BusinessDomainAssignment, error) {
	aggregate := &BusinessDomainAssignment{
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

func (a *BusinessDomainAssignment) Unassign() error {
	event := events.NewCapabilityUnassignedFromDomain(
		a.ID(),
		a.businessDomainID.Value(),
		a.capabilityID.Value(),
	)

	a.raise(event)

	return nil
}

func (a *BusinessDomainAssignment) raise(event domain.DomainEvent) {
	if err := a.apply(event); err != nil {
		panic(fmt.Sprintf("capabilitymapping: in-process apply failed: %v", err))
	}
	a.RaiseEvent(event)
}

func (a *BusinessDomainAssignment) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.CapabilityAssignedToDomain:
		a.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		businessDomainID, err := valueobjects.NewBusinessDomainIDFromString(e.BusinessDomainID)
		if err != nil {
			return fmt.Errorf("%w: business domain ID %q: %v", domain.ErrCorruptedEvent, e.BusinessDomainID, err)
		}
		a.businessDomainID = businessDomainID
		capabilityID, err := valueobjects.NewCapabilityIDFromString(e.CapabilityID)
		if err != nil {
			return fmt.Errorf("%w: capability ID %q: %v", domain.ErrCorruptedEvent, e.CapabilityID, err)
		}
		a.capabilityID = capabilityID
		a.assignedAt = e.AssignedAt
	case events.CapabilityUnassignedFromDomain:
	}
	return nil
}

func (a *BusinessDomainAssignment) BusinessDomainID() valueobjects.BusinessDomainID {
	return a.businessDomainID
}

func (a *BusinessDomainAssignment) CapabilityID() valueobjects.CapabilityID {
	return a.capabilityID
}

func (a *BusinessDomainAssignment) AssignedAt() time.Time {
	return a.assignedAt
}
