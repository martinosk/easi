package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrBuiltByRelationshipNotFound = errors.New("built by relationship not found")

type BuiltByRelationshipRepository struct {
	*repository.EventSourcedRepository[*aggregates.BuiltByRelationship]
}

func NewBuiltByRelationshipRepository(eventStore eventstore.EventStore) *BuiltByRelationshipRepository {
	return &BuiltByRelationshipRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			builtByRelationshipEventDeserializers,
			aggregates.LoadBuiltByRelationshipFromHistory,
			ErrBuiltByRelationshipNotFound,
		),
	}
}

var builtByRelationshipEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"BuiltByRelationshipCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			internalTeamID, err := repository.GetRequiredString(data, "internalTeamId")
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

			evt := events.NewBuiltByRelationshipCreated(id, internalTeamID, componentID, notes)
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"BuiltByRelationshipDeleted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			internalTeamID, err := repository.GetRequiredString(data, "internalTeamId")
			if err != nil {
				return nil, err
			}
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}

			return events.NewBuiltByRelationshipDeleted(id, internalTeamID, componentID), nil
		},
	},
)
