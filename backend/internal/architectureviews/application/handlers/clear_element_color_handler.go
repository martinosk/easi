package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type ClearElementColorHandler struct {
	layoutRepository *repositories.ViewLayoutRepository
}

func NewClearElementColorHandler(layoutRepository *repositories.ViewLayoutRepository) *ClearElementColorHandler {
	return &ClearElementColorHandler{
		layoutRepository: layoutRepository,
	}
}

func (h *ClearElementColorHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.ClearElementColor)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	var elementType repositories.ElementType
	switch command.ElementType {
	case "component":
		elementType = repositories.ElementTypeComponent
	case "capability":
		elementType = repositories.ElementTypeCapability
	default:
		return errors.New("invalid element type: must be 'component' or 'capability'")
	}

	return h.layoutRepository.ClearElementColor(ctx, command.ViewID, command.ElementID, elementType)
}
