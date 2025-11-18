package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DeleteCapabilityDependencyHandler struct {
	repository *repositories.DependencyRepository
}

func NewDeleteCapabilityDependencyHandler(repository *repositories.DependencyRepository) *DeleteCapabilityDependencyHandler {
	return &DeleteCapabilityDependencyHandler{
		repository: repository,
	}
}

func (h *DeleteCapabilityDependencyHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.DeleteCapabilityDependency)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	dependency, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	if err := dependency.Delete(); err != nil {
		return err
	}

	return h.repository.Save(ctx, dependency)
}
