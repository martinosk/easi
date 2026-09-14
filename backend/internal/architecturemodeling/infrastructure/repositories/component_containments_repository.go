package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrComponentContainmentsNotFound = errors.New("component containments not found")

type ComponentContainmentsRepository struct {
	*repository.EventSourcedRepository[*aggregates.ComponentContainments]
}

func NewComponentContainmentsRepository(eventStore eventstore.EventStore) *ComponentContainmentsRepository {
	return &ComponentContainmentsRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			componentContainmentsEventDeserializers,
			aggregates.LoadComponentContainmentsFromHistory,
			ErrComponentContainmentsNotFound,
		),
	}
}

var componentContainmentsEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		archPL.ComponentAttached: repository.JSONDeserializer[events.ComponentAttached],
		archPL.ComponentDetached: repository.JSONDeserializer[events.ComponentDetached],
	},
)
