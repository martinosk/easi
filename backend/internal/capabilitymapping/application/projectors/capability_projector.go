package projectors

import (
	"context"
	"encoding/json"
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

func (p *CapabilityProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"CapabilityCreated":         p.handleCapabilityCreated,
		"CapabilityUpdated":         p.handleCapabilityUpdated,
		"CapabilityMetadataUpdated": p.handleCapabilityMetadataUpdated,
		"CapabilityExpertAdded":     p.handleCapabilityExpertAdded,
		"CapabilityExpertRemoved":   p.handleCapabilityExpertRemoved,
		"CapabilityTagAdded":        p.handleCapabilityTagAdded,
		"CapabilityParentChanged":   p.handleCapabilityParentChanged,
		"CapabilityLevelChanged":    p.handleCapabilityLevelChanged,
		"CapabilityDeleted":         p.handleCapabilityDeleted,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *CapabilityProjector) handleCapabilityCreated(ctx context.Context, eventData []byte) error {
	var event events.CapabilityCreated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityCreated event: %v", err)
		return err
	}

	dto := readmodels.CapabilityDTO{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		ParentID:    event.ParentID,
		Level:       event.Level,
		CreatedAt:   event.CreatedAt,
	}
	return p.readModel.Insert(ctx, dto)
}

func (p *CapabilityProjector) handleCapabilityUpdated(ctx context.Context, eventData []byte) error {
	var event events.CapabilityUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityUpdated event: %v", err)
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
		return nil
	}

	if err := p.assignmentReadModel.UpdateCapabilityInfo(ctx, event.ID, capability.Name, capability.Description, capability.Level); err != nil {
		log.Printf("Failed to update assignments for capability %s: %v", event.ID, err)
		return err
	}

	return nil
}

func (p *CapabilityProjector) handleCapabilityMetadataUpdated(ctx context.Context, eventData []byte) error {
	var event events.CapabilityMetadataUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityMetadataUpdated event: %v", err)
		return err
	}

	return p.readModel.UpdateMetadata(ctx, event.ID, readmodels.CapabilityMetadataUpdate{
		MaturityValue:  event.MaturityValue,
		OwnershipModel: event.OwnershipModel,
		PrimaryOwner:   event.PrimaryOwner,
		EAOwner:        event.EAOwner,
		Status:         event.Status,
	})
}

func (p *CapabilityProjector) handleCapabilityExpertAdded(ctx context.Context, eventData []byte) error {
	var event events.CapabilityExpertAdded
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityExpertAdded event: %v", err)
		return err
	}
	return p.readModel.AddExpert(ctx, readmodels.ExpertInfo{
		CapabilityID: event.CapabilityID,
		Name:         event.ExpertName,
		Role:         event.ExpertRole,
		Contact:      event.ContactInfo,
		AddedAt:      event.AddedAt,
	})
}

func (p *CapabilityProjector) handleCapabilityExpertRemoved(ctx context.Context, eventData []byte) error {
	var event events.CapabilityExpertRemoved
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityExpertRemoved event: %v", err)
		return err
	}
	return p.readModel.RemoveExpert(ctx, readmodels.ExpertInfo{
		CapabilityID: event.CapabilityID,
		Name:         event.ExpertName,
		Role:         event.ExpertRole,
		Contact:      event.ContactInfo,
	})
}

func (p *CapabilityProjector) handleCapabilityTagAdded(ctx context.Context, eventData []byte) error {
	var event events.CapabilityTagAdded
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityTagAdded event: %v", err)
		return err
	}
	return p.readModel.AddTag(ctx, readmodels.TagInfo{
		CapabilityID: event.CapabilityID,
		Tag:          event.Tag,
		AddedAt:      event.AddedAt,
	})
}

func (p *CapabilityProjector) handleCapabilityParentChanged(ctx context.Context, eventData []byte) error {
	var event events.CapabilityParentChanged
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityParentChanged event: %v", err)
		return err
	}
	return p.readModel.UpdateParent(ctx, readmodels.ParentUpdate{
		ID:       event.CapabilityID,
		ParentID: event.NewParentID,
		Level:    event.NewLevel,
	})
}

func (p *CapabilityProjector) handleCapabilityLevelChanged(ctx context.Context, eventData []byte) error {
	var event events.CapabilityLevelChanged
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityLevelChanged event: %v", err)
		return err
	}
	return p.readModel.UpdateLevel(ctx, event.CapabilityID, event.NewLevel)
}

func (p *CapabilityProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
	var event events.CapabilityDeleted
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityDeleted event: %v", err)
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}
