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
		"EnterpriseStrategicImportanceSet": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			enterpriseCapabilityID, _ := data["enterpriseCapabilityId"].(string)
			pillarID, _ := data["pillarId"].(string)
			pillarName, _ := data["pillarName"].(string)
			importance := int(data["importance"].(float64))
			rationale, _ := data["rationale"].(string)
			setAtStr, _ := data["setAt"].(string)
			setAt, _ := time.Parse(time.RFC3339Nano, setAtStr)

			evt := events.NewEnterpriseStrategicImportanceSet(events.EnterpriseStrategicImportanceSetParams{
				ID:                     id,
				EnterpriseCapabilityID: enterpriseCapabilityID,
				PillarID:               pillarID,
				PillarName:             pillarName,
				Importance:             importance,
				Rationale:              rationale,
			})
			evt.SetAt = setAt
			return evt
		},
		"EnterpriseStrategicImportanceUpdated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			importance := int(data["importance"].(float64))
			rationale, _ := data["rationale"].(string)
			oldImportance := int(data["oldImportance"].(float64))
			oldRationale, _ := data["oldRationale"].(string)

			return events.NewEnterpriseStrategicImportanceUpdated(events.EnterpriseStrategicImportanceUpdatedParams{
				ID:            id,
				Importance:    importance,
				Rationale:     rationale,
				OldImportance: oldImportance,
				OldRationale:  oldRationale,
			})
		},
		"EnterpriseStrategicImportanceRemoved": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			enterpriseCapabilityID, _ := data["enterpriseCapabilityId"].(string)
			pillarID, _ := data["pillarId"].(string)

			return events.NewEnterpriseStrategicImportanceRemoved(id, enterpriseCapabilityID, pillarID)
		},
	},
)
