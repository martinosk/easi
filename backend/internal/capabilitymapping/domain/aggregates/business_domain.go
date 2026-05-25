package aggregates

import (
	"fmt"
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

	aggregate.raise(event)

	return aggregate, nil
}

func LoadBusinessDomainFromHistory(events []domain.DomainEvent) (*BusinessDomain, error) {
	aggregate := &BusinessDomain{
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

func (b *BusinessDomain) Update(name valueobjects.DomainName, description valueobjects.Description, domainArchitectID string) error {
	event := events.NewBusinessDomainUpdated(
		b.ID(),
		name.Value(),
		description.Value(),
		domainArchitectID,
	)

	b.raise(event)

	return nil
}

func (b *BusinessDomain) Delete() error {
	event := events.NewBusinessDomainDeleted(b.ID())

	b.raise(event)

	return nil
}

func (b *BusinessDomain) raise(event domain.DomainEvent) {
	if err := b.apply(event); err != nil {
		panic(fmt.Sprintf("capabilitymapping: in-process apply failed: %v", err))
	}
	b.RaiseEvent(event)
}

func (b *BusinessDomain) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.BusinessDomainCreated:
		b.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		name, err := valueobjects.NewDomainName(e.Name)
		if err != nil {
			return fmt.Errorf("%w: domain name %q: %v", domain.ErrCorruptedEvent, e.Name, err)
		}
		b.name = name
		b.description = valueobjects.MustNewDescription(e.Description)
		b.domainArchitectID = e.DomainArchitectID
		b.createdAt = e.CreatedAt
	case events.BusinessDomainUpdated:
		name, err := valueobjects.NewDomainName(e.Name)
		if err != nil {
			return fmt.Errorf("%w: domain name %q: %v", domain.ErrCorruptedEvent, e.Name, err)
		}
		b.name = name
		b.description = valueobjects.MustNewDescription(e.Description)
		b.domainArchitectID = e.DomainArchitectID
	case events.BusinessDomainDeleted:
	}
	return nil
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
