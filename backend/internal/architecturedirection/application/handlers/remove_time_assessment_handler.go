package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/shared/cqrs"
)

var ErrTimeAssessmentNotFoundForPair = errors.New("no time assessment exists for this capability and component pair")

type RemoveTimeAssessmentHandler struct {
	repo   TimeAssessmentRepository
	lookup ExistingTimeAssessmentLookup
}

func NewRemoveTimeAssessmentHandler(repo TimeAssessmentRepository, lookup ExistingTimeAssessmentLookup) *RemoveTimeAssessmentHandler {
	return &RemoveTimeAssessmentHandler{repo: repo, lookup: lookup}
}

func (h *RemoveTimeAssessmentHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.RemoveTimeAssessment)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	aggregateID, exists, err := h.lookup.FindAggregateIDForPair(ctx, command.CapabilityID, command.ComponentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if !exists {
		return cqrs.EmptyResult(), ErrTimeAssessmentNotFoundForPair
	}
	ta, err := h.repo.GetByID(ctx, aggregateID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := ta.Remove(command.RemovedBy); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, ta); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(ta.ID()), nil
}
