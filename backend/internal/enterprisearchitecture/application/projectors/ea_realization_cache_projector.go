package projectors

import (
	"context"
	"encoding/json"
	"log"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	domain "easi/backend/internal/shared/eventsourcing"
)

type EARealizationCacheProjector struct {
	readModel *readmodels.EARealizationCacheReadModel
}

func NewEARealizationCacheProjector(readModel *readmodels.EARealizationCacheReadModel) *EARealizationCacheProjector {
	return &EARealizationCacheProjector{readModel: readModel}
}

func (p *EARealizationCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *EARealizationCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		cmPL.SystemLinkedToCapability:   p.handleSystemLinkedToCapability,
		cmPL.SystemRealizationDeleted:   p.handleSystemRealizationDeleted,
		cmPL.CapabilityDeleted:          p.handleCapabilityDeleted,
		amPL.ApplicationComponentUpdated: p.handleApplicationComponentUpdated,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

type systemLinkedToCapabilityEvent struct {
	ID               string `json:"id"`
	CapabilityID     string `json:"capabilityId"`
	ComponentID      string `json:"componentId"`
	ComponentName    string `json:"componentName"`
	RealizationLevel string `json:"realizationLevel"`
}

func (p *EARealizationCacheProjector) handleSystemLinkedToCapability(ctx context.Context, eventData []byte) error {
	var event systemLinkedToCapabilityEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal SystemLinkedToCapability event: %v", err)
		return err
	}

	return p.readModel.Upsert(ctx, readmodels.RealizationEntry{
		RealizationID: event.ID,
		CapabilityID:  event.CapabilityID,
		ComponentID:   event.ComponentID,
		ComponentName: event.ComponentName,
		Origin:        event.RealizationLevel,
	})
}

type systemRealizationDeletedEvent struct {
	ID string `json:"id"`
}

func (p *EARealizationCacheProjector) handleSystemRealizationDeleted(ctx context.Context, eventData []byte) error {
	var event systemRealizationDeletedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal SystemRealizationDeleted event: %v", err)
		return err
	}

	return p.readModel.Delete(ctx, event.ID)
}

type realizationCapabilityDeletedEvent struct {
	ID string `json:"id"`
}

func (p *EARealizationCacheProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
	var event realizationCapabilityDeletedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityDeleted event: %v", err)
		return err
	}

	return p.readModel.DeleteByCapabilityID(ctx, event.ID)
}

type applicationComponentUpdatedEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (p *EARealizationCacheProjector) handleApplicationComponentUpdated(ctx context.Context, eventData []byte) error {
	var event applicationComponentUpdatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ApplicationComponentUpdated event: %v", err)
		return err
	}

	return p.readModel.UpdateComponentName(ctx, event.ID, event.Name)
}
