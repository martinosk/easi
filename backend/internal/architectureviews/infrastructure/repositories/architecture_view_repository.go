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
	domainEvents := r.deserializeEvents(storedEvents)

	return aggregates.LoadArchitectureViewFromHistory(domainEvents)
}

type eventDeserializer func(map[string]interface{}) domain.DomainEvent

var eventDeserializers = map[string]eventDeserializer{
	"ViewCreated": deserializeViewCreated,
	"ComponentAddedToView": func(data map[string]interface{}) domain.DomainEvent {
		viewID, componentID, x, y := extractComponentPosition(data)
		return events.ComponentAddedToView{
			BaseEvent:   domain.NewBaseEvent(viewID),
			ViewID:      viewID,
			ComponentID: componentID,
			X:           x,
			Y:           y,
		}
	},
	"ComponentPositionUpdated": func(data map[string]interface{}) domain.DomainEvent {
		viewID, componentID, x, y := extractComponentPosition(data)
		return events.ComponentPositionUpdated{
			BaseEvent:   domain.NewBaseEvent(viewID),
			ViewID:      viewID,
			ComponentID: componentID,
			X:           x,
			Y:           y,
		}
	},
	"ComponentRemovedFromView": deserializeComponentRemoved,
	"ViewRenamed":              deserializeViewRenamed,
	"ViewDeleted":              deserializeViewDeleted,
	"DefaultViewChanged":       deserializeDefaultViewChanged,
}

func (r *ArchitectureViewRepository) deserializeEvents(storedEvents []domain.DomainEvent) []domain.DomainEvent {
	domainEvents := make([]domain.DomainEvent, 0, len(storedEvents))

	for _, event := range storedEvents {
		deserializer, exists := eventDeserializers[event.EventType()]
		if !exists {
			continue
		}

		domainEvents = append(domainEvents, deserializer(event.EventData()))
	}

	return domainEvents
}

func deserializeViewCreated(data map[string]interface{}) domain.DomainEvent {
	id, _ := data["id"].(string)
	name, _ := data["name"].(string)
	description, _ := data["description"].(string)

	evt := events.ViewCreated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
	}

	if createdAtStr, ok := data["createdAt"].(string); ok {
		if createdAt, err := json.Marshal(createdAtStr); err == nil {
			_ = json.Unmarshal(createdAt, &evt.CreatedAt)
		}
	}

	return evt
}

func extractComponentPosition(data map[string]interface{}) (viewID, componentID string, x, y float64) {
	viewID, _ = data["viewId"].(string)
	componentID, _ = data["componentId"].(string)
	x, _ = data["x"].(float64)
	y, _ = data["y"].(float64)
	return
}

func deserializeComponentRemoved(data map[string]interface{}) domain.DomainEvent {
	viewID, _ := data["viewId"].(string)
	componentID, _ := data["componentId"].(string)

	return events.ComponentRemovedFromView{
		BaseEvent:   domain.NewBaseEvent(viewID),
		ViewID:      viewID,
		ComponentID: componentID,
	}
}

func deserializeViewRenamed(data map[string]interface{}) domain.DomainEvent {
	viewID, _ := data["viewId"].(string)
	oldName, _ := data["oldName"].(string)
	newName, _ := data["newName"].(string)

	return events.ViewRenamed{
		BaseEvent: domain.NewBaseEvent(viewID),
		ViewID:    viewID,
		OldName:   oldName,
		NewName:   newName,
	}
}

func deserializeViewDeleted(data map[string]interface{}) domain.DomainEvent {
	viewID, _ := data["viewId"].(string)

	return events.ViewDeleted{
		BaseEvent: domain.NewBaseEvent(viewID),
		ViewID:    viewID,
	}
}

func deserializeDefaultViewChanged(data map[string]interface{}) domain.DomainEvent {
	viewID, _ := data["viewId"].(string)
	isDefault, _ := data["isDefault"].(bool)

	return events.DefaultViewChanged{
		BaseEvent: domain.NewBaseEvent(viewID),
		ViewID:    viewID,
		IsDefault: isDefault,
	}
}
