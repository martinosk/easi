package repositories

import (
	"errors"

	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
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
		"EnterpriseCapabilityLinked":   repository.JSONDeserializer[events.EnterpriseCapabilityLinked],
		"EnterpriseCapabilityUnlinked": repository.JSONDeserializer[events.EnterpriseCapabilityUnlinked],
	},
)
