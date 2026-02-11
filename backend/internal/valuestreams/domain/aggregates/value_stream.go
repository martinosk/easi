package aggregates

import (
	"time"

	"easi/backend/internal/valuestreams/domain/events"
	"easi/backend/internal/valuestreams/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ValueStream struct {
	domain.AggregateRoot
	name        valueobjects.ValueStreamName
	description valueobjects.Description
	createdAt   time.Time
}

func NewValueStream(
	name valueobjects.ValueStreamName,
	description valueobjects.Description,
) (*ValueStream, error) {
	id := valueobjects.NewValueStreamID()
	aggregate := &ValueStream{
		AggregateRoot: domain.NewAggregateRootWithID(id.Value()),
	}

	event := events.NewValueStreamCreated(
		aggregate.ID(),
		name.Value(),
		description.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadValueStreamFromHistory(events []domain.DomainEvent) (*ValueStream, error) {
	aggregate := &ValueStream{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (v *ValueStream) Update(name valueobjects.ValueStreamName, description valueobjects.Description) error {
	event := events.NewValueStreamUpdated(
		v.ID(),
		name.Value(),
		description.Value(),
	)

	v.apply(event)
	v.RaiseEvent(event)

	return nil
}

func (v *ValueStream) Delete() error {
	event := events.NewValueStreamDeleted(v.ID())

	v.apply(event)
	v.RaiseEvent(event)

	return nil
}

func (v *ValueStream) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.ValueStreamCreated:
		v.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		v.name, _ = valueobjects.NewValueStreamName(e.Name)
		v.description = valueobjects.MustNewDescription(e.Description)
		v.createdAt = e.CreatedAt
	case events.ValueStreamUpdated:
		v.name, _ = valueobjects.NewValueStreamName(e.Name)
		v.description = valueobjects.MustNewDescription(e.Description)
	case events.ValueStreamDeleted:
	}
}

func (v *ValueStream) Name() valueobjects.ValueStreamName {
	return v.name
}

func (v *ValueStream) Description() valueobjects.Description {
	return v.description
}

func (v *ValueStream) CreatedAt() time.Time {
	return v.createdAt
}
