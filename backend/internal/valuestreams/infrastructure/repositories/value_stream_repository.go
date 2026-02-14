package repositories

import (
	"errors"

	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/events"
	"easi/backend/internal/valuestreams/publishedlanguage"
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
		publishedlanguage.ValueStreamCreated:                repository.JSONDeserializer[events.ValueStreamCreated],
		publishedlanguage.ValueStreamUpdated:                repository.JSONDeserializer[events.ValueStreamUpdated],
		publishedlanguage.ValueStreamDeleted:                repository.JSONDeserializer[events.ValueStreamDeleted],
		publishedlanguage.ValueStreamStageAdded:             repository.JSONDeserializer[events.ValueStreamStageAdded],
		publishedlanguage.ValueStreamStageUpdated:           repository.JSONDeserializer[events.ValueStreamStageUpdated],
		publishedlanguage.ValueStreamStageRemoved:           repository.JSONDeserializer[events.ValueStreamStageRemoved],
		publishedlanguage.ValueStreamStagesReordered:        repository.JSONDeserializer[events.ValueStreamStagesReordered],
		publishedlanguage.ValueStreamStageCapabilityAdded:   repository.JSONDeserializer[events.ValueStreamStageCapabilityAdded],
		publishedlanguage.ValueStreamStageCapabilityRemoved: repository.JSONDeserializer[events.ValueStreamStageCapabilityRemoved],
	},
)
