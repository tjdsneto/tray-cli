# Project instructions (Codex)

This file mirrors [Cursor rules](.cursor/rules/) for Codex / CLI context. **Keep it in sync** with `.cursor/rules/*.mdc` when either changes—see *AI instruction parity* below.

For full CLI layout, testing, docs, and skill pointers, see [`CLAUDE.md`](CLAUDE.md) (same guidance; Claude-oriented install wording).

---

## End-user agent skills (`tray` CLI)

Canonical skill: [`skills/tray-cli/SKILL.md`](skills/tray-cli/SKILL.md).

- **Agent session trays:** export `TRAY_AGENT_SESSION_ID` (same as agent-registry); check inbox with `tray list --agent-session-id --remote`; hand off with `tray add "…" --agent-session-id <other> --remote`; optional dial ping; `tray prune --remote` drops idle empty session trays (add recreates). Details: [`skills/tray-cli/SKILL.md`](skills/tray-cli/SKILL.md), [`docs/user/trays.md`](docs/user/trays.md#agent-session-trays).

---

## AI instruction parity (Cursor ↔ Codex)

| Surface | Location |
|--------|----------|
| **Cursor** | `.cursor/rules/*.mdc` |
| **Codex** | this file (`AGENTS.md`) |

**When creating, updating, or removing project-wide AI guidance in one place, apply the same change to the other** so Cursor and Codex stay aligned. If guidance is intentionally tool-specific, note that in **both** places in one line.
