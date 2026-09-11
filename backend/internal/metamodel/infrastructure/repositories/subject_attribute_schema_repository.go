package repositories

import (
	"errors"

	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/events"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrSubjectAttributeSchemaNotFound = errors.New("subject attribute schema not found")

type SubjectAttributeSchemaRepository struct {
	*repository.EventSourcedRepository[*aggregates.SubjectAttributeSchema]
}

func NewSubjectAttributeSchemaRepository(eventStore eventstore.EventStore) *SubjectAttributeSchemaRepository {
	return &SubjectAttributeSchemaRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			subjectAttributeSchemaEventDeserializers,
			aggregates.LoadSubjectAttributeSchemaFromHistory,
			ErrSubjectAttributeSchemaNotFound,
		),
	}
}

var subjectAttributeSchemaEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		events.TypeSubjectAttributeSchemaCreated: repository.JSONDeserializer[events.SubjectAttributeSchemaCreated],
		events.TypeSubjectAttributeDefined:       repository.JSONDeserializer[events.SubjectAttributeDefined],
		events.TypeSubjectAttributeRenamed:       repository.JSONDeserializer[events.SubjectAttributeRenamed],
		events.TypeSubjectAttributeRetired:       repository.JSONDeserializer[events.SubjectAttributeRetired],
		events.TypeSubjectAttributeReactivated:   repository.JSONDeserializer[events.SubjectAttributeReactivated],
		events.TypeSubjectAttributeOptionAdded:   repository.JSONDeserializer[events.SubjectAttributeOptionAdded],
		events.TypeSubjectAttributeOptionRetired: repository.JSONDeserializer[events.SubjectAttributeOptionRetired],
		events.TypeSubjectAttributeBoundsChanged: repository.JSONDeserializer[events.SubjectAttributeBoundsChanged],
	},
)
