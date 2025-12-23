package aggregates

import (
	"errors"
	"time"

	"easi/backend/internal/architectureviews/domain/events"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/shared/eventsourcing"
)

var (
	// ErrComponentNotFound is returned when trying to update position of non-existent component
	ErrComponentNotFound = errors.New("component not found in view")

	// ErrComponentAlreadyInView is returned when trying to add a component that's already in the view
	ErrComponentAlreadyInView = errors.New("component already exists in view")

	// ErrCannotDeleteDefaultView is returned when trying to delete the default view
	ErrCannotDeleteDefaultView = errors.New("cannot delete the default view")

	// ErrViewAlreadyDeleted is returned when trying to perform operations on a deleted view
	ErrViewAlreadyDeleted = errors.New("view has been deleted")
)

type ArchitectureView struct {
	domain.AggregateRoot
	name        valueobjects.ViewName
	description valueobjects.ViewDescription
	components  valueobjects.ComponentMembership
	isDefault   bool
	isDeleted   bool
	createdAt   time.Time
}

func NewArchitectureView(name valueobjects.ViewName, description string, isDefault bool) (*ArchitectureView, error) {
	aggregate := &ArchitectureView{
		AggregateRoot: domain.NewAggregateRoot(),
		components:    valueobjects.NewComponentMembership(),
	}

	viewCreatedEvent := events.NewViewCreated(
		aggregate.ID(),
		name.Value(),
		description,
	)

	aggregate.apply(viewCreatedEvent)
	aggregate.RaiseEvent(viewCreatedEvent)

	if isDefault {
		defaultEvent := events.NewDefaultViewChanged(aggregate.ID(), true)
		aggregate.apply(defaultEvent)
		aggregate.RaiseEvent(defaultEvent)
	}

	return aggregate, nil
}

func (v *ArchitectureView) checkNotDeleted() error {
	if v.isDeleted {
		return ErrViewAlreadyDeleted
	}
	return nil
}

func (v *ArchitectureView) applyAndRaise(event domain.DomainEvent) {
	v.apply(event)
	v.RaiseEvent(event)
}

func (v *ArchitectureView) AddComponent(componentID string) error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if v.components.Contains(componentID) {
		return ErrComponentAlreadyInView
	}

	v.applyAndRaise(events.NewComponentAddedToView(v.ID(), componentID, 0, 0))
	return nil
}

func (v *ArchitectureView) RemoveComponent(componentID string) error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if !v.components.Contains(componentID) {
		return ErrComponentNotFound
	}

	v.applyAndRaise(events.NewComponentRemovedFromView(v.ID(), componentID))
	return nil
}

func (v *ArchitectureView) Rename(newName valueobjects.ViewName) error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if v.name.Value() == newName.Value() {
		return nil
	}

	v.applyAndRaise(events.NewViewRenamed(v.ID(), v.name.Value(), newName.Value()))
	return nil
}

func (v *ArchitectureView) Delete() error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if v.isDefault {
		return ErrCannotDeleteDefaultView
	}

	v.applyAndRaise(events.NewViewDeleted(v.ID()))
	return nil
}

func (v *ArchitectureView) SetAsDefault() error {
	return v.setDefaultStatus(true)
}

func (v *ArchitectureView) UnsetAsDefault() error {
	return v.setDefaultStatus(false)
}

func (v *ArchitectureView) setDefaultStatus(makeDefault bool) error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if v.isDefault == makeDefault {
		return nil
	}

	v.applyAndRaise(events.NewDefaultViewChanged(v.ID(), makeDefault))
	return nil
}

func LoadArchitectureViewFromHistory(events []domain.DomainEvent) (*ArchitectureView, error) {
	aggregate := &ArchitectureView{
		AggregateRoot: domain.NewAggregateRoot(),
		components:    valueobjects.NewComponentMembership(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (v *ArchitectureView) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.ViewCreated:
		v.applyViewCreated(e)
	case events.ComponentAddedToView:
		v.applyComponentAdded(e)
	case events.ComponentRemovedFromView:
		v.applyComponentRemoved(e)
	case events.ViewRenamed:
		v.applyViewRenamed(e)
	case events.ViewDeleted:
		v.applyViewDeleted()
	case events.DefaultViewChanged:
		v.applyDefaultViewChanged(e)
	case events.ComponentPositionUpdated, events.ViewEdgeTypeUpdated, events.ViewLayoutDirectionUpdated:
	}
}

func (v *ArchitectureView) applyViewCreated(e events.ViewCreated) {
	v.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	v.name, _ = valueobjects.NewViewName(e.Name)
	v.description = valueobjects.NewViewDescription(e.Description)
	v.createdAt = e.CreatedAt
}

func (v *ArchitectureView) applyComponentAdded(e events.ComponentAddedToView) {
	v.components.Add(e.ComponentID)
}

func (v *ArchitectureView) applyComponentRemoved(e events.ComponentRemovedFromView) {
	v.components.Remove(e.ComponentID)
}

func (v *ArchitectureView) applyViewRenamed(e events.ViewRenamed) {
	v.name, _ = valueobjects.NewViewName(e.NewName)
}

func (v *ArchitectureView) applyViewDeleted() {
	v.isDeleted = true
}

func (v *ArchitectureView) applyDefaultViewChanged(e events.DefaultViewChanged) {
	v.isDefault = e.IsDefault
}

func (v *ArchitectureView) Name() valueobjects.ViewName {
	return v.name
}

func (v *ArchitectureView) Description() string {
	return v.description.Value()
}

func (v *ArchitectureView) Components() []string {
	return v.components.GetAll()
}

func (v *ArchitectureView) HasComponent(componentID string) bool {
	return v.components.Contains(componentID)
}

func (v *ArchitectureView) CreatedAt() time.Time {
	return v.createdAt
}

func (v *ArchitectureView) IsDefault() bool {
	return v.isDefault
}

func (v *ArchitectureView) IsDeleted() bool {
	return v.isDeleted
}
