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

func (h *CancelImportHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.CancelImport)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	session, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	if err := session.Cancel(); err != nil {
		return err
	}

	return h.repository.Save(ctx, session)
}
