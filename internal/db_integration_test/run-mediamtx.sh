#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "$0")/../.." && pwd)
export MEDIAMTX_INTEGRATION_BINARY="${MEDIAMTX_INTEGRATION_BINARY:-$root/mediamtx_v1.17.0_linux_amd64/mediamtx}"
[[ -x "$MEDIAMTX_INTEGRATION_BINARY" ]] || { echo "MediaMTX binary not found: $MEDIAMTX_INTEGRATION_BINARY" >&2; exit 1; }
command -v node >/dev/null
export GOCACHE="${GOCACHE:-/tmp/spotted-integration-go-cache}"
cd "$root"
go test -count=1 -timeout=10m ./internal/db_integration_test/inc3 "$@"
