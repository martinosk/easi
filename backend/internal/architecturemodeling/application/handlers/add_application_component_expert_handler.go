package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/entities"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type AddApplicationComponentExpertHandler struct {
	repository *repositories.ApplicationComponentRepository
}

func NewAddApplicationComponentExpertHandler(repository *repositories.ApplicationComponentRepository) *AddApplicationComponentExpertHandler {
	return &AddApplicationComponentExpertHandler{
		repository: repository,
	}
}

func (h *AddApplicationComponentExpertHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AddApplicationComponentExpert)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	component, err := h.repository.GetByID(ctx, command.ComponentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	expert, err := entities.NewExpert(command.ExpertName, command.ExpertRole, command.ContactInfo)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := component.AddExpert(expert); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, component); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
