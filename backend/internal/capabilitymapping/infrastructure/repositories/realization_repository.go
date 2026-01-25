package repositories

import (
	"errors"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrRealizationNotFound = errors.New("realization not found")

type RealizationRepository struct {
	*repository.EventSourcedRepository[*aggregates.CapabilityRealization]
}

func NewRealizationRepository(eventStore eventstore.EventStore) *RealizationRepository {
	return &RealizationRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			realizationEventDeserializers,
			aggregates.LoadCapabilityRealizationFromHistory,
			ErrRealizationNotFound,
		),
	}
}

var realizationEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"SystemLinkedToCapability": repository.JSONDeserializer[events.SystemLinkedToCapability],
		"SystemRealizationUpdated": repository.JSONDeserializer[events.SystemRealizationUpdated],
		"SystemRealizationDeleted": repository.JSONDeserializer[events.SystemRealizationDeleted],
	},
)
