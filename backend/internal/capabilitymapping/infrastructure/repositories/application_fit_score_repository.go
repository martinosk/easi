package repositories

import (
	"errors"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrApplicationFitScoreNotFound = errors.New("application fit score not found")

type ApplicationFitScoreRepository struct {
	*repository.EventSourcedRepository[*aggregates.ApplicationFitScore]
}

func NewApplicationFitScoreRepository(eventStore eventstore.EventStore) *ApplicationFitScoreRepository {
	return &ApplicationFitScoreRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			applicationFitScoreEventDeserializers,
			aggregates.LoadApplicationFitScoreFromHistory,
			ErrApplicationFitScoreNotFound,
		),
	}
}

var applicationFitScoreEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"ApplicationFitScoreSet": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			componentID, err := repository.GetRequiredString(data, "componentId")
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
			score, err := repository.GetRequiredInt(data, "score")
			if err != nil {
				return nil, err
			}
			rationale, err := repository.GetOptionalString(data, "rationale", "")
			if err != nil {
				return nil, err
			}
			scoredAt, err := repository.GetRequiredTime(data, "scoredAt")
			if err != nil {
				return nil, err
			}
			scoredBy, err := repository.GetRequiredString(data, "scoredBy")
			if err != nil {
				return nil, err
			}

			evt := events.NewApplicationFitScoreSet(events.ApplicationFitScoreSetParams{
				ID:          id,
				ComponentID: componentID,
				PillarID:    pillarID,
				PillarName:  pillarName,
				Score:       score,
				Rationale:   rationale,
				ScoredBy:    scoredBy,
			})
			evt.ScoredAt = scoredAt
			return evt, nil
		},
		"ApplicationFitScoreUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			score, err := repository.GetRequiredInt(data, "score")
			if err != nil {
				return nil, err
			}
			rationale, err := repository.GetOptionalString(data, "rationale", "")
			if err != nil {
				return nil, err
			}
			oldScore, err := repository.GetRequiredInt(data, "oldScore")
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
			updatedBy, err := repository.GetRequiredString(data, "updatedBy")
			if err != nil {
				return nil, err
			}

			evt := events.NewApplicationFitScoreUpdated(events.ApplicationFitScoreUpdatedParams{
				ID:           id,
				Score:        score,
				Rationale:    rationale,
				OldScore:     oldScore,
				OldRationale: oldRationale,
				UpdatedBy:    updatedBy,
			})
			evt.UpdatedAt = updatedAt
			return evt, nil
		},
		"ApplicationFitScoreRemoved": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			pillarID, err := repository.GetRequiredString(data, "pillarId")
			if err != nil {
				return nil, err
			}
			removedBy, err := repository.GetRequiredString(data, "removedBy")
			if err != nil {
				return nil, err
			}

			return events.NewApplicationFitScoreRemoved(id, componentID, pillarID, removedBy), nil
		},
	},
)
