# Projector Idempotency and Replay Safety

## Description
Guarantee that projector handlers are safe under retries and event replays, so transient failures (spec 140) do not create duplicate or divergent read-model state.

## Why
After enabling strict error propagation, retries become expected behavior. Projectors must therefore be idempotent for the same event and deterministic across replay.

## Scope
- Projectors and read-model write paths touched by event retries in:
  - `capabilitymapping`
  - `architecturemodeling`
  - `enterprisearchitecture`

## Out of Scope
- Global event bus redesign
- Exactly-once delivery guarantees at infrastructure level

## Required Changes

### 1) Verify idempotent write patterns
Use `UPSERT`, guarded `DELETE`, and deterministic recomputation where possible so replaying the same event does not duplicate rows or corrupt derived values.

### 2) Remove side effects that are replay-unsafe
Avoid non-deterministic behavior in projector handlers (for example time-based mutation without event timestamp input).

### 3) Add focused replay tests
Add tests that run the same event twice and assert read-model end state is unchanged after first successful projection.

## Acceptance Criteria
- Replaying an already-processed event produces no duplicate read-model rows.
- Retry after transient failure converges to correct state.
- Projector tests cover duplicate-delivery scenarios for critical projectors.

## Verification
- `go test ./internal/...`
- Targeted tests for duplicate-event handling in affected projector packages.

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Focused replay tests added
- [x] Targeted integration verification passed
- [x] User sign-off

## Implementation Summary
- Added replay-safe `DELETE + INSERT` idempotency for create-style read-model write paths in:
  - capabilitymapping read models
  - architecturemodeling read models
  - enterprisearchitecture read models
- Pattern: DELETE existing row by business key (tenant_id + id), then plain INSERT. On first run the delete is a no-op; on replay it removes the stale row before re-inserting. This avoids any reliance on database constraints for correctness.
- Added focused duplicate-delivery integration tests:
  - `TestBusinessDomainReadModel_Insert_IdempotentReplay`
  - `TestDomainCapabilityAssignmentReadModel_Insert_IdempotentReplay`
  - `TestApplicationComponentReadModel_Insert_IdempotentReplay`
  - `TestEnterpriseCapabilityLinkReadModel_Insert_IdempotentReplay`
