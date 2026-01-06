package aggregates

import (
	"errors"
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
	case events.ViewVisibilityChanged:
		v.applyViewVisibilityChanged(e)
	}
}

func (v *ArchitectureView) applyViewCreated(e events.ViewCreated) {
	v.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	v.name, _ = valueobjects.NewViewName(e.Name)
	v.description = valueobjects.NewViewDescription(e.Description)
	v.visibility = valueobjects.NewViewVisibility(e.IsPrivate)
	v.owner, _ = valueobjects.NewViewOwner(e.OwnerUserID, e.OwnerEmail)
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

	v.applyAndRaise(events.NewViewVisibilityChanged(v.ID(), false, newOwner.UserID(), newOwner.Email()))
	return nil
}

func (v *ArchitectureView) MakePrivate() error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if v.visibility.IsPrivate() {
		return ErrViewAlreadyPrivate
	}

	v.applyAndRaise(events.NewViewVisibilityChanged(v.ID(), true, v.owner.UserID(), v.owner.Email()))
	return nil
}

func (v *ArchitectureView) applyViewVisibilityChanged(e events.ViewVisibilityChanged) {
	v.visibility = valueobjects.NewViewVisibility(e.IsPrivate)
	v.owner, _ = valueobjects.NewViewOwner(e.OwnerUserID, e.OwnerEmail)
}
