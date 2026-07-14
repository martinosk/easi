package repositories

import (
	"errors"

	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/events"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrTimeAssessmentNotFound = errors.New("time assessment not found")

type TimeAssessmentRepository struct {
	*repository.EventSourcedRepository[*aggregates.TimeAssessment]
}

func NewTimeAssessmentRepository(eventStore eventstore.EventStore) *TimeAssessmentRepository {
	return &TimeAssessmentRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			timeAssessmentEventDeserializers,
			aggregates.LoadTimeAssessmentFromHistory,
			ErrTimeAssessmentNotFound,
		),
	}
}

var timeAssessmentEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		pl.TimeAssessmentRecorded: repository.JSONDeserializer[events.TimeAssessmentRecorded],
		pl.TimeAssessmentRemoved:  repository.JSONDeserializer[events.TimeAssessmentRemoved],
	},
)
