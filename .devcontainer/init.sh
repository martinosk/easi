#!/bin/sh
set -e

# Persist ~/.claude.json (auth tokens) inside the named volume
PERSISTED_CONFIG=/home/node/.claude/global-config.json
if [ ! -L /home/node/.claude.json ]; then
  if [ -f /home/node/.claude.json ]; then
    mv /home/node/.claude.json "$PERSISTED_CONFIG"
  else
    touch "$PERSISTED_CONFIG"
  fi
  ln -sf "$PERSISTED_CONFIG" /home/node/.claude.json
fi

# Always install/upgrade the latest CLI and MCP tooling.
npm install -g @anthropic-ai/claude-code@latest
npm install -g @codescene/codehealth-mcp@latest

# Symlink opencode skills into .claude/skills
mkdir -p /workspace/.claude/skills
for d in /workspace/.opencode/skills/*/; do
  [ -d "$d" ] || continue
  name=$(basename "$d")
  ln -sfn "$d" "/workspace/.claude/skills/$name"
done
