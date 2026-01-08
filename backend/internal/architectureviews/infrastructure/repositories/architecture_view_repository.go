package repositories

import (
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
		"ViewCreated":                deserializeViewCreated,
		"ComponentAddedToView":       deserializeComponentAddedToView,
		"ComponentRemovedFromView":   deserializeComponentRemoved,
		"ViewRenamed":                deserializeViewRenamed,
		"ViewDeleted":                deserializeViewDeleted,
		"DefaultViewChanged":         deserializeDefaultViewChanged,
		"ViewVisibilityChanged":      deserializeViewVisibilityChanged,
	},
)

func deserializeViewCreated(data map[string]interface{}) (domain.DomainEvent, error) {
	id, err := repository.GetRequiredString(data, "id")
	if err != nil {
		return nil, err
	}
	name, err := repository.GetRequiredString(data, "name")
	if err != nil {
		return nil, err
	}
	description, err := repository.GetRequiredString(data, "description")
	if err != nil {
		return nil, err
	}
	createdAt, err := repository.GetRequiredTime(data, "createdAt")
	if err != nil {
		return nil, err
	}
	isPrivate, err := repository.GetOptionalBool(data, "isPrivate", false)
	if err != nil {
		return nil, err
	}
	ownerUserID, err := repository.GetOptionalString(data, "ownerUserId", "")
	if err != nil {
		return nil, err
	}
	ownerEmail, err := repository.GetOptionalString(data, "ownerEmail", "")
	if err != nil {
		return nil, err
	}

	evt := events.ViewCreated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		IsPrivate:   isPrivate,
		OwnerUserID: ownerUserID,
		OwnerEmail:  ownerEmail,
	}
	evt.CreatedAt = createdAt

	return evt, nil
}

func deserializeComponentAddedToView(data map[string]interface{}) (domain.DomainEvent, error) {
	viewID, err := repository.GetRequiredString(data, "viewId")
	if err != nil {
		return nil, err
	}
	componentID, err := repository.GetRequiredString(data, "componentId")
	if err != nil {
		return nil, err
	}
	x, err := repository.GetRequiredFloat64(data, "x")
	if err != nil {
		return nil, err
	}
	y, err := repository.GetRequiredFloat64(data, "y")
	if err != nil {
		return nil, err
	}

	return events.ComponentAddedToView{
		BaseEvent:   domain.NewBaseEvent(viewID),
		ViewID:      viewID,
		ComponentID: componentID,
		X:           x,
		Y:           y,
	}, nil
}

func deserializeComponentRemoved(data map[string]interface{}) (domain.DomainEvent, error) {
	viewID, err := repository.GetRequiredString(data, "viewId")
	if err != nil {
		return nil, err
	}
	componentID, err := repository.GetRequiredString(data, "componentId")
	if err != nil {
		return nil, err
	}

	return events.ComponentRemovedFromView{
		BaseEvent:   domain.NewBaseEvent(viewID),
		ViewID:      viewID,
		ComponentID: componentID,
	}, nil
}

func deserializeViewRenamed(data map[string]interface{}) (domain.DomainEvent, error) {
	viewID, err := repository.GetRequiredString(data, "viewId")
	if err != nil {
		return nil, err
	}
	oldName, err := repository.GetRequiredString(data, "oldName")
	if err != nil {
		return nil, err
	}
	newName, err := repository.GetRequiredString(data, "newName")
	if err != nil {
		return nil, err
	}

	return events.ViewRenamed{
		BaseEvent: domain.NewBaseEvent(viewID),
		ViewID:    viewID,
		OldName:   oldName,
		NewName:   newName,
	}, nil
}

func deserializeViewDeleted(data map[string]interface{}) (domain.DomainEvent, error) {
	viewID, err := repository.GetRequiredString(data, "viewId")
	if err != nil {
		return nil, err
	}

	return events.ViewDeleted{
		BaseEvent: domain.NewBaseEvent(viewID),
		ViewID:    viewID,
	}, nil
}

func deserializeDefaultViewChanged(data map[string]interface{}) (domain.DomainEvent, error) {
	viewID, err := repository.GetRequiredString(data, "viewId")
	if err != nil {
		return nil, err
	}
	isDefault, err := repository.GetRequiredBool(data, "isDefault")
	if err != nil {
		return nil, err
	}

	return events.DefaultViewChanged{
		BaseEvent: domain.NewBaseEvent(viewID),
		ViewID:    viewID,
		IsDefault: isDefault,
	}, nil
}

func deserializeViewVisibilityChanged(data map[string]interface{}) (domain.DomainEvent, error) {
	viewID, err := repository.GetRequiredString(data, "viewId")
	if err != nil {
		return nil, err
	}
	isPrivate, err := repository.GetRequiredBool(data, "isPrivate")
	if err != nil {
		return nil, err
	}
	ownerUserID, err := repository.GetOptionalString(data, "ownerUserId", "")
	if err != nil {
		return nil, err
	}
	ownerEmail, err := repository.GetOptionalString(data, "ownerEmail", "")
	if err != nil {
		return nil, err
	}

	return events.ViewVisibilityChanged{
		BaseEvent:   domain.NewBaseEvent(viewID),
		ViewID:      viewID,
		IsPrivate:   isPrivate,
		OwnerUserID: ownerUserID,
		OwnerEmail:  ownerEmail,
	}, nil
}
