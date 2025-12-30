package handlers

import (
	"context"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type EnableUserHandler struct {
	repository *repositories.UserAggregateRepository
}

func NewEnableUserHandler(repository *repositories.UserAggregateRepository) *EnableUserHandler {
	return &EnableUserHandler{
		repository: repository,
	}
}

func (h *EnableUserHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.EnableUser)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	user, err := h.repository.GetByID(ctx, command.UserID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := user.Enable(command.EnabledByID); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, user); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
