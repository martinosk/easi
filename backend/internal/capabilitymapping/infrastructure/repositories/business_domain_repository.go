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
	ErrBusinessDomainNotFound = errors.New("business domain not found")
)

type BusinessDomainRepository struct {
	eventStore eventstore.EventStore
}

func NewBusinessDomainRepository(eventStore eventstore.EventStore) *BusinessDomainRepository {
	return &BusinessDomainRepository{
		eventStore: eventStore,
	}
}

func (r *BusinessDomainRepository) Save(ctx context.Context, businessDomain *aggregates.BusinessDomain) error {
	uncommittedEvents := businessDomain.GetUncommittedChanges()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	err := r.eventStore.SaveEvents(ctx, businessDomain.ID(), uncommittedEvents, businessDomain.Version()-len(uncommittedEvents))
	if err != nil {
		return err
	}

	businessDomain.MarkChangesAsCommitted()
	return nil
}

func (r *BusinessDomainRepository) GetByID(ctx context.Context, id string) (*aggregates.BusinessDomain, error) {
	storedEvents, err := r.eventStore.GetEvents(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(storedEvents) == 0 {
		return nil, ErrBusinessDomainNotFound
	}

	domainEvents, err := r.deserializeEvents(storedEvents)
	if err != nil {
		return nil, err
	}

	return aggregates.LoadBusinessDomainFromHistory(domainEvents)
}

func (r *BusinessDomainRepository) deserializeEvents(storedEvents []domain.DomainEvent) ([]domain.DomainEvent, error) {
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		eventData := event.EventData()

		switch event.EventType() {
		case "BusinessDomainCreated":
			id, _ := eventData["id"].(string)
			name, _ := eventData["name"].(string)
			description, _ := eventData["description"].(string)
			createdAtStr, _ := eventData["createdAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

			concreteEvent := events.NewBusinessDomainCreated(id, name, description)
			concreteEvent.CreatedAt = createdAt
			domainEvents = append(domainEvents, concreteEvent)

		case "BusinessDomainUpdated":
			id, _ := eventData["id"].(string)
			name, _ := eventData["name"].(string)
			description, _ := eventData["description"].(string)

			concreteEvent := events.NewBusinessDomainUpdated(id, name, description)
			domainEvents = append(domainEvents, concreteEvent)

		case "BusinessDomainDeleted":
			id, _ := eventData["id"].(string)

			concreteEvent := events.NewBusinessDomainDeleted(id)
			domainEvents = append(domainEvents, concreteEvent)

		default:
			continue
		}
	}

	return domainEvents, nil
}
