package repositories

import (
	"context"

	"github.com/easi/backend/internal/architecturemodeling/domain/aggregates"
	"github.com/easi/backend/internal/infrastructure/eventstore"
	"github.com/easi/backend/internal/shared/domain"
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

	// Deserialize events (simplified)
	domainEvents, err := r.deserializeEvents(storedEvents)
	if err != nil {
		return nil, err
	}

	return aggregates.LoadComponentRelationFromHistory(domainEvents)
}

// deserializeEvents converts stored events to domain events
func (r *ComponentRelationRepository) deserializeEvents(storedEvents []domain.DomainEvent) ([]domain.DomainEvent, error) {
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		domainEvents = append(domainEvents, event)
	}

	return domainEvents, nil
}
