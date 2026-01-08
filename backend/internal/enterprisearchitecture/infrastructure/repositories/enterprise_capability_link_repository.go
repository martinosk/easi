package repositories

import (
	"errors"

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
		"EnterpriseCapabilityLinked": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			enterpriseCapabilityID, err := repository.GetRequiredString(data, "enterpriseCapabilityId")
			if err != nil {
				return nil, err
			}
			domainCapabilityID, err := repository.GetRequiredString(data, "domainCapabilityId")
			if err != nil {
				return nil, err
			}
			linkedBy, err := repository.GetRequiredString(data, "linkedBy")
			if err != nil {
				return nil, err
			}
			linkedAt, err := repository.GetRequiredTime(data, "linkedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewEnterpriseCapabilityLinked(id, enterpriseCapabilityID, domainCapabilityID, linkedBy)
			evt.LinkedAt = linkedAt
			return evt, nil
		},
		"EnterpriseCapabilityUnlinked": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			enterpriseCapabilityID, err := repository.GetRequiredString(data, "enterpriseCapabilityId")
			if err != nil {
				return nil, err
			}
			domainCapabilityID, err := repository.GetRequiredString(data, "domainCapabilityId")
			if err != nil {
				return nil, err
			}

			return events.NewEnterpriseCapabilityUnlinked(id, enterpriseCapabilityID, domainCapabilityID), nil
		},
	},
)
