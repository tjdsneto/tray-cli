# macOS menu bar (`tray bar`) — design

**Date:** 2026-09-24  
**Status:** Approved for planning  
**Platform (v1):** macOS only

## Problem

Tray is CLI-first. Users want a persistent **menu bar** presence: see how many items need attention, browse them by tray, and hand an item to an agent session without leaving the desktop chrome.

## Goals

- Show a **badge** for items that need attention.
- On click, show a **menu** of those items **grouped by tray** (local and remote together).
- On row click, **copy `tray-ref item-id`** to the clipboard for paste into an agent session.
- Agent resolves via existing CLI: fetch → **accept** (if remote pending) → work → **complete**.
- Teach that flow in the **existing** `tray-cli` skill (no new skill).

## Non-goals (v1)

- Linux / Windows system tray.
- Full UI to browse all trays and all item statuses.
- In-menu accept / decline / complete / snooze.
- Merging menu bar into `tray listen`.
- LaunchAgent / login-item installer.
- A separate resolve skill or marketplace package.

## Product behavior

### Badge

Count = **local open** + **remote pending** (owned trays).

- Accepting a remote pending item **decrements the badge** on the next refresh (before complete).
- Zero: show icon without a numeric badge (hide `0`).

### Menu

- Rows = same set as the badge (attention backlog only).
- **Grouped by tray name**; local and remote share grouping. If the same name exists in both scopes, disambiguate with a small `local` / `remote` label.
- Tray sections sorted A–Z; within a tray, **newest-first**.
- **After accept:** row **leaves** the menu on refresh. Accepted-but-incomplete work is not shown in the bar.
- Footer: **Refresh**, **Quit**. No other chrome in v1.

### Clipboard

Exactly two tokens: `<tray-ref> <item-id>` (e.g. `inbox a1b2c3d4`).

- Agent (via `tray-cli` skill) fetches any further detail with the CLI.
- No title, JSON, or ready-made `tray complete` command on the clipboard.

### Auth

- Signed out: local open items only.
- Signed in: add remote pending on owned trays.
- Remote fetch failure: keep last good snapshot; show a muted “remote unavailable” line when useful.

## Architecture

Long-running command **`tray bar`**, separate from `tray listen`.

| Layer | Responsibility |
|--------|----------------|
| Aggregator | Pure snapshot: collect local open + remote pending, group by tray, badge count, clipboard strings. Unit-tested. |
| Refresh loop | Timer (flag `--interval`); optional FS watch on local tray index. Poll remote when session valid. Reuse existing list/review-style APIs—do not share process with listen. |
| Menu bar UI | Darwin + CGO/systray: icon, badge, menu, clipboard on click. Other GOOS: clear “macOS only” error. |
| Daemon mode | Optional `--daemon`: pid + log under config dir (same pattern as `listen --daemon`). |

```
┌─────────────┐     refresh      ┌──────────────┐
│  tray bar   │◄────────────────►│  Aggregator  │
│  (systray)  │                  │  (pure)      │
└──────┬──────┘                  └──────┬───────┘
       │ copy tray id                   │
       ▼                                ▼
  clipboard                      local FS + remote API
```

## Agent surface (`tray-cli` skill)

Extend `skills/tray-cli/SKILL.md` (and symlinks):

- Document `tray bar` and the clipboard handoff.
- When user pastes `tray-ref item-id` (or points at a menu-bar handoff): fetch → if remote **pending**, **`tray accept`** first → do work → **`tray complete`**.
- Local open: skip accept.
- Already accepted: skip accept; complete when done.
- Do not invent status; surface CLI errors.

Human docs: short `docs/user/` page (or section) for `tray bar`; link from user README. Maintainer note in distribution: darwin binary / CGO for systray.

## CLI

```text
tray bar [--interval duration] [--daemon]
```

- Default foreground (menu bar in this process).
- `--daemon`: background with pid/log under config dir.
- Non-darwin: fail fast with a friendly message.

## Errors & reliability

- Aggregator/API errors must not tear down the icon; log and retain last snapshot.
- Clipboard failure: log; do not crash.
- Single-instance: pid file when daemonized (mirrors listen).

## Testing

- **Unit:** aggregator grouping, badge math, exclusion of accepted/completed, clipboard format, name-collision labels.
- **UI:** no heavy automated coverage required in v1; manual check on macOS.

## Future (explicitly deferred)

- Cross-platform tray icons.
- Rich window listing all trays/items.
- Actionable menu (accept/complete in-bar).
- Optional integration or shared event bus with `tray listen`.
- Login-item / LaunchAgent helper.

## Decisions log

| Topic | Choice |
|--------|--------|
| Primary UX | Badge + dropdown list |
| Item scope | Local open + remote pending, grouped by tray |
| Row action | Copy `tray id` for agent |
| Accept vs badge | Accept decrements badge; row leaves menu |
| Skill | Extend `tray-cli` only |
| Platform | macOS v1 |
| Process | New `tray bar`, not inside `listen` |
