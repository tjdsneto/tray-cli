# Distribution

This project ships a single static binary (`tray`) built from [`cmd/tray`](../cmd/tray).

## Building locally

From the repo root (embeds `TRAY_SUPABASE_*` from `.env` when present):

```bash
./build.sh
```

Cross-compile release tarballs (macOS + Linux, `amd64` + `arm64`):

```bash
./scripts/build-release.sh
# artifacts: dist/tray_<os>_<arch>.tar.gz
```

Or use Make:

```bash
make release
```

## Versioning

- **Policy:** [Semantic versioning](https://semver.org/) with a `v` prefix on git tags: `vMAJOR.MINOR.PATCH` (optional prerelease: `v1.0.0-rc.1`).
- **Source of truth:** annotated **git tags** on `main` (or your release branch). The same tag string is embedded in release binaries (`tray --version`).
- **Dev builds:** `go run` / `./run.sh` without `TRAY_RELEASE_*` show `dev` (see [`internal/config/version.go`](../internal/config/version.go)).

### Where is the “current version” in git?

There is **no** `VERSION` file checked into the repo. **`main` at HEAD is not a numbered release** until you cut a tag.

| What you care about | Where it lives |
|---------------------|----------------|
| Latest **released** version | **Git tag** (e.g. `v1.2.3`) and **GitHub Release** with that name |
| What users download | Release assets + **install script** (`TRAY_VERSION=v1.2.3` or `latest`) |
| What `tray --version` shows | **Release binary:** tag + short commit (embedded at link time). **Dev build:** `dev` |

To see the nearest tag on a clone: `git describe --tags --always`.

### First public release (before 1.0)

Use **`v0.x.y`** tags — while major is **0**, semver treats the API as unstable. A common first tag is **`v0.1.0`** (“start of the 0.1 line”). **`v0.0.1`** is valid but often reads oddly as a first tag; prefer **`v0.1.0`** unless you really want a three-part patch series from day one.

To say “beta” **in the version string**, add a [prerelease](https://semver.org/#spec-item-9) suffix, e.g. **`v0.1.0-beta.1`** or **`v0.1.0-rc.1`**. `publish-release.sh` accepts these. On GitHub you can mark that GitHub Release as **Pre-release**; when you later tag **`v0.1.0`** (no suffix), that becomes the stable **0.1** release for `latest` / default installs.

## Publishing (GitHub Releases)

Recommended: one command from a **clean** working tree (runs tests, builds tarballs with embedded version, creates tag, pushes, uploads):

```bash
./scripts/publish-release.sh v0.1.0
```

Or:

```bash
make publish-release VERSION=v0.1.0
```

Requirements:

- `origin` remote configured (for `git push` / tag checks).
- Optional: [GitHub CLI](https://cli.github.com/) (`gh`) for `gh release create` + asset upload. Without `gh`, upload the four `dist/tray_*.tar.gz` files manually.

Manual equivalent:

1. `TRAY_RELEASE_VERSION=v0.1.0 TRAY_RELEASE_COMMIT=$(git rev-parse --short HEAD) ./scripts/build-release.sh`
2. `git tag -a v0.1.0 -m "Release v0.1.0" && git push origin v0.1.0`
3. Upload **these four** assets (same **filenames** every release so install URLs stay stable):
   - `tray_darwin_amd64.tar.gz`
   - `tray_darwin_arm64.tar.gz`
   - `tray_linux_amd64.tar.gz`
   - `tray_linux_arm64.tar.gz`

Each tarball contains a single `tray` binary at the archive root.

Runtime config still uses `TRAY_SUPABASE_URL` and `TRAY_SUPABASE_ANON_KEY` (or values embedded at build time). Public installs usually rely on **environment variables**, not embeds.

## One-line install (end users)

After releases exist on GitHub:

```bash
curl -fsSL https://raw.githubusercontent.com/tjdsneto/tray-cli/main/scripts/install.sh | bash
```

Pinned version:

```bash
TRAY_VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/tjdsneto/tray-cli/main/scripts/install.sh | bash
```

Override install location:

```bash
TRAY_INSTALL_DIR="$HOME/bin" curl -fsSL https://raw.githubusercontent.com/tjdsneto/tray-cli/main/scripts/install.sh | bash
```

## Install with Go (developers)

If you have Go 1.22+ and do not need release tarballs:

```bash
go install github.com/tjdsneto/tray-cli/cmd/tray@latest
```

Ensure `$(go env GOPATH)/bin` is on your `PATH`.

## Windows

The install script targets macOS/Linux shells. On Windows, use WSL, or ship a separate installer (Scoop/Chocolatey) later.
