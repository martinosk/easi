package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	domain "easi/backend/internal/shared/eventsourcing"
)

func dispatchReferenceEvent(ctx context.Context, event domain.DomainEvent, project func(context.Context, string, []byte) error) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return project(ctx, event.EventType(), eventData)
}

func dispatchByReferenceID[ID ~string](ctx context.Context, eventData []byte, run func(context.Context, ID) error) error {
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal reference id payload: %w", err)
	}
	if payload.ID == "" {
		return nil
	}
	return run(ctx, ID(payload.ID))
}

func dispatchReferenceNameChange[ID ~string](
	ctx context.Context,
	eventData []byte,
	cache func(context.Context, ID, string) error,
	update func(context.Context, ID, string) error,
) error {
	var payload struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal reference name-change payload: %w", err)
	}
	if payload.ID == "" {
		return nil
	}
	if err := cache(ctx, ID(payload.ID), payload.Name); err != nil {
		return err
	}
	return update(ctx, ID(payload.ID), payload.Name)
}

func dispatchReferenceUserCreated(ctx context.Context, eventData []byte, save func(context.Context, string, string) error) error {
	var payload struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal UserCreated payload: %w", err)
	}
	if payload.Email == "" || payload.Name == "" {
		return nil
	}
	return save(ctx, payload.Email, payload.Name)
}
