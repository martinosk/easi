package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

var ErrUnknownAdvanceTarget = errors.New("advance target must be 'proposed' or 'agreed'")

type DirectionLoaderRepository interface {
	DirectionRepository
	GetByID(ctx context.Context, id string) (*aggregates.Direction, error)
}

type mutationHandler[T cqrs.Command] struct {
	repo          DirectionLoaderRepository
	directionIDOf func(T) string
	apply         func(T, *aggregates.Direction) error
}

func (h *mutationHandler[T]) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(T)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	direction, err := h.repo.GetByID(ctx, h.directionIDOf(command))
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.apply(command, direction); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, direction); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.EmptyResult(), nil
}

func NewAdvanceDirectionHandler(repo DirectionLoaderRepository) cqrs.CommandHandler {
	return &mutationHandler[*commands.AdvanceDirection]{
		repo:          repo,
		directionIDOf: func(c *commands.AdvanceDirection) string { return c.DirectionID },
		apply:         applyAdvance,
	}
}

func NewRejectDirectionHandler(repo DirectionLoaderRepository) cqrs.CommandHandler {
	return &mutationHandler[*commands.RejectDirection]{
		repo:          repo,
		directionIDOf: func(c *commands.RejectDirection) string { return c.DirectionID },
		apply:         func(_ *commands.RejectDirection, d *aggregates.Direction) error { return d.Reject() },
	}
}

func NewUpdateDirectionNarrativeHandler(repo DirectionLoaderRepository) cqrs.CommandHandler {
	return &mutationHandler[*commands.UpdateDirectionNarrative]{
		repo:          repo,
		directionIDOf: func(c *commands.UpdateDirectionNarrative) string { return c.DirectionID },
		apply:         applyNarrative,
	}
}

func NewUpdateDirectionHorizonHandler(repo DirectionLoaderRepository) cqrs.CommandHandler {
	return &mutationHandler[*commands.UpdateDirectionHorizon]{
		repo:          repo,
		directionIDOf: func(c *commands.UpdateDirectionHorizon) string { return c.DirectionID },
		apply:         applyHorizon,
	}
}

func NewUpdateDirectionSourceCapabilitiesHandler(repo DirectionLoaderRepository) cqrs.CommandHandler {
	return &mutationHandler[*commands.UpdateDirectionSourceCapabilities]{
		repo:          repo,
		directionIDOf: func(c *commands.UpdateDirectionSourceCapabilities) string { return c.DirectionID },
		apply:         applySources,
	}
}

func NewUpdateDirectionPlacementsHandler(repo DirectionLoaderRepository) cqrs.CommandHandler {
	return &mutationHandler[*commands.UpdateDirectionPlacements]{
		repo:          repo,
		directionIDOf: func(c *commands.UpdateDirectionPlacements) string { return c.DirectionID },
		apply:         applyPlacements,
	}
}

func applyAdvance(c *commands.AdvanceDirection, d *aggregates.Direction) error {
	switch c.TargetStatus {
	case valueobjects.DirectionStatusProposed:
		return d.Propose()
	case valueobjects.DirectionStatusAgreed:
		return d.Agree()
	default:
		return ErrUnknownAdvanceTarget
	}
}

func applyNarrative(c *commands.UpdateDirectionNarrative, d *aggregates.Direction) error {
	narrative, err := valueobjects.NewNarrative(c.Narrative)
	if err != nil {
		return err
	}
	return d.UpdateNarrative(narrative)
}

func applyHorizon(c *commands.UpdateDirectionHorizon, d *aggregates.Direction) error {
	horizon, err := valueobjects.NewHorizon(c.Horizon)
	if err != nil {
		return err
	}
	return d.ChangeHorizon(horizon)
}

func applySources(c *commands.UpdateDirectionSourceCapabilities, d *aggregates.Direction) error {
	refs, err := buildSourceRefs(c.SourceCapabilityIDs)
	if err != nil {
		return err
	}
	return d.ChangeSourceCapabilities(refs)
}

func applyPlacements(c *commands.UpdateDirectionPlacements, d *aggregates.Direction) error {
	placements, err := buildPlacements(c.Placements)
	if err != nil {
		return err
	}
	return d.ChangePlacements(placements)
}
