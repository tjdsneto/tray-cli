# Tray desktop app — design

**Date:** 2026-11-17
**Status:** Pending review
**Platform (v1):** macOS for the menu bar. The window does not depend on AppKit.

## Problem

Tray is a CLI, plus a macOS menu bar that shows an attention badge and copies a handoff line. The inbox itself — trays, items, bodies, and triage — has no window. This spec is the first official desktop app: the window you triage in, and the menu bar living in that app.

## Goals

- A Tauri window (React, Vite, Tailwind) that lists trays and items, reads a body, and adds, accepts, declines, completes, snoozes, and archives.
- The app owns a menu bar with today's behavior, and can open the window.
- One Go process, `tray api`, performs every tray operation by calling the packages the CLI already uses.
- The window and the menu bar share one snapshot.
- Behavior is covered by Go tests. The window test checks which actions a status shows.

## Non-goals (v1)

- Invite, members, join, listen, prune, upgrade, rename, delete, reorder (`item up` / `down`), and remove.
- A sign-in screen. The app uses the CLI's existing session.
- A settings screen. Refresh is 30 seconds.
- Bundling the `tray` binary inside the app. v1 expects `tray` on `PATH`.
- Editing a body after create. Completion notes. A free-form date picker for snooze.
- Tauri end-to-end tests.
- Retiring `tray bar`. It stays for people who only use the CLI.
- The ChatGPT gateway. `tray api` is a different, loopback-only server.

## Who it is for

The official app. One person uses it at first, so v1 cuts the installer and the login screen. The seams those need are the local API and `GET /v1/session`, not a second tray implementation.

## Window

Three panes.

| Pane | Contents |
|------|----------|
| Trays | **Attention** first (the badge count), then local trays, then owned remote trays. A tray's count is its open work. |
| Items | The selected tray's open work, newest first. A field to add a title, with an optional body. |
| Reading | Title, status, body, and the actions that apply. |

**Attention** is the menu-bar set: local open items plus remote pending items on trays you own. Order matches the existing snapshot: tray name A–Z, newest first inside a tray. It has no add field. Adding happens on a selected tray. Ids in the API are the full item id and the same tray-ref the menu bar copies, not a short prefix.

**Open work** is the default list. Local: status `open`. Remote: `pending`, `accepted`, and `snoozed`. **All** is a control on the list header and adds `done`, `declined`, `completed`, and `archived`.

Selecting a row fades the reading pane in. The body is loaded for that row, not with the list.

Actions that do not apply are absent:

| Item | Actions |
|------|---------|
| Local `open` | Complete |
| Remote `pending` | Accept, Decline, Snooze |
| Remote `accepted` | Complete, Archive. The window does not offer Snooze here, even though the CLI can snooze an accepted item. |
| Remote `snoozed` | Archive. The pane shows `snooze_until`. |

Decline reveals an optional reason. Empty is allowed. Snooze offers three presets and sends the resulting RFC3339 instant: one hour from now, tomorrow at 09:00 local, or seven days from now at 09:00 local. Complete does not ask for a message.

Signed out: local trays work, and the remote group is one muted line. Remote fetch failure: the last remote rows stay, with "remote unavailable".

The visual reference is the approved mock: light surface, one dark accent, system font, pill buttons.

## Menu bar

The status item lives in the Tauri app. Clicking it opens the menu, as on macOS today. Rows are the attention snapshot. Clicking a row copies `tray-ref item-id` using the existing clipboard line. **Open** shows the window. So does the Dock icon.

`tray bar` is unchanged. Running it beside the app shows two icons. v1 does not pick a winner.

A second launch of the app focuses the existing window and does not start a second API.

## Architecture

```text
Tauri app (desktop/)                tray api (child process)
  window  ──┐                         loopback HTTP
  menu bar ─┴── Bearer token ───────► existing Go packages
                                      local files + remote API
                                      (same config dir and login as the CLI)
```

On launch the app looks up `tray` on `PATH` and starts `tray api`. The child prints one JSON line on stdout, then logs on stderr:

```json
{"port": 12345, "token": "<per-launch secret>"}
```

The token is not the Supabase session. It dies with the process. The server binds `127.0.0.1` only. The app closes the child's stdin and sends SIGTERM on quit. The child exits on stdin EOF or SIGTERM. If `tray` is missing, the window explains that and does not pretend to have trays.

`tray api` is a thin `net/http` wrapper. Request shapes, response shapes, and the action table above are pure functions under `internal/api`. Handlers call `internal/bar`, `internal/localtray`, and the domain services. They do not shell out to Cobra and they do not import the ChatGPT gateway.

The API keeps one snapshot. It rebuilds every 30 seconds and after each successful write. Rebuild loads local open items and, when signed in, remote items. A failed remote fetch keeps the last remote rows and sets the remote-unavailable line, which is the menu bar's rule today. The window and the menu bar both `GET` that snapshot.

### Routes

All routes require `Authorization: Bearer <token>`.

| Route | Role |
|-------|------|
| `GET /v1/snapshot` | Badge, sections, clipboard lines, remote status |
| `GET /v1/trays` | Local trays and owned remote trays, with counts |
| `GET /v1/items?tray=<ref>&filter=open\|all` | Titles and status, no body |
| `GET /v1/items/{id}` | One item, including body |
| `GET /v1/session` | `{ "signedIn": true \| false }` |
| `POST /v1/items` | Add `{ "tray", "title", "body"? }` |
| `POST /v1/items/{id}/accept` | |
| `POST /v1/items/{id}/decline` | `{ "reason"? }` |
| `POST /v1/items/{id}/complete` | |
| `POST /v1/items/{id}/snooze` | `{ "until": "<RFC3339>" }` |
| `POST /v1/items/{id}/archive` | |

A failed write returns `4xx` or `5xx` and does not change the snapshot. The pane shows `message`. The list stays as it was.

```json
{"error": "signed_out | remote_unavailable | not_found | invalid | unavailable", "message": "<text>"}
```

A rejected token means the child is stale. The app restarts `tray api` once, then surfaces the error if it fails again.

## Errors

- Local store unreadable: the window shows the message. The menu bar keeps the last snapshot.
- Signed out during a remote write: `signed_out`. Local complete still works.
- Empty title or a snooze instant in the past: `invalid`.
- Unknown id: `not_found`.

## Testing

- **Go, pure:** snapshot mapping stays covered by `internal/bar`. New tests for the action table, request and response shapes, and the snooze presets.
- **Go, handlers:** fake local files and a fake remote. A failed remote keeps the last remote rows. A failed write does not replace the snapshot.
- **Window, Vitest:** Attention renders the snapshot. A pending row offers Accept, Decline, and Snooze. An accepted row offers Complete and Archive. A local open row offers Complete.
- No Tauri end-to-end run in v1.

## Docs

A short `docs/user/` page for launching the app and for what the menu bar still copies. Link it from the user README. State that `tray bar` remains, and that both at once means two icons.

## Decisions log

| Topic | Choice |
|-------|--------|
| Shell | Tauri, React, Vite, Tailwind, in `desktop/` |
| Domain | `tray api` child, loopback, per-launch token, existing Go packages |
| Session | CLI login. No sign-in screen. `GET /v1/session` is the seam |
| Binary | `tray` on `PATH`. Bundling comes later |
| Window | Three panes. Attention first. Open work by default. Newest first |
| Body | Loaded when a row is opened |
| Snooze | Three presets. No calendar |
| Decline | Optional reason |
| Menu bar | Same badge and copy behavior, plus Open. `tray bar` stays |
| Refresh | 30 seconds, and after each successful write |
| Tests | Go owns behavior. Vitest owns which buttons show |
