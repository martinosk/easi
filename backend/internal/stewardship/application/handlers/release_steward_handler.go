package handlers

import (
	"context"

	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/stewardship/application/commands"
)

type ReleaseStewardDeps struct {
	Repository StewardshipRepository
	Lookup     LiveStewardshipLookup
	Domains    DomainDirectory
}

type ReleaseStewardHandler struct {
	deps ReleaseStewardDeps
	live liveStewardships
}

func NewReleaseStewardHandler(deps ReleaseStewardDeps) *ReleaseStewardHandler {
	return &ReleaseStewardHandler{deps: deps, live: liveStewardships{lookup: deps.Lookup, repository: deps.Repository}}
}

func (h *ReleaseStewardHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ReleaseSteward)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	concern, err := concernOfExistingDomain(ctx, h.deps.Domains, command.DomainID, command.Concern)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	stewardship, err := h.live.find(ctx, command.DomainID, concern)
	if err != nil || stewardship == nil {
		return cqrs.EmptyResult(), err
	}
	if err := stewardship.Release(command.ReleasedBy); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.deps.Repository.Save(ctx, stewardship); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(stewardship.ID()), nil
}
