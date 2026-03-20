# Tray-CLI

CLI-first **tray** (shared inbox-tray / attention queue): **Go** client, **Supabase** backend.

## Dev

```bash
go test ./...
go run ./cmd/tray --help
```

Config directory: `$XDG_CONFIG_HOME/tray` or `~/.config/tray`, or override with `TRAY_CONFIG_DIR`.

Supabase: `TRAY_SUPABASE_URL` (e.g. `https://xxxx.supabase.co`), `TRAY_SUPABASE_ANON_KEY`.

**Login (token flow):** `tray login --token '<access_jwt>'` — validates via `GET /auth/v1/user` and writes `credentials.json` under the config directory.

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
