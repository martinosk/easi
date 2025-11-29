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
	ErrAssignmentNotFound = errors.New("business domain assignment not found")
)

type BusinessDomainAssignmentRepository struct {
	eventStore eventstore.EventStore
}

func NewBusinessDomainAssignmentRepository(eventStore eventstore.EventStore) *BusinessDomainAssignmentRepository {
	return &BusinessDomainAssignmentRepository{
		eventStore: eventStore,
	}
}

func (r *BusinessDomainAssignmentRepository) Save(ctx context.Context, assignment *aggregates.BusinessDomainAssignment) error {
	uncommittedEvents := assignment.GetUncommittedChanges()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	err := r.eventStore.SaveEvents(ctx, assignment.ID(), uncommittedEvents, assignment.Version()-len(uncommittedEvents))
	if err != nil {
		return err
	}

	assignment.MarkChangesAsCommitted()
	return nil
}

func (r *BusinessDomainAssignmentRepository) GetByID(ctx context.Context, id string) (*aggregates.BusinessDomainAssignment, error) {
	storedEvents, err := r.eventStore.GetEvents(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(storedEvents) == 0 {
		return nil, ErrAssignmentNotFound
	}

	domainEvents, err := r.deserializeEvents(storedEvents)
	if err != nil {
		return nil, err
	}

	return aggregates.LoadBusinessDomainAssignmentFromHistory(domainEvents)
}

func (r *BusinessDomainAssignmentRepository) deserializeEvents(storedEvents []domain.DomainEvent) ([]domain.DomainEvent, error) {
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		eventData := event.EventData()

		switch event.EventType() {
		case "CapabilityAssignedToDomain":
			id, _ := eventData["id"].(string)
			businessDomainID, _ := eventData["businessDomainId"].(string)
			capabilityID, _ := eventData["capabilityId"].(string)
			assignedAtStr, _ := eventData["assignedAt"].(string)
			assignedAt, _ := time.Parse(time.RFC3339Nano, assignedAtStr)

			concreteEvent := events.NewCapabilityAssignedToDomain(id, businessDomainID, capabilityID)
			concreteEvent.AssignedAt = assignedAt
			domainEvents = append(domainEvents, concreteEvent)

		case "CapabilityUnassignedFromDomain":
			id, _ := eventData["id"].(string)
			businessDomainID, _ := eventData["businessDomainId"].(string)
			capabilityID, _ := eventData["capabilityId"].(string)

			concreteEvent := events.NewCapabilityUnassignedFromDomain(id, businessDomainID, capabilityID)
			domainEvents = append(domainEvents, concreteEvent)

		default:
			continue
		}
	}

	return domainEvents, nil
}
