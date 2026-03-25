#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/tray-env.sh
source "${ROOT}/scripts/tray-env.sh"

load_tray_env "${ROOT}"
ensure_go
cd "${ROOT}"

# Optional: prebuilt binary (e.g. scripts/generate-cli-report.sh sets TRAY_CLI_BIN) to avoid `go run` per invocation.
if [[ -n "${TRAY_CLI_BIN:-}" ]]; then
	_bin="${TRAY_CLI_BIN}"
	if [[ "${_bin}" != /* ]]; then
		_bin="${ROOT}/${_bin}"
	fi
	if [[ -x "${_bin}" ]]; then
		exec "${_bin}" "$@"
	fi
fi

exec go run -ldflags "$(tray_ldflags)" ./cmd/tray "$@"
