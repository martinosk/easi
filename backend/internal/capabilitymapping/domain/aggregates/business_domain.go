package aggregates

import (
	"time"

	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/domain"
)

type BusinessDomain struct {
	domain.AggregateRoot
	name        valueobjects.DomainName
	description valueobjects.Description
	createdAt   time.Time
}

func NewBusinessDomain(
	name valueobjects.DomainName,
	description valueobjects.Description,
) (*BusinessDomain, error) {
	aggregate := &BusinessDomain{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewBusinessDomainCreated(
		aggregate.ID(),
		name.Value(),
		description.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadBusinessDomainFromHistory(events []domain.DomainEvent) (*BusinessDomain, error) {
	aggregate := &BusinessDomain{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (b *BusinessDomain) Update(name valueobjects.DomainName, description valueobjects.Description) error {
	event := events.NewBusinessDomainUpdated(
		b.ID(),
		name.Value(),
		description.Value(),
	)

	b.apply(event)
	b.RaiseEvent(event)

	return nil
}

func (b *BusinessDomain) Delete() error {
	event := events.NewBusinessDomainDeleted(b.ID())

	b.apply(event)
	b.RaiseEvent(event)

	return nil
}

func (b *BusinessDomain) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.BusinessDomainCreated:
		b.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		b.name, _ = valueobjects.NewDomainName(e.Name)
		b.description = valueobjects.NewDescription(e.Description)
		b.createdAt = e.CreatedAt
	case events.BusinessDomainUpdated:
		b.name, _ = valueobjects.NewDomainName(e.Name)
		b.description = valueobjects.NewDescription(e.Description)
	case events.BusinessDomainDeleted:
	}
}

func (b *BusinessDomain) Name() valueobjects.DomainName {
	return b.name
}

func (b *BusinessDomain) Description() valueobjects.Description {
	return b.description
}

func (b *BusinessDomain) CreatedAt() time.Time {
	return b.createdAt
}
