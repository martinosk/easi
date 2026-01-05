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
		"SystemLinkedToCapability": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			capabilityID, _ := data["capabilityId"].(string)
			componentID, _ := data["componentId"].(string)
			componentName, _ := data["componentName"].(string)
			realizationLevel, _ := data["realizationLevel"].(string)
			notes, _ := data["notes"].(string)
			linkedAtStr, _ := data["linkedAt"].(string)
			linkedAt, _ := time.Parse(time.RFC3339Nano, linkedAtStr)

			evt := events.NewSystemLinkedToCapability(id, capabilityID, componentID, componentName, realizationLevel, notes)
			evt.LinkedAt = linkedAt
			return evt
		},
		"SystemRealizationUpdated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			realizationLevel, _ := data["realizationLevel"].(string)
			notes, _ := data["notes"].(string)

			return events.NewSystemRealizationUpdated(id, realizationLevel, notes)
		},
		"SystemRealizationDeleted": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			deletedAtStr, _ := data["deletedAt"].(string)
			deletedAt, _ := time.Parse(time.RFC3339Nano, deletedAtStr)

			evt := events.NewSystemRealizationDeleted(id)
			evt.DeletedAt = deletedAt
			return evt
		},
	},
)
