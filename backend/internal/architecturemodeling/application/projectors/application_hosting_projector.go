package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	"easi/backend/internal/architecturemodeling/domain/events"
	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ComponentHostingWriter interface {
	SetHosting(ctx context.Context, componentID, hosting string) error
}

type ApplicationHostingProjector struct {
	writer ComponentHostingWriter
}

func NewApplicationHostingProjector(writer ComponentHostingWriter) *ApplicationHostingProjector {
	return &ApplicationHostingProjector{writer: writer}
}

func (p *ApplicationHostingProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *ApplicationHostingProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	if eventType != archPL.ApplicationHostingClassified {
		return nil
	}
	return projectEvent(ctx, eventData, "ApplicationHostingClassified", func(ctx context.Context, event *events.ApplicationHostingClassified) error {
		return p.writer.SetHosting(ctx, event.ComponentID, event.Hosting)
	})
}
