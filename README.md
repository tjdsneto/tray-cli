# Tray-CLI

CLI-first **tray** (shared inbox-tray / attention queue): **Go** client, **Supabase** backend.

## Dev

```bash
go test ./...
go run ./cmd/tray --help
```

Config directory: `$XDG_CONFIG_HOME/tray` or `~/.config/tray`, or override with `TRAY_CONFIG_DIR`.

Supabase (later commands): `TRAY_SUPABASE_URL`, `TRAY_SUPABASE_ANON_KEY`.
