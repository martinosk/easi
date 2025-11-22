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
	switch eventType {
	case "CapabilityCreated":
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
	case "CapabilityUpdated":
		var event events.CapabilityUpdated
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal CapabilityUpdated event: %v", err)
			return err
		}

		return p.readModel.Update(ctx, event.ID, event.Name, event.Description)
	case "CapabilityMetadataUpdated":
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
	case "CapabilityExpertAdded":
		var event events.CapabilityExpertAdded
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal CapabilityExpertAdded event: %v", err)
			return err
		}

		return p.readModel.AddExpert(ctx, event.CapabilityID, event.ExpertName, event.ExpertRole, event.ContactInfo, event.AddedAt)
	case "CapabilityTagAdded":
		var event events.CapabilityTagAdded
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal CapabilityTagAdded event: %v", err)
			return err
		}

		return p.readModel.AddTag(ctx, event.CapabilityID, event.Tag, event.AddedAt)
	case "CapabilityParentChanged":
		var event events.CapabilityParentChanged
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal CapabilityParentChanged event: %v", err)
			return err
		}

		return p.readModel.UpdateParent(ctx, event.CapabilityID, event.NewParentID, event.NewLevel)
	case "CapabilityDeleted":
		var event events.CapabilityDeleted
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal CapabilityDeleted event: %v", err)
			return err
		}

		return p.readModel.Delete(ctx, event.ID)
	}

	return nil
}
