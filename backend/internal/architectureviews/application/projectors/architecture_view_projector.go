package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/domain/events"
	viewsPL "easi/backend/internal/architectureviews/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

// ArchitectureViewProjector projects events to read models
type ArchitectureViewProjector struct {
	readModel *readmodels.ArchitectureViewReadModel
}

// NewArchitectureViewProjector creates a new projector
func NewArchitectureViewProjector(readModel *readmodels.ArchitectureViewReadModel) *ArchitectureViewProjector {
	return &ArchitectureViewProjector{
		readModel: readModel,
	}
}

// Handle implements the EventHandler interface for the event bus
func (p *ArchitectureViewProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

// ProjectEvent projects a domain event to the read model
func (p *ArchitectureViewProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case viewsPL.ViewCreated:
		return p.projectViewCreated(ctx, eventData)
	case viewsPL.ComponentAddedToView:
		return p.projectComponentAdded(ctx, eventData)
	case viewsPL.ComponentRemovedFromView:
		return p.projectComponentRemoved(ctx, eventData)
	case viewsPL.ViewRenamed:
		return p.projectViewRenamed(ctx, eventData)
	case viewsPL.ViewDeleted:
		return p.projectViewDeleted(ctx, eventData)
	case viewsPL.DefaultViewChanged:
		return p.projectDefaultViewChanged(ctx, eventData)
	case viewsPL.ViewVisibilityChanged:
		return p.projectViewVisibilityChanged(ctx, eventData)
	}

	return nil
}

func (p *ArchitectureViewProjector) projectViewCreated(ctx context.Context, eventData []byte) error {
	var event events.ViewCreated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ViewCreated event: %v", err)
		return err
	}

	dto := readmodels.ArchitectureViewDTO{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		IsPrivate:   event.IsPrivate,
		OwnerUserID: stringPtrIfNotEmpty(event.OwnerUserID),
		OwnerEmail:  stringPtrIfNotEmpty(event.OwnerEmail),
		CreatedAt:   event.CreatedAt,
		Components:  make([]readmodels.ComponentPositionDTO, 0),
	}

	return p.readModel.InsertView(ctx, dto)
}

func stringPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (p *ArchitectureViewProjector) projectComponentAdded(ctx context.Context, eventData []byte) error {
	var event events.ComponentAddedToView
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ComponentAddedToView event: %v", err)
		return err
	}

	pos := readmodels.Position{X: event.X, Y: event.Y}
	return p.readModel.AddComponent(ctx, readmodels.ViewID(event.ViewID), readmodels.ComponentID(event.ComponentID), pos)
}

func (p *ArchitectureViewProjector) projectComponentRemoved(ctx context.Context, eventData []byte) error {
	var event events.ComponentRemovedFromView
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ComponentRemovedFromView event: %v", err)
		return err
	}

	return p.readModel.RemoveComponent(ctx, event.ViewID, event.ComponentID)
}

func (p *ArchitectureViewProjector) projectViewRenamed(ctx context.Context, eventData []byte) error {
	var event events.ViewRenamed
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ViewRenamed event: %v", err)
		return err
	}

	return p.readModel.UpdateViewName(ctx, event.ViewID, event.NewName)
}

func (p *ArchitectureViewProjector) projectViewDeleted(ctx context.Context, eventData []byte) error {
	var event events.ViewDeleted
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ViewDeleted event: %v", err)
		return err
	}

	return p.readModel.MarkViewAsDeleted(ctx, event.ViewID)
}

func (p *ArchitectureViewProjector) projectDefaultViewChanged(ctx context.Context, eventData []byte) error {
	var event events.DefaultViewChanged
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal DefaultViewChanged event: %v", err)
		return err
	}

	return p.readModel.SetViewAsDefault(ctx, event.ViewID, event.IsDefault)
}

func (p *ArchitectureViewProjector) projectViewVisibilityChanged(ctx context.Context, eventData []byte) error {
	var event events.ViewVisibilityChanged
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ViewVisibilityChanged event: %v", err)
		return err
	}

	return p.readModel.UpdateVisibility(ctx, readmodels.VisibilityUpdate{
		ViewID:      event.ViewID,
		IsPrivate:   event.IsPrivate,
		OwnerUserID: event.OwnerUserID,
		OwnerEmail:  event.OwnerEmail,
	})
}
