# Tray-CLI

CLI-first **tray** (shared inbox-tray / attention queue): **Go** client, **Supabase** backend.

**Output & UX notes:** keep a local **`scratch/`** directory (gitignored) for brainstorming—see the **Output** section below and [`docs/testing.md`](docs/testing.md) for what ships in-repo.

## Dev

You need **[Go 1.22+](https://go.dev/dl/)** on your `PATH` (e.g. `brew install go` on macOS). `./run.sh` and `./build.sh` also look under `/opt/homebrew/bin` and `/usr/local/go/bin` if `go` is missing from PATH.

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

**Login:** Run `./run.sh login` to open a **local web page** where you pick Google, GitHub, etc. (enable each in Supabase). Use `./run.sh login --provider google` or **`TRAY_OAUTH_PROVIDER`** in `.env` to skip the picker. If you already have a **valid saved session**, the CLI skips the browser until you run **`./run.sh login --force`**.

**Status:** `./run.sh status` checks credentials and validates the session with Supabase (`--format json` for scripts; exit code **0** if signed in, **1** if not).

During OAuth, the CLI starts a **short-lived local HTTP server** on `127.0.0.1` with a **random free port** (`:0`) so the browser can return the auth code—same for all users, including production installs. That is normal; it is not listening on the network.

**Google / GitHub OAuth apps** (outside Supabase): **Authorized redirect URI** = **`https://<project-ref>.supabase.co/auth/v1/callback`**. Tokens are written to `credentials.json` under the config directory.

**Maintainer builds only:** To embed extra OAuth redirect/callback diagnostics in the binary (for people working on the CLI), set **`TRAY_EMBED_DEV_OAUTH_HINTS=1`** in `.env` when using `./run.sh` / `./build.sh`. Release artifacts for end users should **not** set this. It is not a runtime env var—only affects `-ldflags` at compile time.

**Login (manual token):** `tray login --token '<access_jwt>'` — validates via `GET /auth/v1/user` and writes credentials (no browser).

**Troubleshooting OAuth:** If you see `Unsupported provider: provider is not enabled`, open **Supabase Dashboard → Authentication → Providers**, turn **Google** on, and paste the **Client ID** and **Client secret** from Google Cloud (same OAuth client whose redirect URI is `https://<project-ref>.supabase.co/auth/v1/callback`). Then run `tray login --provider google` again. The CLI cannot enable providers; it must be done in the dashboard.

**Trays:** `./run.sh create <name>` creates a tray; **`rename`**, **`delete-tray`** (owner); **`ls`** lists trays; **`join`** uses an invite token or link; **`invite`** / **`rotate-invite`** or **`invite --rotate`** manage invite tokens (owner); **`members`**, **`revoke`**, **`leave`** for membership.

**Items:** **`./run.sh add "title" <tray>`** adds a pending item (tray = name, id, or **`remote`** alias). **`list`** / **`list <tray>`**, **`contributed`**, **`remove <item-id>`** (owner deletes any item; contributor can delete own **pending** items).

**Triage (tray owner):** **`accept`**, **`decline`** (**`--reason`**), **`snooze`** (**`--until` RFC3339**), **`complete`** (**`--message`**), **`archive`**. Use item ids from **`tray list --format json`**.

**Remote aliases:** **`./run.sh remote add <alias> <invite-url-or-token>`** joins and saves `remotes.json` under your config dir so you can use **`./run.sh add "…" <alias>`** later. **`./run.sh remote ls`** / **`./run.sh remote remove <alias>`** manage that file.

Use the same `./run.sh` pattern so Supabase URL/key are embedded, or export env vars and run `tray` directly.

**Database:** SQL migrations live under [`supabase/migrations/`](supabase/migrations/). Link the repo to your Supabase project and run **`supabase db push`** (or paste SQL in the dashboard) so row-level security matches the CLI. If `create` fails with a policy / recursion error, your remote DB is usually missing a newer migration.

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

### Architecture

- **`internal/domain`** — types (`Tray`, `Item`, `Session`, …) and **service interfaces**: `TrayService`, `ItemService`. The CLI depends on these, not on HTTP paths.
- **`internal/domain.Services`** — bundles `Trays` + `Items` for a single injection point.
- **`internal/adapters/postgrest`** — PostgREST (Supabase Data API) implementations of those services; `postgrest.Dial` / `postgrest.NewServices`. A future Firebase adapter would live alongside as another implementation.
- **`internal/credentials`** — persisted session file (`credentials.json`).
- **`internal/auth`** — Supabase Auth helpers (e.g. fetch current user).
- **`internal/supabase`** — tiny HTTP client (`apikey` + `Authorization` only).
