package handlers

import (
	"context"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DisableUserHandler struct {
	repository    *repositories.UserAggregateRepository
	userReadModel *readmodels.UserReadModel
}

func NewDisableUserHandler(
	repository *repositories.UserAggregateRepository,
	userReadModel *readmodels.UserReadModel,
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
