package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var ErrNoPendingInvitation = errors.New("no pending invitation found for this email")

type AcceptInvitationHandler struct {
	repository *repositories.InvitationRepository
	readModel  *readmodels.InvitationReadModel
}

func NewAcceptInvitationHandler(
	repository *repositories.InvitationRepository,
	readModel *readmodels.InvitationReadModel,
) *AcceptInvitationHandler {
	return &AcceptInvitationHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *AcceptInvitationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.AcceptInvitation)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	pendingInvitation, err := h.readModel.GetPendingByEmail(ctx, command.Email)
	if err != nil {
		return err
	}
	if pendingInvitation == nil {
		return ErrNoPendingInvitation
	}

	invitation, err := h.repository.GetByID(ctx, pendingInvitation.ID)
	if err != nil {
		return err
	}

	if err := invitation.Accept(); err != nil {
		return err
	}

	return h.repository.Save(ctx, invitation)
}
