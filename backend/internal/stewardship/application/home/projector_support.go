package home

import (
	"context"
	"encoding/json"
	"fmt"

	domain "easi/backend/internal/shared/eventsourcing"
)

type eventProjection[P any] func(ctx context.Context, payload P) error

type projection struct {
	project func(ctx context.Context, event domain.DomainEvent) error
}

func on[P any](handle eventProjection[P]) projection {
	return projection{project: func(ctx context.Context, event domain.DomainEvent) error {
		payload, err := decodePayload[P](event)
		if err != nil {
			return err
		}
		if err := handle(ctx, payload); err != nil {
			return fmt.Errorf("project %s of %s: %w", event.EventType(), event.AggregateID(), err)
		}
		return nil
	}}
}

type projections map[string]projection

func (p projections) Handle(ctx context.Context, event domain.DomainEvent) error {
	if handler, tracked := p[event.EventType()]; tracked {
		return handler.project(ctx, event)
	}
	return nil
}

func decodePayload[T any](event domain.DomainEvent) (T, error) {
	var payload T
	data, err := json.Marshal(event.EventData())
	if err != nil {
		return payload, fmt.Errorf("marshal %s payload of %s: %w", event.EventType(), event.AggregateID(), err)
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return payload, fmt.Errorf("unmarshal %s payload of %s: %w", event.EventType(), event.AggregateID(), err)
	}
	return payload, nil
}
