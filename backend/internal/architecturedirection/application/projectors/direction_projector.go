package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/events"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

func handleProjection[T any](ctx context.Context, eventData []byte, fn func(context.Context, T) error) error {
	var event T
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal %T event data: %w", event, err)
	}
	return fn(ctx, event)
}

type DirectionStore interface {
	Insert(ctx context.Context, p readmodels.InsertDirectionParams) error
	UpdateStatus(ctx context.Context, id, status string) error
	UpdateNarrative(ctx context.Context, id, narrative string) error
	UpdateHorizon(ctx context.Context, id, horizon string) error
	UpdatePlacements(ctx context.Context, id string, placements []readmodels.DirectionPlacementDTO) error
	ReplaceSourceCapabilities(ctx context.Context, id string, sourceCapabilityIDs []string) error
}

type DirectionProjector struct {
	readModel DirectionStore
}

func NewDirectionProjector(readModel DirectionStore) *DirectionProjector {
	return &DirectionProjector{readModel: readModel}
}

func (p *DirectionProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *DirectionProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		pl.DirectionDrafted:                   p.handleDrafted,
		pl.DirectionProposed:                  p.handleStatusEvent("proposed"),
		pl.DirectionAgreed:                    p.handleStatusEvent("agreed"),
		pl.DirectionRejected:                  p.handleStatusEvent("rejected"),
		pl.DirectionNarrativeUpdated:          p.handleNarrativeUpdated,
		pl.DirectionHorizonChanged:            p.handleHorizonChanged,
		pl.DirectionPlacementsChanged:         p.handlePlacementsChanged,
		pl.DirectionSourceCapabilitiesChanged: p.handleSourcesChanged,
	}
	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *DirectionProjector) handleDrafted(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, e events.DirectionDrafted) error {
		placements := make([]readmodels.DirectionPlacementDTO, len(e.Placements))
		for i, pl := range e.Placements {
			placements[i] = readmodels.DirectionPlacementDTO{
				TargetBusinessDomainID: pl.TargetBusinessDomainID,
				ResultingName:          pl.ResultingName,
			}
		}
		err := p.readModel.Insert(ctx, readmodels.InsertDirectionParams{
			ID:                     e.ID,
			EnterpriseCapabilityID: e.EnterpriseCapabilityID,
			Type:                   e.Type,
			Status:                 "draft",
			Horizon:                e.Horizon,
			Narrative:              e.Narrative,
			SourceCapabilityIDs:    e.SourceCapabilityIDs,
			Placements:             placements,
			CreatedAt:              e.CreatedAt,
		})
		if errors.Is(err, readmodels.ErrActiveDirectionAlreadyExists) {
			log.Printf("architecturedirection: orphan DirectionDrafted event %s for EC %s lost the race against an existing active direction; reject it before retrying capture",
				e.ID, e.EnterpriseCapabilityID)
		}
		return err
	})
}

func (p *DirectionProjector) handleStatusEvent(status string) func(context.Context, []byte) error {
	return func(ctx context.Context, eventData []byte) error {
		var generic struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(eventData, &generic); err != nil {
			return err
		}
		return p.readModel.UpdateStatus(ctx, generic.ID, status)
	}
}

func (p *DirectionProjector) handleNarrativeUpdated(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, e events.DirectionNarrativeUpdated) error {
		return p.readModel.UpdateNarrative(ctx, e.ID, e.Narrative)
	})
}

func (p *DirectionProjector) handleHorizonChanged(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, e events.DirectionHorizonChanged) error {
		return p.readModel.UpdateHorizon(ctx, e.ID, e.Horizon)
	})
}

func (p *DirectionProjector) handlePlacementsChanged(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, e events.DirectionPlacementsChanged) error {
		placements := make([]readmodels.DirectionPlacementDTO, len(e.Placements))
		for i, pl := range e.Placements {
			placements[i] = readmodels.DirectionPlacementDTO{
				TargetBusinessDomainID: pl.TargetBusinessDomainID,
				ResultingName:          pl.ResultingName,
			}
		}
		return p.readModel.UpdatePlacements(ctx, e.ID, placements)
	})
}

func (p *DirectionProjector) handleSourcesChanged(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, e events.DirectionSourceCapabilitiesChanged) error {
		return p.readModel.ReplaceSourceCapabilities(ctx, e.ID, e.SourceCapabilityIDs)
	})
}
