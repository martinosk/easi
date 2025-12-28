package repositories

import (
	"encoding/json"
	"errors"

	"easi/backend/internal/architectureviews/domain/aggregates"
	"easi/backend/internal/architectureviews/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrViewNotFound = errors.New("view not found")

type ArchitectureViewRepository struct {
	*repository.EventSourcedRepository[*aggregates.ArchitectureView]
}

func NewArchitectureViewRepository(eventStore eventstore.EventStore) *ArchitectureViewRepository {
	return &ArchitectureViewRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			eventDeserializers,
			aggregates.LoadArchitectureViewFromHistory,
			ErrViewNotFound,
		),
	}
}

var eventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
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
	},
)

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
