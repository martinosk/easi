package repositories

import (
	"errors"

	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
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
		"UserCreated":     repository.JSONDeserializer[events.UserCreated],
		"UserRoleChanged": repository.JSONDeserializer[events.UserRoleChanged],
		"UserDisabled":    repository.JSONDeserializer[events.UserDisabled],
		"UserEnabled":     repository.JSONDeserializer[events.UserEnabled],
	},
)
