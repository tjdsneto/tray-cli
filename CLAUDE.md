# Project instructions (Claude)

This file mirrors [Cursor rules](.cursor/rules/) for Claude Code / CLI context. **Keep it in sync** with `.cursor/rules/*.mdc` when either changes—see *AI instruction parity* below.

---

## Pure functions for unit tests

Unless it would push the design too far from the simplest or clearest approach, **prefer abstractions that isolate pure functions** (deterministic inputs → outputs, no I/O) so behavior can be covered with fast, focused unit tests.

- Put URL/query builders, request body maps, decoders, string normalizers, and validation in **pure helpers**; keep HTTP, filesystem, and cobra wiring in **thin wrappers**.
- **Do not** twist the domain model or add layers of indirection only for purity—when a bit of I/O-bound code is obviously clearer, keep it.

See also: [`docs/maintainers/testing.md`](docs/maintainers/testing.md).

---

## CLI layout & errors

- **Root** lives in [`internal/cli`](internal/cli) (`Execute`, `NewRootCmd`, `requireAuth`, `ConfigDir`, `UserFacingError`).
- **Subcommands** live in [`internal/cli/commands`](internal/cli/commands); `commands.Register` groups them (see comments in `register.go`).
- **Missing server URL/key** uses [`internal/cli/errs`](internal/cli/errs) `MissingBackendConfig` — user-facing text is in `UserFacingError`; **`TRAY_DEBUG=1`** prints the raw error on stderr first (see `cmd/tray/main.go`).
- **Tray resolution** without extra I/O: [`internal/cli/trayref`](internal/cli/trayref).
- **Session:** OAuth stores a refresh token; [`requireAuth`](internal/cli/auth.go) uses [`internal/auth`](internal/auth) `EnsureFreshCredentials` (JWT `exp` + Supabase `grant_type=refresh_token`) before commands; `tray login --token` has no refresh.
- **Distribution:** release tarballs via [`scripts/build-release.sh`](scripts/build-release.sh) / `make release`; publish with [`scripts/publish-release.sh`](scripts/publish-release.sh) / `make publish-release VERSION=v…`; install script [`scripts/install.sh`](scripts/install.sh) — see [`docs/maintainers/distribution.md`](docs/maintainers/distribution.md).
- **Listen hooks (CLI users):** [`docs/user/hooks.md`](docs/user/hooks.md) — `tray listen` + `hooks.json` (events, `TRAY_*` env, recipes).
- **Docs layout:** [`docs/README.md`](docs/README.md) — `docs/user/` vs `docs/maintainers/`.
- **Repository README:** aimed at **CLI users**; **maintainers** → [`docs/maintainers/README.md`](docs/maintainers/README.md) (linked at the bottom of `README.md`).

---

## Local `scratch/` directory

The **`scratch/`** directory is **gitignored**—use it for private brainstorming and notes. Nothing under `scratch/` is committed. Promote ideas into `README.md`, `docs/user/` or `docs/maintainers/`, or code when they should ship.

(See [`.cursor/rules/scratch-local-brainstorm.mdc`](.cursor/rules/scratch-local-brainstorm.mdc).)

---

## End-user agent skills (`tray` CLI)

Help for **using** the shipped `tray` binary (not developing this repo) lives in **[`skills/README.md`](skills/README.md)**. The canonical body is **[`skills/tray-cli/SKILL.md`](skills/tray-cli/SKILL.md)**; the repository symlinks it into **`.cursor/skills/tray-cli/`** and **`.claude/skills/tray-cli/`** so Cursor and Claude Code load it in a normal clone.

---

## AI instruction parity (Cursor ↔ Claude)

| Surface | Location |
|--------|----------|
| **Cursor** | `.cursor/rules/*.mdc` |
| **Claude** | this file (`CLAUDE.md`) |

**When creating, updating, or removing project-wide AI guidance in one place, apply the same change to the other** so Cursor and Claude stay aligned. If guidance is intentionally tool-specific, note that in **both** places in one line.
