# Import Execution Safety in Confirm Handler

## Description
Harden asynchronous import execution so panics, hangs, and shutdown events do not leave sessions in ambiguous state.

## Scope
- `backend/internal/importing/application/handlers/confirm_import_handler.go`
- Adjacent tests for handler behavior

## Out of Scope
- Replacing asynchronous execution model
- Re-architecting import orchestration flow

## Required Changes

### 1) Panic-safe goroutine
Wrap background execution with `defer recover()` and convert panic into import-session failure state.

### 2) Bounded execution time
Use `context.WithTimeout` for background import execution (default target: 30 minutes; configurable constant in handler package).

### 3) Shutdown-aware parent context
Do not root execution in `context.Background()` directly inside handler logic. Use injected long-lived parent context (or injected context provider) so graceful shutdown cancels outstanding imports.

### 4) Deterministic terminal state
Ensure panic/timeout/cancel paths all attempt to persist a terminal status (`failed` with reason) if not already terminal.

## Acceptance Criteria
- Panic in import execution does not crash process and results in failed import session.
- Timeout expiration marks session failed with timeout reason.
- Server shutdown context cancellation propagates into in-flight import execution.
- Unit tests cover panic, timeout, and cancellation flows.

## Verification
- `go test ./internal/importing/...`
- `go test ./internal/...`

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Verification done
- [x] User sign-off

## Implementation Status
- Added panic-safe import execution with recovery-to-failure behavior in confirm handler.
- Added bounded execution timeout with configurable handler-level default.
- Added shutdown-aware parent execution context injection from app bootstrap.
- Ensured cancellation/timeout/panic paths persist terminal `failed` state when session is not terminal.

## Verification Results
- `go test ./internal/importing/application/handlers -run ConfirmImportHandler` passed.
- `go test ./internal/importing/...` passed.
- `go test ./internal/...` passed.

