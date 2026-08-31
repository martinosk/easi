package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"easi/backend/internal/architecturedirection/application/readmodels"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
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

func decodeRealizationEvent[T any](eventName string, eventData []byte) (T, error) {
	var event T
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal %s event data in EA realization cache projector: %w", eventName, err)
		log.Printf("failed to unmarshal %s event: %v", eventName, wrappedErr)
		return event, wrappedErr
	}
	return event, nil
}

type systemLinkedToCapabilityEvent struct {
	ID               string `json:"id"`
	CapabilityID     string `json:"capabilityId"`
	ComponentID      string `json:"componentId"`
	ComponentName    string `json:"componentName"`
	RealizationLevel string `json:"realizationLevel"`
}

func (p *EARealizationCacheProjector) handleSystemLinkedToCapability(ctx context.Context, eventData []byte) error {
	event, err := decodeRealizationEvent[systemLinkedToCapabilityEvent](cmPL.SystemLinkedToCapability, eventData)
	if err != nil {
		return err
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

type identifiedEvent struct {
	ID string `json:"id"`
}

type identifiedEventAction struct {
	eventName          string
	failureDescription string
	run                func(ctx context.Context, id string) error
}

func handleIdentifiedEvent(ctx context.Context, eventData []byte, action identifiedEventAction) error {
	event, err := decodeRealizationEvent[identifiedEvent](action.eventName, eventData)
	if err != nil {
		return err
	}
	if err := action.run(ctx, event.ID); err != nil {
		return fmt.Errorf("project %s EA realization cache %s %s: %w", action.eventName, action.failureDescription, event.ID, err)
	}
	return nil
}

func (p *EARealizationCacheProjector) handleSystemRealizationDeleted(ctx context.Context, eventData []byte) error {
	return handleIdentifiedEvent(ctx, eventData, identifiedEventAction{
		eventName:          cmPL.SystemRealizationDeleted,
		failureDescription: "delete for realization",
		run:                p.readModel.Delete,
	})
}

func (p *EARealizationCacheProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
	return handleIdentifiedEvent(ctx, eventData, identifiedEventAction{
		eventName:          cmPL.CapabilityDeleted,
		failureDescription: "delete by capability",
		run:                p.readModel.DeleteByCapabilityID,
	})
}

type applicationComponentUpdatedEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (p *EARealizationCacheProjector) handleApplicationComponentUpdated(ctx context.Context, eventData []byte) error {
	event, err := decodeRealizationEvent[applicationComponentUpdatedEvent](amPL.ApplicationComponentUpdated, eventData)
	if err != nil {
		return err
	}
	if err := p.readModel.UpdateComponentName(ctx, event.ID, event.Name); err != nil {
		return fmt.Errorf("project ApplicationComponentUpdated EA realization cache component rename for component %s: %w", event.ID, err)
	}
	return nil
}
