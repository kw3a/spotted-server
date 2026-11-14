#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "$0")/../.." && pwd)
command -v goose >/dev/null || { echo 'goose is required' >&2; exit 1; }
name="spotted-db-integration-$(date +%s)-$$"
password=$(od -An -N16 -tx1 /dev/urandom | tr -d ' \n')
cleanup() { docker rm -fv "$name" >/dev/null; }
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
docker run --detach --name "$name" --label spotted.db-integration=true \
 --publish 127.0.0.1::3306 --env MYSQL_ROOT_PASSWORD="$password" \
 --env MYSQL_ROOT_HOST=% mysql:8.0 >/dev/null
ready=false
for ((attempt=0; attempt<90; attempt++)); do
 if docker exec "$name" sh -c 'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" mysql -uroot -h127.0.0.1 -e "SELECT 1"' >/dev/null 2>&1; then
  ready=true; break
 fi
 sleep 1
done
if [[ "$ready" != true ]]; then echo 'Temporary MySQL did not become ready' >&2; exit 1; fi
port=$(docker port "$name" 3306/tcp)
export DB_INTEGRATION_DSN="root:$password@tcp($port)/?parseTime=true"
export GOCACHE="${GOCACHE:-/tmp/spotted-integration-go-cache}"
cd "$root"
go test -count=1 -timeout=5m ./internal/db_integration_test/inc1 ./internal/db_integration_test/inc2 "$@"
