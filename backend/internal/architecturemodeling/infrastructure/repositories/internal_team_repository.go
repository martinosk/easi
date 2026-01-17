package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrInternalTeamNotFound = errors.New("internal team not found")

type InternalTeamRepository struct {
	*repository.EventSourcedRepository[*aggregates.InternalTeam]
}

func NewInternalTeamRepository(eventStore eventstore.EventStore) *InternalTeamRepository {
	return &InternalTeamRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			internalTeamEventDeserializers,
			aggregates.LoadInternalTeamFromHistory,
			ErrInternalTeamNotFound,
		),
	}
}

var internalTeamEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"InternalTeamCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			department, _ := repository.GetOptionalString(data, "department", "")
			contactPerson, _ := repository.GetOptionalString(data, "contactPerson", "")
			notes, _ := repository.GetOptionalString(data, "notes", "")
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewInternalTeamCreated(id, name, department, contactPerson, notes)
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"InternalTeamUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			department, _ := repository.GetOptionalString(data, "department", "")
			contactPerson, _ := repository.GetOptionalString(data, "contactPerson", "")
			notes, _ := repository.GetOptionalString(data, "notes", "")

			return events.NewInternalTeamUpdated(id, name, department, contactPerson, notes), nil
		},
		"InternalTeamDeleted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}

			return events.NewInternalTeamDeleted(id, name), nil
		},
	},
)
