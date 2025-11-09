package repositories

import (
	"context"

	"github.com/easi/backend/internal/architectureviews/domain/aggregates"
	"github.com/easi/backend/internal/infrastructure/eventstore"
	"github.com/easi/backend/internal/shared/domain"
)

// ArchitectureViewRepository manages persistence of architecture views
type ArchitectureViewRepository struct {
	eventStore eventstore.EventStore
}

// NewArchitectureViewRepository creates a new repository
func NewArchitectureViewRepository(eventStore eventstore.EventStore) *ArchitectureViewRepository {
	return &ArchitectureViewRepository{
		eventStore: eventStore,
	}
}

// Save persists an architecture view aggregate
func (r *ArchitectureViewRepository) Save(ctx context.Context, view *aggregates.ArchitectureView) error {
	uncommittedEvents := view.GetUncommittedChanges()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	err := r.eventStore.SaveEvents(ctx, view.ID(), uncommittedEvents, view.Version()-len(uncommittedEvents))
	if err != nil {
		return err
	}

	view.MarkChangesAsCommitted()
	return nil
}

// GetByID retrieves an architecture view by ID
func (r *ArchitectureViewRepository) GetByID(ctx context.Context, id string) (*aggregates.ArchitectureView, error) {
	storedEvents, err := r.eventStore.GetEvents(ctx, id)
	if err != nil {
		return nil, err
	}

	// Deserialize events (simplified)
	domainEvents, err := r.deserializeEvents(storedEvents)
	if err != nil {
		return nil, err
	}

	return aggregates.LoadArchitectureViewFromHistory(domainEvents)
}

// deserializeEvents converts stored events to domain events
func (r *ArchitectureViewRepository) deserializeEvents(storedEvents []domain.DomainEvent) ([]domain.DomainEvent, error) {
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		domainEvents = append(domainEvents, event)
	}

	return domainEvents, nil
}
