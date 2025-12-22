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

func (h *ChangeUserRoleHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.ChangeUserRole)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	newRole, err := valueobjects.RoleFromString(command.NewRole)
	if err != nil {
		return err
	}

	isLastAdmin, err := h.userReadModel.IsLastActiveAdmin(ctx, command.UserID)
	if err != nil {
		return err
	}

	user, err := h.repository.GetByID(ctx, command.UserID)
	if err != nil {
		return err
	}

	if err := user.ChangeRole(newRole, command.ChangedByID, isLastAdmin); err != nil {
		return err
	}

	return h.repository.Save(ctx, user)
}
