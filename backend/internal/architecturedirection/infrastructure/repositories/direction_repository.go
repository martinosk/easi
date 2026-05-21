package repositories

import (
	"errors"

	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/events"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrDirectionNotFound = errors.New("direction not found")

type DirectionRepository struct {
	*repository.EventSourcedRepository[*aggregates.Direction]
}

func NewDirectionRepository(eventStore eventstore.EventStore) *DirectionRepository {
	return &DirectionRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			directionEventDeserializers,
			aggregates.LoadDirectionFromHistory,
			ErrDirectionNotFound,
		),
	}
}

var directionEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		pl.DirectionDrafted:                   repository.JSONDeserializer[events.DirectionDrafted],
		pl.DirectionProposed:                  repository.JSONDeserializer[events.DirectionProposed],
		pl.DirectionAgreed:                    repository.JSONDeserializer[events.DirectionAgreed],
		pl.DirectionRejected:                  repository.JSONDeserializer[events.DirectionRejected],
		pl.DirectionNarrativeUpdated:          repository.JSONDeserializer[events.DirectionNarrativeUpdated],
		pl.DirectionHorizonChanged:            repository.JSONDeserializer[events.DirectionHorizonChanged],
		pl.DirectionPlacementsChanged:         repository.JSONDeserializer[events.DirectionPlacementsChanged],
		pl.DirectionSourceCapabilitiesChanged: repository.JSONDeserializer[events.DirectionSourceCapabilitiesChanged],
	},
)
