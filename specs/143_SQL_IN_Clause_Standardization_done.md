# SQL IN-Clause Standardization in Read Models

## Description
Standardize dynamic `IN` filtering in read models to `= ANY($N)` with `pq.Array(...)` and remove custom placeholder builders where no longer needed.

## Scope
- Runtime SQL in read models/projectors where `IN (...)` lists are currently assembled via string formatting.

## Out of Scope
- Static SQL statements with fixed predicates
- Migration SQL files

## Required Changes

### 1) Replace dynamic `IN (...)` construction
For variable-length ID lists, use:
- SQL: `column = ANY($N)`
- Args: `pq.Array(ids)`

### 2) Remove obsolete helper utilities
Delete or inline helper functions that only exist to build placeholder strings for `IN (...)` lists.

### 3) Preserve empty-list semantics
Handle empty arrays explicitly (no-op delete/update or short-circuit return) so behavior is deterministic and no invalid SQL is generated.

## Acceptance Criteria
- No read model/projector builds dynamic `IN (...)` lists with `fmt.Sprintf` or string concatenation.
- All converted paths preserve previous behavior for empty and non-empty list inputs.
- Integration tests for affected read models pass.

## Verification
- `go test ./internal/...`
- Focused grep for dynamic `IN (` query builders in affected packages.

## Implementation Status
- [x] Specification ready
- [x] Implementation done
- [x] Unit/integration tests implemented and passing for affected paths
- [x] User sign-off

## Validation Notes
- Converted read-model runtime dynamic `IN (...)` construction to `= ANY($N)` + `pq.Array(...)` in affected paths.
- Removed obsolete placeholder-builder helpers used only for dynamic `IN` construction.
- Verified targeted integration tests pass:
	- `TestBusinessDomainReadModel_Insert_IdempotentReplay`
	- `TestDomainCapabilityAssignmentReadModel_Insert_IdempotentReplay`
	- `TestApplicationComponentReadModel_Insert_IdempotentReplay`
	- `TestEnterpriseCapabilityLinkReadModel_Insert_IdempotentReplay`

