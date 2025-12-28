package repositories

import (
	"errors"
	"time"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrCapabilityNotFound = errors.New("capability not found")

type CapabilityRepository struct {
	*repository.EventSourcedRepository[*aggregates.Capability]
}

func NewCapabilityRepository(eventStore eventstore.EventStore) *CapabilityRepository {
	return &CapabilityRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			capabilityEventDeserializers,
			aggregates.LoadCapabilityFromHistory,
			ErrCapabilityNotFound,
		),
	}
}

var capabilityEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"CapabilityCreated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			name, _ := data["name"].(string)
			description, _ := data["description"].(string)
			parentID, _ := data["parentId"].(string)
			level, _ := data["level"].(string)
			createdAtStr, _ := data["createdAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

			evt := events.NewCapabilityCreated(id, name, description, parentID, level)
			evt.CreatedAt = createdAt
			return evt
		},
		"CapabilityUpdated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			name, _ := data["name"].(string)
			description, _ := data["description"].(string)

			return events.NewCapabilityUpdated(id, name, description)
		},
		"CapabilityMetadataUpdated": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			strategyPillar, _ := data["strategyPillar"].(string)
			pillarWeightFloat, _ := data["pillarWeight"].(float64)
			pillarWeight := int(pillarWeightFloat)
			maturityValueFloat, _ := data["maturityValue"].(float64)
			maturityValue := int(maturityValueFloat)
			ownershipModel, _ := data["ownershipModel"].(string)
			primaryOwner, _ := data["primaryOwner"].(string)
			eaOwner, _ := data["eaOwner"].(string)
			status, _ := data["status"].(string)

			return events.NewCapabilityMetadataUpdated(id, strategyPillar, pillarWeight, maturityValue, ownershipModel, primaryOwner, eaOwner, status)
		},
		"CapabilityExpertAdded": func(data map[string]interface{}) domain.DomainEvent {
			capabilityID, _ := data["capabilityId"].(string)
			expertName, _ := data["expertName"].(string)
			expertRole, _ := data["expertRole"].(string)
			contactInfo, _ := data["contactInfo"].(string)

			return events.NewCapabilityExpertAdded(capabilityID, expertName, expertRole, contactInfo)
		},
		"CapabilityTagAdded": func(data map[string]interface{}) domain.DomainEvent {
			capabilityID, _ := data["capabilityId"].(string)
			tag, _ := data["tag"].(string)

			return events.NewCapabilityTagAdded(capabilityID, tag)
		},
		"CapabilityParentChanged": func(data map[string]interface{}) domain.DomainEvent {
			capabilityID, _ := data["capabilityId"].(string)
			oldParentID, _ := data["oldParentId"].(string)
			newParentID, _ := data["newParentId"].(string)
			oldLevel, _ := data["oldLevel"].(string)
			newLevel, _ := data["newLevel"].(string)

			return events.NewCapabilityParentChanged(capabilityID, oldParentID, newParentID, oldLevel, newLevel)
		},
	},
	CapabilityMetadataUpdatedV1ToV2Upcaster{},
)
