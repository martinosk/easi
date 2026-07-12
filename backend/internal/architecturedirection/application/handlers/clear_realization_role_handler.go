package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

var ErrRealizationRoleNotFoundForPair = errors.New("no realization role exists for this capability and component pair")

type ClearRealizationRoleHandler struct {
	repo   RealizationRolesRepository
	lookup ExistingRealizationRolesLookup
}

func NewClearRealizationRoleHandler(repo RealizationRolesRepository, lookup ExistingRealizationRolesLookup) *ClearRealizationRoleHandler {
	return &ClearRealizationRoleHandler{repo: repo, lookup: lookup}
}

func (h *ClearRealizationRoleHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ClearRealizationRole)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	component, err := valueobjects.NewApplicationRef(command.ComponentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	aggregateID, exists, err := h.lookup.FindAggregateIDForCapability(ctx, command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if !exists {
		return cqrs.EmptyResult(), ErrRealizationRoleNotFoundForPair
	}
	rr, err := h.repo.GetByID(ctx, aggregateID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := rr.Clear(component, command.ClearedBy); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, rr); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(rr.ID()), nil
}
