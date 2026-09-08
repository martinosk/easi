#!/usr/bin/env bash
# One-time provisioning after the dev container is created.
set -euo pipefail

cd /workspace

echo "==> Starting PostgreSQL"
bash .devcontainer/setup-postgres.sh

echo "==> Running database migrations"
(
  cd backend
  DB_ADMIN_CONN_STRING="host=localhost port=5432 user=easi password=easi dbname=easi sslmode=disable" \
  EASI_APP_PASSWORD=localdev \
  EASI_ADMIN_PASSWORD=localdev \
    go run cmd/migrate/main.go
)

echo "==> Downloading Go modules"
(cd backend && go mod download)

echo "==> Installing Go tools (swag)"
go install github.com/swaggo/swag/cmd/swag@v1.16.6

echo "==> Installing frontend dependencies"
(cd frontend && npm install)

echo "==> Installing Playwright browsers"
(cd frontend && npx playwright install chromium)

echo "==> Installing CodeHealth MCP"
npm install -g @codescene/codehealth-mcp@latest

echo "==> Persisting Claude config in the named volume"
persisted_config=/home/node/.claude/global-config.json
if [ ! -L /home/node/.claude.json ]; then
  if [ -f /home/node/.claude.json ]; then
    mv /home/node/.claude.json "${persisted_config}"
  else
    touch "${persisted_config}"
  fi
  ln -sf "${persisted_config}" /home/node/.claude.json
fi

echo "==> Linking opencode skills into .claude/skills"
mkdir -p /workspace/.claude/skills
for d in /workspace/.opencode/skills/*/; do
  [ -d "${d}" ] || continue
  ln -sfn "${d}" "/workspace/.claude/skills/$(basename "${d}")"
done

echo "==> Dev container ready"
