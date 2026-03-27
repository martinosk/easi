package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type ColorSchemeUpdater interface {
	UpdatePreference(ctx context.Context, viewID string, key repositories.PreferenceKey, value string) error
}

type UpdateViewColorSchemeHandler struct {
	layoutRepository ColorSchemeUpdater
}

func NewUpdateViewColorSchemeHandler(layoutRepository ColorSchemeUpdater) *UpdateViewColorSchemeHandler {
	return &UpdateViewColorSchemeHandler{
		layoutRepository: layoutRepository,
	}
}

func (h *UpdateViewColorSchemeHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateViewColorScheme)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	_, err := valueobjects.NewColorScheme(command.ColorScheme)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.layoutRepository.UpdatePreference(ctx, command.ViewID, repositories.PreferenceKeyColorScheme, command.ColorScheme); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
