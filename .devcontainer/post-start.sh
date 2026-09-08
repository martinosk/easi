#!/usr/bin/env bash
# Runs every time the dev container starts.
set -euo pipefail

sudo chown -R node:node \
  /home/node/go \
  /home/node/.npm \
  /home/node/.claude \
  /home/node/pgdata 2>/dev/null || true

bash /workspace/.devcontainer/setup-postgres.sh
