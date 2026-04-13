# Maintainers and contributors

This is the home for **working on the repository**: tests, local builds, releases, database migrations, and code layout. **CLI user docs** (install, hooks, day-to-day commands) live in **[`docs/user/`](../user/README.md)** and the [repository **`README.md`**](../../README.md).

## Contents

| Doc | What |
|-----|------|
| **[Distribution](distribution.md)** | Local builds, release tarballs, `publish-release.sh`, versioning, install script behavior |
| **[Unit testing](testing.md)** | `make test`, coverage, what to prioritize in tests |
| **Configuration and runtime** (below) | `TRAY_CONFIG_DIR`, Supabase env overrides — for development and custom builds, not normal installs |
| **Login and OAuth** (below) | Browser flow, Supabase, JWT refresh, providers, redirect URIs — not required reading for normal CLI users |
| **Debugging** (below) | **`TRAY_DEBUG`** — verbose API errors when developing or diagnosing failures |

## Development setup

You need **[Go 1.25+](https://go.dev/dl/)** on your `PATH` (e.g. `brew install go` on macOS). `./run.sh` and `./build.sh` also look under `/opt/homebrew/bin` and `/usr/local/go/bin` if `go` is missing from `PATH`.

```bash
make test              # or: go test ./... -race -count=1
cp .env.example .env   # then edit with your Supabase URL + anon key
./run.sh --help
```

Or set `TRAY_SUPABASE_URL` and `TRAY_SUPABASE_ANON_KEY` in the environment and use `go run ./cmd/tray` (without embeds unless you pass the same `-ldflags` as in [`run.sh`](../../run.sh)).

**Testing:** [`testing.md`](testing.md) — `make test`, coverage reports.

**AI coding rules:** [`CLAUDE.md`](../../CLAUDE.md) and [`.cursor/rules/`](../../.cursor/rules/) — keep them aligned when you change either.

**Release-style binary:** `./build.sh` writes `./tray` with Supabase settings embedded from `.env` (or from already-exported env vars). CI can set the same variables and invoke `go build -ldflags "..."` the same way.

**Maintainer-only compile flag:** To embed extra OAuth redirect/callback diagnostics in the binary, set **`TRAY_EMBED_DEV_OAUTH_HINTS=1`** in `.env` when using `./run.sh` / `./build.sh`. Release artifacts for end users should **not** set this.

**Local notes:** keep a **`scratch/`** directory (gitignored) for brainstorming—see [`.cursor/rules/scratch-local-brainstorm.mdc`](../../.cursor/rules/scratch-local-brainstorm.mdc).

## Configuration and runtime (not needed for normal installs)

End users of a **release binary** from the install script do **not** need to set backend variables; defaults are embedded at build time.

**Config directory:** override with **`TRAY_CONFIG_DIR`**; otherwise **Windows** uses **`%APPDATA%\tray`**, **macOS/Linux** use **`$XDG_CONFIG_HOME/tray`** if set, else **`~/.config/tray`** (see `internal/config/paths.go`).

**Supabase overrides:** **`TRAY_SUPABASE_URL`**, **`TRAY_SUPABASE_ANON_KEY`** — environment variables override values embedded at link time (development, self-hosted backend, or custom builds). See **[`.env.example`](../../.env.example)**.

## Login and OAuth (technical)

This is background for **developing** the client or **self-hosting** auth—not needed for people who only install a release binary and run **`tray login`**.

**Flow:** **`tray login`** opens a **local web page** (default sign-in is often **Google**; enable the provider in **Supabase**). **`tray login --provider <id>`** or **`TRAY_OAUTH_PROVIDER`** in `.env` skips the picker and uses a specific provider your project has enabled. If a **saved session** is still valid, the CLI skips the browser until **`tray login --force`**. After OAuth, the CLI **refreshes the access JWT** using the stored refresh token when it is expired or near expiry. **`tray login --token '<jwt>'`** stores only an access token (no refresh)—prefer OAuth for long-term use.

**Status:** **`tray status`** validates the session with Supabase (**`--format json`** for scripts; exit **0** if signed in, **1** if not).

**Local callback:** during OAuth the CLI starts a **short-lived HTTP server** on **`127.0.0.1`** on a **random port** so the browser can return the auth code; it is not meant to be reachable on the network.

**Redirect URI (Google / GitHub OAuth apps outside Supabase):** **`https://<project-ref>.supabase.co/auth/v1/callback`**. Session tokens are written to **`credentials.json`** under the [config directory](#configuration-and-runtime-not-needed-for-normal-installs).

**Manual token:** **`tray login --token`** validates via **`GET /auth/v1/user`** and writes credentials (no browser).

**Troubleshooting:** **`Unsupported provider: provider is not enabled`** — in **Supabase Dashboard → Authentication → Providers**, enable the provider and set **Client ID** / **Client secret**. The CLI cannot enable providers from the client.

## Debugging (CLI)

**`TRAY_DEBUG=1`** prints full PostgREST response bodies when something fails. By default, errors are shortened for readability.

## Database and migrations

SQL migrations live under [`supabase/migrations/`](../../supabase/migrations/). Link the repo to your Supabase project and run **`supabase db push`** (or paste SQL in the dashboard) so row-level security matches the CLI. If `create` fails with a policy / recursion error, your remote DB is usually missing a newer migration. Item list **By** resolves contributor **name or email** from the `profiles` migration when it is applied. **`items`** rows include **`accepted_at`**, **`declined_at`**, **`completed_at`**, **`archived_at`**, and **`snoozed_at`** (set by the database when status changes; see migrations).

## Architecture (code layout)

- **`internal/domain`** — types (`Tray`, `Item`, `Session`, …) and **service interfaces**: `TrayService`, `ItemService`. The CLI depends on these, not on HTTP paths.
- **`internal/domain.Services`** — bundles `Trays` + `Items` for a single injection point.
- **`internal/adapters/postgrest/pghttp`** — generic JSON REST client + HTTP status → user-facing errors (no domain types).
- **`internal/adapters/postgrest`** — PostgREST / Supabase Data API adapter: `TrayService` / `ItemService` via `pghttp`; row types and mapping in **`item.go`**, **`tray.go`**, **`member.go`**; URL helpers live next to **`item_service.go`** / **`tray_service.go`**.
- **`internal/timex`** — small time helpers (e.g. RFC3339 parsing for JSON timestamps), shared where useful outside the adapter.
- **`internal/cli/commands`** — Cobra subcommands (grouped in `register.go`); wired from `internal/cli` with `commands.Deps`.
- **`internal/cli/trayref`** — pure tray name / id / alias resolution.
- **`internal/remotesfile`** — `remotes.json` load/save for local tray aliases.
- **`internal/credentials`** — persisted session file (`credentials.json`).
- **`internal/auth`** — OAuth and user session helpers for the configured auth server.
- **`internal/supabase`** — tiny HTTP client (`apikey` + `Authorization` only).

## See also

- **[`README.md` § Maintainers and contributors](../../README.md#maintainers-and-contributors)** — short pointer from the repo root.
- **[`docs/README.md`](../README.md)** — index of `docs/user/` vs `docs/maintainers/`.
