package handlers

import (
	"context"

	"easi/backend/internal/importing/application/commands"
	"easi/backend/internal/importing/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CancelImportHandler struct {
	repository *repositories.ImportSessionRepository
}

func NewCancelImportHandler(repository *repositories.ImportSessionRepository) *CancelImportHandler {
	return &CancelImportHandler{repository: repository}
}

func (h *CancelImportHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CancelImport)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	session, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := session.Cancel(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, session); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
