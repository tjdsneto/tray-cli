# Listen hooks (`hooks.json`)

`tray listen` polls your trays in the background and can run **hooks** (arbitrary commands) when something changes. Configuration is a JSON file; state is **in memory only** (no local database—after a restart, the first poll re-seeds so you are not flooded with old events).

**Default path:** `$TRAY_CONFIG_DIR/hooks.json` (see [`internal/config/paths.go`](../../internal/config/paths.go) and env `TRAY_CONFIG_DIR`). Use `tray listen --hooks /path/to/hooks.json` or `--no-hooks` to disable.

## Roles: who does what

- **Tray owner** — triages items **on their tray**: accept, decline, snooze, complete, archive. Only the owner changes those statuses for items on that tray.
- **Contributors / members** — can **add** items to a tray they can access; they do **not** complete someone else’s triage on their behalf.

So: **you** (as owner) **complete** items that others added **to your tray**. Another person does **not** “complete” your item for you—**you** mark it done when you triage. Hooks for **`item.completed`** (and accepted/declined) in this doc refer to **outbox** items: things **you** filed on **someone else’s** tray, and **they** accepted, declined, or completed them.

## Events

| Event | Meaning | Scope |
|--------|---------|--------|
| `item.pending` | A new **pending** line appears (by default: on a tray **you own**, from **someone else**). | `inbox_owned` (default) |
| `item.completed` | An item **you** added on another’s tray moved to **completed** (they finished it). | `outbox` (default) |
| `item.accepted` | An item **you** added on another’s tray was **accepted** by the owner. | `outbox` (default) |
| `item.declined` | An item **you** added was **declined** (see `TRAY_ITEM_DECLINE_REASON`). | `outbox` (default) |

Optional per-rule **`tray`**: tray name, id, or remote alias—limits that rule to one tray.

## Environment variables

Hooks run with your normal environment plus **TRAY\_** variables (canonical names live in [`internal/cli/listenhook/env.go`](../../internal/cli/listenhook/env.go) in the source tree). Commonly:

- **`TRAY_HOOK_EVENT`** — e.g. `item.pending`, `item.accepted`
- **`TRAY_SESSION_USER_ID`** — signed-in user
- **`TRAY_ITEM_ID`**, **`TRAY_ITEM_TRAY_ID`**, **`TRAY_ITEM_TITLE`**, **`TRAY_ITEM_STATUS`**
- **`TRAY_ITEM_ADDED_BY_USER_ID`**, **`TRAY_ITEM_ADDED_BY_DISPLAY_NAME`** (when profiles resolve)
- **`TRAY_ITEM_DECLINE_REASON`** — owner’s decline message when applicable (newlines collapsed to spaces)
- **Timestamps** when set: **`TRAY_ITEM_COMPLETED_AT`**, **`TRAY_ITEM_ACCEPTED_AT`**, **`TRAY_ITEM_DECLINED_AT`**

## Example `hooks.json`

```json
{
  "hooks": [
    {
      "event": "item.pending",
      "command": ["/usr/bin/notify-send", "Tray", "New request on your tray"]
    },
    {
      "event": "item.accepted",
      "command": ["/usr/bin/notify-send", "Tray", "They accepted your item"]
    },
    {
      "event": "item.declined",
      "command": ["/bin/sh", "-c", "/usr/bin/notify-send Tray \"Declined: $TRAY_ITEM_DECLINE_REASON\""]
    },
    {
      "event": "item.completed",
      "command": ["/usr/bin/afplay", "/System/Library/Sounds/Glass.aiff"]
    }
  ]
}
```

## Recipes

**Desktop notification (Linux).** Use `notify-send` as in the example; install `libnotify` where needed.

**Desktop notification (macOS).** Either call `osascript -e 'display notification "…" with title "Tray"'` with env interpolated in a `sh -c` wrapper, or use a third-party notifier.

**Sound only.** Point `command` at `afplay` (macOS) or `paplay` / `ffplay` (Linux) on a short sound file.

**Log to a file.** `["/bin/sh", "-c", "echo \"$(date) $TRAY_HOOK_EVENT $TRAY_ITEM_TITLE\" >> \"$HOME/tray-hooks.log\""]` (adjust `date` flags if you want ISO timestamps—GNU `date -Iseconds` is Linux-only).

**Quiet terminal, hooks only.** `tray listen --quiet` so new rows are not printed; hooks still run.

**JSON / automation.** Combine with `--format json` for the printed stream (hooks are separate); for scripting, prefer the env vars inside the hook command.

**Filter to one tray.** Add `"tray": "work"` (or id / alias) to a rule.

---

CLI flags: `tray listen --help` (`--mode auto|realtime|poll`, `--interval`, `--once` snapshot without hooks, `--hooks`, `--no-hooks`, `--quiet`, `--exec`, `--exec-pattern`).

- `--mode auto` (default): try realtime subscription first; if it fails, fall back to polling.
- `--mode realtime`: require websocket subscription (no fallback).
- `--mode poll`: always poll on `--interval`.
