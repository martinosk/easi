package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"easi/backend/internal/architecturedirection/application/readmodels"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

// capabilityDeletedPayload mirrors capabilitymapping's CapabilityDeletedPayload.
// The architecture-test forbids importing /publishedlanguage/contracts across
// bounded contexts, so the shape is duplicated here. If the upstream contract
// changes, update this struct alongside it.
type capabilityDeletedPayload struct {
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

type StaleReferenceStore interface {
	MarkSourceCapabilityStale(ctx context.Context, capabilityID readmodels.CapabilityID) error
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
	var payload capabilityDeletedPayload
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal CapabilityDeleted payload: %w", err)
	}
	if payload.ID == "" {
		return nil
	}
	return p.readModel.MarkSourceCapabilityStale(ctx, readmodels.CapabilityID(payload.ID))
}
