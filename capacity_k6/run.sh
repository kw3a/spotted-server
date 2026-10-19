#!/usr/bin/env bash
set -euo pipefail
capacity_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$capacity_dir"
case "${1:-}" in
  seed|cleanup|verify-judge0)
    printf 'Comando de mantenimiento separado: utiliza ./%s.sh (sin subcomando).\n' "$1" >&2
    exit 2
    ;;
esac
mkdir -p .bin
printf 'Compilando el lanzador HTTP de capacity_k6...\n' >&2
go build -o .bin/capacity-runner \
  ./fixtures/runner_main.go ./fixtures/run_command.go ./fixtures/runner.go \
  ./fixtures/run_config.go ./fixtures/files.go ./fixtures/types.go ./fixtures/run_identity.go
printf 'Compilacion completada; iniciando carga HTTP...\n' >&2
exec .bin/capacity-runner "$@"
