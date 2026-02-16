package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type CapabilityProjector struct {
	readModel           *readmodels.CapabilityReadModel
	assignmentReadModel *readmodels.DomainCapabilityAssignmentReadModel
}

func NewCapabilityProjector(readModel *readmodels.CapabilityReadModel, assignmentReadModel *readmodels.DomainCapabilityAssignmentReadModel) *CapabilityProjector {
	return &CapabilityProjector{
		readModel:           readModel,
		assignmentReadModel: assignmentReadModel,
	}
}

func (p *CapabilityProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func projectionHandler[T any](project func(context.Context, T) error) func(context.Context, []byte) error {
	return func(ctx context.Context, eventData []byte) error {
		return handleProjection(ctx, eventData, project)
	}
}

func (p *CapabilityProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"CapabilityCreated":         projectionHandler(p.projectCreated),
		"CapabilityUpdated":         p.handleCapabilityUpdated,
		"CapabilityMetadataUpdated": projectionHandler(p.projectMetadataUpdated),
		"CapabilityExpertAdded":     projectionHandler(p.projectExpertAdded),
		"CapabilityExpertRemoved":   projectionHandler(p.projectExpertRemoved),
		"CapabilityTagAdded":        projectionHandler(p.projectTagAdded),
		"CapabilityParentChanged":   projectionHandler(p.projectParentChanged),
		"CapabilityLevelChanged":    projectionHandler(p.projectLevelChanged),
		"CapabilityDeleted":         projectionHandler(p.projectDeleted),
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func unmarshalEvent[T any](eventData []byte) (T, error) {
	var event T
	if err := json.Unmarshal(eventData, &event); err != nil {
		return event, fmt.Errorf("unmarshal %T: %w", event, err)
	}
	return event, nil
}

func handleProjection[T any](ctx context.Context, eventData []byte, project func(context.Context, T) error) error {
	event, err := unmarshalEvent[T](eventData)
	if err != nil {
		return err
	}
	return project(ctx, event)
}

func (p *CapabilityProjector) projectCreated(ctx context.Context, event events.CapabilityCreated) error {
	return p.readModel.Insert(ctx, readmodels.CapabilityDTO{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		ParentID:    event.ParentID,
		Level:       event.Level,
		CreatedAt:   event.CreatedAt,
	})
}

func (p *CapabilityProjector) handleCapabilityUpdated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.CapabilityUpdated](eventData)
	if err != nil {
		return err
	}

	if err := p.readModel.Update(ctx, readmodels.CapabilityUpdate{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
	}); err != nil {
		return err
	}

	capability, err := p.readModel.GetByID(ctx, event.ID)
	if err != nil {
		log.Printf("Failed to get capability %s after update: %v", event.ID, err)
		return err
	}
	if capability == nil {
		log.Printf("Capability not found after update: %s", event.ID)
		return fmt.Errorf("capability %s not found after update in capability projector", event.ID)
	}

	if err := p.assignmentReadModel.UpdateCapabilityInfo(ctx, readmodels.CapabilityInfoUpdate{
		CapabilityID: event.ID, Name: capability.Name, Description: capability.Description, Level: capability.Level,
	}); err != nil {
		log.Printf("Failed to update assignments for capability %s: %v", event.ID, err)
		return err
	}

	return nil
}

func (p *CapabilityProjector) projectMetadataUpdated(ctx context.Context, event events.CapabilityMetadataUpdated) error {
	return p.readModel.UpdateMetadata(ctx, event.ID, readmodels.CapabilityMetadataUpdate{
		MaturityValue:  event.MaturityValue,
		OwnershipModel: event.OwnershipModel,
		PrimaryOwner:   event.PrimaryOwner,
		EAOwner:        event.EAOwner,
		Status:         event.Status,
	})
}

func (p *CapabilityProjector) projectExpertAdded(ctx context.Context, event events.CapabilityExpertAdded) error {
	return p.readModel.AddExpert(ctx, readmodels.ExpertInfo{
		CapabilityID: event.CapabilityID,
		Name:         event.ExpertName,
		Role:         event.ExpertRole,
		Contact:      event.ContactInfo,
		AddedAt:      event.AddedAt,
	})
}

func (p *CapabilityProjector) projectExpertRemoved(ctx context.Context, event events.CapabilityExpertRemoved) error {
	return p.readModel.RemoveExpert(ctx, readmodels.ExpertInfo{
		CapabilityID: event.CapabilityID,
		Name:         event.ExpertName,
		Role:         event.ExpertRole,
		Contact:      event.ContactInfo,
	})
}

func (p *CapabilityProjector) projectTagAdded(ctx context.Context, event events.CapabilityTagAdded) error {
	return p.readModel.AddTag(ctx, readmodels.TagInfo{
		CapabilityID: event.CapabilityID,
		Tag:          event.Tag,
		AddedAt:      event.AddedAt,
	})
}

func (p *CapabilityProjector) projectParentChanged(ctx context.Context, event events.CapabilityParentChanged) error {
	return p.readModel.UpdateParent(ctx, readmodels.ParentUpdate{
		ID:       event.CapabilityID,
		ParentID: event.NewParentID,
		Level:    event.NewLevel,
	})
}

func (p *CapabilityProjector) projectLevelChanged(ctx context.Context, event events.CapabilityLevelChanged) error {
	return p.readModel.UpdateLevel(ctx, event.CapabilityID, event.NewLevel)
}

func (p *CapabilityProjector) projectDeleted(ctx context.Context, event events.CapabilityDeleted) error {
	return p.readModel.Delete(ctx, event.ID)
}
