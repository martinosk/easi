---
name: easi-backend-testing
description: MUST load when writing, running, or reviewing backend Go tests in EASI. Load when adding a new test file, choosing between unit and integration tests, running tests in CI, or checking test coverage.
compatibility: opencode
---

# EASI Backend Testing

## Overview

The EASI backend has two test tiers: fast unit tests (no external deps) and integration tests (full stack with PostgreSQL). They are kept strictly separate via build tags and file naming conventions.

## Test Tiers

### Unit Tests

Fast, isolated — no database, no Docker.

- **File naming**: `*_test.go` (no special suffix)
- **Build tag**: none
- **Location examples**:
  - `internal/architecturemodeling/domain/valueobjects/component_name_test.go`
  - `internal/architecturemodeling/domain/aggregates/application_component_test.go`
  - `internal/infrastructure/eventstore/event_store_test.go`

```bash
# Run all unit tests
cd backend && go test ./...

# Run unit tests for a specific package
go test -v ./internal/architecturemodeling/domain/...

# Run with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Integration Tests

Full-stack — requires a running PostgreSQL database.

- **File naming**: `*_integration_test.go`
- **Build tag**: `// +build integration` at top of file

```bash
# Start the database first
cd /workspace
docker-compose up -d   # or: podman compose up -d

# Run all integration tests
cd backend && ./test_integration.sh

# Or manually:
go test -v -tags=integration ./internal/architecturemodeling/infrastructure/api/... -count=1

# Run integration tests for a specific package
go test -v -tags=integration ./...

# Integration tests with coverage
go test -v -tags=integration -coverprofile=coverage_integration.out \
  ./internal/architecturemodeling/infrastructure/api/...
go tool cover -html=coverage_integration.out -o coverage_integration.html
```

## Separation Rules

| Test type | File suffix | Build tag | External deps |
|-----------|------------|-----------|---------------|
| Unit | `_test.go` | none | None allowed |
| Integration | `_integration_test.go` | `// +build integration` | PostgreSQL required |

Running `go test ./...` (no tags) executes **only unit tests**. Integration tests are excluded by default — this keeps the standard test command fast and CI-safe without a database.

## Test Placement by Layer

| Layer | Test type | Example package |
|-------|-----------|----------------|
| Domain value objects | Unit | `domain/valueobjects/` |
| Domain aggregates | Unit | `domain/aggregates/` |
| Event store | Unit | `infrastructure/eventstore/` |
| HTTP handlers / API | Integration | `infrastructure/api/` |
| Repository projectors | Integration | `infrastructure/repository/` |

## Guidelines

1. **Write unit tests for all domain logic** — aggregates, value objects, domain services, event deserialization
2. **Write integration tests for HTTP handlers** — they exercise the full stack including DB
3. **Never skip the build tag** on integration test files — without it, `go test ./...` will try to run them without a DB and fail
4. **Use `-count=1`** on integration tests to bypass the test cache when verifying DB interactions
5. **Run unit tests in CI without a database** — integration tests require the compose stack

## Rule: Tests Must Call the Production Type

A unit test for `FooProjector` must construct `NewFooProjector(...)` (or the real struct literal) and exercise its actual methods. **Re-implementing the projector / handler / dispatcher inside `_test.go` as a `testableFoo` shadow type is forbidden.**

**When you need to inject a mock collaborator into a Go projector/handler/service**:
- Promote the collaborator's contract to an exported (or package-level) interface that the production constructor accepts. Production code keeps using the concrete read model; tests pass a mock that implements the interface.
- Mocks may stand in for *collaborators* (read models, repositories, command bus). Mocks must never stand in for the *system under test*.

**Review smell to grep for**: `type testable`, `newTestable*`, or any test-file type whose `ProjectEvent` / `Handle` / dispatcher method mirrors the production type's. Treat each as suspect until you confirm the production constructor is actually called in the test bodies.