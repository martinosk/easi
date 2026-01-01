package repositories

import (
	"errors"
	"time"

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
		"EnterpriseCapabilityCreated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			name, _ := data["name"].(string)
			description, _ := data["description"].(string)
			category, _ := data["category"].(string)
			active, _ := data["active"].(bool)
			createdAtStr, _ := data["createdAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

			evt := events.NewEnterpriseCapabilityCreated(id, name, description, category)
			evt.Active = active
			evt.CreatedAt = createdAt
			return evt
		},
		"EnterpriseCapabilityUpdated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			name, _ := data["name"].(string)
			description, _ := data["description"].(string)
			category, _ := data["category"].(string)

			return events.NewEnterpriseCapabilityUpdated(id, name, description, category)
		},
		"EnterpriseCapabilityDeleted": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)

			return events.NewEnterpriseCapabilityDeleted(id)
		},
		"EnterpriseCapabilityTargetMaturitySet": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			targetMaturity, _ := data["targetMaturity"].(float64)

			return events.NewEnterpriseCapabilityTargetMaturitySet(id, int(targetMaturity))
		},
	},
)
