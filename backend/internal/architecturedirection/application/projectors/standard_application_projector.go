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

type StandardApplicationStore interface {
	UpsertCurrent(ctx context.Context, p readmodels.UpsertStandardApplicationParams) error
	AppendHistory(ctx context.Context, p readmodels.AppendStandardApplicationHistoryParams) error
}

type StandardApplicationProjector struct {
	readModel StandardApplicationStore
}

func NewStandardApplicationProjector(readModel StandardApplicationStore) *StandardApplicationProjector {
	return &StandardApplicationProjector{readModel: readModel}
}

func (p *StandardApplicationProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *StandardApplicationProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	if eventType != pl.StandardApplicationSet {
		return nil
	}
	return handleProjection(ctx, eventData, p.applyStandardApplicationSet)
}

func (p *StandardApplicationProjector) applyStandardApplicationSet(ctx context.Context, evt events.StandardApplicationSet) error {
	if err := p.readModel.UpsertCurrent(ctx, readmodels.UpsertStandardApplicationParams{
		ID:                     evt.ID,
		EnterpriseCapabilityID: evt.EnterpriseCapabilityID,
		ApplicationID:          evt.ApplicationID,
		Narrative:              evt.Narrative,
		SetAt:                  evt.OccurredOn,
	}); err != nil {
		return err
	}
	return p.readModel.AppendHistory(ctx, readmodels.AppendStandardApplicationHistoryParams{
		StandardApplicationID: evt.ID,
		ApplicationID:         evt.ApplicationID,
		PreviousApplicationID: evt.PreviousApplicationID,
		Narrative:             evt.Narrative,
		SetAt:                 evt.OccurredOn,
	})
}
