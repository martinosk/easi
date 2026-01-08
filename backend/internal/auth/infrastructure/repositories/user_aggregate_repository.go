package repositories

import (
	"errors"

	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var (
	ErrUserAggregateNotFound = errors.New("user aggregate not found")
)

type UserAggregateRepository struct {
	*repository.EventSourcedRepository[*aggregates.User]
}

func NewUserAggregateRepository(eventStore eventstore.EventStore) *UserAggregateRepository {
	return &UserAggregateRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			userAggregateEventDeserializers,
			aggregates.LoadUserFromHistory,
			ErrUserAggregateNotFound,
		),
	}
}

var userAggregateEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"UserCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			email, err := repository.GetRequiredString(data, "email")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			role, err := repository.GetRequiredString(data, "role")
			if err != nil {
				return nil, err
			}
			externalId, err := repository.GetRequiredString(data, "externalId")
			if err != nil {
				return nil, err
			}
			invitationId, err := repository.GetRequiredString(data, "invitationId")
			if err != nil {
				return nil, err
			}
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewUserCreated(id, email, name, role, externalId, invitationId)
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"UserRoleChanged": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			oldRole, err := repository.GetRequiredString(data, "oldRole")
			if err != nil {
				return nil, err
			}
			newRole, err := repository.GetRequiredString(data, "newRole")
			if err != nil {
				return nil, err
			}
			changedById, err := repository.GetRequiredString(data, "changedById")
			if err != nil {
				return nil, err
			}
			changedAt, err := repository.GetRequiredTime(data, "changedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewUserRoleChanged(id, oldRole, newRole, changedById)
			evt.ChangedAt = changedAt
			return evt, nil
		},
		"UserDisabled": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			disabledBy, err := repository.GetRequiredString(data, "disabledBy")
			if err != nil {
				return nil, err
			}
			disabledAt, err := repository.GetRequiredTime(data, "disabledAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewUserDisabled(id, disabledBy)
			evt.DisabledAt = disabledAt
			return evt, nil
		},
		"UserEnabled": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			enabledBy, err := repository.GetRequiredString(data, "enabledBy")
			if err != nil {
				return nil, err
			}
			enabledAt, err := repository.GetRequiredTime(data, "enabledAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewUserEnabled(id, enabledBy)
			evt.EnabledAt = enabledAt
			return evt, nil
		},
	},
)
