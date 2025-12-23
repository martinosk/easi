package repository

import (
	"context"
	"errors"

	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/eventsourcing"
)

type LoadFromHistoryFunc[T domain.EventSourcedAggregate] func(events []domain.DomainEvent) (T, error)

type EventSourcedRepository[T domain.EventSourcedAggregate] struct {
	eventStore      eventstore.EventStore
	deserializers   EventDeserializers
	loadFromHistory LoadFromHistoryFunc[T]
	notFoundErr     error
}

func NewEventSourcedRepository[T domain.EventSourcedAggregate](
	eventStore eventstore.EventStore,
	deserializers EventDeserializers,
	loadFromHistory LoadFromHistoryFunc[T],
	notFoundErr error,
) *EventSourcedRepository[T] {
	if notFoundErr == nil {
		notFoundErr = ErrAggregateNotFound
	}
	return &EventSourcedRepository[T]{
		eventStore:      eventStore,
		deserializers:   deserializers,
		loadFromHistory: loadFromHistory,
		notFoundErr:     notFoundErr,
	}
}

func (r *EventSourcedRepository[T]) Save(ctx context.Context, aggregate T) error {
	uncommittedEvents := aggregate.GetUncommittedChanges()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	expectedVersion := aggregate.Version() - len(uncommittedEvents)

	err := r.eventStore.SaveEvents(ctx, aggregate.ID(), uncommittedEvents, expectedVersion)
	if err != nil {
		return err
	}

	aggregate.MarkChangesAsCommitted()
	return nil
}

func (r *EventSourcedRepository[T]) GetByID(ctx context.Context, id string) (T, error) {
	var zero T

	storedEvents, err := r.eventStore.GetEvents(ctx, id)
	if err != nil {
		return zero, err
	}

	if len(storedEvents) == 0 {
		return zero, r.notFoundErr
	}

	domainEvents := r.deserializers.Deserialize(storedEvents)

	return r.loadFromHistory(domainEvents)
}

func (r *EventSourcedRepository[T]) NotFoundError() error {
	return r.notFoundErr
}

func (r *EventSourcedRepository[T]) IsNotFoundError(err error) bool {
	return errors.Is(err, r.notFoundErr)
}
