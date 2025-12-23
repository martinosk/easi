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
	ErrUserAggregateNotFound = errors.New("user aggregate not found")
)

type UserAggregateRepository struct {
	eventStore eventstore.EventStore
}

func NewUserAggregateRepository(eventStore eventstore.EventStore) *UserAggregateRepository {
	return &UserAggregateRepository{
		eventStore: eventStore,
	}
}

func (r *UserAggregateRepository) Save(ctx context.Context, user *aggregates.User) error {
	uncommittedEvents := user.GetUncommittedChanges()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	err := r.eventStore.SaveEvents(ctx, user.ID(), uncommittedEvents, user.Version()-len(uncommittedEvents))
	if err != nil {
		return err
	}

	user.MarkChangesAsCommitted()
	return nil
}

func (r *UserAggregateRepository) GetByID(ctx context.Context, id string) (*aggregates.User, error) {
	storedEvents, err := r.eventStore.GetEvents(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(storedEvents) == 0 {
		return nil, ErrUserAggregateNotFound
	}

	domainEvents := r.deserializeEvents(storedEvents)

	return aggregates.LoadUserFromHistory(domainEvents)
}

func (r *UserAggregateRepository) deserializeEvents(storedEvents []domain.DomainEvent) []domain.DomainEvent {
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		eventData := event.EventData()

		switch event.EventType() {
		case "UserCreated":
			id, _ := eventData["id"].(string)
			email, _ := eventData["email"].(string)
			name, _ := eventData["name"].(string)
			role, _ := eventData["role"].(string)
			externalID, _ := eventData["externalId"].(string)
			invitationID, _ := eventData["invitationId"].(string)
			createdAtStr, _ := eventData["createdAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339, createdAtStr)

			concreteEvent := events.NewUserCreated(id, email, name, role, externalID, invitationID)
			concreteEvent.CreatedAt = createdAt
			domainEvents = append(domainEvents, concreteEvent)

		case "UserRoleChanged":
			id, _ := eventData["id"].(string)
			oldRole, _ := eventData["oldRole"].(string)
			newRole, _ := eventData["newRole"].(string)
			changedByID, _ := eventData["changedById"].(string)
			changedAtStr, _ := eventData["changedAt"].(string)
			changedAt, _ := time.Parse(time.RFC3339, changedAtStr)

			concreteEvent := events.NewUserRoleChanged(id, oldRole, newRole, changedByID)
			concreteEvent.ChangedAt = changedAt
			domainEvents = append(domainEvents, concreteEvent)

		case "UserDisabled":
			id, _ := eventData["id"].(string)
			disabledBy, _ := eventData["disabledBy"].(string)
			disabledAtStr, _ := eventData["disabledAt"].(string)
			disabledAt, _ := time.Parse(time.RFC3339, disabledAtStr)

			concreteEvent := events.NewUserDisabled(id, disabledBy)
			concreteEvent.DisabledAt = disabledAt
			domainEvents = append(domainEvents, concreteEvent)

		case "UserEnabled":
			id, _ := eventData["id"].(string)
			enabledBy, _ := eventData["enabledBy"].(string)
			enabledAtStr, _ := eventData["enabledAt"].(string)
			enabledAt, _ := time.Parse(time.RFC3339, enabledAtStr)

			concreteEvent := events.NewUserEnabled(id, enabledBy)
			concreteEvent.EnabledAt = enabledAt
			domainEvents = append(domainEvents, concreteEvent)

		default:
			continue
		}
	}

	return domainEvents
}
