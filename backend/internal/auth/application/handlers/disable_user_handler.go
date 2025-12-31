package handlers

import (
	"context"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/shared/cqrs"
)

type DisableUserRepository interface {
	Save(ctx context.Context, user *aggregates.User) error
	GetByID(ctx context.Context, id string) (*aggregates.User, error)
}

type DisableUserReadModel interface {
	IsLastActiveAdmin(ctx context.Context, userID string) (bool, error)
}

type DisableUserHandler struct {
	repository    DisableUserRepository
	userReadModel DisableUserReadModel
}

func NewDisableUserHandler(
	repository DisableUserRepository,
	userReadModel DisableUserReadModel,
) *DisableUserHandler {
	return &DisableUserHandler{
		repository:    repository,
		userReadModel: userReadModel,
	}
}

func (h *DisableUserHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DisableUser)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	isCurrentUser := command.UserID == command.DisabledByID

	isLastAdmin, err := h.userReadModel.IsLastActiveAdmin(ctx, command.UserID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	user, err := h.repository.GetByID(ctx, command.UserID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := user.Disable(command.DisabledByID, isCurrentUser, isLastAdmin); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, user); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
