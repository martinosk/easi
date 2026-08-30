package handlers

import (
	"context"
	"errors"
	"fmt"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/shared/cqrs"
)

var ErrEmailDomainNotAllowed = errors.New("email domain is not registered to this tenant")

type UserDirectory interface {
	GetByEmail(ctx context.Context, email string) (*readmodels.UserDTO, error)
}

type PendingInvitations interface {
	ExistsPendingForEmail(ctx context.Context, email string) (bool, error)
}

type EmailDomainPolicy interface {
	IsDomainAllowed(ctx context.Context, email string) (bool, error)
}

type EnsureInvitationHandler struct {
	repository  InvitationStore
	users       UserDirectory
	invitations PendingInvitations
	domains     EmailDomainPolicy
}

func NewEnsureInvitationHandler(
	repository InvitationStore,
	users UserDirectory,
	invitations PendingInvitations,
	domains EmailDomainPolicy,
) *EnsureInvitationHandler {
	return &EnsureInvitationHandler{
		repository:  repository,
		users:       users,
		invitations: invitations,
		domains:     domains,
	}
}

func (h *EnsureInvitationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.EnsureInvitation)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	needed, err := h.isInvitationNeeded(ctx, command.Email)
	if err != nil || !needed {
		return cqrs.EmptyResult(), err
	}

	id, err := storeInvitation(ctx, h.repository, commands.CreateInvitation{
		Email:        command.Email,
		Role:         command.Role,
		InviterID:    command.InviterID,
		InviterEmail: command.InviterEmail,
	})
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(id), nil
}

func (h *EnsureInvitationHandler) isInvitationNeeded(ctx context.Context, email string) (bool, error) {
	user, err := h.users.GetByEmail(ctx, email)
	if err != nil {
		return false, fmt.Errorf("look up user %s: %w", email, err)
	}
	if user != nil {
		return false, nil
	}

	invited, err := h.invitations.ExistsPendingForEmail(ctx, email)
	if err != nil {
		return false, fmt.Errorf("look up pending invitation for %s: %w", email, err)
	}
	if invited {
		return false, nil
	}

	allowed, err := h.domains.IsDomainAllowed(ctx, email)
	if err != nil {
		return false, fmt.Errorf("check email domain of %s: %w", email, err)
	}
	if !allowed {
		return false, ErrEmailDomainNotAllowed
	}

	return true, nil
}
