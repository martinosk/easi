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
		"UserCreated": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewUserCreated(
				repository.GetString(data, "id"),
				repository.GetString(data, "email"),
				repository.GetString(data, "name"),
				repository.GetString(data, "role"),
				repository.GetString(data, "externalId"),
				repository.GetString(data, "invitationId"),
			)
			evt.CreatedAt = repository.GetTimeRFC3339(data, "createdAt")
			return evt
		},
		"UserRoleChanged": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewUserRoleChanged(
				repository.GetString(data, "id"),
				repository.GetString(data, "oldRole"),
				repository.GetString(data, "newRole"),
				repository.GetString(data, "changedById"),
			)
			evt.ChangedAt = repository.GetTimeRFC3339(data, "changedAt")
			return evt
		},
		"UserDisabled": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewUserDisabled(
				repository.GetString(data, "id"),
				repository.GetString(data, "disabledBy"),
			)
			evt.DisabledAt = repository.GetTimeRFC3339(data, "disabledAt")
			return evt
		},
		"UserEnabled": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewUserEnabled(
				repository.GetString(data, "id"),
				repository.GetString(data, "enabledBy"),
			)
			evt.EnabledAt = repository.GetTimeRFC3339(data, "enabledAt")
			return evt
		},
	},
)
