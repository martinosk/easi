package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/events"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type RealizationRoleStore interface {
	RegisterAggregate(ctx context.Context, capabilityID, aggregateID string) error
	UpsertRole(ctx context.Context, p readmodels.UpsertRealizationRoleParams) error
	DeleteRole(ctx context.Context, capabilityID, componentID string) error
}

type RealizationRoleProjector struct {
	readModel RealizationRoleStore
}

func NewRealizationRoleProjector(readModel RealizationRoleStore) *RealizationRoleProjector {
	return &RealizationRoleProjector{readModel: readModel}
}

func (p *RealizationRoleProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *RealizationRoleProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case pl.RealizationRoleAssigned:
		return p.handleAssigned(ctx, eventData)
	case pl.RealizationRoleCleared:
		return p.handleCleared(ctx, eventData)
	default:
		return nil
	}
}

func (p *RealizationRoleProjector) handleAssigned(ctx context.Context, eventData []byte) error {
	var evt events.RealizationRoleAssigned
	if err := json.Unmarshal(eventData, &evt); err != nil {
		return fmt.Errorf("unmarshal RealizationRoleAssigned payload: %w", err)
	}
	if err := p.readModel.RegisterAggregate(ctx, evt.CapabilityID, evt.ID); err != nil {
		return err
	}
	return p.readModel.UpsertRole(ctx, readmodels.UpsertRealizationRoleParams{
		CapabilityID:         evt.CapabilityID,
		ComponentID:          evt.ComponentID,
		RealizationID:        evt.RealizationID,
		Role:                 evt.Role,
		AssignedBy:           evt.AssignedBy,
		AssignedAt:           evt.OccurredOn,
		AggregateID:          evt.ID,
		DisplacedComponentID: evt.DisplacedComponentID,
	})
}

func (p *RealizationRoleProjector) handleCleared(ctx context.Context, eventData []byte) error {
	var evt events.RealizationRoleCleared
	if err := json.Unmarshal(eventData, &evt); err != nil {
		return fmt.Errorf("unmarshal RealizationRoleCleared payload: %w", err)
	}
	return p.readModel.DeleteRole(ctx, evt.CapabilityID, evt.ComponentID)
}
