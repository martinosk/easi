package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/events"
	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ComponentContainmentWriter interface {
	SetParent(ctx context.Context, record readmodels.ContainmentRecord) error
	ClearParent(ctx context.Context, partID string) error
	RegisterContainmentsAggregate(ctx context.Context, aggregateID string) error
}

type ComponentContainmentProjector struct {
	writer ComponentContainmentWriter
}

func NewComponentContainmentProjector(writer ComponentContainmentWriter) *ComponentContainmentProjector {
	return &ComponentContainmentProjector{writer: writer}
}

func (p *ComponentContainmentProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *ComponentContainmentProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case archPL.ComponentAttached:
		return projectEvent(ctx, eventData, "ComponentAttached", p.projectAttached)
	case archPL.ComponentDetached:
		return projectEvent(ctx, eventData, "ComponentDetached", func(ctx context.Context, event *events.ComponentDetached) error {
			return p.writer.ClearParent(ctx, event.PartID)
		})
	}
	return nil
}

func (p *ComponentContainmentProjector) projectAttached(ctx context.Context, event *events.ComponentAttached) error {
	if err := p.writer.RegisterContainmentsAggregate(ctx, event.ContainmentsID); err != nil {
		return err
	}
	return p.writer.SetParent(ctx, readmodels.ContainmentRecord{PartID: event.PartID, ParentID: event.ParentID, Kind: event.Kind})
}
