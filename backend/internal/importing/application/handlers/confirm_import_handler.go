package handlers

import (
	"context"
	"log"

	"easi/backend/internal/importing/application/commands"
	"easi/backend/internal/importing/application/orchestrator"
	"easi/backend/internal/importing/infrastructure/repositories"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type ConfirmImportHandler struct {
	repository   *repositories.ImportSessionRepository
	orchestrator *orchestrator.ImportOrchestrator
}

func NewConfirmImportHandler(repository *repositories.ImportSessionRepository, orch *orchestrator.ImportOrchestrator) *ConfirmImportHandler {
	return &ConfirmImportHandler{
		repository:   repository,
		orchestrator: orch,
	}
}

func (h *ConfirmImportHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ConfirmImport)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	session, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := session.StartImport(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, session); err != nil {
		return cqrs.EmptyResult(), err
	}

	tenantID := sharedctx.GetTenantOrDefault(ctx)
	bgCtx := sharedctx.WithTenant(context.Background(), tenantID)
	go h.executeImport(bgCtx, command.ID)

	return cqrs.EmptyResult(), nil
}

func (h *ConfirmImportHandler) executeImport(ctx context.Context, sessionID string) {
	session, err := h.repository.GetByID(ctx, sessionID)
	if err != nil {
		log.Printf("Failed to load import session %s: %v", sessionID, err)
		return
	}

	_, err = h.orchestrator.Execute(ctx, session)
	if err != nil {
		log.Printf("Import execution failed for session %s: %v", sessionID, err)
		_ = session.Fail(err.Error())
	}

	if err := h.repository.Save(ctx, session); err != nil {
		log.Printf("Failed to save import session %s: %v", sessionID, err)
	}
}
