package repositories

import (
	"errors"

	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
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
		"EnterpriseStrategicImportanceSet": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			enterpriseCapabilityID, err := repository.GetRequiredString(data, "enterpriseCapabilityId")
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

			evt := events.NewEnterpriseStrategicImportanceSet(events.EnterpriseStrategicImportanceSetParams{
				ID:                     id,
				EnterpriseCapabilityID: enterpriseCapabilityID,
				PillarID:               pillarID,
				PillarName:             pillarName,
				Importance:             importance,
				Rationale:              rationale,
			})
			evt.SetAt = setAt
			return evt, nil
		},
		"EnterpriseStrategicImportanceUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
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

			return events.NewEnterpriseStrategicImportanceUpdated(events.EnterpriseStrategicImportanceUpdatedParams{
				ID:            id,
				Importance:    importance,
				Rationale:     rationale,
				OldImportance: oldImportance,
				OldRationale:  oldRationale,
			}), nil
		},
		"EnterpriseStrategicImportanceRemoved": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			enterpriseCapabilityID, err := repository.GetRequiredString(data, "enterpriseCapabilityId")
			if err != nil {
				return nil, err
			}
			pillarID, err := repository.GetRequiredString(data, "pillarId")
			if err != nil {
				return nil, err
			}

			return events.NewEnterpriseStrategicImportanceRemoved(id, enterpriseCapabilityID, pillarID), nil
		},
	},
)
