#!/usr/bin/env bash
# Shared by build.sh and run.sh: load repo-root .env and produce -ldflags for embeds.

load_tray_env() {
	local root="$1"
	if [[ -f "${root}/.env" ]]; then
		set -a
		# shellcheck disable=SC1091
		source "${root}/.env"
		set +a
	fi
}

# Prints one -ldflags argument. Requires TRAY_* from env or .env.
# Set TRAY_EMBED_DEV_OAUTH_HINTS=1 in .env for maintainer builds only (OAuth redirect diagnostics in tray login).
tray_ldflags() {
	local pkg="github.com/tjdsneto/tray-cli/internal/config"
	local flags="-X ${pkg}.EmbeddedSupabaseURL=${TRAY_SUPABASE_URL-} -X ${pkg}.EmbeddedSupabaseAnonKey=${TRAY_SUPABASE_ANON_KEY-}"
	if [[ "${TRAY_EMBED_DEV_OAUTH_HINTS-}" == "1" ]]; then
		flags="${flags} -X ${pkg}.EmbeddedDevOAuthHints=1"
	fi
	printf '%s' "${flags}"
}

# Prepend common install dirs when `go` is missing from PATH (e.g. fresh Terminal, GUI apps).
ensure_go() {
	if command -v go >/dev/null 2>&1; then
		return 0
	fi
	local d
	for d in /opt/homebrew/bin /usr/local/go/bin "${HOME}/go/bin"; do
		if [[ -x "${d}/go" ]]; then
			export PATH="${d}:${PATH}"
			return 0
		fi
	done
	echo "tray: 'go' not found. Install Go (https://go.dev/dl/) or add it to PATH." >&2
	echo "  macOS (Homebrew): brew install go" >&2
	exit 127
}
