# Testing Guide

## Test Structure

The EASI backend has two types of tests:

### 1. Unit Tests
Fast, isolated tests that don't require external dependencies.

**Location**: `*_test.go` files without build tags

**Run:**
```bash
go test ./...
```

**Examples:**
- `internal/architecturemodeling/domain/valueobjects/component_name_test.go`
- `internal/architecturemodeling/domain/aggregates/application_component_test.go`
- `internal/infrastructure/eventstore/event_store_test.go`

### 2. Integration Tests
Tests that verify the full stack including database interactions.

**Location**: `*_integration_test.go` files with `// +build integration` tag

**Run:**
```bash
# Start database
cd /home/devuser/repos/easi
podman compose up -d

# Run integration tests
cd backend
./test_integration.sh

# Or manually:
go test -v -tags=integration ./internal/architecturemodeling/infrastructure/api/... -count=1
```

## Running Specific Tests

### Run only integration tests:
```bash
go test -v -tags=integration ./...
```

### Run only unit tests (exclude integration):
```bash
go test -v ./...
```

### Run tests for specific package:
```bash
# Unit tests for domain layer
go test -v ./internal/architecturemodeling/domain/...

# Integration tests for API layer
go test -v -tags=integration ./internal/architecturemodeling/infrastructure/api/...
```

### Run with coverage:
```bash
# Unit tests
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Integration tests
go test -v -tags=integration -coverprofile=coverage_integration.out ./internal/architecturemodeling/infrastructure/api/...
go tool cover -html=coverage_integration.out -o coverage_integration.html
```