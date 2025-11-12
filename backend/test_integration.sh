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
echo "✓ All integration tests complete!"
