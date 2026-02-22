package repositories

import (
	"errors"

	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/events"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrMetaModelConfigurationNotFound = errors.New("meta model configuration not found")

type MetaModelConfigurationRepository struct {
	*repository.EventSourcedRepository[*aggregates.MetaModelConfiguration]
}

func NewMetaModelConfigurationRepository(eventStore eventstore.EventStore) *MetaModelConfigurationRepository {
	return &MetaModelConfigurationRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			metaModelEventDeserializers,
			aggregates.LoadMetaModelConfigurationFromHistory,
			ErrMetaModelConfigurationNotFound,
		),
	}
}

var metaModelEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"MetaModelConfigurationCreated": repository.JSONDeserializer[events.MetaModelConfigurationCreated],
		"MaturityScaleConfigUpdated":    repository.JSONDeserializer[events.MaturityScaleConfigUpdated],
		"MaturityScaleConfigReset":      repository.JSONDeserializer[events.MaturityScaleConfigReset],
		"StrategyPillarAdded":           repository.JSONDeserializer[events.StrategyPillarAdded],
		"StrategyPillarUpdated":         repository.JSONDeserializer[events.StrategyPillarUpdated],
		"StrategyPillarRemoved":         repository.JSONDeserializer[events.StrategyPillarRemoved],
		"PillarFitConfigurationUpdated": repository.JSONDeserializer[events.PillarFitConfigurationUpdated],
	},
)
