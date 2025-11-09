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
)

// ArchitectureView represents an architecture view aggregate
type ArchitectureView struct {
	domain.AggregateRoot
	name        valueobjects.ViewName
	description string
	components  map[string]entities.ViewComponent // componentID -> ViewComponent
	createdAt   time.Time
}

// NewArchitectureView creates a new architecture view
func NewArchitectureView(name valueobjects.ViewName, description string) (*ArchitectureView, error) {
	aggregate := &ArchitectureView{
		AggregateRoot: domain.NewAggregateRoot(),
		components:    make(map[string]entities.ViewComponent),
	}

	// Raise creation event
	event := events.NewViewCreated(
		aggregate.ID(),
		name.Value(),
		description,
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

// AddComponent adds a component to the view at a specific position
func (v *ArchitectureView) AddComponent(componentID string, position valueobjects.ComponentPosition) error {
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
func (v *ArchitectureView) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.ViewCreated:
		v.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		v.name, _ = valueobjects.NewViewName(e.Name)
		v.description = e.Description
		v.createdAt = e.CreatedAt
		v.IncrementVersion()

	case events.ComponentAddedToView:
		position := valueobjects.NewComponentPosition(e.X, e.Y)
		viewComponent := entities.NewViewComponent(e.ComponentID, position)
		v.components[e.ComponentID] = viewComponent
		v.IncrementVersion()

	case events.ComponentPositionUpdated:
		if viewComponent, exists := v.components[e.ComponentID]; exists {
			newPosition := valueobjects.NewComponentPosition(e.X, e.Y)
			v.components[e.ComponentID] = viewComponent.UpdatePosition(newPosition)
		}
		v.IncrementVersion()
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
