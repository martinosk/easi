#!/usr/bin/env bash
set -euo pipefail

env_file="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/.env"
vars=(CS_ACCESS_TOKEN CS_ONPREM_URL)
mode="${1:-}"

read_from_file() {
  [ -f "$env_file" ] || return 0
  sed -n "s/^$1=//p" "$env_file" | tail -n1
}

write_to_file() {
  local tmp="$env_file.tmp"
  {
    [ -f "$env_file" ] && grep -v "^$1=" "$env_file" || true
    printf '%s=%s\n' "$1" "$2"
  } > "$tmp"
  mv "$tmp" "$env_file"
  chmod 600 "$env_file"
}

prompt_for() {
  local value
  if [ "$1" = CS_ACCESS_TOKEN ]; then
    read -rs -p "Enter $1 (input hidden): " value
    echo >&2
  else
    read -r -p "Enter $1: " value
  fi
  printf '%s' "$value"
}

[ -f "$env_file" ] || { touch "$env_file"; chmod 600 "$env_file"; }

missing=()
for name in "${vars[@]}"; do
  stored="$(read_from_file "$name")"
  value="${!name:-}"
  [ -n "$value" ] || value="$stored"
  if [ -z "$value" ] && [ -t 0 ]; then
    value="$(prompt_for "$name")"
  fi
  if [ -z "$value" ]; then
    missing+=("$name")
    continue
  fi
  [ "$value" = "$stored" ] || write_to_file "$name" "$value"
  if [ "$mode" = --export ]; then
    printf 'export %s=%q\n' "$name" "$value"
  fi
done

if [ "${#missing[@]}" -gt 0 ]; then
  echo "cs-env: ${missing[*]} not set. Export them on the host or add them to $env_file, then reopen the dev container (or open a new terminal inside it to be prompted)." >&2
fi
