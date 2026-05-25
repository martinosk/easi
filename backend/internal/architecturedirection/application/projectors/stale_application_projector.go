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

type StaleApplicationReferenceStore interface {
	MarkApplicationStale(ctx context.Context, applicationID string) error
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
	if eventType != amPL.ApplicationComponentDeleted {
		return nil
	}
	var payload applicationComponentDeletedPayload
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal ApplicationComponentDeleted payload: %w", err)
	}
	if payload.ID == "" {
		return nil
	}
	return p.readModel.MarkApplicationStale(ctx, payload.ID)
}
