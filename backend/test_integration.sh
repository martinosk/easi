#!/bin/bash

# Integration test runner script
# Requires PostgreSQL to be running on localhost:5432

set -e

echo "Running integration tests..."

# Run integration tests for architecture modeling
echo "Running architecture modeling integration tests..."
go test -v -tags=integration ./internal/architecturemodeling/infrastructure/api/... -count=1

echo ""

# Run integration tests for architecture views
echo "Running architecture views integration tests..."
go test -v -tags=integration ./internal/architectureviews/infrastructure/api/... -count=1

echo ""

# Run integration tests for capability mapping
echo "Running capability mapping integration tests..."
go test -v -tags=integration ./internal/capabilitymapping/infrastructure/api/... -count=1

echo ""

# Run integration tests for auth
echo "Running auth integration tests..."
go test -v -tags=integration ./internal/auth/infrastructure/api/... -count=1

echo ""

# Run integration tests for platform (tenant management)
echo "Running platform integration tests..."
go test -v -tags=integration ./internal/platform/infrastructure/api/... -count=1

echo ""

# Run integration tests for importing
echo "Running importing integration tests..."
go test -v -tags=integration ./internal/importing/application/parsers/... -count=1

echo ""

# Run integration tests for database tenant isolation
echo "Running database tenant isolation integration tests..."
go test -v -tags=integration ./internal/infrastructure/database/... -count=1

echo ""

# Run integration tests for metamodel
echo "Running metamodel integration tests..."
go test -v -tags=integration ./internal/metamodel/infrastructure/api/... -count=1

echo ""

# Run integration tests for enterprise architecture
echo "Running enterprise architecture integration tests..."
go test -v -tags=integration ./internal/enterprisearchitecture/application/... -count=1

echo ""

# Run integration tests for audit
echo "Running audit integration tests..."
go test -v -tags=integration ./internal/shared/audit/... -count=1

echo ""

# Run integration tests for test fixtures
echo "Running test fixtures integration tests..."
go test -v -tags=integration ./internal/testing/... -count=1

echo ""
echo "✓ All integration tests complete!"
