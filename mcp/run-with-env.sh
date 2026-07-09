#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ -f "$ROOT_DIR/.env" ]; then
  set -a
  # shellcheck disable=SC1091
  . "$ROOT_DIR/.env"
  set +a
fi

export MINIAPPS_API_URL="${MINIAPPS_API_URL:-${API_URL:-https://api.nimiqminiapps.com}}"
export MINIAPPS_ADMIN_TOKEN="${MINIAPPS_ADMIN_TOKEN:-${ADMIN_TOKEN:-}}"

if [ ! -f "$ROOT_DIR/mcp/dist/index.js" ]; then
  npm --prefix "$ROOT_DIR/mcp" run build
fi

exec node "$ROOT_DIR/mcp/dist/index.js"
