package repositories

import (
	"errors"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
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
		"SystemLinkedToCapability": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			capabilityID, err := repository.GetRequiredString(data, "capabilityId")
			if err != nil {
				return nil, err
			}
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			componentName, err := repository.GetRequiredString(data, "componentName")
			if err != nil {
				return nil, err
			}
			realizationLevel, err := repository.GetRequiredString(data, "realizationLevel")
			if err != nil {
				return nil, err
			}
			notes, err := repository.GetOptionalString(data, "notes", "")
			if err != nil {
				return nil, err
			}
			linkedAt, err := repository.GetRequiredTime(data, "linkedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewSystemLinkedToCapability(id, capabilityID, componentID, componentName, realizationLevel, notes)
			evt.LinkedAt = linkedAt
			return evt, nil
		},
		"SystemRealizationUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			realizationLevel, err := repository.GetRequiredString(data, "realizationLevel")
			if err != nil {
				return nil, err
			}
			notes, err := repository.GetOptionalString(data, "notes", "")
			if err != nil {
				return nil, err
			}

			return events.NewSystemRealizationUpdated(id, realizationLevel, notes), nil
		},
		"SystemRealizationDeleted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			deletedAt, err := repository.GetRequiredTime(data, "deletedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewSystemRealizationDeleted(id)
			evt.DeletedAt = deletedAt
			return evt, nil
		},
	},
)
