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
		"ImportSessionCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			sourceFormat, err := repository.GetRequiredString(data, "sourceFormat")
			if err != nil {
				return nil, err
			}
			businessDomainId, err := repository.GetRequiredString(data, "businessDomainId")
			if err != nil {
				return nil, err
			}
			capabilityEAOwner, err := repository.GetOptionalString(data, "capabilityEAOwner", "")
			if err != nil {
				return nil, err
			}
			preview, err := repository.GetOptionalMap(data, "preview")
			if err != nil {
				return nil, err
			}
			parsedData, err := repository.GetOptionalMap(data, "parsedData")
			if err != nil {
				return nil, err
			}
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewImportSessionCreated(id, sourceFormat, businessDomainId, capabilityEAOwner, preview, parsedData)
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"ImportStarted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			totalItems, err := repository.GetRequiredInt(data, "totalItems")
			if err != nil {
				return nil, err
			}
			startedAt, err := repository.GetRequiredTime(data, "startedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewImportStarted(id, totalItems)
			evt.StartedAt = startedAt
			return evt, nil
		},
		"ImportProgressUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			phase, err := repository.GetRequiredString(data, "phase")
			if err != nil {
				return nil, err
			}
			totalItems, err := repository.GetRequiredInt(data, "totalItems")
			if err != nil {
				return nil, err
			}
			completedItems, err := repository.GetRequiredInt(data, "completedItems")
			if err != nil {
				return nil, err
			}
			updatedAt, err := repository.GetRequiredTime(data, "updatedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewImportProgressUpdated(id, phase, totalItems, completedItems)
			evt.UpdatedAt = updatedAt
			return evt, nil
		},
		"ImportCompleted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			capabilitiesCreated, err := repository.GetRequiredInt(data, "capabilitiesCreated")
			if err != nil {
				return nil, err
			}
			componentsCreated, err := repository.GetRequiredInt(data, "componentsCreated")
			if err != nil {
				return nil, err
			}
			realizationsCreated, err := repository.GetRequiredInt(data, "realizationsCreated")
			if err != nil {
				return nil, err
			}
			domainAssignments, err := repository.GetRequiredInt(data, "domainAssignments")
			if err != nil {
				return nil, err
			}
			errors, err := repository.GetOptionalMapSlice(data, "errors")
			if err != nil {
				return nil, err
			}
			completedAt, err := repository.GetRequiredTime(data, "completedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewImportCompleted(id, capabilitiesCreated, componentsCreated, realizationsCreated, domainAssignments, errors)
			evt.CompletedAt = completedAt
			return evt, nil
		},
		"ImportFailed": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			reason, err := repository.GetRequiredString(data, "reason")
			if err != nil {
				return nil, err
			}
			failedAt, err := repository.GetRequiredTime(data, "failedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewImportFailed(id, reason)
			evt.FailedAt = failedAt
			return evt, nil
		},
		"ImportSessionCancelled": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			cancelledAt, err := repository.GetRequiredTime(data, "cancelledAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewImportSessionCancelled(id)
			evt.CancelledAt = cancelledAt
			return evt, nil
		},
	},
)
