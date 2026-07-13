package handlers

import (
	"context"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/shared/cqrs"
)

type journeyMutationHandler[T cqrs.Command] struct {
	repo        CapabilityJourneyRepository
	journeyIDOf func(T) string
	apply       func(T, *aggregates.CapabilityJourney) error
}

func (h *journeyMutationHandler[T]) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(T)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	journey, err := h.repo.GetByID(ctx, h.journeyIDOf(command))
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.apply(command, journey); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, journey); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.EmptyResult(), nil
}

func NewStartJourneyHandler(repo CapabilityJourneyRepository) cqrs.CommandHandler {
	return &journeyMutationHandler[*commands.StartJourney]{
		repo:        repo,
		journeyIDOf: func(c *commands.StartJourney) string { return c.JourneyID },
		apply:       func(c *commands.StartJourney, j *aggregates.CapabilityJourney) error { return j.Start(c.Actor) },
	}
}

func NewCompleteJourneyHandler(repo CapabilityJourneyRepository) cqrs.CommandHandler {
	return &journeyMutationHandler[*commands.CompleteJourney]{
		repo:        repo,
		journeyIDOf: func(c *commands.CompleteJourney) string { return c.JourneyID },
		apply:       func(c *commands.CompleteJourney, j *aggregates.CapabilityJourney) error { return j.Complete(c.Actor) },
	}
}

func NewAbandonJourneyHandler(repo CapabilityJourneyRepository) cqrs.CommandHandler {
	return &journeyMutationHandler[*commands.AbandonJourney]{
		repo:        repo,
		journeyIDOf: func(c *commands.AbandonJourney) string { return c.JourneyID },
		apply:       func(c *commands.AbandonJourney, j *aggregates.CapabilityJourney) error { return j.Abandon(c.Actor) },
	}
}
