package repositories

import (
	"errors"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrAssignmentNotFound = errors.New("business domain assignment not found")

type BusinessDomainAssignmentRepository struct {
	*repository.EventSourcedRepository[*aggregates.BusinessDomainAssignment]
}

func NewBusinessDomainAssignmentRepository(eventStore eventstore.EventStore) *BusinessDomainAssignmentRepository {
	return &BusinessDomainAssignmentRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			assignmentEventDeserializers,
			aggregates.LoadBusinessDomainAssignmentFromHistory,
			ErrAssignmentNotFound,
		),
	}
}

var assignmentEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"CapabilityAssignedToDomain": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			businessDomainID, err := repository.GetRequiredString(data, "businessDomainId")
			if err != nil {
				return nil, err
			}
			capabilityID, err := repository.GetRequiredString(data, "capabilityId")
			if err != nil {
				return nil, err
			}
			assignedAt, err := repository.GetRequiredTime(data, "assignedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewCapabilityAssignedToDomain(id, businessDomainID, capabilityID)
			evt.AssignedAt = assignedAt
			return evt, nil
		},
		"CapabilityUnassignedFromDomain": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			businessDomainID, err := repository.GetRequiredString(data, "businessDomainId")
			if err != nil {
				return nil, err
			}
			capabilityID, err := repository.GetRequiredString(data, "capabilityId")
			if err != nil {
				return nil, err
			}

			return events.NewCapabilityUnassignedFromDomain(id, businessDomainID, capabilityID), nil
		},
	},
)
