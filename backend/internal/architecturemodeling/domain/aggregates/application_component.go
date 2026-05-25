package aggregates

import (
	"errors"
	"fmt"
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrDuplicateExpert = errors.New("this exact expert entry already exists on this component")
)

// ApplicationComponent represents an application component aggregate
type ApplicationComponent struct {
	domain.AggregateRoot
	name        valueobjects.ComponentName
	description valueobjects.Description
	createdAt   time.Time
	isDeleted   bool
	experts     []valueobjects.Expert
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

	if err := aggregate.apply(event); err != nil {
		return nil, err
	}
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadApplicationComponentFromHistory(events []domain.DomainEvent) (*ApplicationComponent, error) {
	aggregate := &ApplicationComponent{
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

func (a *ApplicationComponent) Update(name valueobjects.ComponentName, description valueobjects.Description) error {
	event := events.NewApplicationComponentUpdated(
		a.ID(),
		name.Value(),
		description.Value(),
	)

	if err := a.apply(event); err != nil {
		return err
	}
	a.RaiseEvent(event)

	return nil
}

func (a *ApplicationComponent) Delete() error {
	event := events.NewApplicationComponentDeleted(
		a.ID(),
		a.name.Value(),
	)

	if err := a.apply(event); err != nil {
		return err
	}
	a.RaiseEvent(event)

	return nil
}

func (a *ApplicationComponent) AddExpert(expert valueobjects.Expert) error {
	for _, existing := range a.experts {
		if existing.Equals(expert) {
			return ErrDuplicateExpert
		}
	}

	event := events.NewApplicationComponentExpertAdded(
		a.ID(),
		expert.Name().Value(),
		expert.Role().Value(),
		expert.Contact().Value(),
	)

	if err := a.apply(event); err != nil {
		return err
	}
	a.RaiseEvent(event)

	return nil
}

func (a *ApplicationComponent) RemoveExpert(expert valueobjects.Expert) error {
	event := events.NewApplicationComponentExpertRemoved(
		a.ID(),
		expert.Name().Value(),
		expert.Role().Value(),
		expert.Contact().Value(),
	)

	if err := a.apply(event); err != nil {
		return err
	}
	a.RaiseEvent(event)

	return nil
}

func (a *ApplicationComponent) Experts() []valueobjects.Expert {
	return a.experts
}

func (a *ApplicationComponent) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.ApplicationComponentCreated:
		return a.applyCreated(e)
	case events.ApplicationComponentUpdated:
		return a.applyUpdated(e)
	case events.ApplicationComponentDeleted:
		a.isDeleted = true
	case events.ApplicationComponentExpertAdded:
		return a.applyExpertAdded(e)
	case events.ApplicationComponentExpertRemoved:
		a.experts = removeExpert(a.experts, e.ExpertName, e.ExpertRole, e.ContactInfo)
	}
	return nil
}

func (a *ApplicationComponent) applyCreated(e events.ApplicationComponentCreated) error {
	name, err := valueobjects.NewComponentName(e.Name)
	if err != nil {
		return fmt.Errorf("%w: component name %q: %v", domain.ErrCorruptedEvent, e.Name, err)
	}
	description, err := valueobjects.NewDescription(e.Description)
	if err != nil {
		return fmt.Errorf("%w: description: %v", domain.ErrCorruptedEvent, err)
	}
	a.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	a.name = name
	a.description = description
	a.createdAt = e.CreatedAt
	return nil
}

func (a *ApplicationComponent) applyUpdated(e events.ApplicationComponentUpdated) error {
	name, err := valueobjects.NewComponentName(e.Name)
	if err != nil {
		return fmt.Errorf("%w: component name %q: %v", domain.ErrCorruptedEvent, e.Name, err)
	}
	description, err := valueobjects.NewDescription(e.Description)
	if err != nil {
		return fmt.Errorf("%w: description: %v", domain.ErrCorruptedEvent, err)
	}
	a.name = name
	a.description = description
	return nil
}

func (a *ApplicationComponent) applyExpertAdded(e events.ApplicationComponentExpertAdded) error {
	expert, err := valueobjects.NewExpert(e.ExpertName, e.ExpertRole, e.ContactInfo, e.AddedAt)
	if err != nil {
		return fmt.Errorf("%w: expert: %v", domain.ErrCorruptedEvent, err)
	}
	a.experts = append(a.experts, expert)
	return nil
}

func removeExpert(experts []valueobjects.Expert, name, role, contact string) []valueobjects.Expert {
	result := make([]valueobjects.Expert, 0, len(experts))
	for _, expert := range experts {
		if !expert.MatchesValues(name, role, contact) {
			result = append(result, expert)
		}
	}
	return result
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
