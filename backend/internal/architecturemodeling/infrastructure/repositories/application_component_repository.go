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
	// ErrComponentNotFound is returned when a component is not found
	ErrComponentNotFound = errors.New("component not found")
)

// ApplicationComponentRepository manages persistence of application components
type ApplicationComponentRepository struct {
	eventStore eventstore.EventStore
}

// NewApplicationComponentRepository creates a new repository
func NewApplicationComponentRepository(eventStore eventstore.EventStore) *ApplicationComponentRepository {
	return &ApplicationComponentRepository{
		eventStore: eventStore,
	}
}

// Save persists an application component aggregate
func (r *ApplicationComponentRepository) Save(ctx context.Context, component *aggregates.ApplicationComponent) error {
	uncommittedEvents := component.GetUncommittedChanges()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	err := r.eventStore.SaveEvents(ctx, component.ID(), uncommittedEvents, component.Version()-len(uncommittedEvents))
	if err != nil {
		return err
	}

	component.MarkChangesAsCommitted()
	return nil
}

// GetByID retrieves an application component by ID
func (r *ApplicationComponentRepository) GetByID(ctx context.Context, id string) (*aggregates.ApplicationComponent, error) {
	storedEvents, err := r.eventStore.GetEvents(ctx, id)
	if err != nil {
		return nil, err
	}

	// If no events found, the aggregate doesn't exist
	if len(storedEvents) == 0 {
		return nil, ErrComponentNotFound
	}

	// Deserialize events (simplified - would use event registry in production)
	domainEvents, err := r.deserializeEvents(storedEvents)
	if err != nil {
		return nil, err
	}

	return aggregates.LoadApplicationComponentFromHistory(domainEvents)
}

// deserializeEvents converts stored events to domain events
func (r *ApplicationComponentRepository) deserializeEvents(storedEvents []domain.DomainEvent) ([]domain.DomainEvent, error) {
	// Convert generic events back to concrete event types
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		eventData := event.EventData()

		switch event.EventType() {
		case "ApplicationComponentCreated":
			// Extract fields from event data
			id, _ := eventData["id"].(string)
			name, _ := eventData["name"].(string)
			description, _ := eventData["description"].(string)
			createdAtStr, _ := eventData["createdAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

			// Create concrete event
			concreteEvent := events.NewApplicationComponentCreated(id, name, description)
			concreteEvent.CreatedAt = createdAt
			domainEvents = append(domainEvents, concreteEvent)

		case "ApplicationComponentUpdated":
			// Extract fields from event data
			id, _ := eventData["id"].(string)
			name, _ := eventData["name"].(string)
			description, _ := eventData["description"].(string)

			// Create concrete event
			concreteEvent := events.NewApplicationComponentUpdated(id, name, description)
			domainEvents = append(domainEvents, concreteEvent)

		case "ApplicationComponentDeleted":
			// Extract fields from event data
			id, _ := eventData["id"].(string)
			name, _ := eventData["name"].(string)

			// Create concrete event
			concreteEvent := events.NewApplicationComponentDeleted(id, name)
			domainEvents = append(domainEvents, concreteEvent)

		default:
			// Unknown event type, skip it
			continue
		}
	}

	return domainEvents, nil
}
