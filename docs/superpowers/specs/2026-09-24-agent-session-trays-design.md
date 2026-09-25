# Agent session trays (remote inbox) — design

**Date:** 2026-09-24  
**Status:** Approved for planning  
**Related:** phone-trays / chatgpt-handoff skills (directional conventions); GoFlow GF-37 (`--agent-session-id`); agent-registry (session presence)

## Problem

Agents need a **non-interruptive** way to hand work to another agent session. Dial / hotline is a live interrupt. ChatGPT handoff uses skill-level directional trays (`to-chatgpt` / `from-chatgpt`) with title-only packets. Neither is a first-class, session-scoped inbox.

## Goals

- Each AI agent session can have a **remote tray** that acts as its **inbox**.
- Other sessions write items into that tray by **agent session id**.
- The owning session drains when ready (`list` → work → `complete`).
- **Create-on-add**: writing always recreates the tray if it was pruned.
- **Prune** empty session trays after an idle window (safe because add recreates).
- Stamp **who created** an item with a first-class `agent_session_id` (source), aligned with GoFlow.
- Skills + CLAUDE.md / AGENTS.md make sessions aware they may have inbox items; dial/hotline “you’ve got mail” is **optional**, not required.

## Non-goals (v1)

- New cross-user ACL beyond existing remote join / contribute.
- Local-only session trays as the primary path (v1 is **remote**).
- Replacing dial/hotline; this complements them.
- Encoding handoff packets in a new format (title remains free-form; skills may still use packet conventions).

## Approach

**One remote tray per agent session** (Approach A).

| Concept | Choice |
|--------|--------|
| Tray identity | Remote tray keyed by agent session id; ref `agent-session:<id>` |
| Target (inbox) | Which tray you write to — `tray add … --agent-session-id <B>` |
| Source (creator) | Item field `agent_session_id` — who created the item |
| Naming | Match GoFlow GF-37: flag `--agent-session-id`, env `TRAY_AGENT_SESSION_ID`, field `agent_session_id` |
| Discovery | Caller hands over id, or resolve via **agent-registry** `list` (public `session_id`) |
| Interrupt | Optional dial/hotline ping after add — skill recipe only |

Rejected for v1: shared single inbox + filter-only (Approach B/C) — weaker “my tray” model than literal tray-per-session.

## Architecture

```
Agent A                         Remote                          Agent B
───────                         ──────                          ───────
knows B's agent session id
(agent-registry or handoff)
        │
        ├── tray add "…" --agent-session-id <B> --remote
        │     → ensure/recreate agent-session:<B>
        │     → stamp agent_session_id = A's id (env/flag)
        │
        └── [optional] dial/hotline: "item on your tray"
                                                         B: tray list --agent-session-id [--remote]
                                                         B: complete items when done
```

### Ownership and writers

- Requires `tray login` (remote).
- Tray owned by the **user account** that first create-on-add’d it.
- Typical: one human, many agent sessions → one owner, many `agent-session:<id>` trays.
- **Same owner:** any of that user’s sessions can add to any of their session trays.
- **Other users:** existing join / contribute only; no special agent ACL in v1.
- Recreate-after-prune: same-owner recreate is fine. Cross-user must not silently steal ownership — name/key stays under the owning user’s namespace; foreign writers need membership.

### Prune

- Eligible: agent-session tray with **no open items** (empty or all completed).
- Idle window: TBD default (candidate: 7 days since emptied / last activity); same spirit as local `tray prune` for empty dir/branch trays.
- v1: extend CLI `tray prune` (and/or documented GC); server-side auto-GC optional later.
- Protocol remains valid after prune because **add recreates**.

### Hooks

- Existing remote listen events (`item.pending`, etc.) apply to owned session trays.
- Expose source `agent_session_id` to hooks as `TRAY_ITEM_AGENT_SESSION_ID` (do not overload session-identity `TRAY_AGENT_SESSION_ID`).
- B can wake on new mail without dial.

## CLI surface

```text
# Writer A → B's inbox (create-on-add)
tray add "Review the auth PR" --agent-session-id <B-id> --remote

# Optional: override source stamp (default: $TRAY_AGENT_SESSION_ID)
tray add "…" --agent-session-id <B-id> --from-agent-session-id <A-id> --remote

# Owner B drains inbox
tray list --agent-session-id <my-id> --remote
tray list --agent-session-id --remote   # “mine” when TRAY_AGENT_SESSION_ID is set
tray complete <item-id>

# Prune empty agent-session trays (idle)
tray prune …   # extend to cover agent-session kind
```

### Flag / env / field map

| Role | CLI | Env | Stored |
|------|-----|-----|--------|
| Target inbox | `--agent-session-id <id>` | `TRAY_AGENT_SESSION_ID` when listing “mine” | Tray key / name |
| Source creator | `--from-agent-session-id <id>` (optional) | `TRAY_AGENT_SESSION_ID` on add | Item `agent_session_id` |

Do **not** use bare `--session` (collides with auth session). Align vocabulary with GoFlow (`gtr start --agent-session-id`, `$GOFLOW_AGENT_SESSION_ID`).

## Data model

- New remote tray kind or naming convention for session trays (exact DB shape TBD in implementation plan: dedicated kind vs named tray under a reserved prefix).
- New optional column / field on items: `agent_session_id` (text, nullable) — **source** session that created the item.
- List/json output includes `agent_session_id` when set.
- Local items: out of scope for v1 primary path; may stamp the field later for consistency.

## Agent awareness (docs / skills)

- Document in user docs + skill (and CLAUDE.md / AGENTS.md mirrors): every agent session may have a remote inbox at `agent-session:<its-id>`; check periodically or on listen.
- Export `TRAY_AGENT_SESSION_ID` (and optionally GoFlow’s env) at session start — same id as agent-registry registration.
- Optional skill recipe: after `add`, dial/hotline the target with a one-line “item waiting” notice.

## Error handling

- Not signed in → same remote auth errors as today.
- Invalid / empty agent session id → clear validation error.
- Pruned tray + add → recreate, then add (no user-facing “tray missing” for the happy path).
- List “mine” without env or id → prompt to pass `--agent-session-id` or set `TRAY_AGENT_SESSION_ID`.

## Testing

- Pure helpers: tray ref parse (`agent-session:<id>`), create-on-add ensure logic inputs→outputs, prune eligibility (empty + idle).
- CLI: add with `--agent-session-id` stamps source from env; list filters/resolves mine; prune removes empty session trays only.
- Integration: recreate after prune; hooks env includes source id when present.

## Success criteria

- A can file work into B’s session inbox by id without interrupting B.
- B can list and complete that work when ready.
- Empty session trays can be pruned; subsequent add recreates them.
- Source attribution is first-class and GoFlow-aligned.
- Agents are instructed (skill/docs) to know about and check their session tray.

## Locked decisions (implementation plan)

1. **Persistence:** reserved remote tray **name** `agent-session:<id>` — no `kind` column on `trays` in v1.
2. **Prune idle:** **7 days** since last activity (tray `updated_at` / latest item activity), with no open items.
3. **Create-on-add:** creates under the **writer’s** user account (same as `TrayService.Create`). Cross-user writers only add if the tray already exists and they have contribute access — no foreign auto-create under another owner.
4. **Listen:** do **not** default-filter to “mine” in v1 (YAGNI).
