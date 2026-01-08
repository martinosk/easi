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
		"InvitationCreated": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewInvitationCreated(
				repository.GetString(data, "id"),
				repository.GetString(data, "email"),
				repository.GetString(data, "role"),
				repository.GetString(data, "inviterID"),
				repository.GetString(data, "inviterEmail"),
				repository.GetTime(data, "expiresAt"),
			)
			evt.CreatedAt = repository.GetTime(data, "createdAt")
			return evt
		},
		"InvitationAccepted": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewInvitationAccepted(
				repository.GetString(data, "id"),
				repository.GetString(data, "email"),
			)
			evt.AcceptedAt = repository.GetTime(data, "acceptedAt")
			return evt
		},
		"InvitationRevoked": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewInvitationRevoked(
				repository.GetString(data, "id"),
			)
			evt.RevokedAt = repository.GetTime(data, "revokedAt")
			return evt
		},
		"InvitationExpired": func(data map[string]interface{}) domain.DomainEvent {
			evt := events.NewInvitationExpired(
				repository.GetString(data, "id"),
			)
			evt.ExpiredAt = repository.GetTime(data, "expiredAt")
			return evt
		},
	},
)
