package repositories

import (
	"errors"

	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrEnterpriseCapabilityNotFound = errors.New("enterprise capability not found")

type EnterpriseCapabilityRepository struct {
	*repository.EventSourcedRepository[*aggregates.EnterpriseCapability]
}

func NewEnterpriseCapabilityRepository(eventStore eventstore.EventStore) *EnterpriseCapabilityRepository {
	return &EnterpriseCapabilityRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			enterpriseCapabilityEventDeserializers,
			aggregates.LoadEnterpriseCapabilityFromHistory,
			ErrEnterpriseCapabilityNotFound,
		),
	}
}

var enterpriseCapabilityEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"EnterpriseCapabilityCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			description, err := repository.GetRequiredString(data, "description")
			if err != nil {
				return nil, err
			}
			category, err := repository.GetRequiredString(data, "category")
			if err != nil {
				return nil, err
			}
			active, err := repository.GetOptionalBool(data, "active", true)
			if err != nil {
				return nil, err
			}
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewEnterpriseCapabilityCreated(id, name, description, category)
			evt.Active = active
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"EnterpriseCapabilityUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			description, err := repository.GetRequiredString(data, "description")
			if err != nil {
				return nil, err
			}
			category, err := repository.GetRequiredString(data, "category")
			if err != nil {
				return nil, err
			}

			return events.NewEnterpriseCapabilityUpdated(id, name, description, category), nil
		},
		"EnterpriseCapabilityDeleted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}

			return events.NewEnterpriseCapabilityDeleted(id), nil
		},
		"EnterpriseCapabilityTargetMaturitySet": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			targetMaturity, err := repository.GetRequiredInt(data, "targetMaturity")
			if err != nil {
				return nil, err
			}

			return events.NewEnterpriseCapabilityTargetMaturitySet(id, targetMaturity), nil
		},
	},
)
