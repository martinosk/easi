package repositories

import (
	"context"
	"errors"
	"time"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/domain"
)

var (
	ErrDependencyNotFound = errors.New("dependency not found")
)

type DependencyRepository struct {
	eventStore eventstore.EventStore
}

func NewDependencyRepository(eventStore eventstore.EventStore) *DependencyRepository {
	return &DependencyRepository{
		eventStore: eventStore,
	}
}

func (r *DependencyRepository) Save(ctx context.Context, dependency *aggregates.CapabilityDependency) error {
	uncommittedEvents := dependency.GetUncommittedChanges()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	err := r.eventStore.SaveEvents(ctx, dependency.ID(), uncommittedEvents, dependency.Version()-len(uncommittedEvents))
	if err != nil {
		return err
	}

	dependency.MarkChangesAsCommitted()
	return nil
}

func (r *DependencyRepository) GetByID(ctx context.Context, id string) (*aggregates.CapabilityDependency, error) {
	storedEvents, err := r.eventStore.GetEvents(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(storedEvents) == 0 {
		return nil, ErrDependencyNotFound
	}

	domainEvents, err := r.deserializeEvents(storedEvents)
	if err != nil {
		return nil, err
	}

	return aggregates.LoadCapabilityDependencyFromHistory(domainEvents)
}

func (r *DependencyRepository) deserializeEvents(storedEvents []domain.DomainEvent) ([]domain.DomainEvent, error) {
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		eventData := event.EventData()

		switch event.EventType() {
		case "CapabilityDependencyCreated":
			id, _ := eventData["id"].(string)
			sourceCapabilityID, _ := eventData["sourceCapabilityId"].(string)
			targetCapabilityID, _ := eventData["targetCapabilityId"].(string)
			dependencyType, _ := eventData["dependencyType"].(string)
			description, _ := eventData["description"].(string)
			createdAtStr, _ := eventData["createdAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

			concreteEvent := events.NewCapabilityDependencyCreated(id, sourceCapabilityID, targetCapabilityID, dependencyType, description)
			concreteEvent.CreatedAt = createdAt
			domainEvents = append(domainEvents, concreteEvent)

		case "CapabilityDependencyDeleted":
			id, _ := eventData["id"].(string)
			deletedAtStr, _ := eventData["deletedAt"].(string)
			deletedAt, _ := time.Parse(time.RFC3339Nano, deletedAtStr)

			concreteEvent := events.NewCapabilityDependencyDeleted(id)
			concreteEvent.DeletedAt = deletedAt
			domainEvents = append(domainEvents, concreteEvent)

		default:
			continue
		}
	}

	return domainEvents, nil
}
