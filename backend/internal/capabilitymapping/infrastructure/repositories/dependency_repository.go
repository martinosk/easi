package repositories

import (
	"errors"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrDependencyNotFound = errors.New("dependency not found")

type DependencyRepository struct {
	*repository.EventSourcedRepository[*aggregates.CapabilityDependency]
}

func NewDependencyRepository(eventStore eventstore.EventStore) *DependencyRepository {
	return &DependencyRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			dependencyEventDeserializers,
			aggregates.LoadCapabilityDependencyFromHistory,
			ErrDependencyNotFound,
		),
	}
}

var dependencyEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"CapabilityDependencyCreated": repository.JSONDeserializer[events.CapabilityDependencyCreated],
		"CapabilityDependencyDeleted": repository.JSONDeserializer[events.CapabilityDependencyDeleted],
	},
)
