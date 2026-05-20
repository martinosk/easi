package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
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
	UpdateField(ctx context.Context, u readmodels.FieldUpdate) error
	UpdatePlacements(ctx context.Context, u readmodels.PlacementsUpdate) error
	ReplaceSourceCapabilities(ctx context.Context, u readmodels.SourceCapabilitiesUpdate) error
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
		pl.DirectionProposed:                  p.handleStatusEvent(valueobjects.DirectionStatusProposed),
		pl.DirectionAgreed:                    p.handleStatusEvent(valueobjects.DirectionStatusAgreed),
		pl.DirectionRejected:                  p.handleStatusEvent(valueobjects.DirectionStatusRejected),
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
		err := p.readModel.Insert(ctx, draftedToInsertParams(e))
		if errors.Is(err, services.ErrActiveDirectionAlreadyExists) {
			log.Printf("architecturedirection: orphan DirectionDrafted event %s for EC %s lost the race against an existing active direction; reject it before retrying capture",
				e.ID, e.EnterpriseCapabilityID)
		}
		return err
	})
}

func draftedToInsertParams(e events.DirectionDrafted) readmodels.InsertDirectionParams {
	return readmodels.InsertDirectionParams{
		ID:                     readmodels.DirectionID(e.ID),
		EnterpriseCapabilityID: e.EnterpriseCapabilityID,
		Type:                   e.Type,
		Status:                 valueobjects.DirectionStatusDraft,
		Horizon:                e.Horizon,
		Narrative:              e.Narrative,
		SourceCapabilityIDs:    toCapabilityIDs(e.SourceCapabilityIDs),
		Placements:             toPlacementDTOs(e.Placements),
		CreatedAt:              e.CreatedAt,
	}
}

func (p *DirectionProjector) handleStatusEvent(status string) func(context.Context, []byte) error {
	return func(ctx context.Context, eventData []byte) error {
		var generic struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(eventData, &generic); err != nil {
			return err
		}
		return p.readModel.UpdateField(ctx, readmodels.FieldUpdate{
			DirectionID: readmodels.DirectionID(generic.ID),
			Field:       readmodels.DirectionFieldStatus,
			Value:       status,
		})
	}
}

func (p *DirectionProjector) handleNarrativeUpdated(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, e events.DirectionNarrativeUpdated) error {
		return p.readModel.UpdateField(ctx, readmodels.FieldUpdate{
			DirectionID: readmodels.DirectionID(e.ID),
			Field:       readmodels.DirectionFieldNarrative,
			Value:       e.Narrative,
		})
	})
}

func (p *DirectionProjector) handleHorizonChanged(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, e events.DirectionHorizonChanged) error {
		return p.readModel.UpdateField(ctx, readmodels.FieldUpdate{
			DirectionID: readmodels.DirectionID(e.ID),
			Field:       readmodels.DirectionFieldHorizon,
			Value:       e.Horizon,
		})
	})
}

func (p *DirectionProjector) handlePlacementsChanged(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, e events.DirectionPlacementsChanged) error {
		return p.readModel.UpdatePlacements(ctx, readmodels.PlacementsUpdate{
			DirectionID: readmodels.DirectionID(e.ID),
			Placements:  toPlacementDTOs(e.Placements),
		})
	})
}

func (p *DirectionProjector) handleSourcesChanged(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, e events.DirectionSourceCapabilitiesChanged) error {
		return p.readModel.ReplaceSourceCapabilities(ctx, readmodels.SourceCapabilitiesUpdate{
			DirectionID:         readmodels.DirectionID(e.ID),
			SourceCapabilityIDs: toCapabilityIDs(e.SourceCapabilityIDs),
		})
	})
}

func toCapabilityIDs(ids []string) []readmodels.CapabilityID {
	out := make([]readmodels.CapabilityID, len(ids))
	for i, id := range ids {
		out[i] = readmodels.CapabilityID(id)
	}
	return out
}

func toPlacementDTOs(in []events.PlacementData) []readmodels.DirectionPlacementDTO {
	out := make([]readmodels.DirectionPlacementDTO, len(in))
	for i, p := range in {
		out[i] = readmodels.DirectionPlacementDTO{
			TargetBusinessDomainID: p.TargetBusinessDomainID,
			ResultingName:          p.ResultingName,
		}
	}
	return out
}
