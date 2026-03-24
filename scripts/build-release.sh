#!/usr/bin/env bash
# Build static-ish release tarballs for common platforms (darwin/linux, amd64/arm64).
# Artifacts: dist/tray_${GOOS}_${GOARCH}.tar.gz suitable for GitHub Releases uploads.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=scripts/tray-env.sh
source "${ROOT}/scripts/tray-env.sh"

load_tray_env "${ROOT}"
ensure_go
cd "${ROOT}"

DIST="${ROOT}/dist"
mkdir -p "${DIST}"

# Avoid macOS extended attributes leaking into tarballs.
export COPYFILE_DISABLE=1

# Release tarballs include version/commit when TRAY_RELEASE_* is set (see publish-release.sh).
LDFLAGS="$(tray_ldflags) $(tray_version_ldflags)"

build_one() {
	local goos="$1"
	local goarch="$2"
	local name="tray_${goos}_${goarch}"
	local tmp="${DIST}/build-${name}"
	rm -rf "${tmp}"
	mkdir -p "${tmp}"
	echo "Building ${goos}/${goarch}..."
	GOOS="${goos}" GOARCH="${goarch}" CGO_ENABLED=0 go build -trimpath \
		-ldflags "${LDFLAGS}" \
		-o "${tmp}/tray" ./cmd/tray
	(
		cd "${tmp}"
		tar -czf "${DIST}/${name}.tar.gz" tray
	)
	rm -rf "${tmp}"
	echo "Wrote ${DIST}/${name}.tar.gz"
}

for goos in darwin linux; do
	for goarch in amd64 arm64; do
		build_one "${goos}" "${goarch}"
	done
done

echo
echo "Done. Upload each dist/tray_*.tar.gz to a GitHub Release (same filenames each release for stable install URLs)."
