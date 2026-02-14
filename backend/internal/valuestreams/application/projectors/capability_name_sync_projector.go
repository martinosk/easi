package projectors

import (
	"context"
	"encoding/json"
	"log"

	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type StageCapabilityNameUpdater interface {
	UpdateStageCapabilityName(ctx context.Context, capabilityID, name string) error
}

type CapabilityNameSyncProjector struct {
	readModel StageCapabilityNameUpdater
}

func NewCapabilityNameSyncProjector(readModel StageCapabilityNameUpdater) *CapabilityNameSyncProjector {
	return &CapabilityNameSyncProjector{readModel: readModel}
}

func (p *CapabilityNameSyncProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *CapabilityNameSyncProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	if eventType != cmPL.CapabilityUpdated {
		return nil
	}

	var event struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityUpdated event for name sync: %v", err)
		return err
	}

	return p.readModel.UpdateStageCapabilityName(ctx, event.ID, event.Name)
}
