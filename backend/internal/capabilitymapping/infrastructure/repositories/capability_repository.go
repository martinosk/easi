package repositories

import (
	"errors"

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
		"CapabilityCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			description, err := repository.GetRequiredString(data, "description")
			if err != nil {
				return nil, err
			}
			parentID, err := repository.GetOptionalString(data, "parentId", "")
			if err != nil {
				return nil, err
			}
			level, err := repository.GetRequiredString(data, "level")
			if err != nil {
				return nil, err
			}
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewCapabilityCreated(id, name, description, parentID, level)
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"CapabilityUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			description, err := repository.GetRequiredString(data, "description")
			if err != nil {
				return nil, err
			}

			return events.NewCapabilityUpdated(id, name, description), nil
		},
		"CapabilityMetadataUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			strategyPillar, err := repository.GetOptionalString(data, "strategyPillar", "")
			if err != nil {
				return nil, err
			}
			pillarWeight, err := repository.GetOptionalInt(data, "pillarWeight", 0)
			if err != nil {
				return nil, err
			}
			maturityValue, err := repository.GetOptionalInt(data, "maturityValue", 0)
			if err != nil {
				return nil, err
			}
			ownershipModel, err := repository.GetOptionalString(data, "ownershipModel", "")
			if err != nil {
				return nil, err
			}
			primaryOwner, err := repository.GetOptionalString(data, "primaryOwner", "")
			if err != nil {
				return nil, err
			}
			eaOwner, err := repository.GetOptionalString(data, "eaOwner", "")
			if err != nil {
				return nil, err
			}
			status, err := repository.GetOptionalString(data, "status", "")
			if err != nil {
				return nil, err
			}

			return events.NewCapabilityMetadataUpdated(id, strategyPillar, pillarWeight, maturityValue, ownershipModel, primaryOwner, eaOwner, status), nil
		},
		"CapabilityExpertAdded": func(data map[string]interface{}) (domain.DomainEvent, error) {
			capabilityID, err := repository.GetRequiredString(data, "capabilityId")
			if err != nil {
				return nil, err
			}
			expertName, err := repository.GetRequiredString(data, "expertName")
			if err != nil {
				return nil, err
			}
			expertRole, err := repository.GetRequiredString(data, "expertRole")
			if err != nil {
				return nil, err
			}
			contactInfo, err := repository.GetOptionalString(data, "contactInfo", "")
			if err != nil {
				return nil, err
			}

			return events.NewCapabilityExpertAdded(capabilityID, expertName, expertRole, contactInfo), nil
		},
		"CapabilityExpertRemoved": func(data map[string]interface{}) (domain.DomainEvent, error) {
			capabilityID, err := repository.GetRequiredString(data, "capabilityId")
			if err != nil {
				return nil, err
			}
			expertName, err := repository.GetRequiredString(data, "expertName")
			if err != nil {
				return nil, err
			}
			expertRole, err := repository.GetRequiredString(data, "expertRole")
			if err != nil {
				return nil, err
			}
			contactInfo, err := repository.GetOptionalString(data, "contactInfo", "")
			if err != nil {
				return nil, err
			}

			return events.NewCapabilityExpertRemoved(capabilityID, expertName, expertRole, contactInfo), nil
		},
		"CapabilityTagAdded": func(data map[string]interface{}) (domain.DomainEvent, error) {
			capabilityID, err := repository.GetRequiredString(data, "capabilityId")
			if err != nil {
				return nil, err
			}
			tag, err := repository.GetRequiredString(data, "tag")
			if err != nil {
				return nil, err
			}

			return events.NewCapabilityTagAdded(capabilityID, tag), nil
		},
		"CapabilityParentChanged": func(data map[string]interface{}) (domain.DomainEvent, error) {
			capabilityID, err := repository.GetRequiredString(data, "capabilityId")
			if err != nil {
				return nil, err
			}
			oldParentID, err := repository.GetOptionalString(data, "oldParentId", "")
			if err != nil {
				return nil, err
			}
			newParentID, err := repository.GetOptionalString(data, "newParentId", "")
			if err != nil {
				return nil, err
			}
			oldLevel, err := repository.GetRequiredString(data, "oldLevel")
			if err != nil {
				return nil, err
			}
			newLevel, err := repository.GetRequiredString(data, "newLevel")
			if err != nil {
				return nil, err
			}

			return events.NewCapabilityParentChanged(capabilityID, oldParentID, newParentID, oldLevel, newLevel), nil
		},
	},
	CapabilityMetadataUpdatedV1ToV2Upcaster{},
)
