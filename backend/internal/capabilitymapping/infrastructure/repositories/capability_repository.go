package repositories

import (
	"errors"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrCapabilityNotFound = errors.New("capability not found")

type CapabilityRepository struct {
	*repository.EventSourcedRepository[*aggregates.Capability]
}

func NewCapabilityRepository(eventStore eventstore.EventStore) *CapabilityRepository {
	return &CapabilityRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			capabilityEventDeserializers,
			aggregates.LoadCapabilityFromHistory,
			ErrCapabilityNotFound,
		),
	}
}

var capabilityEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"CapabilityCreated":                 repository.JSONDeserializer[events.CapabilityCreated],
		"CapabilityUpdated":                 repository.JSONDeserializer[events.CapabilityUpdated],
		"CapabilityDeleted":                 repository.JSONDeserializer[events.CapabilityDeleted],
		"CapabilityMetadataUpdated":         repository.JSONDeserializer[events.CapabilityMetadataUpdated],
		"CapabilityExpertAdded":             repository.JSONDeserializer[events.CapabilityExpertAdded],
		"CapabilityExpertRemoved":           repository.JSONDeserializer[events.CapabilityExpertRemoved],
		"CapabilityTagAdded":                repository.JSONDeserializer[events.CapabilityTagAdded],
		"CapabilityParentChanged":           repository.JSONDeserializer[events.CapabilityParentChanged],
		"CapabilityLevelChanged":            repository.JSONDeserializer[events.CapabilityLevelChanged],
		"CapabilityRealizationsInherited":   repository.JSONDeserializer[events.CapabilityRealizationsInherited],
		"CapabilityRealizationsUninherited": repository.JSONDeserializer[events.CapabilityRealizationsUninherited],
	},
	CapabilityMetadataUpdatedV1ToV2Upcaster{},
)
