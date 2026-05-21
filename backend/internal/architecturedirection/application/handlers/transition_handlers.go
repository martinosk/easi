package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
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

func NewUpdateDirectionHandler(repo DirectionLoaderRepository) cqrs.CommandHandler {
	return &mutationHandler[*commands.UpdateDirection]{
		repo:          repo,
		directionIDOf: func(c *commands.UpdateDirection) string { return c.DirectionID },
		apply:         applyUpdateDirection,
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

func applyUpdateDirection(c *commands.UpdateDirection, d *aggregates.Direction) error {
	if err := applyOptional(c.Narrative, sharedvo.NewDescription, d.UpdateNarrative); err != nil {
		return err
	}
	if err := applyOptional(c.Horizon, valueobjects.NewHorizon, d.ChangeHorizon); err != nil {
		return err
	}
	if err := applyOptional(c.SourceCapabilityIDs, buildSourceRefs, d.ChangeSourceCapabilities); err != nil {
		return err
	}
	return applyOptional(c.Placements, buildPlacements, d.ChangePlacements)
}

func applyOptional[Raw any, Parsed any](
	value *Raw,
	parse func(Raw) (Parsed, error),
	apply func(Parsed) error,
) error {
	if value == nil {
		return nil
	}
	parsed, err := parse(*value)
	if err != nil {
		return err
	}
	return apply(parsed)
}
