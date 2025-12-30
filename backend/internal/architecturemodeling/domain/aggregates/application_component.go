package aggregates

import (
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

// ApplicationComponent represents an application component aggregate
type ApplicationComponent struct {
	domain.AggregateRoot
	name        valueobjects.ComponentName
	description valueobjects.Description
	createdAt   time.Time
	isDeleted   bool
}

func NewApplicationComponent(name valueobjects.ComponentName, description valueobjects.Description) (*ApplicationComponent, error) {
	aggregate := &ApplicationComponent{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewApplicationComponentCreated(
		aggregate.ID(),
		name.Value(),
		description.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadApplicationComponentFromHistory(events []domain.DomainEvent) (*ApplicationComponent, error) {
	aggregate := &ApplicationComponent{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (a *ApplicationComponent) Update(name valueobjects.ComponentName, description valueobjects.Description) error {
	event := events.NewApplicationComponentUpdated(
		a.ID(),
		name.Value(),
		description.Value(),
	)

	a.apply(event)
	a.RaiseEvent(event)

	return nil
}

func (a *ApplicationComponent) Delete() error {
	event := events.NewApplicationComponentDeleted(
		a.ID(),
		a.name.Value(),
	)

	a.apply(event)
	a.RaiseEvent(event)

	return nil
}

// Note: This method should NOT increment the version - that's handled by LoadFromHistory or RaiseEvent
func (a *ApplicationComponent) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.ApplicationComponentCreated:
		a.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		a.name, _ = valueobjects.NewComponentName(e.Name)
		a.description = valueobjects.MustNewDescription(e.Description)
		a.createdAt = e.CreatedAt
	case events.ApplicationComponentUpdated:
		a.name, _ = valueobjects.NewComponentName(e.Name)
		a.description = valueobjects.MustNewDescription(e.Description)
	case events.ApplicationComponentDeleted:
		a.isDeleted = true
	}
}

func (a *ApplicationComponent) Name() valueobjects.ComponentName {
	return a.name
}

func (a *ApplicationComponent) Description() valueobjects.Description {
	return a.description
}

func (a *ApplicationComponent) CreatedAt() time.Time {
	return a.createdAt
}

func (a *ApplicationComponent) IsDeleted() bool {
	return a.isDeleted
}
