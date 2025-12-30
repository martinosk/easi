package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateElementColorHandler struct {
	layoutRepository *repositories.ViewLayoutRepository
}

func NewUpdateElementColorHandler(layoutRepository *repositories.ViewLayoutRepository) *UpdateElementColorHandler {
	return &UpdateElementColorHandler{
		layoutRepository: layoutRepository,
	}
}

func (h *UpdateElementColorHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateElementColor)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	_, err := valueobjects.NewHexColor(command.Color)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	var elementType repositories.ElementType
	switch command.ElementType {
	case "component":
		elementType = repositories.ElementTypeComponent
	case "capability":
		elementType = repositories.ElementTypeCapability
	default:
		return cqrs.EmptyResult(), errors.New("invalid element type: must be 'component' or 'capability'")
	}

	if err := h.layoutRepository.UpdateElementColor(ctx, command.ViewID, command.ElementID, elementType, command.Color); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
