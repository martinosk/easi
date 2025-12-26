package repositories

import (
	"errors"
	"time"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/eventsourcing"
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
		"ApplicationComponentCreated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			name, _ := data["name"].(string)
			description, _ := data["description"].(string)
			createdAtStr, _ := data["createdAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

			evt := events.NewApplicationComponentCreated(id, name, description)
			evt.CreatedAt = createdAt
			return evt
		},
		"ApplicationComponentUpdated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			name, _ := data["name"].(string)
			description, _ := data["description"].(string)

			return events.NewApplicationComponentUpdated(id, name, description)
		},
		"ApplicationComponentDeleted": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			name, _ := data["name"].(string)

			return events.NewApplicationComponentDeleted(id, name)
		},
	},
)
