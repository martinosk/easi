package handlers

import (
	"context"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type MarkInvitationExpiredHandler struct {
	repository *repositories.InvitationRepository
}

func NewMarkInvitationExpiredHandler(repository *repositories.InvitationRepository) *MarkInvitationExpiredHandler {
	return &MarkInvitationExpiredHandler{
		repository: repository,
	}
}

func (h *MarkInvitationExpiredHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.MarkInvitationExpired)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	invitation, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := invitation.MarkExpired(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, invitation); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
