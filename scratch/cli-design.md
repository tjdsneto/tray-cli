# CLI Design

Specification for **Tray CLI** (official tool name; repo: attention-queue): commands, flags, remotes (aliases), output format, and local storage. **Stack:** Go implementation talking to **Supabase** (see [backend-spec.md](backend-spec.md)). Binary on `$PATH`: **`tray`**. For MVP scope and launch checklist see [launch-requirements.md](launch-requirements.md); for flows and scenarios see [invite-flow-and-scenarios.md](invite-flow-and-scenarios.md).

---

## 1. Command set

### Auth

| Command | Purpose |
|---------|---------|
| `tray login` | Authenticate (browser or device flow); store token for subsequent requests. |

### Trays (owner)

| Command | Purpose |
|---------|---------|
| `tray create <name>` | Create a named tray. |
| `tray ls` | List my trays. |
| `tray invite <tray>` | Show or generate shareable invite link/token for the tray (Model B). |
| `tray rotate-invite <tray>` | Rotate invite token; old link stops working, new link printed. |
| `tray members <tray>` | List members (and optionally pending invites). |
| `tray revoke <tray> <user>` | Remove a member from the tray. |

### Remotes (aliases for others’ trays)

Members refer to “someone else’s tray” by a **local alias** (remote), not by tray name or ID, to avoid ambiguity. See [invite-flow-and-scenarios.md](invite-flow-and-scenarios.md#identifying-buckets-remotes-aliases-vs-owner-provided-names) (anchor may still say “buckets” until that doc is updated).

| Command | Purpose |
|---------|---------|
| `tray remote add <alias> <invite-url-or-token>` | Add a remote: alias → tray (resolved via invite). |
| `tray remote ls` | List remotes (alias → tray/owner summary). |
| `tray remote remove <alias>` | Remove a remote. |
| `tray join <url-or-token>` | Join a tray via shareable invite; optionally prompt for alias and create remote. |

### Items

| Command | Purpose |
|---------|---------|
| `tray add "title" [tray-or-alias]` | Add item. Owner uses tray name; member uses remote alias. Optional: `--due <date>`. |
| `tray list [tray]` | List items (default: active only). Owner uses tray name; no tray = all trays. |
| `tray review [tray]` | Interactive triage (optional). |

### Triage (item status)

| Command | Purpose |
|---------|---------|
| `tray accept <id>` | Set status to `accepted`. |
| `tray decline <id> [--reason "..."]` | Set status to `declined`; optional reason for the member who added it. |
| `tray snooze <id> [--until <date>]` | Set status to `snoozed` until optional date. |
| `tray complete <id> [--message "..."]` | Set status to `completed`; optional completion message. |
| `tray archive <id>` | Set status to `archived`. |

### Member view

| Command | Purpose |
|---------|---------|
| `tray contributed` | List items I added to others’ trays, with status (pending, accepted, completed, declined, etc.). |

### Invite (Model A — optional / post-MVP)

| Command | Purpose |
|---------|---------|
| `tray invite-email <email> [tray]` | Invite by email (sends link); creates pending invite. Distinct from `tray invite <tray>` (shareable link). |

### Listen (post-MVP)

| Command | Purpose |
|---------|---------|
| `tray listen [tray]` | Watch for new/updated items (poll or realtime). Flags: `--interval`, `--json`, `--exec <cmd>`. |

---

## 2. Flags and output

### JSON output

- All **read** commands support `--json` for machine-readable output (e.g. `tray list --json`, `tray contributed --json`, `tray ls --json`, `tray remote ls --json`).
- Schema must be stable and documented so agents and scripts can parse it.

### List order (queue vs stack)

- **Default:** FIFO (oldest first) — queue semantics.
- **Option:** `--newest-first` (or config `list.order: newest_first`) for stack-style view. Same data; only presentation order changes.

### List scope

- **Default:** Active items only (`pending`, `accepted`, `snoozed` if past `snooze_until`). Exclude `declined`, `completed`, `archived`.
- **Option:** e.g. `--include archived` or `--status completed` to include or filter by status.

### Due date

- `tray add "..." <tray-or-alias> --due <date>` to set optional due date.
- `tray list [tray] --due-before <date>` to filter by due date (optional).

---

## 3. Local storage (local “DB”)

The CLI needs **local state** for auth, remotes, and config. No separate local database is required; file-based storage is enough.

### Config directory

- **Default:** `~/.config/tray/` (or `$XDG_CONFIG_HOME/tray/` when set).
- **Override:** e.g. `$TRAY_CONFIG_DIR` or `--config-dir` for tests/portability.

### Files (under config dir)

| File | Purpose |
|------|---------|
| `credentials` (or keychain) | Stored auth token (e.g. JWT or refresh token). Should not be world-readable. |
| `remotes` | Map alias → tray_id or invite_token (or URL). Used to resolve `tray add "..." <alias>`. Format: e.g. JSON or simple `alias=tray_id` lines. |
| `config` (optional) | User preferences: e.g. `list.order`, `list.default_tray`, default `--interval` for listen. |

### Remotes resolution

- On `tray add "..." <alias>`: CLI reads `remotes`, looks up alias → tray_id (or invite_token). If token, resolve once via API to get tray_id, then use tray_id for the add request. Prefer storing tray_id after first resolution so the CLI does not depend on invite token long-term.
- Remotes are **local only**; not synced to the server (no “remotes” table on backend required for MVP).

### No local DB engine

- **No SQLite or similar** for MVP. Plain files (JSON or line-based) for remotes and config are sufficient. If we later add offline queue or richer local cache, we can introduce a local DB then.

---

## 4. Agent-friendly

- **Stable commands and flags** so agents can rely on them.
- **`--json`** on all read commands; **documented schema** (field names, types).
- **Non-interactive by default** (no prompts unless necessary, e.g. `tray join` may prompt for alias).
- Optional **MCP server** (post-MVP) that wraps the same API for Claude/Cursor.

---

## 5. Distribution (summary)

- **Single Go binary** per OS/arch; no runtime on user machine.
- **GitHub Releases** for binaries; **install script** (one-liner) that detects OS/arch and installs to `$PATH`.
- **Homebrew tap** for macOS/Linux; optional **Scoop** for Windows.
- Details: [research-and-stack.md](research-and-stack.md#cli-distribution).

---

## 6. References

| Doc | Content |
|-----|---------|
| [launch-requirements.md](launch-requirements.md) | MVP scope, CLI checklist, done criteria. |
| [invite-flow-and-scenarios.md](invite-flow-and-scenarios.md) | Invite models, remotes rationale, item lifecycle, scenarios. |
| [research-and-stack.md](research-and-stack.md) | Stack choice, listen/watch, queue vs stack, distribution. |
| [repo-visibility-and-distribution.md](repo-visibility-and-distribution.md) | Public vs private repo; GitHub Releases, Pages, install scripts. |
| [backend-spec.md](backend-spec.md) | API, auth, DB structure; backend contract for the CLI. |
| [naming-ideas.md](naming-ideas.md) | Product name Tray, Tray.io conflict, API vocabulary. |

---

*Last updated: 2026-03-20 (Tray CLI + Go + Supabase)*
