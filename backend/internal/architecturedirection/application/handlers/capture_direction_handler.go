package handlers

import (
	"context"
	"errors"
	"fmt"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

var ErrActiveDirectionAlreadyExists = readmodels.ErrActiveDirectionAlreadyExists

var ErrReferencedEntityNotFound = errors.New("a referenced entity does not exist or is not accessible in this tenant")

type ReferenceChecker interface {
	EnterpriseCapabilityExists(ctx context.Context, id string) (bool, error)
	PhysicalCapabilityExists(ctx context.Context, id string) (bool, error)
	BusinessDomainExists(ctx context.Context, id string) (bool, error)
}

type DirectionRepository interface {
	Save(ctx context.Context, d *aggregates.Direction) error
}

type ActiveDirectionLookup interface {
	HasActiveDirectionForEnterpriseCapability(ctx context.Context, enterpriseCapabilityID string) (bool, error)
}

type CaptureDirectionHandler struct {
	repo       DirectionRepository
	lookup     ActiveDirectionLookup
	references ReferenceChecker
}

func NewCaptureDirectionHandler(repo DirectionRepository, lookup ActiveDirectionLookup, references ReferenceChecker) *CaptureDirectionHandler {
	return &CaptureDirectionHandler{repo: repo, lookup: lookup, references: references}
}

func (h *CaptureDirectionHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CaptureDirection)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	if err := h.verifyReferences(ctx, command); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.ensureNoActiveDirection(ctx, command.EnterpriseCapabilityID); err != nil {
		return cqrs.EmptyResult(), err
	}
	params, err := captureParamsFromCommand(command)
	if err != nil {
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

func (h *CaptureDirectionHandler) verifyReferences(ctx context.Context, cmd *commands.CaptureDirection) error {
	if h.references == nil {
		return nil
	}
	if err := requireExists(ctx, h.references.EnterpriseCapabilityExists, cmd.EnterpriseCapabilityID, "enterprise capability"); err != nil {
		return err
	}
	if err := verifyAll(ctx, h.references.PhysicalCapabilityExists, cmd.SourceCapabilityIDs, "source capability"); err != nil {
		return err
	}
	return verifyAll(ctx, h.references.BusinessDomainExists, placementDomainIDs(cmd.Placements), "target business domain")
}

func placementDomainIDs(placements []commands.PlacementInput) []string {
	ids := make([]string, len(placements))
	for i, p := range placements {
		ids[i] = p.TargetBusinessDomainID
	}
	return ids
}

func verifyAll(ctx context.Context, check func(context.Context, string) (bool, error), ids []string, label string) error {
	for _, id := range ids {
		if err := requireExists(ctx, check, id, label); err != nil {
			return err
		}
	}
	return nil
}

func requireExists(ctx context.Context, check func(context.Context, string) (bool, error), id, label string) error {
	exists, err := check(ctx, id)
	if err != nil {
		return fmt.Errorf("verify %s %s: %w", label, id, err)
	}
	if !exists {
		return fmt.Errorf("%w: %s %s", ErrReferencedEntityNotFound, label, id)
	}
	return nil
}

func (h *CaptureDirectionHandler) ensureNoActiveDirection(ctx context.Context, enterpriseCapabilityID string) error {
	hasActive, err := h.lookup.HasActiveDirectionForEnterpriseCapability(ctx, enterpriseCapabilityID)
	if err != nil {
		return err
	}
	if hasActive {
		return ErrActiveDirectionAlreadyExists
	}
	return nil
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
