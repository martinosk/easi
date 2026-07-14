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

type CapabilityJourneyStore interface {
	InsertJourney(ctx context.Context, p readmodels.InsertJourneyParams) error
	UpdateStatus(ctx context.Context, p readmodels.UpdateJourneyStatusParams) error
	UpdateProgress(ctx context.Context, journeyID string, progress int) error
	UpdateDetails(ctx context.Context, p readmodels.UpdateJourneyDetailsParams) error
	ReplaceSources(ctx context.Context, journeyID string, componentIDs []string) error
	AddMilestone(ctx context.Context, p readmodels.JourneyMilestoneUpsertParams) error
	UpdateMilestone(ctx context.Context, p readmodels.JourneyMilestoneUpsertParams) error
	RemoveMilestone(ctx context.Context, journeyID, milestoneID string) error
}

type CapabilityJourneyProjector struct {
	readModel CapabilityJourneyStore
}

func NewCapabilityJourneyProjector(readModel CapabilityJourneyStore) *CapabilityJourneyProjector {
	return &CapabilityJourneyProjector{readModel: readModel}
}

func (p *CapabilityJourneyProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *CapabilityJourneyProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		pl.JourneyPlanned: func(ctx context.Context, data []byte) error {
			return handleProjection(ctx, data, p.applyJourneyPlanned)
		},
		pl.JourneyStarted: func(ctx context.Context, data []byte) error {
			return handleProjection(ctx, data, p.applyJourneyStarted)
		},
		pl.JourneyCompleted: func(ctx context.Context, data []byte) error {
			return handleProjection(ctx, data, p.applyJourneyCompleted)
		},
		pl.JourneyAbandoned: func(ctx context.Context, data []byte) error {
			return handleProjection(ctx, data, p.applyJourneyAbandoned)
		},
		pl.JourneyProgressUpdated: func(ctx context.Context, data []byte) error {
			return handleProjection(ctx, data, p.applyJourneyProgressUpdated)
		},
		pl.JourneyDetailsUpdated: func(ctx context.Context, data []byte) error {
			return handleProjection(ctx, data, p.applyJourneyDetailsUpdated)
		},
		pl.JourneySourceApplicationsChanged: func(ctx context.Context, data []byte) error {
			return handleProjection(ctx, data, p.applyJourneySourceApplicationsChanged)
		},
		pl.JourneyMilestoneAdded: func(ctx context.Context, data []byte) error {
			return handleProjection(ctx, data, p.applyJourneyMilestoneAdded)
		},
		pl.JourneyMilestoneUpdated: func(ctx context.Context, data []byte) error {
			return handleProjection(ctx, data, p.applyJourneyMilestoneUpdated)
		},
		pl.JourneyMilestoneRemoved: func(ctx context.Context, data []byte) error {
			return handleProjection(ctx, data, p.applyJourneyMilestoneRemoved)
		},
	}
	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *CapabilityJourneyProjector) applyJourneyPlanned(ctx context.Context, evt events.JourneyPlanned) error {
	year, quarter := targetPeriodParts(evt.TargetPeriod)
	return p.readModel.InsertJourney(ctx, readmodels.InsertJourneyParams{
		ID:               evt.ID,
		CapabilityID:     evt.CapabilityID,
		Kind:             evt.Kind,
		FromComponentIDs: evt.FromComponentIDs,
		ToComponentID:    evt.ToComponentID,
		Note:             evt.Note,
		TargetYear:       year,
		TargetQuarter:    quarter,
		TargetDomainID:   evt.TargetDomainID,
		TargetParentID:   evt.TargetParentID,
		ResultingName:    evt.ResultingName,
		PlannedBy:        evt.PlannedBy,
		PlannedAt:        evt.OccurredOn,
	})
}

func (p *CapabilityJourneyProjector) applyJourneyStarted(ctx context.Context, evt events.JourneyStarted) error {
	return p.readModel.UpdateStatus(ctx, readmodels.UpdateJourneyStatusParams{
		JourneyID: evt.ID, Status: "in-flight", Column: readmodels.JourneyTimestampStarted, OccurredAt: evt.OccurredOn,
	})
}

func (p *CapabilityJourneyProjector) applyJourneyCompleted(ctx context.Context, evt events.JourneyCompleted) error {
	return p.readModel.UpdateStatus(ctx, readmodels.UpdateJourneyStatusParams{
		JourneyID: evt.ID, Status: "done", Column: readmodels.JourneyTimestampCompleted, OccurredAt: evt.OccurredOn,
	})
}

func (p *CapabilityJourneyProjector) applyJourneyAbandoned(ctx context.Context, evt events.JourneyAbandoned) error {
	return p.readModel.UpdateStatus(ctx, readmodels.UpdateJourneyStatusParams{
		JourneyID: evt.ID, Status: "abandoned", Column: readmodels.JourneyTimestampAbandoned, OccurredAt: evt.OccurredOn,
	})
}

func (p *CapabilityJourneyProjector) applyJourneyProgressUpdated(ctx context.Context, evt events.JourneyProgressUpdated) error {
	return p.readModel.UpdateProgress(ctx, evt.ID, evt.Progress)
}

func (p *CapabilityJourneyProjector) applyJourneyDetailsUpdated(ctx context.Context, evt events.JourneyDetailsUpdated) error {
	year, quarter := targetPeriodParts(evt.TargetPeriod)
	return p.readModel.UpdateDetails(ctx, readmodels.UpdateJourneyDetailsParams{
		JourneyID: evt.ID, Note: evt.Note, TargetYear: year, TargetQuarter: quarter, ResultingName: evt.ResultingName,
	})
}

func (p *CapabilityJourneyProjector) applyJourneySourceApplicationsChanged(ctx context.Context, evt events.JourneySourceApplicationsChanged) error {
	return p.readModel.ReplaceSources(ctx, evt.ID, evt.FromComponentIDs)
}

func (p *CapabilityJourneyProjector) applyJourneyMilestoneAdded(ctx context.Context, evt events.JourneyMilestoneAdded) error {
	year, quarter := targetPeriodParts(evt.TargetPeriod)
	return p.readModel.AddMilestone(ctx, readmodels.JourneyMilestoneUpsertParams{
		JourneyID: evt.ID, MilestoneID: evt.MilestoneID, Label: evt.Label,
		TargetYear: year, TargetQuarter: quarter, Status: evt.Status, UpdatedAt: evt.OccurredOn,
	})
}

func (p *CapabilityJourneyProjector) applyJourneyMilestoneUpdated(ctx context.Context, evt events.JourneyMilestoneUpdated) error {
	year, quarter := targetPeriodParts(evt.TargetPeriod)
	return p.readModel.UpdateMilestone(ctx, readmodels.JourneyMilestoneUpsertParams{
		JourneyID: evt.ID, MilestoneID: evt.MilestoneID, Label: evt.Label,
		TargetYear: year, TargetQuarter: quarter, Status: evt.Status, UpdatedAt: evt.OccurredOn,
	})
}

func (p *CapabilityJourneyProjector) applyJourneyMilestoneRemoved(ctx context.Context, evt events.JourneyMilestoneRemoved) error {
	return p.readModel.RemoveMilestone(ctx, evt.ID, evt.MilestoneID)
}

func targetPeriodParts(tp *events.TargetPeriodData) (*int, *int) {
	if tp == nil {
		return nil, nil
	}
	year, quarter := tp.Year, tp.Quarter
	return &year, &quarter
}
