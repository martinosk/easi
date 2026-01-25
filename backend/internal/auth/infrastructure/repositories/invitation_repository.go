package repositories

import (
	"errors"

	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var (
	ErrInvitationNotFound = errors.New("invitation not found")
)

type InvitationRepository struct {
	*repository.EventSourcedRepository[*aggregates.Invitation]
}

func NewInvitationRepository(eventStore eventstore.EventStore) *InvitationRepository {
	return &InvitationRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			invitationEventDeserializers,
			aggregates.LoadInvitationFromHistory,
			ErrInvitationNotFound,
		),
	}
}

var invitationEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"InvitationCreated":  repository.JSONDeserializer[events.InvitationCreated],
		"InvitationAccepted": repository.JSONDeserializer[events.InvitationAccepted],
		"InvitationRevoked":  repository.JSONDeserializer[events.InvitationRevoked],
		"InvitationExpired":  repository.JSONDeserializer[events.InvitationExpired],
	},
)
