package handlers

import (
	"context"
	"time"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type AddApplicationComponentExpertRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ApplicationComponent, error)
	Save(ctx context.Context, component *aggregates.ApplicationComponent) error
}

type AddApplicationComponentExpertHandler struct {
	repository AddApplicationComponentExpertRepository
}

func NewAddApplicationComponentExpertHandler(repository AddApplicationComponentExpertRepository) *AddApplicationComponentExpertHandler {
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

	expert, err := valueobjects.NewExpert(command.ExpertName, command.ExpertRole, command.ContactInfo, time.Now().UTC())
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
