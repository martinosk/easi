package handlers

import (
	"context"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type DirectionRepository interface {
	Save(ctx context.Context, d *aggregates.Direction) error
}

type CaptureDirectionHandler struct {
	repo   DirectionRepository
	policy *services.DirectionReferenceService
}

func NewCaptureDirectionHandler(repo DirectionRepository, policy *services.DirectionReferenceService) *CaptureDirectionHandler {
	return &CaptureDirectionHandler{repo: repo, policy: policy}
}

func (h *CaptureDirectionHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CaptureDirection)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	params, err := captureParamsFromCommand(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.policy.VerifyCanCapture(ctx, params); err != nil {
		return cqrs.EmptyResult(), err
	}
	direction, err := aggregates.DraftDirection(params)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, direction); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(direction.ID()), nil
}

func captureParamsFromCommand(command *commands.CaptureDirection) (aggregates.DraftParams, error) {
	ecRef, err := valueobjects.NewEnterpriseCapabilityRef(command.EnterpriseCapabilityID)
	if err != nil {
		return aggregates.DraftParams{}, err
	}
	dt, err := valueobjects.NewDirectionType(command.Type)
	if err != nil {
		return aggregates.DraftParams{}, err
	}
	horizon, err := valueobjects.NewHorizon(command.Horizon)
	if err != nil {
		return aggregates.DraftParams{}, err
	}
	narrative, err := valueobjects.NewNarrative(command.Narrative)
	if err != nil {
		return aggregates.DraftParams{}, err
	}
	sourceRefs, err := buildSourceRefs(command.SourceCapabilityIDs)
	if err != nil {
		return aggregates.DraftParams{}, err
	}
	placements, err := buildPlacements(command.Placements)
	if err != nil {
		return aggregates.DraftParams{}, err
	}
	return aggregates.DraftParams{
		EnterpriseCapabilityID: ecRef,
		Type:                   dt,
		SourceCapabilityIDs:    sourceRefs,
		Placements:             placements,
		Horizon:                horizon,
		Narrative:              narrative,
	}, nil
}

func buildSourceRefs(ids []string) ([]valueobjects.PhysicalCapabilityRef, error) {
	refs := make([]valueobjects.PhysicalCapabilityRef, 0, len(ids))
	for _, id := range ids {
		ref, err := valueobjects.NewPhysicalCapabilityRef(id)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

func buildPlacements(inputs []commands.PlacementInput) ([]valueobjects.Placement, error) {
	placements := make([]valueobjects.Placement, 0, len(inputs))
	for _, in := range inputs {
		p, err := valueobjects.NewPlacement(in.TargetBusinessDomainID, in.ResultingName)
		if err != nil {
			return nil, err
		}
		placements = append(placements, p)
	}
	return placements, nil
}
