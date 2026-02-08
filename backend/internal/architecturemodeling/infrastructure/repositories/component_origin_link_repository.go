package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrComponentOriginLinkNotFound = errors.New("component origin link not found")

type ComponentOriginLinkRepository struct {
	*repository.EventSourcedRepository[*aggregates.ComponentOriginLink]
}

func NewComponentOriginLinkRepository(eventStore eventstore.EventStore) *ComponentOriginLinkRepository {
	return &ComponentOriginLinkRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			componentOriginLinkEventDeserializers,
			aggregates.LoadComponentOriginLinkFromHistory,
			ErrComponentOriginLinkNotFound,
		),
	}
}

var componentOriginLinkEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"OriginLinkCreated":      repository.JSONDeserializer[events.OriginLinkCreated],
		"OriginLinkSet":          repository.JSONDeserializer[events.OriginLinkSet],
		"OriginLinkReplaced":     repository.JSONDeserializer[events.OriginLinkReplaced],
		"OriginLinkNotesUpdated": repository.JSONDeserializer[events.OriginLinkNotesUpdated],
		"OriginLinkCleared":      repository.JSONDeserializer[events.OriginLinkCleared],
		"OriginLinkDeleted":      repository.JSONDeserializer[events.OriginLinkDeleted],
	},
)
