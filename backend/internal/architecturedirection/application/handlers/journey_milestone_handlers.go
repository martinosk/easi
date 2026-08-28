package handlers

import (
	"context"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/google/uuid"
)

type AddJourneyMilestoneHandler struct {
	repo CapabilityJourneyRepository
}

func NewAddJourneyMilestoneHandler(repo CapabilityJourneyRepository) *AddJourneyMilestoneHandler {
	return &AddJourneyMilestoneHandler{repo: repo}
}

func (h *AddJourneyMilestoneHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AddJourneyMilestone)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	facts, err := parseMilestoneFacts(milestoneCommandFields{
		milestoneID:   uuid.New().String(),
		label:         command.Label,
		targetYear:    command.TargetYear,
		targetQuarter: command.TargetQuarter,
		status:        command.Status,
		actor:         command.Actor,
	})
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	journey, err := h.repo.GetByID(ctx, command.JourneyID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := journey.AddMilestone(facts); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, journey); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.EmptyResult(), nil
}

func NewUpdateJourneyMilestoneHandler(repo CapabilityJourneyRepository) cqrs.CommandHandler {
	return &journeyMutationHandler[*commands.UpdateJourneyMilestone]{
		repo:        repo,
		journeyIDOf: func(c *commands.UpdateJourneyMilestone) string { return c.JourneyID },
		apply:       applyUpdateJourneyMilestone,
	}
}

func applyUpdateJourneyMilestone(c *commands.UpdateJourneyMilestone, j *aggregates.CapabilityJourney) error {
	facts, err := parseMilestoneFacts(milestoneCommandFields{
		milestoneID:   c.MilestoneID,
		label:         c.Label,
		targetYear:    c.TargetYear,
		targetQuarter: c.TargetQuarter,
		status:        c.Status,
		actor:         c.Actor,
	})
	if err != nil {
		return err
	}
	return j.UpdateMilestone(facts)
}

func NewRemoveJourneyMilestoneHandler(repo CapabilityJourneyRepository) cqrs.CommandHandler {
	return &journeyMutationHandler[*commands.RemoveJourneyMilestone]{
		repo:        repo,
		journeyIDOf: func(c *commands.RemoveJourneyMilestone) string { return c.JourneyID },
		apply: func(c *commands.RemoveJourneyMilestone, j *aggregates.CapabilityJourney) error {
			return j.RemoveMilestone(c.MilestoneID, c.Actor)
		},
	}
}

func NewReorderJourneyMilestonesHandler(repo CapabilityJourneyRepository) cqrs.CommandHandler {
	return &journeyMutationHandler[*commands.ReorderJourneyMilestones]{
		repo:        repo,
		journeyIDOf: func(c *commands.ReorderJourneyMilestones) string { return c.JourneyID },
		apply: func(c *commands.ReorderJourneyMilestones, j *aggregates.CapabilityJourney) error {
			return j.ReorderMilestones(c.MilestoneIDs, c.Actor)
		},
	}
}

type milestoneCommandFields struct {
	milestoneID   string
	label         string
	targetYear    *int
	targetQuarter *int
	status        string
	actor         string
}

func parseMilestoneFacts(f milestoneCommandFields) (aggregates.MilestoneFacts, error) {
	status, err := valueobjects.NewMilestoneStatus(f.status)
	if err != nil {
		return aggregates.MilestoneFacts{}, err
	}
	targetPeriod, err := buildTargetPeriod(f.targetYear, f.targetQuarter)
	if err != nil {
		return aggregates.MilestoneFacts{}, err
	}
	return aggregates.MilestoneFacts{
		MilestoneID:  f.milestoneID,
		Label:        f.label,
		TargetPeriod: targetPeriod,
		Status:       status,
		Actor:        f.actor,
	}, nil
}
