package repositories

import (
	"errors"

	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrEnterpriseStrategicImportanceNotFound = errors.New("enterprise strategic importance not found")

type EnterpriseStrategicImportanceRepository struct {
	*repository.EventSourcedRepository[*aggregates.EnterpriseStrategicImportance]
}

func NewEnterpriseStrategicImportanceRepository(eventStore eventstore.EventStore) *EnterpriseStrategicImportanceRepository {
	return &EnterpriseStrategicImportanceRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			enterpriseStrategicImportanceEventDeserializers,
			aggregates.LoadEnterpriseStrategicImportanceFromHistory,
			ErrEnterpriseStrategicImportanceNotFound,
		),
	}
}

var enterpriseStrategicImportanceEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"EnterpriseStrategicImportanceSet":     repository.JSONDeserializer[events.EnterpriseStrategicImportanceSet],
		"EnterpriseStrategicImportanceUpdated": repository.JSONDeserializer[events.EnterpriseStrategicImportanceUpdated],
		"EnterpriseStrategicImportanceRemoved": repository.JSONDeserializer[events.EnterpriseStrategicImportanceRemoved],
	},
)
