package aggregates

import (
	"errors"
	"time"

	"easi/backend/internal/architectureviews/domain/entities"
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
	name        valueobjects.ViewName
	description string
	components  map[string]entities.ViewComponent // componentID -> ViewComponent
	isDefault   bool
	isDeleted   bool
	createdAt   time.Time
}

// NewArchitectureView creates a new architecture view
func NewArchitectureView(name valueobjects.ViewName, description string, isDefault bool) (*ArchitectureView, error) {
	aggregate := &ArchitectureView{
		AggregateRoot: domain.NewAggregateRoot(),
		components:    make(map[string]entities.ViewComponent),
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

// AddComponent adds a component to the view at a specific position
func (v *ArchitectureView) AddComponent(componentID string, position valueobjects.ComponentPosition) error {
	// Check if view is deleted
	if v.isDeleted {
		return ErrViewAlreadyDeleted
	}

	// Check if component already exists
	if _, exists := v.components[componentID]; exists {
		return ErrComponentAlreadyInView
	}

	// Raise event
	event := events.NewComponentAddedToView(
		v.ID(),
		componentID,
		position.X(),
		position.Y(),
	)

	v.apply(event)
	v.RaiseEvent(event)

	return nil
}

// UpdateComponentPosition updates the position of a component in the view
func (v *ArchitectureView) UpdateComponentPosition(componentID string, newPosition valueobjects.ComponentPosition) error {
	// Check if view is deleted
	if v.isDeleted {
		return ErrViewAlreadyDeleted
	}

	// Check if component exists
	if _, exists := v.components[componentID]; !exists {
		return ErrComponentNotFound
	}

	// Raise event
	event := events.NewComponentPositionUpdated(
		v.ID(),
		componentID,
		newPosition.X(),
		newPosition.Y(),
	)

	v.apply(event)
	v.RaiseEvent(event)

	return nil
}

// RemoveComponent removes a component from the view
func (v *ArchitectureView) RemoveComponent(componentID string) error {
	// Check if view is deleted
	if v.isDeleted {
		return ErrViewAlreadyDeleted
	}

	// Check if component exists
	if _, exists := v.components[componentID]; !exists {
		return ErrComponentNotFound
	}

	// Raise event
	event := events.NewComponentRemovedFromView(
		v.ID(),
		componentID,
	)

	v.apply(event)
	v.RaiseEvent(event)

	return nil
}

// Rename renames the view
func (v *ArchitectureView) Rename(newName valueobjects.ViewName) error {
	// Check if view is deleted
	if v.isDeleted {
		return ErrViewAlreadyDeleted
	}

	// Check if name is different
	if v.name.Value() == newName.Value() {
		return nil // No-op if name is the same
	}

	// Raise event
	event := events.NewViewRenamed(
		v.ID(),
		v.name.Value(),
		newName.Value(),
	)

	v.apply(event)
	v.RaiseEvent(event)

	return nil
}

// Delete marks the view as deleted
func (v *ArchitectureView) Delete() error {
	// Check if view is already deleted
	if v.isDeleted {
		return ErrViewAlreadyDeleted
	}

	// Cannot delete default view
	if v.isDefault {
		return ErrCannotDeleteDefaultView
	}

	// Raise event
	event := events.NewViewDeleted(v.ID())

	v.apply(event)
	v.RaiseEvent(event)

	return nil
}

// SetAsDefault sets this view as the default view
func (v *ArchitectureView) SetAsDefault() error {
	// Check if view is deleted
	if v.isDeleted {
		return ErrViewAlreadyDeleted
	}

	// If already default, no-op
	if v.isDefault {
		return nil
	}

	// Raise event
	event := events.NewDefaultViewChanged(v.ID(), true)

	v.apply(event)
	v.RaiseEvent(event)

	return nil
}

// UnsetAsDefault unsets this view as the default view
func (v *ArchitectureView) UnsetAsDefault() error {
	// Check if view is deleted
	if v.isDeleted {
		return ErrViewAlreadyDeleted
	}

	// If not default, no-op
	if !v.isDefault {
		return nil
	}

	// Raise event
	event := events.NewDefaultViewChanged(v.ID(), false)

	v.apply(event)
	v.RaiseEvent(event)

	return nil
}

// LoadFromHistory reconstructs the aggregate from events
func LoadArchitectureViewFromHistory(events []domain.DomainEvent) (*ArchitectureView, error) {
	aggregate := &ArchitectureView{
		AggregateRoot: domain.NewAggregateRoot(),
		components:    make(map[string]entities.ViewComponent),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

// apply applies an event to the aggregate
// Note: This method should NOT increment the version - that's handled by LoadFromHistory or RaiseEvent
func (v *ArchitectureView) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.ViewCreated:
		v.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		v.name, _ = valueobjects.NewViewName(e.Name)
		v.description = e.Description
		v.createdAt = e.CreatedAt

	case events.ComponentAddedToView:
		position := valueobjects.NewComponentPosition(e.X, e.Y)
		viewComponent := entities.NewViewComponent(e.ComponentID, position)
		v.components[e.ComponentID] = viewComponent

	case events.ComponentPositionUpdated:
		if viewComponent, exists := v.components[e.ComponentID]; exists {
			newPosition := valueobjects.NewComponentPosition(e.X, e.Y)
			v.components[e.ComponentID] = viewComponent.UpdatePosition(newPosition)
		}

	case events.ComponentRemovedFromView:
		delete(v.components, e.ComponentID)

	case events.ViewRenamed:
		v.name, _ = valueobjects.NewViewName(e.NewName)

	case events.ViewDeleted:
		v.isDeleted = true

	case events.DefaultViewChanged:
		v.isDefault = e.IsDefault
	}
}

// Name returns the view name
func (v *ArchitectureView) Name() valueobjects.ViewName {
	return v.name
}

// Description returns the view description
func (v *ArchitectureView) Description() string {
	return v.description
}

// Components returns all components in the view
func (v *ArchitectureView) Components() map[string]entities.ViewComponent {
	// Return copy to maintain immutability
	componentsCopy := make(map[string]entities.ViewComponent)
	for k, v := range v.components {
		componentsCopy[k] = v
	}
	return componentsCopy
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
