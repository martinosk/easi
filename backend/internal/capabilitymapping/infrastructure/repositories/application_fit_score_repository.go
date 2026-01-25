package repositories

import (
	"errors"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrApplicationFitScoreNotFound = errors.New("application fit score not found")

type ApplicationFitScoreRepository struct {
	*repository.EventSourcedRepository[*aggregates.ApplicationFitScore]
}

func NewApplicationFitScoreRepository(eventStore eventstore.EventStore) *ApplicationFitScoreRepository {
	return &ApplicationFitScoreRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			applicationFitScoreEventDeserializers,
			aggregates.LoadApplicationFitScoreFromHistory,
			ErrApplicationFitScoreNotFound,
		),
	}
}

var applicationFitScoreEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"ApplicationFitScoreSet":     repository.JSONDeserializer[events.ApplicationFitScoreSet],
		"ApplicationFitScoreUpdated": repository.JSONDeserializer[events.ApplicationFitScoreUpdated],
		"ApplicationFitScoreRemoved": repository.JSONDeserializer[events.ApplicationFitScoreRemoved],
	},
)
