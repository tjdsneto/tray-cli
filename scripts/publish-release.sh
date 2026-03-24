#!/usr/bin/env bash
# Cut a semver Git tag, build release tarballs with embedded version metadata, push the tag,
# and optionally create a GitHub Release with gh(1).
#
# Versioning policy: **git tags are the source of truth** — use annotated tags vMAJOR.MINOR.PATCH
# (optionally with a prerelease suffix, e.g. v1.0.0-rc.1). The same string is embedded in the
# binary via -ldflags (TRAY_RELEASE_VERSION / TRAY_RELEASE_COMMIT).
#
# Usage:
#   ./scripts/publish-release.sh v1.2.3
#
# Prerequisites:
#   - clean git working tree
#   - tag must not already exist locally or on origin
#   - Optional: GitHub CLI (`gh`) for `gh release create` + asset upload
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

VERSION="${1:-}"
if [[ -z "${VERSION}" ]]; then
	echo "usage: $0 <vMAJOR.MINOR.PATCH>" >&2
	exit 1
fi

if [[ ! "${VERSION}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
	echo "invalid version: ${VERSION} (expected v1.2.3 or v1.0.0-rc.1)" >&2
	exit 1
fi

if [[ -n "$(git status --porcelain 2>/dev/null)" ]]; then
	echo "refusing to publish: working tree is not clean (commit or stash first)" >&2
	exit 1
fi

if git rev-parse "${VERSION}" >/dev/null 2>&1; then
	echo "refusing to publish: tag ${VERSION} already exists locally" >&2
	exit 1
fi

if git remote get-url origin >/dev/null 2>&1; then
	if git ls-remote --tags origin "refs/tags/${VERSION}" | grep -q .; then
		echo "refusing to publish: tag ${VERSION} already exists on origin" >&2
		exit 1
	fi
fi

echo "Running tests..."
go test ./... -count=1

export TRAY_RELEASE_VERSION="${VERSION}"
export TRAY_RELEASE_COMMIT="$(git rev-parse --short HEAD)"

echo "Building release artifacts for ${TRAY_RELEASE_VERSION} (${TRAY_RELEASE_COMMIT})..."
./scripts/build-release.sh

echo "Creating annotated tag ${VERSION}..."
git tag -a "${VERSION}" -m "Release ${VERSION}"

echo "Pushing tag to origin..."
git push origin "${VERSION}"

if command -v gh >/dev/null 2>&1; then
	echo "Creating GitHub Release and uploading dist/tray_*.tar.gz..."
	gh release create "${VERSION}" "${ROOT}/dist"/tray_*.tar.gz \
		--title "${VERSION}" \
		--generate-notes
else
	echo
	echo "gh not found — create the GitHub Release manually and upload:"
	echo "  ${ROOT}/dist/tray_*.tar.gz"
fi

echo
echo "Done. Users can install with scripts/install.sh (TRAY_VERSION=${VERSION})."
