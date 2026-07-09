#!/usr/bin/env bash
# Validate docs/openapi.yaml and bundle JSON for the Go backend embed.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT/docs/openapi.yaml"
OUT_YAML="$ROOT/backend/openapi.yaml"
OUT_JSON="$ROOT/backend/openapi.json"

if [[ ! -f "$SRC" ]]; then
  echo "missing $SRC" >&2
  exit 1
fi

CLI=(npx --yes @apidevtools/swagger-cli)

echo "validating $SRC"
"${CLI[@]}" validate "$SRC"

cp "$SRC" "$OUT_YAML"
"${CLI[@]}" bundle "$SRC" -o "$OUT_JSON" -t json

echo "wrote $OUT_YAML"
echo "wrote $OUT_JSON"
