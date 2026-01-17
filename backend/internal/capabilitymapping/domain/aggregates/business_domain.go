package aggregates

import (
	"time"

	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type BusinessDomain struct {
	domain.AggregateRoot
	name              valueobjects.DomainName
	description       valueobjects.Description
	domainArchitectID string
	createdAt         time.Time
}

func NewBusinessDomain(
	name valueobjects.DomainName,
	description valueobjects.Description,
	domainArchitectID string,
) (*BusinessDomain, error) {
	id := valueobjects.NewBusinessDomainID()
	aggregate := &BusinessDomain{
		AggregateRoot: domain.NewAggregateRootWithID(id.Value()),
	}

	event := events.NewBusinessDomainCreated(
		aggregate.ID(),
		name.Value(),
		description.Value(),
		domainArchitectID,
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

func (b *BusinessDomain) Update(name valueobjects.DomainName, description valueobjects.Description, domainArchitectID string) error {
	event := events.NewBusinessDomainUpdated(
		b.ID(),
		name.Value(),
		description.Value(),
		domainArchitectID,
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
		b.description = valueobjects.MustNewDescription(e.Description)
		b.domainArchitectID = e.DomainArchitectID
		b.createdAt = e.CreatedAt
	case events.BusinessDomainUpdated:
		b.name, _ = valueobjects.NewDomainName(e.Name)
		b.description = valueobjects.MustNewDescription(e.Description)
		b.domainArchitectID = e.DomainArchitectID
	case events.BusinessDomainDeleted:
	}
}

func (b *BusinessDomain) Name() valueobjects.DomainName {
	return b.name
}

func (b *BusinessDomain) Description() valueobjects.Description {
	return b.description
}

func (b *BusinessDomain) DomainArchitectID() string {
	return b.domainArchitectID
}

func (b *BusinessDomain) CreatedAt() time.Time {
	return b.createdAt
}
