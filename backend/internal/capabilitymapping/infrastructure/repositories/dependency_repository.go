package repositories

import (
	"errors"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
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
		"CapabilityDependencyCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			sourceCapabilityID, err := repository.GetRequiredString(data, "sourceCapabilityId")
			if err != nil {
				return nil, err
			}
			targetCapabilityID, err := repository.GetRequiredString(data, "targetCapabilityId")
			if err != nil {
				return nil, err
			}
			dependencyType, err := repository.GetRequiredString(data, "dependencyType")
			if err != nil {
				return nil, err
			}
			description, err := repository.GetRequiredString(data, "description")
			if err != nil {
				return nil, err
			}
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewCapabilityDependencyCreated(id, sourceCapabilityID, targetCapabilityID, dependencyType, description)
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"CapabilityDependencyDeleted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			deletedAt, err := repository.GetRequiredTime(data, "deletedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewCapabilityDependencyDeleted(id)
			evt.DeletedAt = deletedAt
			return evt, nil
		},
	},
)
