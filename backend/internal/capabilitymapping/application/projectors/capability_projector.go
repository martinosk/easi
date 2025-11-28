package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/shared/domain"
)

type CapabilityProjector struct {
	readModel *readmodels.CapabilityReadModel
}

func NewCapabilityProjector(readModel *readmodels.CapabilityReadModel) *CapabilityProjector {
	return &CapabilityProjector{
		readModel: readModel,
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
		"CapabilityTagAdded":        p.handleCapabilityTagAdded,
		"CapabilityParentChanged":   p.handleCapabilityParentChanged,
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
	return p.readModel.Update(ctx, event.ID, event.Name, event.Description)
}

func (p *CapabilityProjector) handleCapabilityMetadataUpdated(ctx context.Context, eventData []byte) error {
	var event events.CapabilityMetadataUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityMetadataUpdated event: %v", err)
		return err
	}

	return p.readModel.UpdateMetadata(ctx, event.ID, readmodels.CapabilityMetadataUpdate{
		StrategyPillar: event.StrategyPillar,
		PillarWeight:   event.PillarWeight,
		MaturityLevel:  event.MaturityLevel,
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
	return p.readModel.AddExpert(ctx, event.CapabilityID, event.ExpertName, event.ExpertRole, event.ContactInfo, event.AddedAt)
}

func (p *CapabilityProjector) handleCapabilityTagAdded(ctx context.Context, eventData []byte) error {
	var event events.CapabilityTagAdded
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityTagAdded event: %v", err)
		return err
	}
	return p.readModel.AddTag(ctx, event.CapabilityID, event.Tag, event.AddedAt)
}

func (p *CapabilityProjector) handleCapabilityParentChanged(ctx context.Context, eventData []byte) error {
	var event events.CapabilityParentChanged
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityParentChanged event: %v", err)
		return err
	}
	return p.readModel.UpdateParent(ctx, event.CapabilityID, event.NewParentID, event.NewLevel)
}

func (p *CapabilityProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
	var event events.CapabilityDeleted
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityDeleted event: %v", err)
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}
