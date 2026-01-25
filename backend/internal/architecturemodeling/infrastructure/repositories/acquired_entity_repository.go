package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrAcquiredEntityNotFound = errors.New("acquired entity not found")

type AcquiredEntityRepository struct {
	*repository.EventSourcedRepository[*aggregates.AcquiredEntity]
}

func NewAcquiredEntityRepository(eventStore eventstore.EventStore) *AcquiredEntityRepository {
	return &AcquiredEntityRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			acquiredEntityEventDeserializers,
			aggregates.LoadAcquiredEntityFromHistory,
			ErrAcquiredEntityNotFound,
		),
	}
}

var acquiredEntityEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"AcquiredEntityCreated": repository.JSONDeserializer[events.AcquiredEntityCreated],
		"AcquiredEntityUpdated": repository.JSONDeserializer[events.AcquiredEntityUpdated],
		"AcquiredEntityDeleted": repository.JSONDeserializer[events.AcquiredEntityDeleted],
	},
)
