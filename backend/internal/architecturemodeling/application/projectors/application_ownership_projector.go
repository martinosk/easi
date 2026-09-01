package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ComponentOwnershipWriter interface {
	SetOwnership(ctx context.Context, componentID string, record readmodels.OwnershipRecord) error
	ClearOwnership(ctx context.Context, componentID string) error
}

type ApplicationOwnershipProjector struct {
	writer ComponentOwnershipWriter
}

func NewApplicationOwnershipProjector(writer ComponentOwnershipWriter) *ApplicationOwnershipProjector {
	return &ApplicationOwnershipProjector{writer: writer}
}

func (p *ApplicationOwnershipProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *ApplicationOwnershipProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case archPL.ApplicationOwnerNominated:
		return p.projectNominated(ctx, eventData)
	case archPL.ApplicationOwnershipConfirmed:
		return p.projectConfirmed(ctx, eventData)
	case archPL.ApplicationOwnerAssigned:
		return p.projectAssigned(ctx, eventData)
	case archPL.ApplicationOwnershipCleared:
		return p.projectCleared(ctx, eventData)
	}
	return nil
}

func (p *ApplicationOwnershipProjector) projectNominated(ctx context.Context, eventData []byte) error {
	return projectEvent(ctx, eventData, "ApplicationOwnerNominated", func(ctx context.Context, event *events.ApplicationOwnerNominated) error {
		return p.writer.SetOwnership(ctx, event.ComponentID, readmodels.OwnershipRecord{
			State:     valueobjects.OwnershipStateNominated,
			OwnerKind: event.OwnerKind,
			OwnerID:   event.OwnerID,
		})
	})
}

func (p *ApplicationOwnershipProjector) projectConfirmed(ctx context.Context, eventData []byte) error {
	return projectEvent(ctx, eventData, "ApplicationOwnershipConfirmed", func(ctx context.Context, event *events.ApplicationOwnershipConfirmed) error {
		return p.writer.SetOwnership(ctx, event.ComponentID, readmodels.OwnershipRecord{
			State:     event.OwnershipState,
			OwnerKind: event.OwnerKind,
			OwnerID:   event.OwnerID,
		})
	})
}

func (p *ApplicationOwnershipProjector) projectAssigned(ctx context.Context, eventData []byte) error {
	return projectEvent(ctx, eventData, "ApplicationOwnerAssigned", func(ctx context.Context, event *events.ApplicationOwnerAssigned) error {
		return p.writer.SetOwnership(ctx, event.ComponentID, readmodels.OwnershipRecord{
			State:     event.OwnershipState,
			OwnerKind: event.OwnerKind,
			OwnerID:   event.OwnerID,
		})
	})
}

func (p *ApplicationOwnershipProjector) projectCleared(ctx context.Context, eventData []byte) error {
	return projectEvent(ctx, eventData, "ApplicationOwnershipCleared", func(ctx context.Context, event *events.ApplicationOwnershipCleared) error {
		return p.writer.ClearOwnership(ctx, event.ComponentID)
	})
}
