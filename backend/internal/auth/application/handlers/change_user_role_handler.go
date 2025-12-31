package handlers

import (
	"context"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type ChangeUserRoleRepository interface {
	Save(ctx context.Context, user *aggregates.User) error
	GetByID(ctx context.Context, id string) (*aggregates.User, error)
}

type ChangeUserRoleReadModel interface {
	IsLastActiveAdmin(ctx context.Context, userID string) (bool, error)
}

type ChangeUserRoleHandler struct {
	repository    ChangeUserRoleRepository
	userReadModel ChangeUserRoleReadModel
}

func NewChangeUserRoleHandler(
	repository ChangeUserRoleRepository,
	userReadModel ChangeUserRoleReadModel,
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
