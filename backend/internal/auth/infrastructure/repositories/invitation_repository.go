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
		"InvitationCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			email, err := repository.GetRequiredString(data, "email")
			if err != nil {
				return nil, err
			}
			role, err := repository.GetRequiredString(data, "role")
			if err != nil {
				return nil, err
			}
			inviterID, err := repository.GetRequiredString(data, "inviterID")
			if err != nil {
				return nil, err
			}
			inviterEmail, err := repository.GetRequiredString(data, "inviterEmail")
			if err != nil {
				return nil, err
			}
			expiresAt, err := repository.GetRequiredTime(data, "expiresAt")
			if err != nil {
				return nil, err
			}
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewInvitationCreated(id, email, role, inviterID, inviterEmail, expiresAt)
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"InvitationAccepted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			email, err := repository.GetRequiredString(data, "email")
			if err != nil {
				return nil, err
			}
			acceptedAt, err := repository.GetRequiredTime(data, "acceptedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewInvitationAccepted(id, email)
			evt.AcceptedAt = acceptedAt
			return evt, nil
		},
		"InvitationRevoked": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			revokedAt, err := repository.GetRequiredTime(data, "revokedAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewInvitationRevoked(id)
			evt.RevokedAt = revokedAt
			return evt, nil
		},
		"InvitationExpired": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			expiredAt, err := repository.GetRequiredTime(data, "expiredAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewInvitationExpired(id)
			evt.ExpiredAt = expiredAt
			return evt, nil
		},
	},
)
