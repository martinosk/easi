package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrRelationNotFound = errors.New("relation not found")

type ComponentRelationRepository struct {
	*repository.EventSourcedRepository[*aggregates.ComponentRelation]
}

func NewComponentRelationRepository(eventStore eventstore.EventStore) *ComponentRelationRepository {
	return &ComponentRelationRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			relationEventDeserializers,
			aggregates.LoadComponentRelationFromHistory,
			ErrRelationNotFound,
		),
	}
}

var relationEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"ComponentRelationCreated": repository.JSONDeserializer[events.ComponentRelationCreated],
		"ComponentRelationUpdated": repository.JSONDeserializer[events.ComponentRelationUpdated],
	},
)
