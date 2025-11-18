package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/shared/domain"
)

type RealizationProjector struct {
	readModel *readmodels.RealizationReadModel
}

func NewRealizationProjector(readModel *readmodels.RealizationReadModel) *RealizationProjector {
	return &RealizationProjector{
		readModel: readModel,
	}
}

func (p *RealizationProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *RealizationProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "SystemLinkedToCapability":
		var event events.SystemLinkedToCapability
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal SystemLinkedToCapability event: %v", err)
			return err
		}

		dto := readmodels.RealizationDTO{
			ID:               event.ID,
			CapabilityID:     event.CapabilityID,
			ComponentID:      event.ComponentID,
			RealizationLevel: event.RealizationLevel,
			Notes:            event.Notes,
			LinkedAt:         event.LinkedAt,
		}

		return p.readModel.Insert(ctx, dto)

	case "SystemRealizationUpdated":
		var event events.SystemRealizationUpdated
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal SystemRealizationUpdated event: %v", err)
			return err
		}

		return p.readModel.Update(ctx, event.ID, event.RealizationLevel, event.Notes)

	case "SystemRealizationDeleted":
		var event events.SystemRealizationDeleted
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal SystemRealizationDeleted event: %v", err)
			return err
		}

		return p.readModel.Delete(ctx, event.ID)
	}

	return nil
}
