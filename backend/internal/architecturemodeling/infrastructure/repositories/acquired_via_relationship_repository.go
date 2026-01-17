package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrAcquiredViaRelationshipNotFound = errors.New("acquired via relationship not found")

type AcquiredViaRelationshipRepository struct {
	*repository.EventSourcedRepository[*aggregates.AcquiredViaRelationship]
}

func NewAcquiredViaRelationshipRepository(eventStore eventstore.EventStore) *AcquiredViaRelationshipRepository {
	return &AcquiredViaRelationshipRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			acquiredViaRelationshipEventDeserializers,
			aggregates.LoadAcquiredViaRelationshipFromHistory,
			ErrAcquiredViaRelationshipNotFound,
		),
	}
}

var acquiredViaRelationshipEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"AcquiredViaRelationshipCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			acquiredEntityID, err := repository.GetRequiredString(data, "acquiredEntityId")
			if err != nil {
				return nil, err
			}
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			notes, _ := repository.GetOptionalString(data, "notes", "")
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewAcquiredViaRelationshipCreated(id, acquiredEntityID, componentID, notes)
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"AcquiredViaRelationshipDeleted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			acquiredEntityID, err := repository.GetRequiredString(data, "acquiredEntityId")
			if err != nil {
				return nil, err
			}
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}

			return events.NewAcquiredViaRelationshipDeleted(id, acquiredEntityID, componentID), nil
		},
	},
)
