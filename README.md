# Tray-CLI

CLI-first **tray** (shared inbox-tray / attention queue): **Go** client, **Supabase** backend.

## Dev

```bash
go test ./...
go run ./cmd/tray --help
```

Config directory: `$XDG_CONFIG_HOME/tray` or `~/.config/tray`, or override with `TRAY_CONFIG_DIR`.

Supabase (later commands): `TRAY_SUPABASE_URL`, `TRAY_SUPABASE_ANON_KEY`.

### Output (list-style commands)

| Flag | Purpose |
|------|---------|
| `-o table` (default) | Columnar text for humans |
| `-o json` or `--json` | Stable JSON for scripts / tools |
| `-o markdown` / `-o md` | Markdown tables — easy to paste into chats and for models to read |

`--json` is shorthand for `-o json` and cannot be combined with another `-o` value.

### Architecture

- **`internal/backend`** — `Backend` interface + domain types (`Tray`, `Item`, `Session`, …). Commands depend on this, not on PostgREST paths.
- **`internal/backend` (Supabase implementation)** — `NewSupabase` / `DialSupabase` map those operations to the Supabase Data API. Other backends could implement the same interface later.
- **`internal/supabase`** — low-level HTTP client (`apikey` + `Authorization` headers only).
