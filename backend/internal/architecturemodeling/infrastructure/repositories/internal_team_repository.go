package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrInternalTeamNotFound = errors.New("internal team not found")

type InternalTeamRepository struct {
	*repository.EventSourcedRepository[*aggregates.InternalTeam]
}

func NewInternalTeamRepository(eventStore eventstore.EventStore) *InternalTeamRepository {
	return &InternalTeamRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			internalTeamEventDeserializers,
			aggregates.LoadInternalTeamFromHistory,
			ErrInternalTeamNotFound,
		),
	}
}

var internalTeamEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"InternalTeamCreated": repository.JSONDeserializer[events.InternalTeamCreated],
		"InternalTeamUpdated": repository.JSONDeserializer[events.InternalTeamUpdated],
		"InternalTeamDeleted": repository.JSONDeserializer[events.InternalTeamDeleted],
	},
)
