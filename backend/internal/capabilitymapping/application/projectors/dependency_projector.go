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
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *DependencyProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "CapabilityDependencyCreated":
		var event events.CapabilityDependencyCreated
		if err := json.Unmarshal(eventData, &event); err != nil {
			wrappedErr := fmt.Errorf("unmarshal CapabilityDependencyCreated event data: %w", err)
			log.Printf("failed to unmarshal CapabilityDependencyCreated event: %v", wrappedErr)
			return wrappedErr
		}

		dto := readmodels.DependencyDTO{
			ID:                 event.ID,
			SourceCapabilityID: event.SourceCapabilityID,
			TargetCapabilityID: event.TargetCapabilityID,
			DependencyType:     event.DependencyType,
			Description:        event.Description,
			CreatedAt:          event.CreatedAt,
		}

		if err := p.readModel.Insert(ctx, dto); err != nil {
			return fmt.Errorf("project CapabilityDependencyCreated for dependency %s: %w", event.ID, err)
		}
		return nil
	case "CapabilityDependencyDeleted":
		var event events.CapabilityDependencyDeleted
		if err := json.Unmarshal(eventData, &event); err != nil {
			wrappedErr := fmt.Errorf("unmarshal CapabilityDependencyDeleted event data: %w", err)
			log.Printf("failed to unmarshal CapabilityDependencyDeleted event: %v", wrappedErr)
			return wrappedErr
		}
		if err := p.readModel.Delete(ctx, event.ID); err != nil {
			return fmt.Errorf("project CapabilityDependencyDeleted for dependency %s: %w", event.ID, err)
		}
		return nil
	}

	return nil
}
