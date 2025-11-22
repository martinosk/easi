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
	ErrCapabilityNotFound = errors.New("capability not found")
)

type CapabilityRepository struct {
	eventStore eventstore.EventStore
}

func NewCapabilityRepository(eventStore eventstore.EventStore) *CapabilityRepository {
	return &CapabilityRepository{
		eventStore: eventStore,
	}
}

func (r *CapabilityRepository) Save(ctx context.Context, capability *aggregates.Capability) error {
	uncommittedEvents := capability.GetUncommittedChanges()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	err := r.eventStore.SaveEvents(ctx, capability.ID(), uncommittedEvents, capability.Version()-len(uncommittedEvents))
	if err != nil {
		return err
	}

	capability.MarkChangesAsCommitted()
	return nil
}

func (r *CapabilityRepository) GetByID(ctx context.Context, id string) (*aggregates.Capability, error) {
	storedEvents, err := r.eventStore.GetEvents(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(storedEvents) == 0 {
		return nil, ErrCapabilityNotFound
	}

	domainEvents, err := r.deserializeEvents(storedEvents)
	if err != nil {
		return nil, err
	}

	return aggregates.LoadCapabilityFromHistory(domainEvents)
}

func (r *CapabilityRepository) deserializeEvents(storedEvents []domain.DomainEvent) ([]domain.DomainEvent, error) {
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		eventData := event.EventData()

		switch event.EventType() {
		case "CapabilityCreated":
			id, _ := eventData["id"].(string)
			name, _ := eventData["name"].(string)
			description, _ := eventData["description"].(string)
			parentID, _ := eventData["parentId"].(string)
			level, _ := eventData["level"].(string)
			createdAtStr, _ := eventData["createdAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

			concreteEvent := events.NewCapabilityCreated(id, name, description, parentID, level)
			concreteEvent.CreatedAt = createdAt
			domainEvents = append(domainEvents, concreteEvent)

		case "CapabilityUpdated":
			id, _ := eventData["id"].(string)
			name, _ := eventData["name"].(string)
			description, _ := eventData["description"].(string)

			concreteEvent := events.NewCapabilityUpdated(id, name, description)
			domainEvents = append(domainEvents, concreteEvent)

		case "CapabilityMetadataUpdated":
			id, _ := eventData["id"].(string)
			strategyPillar, _ := eventData["strategyPillar"].(string)
			pillarWeightFloat, _ := eventData["pillarWeight"].(float64)
			pillarWeight := int(pillarWeightFloat)
			maturityLevel, _ := eventData["maturityLevel"].(string)
			ownershipModel, _ := eventData["ownershipModel"].(string)
			primaryOwner, _ := eventData["primaryOwner"].(string)
			eaOwner, _ := eventData["eaOwner"].(string)
			status, _ := eventData["status"].(string)

			concreteEvent := events.NewCapabilityMetadataUpdated(id, strategyPillar, pillarWeight, maturityLevel, ownershipModel, primaryOwner, eaOwner, status)
			domainEvents = append(domainEvents, concreteEvent)

		case "CapabilityExpertAdded":
			capabilityID, _ := eventData["capabilityId"].(string)
			expertName, _ := eventData["expertName"].(string)
			expertRole, _ := eventData["expertRole"].(string)
			contactInfo, _ := eventData["contactInfo"].(string)

			concreteEvent := events.NewCapabilityExpertAdded(capabilityID, expertName, expertRole, contactInfo)
			domainEvents = append(domainEvents, concreteEvent)

		case "CapabilityTagAdded":
			capabilityID, _ := eventData["capabilityId"].(string)
			tag, _ := eventData["tag"].(string)

			concreteEvent := events.NewCapabilityTagAdded(capabilityID, tag)
			domainEvents = append(domainEvents, concreteEvent)

		case "CapabilityParentChanged":
			capabilityID, _ := eventData["capabilityId"].(string)
			oldParentID, _ := eventData["oldParentId"].(string)
			newParentID, _ := eventData["newParentId"].(string)
			oldLevel, _ := eventData["oldLevel"].(string)
			newLevel, _ := eventData["newLevel"].(string)

			concreteEvent := events.NewCapabilityParentChanged(capabilityID, oldParentID, newParentID, oldLevel, newLevel)
			domainEvents = append(domainEvents, concreteEvent)

		default:
			continue
		}
	}

	return domainEvents, nil
}
