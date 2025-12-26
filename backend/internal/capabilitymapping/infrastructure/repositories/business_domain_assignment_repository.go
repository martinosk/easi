package repositories

import (
	"errors"
	"time"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/eventsourcing"
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
		"CapabilityAssignedToDomain": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			businessDomainID, _ := data["businessDomainId"].(string)
			capabilityID, _ := data["capabilityId"].(string)
			assignedAtStr, _ := data["assignedAt"].(string)
			assignedAt, _ := time.Parse(time.RFC3339Nano, assignedAtStr)

			evt := events.NewCapabilityAssignedToDomain(id, businessDomainID, capabilityID)
			evt.AssignedAt = assignedAt
			return evt
		},
		"CapabilityUnassignedFromDomain": func(data map[string]interface{}) domain.DomainEvent {
			id, _ := data["id"].(string)
			businessDomainID, _ := data["businessDomainId"].(string)
			capabilityID, _ := data["capabilityId"].(string)

			return events.NewCapabilityUnassignedFromDomain(id, businessDomainID, capabilityID)
		},
	},
)
