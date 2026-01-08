package repositories

import (
	"errors"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrBusinessDomainNotFound = errors.New("business domain not found")

type BusinessDomainRepository struct {
	*repository.EventSourcedRepository[*aggregates.BusinessDomain]
}

func NewBusinessDomainRepository(eventStore eventstore.EventStore) *BusinessDomainRepository {
	return &BusinessDomainRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			businessDomainEventDeserializers,
			aggregates.LoadBusinessDomainFromHistory,
			ErrBusinessDomainNotFound,
		),
	}
}

var businessDomainEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"BusinessDomainCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
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
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewBusinessDomainCreated(id, name, description)
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"BusinessDomainUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
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

			return events.NewBusinessDomainUpdated(id, name, description), nil
		},
		"BusinessDomainDeleted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}

			return events.NewBusinessDomainDeleted(id), nil
		},
	},
)
