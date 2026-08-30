package handlers

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type InvitationStore interface {
	Save(ctx context.Context, invitation *aggregates.Invitation) error
}

type CreateInvitationHandler struct {
	repository InvitationStore
}

func NewCreateInvitationHandler(repository InvitationStore) *CreateInvitationHandler {
	return &CreateInvitationHandler{
		repository: repository,
	}
}

func (h *CreateInvitationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateInvitation)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	id, err := storeInvitation(ctx, h.repository, *command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(id), nil
}

func storeInvitation(ctx context.Context, store InvitationStore, command commands.CreateInvitation) (string, error) {
	email, err := valueobjects.NewEmail(command.Email)
	if err != nil {
		return "", err
	}

	role, err := valueobjects.RoleFromString(command.Role)
	if err != nil {
		return "", err
	}

	inviterInfo, err := parseInviterInfo(command.InviterID, command.InviterEmail)
	if err != nil {
		return "", err
	}

	invitation, err := aggregates.NewInvitation(email, role, inviterInfo)
	if err != nil {
		return "", err
	}

	if err := store.Save(ctx, invitation); err != nil {
		return "", fmt.Errorf("save invitation for %s: %w", command.Email, err)
	}

	return invitation.ID(), nil
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
