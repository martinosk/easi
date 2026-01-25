package repositories

import (
	"errors"

	"easi/backend/internal/architectureviews/domain/aggregates"
	"easi/backend/internal/architectureviews/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrViewNotFound = errors.New("view not found")

type ArchitectureViewRepository struct {
	*repository.EventSourcedRepository[*aggregates.ArchitectureView]
}

func NewArchitectureViewRepository(eventStore eventstore.EventStore) *ArchitectureViewRepository {
	return &ArchitectureViewRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			eventDeserializers,
			aggregates.LoadArchitectureViewFromHistory,
			ErrViewNotFound,
		),
	}
}

var eventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"ViewCreated":              repository.JSONDeserializer[events.ViewCreated],
		"ComponentAddedToView":     repository.JSONDeserializer[events.ComponentAddedToView],
		"ComponentRemovedFromView": repository.JSONDeserializer[events.ComponentRemovedFromView],
		"ViewRenamed":              repository.JSONDeserializer[events.ViewRenamed],
		"ViewDeleted":              repository.JSONDeserializer[events.ViewDeleted],
		"DefaultViewChanged":       repository.JSONDeserializer[events.DefaultViewChanged],
		"ViewVisibilityChanged":    repository.JSONDeserializer[events.ViewVisibilityChanged],
	},
)
