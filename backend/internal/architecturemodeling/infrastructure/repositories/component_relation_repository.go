package repositories

import (
	"context"
	"errors"
	"time"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/domain"
)

var (
	// ErrRelationNotFound is returned when a relation is not found
	ErrRelationNotFound = errors.New("relation not found")
)

// ComponentRelationRepository manages persistence of component relations
type ComponentRelationRepository struct {
	eventStore eventstore.EventStore
}

// NewComponentRelationRepository creates a new repository
func NewComponentRelationRepository(eventStore eventstore.EventStore) *ComponentRelationRepository {
	return &ComponentRelationRepository{
		eventStore: eventStore,
	}
}

// Save persists a component relation aggregate
func (r *ComponentRelationRepository) Save(ctx context.Context, relation *aggregates.ComponentRelation) error {
	uncommittedEvents := relation.GetUncommittedChanges()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	err := r.eventStore.SaveEvents(ctx, relation.ID(), uncommittedEvents, relation.Version()-len(uncommittedEvents))
	if err != nil {
		return err
	}

	relation.MarkChangesAsCommitted()
	return nil
}

// GetByID retrieves a component relation by ID
func (r *ComponentRelationRepository) GetByID(ctx context.Context, id string) (*aggregates.ComponentRelation, error) {
	storedEvents, err := r.eventStore.GetEvents(ctx, id)
	if err != nil {
		return nil, err
	}

	// If no events found, the aggregate doesn't exist
	if len(storedEvents) == 0 {
		return nil, ErrRelationNotFound
	}

	// Deserialize events (simplified)
	domainEvents := r.deserializeEvents(storedEvents)

	return aggregates.LoadComponentRelationFromHistory(domainEvents)
}

// deserializeEvents converts stored events to domain events
func (r *ComponentRelationRepository) deserializeEvents(storedEvents []domain.DomainEvent) []domain.DomainEvent {
	// Convert generic events back to concrete event types
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		eventData := event.EventData()

		switch event.EventType() {
		case "ComponentRelationCreated":
			// Extract fields from event data
			id, _ := eventData["id"].(string)
			sourceComponentID, _ := eventData["sourceComponentId"].(string)
			targetComponentID, _ := eventData["targetComponentId"].(string)
			relationType, _ := eventData["relationType"].(string)
			name, _ := eventData["name"].(string)
			description, _ := eventData["description"].(string)
			createdAtStr, _ := eventData["createdAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

			concreteEvent := events.NewComponentRelationCreated(events.ComponentRelationParams{
				ID:          id,
				SourceID:    sourceComponentID,
				TargetID:    targetComponentID,
				Type:        relationType,
				Name:        name,
				Description: description,
			})
			concreteEvent.CreatedAt = createdAt
			domainEvents = append(domainEvents, concreteEvent)

		case "ComponentRelationUpdated":
			// Extract fields from event data
			id, _ := eventData["id"].(string)
			name, _ := eventData["name"].(string)
			description, _ := eventData["description"].(string)

			// Create concrete event
			concreteEvent := events.NewComponentRelationUpdated(id, name, description)
			domainEvents = append(domainEvents, concreteEvent)

		default:
			// Unknown event type, skip it
			continue
		}
	}

	return domainEvents
}
