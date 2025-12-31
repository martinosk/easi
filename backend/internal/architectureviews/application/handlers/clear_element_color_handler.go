package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type ElementColorClearer interface {
	ClearElementColor(ctx context.Context, viewID, elementID string, elementType repositories.ElementType) error
}

type ClearElementColorHandler struct {
	layoutRepository ElementColorClearer
}

func NewClearElementColorHandler(layoutRepository ElementColorClearer) *ClearElementColorHandler {
	return &ClearElementColorHandler{
		layoutRepository: layoutRepository,
	}
}

func (h *ClearElementColorHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ClearElementColor)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
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

	if err := h.layoutRepository.ClearElementColor(ctx, command.ViewID, command.ElementID, elementType); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
