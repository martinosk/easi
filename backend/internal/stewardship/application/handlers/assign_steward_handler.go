package handlers

import (
	"context"
	"fmt"

	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/stewardship/application/commands"
	"easi/backend/internal/stewardship/domain/aggregates"
	"easi/backend/internal/stewardship/domain/valueobjects"
)

type AssignStewardDeps struct {
	Repository StewardshipRepository
	Lookup     LiveStewardshipLookup
	Domains    DomainDirectory
	Users      UserDirectory
}

type AssignStewardHandler struct {
	deps AssignStewardDeps
	live liveStewardships
}

func NewAssignStewardHandler(deps AssignStewardDeps) *AssignStewardHandler {
	return &AssignStewardHandler{deps: deps, live: liveStewardships{lookup: deps.Lookup, repository: deps.Repository}}
}

func (h *AssignStewardHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AssignSteward)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	concern, err := concernOfExistingDomain(ctx, h.deps.Domains, command.DomainID, command.Concern)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	steward, err := h.stewardRef(ctx, command.StewardID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	stewardship, err := h.assign(ctx, command, concern, steward)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.deps.Repository.Save(ctx, stewardship); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(stewardship.ID()), nil
}

func (h *AssignStewardHandler) stewardRef(ctx context.Context, userID string) (valueobjects.StewardRef, error) {
	standing, err := h.deps.Users.UserStanding(ctx, userID)
	if err != nil {
		return valueobjects.StewardRef{}, fmt.Errorf("check standing of user %s: %w", userID, err)
	}
	return valueobjects.NewStewardRef(userID, standing)
}

func (h *AssignStewardHandler) assign(ctx context.Context, command *commands.AssignSteward, concern valueobjects.Concern, steward valueobjects.StewardRef) (*aggregates.Stewardship, error) {
	existing, err := h.live.find(ctx, command.DomainID, concern)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return aggregates.NewStewardship(aggregates.Assignment{
			DomainID:   command.DomainID,
			Concern:    concern,
			Steward:    steward,
			AssignedBy: command.AssignedBy,
		})
	}
	if err := existing.Assign(steward, command.AssignedBy); err != nil {
		return nil, err
	}
	return existing, nil
}
