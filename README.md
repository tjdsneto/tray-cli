# Tray-CLI

CLI-first **tray** (shared inbox-tray / attention queue): **Go** client, **Supabase** backend.

**Output & UX notes:** keep a local **`scratch/`** directory (gitignored) for brainstorming—see the **Output** section below and [`docs/testing.md`](docs/testing.md) for what ships in-repo.

## Install

**From GitHub Releases** (after maintainers publish `tray_*.tar.gz` assets — see [`docs/distribution.md`](docs/distribution.md)):

```bash
curl -fsSL https://raw.githubusercontent.com/tjdsneto/tray-cli/main/scripts/install.sh | bash
```

**Where it installs:** the script **does not run `sudo` unless you set `TRAY_INSTALL_USE_SUDO=1`**. It picks the first **writable** directory: reuse the path if `tray` is already on `PATH`, else **`/usr/local/bin`** or **`/opt/homebrew/bin`** (macOS) when your user can write there, else **`~/.local/bin`**. That last path is often not on `PATH` on macOS until you add it (the installer prints copy-paste steps). System-wide install without write access: `TRAY_INSTALL_USE_SUDO=1 TRAY_INSTALL_DIR=/usr/local/bin` (one password prompt). Override directory with `TRAY_INSTALL_DIR`.

**Upgrades:** run the same `curl … | bash` line again with default **`TRAY_VERSION=latest`** (the default). It downloads the newest GitHub Release and replaces the binary in the install directory. To stay on a specific version, set `TRAY_VERSION=v0.1.0` (or pin in your docs).

**With Go** (builds from source; needs Go 1.25+ for current Charm deps):

```bash
go install github.com/tjdsneto/tray-cli/cmd/tray@latest
```

More detail: [`docs/distribution.md`](docs/distribution.md) (versioning, `publish-release`, and release tarballs).

**Version in git:** releases are **git tags** (`v1.2.3`); there is no checked-in `VERSION` file. `main` is unreleased until tagged. See **“Where is the current version”** in [`docs/distribution.md`](docs/distribution.md).

## Dev

You need **[Go 1.25+](https://go.dev/dl/)** on your `PATH` (e.g. `brew install go` on macOS). `./run.sh` and `./build.sh` also look under `/opt/homebrew/bin` and `/usr/local/go/bin` if `go` is missing from PATH.

```bash
make test              # or: go test ./... -race -count=1
cp .env.example .env   # then edit with your Supabase URL + anon key
./run.sh --help
```

Or set `TRAY_SUPABASE_URL` and `TRAY_SUPABASE_ANON_KEY` in the environment and use `go run ./cmd/tray` (without embeds unless you pass the same `-ldflags` as in [`run.sh`](run.sh)).

**Testing:** [`docs/testing.md`](docs/testing.md) — `make test`, coverage reports.

**AI coding rules:** [`CLAUDE.md`](CLAUDE.md) and [`.cursor/rules/`](.cursor/rules/) — keep them aligned when you change either.

**Release-style binary:** `./build.sh` writes `./tray` with Supabase settings embedded from `.env` (or from already-exported env vars). CI can set the same variables and invoke `go build -ldflags "..."` the same way.

Config directory (see `internal/config/paths.go`): override with **`TRAY_CONFIG_DIR`**; otherwise **Windows** uses `%APPDATA%\tray`, **macOS/Linux** use `$XDG_CONFIG_HOME/tray` if set, else `~/.config/tray`.

**Supabase:** `TRAY_SUPABASE_URL` (e.g. `https://xxxx.supabase.co`), `TRAY_SUPABASE_ANON_KEY`. At runtime, **environment variables override** values embedded at build time. See [`.env.example`](.env.example).

**Login:** Run `./run.sh login` to open a **local web page** where you pick Google, GitHub, etc. (enable each in Supabase). Use `./run.sh login --provider google` or **`TRAY_OAUTH_PROVIDER`** in `.env` to skip the picker. If you already have a **valid saved session**, the CLI skips the browser until you run **`./run.sh login --force`**. After **OAuth**, the CLI **refreshes the access JWT** using the stored refresh token when it is expired or near expiry (until Supabase invalidates the refresh token). **`tray login --token`** stores only an access token—use OAuth for automatic refresh.

**Status:** `./run.sh status` checks credentials and validates the session with Supabase (`--format json` for scripts; exit code **0** if signed in, **1** if not).

During OAuth, the CLI starts a **short-lived local HTTP server** on `127.0.0.1` with a **random free port** (`:0`) so the browser can return the auth code—same for all users, including production installs. That is normal; it is not listening on the network.

**Google / GitHub OAuth apps** (outside Supabase): **Authorized redirect URI** = **`https://<project-ref>.supabase.co/auth/v1/callback`**. Tokens are written to `credentials.json` under the config directory.

**Maintainer builds only:** To embed extra OAuth redirect/callback diagnostics in the binary (for people working on the CLI), set **`TRAY_EMBED_DEV_OAUTH_HINTS=1`** in `.env` when using `./run.sh` / `./build.sh`. Release artifacts for end users should **not** set this. It is not a runtime env var—only affects `-ldflags` at compile time.

**Login (manual token):** `tray login --token '<access_jwt>'` — validates via `GET /auth/v1/user` and writes credentials (no browser).

**Troubleshooting OAuth:** If you see `Unsupported provider: provider is not enabled`, open **Supabase Dashboard → Authentication → Providers**, turn **Google** on, and paste the **Client ID** and **Client secret** from Google Cloud (same OAuth client whose redirect URI is `https://<project-ref>.supabase.co/auth/v1/callback`). Then run `tray login --provider google` again. The CLI cannot enable providers; it must be done in the dashboard.

**Trays:** `./run.sh create <name>` creates a tray; **`rename <tray> <new-name>`** sets the tray name on the server (owner only); **`delete-tray`** (owner); **`ls`** lists trays; **`join <token-or-url> [local-alias]`** joins via invite and optionally saves a local short name; **`invite`** / **`rotate-invite`** or **`invite --rotate`** manage invite tokens (owner); **`members`**, **`revoke`**, **`leave`** for membership.

**Items:** **`./run.sh add "title" <tray>`** adds a pending item (tray = name, id, or **`remote`** alias). **`list`** / **`list <tray>`**, **`contributed`**, **`remove <item-id>`** (owner deletes any item; contributor can delete own **pending** items).

**Triage (tray owner):** **`accept`**, **`decline`** (**`--reason`**), **`snooze`** (**`--until` RFC3339**), **`complete`** (**`--message`**), **`archive`**. Use item ids from **`tray list --format json`**.

**Remote aliases:** **`join … <alias>`** or **`./run.sh remote add <alias> <invite-url-or-token>`** saves `remotes.json`. **`./run.sh remote rename <current> <new>`** renames an existing local alias, or sets a first alias when `<current>` is a tray name/id from **`tray ls`** (after you already joined). **`./run.sh remote ls`** / **`./run.sh remote remove <alias>`** manage that file.

Use the same `./run.sh` pattern so Supabase URL/key are embedded, or export env vars and run `tray` directly.

**Database:** SQL migrations live under [`supabase/migrations/`](supabase/migrations/). Link the repo to your Supabase project and run **`supabase db push`** (or paste SQL in the dashboard) so row-level security matches the CLI. If `create` fails with a policy / recursion error, your remote DB is usually missing a newer migration. Item list **By** resolves contributor **name or email** from the `profiles` migration when it is applied; otherwise you see a short id. **`items`** rows include **`accepted_at`**, **`declined_at`**, **`completed_at`**, **`archived_at`**, and **`snoozed_at`** (set by the database when status changes; see migrations).

**Verbose API errors:** set **`TRAY_DEBUG=1`** when running `tray` to print full PostgREST response bodies. By default, errors are shortened for end users.

### Output (list-style commands)

**Default is human-friendly:** tables, local dates, and contextual hints where we’ve added them.

| Flag | Purpose |
|------|---------|
| **`--format human`** (default) | Friendly tables and hints — what you want for day-to-day use |
| **`--format json`**, **`--format machine`**, or **`--json`** | Stable, machine-readable JSON for scripts and automation |
| **`--format markdown`** / **`md`** | Markdown tables — easy to paste into chats and for models to read |

**Deprecated but still works:** `-o` / `--output` (same values as `--format`). Prefer **`--format`**.

`--json` is shorthand for `--format json` and must not be combined with another explicit format (e.g. `--format markdown`).

For **trays**, the default **human** output shows **name**, **item count**, and **created** (in your **local timezone**; set **`TZ`** if needed). An **empty** list suggests **`tray create <name>`**; after **`create`**, human and markdown print a short **“Created tray …”** line before the table (JSON is data-only). After `create` or `ls` with rows, the CLI prints **next-step hints** (`tray add …`, `tray invite …`). Tray **IDs** (UUIDs) and **`item_count`** appear in **`--format json`** / **`--json`**.

For **items** (`list`, `review`, …), human output includes **who added** the item (`you` vs a short id), **created** as a **relative time** when recent (e.g. `20 minutes ago`), and **status** colors on a TTY. Set **`NO_COLOR=1`** to disable ANSI colors. Use **`tray triage`** for an interactive pending queue (TTY); **`tray review`** stays a non-interactive list.

### Architecture

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
