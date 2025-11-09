package repositories

import (
	"context"

	"github.com/easi/backend/internal/architecturemodeling/domain/aggregates"
	"github.com/easi/backend/internal/infrastructure/eventstore"
	"github.com/easi/backend/internal/shared/domain"
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

	// Deserialize events (simplified - would use event registry in production)
	domainEvents, err := r.deserializeEvents(storedEvents)
	if err != nil {
		return nil, err
	}

	return aggregates.LoadApplicationComponentFromHistory(domainEvents)
}

// deserializeEvents converts stored events to domain events
func (r *ApplicationComponentRepository) deserializeEvents(storedEvents []domain.DomainEvent) ([]domain.DomainEvent, error) {
	// This is a simplified implementation
	// In production, you would use an event type registry
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		// For now, we'll just return the events as-is
		// In a full implementation, you would deserialize based on event type
		domainEvents = append(domainEvents, event)
	}

	return domainEvents, nil
}
