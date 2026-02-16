# Platform/Auth Boundary Decoupling

## Description
Remove direct `platform -> auth` internal-package coupling and relocate generic HTTP middleware out of a bounded-context package.

## Scope
- `platform` route wiring that currently constructs/depends on auth internals.
- Shared rate limiter middleware currently exposed from platform infrastructure.
- Architecture guardrail allowlist entries tied to this coupling.

## Out of Scope
- Forcing all architecture allowlists to zero in one step.

## Required Changes

### 1) Platform consumes auth via published contracts
Replace imports of `auth/application/*` and `auth/infrastructure/*` from platform with published language DTOs/contracts plus shared command bus integration.

### 2) Move generic middleware to shared infrastructure
Move rate limiting middleware to shared infrastructure package (`internal/infrastructure/api/middleware` or equivalent shared location) and update all consumers.

### 3) Remove obsolete architecture exceptions
Delete only the allowlist entries that were needed for `platform -> auth` and shared-middleware leakage after migration is complete.

## Acceptance Criteria
- No direct imports from platform into auth internal application/infrastructure packages.
- Rate limiter is imported from shared infrastructure, not platform BC package.
- Architecture tests pass with related exceptions removed.
- No behavior change in auth flows or endpoint protections.

## Verification
- `go test ./internal/...`
- Guard test run proving removed exceptions are no longer required.
