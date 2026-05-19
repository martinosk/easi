package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type StaleReferenceStore interface {
	MarkSourceCapabilityStale(ctx context.Context, capabilityID string) error
}

type StaleReferenceProjector struct {
	readModel StaleReferenceStore
}

func NewStaleReferenceProjector(readModel StaleReferenceStore) *StaleReferenceProjector {
	return &StaleReferenceProjector{readModel: readModel}
}

func (p *StaleReferenceProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *StaleReferenceProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	if eventType != cmPL.CapabilityDeleted {
		return nil
	}
	var generic struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(eventData, &generic); err != nil {
		return err
	}
	if generic.ID == "" {
		return nil
	}
	return p.readModel.MarkSourceCapabilityStale(ctx, generic.ID)
}
