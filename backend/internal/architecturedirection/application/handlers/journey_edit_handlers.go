package handlers

import (
	"context"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

func NewUpdateJourneyProgressHandler(repo CapabilityJourneyRepository) cqrs.CommandHandler {
	return &journeyMutationHandler[*commands.UpdateJourneyProgress]{
		repo:        repo,
		journeyIDOf: func(c *commands.UpdateJourneyProgress) string { return c.JourneyID },
		apply:       applyUpdateJourneyProgress,
	}
}

func applyUpdateJourneyProgress(c *commands.UpdateJourneyProgress, j *aggregates.CapabilityJourney) error {
	progress, err := valueobjects.NewJourneyProgress(c.Progress)
	if err != nil {
		return err
	}
	return j.UpdateProgress(progress, c.Actor)
}

func NewUpdateJourneyDetailsHandler(repo CapabilityJourneyRepository) cqrs.CommandHandler {
	return &journeyMutationHandler[*commands.UpdateJourneyDetails]{
		repo:        repo,
		journeyIDOf: func(c *commands.UpdateJourneyDetails) string { return c.JourneyID },
		apply:       applyUpdateJourneyDetails,
	}
}

func applyUpdateJourneyDetails(c *commands.UpdateJourneyDetails, j *aggregates.CapabilityJourney) error {
	note, err := sharedvo.NewDescription(c.Note)
	if err != nil {
		return err
	}
	targetPeriod, err := buildTargetPeriod(c.TargetYear, c.TargetQuarter)
	if err != nil {
		return err
	}
	return j.UpdateDetails(aggregates.JourneyDetailsFacts{
		Note:          note,
		TargetPeriod:  targetPeriod,
		ResultingName: c.ResultingName,
		Actor:         c.Actor,
	})
}

type ChangeJourneySourceApplicationsHandler struct {
	repo            CapabilityJourneyRepository
	componentExists services.ComponentExists
}

func NewChangeJourneySourceApplicationsHandler(repo CapabilityJourneyRepository, componentExists services.ComponentExists) *ChangeJourneySourceApplicationsHandler {
	return &ChangeJourneySourceApplicationsHandler{repo: repo, componentExists: componentExists}
}

func (h *ChangeJourneySourceApplicationsHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ChangeJourneySourceApplications)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	fromApps, err := parseApplicationRefs(command.FromComponentIDs)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := verifyComponentsExist(ctx, h.componentExists, fromApps); err != nil {
		return cqrs.EmptyResult(), err
	}
	journey, err := h.repo.GetByID(ctx, command.JourneyID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := journey.ChangeSourceApplications(fromApps, command.Actor); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, journey); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.EmptyResult(), nil
}
