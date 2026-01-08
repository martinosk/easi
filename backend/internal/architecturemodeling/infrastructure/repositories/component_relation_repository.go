package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrRelationNotFound = errors.New("relation not found")

type ComponentRelationRepository struct {
	*repository.EventSourcedRepository[*aggregates.ComponentRelation]
}

func NewComponentRelationRepository(eventStore eventstore.EventStore) *ComponentRelationRepository {
	return &ComponentRelationRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			relationEventDeserializers,
			aggregates.LoadComponentRelationFromHistory,
			ErrRelationNotFound,
		),
	}
}

var relationEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"ComponentRelationCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			sourceComponentID, err := repository.GetRequiredString(data, "sourceComponentId")
			if err != nil {
				return nil, err
			}
			targetComponentID, err := repository.GetRequiredString(data, "targetComponentId")
			if err != nil {
				return nil, err
			}
			relationType, err := repository.GetRequiredString(data, "relationType")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			description, err := repository.GetRequiredString(data, "description")
			if err != nil {
				return nil, err
			}
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewComponentRelationCreated(events.ComponentRelationParams{
				ID:          id,
				SourceID:    sourceComponentID,
				TargetID:    targetComponentID,
				Type:        relationType,
				Name:        name,
				Description: description,
			})
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"ComponentRelationUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			description, err := repository.GetRequiredString(data, "description")
			if err != nil {
				return nil, err
			}

			return events.NewComponentRelationUpdated(id, name, description), nil
		},
	},
)
