package repositories

import (
	"errors"

	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrValueStreamNotFound = errors.New("value stream not found")

type ValueStreamRepository struct {
	*repository.EventSourcedRepository[*aggregates.ValueStream]
}

func NewValueStreamRepository(eventStore eventstore.EventStore) *ValueStreamRepository {
	return &ValueStreamRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			valueStreamEventDeserializers,
			aggregates.LoadValueStreamFromHistory,
			ErrValueStreamNotFound,
		),
	}
}

var valueStreamEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"ValueStreamCreated": repository.JSONDeserializer[events.ValueStreamCreated],
		"ValueStreamUpdated": repository.JSONDeserializer[events.ValueStreamUpdated],
		"ValueStreamDeleted": repository.JSONDeserializer[events.ValueStreamDeleted],
	},
)
