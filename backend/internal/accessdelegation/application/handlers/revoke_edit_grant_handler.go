package handlers

import (
	"context"

	"easi/backend/internal/accessdelegation/application/commands"
	"easi/backend/internal/accessdelegation/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type RevokeEditGrantHandler struct {
	repository *repositories.EditGrantRepository
}

func NewRevokeEditGrantHandler(repository *repositories.EditGrantRepository) *RevokeEditGrantHandler {
	return &RevokeEditGrantHandler{repository: repository}
}

func (h *RevokeEditGrantHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.RevokeEditGrant)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	grant, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := grant.Revoke(command.RevokedBy); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, grant); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
