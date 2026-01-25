package repositories

import (
	"errors"

	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
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
		"EnterpriseCapabilityCreated":           repository.JSONDeserializer[events.EnterpriseCapabilityCreated],
		"EnterpriseCapabilityUpdated":           repository.JSONDeserializer[events.EnterpriseCapabilityUpdated],
		"EnterpriseCapabilityDeleted":           repository.JSONDeserializer[events.EnterpriseCapabilityDeleted],
		"EnterpriseCapabilityTargetMaturitySet": repository.JSONDeserializer[events.EnterpriseCapabilityTargetMaturitySet],
	},
)
