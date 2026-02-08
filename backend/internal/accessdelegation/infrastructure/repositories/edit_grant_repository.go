package repositories

import (
	"errors"

	"easi/backend/internal/accessdelegation/domain/aggregates"
	"easi/backend/internal/accessdelegation/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var (
	ErrEditGrantNotFound = errors.New("edit grant not found")
)

type EditGrantRepository struct {
	*repository.EventSourcedRepository[*aggregates.EditGrant]
}

func NewEditGrantRepository(eventStore eventstore.EventStore) *EditGrantRepository {
	return &EditGrantRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			editGrantEventDeserializers,
			aggregates.LoadEditGrantFromHistory,
			ErrEditGrantNotFound,
		),
	}
}

var editGrantEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"EditGrantActivated": repository.JSONDeserializer[events.EditGrantActivated],
		"EditGrantRevoked":   repository.JSONDeserializer[events.EditGrantRevoked],
		"EditGrantExpired":   repository.JSONDeserializer[events.EditGrantExpired],
	},
)
