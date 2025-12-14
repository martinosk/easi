package handlers

import (
	"context"

	"github.com/google/uuid"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CreateInvitationHandler struct {
	repository *repositories.InvitationRepository
}

func NewCreateInvitationHandler(repository *repositories.InvitationRepository) *CreateInvitationHandler {
	return &CreateInvitationHandler{
		repository: repository,
	}
}

func (h *CreateInvitationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.CreateInvitation)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	email, err := valueobjects.NewEmail(command.Email)
	if err != nil {
		return err
	}

	role, err := valueobjects.RoleFromString(command.Role)
	if err != nil {
		return err
	}

	inviterInfo, err := parseInviterInfo(command.InviterID, command.InviterEmail)
	if err != nil {
		return err
	}

	invitation, err := aggregates.NewInvitation(email, role, inviterInfo)
	if err != nil {
		return err
	}

	command.ID = invitation.ID()

	return h.repository.Save(ctx, invitation)
}

func parseInviterInfo(inviterID, inviterEmail string) (*valueobjects.InviterInfo, error) {
	if inviterID == "" || inviterEmail == "" {
		return nil, nil
	}

	id, err := uuid.Parse(inviterID)
	if err != nil {
		return nil, err
	}

	email, err := valueobjects.NewEmail(inviterEmail)
	if err != nil {
		return nil, err
	}

	info, err := valueobjects.NewInviterInfo(id, email)
	if err != nil {
		return nil, err
	}

	return &info, nil
}
