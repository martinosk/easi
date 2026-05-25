package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type applicationComponentDeletedPayload struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	DeletedAt time.Time `json:"deletedAt"`
}

type applicationComponentCreatedPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type applicationComponentUpdatedPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type StaleApplicationReferenceStore interface {
	MarkApplicationStale(ctx context.Context, applicationID string) error
	CacheApplicationName(ctx context.Context, applicationID, name string) error
	UpdateApplicationName(ctx context.Context, applicationID, name string) error
}

type StaleApplicationProjector struct {
	readModel StaleApplicationReferenceStore
}

func NewStaleApplicationProjector(readModel StaleApplicationReferenceStore) *StaleApplicationProjector {
	return &StaleApplicationProjector{readModel: readModel}
}

func (p *StaleApplicationProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *StaleApplicationProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case amPL.ApplicationComponentDeleted:
		return p.handleDeleted(ctx, eventData)
	case amPL.ApplicationComponentCreated:
		return p.handleCreated(ctx, eventData)
	case amPL.ApplicationComponentUpdated:
		return p.handleUpdated(ctx, eventData)
	default:
		return nil
	}
}

func (p *StaleApplicationProjector) handleDeleted(ctx context.Context, eventData []byte) error {
	var payload applicationComponentDeletedPayload
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal ApplicationComponentDeleted payload: %w", err)
	}
	if payload.ID == "" {
		return nil
	}
	return p.readModel.MarkApplicationStale(ctx, payload.ID)
}

func (p *StaleApplicationProjector) handleCreated(ctx context.Context, eventData []byte) error {
	var payload applicationComponentCreatedPayload
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal ApplicationComponentCreated payload: %w", err)
	}
	if payload.ID == "" {
		return nil
	}
	if err := p.readModel.CacheApplicationName(ctx, payload.ID, payload.Name); err != nil {
		return err
	}
	return p.readModel.UpdateApplicationName(ctx, payload.ID, payload.Name)
}

func (p *StaleApplicationProjector) handleUpdated(ctx context.Context, eventData []byte) error {
	var payload applicationComponentUpdatedPayload
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal ApplicationComponentUpdated payload: %w", err)
	}
	if payload.ID == "" {
		return nil
	}
	if err := p.readModel.CacheApplicationName(ctx, payload.ID, payload.Name); err != nil {
		return err
	}
	return p.readModel.UpdateApplicationName(ctx, payload.ID, payload.Name)
}
