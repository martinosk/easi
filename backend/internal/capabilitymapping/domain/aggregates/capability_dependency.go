package aggregates

import (
	"errors"
	"time"

	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrCannotCreateSelfDependency = errors.New("cannot create self-dependency")
)

type CapabilityDependency struct {
	domain.AggregateRoot
	sourceCapabilityID valueobjects.CapabilityID
	targetCapabilityID valueobjects.CapabilityID
	dependencyType     valueobjects.DependencyType
	description        valueobjects.Description
	createdAt          time.Time
}

func NewCapabilityDependency(
	sourceCapabilityID valueobjects.CapabilityID,
	targetCapabilityID valueobjects.CapabilityID,
	dependencyType valueobjects.DependencyType,
	description valueobjects.Description,
) (*CapabilityDependency, error) {
	if sourceCapabilityID.Value() == targetCapabilityID.Value() {
		return nil, ErrCannotCreateSelfDependency
	}

	aggregate := &CapabilityDependency{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewCapabilityDependencyCreated(
		aggregate.ID(),
		sourceCapabilityID.Value(),
		targetCapabilityID.Value(),
		dependencyType.Value(),
		description.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadCapabilityDependencyFromHistory(events []domain.DomainEvent) (*CapabilityDependency, error) {
	aggregate := &CapabilityDependency{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (cd *CapabilityDependency) Delete() error {
	event := events.NewCapabilityDependencyDeleted(cd.ID())

	cd.apply(event)
	cd.RaiseEvent(event)

	return nil
}

func (cd *CapabilityDependency) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.CapabilityDependencyCreated:
		cd.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		cd.sourceCapabilityID, _ = valueobjects.NewCapabilityIDFromString(e.SourceCapabilityID)
		cd.targetCapabilityID, _ = valueobjects.NewCapabilityIDFromString(e.TargetCapabilityID)
		cd.dependencyType, _ = valueobjects.NewDependencyType(e.DependencyType)
		cd.description = valueobjects.NewDescription(e.Description)
		cd.createdAt = e.CreatedAt
	case events.CapabilityDependencyDeleted:
	}
}

func (cd *CapabilityDependency) SourceCapabilityID() valueobjects.CapabilityID {
	return cd.sourceCapabilityID
}

func (cd *CapabilityDependency) TargetCapabilityID() valueobjects.CapabilityID {
	return cd.targetCapabilityID
}

func (cd *CapabilityDependency) DependencyType() valueobjects.DependencyType {
	return cd.dependencyType
}

func (cd *CapabilityDependency) Description() valueobjects.Description {
	return cd.description
}

func (cd *CapabilityDependency) CreatedAt() time.Time {
	return cd.createdAt
}
