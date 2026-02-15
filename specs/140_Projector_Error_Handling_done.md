# Projector Error Handling: Fail Fast on Projection Failure

## Description
Stop silent projector corruption by ensuring failed projector operations return errors to the event processing pipeline instead of logging and returning success.

## Why
Projectors are write-side consumers for read models. Returning `nil` on real failures acknowledges an event as processed while leaving read models stale or inconsistent.

## Scope
- `backend/internal/capabilitymapping/application/projectors/*.go`
- `backend/internal/enterprisearchitecture/application/projectors/*.go`
- `backend/internal/architecturemodeling/application/projectors/*.go`

## Out of Scope
- Projector rewrites or routing refactors
- Changing retry policy in event store infrastructure

## Required Changes

### 1) Eliminate swallowed errors
Replace all `log + return nil` paths where an operation failed with `return fmt.Errorf(...: %w, err)` including operation and entity context.

### 2) Treat impossible state as failure
If a projector writes and immediately cannot read expected data (for example `GetByID` returns `nil` after successful update), return a consistency error, do not treat as success.

### 3) Keep logging optional and non-authoritative
If logging is retained for diagnostics, it must be followed by returning the wrapped error. Logging alone is not completion.

## Acceptance Criteria
- No projector path returns `nil` after a failed DB call, recomputation, unmarshal, or downstream projector action.
- Returned errors use `%w` and include operation plus key identifiers.
- Event reprocessing semantics are preserved: failure propagates to caller for retry.
- Existing projector tests pass.

## Verification
- `go test ./internal/...`
- Targeted grep check for suspicious pattern: error logged and then `return nil` in projector files.

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests passing
- [x] Verification done
- [ ] User sign-off

## Implementation Status
- Implemented in scoped projector files for `capabilitymapping` and `enterprisearchitecture`
- Added/updated targeted projector tests for fail-fast behavior

## Verification Results
- `go test ./internal/...` passed
- Targeted grep check for `log ... return nil` in scoped projector files found no matches
