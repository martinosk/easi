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
		"ApplicationFitScoreSet": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			componentID, _ := data["componentId"].(string)
			pillarID, _ := data["pillarId"].(string)
			pillarName, _ := data["pillarName"].(string)
			score, _ := data["score"].(float64)
			rationale, _ := data["rationale"].(string)
			scoredAtStr, _ := data["scoredAt"].(string)
			scoredAt, _ := time.Parse(time.RFC3339Nano, scoredAtStr)
			scoredBy, _ := data["scoredBy"].(string)

			evt := events.NewApplicationFitScoreSet(events.ApplicationFitScoreSetParams{
				ID:          id,
				ComponentID: componentID,
				PillarID:    pillarID,
				PillarName:  pillarName,
				Score:       int(score),
				Rationale:   rationale,
				ScoredBy:    scoredBy,
			})
			evt.ScoredAt = scoredAt
			return evt
		},
		"ApplicationFitScoreUpdated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			score, _ := data["score"].(float64)
			rationale, _ := data["rationale"].(string)
			oldScore, _ := data["oldScore"].(float64)
			oldRationale, _ := data["oldRationale"].(string)
			updatedAtStr, _ := data["updatedAt"].(string)
			updatedAt, _ := time.Parse(time.RFC3339Nano, updatedAtStr)
			updatedBy, _ := data["updatedBy"].(string)

			evt := events.NewApplicationFitScoreUpdated(events.ApplicationFitScoreUpdatedParams{
				ID:           id,
				Score:        int(score),
				Rationale:    rationale,
				OldScore:     int(oldScore),
				OldRationale: oldRationale,
				UpdatedBy:    updatedBy,
			})
			evt.UpdatedAt = updatedAt
			return evt
		},
		"ApplicationFitScoreRemoved": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			componentID, _ := data["componentId"].(string)
			pillarID, _ := data["pillarId"].(string)
			removedBy, _ := data["removedBy"].(string)

			return events.NewApplicationFitScoreRemoved(id, componentID, pillarID, removedBy)
		},
	},
)
