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
	local cgo=0

	if [[ "${goos}" == "darwin" ]]; then
		if [[ "$(uname -s)" != "Darwin" ]]; then
			echo "Skipping ${goos}/${goarch}: tray bar needs CGO; build darwin artifacts on macOS."
			return 0
		fi
		local native_arch
		native_arch="$(uname -m)"
		case "${native_arch}" in
		x86_64) native_arch="amd64" ;;
		arm64) ;;
		*)
			echo "Skipping ${goos}/${goarch}: unknown native arch ${native_arch}."
			return 0
			;;
		esac
		if [[ "${goarch}" != "${native_arch}" ]]; then
			echo "Skipping ${goos}/${goarch}: CGO cross-compile not supported; build ${goos}/${native_arch} natively on this Mac."
			return 0
		fi
		cgo=1
	fi

	local name="tray_${goos}_${goarch}"
	local tmp="${DIST}/build-${name}"
	rm -rf "${tmp}"
	mkdir -p "${tmp}"
	echo "Building ${goos}/${goarch} (CGO_ENABLED=${cgo})..."
	GOOS="${goos}" GOARCH="${goarch}" CGO_ENABLED="${cgo}" go build -trimpath \
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
