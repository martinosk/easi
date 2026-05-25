package aggregates

import (
	"errors"
	"fmt"
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

	aggregate.raise(event)

	return aggregate, nil
}

func LoadCapabilityDependencyFromHistory(events []domain.DomainEvent) (*CapabilityDependency, error) {
	aggregate := &CapabilityDependency{
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

func (cd *CapabilityDependency) Delete() error {
	event := events.NewCapabilityDependencyDeleted(cd.ID())

	cd.raise(event)

	return nil
}

func (cd *CapabilityDependency) raise(event domain.DomainEvent) {
	if err := cd.apply(event); err != nil {
		panic(fmt.Sprintf("capabilitymapping: in-process apply failed: %v", err))
	}
	cd.RaiseEvent(event)
}

func (cd *CapabilityDependency) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.CapabilityDependencyCreated:
		cd.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		sourceCapabilityID, err := valueobjects.NewCapabilityIDFromString(e.SourceCapabilityID)
		if err != nil {
			return fmt.Errorf("%w: source capability ID %q: %v", domain.ErrCorruptedEvent, e.SourceCapabilityID, err)
		}
		cd.sourceCapabilityID = sourceCapabilityID
		targetCapabilityID, err := valueobjects.NewCapabilityIDFromString(e.TargetCapabilityID)
		if err != nil {
			return fmt.Errorf("%w: target capability ID %q: %v", domain.ErrCorruptedEvent, e.TargetCapabilityID, err)
		}
		cd.targetCapabilityID = targetCapabilityID
		dependencyType, err := valueobjects.NewDependencyType(e.DependencyType)
		if err != nil {
			return fmt.Errorf("%w: dependency type %q: %v", domain.ErrCorruptedEvent, e.DependencyType, err)
		}
		cd.dependencyType = dependencyType
		cd.description = valueobjects.MustNewDescription(e.Description)
		cd.createdAt = e.CreatedAt
	case events.CapabilityDependencyDeleted:
	}
	return nil
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
