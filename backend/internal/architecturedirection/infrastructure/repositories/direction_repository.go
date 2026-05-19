package repositories

import (
	"errors"

	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/events"
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
		"DirectionDrafted":                   repository.JSONDeserializer[events.DirectionDrafted],
		"DirectionProposed":                  repository.JSONDeserializer[events.DirectionProposed],
		"DirectionAgreed":                    repository.JSONDeserializer[events.DirectionAgreed],
		"DirectionRejected":                  repository.JSONDeserializer[events.DirectionRejected],
		"DirectionNarrativeUpdated":          repository.JSONDeserializer[events.DirectionNarrativeUpdated],
		"DirectionHorizonChanged":            repository.JSONDeserializer[events.DirectionHorizonChanged],
		"DirectionPlacementsChanged":         repository.JSONDeserializer[events.DirectionPlacementsChanged],
		"DirectionSourceCapabilitiesChanged": repository.JSONDeserializer[events.DirectionSourceCapabilitiesChanged],
	},
)
