package repositories

import (
	"context"
	"encoding/json"
	"errors"

	"easi/backend/internal/architectureviews/domain/aggregates"
	"easi/backend/internal/architectureviews/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/domain"
)

var (
	// ErrViewNotFound is returned when a view is not found
	ErrViewNotFound = errors.New("view not found")
)

// ArchitectureViewRepository manages persistence of architecture views
type ArchitectureViewRepository struct {
	eventStore eventstore.EventStore
}

// NewArchitectureViewRepository creates a new repository
func NewArchitectureViewRepository(eventStore eventstore.EventStore) *ArchitectureViewRepository {
	return &ArchitectureViewRepository{
		eventStore: eventStore,
	}
}

// Save persists an architecture view aggregate
func (r *ArchitectureViewRepository) Save(ctx context.Context, view *aggregates.ArchitectureView) error {
	uncommittedEvents := view.GetUncommittedChanges()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	// Calculate the expected version: current version minus the number of uncommitted events
	// This represents the version before the new events were applied
	expectedVersion := view.Version() - len(uncommittedEvents)

	err := r.eventStore.SaveEvents(ctx, view.ID(), uncommittedEvents, expectedVersion)
	if err != nil {
		return err
	}

	view.MarkChangesAsCommitted()
	return nil
}

// GetByID retrieves an architecture view by ID
func (r *ArchitectureViewRepository) GetByID(ctx context.Context, id string) (*aggregates.ArchitectureView, error) {
	storedEvents, err := r.eventStore.GetEvents(ctx, id)
	if err != nil {
		return nil, err
	}

	// If no events found, the aggregate doesn't exist
	if len(storedEvents) == 0 {
		return nil, ErrViewNotFound
	}

	// Deserialize events (simplified)
	domainEvents, err := r.deserializeEvents(storedEvents)
	if err != nil {
		return nil, err
	}

	return aggregates.LoadArchitectureViewFromHistory(domainEvents)
}

// deserializeEvents converts stored events to domain events
func (r *ArchitectureViewRepository) deserializeEvents(storedEvents []domain.DomainEvent) ([]domain.DomainEvent, error) {
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		// Get the event data as a map
		eventData := event.EventData()

		switch event.EventType() {
		case "ViewCreated":
			// Manually reconstruct the event from the map
			id, _ := eventData["id"].(string)
			name, _ := eventData["name"].(string)
			description, _ := eventData["description"].(string)

			specificEvent := events.ViewCreated{
				BaseEvent:   domain.NewBaseEvent(id),
				ID:          id,
				Name:        name,
				Description: description,
			}
			// Parse CreatedAt if present
			if createdAtStr, ok := eventData["createdAt"].(string); ok {
				if createdAt, err := json.Marshal(createdAtStr); err == nil {
					json.Unmarshal(createdAt, &specificEvent.CreatedAt)
				}
			}
			domainEvents = append(domainEvents, specificEvent)

		case "ComponentAddedToView":
			// Manually reconstruct the event from the map
			viewID, _ := eventData["viewId"].(string)
			componentID, _ := eventData["componentId"].(string)
			x, _ := eventData["x"].(float64)
			y, _ := eventData["y"].(float64)

			specificEvent := events.ComponentAddedToView{
				BaseEvent:   domain.NewBaseEvent(viewID),
				ViewID:      viewID,
				ComponentID: componentID,
				X:           x,
				Y:           y,
			}
			domainEvents = append(domainEvents, specificEvent)

		case "ComponentPositionUpdated":
			// Manually reconstruct the event from the map
			viewID, _ := eventData["viewId"].(string)
			componentID, _ := eventData["componentId"].(string)
			x, _ := eventData["x"].(float64)
			y, _ := eventData["y"].(float64)

			specificEvent := events.ComponentPositionUpdated{
				BaseEvent:   domain.NewBaseEvent(viewID),
				ViewID:      viewID,
				ComponentID: componentID,
				X:           x,
				Y:           y,
			}
			domainEvents = append(domainEvents, specificEvent)

		case "ComponentRemovedFromView":
			// Manually reconstruct the event from the map
			viewID, _ := eventData["viewId"].(string)
			componentID, _ := eventData["componentId"].(string)

			specificEvent := events.ComponentRemovedFromView{
				BaseEvent:   domain.NewBaseEvent(viewID),
				ViewID:      viewID,
				ComponentID: componentID,
			}
			domainEvents = append(domainEvents, specificEvent)

		case "ViewRenamed":
			// Manually reconstruct the event from the map
			viewID, _ := eventData["viewId"].(string)
			oldName, _ := eventData["oldName"].(string)
			newName, _ := eventData["newName"].(string)

			specificEvent := events.ViewRenamed{
				BaseEvent: domain.NewBaseEvent(viewID),
				ViewID:    viewID,
				OldName:   oldName,
				NewName:   newName,
			}
			domainEvents = append(domainEvents, specificEvent)

		case "ViewDeleted":
			// Manually reconstruct the event from the map
			viewID, _ := eventData["viewId"].(string)

			specificEvent := events.ViewDeleted{
				BaseEvent: domain.NewBaseEvent(viewID),
				ViewID:    viewID,
			}
			domainEvents = append(domainEvents, specificEvent)

		case "DefaultViewChanged":
			// Manually reconstruct the event from the map
			viewID, _ := eventData["viewId"].(string)
			isDefault, _ := eventData["isDefault"].(bool)

			specificEvent := events.DefaultViewChanged{
				BaseEvent: domain.NewBaseEvent(viewID),
				ViewID:    viewID,
				IsDefault: isDefault,
			}
			domainEvents = append(domainEvents, specificEvent)

		default:
			// Unknown event type, skip it
			continue
		}
	}

	return domainEvents, nil
}
