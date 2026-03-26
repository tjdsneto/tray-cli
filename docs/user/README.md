# CLI user documentation

These pages are for **end users** of the `tray` binary: installing it, configuring auth, using commands, and automation (hooks, JSON output).

## Contents

- **[Trays: yours vs joined](trays.md)** — `tray ls`, `tray remote ls`, `tray list`, and `tray contributed`.
- **[Listen hooks (`hooks.json`)](hooks.md)** — `tray listen`, events, `TRAY_*` environment variables, and recipes (notifications, sounds, logging).

## See also

- **Install and upgrades** — [README.md § Install](../../README.md#install) in the repository root (curl installer, `go install`, config directory).
- **Command overview** — same README (trays, items, triage, remotes) and `tray --help`.
- **Developing the client** — not required to use the CLI; see [`docs/maintainers/`](../maintainers/README.md) if you hack on this repo.
