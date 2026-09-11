#!/bin/bash
# Runs every integration test package against the migrated PostgreSQL named by INTEGRATION_TEST_DB_*
# (defaults: localhost:5432, easi_app/localdev; the dev container presets INTEGRATION_TEST_DB_HOST=postgres).
# The auth package needs Dex (OIDC); it is skipped when Dex does not answer.

set -e

dex_url="${DEX_URL:-http://${DEX_HOST:-localhost}:5556/dex/.well-known/openid-configuration}"
packages=$(go list -tags=integration ./...)

if ! curl -sf -o /dev/null "$dex_url"; then
  echo "Dex not reachable at $dex_url — skipping ./internal/auth/infrastructure/api"
  packages=$(echo "$packages" | grep -v '/internal/auth/infrastructure/api$')
fi

echo "Running integration tests..."
# shellcheck disable=SC2086
go test -tags=integration -count=1 $packages
echo "✓ All integration tests complete!"
