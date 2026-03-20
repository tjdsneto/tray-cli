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

# Prints one -ldflags argument (two -X symbols). Requires TRAY_* from env or .env.
tray_ldflags() {
	local pkg="github.com/tjdsneto/tray-cli/internal/config"
	printf '%s' "-X ${pkg}.EmbeddedSupabaseURL=${TRAY_SUPABASE_URL-} -X ${pkg}.EmbeddedSupabaseAnonKey=${TRAY_SUPABASE_ANON_KEY-}"
}
