package aggregates

import (
	"time"

	"easi/backend/internal/enterprisearchitecture/domain/events"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapability struct {
	domain.AggregateRoot
	name           valueobjects.EnterpriseCapabilityName
	description    valueobjects.Description
	category       valueobjects.Category
	targetMaturity *valueobjects.TargetMaturity
	active         bool
	createdAt      time.Time
}

func NewEnterpriseCapability(
	name valueobjects.EnterpriseCapabilityName,
	description valueobjects.Description,
	category valueobjects.Category,
) (*EnterpriseCapability, error) {
	id := valueobjects.NewEnterpriseCapabilityID()
	aggregate := &EnterpriseCapability{
		AggregateRoot: domain.NewAggregateRootWithID(id.Value()),
	}

	event := events.NewEnterpriseCapabilityCreated(
		aggregate.ID(),
		name.Value(),
		description.Value(),
		category.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadEnterpriseCapabilityFromHistory(eventHistory []domain.DomainEvent) (*EnterpriseCapability, error) {
	aggregate := &EnterpriseCapability{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(eventHistory, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (e *EnterpriseCapability) Update(
	name valueobjects.EnterpriseCapabilityName,
	description valueobjects.Description,
	category valueobjects.Category,
) error {
	event := events.NewEnterpriseCapabilityUpdated(
		e.ID(),
		name.Value(),
		description.Value(),
		category.Value(),
	)

	e.apply(event)
	e.RaiseEvent(event)

	return nil
}

func (e *EnterpriseCapability) Delete() error {
	event := events.NewEnterpriseCapabilityDeleted(e.ID())

	e.apply(event)
	e.RaiseEvent(event)

	return nil
}

func (e *EnterpriseCapability) SetTargetMaturity(targetMaturity valueobjects.TargetMaturity) error {
	event := events.NewEnterpriseCapabilityTargetMaturitySet(e.ID(), targetMaturity.Value())

	e.apply(event)
	e.RaiseEvent(event)

	return nil
}

func (e *EnterpriseCapability) apply(event domain.DomainEvent) {
	switch evt := event.(type) {
	case events.EnterpriseCapabilityCreated:
		e.AggregateRoot = domain.NewAggregateRootWithID(evt.ID)
		e.name = mustNewEnterpriseCapabilityName(evt.Name)
		e.description = valueobjects.MustNewDescription(evt.Description)
		e.category = mustNewCategory(evt.Category)
		e.active = evt.Active
		e.createdAt = evt.CreatedAt
	case events.EnterpriseCapabilityUpdated:
		e.name = mustNewEnterpriseCapabilityName(evt.Name)
		e.description = valueobjects.MustNewDescription(evt.Description)
		e.category = mustNewCategory(evt.Category)
	case events.EnterpriseCapabilityDeleted:
		e.active = false
	case events.EnterpriseCapabilityTargetMaturitySet:
		maturity := mustNewTargetMaturity(evt.TargetMaturity)
		e.targetMaturity = &maturity
	}
}

func mustNewEnterpriseCapabilityName(value string) valueobjects.EnterpriseCapabilityName {
	name, err := valueobjects.NewEnterpriseCapabilityName(value)
	if err != nil {
		panic("corrupted event store: invalid enterprise capability name: " + value)
	}
	return name
}

func mustNewCategory(value string) valueobjects.Category {
	cat, err := valueobjects.NewCategory(value)
	if err != nil {
		panic("corrupted event store: invalid category: " + value)
	}
	return cat
}

func mustNewTargetMaturity(value int) valueobjects.TargetMaturity {
	maturity, err := valueobjects.NewTargetMaturity(value)
	if err != nil {
		panic("corrupted event store: invalid target maturity")
	}
	return maturity
}

func (e *EnterpriseCapability) Name() valueobjects.EnterpriseCapabilityName {
	return e.name
}

func (e *EnterpriseCapability) Description() valueobjects.Description {
	return e.description
}

func (e *EnterpriseCapability) Category() valueobjects.Category {
	return e.category
}

func (e *EnterpriseCapability) IsActive() bool {
	return e.active
}

func (e *EnterpriseCapability) CreatedAt() time.Time {
	return e.createdAt
}

func (e *EnterpriseCapability) TargetMaturity() *valueobjects.TargetMaturity {
	return e.targetMaturity
}
