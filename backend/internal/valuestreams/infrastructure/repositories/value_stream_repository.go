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
		"ValueStreamCreated":                repository.JSONDeserializer[events.ValueStreamCreated],
		"ValueStreamUpdated":                repository.JSONDeserializer[events.ValueStreamUpdated],
		"ValueStreamDeleted":                repository.JSONDeserializer[events.ValueStreamDeleted],
		"ValueStreamStageAdded":             repository.JSONDeserializer[events.ValueStreamStageAdded],
		"ValueStreamStageUpdated":           repository.JSONDeserializer[events.ValueStreamStageUpdated],
		"ValueStreamStageRemoved":           repository.JSONDeserializer[events.ValueStreamStageRemoved],
		"ValueStreamStagesReordered":        repository.JSONDeserializer[events.ValueStreamStagesReordered],
		"ValueStreamStageCapabilityAdded":   repository.JSONDeserializer[events.ValueStreamStageCapabilityAdded],
		"ValueStreamStageCapabilityRemoved": repository.JSONDeserializer[events.ValueStreamStageCapabilityRemoved],
	},
)
