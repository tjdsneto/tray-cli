# Maintainers and contributors

This is the home for **working on the repository**: tests, local builds, releases, database migrations, and code layout. **CLI user docs** (install, hooks, day-to-day commands) live in **[`docs/user/`](../user/README.md)** and the [repository **`README.md`**](../../README.md).

## Contents

| Doc | What |
|-----|------|
| **[Distribution](distribution.md)** | Local builds, release tarballs, `publish-release.sh`, versioning, install script behavior |
| **[Unit testing](testing.md)** | `make test`, coverage, what to prioritize in tests |

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
