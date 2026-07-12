package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type RealizationRoleReferenceStore interface {
	DeleteByCapabilityID(ctx context.Context, capabilityID string) error
	DeleteByComponentID(ctx context.Context, componentID string) error
	CacheCapabilityName(ctx context.Context, capabilityID, name string) error
	UpdateCapabilityName(ctx context.Context, capabilityID, name string) error
	CacheComponentName(ctx context.Context, componentID, name string) error
	UpdateComponentName(ctx context.Context, componentID, name string) error
}

type RealizationRoleReferenceProjector struct {
	readModel RealizationRoleReferenceStore
}

func NewRealizationRoleReferenceProjector(readModel RealizationRoleReferenceStore) *RealizationRoleReferenceProjector {
	return &RealizationRoleReferenceProjector{readModel: readModel}
}

func (p *RealizationRoleReferenceProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *RealizationRoleReferenceProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case cmPL.CapabilityDeleted:
		return p.handleIdentified(ctx, eventData, p.readModel.DeleteByCapabilityID)
	case amPL.ApplicationComponentDeleted:
		return p.handleIdentified(ctx, eventData, p.readModel.DeleteByComponentID)
	case cmPL.CapabilityCreated, cmPL.CapabilityUpdated:
		return p.handleNameChange(ctx, eventData, p.readModel.CacheCapabilityName, p.readModel.UpdateCapabilityName)
	case amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated:
		return p.handleNameChange(ctx, eventData, p.readModel.CacheComponentName, p.readModel.UpdateComponentName)
	default:
		return nil
	}
}

func (p *RealizationRoleReferenceProjector) handleIdentified(ctx context.Context, eventData []byte, run func(context.Context, string) error) error {
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal identified event payload: %w", err)
	}
	if payload.ID == "" {
		return nil
	}
	return run(ctx, payload.ID)
}

func (p *RealizationRoleReferenceProjector) handleNameChange(
	ctx context.Context,
	eventData []byte,
	cache func(context.Context, string, string) error,
	update func(context.Context, string, string) error,
) error {
	var payload struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal name-change event payload: %w", err)
	}
	if payload.ID == "" {
		return nil
	}
	if err := cache(ctx, payload.ID, payload.Name); err != nil {
		return err
	}
	return update(ctx, payload.ID, payload.Name)
}
