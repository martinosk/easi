package repositories

import (
	"errors"

	"easi/backend/internal/importing/domain/aggregates"
	"easi/backend/internal/importing/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
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
		"ImportSessionCreated": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewImportSessionCreated(
				repository.GetString(data, "id"),
				repository.GetString(data, "sourceFormat"),
				repository.GetString(data, "businessDomainId"),
				repository.GetMap(data, "preview"),
				repository.GetMap(data, "parsedData"),
			)
			evt.CreatedAt = repository.GetTime(data, "createdAt")
			return evt
		},
		"ImportStarted": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewImportStarted(
				repository.GetString(data, "id"),
				repository.GetInt(data, "totalItems"),
			)
			evt.StartedAt = repository.GetTime(data, "startedAt")
			return evt
		},
		"ImportProgressUpdated": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewImportProgressUpdated(
				repository.GetString(data, "id"),
				repository.GetString(data, "phase"),
				repository.GetInt(data, "totalItems"),
				repository.GetInt(data, "completedItems"),
			)
			evt.UpdatedAt = repository.GetTime(data, "updatedAt")
			return evt
		},
		"ImportCompleted": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewImportCompleted(
				repository.GetString(data, "id"),
				repository.GetInt(data, "capabilitiesCreated"),
				repository.GetInt(data, "componentsCreated"),
				repository.GetInt(data, "realizationsCreated"),
				repository.GetInt(data, "domainAssignments"),
				repository.GetMapSlice(data, "errors"),
			)
			evt.CompletedAt = repository.GetTime(data, "completedAt")
			return evt
		},
		"ImportFailed": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewImportFailed(
				repository.GetString(data, "id"),
				repository.GetString(data, "reason"),
			)
			evt.FailedAt = repository.GetTime(data, "failedAt")
			return evt
		},
		"ImportSessionCancelled": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewImportSessionCancelled(
				repository.GetString(data, "id"),
			)
			evt.CancelledAt = repository.GetTime(data, "cancelledAt")
			return evt
		},
	},
)
