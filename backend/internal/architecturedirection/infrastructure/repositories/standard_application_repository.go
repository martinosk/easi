package repositories

import (
	"errors"

	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/events"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrStandardApplicationNotFound = errors.New("standard application not found")

type StandardApplicationRepository struct {
	*repository.EventSourcedRepository[*aggregates.StandardApplication]
}

func NewStandardApplicationRepository(eventStore eventstore.EventStore) *StandardApplicationRepository {
	return &StandardApplicationRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			standardApplicationEventDeserializers,
			aggregates.LoadStandardApplicationFromHistory,
			ErrStandardApplicationNotFound,
		),
	}
}

var standardApplicationEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		pl.StandardApplicationSet: repository.JSONDeserializer[events.StandardApplicationSet],
	},
)
