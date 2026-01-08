package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrComponentNotFound = errors.New("component not found")

type ApplicationComponentRepository struct {
	*repository.EventSourcedRepository[*aggregates.ApplicationComponent]
}

func NewApplicationComponentRepository(eventStore eventstore.EventStore) *ApplicationComponentRepository {
	return &ApplicationComponentRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			componentEventDeserializers,
			aggregates.LoadApplicationComponentFromHistory,
			ErrComponentNotFound,
		),
	}
}

var componentEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"ApplicationComponentCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
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

			evt := events.NewApplicationComponentCreated(id, name, description)
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"ApplicationComponentUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
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

			return events.NewApplicationComponentUpdated(id, name, description), nil
		},
		"ApplicationComponentDeleted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}

			return events.NewApplicationComponentDeleted(id, name), nil
		},
	},
)
