package handlers

import (
	"context"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type RevokeInvitationHandler struct {
	repository *repositories.InvitationRepository
}

func NewRevokeInvitationHandler(repository *repositories.InvitationRepository) *RevokeInvitationHandler {
	return &RevokeInvitationHandler{
		repository: repository,
	}
}

func (h *RevokeInvitationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.RevokeInvitation)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	invitation, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := invitation.Revoke(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, invitation); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
