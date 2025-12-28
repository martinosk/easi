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
		"CapabilityDependencyCreated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			sourceCapabilityID, _ := data["sourceCapabilityId"].(string)
			targetCapabilityID, _ := data["targetCapabilityId"].(string)
			dependencyType, _ := data["dependencyType"].(string)
			description, _ := data["description"].(string)
			createdAtStr, _ := data["createdAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

			evt := events.NewCapabilityDependencyCreated(id, sourceCapabilityID, targetCapabilityID, dependencyType, description)
			evt.CreatedAt = createdAt
			return evt
		},
		"CapabilityDependencyDeleted": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			deletedAtStr, _ := data["deletedAt"].(string)
			deletedAt, _ := time.Parse(time.RFC3339Nano, deletedAtStr)

			evt := events.NewCapabilityDependencyDeleted(id)
			evt.DeletedAt = deletedAt
			return evt
		},
	},
)
