package repositories

import (
	"errors"

	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/events"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrOnePagerFactsNotFound = errors.New("one-pager facts not found")

type OnePagerFactsRepository struct {
	*repository.EventSourcedRepository[*aggregates.OnePagerFacts]
}

func NewOnePagerFactsRepository(eventStore eventstore.EventStore) *OnePagerFactsRepository {
	return &OnePagerFactsRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			onePagerFactsEventDeserializers,
			aggregates.LoadOnePagerFactsFromHistory,
			ErrOnePagerFactsNotFound,
		),
	}
}

var onePagerFactsEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"FieldValueRecorded":    repository.JSONDeserializer[events.FieldValueRecorded],
		"FieldValueCleared":     repository.JSONDeserializer[events.FieldValueCleared],
		"OnePagerFactsArchived": repository.JSONDeserializer[events.OnePagerFactsArchived],
	},
)
