package repositories

import (
	"errors"
	"time"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrAcquiredEntityNotFound = errors.New("acquired entity not found")

type AcquiredEntityRepository struct {
	*repository.EventSourcedRepository[*aggregates.AcquiredEntity]
}

func NewAcquiredEntityRepository(eventStore eventstore.EventStore) *AcquiredEntityRepository {
	return &AcquiredEntityRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			acquiredEntityEventDeserializers,
			aggregates.LoadAcquiredEntityFromHistory,
			ErrAcquiredEntityNotFound,
		),
	}
}

var acquiredEntityEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"AcquiredEntityCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			integrationStatus, _ := repository.GetOptionalString(data, "integrationStatus", "")
			notes, _ := repository.GetOptionalString(data, "notes", "")
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}
			acquisitionDate, _ := repository.GetOptionalTime(data, "acquisitionDate", time.Time{})

			var acqDatePtr *time.Time
			if !acquisitionDate.IsZero() {
				acqDatePtr = &acquisitionDate
			}

			evt := events.NewAcquiredEntityCreated(id, name, acqDatePtr, integrationStatus, notes)
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"AcquiredEntityUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			integrationStatus, _ := repository.GetOptionalString(data, "integrationStatus", "")
			notes, _ := repository.GetOptionalString(data, "notes", "")
			acquisitionDate, _ := repository.GetOptionalTime(data, "acquisitionDate", time.Time{})

			var acqDatePtr *time.Time
			if !acquisitionDate.IsZero() {
				acqDatePtr = &acquisitionDate
			}

			return events.NewAcquiredEntityUpdated(id, name, acqDatePtr, integrationStatus, notes), nil
		},
		"AcquiredEntityDeleted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}

			return events.NewAcquiredEntityDeleted(id, name), nil
		},
	},
)
