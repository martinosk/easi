package aggregates

import (
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

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadBusinessDomainAssignmentFromHistory(events []domain.DomainEvent) (*BusinessDomainAssignment, error) {
	aggregate := &BusinessDomainAssignment{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (a *BusinessDomainAssignment) Unassign() error {
	event := events.NewCapabilityUnassignedFromDomain(
		a.ID(),
		a.businessDomainID.Value(),
		a.capabilityID.Value(),
	)

	a.apply(event)
	a.RaiseEvent(event)

	return nil
}

func (a *BusinessDomainAssignment) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.CapabilityAssignedToDomain:
		a.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		a.businessDomainID, _ = valueobjects.NewBusinessDomainIDFromString(e.BusinessDomainID)
		a.capabilityID, _ = valueobjects.NewCapabilityIDFromString(e.CapabilityID)
		a.assignedAt = e.AssignedAt
	case events.CapabilityUnassignedFromDomain:
	}
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
