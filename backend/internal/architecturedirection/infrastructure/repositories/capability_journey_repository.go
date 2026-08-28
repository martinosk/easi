package repositories

import (
	"errors"

	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/events"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrCapabilityJourneyNotFound = errors.New("capability journey not found")

type CapabilityJourneyRepository struct {
	*repository.EventSourcedRepository[*aggregates.CapabilityJourney]
}

func NewCapabilityJourneyRepository(eventStore eventstore.EventStore) *CapabilityJourneyRepository {
	return &CapabilityJourneyRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			capabilityJourneyEventDeserializers,
			aggregates.LoadCapabilityJourneyFromHistory,
			ErrCapabilityJourneyNotFound,
		),
	}
}

var capabilityJourneyEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		pl.JourneyPlanned:                   repository.JSONDeserializer[events.JourneyPlanned],
		pl.JourneyStarted:                   repository.JSONDeserializer[events.JourneyStarted],
		pl.JourneyCompleted:                 repository.JSONDeserializer[events.JourneyCompleted],
		pl.JourneyAbandoned:                 repository.JSONDeserializer[events.JourneyAbandoned],
		pl.JourneyProgressUpdated:           repository.JSONDeserializer[events.JourneyProgressUpdated],
		pl.JourneyDetailsUpdated:            repository.JSONDeserializer[events.JourneyDetailsUpdated],
		pl.JourneyMilestoneAdded:            repository.JSONDeserializer[events.JourneyMilestoneAdded],
		pl.JourneyMilestoneUpdated:          repository.JSONDeserializer[events.JourneyMilestoneUpdated],
		pl.JourneyMilestoneRemoved:          repository.JSONDeserializer[events.JourneyMilestoneRemoved],
		pl.JourneyMilestonesReordered:       repository.JSONDeserializer[events.JourneyMilestonesReordered],
		pl.JourneySourceApplicationsChanged: repository.JSONDeserializer[events.JourneySourceApplicationsChanged],
	},
)
