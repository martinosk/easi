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

func (h *RevokeInvitationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.RevokeInvitation)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	invitation, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	if err := invitation.Revoke(); err != nil {
		return err
	}

	return h.repository.Save(ctx, invitation)
}
