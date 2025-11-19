package aggregates

import (
	"errors"
	"time"

	"easi/backend/internal/architectureviews/domain/events"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/shared/domain"
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

// ArchitectureView represents an architecture view aggregate
type ArchitectureView struct {
	domain.AggregateRoot
	name            valueobjects.ViewName
	description     string
	components      map[string]bool // componentID -> exists (membership set)
	isDefault       bool
	isDeleted       bool
	createdAt       time.Time
}

// NewArchitectureView creates a new architecture view
func NewArchitectureView(name valueobjects.ViewName, description string, isDefault bool) (*ArchitectureView, error) {
	aggregate := &ArchitectureView{
		AggregateRoot: domain.NewAggregateRoot(),
		components:    make(map[string]bool),
	}

	// Raise creation event
	viewCreatedEvent := events.NewViewCreated(
		aggregate.ID(),
		name.Value(),
		description,
	)

	aggregate.apply(viewCreatedEvent)
	aggregate.RaiseEvent(viewCreatedEvent)

	// If this is the default view, raise the default view changed event
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

	if v.components[componentID] {
		return ErrComponentAlreadyInView
	}

	v.applyAndRaise(events.NewComponentAddedToView(v.ID(), componentID, 0, 0))
	return nil
}

func (v *ArchitectureView) RemoveComponent(componentID string) error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if !v.components[componentID] {
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
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if v.isDefault {
		return nil
	}

	v.applyAndRaise(events.NewDefaultViewChanged(v.ID(), true))
	return nil
}

func (v *ArchitectureView) UnsetAsDefault() error {
	if err := v.checkNotDeleted(); err != nil {
		return err
	}

	if !v.isDefault {
		return nil
	}

	v.applyAndRaise(events.NewDefaultViewChanged(v.ID(), false))
	return nil
}


// LoadFromHistory reconstructs the aggregate from events
func LoadArchitectureViewFromHistory(events []domain.DomainEvent) (*ArchitectureView, error) {
	aggregate := &ArchitectureView{
		AggregateRoot: domain.NewAggregateRoot(),
		components:    make(map[string]bool),
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
	case events.ComponentPositionUpdated:
	case events.ComponentRemovedFromView:
		v.applyComponentRemoved(e)
	case events.ViewRenamed:
		v.applyViewRenamed(e)
	case events.ViewDeleted:
		v.applyViewDeleted()
	case events.DefaultViewChanged:
		v.applyDefaultViewChanged(e)
	case events.ViewEdgeTypeUpdated:
	case events.ViewLayoutDirectionUpdated:
	}
}

func (v *ArchitectureView) applyViewCreated(e events.ViewCreated) {
	v.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	v.name, _ = valueobjects.NewViewName(e.Name)
	v.description = e.Description
	v.createdAt = e.CreatedAt
}

func (v *ArchitectureView) applyComponentAdded(e events.ComponentAddedToView) {
	v.components[e.ComponentID] = true
}

func (v *ArchitectureView) applyComponentRemoved(e events.ComponentRemovedFromView) {
	delete(v.components, e.ComponentID)
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

// Name returns the view name
func (v *ArchitectureView) Name() valueobjects.ViewName {
	return v.name
}

// Description returns the view description
func (v *ArchitectureView) Description() string {
	return v.description
}

// Components returns all component IDs in the view (membership set)
func (v *ArchitectureView) Components() []string {
	componentIDs := make([]string, 0, len(v.components))
	for componentID := range v.components {
		componentIDs = append(componentIDs, componentID)
	}
	return componentIDs
}

// HasComponent checks if a component is in the view
func (v *ArchitectureView) HasComponent(componentID string) bool {
	return v.components[componentID]
}

// CreatedAt returns when the view was created
func (v *ArchitectureView) CreatedAt() time.Time {
	return v.createdAt
}

// IsDefault returns whether this is the default view
func (v *ArchitectureView) IsDefault() bool {
	return v.isDefault
}

// IsDeleted returns whether this view has been deleted
func (v *ArchitectureView) IsDeleted() bool {
	return v.isDeleted
}
