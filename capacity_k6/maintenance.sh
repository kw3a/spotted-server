#!/usr/bin/env bash
set -euo pipefail
capacity_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$capacity_dir"
case "${1:-}" in
  seed|cleanup|verify-judge0) ;;
  *) printf 'Uso: ./maintenance.sh {seed|cleanup|verify-judge0} [flags]\n' >&2; exit 2 ;;
esac
mkdir -p .bin
printf 'Compilando el ejecutable de mantenimiento...\n' >&2
go build -o .bin/capacity-maintenance ./fixtures
printf 'Compilacion completada; iniciando %s...\n' "$1" >&2
exec .bin/capacity-maintenance "$@"
