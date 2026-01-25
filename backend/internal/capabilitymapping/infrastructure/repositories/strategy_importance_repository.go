package repositories

import (
	"errors"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrStrategyImportanceNotFound = errors.New("strategy importance not found")

type StrategyImportanceRepository struct {
	*repository.EventSourcedRepository[*aggregates.StrategyImportance]
}

func NewStrategyImportanceRepository(eventStore eventstore.EventStore) *StrategyImportanceRepository {
	return &StrategyImportanceRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			strategyImportanceEventDeserializers,
			aggregates.LoadStrategyImportanceFromHistory,
			ErrStrategyImportanceNotFound,
		),
	}
}

var strategyImportanceEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"StrategyImportanceSet":     repository.JSONDeserializer[events.StrategyImportanceSet],
		"StrategyImportanceUpdated": repository.JSONDeserializer[events.StrategyImportanceUpdated],
		"StrategyImportanceRemoved": repository.JSONDeserializer[events.StrategyImportanceRemoved],
	},
)
