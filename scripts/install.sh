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

# True if $1 is already on PATH (normalized paths; macOS-friendly).
tray_install_dir_on_path() {
	local dest="$1"
	local dest_abs p dir_abs
	[[ -d "$dest" ]] || return 1
	dest_abs="$(cd "$dest" && pwd -P)" || return 1
	IFS=':' read -r -a __path_parts <<<"${PATH:-}"
	for p in "${__path_parts[@]}"; do
		[[ -z "$p" ]] && continue
		[[ -d "$p" ]] || continue
		dir_abs="$(cd "$p" && pwd -P)" || continue
		if [[ "$dir_abs" == "$dest_abs" ]]; then
			return 0
		fi
	done
	return 1
}

tray_print_path_instructions() {
	local dest="$1"
	{
		echo ""
		echo "--------------------------------------------------------------------------------"
		echo "  tray was installed to: ${dest}/tray"
		echo ""
		echo "  That directory is not on your PATH, so this shell cannot run \"tray\" yet."
		echo ""
		echo "  Fix for this terminal only (copy-paste both lines):"
		echo "    export PATH=\"${dest}:\$PATH\""
		echo "    tray --version"
		echo ""
		echo "  Fix for every new terminal (zsh — default on macOS):"
		echo "    echo 'export PATH=\"${dest}:\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
		echo ""
		echo "  Fix for bash on macOS:"
		echo "    echo 'export PATH=\"${dest}:\$PATH\"' >> ~/.bash_profile && source ~/.bash_profile"
		echo ""
		echo "  Or install to /usr/local/bin instead (may ask for your password):"
		echo "    curl -fsSL https://raw.githubusercontent.com/${REPO}/main/scripts/install.sh | TRAY_INSTALL_DIR=/usr/local/bin bash"
		echo "--------------------------------------------------------------------------------"
		echo ""
	} >&2
}

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
if ! tray_install_dir_on_path "${DEST}"; then
	tray_print_path_instructions "${DEST}"
elif ! command -v tray >/dev/null 2>&1; then
	tray_print_path_instructions "${DEST}"
fi
