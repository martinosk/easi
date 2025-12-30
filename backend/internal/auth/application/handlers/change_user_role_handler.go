package handlers

import (
	"context"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type ChangeUserRoleHandler struct {
	repository    *repositories.UserAggregateRepository
	userReadModel *readmodels.UserReadModel
}

func NewChangeUserRoleHandler(
	repository *repositories.UserAggregateRepository,
	userReadModel *readmodels.UserReadModel,
) *ChangeUserRoleHandler {
	return &ChangeUserRoleHandler{
		repository:    repository,
		userReadModel: userReadModel,
	}
}

func (h *ChangeUserRoleHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ChangeUserRole)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	newRole, err := valueobjects.RoleFromString(command.NewRole)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	isLastAdmin, err := h.userReadModel.IsLastActiveAdmin(ctx, command.UserID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	user, err := h.repository.GetByID(ctx, command.UserID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := user.ChangeRole(newRole, command.ChangedByID, isLastAdmin); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, user); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
