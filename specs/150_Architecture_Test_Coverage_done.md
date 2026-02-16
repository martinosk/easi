# Architecture SQL Guardrail Coverage Hardening

## Description
Strengthen `architecture_sql_test.go` so SQL ownership checks cover all relevant runtime SQL locations without introducing noisy false positives.

## Scope
- `backend/internal/architecture_sql_test.go`

## Required Changes

### 1) Expand scanner target paths conservatively
Include runtime SQL hot spots beyond read models/projectors (for example repositories/adapters where raw SQL is allowed by design).

### 2) Add a negative guard for unscanned paths
Introduce a complementary test that detects SQL-like table operations in non-scanned production files and fails with guidance to either:
- move SQL to approved locations, or
- explicitly extend scanner patterns.

### 3) Reduce extractor blind spots
Improve SQL extraction to handle multiline queries and common formatting variants used in this repository.

### 4) Document guardrail boundaries
State in-test what is intentionally scanned and why, so future contributors understand where SQL is expected.

## Acceptance Criteria
- Architecture SQL test catches ownership violations in all approved SQL locations.
- SQL appearing outside approved locations causes a clear failing test.
- Existing architecture tests remain stable and pass in CI.

## Verification
- `go test ./internal -run Architecture`
- `go test ./internal/...`
