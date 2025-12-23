package repositories

import (
	"errors"
	"time"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/eventsourcing"
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

var relationEventDeserializers = repository.EventDeserializers{
	"ComponentRelationCreated": func(data map[string]interface{}) domain.DomainEvent {
		id, _ := data["id"].(string)
		sourceComponentID, _ := data["sourceComponentId"].(string)
		targetComponentID, _ := data["targetComponentId"].(string)
		relationType, _ := data["relationType"].(string)
		name, _ := data["name"].(string)
		description, _ := data["description"].(string)
		createdAtStr, _ := data["createdAt"].(string)
		createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

		evt := events.NewComponentRelationCreated(events.ComponentRelationParams{
			ID:          id,
			SourceID:    sourceComponentID,
			TargetID:    targetComponentID,
			Type:        relationType,
			Name:        name,
			Description: description,
		})
		evt.CreatedAt = createdAt
		return evt
	},
	"ComponentRelationUpdated": func(data map[string]interface{}) domain.DomainEvent {
		id, _ := data["id"].(string)
		name, _ := data["name"].(string)
		description, _ := data["description"].(string)

		return events.NewComponentRelationUpdated(id, name, description)
	},
}
