#!/bin/sh
set -e

VOLUME=/home/node/frontend-node-modules
TARGET=/workspace/frontend/node_modules

sudo mkdir -p "$VOLUME"
sudo chown node:node "$VOLUME"

mkdir -p "$TARGET"

if mountpoint -q "$TARGET"; then
  exit 0
fi

sudo mount --bind "$VOLUME" "$TARGET"
