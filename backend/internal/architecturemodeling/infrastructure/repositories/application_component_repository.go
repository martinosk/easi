package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrComponentNotFound = errors.New("component not found")

type ApplicationComponentRepository struct {
	*repository.EventSourcedRepository[*aggregates.ApplicationComponent]
}

func NewApplicationComponentRepository(eventStore eventstore.EventStore) *ApplicationComponentRepository {
	return &ApplicationComponentRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			componentEventDeserializers,
			aggregates.LoadApplicationComponentFromHistory,
			ErrComponentNotFound,
		),
	}
}

var componentEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"ApplicationComponentCreated":       repository.JSONDeserializer[events.ApplicationComponentCreated],
		"ApplicationComponentUpdated":       repository.JSONDeserializer[events.ApplicationComponentUpdated],
		"ApplicationComponentDeleted":       repository.JSONDeserializer[events.ApplicationComponentDeleted],
		"ApplicationComponentExpertAdded":   repository.JSONDeserializer[events.ApplicationComponentExpertAdded],
		"ApplicationComponentExpertRemoved": repository.JSONDeserializer[events.ApplicationComponentExpertRemoved],
	},
)
