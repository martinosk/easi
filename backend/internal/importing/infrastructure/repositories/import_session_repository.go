package repositories

import (
	"errors"

	"easi/backend/internal/importing/domain/aggregates"
	"easi/backend/internal/importing/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var (
	ErrImportSessionNotFound = errors.New("import session not found")
)

type ImportSessionRepository struct {
	*repository.EventSourcedRepository[*aggregates.ImportSession]
}

func NewImportSessionRepository(eventStore eventstore.EventStore) *ImportSessionRepository {
	return &ImportSessionRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			importSessionEventDeserializers,
			aggregates.LoadImportSessionFromHistory,
			ErrImportSessionNotFound,
		),
	}
}

var importSessionEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"ImportSessionCreated":   repository.JSONDeserializer[events.ImportSessionCreated],
		"ImportStarted":          repository.JSONDeserializer[events.ImportStarted],
		"ImportProgressUpdated":  repository.JSONDeserializer[events.ImportProgressUpdated],
		"ImportCompleted":        repository.JSONDeserializer[events.ImportCompleted],
		"ImportFailed":           repository.JSONDeserializer[events.ImportFailed],
		"ImportSessionCancelled": repository.JSONDeserializer[events.ImportSessionCancelled],
	},
)
