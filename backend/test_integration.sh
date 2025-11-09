#!/bin/bash

# Integration test runner script
# Requires PostgreSQL to be running on localhost:5432

set -e

echo "Running integration tests..."
echo "Prerequisites:"
echo "  - PostgreSQL running on localhost:5432"
echo "  - Database 'easi' with user 'easi' password 'easi'"
echo ""

# Run integration tests
echo "Running tests with -tags=integration..."
go test -v -tags=integration ./internal/architecturemodeling/infrastructure/api/... -count=1

echo ""
echo "✓ Integration tests complete!"
