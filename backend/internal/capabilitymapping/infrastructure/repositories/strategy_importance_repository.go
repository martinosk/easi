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
		"StrategyImportanceSet": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			businessDomainID, _ := data["businessDomainId"].(string)
			capabilityID, _ := data["capabilityId"].(string)
			pillarID, _ := data["pillarId"].(string)
			pillarName, _ := data["pillarName"].(string)
			importance, _ := data["importance"].(float64)
			rationale, _ := data["rationale"].(string)
			setAtStr, _ := data["setAt"].(string)
			setAt, _ := time.Parse(time.RFC3339Nano, setAtStr)

			evt := events.NewStrategyImportanceSet(events.StrategyImportanceSetParams{
				ID:               id,
				BusinessDomainID: businessDomainID,
				CapabilityID:     capabilityID,
				PillarID:         pillarID,
				PillarName:       pillarName,
				Importance:       int(importance),
				Rationale:        rationale,
			})
			evt.SetAt = setAt
			return evt
		},
		"StrategyImportanceUpdated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			importance, _ := data["importance"].(float64)
			rationale, _ := data["rationale"].(string)
			oldImportance, _ := data["oldImportance"].(float64)
			oldRationale, _ := data["oldRationale"].(string)
			updatedAtStr, _ := data["updatedAt"].(string)
			updatedAt, _ := time.Parse(time.RFC3339Nano, updatedAtStr)

			evt := events.NewStrategyImportanceUpdated(events.StrategyImportanceUpdatedParams{
				ID:            id,
				Importance:    int(importance),
				Rationale:     rationale,
				OldImportance: int(oldImportance),
				OldRationale:  oldRationale,
			})
			evt.UpdatedAt = updatedAt
			return evt
		},
		"StrategyImportanceRemoved": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			businessDomainID, _ := data["businessDomainId"].(string)
			capabilityID, _ := data["capabilityId"].(string)
			pillarID, _ := data["pillarId"].(string)

			return events.NewStrategyImportanceRemoved(id, businessDomainID, capabilityID, pillarID)
		},
	},
)
