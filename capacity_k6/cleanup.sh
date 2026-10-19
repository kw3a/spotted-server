#!/usr/bin/env bash
set -euo pipefail
capacity_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
exec "$capacity_dir/maintenance.sh" cleanup "$@"
