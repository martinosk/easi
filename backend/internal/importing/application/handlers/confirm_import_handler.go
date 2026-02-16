package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"easi/backend/internal/importing/application/commands"
	"easi/backend/internal/importing/application/saga"
	"easi/backend/internal/importing/domain/aggregates"
	"easi/backend/internal/importing/infrastructure/repositories"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

const (
	DefaultImportExecutionTimeout   = 30 * time.Minute
	terminalStatePersistenceTimeout = 5 * time.Second
	reasonImportExecutionTimedOut   = "import execution timed out"
	reasonImportExecutionCancelled  = "import execution cancelled"
)

type ConfirmImportHandler struct {
	repository       *repositories.ImportSessionRepository
	importSaga       *saga.ImportSaga
	executionParent  context.Context
	executionTimeout time.Duration
}

func NewConfirmImportHandler(repository *repositories.ImportSessionRepository, importSaga *saga.ImportSaga) *ConfirmImportHandler {
	return NewConfirmImportHandlerWithExecutionContext(repository, importSaga, context.Background(), DefaultImportExecutionTimeout)
}

func NewConfirmImportHandlerWithExecutionContext(repository *repositories.ImportSessionRepository, importSaga *saga.ImportSaga, executionParent context.Context, executionTimeout time.Duration) *ConfirmImportHandler {
	if executionParent == nil {
		executionParent = context.Background()
	}
	if executionTimeout <= 0 {
		executionTimeout = DefaultImportExecutionTimeout
	}

	return &ConfirmImportHandler{
		repository:       repository,
		importSaga:       importSaga,
		executionParent:  executionParent,
		executionTimeout: executionTimeout,
	}
}

func (h *ConfirmImportHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ConfirmImport)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	session, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), fmt.Errorf("load import session %s: %w", command.ID, err)
	}

	if err := session.StartImport(); err != nil {
		return cqrs.EmptyResult(), fmt.Errorf("start import session %s: %w", command.ID, err)
	}

	if err := h.repository.Save(ctx, session); err != nil {
		return cqrs.EmptyResult(), fmt.Errorf("persist started import session %s: %w", command.ID, err)
	}

	tenantID := sharedctx.GetTenantOrDefault(ctx)
	bgCtx := sharedctx.WithTenant(h.executionParent, tenantID)
	if actor, ok := sharedctx.GetActor(ctx); ok {
		bgCtx = sharedctx.WithActor(bgCtx, actor)
	}
	go h.executeImport(bgCtx, command.ID)

	return cqrs.EmptyResult(), nil
}

func (h *ConfirmImportHandler) executeImport(ctx context.Context, sessionID string) {
	execCtx, cancel := context.WithTimeout(ctx, h.executionTimeout)
	defer cancel()

	session, err := h.repository.GetByID(execCtx, sessionID)
	if err != nil {
		log.Printf("failed to load import session %s for execution: %v", sessionID, err)
		return
	}

	importResult, reason := h.executeImportWithRecovery(execCtx, session)
	if reason != "" {
		h.failSessionWithReason(ctx, sessionID, reason)
		return
	}

	if err := session.Complete(importResult); err != nil {
		log.Printf("failed to mark import session %s complete: %v", sessionID, err)
		return
	}

	if err := h.repository.Save(ctx, session); err != nil {
		log.Printf("failed to persist completed import session %s: %v", sessionID, err)
	}
}

type executionResult struct {
	result aggregates.ImportResult
	panicV any
}

func (h *ConfirmImportHandler) executeImportWithRecovery(execCtx context.Context, session *aggregates.ImportSession) (aggregates.ImportResult, string) {
	done := make(chan executionResult, 1)
	go func() {
		defer func() {
			if panicValue := recover(); panicValue != nil {
				done <- executionResult{panicV: panicValue}
			}
		}()
		result := h.importSaga.Execute(execCtx, session.ParsedData(), session.BusinessDomainID(), session.CapabilityEAOwner())
		done <- executionResult{result: result}
	}()

	select {
	case <-execCtx.Done():
		return aggregates.ImportResult{}, failureReasonFromContextError(execCtx.Err())
	case result := <-done:
		if result.panicV != nil {
			return aggregates.ImportResult{}, fmt.Sprintf("import execution panic: %v", result.panicV)
		}
		return result.result, ""
	}
}

func failureReasonFromContextError(err error) string {
	if err == nil {
		return reasonImportExecutionCancelled
	}
	if err == context.DeadlineExceeded {
		return reasonImportExecutionTimedOut
	}
	return reasonImportExecutionCancelled
}

func (h *ConfirmImportHandler) failSessionWithReason(ctx context.Context, sessionID, reason string) {
	persistenceCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), terminalStatePersistenceTimeout)
	defer cancel()

	session, err := h.repository.GetByID(persistenceCtx, sessionID)
	if err != nil {
		log.Printf("failed to reload import session %s for failure persistence: %v", sessionID, err)
		return
	}

	if session.Status().IsCompleted() || session.Status().IsFailed() {
		return
	}

	if err := session.Fail(reason); err != nil {
		log.Printf("failed to mark import session %s as failed: %v", sessionID, err)
		return
	}

	if err := h.repository.Save(persistenceCtx, session); err != nil {
		log.Printf("failed to persist failed import session %s: %v", sessionID, err)
	}
}
