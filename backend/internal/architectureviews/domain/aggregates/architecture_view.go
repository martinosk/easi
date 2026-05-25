package aggregates

import (
	"errors"
	"fmt"
	"time"

	"easi/backend/internal/architectureviews/domain/events"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrComponentNotFound       = errors.New("component not found in view")
	ErrComponentAlreadyInView  = errors.New("component already exists in view")
	ErrCannotDeleteDefaultView = errors.New("cannot delete the default view")
	ErrViewAlreadyDeleted      = errors.New("view has been deleted")
	ErrOnlyOwnerCanMakePrivate = errors.New("only the owner can make a view private")
	ErrViewAlreadyPrivate      = errors.New("view is already private")
	ErrViewAlreadyPublic       = errors.New("view is already public")
)

type ArchitectureView struct {
	domain.AggregateRoot
	name        valueobjects.ViewName
	description valueobjects.ViewDescription
	components  valueobjects.ComponentMembership
	owner       valueobjects.ViewOwner
	visibility  valueobjects.ViewVisibility
	isDefault   bool
	isDeleted   bool
	createdAt   time.Time
}

const defaultViewVisibilityIsPrivate = true

func NewArchitectureView(name valueobjects.ViewName, description string, isDefault bool, owner valueobjects.ViewOwner) (*ArchitectureView, error) {
	aggregate := &ArchitectureView{
		AggregateRoot: domain.NewAggregateRoot(),
		components:    valueobjects.NewComponentMembership(),
	}

	viewCreatedEvent := events.NewViewCreated(
		aggregate.ID(),
		name.Value(),
		description,
		defaultViewVisibilityIsPrivate,
		owner.UserID(),
		owner.Email(),
	)

	if err := aggregate.apply(viewCreatedEvent); err != nil {
		return nil, err
	}
	aggregate.RaiseEvent(viewCreatedEvent)

	if isDefault {
		defaultEvent := events.NewDefaultViewChanged(aggregate.ID(), true)
		if err := aggregate.apply(defaultEvent); err != nil {
			return nil, err
		}
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

func (v *ArchitectureView) applyAndRaise(event domain.DomainEvent) error {
	if err := v.apply(event); err != nil {
		return err
	}
	v.RaiseEvent(event)
	return nil
}

func (v *ArchitectureView) AddComponent(componentID string) error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if v.components.Contains(componentID) {
		return ErrComponentAlreadyInView
	}

	return v.applyAndRaise(events.NewComponentAddedToView(v.ID(), componentID, 0, 0))
}

func (v *ArchitectureView) RemoveComponent(componentID string) error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if !v.components.Contains(componentID) {
		return ErrComponentNotFound
	}

	return v.applyAndRaise(events.NewComponentRemovedFromView(v.ID(), componentID))
}

func (v *ArchitectureView) Rename(newName valueobjects.ViewName) error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if v.name.Value() == newName.Value() {
		return nil
	}

	return v.applyAndRaise(events.NewViewRenamed(v.ID(), v.name.Value(), newName.Value()))
}

func (v *ArchitectureView) Delete() error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if v.isDefault {
		return ErrCannotDeleteDefaultView
	}

	return v.applyAndRaise(events.NewViewDeleted(v.ID()))
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

	return v.applyAndRaise(events.NewDefaultViewChanged(v.ID(), makeDefault))
}

func LoadArchitectureViewFromHistory(events []domain.DomainEvent) (*ArchitectureView, error) {
	aggregate := &ArchitectureView{
		AggregateRoot: domain.NewAggregateRoot(),
		components:    valueobjects.NewComponentMembership(),
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

func (v *ArchitectureView) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.ViewCreated:
		return v.applyViewCreated(e)
	case events.ComponentAddedToView:
		v.applyComponentAdded(e)
	case events.ComponentRemovedFromView:
		v.applyComponentRemoved(e)
	case events.ViewRenamed:
		return v.applyViewRenamed(e)
	case events.ViewDeleted:
		v.applyViewDeleted()
	case events.DefaultViewChanged:
		v.applyDefaultViewChanged(e)
	case events.ViewVisibilityChanged:
		return v.applyViewVisibilityChanged(e)
	}
	return nil
}

func (v *ArchitectureView) applyViewCreated(e events.ViewCreated) error {
	name, err := valueobjects.NewViewName(e.Name)
	if err != nil {
		return fmt.Errorf("%w: view name %q: %v", domain.ErrCorruptedEvent, e.Name, err)
	}
	owner, err := valueobjects.NewViewOwner(e.OwnerUserID, e.OwnerEmail)
	if err != nil {
		return fmt.Errorf("%w: view owner: %v", domain.ErrCorruptedEvent, err)
	}
	v.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	v.name = name
	v.description = valueobjects.NewViewDescription(e.Description)
	v.visibility = valueobjects.NewViewVisibility(e.IsPrivate)
	v.owner = owner
	v.createdAt = e.CreatedAt
	return nil
}

func (v *ArchitectureView) applyComponentAdded(e events.ComponentAddedToView) {
	v.components.Add(e.ComponentID)
}

func (v *ArchitectureView) applyComponentRemoved(e events.ComponentRemovedFromView) {
	v.components.Remove(e.ComponentID)
}

func (v *ArchitectureView) applyViewRenamed(e events.ViewRenamed) error {
	name, err := valueobjects.NewViewName(e.NewName)
	if err != nil {
		return fmt.Errorf("%w: view name %q: %v", domain.ErrCorruptedEvent, e.NewName, err)
	}
	v.name = name
	return nil
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

func (v *ArchitectureView) Owner() valueobjects.ViewOwner {
	return v.owner
}

func (v *ArchitectureView) IsPrivate() bool {
	return v.visibility.IsPrivate()
}

func (v *ArchitectureView) CanBeEditedBy(actorID string, hasViewsWritePermission bool) bool {
	if v.visibility.IsPrivate() {
		return v.owner.UserID() == actorID
	}
	return hasViewsWritePermission
}

func (v *ArchitectureView) SetOwner(owner valueobjects.ViewOwner) error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}
	v.owner = owner
	return nil
}

func (v *ArchitectureView) MakePublic(newOwner valueobjects.ViewOwner) error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if v.visibility.IsPublic() {
		return ErrViewAlreadyPublic
	}

	return v.applyAndRaise(events.NewViewVisibilityChanged(v.ID(), false, newOwner.UserID(), newOwner.Email()))
}

func (v *ArchitectureView) MakePrivate() error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if v.visibility.IsPrivate() {
		return ErrViewAlreadyPrivate
	}

	return v.applyAndRaise(events.NewViewVisibilityChanged(v.ID(), true, v.owner.UserID(), v.owner.Email()))
}

func (v *ArchitectureView) applyViewVisibilityChanged(e events.ViewVisibilityChanged) error {
	owner, err := valueobjects.NewViewOwner(e.OwnerUserID, e.OwnerEmail)
	if err != nil {
		return fmt.Errorf("%w: view owner: %v", domain.ErrCorruptedEvent, err)
	}
	v.visibility = valueobjects.NewViewVisibility(e.IsPrivate)
	v.owner = owner
	return nil
}
