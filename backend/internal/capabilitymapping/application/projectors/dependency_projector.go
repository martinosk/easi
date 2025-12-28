package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type DependencyProjector struct {
	readModel *readmodels.DependencyReadModel
}

func NewDependencyProjector(readModel *readmodels.DependencyReadModel) *DependencyProjector {
	return &DependencyProjector{
		readModel: readModel,
	}
}

func (p *DependencyProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *DependencyProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "CapabilityDependencyCreated":
		var event events.CapabilityDependencyCreated
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal CapabilityDependencyCreated event: %v", err)
			return err
		}

		dto := readmodels.DependencyDTO{
			ID:                 event.ID,
			SourceCapabilityID: event.SourceCapabilityID,
			TargetCapabilityID: event.TargetCapabilityID,
			DependencyType:     event.DependencyType,
			Description:        event.Description,
			CreatedAt:          event.CreatedAt,
		}

		return p.readModel.Insert(ctx, dto)
	case "CapabilityDependencyDeleted":
		var event events.CapabilityDependencyDeleted
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal CapabilityDependencyDeleted event: %v", err)
			return err
		}

		return p.readModel.Delete(ctx, event.ID)
	}

	return nil
}
