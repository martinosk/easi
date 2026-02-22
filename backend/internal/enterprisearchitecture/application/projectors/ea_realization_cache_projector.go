package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	domain "easi/backend/internal/shared/eventsourcing"
)

type RealizationCacheWriter interface {
	Upsert(ctx context.Context, entry readmodels.RealizationEntry) error
	Delete(ctx context.Context, realizationID string) error
	DeleteByCapabilityID(ctx context.Context, capabilityID string) error
	UpdateComponentName(ctx context.Context, componentID, componentName string) error
}

type EARealizationCacheProjector struct {
	readModel RealizationCacheWriter
}

func NewEARealizationCacheProjector(readModel RealizationCacheWriter) *EARealizationCacheProjector {
	return &EARealizationCacheProjector{readModel: readModel}
}

func (p *EARealizationCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *EARealizationCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		cmPL.SystemLinkedToCapability:    p.handleSystemLinkedToCapability,
		cmPL.SystemRealizationDeleted:    p.handleSystemRealizationDeleted,
		cmPL.CapabilityDeleted:           p.handleCapabilityDeleted,
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
		wrappedErr := fmt.Errorf("unmarshal SystemLinkedToCapability event data in EA realization cache projector: %w", err)
		log.Printf("failed to unmarshal SystemLinkedToCapability event: %v", wrappedErr)
		return wrappedErr
	}
	if err := p.readModel.Upsert(ctx, readmodels.RealizationEntry{
		RealizationID: event.ID,
		CapabilityID:  event.CapabilityID,
		ComponentID:   event.ComponentID,
		ComponentName: event.ComponentName,
		Origin:        event.RealizationLevel,
	}); err != nil {
		return fmt.Errorf("project SystemLinkedToCapability EA realization cache upsert for realization %s: %w", event.ID, err)
	}
	return nil
}

type systemRealizationDeletedEvent struct {
	ID string `json:"id"`
}

func (p *EARealizationCacheProjector) handleSystemRealizationDeleted(ctx context.Context, eventData []byte) error {
	var event systemRealizationDeletedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal SystemRealizationDeleted event data in EA realization cache projector: %w", err)
		log.Printf("failed to unmarshal SystemRealizationDeleted event: %v", wrappedErr)
		return wrappedErr
	}
	if err := p.readModel.Delete(ctx, event.ID); err != nil {
		return fmt.Errorf("project SystemRealizationDeleted EA realization cache delete for realization %s: %w", event.ID, err)
	}
	return nil
}

type realizationCapabilityDeletedEvent struct {
	ID string `json:"id"`
}

func (p *EARealizationCacheProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
	var event realizationCapabilityDeletedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal CapabilityDeleted event data in EA realization cache projector: %w", err)
		log.Printf("failed to unmarshal CapabilityDeleted event: %v", wrappedErr)
		return wrappedErr
	}
	if err := p.readModel.DeleteByCapabilityID(ctx, event.ID); err != nil {
		return fmt.Errorf("project CapabilityDeleted EA realization cache delete by capability %s: %w", event.ID, err)
	}
	return nil
}

type applicationComponentUpdatedEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (p *EARealizationCacheProjector) handleApplicationComponentUpdated(ctx context.Context, eventData []byte) error {
	var event applicationComponentUpdatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal ApplicationComponentUpdated event data in EA realization cache projector: %w", err)
		log.Printf("failed to unmarshal ApplicationComponentUpdated event: %v", wrappedErr)
		return wrappedErr
	}
	if err := p.readModel.UpdateComponentName(ctx, event.ID, event.Name); err != nil {
		return fmt.Errorf("project ApplicationComponentUpdated EA realization cache component rename for component %s: %w", event.ID, err)
	}
	return nil
}
