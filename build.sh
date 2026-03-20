#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/tray-env.sh
source "${ROOT}/scripts/tray-env.sh"

load_tray_env "${ROOT}"
ensure_go
cd "${ROOT}"

exec go build -ldflags "$(tray_ldflags)" -o tray ./cmd/tray
