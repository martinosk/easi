package repositories

import (
	"errors"

	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/events"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrRealizationRolesNotFound = errors.New("realization roles not found")

type RealizationRolesRepository struct {
	*repository.EventSourcedRepository[*aggregates.RealizationRoles]
}

func NewRealizationRolesRepository(eventStore eventstore.EventStore) *RealizationRolesRepository {
	return &RealizationRolesRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			realizationRolesEventDeserializers,
			aggregates.LoadRealizationRolesFromHistory,
			ErrRealizationRolesNotFound,
		),
	}
}

var realizationRolesEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		pl.RealizationRoleAssigned: repository.JSONDeserializer[events.RealizationRoleAssigned],
		pl.RealizationRoleCleared:  repository.JSONDeserializer[events.RealizationRoleCleared],
	},
)
