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

var ErrEnterpriseCapabilityLinkNotFound = errors.New("enterprise capability link not found")

type EnterpriseCapabilityLinkRepository struct {
	*repository.EventSourcedRepository[*aggregates.EnterpriseCapabilityLink]
}

func NewEnterpriseCapabilityLinkRepository(eventStore eventstore.EventStore) *EnterpriseCapabilityLinkRepository {
	return &EnterpriseCapabilityLinkRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			enterpriseCapabilityLinkEventDeserializers,
			aggregates.LoadEnterpriseCapabilityLinkFromHistory,
			ErrEnterpriseCapabilityLinkNotFound,
		),
	}
}

var enterpriseCapabilityLinkEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"EnterpriseCapabilityLinked": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			enterpriseCapabilityID, _ := data["enterpriseCapabilityId"].(string)
			domainCapabilityID, _ := data["domainCapabilityId"].(string)
			linkedBy, _ := data["linkedBy"].(string)
			linkedAtStr, _ := data["linkedAt"].(string)
			linkedAt, _ := time.Parse(time.RFC3339Nano, linkedAtStr)

			evt := events.NewEnterpriseCapabilityLinked(id, enterpriseCapabilityID, domainCapabilityID, linkedBy)
			evt.LinkedAt = linkedAt
			return evt
		},
		"EnterpriseCapabilityUnlinked": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			enterpriseCapabilityID, _ := data["enterpriseCapabilityId"].(string)
			domainCapabilityID, _ := data["domainCapabilityId"].(string)

			return events.NewEnterpriseCapabilityUnlinked(id, enterpriseCapabilityID, domainCapabilityID)
		},
	},
)
