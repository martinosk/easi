package repositories

import (
	"errors"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
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
		"CapabilityAssignedToDomain":     repository.JSONDeserializer[events.CapabilityAssignedToDomain],
		"CapabilityUnassignedFromDomain": repository.JSONDeserializer[events.CapabilityUnassignedFromDomain],
	},
)
