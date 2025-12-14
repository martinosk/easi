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
	ErrRealizationNotFound = errors.New("realization not found")
)

type RealizationRepository struct {
	eventStore eventstore.EventStore
}

func NewRealizationRepository(eventStore eventstore.EventStore) *RealizationRepository {
	return &RealizationRepository{
		eventStore: eventStore,
	}
}

func (r *RealizationRepository) Save(ctx context.Context, realization *aggregates.CapabilityRealization) error {
	uncommittedEvents := realization.GetUncommittedChanges()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	err := r.eventStore.SaveEvents(ctx, realization.ID(), uncommittedEvents, realization.Version()-len(uncommittedEvents))
	if err != nil {
		return err
	}

	realization.MarkChangesAsCommitted()
	return nil
}

func (r *RealizationRepository) GetByID(ctx context.Context, id string) (*aggregates.CapabilityRealization, error) {
	storedEvents, err := r.eventStore.GetEvents(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(storedEvents) == 0 {
		return nil, ErrRealizationNotFound
	}

	domainEvents := r.deserializeEvents(storedEvents)

	return aggregates.LoadCapabilityRealizationFromHistory(domainEvents)
}

func (r *RealizationRepository) deserializeEvents(storedEvents []domain.DomainEvent) []domain.DomainEvent {
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		eventData := event.EventData()

		switch event.EventType() {
		case "SystemLinkedToCapability":
			id, _ := eventData["id"].(string)
			capabilityID, _ := eventData["capabilityId"].(string)
			componentID, _ := eventData["componentId"].(string)
			realizationLevel, _ := eventData["realizationLevel"].(string)
			notes, _ := eventData["notes"].(string)
			linkedAtStr, _ := eventData["linkedAt"].(string)
			linkedAt, _ := time.Parse(time.RFC3339Nano, linkedAtStr)

			concreteEvent := events.NewSystemLinkedToCapability(id, capabilityID, componentID, realizationLevel, notes)
			concreteEvent.LinkedAt = linkedAt
			domainEvents = append(domainEvents, concreteEvent)

		case "SystemRealizationUpdated":
			id, _ := eventData["id"].(string)
			realizationLevel, _ := eventData["realizationLevel"].(string)
			notes, _ := eventData["notes"].(string)

			concreteEvent := events.NewSystemRealizationUpdated(id, realizationLevel, notes)
			domainEvents = append(domainEvents, concreteEvent)

		case "SystemRealizationDeleted":
			id, _ := eventData["id"].(string)
			deletedAtStr, _ := eventData["deletedAt"].(string)
			deletedAt, _ := time.Parse(time.RFC3339Nano, deletedAtStr)

			concreteEvent := events.NewSystemRealizationDeleted(id)
			concreteEvent.DeletedAt = deletedAt
			domainEvents = append(domainEvents, concreteEvent)

		default:
			continue
		}
	}

	return domainEvents
}
