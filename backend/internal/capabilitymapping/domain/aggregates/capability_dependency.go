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
	ErrCircularDependencyDetected = errors.New("circular dependency detected")
	ErrDuplicateDependencyExists  = errors.New("dependency already exists between these capabilities")
)

type ExistingDependency struct {
	SourceID valueobjects.CapabilityID
	TargetID valueobjects.CapabilityID
}

type CapabilityDependency struct {
	domain.AggregateRoot
	sourceCapabilityID valueobjects.CapabilityID
	targetCapabilityID valueobjects.CapabilityID
	dependencyType     valueobjects.DependencyType
	description        valueobjects.Description
	createdAt          time.Time
}

func ValidateNewDependency(
	sourceID valueobjects.CapabilityID,
	targetID valueobjects.CapabilityID,
	existingDeps []ExistingDependency,
) error {
	if sourceID.Value() == targetID.Value() {
		return ErrCannotCreateSelfDependency
	}

	for _, dep := range existingDeps {
		if dep.SourceID.Value() == sourceID.Value() && dep.TargetID.Value() == targetID.Value() {
			return ErrDuplicateDependencyExists
		}
	}

	if wouldCreateCycle(sourceID, targetID, existingDeps) {
		return ErrCircularDependencyDetected
	}

	return nil
}

func wouldCreateCycle(sourceID, targetID valueobjects.CapabilityID, existingDeps []ExistingDependency) bool {
	visited := make(map[string]bool)
	return canReach(targetID.Value(), sourceID.Value(), existingDeps, visited)
}

func canReach(from, to string, deps []ExistingDependency, visited map[string]bool) bool {
	if from == to {
		return true
	}
	if visited[from] {
		return false
	}
	visited[from] = true

	for _, dep := range deps {
		if dep.SourceID.Value() == from {
			if canReach(dep.TargetID.Value(), to, deps, visited) {
				return true
			}
		}
	}
	return false
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
		cd.description = valueobjects.MustNewDescription(e.Description)
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
