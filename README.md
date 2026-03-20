# Tray-CLI

CLI-first **tray** (shared inbox-tray / attention queue): **Go** client, **Supabase** backend.

## Dev

```bash
go test ./...
cp .env.example .env   # then edit with your Supabase URL + anon key
./run.sh --help
```

Or set `TRAY_SUPABASE_URL` and `TRAY_SUPABASE_ANON_KEY` in the environment and use `go run ./cmd/tray` (without embeds unless you pass the same `-ldflags` as in [`run.sh`](run.sh)).

**Release-style binary:** `./build.sh` writes `./tray` with Supabase settings embedded from `.env` (or from already-exported env vars). CI can set the same variables and invoke `go build -ldflags "..."` the same way.

Config directory (see `internal/config/paths.go`): override with **`TRAY_CONFIG_DIR`**; otherwise **Windows** uses `%APPDATA%\tray`, **macOS/Linux** use `$XDG_CONFIG_HOME/tray` if set, else `~/.config/tray`.

**Supabase:** `TRAY_SUPABASE_URL` (e.g. `https://xxxx.supabase.co`), `TRAY_SUPABASE_ANON_KEY`. At runtime, **environment variables override** values embedded at build time. See [`.env.example`](.env.example).

**Login (token flow):** `tray login --token '<access_jwt>'` — validates via `GET /auth/v1/user` and writes `credentials.json` under the config directory.

**Trays:** `./run.sh create <name>` creates a tray; `./run.sh ls` lists trays you can access (owned and joined). Use the same `./run.sh` pattern so Supabase URL/key are embedded, or export env vars and run `tray` directly.

### Output (list-style commands)

| Flag | Purpose |
|------|---------|
| `-o table` (default) | Columnar text for humans |
| `-o json` or `--json` | Stable JSON for scripts / tools |
| `-o markdown` / `-o md` | Markdown tables — easy to paste into chats and for models to read |

`--json` is shorthand for `-o json` and cannot be combined with another `-o` value.

### Architecture

- **`internal/domain`** — types (`Tray`, `Item`, `Session`, …) and **service interfaces**: `TrayService`, `ItemService`. The CLI depends on these, not on HTTP paths.
- **`internal/domain.Services`** — bundles `Trays` + `Items` for a single injection point.
- **`internal/adapters/postgrest`** — PostgREST (Supabase Data API) implementations of those services; `postgrest.Dial` / `postgrest.NewServices`. A future Firebase adapter would live alongside as another implementation.
- **`internal/credentials`** — persisted session file (`credentials.json`).
- **`internal/auth`** — Supabase Auth helpers (e.g. fetch current user).
- **`internal/supabase`** — tiny HTTP client (`apikey` + `Authorization` only).
