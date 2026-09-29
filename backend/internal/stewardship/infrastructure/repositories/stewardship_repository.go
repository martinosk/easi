package repositories

import (
	"errors"

	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
	"easi/backend/internal/stewardship/domain/aggregates"
	"easi/backend/internal/stewardship/domain/events"
	pl "easi/backend/internal/stewardship/publishedlanguage"
)

var ErrStewardshipNotFound = errors.New("stewardship not found")

type StewardshipRepository struct {
	*repository.EventSourcedRepository[*aggregates.Stewardship]
}

func NewStewardshipRepository(eventStore eventstore.EventStore) *StewardshipRepository {
	return &StewardshipRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			stewardshipEventDeserializers,
			aggregates.LoadStewardshipFromHistory,
			ErrStewardshipNotFound,
		),
	}
}

var stewardshipEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		pl.StewardAssigned: repository.JSONDeserializer[events.StewardAssigned],
		pl.StewardReleased: repository.JSONDeserializer[events.StewardReleased],
	},
)
