# CLI user documentation

These pages are for **end users** of the `tray` binary: installing it, configuring auth, using commands, and automation (hooks, JSON output).

## Contents

- **[Trays: yours vs joined](trays.md)** — `tray ls`, `tray remote ls`, `tray list`, and `tray contributed`.
- **[Local trays](local-trays.md)** — directory, branch, and global trays on this machine (no sign-in).
- **[Listen hooks (`hooks.json`)](hooks.md)** — `tray listen`, events, `TRAY_*` environment variables, and recipes (notifications, sounds, logging).
- **[macOS menu bar (`tray bar`)](menu-bar.md)** — badge for open/pending items, grouped menu, clipboard handoff to agents (macOS only).

## See also

- **Install, first use, upgrades** — [README.md](../../README.md#install) in the repository root (`curl` installer; [First use](../../README.md#first-use), [Upgrades](../../README.md#upgrades), [Install troubleshooting](../../README.md#install-troubleshooting)).
- **Command overview** — same README (trays, items, triage, remotes) and `tray --help`.
- **Agent skills** — [`skills/README.md`](../../skills/README.md): guidance for AI assistants helping you run `tray` (canonical [`skills/tray-cli/SKILL.md`](../../skills/tray-cli/SKILL.md)).
- **Developing the client** — not required to use the CLI; see [`docs/maintainers/`](../maintainers/README.md) if you hack on this repo.
