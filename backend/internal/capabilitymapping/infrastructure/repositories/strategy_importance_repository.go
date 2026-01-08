package repositories

import (
	"errors"

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
		"StrategyImportanceSet": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			businessDomainID, err := repository.GetRequiredString(data, "businessDomainId")
			if err != nil {
				return nil, err
			}
			capabilityID, err := repository.GetRequiredString(data, "capabilityId")
			if err != nil {
				return nil, err
			}
			pillarID, err := repository.GetRequiredString(data, "pillarId")
			if err != nil {
				return nil, err
			}
			pillarName, err := repository.GetRequiredString(data, "pillarName")
			if err != nil {
				return nil, err
			}
			importance, err := repository.GetRequiredInt(data, "importance")
			if err != nil {
				return nil, err
			}
			rationale, err := repository.GetOptionalString(data, "rationale", "")
			if err != nil {
				return nil, err
			}
			setAt, err := repository.GetRequiredTime(data, "setAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewStrategyImportanceSet(events.StrategyImportanceSetParams{
				ID:               id,
				BusinessDomainID: businessDomainID,
				CapabilityID:     capabilityID,
				PillarID:         pillarID,
				PillarName:       pillarName,
				Importance:       importance,
				Rationale:        rationale,
			})
			evt.SetAt = setAt
			return evt, nil
		},
		"StrategyImportanceUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			importance, err := repository.GetRequiredInt(data, "importance")
			if err != nil {
				return nil, err
			}
			rationale, err := repository.GetOptionalString(data, "rationale", "")
			if err != nil {
				return nil, err
			}
			oldImportance, err := repository.GetRequiredInt(data, "oldImportance")
			if err != nil {
				return nil, err
			}
			oldRationale, err := repository.GetOptionalString(data, "oldRationale", "")
			if err != nil {
				return nil, err
			}
			updatedAt, err := repository.GetRequiredTime(data, "updatedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewStrategyImportanceUpdated(events.StrategyImportanceUpdatedParams{
				ID:            id,
				Importance:    importance,
				Rationale:     rationale,
				OldImportance: oldImportance,
				OldRationale:  oldRationale,
			})
			evt.UpdatedAt = updatedAt
			return evt, nil
		},
		"StrategyImportanceRemoved": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			businessDomainID, err := repository.GetRequiredString(data, "businessDomainId")
			if err != nil {
				return nil, err
			}
			capabilityID, err := repository.GetRequiredString(data, "capabilityId")
			if err != nil {
				return nil, err
			}
			pillarID, err := repository.GetRequiredString(data, "pillarId")
			if err != nil {
				return nil, err
			}

			return events.NewStrategyImportanceRemoved(id, businessDomainID, capabilityID, pillarID), nil
		},
	},
)
