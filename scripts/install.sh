#!/usr/bin/env bash
# Install the tray CLI from GitHub Releases (prebuilt tarballs).
#
# One-liner (after you publish release assets — see docs/distribution.md):
#   curl -fsSL https://raw.githubusercontent.com/tjdsneto/tray-cli/main/scripts/install.sh | bash
#
# Environment:
#   TRAY_INSTALL_REPO   default: tjdsneto/tray-cli
#   TRAY_VERSION        default: latest  (Git tag like v1.2.3, or "latest")
#   TRAY_INSTALL_DIR      default: /usr/local/bin if writable, else ~/.local/bin
#
set -euo pipefail

REPO="${TRAY_INSTALL_REPO:-tjdsneto/tray-cli}"
VERSION="${TRAY_VERSION:-latest}"
DEST="${TRAY_INSTALL_DIR:-}"

case "$(uname -s)" in
Darwin) OS=darwin ;;
Linux) OS=linux ;;
*)
	echo "tray install: unsupported OS $(uname -s) (only darwin and linux)" >&2
	exit 1
	;;
esac

case "$(uname -m)" in
x86_64 | amd64) ARCH=amd64 ;;
arm64 | aarch64) ARCH=arm64 ;;
*)
	echo "tray install: unsupported CPU $(uname -m) (only amd64 and arm64)" >&2
	exit 1
	;;
esac

ASSET="tray_${OS}_${ARCH}.tar.gz"
if [[ "${VERSION}" == "latest" ]]; then
	URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"
else
	URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"
fi

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

echo "Downloading ${URL}"
if ! curl -fsSL "${URL}" -o "${TMP}/tray.tgz"; then
	echo >&2
	echo "tray install: download failed. Check that a release exists for ${REPO} with asset ${ASSET}." >&2
	echo "Alternative (requires Go): go install github.com/tjdsneto/tray-cli/cmd/tray@latest" >&2
	exit 1
fi

tar -xzf "${TMP}/tray.tgz" -C "${TMP}"
if [[ ! -x "${TMP}/tray" ]]; then
	echo "tray install: expected a 'tray' binary inside the tarball" >&2
	exit 1
fi

if [[ -z "${DEST}" ]]; then
	if [[ -w "/usr/local/bin" ]] 2>/dev/null; then
		DEST="/usr/local/bin"
	else
		DEST="${HOME}/.local/bin"
	fi
fi

mkdir -p "${DEST}"
INSTALL_CMD=(install -m 0755 "${TMP}/tray" "${DEST}/tray")
if [[ ! -w "${DEST}" ]]; then
	INSTALL_CMD=(sudo "${INSTALL_CMD[@]}")
fi

"${INSTALL_CMD[@]}"
echo "Installed tray -> ${DEST}/tray"
if ! command -v tray >/dev/null 2>&1; then
	echo "Note: add ${DEST} to your PATH if needed (e.g. export PATH=\"${DEST}:\$PATH\")." >&2
fi
