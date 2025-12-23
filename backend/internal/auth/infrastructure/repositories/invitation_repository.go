package repositories

import (
	"context"
	"errors"
	"time"

	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvitationNotFound = errors.New("invitation not found")
)

type InvitationRepository struct {
	eventStore eventstore.EventStore
}

func NewInvitationRepository(eventStore eventstore.EventStore) *InvitationRepository {
	return &InvitationRepository{
		eventStore: eventStore,
	}
}

func (r *InvitationRepository) Save(ctx context.Context, invitation *aggregates.Invitation) error {
	uncommittedEvents := invitation.GetUncommittedChanges()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	err := r.eventStore.SaveEvents(ctx, invitation.ID(), uncommittedEvents, invitation.Version()-len(uncommittedEvents))
	if err != nil {
		return err
	}

	invitation.MarkChangesAsCommitted()
	return nil
}

func (r *InvitationRepository) GetByID(ctx context.Context, id string) (*aggregates.Invitation, error) {
	storedEvents, err := r.eventStore.GetEvents(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(storedEvents) == 0 {
		return nil, ErrInvitationNotFound
	}

	domainEvents := r.deserializeEvents(storedEvents)

	return aggregates.LoadInvitationFromHistory(domainEvents)
}

func (r *InvitationRepository) deserializeEvents(storedEvents []domain.DomainEvent) []domain.DomainEvent {
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		eventData := event.EventData()

		switch event.EventType() {
		case "InvitationCreated":
			id, _ := eventData["id"].(string)
			email, _ := eventData["email"].(string)
			role, _ := eventData["role"].(string)
			inviterID, _ := eventData["inviterID"].(string)
			inviterEmail, _ := eventData["inviterEmail"].(string)
			createdAtStr, _ := eventData["createdAt"].(string)
			expiresAtStr, _ := eventData["expiresAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339, createdAtStr)
			expiresAt, _ := time.Parse(time.RFC3339, expiresAtStr)

			concreteEvent := events.NewInvitationCreated(id, email, role, inviterID, inviterEmail, expiresAt)
			concreteEvent.CreatedAt = createdAt
			domainEvents = append(domainEvents, concreteEvent)

		case "InvitationAccepted":
			id, _ := eventData["id"].(string)
			email, _ := eventData["email"].(string)
			acceptedAtStr, _ := eventData["acceptedAt"].(string)
			acceptedAt, _ := time.Parse(time.RFC3339, acceptedAtStr)

			concreteEvent := events.NewInvitationAccepted(id, email)
			concreteEvent.AcceptedAt = acceptedAt
			domainEvents = append(domainEvents, concreteEvent)

		case "InvitationRevoked":
			id, _ := eventData["id"].(string)
			revokedAtStr, _ := eventData["revokedAt"].(string)
			revokedAt, _ := time.Parse(time.RFC3339, revokedAtStr)

			concreteEvent := events.NewInvitationRevoked(id)
			concreteEvent.RevokedAt = revokedAt
			domainEvents = append(domainEvents, concreteEvent)

		case "InvitationExpired":
			id, _ := eventData["id"].(string)
			expiredAtStr, _ := eventData["expiredAt"].(string)
			expiredAt, _ := time.Parse(time.RFC3339, expiredAtStr)

			concreteEvent := events.NewInvitationExpired(id)
			concreteEvent.ExpiredAt = expiredAt
			domainEvents = append(domainEvents, concreteEvent)

		default:
			continue
		}
	}

	return domainEvents
}
