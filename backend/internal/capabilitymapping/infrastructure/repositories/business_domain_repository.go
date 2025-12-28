package repositories

import (
	"errors"
	"time"

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
		"BusinessDomainCreated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			name, _ := data["name"].(string)
			description, _ := data["description"].(string)
			createdAtStr, _ := data["createdAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

			evt := events.NewBusinessDomainCreated(id, name, description)
			evt.CreatedAt = createdAt
			return evt
		},
		"BusinessDomainUpdated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			name, _ := data["name"].(string)
			description, _ := data["description"].(string)

			return events.NewBusinessDomainUpdated(id, name, description)
		},
		"BusinessDomainDeleted": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)

			return events.NewBusinessDomainDeleted(id)
		},
	},
)
