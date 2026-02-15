package handlers

import (
	"context"
	"log"

	"easi/backend/internal/importing/application/commands"
	"easi/backend/internal/importing/application/saga"
	"easi/backend/internal/importing/infrastructure/repositories"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type ConfirmImportHandler struct {
	repository  *repositories.ImportSessionRepository
	importSaga  *saga.ImportSaga
}

func NewConfirmImportHandler(repository *repositories.ImportSessionRepository, importSaga *saga.ImportSaga) *ConfirmImportHandler {
	return &ConfirmImportHandler{
		repository: repository,
		importSaga: importSaga,
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
	if actor, ok := sharedctx.GetActor(ctx); ok {
		bgCtx = sharedctx.WithActor(bgCtx, actor)
	}
	go h.executeImport(bgCtx, command.ID)

	return cqrs.EmptyResult(), nil
}

func (h *ConfirmImportHandler) executeImport(ctx context.Context, sessionID string) {
	session, err := h.repository.GetByID(ctx, sessionID)
	if err != nil {
		log.Printf("Failed to load import session %s: %v", sessionID, err)
		return
	}

	result := h.importSaga.Execute(ctx, session.ParsedData(), session.BusinessDomainID(), session.CapabilityEAOwner())

	if err := session.Complete(result); err != nil {
		log.Printf("Failed to mark import session %s complete: %v", sessionID, err)
	}

	if err := h.repository.Save(ctx, session); err != nil {
		log.Printf("Failed to save import session %s: %v", sessionID, err)
	}
}
