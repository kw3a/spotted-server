#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "$0")/../.." && pwd)
# Fail before creating resources if a required image is not installed.
for image in judge0/judge0:1.13.1 postgres:16.2 redis:7.2.4; do
 docker image inspect "$image" >/dev/null 2>&1 || {
  echo "Required local image is missing: $image (no images will be downloaded)" >&2
  exit 1
 }
done
project="spotted-judge0-integration-$(date +%s)-$$"
config_dir=$(mktemp -d /tmp/spotted-judge0-integration.XXXXXX)
export JUDGE0_TEST_CONFIG="$config_dir/judge0.conf"
random_secret() { od -An -N16 -tx1 /dev/urandom | tr -d ' \n'; }
token=$(random_secret)
umask 077
cat > "$JUDGE0_TEST_CONFIG" <<CONFIG
AUTHN_HEADER=X-Auth-Token
AUTHN_TOKEN=$token
REDIS_HOST=redis
REDIS_PASSWORD=$(random_secret)
POSTGRES_HOST=db
POSTGRES_DB=judge0
POSTGRES_USER=judge0
POSTGRES_PASSWORD=$(random_secret)
RAILS_ENV=production
RAILS_MAX_THREADS=2
RAILS_SERVER_PROCESSES=1
ENABLE_CALLBACKS=true
ENABLE_BATCHED_SUBMISSIONS=true
CALLBACKS_TIMEOUT=10
COUNT=2
INTERVAL=0.1
CONFIG
# The container's non-root API user must be able to read the bind-mounted config.
chmod 644 "$JUDGE0_TEST_CONFIG"
compose=(docker compose --env-file "$JUDGE0_TEST_CONFIG" -p "$project" -f "$root/internal/db_integration_test/judge0.compose.yaml")
cleanup() {
 "${compose[@]}" down --volumes --remove-orphans >/dev/null
 rm -rf "$config_dir"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
"${compose[@]}" up --detach --pull never --no-build --wait --wait-timeout 180
port=$("${compose[@]}" port server 2358)
gateway=$(docker network inspect "${project}_default" --format '{{(index .IPAM.Config 0).Gateway}}')
export JUDGE0_INTEGRATION_URL="http://$port"
export JUDGE0_INTEGRATION_CALLBACK_HOST="$gateway"
export JUDGE0_INTEGRATION_TOKEN="$token"
export GOCACHE="${GOCACHE:-/tmp/spotted-integration-go-cache}"
cd "$root"
go test -count=1 -timeout=10m ./internal/db_integration_test/inc1 -run '^TestJudge0' "$@"
