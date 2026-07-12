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

type TimeAssessmentStore interface {
	UpsertCurrent(ctx context.Context, p readmodels.UpsertTimeAssessmentParams) error
	Delete(ctx context.Context, id string) error
}

type TimeAssessmentProjector struct {
	readModel TimeAssessmentStore
}

func NewTimeAssessmentProjector(readModel TimeAssessmentStore) *TimeAssessmentProjector {
	return &TimeAssessmentProjector{readModel: readModel}
}

func (p *TimeAssessmentProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *TimeAssessmentProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case pl.TimeAssessmentRecorded:
		return p.handleRecorded(ctx, eventData)
	case pl.TimeAssessmentRemoved:
		return p.handleRemoved(ctx, eventData)
	default:
		return nil
	}
}

func (p *TimeAssessmentProjector) handleRecorded(ctx context.Context, eventData []byte) error {
	var evt events.TimeAssessmentRecorded
	if err := json.Unmarshal(eventData, &evt); err != nil {
		return fmt.Errorf("unmarshal TimeAssessmentRecorded payload: %w", err)
	}
	return p.readModel.UpsertCurrent(ctx, readmodels.UpsertTimeAssessmentParams{
		ID:            evt.ID,
		CapabilityID:  evt.CapabilityID,
		ComponentID:   evt.ComponentID,
		RealizationID: evt.RealizationID,
		Grade:         evt.Grade,
		Rationale:     evt.Rationale,
		AssessedBy:    evt.AssessedBy,
		AssessedAt:    evt.OccurredOn,
	})
}

func (p *TimeAssessmentProjector) handleRemoved(ctx context.Context, eventData []byte) error {
	var evt events.TimeAssessmentRemoved
	if err := json.Unmarshal(eventData, &evt); err != nil {
		return fmt.Errorf("unmarshal TimeAssessmentRemoved payload: %w", err)
	}
	return p.readModel.Delete(ctx, evt.ID)
}
